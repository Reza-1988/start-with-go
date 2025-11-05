package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	port = "4001"
	path = "http://localhost:4001"
)

var serverSingleton *Server

// getServer ensures that only ONE instance of the server_httpclient is created and running.
// This function follows the Singleton pattern:
//
// Why we need this:
// ------------------
//   - In Go tests, each test can run independently.
//   - If every test tried to start the server_httpclient on the same port, we would get:
//     "address already in use"
//   - By checking if `serverSingleton` is nil, we guarantee:
//     → NewServer() + Start() is executed ONLY ONCE during all tests.
//
// How it works:
// -------------
//   - The first time getServer() is called, serverSingleton == nil:
//     → create the server_httpclient
//     → start it in a goroutine so the test thread does not block
//   - Subsequent calls simply return the already-running server_httpclient.
func getServer() *Server {
	if serverSingleton == nil {
		serverSingleton = NewServer(port) // create the server_httpclient only once

		// Start server_httpclient in background goroutine
		// This allows the test execution to continue without blocking.
		go serverSingleton.Start()

		// A small delay to give the server_httpclient time to start listening on the port.
		// If we remove this delay, tests may try to connect before the server_httpclient is ready.
		time.Sleep(1000 * time.Millisecond)
	}

	return serverSingleton // always return the same running instance
}

func TestServerCreation(t *testing.T) {
	s := getServer()
	assert.NotNil(t, s)
}

type responseForm struct {
	Result string
	Error  string
}

func addBook(title, author string, duplicate bool) error {

	// Build the request body as FORM data (not JSON).
	// The test uses: http.DefaultClient.PostForm(...)
	// So the server_httpclient must read values using r.ParseForm() / r.FormValue("title")
	data := url.Values{
		"title":  []string{title},
		"author": []string{author},
	}

	// Send POST /book request.
	// PostForm automatically encodes data using:
	//     Content-Type: application/x-www-form-urlencoded
	resp, err := http.DefaultClient.PostForm(path+"/book", data)
	if err != nil {
		// Network / connection failure (server_httpclient not running?)
		return err
	}
	defer resp.Body.Close()

	// In this test suite, POST should always return 200,
	// even when the book already exists in the library.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not OK: %d", resp.StatusCode)
	}

	// This struct matches: {"Result": "...", "Error": "..."}
	var rf responseForm

	// Decode server_httpclient JSON response into struct.
	// If the JSON cannot be unmarshalled, something is wrong on server_httpclient-side.
	if err := json.NewDecoder(resp.Body).Decode(&rf); err != nil {
		return err
	}

	// When duplicate == false
	// We expect the newly added book message:
	//     "added book <title> by <author>"
	// Test compares values in lowercase, so we lowercase title and author.
	expectedAddMsg := fmt.Sprintf(
		"added book %s by %s",
		strings.ToLower(title),
		strings.ToLower(author),
	)

	// If this is NOT duplicate, server_httpclient *must* send the "added book..." message.
	if !duplicate && rf.Result != expectedAddMsg {
		return fmt.Errorf("unexpected result: %s", rf.Result)
	}

	// When duplicate == true, it means the test EXPECTS the book to already exist.
	// In that case, the correct server_httpclient response should be:
	//     "this book is already in the library"
	//
	// Important:
	// ----------
	// Returning an error here does NOT mean something went wrong!
	// In the context of the test, returning an error is how we SIGNAL
	// that the duplicate-case behavior occurred as expected.
	// The test calling this function will check for a non-nil error and PASS.
	//
	// In short: error here = expected result for duplicate add attempt.
	if duplicate && rf.Result == "this book is already in the library" {
		return errors.New(rf.Result)
	}

	// No errors → request was valid and behavior matched expectations.
	return nil
}

func TestAddBook(t *testing.T) {
	err := addBook("alice in wonderland", "JC", false)
	assert.Nil(t, err)
}

type testBook struct {
	Title  string
	Author string
}

func getBook(title, author string) (testBook, error) {
	// Encode query parameters to make them URL-safe.
	// e.g., "Alice in Wonderland" => "Alice%20in%20Wonderland"
	// Prevents breaking the URL when there are spaces or special characters (&, ?, =, /, ...).
	title = url.QueryEscape(title)
	author = url.QueryEscape(author)

	// Build and send GET /book?title=...&author=...
	// The server_httpclient is expected to read title/author from the query string.
	resp, err := http.DefaultClient.Get(fmt.Sprintf("%s/book?title=%s&author=%s", path, title, author))
	if err != nil {
		// Network/connection error (server_httpclient not reachable, DNS, etc.)
		return testBook{}, err
	}
	defer resp.Body.Close()

	// Contract: on success, server_httpclient must return 200 OK.
	// Any other status means the request did not meet the requirements
	// (e.g., missing params, book not found, or borrowed).
	if resp.StatusCode != http.StatusOK {
		return testBook{}, fmt.Errorf("invalid status code %d", resp.StatusCode)
	}

	// Read the response body fully.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return testBook{}, err
	}

	// Expected success payload shape: {"title":"...", "author":"..."}
	var result testBook
	if err := json.Unmarshal(body, &result); err != nil {
		// Malformed JSON or unexpected response schema from server_httpclient.
		return testBook{}, err
	}

	// Return the parsed book info to the test.
	return result, nil
}

func TestGetBook(t *testing.T) {
	book, err := getBook("alice in wonderland", "JC")
	assert.Nil(t, err)
	assert.Equal(t, "alice in wonderland", book.Title)
	assert.Equal(t, "jc", book.Author)
}
