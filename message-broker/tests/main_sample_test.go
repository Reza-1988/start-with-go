package test

import (
	"QueraMQ/server"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestBasicFunctionality is a high-level integration test that exercises the
// "happy path" of the broker over real TCP connections:
//
//   - Start the server in the background.
//   - Start a consumer that subscribes to a topic and waits for a delivered message.
//   - Start a producer that publishes one message to that same topic.
//   - Assert the consumer receives a delivery and the topic exists on the server.
//   - Finally, request a graceful shutdown of the server.
//
// This test validates basic concurrency (server + clients), JSON protocol handling,
// topic creation, and message delivery end-to-end.
func TestBasicFunctionality(t *testing.T) {
	// Use a loopback TCP address for an isolated local test server.
	// Note: hard-coding a port can cause flakiness if the port is in use.
	// (Many tests instead choose ":0" and read the actual port, but here the
	// assignment/test harness expects a fixed address.)
	serverAddr := "127.0.0.1:8080"

	// Construct the server instance. NewServer typically just sets fields;
	// the server does not start listening until Run() is called.
	srv := server.NewServer(serverAddr)

	// Start the server in a separate goroutine because Run() is expected to
	// block while accepting connections. If Run returns an error, report it
	// via the test framework.
	go func() {
		if err := srv.Run(); err != nil {
			// Using t.Errorf here records the failure without stopping the goroutine.
			// In production-quality tests, you often coordinate shutdown and
			// avoid logging errors after test completion, but this is sufficient
			// for a simple sample.
			t.Errorf("Server failed to run: %v", err)
		}
	}()

	// Give the server time to start listening before clients dial in.
	// Best practice is to wait on a readiness signal (channel/health check)
	// instead of sleeping, but sleep is used here for simplicity.
	time.Sleep(100 * time.Millisecond)

	// WaitGroup is used to coordinate that goroutines reach certain milestones.
	// In this test, runConsumer calls wg.Done() once it has successfully subscribed,
	// and runProducer calls wg.Done() once it has successfully published.
	var wg sync.WaitGroup

	// Start a consumer in a goroutine. wg.Add(1) must happen before starting
	// the goroutine to avoid a race where Done() is called before Add().
	wg.Add(1)
	go runConsumer(t, serverAddr, "test_topic", &wg)

	// Give the consumer time to connect and subscribe before producing a message.
	// Without this, the producer might publish before the consumer subscribes,
	// and the spec says subscribers do not receive messages published before they subscribe.
	time.Sleep(100 * time.Millisecond)

	// Start the producer in a goroutine to publish one message to the same topic.
	// Priority=1 means "highest priority" (smaller number => higher priority) per the spec.
	wg.Add(1)
	go runProducer(t, serverAddr, "test_topic", "Hello, World!", 1, &wg)

	// Wait until both:
	//   - the consumer has subscribed successfully, and
	//   - the producer has published successfully.
	//
	// Note: runConsumer continues after wg.Done() to block waiting for "deliver";
	// wg.Wait() does NOT wait for the consumer to receive the delivered message.
	wg.Wait()

	// Verify that the topic exists on the server. According to the assignment,
	// GetTopic should create a topic if missing, and return (topic, true) if it
	// already existed. Since the consumer/producer used "test_topic", it should exist now.
	_, exists := srv.GetTopic("test_topic")
	require.True(t, exists, "Failed to get topic")

	// Open a fresh TCP connection to send a shutdown request.
	// The test does not reuse producer/consumer connections; it creates a new one
	// to simulate an arbitrary client requesting shutdown.
	conn, err := net.Dial("tcp", serverAddr)
	require.NoError(t, err, "Producer failed to connect")
	defer conn.Close()

	// JSON encoder writes newline-delimited JSON objects to the TCP stream.
	// This matches the typical "one JSON object per message" framing used with json.Encoder.
	encoder := json.NewEncoder(conn)

	// Request a graceful server shutdown. The server should close all client connections
	// and stop accepting new ones.
	message := map[string]interface{}{"action": "shutdown"}

	// The sample test does not check the response here, but your server should
	// ideally respond with {"status":"ok"} before closing connections.
	encoder.Encode(message)
}

// runConsumer acts like a broker "consumer" client for the integration test.
//
// It connects to the TCP server, subscribes to a topic via the JSON protocol,
// verifies the server replies with {"status":"ok"}, then blocks waiting for a
// delivered message. After receiving one "deliver" message, it requests
// connection closure.
//
// The WaitGroup is used only to signal "subscription completed" back to the
// parent test, so the producer can publish after the consumer is ready.
// (It does NOT represent the full lifetime of this goroutine.)
func runConsumer(t *testing.T, address, topic string, wg *sync.WaitGroup) {
	// Establish a TCP connection to the broker. This connection stays open and is
	// reused for both sending requests and receiving asynchronous deliveries.
	consumerConn, err := net.Dial("tcp", address)
	require.NoError(t, err, "Consumer failed to connect")
	// Ensure the connection is closed when this goroutine exits.
	// Closing the connection helps free resources and unblocks server-side reads.
	defer consumerConn.Close()

	// json.Encoder / json.Decoder provide a convenient stream protocol:
	// each Encode writes one JSON value (typically newline-delimited),
	// each Decode reads one JSON value in order from the stream.
	consumerEncoder := json.NewEncoder(consumerConn)
	consumerDecoder := json.NewDecoder(consumerConn)

	// Build and send the subscribe request according to the assignment protocol.
	// Using map[string]interface{} is flexible for tests, though in production
	// code you'd usually define request/response structs for type safety.
	subscribeRequest := map[string]interface{}{
		"action": "subscribe",
		"topic":  topic,
	}
	err = consumerEncoder.Encode(subscribeRequest)
	require.NoError(t, err, "Failed to send subscribe request")

	// The server must acknowledge a successful subscribe with {"status":"ok"}.
	// Decode blocks until a full JSON object arrives or the connection closes/errors.
	var subscribeResponse map[string]interface{}
	err = consumerDecoder.Decode(&subscribeResponse)
	require.NoError(t, err, "Failed to read subscribe response")
	require.Equal(t, "ok", subscribeResponse["status"], "Subscribe failed")

	// Signal to the parent test that the consumer is now subscribed and ready.
	// This allows the producer goroutine to publish after subscription is confirmed.
	wg.Done()

	// Wait for the next JSON message from the server, which should be an async
	// delivery event triggered by a producer publishing to the subscribed topic.
	var deliverMessage map[string]interface{}
	err = consumerDecoder.Decode(&deliverMessage)
	require.NoError(t, err, "Failed to read delivered message")

	// The server should send messages to consumers with action "deliver".
	require.Equal(t, "deliver", deliverMessage["action"], "Expected action 'deliver'")

	// Basic validation that the delivered payload includes a nested "message" object.
	// This test only checks that it is an object-like structure, not the full schema.
	_, ok := deliverMessage["message"].(map[string]interface{})
	require.True(t, ok, "Invalid message format")

	// Ask the server to close this client connection gracefully.
	// Per the spec, the server should unsubscribe this client from all topics
	// before closing the underlying TCP connection.
	message := map[string]interface{}{"action": "close_connection"}

	// The test does not assert the response here, but a robust server would likely
	// reply {"status":"ok"} (or just close, depending on your chosen behavior).
	consumerEncoder.Encode(message)
}

// runProducer acts like a broker "producer" client for the integration test.
//
// It connects to the TCP server, publishes a single message to a topic with a
// given priority using the JSON protocol, verifies the server replies with
// {"status":"ok"}, and then asks the server to close the connection.
//
// The WaitGroup is used to let the parent test know when publishing has
// completed (successfully or not).
func runProducer(t *testing.T, address, topic, content string, priority int, wg *sync.WaitGroup) {
	// Ensure we always signal completion to the parent test, even if an assertion fails
	// (require.* will call t.FailNow(), which exits the goroutine, but deferred calls
	// still run).
	defer wg.Done()

	// Establish a TCP connection to the broker. Like the consumer, the producer keeps
	// the connection open long enough to send requests and read responses.
	producerConn, err := net.Dial("tcp", address)
	require.NoError(t, err, "Producer failed to connect")
	// Always close the connection when done to avoid leaking file descriptors.
	defer producerConn.Close()

	// Encoder/Decoder provide stream-based JSON messaging over TCP.
	// Each Encode sends one JSON object; each Decode reads the next JSON object.
	producerEncoder := json.NewEncoder(producerConn)
	producerDecoder := json.NewDecoder(producerConn)

	// Construct the publish request according to the protocol:
	// - action = "publish"
	// - message object contains topic/content/priority
	//
	// In production, you'd typically use typed structs to avoid runtime type assertions
	// and to ensure correct JSON field names via struct tags.
	publishRequest := map[string]interface{}{
		"action": "publish",
		"message": map[string]interface{}{
			"topic":    topic,
			"content":  content,
			"priority": priority,
		},
	}

	// Send the publish request to the server.
	err = producerEncoder.Encode(publishRequest)
	require.NoError(t, err, "Failed to send publish request")

	// The server must acknowledge a successful publish with {"status":"ok"}.
	// Decode blocks until it receives that response or the connection closes/errors.
	var publishResponse map[string]interface{}
	err = producerDecoder.Decode(&publishResponse)
	require.NoError(t, err, "Failed to read publish response")
	require.Equal(t, "ok", publishResponse["status"], "Publish failed")

	// Ask the server to close this producer connection.
	// A correct server will remove this client from any internal registries and then
	// close the underlying TCP connection.
	message := map[string]interface{}{"action": "close_connection"}

	// The test does not assert the response here, but a robust server might respond
	// with {"status":"ok"} before closing.
	producerEncoder.Encode(message)
}
