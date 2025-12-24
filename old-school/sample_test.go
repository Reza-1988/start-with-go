package main

import (
	"encoding/json" // Encode/Decode JSON messages over the TCP connection
	"fmt"           // Print debug output in the test
	"log"           // Log fatal errors if the server can't start
	"net"           // Create TCP connections (net.Dial)
	"testing"       // Go testing framework
	"time"          // Sleep to give the server time to start

	"github.com/stretchr/testify/require" // Assertions: fail test immediately if condition is not met
)

const (
	PORT = "8080" // Port where the test expects the server to listen (TCP)

	// These are the exact method strings sent inside Request.Method.
	// Your server must compare against these to route the request correctly.
	createSchoolMethod      = "/school/create"
	createClassMethod       = "/class/create"
	createPersonMethod      = "/person/create"
	addStudentToClassMethod = "/class/add/student"
	whoAmIMethod            = "/who/am/i"
)

// startTestServer creates a new server instance, starts it in a goroutine,
// waits a short time so it begins listening, then returns the server so the test can stop it later.
func startTestServer(port string) Server {
	// Create the server instance using the project's factory function.
	server := NewServer()

	// Start the server in a separate goroutine so the test can continue running.
	go func() {
		// Start should block (listen + accept loop). If it fails, stop the whole test run.
		err := server.Start(port)
		if err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Small delay to make sure the server is already listening before we Dial.
	time.Sleep(200 * time.Millisecond)

	// Return the server so the caller can Stop() it later.
	return server
}

// createConnection starts the server and opens a TCP connection to it.
// It returns a JSON encoder/decoder bound to that connection so tests can send Request and read Response.
func createConnection(t *testing.T) (*json.Encoder, *json.Decoder) {
	// Start the TCP server before trying to connect to it.
	server := startTestServer(PORT)

	// Stop the server when this function returns.
	// IMPORTANT: this runs immediately after returning encoder/decoder,
	// so Stop() must NOT kill the already-open connection, only stop accepting new ones.
	defer server.Stop()

	// Open a TCP connection to the server (client side).
	conn, err := net.Dial("tcp", "localhost:"+PORT)
	require.NoError(t, err, "Consumer failed to connect")

	// Connection is intentionally NOT closed here (commented out),
	// because the test keeps using encoder/decoder after this function returns.
	// defer conn.Close()

	// Encoder writes JSON objects to the connection (one Encode = one JSON message + newline).
	encoder := json.NewEncoder(conn)

	// Decoder reads JSON objects from the connection (one Decode = read one JSON message).
	decoder := json.NewDecoder(conn)

	// Return tools used by the test to send requests and read responses.
	return encoder, decoder
}

// createSchool sends a "/school/create" request with only the school name,
// reads the response, converts Response.Data into a School struct,
// and asserts the returned school has the expected id and name.
func createSchool(
	t *testing.T,
	encoder *json.Encoder,
	decoder *json.Decoder,
	name string,
	id uint,
) School {
	// Build a request message for "/school/create".
	// Data contains a School with only Name set (Id must be assigned by the server).
	req := Request{
		Method: createSchoolMethod,
		Data: School{
			Name: name,
		},
	}

	// Send the request as one JSON object over the TCP connection.
	err := encoder.Encode(req)
	require.NoError(t, err, "Failed to create school")

	// Read the server response (one JSON object).
	var resp Response
	err = decoder.Decode(&resp)
	require.NoError(t, err, "Failed to create school")

	// Server must mark success.
	require.Equal(t, true, resp.Status, "Failed to create school")

	// Because Response.Data is interface{}, json.Decode fills it as map[string]any.
	// We first assert it's a JSON object.
	actualSchoolMap, ok := resp.Data.(map[string]any)
	require.Equal(t, true, ok, "Bad response")

	// Convert the map back into JSON bytes...
	jsonData, err := json.Marshal(actualSchoolMap)
	require.NoError(t, err, "Failed to create school")

	// ...then unmarshal into the School struct to get typed fields (Id, Name).
	var actualSchool School
	err = json.Unmarshal(jsonData, &actualSchool)
	require.NoError(t, err, "Failed to create school")

	// Expected: server assigned the provided id and echoed the name back.
	expectedSchool := School{
		Id:   id,
		Name: name,
	}

	// Strict equality check (so server must return exactly these fields/values).
	require.Equal(t, expectedSchool, actualSchool, "Bad response")

	// Return the created school for later test steps.
	return actualSchool
}

// createPerson sends a "/person/create" request with only the person's name,
// reads the response, converts Response.Data into a Person struct,
// and asserts the returned person has the expected id and name.
func createPerson(
	t *testing.T,
	encoder *json.Encoder,
	decoder *json.Decoder,
	name string,
	id uint,
) Person {
	// Build a request for "/person/create".
	// Data contains a Person with only Name set; server must assign Id.
	req := Request{
		Method: createPersonMethod,
		Data: Person{
			Name: name,
		},
	}

	// Send request as one JSON message.
	err := encoder.Encode(req)
	require.NoError(t, err, "Failed to create person")

	// Read response as one JSON message.
	var resp Response
	err = decoder.Decode(&resp)
	require.NoError(t, err, "Failed to create person")

	// Server must report success.
	require.Equal(t, true, resp.Status, "Failed to create person")

	// Response.Data decodes as map[string]any because it's interface{}.
	actualPersonMap, ok := resp.Data.(map[string]any)
	require.Equal(t, true, ok, "Bad response")

	// Convert the map back to JSON...
	// (Comment says OR, but this is the standard way to turn map[string]any into a typed struct.)
	jsonData, err := json.Marshal(actualPersonMap)
	require.NoError(t, err, "Failed to create class") // message text is a typo

	// ...then unmarshal into Person to get typed fields (Id, Name, Classes).
	var actualPerson Person
	err = json.Unmarshal(jsonData, &actualPerson)
	require.NoError(t, err, "Failed to create person")

	// Expected: server assigned the given id and echoed name back.
	expectedPerson := Person{
		Id:   id,
		Name: name,
	}

	// Strict equality check.
	require.Equal(t, expectedPerson, actualPerson, "Bad response")

	// Return the created person for later steps (teacher/student).
	return actualPerson
}

// createClass sends a "/class/create" request containing schoolId, teacher info, and class name,
// reads the response, converts Response.Data into a Class struct,
// and asserts the returned class has the expected id and fields
func createClass(
	t *testing.T,
	encoder *json.Encoder,
	decoder *json.Decoder,
	teacher Person,
	schoolId uint,
	name string,
	id uint,
) Class {
	// Build a request for "/class/create".
	// Data includes:
	// - SchoolId: which school this class belongs to
	// - Teacher: the person who will teach it (must already exist)
	// - Name: class name
	// Server must assign the class Id.
	req := Request{
		Method: createClassMethod,
		Data: Class{
			SchoolId: schoolId,
			Teacher:  teacher,
			Name:     name,
		},
	}

	// Send request as one JSON message.
	err := encoder.Encode(req)
	require.NoError(t, err, "Failed to create class")

	// Read response as one JSON message.
	var resp Response
	err = decoder.Decode(&resp)
	require.NoError(t, err, "Failed to create class")

	// Server must report success.
	require.Equal(t, true, resp.Status, "Failed to create class")

	// Response.Data is decoded into map[string]any (because it's interface{}).
	actualClassMap, ok := resp.Data.(map[string]any)
	require.Equal(t, true, ok, "Bad response")

	// Convert map -> JSON -> typed Class struct for strict comparison.
	jsonData, err := json.Marshal(actualClassMap)
	require.NoError(t, err, "Failed to create class")

	var actualClass Class
	err = json.Unmarshal(jsonData, &actualClass)
	require.NoError(t, err, "Failed to create class")

	// Expected: server returns the created class with assigned Id,
	// and echoes SchoolId, Name, and Teacher exactly.
	expectedClass := Class{
		Id:       id,
		Name:     name,
		Teacher:  teacher,
		SchoolId: schoolId,
	}

	// Strict equality check (so fields must match exactly).
	require.Equal(t, expectedClass, actualClass, "Bad response")

	// Return the created class for later steps (enrollment, etc.).
	return actualClass
}

// addStudentToClass sends a "/class/add/student" request with student_id and class_id,
// and expects the server to return the updated student Person with Classes containing the class id(s).
func addStudentToClass(
	t *testing.T,
	encoder *json.Encoder,
	decoder *json.Decoder,
	student Person,
	classId uint,
	expectedClasses []uint,
) {
	// Build a request for "/class/add/student".
	// Data only contains ids (student_id, class_id).
	req := Request{
		Method: addStudentToClassMethod,
		Data: AddStudentToClassReq{
			StudentId: student.Id,
			ClassId:   classId,
		},
	}

	// Send request as one JSON message.
	err := encoder.Encode(req)
	require.NoError(t, err, "Operation Failed")

	// Read response as one JSON message.
	var resp Response
	err = decoder.Decode(&resp)
	require.NoError(t, err, "Operation Failed")

	// Server must report success.
	require.Equal(t, true, resp.Status, "Operation Failed")

	// Response.Data decodes as map[string]any (because Data is interface{}).
	// The server is expected to return the UPDATED student person.
	actualPersonMap, ok := resp.Data.(map[string]any)
	require.Equal(t, true, ok, "Bad response")

	// Convert map -> JSON -> typed Person struct.
	jsonData, err := json.Marshal(actualPersonMap)
	require.NoError(t, err, "Operation Failed")

	var actualPerson Person
	err = json.Unmarshal(jsonData, &actualPerson)
	require.NoError(t, err, "Operation Failed")

	// Expected: same student Id/Name, but Classes updated to include classId.
	expectedPerson := Person{
		Id:      student.Id,
		Name:    student.Name,
		Classes: expectedClasses,
	}

	// Strict equality check.
	require.Equal(t, expectedPerson, actualPerson, "Bad response")
}

// whoAmI sends a "/who/am/i" request with only a person's Id,
// and expects the server to return that person's full info plus their related class ids in Classes.
func whoAmI(
	t *testing.T,
	encoder *json.Encoder,
	decoder *json.Decoder,
	person Person,
	expectedClasses []uint,
) {
	// Build a request for "/who/am/i".
	// Only Id is sent; server must look up the person and fill the rest.
	req := Request{
		Method: whoAmIMethod,
		Data: Person{
			Id: person.Id,
		},
	}

	// Send request as one JSON message.
	err := encoder.Encode(req)
	require.NoError(t, err, "Operation Failed")

	// Read response as one JSON message.
	var resp Response
	err = decoder.Decode(&resp)
	require.NoError(t, err, "Operation Failed")

	// Server must report success.
	require.Equal(t, true, resp.Status, "Operation Failed")

	// Response.Data should be a JSON object representing the found person.
	// It decodes as map[string]any because Data is interface{}.
	actualPersonMap, ok := resp.Data.(map[string]any)
	require.Equal(t, true, ok, "Bad response")

	// Convert map -> JSON -> typed Person struct.
	jsonData, err := json.Marshal(actualPersonMap)
	require.NoError(t, err, "Operation Failed")

	var actualPerson Person
	err = json.Unmarshal(jsonData, &actualPerson)
	require.NoError(t, err, "Operation Failed")

	// Expected: same Id/Name, and Classes filled with class IDs for that person.
	// (Teacher: classes they teach; Student: classes they attend.)
	expectedPerson := Person{
		Id:      person.Id,
		Name:    person.Name,
		Classes: expectedClasses,
	}

	// Strict equality check.
	require.Equal(t, expectedPerson, actualPerson, "Bad response")
}

// TestSimple_1 is an end-to-end test of the socket protocol:
// it creates a school, teacher, class, and student, enrolls the student,
// then verifies WhoAmI returns correct class ids for the teacher.
func TestSimple_1(t *testing.T) {
	// Create a TCP connection to the server and get JSON encoder/decoder.
	encoder, decoder := createConnection(t)

	// 1) Create a school. Server must assign Id=1.
	school_1 := createSchool(t, encoder, decoder, "school_1", 1)

	// 2) Create a person who will act as teacher. Server must assign Id=1 (people ids start from 1).
	teacher_1 := createPerson(t, encoder, decoder, "teacher_1", 1)

	// 3) Create a class in school_1 taught by teacher_1. Server must assign Id=1 (class ids start from 1).
	class_1 := createClass(t, encoder, decoder, teacher_1, school_1.Id, "class_1", 1)

	// 4) Create another person who will act as student. Server must assign Id=2.
	student_1 := createPerson(t, encoder, decoder, "student_1", 2)

	// After enrolling/looking up, the expected class list is just [1].
	expectedClasses := []uint{1}

	// 5) Enroll student_1 into class_1.
	// Server must return the updated student with Classes=[1].
	addStudentToClass(t, encoder, decoder, student_1, class_1.Id, expectedClasses)

	// 6) Ask "Who am I?" for the teacher.
	// Server must return the teacher with Classes=[1] meaning "classes they teach".
	whoAmI(t, encoder, decoder, teacher_1, expectedClasses)

	// Debug printing (not part of assertions).
	fmt.Println(school_1, teacher_1, class_1)
	println("end of test")
}
