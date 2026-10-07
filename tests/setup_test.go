package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"money-tracker-ai/config"
	"money-tracker-ai/controllers"
	"money-tracker-ai/models"
	"money-tracker-ai/repositories"
	"money-tracker-ai/routes"
	"money-tracker-ai/services"
	"money-tracker-ai/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var testRouter *gin.Engine
var mockAIService *mocks.MockAIService
var DefaultTestUser *models.User

func SetupTestDB() {
	_ = godotenv.Load("../.env")
	config.ConnectDatabase()
	_ = config.DB.Migrator().DropTable(
		&models.AILog{},
		&models.Transaction{},
		&models.Wallet{},
		&models.Category{},
		&models.RefreshToken{},
		&models.EmailVerification{},
		&models.User{},
	)
	config.DB.AutoMigrate(
		&models.User{},
		&models.EmailVerification{},
		&models.RefreshToken{},
		&models.Wallet{},
		&models.Category{},
		&models.Transaction{},
		&models.AILog{},
	)
}

func SetupTestRouter() {
	gin.SetMode(gin.TestMode)
	SetupTestDB()

	userRepo := repositories.NewUserRepository(config.DB)
	emailVerifRepo := repositories.NewEmailVerificationRepository(config.DB)
	refreshRepo := repositories.NewRefreshTokenRepository(config.DB)
	walletRepo := repositories.NewWalletRepository(config.DB)
	categoryRepo := repositories.NewCategoryRepository(config.DB)

	emailSvc := services.NewEmailService()
	oauthSvc := services.NewOAuthService()
	authSvc := services.NewAuthService(userRepo, emailVerifRepo, refreshRepo, emailSvc, oauthSvc)
	walletSvc := services.NewWalletService(walletRepo)
	categorySvc := services.NewCategoryService(categoryRepo)
	analyticsSvc := services.NewAnalyticsService(config.DB)

	mockAIService = new(mocks.MockAIService)
	transactionSvc := services.NewTransactionService(mockAIService, walletRepo, categoryRepo, config.DB)

	rootCtrl := controllers.NewRootController()
	authCtrl := controllers.NewAuthController(authSvc, oauthSvc)
	walletCtrl := controllers.NewWalletController(walletSvc)
	categoryCtrl := controllers.NewCategoryController(categorySvc)
	chatCtrl := controllers.NewChatController(transactionSvc)
	analyticsCtrl := controllers.NewAnalyticsController(analyticsSvc)

	testRouter = gin.Default()

	testRouter.GET("/ping", func(c *gin.Context) {
		dbStatus := "connected"
		sqlDB, err := config.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "disconnected"
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"db_status": dbStatus,
		})
	})

	routes.SetupRoutes(testRouter, routes.RouterConfig{
		RootController:      rootCtrl,
		AuthController:      authCtrl,
		WalletController:    walletCtrl,
		CategoryController:  categoryCtrl,
		ChatController:      chatCtrl,
		AnalyticsController: analyticsCtrl,
	})
}

func CleanDatabase() {
	if config.DB != nil {
		config.DB.Exec("TRUNCATE TABLE transactions RESTART IDENTITY CASCADE;")
		config.DB.Exec("TRUNCATE TABLE wallets RESTART IDENTITY CASCADE;")
		config.DB.Exec("TRUNCATE TABLE categories RESTART IDENTITY CASCADE;")
		config.DB.Exec("TRUNCATE TABLE ai_logs RESTART IDENTITY CASCADE;")
		config.DB.Exec("TRUNCATE TABLE refresh_tokens RESTART IDENTITY CASCADE;")
		config.DB.Exec("TRUNCATE TABLE email_verifications RESTART IDENTITY CASCADE;")
		config.DB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE;")
	}

	user := &models.User{
		Email:      "testuser@example.com",
		Name:       "Test User",
		IsVerified: true,
	}
	config.DB.Create(user)
	DefaultTestUser = user
}

func GenerateTestToken(userID uint) string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "supersecretkey"
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := t.SignedString([]byte(secret))
	return tokenStr
}

func DoRequest(method, url string, body interface{}, ip ...string) *httptest.ResponseRecorder {
	var token string
	if DefaultTestUser != nil && DefaultTestUser.ID > 0 {
		token = GenerateTestToken(DefaultTestUser.ID)
	}
	return DoRequestWithToken(method, url, body, token, ip...)
}

func DoRequestWithToken(method, url string, body interface{}, token string, ip ...string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = []byte(v)
		default:
			reqBody, _ = json.Marshal(body)
		}
	}
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if len(ip) > 0 && ip[0] != "" {
		req.Header.Set("X-Forwarded-For", ip[0])
		req.RemoteAddr = ip[0] + ":12345"
	}

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func TestMain(m *testing.M) {
	SetupTestRouter()
	os.Exit(m.Run())
}
