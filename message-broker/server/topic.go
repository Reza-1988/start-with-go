package server

import (
	"QueraMQ/queue"
	"sync"
)

// Topic represents a named channel in the broker.
//
// Each topic owns its own message queue (a priority heap) and will later own
// its own subscriber set. Producers publish messages to a topic, and consumers
// subscribe to receive messages for that topic.
//
// Concurrency note:
// Topic is accessed by multiple goroutines (one per client connection, plus
// possible dispatcher goroutines). The underlying heap implementation is NOT
// thread-safe, so we protect all accesses to MQ (and later subscriber state)
// using mu.
type Topic struct {
	// Name is the unique identifier of the topic (e.g., "chat_room_1").
	Name string

	// MQ is the per-topic priority queue used to order messages by Priority.
	// It is stored as an interface so the rest of the code depends on the
	// abstraction (heap.Interface) rather than the concrete queue type.
	MQ queue.IMessageQueue

	// mu protects all mutable state inside Topic (MQ and, later, subscribers).
	// Use RLock/RUnlock for read-only access and Lock/Unlock for mutations.
	// Never hold this lock while performing network I/O (writes to client connections).
	mu sync.RWMutex
}

// NewTopic creates a new Topic with its own independent priority message queue.
//
// Each topic maintains a separate queue so messages are ordered and managed per topic.
// The returned Topic is ready to be used by the server (publish/subscribe logic will
// add more fields like subscriber sets in later steps).
func NewTopic(name string) *Topic {
	return &Topic{
		// Name is the identifier clients use in subscribe/publish requests.
		Name: name,

		// MQ is a heap-initialized priority queue (container/heap).
		// Note: the queue itself is not concurrency-safe; Topic.mu must guard access.
		MQ: queue.NewMessageQueue(),
	}
}

// GetMessageQueue returns the underlying concrete *queue.MessageQueue used by this Topic.
//
// Why this exists:
//   - The assignment/test harness explicitly requires this exact signature:
//     `func (t *Topic) GetMessageQueue() *queue.MessageQueue`
//   - so grading/tests can call it to inspect the queue inside a Topic.
//
// Why we need a type assertion:
//   - Topic stores the queue behind an interface:
//     `MQ queue.IMessageQueue`
//   - An interface value may hold different concrete implementations.
//   - But this method must return the concrete type *queue.MessageQueue, so we “unwrap” the interface
//     back to that concrete type using a type assertion.
//
// Safety / behavior:
//   - The “ok” result is a guard in case MQ is not actually a *queue.MessageQueue.
//   - In that case we return nil instead of panicking.
//   - In this project it should normally succeed because NewTopic() initializes MQ via
//
// queue.NewMessageQueue(), which returns *queue.MessageQueue.
func (t *Topic) GetMessageQueue() *queue.MessageQueue {
	mq, ok := t.MQ.(*queue.MessageQueue)
	if !ok {
		return nil
	}
	return mq
}

// One-sentence mental model
// 	- `t.MQ` is a “box” (interface) holding the real queue.
//	- `GetMessageQueue()` opens the box and returns the real queue type.
