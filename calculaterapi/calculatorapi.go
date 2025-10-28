package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	Port string
}

type Response struct {
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
	numberString := r.URL.Query().Get("numbers")
	if numberString == "" {
		json.NewEncoder(w).Encode(Response{
			Result: "",
			Error:  "'numbers' parameter missing",
		})
		return
	}
	parts := strings.Split(numberString, ",")

	var res int64 = 0
	for _, p := range parts {
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Error: "Invalid number format"})
			return
		}
		res += n
	}
	json.NewEncoder(w).Encode(Response{
		Result: fmt.Sprintf("The result of your query is: %d", res),
		Error:  "",
	})
}

func (s *Server) Start() {
	http.HandleFunc("/add", s.handleAdd)
	http.HandleFunc("/sub", s.handleSub)
	http.ListenAndServe(":"+s.Port, nil)

}
