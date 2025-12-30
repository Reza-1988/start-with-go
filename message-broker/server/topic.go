package server

import (
	"QueraMQ/queue"
)

type Topic struct {
	Name string
	MQ   queue.IMessageQueue
	// other fields
}

func NewTopic(name string) *Topic {
	return &Topic{
		Name: name,
		MQ:   queue.NewMessageQueue(),
	}
}

func (t *Topic) GetMessageQueue() *queue.MessageQueue
