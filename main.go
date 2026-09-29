package main

import (
	"log"
	"net/http"
	"os"

	"money-tracker-ai/config"

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

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
