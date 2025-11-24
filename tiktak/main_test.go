package main

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
)

func TestTikTakSample(t *testing.T) {
	tests := []struct {
		n      int
		output string
	}{
		{n: 1, output: "TikTak"},
		{n: 2, output: "TikTakTikTak"},
	}

	for _, test := range tests {
		tikCount = 0
		takCount = 0

		// Capture the output
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		tiktak := NewTikTak(test.n)

		var wg sync.WaitGroup
		wg.Add(2)

		// Start the goroutines
		go func() {
			defer wg.Done()
			tiktak.Tik()
		}()

		go func() {
			defer wg.Done()
			tiktak.Tak()
		}()

		wg.Wait()

		// Close the write end of the pipe
		w.Close()

		os.Stdout = originalStdout
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if output != test.output {
			t.Errorf("For n=%d, expected output %q, but got %q", test.n, test.output, output)
		}
	}
}
