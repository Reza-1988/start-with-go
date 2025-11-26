package library

import (
	"strings"
)

type Book struct {
	title    string
	borrowed bool
	borrower string
}
type Library struct {
	books    map[string]*Book
	capacity int
}

func NewLibrary(capacity int) *Library {
	return &Library{
		books:    make(map[string]*Book),
		capacity: capacity,
	}
}

func (library *Library) AddBook(name string) string {
	rawName := strings.TrimSpace(strings.ToLower(name))
	if _, ok := library.books[rawName]; ok {
		return "The book is already in the library"
	}
	if library.capacity <= len(library.books) {
		return "Not enough capacity"
	}
	library.books[rawName] = &Book{
		title:    strings.TrimSpace(name),
		borrowed: false,
		borrower: "",
	}
	return "OK"
}

func (library *Library) BorrowBook(bookName, personName string) string {
	// TODO
	return ""
}

func (library *Library) ReturnBook(bookName string) string {
	// TODO
	return ""
}
