package models

import "time"

type AILog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	RawMessage    string    `gorm:"type:text;not null" json:"raw_message"`
	ExtractedJSON string    `gorm:"type:jsonb;not null" json:"extracted_json"`
	CreatedAt     time.Time `json:"created_at"`
}
