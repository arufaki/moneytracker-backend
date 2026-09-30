package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"money-tracker-ai/config"
	"money-tracker-ai/controllers"
	"money-tracker-ai/models"
	"money-tracker-ai/repositories"
	"money-tracker-ai/routes"
	"money-tracker-ai/services"
	"money-tracker-ai/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
	"testing"
)

var testRouter *gin.Engine
var mockAIService *mocks.MockAIService

func SetupTestDB() {
	_ = godotenv.Load("../.env")
	config.ConnectDatabase()
	config.DB.AutoMigrate(
		&models.Wallet{},
		&models.Category{},
		&models.Transaction{},
		&models.AILog{},
	)
}

func SetupTestRouter() {
	gin.SetMode(gin.TestMode)
	SetupTestDB()

	walletRepo := repositories.NewWalletRepository(config.DB)
	categoryRepo := repositories.NewCategoryRepository(config.DB)

	walletSvc := services.NewWalletService(walletRepo)
	categorySvc := services.NewCategoryService(categoryRepo)
	analyticsSvc := services.NewAnalyticsService(config.DB)

	mockAIService = new(mocks.MockAIService)
	transactionSvc := services.NewTransactionService(mockAIService, walletRepo, categoryRepo, config.DB)

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
	}
}

func DoRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
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

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func TestMain(m *testing.M) {
	SetupTestRouter()
	os.Exit(m.Run())
}

