package db

import (
	"github.com/Reza-1988/start-with-go/gormintroduction/internal/models"
	"gorm.io/gorm"
)

// AutoMigrate runs GORM's automigrate for all core models in Stage 1.
// NOTE:
//   - AutoMigrate will create tables, missing foreign keys, constraints and indexes.
//   - AutoMigrate WILL NOT drop columns or delete indexes (safe for dev).
//   - You should re-run this every time the app starts during development.
//
// As the project grows, simply add new models to this list.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Profile{},
		&models.CreditCard{},
	)
}
