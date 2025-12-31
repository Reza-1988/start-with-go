package server

import (
	"QueraMQ/queue"
	"container/heap"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/google/uuid"
)

// Server is a TCP message broker.
//
// It accepts multiple concurrent client connections (producers and consumers),
// manages topics (creating them on demand), and coordinates graceful shutdown.
//
// Design goals reflected in fields below:
//   - Concurrency safety: maps are protected by RWMutex.
//   - Graceful shutdown: Stop() closes the listener and all connections and
//     waits for handler goroutines to exit.
//   - Test visibility: Addr is exported as required by the assignment.
type Server struct {
	// Addr is the TCP address the server listens on (e.g. "127.0.0.1:8080").
	Addr string

	// ln is the TCP listener created by Run(). Kept so Stop() can close it,
	// which unblocks Accept() and allows Run() to exit gracefully.
	ln net.Listener

	// topics holds all topics known to the server (created on subscribe/publish).
	// Key: topic name.
	topics   map[string]*Topic
	topicsMu sync.RWMutex

	// clients tracks currently connected clients.
	// Key: client ID (best practice here: conn.RemoteAddr().String()).
	// Value: a wrapper that gives safe, serialized writes (JSON Encode).
	clients   map[string]*Client
	clientsMu sync.RWMutex

	// stopCh is closed when the server is shutting down. Goroutines can select
	// on it to exit early.
	stopCh chan struct{}

	// stopOnce ensures Stop() is idempotent and safe to call multiple times.
	stopOnce sync.Once

	// wg tracks goroutines started by the server (e.g. per-connection handlers),
	// so Stop() can wait for them to finish.
	wg sync.WaitGroup
}

// Client wraps a net.Conn and provides safe, serialized JSON writes.
//
// This is important because producers and the server dispatcher may both write
// JSON responses/deliveries to the same connection concurrently; without a write
// mutex, JSON messages can interleave and corrupt the stream.
type Client struct {
	// conn is the underlying TCP connection.
	conn net.Conn

	// enc encodes JSON to conn. Must be protected by mu (Encode is not safe
	// for concurrent use on the same writer).
	enc *json.Encoder

	// mu serializes all writes (enc.Encode) to avoid interleaved JSON output.
	mu sync.Mutex

	// id uniquely identifies the client (typically conn.RemoteAddr().String()).
	id string

	// subs tracks which topics this client is currently subscribed to.
	// This makes "close_connection" easy: remove from all topics.
	subs   map[string]struct{}
	subsMu sync.Mutex
}

// - What is id field in Client?
// 		- The id is a unique identifier for each client so that the server can distinguish each connection from the others
// 	      (for logging, debugging, managing clients, etc.).
//		- We usually get this identifier from the client-side address on TCP.
// - What does conn.RemoteAddr().string() mean?
//		- conn is a TCP connection (net.Conn)
// 		- RemoteAddr() returns the address of the remote party (client)
//		- Usually something like this: "192.168.1.10:53022"
// 			- That is:
// 				- Client IP: 192.168.1.10
// 				- Port from which the client connected to the server: 53022 (temporary/ephemeral port)
// 		- And String() only converts this address to a string.
//	- Why is this a good option?
// 		- Because for each TCP connection the IP:Port combination is usually unique.
// 		- Even if two people connect from the same IP, their port will be different, so the id will be different.
// 		- So without having to create a counter or UUID yourself, you have a ready-made ID.
//	- What is this ID for?
//		1. Logging / Debugging
//		2. Maintaining clients on map
//		3. Identifying which connection the message/request came from
//	- Is it always really “unique”?
// 		- For “connection ID” almost yes.
//		- But if the client disconnects and reconnects, it may come with a new port and the id will change
//	      (which is usually not a problem because it is a new connection).

// NewServer constructs a Server configured to listen on the given address.
//
// It initializes internal maps/channels so the server is ready to Run().
// NewServer does not start listening; Run() is responsible for creating the listener.
func NewServer(address string) *Server {
	return &Server{
		Addr: address,

		// Initialize maps to avoid nil map panics when adding topics/clients.
		topics:  make(map[string]*Topic),
		clients: make(map[string]*Client),

		// stopCh is closed during shutdown so goroutines can exit via select.
		stopCh: make(chan struct{}),
	}
}

// Run starts the TCP server and blocks, accepting connections until the server
// is stopped.
//
// Responsibilities of Run (best-practice split of concerns):
//   - Create and store the listener (so Stop() can close it).
//   - Accept client connections in a loop.
//   - Register each client in the server's connection registry.
//   - Start a goroutine per client to handle JSON requests and async deliveries.
//   - Exit gracefully when Stop() is called (listener closed / stopCh closed).
//
// Protocol handling (publish/subscribe/unsubscribe/close/shutdown) should be
// implemented in s.handleClient; Run should stay focused on networking + lifecycle.
func (s *Server) Run() error {
	// Create the TCP listener. If this fails (e.g. port already in use),
	// return the error so the test can fail early.
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	// Store the listener so Stop() can close it to unblock Accept().
	s.ln = ln
	for {
		// Accept blocks until a new client connects or the listener is closed.
		conn, err := ln.Accept()
		if err != nil {
			// If shutdown has started, Accept will typically fail because
			// the listener is closed. In that case, exit Run cleanly.
			select {
			case <-s.stopCh:
				return nil
			default:
			}
			// net.ErrClosed is the canonical error when a listener is closed.
			// Treat it as a normal shutdown path.
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			// For any other error, return it ( tests should see the failure).
			return nil
		}
		// Wrap the raw connection in a Client to support safe JSON writes.
		// (Encoder writes must be serialized per connection to avoid interleaving.)
		c := &Client{
			conn: conn,
			enc:  json.NewEncoder(conn),
			id:   conn.RemoteAddr().String(),
			subs: make(map[string]struct{}),
		}
		// Track the client so GetClientConnections() can return active conns
		// and Stop() can close all connections during shutdown.
		s.clientsMu.Lock()
		s.clients[c.id] = c
		s.clientsMu.Unlock()

		// Handle this connection concurrently.
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleClient(c) // request loop & protocol
		}()
	}
}

// Stop gracefully shuts down the server.
//
// It is safe to call Stop multiple times. On shutdown it:
//   - broadcasts a stop signal (by closing stopCh),
//   - closes the listener to unblock Accept(),
//   - closes all client connections to unblock per-client Decode loops,
//   - waits for server goroutines to exit.
func (s *Server) Stop() {
	// `stopOnce.Do()` is of type `sync.Once`.
	//	- This means that even if Stop() is called multiple times, the body inside it will only be executed once.
	//	- So stopping again will not cause a panic (e.g. closing the channel again).
	s.stopOnce.Do(func() {
		// Broadcast shutdown to all goroutines that select on stopCh.
		// 	- stopCh is a channel that goroutines may check inside select:
		// 		- By closing it:
		// 			- All goroutines waiting for <-stopCh will immediately wake up
		// 			- and realize they need to exit
		// 		- This is “broadcasting” the shutdown signal.
		close(s.stopCh)

		// Closing the listener unblocks Accept() in Run().
		// 	- `ln` is the TCP listener (net.Listener) that gets stuck on `Accept()` in Run().
		// 	- If you close the listener:
		// 		- Accept() will throw an error and exit the block
		// 		- And the new connection acceptance loop can terminate cleanly
		// 	- `_ =` means we ignore the Close error (usually not important in shutdown).
		if s.ln != nil {
			_ = s.ln.Close()
		}

		// Snapshot client connections, then close them without holding the lock.
		s.clientsMu.RLock()
		clients := make([]*Client, 0, len(s.clients))
		for _, c := range s.clients {
			clients = append(clients, c)
		}
		s.clientsMu.RUnlock()

		// Closing client conns unblocks json.Decoder.Decode() in handleClient.
		for _, c := range clients {
			_ = c.conn.Close()
			// Why `_ = c.conn.Close()`?
			//	- Because in shutdown:
			// 	- Some connections may have already been closed
			// 	- `Close()` may give an error
			// 	- But for shutdown it usually doesn't matter; the goal is just to "close", so we ignore the error.
		}

		// Wait for all goroutines started by Run() to finish.
		s.wg.Wait()
	}) // End of `DO`
}

// Why isn't just close(stopCh) enough?
// 	- Many servers have a goroutine for each client that works like this:
// 	- 	```func (s *Server) handleClient(c *Client) {
//          	dec := json.NewDecoder(c.conn)
//				for {
//        			var req Request
//       			if err := dec.Decode(&req); err != nil {
// 					// Here, when the connection is closed, Decode throws an error, and we exit the loop.
//					return
//        			}
//        	// ... processing the request ...
//    			}
//			}```
// - Note:
//		- `dec.Decode()` usually waits for data to arrive from the network.
//		- So if the client hasn't sent anything, this line can block for a long time (even forever).
// 		- Now if you just close `stopCh`:
// 			- The goroutine that is now stuck inside `Decode()` doesn't have a chance to check `stopCh` at all
// 			- Because it hasn't reached a select or conditional check yet; it's stuck in I/O.
// 		- So to get this goroutine out of the block, you need to make `Decode()` wake up.
// - Solution: Closing the connection wakes up Decode
//		```for _, c := range clients {
//			_ = c.conn.Close()```
//			}
// 		- When you call `c.conn.Close()`:
// 		1) Any goroutines that are currently reading/decoding on this connection…
//			- It comes out of the block.
// 			- `Decode()` returns an error (usually something like EOF or “use of closed network connection”)
//		2) The `handleClient` loop that sees the error…
// 			- returns
// 			- The goroutine ends cleanly
// 			- And if you have `wg.Done()`, `wg.Wait()` will eventually be freed
// 		- That is: the main purpose of closing connections is to “unblock” goroutines that are stuck on the network.
// Why do we take a "snapshot" and release the lock first?
// 	- Important reason:
// 		- If you hold the lock and start Close()ing connections at the same time, two problems can arise:
// 		- Problem 1: Slowness/Stuckness of Others
// 			- `Close()` is usually fast, but it is considered an "I/O operation" and is better not to do it under a lock.
// 			- If you do it under a lock, all other goroutines that want to modify (or remove) clients will be stuck behind the lock.
//		- Problem 2: Possible deadlock in cleanup
// 			- It is very common for handleClient to want to remove itself from s.clients at the end of its work:
// 			 ```defer func() {
//					s.clientsMu.Lock()
//					delete(s.clients, c.id)
//					s.clientsMu.Unlock()
//				}()
// 			- Now imagine:
// 				- `Stop()` holds the lock and is doing a `Close()`
// 				- `Close()` causes handleClient to jump out and into defer and try to get `clientsMu.Lock()`
// 				- But the lock is still in Stop()’s hands…
// 				- This can cause a hang/deadlock or at least a severe slowdown.

// GetTopic returns the Topic for the given name.
//
// Behavior required by the assignment:
//   - If the topic already exists, return (topic, true).
//   - If the topic does not exist, create it, store it, and return (topic, false).
//
// Concurrency:
// Multiple goroutines (client handlers) may call GetTopic at the same time.
// We use a double-check pattern:
//  1. Fast path with RLock for existing topics.
//  2. If missing, upgrade to Lock and check again before creating,
//     to avoid creating the same topic twice.
func (s *Server) GetTopic(topicName string) (*Topic, bool) {
	// Fast path: check under read lock.
	s.topicsMu.RLock()
	t, ok := s.topics[topicName]
	s.topicsMu.RUnlock()
	if ok {
		return t, true // topic already existed
	}

	// Slow path: create under write lock (with double-check).
	s.topicsMu.Lock()
	defer s.topicsMu.Unlock()

	// Another goroutine might have created the topic between locks.
	if t, ok := s.topics[topicName]; ok {
		return t, true
	}

	// Create and store the new topic.
	t = NewTopic(topicName)
	s.topics[topicName] = t

	// Start per-topic dispatcher so messages are delivered in priority order.
	//
	// 	- wg is a sync.WaitGroup. With Add(1) you say:
	// 		- "I'm starting a new goroutine; so the number of running tasks will be +1."
	// 	- This will cause the server to wait until this goroutine has finished completely when you later call Stop() or shutdown and `s.wg.Wait()` is called.
	//	- If you don't add(1), `Wait()` may return early and the server will think everything is finished, while the dispatcher is still running.
	s.wg.Add(1)
	go func() {
		// `Done()` is the exact opposite of `Add(1)`: that is, when the goroutine finishes:
		// 	- "This job is done; WaitGroup counter is set to -1."
		//	- And because it is defer, even if the topicLoop finishes early with error/return, Done() will still be executed and wg will not get stuck.
		defer s.wg.Done()
		// This is where the main dispatcher logic is implemented. Typically, topicLoop does the following:
		// 	- Waits on `t.notify `or `stopCh`
		// 	- When a new message arrives, it pops from t.MQ (heap)
		// 	- Sends messages to the subscriptions of the same topic (observing joinedSeq etc.)
		// Why does this only happen when the topic is newly created?
		// 	- Because the dispatcher needs to be started only once for each topic.
		// 	- If a topic already existed, the dispatcher would have already been started.
		s.topicLoop(t)
	}()
	// This piece of goroutine means:
	// Since we have just created a new topic, start a separate worker/dispatcher for it to consume and
	// distribute messages from this topic in order (priority).
	// And for shutdown, use WaitGroup to make sure the server waits until this worker finishes.

	return t, false // topic was newly created
}

// GetClientConnections returns a snapshot of all active client connections.
//
// The returned slice is a point-in-time view: clients may connect/disconnect
// immediately after this method returns. Callers should not assume the slice
// stays in sync with the server's internal state.
func (s *Server) GetClientConnections() []net.Conn {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	// Preallocate for efficiency (best practice when size is known).
	conns := make([]net.Conn, 0, len(s.clients))
	for _, c := range s.clients {
		conns = append(conns, c.conn)
	}
	return conns
}

// handleClient is the per-connection request loop.
//
// It continuously reads JSON requests from one TCP client (producer or consumer)
// and performs the requested action until the client disconnects or asks to close.
//
// Why this loop is needed:
//   - Clients keep a long-lived TCP connection.
//   - Over that single connection they can send multiple requests (publish/subscribe/...).
//   - The server must respond and also may asynchronously deliver messages.
//
// Protocol rules enforced (per assignment):
//   - subscribe/unsubscribe require "topic".
//   - publish requires "message" and message.topic/content/priority.
//   - success => {"status":"ok"}
//   - errors  => {"error":"..."}
//   - unknown action => {"error":"unknown action"}
//   - close_connection => remove from all topics then close
//   - shutdown => graceful server Stop()
func (s *Server) handleClient(c *Client) {
	// Decoder reads a stream of JSON objects from the TCP connection.
	// Each client request is one JSON object encoded by json.Encoder on the client side.
	dec := json.NewDecoder(c.conn)

	// Always clean up server state when this handler returns:
	//   - remove client from server registry
	//   - unsubscribe from all topics
	//   - close the TCP connection
	//
	// This runs for normal disconnects and also for error paths.
	defer s.removeClient(c)

	for {
		// Decode the next request object from the TCP stream.
		// Decode blocks until:
		//   - a full JSON value arrives, or
		//   - the connection is closed, or
		//   - invalid JSON is received.
		var req clientRequest
		if err := dec.Decode(&req); err != nil {
			// io.EOF means the peer closed the connection cleanly.
			// Any other decode error is treated as "connection broken/invalid input".
			// Either way, we stop handling this client and let deferred cleanup run.
			if errors.Is(err, io.EOF) {
				return
			}
			return
		}

		// Route by action. Each case implements one part of the broker protocol.
		switch req.Action {

		case "subscribe":
			// Validate required field.
			if req.Topic == "" {
				_ = c.Send(map[string]any{"error": "topic is required"})
				continue
			}

			// GetTopic creates the topic if it doesn't exist.
			t, _ := s.GetTopic(req.Topic)

			// Record subscription *at the current sequence* to enforce:
			// "subscribers do not receive messages published before they subscribed".
			//
			// joinedSeq == current t.seq, and later snapshotSubscribers uses
			// joinedSeq < msgSeq to decide eligibility.
			t.mu.Lock()
			t.subscribers[c.id] = subscriber{client: c, joinedSeq: t.seq}
			t.mu.Unlock()

			// Track this topic membership on the client as well, so close_connection
			// can quickly remove the client from all topics without scanning all topics.
			c.subsMu.Lock()
			c.subs[req.Topic] = struct{}{}
			c.subsMu.Unlock()

			// Protocol requires an OK response on successful subscribe.
			_ = c.Send(map[string]any{"status": "ok"})

		case "unsubscribe":
			// Validate required field.
			if req.Topic == "" {
				_ = c.Send(map[string]any{"error": "topic is required"})
				continue
			}

			// Remove the client from the topic's subscriber set.
			// If the topic does not exist, the spec does not demand an error,
			// so treating it as a no-op and returning OK is acceptable.
			s.topicsMu.RLock()
			t := s.topics[req.Topic]
			s.topicsMu.RUnlock()
			if t != nil {
				t.mu.Lock()
				delete(t.subscribers, c.id)
				t.mu.Unlock()
			}

			// Also remove membership from the client's subscription tracking.
			c.subsMu.Lock()
			delete(c.subs, req.Topic)
			c.subsMu.Unlock()

			// Protocol requires an OK response on successful unsubscribe.
			_ = c.Send(map[string]any{"status": "ok"})

		case "publish":
			// Validate publish request according to the exact required error messages.
			//
			// Note: we distinguish between:
			//   - message missing entirely => "message is required"
			//   - message exists but fields missing => specific errors
			if req.Message == nil {
				_ = c.Send(map[string]any{"error": "message is required"})
				continue
			}
			if req.Message.Topic == "" {
				_ = c.Send(map[string]any{"error": "topic is required"})
				continue
			}
			if req.Message.Content == "" {
				_ = c.Send(map[string]any{"error": "message content is required"})
				continue
			}
			// Priority is a pointer so we can detect "missing" vs "present".
			// If the JSON omitted "priority", Priority will be nil.
			if req.Message.Priority == nil {
				_ = c.Send(map[string]any{"error": "priority is required"})
				continue
			}

			// Ensure the topic exists (create if missing).
			t, _ := s.GetTopic(req.Message.Topic)

			// Generate a UUID v4 for the message ID.
			// (The assignment explicitly wants UUID.)
			id, err := uuid.NewRandom()
			if err != nil {
				// Not in the spec, but returning a clear error is helpful.
				_ = c.Send(map[string]any{"error": "cannot generate message id"})
				continue
			}

			// Assign a per-topic sequence number in a critical section.
			// This sequence is used to enforce "no old messages on subscribe".
			//
			// We also push into the heap while holding the topic lock because
			// heap operations are not concurrency-safe.
			t.mu.Lock()
			t.seq++
			seq := t.seq

			msg := &queue.Message{
				ID:       id,
				Topic:    req.Message.Topic,
				Content:  req.Message.Content,
				Priority: *req.Message.Priority,
				Seq:      seq,
			}

			// Push into the per-topic heap (priority queue).
			heap.Push(t.MQ, msg)
			t.mu.Unlock()

			// Notify the dispatcher that there is work to do.
			// Non-blocking send is important: we don't want a publish request
			// to hang if the dispatcher is already awake or the buffer is full.
			select {
			case t.notify <- struct{}{}:
			default:
			}

			// Protocol requires OK response on successful publish.
			_ = c.Send(map[string]any{"status": "ok"})

		case "close_connection":
			// Client requests to close its connection.
			//
			// We acknowledge (optional but friendly), then return.
			// Returning triggers deferred cleanup (removeClient), which unsubscribes
			// from all topics and closes the connection.
			_ = c.Send(map[string]any{"status": "ok"})
			return

		case "shutdown":
			// Client requests server shutdown (graceful shutdown).
			//
			// We acknowledge and then trigger Stop().
			// Stop() may wait on goroutines; calling it in a goroutine avoids the risk
			// of deadlock if Stop waits for this handler to exit.
			_ = c.Send(map[string]any{"status": "ok"})
			go s.Stop()
			return

		default:
			// Any other action is invalid per spec.
			_ = c.Send(map[string]any{"error": "unknown action"})
		}
	}
}

// removeClient cleans up all server-side state for a client connection.
//
// This function is called when a client disconnects (EOF), sends close_connection,
// or when the server is shutting down. It ensures we don't leak memory or keep
// stale subscribers in topics.
//
// What it does (in safe order):
//  1. Remove client from the server-wide clients registry.
//  2. Read and clear the client's subscribed topic list (snapshot).
//  3. For each subscribed topic, remove the client from that topic's subscriber set.
//  4. Close the underlying TCP connection.
//
// Idempotency:
// It is safe to call more than once because deleting from maps is a no-op if
// the key doesn't exist, and closing an already-closed connection just returns an error.
// (We ignore that error here.)
func (s *Server) removeClient(c *Client) {
	// --- 1) Remove from server registry ---
	// Protect the clients map with the server-level lock because other goroutines
	// may be iterating it (GetClientConnections) or adding clients (Run).
	s.clientsMu.Lock()
	delete(s.clients, c.id)
	s.clientsMu.Unlock()

	// --- 2) Snapshot subscribed topics ---
	// We keep per-client subscription tracking so we can efficiently unsubscribe
	// the client from *only* the topics they joined, rather than scanning every topic.
	//
	// We snapshot under lock, then release the lock before touching topics.
	// This avoids holding the client lock while doing slower operations.
	c.subsMu.Lock()
	topics := make([]string, 0, len(c.subs))
	for tn := range c.subs {
		topics = append(topics, tn)
	}

	// Clear the subscription map to make repeated calls cheap and truly idempotent.
	// Also prevents double work if removeClient is triggered twice by different paths.
	c.subs = make(map[string]struct{})
	c.subsMu.Unlock()

	// --- 3) Remove from each topic ---
	// We deliberately do NOT call GetTopic here because cleanup should not create topics.
	for _, tn := range topics {
		// Lookup the topic under the server topics read lock.
		s.topicsMu.RLock()
		t := s.topics[tn]
		s.topicsMu.RUnlock()
		if t == nil {
			continue // topic may have been removed or never existed
		}

		// Remove the client from this topic's subscriber registry.
		// Write lock is needed because we're mutating the subscribers map.
		t.mu.Lock()
		delete(t.subscribers, c.id)
		t.mu.Unlock()
	}

	// --- 4) Close the TCP connection ---
	// Closing the connection unblocks any goroutine currently blocked on reading
	// from this connection (json.Decoder.Decode) and allows graceful shutdown.
	_ = c.conn.Close()
}

// Send writes a single JSON value to the client connection.
//
// Why this method is necessary:
// In this broker, multiple goroutines may need to write to the same client
// connection concurrently, for example:
//   - the per-connection handler goroutine replying {"status":"ok"} / {"error":...}
//   - the topic dispatcher goroutine delivering {"action":"deliver", ...}
//
// json.Encoder (and the underlying net.Conn writer) is NOT safe for concurrent
// use. Without synchronization, two goroutines can interleave bytes on the TCP
// stream, producing invalid JSON and causing the client-side json.Decoder to fail.
//
// This method prevents that by serializing writes with a mutex: only one Encode
// runs at a time per client.
func (c *Client) Send(v any) error {
	// Lock ensures only one goroutine writes to this connection at a time.
	c.mu.Lock()
	defer c.mu.Unlock()

	// Encode writes JSON followed by a newline. This "one JSON object per line"
	// framing works well over TCP when the peer uses json.Decoder.Decode().
	return c.enc.Encode(v)
}
