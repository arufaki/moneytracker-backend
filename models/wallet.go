package models

import "time"

type Wallet struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);unique;not null" json:"name"`
	Balance   float64   `gorm:"type:numeric(15,2);default:0.0" json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
