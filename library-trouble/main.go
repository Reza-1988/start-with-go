package library

import (
	"fmt"
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
	if len(library.books) >= library.capacity {
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
	bookName = strings.TrimSpace(strings.ToLower(bookName))
	b, ok := library.books[bookName]
	if !ok {
		return "The book is not defined in the library"
	}
	if b.borrowed {
		return fmt.Sprintf("The book is already borrowed by %s", b.borrower)
	}
	b.borrowed = true
	b.borrower = strings.TrimSpace(personName)
	return "OK"
}

func (library *Library) ReturnBook(bookName string) string {
	bookName = strings.TrimSpace(strings.ToLower(bookName))
	b, ok := library.books[bookName]
	if !ok {
		return "The book is not defined in the library"
	}
	if !b.borrowed {
		return "The book has not been borrowed"
	}
	b.borrowed = false
	b.borrower = ""
	return "OK"
}
