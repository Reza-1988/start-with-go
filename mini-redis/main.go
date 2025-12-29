package main

import (
	"fmt"
	"sync"
	"time"
)

type TTL int

const (
	TTL0 TTL = iota
	TTL1
	TTL2
	TTL3
)

// Redis defines the public API for our tiny in-memory cache.
//
// This challenge implements a compact subset of Redis-like behavior:
// - Key/value storage in memory
// - Optional TTL (time-to-live) expiration per key (in seconds)
// - Capacity limit with eviction policy when full
// - Thread-safety: methods may be called concurrently by tests
//
// Important behavioral notes inferred from tests:
//   - Get must return an error when the key does not exist or is expired.
//   - Set with ttl=0 means "no expiration" (the key lives until evicted or cleared).
//   - When capacity is full and we insert a NEW key, Evict() is triggered internally.
//     Eviction is based on "least Get() count" (least frequently accessed);
//     ties are broken by earlier insertion time.
type Redis interface {
	Get(key string) (interface{}, error)        // Get returns the stored value for the given key.
	Set(key string, value interface{}, ttl TTL) // Set stores (or updates) the value for a key with an optional TTL.
	Size() int                                  // Size returns the number of currently stored keys.
	Clear()                                     // Clear removes all keys and resets internal state.
	Evict()                                     // Evict removes exactly one key using the eviction policy.
}

// entry represents a single stored record in our in-memory Redis-like DB.
//
// We store extra metadata (beyond key/value) to support the two main features
// required by the tests:
//
// 1) TTL (Time To Live):
//   - If TTL > 0, the key must expire automatically after N seconds.
//   - We implement this by storing an absolute expiration timestamp (expireAt).
//   - If expireAt is the zero value (time.Time{}), it means "never expires" (TTL=0).
//
// 2) Eviction policy (called "LRU" in test names, but behaves like LFU by Get count):
//   - When capacity is full and we insert a NEW key, we must remove one key.
//   - We evict the key with the smallest number of successful Get() calls (gets).
//   - If multiple keys have the same gets count, we evict the one inserted earlier.
//     We implement that tie-breaker via an increasing sequence number (inserted).
type entry struct {
	// value is the user-provided value stored for the key.
	value interface{}

	// expireAt is the absolute time when this key becomes invalid.
	// - Zero value (expireAt.IsZero() == true) means "no expiration" (TTL=0).
	// - If time.Now() is after expireAt, the key is expired and Get must return an error.
	expireAt time.Time

	// gets counts how many times Get(key) successfully returned this entry.
	// This is used as the primary eviction score (smaller gets => more likely to be evicted).
	gets uint64

	// inserted is a monotonically increasing sequence assigned when the key is first inserted.
	// Used only for tie-breaking during eviction when multiple keys have equal gets.
	// Smaller inserted means "older", so it gets evicted first in a tie.
	inserted uint64
}

// DB is an in-memory, Redis-like key/value store with:
//
// - TTL support (keys can expire after N seconds)
// - A fixed capacity (max number of keys stored)
// - An eviction policy when capacity is full:
//   - Evict the key with the smallest successful Get() count (least frequently accessed)
//   - If tied, evict the key inserted earlier (oldest)
//
// Concurrency:
//   - Methods may be called from multiple goroutines (tests do concurrent Set).
//   - We use a single mutex to protect all internal state.
//     (Simple and safe: map writes, counters, and expiration checks are all synchronized.)
type DB struct {
	// mu guards all fields in DB:
	// - items map (Go maps are NOT safe for concurrent writes)
	// - seq counter (used for insertion ordering)
	// - any per-entry updates like incrementing gets or removing expired keys
	mu sync.Mutex

	// cap is the maximum number of keys allowed in memory at once.
	// When inserting a NEW key and len(items) == cap, we must evict one key first.
	//
	// Note: capacity is expected to be at least 1 (per problem statement),
	// but we can still defensively handle cap <= 0 in NewDB.
	cap int

	// seq is a monotonically increasing counter used to assign "inserted order"
	// to each new key. This provides deterministic tie-breaking during eviction.
	//
	// Example: the first inserted key might get inserted=1, next=2, etc.
	seq uint64

	// items stores the actual key -> entry mapping.
	// Each entry contains:
	// - the stored value
	// - expiration timestamp (if any)
	// - usage count (Get hits)
	// - insertion order for eviction tie-breaking
	items map[string]*entry
}

// NewDB constructs a new in-memory DB with a fixed maximum capacity.
//
// Behavior required by tests:
// - capacity controls how many keys can exist at the same time.
// - When capacity is full and a NEW key is inserted, one key must be evicted.
// - The DB must be safe for concurrent use, so we initialize internal state properly.
func NewDB(capacity int) Redis {
	// Defensive behavior (recommended):
	// - If capacity <= 0, we clamp it to 1 to avoid creating a broken cache.
	// - Capacity is guaranteed to be at least 1 by the statement,
	//   but this makes the code safer if someone passes 0 or a negative value.
	if capacity < 1 {
		capacity = 1
	}

	db := &DB{
		// Maximum number of keys allowed in memory.
		cap: capacity,

		// Sequence starts at 0.
		// We will increment it on each NEW key insertion and store that as entry.inserted.
		seq: 0,

		// Initialize the map so Set/Get can safely store/read entries.
		items: make(map[string]*entry),
	}

	return db
}

// Get returns the stored value for a key if it exists and has not expired.
//
// Returns:
// - (value, nil) if the key exists and is not expired.
// - (nil, error) if the key does not exist OR has expired.
//
// Side effects:
//   - On successful Get, increments the entry's access counter (gets),
//     which is later used for eviction decisions.
func (db *DB) Get(key string) (interface{}, error) {
	// Lock the DB because:
	// - maps are not safe for concurrent access
	// - Get updates metadata (gets counter)
	// - Get may delete expired keys
	db.mu.Lock()
	defer db.mu.Unlock()

	// Lookup key in map.
	e, ok := db.items[key]
	if !ok {
		// Key does not exist.
		return nil, fmt.Errorf("key %q not found", key)
	}

	// If the key has an expiration time and that time has passed, treat as expired.
	// Best practice: delete it so it doesn't count toward Size() or eviction.
	if !e.expireAt.IsZero() && time.Now().After(e.expireAt) {
		delete(db.items, key)
		return nil, fmt.Errorf("key %q expired", key)
	}

	// Successful Get: increase access count for eviction policy.
	e.gets++

	// Return the stored value (not the *entry struct).
	return e.value, nil
}

// Set stores (or updates) the value for a key with an optional TTL.
//
// TTL rules (from tests):
// - ttl == 0 : key never expires.
// - ttl > 0  : key expires after ttl seconds.
//
// Capacity + eviction rules (from tests):
// - If inserting a NEW key would exceed capacity, evict exactly one key first.
// - Evict the key with the smallest successful Get() count (gets).
// - If tied on gets, evict the key inserted earlier (smallest inserted sequence).
//
// Concurrency:
// - Set may be called concurrently, so we lock the DB for the whole operation.
func (db *DB) Set(key string, value interface{}, ttl TTL) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Best practice: remove expired keys first so they don't block capacity
	// and don't affect eviction decisions.
	db.cleanupExpiredLocked()

	// If key already exists and is not expired, we update it in place.
	// Updating should NOT force eviction because DB size does not increase.
	if e, ok := db.items[key]; ok {
		e.value = value
		// Update TTL:
		// - ttl==0 => remove expiration (never expires)
		// - ttl>0  => set a new expiration timestamp from now
		if ttl <= 0 {
			e.expireAt = time.Time{} // zero time -> no expiry
		} else {
			e.expireAt = time.Now().Add(time.Duration(ttl) * time.Second)
		}
		// NOTE: We keep e.gets and e.inserted unchanged on update.
		// This is the least surprising behavior and works well with the eviction policy.
		return
	}

	// This is a NEW key.
	// If we are at capacity, we must evict one key before inserting.
	if len(db.items) >= db.cap {
		db.evictOneLocked()
	}
	// Creat a new enty with required metadata
	db.seq++ // advance insertion sequence for deterministic tie-breaking.

	newEntry := &entry{
		value:    value,
		gets:     0,      // new keys start with 0 successful Get hits
		inserted: db.seq, // insertion order for tie-breaking on eviction
	}
	// Apply TTl settings.
	if ttl > 0 {
		newEntry.expireAt = time.Now().Add(time.Duration(ttl) * time.Second)
		// If ttl==0, expireAt stays zero => never expires.
	}
	db.items[key] = newEntry
}

// cleanupExpiredLocked removes expired keys from the DB.
// IMPORTANT: caller must hold db.mu.
func (db *DB) cleanupExpiredLocked() {
	now := time.Now()

	for k, e := range db.items {
		// Zero expireAT means never expires.
		if e.expireAt.IsZero() {
			continue
		}
		// if current time is after expireAt, key is expired -> remove it.
		if now.After(e.expireAt) {
			delete(db.items, k)
		}
	}
}

// evictOneLocked removes exactly one key using the eviction policy.
// IMPORTANT: caller must hold db.mu.
//
// Policy:
// - Evict smallest gets (least accessed).
// - If tied, evict smallest inserted (oldest entry).
func (db *DB) evictOneLocked() {
	// If DB is empty, nothing to evict.
	if len(db.items) == 0 {
		return
	}
	var victimKey string
	var victim *entry
	first := true

	for k, e := range db.items {
		if first {
			victimKey = k
			victim = e
			first = false
			continue
		}
		// Compare eviction priority:
		// 1) smaller gets => evict first
		// 2) if gets equal, smaller inserted (older) => evict first
		if e.gets < victim.gets || (e.gets == victim.gets && e.inserted < victim.inserted) {
			victimKey = k
			victim = e
		}
	}
	delete(db.items, victimKey)
}

// What is different between `Evict()` and `evictOneLocked()`?
// - Evict() (interface method)
// 		- This is a public method required by your Redis interface.
// 		- It must be callable by tests (including hidden tests).
// 		- It should handle locking and be safe to call concurrently.
// - `evictOneLocked()` (helper function)
// 		- This is a private/internal helper (lowercase name).
// 		- It assumes the mutex is already locked (db.mu is held).
// 		- It exists to avoid duplicating eviction logic inside both:
// 			- `Set()` (which needs to evict while already holding the lock)
// 			- `Evict()` (public method that will lock then call this)
// - So the usual pattern is:
// 		- `Set()` → lock → maybe call `evictOneLocked()` → unlock
// 		- `Evict()` → lock → call `evictOneLocked()` → unlock

// Size returns the number of currently stored (non-expired) keys.
//
// Expectations from tests:
// - Size must reflect the live keys in the DB.
// - Expired keys should not be counted.
//
// Concurrency:
//   - Size may be called while other goroutines are calling Set/Get,
//     so we lock the DB to protect the map and to safely clean up expired keys.
func (db *DB) Size() int {
	// Lock protects:
	// - concurrent access to the items map
	// - cleanupExpiredLocked() which may delete keys
	db.mu.Lock()
	defer db.mu.Unlock()

	// Remove expired keys before counting so Size() represents "active" entries only.
	db.cleanupExpiredLocked()

	// Return number of remaining keys.
	return len(db.items)
}

// Clear removes all keys from the DB and resets internal bookkeeping.
//
// Expectations:
// - After Clear(), Size() should be 0.
// - Future Get() calls should return an error until keys are Set again.
//
// What Clear should NOT do:
//   - It should NOT change the configured capacity (cap). Capacity is a DB setting,
//     not part of the stored data.
func (db *DB) Clear() {
	// Lock to prevent concurrent map access while we clear state.
	db.mu.Lock()
	defer db.mu.Unlock()

	// Fast and clean way to remove everything:
	// replace the map with a brand new empty map.
	db.items = make(map[string]*entry)

	// Reset insertion sequence counter (optional but tidy).
	// This does not affect correctness because eviction uses relative ordering,
	// and after Clear() there are no entries anyway.
	db.seq = 0
}

// Evict removes exactly one key from the DB using the eviction policy.
//
// Policy (from tests):
// - Evict the key with the lowest successful Get() count (gets).
// - If tied, evict the key inserted earlier (oldest).
//
// Concurrency:
// - Evict must be safe to call concurrently, so it takes the DB lock.
// - Internally we call evictOneLocked(), which assumes the lock is already held.
func (db *DB) Evict() {
	// Lock the DB to safely inspect and modify the items map.
	db.mu.Lock()
	defer db.mu.Unlock()

	// Best practice: remove expired keys first so we don't evict a valid key
	// while expired keys are still taking space.
	db.cleanupExpiredLocked()

	// Evict exactly one key according to the policy.
	db.evictOneLocked()
}
