package services

import (
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
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	var income, expense float64
	var netBalance float64

	// Total Income
	s.db.Model(&models.Transaction{}).
		Where("type = ? AND date BETWEEN ? AND ?", "income", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&income)

	// Total Expense
	s.db.Model(&models.Transaction{}).
		Where("type = ? AND date BETWEEN ? AND ?", "expense", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&expense)

	// Net Balance (from all wallets)
	s.db.Model(&models.Wallet{}).Select("COALESCE(SUM(balance), 0)").Scan(&netBalance)

	// Category Breakdown
	type Result struct {
		CategoryName string
		Total        float64
		BudgetLimit  float64
	}
	var results []Result

	s.db.Table("transactions").
		Select("categories.name as category_name, COALESCE(SUM(transactions.amount), 0) as total, categories.budget_limit as budget_limit").
		Joins("left join categories on categories.id = transactions.category_id").
		Where("transactions.type = ? AND transactions.date BETWEEN ? AND ?", "expense", startDate, endDate).
		Group("categories.id, categories.name, categories.budget_limit").
		Scan(&results)

	var breakdown []models.CategoryBreakdown
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
