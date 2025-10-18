package main

import (
	"errors"
	"strings"
	"sync"
)

type Player struct {
	name string
	zone int
	ch   chan string
	m    *Map
}

type Map struct {
	id      int
	mu      sync.Mutex
	players map[string]*Player
	ch      chan string
}

type Game struct {
	mu      sync.Mutex
	players map[string]*Player
	maps    map[int]*Map
}

func NewGame(mapIds []int) (*Game, error) {
	// Create an empty map to store all game maps
	// Key: map ID (int)
	// Value: pointer to Map struct
	gameMaps := make(map[int]*Map)

	// Iterate through all provided map IDs
	for _, id := range mapIds {

		// Validate: map ID must be positive (greater than 0)
		if id <= 0 {
			return nil, errors.New("map id is invalid")
		}

		// Validate: check for duplicate map IDs
		if _, ok := gameMaps[id]; ok {
			return nil, errors.New("map id is duplicated")
		}

		// Create a new Map instance
		m := &Map{
			id:      id,                       // map identifier
			players: make(map[string]*Player), // holds players currently in this map
			ch:      make(chan string, 100),   // message channel for communication between players in this map
		}

		// Store the new Map in the gameMaps collection
		gameMaps[id] = m

		// Start the FanOutMessages goroutine for this map
		// It continuously listens for new messages on m.ch
		// and broadcasts them to all players inside this map
		go m.FanOutMessages()
	}

	// Create the Game instance containing:
	// - an empty player list (no one connected yet)
	// - all initialized maps
	g := &Game{
		players: make(map[string]*Player),
		maps:    gameMaps,
	}

	// Return the newly created Game and no error
	return g, nil
}

func (g *Game) ConnectPlayer(name string) error {
	// Convert player name to lowercase to make it case-insensitive.
	// This ensures "Mamad", "mamaD", and "MAMAD" are treated as the same player.
	key := strings.ToLower(name)

	// Lock the game mutex to prevent concurrent access to g.players.
	// Only one goroutine can modify the players map at a time.
	g.mu.Lock()
	defer g.mu.Unlock() // Automatically unlock when the function returns.

	// Check if a player with the same name already exists in the game.
	if _, ok := g.players[key]; ok {
		return errors.New("player already exists")
	}

	// Create a new Player instance.
	p := &Player{
		name: name,                   // The player's display name (case preserved).
		zone: -1,                     // -1 means the player is not in any map yet.
		ch:   make(chan string, 100), // Buffered channel for receiving chat messages.
		m:    nil,                    // No map assigned yet.
	}

	// Add the new player to the game's players map using the lowercase key.
	g.players[key] = p

	// Successfully connected the player, no error to return.
	return nil
}

func (g *Game) SwitchPlayerMap(name string, mapId int) error {
	// Normalize the name so lookups are case-insensitive.
	key := strings.ToLower(name)

	// Lock the Game while we read/modify shared structures (players/maps).
	// This prevents concurrent goroutines from racing on g.players or g.maps.
	g.mu.Lock()
	defer g.mu.Unlock() // Automatically unlock when the function ends.

	// 1) Validate the destination map exists.
	newMap, ok := g.maps[mapId]
	if !ok {
		return errors.New("map not found")
	}

	// 2) Validate the player exists.
	p, ok := g.players[key]
	if !ok {
		return errors.New("player not found")
	}

	// 3) Short-circuit: already in that map? that's an error.
	if p.zone == mapId {
		return errors.New("player is already in this map")
	}

	// 4) Find the player's current map (if any).
	// zone == -1 means the player is not in any map yet (first move).
	// if zone == -1, the oldMap get nil value, we use it below to prevent panic.
	var oldMap *Map
	if p.zone != -1 {
		oldMap = g.maps[p.zone] // safe: p.zone came from our own state
	}

	// 5) Lock the per-map mutexes in a consistent global order to avoid deadlocks.
	// If both old and new exist, lock the one with the smaller id first.
	// If there is no old map (first move), just lock the new map.
	if oldMap != nil && oldMap.id < newMap.id {
		oldMap.mu.Lock()
		newMap.mu.Lock()
	} else {
		newMap.mu.Lock()
		if oldMap != nil {
			oldMap.mu.Lock()
		}
	}

	// 6) Remove the player from the old map if they had one.
	// Important: only touch oldMap if it is non-nil (first move has no old map).
	// Without this check, calling delete on a nil map would cause a panic.
	if oldMap != nil {
		delete(oldMap.players, key)
	}

	// 7) Add the player to the new map and
	// Update both the player's zone (map ID) and its map pointer reference.
	newMap.players[key] = p
	p.zone = mapId
	p.m = newMap

	// 8) Unlock in reverse order of locking.
	// Again, we check for nil to avoid calling Unlock on a nil map (which would panic).
	newMap.mu.Unlock()
	if oldMap != nil {
		oldMap.mu.Unlock()
	}

	return nil
}

func (g *Game) GetPlayer(name string) (*Player, error) {
	key := strings.ToLower(name)

	g.mu.Lock()
	p, ok := g.players[key]
	g.mu.Unlock()
	if !ok {
		return nil, errors.New("player not found")
	}
	return p, nil
}

func (g *Game) GetMap(mapId int) (*Map, error) {
	g.mu.Lock()
	m, ok := g.maps[mapId]
	g.mu.Unlock()
	if !ok {
		return nil, errors.New("map not found")
	}
	return m, nil
}

func (m *Map) FanOutMessages() {
	// Continuously read incoming packets from the map's channel.
	// Each packet is formatted as: "<senderKey>\x00<displayMessage>"
	for packet := range m.ch {
		// Find the delimiter that separates sender key and message body.
		i := strings.IndexByte(packet, '\x00')
		if i < 0 {
			// Malformed packet: skip safely rather than crashing.
			continue
		}

		// Extract sender key (lowercase player name) and the display-ready message.
		senderKey := packet[:i]
		raw := packet[i+1:] // e.g., "Mamad says: salam"

		// Lock the map while iterating over its players to avoid data races.
		m.mu.Lock()
		for key, p := range m.players {
			// Do not echo the message back to the sender.
			// Also double-check the player still belongs to this map (defensive check).
			if key != senderKey && p.m == m {
				// Non-blocking send to the player's channel:
				// - If the channel buffer has space, deliver the message.
				// - If it's full, drop the message to prevent blocking this goroutine.
				select {
				case p.ch <- raw:
				default:
					// Drop when the receiver is slow; keeps fan-out loop responsive.
				}
			}
		}
		m.mu.Unlock()
	}
}

func (p *Player) GetChannel() <-chan string {

	return p.ch
}

func (p *Player) SendMessage(msg string) error {
	// A player must be connected to a map to send a message.
	// If p.m is nil, the player hasn't joined any map yet.
	if p.m == nil {
		return errors.New("player is not connected")
	}

	// Convert player's name to lowercase for internal use.
	lower := strings.ToLower(p.name)

	// Build a display-friendly name: first letter uppercase, rest lowercase.
	display := ""
	if len(lower) > 0 {
		display = strings.ToUpper(lower[:1]) + lower[1:]
	}

	// Create the final text that other players will see.
	// Example: "Mamad says: hello"
	raw := display + " says: " + msg

	// Prepare the message packet.
	// Format: "<senderKey>\x00<displayMessage>"
	// Example: "mamad\x00Mamad says: hello"
	packet := strings.ToLower(p.name) + "\x00" + raw

	// Try to send the packet into the map's channel.
	// Use a non-blocking select to avoid freezing if the map channel is full.
	select {
	case p.m.ch <- packet:
		// Successfully sent.
		return nil
	default:
		// Map's channel buffer (100 messages) is full.
		return errors.New("map channel is full")
	}
}

func (p *Player) GetName() string {

	return p.name
}
