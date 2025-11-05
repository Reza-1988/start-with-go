package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	Port string // TCP port to listen on, e.g. "4001"
}

// JSON response schema required by the exercise
type Respond struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

// Constructor: create a Server instance with the chosen port
func NewServer(port string) *Server {
	srv := &Server{
		Port: port, // use the provided port (don't hardcode)
	}
	return srv
}

// /add handler: sums all numbers from the query (?numbers=10,20,...)
func (s *Server) handleAdd(w http.ResponseWriter, r *http.Request) {
	// Always declare response mime-type for clients
	w.Header().Set("Content-Type", "application/json")

	// Read the query param ?numbers=...
	numberString := r.URL.Query().Get("numbers")
	// Per spec: if missing or empty, return 400 with the exact error message
	if strings.TrimSpace(numberString) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Result: "",
			Error:  "'numbers' parameter missing",
		})
		return
	}

	// Split by comma: "10, 20,30" => ["10"," 20","30"]
	parts := strings.Split(numberString, ",")

	var res int64 = 0 // additive identity (start from 0)
	for _, p := range parts {
		// Be lenient with spaces: " 20 " -> "20"
		p = strings.TrimSpace(p)
		// Empty tokens are considered missing parameter (keeps tests happy)
		if p == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Result: "",
				Error:  "'numbers' parameter missing",
			})
			return
		}

		// Convert token to int64; base 10; store as int64 (exercise limits)
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			// Spec only defines two error messages; reuse the “missing” one
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Result: "",
				Error:  "'numbers' parameter missing",
			})
			return
		}

		// Prevent overflow BEFORE doing res += n
		// If adding n would exceed int64 bounds, bail out with Overflow
		if (n > 0 && res > math.MaxInt64-n) || (n < 0 && res < math.MinInt64-n) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Result: "",
				Error:  "Overflow",
			})
			return
		}
		res += n
	}

	// Success: 200 with the exact phrasing required
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Respond{
		Result: fmt.Sprintf("The result of your query is: %d", res),
		Error:  "",
	})
}

// /sub handler: subtracts subsequent numbers from the first one
// e.g. ?numbers=104,204 => 104 - 204 = -100
func (s *Server) handleSub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	numberString := r.URL.Query().Get("numbers")
	if strings.TrimSpace(numberString) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Result: "",
			Error:  "'numbers' parameter missing",
		})
		return
	}

	parts := strings.Split(numberString, ",")

	// Must have at least one number to seed the subtraction
	if len(parts) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Result: "",
			Error:  "'numbers' parameter missing",
		})
		return
	}

	// Seed result with the first number (subtraction identity strategy)
	first := strings.TrimSpace(parts[0])
	res, err := strconv.ParseInt(first, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Result: "",
			Error:  "'numbers' parameter missing",
		})
		return
	}

	// Subtract all remaining numbers: res = res - n
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if p == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Result: "",
				Error:  "'numbers' parameter missing",
			})
			return
		}
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Result: "",
				Error:  "'numbers' parameter missing",
			})
			return
		}

		// Prevent overflow BEFORE doing res -= n
		// For subtraction, the safe-range checks differ from addition:
		// If n > 0  and res < MinInt64 + n  => res - n would underflow
		// If n < 0  and res > MaxInt64 + n  => res - n would overflow
		if (n > 0 && res < math.MinInt64+n) || (n < 0 && res > math.MaxInt64+n) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Result: "",
				Error:  "Overflow",
			})
			return
		}
		res -= n
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Respond{
		Result: fmt.Sprintf("The result of your query is: %d", res),
		Error:  "",
	})
}

// Server bootstrap: register handlers and start listening
func (s *Server) Start() {
	http.HandleFunc("/add", s.handleAdd) // GET /add?numbers=...
	http.HandleFunc("/sub", s.handleSub) // GET /sub?numbers=...

	// Start the HTTP server_httpclient; log if it fails (e.g., port already in use)
	if err := http.ListenAndServe(":"+s.Port, nil); err != nil {
		fmt.Println("server_httpclient error:", err)
	}
}

// If you want to run the API manually without tests:
// Run: go run main.go
// Then test with browser / curl / Postman / Insomnia
// Example:
//
//	http://localhost:5000/add?numbers=10,20
//	http://localhost:5000/sub?numbers=100,50
func main() {
	ser := NewServer("5000")
	ser.Start()
}
