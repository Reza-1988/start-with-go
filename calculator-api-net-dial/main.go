package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
	"strings"
)

// Server represents our TCP calculator service.
// It stores the listening port so Start() knows where to bind.
type Server struct {
	Port string // TCP port number as a string (e.g. "4001") used to listen on ":" + Port
}

// Consider whether you want Port exported (`Port`) or unexported (`port`).
//	- For this challenge, exported vs unexported doesn’t matter unless tests access it directly (they don’t in your sample).
//	- Unexported is usually cleaner, but we’ll only change it if needed.

// Response is the JSON message our server sends back to the client.
// Exactly one of these should be meaningful at a time:
// - On success: Result is filled and Error is empty.
// - On failure: Result is empty and Error contains the required message.
type Response struct {
	Result string `json:"result"` // Success message: "The result of your query is: %d"
	Error  string `json:"error"`  // Error message (empty string on success)
}

// Request is the JSON message the client sends to the server.
// Example:
//
//	{"action":"add","numbers":"2,1"}
type Request struct {
	Action  string `json:"action"`  // "add" or "sub"
	Numbers string `json:"numbers"` // Comma-separated int64 values (e.g. "2,1,-3")
}

// NewServer constructs a new Server configured to listen on the given TCP port.
// The server does not start listening until Start() is called.
func NewServer(port string) *Server {
	// Store the port so Start() can bind to ":" + port later.
	return &Server{
		Port: port,
	}
}

// Start begins listening on the server's TCP port and accepts client connections forever.
// For each new connection, it starts a separate goroutine to handle JSON requests on that connection.
func (s *Server) Start() {
	// Listen on TCP ":" + port (e.g. ":4001").
	// Using ":" binds on all interfaces; tests connect via localhost so this works.
	ln, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		// Start() has no return value in this challenge.
		// If we cannot listen, we simply stop.
		return
	}
	// Accept loop: runs forever, accepting new client connections.
	for {
		conn, err := ln.Accept()
		if err != nil {
			// If the listener is closed, Accept() may return net.ErrClosed (wrapped).
			// No Stop() is required in this exercise, but we still handle it safely.
			if errors.Is(err, net.ErrClosed) {
				return
			}
			// For transient accept errors, continue accepting future connections.
			continue
		}
		// Handle each connection concurrently so one slow client doesn't block others.
		go s.handleConn(conn)
	}
}

// handleConn processes one client TCP connection.
// It repeatedly reads a JSON request from the connection, computes a result/error, and writes a JSON response back.
// The loop ends when the client closes the connection or sends invalid JSON.
func (s *Server) handleConn(conn net.Conn) {
	// Always close the client connection when we're done handling it.
	defer conn.Close()

	dec := json.NewDecoder(conn) // Reads JSON requests from the connection
	enc := json.NewEncoder(conn) // Writes JSON responses to the connection

	// Allow multiple requests over the same TCP connection.
	for {
		var req Request
		if err := dec.Decode(&req); err != nil {
			// io.EOF means the client closed the connection cleanly.
			if errors.Is(err, io.EOF) {
				return
			}
			// For malformed JSON, we can safely close the connection.
			// The problem statement guarantees request format, so no special message needed.
			return
		}
		// Next step: compute the correct Response based on req.Action and req.Numbers
		// and send it back via enc.Encode(resp).
		// Route and compute the response based on req.Action and req.Numbers
		// and send it back via enc.Encode(resp).
		resp := s.route(req)

		// Write exactly one JSON response for each request.
		// If writing fails (client gone), stop handling this connection.
		if err := enc.Encode(resp); err != nil {
			return
		}
	}
}

// How does the request come from the client and get decoded?
//	- `conn` is like a two-way pipe between client and server.
// 	- On the client side, this line sends JSON bytes into the pipe:
//   	```encoder := json.NewEncoder(conn)
//		   encoder.Encode(addRequest)```
//		- Encode converts the Go map into JSON text (bytes) and writes it to `conn` (the TCP stream).
//	- On the server side, this reads from the same pipe:
//	    ```dec := json.NewDecoder(conn)
//         dec.Decode(&req)```
// 		- Decode blocks (waits) until enough bytes arrive to form one complete JSON object,
//		- then it parses them and fills req.

// route selects the correct operation handler based on the action.
// The problem guarantees action is either "add" or "sub", but we still keep a safe fallback.
// IMPORTANT: The fallback must use only allowed error messages from the prompt.
func (s *Server) route(req Request) Response {
	switch req.Action {
	case "add":
		return s.handleAdd(req.Numbers)
	case "sub":
		return s.handleSub(req.Numbers)
	default:
		// Keep output strictly within the expected/allowed error strings.
		return Response{
			Result: "",
			Error:  "Invalid number format",
		}
	}
}

// handleAdd validates and processes an "add" request.
// It must:
// - return "'numbers' parameter missing" if Numbers is missing/empty
// - return "Invalid number format" if any token isn't a valid int64
// - return "Overflow" if int64 addition overflows
// - otherwise return the success "result" string and empty "error"
func (s *Server) handleAdd(numbers string) Response {
	// Treat empty (or whitespace-only) numbers as "missing" per the prompt.
	if strings.TrimSpace(numbers) == "" {
		return Response{
			Result: "",
			Error:  "'numbers' parameter missing",
		}
	}

	// Split the comma-separated list into individual number strings.
	parts := strings.Split(numbers, ",")

	// We'll accumulate the sum here (real overflow-safe logic will come next).
	var sum int64
	for _, p := range parts {
		// Trim spaces so "1, 2" works.
		p = strings.TrimSpace(p)
		// An empty token like "1,,2" is not "missing parameter"; it's a bad format.
		if p == "" {
			return Response{
				Result: "",
				Error:  "Invalid number format",
			}
		}
		// Parse each token as int64 (base 10).
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return Response{
				Result: "",
				Error:  "Invalid number format",
			}
		}
		// Check int64 overflow BEFORE doing sum += n (Go would wrap otherwise).
		if (n > 0 && sum > math.MaxInt64-n) || (n < 0 && sum < math.MinInt64-n) {
			return Response{
				Result: "",
				Error:  "Overflow",
			}
		}
		sum += n
	}
	// Success response must match the exact required format.
	return Response{
		Result: fmt.Sprintf("The result of your query is: %d", sum),
		Error:  "",
	}
}

// handleSub validates and processes a "sub" request.
//
// Contract (from prompt):
// - If numbers is missing/empty -> "'numbers' parameter missing"
// - If any token is not a valid int64 -> "Invalid number format"
// - If any subtraction overflows int64 -> "Overflow"
// - Otherwise -> result message + empty error
func (s *Server) handleSub(numbers string) Response {
	// // Missing or empty numbers' parameter.
	if strings.TrimSpace(numbers) == "" {
		return Response{
			Result: "",
			Error:  "'numbers' parameter missing",
		}
	}
	parts := strings.Split(numbers, ",")

	// Subtraction must start from the first number
	// Parse the first number to initialize the subtraction chain:
	//	- result = first; then result -= each next number.
	first := strings.TrimSpace(parts[0])
	if first == "" {
		return Response{
			Result: "",
			Error:  "Invalid number format",
		}
	}
	res, err := strconv.ParseInt(first, 10, 64)
	if err != nil {
		return Response{
			Result: "",
			Error:  "Invalid number format",
		}
	}
	// Process remaining numbers.
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if p == "" {
			return Response{
				Result: "",
				Error:  "Invalid number format",
			}
		}

		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return Response{
				Result: "",
				Error:  "Invalid number format",
			}
		}
		// Overflow check for int64 subtraction: sub - n
		// - If n > 0, subtracting it can go below MinInt64.
		// - If n < 0, subtracting it is like adding |n| and can exceed MaxInt64.
		if (n > 0 && res < math.MinInt64+n) || (n < 0 && res > math.MaxInt64+n) {
			return Response{
				Result: "",
				Error:  "Overflow",
			}
		}
		res -= n
	}
	// Success response must match the exact required format.
	return Response{
		Result: fmt.Sprintf("The result of your query is: %d", res),
		Error:  "",
	}
}
