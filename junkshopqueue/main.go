package main

import "strings"

const (
	buyMsg             = "i will buy it"
	failedBuyMsg       = "sold"
	successfullyBuyMsg = "worth it :))"
	whichOneMsg        = "which one?"

	//info
	infoMsg    = "what do dou have?"
	nothingMsg = "i have nothing"

	//sell
	thxMsg = "thank you"
)

// handleMessage only handles the message logic
func handleMessage(msg string, state *ClientState, paintings *[]string, firstUnsold *int) string {
	// info
	if msg == infoMsg {
		// 1. If we have no unsold paintings
		if *firstUnsold >= len(*paintings) {
			return nothingMsg
		}
		// 2. We have a painting
		// Return the first unsold
		// And store in state that the customer has seen this painting
		state.lastSeen = *firstUnsold
		return (*paintings)[*firstUnsold]
	}
	// sell
	if strings.HasPrefix(msg, "i have") {
		name := strings.TrimPrefix(msg, "i have")
		*paintings = append(*paintings, name)
		return thxMsg
	}
	// buy
	if msg == buyMsg {
		// 1. If no painting is seen
		if state.lastSeen == -1 {
			return whichOneMsg
		}
		// 2. If the painting that saw is now before firstUnsold => Sold
		if state.lastSeen < *firstUnsold {
			return failedBuyMsg
		}
		// 3. If he wants exactly the first unsold painting
		if state.lastSeen == *firstUnsold {
			*firstUnsold++
			return successfullyBuyMsg
		}
		return failedBuyMsg
	}
	return ""
}

type ClientState struct {
	lastSeen int // index of last painting the client has seen, -1 if none
}

// ManageShop receives one argument.
// That argument is a channel of channels.
// `ch chan chan string` means:
//	`chan strin`g → a channel used to talk to one client.
//	`chan chan string` → a channel through which new clients send their own channels.
// So:
//	Each time a new client enters the shop, they send their personal communication channel into ch.
//	ManageShop must receive these client channels and talk to each client through their own chan string.
//
// This pattern is common in Go for concurrency when multiple clients need their own private communication channels
// but are introduced through a single central entry point.

func ManageShop(ch chan chan string) {
	var paintings []string                           // names of all paintings
	var firstUnsold int = 0                          // index of first unsold painting
	var clients = make(map[chan string]*ClientState) // maintaining the status of all customers

}
