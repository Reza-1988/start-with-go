package queue

import (
	"container/heap"

	"github.com/google/uuid"
)

// IMessageQueue is the abstraction used by the server/topic layer to work with a
// priority queue without depending on a concrete implementation.
//
// It embeds heap.Interface, which means any type that implements the standard
// heap contract (Len, Less, Swap, Push, Pop) can be used as an IMessageQueue.
//
// In this project, MessageQueue (a slice of *Message) is expected to implement
// heap.Interface, and NewMessageQueue returns it through this interface.
//
// Notes (best practice / gotchas):
//   - heap.Interface's Push/Pop take/return `any` (interface{}), so callers must
//     type-assert to the expected element type (here: *Message).
//   - heap.Interface is NOT concurrency-safe. Protect all heap operations with
//     a mutex at a higher level (e.g., inside Topic) when used concurrently.
type IMessageQueue interface {
	heap.Interface
}

// Message represents a single unit of data published by a producer.
//
// Messages are stored inside a per-topic priority queue (heap). Lower numeric
// Priority means "more important" and should be delivered sooner (e.g. 1 before 3).
//
// Index is required by container/heap: it tracks the message's current position
// inside the heap so the heap implementation can maintain ordering efficiently.
// (It is updated by the heap during Push/Swap/Pop and should not be treated as
// business-level metadata.)
type Message struct {
	// ID is a unique identifier for this message. The project expects UUID v4
	// (random) IDs; this is typically assigned at publish time.
	ID uuid.UUID

	// Content is the message payload to be delivered to consumers.
	Content string

	// Priority controls ordering in the heap. Smaller values are delivered first.
	Priority int

	// Index is the current heap index of this message (managed by heap.Interface).
	// Conventionally set to -1 when the item is removed from the heap.
	Index int

	// Topic is the logical channel this message belongs to. While the Topic has its
	// own queue, including this field is useful because the delivery JSON payload
	// must include the topic name.
	//
	// This is an "allowed" extra field per the assignment ("other fields").
	Topic string

	// Seq is an optional tie-breaker for stable ordering when multiple messages
	// have the same Priority. If you set Seq as a monotonically increasing number
	// at publish time, you can ensure FIFO behavior among equal-priority messages.
	//
	// Recommended for determinism, especially in concurrent tests.
	Seq int64
}

// MessageQueue is a priority queue of *Message items backed by a slice.
//
// It is designed to work with Go's container/heap package by implementing
// heap.Interface (Len, Less, Swap, Push, Pop).
//
// Key points / best practices:
//   - The underlying type is a slice so heap operations can efficiently reorder
//     elements in-place.
//   - The heap package mutates the slice (reorders it) to maintain the heap
//     invariant, so you should not assume the slice is kept in insertion order.
//   - MessageQueue itself is NOT concurrency-safe; protect it with a mutex at a
//     higher level (e.g., inside Topic) when accessed from multiple goroutines.
//   - Elements are pointers (*Message) to avoid copying message data during swaps.
type MessageQueue []*Message

func (mq MessageQueue) Len() int {
	return len(mq)
}

// Less defines which element is "higher priority" in heap terms.
// We want smaller Priority to be popped first.
// If priorities match, use Seq so equal-priority messages behave FIFO.
func (mq MessageQueue) Less(i, j int) bool {
	if mq[i].Priority != mq[j].Priority {
		return mq[i].Priority < mq[j].Priority
	}
	return mq[i].Seq < mq[j].Seq
}

// Swap must swap items and also update their Index fields.
func (mq MessageQueue) Swap(i, j int) {
	mq[i], mq[j] = mq[j], mq[i]
	mq[i].Index = i
	mq[j].Index = j
}

// Push adds an item to the heap.
// Must be pointer receiver because it modifies the slice length.
func (mq *MessageQueue) Push(x any) {
	msg := x.(*Message)
	msg.Index = len(*mq)
	*mq = append(*mq, msg)
}

// Pop removes and returns the last element (heap will ensure it's the min).
// Must be pointer receiver because it modifies the slice length.
func (mq *MessageQueue) Pop() any {
	old := *mq
	n := len(old)
	msg := old[n-1]
	msg.Index = -1 // Convention: mark as removed
	*mq = old[:n-1]
	return msg
}

// NewMessageQueue constructs an empty priority queue and initializes it as a heap.
//
// Why heap.Init is called here:
//   - container/heap expects the provided value to satisfy heap.Interface.
//   - heap.Init arranges (or "heapifies") the underlying slice so it satisfies
//     the heap invariants required for subsequent heap.Push / heap.Pop calls.
//   - For an empty slice, heap.Init doesn't move anything, but calling it is still
//     good practice because it makes initialization explicit and correct even if
//     you later change this function to start with preloaded items.
//
// Why we pass `mq` (a *MessageQueue) to heap.Init:
//   - heap.Init takes a heap.Interface.
//   - heap.Interface includes Push and Pop methods which *must* have pointer
//     receivers in most heap implementations (because they modify the slice length).
//   - Therefore, the heap is usually operated on through a pointer to the slice
//     (e.g., *MessageQueue), not the slice value itself.
//
// Why we can return `mq` after calling heap.Init(mq):
//   - `mq` is a pointer to a MessageQueue slice.
//   - heap.Init does not replace `mq`; it mutates the underlying slice data (and
//     may reorder elements) through the methods you implemented.
//   - After heap.Init returns, `mq` still points to the same queue—now guaranteed
//     to be in a valid heap state.
//
// Note: container/heap does not provide any synchronization. If multiple goroutines
// will access this queue, protect heap operations with a mutex at a higher layer.
func NewMessageQueue() IMessageQueue {
	// Create an empty queue value and take its address so heap can call Push/Pop
	// (which typically need to modify the slice length).
	mq := &MessageQueue{}

	// Initialize the heap invariants. Safe and recommended even if mq is empty.
	heap.Init(mq)

	// Return as the interface type expected by the rest of the project.
	return mq
}
