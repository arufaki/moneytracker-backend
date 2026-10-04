package handler

import (
	"net/http"
	"sync"

	"money-tracker-ai/config"
	"money-tracker-ai/controllers"
	_ "money-tracker-ai/docs"
	"money-tracker-ai/repositories"
	"money-tracker-ai/routes"
	"money-tracker-ai/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	rootCtrl := controllers.NewRootController()
	walletCtrl := controllers.NewWalletController(walletSvc)
	categoryCtrl := controllers.NewCategoryController(categorySvc)
	chatCtrl := controllers.NewChatController(transactionSvc)
	analyticsCtrl := controllers.NewAnalyticsController(analyticsSvc)

	// Swagger UI route
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Healthcheck endpoint
	app.GET("/ping", func(c *gin.Context) {
		dbStatus := "connected"
		if config.DB != nil {
			sqlDB, err := config.DB.DB()
			if err != nil || sqlDB.Ping() != nil {
				dbStatus = "disconnected"
			}
		} else {
			dbStatus = "disconnected"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"db_status": dbStatus,
		})
	})

	routes.SetupRoutes(app, routes.RouterConfig{
		RootController:      rootCtrl,
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
