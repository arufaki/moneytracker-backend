package repositories

import (
	"money-tracker-ai/models"

	"gorm.io/gorm"
)

type WalletRepository interface {
	FindAll() ([]models.Wallet, error)
	FindByID(id uint) (*models.Wallet, error)
	Create(wallet *models.Wallet) error
	UpdateBalance(id uint, newBalance float64) error
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) FindAll() ([]models.Wallet, error) {
	var wallets []models.Wallet
	err := r.db.Find(&wallets).Error
	return wallets, err
}

func (r *walletRepository) FindByID(id uint) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.First(&wallet, id).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) Create(wallet *models.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) UpdateBalance(id uint, newBalance float64) error {
	return r.db.Model(&models.Wallet{}).Where("id = ?", id).Update("balance", newBalance).Error
}
