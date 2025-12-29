package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSetWithoutTTL verifies the "happy path" for setting values that never expire.
//
// Expectations:
// 1) TTL=0 means "no expiration" (the key should remain until evicted or cleared).
// 2) With capacity=2, storing exactly two distinct keys should not trigger eviction.
// 3) After Set, Get should succeed for each key.
// 4) Size should report the number of currently stored (non-expired) keys.
func TestSetWithoutTTL(t *testing.T) {
	// Create a new DB with a maximum of 2 keys stored at once.
	db := NewDB(2)

	// Store the first key with TTL=0 ("keep forever").
	db.Set("key1", "value1", 0)

	// Immediately reading it should work (no error).
	_, err := db.Get("key1")
	assert.NoError(t, err)

	// Store the second key, also without expiration.
	db.Set("key2", "value2", 0)

	// Reading the second key should also work.
	_, err = db.Get("key2")
	assert.NoError(t, err)

	// Size() should reflect how many keys are currently stored.
	// Since we inserted exactly 2 keys into a capacity-2 DB, Size must be 2.
	count := db.Size()
	assert.Equal(t, count, 2, "count should be equal to 2")
}

// TestGet verifies basic read behavior:
//
// Expectations:
// 1) Getting an existing key returns its value and no error.
// 2) Getting a missing key returns an error (do NOT silently return nil / empty value).
func TestGet(t *testing.T) {
	// Create a DB with enough capacity so eviction is not involved in this test.
	db := NewDB(2)

	// Insert one key with TTL=0 so it never expires.
	db.Set("key1", "value1", 0)

	// Getting an existing key must succeed and return exactly the same value we stored.
	val, err := db.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Getting a key that was never inserted must fail with an error.
	// (Returning a zero value without error would hide bugs.)
	_, err = db.Get("key2")
	assert.Error(t, err)
}

// TestSetWithZeroTTL ensures that TTL=0 disables expiration completely.
//
// Expectations:
//  1. A key set with TTL=0 must be retrievable immediately.
//  2. After waiting (sleeping) longer than any normal TTL in this challenge,
//     the key must STILL be retrievable (no expiration should happen).
func TestSetWithZeroTTL(t *testing.T) {
	// Create a DB with capacity > 1 to avoid eviction being involved.
	db := NewDB(2)

	// Store a key with TTL=0 => "no expiry".
	db.Set("key1", "value1", 0)

	// Immediately after Set, Get must work.
	val, err := db.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Wait 3 seconds. If TTL=0 is handled incorrectly (treated like "expire now"),
	// the key would disappear and the next Get would fail.
	time.Sleep(3 * time.Second)

	// Key must still exist because TTL=0 means no expiration.
	val, err = db.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)
}

// TestSetWithTTL verifies expiration when TTL > 0.
//
// Expectations:
//  1. A key with TTL=2 should be readable before 2 seconds pass.
//  2. After waiting longer than TTL (sleep 3 seconds),
//     Get should fail because the key must be expired.
func TestSetWithTTL(t *testing.T) {
	// Capacity doesn't matter here; we only store one key.
	db := NewDB(2)

	// Store key1 with TTL=2 seconds (it should expire around now+2s).
	db.Set("key1", "value1", 2)

	// Immediately after Set, the key must exist and return the correct value.
	val, err := db.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Wait past the TTL duration.
	// Sleeping 3 seconds ensures the key has definitely expired (2 < 3).
	time.Sleep(3 * time.Second)

	// After expiration, Get must return an error (key should behave as "not found").
	_, err = db.Get("key1")
	assert.Error(t, err, "Expected an error because key1 should have expired")
}

// TestConcurrentSet verifies that Set is safe under concurrent use.
//
// Expectations:
//  1. Multiple goroutines can call Set at the same time without data races,
//     panics ("concurrent map writes"), or lost writes.
//  2. After all Set calls complete, Size() must match the number of unique keys inserted.
//  3. Every key inserted concurrently must be retrievable with the correct value.
//
// Why this matters:
// - In Go, writing to a map from multiple goroutines without synchronization will crash.
// - Production caches are commonly used concurrently, so the DB must be thread-safe.
func TestConcurrentSet(t *testing.T) {
	// Capacity is 50 so inserting exactly 50 keys should NOT trigger eviction.
	db := NewDB(50)

	// WaitGroup is used to wait until all goroutines finish their Set operation.
	var wg sync.WaitGroup

	// Number of concurrent Set operations we will run.
	concurrentOperations := 50

	// Pre-build all keys and values so goroutines only perform Set (less noise in the test).
	keys := make([]string, concurrentOperations)
	values := make([]string, concurrentOperations)

	// Generate unique key/value pairs: key0->value0, key1->value1, ...
	for i := 0; i < concurrentOperations; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
		values[i] = fmt.Sprintf("value%d", i)
	}

	// Spawn 50 goroutines. Each goroutine sets one unique key.
	for i := 0; i < concurrentOperations; i++ {
		// Tell the WaitGroup: "we have 1 more goroutine to wait for".
		// This increments the internal counter by 1.
		wg.Add(1)
		go func(index int) {
			// Mark this goroutine as "finished" when the function ends.
			// Done() decrements the counter by 1.
			// Using defer guarantees it runs even if the goroutine exits early.
			defer wg.Done()

			// TTL=0 => no expiration.
			db.Set(keys[index], values[index], 0)
		}(i)
	}

	// Wait for all concurrent Set operations to complete.
	wg.Wait()

	// After all sets, the DB should contain exactly 50 keys.
	assert.Equal(t, concurrentOperations, db.Size(), "There should be 50 entries")

	// Verify every inserted key exists and has the correct value.
	for i := 0; i < concurrentOperations; i++ {
		value, err := db.Get(keys[i])
		assert.NoError(t, err)
		assert.Equal(t, values[i], value, fmt.Sprintf("Mismatched value for %s", keys[i]))
	}
}

// TestConcurrentSetWithTTL verifies thread-safety of Set combined with correct TTL expiration.
//
// Expectations:
//  1. Many goroutines can Set keys concurrently without race conditions.
//  2. Keys set with TTL=2 seconds must expire after 2 seconds.
//  3. After sleeping 3 seconds (longer than TTL), all keys must be gone / expired,
//     so Get should return an error for every key.
//
// Why this matters:
// - Real caches are used concurrently.
// - TTL logic must be correct even when keys are created in parallel.
func TestConcurrentSetWithTTL(t *testing.T) {
	// Capacity matches the number of keys, so eviction should not happen.
	db := NewDB(50)

	var wg sync.WaitGroup

	// Number of concurrent inserts.
	concurrentOperations := 50

	// Arrays to store predictable test data.
	keys := make([]string, concurrentOperations)
	values := make([]string, concurrentOperations)
	ttls := make([]TTL, concurrentOperations)

	// Prepare key/value pairs and set all TTLs to 2 seconds.
	for i := 0; i < concurrentOperations; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
		values[i] = fmt.Sprintf("value%d", i)
		ttls[i] = 2
	}

	// Start 50 goroutines. Each goroutine sets one unique key with TTL=2.
	for i := 0; i < concurrentOperations; i++ {
		// Add 1 to WaitGroup so wg.Wait() knows to wait for this goroutine.
		wg.Add(1)

		go func(index int) {
			// Ensure we always signal completion, even if the goroutine exits early.
			defer wg.Done()

			// Set key with TTL=2 seconds.
			db.Set(keys[index], values[index], ttls[index])
		}(i)
	}

	// Wait until all Set operations have finished.
	wg.Wait()

	// Wait longer than TTL so all keys should definitely be expired.
	time.Sleep(3 * time.Second)

	// Every key should now be expired, so Get must return an error for each one.
	for i := 0; i < concurrentOperations; i++ {
		_, err := db.Get(keys[i])
		assert.Error(t, err, fmt.Sprintf("Expected an error because %s should have expired", keys[i]))
	}
}

// TestLRUEviction verifies eviction behavior when capacity is exceeded.
//
// Important detail:
// Despite the name, eviction in these tests is NOT "LRU by time of last access".
// It behaves like: evict the key with the LOWEST Get() count (least frequently accessed).
// If Get() counts are equal, evict the key that was inserted earlier (oldest entry).
//
// Expectations in this test:
//   - Capacity = 2.
//   - Insert key1 and key2 (cache is now full).
//   - Insert key3 => must evict exactly one key.
//   - Since none of the keys were Get()'d before eviction, all have GetCount = 0,
//     so we break ties by insertion time => key1 is the oldest and must be evicted.
func TestLRUEviction(t *testing.T) {
	// Create a DB that can store at most 2 keys at a time.
	db := NewDB(2)

	// Fill the DB to capacity with two keys.
	db.Set("key1", "value1", 0)
	db.Set("key2", "value2", 0)

	// Add a third key. This exceeds capacity, so the DB must evict one key.
	db.Set("key3", "value3", 0)

	// key1 should be evicted (oldest among keys with equal access counts).
	_, err := db.Get("key1")
	assert.Error(t, err)

	// key2 should still exist.
	val, err := db.Get("key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)

	// key3 (the newly inserted key) should also exist.
	val, err = db.Get("key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val)
}

// TestEvictionOrder verifies eviction chooses the least-accessed key when capacity is full.
//
// Rule being tested:
// - When inserting a new key into a full DB, evict the key with the smallest Get() count.
// - This behaves like LFU (Least Frequently Used) based on successful Get calls.
//
// Expectations in this test:
// - Capacity=3, insert 3 keys => full.
// - Access key1 and key2 once each => they become "more used" than key3.
// - Insert key4 => must evict key3 because it has the lowest Get count (0).
func TestEvictionOrder(t *testing.T) {
	// DB can store at most 3 keys.
	db := NewDB(3)

	// Fill the DB to capacity.
	db.Set("key1", "value1", 0)
	db.Set("key2", "value2", 0)
	db.Set("key3", "value3", 0)

	// Access key1 and key2 once:
	// - key1 GetCount becomes 1
	// - key2 GetCount becomes 1
	// - key3 stays at 0 (least used)
	db.Get("key1")
	db.Get("key2")

	// Insert a 4th key. Since capacity is full, DB must evict one existing key.
	// The least-used key is key3 (0 gets), so key3 should be evicted.
	db.Set("key4", "value4", 0)

	// key3 should have been evicted.
	_, err := db.Get("key3")
	assert.Error(t, err, "key3 should to be evicted")

	// key1 must still exist.
	val, err := db.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// key2 must still exist.
	val, err = db.Get("key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)

	// key4 (newly inserted) must exist.
	val, err = db.Get("key4")
	assert.NoError(t, err)
	assert.Equal(t, "value4", val)
}

// TestEvictionOrder2 verifies eviction tie-breaking when multiple keys have the same Get count.
//
// Rule being tested:
// 1) Evict the key with the smallest Get() count (least frequently accessed).
// 2) If multiple keys share the same Get count, evict the one inserted earlier (oldest).
//
// Expectations in this test:
// - key1 is accessed once => GetCount=1.
// - key2 and key3 are never accessed => GetCount=0 for both.
// - When inserting key4 into a full cache, we must evict one key.
// - Between key2 and key3 (tie at 0), key2 is older => key2 must be evicted.
func TestEvictionOrder2(t *testing.T) {
	// DB capacity is 3.
	db := NewDB(3)

	// Fill the DB to capacity.
	db.Set("key1", "value1", 0)
	db.Set("key2", "value2", 0)
	db.Set("key3", "value3", 0)

	// Access only key1 once:
	// - key1 GetCount becomes 1
	// - key2 GetCount remains 0
	// - key3 GetCount remains 0
	db.Get("key1")

	// Insert a new key; capacity is full, so one key must be evicted.
	// The least-used keys are key2 and key3 (both 0).
	// Tie-breaker: evict the older inserted key => key2.
	db.Set("key4", "value4", 0)

	// key2 should be evicted due to tie-breaking by insertion order.
	_, err := db.Get("key2")
	assert.Error(t, err, "key2 should be evicted")

	// key1 should remain (it was accessed, so it is more "valuable").
	val, err := db.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// key3 should remain (same usage as key2 but inserted later, so it survives).
	val, err = db.Get("key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val)

	// key4 (newly inserted) should remain.
	val, err = db.Get("key4")
	assert.NoError(t, err)
	assert.Equal(t, "value4", val)
}

// TestEvictionAfterMultipleAccess verifies that multiple successful Get() calls
// increase a key's "usage score", making it less likely to be evicted.
//
// Rule being tested:
// - When capacity is exceeded, evict the key with the smallest Get() count.
// - Here, key1 is accessed twice, so key2 becomes the least-used and must be evicted.
//
// Note:
// The error message mentions "least recently used", but this test actually matches
// "least frequently accessed" based on Get counts.
func TestEvictionAfterMultipleAccess(t *testing.T) {
	// DB capacity is 2 keys.
	db := NewDB(2)

	// Fill the DB.
	db.Set("key1", "value1", 0)
	db.Set("key2", "value2", 0)

	// Access key1 twice:
	// - key1 GetCount becomes 2
	// - key2 GetCount stays 0
	db.Get("key1")
	db.Get("key1")

	// Insert a third key; cache is full, so one key must be evicted.
	// key2 has the smallest GetCount (0), so it should be evicted.
	db.Set("key3", "value3", 0)

	// key2 should be evicted because it is the least-used key.
	_, err := db.Get("key2")
	assert.Error(t, err, "key2 should be evicted due to being least recently used")

	// key1 should remain because it was accessed multiple times.
	val, err := db.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// key3 should exist because it was just inserted.
	val, err = db.Get("key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val)
}

// TestEvictionAfterMultipleAccess2 verifies tie-breaking when usage (Get count) is equal.
//
// Rule being tested:
// 1) Evict the key with the smallest number of successful Get() calls.
// 2) If multiple keys have the same Get() count, evict the one inserted earlier (older key).
//
// Expectations in this test:
// - key1 and key2 are both accessed twice => GetCount=2 for both.
// - Inserting key3 into a full DB forces eviction.
// - Because usage is tied, the older key (key1) must be evicted.
func TestEvictionAfterMultipleAccess2(t *testing.T) {
	// DB capacity is 2 keys.
	db := NewDB(2)

	// Fill the DB.
	db.Set("key1", "value1", 0)
	db.Set("key2", "value2", 0)

	// Access both keys the same number of times:
	// - key1 GetCount becomes 2
	// - key2 GetCount becomes 2
	db.Get("key1")
	db.Get("key1")
	db.Get("key2")
	db.Get("key2")

	// Insert a third key; capacity is full so we must evict one existing key.
	// Since usage is tied, we evict the oldest inserted key: key1.
	db.Set("key3", "value3", 0)

	// key1 should be evicted because it was inserted earlier than key2.
	_, err := db.Get("key1")
	assert.Error(t, err, "key1 should be evicted due to being added sooner")

	// key2 should remain (same usage as key1, but inserted later so it survives).
	val, err := db.Get("key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)

	// key3 should exist because it was just inserted.
	val, err = db.Get("key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val)
}

// TestEvictionAfterMultipleAccess3 checks eviction with several keys and different access counts.
//
// Rule being tested:
// - When capacity is exceeded, evict the key with the smallest Get() count (least accessed).
// - This is effectively LFU behavior using "number of successful Get calls" as the score.
// - Tie-breaking by insertion order is not needed here because one key is clearly least used.
//
// Expectations in this test:
// - W is never accessed => GetCount=0.
// - Other keys have at least 1 access.
// - When inserting V into a full cache, W must be evicted.
func TestEvictionAfterMultipleAccess3(t *testing.T) {
	// DB capacity is 4.
	db := NewDB(4)

	// Fill the DB to capacity.
	db.Set("X", "Ex", 0)
	db.Set("Y", "Why", 0)
	db.Set("Z", "Zed", 0)
	db.Set("W", "Double-U", 0)

	// Access pattern:
	// - X accessed twice => GetCount(X)=2
	// - Z accessed once  => GetCount(Z)=1
	// - Y accessed once  => GetCount(Y)=1
	// - W accessed zero  => GetCount(W)=0 (least used)
	db.Get("X")
	db.Get("Z")
	db.Get("X")
	db.Get("Y")

	// Insert a 5th key into capacity=4 DB => must evict one key.
	// W has the lowest GetCount (0), so it should be evicted.
	db.Set("V", "Five", 0)

	// W should be gone.
	_, err := db.Get("W")
	assert.Error(t, err, "W should to be evicted")

	// X should remain.
	val, err := db.Get("X")
	assert.NoError(t, err)
	assert.Equal(t, "Ex", val, "X should remain")

	// Z should remain.
	val, err = db.Get("Z")
	assert.NoError(t, err)
	assert.Equal(t, "Zed", val, "Z should remain")

	// Y should remain.
	val, err = db.Get("Y")
	assert.NoError(t, err)
	assert.Equal(t, "Why", val, "Y should remain")

	// V (newly inserted) should remain.
	val, err = db.Get("V")
	assert.NoError(t, err)
	assert.Equal(t, "Five", val, "V should remain")
}

// TestEvictionWithTTL checks eviction and TTL expiration together.
//
// Scenario:
//   - Capacity=2.
//   - Insert key1 and key2 with TTL=2 seconds => DB becomes full.
//   - Insert key3 with TTL=0 => DB must evict ONE key immediately.
//     At this moment, key1 and key2 both have GetCount=0, so tie-breaker is insertion order:
//     => key1 is older, so key1 should be evicted.
//   - After sleeping 3 seconds, any TTL=2 key that remains should now be expired.
//     => key2 should expire.
//     => key3 should remain because TTL=0 means "never expire".
func TestEvictionWithTTL(t *testing.T) {
	// DB capacity is 2 keys.
	db := NewDB(2)

	// Insert two keys that will expire after 2 seconds.
	db.Set("key1", "value1", 2)
	db.Set("key2", "value2", 2)

	// Insert a third key with no expiration.
	// This exceeds capacity, so DB must evict one key NOW (before any TTL has expired).
	// Both key1 and key2 have the same usage (0 gets), so evict the older one: key1.
	db.Set("key3", "value3", 0)

	// Wait long enough for TTL=2 keys to definitely expire.
	time.Sleep(3 * time.Second)

	// key1 should be missing because it was evicted at the moment we inserted key3.
	// (Even though it *would* also be expired after 3 seconds, eviction already removed it.)
	_, err := db.Get("key1")
	assert.Error(t, err, "Expected an error because key1 should have been evicted")

	// key2 was not evicted, so it stayed in the DB until TTL expiration.
	// After 3 seconds, TTL=2 has passed => key2 must be expired and return an error.
	_, err = db.Get("key2")
	assert.Error(t, err, "Expected an error because key2 should have expired")

	// key3 has TTL=0 ("never expires"), so it must still be retrievable.
	val, err := db.Get("key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val)
}

// TestMultipleKeysWithTTL verifies that each key has its own independent TTL
// and that keys expire according to the TTL value in seconds.
//
// Expectations:
// - key1 (TTL=1) expires after ~1 second.
// - key2 (TTL=2) expires after ~2 seconds.
// - key3 (TTL=3) expires after ~3 seconds.
// - After a key expires, Get must return an error (behave like "not found").
//
// This test also ensures that expiration is not "global" and not tied to insertion order;
// each key must track its own expiresAt timestamp.
func TestMultipleKeysWithTTL(t *testing.T) {
	// Capacity is 3, and we store exactly 3 keys, so eviction should never happen here.
	db := NewDB(3)

	// Set keys with different TTL durations (in seconds).
	db.Set("key1", "value1", 1) // expires after 1s
	db.Set("key2", "value2", 2) // expires after 2s
	db.Set("key3", "value3", 3) // expires after 3s

	// After 1 second, key1 should be expired.
	time.Sleep(1 * time.Second)
	_, err := db.Get("key1")
	assert.Error(t, err, "Expected an error because key1 should have expired")

	// After another 1 second (2 seconds total), key2 should be expired.
	time.Sleep(1 * time.Second)
	_, err = db.Get("key2")
	assert.Error(t, err, "Expected an error because key2 should have expired")

	// After another 1 second (3 seconds total), key3 should be expired.
	time.Sleep(1 * time.Second)
	_, err = db.Get("key3")
	assert.Error(t, err, "Expected an error because key3 should have expired")
}
