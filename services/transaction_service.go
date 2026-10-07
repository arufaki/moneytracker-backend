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

func (s *transactionService) getOrCreateWallet(rawName string, initialBalance float64) (*models.Wallet, error) {
	name := strings.TrimSpace(rawName)
	if name == "" {
		name = "Cash"
	} else {
		name = strings.Title(strings.ToLower(name))
	}

	wallet, err := s.walletRepo.FindByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = &models.Wallet{Name: name, Balance: initialBalance}
			if createErr := s.walletRepo.Create(wallet); createErr != nil {
				return nil, fmt.Errorf("failed to create wallet %s: %w", name, createErr)
			}
			return wallet, nil
		}
		return nil, fmt.Errorf("failed to find wallet %s: %w", name, err)
	}
	return wallet, nil
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
			switch act.Action {
			case models.ActionCreateWallet, models.ActionSetBalance:
				wallet, err := s.getOrCreateWallet(act.Wallet, act.Balance)
				if err != nil {
					return nil, err
				}
				if wallet.Balance != act.Balance {
					if err := s.walletRepo.UpdateBalance(wallet.ID, act.Balance); err != nil {
						return nil, fmt.Errorf("failed to set balance for %s: %w", wallet.Name, err)
					}
					wallet.Balance = act.Balance
					msgs = append(msgs, fmt.Sprintf("Saldo wallet %s berhasil diperbarui menjadi Rp %.0f.", wallet.Name, wallet.Balance))
				} else {
					msgs = append(msgs, fmt.Sprintf("Wallet %s berhasil dibuat dengan saldo Rp %.0f.", wallet.Name, wallet.Balance))
				}
				lastUpdatedWallet = wallet

			case models.ActionDeleteWallet:
				name := strings.Title(strings.ToLower(strings.TrimSpace(act.Wallet)))
				wallet, err := s.walletRepo.FindByName(name)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						msgs = append(msgs, fmt.Sprintf("Wallet %s tidak ditemukan.", name))
					} else {
						return nil, fmt.Errorf("failed to find wallet %s: %w", name, err)
					}
				} else {
					if err := s.walletRepo.Delete(wallet.ID); err != nil {
						return nil, fmt.Errorf("failed to delete wallet %s: %w", name, err)
					}
					msgs = append(msgs, fmt.Sprintf("Wallet %s berhasil dihapus.", wallet.Name))
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
			wallet, err := s.getOrCreateWallet(act.Wallet, 0)
			if err != nil {
				return nil, err
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
				} else if category.BudgetLimit > 0 && totalExpense > category.BudgetLimit {
					warningMsg := fmt.Sprintf("\n⚠️ PERINGATAN: Pengeluaran kategori %s bulan ini sudah mencapai Rp %.0f (Batas Budget: Rp %.0f).",
						category.Name, totalExpense, category.BudgetLimit)
					msgs = append(msgs, warningMsg)
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

		return &models.ChatResponse{
			Success:       true,
			Message:       finalMsg,
			Transaction:   firstTx,
			Transactions:  savedTransactions,
			UpdatedWallet: lastWallet,
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

