package services

import (
	"fmt"
	"time"

	"money-tracker-ai/models"

	"gorm.io/gorm"
)

type AnalyticsService interface {
	GetMonthlySummary(month int, year int) (*models.MonthlySummary, error)
}

type analyticsService struct {
	db *gorm.DB
}

func NewAnalyticsService(db *gorm.DB) AnalyticsService {
	return &analyticsService{db: db}
}

func (s *analyticsService) GetMonthlySummary(month int, year int) (*models.MonthlySummary, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	nextMonthStart := startDate.AddDate(0, 1, 0)

	var income, expense float64
	var netBalance float64

	// Total Income
	if err := s.db.Model(&models.Transaction{}).
		Where("type = ? AND created_at >= ? AND created_at < ?", "income", startDate, nextMonthStart).
		Select("COALESCE(SUM(amount), 0)").Scan(&income).Error; err != nil {
		return nil, fmt.Errorf("failed to query income: %w", err)
	}

	// Total Expense
	if err := s.db.Model(&models.Transaction{}).
		Where("type = ? AND created_at >= ? AND created_at < ?", "expense", startDate, nextMonthStart).
		Select("COALESCE(SUM(amount), 0)").Scan(&expense).Error; err != nil {
		return nil, fmt.Errorf("failed to query expense: %w", err)
	}

	// Net Balance (from all wallets)
	if err := s.db.Model(&models.Wallet{}).
		Select("COALESCE(SUM(balance), 0)").Scan(&netBalance).Error; err != nil {
		return nil, fmt.Errorf("failed to query net balance: %w", err)
	}

	// Category Breakdown
	type Result struct {
		CategoryName string
		Total        float64
		BudgetLimit  float64
	}
	var results []Result

	if err := s.db.Table("transactions").
		Select("categories.name as category_name, COALESCE(SUM(transactions.amount), 0) as total, categories.budget_limit as budget_limit").
		Joins("left join categories on categories.id = transactions.category_id").
		Where("transactions.type = ? AND transactions.created_at >= ? AND transactions.created_at < ?", "expense", startDate, nextMonthStart).
		Group("categories.id, categories.name, categories.budget_limit").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to query category breakdown: %w", err)
	}

	breakdown := make([]models.CategoryBreakdown, 0)
	for _, r := range results {
		percentage := 0.0
		if expense > 0 {
			percentage = (r.Total / expense) * 100
		}
		isOverBudget := false
		if r.BudgetLimit > 0 && r.Total > r.BudgetLimit {
			isOverBudget = true
		}
		breakdown = append(breakdown, models.CategoryBreakdown{
			CategoryName: r.CategoryName,
			Total:        r.Total,
			Percentage:   percentage,
			BudgetLimit:  r.BudgetLimit,
			IsOverBudget: isOverBudget,
		})
	}

	return &models.MonthlySummary{
		Month:      month,
		Year:       year,
		Income:     income,
		Expense:    expense,
		NetBalance: netBalance,
		Breakdown:  breakdown,
	}, nil
}
