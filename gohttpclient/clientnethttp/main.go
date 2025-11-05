package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {

	// Create an HTTP client.
	// This gives us full control over timeouts, headers, cookies, etc.
	client := &http.Client{}

	// Target server endpoint
	url := "http://localhost:8080/sayHelloWorld"

	// Build an HTTP GET request.
	// Unlike http.Get(), NewRequest allows adding headers or body if needed.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Send the request using the client.
	// This does: open connection → send request → wait for response.
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	// Always close response body to avoid memory leaks.
	defer resp.Body.Close()

	// Validate success response.
	// A non-200 status means server returned an error.
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Request failed with status code: %d\n", resp.StatusCode)
		return
	}

	// Read the full body from response (raw bytes)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Print the server response as a string
	fmt.Printf("Response Body:\n%s\n", string(body))
}

/*
======================================================
 Summary of this approach (using net/http standard lib)
======================================================

This method is LOW-LEVEL / manual:

Full control over request (headers, cookies, timeout, body, etc.)
No external dependencies (pure Go standard library)
❌ More boilerplate code:
   - Create request manually (NewRequest)
   - Send it (client.Do)
   - Read the response (io.ReadAll)
   - Decode JSON manually (if needed)

Compared to `go-http-client` (which handles JSON + error wrapping automatically),
here **you must manually manage every step**.

Think of this as:
   Manual Car (more control, more responsibility)
*/