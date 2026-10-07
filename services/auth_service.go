package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"money-tracker-ai/models"
	"money-tracker-ai/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(email, name, password string) error
	VerifyEmail(token string) error
	Login(email, password string) (accessToken, refreshToken string, err error)
	GoogleLogin(code string) (accessToken, refreshToken string, err error)
	RefreshToken(refreshToken string) (newAccessToken, newRefreshToken string, err error)
	Logout(refreshToken string) error
}

type authService struct {
	userRepo       repositories.UserRepository
	emailVerifRepo repositories.EmailVerificationRepository
	refreshRepo    repositories.RefreshTokenRepository
	emailSvc       EmailService
	oauthSvc       OAuthService
}

func NewAuthService(
	userRepo repositories.UserRepository,
	emailVerifRepo repositories.EmailVerificationRepository,
	refreshRepo repositories.RefreshTokenRepository,
	emailSvc EmailService,
	oauthSvc OAuthService,
) AuthService {
	return &authService{
		userRepo:       userRepo,
		emailVerifRepo: emailVerifRepo,
		refreshRepo:    refreshRepo,
		emailSvc:       emailSvc,
		oauthSvc:       oauthSvc,
	}
}

func (s *authService) Register(email, name, password string) error {
	if email == "" || name == "" || password == "" {
		return errors.New("email, name, and password are required")
	}

	existing, err := s.userRepo.FindByEmail(email)
	if err == nil && existing != nil {
		return errors.New("email is already registered")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	hashStr := string(hashedBytes)

	user := &models.User{
		Email:        email,
		Name:         name,
		PasswordHash: &hashStr,
		IsVerified:   false,
	}
	if err := s.userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}
	verifToken := hex.EncodeToString(tokenBytes)

	ev := &models.EmailVerification{
		UserID:    user.ID,
		Token:     verifToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.emailVerifRepo.Create(ev); err != nil {
		return fmt.Errorf("failed to create verification record: %w", err)
	}

	return s.emailSvc.SendVerificationEmail(user.Email, user.Name, verifToken)
}

func (s *authService) VerifyEmail(token string) error {
	ev, err := s.emailVerifRepo.FindByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired verification token")
		}
		return err
	}

	if time.Now().After(ev.ExpiresAt) {
		s.emailVerifRepo.DeleteByUserID(ev.UserID)
		return errors.New("verification token expired")
	}

	user, err := s.userRepo.FindByID(ev.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	user.IsVerified = true
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}

	_ = s.emailVerifRepo.DeleteByUserID(ev.UserID)
	return nil
}

func (s *authService) Login(email, password string) (string, string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("invalid email or password")
		}
		return "", "", err
	}

	if !user.IsVerified {
		return "", "", errors.New("account email is not verified")
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return "", "", errors.New("please login with Google OAuth")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return "", "", errors.New("invalid email or password")
	}

	return s.issueTokens(user.ID)
}

func (s *authService) GoogleLogin(code string) (string, string, error) {
	userInfo, err := s.oauthSvc.ExchangeCode(code)
	if err != nil {
		return "", "", fmt.Errorf("oauth exchange failed: %w", err)
	}

	var user *models.User
	userByGoogle, err := s.userRepo.FindByGoogleID(userInfo.ID)
	if err == nil && userByGoogle != nil {
		user = userByGoogle
	} else {
		userByEmail, err := s.userRepo.FindByEmail(userInfo.Email)
		if err == nil && userByEmail != nil {
			userByEmail.GoogleID = &userInfo.ID
			userByEmail.IsVerified = true
			if err := s.userRepo.Update(userByEmail); err != nil {
				return "", "", fmt.Errorf("failed to link Google ID: %w", err)
			}
			user = userByEmail
		} else {
			newUser := &models.User{
				Email:      userInfo.Email,
				Name:       userInfo.Name,
				GoogleID:   &userInfo.ID,
				IsVerified: true,
			}
			if err := s.userRepo.Create(newUser); err != nil {
				return "", "", fmt.Errorf("failed to create user via Google: %w", err)
			}
			user = newUser
		}
	}

	return s.issueTokens(user.ID)
}

func (s *authService) RefreshToken(rawRefreshToken string) (string, string, error) {
	hashed := hashToken(rawRefreshToken)
	rt, err := s.refreshRepo.FindByToken(hashed)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("invalid refresh token")
		}
		return "", "", err
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = s.refreshRepo.DeleteByToken(hashed)
		return "", "", errors.New("refresh token expired")
	}

	// Token rotation
	_ = s.refreshRepo.DeleteByToken(hashed)

	return s.issueTokens(rt.UserID)
}

func (s *authService) Logout(rawRefreshToken string) error {
	hashed := hashToken(rawRefreshToken)
	return s.refreshRepo.DeleteByToken(hashed)
}

func (s *authService) issueTokens(userID uint) (string, string, error) {
	accessToken, err := s.issueJWT(userID)
	if err != nil {
		return "", "", err
	}

	rawRefresh, hashedRefresh := generateRefreshTokenPair()
	expiryDays := 7
	if daysStr := os.Getenv("REFRESH_TOKEN_EXPIRY_DAYS"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			expiryDays = d
		}
	}

	rt := &models.RefreshToken{
		UserID:    userID,
		Token:     hashedRefresh,
		ExpiresAt: time.Now().Add(time.Duration(expiryDays) * 24 * time.Hour),
	}
	if err := s.refreshRepo.Create(rt); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, rawRefresh, nil
}

func (s *authService) issueJWT(userID uint) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "defaultsecret"
	}

	expiryMinutes := 15
	if expStr := os.Getenv("JWT_EXPIRY_MINUTES"); expStr != "" {
		if m, err := strconv.Atoi(expStr); err == nil && m > 0 {
			expiryMinutes = m
		}
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(expiryMinutes) * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func generateRefreshTokenPair() (string, string) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	raw := hex.EncodeToString(bytes)
	return raw, hashToken(raw)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
