package config

import (
	"fmt"
	"log"
	"os"

	"money-tracker-ai/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {

	var dsn string

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL != "" {
		dsn = databaseURL 
	} else {
		sslmode := os.Getenv("DB_SSLMODE")
		if sslmode == "" {
			sslmode = "disable"
		}

			dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
			sslmode,
		)
	}


	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database!", err)
	}

	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
	}

	fmt.Println("Database connection successfully opened")
	DB = database
}

// MigrateAndSeed menjalankan auto-migration dan seeding data default
func MigrateAndSeed(db *gorm.DB) {
	// Auto-migrate semua model
	err := db.AutoMigrate(
		&models.Wallet{},
		&models.Category{},
		&models.Transaction{},
		&models.AILog{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	fmt.Println("Database migrated successfully")

	// Seed Categories default jika tabel masih kosong
	var categoryCount int64
	db.Model(&models.Category{}).Count(&categoryCount)
	if categoryCount == 0 {
		defaultCategories := []models.Category{
			{Name: "Makanan", Type: "expense"},
			{Name: "Transportasi", Type: "expense"},
			{Name: "Gaji", Type: "income"},
			{Name: "Hiburan", Type: "expense"},
			{Name: "Belanja", Type: "expense"},
			{Name: "Tagihan", Type: "expense"},
		}
		db.Create(&defaultCategories)
		fmt.Println("Categories seeded successfully")
	}

	// Seed Wallets default jika tabel masih kosong
	var walletCount int64
	db.Model(&models.Wallet{}).Count(&walletCount)
	if walletCount == 0 {
		defaultWallets := []models.Wallet{
			{Name: "Cash", Balance: 0},
		}
		db.Create(&defaultWallets)
		fmt.Println("Wallets seeded successfully")
	}
}
