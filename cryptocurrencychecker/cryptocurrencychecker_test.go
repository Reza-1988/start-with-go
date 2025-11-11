package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//
// ──────────────────────────────────────────────────────────────────────────────
//  MOCK SERVER
// ──────────────────────────────────────────────────────────────────────────────
//
// startServer spins up a tiny in-memory HTTP server that mimics the assignment's API.
// - It reads exchange data from a local JSON file (data.json)
// - It exposes the endpoint: /rates?srcCurrency=<src>&dstCurrency=<dst>
// - It returns a JSON payload with the shape:
//     {
//       "status": "OK",
//       "stats": { "btc-rls": { "latest": "28650.385" } }
//     }
// Notes:
//   • This server is ONLY for tests. The production solution should call the real API.
//   • We run it in a goroutine so the test can continue executing concurrently.
//   • It binds to port :4001; make sure this port is free when running tests.
//
func startServer() {
	// Minimal JSON shapes matching the mock API’s contract:
	type stats struct {
		Latest string `json:"latest"` // the numeric price as a string
	}
	type information struct {
		Status string           `json:"status"` // expected to be "OK" for success
		Stats  map[string]stats `json:"stats"`  // e.g., "btc-rls" -> { latest: "..." }
	}

	// datalist holds many entries keyed by "<src>-<dst>" (e.g., "btc-rls", "btc-usdt", ...)
	// The file data.json must decode into: map[string]information
	datalist := make(map[string]information)

	// Open the dataset file that drives server responses.
	// TIP: data.json must be located in the SAME directory where you run `go test`.
	f, err := os.Open("data.json")
	if err != nil {
		// We panic here because the mock server cannot function without data.
		panic(err)
	}
	defer f.Close()

	// Decode the JSON file into our in-memory datastore (datalist).
	if err := json.NewDecoder(f).Decode(&datalist); err != nil {
		panic(err)
	}

	// Register a single handler for /rates that validates query params and responds accordingly.
	http.HandleFunc("/rates", func(w http.ResponseWriter, r *http.Request) {
		// Extract query parameters (expected lowercase in this mock).
		src := r.URL.Query().Get("srcCurrency")
		dst := r.URL.Query().Get("dstCurrency")

		// Build the composite key (e.g., "btc-rls") to look up the rate.
		key := fmt.Sprintf("%s-%s", src, dst)

		// If we have a matching record, return 200 + JSON.
		if info, ok := datalist[key]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(info)
			return
		}

		// Otherwise signal a client error (bad request).
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid Requests"))
	})

	// Blocking call: this will keep serving until the process exits.
	// In tests we call it inside a goroutine so the test can continue.
	log.Fatal(http.ListenAndServe(":4001", nil))
}

//
// ──────────────────────────────────────────────────────────────────────────────
//  TEST-ONLY RESPONSE TYPES
// ──────────────────────────────────────────────────────────────────────────────
//
// These types are used only inside the test to decode the mock server’s response
// (so we can assert against what the mock server returns).
//
type sampleTestExchangeResponse struct {
	Status string                    `json:"status"`
	Stats  map[string]sampleTestStat `json:"stats"`
}

type sampleTestStat struct {
	Latest string `json:"latest"`
}

//
// ──────────────────────────────────────────────────────────────────────────────
//  TEST CASE
// ──────────────────────────────────────────────────────────────────────────────
//
// TestSampleGetBTCtoRials verifies that:
//  1) Our solution function GetExchangeRate("BTC", "") correctly defaults dst="rls"
//  2) It hits the right URL and parses the JSON shape properly
//  3) The returned string equals the mock API’s "latest" field for key "btc-rls"
//
// Why assert against the mock server’s own response?
// - This makes the test robust against data changes in data.json
// - We’re not hard-coding the expected price; we assert “function result == server truth”
//
// Concurrency notes:
// - The mock server runs in a goroutine
// - We sleep briefly to allow the server to start listening before making the first request
//   (In production you might coordinate with sync primitives instead of Sleep.)
//
func TestSampleGetBTCtoRials(t *testing.T) {
	// 1) Boot the mock server in the background.
	go startServer()

	// 2) Give the server a moment to bind to :4001 (quick and dirty; enough for tests).
	time.Sleep(30 * time.Millisecond)

	// 3) Call the function under test: it should default destination to "rls" when empty.
	result, err := GetExchangeRate("BTC", "")
	if err != nil {
		t.Fatalf("Expected no error from GetExchangeRate, got: %v", err)
	}

	// 4) Independently call the same mock endpoint to obtain the ground-truth JSON.
	resp, err := http.DefaultClient.Get("http://localhost:4001/rates?srcCurrency=btc&dstCurrency=rls")
	if err != nil {
		t.Fatalf("Expected no error requesting mock server, got: %v", err)
	}
	defer resp.Body.Close()

	// 5) Read and decode the mock server’s JSON payload.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed reading mock response body: %v", err)
	}

	var serverPayload sampleTestExchangeResponse
	if err := json.Unmarshal(body, &serverPayload); err != nil {
		t.Fatalf("Failed decoding mock response JSON: %v", err)
	}

	// 6) Extract the expected "latest" value for key "btc-rls".
	key := "btc-rls"
	expected := serverPayload.Stats[key].Latest

	// 7) Assertion:
	//    The function’s result must match the mock server’s "latest" field.
	assert.Equal(t, expected, result,
		"GetExchangeRate result must equal server's 'latest' for %s", key)
}
