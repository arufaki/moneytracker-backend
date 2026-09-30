package repositories

import (
	"money-tracker-ai/models"

	"gorm.io/gorm"
)

type AILogRepository interface {
	Create(log *models.AILog) error
}

type aiLogRepository struct {
	db *gorm.DB
}

func NewAILogRepository(db *gorm.DB) AILogRepository {
	return &aiLogRepository{db: db}
}

func (r *aiLogRepository) Create(log *models.AILog) error {
	return r.db.Create(log).Error
}
