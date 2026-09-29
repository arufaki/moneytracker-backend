package controllers

import (
	"money-tracker-ai/models"
	"money-tracker-ai/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	service services.TransactionService
}

func NewChatController(service services.TransactionService) *ChatController {
	return &ChatController{service: service}
}

// Chat handles POST /api/chat
// User mengirim pesan teks, server memproses dan mencatat transaksi
func (ctrl *ChatController) Chat(c *gin.Context) {
	var req models.ChatRequest

	// Bind JSON body ke struct ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Field 'message' wajib diisi",
		})
		return
	}

	// Proses pesan via TransactionService
	resp, err := ctrl.service.ProcessChatMessage(req.Message)
	if err != nil {
		if err.Error() == "insufficient balance" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
