package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// apiResponse represents the exact structure of the JSON returned by the API.
// Example JSON:
//
//	{
//	  "status": "OK",
//	  "stats": { "btc-rls": { "latest": "28650.385" } }
//	}
type apiResponse struct {
	Status string `json:"status"`
	Stats  map[string]struct {
		Latest string `json:"latest"` // Holds the latest exchange rate value
	} `json:"stats"`
}

// GetExchangeRate sends a request to the API and returns the exchange rate between two currencies.
//
// Parameters:
//
//	source      = currency we want to convert from   (e.g., "BTC")
//	destination = currency we want to convert to     (e.g., "USDT")
//
// Returns:
//
//	string = latest rate price returned by server
//	error  = nil on success / error explaining what failed
func GetExchangeRate(source, destination string) (string, error) {

	// Normalize inputs: remove spaces + convert to lowercase
	src := strings.ToLower(strings.TrimSpace(source))
	dst := strings.ToLower(strings.TrimSpace(destination))

	// Source currency must not be empty
	if src == "" {
		return "", fmt.Errorf("source currency cannot be empty")
	}

	// If destination is empty, default to Rials (as required by instructions)
	if dst == "" {
		dst = "rls"
	}

	// Build API URL dynamically based on parameters
	url := fmt.Sprintf("http://localhost:4001/rates?srcCurrency=%s&dstCurrency=%s", src, dst)

	// Send GET request using Go's built-in HTTP client
	resp, err := http.Get(url)
	if err != nil {
		// Network-level errors (connection refused, timeout, ...)
		return "", err
	}
	defer resp.Body.Close() // Prevent connection leak

	// Validate API HTTP status (checks server-side errors)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status: %s", resp.Status)
	}

	// Parse JSON response into struct (apiResponse)
	var data apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err // JSON parsing error
	}

	// API returns a "status" field we should validate
	if strings.ToUpper(data.Status) != "OK" {
		return "", fmt.Errorf("api status: %s", data.Status)
	}

	// Build key format "src-dst" (e.g., "btc-usdt") to access map
	key := fmt.Sprintf("%s-%s", src, dst)

	// Lookup the rate in the Stats map
	stat, ok := data.Stats[key]
	if !ok || stat.Latest == "" {
		return "", fmt.Errorf("rate not found for %s", key)
	}

	// Success — return the exchange rate value
	return stat.Latest, nil
}
