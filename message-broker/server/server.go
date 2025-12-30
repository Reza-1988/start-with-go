package server

import (
	"encoding/json"
	"net"
	"sync"
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

func (s *Server) Run() error {

}

func (s *Server) Stop()

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
