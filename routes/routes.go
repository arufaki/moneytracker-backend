package routes

import (
	"money-tracker-ai/controllers"

	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	WalletController    *controllers.WalletController
	CategoryController  *controllers.CategoryController
	ChatController      *controllers.ChatController
	AnalyticsController *controllers.AnalyticsController
}

func SetupRoutes(r *gin.Engine, cfg RouterConfig) {
	api := r.Group("/api")
	{
		// Wallet routes
		wallets := api.Group("/wallets")
		{
			wallets.GET("", cfg.WalletController.GetAllWallets)
			wallets.GET("/:id", cfg.WalletController.GetWalletByID)
			wallets.POST("", cfg.WalletController.CreateWallet)
		}

		// Category routes
		categories := api.Group("/categories")
		{
			categories.GET("", cfg.CategoryController.GetAllCategories)
			categories.GET("/:id", cfg.CategoryController.GetCategoryByID)
			categories.POST("", cfg.CategoryController.CreateCategory)
		}

		// Chat route (AI-powered transaction)
		api.POST("/chat", cfg.ChatController.Chat)

		// Analytics route
		api.GET("/summary", cfg.AnalyticsController.GetSummary)
	}
}
