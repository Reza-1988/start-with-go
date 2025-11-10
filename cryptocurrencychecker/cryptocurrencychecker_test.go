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

// startServer starts a mock HTTP server used ONLY during tests.
// This server simulates the API specified in the assignment.
func startServer() {

	// Represents the inner JSON structure: { "latest": "1234.56" }
	type stats struct {
		Latest string `json:"latest"`
	}

	// Represents one valid API response structure
	// {
	//   "status": "OK",
	//   "stats": { "btc-rls": { "latest":"28650.385" } }
	// }
	type information struct {
		Status string           `json:"status"`
		Stats  map[string]stats `json:"stats"`
	}

	// datalist will hold all valid exchange rates loaded from data.json
	var datalist = make(map[string]information)

	// Load exchange rates from local file data.json
	f, err := os.Open("data.json")
	if err != nil {
		// If the file doesn't exist, the test cannot proceed
		panic(err)
	}
	defer f.Close()

	// Decode JSON file into map[string]information
	err = json.NewDecoder(f).Decode(&datalist)
	if err != nil {
		panic(err)
	}

	// Setup endpoint: /rates?srcCurrency=btc&dstCurrency=rls
	http.HandleFunc("/rates", func(w http.ResponseWriter, r *http.Request) {
		src := r.URL.Query().Get("srcCurrency")
		dst := r.URL.Query().Get("dstCurrency")

		// Construct key like "btc-rls" and check if we have data for it
		if info, ok := datalist[fmt.Sprintf("%s-%s", src, dst)]; ok {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(info)
			return
		}

		// If key was not found → return HTTP 400
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid Requests"))
	})

	// Start server (blocking call) - so use goroutine in tests
	log.Fatal(http.ListenAndServe(":4001", nil))
}

// Struct used only inside tests for validation
type sampleTestExchangeResponse struct {
	Status string                    `json:"status"`
	Stats  map[string]sampleTestStat `json:"stats"`
}

// Represents { "latest": "value" } inside test JSON
type sampleTestStat struct {
	Latest string `json:"latest"`
}

// TEST CASE:
// Ensures that GetExchangeRate("BTC", "") returns correct RLS (rial) price.
func TestSampleGetBTCtoRials(t *testing.T) {

	// Start the fake API server in background
	go startServer()

	// Wait a bit so server starts before requesting
	time.Sleep(30 * time.Millisecond)

	// Call function we are testing
	result, err := GetExchangeRate("BTC", "")
	if err != nil {
		t.Errorf("Expected no errors, but got: %s", err.Error())
	}

	// Make the *same* request manually to compare result
	resp, err := http.DefaultClient.Get("http://localhost:4001/rates?srcCurrency=btc&dstCurrency=rls")
	if err != nil {
		t.Errorf("Expected no errors, but got: %s", err.Error())
	}
	defer resp.Body.Close()

	// Read response body
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading server response: %s", err.Error())
	}

	// Decode server JSON into struct
	var response sampleTestExchangeResponse
	err = json.Unmarshal(b, &response)
	if err != nil {
		t.Errorf("JSON decoding failed: %s", err.Error())
	}

	// Extract expected value
	expected := response.Stats["btc-rls"].Latest

	// ✅ Compare function output with expected server response
	assert.Equal(t, expected, result)
}
