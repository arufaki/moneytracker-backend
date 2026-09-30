package services

import (
	"errors"
	"money-tracker-ai/models"
	"money-tracker-ai/repositories"
)

type CategoryService interface {
	GetAllCategories() ([]models.Category, error)
	GetCategoryByID(id uint) (*models.Category, error)
	CreateCategory(name string, categoryType string) (*models.Category, error)
}

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) GetAllCategories() ([]models.Category, error) {
	return s.repo.FindAll()
}

func (s *categoryService) GetCategoryByID(id uint) (*models.Category, error) {
	return s.repo.FindByID(id)
}

func (s *categoryService) CreateCategory(name string, categoryType string) (*models.Category, error) {
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}
	if categoryType != "income" && categoryType != "expense" {
		return nil, errors.New("type must be 'income' or 'expense'")
	}

	category := &models.Category{
		Name: name,
		Type: categoryType,
	}
	err := s.repo.Create(category)
	if err != nil {
		return nil, err
	}
	return category, nil
}
