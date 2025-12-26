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
