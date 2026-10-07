package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
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

func formatTitleCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "Cash"
	}
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

func (s *transactionService) ProcessChatMessage(userMessage string) (*models.ChatResponse, error) {
	intent, err := s.aiSvc.ParseIntentPrompt(userMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse intent: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch intent.Intent {
	case models.IntentManageWallet:
		var lastUpdatedWallet *models.Wallet
		var msgs []string

		for _, act := range intent.Actions {
			walletName := formatTitleCase(act.Wallet)
			switch act.Action {
			case models.ActionCreateWallet, models.ActionSetBalance:
				wallet, err := s.walletRepo.FindByName(walletName)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						wallet = &models.Wallet{Name: walletName, Balance: act.Balance}
						if err := s.walletRepo.Create(wallet); err != nil {
							return nil, fmt.Errorf("failed to create wallet %s: %w", walletName, err)
						}
						msgs = append(msgs, fmt.Sprintf("Wallet %s berhasil dibuat dengan saldo Rp %.0f.", wallet.Name, wallet.Balance))
					} else {
						return nil, fmt.Errorf("failed to find wallet %s: %w", walletName, err)
					}
				} else {
					if err := s.walletRepo.UpdateBalance(wallet.ID, act.Balance); err != nil {
						return nil, fmt.Errorf("failed to set balance for %s: %w", walletName, err)
					}
					wallet.Balance = act.Balance
					msgs = append(msgs, fmt.Sprintf("Saldo wallet %s berhasil diperbarui menjadi Rp %.0f.", wallet.Name, wallet.Balance))
				}
				lastUpdatedWallet = wallet

			case models.ActionDeleteWallet:
				wallet, err := s.walletRepo.FindByName(walletName)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						msgs = append(msgs, fmt.Sprintf("Wallet %s tidak ditemukan.", walletName))
					} else {
						return nil, fmt.Errorf("failed to find wallet %s: %w", walletName, err)
					}
				} else {
					if err := s.walletRepo.Delete(wallet.ID); err != nil {
						return nil, fmt.Errorf("failed to delete wallet %s: %w", walletName, err)
					}
					msgs = append(msgs, fmt.Sprintf("Wallet %s berhasil dihapus.", walletName))
				}
			}
		}

		return &models.ChatResponse{
			Success:       true,
			Message:       strings.Join(msgs, " "),
			UpdatedWallet: lastUpdatedWallet,
			ParsedIntent:  intent,
		}, nil

	case models.IntentRecordTransaction:
		var savedTransactions []*models.Transaction
		var lastWallet *models.Wallet
		var msgs []string

		for _, act := range intent.Actions {
			walletName := formatTitleCase(act.Wallet)
			wallet, err := s.walletRepo.FindByName(walletName)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					wallet = &models.Wallet{Name: walletName, Balance: 0}
					if createErr := s.walletRepo.Create(wallet); createErr != nil {
						return nil, fmt.Errorf("failed to create wallet: %w", createErr)
					}
				} else {
					return nil, fmt.Errorf("failed to find wallet: %w", err)
				}
			}

			catName := strings.TrimSpace(act.Category)
			if catName == "" {
				catName = "Lainnya"
			}
			category, err := s.categoryRepo.FindByName(catName)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					category = &models.Category{
						Name: catName,
						Type: string(act.Type),
					}
					if createErr := s.categoryRepo.Create(category); createErr != nil {
						return nil, fmt.Errorf("failed to create category: %w", createErr)
					}
				} else {
					return nil, fmt.Errorf("failed to find category: %w", err)
				}
			}

			var savedTx *models.Transaction
			dbErr := s.db.Transaction(func(tx *gorm.DB) error {
				var newBalance float64
				switch string(act.Type) {
				case "expense":
					if wallet.Balance < act.Amount {
						return errors.New("insufficient balance")
					}
					newBalance = wallet.Balance - act.Amount
				case "income":
					newBalance = wallet.Balance + act.Amount
				default:
					return fmt.Errorf("invalid transaction type: %s", act.Type)
				}

				if err := tx.Model(&models.Wallet{}).Where("id = ?", wallet.ID).Update("balance", newBalance).Error; err != nil {
					return fmt.Errorf("failed to update wallet balance: %w", err)
				}
				wallet.Balance = newBalance

				savedTx = &models.Transaction{
					WalletID:    wallet.ID,
					CategoryID:  category.ID,
					Amount:      act.Amount,
					Type:        string(act.Type),
					Description: act.Description,
				}
				if err := tx.Create(savedTx).Error; err != nil {
					return fmt.Errorf("failed to save transaction: %w", err)
				}
				return nil
			})

			if dbErr != nil {
				return nil, dbErr
			}

			savedTransactions = append(savedTransactions, savedTx)
			lastWallet = wallet

			actionWord := "pengeluaran"
			if string(act.Type) == "income" {
				actionWord = "pemasukan"
			}
			msgs = append(msgs, fmt.Sprintf("Berhasil mencatat %s %s sebesar Rp %.0f dari %s.", actionWord, category.Name, act.Amount, wallet.Name))

			if string(act.Type) == "expense" {
				now := time.Now()
				startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
				nextMonthStart := startDate.AddDate(0, 1, 0)

				var totalExpense float64
				if err := s.db.Model(&models.Transaction{}).
					Where("category_id = ? AND type = ? AND created_at >= ? AND created_at < ?", category.ID, "expense", startDate, nextMonthStart).
					Select("COALESCE(SUM(amount), 0)").Scan(&totalExpense).Error; err != nil {
					log.Printf("[WARN] failed to check budget total expense: %v", err)
				} else {
					var latestCategory models.Category
					if err := s.db.First(&latestCategory, category.ID).Error; err != nil {
						log.Printf("[WARN] failed to fetch latest category for budget check: %v", err)
					} else if latestCategory.BudgetLimit > 0 && totalExpense > latestCategory.BudgetLimit {
						warningMsg := fmt.Sprintf("\n⚠️ PERINGATAN: Pengeluaran kategori %s bulan ini sudah mencapai Rp %.0f (Batas Budget: Rp %.0f).",
							latestCategory.Name, totalExpense, latestCategory.BudgetLimit)
						msgs = append(msgs, warningMsg)
					}
				}
			}
		}

		firstTx := (*models.Transaction)(nil)
		if len(savedTransactions) > 0 {
			firstTx = savedTransactions[0]
		}

		finalMsg := strings.Join(msgs, " ")
		if lastWallet != nil {
			finalMsg += fmt.Sprintf(" Sisa saldo %s kamu sekarang Rp %.0f.", lastWallet.Name, lastWallet.Balance)
		}

		firstParsed := (*models.ParsedTransaction)(nil)
		if len(intent.Actions) > 0 {
			a := intent.Actions[0]
			firstParsed = &models.ParsedTransaction{
				Amount:      a.Amount,
				Type:        a.Type,
				Category:    a.Category,
				Wallet:      formatTitleCase(a.Wallet),
				Description: a.Description,
			}
		}

		return &models.ChatResponse{
			Success:       true,
			Message:       finalMsg,
			Transaction:   firstTx,
			Transactions:  savedTransactions,
			UpdatedWallet: lastWallet,
			ParsedData:    firstParsed,
			ParsedIntent:  intent,
		}, nil

	default:
		return &models.ChatResponse{
			Success:      false,
			Message:      "Maaf, saya tidak mengerti perintah tersebut.",
			ParsedIntent: intent,
		}, nil
	}
}

