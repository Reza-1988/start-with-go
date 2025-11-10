package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// db     → shared GORM DB instance for entire application
// once   → ensures database initialization happens ONLY ONCE (even with multi goroutines)
var (
	db   *gorm.DB
	once sync.Once
)

// GetDB returns the DB instance. initializeDB() only runs once.
// Using sync.Once makes the connection setup thread-safe and lazy-initialized.
func GetDB() *gorm.DB {
	once.Do(func() {
		initializeDB() // runs only once
	})
	return db
}

// initializeDB loads .env, builds DSN, opens DB connection, configures pooling.
func initializeDB() {

	// Load .env file into environment variables (LOCAL development use case)
	// Not needed in production — production uses system environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found (environment variables may come from OS)")
	}

	// Retrieve database credentials stored in environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// DSN (Data Source Name) → "connection string" that tells GORM how to connect to MySQL
	// Format: user:password@tcp(host:port)/dbname?params...
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Open connection to the database through GORM
	// mysql.Open(dsn) selects MySQL driver
	// &gorm.Config{} passes GORM configuration options
	connection, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("❌ Failed to connect to the database")
	}

	// Extract underlying sql.DB object to configure connection pooling
	// (GORM is built on top of database/sql)
	if sqlDB, err := connection.DB(); err == nil {

		// Max number of active / open DB connections
		sqlDB.SetMaxOpenConns(25)

		// Number of idle connections kept ready for next use (improves performance)
		sqlDB.SetMaxIdleConns(25)

		// Maximum lifetime of a DB connection before being recycled
		sqlDB.SetConnMaxLifetime(5 * time.Minute)

		// Test DB connection immediately to ensure database is reachable
		if err := sqlDB.Ping(); err != nil {
			panic(fmt.Errorf("❌ DB ping failed: %w", err))
		}
	}

	// Store initialized DB connection into shared global reference
	db = connection
}
