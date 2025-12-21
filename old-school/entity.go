package main

// School represents a school/university in the system.
//
// Meaning in the domain:
// - A School can have multiple Classes.
// - Each Class belongs to exactly one School (via Class.SchoolId).
//
// Meaning in the JSON protocol/tests:
// - When creating a school, the client sends: { "name": "..." } (Id omitted).
// - The server must assign a unique Id and return the created school in Response.Data.
// - Tests compare only Id and Name in responses; Classes is optional and usually omitted.
type School struct {
	Id      uint    `json:"id,omitempty"`
	Name    string  `json:"name,omitempty"`
	Classes []Class `json:"classes,omitempty"`
}

// Class represents a course/class offered by a specific school.
//
// Meaning in the domain:
// - Each Class belongs to one School (SchoolId).
// - Each Class has exactly one Teacher (Teacher).
// - A Class can have many Students (Students).
//
// Meaning in the JSON protocol/tests:
//   - When creating a class, the client sends SchoolId, Name, and Teacher (as a Person object).
//     Example payload: { "school_id": 1, "name": "...", "teacher": { "id": 1, "name": "..." } }
//   - The server must assign a unique class Id and return it.
//   - Tests compare Id, Name, SchoolId, and Teacher; Students is typically omitted.
type Class struct {
	Id       uint     `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	SchoolId uint     `json:"school_id,omitempty"`
	Teacher  Person   `json:"teacher,omitempty"`
	Students []Person `json:"students,omitempty"`
}

// Person represents a user of the system.
// In the domain, a Person can act as a Teacher OR a Student, but not both.
//
// Meaning in the domain:
// - If a person is used as a Teacher, their Classes list should contain class IDs they teach.
// - If a person is a Student, their Classes list should contain class IDs they are enrolled in.
// - Hidden tests typically enforce:
//   - A person cannot be both teacher and student.
//   - A student can only enroll in classes of one school (but can take multiple classes in that school).
//   - A teacher can teach multiple classes across multiple schools.
//
// Meaning in the JSON protocol/tests:
// - When creating a person, the client sends { "name": "..." } and server assigns Id.
// - For WhoAmI and AddStudentToClass responses, server returns the person with the Classes field filled.
type Person struct {
	Id      uint   `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Classes []uint `json:"classes,omitempty"`
}

// Request is the top-level message sent from client (tests) to the TCP server.
//
// Protocol notes:
//   - Communication is JSON over raw TCP (not HTTP).
//   - Client uses json.Encoder.Encode(...) which writes one JSON object per request.
//   - Method selects which operation to perform (e.g. "/school/create").
//   - Data holds the payload for the operation, but because Data is interface{},
//     when decoding it will usually come in as map[string]any.
//     The server typically re-marshals Data to JSON and unmarshals into the correct struct
//     depending on Method.
type Request struct {
	Method string      `json:"method,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

// Response is the top-level message sent from server to client (tests).
//
// Test expectations:
// - Status must be true on success (tests assert this).
// - Data must be a JSON object matching the expected entity (School/Person/Class).
// - Tests decode Response.Data into map[string]any, then marshal+unmarshal into the entity struct.
// - Message is not heavily asserted in the visible test, but is useful for debugging.
type Response struct {
	Status  bool        `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// AddStudentToClassReq is the payload for enrolling a student in a class.
//
// Protocol notes:
//   - Client sends this as Request.Data for method "/class/add/student".
//   - On success, server responds with the updated Student Person as Response.Data,
//     including the updated list of class IDs in Person.Classes.
type AddStudentToClassReq struct {
	StudentId uint `json:"student_id,omitempty"`
	ClassId   uint `json:"class_id,omitempty"`
}

// Note: Why Id is `unit`?
// - IDs are never negative, so `unit`(unsigned integer) expresses that intent.
// - In many Go + ORM codebases (especially with GORM), `unit` is commonly used for primary keys (It maps cleanly to SQLite `INTEGER`)

// Note: What does `omitempty` mean?
//	- `omitempty` is JSON tag option that tells Go's JSON encoder:
//		- If this field has the zero value, don't include in the output JSON.
// 		- `unit` zero value is 0
//		- `string` zero value is ""
//		- slices/maps zero value is nil (note: empty slice `[]` is not nil)
//	- For example: if `id == 0`, JSON output won't contain "id" at all.
// 	- Why it’s useful here:
//		- When creating objects, the client sends `{ "name": "..." }` without an id.
// 		- If you encode a struct with `Id = 0`, `omitempty` prevents sending "id":0, keeping payload clean.
