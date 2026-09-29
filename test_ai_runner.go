//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"money-tracker-ai/config"
	"money-tracker-ai/repositories"
	"money-tracker-ai/services"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables dari .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Fatal("Error: GEMINI_API_KEY belum diset di .env")
	}

	// Connect ke database (diperlukan untuk ai_logs)
	config.ConnectDatabase()

	// Inisialisasi Repository & Service
	aiLogRepo := repositories.NewAILogRepository(config.DB)
	aiSvc := services.NewAIService(aiLogRepo)
	defer aiSvc.Close()

	prompt := "Beli kopi 25k pake Cash"
	fmt.Printf("Mengirim prompt: %q\n", prompt)
	fmt.Println("Tunggu sebentar, sedang diproses Gemini...")

	// Panggil Gemini AI
	result, err := aiSvc.ParseTransactionPrompt(prompt)
	if err != nil {
		log.Fatalf("Failed to parse transaction: %v", err)
	}

	// Tampilkan Hasil
	fmt.Printf("\n===== HASIL EKSTRAKSI AI =====\n")
	fmt.Printf("Amount      : %.0f\n", result.Amount)
	fmt.Printf("Type        : %s\n", result.Type)
	fmt.Printf("Category    : %s\n", result.Category)
	fmt.Printf("Wallet      : %s\n", result.Wallet)
	fmt.Printf("Description : %s\n", result.Description)
	fmt.Println("==============================")
}
