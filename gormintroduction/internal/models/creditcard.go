package models

import "gorm.io/gorm"

// CreditCard is a "has many" relation from User (1:N).
// Foreign key lives on CreditCard (user_id).
type CreditCard struct {
	gorm.Model
	UserID uint
	Number string `gorm:"size:64;index"` // do NOT store real cards in plaintext in real apps
	Label  string `gorm:"size:50"`       // e.g., "Personal", "Work"
}
