package main

// DO NOT USE ANY IMPORT

type Qutex struct {
	ch chan struct{}
}

func NewQutex() *Qutex {
	return &Qutex{
		ch: make(chan struct{}, 1),
	}
}

func (q *Qutex) Lock() {
	q.ch <- struct{}{}
}

func (q *Qutex) Unlock() {
	select {
	case <-q.ch:
	default:
		panic("unlock of unlocked qux")
	}
}
