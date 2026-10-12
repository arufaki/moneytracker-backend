package main

import (
	"log"
	"net/http"
	"os"

	"money-tracker-ai/config"
	"money-tracker-ai/controllers"
	_ "money-tracker-ai/docs"
	"money-tracker-ai/middleware"
	"money-tracker-ai/repositories"
	"money-tracker-ai/routes"
	"money-tracker-ai/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           MoneyTracker API
// @version         1.0
// @description     REST API untuk aplikasi pencatat keuangan berbasis AI.

// @host      localhost:8080
// @BasePath  /api

// @schemes http https
func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, relying on environment variables")
	}

	// Initialize database connection
	config.ConnectDatabase()

	// Run migrations and seed default data
	config.MigrateAndSeed(config.DB)

	// --- Dependency Injection (manual wiring) ---
	// Repositories
	userRepo := repositories.NewUserRepository(config.DB)
	emailVerifRepo := repositories.NewEmailVerificationRepository(config.DB)
	refreshRepo := repositories.NewRefreshTokenRepository(config.DB)
	walletRepo := repositories.NewWalletRepository(config.DB)
	categoryRepo := repositories.NewCategoryRepository(config.DB)

	// Services
	emailSvc := services.NewEmailService()
	oauthSvc := services.NewOAuthService()
	authSvc := services.NewAuthService(userRepo, emailVerifRepo, refreshRepo, emailSvc, oauthSvc)
	walletSvc := services.NewWalletService(walletRepo)
	categorySvc := services.NewCategoryService(categoryRepo)

	// AI Setup
	aiLogRepo := repositories.NewAILogRepository(config.DB)
	aiSvc := services.NewAIService(aiLogRepo)
	defer aiSvc.Close()
	log.Println("AI Service initialized")

	transactionSvc := services.NewTransactionService(aiSvc, walletRepo, categoryRepo, config.DB)
	analyticsSvc := services.NewAnalyticsService(config.DB)

	// Controllers
	rootCtrl := controllers.NewRootController()
	authCtrl := controllers.NewAuthController(authSvc, oauthSvc)
	walletCtrl := controllers.NewWalletController(walletSvc)
	categoryCtrl := controllers.NewCategoryController(categorySvc)
	chatCtrl := controllers.NewChatController(transactionSvc)
	analyticsCtrl := controllers.NewAnalyticsController(analyticsSvc)

	// Setup Gin router
	r := gin.Default()
	r.Use(middleware.MaxBodySize(1024 * 1024)) // Global 1MB body limit

	// Swagger UI route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Healthcheck endpoint
	r.GET("/ping", func(c *gin.Context) {
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

	// Register all API routes
	routes.SetupRoutes(r, routes.RouterConfig{
		RootController:      rootCtrl,
		AuthController:      authCtrl,
		WalletController:    walletCtrl,
		CategoryController:  categoryCtrl,
		ChatController:      chatCtrl,
		AnalyticsController: analyticsCtrl,
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
