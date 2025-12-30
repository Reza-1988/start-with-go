package server

import (
	"QueraMQ/queue"
	"sync"
)

// subscriber holds a client and the sequence number at which they subscribed.
// A client should only receive messages with Seq > joinedSeq, so they do not
// receive messages published before they subscribed.
type subscriber struct {
	client    *Client
	joinedSeq int64
}

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

	// subscribers are the currently subscribed clients for this topic.
	// Key: client ID (RemoteAddr string).
	subscribers map[string]subscriber

	// seq is a monotonically increasing counter for messages in this topic.
	// It is used to ensure "no old messages on subscribe".
	seq int64

	// notify wakes the topic dispatcher when new messages arrive.
	// Buffered to avoid blocking publishers.
	notify chan struct{}
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
		MQ:          queue.NewMessageQueue(),
		subscribers: make(map[string]subscriber),
		notify:      make(chan struct{}, 1),
	}
}

// What exactly is notify?
//		- Its type is `chan struct{}`, meaning it is a channel that does not carry any data; it only "signals".
// 		- `struct{}` is very lightweight and common for signaling because it is zero bytes.
// Why is it there?
// 		- You usually have a goroutine (dispatcher) for each topic whose job is to:
// 			- Wait for new messages to arrive
// 			- When they arrive, take the messages from MQ and deliver them to subscribers
// 			- If there is no notify, the dispatcher should either:
// 				- Continuously loop and check if the MQ is empty (busy-wait → burns CPU)
// 				- Or work with more complex sleep/poll methods
// 			- notify makes the dispatcher sleep and wake up only when a new message arrives.
// 	Why is it “Buffered”? What is the advantage?
// 		- Suppose the publisher publishes a message and wants to wake up the dispatcher:
// 		`t.notify <- struct{}`
// 			- If the channel was unbuffered (`make(chan struct{})`):
// 				- If the dispatcher is not ready to receive at this moment, the publisher will block and get stuck.
// 			- But since it has a buffer of 1:
// 				- If the buffer is empty, the signal goes into the buffer and the publisher continues immediately.
// 				- If the buffer is full (i.e. there is already a signal and the dispatcher has not read it yet),
//				  retransmission is not necessary, and it is better not to block.
// 			- Usually the correct pattern is:
// 			 ```select {
//				case t.notify <- struct{}:
//				default:
//				}```
//			- That is:
// 				- If there is room, send a signal
// 				- If there is no room (the buffer is full), don't worry; because "this one signal" is enough to wake up the dispatcher.
// How does the dispatcher use it?
// 		- A common dispatcher pattern is:
// 		 ```for {
//				select {
//				case <-t.notify:
//					// new message arrived, go read from MQ and send
//				case <-stopCh:
//					return
//				}
//			}
//		- This way:
//			- the dispatcher does not consume CPU until a message arrives
// 			- when a message arrives, it wakes up and drains all existing messages

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
