package calculater_api_net_dial

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// connectToServer opens a TCP connection to localhost:<port>.
func connectToServer(t *testing.T, port string) net.Conn {
	// Dial means “connect to a TCP server”.
	conn, err := net.Dial("tcp", "localhost:"+port)

	// If connection fails, stop the test immediately.
	require.NoError(t, err, "Failed to connect to server")
	return conn
}

func TestTCPServer_Add(t *testing.T) {
	// Pick a port for the server to listen on.
	port := "4001"

	// Create a new server instance configured for this port.
	srv := NewServer(port)

	// Start the server in a goroutine so the test can continue.
	// Start() should block forever (accept loop), so we must run it in background.
	go srv.Start()

	// Give the server a short time to start listening before we connect.
	time.Sleep(100 * time.Millisecond)

	// Connect to the running TCP server.
	conn := connectToServer(t, port)

	// Make sure the connection is closed when the test ends.
	defer conn.Close()

	// Encoder writes JSON to the connection.
	encoder := json.NewEncoder(conn)

	// Decoder reads JSON from the connection.
	decoder := json.NewDecoder(conn)

	// Build the request object we want to send.
	// action="add" means sum, numbers is a comma-separated string.
	addRequest := map[string]interface{}{
		"action":  "add",
		"numbers": "2,1",
	}

	// Send the JSON request over TCP.
	// Encode() writes one JSON object (and a newline).
	err := encoder.Encode(addRequest)
	require.NoError(t, err, "Failed to send add request")

	// Read one JSON response back from the server.
	// The server must respond with a JSON object that matches Response struct.
	var response Response
	err = decoder.Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	// Check the server computed 2 + 1 = 3 and formatted the result message correctly.
	require.Equal(
		t,
		fmt.Sprintf("The result of your query is: %d", 3),
		response.Result,
		"Incorrect sum",
	)

	// For a valid request, error must be an empty string.
	require.Equal(t, "", response.Error, "Error should be empty")
}
