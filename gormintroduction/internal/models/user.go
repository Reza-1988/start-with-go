package models

import "gorm.io/gorm"

// User demonstrates multiple association types:
// - belongs to Company (CompanyID lives on User)
// - has one Profile (FK lives on Profile)
// - has many CreditCards (FK lives on CreditCard)
type User struct {
	// gorm.Model adds the following fields automatically:
	//   ID        uint           `gorm:"primaryKey"`
	//   CreatedAt time.Time
	//   UpdatedAt time.Time
	//   DeletedAt gorm.DeletedAt `gorm:"index"`  // enables soft delete
	gorm.Model
	Email       string       `gorm:"size:200;uniqueIndex"`
	Name        string       `gorm:"size:120;index"`
	CompanyID   *uint        `gorm:"index"` // nullable: user may have no company
	Company     *Company     // Belongs To: by default uses CompanyID as FK
	Profile     *Profile     // Has One: FK on profile.user_id
	CreditCards []CreditCard // Has Many: FK on credit_cards.user_id
}

// Example Hooks (optional, nice to learn):
// BeforeCreate/BeforeSave to normalize email, etc.
// func (u *User) BeforeCreate(tx *gorm.DB) (err error) { ... }
