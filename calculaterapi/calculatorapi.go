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
	Port string
}

type Respond struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

func NewServer(port string) *Server {
	srv := &Server{
		Port: port,
	}
	return srv
}

func (s *Server) handleAdd(w http.ResponseWriter, r *http.Request) {
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

	var res int64 = 0
	for _, p := range parts {
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Respond{
		Result: fmt.Sprintf("The result of your query is: %d", res),
		Error:  "",
	})
}

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

	if len(parts) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Result: "",
			Error:  "'numbers' parameter missing",
		})
		return
	}

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

func (s *Server) Start() {
	http.HandleFunc("/add", s.handleAdd)
	http.HandleFunc("/sub", s.handleSub)
	if err := http.ListenAndServe(":"+s.Port, nil); err != nil {
		fmt.Println("server error:", err)
	}

}
