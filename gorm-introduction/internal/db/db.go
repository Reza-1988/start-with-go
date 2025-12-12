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
	//
	// At this step we are creating the high-level GORM connection object (*gorm.DB).
	// This is the ORM layer. It does NOT represent the actual raw database connections.
	// Instead, it is a feature-rich wrapper around the lower-level *sql.DB.
	//
	// - It knows how to map Go structs to tables
	// - It can build SQL queries for us
	// - It handles model-based operations (Create, Find, Update, etc.)
	// - It provides hooks, logging, migrations, and many ORM features
	//
	// IMPORTANT:
	// gorm.Open() does NOT immediately open a network connection.
	// It prepares the ORM and only initializes the underlying *sql.DB internally.
	conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger, // GORM uses this logger for SQL query logs and slow-query warnings
	})
	if err != nil {
		return fmt.Errorf("open mysql connection: %w", err)
	}

	// 6) Extract the underlying *sql.DB to tune the pool and do health checks.
	//
	// Even though GORM gives us *gorm.DB (conn), the actual database connection pool
	// lives inside the standard library type *sql.DB.
	//
	// We MUST extract it when we want to:
	//
	// - Configure connection pooling limits (max open / max idle connections)
	// - Configure connection lifetime limits
	// - Run health checks (Ping, PingContext)
	// - Close the database (sqlDB.Close()) on application shutdown
	//
	// Think of *gorm.DB as the "smart frontend" and *sql.DB as the "real engine".
	sqlDB, err := conn.DB()
	if err != nil {
		return fmt.Errorf("unwrap *sql.DB: %w", err)
	}

	// Configure connection pool behavior.
	// These settings control how many connections to the database server
	// can be opened simultaneously, how many idle (ready) connections are kept,
	// and how long a single connection is allowed to live before being recycled.
	sqlDB.SetMaxOpenConns(maxOpen)                                 // Maximum number of total open connections (in use + idle)
	sqlDB.SetMaxIdleConns(maxIdle)                                 // Maximum number of idle connections kept ready in the pool
	sqlDB.SetConnMaxLifetime(time.Duration(lifeMin) * time.Minute) // Recycle connections periodically to avoid stale connections

	// 7) Proactive health check (fast-fail if DB is unreachable)
	//
	// Before we consider the connection ready, we perform an explicit Ping using a context with timeout.
	// This ensures that:
	//
	// - If DB credentials are wrong → fail immediately
	// - If DB server is down → fail immediately
	// - If network is unreachable → fail immediately
	// - If connection handshake hangs → timeout in 3 seconds
	//
	// This prevents your application from starting successfully while the DB is actually unreachable.
	// It is MUCH better to fail fast on startup than to crash later when the first query runs.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := pingContext(ctx, sqlDB); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 8) Publish the initialized connection to the package-level singleton.
	//
	// At this point, the GORM connection (conn = *gorm.DB) is fully tested and safe.
	// We save it into the package-level variable `db` so that GetDB() can return it.
	//
	// NOTE:
	// The application will primarily use `conn` for all ORM operations.
	// The underlying *sql.DB (sqlDB) is mostly needed only here during initialization
	// or when closing the database at shutdown (sqlDB.Close()).
	db = conn
	return nil
}

// pingContext attempts to ping the database using PingContext (with timeout support).
// If the database driver does not support PingContext, it falls back to the regular Ping().
//
// In simple terms:
// - Try the modern method (PingContext): supports cancellation and timeouts.
// - If not available, use the classic Ping(): no timeout, but works for all drivers.
//
// Why this matters:
// Some database drivers implement PingContext(), some only implement Ping().
// This function safely handles both without breaking.
func pingContext(ctx context.Context, s *sql.DB) error {

	// Define a small interface that represents anything that has PingContext().
	// If *sql.DB or the underlying driver supports it, we will detect it.
	type pinger interface {
		PingContext(context.Context) error
	}

	// Check if 's' (the *sql.DB) supports PingContext().
	// If yes → use PingContext with timeout/cancellation support.
	if p, ok := interface{}(s).(pinger); ok {
		return p.PingContext(ctx)
	}

	// Otherwise → fallback to the basic Ping() (no context, no timeout).
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
