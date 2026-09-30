package handler

import (
	"net/http"
	"sync"

	"money-tracker-ai/config"
	"money-tracker-ai/controllers"
	"money-tracker-ai/repositories"
	"money-tracker-ai/routes"
	"money-tracker-ai/services"

	"github.com/gin-gonic/gin"
)

var (
	app  *gin.Engine
	once sync.Once
)

func initApp() {
	gin.SetMode(gin.ReleaseMode)
	app = gin.New()
	app.Use(gin.Recovery())

	config.ConnectDatabase()
	config.MigrateAndSeed(config.DB)

	walletRepo := repositories.NewWalletRepository(config.DB)
	categoryRepo := repositories.NewCategoryRepository(config.DB)

	walletSvc := services.NewWalletService(walletRepo)
	categorySvc := services.NewCategoryService(categoryRepo)

	aiLogRepo := repositories.NewAILogRepository(config.DB)
	aiSvc := services.NewAIService(aiLogRepo)

	transactionSvc := services.NewTransactionService(aiSvc, walletRepo, categoryRepo, config.DB)
	analyticsSvc := services.NewAnalyticsService(config.DB)

	walletCtrl := controllers.NewWalletController(walletSvc)
	categoryCtrl := controllers.NewCategoryController(categorySvc)
	chatCtrl := controllers.NewChatController(transactionSvc)
	analyticsCtrl := controllers.NewAnalyticsController(analyticsSvc)

	routes.SetupRoutes(app, routes.RouterConfig{
		WalletController:    walletCtrl,
		CategoryController:  categoryCtrl,
		ChatController:      chatCtrl,
		AnalyticsController: analyticsCtrl,
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		initApp()
	})
	app.ServeHTTP(w, r)
}
