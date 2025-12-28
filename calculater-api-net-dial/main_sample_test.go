package calculater_api_net_dial

import (
	"encoding/json" // JSON encoder/decoder for sending requests and receiving responses over the TCP connection.
	"fmt"           // Used here to format the expected result string (matching the server's required output format).
	"net"           // Provides net.Dial for opening a TCP client connection to our server.
	"testing"       // Go's standard testing framework.
	"time"          // Used for a short sleep to give the server time to start listening before we connect.

	"github.com/stretchr/testify/require" // Helpful test assertions: fail the test immediately when a condition is not met.
)

// connectToServer is a small helper that opens a TCP connection to the server under test.
// It keeps the test body cleaner and ensures connection errors fail the test immediately.
func connectToServer(t *testing.T, port string) net.Conn {
	t.Helper() // Marks this function as a test helper so failures are reported at the caller line.

	// Try to connect to the TCP server running on localhost at the given port.
	conn, err := net.Dial("tcp", "localhost:"+port)

	// If we cannot connect, the server is not reachable (not started, wrong port, etc.).
	// require.NoError stops the test immediately if err != nil.
	require.NoError(t, err, "Failed to connect to server")

	// Return the open connection so the caller can send requests and read responses.
	return conn
}

// TestTCPServer_Add verifies that the TCP calculator server correctly handles an "add" request.
// The test starts the server, connects as a client, sends a JSON request, reads the JSON response,
// and checks that the result and error fields match the expected contract.
func TestTCPServer_Add(t *testing.T) {
	// Choose a test port. In real test suites you might randomize or allocate a free port,
	// but here a fixed port is used for simplicity.
	port := "4001"

	// Build a new server configured to listen on the chosen port.
	srv := NewServer(port)

	// Start the server in a goroutine so the test can continue.
	// Start() is expected to block while serving, so we run it concurrently.
	go srv.Start()

	// Give the goroutine time to start listening on the port.
	// Without this, the Dial below might race and fail if the server isn't ready yet.
	time.Sleep(100 * time.Millisecond)

	// Open a TCP connection to the server (client side).
	conn := connectToServer(t, port)

	// Ensure the connection is closed at the end of the test, even if assertions fail.
	defer conn.Close()

	// Create a JSON encoder that writes JSON values to the TCP connection.
	// encoder.Encode(...) will send a JSON object followed by a newline.
	encoder := json.NewEncoder(conn)

	// Create a JSON decoder that reads JSON values from the TCP connection.
	// decoder.Decode(...) will block until it can parse one full JSON value.
	decoder := json.NewDecoder(conn)

	// Build the request payload as a generic JSON object.
	// The server contract expects:
	// - "action": either "add" or "sub"
	// - "numbers": a comma-separated string of integers
	addRequest := map[string]interface{}{
		"action":  "add", // Request addition.
		"numbers": "2,1", // Add 2 and 1 -> expected result is 3.
	}

	// Send the request to the server as JSON.
	// Encode writes the JSON object to conn (and typically appends a newline).
	err := encoder.Encode(addRequest)

	// If the request cannot be written, something is wrong with the connection/server.
	require.NoError(t, err, "Failed to send add request")

	// Read one JSON response object back from the server.
	// Response should have exported fields like Result and Error to decode into.
	var response Response
	err = decoder.Decode(&response)

	// If we cannot decode a response, server didn't respond with valid JSON.
	require.NoError(t, err, "Failed to decode response")

	// Verify the server returned the exact expected result string.
	// The problem statement requires: "The result of your query is: %d"
	require.Equal(
		t,
		fmt.Sprintf("The result of your query is: %d", 3),
		response.Result,
		"Incorrect sum",
	)

	// Verify the error field is empty for a successful request.
	require.Equal(t, "", response.Error, "Error should be empty")
}
