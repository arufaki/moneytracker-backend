package repositories

import (
	"money-tracker-ai/models"

	"gorm.io/gorm"
)

type EmailVerificationRepository interface {
	Create(ev *models.EmailVerification) error
	FindByToken(token string) (*models.EmailVerification, error)
	DeleteByUserID(userID uint) error
}

type emailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(db *gorm.DB) EmailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

func (r *emailVerificationRepository) Create(ev *models.EmailVerification) error {
	return r.db.Create(ev).Error
}

func (r *emailVerificationRepository) FindByToken(token string) (*models.EmailVerification, error) {
	var ev models.EmailVerification
	err := r.db.Where("token = ?", token).First(&ev).Error
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func (r *emailVerificationRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.EmailVerification{}).Error
}
