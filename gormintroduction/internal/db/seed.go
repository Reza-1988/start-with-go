package db

import (
	"fmt"

	"github.com/Reza-1988/start-with-go/gormintroduction/internal/models"
	"gorm.io/gorm"
)

// SeedInitialData inserts a small set of sample data into the database.
//
// WHY THIS FILE EXISTS (IMPORTANT):
// ---------------------------------
// In real projects and in learning environments, it is extremely helpful
// to have some initial "example" data in your database.
// This is called "seeding".
//
// Seeding helps you:
//
//  1. Quickly test queries without manually inserting rows.
//  2. Verify that associations (belongs to, has many, has one) are working.
//  3. Learn Preload(), Joins, CRUD flows with real data.
//  4. Avoid repeating tedious manual insert steps every time you restart.
//
// For now, this seed only inserts:
//   - Two companies
//   - Three users (some with company, some without)
//   - One profile (for Alice)
//   - Two credit cards (for Bob)
//
// This is enough to learn 90% of basic GORM behaviors.
//
// NOTES:
//   - This seed function is idempotent-ish: running it multiple times
//     will NOT create duplicates (because we check first).
//   - In production, seeding is handled differently. This is only for learning.
func SeedInitialData(db *gorm.DB) error {

	// Helper: create company only if it doesn't already exist.
	ensureCompany := func(name string) (*models.Company, error) {
		var company models.Company
		if err := db.Where("name = ?", name).First(&company).Error; err == nil {
			// Company already exists → return it
			return &company, nil
		}

		// Create a new company
		company = models.Company{Name: name}
		if err := db.Create(&company).Error; err != nil {
			return nil, err
		}
		return &company, nil
	}

	// Create two companies
	acme, err := ensureCompany("Acme Inc")
	if err != nil {
		return fmt.Errorf("seed: failed to create Acme: %w", err)
	}

	contoso, err := ensureCompany("Contoso")
	if err != nil {
		return fmt.Errorf("seed: failed to create Contoso: %w", err)
	}

	// USERS
	// -----
	// We will insert three sample users:
	//   1) Alice → belongs to Acme → has a Profile
	//   2) Bob   → belongs to Contoso → has two credit cards
	//   3) Charlie → no company, no relations
	//
	// This gives us both "belongs to" and "has one" and "has many" examples.

	users := []models.User{
		{Email: "alice@example.com", Name: "Alice", CompanyID: &acme.ID},
		{Email: "bob@example.com", Name: "Bob", CompanyID: &contoso.ID},
		{Email: "charlie@example.com", Name: "Charlie", CompanyID: nil},
	}

	// Insert each user if not already present.
	for _, u := range users {
		var existing models.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err == nil {
			// user exists → skip
			continue
		}
		if err := db.Create(&u).Error; err != nil {
			return fmt.Errorf("seed: failed to create user %s: %w", u.Email, err)
		}
	}

	// PROFILE for Alice
	// -----------------
	var alice models.User
	if err := db.Where("email = ?", "alice@example.com").First(&alice).Error; err == nil {
		var p models.Profile
		if err := db.Where("user_id = ?", alice.ID).First(&p).Error; err != nil {
			// Create only if not exists
			if err := db.Create(&models.Profile{
				UserID: alice.ID,
				Bio:    "Senior Developer",
				Phone:  "+1-555-111",
			}).Error; err != nil {
				return fmt.Errorf("seed: failed to create profile for Alice: %w", err)
			}
		}
	}

	// CREDIT CARDS for Bob
	// ----------------------
	var bob models.User
	if err := db.Where("email = ?", "bob@example.com").First(&bob).Error; err == nil {
		var count int64
		db.Model(&models.CreditCard{}).Where("user_id = ?", bob.ID).Count(&count)
		if count == 0 {
			// Insert two credit cards for Bob
			cards := []models.CreditCard{
				{UserID: bob.ID, Number: "CARD-1111", Label: "Personal"},
				{UserID: bob.ID, Number: "CARD-2222", Label: "Work"},
			}
			if err := db.Create(&cards).Error; err != nil {
				return fmt.Errorf("seed: failed to create Bob's credit cards: %w", err)
			}
		}
	}

	return nil
}
