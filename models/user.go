package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"type:varchar(191);unique;not null" json:"email"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	PasswordHash *string   `gorm:"type:varchar(255)" json:"-"`
	GoogleID     *string   `gorm:"type:varchar(191);unique" json:"google_id,omitempty"`
	IsVerified   bool      `gorm:"default:false" json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
