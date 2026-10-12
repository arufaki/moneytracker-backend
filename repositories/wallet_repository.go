package repositories

import (
	"money-tracker-ai/models"

	"gorm.io/gorm"
)

type WalletRepository interface {
	FindAll(userID uint) ([]models.Wallet, error)
	FindByID(id uint, userID uint) (*models.Wallet, error)
	Create(wallet *models.Wallet) error
	UpdateBalance(id uint, userID uint, newBalance float64) error
	FindByName(name string, userID uint) (*models.Wallet, error)
	Delete(id uint, userID uint) error
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) FindAll(userID uint) ([]models.Wallet, error) {
	var wallets []models.Wallet
	err := r.db.Where("user_id = ?", userID).Find(&wallets).Error
	return wallets, err
}

func (r *walletRepository) FindByID(id uint, userID uint) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) Create(wallet *models.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) UpdateBalance(id uint, userID uint, newBalance float64) error {
	result := r.db.Model(&models.Wallet{}).Where("id = ? AND user_id = ?", id, userID).Update("balance", newBalance)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *walletRepository) FindByName(name string, userID uint) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Where("user_id = ? AND LOWER(name) = LOWER(?)", userID, name).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) Delete(id uint, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Wallet{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
