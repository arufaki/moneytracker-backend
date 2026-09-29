package models

import "time"

type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);unique;not null" json:"name"`
	Type      string    `gorm:"type:varchar(20);not null" json:"type"` // 'income' / 'expense'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
