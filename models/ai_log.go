package models

import "time"

type AILog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	User          User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
	RawMessage    string    `gorm:"type:text;not null" json:"raw_message"`
	ExtractedJSON string    `gorm:"type:jsonb;not null" json:"extracted_json"`
	CreatedAt     time.Time `json:"created_at"`
}
