package services

import (
	"errors"
	"money-tracker-ai/models"
	"money-tracker-ai/repositories"
)

type CategoryService interface {
	GetAllCategories(userID uint) ([]models.Category, error)
	GetCategoryByID(id uint, userID uint) (*models.Category, error)
	CreateCategory(userID uint, name string, categoryType string) (*models.Category, error)
	DeleteCategory(id uint, userID uint) error
}

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) GetAllCategories(userID uint) ([]models.Category, error) {
	return s.repo.FindAll(userID)
}

func (s *categoryService) GetCategoryByID(id uint, userID uint) (*models.Category, error) {
	return s.repo.FindByID(id, userID)
}

func (s *categoryService) CreateCategory(userID uint, name string, categoryType string) (*models.Category, error) {
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}
	if categoryType != "income" && categoryType != "expense" {
		return nil, errors.New("type must be 'income' or 'expense'")
	}

	existing, _ := s.repo.FindByName(name, userID)
	if existing != nil {
		return nil, errors.New("category already exists")
	}

	uid := userID
	category := &models.Category{
		UserID: &uid,
		Name:   name,
		Type:   categoryType,
	}
	err := s.repo.Create(category)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (s *categoryService) DeleteCategory(id uint, userID uint) error {
	cat, err := s.repo.FindByID(id, userID)
	if err != nil {
		return errors.New("category not found")
	}
	if cat.UserID == nil {
		return errors.New("cannot delete global default category")
	}
	return s.repo.Delete(id, userID)
}
