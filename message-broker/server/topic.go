package server

import (
	"QueraMQ/queue"
	"container/heap"
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

// topicLoop is the per-topic "dispatcher" goroutine.
//
// Why we need this loop (big picture):
//   - Producers publish messages into a topic's priority queue (heap).
//   - Consumers subscribed to the topic should receive messages immediately.
//   - Messages must be delivered in PRIORITY order (higher priority first).
//
// Instead of delivering directly inside the publish handler (which would mix
// concerns and risk blocking on slow clients), we:
//  1. Push messages into the topic heap quickly.
//  2. Notify this loop via t.notify.
//  3. This loop pops messages in priority order and broadcasts them.
//
// Shutdown behavior:
//   - When s.stopCh is closed (Stop/shutdown), this loop exits gracefully.
func (s *Server) topicLoop(t *Topic) {
	for {
		select {
		// t.notify is a "wake up" signal: at least one new message was published.
		// It does not carry the message itself; it only tells us to check the heap.
		case <-t.notify:

			// Drain all available messages from the heap.
			//
			// We drain in a loop because:
			//   - Multiple messages can be published before we wake up.
			//   - t.notify may be buffered / coalesced (we might receive only one signal
			//     for many publishes), so we must empty the heap until it's truly empty.
			for {
				// popMessage pops the highest-priority message (smallest Priority number)
				// from the topic heap. It returns nil when the heap is empty.
				msg := s.popMessage(t)
				if msg == nil {
					break // nothing left to deliver right now
				}

				// snapshotSubscribers returns a slice of clients who should receive this
				// message, based on "subscribe after publish should not receive old messages".
				//
				// Important best practice: we take a snapshot while holding locks, then
				// release locks before doing network I/O. Network writes can block, and
				// holding locks during I/O can deadlock or stall the entire server.
				subs := s.snapshotSubscribers(t, msg.Seq)

				// Broadcast the message to all eligible subscribers.
				//
				// We ignore send errors here because delivery is "best effort":
				// if a client disconnected or is broken, Send will fail and the client
				// cleanup path (handleClient/removeClient) should eventually remove it.
				for _, c := range subs {
					_ = c.Send(map[string]any{
						"action": "deliver",
						"message": map[string]any{
							// message_id must be a string in the JSON protocol.
							"message_id": msg.ID.String(),
							"topic":      msg.Topic,
							"content":    msg.Content,
							"priority":   msg.Priority,
						},
					})
				}
			}

		// stopCh is closed during server shutdown. Selecting on it lets the loop
		// exit cleanly without leaking goroutines.
		case <-s.stopCh:
			return
		}
	}
}

// popMessage removes and returns the next message to be delivered from a topic.
//
// Why this exists:
//   - The topic's message queue is implemented as a heap (priority queue).
//   - To deliver messages in priority order, we must "pop" from the heap.
//   - heap.Pop returns the highest-priority element according to the heap's Less()
//     method (in this project: smaller Priority value is higher priority).
//
// Concurrency:
//   - container/heap is NOT safe for concurrent use.
//   - Therefore we must hold the topic lock while checking length and popping.
//
// Returns:
//   - The next *queue.Message if available.
//   - nil if the heap is empty (no message to deliver).
func (s *Server) popMessage(t *Topic) *queue.Message {
	// We take an exclusive lock because heap.Pop mutates the heap/slice.
	t.mu.Lock()
	defer t.mu.Unlock()

	// Fast empty check to avoid calling heap.Pop on an empty heap.
	if t.MQ.Len() == 0 {
		return nil
	}

	// heap.Pop removes and returns the top element of the heap.
	// We type-assert because heap.Interface uses `any` (interface{}) for Pop.
	return heap.Pop(t.MQ).(*queue.Message)
}

// snapshotSubscribers returns a point-in-time list of clients that should
// receive a specific message.
//
// Why this function exists:
// The assignment requires that when a client subscribes to a topic, it must NOT
// receive messages that were published before the subscription.
//
// We enforce that rule by tracking, for each subscriber, the topic sequence value
// at the moment they subscribed (joinedSeq). Each published message gets its own
// sequence number (msgSeq). A subscriber is eligible if:
//
//	joinedSeq < msgSeq
//
// Meaning: "this subscriber joined before this message was published".
//
// Concurrency and best practice:
//   - We use RLock because we only need to read the subscribers map.
//   - We return a *snapshot* slice so the caller can release the lock before
//     performing network I/O (sending messages), which might block.
func (s *Server) snapshotSubscribers(t *Topic, msgSeq int64) []*Client {
	// Read lock: safe concurrent access to the subscribers map.
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Preallocate output slice based on current subscriber count.
	// This is efficient and avoids repeated reallocations.
	out := make([]*Client, 0, len(t.subscribers))

	// Iterate over current subscribers and select only those who are eligible
	// for this message according to the "no old messages on subscribe" rule.
	for _, sub := range t.subscribers {
		if sub.joinedSeq < msgSeq {
			out = append(out, sub.client)
		}
	}

	// The caller can now send to these clients without holding t.mu.
	return out
}
