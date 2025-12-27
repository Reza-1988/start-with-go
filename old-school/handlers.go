package main

import (
	"database/sql"
	"encoding/json"
	"strings"
)

// handleCreateSchool handles the requests for creating a school.
//   - Input `data` is whatever came from Request.Data
//   - It returns a Response (success/failure + data)
func (s *server) handleCreateSchool(data interface{}) Response {

	// Because Request.Data is interface{}, when JSON is decoded, Go usually stores objects as `map[string]any`
	// Here we try to case `data` to a map.
	//	- `m` becomes the JSON object like: `{"name": "school_1"}
	//	- `ok` is `true` if the cast succeeded.
	m, ok := data.(map[string]any)
	// If data is not the expected map format, return an error response.
	if !ok {
		return Response{
			Status:  false,
			Message: "bad data format",
		}
	}
	// Convert the map (`m`) back into JSON bytes.
	// Example result: `raw = []byte({"name": "school_1"})`
	raw, err := json.Marshal(m)
	// If marshaling fails, return error.
	if err != nil {
		return Response{
			Status:  false,
			Message: "name bad school playload",
		}
	}
	// Create an empty School struct variable.
	// Convert the JSON bytes into the struct School
	// Now `sch.Name` should be `"school_1"`
	// This map → JSON → struct, trick is common because it’s simple and works well.
	var sch School
	if err := json.Unmarshal(raw, &sch); err != nil {
		return Response{
			Status:  false,
			Message: "bad school payload",
		}
	}
	// Validate input: school name must not be empty. if empty return failure.
	if sch.Name == "" {
		return Response{
			Status:  false,
			Message: "Name is required",
		}
	}
	// Run SQL insert query to store school in SQLite
	// 	- `?` is placeholder, so it safety insets `sch.Name`(avoid SQL injection and formatting issues)
	//	- `res` is the SQL result object(it contains info like last inserted id)
	res, err := s.db.Exec(`INSERT INTO schools(name) VALUES(?)`, sch.Name)
	// If insert fails (db issue), return error.
	if err != nil {
		return Response{
			Status:  false,
			Message: "db insert failed",
		}
	}
	// After inserting a row, SQLite generates an auto-increment ID.
	// LastInsertId() gets that new ID (example: 1).
	newID, err := res.LastInsertId()
	// If it can't get the ID, return error.
	if err != nil {
		return Response{
			Status:  false,
			Message: "db is failed",
		}
	}
	// Put the new database ID inside the struct.
	// Convert to uint because `School.Id` is `uint`.
	sch.Id = uint(newID)
	// Return a success response.
	// Data contains the created school including its new Id.
	return Response{
		Status:  true,
		Message: "school created",
		Data:    sch,
	}
}

// More Explanation:
//	- Why not decode directly into School?
// 		- Because `Decode(&req)` happens before you check `req.Method`.
// 		- At the time of decoding, Go only sees `interface{}`. It cannot guess:
// 			- “Oh, if method is `/school/create` then Data should be School”
// 		- So you must decode in two steps:
// 			- Decode the envelope (`Request`) to know the method
// 			- Then decode the `Data` payload into the correct struct
// - Why do map → marshal → unmarshal?
// 	- Because you want to turn this: `map[string]any{"name": "school_1"}`
// 		- into this: `School{Name: "school_1"}`
//		- Go cannot “cast” a map into a struct directly.
// 		- So the easy and reliable trick is:
// 			1. marshal the map back into JSON bytes
// 				- now it becomes valid JSON again
//			2. unmarshal those bytes into the struct
// 				- now Go can fill struct fields using JSON tags
// 			- It’s basically:
// 				- Convert generic data back to JSON, then decode it properly
// - A tiny example:
//	- Request JSON: `{"method":"/school/create","data":{"name":"school_1"}}`
// 	- after `Decode(&req):
//		- `req.Method == "school/createa"`
//		- `req.Data == map[string]any{"name":"school_1"}`
//	- Then:
//		- marshal map → `{"name":"school_1"}`
//		- unmarshal into School → `School{Name:"school_1"}`

// handleCreatePerson handles the requests for creating a person.
//   - Input `data` is whatever came from Request.Data
//   - It returns a Response (success/failure + data)
func (s *server) handleCreatePerson(data interface{}) Response {
	m, ok := data.(map[string]any)
	if !ok {
		return Response{
			Status:  false,
			Message: "bad data format",
		}
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return Response{
			Status:  false,
			Message: "bad data",
		}
	}
	var p Person
	if err := json.Unmarshal(raw, &p); err != nil {
		return Response{
			Status:  false,
			Message: "bad person payload",
		}
	}
	if p.Name == "" {
		return Response{
			Status:  false,
			Message: "name is required",
		}
	}
	res, err := s.db.Exec(`INSERT INTO people(name) VALUES(?)`, p.Name)
	if err != nil {
		return Response{
			Status:  false,
			Message: "db insert failed",
		}
	}
	newID, err := res.LastInsertId()
	if err != nil {
		return Response{
			Status:  false,
			Message: "db id failed",
		}
	}
	p.Id = uint(newID)
	return Response{
		Status:  true,
		Message: "person created",
		Data:    p,
	}
}

// handleCreateClass handles the requests for creating a class.
//   - Input `data` is whatever came from Request.Data
//   - It returns a Response (success/failure + data)
func (s *server) handleCreateClass(data interface{}) Response {
	// 1. Convert Request.Data (Interface{}) to Class
	m, ok := data.(map[string]any)
	if !ok {
		return Response{
			Status:  false,
			Message: "bad data format",
		}
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return Response{
			Status:  false,
			Message: "bad data",
		}
	}
	var c Class
	if err := json.Unmarshal(raw, &c); err != nil {
		return Response{
			Status:  false,
			Message: "bad class payload",
		}
	}
	// 2. Basic validation
	if c.Name == "" {
		return Response{
			Status:  false,
			Message: "name is required",
		}
	}
	if c.SchoolId == 0 {
		return Response{
			Status:  false,
			Message: "school_id is required",
		}
	}
	if c.Teacher.Id == 0 {
		return Response{
			Status:  false,
			Message: "teacher id is required",
		}
	}
	// 3. Check school exist
	// We create a variable to hold the school id if it exists.
	// We don’t actually need the value, we just want to know “does a row exist?”
	var schoolExists uint
	// `QueryRow(...)` runs the SQL query that should return one row.
	// The SQL means: “find a school row where id = c.SchoolId”.
	// `?` is a safe placeholder; `c.SchoolId` is passed separately.
	// `.Scan(&schoolExists)` takes the returned column (id) and puts it into schoolExists.
	err = s.db.QueryRow(`SELECT id FROM schools WHERE id = ?`, c.SchoolId).Scan(&schoolExists)
	if err == sql.ErrNoRows {
		return Response{
			Status:  false,
			Message: "school not found",
		}
	}
	if err != nil {
		return Response{
			Status:  false,
			Message: "school not found",
		}
	}
	// 4.Check teacher exists (and load name to return consistent Teacher object)
	// Create a Person struct to store the teacher data we read from DB.
	var teacher Person
	// Same idea as before, but now we select two columns: id and name.
	// We search in people table using teacher id.
	// Scan(&teacher.Id, &teacher.Name) fills the struct fields.
	// If the teacher exists, you now have a clean teacher object from DB.
	err = s.db.QueryRow(`SELECT id, name FROM people WHERE id = ?`, c.Teacher.Id).Scan(&teacher.Id, &teacher.Name)
	// If teacher id doesn’t exist → query returns no rows → Scan returns error.
	// Then we return failure because you can’t create a class with a teacher that doesn’t exist.
	if err == sql.ErrNoRows {
		return Response{
			Status:  false,
			Message: "teacher not found",
		}
	}
	if err != nil {
		return Response{
			Status:  false,
			Message: "teacher not found",
		}
	}
	// 5. Rule: a person cannot be both student and teacher.
	//	- If this person is enrolled as a student anywhere, they cannot be assigned as a teacher.
	var tmp int
	err = s.db.QueryRow(`SELECT 1 FROM class_students WHERE student_id = ? LIMIT 1`,
		teacher.Id).Scan(&tmp)

	if err == nil {
		return Response{
			Status:  false,
			Message: "student cannot be teacher",
		}
	}
	if err != nil && err != sql.ErrNoRows {
		return Response{
			Status:  false,
			Message: "db error",
		}
	}
	// 5. Insert class
	res, err := s.db.Exec(
		`INSERT INTO classes(name, school_id, teacher_id) VALUES(?,?,?)`,
		c.Name, c.SchoolId, teacher.Id,
	)
	if err != nil {
		return Response{
			Status:  false,
			Message: "db insert failed",
		}
	}
	newID, err := res.LastInsertId()
	if err != nil {
		return Response{Status: false, Message: "db id failed"}
	}
	// 6. Build response object exactly like tests expect
	c.Id = uint(newID)
	c.Teacher = teacher
	// c.Students left empty/omitted because at class creation time there are no students yet,
	// and also because the tests don’t expect students in the create-class response.
	return Response{
		Status:  true,
		Message: "class created",
		Data:    c,
	}
}

func (s *server) handleAddStudentToClass(data interface{}) Response {
	// 1. Parse payload
	m, ok := data.(map[string]any)
	if !ok {
		return Response{
			Status:  false,
			Message: "bad data format",
		}
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return Response{
			Status:  false,
			Message: "bad data",
		}
	}
	var req AddStudentToClassReq
	if err := json.Unmarshal(raw, &req); err != nil {
		return Response{
			Status:  false,
			Message: "bad payload",
		}
	}
	if req.StudentId == 0 && req.ClassId == 0 {
		return Response{
			Status:  false,
			Message: "student_id and class_id are required",
		}
	}
	// Use a transaction to checks + insert are consistent
	//	- What is a transaction? A transaction is like doing several database steps as one safe unit.
	// 		- It means:
	// 			- either all steps succeed → we `Commit()`
	// 			- or any step fails → everything is cancelled → we `Rollback()`
	tx, err := s.db.Begin() // Start a transaction, tx is like a temporary DB “session” where all queries are grouped.
	if err != nil {         // If DB can’t start a transaction, we fail.
		return Response{
			Status:  false,
			Message: "db error",
		}
	}
	// This is a safety net.
	// 	- If the function returns early (because of any error), rollback will run automatically.
	// 	- If later we successfully call `tx.Commit()`, the rollback does nothing (safe).
	defer tx.Rollback()
	// Why do we use a transaction here?
	// 	- In AddStudentToClass we do multiple steps:
	// 		1. check student exists
	// 		2. check class exists
	// 		3. check rules (teacher/student, same school)
	// 		4. insert into class_students
	// 	- If we do these steps without a transaction, this can happen:
	// 		- You check rule
	// 		- Before you insert, another request changes the data (race condition)
	// 		- Now your insert could break the rules or create inconsistent state
	// 		- A transaction makes sure:
	// 			- The checks and the insert happen together on a consistent view of the database

	// 2. Load student (must exist)
	var student Person
	err = tx.QueryRow(`SELECT id, name FROM people WHERE id = ?`, req.StudentId).
		Scan(&student.Id, &student.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return Response{
				Status:  false,
				Message: "student not found",
			}
		}
		return Response{
			Status:  false,
			Message: "db error",
		}
	}
	// 3. Load class + school (must exist)
	var classSchoolID uint
	err = tx.QueryRow(`SELECT school_id FROM classes WHERE id = ?`, req.ClassId).
		Scan(&classSchoolID)
	if err != nil {
		if err == sql.ErrNoRows {
			return Response{
				Status:  false,
				Message: "class not found",
			}
		}
		return Response{
			Status:  false,
			Message: "db error",
		}
	}
	// 4. Rule: a person cannot be both teacher and student
	// 	- If this person teaches any class, reject enrollment as student.
	var tmp int // We create a temporary variable to store the result. We don’t care about the actual value, we just want to know if a row exists.
	err = tx.QueryRow(`SELECT 1 FROM classes WHERE teacher_id = ? LIMIT 1`, req.StudentId).
		Scan(&tmp)
	// - `SELECT 1` means:
	//		- If a row exists, return the number 1
	//		- We’re not selecting class data; we only want yes/no.
	// - `FROM classes`
	// 		- look inside the `classes` table.
	// - `WHERE teacher_id = ?`
	// 		- find classes where the teacher id equals this person’s id.
	// 		- `?` is a placeholder, and we pass `req.StudentId` safely.
	// - LIMIT 1
	// 		- stop after finding the first match (faster).
	// - What QueryRow(...).Scan(&tmp) does:
	// 	- If the query finds a row:
	//		- it returns one row containing 1
	// 		- Scan(&tmp) succeeds
	// 		- so err == nil
	// 	- If the query finds no rows:
	// 		- Scan returns sql.ErrNoRows
	// - If there is a real database problem:
	// 	- you get some other error
	if err == nil {
		return Response{
			Status:  false,
			Message: "teacher cannot be student",
		}
	}
	if err != nil && err != sql.ErrNoRows {
		return Response{
			Status:  false,
			Message: "db error",
		}
	}
	// 5. Rule: student can only enroll in classes of ONE school
	// 	- If student already enrolled in another school's class, reject.
	err = tx.QueryRow(`
		SELECT 1
		FROM class_sutdents cs
		JOIN classes c ON c.id = cs.class_id
		where cs.student_id = ? AND c.school_id <> ?
		LIMIT 1
    `, req.StudentId, classSchoolID).Scan(&tmp)
	// What data do we have here?
	// 	- `req.StudentId` → which student wants to enroll
	// 	- `classSchoolID`→ the school of the class they are trying to join (we already read it before)
	// 	- `class_students` table → stores enrollments (class_id, student_id)
	// 	- `classes table` → tells us each class belongs to which school (school_id)
	// The SQL query meaning:
	//	- `SELECT 1`
	//		- We only want to know: “does such a row exist?
	//		- Return 1 if found.
	// 	- `FROM class_students cs`
	// 		- Look at enrollments.
	// 		- cs is just a short name (alias).
	//	- `JOIN classes c ON c.id = cs.class_id`
	// 		- For each enrollment row, connect it to its class row.
	// 		- This gives us access to `c.school_id`.
	// 	- `WHERE cs.student_id = ?`
	// 		- Only check enrollments for this student.
	//	- `AND c.school_id <> ?`
	// 		- <> means “not equal”.
	// 		- So: find enrollments where the class’s school is different from the school we are trying to join now.
	// 	- `LIMIT 1`
	// 		- Stop early if we find one such case.
	//
	// Tiny example:
	// 	- Student 2 is enrolled in:
	// 		- class 10 (school 1)
	// 	- Now they try to join:
	// 		- class 20 (school 2)
	// - Query checks: does student 2 have an enrollment with school_id != 2 ?
	// 		- Yes (school 1 != 2) → reject
	// - If they try class 11 (school 1):
	// 		- school 1 != 1 is false → no row → allowed

	if err == nil {
		return Response{
			Status:  false,
			Message: "student can only enroll in one school",
		}
	}
	if err != nil && err != sql.ErrNoRows {
		return Response{
			Status:  false,
			Message: "db error",
		}
	}
	// 6. Inser enrollment
	_, err = tx.Exec(`INSERT INTO class_sutudents(class_id, student_id) VALUES(?, ?) `,
		req.ClassId, req.StudentId)
	if err != nil {
		// Duplicate enrollment (PRIMARY KEY constraint)
		//	- Because class_students table has this rule: `PRIMARY KEY (class_id, student_id)`
		//	- That means: the pair (class_id, student_id) must be unique, you cannot insert the same pair again
		// Explanation:
		//	- SQLite returns an error message text when the constraint is violated.
		// 	- The exact error string can vary by driver/version, but usually includes phrases like:
		// 		- "UNIQUE constraint failed"
		// 		- "PRIMARY KEY constraint failed" (or mentions PRIMARY KEY)
		//	- So this code checks the error message text:
		// 	- If it looks like a duplicate constraint error, we return a friendly message: "already enrolled"
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "PRIMARY KEY") {
			return Response{
				Status:  false,
				Message: "already enrolled",
			}
		}
		return Response{
			Status:  false,
			Message: "db insert failed",
		}
		// Database handle this rules already we define in database, just We are interpreting the error to give the correct message.
		// When `err != nil`, we want to know why it failed:
		// 	- If it failed because the student is already enrolled → return "already enrolled"
		// 	- If it failed for another reason (DB locked, foreign key error, syntax error) → return "db insert failed"
	}
	if err := tx.Commit(); err != nil {
		return Response{
			Status:  false,
			Message: "db commit failed",
		}
	}
	// 7. Return updated student with class ids.
	// 	- `student` is a Person struct representing that student:
	// 		- `student.Id`
	// 		- `student.Name`
	// 	- At that moment, `student.Classes` is still empty (zero value).
	//	- `Classes` is not stored inside the people table directly. It’s a computed field: we fill it by querying `class_students`.
	rows, err := s.db.Query(`SELECT class_id FROM class_students WHERE student_id = ? ORDER BY class_id`,
		req.StudentId) // This query: returns all class IDs the student is enrolled in.
	if err != nil {
		return Response{
			Status:  false,
			Message: "db error",
		}
		defer rows.Close()
	}
	var classIDs []uint // We collect all class IDs inside this slice
	for rows.Next() {
		var cid uint
		if err := rows.Scan(&cid); err != nil {
			return Response{
				Status:  false,
				Message: "db error",
			}
		}
		classIDs = append(classIDs, cid)
	}
	student.Classes = classIDs
	return Response{
		Status:  true,
		Message: "student added to class",
		Data:    student,
	}
}

func (s *server) handleWhoAmI(data interface{}) Response {
	// 1. Parse payload into Person (we only need Id).
	m, ok := data.(map[string]any)
	if !ok {
		return Response{
			Status:  false,
			Message: "bad data format",
		}
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return Response{
			Status:  false,
			Message: "bad data",
		}
	}
	var req Person
	if err := json.Unmarshal(raw, &req); err != nil {
		return Response{
			Status:  false,
			Message: "bad payload",
		}
	}
	if req.Id == 0 {
		return Response{
			Status:  false,
			Message: "id is required",
		}
	}

	// 2. Load the person from DB (must exists)
	var p Person
	err = s.db.QueryRow(`SELECT id, name FROM people WHERE id = ?`, req.Id).
		Scan(&p.Id, &p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return Response{
				Status:  false,
				Message: "person not found",
			}
		}
		return Response{
			Status:  false,
			Message: "db error",
		}
	}

	// 3. Check if this person is a teacher (classes where teacher_id = p.id)
	teacherClasses, err := s.getTeacherClassIDs(p.Id)
	if err != nil {
		return Response{
			Status:  false,
			Message: "db error",
		}
	}
	if len(teacherClasses) > 0 {
		p.Classes = teacherClasses
		return Response{
			Status:  true,
			Message: "ok",
			Data:    p,
		}
	}
	// 4. Otherwise treat as student: classes enrolled class_students
	studentClasses, err := s.getStudentClassIDs(p.Id)
	if err != nil {
		return Response{
			Status:  false,
			Message: "db error",
		}
	}
	p.Classes = studentClasses
	return Response{
		Status:  true,
		Message: "ok",
		Data:    p,
	}
}

// getTeacherClassIDs returns class IDs taught by a teacher.
func (s *server) getTeacherClassIDs(teacherID uint) ([]uint, error) {
	rows, err := s.db.Query(`SELECT id FROM classes WHERE teacher_id = ? ORDER BY id`, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uint
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// getStudentClassIDs returns class IDs a student is enrolled in.
func (s *server) getStudentClassIDs(studentID uint) ([]uint, error) {
	rows, err := s.db.Query(`SELECT class_id FROM class_students WHERE student_id = ? ORDER BY  class_id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uint
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
