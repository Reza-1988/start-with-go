What I gave you is **correct and complete for the core requirements**, but if you’re putting this into `ARCHITECTURE.md` and want it to be truly “complete + wise”, there are a few important pieces worth adding so future-you won’t get surprised:

* **Where locks must/ must not be held** (most common source of deadlocks)
* **Message framing over TCP** (why `json.Encoder/Decoder` works here)
* **What “immediate delivery” really means** (and why we use `notify` + dispatcher)
* **Edge cases + exact error rules** (Quera is picky)
* **Limitations / simplifications** (what your broker does *not* guarantee)

Below is a refined, more “final” `ARCHITECTURE.md` style explanation you can paste.

---

# ARCHITECTURE.md (QueraMQ)

## Goal

Build a simple TCP message broker (RabbitMQ-like) to practice networking and concurrency:

* Producers publish messages to topics.
* Consumers subscribe/unsubscribe to topics.
* Messages are delivered immediately to current subscribers.
* Each topic has its own priority queue (heap): smaller `priority` number = higher priority.
* Server supports multiple concurrent clients.
* Server supports graceful shutdown.

---

## Components

### 1) Queue (package `queue`)

**Purpose:** Maintain messages in priority order.

* `Message` holds message data:

    * `ID` (UUID v4), `Content`, `Priority`
    * optional but recommended: `Topic`, `Seq`
* `MessageQueue` implements `heap.Interface`:

    * `Less(i,j)` returns `Priority` ordering (smaller = higher priority)
    * heap operations are `O(log n)` and retrieving top is `O(1)`

**Important:** `container/heap` is **not** concurrency-safe. Heap operations must be guarded by a mutex in the topic.

---

### 2) Topic (package `server`)

**Purpose:** Represents a channel (chat-room like) with its own queue and subscriber list.

Each `Topic` contains:

* `Name`
* `MQ` (priority heap)
* `subscribers` (map of active subscribers)
* `seq` (monotonic publish counter)
* `notify` channel (wakes dispatcher)
* `mu` lock (protects all topic state)

#### Enforcing “no old messages on subscribe”

Requirement: subscriber receives only messages published *after* subscribing.

Mechanism:

* Topic maintains `seq` as logical time.
* On subscribe: store `joinedSeq = current t.seq`.
* On publish: increment `t.seq` and assign `msg.Seq = t.seq`.
* On delivery: subscriber eligible only if `joinedSeq < msg.Seq`.

This ensures new subscribers never see older messages, even if older messages still exist in the heap.

---

### 3) Server (package `server`)

**Purpose:** Accept TCP connections, parse JSON actions, route to topics, and manage shutdown.

Server contains:

* `Addr` (listen address)
* `ln` (listener, closed on Stop)
* `topics` map + lock
* `clients` map + lock
* `stopCh` (closed to broadcast shutdown)
* `wg` (wait for goroutines)
* `stopOnce` (Stop is idempotent)

---

### 4) Client wrapper (`server.Client`)

**Purpose:** Safe JSON writes to a single TCP connection.

Why needed:

* The request handler goroutine writes responses (`{"status":"ok"}`, errors).
* The topic dispatcher writes deliveries (`{"action":"deliver", ...}`).
  These may happen concurrently.

Solution:

* `Client.Send()` uses a mutex to serialize `json.Encoder.Encode()` so JSON messages never interleave on the TCP stream.

---

## Execution Flow

### Server startup (`Run`)

1. `net.Listen("tcp", Addr)`
2. Loop:

    * `Accept()` blocks until a client connects
    * create `Client` wrapper (`id := conn.RemoteAddr().String()`)
    * store client in `s.clients`
    * start goroutine `handleClient(client)`

Server can handle many clients because each connection is handled in its own goroutine.

---

### Client request loop (`handleClient`)

Each client connection runs this loop:

1. `json.Decoder.Decode(&req)` blocks for the next JSON request
2. Switch on `req.Action`:

#### `subscribe`

* validate `topic`
* `GetTopic(topic)` (create if missing)
* add client to `topic.subscribers` with `joinedSeq = t.seq`
* record membership in `client.subs`
* respond `{"status":"ok"}`

#### `unsubscribe`

* validate `topic`
* remove from `topic.subscribers` if topic exists
* remove from `client.subs`
* respond `{"status":"ok"}`

#### `publish`

* validate message presence and fields:

    * missing message => `"message is required"`
    * missing topic => `"topic is required"`
    * missing content => `"message content is required"`
    * missing priority => `"priority is required"` (priority decoded as `*int` to detect omission)
* `GetTopic(message.topic)`
* generate UUID
* lock topic:

    * increment `t.seq`
    * push message into heap (`heap.Push`)
* non-blocking notify: `t.notify <- struct{}{}` (if possible)
* respond `{"status":"ok"}`

#### `close_connection`

* (optional) respond ok
* return → deferred cleanup runs

#### `shutdown`

* respond ok
* call `Stop()` (often in a goroutine)
* return

#### unknown action

* respond `{"error":"unknown action"}`

---

### Topic dispatcher (`topicLoop`)

Each topic has a dispatcher goroutine:

1. Wait on either:

    * `t.notify` (new message published)
    * `s.stopCh` (shutdown)
2. When notified:

    * pop messages from heap until empty (priority order)
    * snapshot eligible subscribers (no locks during I/O)
    * send `"deliver"` JSON to each subscriber via `Client.Send()`

**Important best practice:** never hold `topic.mu` while sending over the network.

---

## Cleanup and shutdown

### Client cleanup (`removeClient`)

When a client disconnects or requests close:

* remove from `s.clients`
* snapshot and clear `client.subs`
* for each subscribed topic:

    * remove client from `topic.subscribers`
* close the TCP connection

### Graceful shutdown (`Stop`)

Stop must end all blocking calls:

* close `stopCh` (broadcast shutdown to loops)
* close listener (unblocks `Accept`)
* close all client conns (unblocks `Decode`)
* wait for goroutines to exit (`wg.Wait`)

---

## Concurrency rules (critical)

* Protect shared maps with locks:

    * `Server.topics`, `Server.clients`
    * `Topic.MQ`, `Topic.subscribers`, `Topic.seq`
* Never hold locks during network writes.
* Serialize per-connection writes with `Client.Send()` mutex.

---

## TCP + JSON framing

Using `json.Encoder.Encode()` sends one JSON value and a newline.
Using `json.Decoder.Decode()` reads one JSON value at a time from the stream.
This makes a simple “one JSON object per message” protocol over TCP.

---

## Known simplifications / limitations

* No persistence: messages are in memory only.
* No delivery acknowledgements / retries.
* Slow consumers may miss messages if their connection is broken (best-effort).
* If you don’t implement a stable tie-breaker for equal priority, ordering among equal priorities may be non-deterministic.

