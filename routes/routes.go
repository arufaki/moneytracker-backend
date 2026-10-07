package routes

import (
	"time"

	"money-tracker-ai/controllers"
	"money-tracker-ai/middleware"

	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	RootController      *controllers.RootController
	WalletController    *controllers.WalletController
	CategoryController  *controllers.CategoryController
	ChatController      *controllers.ChatController
	AnalyticsController *controllers.AnalyticsController
}

func SetupRoutes(r *gin.Engine, cfg RouterConfig) {
	r.GET("/", cfg.RootController.GetAPIDocumentation)

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

		// Chat route (AI-powered transaction with rate limit & 4KB body limit)
		api.POST("/chat", middleware.RateLimiter(10, time.Minute), middleware.MaxBodySize(4096), cfg.ChatController.Chat)

		// Analytics route
		api.GET("/summary", cfg.AnalyticsController.GetSummary)
	}
}

