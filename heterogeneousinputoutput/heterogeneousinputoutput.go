package main

import (
	"sync/atomic"
	"time"
)

type FutureResult struct {
	Done       atomic.Bool
	ResultChan chan string
}

// Task defines a semantic type for a unit of work that runs asynchronously.
// Any function that takes no input and returns a string can be used as a Task.
// This makes the code more readable and allows consistent handling of async jobs.
type Task func() string

// Async executes the provided Task asynchronously in a separate goroutine.
// It returns a FutureResult that can be used to retrieve the Task's output later.
//
// The Task is executed concurrently, and once it finishes:
//   - The resulting string is sent to the ResultChan channel.
//   - The Done atomic flag is set to true, indicating completion.
//
// This allows non-blocking execution of long-running tasks while still
// providing a simple way to access the result once available.
func Async(t Task) *FutureResult {
	// Create a new FutureResult with a buffered channel (capacity 1)
	// to hold the result of the asynchronous Task.
	fr := &FutureResult{ResultChan: make(chan string, 1)}

	// Run the Task in a separate goroutine.
	go func() {
		// Execute the Task and send its result to the channel.
		res := t()
		fr.ResultChan <- res

		// Mark the Task as completed using the atomic flag.
		fr.Done.Store(true)
	}()

	// Return the FutureResult to allow the caller to access the result later.
	return fr
}

// AsyncWithTimeout executes the given Task asynchronously in a separate goroutine,
// but limits the maximum waiting time for its result.
//
// It returns a FutureResult that contains a ResultChan, which will receive either:
//   - the Task's actual result (if it finishes before the timeout), or
//   - the string "timeout" (if the specified duration expires first).
//
// The Done atomic flag is only set to true when the Task completes successfully;
// if the timeout occurs first, Done remains false.
//
// Internally, this function starts two goroutines:
//  1. One runs the Task in the background and sends its result to a temporary channel.
//  2. Another uses a select statement to wait for either:
//     - a message from that temporary channel, or
//     - a signal from time.After(timeout), which sends a value on its own channel after the duration.
//
// This design allows non-blocking concurrent execution with built-in timeout control.
func AsyncWithTimeout(t Task, timeout time.Duration) *FutureResult {
	fr := &FutureResult{ResultChan: make(chan string, 1)}
	tempCh := make(chan string, 1)

	// Run the Task asynchronously and send its result to a temporary channel.
	go func() {
		tempCh <- t()
	}()

	// Wait for either the Task result or the timeout signal.
	go func() {
		select {
		case res := <-tempCh:
			fr.ResultChan <- res // Task finished before timeout
			fr.Done.Store(true)  // Mark task as done
		case <-time.After(timeout):
			fr.ResultChan <- "timeout" // Timeout occurred
			// Done remains false in this case
		}
	}()

	return fr
}

// Await blocks the caller until the Task finishes and a value
// is sent into the ResultChan. It then returns that result.
//
// This works for both Async and AsyncWithTimeout —
// in the timeout case, the returned value will be "timeout".
func (fResult *FutureResult) Await() string {
	// Wait until something is received from ResultChan
	// (either the actual result or "timeout")
	result := <-fResult.ResultChan
	return result
}

// CombineFutureResults takes multiple FutureResult objects (each created by Async or AsyncWithTimeout)
// and combines their results into a single FutureResult.
//
// It creates a new FutureResult with a buffered ResultChan large enough to hold one result
// for each of the input FutureResults.
//
// A new goroutine is started that sequentially reads one value from each input's ResultChan
// and forwards it into the combined ResultChan.
//
// The order of results in the combined channel matches the order of the input FutureResults,
// regardless of which tasks finish first.
//
// Note: This function does not start any new tasks; it only merges the results of already-running ones.
func CombineFutureResults(fResults ...*FutureResult) *FutureResult {
	combined := &FutureResult{ResultChan: make(chan string, len(fResults))}

	go func() {
		for _, fr := range fResults {
			res := <-fr.ResultChan
			combined.ResultChan <- res
		}
	}()
	return combined
}
