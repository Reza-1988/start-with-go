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
				Error:  "title or author cannot be empty",
			})
			return
		}
		var input Book
		er := json.Unmarshal(body, &input)
		if er != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "title or author cannot be empty",
			})
			return
		}
		input.Title = strings.TrimSpace(input.Title)
		input.Author = strings.TrimSpace(input.Author)
		if input.Title == "" || input.Author == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "title or author cannot be empty",
			})
			return
		}
		// logic of adding books
		t := strings.ToLower(input.Title)
		a := strings.ToLower(input.Author)
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
		json.NewEncoder(w).Encode(ErrorResponse{
			Result: fmt.Sprintf("add book %s by %s", input.Title, input.Author),
			Error:  "",
		})
		return

	// GET Methode
	case http.MethodGet:
		title := strings.TrimSpace(r.URL.Query().Get("title"))
		author := strings.TrimSpace(r.URL.Query().Get("author"))
		if title == "" || author == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "title or author cannot be empty",
			})
			return
		}

		// logic of return book information
		t := strings.ToLower(title)
		a := strings.ToLower(author)
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
		return

	// PUT method
	case http.MethodPut:
		title := strings.TrimSpace(r.URL.Query().Get("title"))
		author := strings.TrimSpace(r.URL.Query().Get("author"))

		if title == "" || author == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "title or author cannot be empty",
			})
			return
		}
		// Body must be: { "borrow": true|false }
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "borrow value cannot be empty",
			})
			return
		}
		// We use *bool (a pointer to bool) instead of bool to distinguish three cases:
		// 1. borrow = true   → user wants to borrow the book
		// 2. borrow = false  → user wants to return the book
		// 3. borrow is missing from the request body → invalid request
		//
		// If we use a plain `bool`, missing "borrow" would automatically become `false`
		// (the zero-value of bool), making it impossible to know whether the user
		// really sent `false` or omitted the field entirely. Using *bool allows us to
		// detect the absence of the field because the value will be `nil` when not provided.

		var payload struct {
			Borrow *bool `json:"borrow"`
		}
		if err := json.Unmarshal(body, &payload); err != nil || payload.Borrow == nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "",
				Error:  "borrow value cannot be empty",
			})
			return
		}

		t := strings.ToLower(title)
		a := strings.ToLower(author)
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
		if *payload.Borrow {
			if b.Borrowed {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Result: "",
					Error:  "this book is already borrowed",
				})
				return
			}
			b.Borrowed = true
			s.books[key] = b
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "you have borrowed this book successfully",
				Error:  "",
			})
			return
		} else {
			if !b.Borrowed {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Result: "",
					Error:  "this book is already in the library",
				})
				return
			}
			b.Borrowed = false
			s.books[key] = b
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ErrorResponse{
				Result: "thank you for returning this book",
				Error:  "",
			})
			return
		}

	}

}
