package main

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

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
		// In-memory database (no file on disk):
		//   - "file:memdb1?mode=memory&cache=shared..."
		//
		// What it means:
		// - Data lives only in RAM and is deleted when the program ends.
		// - No .db file is created.
		// - Great for tests: every run starts fresh, so IDs start from 1.
		//
		// Why "cache=shared"? (in simple words)
		// - It helps all parts of this program reuse the same in-memory DB instance.
		//
		// Other flags:
		// - _busy_timeout=5000 : if DB is locked, wait up to 5 seconds
		// - _foreign_keys=1    : enforce foreign key rules (safer relations)
		db, err := sql.Open("sqlite", "file:memdb1?mode=memory&cache=shared&_busy_timeout=5000&_foreign_keys=1")
		//
		//On-disk database (hard / file-based):
		//   "file:school.db?..."
		// What it means:
		// - SQLite stores data in a real file (school.db) on disk.
		// - Data stays even after program stops (persistent).
		// - Not ideal for tests unless you clear the tables, because old data remains and IDs keep increasing.
		// db, err := sql.Open("sqlite", "file:school.db?_busy_timeout=5000&_foreign_keys=1")

		// If opening the DB fails, save the error into initErr and stop initialization.
		if err != nil {
			initErr = err
			return
		}
		// Small connection pool settings for local sqlite usage
		// 	- SQLite works best with a single open connection for this kind of project.
		// 	- This limits `database/sql` from opening many connections at the same time.
		//	- Keeps things simpler and avoids "database is locked" errors.
		//	- Keeping MaxOpenConns(1) also helps ensure the in-memory DB stays consistent.
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)

		// Create school table
		// TODO: Implement other tables
		// Run a SQL command that creates the schools table if it doesn’t exist.
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS schools (
			    id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL
			);
        `)
		// If school table creation fails:
		// 	- close the database handle (cleanup)
		// 	- store the error in initErr
		// 	- stop
		if err != nil {
			_ = db.Close()
			initErr = err
			return
		}
		// Create people table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS people (
			    id INTEGER PRIMARY KEY AUTOINCREMENT,
			    name TEXT NOT NULL                             
			);
        `)
		if err != nil {
			_ = db.Close()
			initErr = err
			return
		}
		// create classes table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS classes (
			    id INTEGER PRIMARY KEY AUTOINCREMENT,
			    name TEXT NOT NULL,
			    school_id INTEGER NOT NULL,
			    teacher_id INTEGER NOT NULL, 
			    FOREIGN KEY (school_id) REFERENCES schools(id),
			    FOREIGN KEY (teacher_id) REFERENCES people(id)
			);
        `)
		if err != nil {
			_ = db.Close()
			initErr = err
			return
		}
		// In our system:
		// 	- One Class can have many Students
		// 	- One Student can join many Classes (within the same school)
		// 	- That is a classic many-to-many relationship, and in SQL you represent it with a separate table (a “link table”).
		// Why not store students inside classes table directly?
		// 	- Because classes would need to store a list of students, and SQL tables don’t store lists well.
		// Create class_student (many-to-many join table)
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS class_students (
			    class_id INTEGER NOT NULL,
			    student_id INTEGER NOT NULL,
			    PRIMARY KEY (class_id, student_id),
			    FOREIGN KEY (class_id) REFERENCES classes(id),
			    FOREIGN KEY (student_id) REFERENCES people(id)
			);
        `)
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
