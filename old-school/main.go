package main

import (
	"net"
)

const (
	CreateSchoolMethod      = "/school/create"
	CreateClassMethod       = "/class/create"
	CreatePersonMethod      = "/person/create"
	AddStudentToClassMethod = "/class/add/student"
	WhoAmIMethod            = "/who/am/i"
)

func main() {
}

type Server interface {
	Start(port string) error
	Stop() error
}
type server struct {
	listener net.Listener
}

// Start begins listening on the given TCP port and accepts client connections.
// It blocks in an accept loop (so callers typically run it in a goroutine) and
// spawns a goroutine per connection to handle JSON requests/responses.
func (s *server) Start(port string) error {
	// Listen on TCP port (tests call net.Dial("tcp", "localhost:"+PORT)).
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

	// Accept loop: handle each connection concurrently.
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// If the listener was closed via Stop(), Accept returns an error.
			// Treat that as a normal shutdown.
			if ne, ok := err.(*net.OpError); ok && !ne.Temporary() {
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

		// Handle the connection in its own goroutine.
		go func(c net.Conn) {
			// For now, keep the connection open.
			// Later we will decode Request JSONs and encode Response JSONs here.
			// IMPORTANT: do not close the connection here yet, because tests keep using it.
			_ = c
		}(conn)
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

func NewServer() Server {

	return &server{}
}


}
