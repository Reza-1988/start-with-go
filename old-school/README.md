# Old School – TCP Socket JSON Server (SQLite)

A minimal school management system implemented as a **raw TCP socket server** (no HTTP)
where clients communicate with the server using **JSON messages**.
This project supports creating **Schools**, **People**, and **Classes**, enrolling **Students** into classes, 
and a `WhoAmI` query that returns a person’s class IDs based on their role in the system.

---
## Key Concepts

- **Transport**: TCP socket (`net.Listener`, `net.Conn`)
- **Protocol**: JSON request/response objects (not HTTP)
- **Persistence**: SQLite (in-memory for test stability)
- **Roles (Teacher / Student)**:
  - A person becomes a **Teacher** if they are assigned as `teacher_id` in any class. 
  - A person becomes a **Student** if they appear in `class_students`. 
  - A person **cannot be both**.

---

## Feature & Business Rules

### School & Classes

- You can create new schools. 
- You can create classes only if:
  - The school exists. 
  - The teacher exists. 
  - The teacher is not already a student.

## People (Teacher/Student)

- People are created without an explicit “role”. 
- The role is inferred:
  - **Teacher**: appears in `classes.teacher_id`
  - **Student**: appears in `class_students.student_id`
- A person cannot be both teacher and student.

## Enrolling Students

- Students can enroll in multiple classes, **but only within a single school**. 
- A class can have many students. 
- Duplicate enrollment is prevented.

---

## Request/Response Protocol

### Request Format

Each client request is a single JSON object: 
```json
{
  "method": "/school/create",
  "data": { ... }
}
```
- `method`: identifies the action.
- `data`: payload for the method.

### Response Format

Server replies with: 
```json
{
  "status": true,
  "message": "ok",
  "data": { ... }
}
```
- `status`: success/failure 
- `message`: human-readable info 
- `data`: optional payload on success (or sometimes on failure)

### Message Framing
Clients send multiple requests over the same TCP connection.
Each `json.Encoder.Encode(...)` sends one JSON object; 
server `uses json.Decoder.Decode(...)` to read objects sequentially.

---

## Supported Methods

1. Create School
   - **Method**: `/school/create`
   - **Request Data**: `School` (usually `{ "name": "..." }`)
   - **Response Data**: created `School` with assigned `id`

2. Create Person
   - **Method**: `/person/create`
   - **Request Data**: `Person` (usually `{ "name": "..." }`)
   - **Response Data**: created `Person` with assigned `id`

3. Create Class
   - **Method**: `/class/create`
   - **Request Data**: `Class` (requires `school_id`, `name`, `teacher.id`)
   - **Validations**:
     - school exists 
     - teacher exists 
     - teacher is not a student
   - Response Data: created Class with assigned id

4. Add Student to Class
   - **Method**: `/class/add/student`
   - **Request Data**:
   ```json
   { "student_id": 2, "class_id": 1 }
   ```
   - **Validations**:
     - student exists
     - class exists
     - student is not a teacher
     - student is not already enrolled in a different school
     - prevents duplicate enrollment
     - Response Data: updated `Person` (student) with `classes: [classId, ...]`

5. Who Am I
   - **Method**: `/who/am/i `
   - **Request Data**: `Person` with `id` only (e.g. `{ "id": 1 }`)
   - **Response Data**: the person + their `classes`:
     - If they are a **teacher** → classes they teach
     - Else → classes they are enrolled in (student)

---

## Database Schema (SQLite)

- SQLite tables (minimal normalized schema):
  - `schools(id, name)`
  - `people(id, name)`
  - `classes(id, name, school_id, teacher_id)`
  - `class_students(class_id, student_id`) (join table, many-to-many)

### Why `class_students`?
- Because:
  - One class has many students 
  - One student can join many classes 
  - This is a classic many-to-many relationship, represented via a join table.

## In-Memory vs On-Disk SQLite

- This project uses in-memory SQLite by default:
  - In-memory: data is stored in RAM and disappears after process exits
  ✅ Great for tests: clean DB each run, IDs start indicating from 1
  - On-disk: data is stored in a .db file and persists across runs
  ✅ Good for real usage, but tests may fail unless tables are cleared

---

## Project Structure

Example structure (recommended):
```pgsql
.
├── entity.go        # entities + Request/Response structs (given)
├── main.go          # constants + Server interface + NewServer()
├── server.go        # Start/Stop, accept loop, connection handling, route()
├── storage.go       # initDBOnce + schema
├── handlers.go      # method handlers (create school/person/class, add student, who am i)
├── sample_test.go   # provided tests
├── go.mod
└── go.sum

```
---

## How to Run 

### Run Tests (Quera-style)

```bash
go test -v
```

### Run Server Manually (optional)

If you add a basic main() that starts the server:

```go
func main() {
    srv := NewServer()
    if err := srv.Start("8080"); err != nil {
        panic(err)
    }
}
```

Then run: 

```bash
go run .
```
---

## Implementation Notes / Best Practices Used
- **Single listener** for the server; one `Accept()` loop. 
- **Per-connection handler** reads multiple JSON requests until disconnect. 
- **SQLite foreign keys enabled** for better integrity. 
- **Transactions** used for `AddStudentToClass` to keep checks + insert consistent. 
- **Normalized schema** using a join table for enrollments.