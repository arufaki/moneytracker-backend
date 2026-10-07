package repositories

import (
	"money-tracker-ai/models"

	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(rt *models.RefreshToken) error
	FindByToken(hashedToken string) (*models.RefreshToken, error)
	DeleteByToken(hashedToken string) error
	DeleteByUserID(userID uint) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(rt *models.RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *refreshTokenRepository) FindByToken(hashedToken string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.Where("token = ?", hashedToken).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) DeleteByToken(hashedToken string) error {
	return r.db.Where("token = ?", hashedToken).Delete(&models.RefreshToken{}).Error
}

func (r *refreshTokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}
