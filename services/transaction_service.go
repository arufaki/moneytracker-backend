package services

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"

	"money-tracker-ai/models"
	"money-tracker-ai/repositories"
)

type TransactionService interface {
	ProcessChatMessage(userMessage string) (*models.ChatResponse, error)
}

type transactionService struct {
	aiSvc        AIService
	walletRepo   repositories.WalletRepository
	categoryRepo repositories.CategoryRepository
	db           *gorm.DB
	mu           sync.Mutex
}

func NewTransactionService(
	aiSvc AIService,
	walletRepo repositories.WalletRepository,
	categoryRepo repositories.CategoryRepository,
	db *gorm.DB,
) TransactionService {
	return &transactionService{
		aiSvc:        aiSvc,
		walletRepo:   walletRepo,
		categoryRepo: categoryRepo,
		db:           db,
	}
}

func (s *transactionService) ProcessChatMessage(userMessage string) (*models.ChatResponse, error) {
	// --- Step 1: Parse pesan menggunakan AI ---
	parsed, err := s.aiSvc.ParseTransactionPrompt(userMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}

	// --- Step 2: Cari atau buat Wallet berdasarkan nama dari AI ---
	// Lock SEBELUM FindByName agar tidak ada race condition
	s.mu.Lock()
	defer s.mu.Unlock()

	wallet, err := s.walletRepo.FindByName(parsed.Wallet)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Wallet belum ada, buat baru dengan saldo awal 0
			wallet = &models.Wallet{Name: parsed.Wallet, Balance: 0}
			if createErr := s.walletRepo.Create(wallet); createErr != nil {
				return nil, fmt.Errorf("failed to create wallet: %w", createErr)
			}
		} else {
			return nil, fmt.Errorf("failed to find wallet: %w", err)
		}
	}

	// --- Step 3: Cari atau buat Category berdasarkan nama dari AI ---
	category, err := s.categoryRepo.FindByName(parsed.Category)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Kategori belum ada, buat baru
			category = &models.Category{
				Name: parsed.Category,
				Type: string(parsed.Type),
			}
			if createErr := s.categoryRepo.Create(category); createErr != nil {
				return nil, fmt.Errorf("failed to create category: %w", createErr)
			}
		} else {
			return nil, fmt.Errorf("failed to find category: %w", err)
		}
	}

	// --- Step 4: Jalankan transaksi DB (atomic) ---
	var savedTransaction *models.Transaction

	dbErr := s.db.Transaction(func(tx *gorm.DB) error {
		// 4a. Hitung saldo baru
		var newBalance float64
		switch string(parsed.Type) {
		case "expense":
			if wallet.Balance < parsed.Amount {
				return errors.New("insufficient balance")
			}
			newBalance = wallet.Balance - parsed.Amount
		case "income":
			newBalance = wallet.Balance + parsed.Amount
		default:
			return fmt.Errorf("invalid transaction type: %s", parsed.Type)
		}

		// 4b. Update saldo wallet di DB
		if err := tx.Model(&models.Wallet{}).Where("id = ?", wallet.ID).Update("balance", newBalance).Error; err != nil {
			return fmt.Errorf("failed to update wallet balance: %w", err)
		}
		wallet.Balance = newBalance

		// 4c. Simpan record transaksi baru
		savedTransaction = &models.Transaction{
			WalletID:    wallet.ID,
			CategoryID:  category.ID,
			Amount:      parsed.Amount,
			Type:        string(parsed.Type),
			Description: parsed.Description,
		}
		if err := tx.Create(savedTransaction).Error; err != nil {
			return fmt.Errorf("failed to save transaction: %w", err)
		}

		return nil // nil = commit
	})

	if dbErr != nil {
		return nil, dbErr
	}

	// --- Step 5: Susun pesan konfirmasi ramah ---
	actionWord := "pengeluaran"
	if string(parsed.Type) == "income" {
		actionWord = "pemasukan"
	}

	confirmationMsg := fmt.Sprintf(
		"Berhasil mencatat %s %s sebesar Rp %.0f dari %s. Sisa saldo %s kamu sekarang Rp %.0f.",
		actionWord,
		category.Name,
		parsed.Amount,
		wallet.Name,
		wallet.Name,
		wallet.Balance,
	)

	// --- Step 6: Budget Warning ---
	if string(parsed.Type) == "expense" {
		now := time.Now()
		startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		nextMonthStart := startDate.AddDate(0, 1, 0)

		var totalExpense float64
		if err := s.db.Model(&models.Transaction{}).
			Where("category_id = ? AND type = ? AND date >= ? AND date < ?", category.ID, "expense", startDate, nextMonthStart).
			Select("COALESCE(SUM(amount), 0)").Scan(&totalExpense).Error; err != nil {
			log.Printf("[WARN] failed to check budget total expense: %v", err)
		} else {
			var latestCategory models.Category
			if err := s.db.First(&latestCategory, category.ID).Error; err != nil {
				log.Printf("[WARN] failed to fetch latest category for budget check: %v", err)
			} else if latestCategory.BudgetLimit > 0 && totalExpense > latestCategory.BudgetLimit {
				warningMsg := fmt.Sprintf("\n⚠️ PERINGATAN: Pengeluaran kategori %s bulan ini sudah mencapai Rp %.0f (Batas Budget: Rp %.0f).",
					latestCategory.Name, totalExpense, latestCategory.BudgetLimit)
				confirmationMsg += warningMsg
			}
		}
	}

	return &models.ChatResponse{
		Success:       true,
		Message:       confirmationMsg,
		Transaction:   savedTransaction,
		UpdatedWallet: wallet,
		ParsedData:    parsed,
	}, nil
}
