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

// Chat godoc
// @Summary     AI Chat untuk input transaksi
// @Description Mengirim pesan natural language untuk merekam transaksi via AI
// @Tags        AI Chat
// @Accept      json
// @Produce     json
// @Param       body  body      models.ChatRequest  true  "Chat payload"
// @Success     200   {object}  models.ChatResponse
// @Failure     400   {object}  map[string]interface{}
// @Failure     500   {object}  map[string]interface{}
// @Router      /chat [post]
func (ctrl *ChatController) Chat(c *gin.Context) {
	var req models.ChatRequest

	// Bind JSON body ke struct ChatRequest
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
		log.Printf("[ERROR] ProcessChatMessage failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

