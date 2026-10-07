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
	FindByName(name string) (*models.Wallet, error)
	Delete(id uint) error
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
	result := r.db.Model(&models.Wallet{}).Where("id = ?", id).Update("balance", newBalance)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *walletRepository) FindByName(name string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Wallet{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

