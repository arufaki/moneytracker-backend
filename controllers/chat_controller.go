package controllers

import (
	"log"
	"money-tracker-ai/models"
	"money-tracker-ai/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	service services.TransactionService
}

func NewChatController(service services.TransactionService) *ChatController {
	return &ChatController{service: service}
}

func (ctrl *ChatController) Chat(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var req models.ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if strings.Contains(err.Error(), "too large") {
			log.Printf("[SECURITY] Request body too large on /api/chat from IP: %s", c.ClientIP())
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"error":   "request body too large",
			})
			return
		}
		log.Printf("[SECURITY] Validation failure on /api/chat from IP %s", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Field 'message' wajib diisi",
		})
		return
	}

	resp, err := ctrl.service.ProcessChatMessage(userID, req.Message)
	if err != nil {
		if err.Error() == "insufficient balance" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		log.Printf("[ERROR] ProcessChatMessage failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
