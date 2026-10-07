package models

import "time"

type EmailVerification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
	Token     string    `gorm:"type:varchar(255);unique;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
