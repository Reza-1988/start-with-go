package models

import "gorm.io/gorm"

// Profile is a "has one" detail of a User (1:1).
// Foreign key lives on Profile (user_id), so User has one Profile.
type Profile struct {
	gorm.Model
	UserID uint   `gorm:"uniqueIndex"` // ensure 1:1 (one profile per user)
	Bio    string `gorm:"size:500"`
	Phone  string `gorm:"size:50"`
}
