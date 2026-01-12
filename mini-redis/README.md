# Mini Redis (TTL + Eviction) — Go

A small, in-memory, Redis-inspired key/value store implemented in Go.
Built to satisfy Quera’s TDD-style tests: **TTL expiration**, **capacity limit**, and **deterministic eviction** under **concurrent access**.

---

## Features

### 1) Key/Value Storage

Store arbitrary values (`interface{}`) under string keys.

### 2) TTL (Time To Live)

Each key may have a TTL in **seconds**:

* `ttl == 0` → **no expiration** (the key stays until evicted or cleared)
* `ttl > 0`  → key **expires after `ttl` seconds**

Expired keys behave as **not found**:

* `Get(key)` returns an error
* expired keys are cleaned up so they don’t count toward `Size()` and don’t block capacity

> Implementation uses an absolute timestamp (`expireAt`) rather than counting seconds.

### 3) Capacity Limit + Eviction Policy

The DB has a fixed capacity (minimum 1). When inserting a **new key** while full, exactly **one** key is evicted.

Eviction policy (as inferred from tests):

1. Evict the key with the **lowest number of successful `Get` calls** (least “Get خورده”)
2. If tied, evict the key that was **inserted earlier** (older entry)

> Although some tests name it “LRU”, the behavior matches **LFU-by-Get-count** with an insertion-order tie-breaker.

### 4) Concurrency-Safe

All public operations are safe under concurrent usage (tests run concurrent `Set`).
A single `sync.Mutex` guards internal state, ensuring correctness and preventing “concurrent map writes”.

---

## Public API

```go
type Redis interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, ttl TTL)
    Size() int
    Clear()
    Evict()
}
```

### Behavior Summary

#### `Set(key, value, ttl)`

* Inserts or updates a key.
* If inserting a **new** key and capacity is full → evicts one key first.
* Applies TTL rules (0 = never expire, >0 = expire after N seconds).

#### `Get(key)`

* Returns `(value, nil)` if present and not expired.
* Returns `(nil, error)` if missing or expired.
* On success, increments the key’s **access counter** used for eviction.

#### `Size()`

* Returns the number of **currently active** (non-expired) keys.

#### `Clear()`

* Removes all keys.
* Does **not** change capacity.

#### `Evict()`

* Removes exactly one key using the eviction policy.
* Cleans expired keys first so eviction doesn’t wastefully remove a valid key.

---

## Implementation Notes

### Data Model

Each key maps to an `entry`:

* `value` — stored value
* `expireAt` — expiration timestamp (zero value means no expiry)
* `gets` — successful `Get` count (eviction score)
* `inserted` — monotonic insertion sequence number (tie-breaker)

The DB stores:

* `items map[string]*entry`
* `seq uint64` for insertion ordering
* `cap int` for capacity
* `mu sync.Mutex` for thread safety

### Expiration Strategy

This implementation performs **lazy/clean-up deletion**:

* expired keys are removed when detected (e.g., during `Get`, `Size`, `Set`, `Evict`)
* keeps logic simple and deterministic for tests

### Complexity

* `Get`: **O(1)** average (plus constant-time expiration check)
* `Set`: **O(1)** average, but eviction requires scanning keys → **O(n)** in worst case
* `Evict`: **O(n)** (scans all keys to find the best eviction candidate)
* `Size`: **O(n)** if cleanup removes expired keys

Given small capacities in the tests, an O(n) eviction scan is practical and keeps the code easy to reason about.

---

## Running Tests Locally

```bash
go test ./...
```

---

## Example Usage

```go
db := NewDB(2)

db.Set("a", "hello", 0) // never expires
db.Set("b", "world", 2) // expires in 2 seconds

v, _ := db.Get("a") // "hello"

time.Sleep(3 * time.Second)
_, err := db.Get("b") // err != nil (expired)

db.Set("c", "!", 0) // capacity eviction may happen depending on current size
```

---

## Notes / Guarantees

* TTL unit is **seconds**
* TTL=0 means **no expiration**
* Eviction is deterministic: **lowest `Get` count**, then **oldest insertion**
* Safe for concurrent access using a single mutex

