package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// db      → shared GORM DB instance for the whole application.
// once    → guarantees the initialization logic runs exactly once (thread-safe lazy init).
var (
	db   *gorm.DB
	once sync.Once
)

// GetDB returns a singleton *gorm.DB.
// It initializes the connection lazily and exactly once.
// If initialization fails, this function panics with a descriptive error message.
// (Keeping panic is consistent with many GORM snippets and your original API.
// If you prefer error returns, we can switch to `GetDB() (*gorm.DB, error)` later.)
func GetDB() *gorm.DB {
	once.Do(func() {
		if err := initializeDB(); err != nil {
			// Fail-fast so callers don't receive a nil *gorm.DB.
			panic(fmt.Errorf("database initialization failed: %w", err))
		}
	})
	return db
}

// Close gracefully closes the underlying *sql.DB connection pool.
// Call this in your application's shutdown hook (e.g., on SIGTERM/SIGINT).
func Close() error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// initializeDB loads configuration, builds the DSN, opens the GORM connection,
// configures the connection pool, sets up logging, and verifies connectivity.
// It returns an error on any failure (which GetDB converts to a panic to keep your original API).
func initializeDB() error {
	// 1) Load environment variables (only in non-production environments).
	// In production, rely on the process environment (containers/secrets/etc).
	if os.Getenv("APP_ENV") != "production" {
		// godotenv.Load() is safe: it only populates missing keys, it won't overwrite existing env.
		_ = godotenv.Load()
	}

	// 2) Read essential DB configuration from env with sensible defaults.
	//    Using defaults is convenient for local development and CI.
	user := getenv("DB_USER", "root")
	pass := os.Getenv("DB_PASSWORD") // may be empty
	host := getenv("DB_HOST", "127.0.0.1")
	port := getenv("DB_PORT", "3306")
	name := getenv("DB_NAME", "gorm_lab")

	// Connection pool tuning (read from env; fallback to safe defaults).
	maxOpen := getenvInt("DB_MAX_OPEN_CONNS", 25)       // maximum open connections
	maxIdle := getenvInt("DB_MAX_IDLE_CONNS", 25)       // maximum idle connections
	lifeMin := getenvInt("DB_CONN_MAX_LIFETIME_MIN", 5) // minutes before a connection is recycled

	// MySQL driver timeouts (protect against hanging connects/reads/writes).
	// See https://github.com/go-sql-driver/mysql#parameters (for reference).
	timeout := getenv("DB_TIMEOUT", "5s")       // connection timeout
	readTO := getenv("DB_READ_TIMEOUT", "5s")   // read timeout
	writeTO := getenv("DB_WRITE_TIMEOUT", "5s") // write timeout

	// 3) Build DSN (Data Source Name).
	//    GORM expects a driver DSN. For MySQL:
	//    user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local...
	addr := net.JoinHostPort(host, port)
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s",
		user, pass, addr, name, timeout, readTO, writeTO,
	)

	// 4) Configure GORM logger.
	//    - In development, show Info to see executed SQL (very useful for learning).
	//    - In production, use Warn (to reduce noise, but still log slow queries).
	//    - The log levels in GORM are:
	//      - Silent → Prints nothing.
	//      - Error → Only errors.
	//      - Warn → Errors + warnings (like slow queries).
	//      - Info → Everything (including normal queries).
	var lvl logger.LogLevel = logger.Warn
	if os.Getenv("APP_ENV") != "production" {
		lvl = logger.Info
	}
	gormLogger := logger.New(
		log.New(os.Stdout, "gorm: ", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // log queries slower than this
			LogLevel:                  lvl,                    // Info in dev, Warn in prod
			IgnoreRecordNotFoundError: true,                   // don't spam on ErrRecordNotFound
			Colorful:                  os.Getenv("NO_COLOR") == "",
		},
	)

	// 5) Open the database connection via GORM.
	conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		// Add more GORM config here if needed (e.g., NowFunc, DisableForeignKeyConstraintWhenMigrating, etc.)
	})
	if err != nil {
		return fmt.Errorf("open mysql connection: %w", err)
	}

	// 6) Extract the underlying *sql.DB to tune the pool and do health checks.
	sqlDB, err := conn.DB()
	if err != nil {
		return fmt.Errorf("unwrap *sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(maxOpen)                                 // cap concurrent open connections
	sqlDB.SetMaxIdleConns(maxIdle)                                 // keep idle connections ready
	sqlDB.SetConnMaxLifetime(time.Duration(lifeMin) * time.Minute) // recycle connections periodically

	// 7) Proactive health check (fast-fail if DB is unreachable).
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := pingContext(ctx, sqlDB); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 8) Publish the initialized connection to the package-level singleton.
	db = conn
	return nil
}

// pingContext pings the DB with context when supported; falls back to Ping().
func pingContext(ctx context.Context, s *sql.DB) error {
	type pinger interface {
		PingContext(context.Context) error
	}
	if p, ok := interface{}(s).(pinger); ok {
		return p.PingContext(ctx)
	}
	return s.Ping()
}

// getenv returns the value of an env var or a default if unset/empty.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// getenvInt returns the integer value of an env var or a default on missing/parse error.
func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
