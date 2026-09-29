package main

import (
	"log"
	"net/http"
	"os"

	"money-tracker-ai/config"
	"money-tracker-ai/controllers"
	"money-tracker-ai/repositories"
	"money-tracker-ai/routes"
	"money-tracker-ai/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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
	walletRepo := repositories.NewWalletRepository(config.DB)
	categoryRepo := repositories.NewCategoryRepository(config.DB)

	// Services
	walletSvc := services.NewWalletService(walletRepo)
	categorySvc := services.NewCategoryService(categoryRepo)

	// AI Setup
	aiLogRepo := repositories.NewAILogRepository(config.DB)
	aiSvc := services.NewAIService(aiLogRepo)
	defer aiSvc.Close()
	log.Println("AI Service initialized")

	transactionSvc := services.NewTransactionService(aiSvc, walletRepo, categoryRepo, config.DB)

	// Controllers
	walletCtrl := controllers.NewWalletController(walletSvc)
	categoryCtrl := controllers.NewCategoryController(categorySvc)
	chatCtrl := controllers.NewChatController(transactionSvc)

	// Setup Gin router
	r := gin.Default()

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
		WalletController:   walletCtrl,
		CategoryController: categoryCtrl,
		ChatController:     chatCtrl,
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
