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
// @Summary     Get all categories
// @Description Mengambil semua kategori transaksi
// @Tags        Categories
// @Produce     json
// @Success     200  {object}  map[string]interface{}
// @Failure     500  {object}  map[string]interface{}
// @Router      /categories [get]
func (ctrl *CategoryController) GetAllCategories(c *gin.Context) {
	categories, err := ctrl.service.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// GetCategoryByID godoc
// @Summary     Get category by ID
// @Description Mengambil kategori berdasarkan ID
// @Tags        Categories
// @Produce     json
// @Param       id   path      int  true  "Category ID"
// @Success     200  {object}  map[string]interface{}
// @Failure     400  {object}  map[string]interface{}
// @Failure     404  {object}  map[string]interface{}
// @Router      /categories/{id} [get]
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
// @Summary     Create a new category
// @Description Membuat kategori transaksi baru
// @Tags        Categories
// @Accept      json
// @Produce     json
// @Param       body  body      object{name=string,type=string}  true  "Category payload"
// @Success     201   {object}  map[string]interface{}
// @Failure     400   {object}  map[string]interface{}
// @Router      /categories [post]
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
