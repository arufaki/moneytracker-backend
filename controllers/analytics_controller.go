package controllers

import (
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
	now := time.Now()
	monthStr := ctx.Query("month")
	yearStr := ctx.Query("year")

	month := int(now.Month())
	year := now.Year()

	if monthStr != "" {
		m, err := strconv.Atoi(monthStr)
		if err == nil && m >= 1 && m <= 12 {
			month = m
		} else if err != nil || m < 1 || m > 12 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month"})
			return
		}
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

	summary, err := c.analyticsService.GetMonthlySummary(month, year)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}
