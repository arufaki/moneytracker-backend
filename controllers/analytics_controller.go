package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"money-tracker-ai/services"

	"github.com/gin-gonic/gin"
)

type AnalyticsController struct {
	analyticsService services.AnalyticsService
}

func NewAnalyticsController(analyticsService services.AnalyticsService) *AnalyticsController {
	return &AnalyticsController{analyticsService: analyticsService}
}

func (c *AnalyticsController) GetSummary(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uint)
	now := time.Now()
	monthStr := ctx.Query("month")
	yearStr := ctx.Query("year")

	month := int(now.Month())
	year := now.Year()

	if monthStr != "" {
		m, err := strconv.Atoi(monthStr)
		if err != nil || m < 1 || m > 12 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "month must be between 1 and 12"})
			return
		}
		month = m
	}

	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err == nil {
			year = y
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
			return
		}
	}

	summary, err := c.analyticsService.GetMonthlySummary(userID, month, year)
	if err != nil {
		log.Printf("[ERROR] GetSummary: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}
