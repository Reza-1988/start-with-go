# Calculator TCP API (Add/Sub)

A small TCP server that exposes a **simple calculator API** over a raw TCP connection using **JSON messages**.
It supports only two actions:

* `add` → sum all numbers
* `sub` → subtract numbers in order (`first - second - third ...`)

The server returns responses as JSON with a strict output format required by the challenge.

---

## Project Structure

```
.
├── go.mod
├── go.sum
├── main.go
└── main_sample_test.go
```

---

## How It Works

### Transport

* **TCP** server (not HTTP).
* Clients connect to: `localhost:<port>`
* Client sends **one JSON object** (newline-delimited JSON is typical).
* Server replies with **one JSON object**.
* A single connection may send **multiple requests**, and the server responds to each one.

---

## Request Format

A request is a JSON object:

```json
{
  "action": "add",
  "numbers": "2,1,-3"
}
```

### Fields

* `action`: must be either:

    * `"add"`
    * `"sub"`
* `numbers`: a **comma-separated string** of integers (int64)

Examples:

```json
{ "action": "add", "numbers": "2,1" }
{ "action": "sub", "numbers": "2,1,55,100,1,-3" }
```

---

## Response Format

Server always replies with:

```json
{
  "result": "...",
  "error":  "..."
}
```

### Success

When the request is valid:

* `error` is empty
* `result` matches this exact format:

`The result of your query is: %d`

Example:

```json
{ "result": "The result of your query is: 3", "error": "" }
```

### Errors (exact messages)

The challenge requires **exact** error texts (case-sensitive):

1. Missing numbers parameter
   If `numbers` is missing or empty/whitespace:

```json
{ "result": "", "error": "'numbers' parameter missing" }
```

2. Invalid number format
   If any token is not a valid int64 integer (or malformed like `1,,2`):

```json
{ "result": "", "error": "Invalid number format" }
```

3. Overflow
   If the operation would overflow an `int64` at any step:

```json
{ "result": "", "error": "Overflow" }
```

---

## Overflow Handling (int64)

Go integers do not throw on overflow—they **wrap around**.
This server explicitly checks overflow boundaries (`math.MaxInt64`, `math.MinInt64`) before applying operations.

Example of overflow case:

* `9223372036854775807 + 1` must return `"Overflow"` rather than wrapping to a negative number.

---

## Running the Server

Example (port can be anything, test uses `4001`):

```bash
go run . 4001
```

(Depending on your `main.go`, port may be hard-coded or passed via args—use the method implemented in your file.)

---

## Testing

Run unit tests:

```bash
go test ./...
```

---

## Manual Testing (TCP Client)

### Using `nc` (netcat)

1. Start server on port `4001`.
2. In another terminal:

```bash
nc localhost 4001
```

Send (press Enter after the JSON):

```json
{"action":"add","numbers":"2,1"}
```

You should receive:

```json
{"result":"The result of your query is: 3","error":""}
```

Try invalid input:

```json
{"action":"add"}
```

Response:

```json
{"result":"","error":"'numbers' parameter missing"}
```

---

## Implementation Notes (Best Practices Used)

* **Goroutine per connection** so slow clients don’t block new connections.
* **Streaming JSON decoding** using `json.Decoder` on the TCP connection.
* **Strict response contract**: only the allowed error strings and exact success message.
* **Overflow-safe arithmetic** for `int64` operations.
* Clean separation of concerns:

    * accept loop (`Start`)
    * per-connection handler (`handleConn`)
    * routing by action (`route`)
    * operation handlers (`handleAdd`, `handleSub`)

