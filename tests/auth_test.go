package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"money-tracker-ai/config"
	"money-tracker-ai/models"
	"money-tracker-ai/repositories"
	"money-tracker-ai/services"
	"money-tracker-ai/tests/mocks"

	"github.com/stretchr/testify/assert"
)

func TestAuth_Flow(t *testing.T) {
	t.Run("Registrasi manual & verifikasi email & login flow", func(t *testing.T) {
		CleanDatabase()

		// 1. Register
		regBody := map[string]interface{}{
			"email":    "user1@example.com",
			"name":     "User One",
			"password": "password123",
		}
		w := DoRequestWithToken("POST", "/auth/register", regBody, "")
		assert.Equal(t, http.StatusCreated, w.Code)

		// Try registering duplicate email -> error
		wDup := DoRequestWithToken("POST", "/auth/register", regBody, "")
		assert.Equal(t, http.StatusBadRequest, wDup.Code)

		// Try login before verified -> error
		loginBody := map[string]interface{}{
			"email":    "user1@example.com",
			"password": "password123",
		}
		wLoginUnverif := DoRequestWithToken("POST", "/auth/login", loginBody, "")
		assert.Equal(t, http.StatusUnauthorized, wLoginUnverif.Code)

		// 2. Verify email
		var ev models.EmailVerification
		err := config.DB.First(&ev).Error
		assert.NoError(t, err)

		wVerif := DoRequestWithToken("GET", fmt.Sprintf("/auth/verify?token=%s", ev.Token), nil, "")
		assert.Equal(t, http.StatusOK, wVerif.Code)

		// 3. Login after verified
		wLogin := DoRequestWithToken("POST", "/auth/login", loginBody, "")
		assert.Equal(t, http.StatusOK, wLogin.Code)

		var loginResp map[string]interface{}
		json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
		accessToken := loginResp["access_token"].(string)
		refreshToken := loginResp["refresh_token"].(string)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)

		// 4. Refresh token
		refBody := map[string]interface{}{
			"refresh_token": refreshToken,
		}
		wRefresh := DoRequestWithToken("POST", "/auth/refresh", refBody, "")
		assert.Equal(t, http.StatusOK, wRefresh.Code)
		var refResp map[string]interface{}
		json.Unmarshal(wRefresh.Body.Bytes(), &refResp)
		newAccessToken := refResp["access_token"].(string)
		newRefreshToken := refResp["refresh_token"].(string)
		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)

		// Old refresh token rotated out -> should fail if reused
		wOldRefresh := DoRequestWithToken("POST", "/auth/refresh", refBody, "")
		assert.Equal(t, http.StatusUnauthorized, wOldRefresh.Code)

		// 5. Logout
		wLogout := DoRequestWithToken("POST", "/auth/logout", map[string]interface{}{
			"refresh_token": newRefreshToken,
		}, newAccessToken)
		assert.Equal(t, http.StatusOK, wLogout.Code)
	})

	t.Run("JWT Middleware protection", func(t *testing.T) {
		CleanDatabase()
		// Request protected endpoint without token -> 401
		wNoToken := DoRequestWithToken("GET", "/api/wallets", nil, "")
		assert.Equal(t, http.StatusUnauthorized, wNoToken.Code)

		// Request protected endpoint with invalid token -> 401
		wInvalid := DoRequestWithToken("GET", "/api/wallets", nil, "invalid.jwt.token")
		assert.Equal(t, http.StatusUnauthorized, wInvalid.Code)
	})

	t.Run("Isolasi Data antar user", func(t *testing.T) {
		CleanDatabase()

		// User A creates wallet
		userA := &models.User{Email: "usera@example.com", Name: "User A", IsVerified: true}
		config.DB.Create(userA)
		walletA := &models.Wallet{UserID: userA.ID, Name: "Wallet User A", Balance: 100000}
		config.DB.Create(walletA)

		// User B creates wallet
		userB := &models.User{Email: "userb@example.com", Name: "User B", IsVerified: true}
		config.DB.Create(userB)
		walletB := &models.Wallet{UserID: userB.ID, Name: "Wallet User B", Balance: 200000}
		config.DB.Create(walletB)

		tokenA := GenerateTestToken(userA.ID)
		tokenB := GenerateTestToken(userB.ID)

		// User A lists wallets -> sees only Wallet A
		wA := DoRequestWithToken("GET", "/api/wallets", nil, tokenA)
		assert.Equal(t, http.StatusOK, wA.Code)
		var respA map[string]interface{}
		json.Unmarshal(wA.Body.Bytes(), &respA)
		walletsA := respA["data"].([]interface{})
		assert.Len(t, walletsA, 1)
		assert.Equal(t, "Wallet User A", walletsA[0].(map[string]interface{})["name"])

		// User A attempts to access Wallet B by ID -> 404
		wAccessB := DoRequestWithToken("GET", fmt.Sprintf("/api/wallets/%d", walletB.ID), nil, tokenA)
		assert.Equal(t, http.StatusNotFound, wAccessB.Code)

		// User B lists wallets -> sees only Wallet B
		wB := DoRequestWithToken("GET", "/api/wallets", nil, tokenB)
		assert.Equal(t, http.StatusOK, wB.Code)
		var respB map[string]interface{}
		json.Unmarshal(wB.Body.Bytes(), &respB)
		walletsB := respB["data"].([]interface{})
		assert.Len(t, walletsB, 1)
		assert.Equal(t, "Wallet User B", walletsB[0].(map[string]interface{})["name"])
	})

	t.Run("Category CRUD & isolation & global category protection", func(t *testing.T) {
		CleanDatabase()

		// Global Category
		globalCat := &models.Category{Name: "Global Cat", Type: "expense"}
		config.DB.Create(globalCat)

		// User category
		uid := DefaultTestUser.ID
		userCat := &models.Category{UserID: &uid, Name: "User Cat", Type: "expense"}
		config.DB.Create(userCat)

		// User lists categories -> gets global + user categories
		token := GenerateTestToken(DefaultTestUser.ID)
		wList := DoRequestWithToken("GET", "/api/categories", nil, token)
		assert.Equal(t, http.StatusOK, wList.Code)
		var resp map[string]interface{}
		json.Unmarshal(wList.Body.Bytes(), &resp)
		cats := resp["data"].([]interface{})
		assert.Len(t, cats, 2)

		// User cannot delete global category -> error
		wDelGlobal := DoRequestWithToken("DELETE", fmt.Sprintf("/api/categories/%d", globalCat.ID), nil, token)
		assert.Equal(t, http.StatusBadRequest, wDelGlobal.Code)

		// User can delete their own category
		wDelUser := DoRequestWithToken("DELETE", fmt.Sprintf("/api/categories/%d", userCat.ID), nil, token)
		assert.Equal(t, http.StatusOK, wDelUser.Code)
	})

	t.Run("Verifikasi email dengan token expired", func(t *testing.T) {
		CleanDatabase()
		user := &models.User{Email: "exp@example.com", Name: "Exp User", IsVerified: false}
		config.DB.Create(user)
		ev := &models.EmailVerification{
			UserID:    user.ID,
			Token:     "expired-token",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		config.DB.Create(ev)

		w := DoRequestWithToken("GET", "/auth/verify?token=expired-token", nil, "")
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "verification token expired", resp["error"])
	})

	t.Run("Refresh token expired", func(t *testing.T) {
		CleanDatabase()
		user := &models.User{Email: "exp-refresh@example.com", Name: "Exp Refresh", IsVerified: true}
		config.DB.Create(user)

		sum := sha256.Sum256([]byte("expired-raw-token"))
		hashedToken := hex.EncodeToString(sum[:])
		rt := &models.RefreshToken{
			UserID:    user.ID,
			Token:     hashedToken,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		config.DB.Create(rt)

		w := DoRequestWithToken("POST", "/auth/refresh", map[string]interface{}{
			"refresh_token": "expired-raw-token",
		}, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "refresh token expired", resp["error"])
	})

	t.Run("Google OAuth callback (user baru, user linking, user sudah ada)", func(t *testing.T) {
		CleanDatabase()

		mockOAuth := new(mocks.MockOAuthService)
		userRepo := repositories.NewUserRepository(config.DB)
		emailVerifRepo := repositories.NewEmailVerificationRepository(config.DB)
		refreshRepo := repositories.NewRefreshTokenRepository(config.DB)
		emailSvc := services.NewEmailService()
		authSvcWithMock := services.NewAuthService(userRepo, emailVerifRepo, refreshRepo, emailSvc, mockOAuth)

		// 1. Google OAuth user baru
		mockOAuth.On("ExchangeCode", "code-new").Return(&services.GoogleUserInfo{
			ID:    "google-id-1",
			Email: "gnew@example.com",
			Name:  "Google New",
		}, nil).Once()

		acc1, ref1, err := authSvcWithMock.GoogleLogin("code-new")
		assert.NoError(t, err)
		assert.NotEmpty(t, acc1)
		assert.NotEmpty(t, ref1)

		createdUser, err := userRepo.FindByEmail("gnew@example.com")
		assert.NoError(t, err)
		assert.True(t, createdUser.IsVerified)
		assert.Equal(t, "google-id-1", *createdUser.GoogleID)

		// 2. Google OAuth user lama linking (email sudah terdaftar via manual tapi belum linking)
		manualUser := &models.User{
			Email:      "glink@example.com",
			Name:       "Manual Link",
			IsVerified: true,
		}
		config.DB.Create(manualUser)

		mockOAuth.On("ExchangeCode", "code-link").Return(&services.GoogleUserInfo{
			ID:    "google-id-2",
			Email: "glink@example.com",
			Name:  "Manual Link",
		}, nil).Once()

		acc2, ref2, err := authSvcWithMock.GoogleLogin("code-link")
		assert.NoError(t, err)
		assert.NotEmpty(t, acc2)
		assert.NotEmpty(t, ref2)

		linkedUser, err := userRepo.FindByEmail("glink@example.com")
		assert.NoError(t, err)
		assert.Equal(t, manualUser.ID, linkedUser.ID)
		assert.Equal(t, "google-id-2", *linkedUser.GoogleID)

		// 3. Google OAuth user sudah ada (login langsung by GoogleID)
		mockOAuth.On("ExchangeCode", "code-existing").Return(&services.GoogleUserInfo{
			ID:    "google-id-1",
			Email: "gnew@example.com",
			Name:  "Google New",
		}, nil).Once()

		acc3, ref3, err := authSvcWithMock.GoogleLogin("code-existing")
		assert.NoError(t, err)
		assert.NotEmpty(t, acc3)
		assert.NotEmpty(t, ref3)
	})
}
