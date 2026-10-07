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

func (ctrl *WalletController) GetAllWallets(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	wallets, err := ctrl.service.GetAllWallets(userID)
	if err != nil {
		log.Printf("[ERROR] GetAllWallets: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wallets})
}

func (ctrl *WalletController) GetWalletByID(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	wallet, err := ctrl.service.GetWalletByID(uint(id), userID)
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

func (ctrl *WalletController) CreateWallet(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var body struct {
		Name    string  `json:"name" binding:"required,max=100"`
		Balance float64 `json:"balance" binding:"gte=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[SECURITY] Validation failure on POST /api/wallets from IP %s", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body or missing required fields"})
		return
	}

	wallet, err := ctrl.service.CreateWallet(userID, body.Name, body.Balance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": wallet, "message": "Wallet created successfully"})
}
