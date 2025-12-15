package smart_thermostat

import (
	"context"
	"fmt"
	"sync"
)

// SystemController manages rooms centrally
type SystemController struct {
	Rooms   map[string]*Room
	Reports map[string]string // populated by GenerateReports()
	mu      sync.RWMutex
}

// Room represents a room and its thermostat + fan state
type Room struct {
	ID         string
	Thermostat Thermostat
	Occupied   bool
	FanRunning bool
	mu         sync.RWMutex
	cancel     context.CancelFunc // non-nil only while fan goroutine is active
}

// Why we add `cancel` field?
// Because the fan runs in a goroutine (background loop ticking every second.)
// We need a safe way to stop the goroutine when:
//	1. room becomes empty
//	2. temperature reaches the target
//	3. controller updates the temp to equal target
// 	4. or `StopFan()` is called.
// `cancel` (from `context.WithCancel`) is like a remote OFF button for the goroutine:
//	- the goroutine checks `ctx.Done()` and exits
// 	- calling `cancel()` is safe even if called multiple times
// 	- it avoids bugs like closing a channel twice (which can crash your program)

// Thermostat hold temperature state
type Thermostat struct {
	CurrentTemperature int
	TargetTemperature  int
}

// NewSystemController creates an instance of SystemController
func NewSystemController() *SystemController {
	return &SystemController{
		Rooms:   make(map[string]*Room),
		Reports: make(map[string]string),
	}
}

// AddRoom adds a new room to the system.
// This operation may be performed concurrently.
func (s *SystemController) AddRoom(room *Room) error {
	if room == nil || room.ID == "" {
		return fmt.Errorf("room ID is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Rooms[room.ID]; ok {
		return fmt.Errorf("room already exists")
	}

	s.Rooms[room.ID] = room
	return nil
}

func (s *SystemController) UpdateRoomTemperature(roomID string, newTemp int) error {
	return nil
}

func (s *SystemController) GenerateReports() map[string]string {
	return map[string]string{}
}

// GetCurrentTemperature returns the room current temperature
func (r *Room) GetCurrentTemperature() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Thermostat.CurrentTemperature
}

// GetTargetTemperature returns the room target temperature
func (r *Room) GetTargetTemperature() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Thermostat.TargetTemperature
}

// GetIsRoomOccupied returns the occupancy status of the room
func (r *Room) GetIsRoomOccupied() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Occupied
}

// GetIsFanRunning indicates whether the room fan is ON or OFF
func (r *Room) GetIsFanRunning() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.FanRunning
}

func (r *Room) SetOccupancy(occupied bool) {

}

func (r *Room) StartFan() error {
	return nil
}

func (r *Room) StopFan() error {
	return nil
}
