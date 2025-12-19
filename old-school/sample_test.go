package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	PORT = "8080"

	createSchoolMethod      = "/school/create"
	createClassMethod       = "/class/create"
	createPersonMethod      = "/person/create"
	addStudentToClassMethod = "/class/add/student"
	whoAmIMethod            = "/who/am/i"
)

func startTestServer(port string) Server {
	server := NewServer()
	go func() {
		err := server.Start(port)
		if err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Give server some time to start
	time.Sleep(200 * time.Millisecond)

	return server
}

func createConnection(t *testing.T) (*json.Encoder, *json.Decoder) {
	server := startTestServer(PORT)
	defer server.Stop()

	conn, err := net.Dial("tcp", "localhost:"+PORT)
	require.NoError(t, err, "Consumer failed to connect")
	// defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	return encoder, decoder
}

func createSchool(
	t *testing.T,
	encoder *json.Encoder,
	decoder *json.Decoder,
	name string,
	id uint,
) School {
	req := Request{
		Method: createSchoolMethod,
		Data: School{
			Name: name,
		},
	}
	err := encoder.Encode(req)
	require.NoError(t, err, "Failed to create school")

	var resp Response
	err = decoder.Decode(&resp)
	require.NoError(t, err, "Failed to create school")
	require.Equal(t, true, resp.Status, "Failed to create school")

	actualSchoolMap, ok := resp.Data.(map[string]any)
	require.Equal(t, true, ok, "Bad response")

	jsonData, err := json.Marshal(actualSchoolMap)
	require.NoError(t, err, "Failed to create school")

	var actualSchool School
	err = json.Unmarshal(jsonData, &actualSchool)
	require.NoError(t, err, "Failed to create school")

	expectedSchool := School{
		Id:   id,
		Name: name,
	}
	require.Equal(t, expectedSchool, actualSchool, "Bad response")
	return actualSchool
}

func createPerson(
	t *testing.T,
	encoder *json.Encoder,
	decoder *json.Decoder,
	name string,
	id uint,
) Person {
	req := Request{
		Method: createPersonMethod,
		Data: Person{
			Name: name,
		},
	}
	err := encoder.Encode(req)
	require.NoError(t, err, "Failed to create person")

	var resp Response
	err = decoder.Decode(&resp)
	require.NoError(t, err, "Failed to create person")
	require.Equal(t, true, resp.Status, "Failed to create school")

	actualPersonMap, ok := resp.Data.(map[string]any)
	require.Equal(t, true, ok, "Bad response")

	// OR
	jsonData, err := json.Marshal(actualPersonMap)
	require.NoError(t, err, "Failed to create class")

	var actualPerson Person
	err = json.Unmarshal(jsonData, &actualPerson)
	require.NoError(t, err, "Failed to create person")

	expectedPerson := Person{
		Id:   id,
		Name: name,
	}
	require.Equal(t, expectedPerson, actualPerson, "Bad response")
	return actualPerson
}

func createClass(
	t *testing.T,
	encoder *json.Encoder,
	decoder *json.Decoder,
	teacher Person,
	schoolId uint,
	name string,
	id uint,
) Class {
	req := Request{
		Method: createClassMethod,
		Data: Class{
			SchoolId: schoolId,
			Teacher:  teacher,
			Name:     name,
		},
	}
	err := encoder.Encode(req)
	require.NoError(t, err, "Failed to create class")

	var resp Response
	err = decoder.Decode(&resp)
	require.NoError(t, err, "Failed to create class")
	require.Equal(t, true, resp.Status, "Failed to create class")

	actualClassMap, ok := resp.Data.(map[string]any)
	require.Equal(t, true, ok, "Bad response")

	jsonData, err := json.Marshal(actualClassMap)
	require.NoError(t, err, "Failed to create class")

	var actualClass Class
	err = json.Unmarshal(jsonData, &actualClass)
	require.NoError(t, err, "Failed to create class")

	expectedClass := Class{
		Id:       id,
		Name:     name,
		Teacher:  teacher,
		SchoolId: schoolId,
	}
	require.Equal(t, expectedClass, actualClass, "Bad response")
	return actualClass
}

// add student to class
func addStudentToClass(
	t *testing.T,
	encoder *json.Encoder,
	decoder *json.Decoder,
	student Person,
	classId uint,
	expectedClasses []uint,
) {
	req := Request{
		Method: addStudentToClassMethod,
		Data: AddStudentToClassReq{
			StudentId: student.Id,
			ClassId:   classId,
		},
	}
	err := encoder.Encode(req)
	require.NoError(t, err, "Operation Failed")

	var resp Response
	err = decoder.Decode(&resp)
	require.NoError(t, err, "Operation Failed")
	require.Equal(t, true, resp.Status, "Operation Failed")

	actualPersonMap, ok := resp.Data.(map[string]any)
	require.Equal(t, true, ok, "Bad response")

	jsonData, err := json.Marshal(actualPersonMap)
	require.NoError(t, err, "Operation Failed")

	var actualPerson Person
	err = json.Unmarshal(jsonData, &actualPerson)
	require.NoError(t, err, "Operation Failed")

	expectedPerson := Person{
		Id:      student.Id,
		Name:    student.Name,
		Classes: expectedClasses,
	}
	require.Equal(t, expectedPerson, actualPerson, "Bad response")
}

func whoAmI(
	t *testing.T,
	encoder *json.Encoder,
	decoder *json.Decoder,
	person Person,
	expectedClasses []uint,
) {
	req := Request{
		Method: whoAmIMethod,
		Data: Person{
			Id: person.Id,
		},
	}
	err := encoder.Encode(req)
	require.NoError(t, err, "Operation Failed")

	var resp Response
	err = decoder.Decode(&resp)
	require.NoError(t, err, "Operation Failed")
	require.Equal(t, true, resp.Status, "Operation Failed")

	actualPersonMap, ok := resp.Data.(map[string]any)
	require.Equal(t, true, ok, "Bad response")

	jsonData, err := json.Marshal(actualPersonMap)
	require.NoError(t, err, "Operation Failed")

	var actualPerson Person
	err = json.Unmarshal(jsonData, &actualPerson)
	require.NoError(t, err, "Operation Failed")

	expectedPerson := Person{
		Id:      person.Id,
		Name:    person.Name,
		Classes: expectedClasses,
	}
	require.Equal(t, expectedPerson, actualPerson, "Bad response")
}

func TestSimple_1(t *testing.T) {
	encoder, decoder := createConnection(t)
	school_1 := createSchool(t, encoder, decoder, "school_1", 1)
	teacher_1 := createPerson(t, encoder, decoder, "teacher_1", 1)
	class_1 := createClass(t, encoder, decoder, teacher_1, school_1.Id, "class_1", 1)
	student_1 := createPerson(t, encoder, decoder, "student_1", 2)

	expectedClasses := []uint{1}
	addStudentToClass(t, encoder, decoder, student_1, class_1.Id, expectedClasses)
	whoAmI(t, encoder, decoder, teacher_1, expectedClasses)
	fmt.Println(school_1, teacher_1, class_1)
	println("end of test")
}
