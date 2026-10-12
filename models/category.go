package models

import "time"

type Category struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      *uint     `gorm:"index" json:"user_id,omitempty"`
	User        *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Type        string    `gorm:"type:varchar(20);not null" json:"type"` // 'income' / 'expense'
	BudgetLimit float64   `gorm:"type:numeric(15,2);default:0" json:"budget_limit"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
