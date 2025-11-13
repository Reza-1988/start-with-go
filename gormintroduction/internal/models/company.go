package models

import "gorm.io/gorm"

// Company represents an organization a user belongs to.
// - One-to-many with User (Company has many Users; User belongs to Company).
// - Soft delete is optional; enable if you want to practice Unscoped().
type Company struct {
	gorm.Model
	Name string `gorm:"size:200;uniqueIndex"` // unique company name (demo purpose)
	// Users []User // optional backref; not required to create the relation
}
