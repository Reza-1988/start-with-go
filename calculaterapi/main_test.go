package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	port = "4001"
	path = "http://127.0.0.1:4001"
)

// This struct matches the JSON response format of the API.
// json.Unmarshal will automatically map JSON fields into these struct fields.
type Response struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

// Global variable to keep a reference to the started server_httpclient
// (so it won't start multiple instances)
var testServer *Server

// Helper function to start server_httpclient once and return it.
// Runs server_httpclient in a goroutine (non-blocking) so tests can continue.
func getServer() *Server {
	if testServer == nil {
		testServer = NewServer(port)

		// Start server_httpclient in a new goroutine
		// Without "go", Start() blocks forever and test does not run
		go testServer.Start()
	}

	// Give server_httpclient a short moment to fully start (port binding)
	time.Sleep(100 * time.Millisecond)
	return testServer
}

// ------------ TESTS ------------

// Ensures that the server_httpclient object can be created successfully
func TestSampleCreation(t *testing.T) {
	s := getServer()

	// assert.NotNil fails test if "s" is nil
	assert.NotNil(t, s)
}

// Ensures that Start() function actually opens the TCP port
func TestSampleServerStart(t *testing.T) {
	getServer()

	// net.Dial tries to connect to the server_httpclient TCP port.
	// If server_httpclient is not running, err will NOT be nil.
	conn, err := net.Dial("tcp", "localhost:"+port)

	assert.Nil(t, err) // connection should succeed

	defer conn.Close()
}

// Ensures that /add endpoint responds with StatusCode 200
func TestSampleAddHandler(t *testing.T) {
	getServer()

	// Sending HTTP GET request to our API
	resp, err := http.DefaultClient.Get(path + "/add?numbers=1,2,3")

	assert.Nil(t, err)                    // ensure request was successful
	assert.Equal(t, 200, resp.StatusCode) // API should return HTTP 200

	defer resp.Body.Close()
}

// Ensures subtraction endpoint returns correct result
func TestSampleSubtraction(t *testing.T) {
	getServer()

	// Example: 104 - 204 = -100
	resp, err := http.DefaultClient.Get(path + "/sub?numbers=104,204")
	assert.Nil(t, err)

	defer resp.Body.Close()

	// Read response body as []byte
	s, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	// Create Zero-value struct to receive the JSON
	var response Response

	// Pass pointer to Unmarshal so it can fill struct fields
	// Unmarshal modifies the actual struct, not a copy.
	err = json.Unmarshal(s, &response)
	assert.Nil(t, err)

	// Validate JSON result field
	assert.Equal(t, "The result of your query is: -100", response.Result)
}
