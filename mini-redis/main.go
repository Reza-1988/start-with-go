package main

type TTL int

const (
	TTL0 TTL = iota
	TTL1
	TTL2
	TTL3
)

type Redis interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl TTL)
	Size() int
	Clear()
	Evict()
}

type DB struct {
	// TODO
}

func NewDB(capacity int) Redis {
	db := &DB{
		// TODO
	}

	return db
}

// ...
