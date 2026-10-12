package routes

import (
	"time"

	"money-tracker-ai/controllers"
	"money-tracker-ai/middleware"

	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	RootController      *controllers.RootController
	AuthController      *controllers.AuthController
	WalletController    *controllers.WalletController
	CategoryController  *controllers.CategoryController
	ChatController      *controllers.ChatController
	AnalyticsController *controllers.AnalyticsController
}

func SetupRoutes(r *gin.Engine, cfg RouterConfig) {
	r.GET("/", cfg.RootController.GetAPIDocumentation)

	// Public Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", cfg.AuthController.Register)
		auth.GET("/verify", cfg.AuthController.VerifyEmail)
		auth.POST("/login", cfg.AuthController.Login)
		auth.POST("/refresh", cfg.AuthController.Refresh)
		auth.GET("/google", cfg.AuthController.GoogleRedirect)
		auth.GET("/google/callback", cfg.AuthController.GoogleCallback)

		// Protected Auth route
		auth.POST("/logout", middleware.JWTAuth(), cfg.AuthController.Logout)
	}

	// Protected API routes
	api := r.Group("/api", middleware.JWTAuth())
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
			categories.DELETE("/:id", cfg.CategoryController.DeleteCategory)
		}

		// Chat route (AI-powered transaction with rate limit & 4KB body limit)
		api.POST("/chat", middleware.RateLimiter(10, time.Minute), middleware.MaxBodySize(4096), cfg.ChatController.Chat)

		// Analytics route
		api.GET("/summary", cfg.AnalyticsController.GetSummary)
	}
}
