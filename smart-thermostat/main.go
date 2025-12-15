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

// UpdateRoomTemperature regulates the occupancy of the room.
// If there are people in the room, it will turn on to reach the target temperature,
// and if the room is empty, it will turn off.
// Of course, the thermostat also has an effect on turning the fan on or off.
// But generally, if the room is empty, the fan will not turn on.
func (s *SystemController) UpdateRoomTemperature(roomID string, newTemp int) error {
	if newTemp < 0 {
		return fmt.Errorf("invalid target temperature")
	}
	s.mu.RLock()
	room, ok := s.Rooms[roomID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("room does not exist")
	}

	// Update current temperature and snapshot state
	room.mu.Lock()
	room.Thermostat.CurrentTemperature = newTemp
	target := room.Thermostat.TargetTemperature
	occupied := room.Occupied
	running := room.FanRunning
	room.mu.Unlock()

	// apply fan rules
	if !occupied || newTemp == target {
		if running {
			_ = room.StartFan() //
		}
		return nil
	}
	// in this clause the room is occupied and `newTemp != target` (so temperature needs to change and fan should be ON)
	// if `running == true`, fan is already ON and do nothing
	// if `running != true`, fan is OFF, and we need to start
	// Why we ignore "fan already running" error?
	//	- Because concurrency. Imagine two goroutines at the same time both decide the fan should start:
	// 		- Goroutine A checks running == false
	//		- Goroutine B checks running == false
	// 		- Both call StartFan()
	// 		- A starts it successfully
	// 		- B now sees it’s already running and StartFan() returns error "fan already running"
	// 		- That error is not a “real failure” — the fan is already ON, which is exactly what we want. So we ignore it.
	if !running {
		// should succeed; if it returns "fan already running" due to race, you can ignore it
		if err := room.StartFan(); err != nil && err.Error() != "fan already running" {
			return err
		}
		// Meaning:
		//	- If StartFan returns nil, good
		// 	- If StartFan returns "fan already running" → still good (goal achieved)
		// 	- If StartFan returns any other error (like "room is not occupied" or "no adjustment needed")
		//	- return it because that’s unexpected here
	}
	return nil
}

// GenerateReports generates reports on the temperature status and fan operation mode in the system.
// These reports include the room ID and the thermostat operation mode of each room.
// The fan operation mode is calculated as follows:
//   - "cooling" if the fan is cooling the room,
//   - "heating" if it is heating the room,
//   - "off" if fan is off.
func (s *SystemController) GenerateReports() map[string]string {
	// 1. snapshot room pointers safely (don’t hold lock too long)
	s.mu.RLock()
	rooms := make([]*Room, 0, len(s.Rooms))
	for _, r := range s.Rooms {
		rooms = append(rooms, r)
	}
	s.mu.RUnlock()
	// 2.compute reports without holding controller lock
	newReports := make(map[string]string, len(rooms))
	for _, r := range rooms {
		r.mu.RLock()
		id := r.ID
		current := r.Thermostat.CurrentTemperature
		target := r.Thermostat.TargetTemperature
		running := r.FanRunning
		r.mu.RUnlock()
		//
		state := "off"
		if running {
			if current < target {
				state = "heating"
			} else if current > target {
				state = "cooling"
			} else {
				state = "off"
			}
		}
		newReports[id] = state
		// 3.Store it automatically
		s.mu.Lock()
		if s.Reports == nil {
			s.Reports = make(map[string]string)
		}
		s.Reports = newReports
		s.mu.Unlock()

	}
	return newReports
}

// Why we take a snapshot of rooms in GenerateReports()?
// 	- Because `s.Rooms` is a map, and other goroutines may call `AddRoom()` at the same time.
// 	- If we loop directly over `s.Rooms` while another goroutine writes to it, Go can crash with:
// 		- fatal error: concurrent map iteration and map write
// 	- So we do this:
// 		- Lock controller (RLock)
// 		- Copy all `*Room` pointers into a slice (`rooms := []*Room{...}`)
// 		- Unlock controller
// 		- Loop over the slice safely (because the slice won’t change)
// 		- That’s the snapshot.
//	- Extra benefit:
//		- We also avoid holding the controller lock for a long time while we:
// 		- lock each room, compute strings, build report map
// 		- So other operations (like AddRoom / UpdateRoomTemperature) don’t get blocked too much.

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
