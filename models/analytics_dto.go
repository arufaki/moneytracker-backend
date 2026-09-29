package models

type CategoryBreakdown struct {
	CategoryName string  `json:"category_name"`
	Total        float64 `json:"total"`
	Percentage   float64 `json:"percentage"`
	BudgetLimit  float64 `json:"budget_limit"`
	IsOverBudget bool    `json:"is_over_budget"`
}

type MonthlySummary struct {
	Month      int                 `json:"month"`
	Year       int                 `json:"year"`
	Income     float64             `json:"income"`
	Expense    float64             `json:"expense"`
	NetBalance float64             `json:"net_balance"`
	Breakdown  []CategoryBreakdown `json:"breakdown"`
}
