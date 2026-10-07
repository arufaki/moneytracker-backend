package controllers

import (
	"errors"
	"log"
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

func (ctrl *CategoryController) GetAllCategories(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	categories, err := ctrl.service.GetAllCategories(userID)
	if err != nil {
		log.Printf("[ERROR] GetAllCategories: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

func (ctrl *CategoryController) GetCategoryByID(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	category, err := ctrl.service.GetCategoryByID(uint(id), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		} else {
			log.Printf("[ERROR] GetCategoryByID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": category})
}

func (ctrl *CategoryController) CreateCategory(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var body struct {
		Name string `json:"name" binding:"required,max=100"`
		Type string `json:"type" binding:"required,oneof=income expense"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[SECURITY] Validation failure on POST /api/categories from IP %s", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body or missing required fields"})
		return
	}

	category, err := ctrl.service.CreateCategory(userID, body.Name, body.Type)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": category, "message": "Category created successfully"})
}

func (ctrl *CategoryController) DeleteCategory(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	if err := ctrl.service.DeleteCategory(uint(id), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
