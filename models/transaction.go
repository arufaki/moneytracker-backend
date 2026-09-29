package models

import "time"

type Transaction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	WalletID    uint      `gorm:"not null" json:"wallet_id"`
	Wallet      Wallet    `gorm:"foreignKey:WalletID;constraint:OnDelete:CASCADE;" json:"wallet"`
	CategoryID  uint      `gorm:"not null" json:"category_id"`
	Category    Category  `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE;" json:"category"`
	Amount      float64   `gorm:"type:numeric(15,2);not null" json:"amount"`
	Type        string    `gorm:"type:varchar(20);not null" json:"type"` // 'income' / 'expense'
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
