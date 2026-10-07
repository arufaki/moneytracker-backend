package services

import (
	"errors"
	"fmt"
	"money-tracker-ai/models"
	"money-tracker-ai/repositories"
	"sync"
)

type WalletService interface {
	GetAllWallets(userID uint) ([]models.Wallet, error)
	GetWalletByID(id uint, userID uint) (*models.Wallet, error)
	CreateWallet(userID uint, name string, initialBalance float64) (*models.Wallet, error)
	UpdateWalletBalance(id uint, userID uint, amount float64, transactionType string) error
}

type walletService struct {
	repo repositories.WalletRepository
	mu   *sync.Mutex
}

func NewWalletService(repo repositories.WalletRepository) WalletService {
	return &walletService{repo: repo, mu: &sync.Mutex{}}
}

func (s *walletService) GetAllWallets(userID uint) ([]models.Wallet, error) {
	return s.repo.FindAll(userID)
}

func (s *walletService) GetWalletByID(id uint, userID uint) (*models.Wallet, error) {
	return s.repo.FindByID(id, userID)
}

func (s *walletService) CreateWallet(userID uint, name string, initialBalance float64) (*models.Wallet, error) {
	if name == "" {
		return nil, errors.New("wallet name cannot be empty")
	}
	if initialBalance < 0 {
		return nil, errors.New("initial balance cannot be negative")
	}
	wallet := &models.Wallet{
		UserID:  userID,
		Name:    name,
		Balance: initialBalance,
	}
	err := s.repo.Create(wallet)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *walletService) UpdateWalletBalance(id uint, userID uint, amount float64, transactionType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wallet, err := s.repo.FindByID(id, userID)
	if err != nil {
		return errors.New("wallet not found")
	}

	switch transactionType {
	case "income":
		wallet.Balance += amount
	case "expense":
		if wallet.Balance < amount {
			return errors.New("insufficient balance")
		}
		wallet.Balance -= amount
	default:
		return errors.New("invalid transaction type, must be 'income' or 'expense'")
	}

	err = s.repo.UpdateBalance(id, userID, wallet.Balance)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance for id %d: %w", id, err)
	}
	return nil
}
