package models

// Package models - base.go
//
// NOTE FOR FUTURE ME:
//
// This file is intentionally empty in the early stages of this project.
// In many real-world Go/GORM applications, developers define a "BaseModel"
// to hold fields or behaviors that should be shared across multiple models.
// Examples include:
//
//   - Custom ID types (UUID instead of uint)
//   - Custom timestamp fields (CreatedAt / UpdatedAt / DeletedAt)
//   - Audit fields (CreatedBy, UpdatedBy, etc.)
//   - Common Scopes (e.g., IsActive, WithTenantID)
//   - Soft delete helpers
//   - Hooks that apply to many models
//
// Right now, this project is focused on learning the fundamentals of GORM:
// associations, CRUD operations, AutoMigrate, and query patterns.
// Introducing a BaseModel too early would add unnecessary abstraction and
// could make the learning process more confusing.
//
// As the project grows and the number of models increases,
// I can return to this file and introduce shared behaviors
// **only when it becomes helpful** and not before.
//
// Summary:
//   Keep this file empty for now.
//   Add shared model features here **later**, when the project becomes larger.
//   This file exists as a reminder and a placeholder for future improvements.
//
// End of note.
