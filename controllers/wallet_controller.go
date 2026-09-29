package controllers

import (
	"money-tracker-ai/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WalletController struct {
	service services.WalletService
}

func NewWalletController(service services.WalletService) *WalletController {
	return &WalletController{service: service}
}

// GetAllWallets godoc
// GET /api/wallets
func (ctrl *WalletController) GetAllWallets(c *gin.Context) {
	wallets, err := ctrl.service.GetAllWallets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wallets})
}

// GetWalletByID godoc
// GET /api/wallets/:id
func (ctrl *WalletController) GetWalletByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	wallet, err := ctrl.service.GetWalletByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": wallet})
}

// CreateWallet godoc
// POST /api/wallets
// Body: { "name": "BCA", "balance": 1000000 }
func (ctrl *WalletController) CreateWallet(c *gin.Context) {
	var body struct {
		Name    string  `json:"name" binding:"required"`
		Balance float64 `json:"balance"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet, err := ctrl.service.CreateWallet(body.Name, body.Balance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": wallet, "message": "Wallet created successfully"})
}
