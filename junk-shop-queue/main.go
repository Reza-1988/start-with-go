package main

import "strings"

type ClientState struct {
	lastSeen int // index of last painting the client has seen, -1 if none
}

// handleMessage only handles the message logic for a single client request.
func handleMessage(msg string, state *ClientState, paintings *[]string, firstUnsold *int) string {
	// info: "what do dou have?"
	if msg == "what do dou have?" {
		// 1. No unsold paintings left
		if *firstUnsold >= len(*paintings) {
			return "i have nothing"
		}
		// 2. We have at least one unsold painting.
		// Return the first unsold and remember that this client has seen it.
		state.lastSeen = *firstUnsold
		return (*paintings)[*firstUnsold]
	}

	// sell: "i have <name>"
	if strings.HasPrefix(msg, "i have ") {
		name := strings.TrimPrefix(msg, "i have ")
		*paintings = append(*paintings, name)
		return "thank you"
	}

	// buy: "i will buy it"
	if msg == "i will buy it" {
		// 1. Client has not seen any painting yet
		if state.lastSeen == -1 {
			return "which one?"
		}
		// 2. The painting they saw is already before firstUnsold → already sold
		if state.lastSeen < *firstUnsold {
			return "sold"
		}
		// 3. Client is buying exactly the first unsold painting
		if state.lastSeen == *firstUnsold {
			*firstUnsold++
			return "worth it :))"
		}
		// Fallback (shouldn't normally happen)
		return "sold"
	}

	// Unknown message (not used in tests)
	return ""
}

// ManageShop receives one argument: That argument is a channel of client channels.
// Sch: chan chan string
//
//   - chan string  → a channel used to talk to one client.
//   - chan chan string → a channel through which new clients send their own channels.
//
// So:
// Each time a new client enters the shop, they send their personal channel into Sch.
// ManageShop receives that channel and starts a goroutine to talk to that client.
func ManageShop(Sch chan chan string) {
	// paintings warehouse (shared state)
	paintings := []string{} // names of all paintings
	firstUnsold := 0        // index of first unsold painting

	// per-client state, keyed by their channel
	clients := make(map[chan string]*ClientState)

	for {
		// Wait for a new client to send us their channel.
		clientChan := <-Sch

		// If we don't have state for this channel yet, let's create it
		state, ok := clients[clientChan]
		if !ok {
			state = &ClientState{lastSeen: -1}
			clients[clientChan] = state
		}

		// Dedicated goroutine for this specific client.
		go func(ch chan string, st *ClientState) {
			for {
				// Receive a message from this client.
				msg := <-ch

				// Process the message using shared shop state + this client state.
				reply := handleMessage(msg, st, &paintings, &firstUnsold)

				// Send the response back on the same channel.
				ch <- reply
			}
		}(clientChan, state)
	}
}
