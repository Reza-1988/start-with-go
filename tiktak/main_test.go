package main

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
)

//Any function that starts with Test and takes the input *testing.T is recognized by go test as a test.
//This function is located in a file like main_test.go.
//So if you want to write your own test, the basic format is always:
// ```
// func TestSomething(t *testing.T) {
//    // ...
//  }

func TestTikTakSample(t *testing.T) {
	// Table-driven tests:
	// Each item contains an input (n) and the expected output string.
	tests := []struct {
		n      int
		output string
	}{
		{n: 1, output: "TikTak"},
		{n: 2, output: "TikTakTikTak"},
	}

	for _, test := range tests {
		// Reset global counters used by SayTik() and SayTak().
		// This ensures each test starts from a clean state.
		tikCount = 0
		takCount = 0

		// -----------------------------
		// 1) Redirect stdout to capture printed output
		// -----------------------------

		// Save the original stdout so we can restore it later.
		originalStdout := os.Stdout

		// Create a pipe: anything written to 'w' can be read from 'r'.
		r, w, _ := os.Pipe()

		// Redirect stdout so all prints go into the pipe instead of the terminal.
		os.Stdout = w

		// -----------------------------
		// 2) Create the TikTak instance
		// -----------------------------
		tiktak := NewTikTak(test.n)

		// -----------------------------
		// 3) Run Tik() and Tak() concurrently
		// -----------------------------

		var wg sync.WaitGroup
		wg.Add(2) // We will wait for two goroutines: Tik and Tak.

		// Start Tik() in its own goroutine.
		go func() {
			defer wg.Done() // Mark this goroutine as finished when returning.
			tiktak.Tik()
		}()

		// Start Tak() in its own goroutine.
		go func() {
			defer wg.Done()
			tiktak.Tak()
		}()

		// Wait for both goroutines to complete.
		wg.Wait()

		// -----------------------------
		// 4) Stop capturing stdout
		// -----------------------------

		// Close the writer end of the pipe.
		// This signals "no more output is coming".
		w.Close()

		// Restore original stdout, so future prints go to terminal again.
		os.Stdout = originalStdout

		// -----------------------------
		// 5) Read captured output into a buffer
		// -----------------------------
		var buf bytes.Buffer

		// Copy everything that was printed (now inside r) into the buffer.
		io.Copy(&buf, r)

		// Convert buffer to a normal string.
		output := buf.String()

		// -----------------------------
		// 6) Compare actual output with expected output
		// -----------------------------
		if output != test.output {
			t.Errorf("For n=%d, expected output %q, but got %q", test.n, test.output, output)
		}
	}
}
