package server

// clientRequest represents one JSON request sent by a client over the TCP connection.
//
// The protocol supports multiple actions, and the shape of the JSON depends on
// the action:
//
//   - subscribe / unsubscribe:
//     {"action":"subscribe","topic":"topic_name"}
//     {"action":"unsubscribe","topic":"topic_name"}
//
//   - publish:
//     {"action":"publish","message":{"topic":"t","content":"c","priority":3}}
//
// Notes:
//   - Topic is used directly for subscribe/unsubscribe.
//   - Message is used for publish and is nil if the client didn't include it.
//     This lets us distinguish "message is required" from other validation errors.
type clientRequest struct {
	// Action determines which operation the server should perform.
	Action string `json:"action"`

	// Topic is required for subscribe/unsubscribe requests.
	// It is ignored for publish requests (publish uses Message.Topic instead).
	Topic string `json:"topic"`

	// Message holds the publish payload for "publish" requests.
	// It will be nil if the JSON request does not include the "message" field.
	Message *publishMessage `json:"message"`
}

// publishMessage is the nested "message" object used in publish requests.
//
// Example:
//
//	{"action":"publish","message":{"topic":"t","content":"c","priority":1}}
//
// Priority is a pointer so we can detect whether the field was omitted.
// This matters because the spec requires a specific error when "priority" is missing:
//   - if "priority" is missing entirely => "priority is required"
//
// If Priority were an int (not a *int), missing priority would decode as 0,
// and you could not tell the difference between:
//   - missing field (invalid) and
//   - explicitly provided priority = 0 (also likely invalid depending on your rules).
type publishMessage struct {
	// Topic is the target topic name the message should be published to.
	Topic string `json:"topic"`

	// Content is the message payload to deliver to subscribers.
	Content string `json:"content"`

	// Priority controls delivery ordering in the per-topic heap.
	// Lower numeric values mean higher priority (1 is higher than 3).
	//
	// Pointer form allows validation of "field missing".
	Priority *int `json:"priority"`
}
