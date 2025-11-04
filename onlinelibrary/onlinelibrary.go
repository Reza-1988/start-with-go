package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Server struct {
	Port  string
	books map[string]Book
}

type ErrorResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

type BookResponse struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

type Book struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Borrowed bool   `json:"-"`
}

func NewServer(port string) *Server {
	srv := &Server{
		Port:  port,
		books: make(map[string]Book),
	}
	return srv
}

func (s *Server) Start() {
	http.HandleFunc("/book", s.handleLibrary)
	if err := http.ListenAndServe(":"+s.Port, nil); err != nil {
		fmt.Println("server error:", err)
	}

}

func (s *Server) handleLibrary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	// POST Methode
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "Error reading request body",
			})
			return
		}
		var input Book
		er := json.Unmarshal(body, &input)
		if er != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "Error the format is incorrect",
			})
			return
		}
		if input.Title == "" || input.Author == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "title or author cannot be empty",
			})
			return
		}
		// logic of adding books
		t := strings.ToLower(strings.TrimSpace(input.Title))
		a := strings.ToLower(strings.TrimSpace(input.Author))
		key := t + "|" + a

		if _, ok := s.books[key]; ok {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "this book is already in the library",
			})
			return
		}
		// Add new book (keep user initials)
		s.books[key] = Book{
			Title:    input.Title,
			Author:   input.Author,
			Borrowed: false,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BookResponse{
			Title:  input.Title,
			Author: input.Author,
		})
		return
	// GET Methode
	case http.MethodGet:
		tQuery := r.URL.Query().Get("title")
		aQuery := r.URL.Query().Get("author")
		if (strings.TrimSpace(tQuery) == "") || (strings.TrimSpace(aQuery) == "") {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "title or author cannot be empty",
			})
			return
		}

		// logic of return book information
		t := strings.ToLower(strings.TrimSpace(tQuery))
		a := strings.ToLower(strings.TrimSpace(aQuery))
		key := t + "|" + a
		b, ok := s.books[key]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "this book does not exist",
			})
			return
		}
		if b.Borrowed {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "this book is borrowed",
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BookResponse{
			Title:  b.Title,
			Author: b.Author,
		})
	}

}
