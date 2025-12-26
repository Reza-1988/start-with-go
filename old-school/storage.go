package main

import "database/sql"

// initDBOnce opens SQLite and creates tables if needed.
// This function makes sure the SQLite database is ready to use.
// "Once" means: it should not run multiple times.
// We do NOT close db in Stop(), because tests call Stop() early still using the connection.
func (s *server) initDBOnce() error {
	// We create a variable to store any error that happens during initialization.
	// We use this because inside `Do(...)` we can't easily `return err` directly.
	var initErr error

	// dbOnce is `sync.Once`.
	// Do(func(){...}) guarantees the coe inside runs only one time, even if many goroutine call initDBOnce().
	s.dbOnce.Do(func() {
		// Open a SQLite database file named `school.db.
		// if the file doesn't exist, SQLite creates it automatically.
		//	- "_busy_timeout=5000" means: if database is busy/locked, wait up to 5 seconds.
		//	- "_foreign_keys=1" enables foreign key rules (useful later when you add class/person relations)
		db, err := sql.Open("sqlite", "file:school.db?_busy_timeout=5000&_foreign_keys=1")

		// If opening the DB fails, save the error into initErr and stop initialization.
		if err != nil {
			initErr = err
			return
		}
		// Small connection pool settings for local sqlite usage
		// 	- SQLite works best with one writer connection.
		// 	- This limits `database/sql` from opening many connections at the same time.
		//	- Keeps things simpler and avoids "database is locked" errors.
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		// Create minimal table fo this step
		// TODO: Implement other tables
		// Run a SQL command that creates the schools table if it doesn’t exist.
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS schools (
			    id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL
			);
        `)
		// If table creation fails:
		// 	- close the database handle (cleanup)
		// 	- store the error in initErr
		// 	- stop
		if err != nil {
			_ = db.Close()
			initErr = err
			return
		}
		// Save the opened database handle into the server struct (s.db)
		// So later handlers can do `s.db.Exec(...)` / `s.db.QueryRow(...)`.
		s.db = db

	})
	// If everything went OK, initErr is nil.
	// If anything failed, initErr contains the error.
	return initErr
}
