package main

import "encoding/json"

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

// handlerCreatePerson handles the requests for creating a person.
//   - Input `data` is whatever came from Request.Data
//   - It returns a Response (success/failure + data)
func (s *server) handlerCreatePerson(data interface{}) Response {
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

func (s *server) handlerCreateClass(data interface{}) Response {
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
			Message: "name is requires",
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
			Message: "teache id is required",
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
	if err != nil {
		return Response{
			Status:  false,
			Message: "teacher not found",
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
	if err != nil {
		return Response{
			Status:  false,
			Message: "teacher not found",
		}
	}
	// 5. Insert class
	res, err := s.db.Exec(
		`INSERT INTO classes(name, school_id, teacher_id, VALUES(?,?,?))`,
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
	// c.Students left empty/omitted // TODO: why?
	return Response{
		Status:  true,
		Message: "class created",
		Data:    c,
	}
}
