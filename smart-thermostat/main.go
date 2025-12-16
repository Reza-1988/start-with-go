package main

import (
	"fmt"
	"sync"
	"time"
)

// SystemController manages rooms centrally
type SystemController struct {
	Rooms map[string]*Room
	mu    sync.RWMutex
}

// Room represents a room and its thermostat + fan state
type Room struct {
	ID         string
	Thermostat Thermostat
	Occupied   bool
	FanRunning bool
	mu         sync.Mutex
}

// Thermostat hold temperature state
type Thermostat struct {
	CurrentTemperature int
	TargetTemperature  int
}

// NewSystemController creates an instance of SystemController
func NewSystemController() *SystemController {
	return &SystemController{
		Rooms: make(map[string]*Room),
	}
}

// AddRoom adds a new room to the system. This operation may be performed concurrently.
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
	// Validate temperature
	if newTemp < 0 {
		return fmt.Errorf("invalid target temperature")
	}

	// Find the room safely
	s.mu.RLock()
	room, ok := s.Rooms[roomID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("room does not exist")
	}

	// Update current temperature and read state under room lock
	room.mu.Lock()
	room.Thermostat.CurrentTemperature = newTemp
	current := room.Thermostat.CurrentTemperature
	target := room.Thermostat.TargetTemperature
	occupied := room.Occupied
	room.mu.Unlock()
	// TODO
	// Case 1: current == target then, fan should be OFF
	if current == target {
		if err := room.StopFan(); err != nil {
			// If fan is already stopped, that's fine
			if err.Error() != "fan not running" {
				return err
			}
		}
		return nil
	}

	// Case 2: current != target then, fan should be ON, but only if room is occupied
	// TODO
	// If room is not occupied, ensure fan is off and exit
	if !occupied {
		if err := room.StopFan(); err != nil {
			// If fan already off, ignore
			if err.Error() != "fan not running" {
				return err
			}
		}
		return nil
	}
	// Room is occupied and needs adjustment then, start fan
	if err := room.StartFan(); err != nil {
		switch err.Error() {
		case "fan already running", "no adjustment needed", "room is not occupied":
			// These are acceptable races from our perspective:
			// - already running -> good
			// - no adjustment needed -> temp changed in between
			// - room is not occupied -> occupancy changed
			return nil
		default:
			return err
		}
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
	reports := make(map[string]string)

	// Take a snapshot of rooms under read lock
	s.mu.RLock()
	rooms := make([]*Room, 0, len(s.Rooms))
	for _, room := range s.Rooms {
		rooms = append(rooms, room)
	}
	s.mu.RUnlock()

	// Inspect each room independently
	for _, room := range rooms {
		if room == nil {
			continue
		}

		room.mu.Lock()
		id := room.ID
		current := room.Thermostat.CurrentTemperature
		target := room.Thermostat.TargetTemperature
		running := room.FanRunning
		room.mu.Unlock()

		// Default mode
		mode := "off"

		if running {
			if current > target {
				mode = "cooling"
			} else if current < target {
				mode = "heating"
			} else {
				// current == target but fan running -> logically off,
				// but this should rarely happen if other logic is correct.
				mode = "off"
			}
		}

		if id != "" {
			reports[id] = mode
		}
	}

	return reports
}

// Why we take a snapshot of rooms in GenerateReports?
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
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Thermostat.CurrentTemperature
}

// GetTargetTemperature returns the room target temperature
func (r *Room) GetTargetTemperature() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Thermostat.TargetTemperature
}

// GetIsRoomOccupied returns the occupancy status of the room
func (r *Room) GetIsRoomOccupied() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Occupied
}

// GetIsFanRunning indicates whether the room fan is ON or OFF
func (r *Room) GetIsFanRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.FanRunning
}

// SetOccupancy adjusts the occupancy status of the room.
//   - If there are people in the room, the fan will turn on to bring the temperature to the target temperature,
//   - if the room is empty, the fan will stop.
//   - Of course, the thermostat temperature also affects whether the fan turns on or off.
//   - But in general, the fan will not turn on if the room is empty.
func (r *Room) SetOccupancy(occupied bool) {
	// Update occupancy and take a snapshot of temperatures under the room lock
	r.mu.Lock()
	r.Occupied = occupied
	current := r.Thermostat.CurrentTemperature
	target := r.Thermostat.TargetTemperature
	running := r.FanRunning
	r.mu.Unlock()

	// If the room is now empty, we must ensure the fan is stopped.
	if !occupied {
		if running {
			_ = r.StartFan()
		}
		return
	}

	// Room is occupied from here on.

	// If temperature is already at target, no need to start the fan.
	if current == target {
		if running {
			_ = r.StopFan()
		}
		return
	}

	// Otherwise, room is occupied AND needs adjustment, then try to start the fan.
	if !running {
		_ = r.StartFan()
	}

}

// TODO

// StartFan This function activates the room fan.
// The fans are able to gradually change the room temperature to bring it to its target temperature.
// Their temperature change rate is 1°C/second.
//   - Note that each room fan can be turned on or off, and they must all be able to operate in their room at the same time.
func (r *Room) StartFan() error {
	r.mu.Lock()

	// 1. If fan is already running
	if r.FanRunning {
		r.mu.Unlock()
		return fmt.Errorf("fan already running")
	}

	// 2. If room is not occupied
	if !r.Occupied {
		r.mu.Unlock()
		return fmt.Errorf("room is not occupied")
	}

	// 3. If temperature does not need adjustment
	if r.Thermostat.CurrentTemperature == r.Thermostat.TargetTemperature {
		r.mu.Unlock()
		return fmt.Errorf("no adjustment needed")
	}

	// 4. Otherwise, start the fan
	r.FanRunning = true
	r.mu.Unlock()

	// Start a background goroutine to adjust temperature 1°C per second
	go func(room *Room) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			room.mu.Lock()

			// If fan was turned off or room became empty, stop the goroutine
			if !room.FanRunning || !room.Occupied {
				room.mu.Unlock()
				return
			}

			cur := room.Thermostat.CurrentTemperature
			target := room.Thermostat.TargetTemperature

			// If we've already reached the target, stop the fan
			if cur == target {
				room.FanRunning = false
				room.mu.Unlock()
				return
			}

			// Move 1 degree toward the target
			if cur < target {
				room.Thermostat.CurrentTemperature++
			} else {
				room.Thermostat.CurrentTemperature--
			}

			// If we reached the target after moving, stop the fan
			if room.Thermostat.CurrentTemperature == target {
				room.FanRunning = false
				room.mu.Unlock()
				return
			}

			room.mu.Unlock()
		}
	}(r)

	return nil
}

// StopFan stops the room fan.
// If the fan is not running, it returns "fan not running" as specified.
func (r *Room) StopFan() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.FanRunning {
		return fmt.Errorf("fan not running")
	}

	r.FanRunning = false
	return nil
}
