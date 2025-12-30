package queue

import (
	"container/heap"

	"github.com/google/uuid"
)

type IMessageQueue interface {
	heap.Interface
}

type Message struct {
	ID       uuid.UUID
	Content  string
	Priority int
	Index    int
	// other fields
}

type MessageQueue []*Message

func NewMessageQueue() IMessageQueue {
	mq := &MessageQueue{}
	heap.Init(mq)
	return mq
}
