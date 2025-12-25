package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
)

type server struct {
	listener net.Listener

	// db is a pointer to a database handle from Go’s `database/sql` package.
	// It represents your connection access to SQLite (and manages things like pooling internally).
	// You store it on the server so any handler (`/school/create`, `/person/create`, etc.) can run queries using the same DB handle:
	// 	- s.db.Exec(...)
	//	- s.db.QueryRow(...)
	// 	- Think of db as: “the server’s database gateway”.
	db *sql.DB

	// dbOnce is a concurrency-safe tool that guarantees a function runs only one time,
	// even if multiple goroutines try to run it at the same time.
	// In a socket server, many clients can connect at once, and you may call DB initialization from Start() or connection handlers.
	// sync.Once prevents “init DB twice” bugs.
	// Example idea:
	// 	- Many goroutines call `initDBOnce()`
	// 	- But this ensures the DB is opened + tables created exactly once.
	// 	- So dbOnce is: “a guard that prevents double initialization”.
	dbOnce sync.Once
}

// Start begins listening on the given TCP port and accepts client connections.
// It blocks in an accept loop (so callers typically run it in a goroutine) and
// spawns a goroutine per connection to handle JSON requests/responses.
func (s *server) Start(port string) error {
	// 1. Init DB once
	if err := s.initDBOnce(); err != nil {
		return err
	}

	// 2. Listen on TCP port (tests call net.Dial("tcp", "localhost:"+PORT)).
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	// Why do we save `ln` in `s.listener`?
	//	- Because you need it later in `Stop()`.
	// 	- `ln` is a `net.Listener` that represents “the thing that is listening”.
	// 	- To stop the server, you need to call: `s.listener.Close()`
	//	- If you don’t store it, `Stop()` has nothing to close.
	// 	- Also: closing the listener is how you make the `Accept()` loop exit gracefully.
	s.listener = ln

	// 3. Accept loop: handle each connection concurrently.
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// If the listener was closed via Stop(), Accept returns an error.
			// Treat that as a normal shutdown.
			if ne, ok := err.(*net.OpError); ok && !ne.Temporary() { // TODO
				// This is handling the special case when you stop the server.
				// 	- When you call `s.listener.Close()` in `Stop()`, the listener becomes closed.
				// 	- Then the blocking call `Accept()` wakes up and returns an error like “use of closed network connection”.
				// 	- That error is usually a `*net.OpError.`
				// 	- So this code does:
				// 		- `err.(*net.OpError)` → “Is this error a network operation error?”
				// 		- ok is true if the cast succeeded
				// 		- `!ne.Temporary()` → means “this is not a temporary error” (it won’t go away by retrying)
				//	- So if:
				// 		- the listener was closed (normal shutdown)
				// 		- `Accept`() returns a non-temporary OpError
				// 	- then we `return nil` meaning:
				// 		- “This is a normal stop, not a real failure.”
				// 		- Without this, `Start()` would return an error on shutdown and tests might treat that as “server failed to start”.
				return nil
			}
			// Any other error: return it so Start reports failure.
			return err
		}

		// 4. Handle the connection in its own goroutine.
		go s.handleConn(conn)
	}
}

// Stop gracefully stops the server from accepting NEW connections by closing the listener.
// It intentionally does not force-close already accepted connections, so in-flight clients
// can continue using their existing net.Conn (the tests rely on this behavior).
func (s *server) Stop() error {
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}

// What is the “listener” vs an “active connection”?
// 	- istener (`net.Listener`) = the thing created by `net.Listen(...)`
// 		- Its job: wait for new clients
// 		- It only does: `Accept()` → gives you a `conn`
// 	- Connection (`net.Conn`) = the “pipe” created after a client connects
// 		- Its job: send/receive data
// 		- This is what `json.Encoder/Decoder` use
// 	- So:
// 		- Listener = “door of the server”
// 		- Conn = “a phone call already in progress”
//
// Why “close only the listener” in `Stop()`?
// 	- In your test, this happens:
// 		1. Server starts
// 		2. Test calls `net.Dial(...)` → connection is created
//		3. Test creates `encoder/decoder` and returns them
//		4. Immediately after returning, `defer server.Stop()` runs (because it was deferred inside `createConnection`)
//	- But the test still wants to use the same connection to send requests after that.
// 	- So `Stop()` must behave like:
// 		- Close the “door” so no new clients can connect
//		-  Do NOT hang up the ongoing “phone call” (the existing connection), because the test is still using it.

// handleConn reads Request JSON messages from one TCP connection and writes Response JSON messages back.
// Protocol: each json.Encoder.Encode(...) from the client produces one JSON object to Decode.
func (s *server) handleConn(conn net.Conn) {
	// It's fine to close when the client disconnects(Decode returns EOF)
	// Test don't close the conn explicitly, so this handler should keep running.
	defer conn.Close()
	// Why do we close the connection in handleConn?
	//	- `defer conn.Close()` does NOT close the connection immediately.
	//	- It means: “Close it when handleConn finishes.”
	// 	- And handleConn only finishes when:
	// 		- the client disconnects (Decode returns io.EOF), or
	// 		- there is a fatal error (bad JSON / broken connection)
	// 		- So we close it for cleanup.
	//		- Otherwise connections would stay open forever and leak resources.
	// Important point:
	// 	- While the test is running and sending requests, handleConn is still looping, so the connection stays open.

	dec := json.NewDecoder(conn) // `dec` read JSON objects from the connection.
	enc := json.NewEncoder(conn) // `enc` writes JSON objects to the connection.

	// Infinite loop.
	// Because this connection can be used for many requests, not just one.
	for {
		var req Request // Create a variable to store one incoming request.
		// Start of decode error handling
		// Try to read one JSON object from the connection and decode it into req.
		// If the client sends invalid JSON or disconnects, Decode returns an error.
		if err := dec.Decode(&req); err != nil {
			// Normal end of connection
			// 	- io.EOF means “end of file” → in sockets it means the client closed the connection.
			// 	- That’s not a bug. We just exit handleConn, and because of defer conn.Close(), it cleans up.
			if errors.Is(err, io.EOF) {
				return
			}
			// Bad JSON / decode error: respond and then close connection
			// 	- If it’s not EOF, it’s an actual decode problem (bad JSON, broken stream, etc.).
			// 	- We try to send an error response back.
			// 	- `_ = enc.Encode(...)` ignores the encode error (because if decoding failed, connection might already be broken).
			// 	- Then return ends the function and closes the connection.
			_ = enc.Encode(Response{
				Status:  false,
				Message: "Invalid request",
			})
			return
		} // End of decode error handling.

		// We now have a valid request.
		//	- We pass it to `route` which decides what to do based on `req.Method`
		//	- (e.g. /school/create) and returns a Response.
		resp := s.route(req)
		// Send the response as JSON back to the client.
		// Again we ignore encode errors because if the client disappears suddenly, there’s nothing useful to do here.
		_ = enc.Encode(resp)
	} // Loop continues, waiting for the next request on the same connection.
}
