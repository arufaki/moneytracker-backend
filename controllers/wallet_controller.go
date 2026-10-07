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

type WalletController struct {
	service services.WalletService
}

func NewWalletController(service services.WalletService) *WalletController {
	return &WalletController{service: service}
}

// GetAllWallets godoc
// @Summary     Get all wallets
// @Description Mengambil semua data wallet yang tersedia
// @Tags        Wallets
// @Produce     json
// @Success     200  {object}  map[string]interface{}
// @Failure     500  {object}  map[string]interface{}
// @Router      /wallets [get]
func (ctrl *WalletController) GetAllWallets(c *gin.Context) {
	wallets, err := ctrl.service.GetAllWallets()
	if err != nil {
		log.Printf("[ERROR] GetAllWallets: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wallets})
}

// GetWalletByID godoc
// @Summary     Get wallet by ID
// @Description Mengambil data wallet berdasarkan ID
// @Tags        Wallets
// @Produce     json
// @Param       id   path      int  true  "Wallet ID"
// @Success     200  {object}  map[string]interface{}
// @Failure     400  {object}  map[string]interface{}
// @Failure     404  {object}  map[string]interface{}
// @Router      /wallets/{id} [get]
func (ctrl *WalletController) GetWalletByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	wallet, err := ctrl.service.GetWalletByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		} else {
			log.Printf("[ERROR] GetWalletByID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wallet})
}

// CreateWallet godoc
// @Summary     Create a new wallet
// @Description Membuat wallet baru dengan nama dan saldo awal
// @Tags        Wallets
// @Accept      json
// @Produce     json
// @Param       body  body      object{name=string,balance=number}  true  "Wallet payload"
// @Success     201   {object}  map[string]interface{}
// @Failure     400   {object}  map[string]interface{}
// @Router      /wallets [post]
func (ctrl *WalletController) CreateWallet(c *gin.Context) {
	var body struct {
		Name    string  `json:"name" binding:"required,max=100"`
		Balance float64 `json:"balance" binding:"gte=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[SECURITY] Validation failure on POST /api/wallets from IP %s", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body or missing required fields"})
		return
	}

	wallet, err := ctrl.service.CreateWallet(body.Name, body.Balance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": wallet, "message": "Wallet created successfully"})
}

