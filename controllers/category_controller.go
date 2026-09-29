package controllers

import (
	"errors"
	"money-tracker-ai/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryController struct {
	service services.CategoryService
}

func NewCategoryController(service services.CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

// GetAllCategories godoc
// GET /api/categories
func (ctrl *CategoryController) GetAllCategories(c *gin.Context) {
	categories, err := ctrl.service.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// GetCategoryByID godoc
// GET /api/categories/:id
func (ctrl *CategoryController) GetCategoryByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	category, err := ctrl.service.GetCategoryByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": category})
}

// CreateCategory godoc
// POST /api/categories
// Body: { "name": "Investasi", "type": "income" }
func (ctrl *CategoryController) CreateCategory(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
		Type string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := ctrl.service.CreateCategory(body.Name, body.Type)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": category, "message": "Category created successfully"})
}
