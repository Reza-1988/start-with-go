package main

import (
	"context"
	"fmt"
	"log"

	gohttpclient "github.com/bozd4g/go-http-client"
)

type Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func main() {

	// Create a new HTTP client_httpclient with a base URL.
	// This library simplifies making HTTP requests compared to net/http.
	client := gohttpclient.New("http://localhost:8080/sayHelloWorld")

	// Send a GET request to the base URL.
	// Context allows cancellation or timeout handling (if needed).
	response, err := client.Get(context.Background(), "")
	f

	// Declare a struct to store the JSON response body.
	var res Response

	// Automatically reads the HTTP body and unmarshal JSON into struct.
	// This replaces manual io.ReadAll + json.Unmarshal in the standard net/http package.
	if err := response.Unmarshal(&res); err != nil {
		log.Fatalf("error decoding JSON response: %v", err)
	}

	// Print the parsed response struct.
	// "%+v" prints field names and values.
	fmt.Printf("%+v\n", res)
}

/*
==============================================================
Difference between this approach and net/http (standard library)
==============================================================

Using go-http-client_httpclient:
---------------------
Less boilerplate code
Automatic JSON unmarshalling (response.Unmarshal(&struct))
Cleaner API for GET / POST / PUT / DELETE
🚫 Requires an external dependency (third-party library)

Example:
	response, err := client_httpclient.Get(ctx, "")
	response.Unmarshal(&res)

Using net/http (standard Go library):
-------------------------------------
No external dependencies (built-in library)
Full low-level control over the HTTP request
🚫 More manual work:
	- Build request manually
	- Send request via client_httpclient.Do()
	- Read body manually (io.ReadAll)
	- Decode JSON manually (json.Unmarshal)

Example (net/http version):
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := client_httpclient.Do(req)
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &res)

Summary:
--------
- `go-http-client_httpclient` → Like **automatic car** (easier, less code)
- `net/http` → Like **manual car** (more control, more code)

Both reach the same result, but with different effort.
*/
