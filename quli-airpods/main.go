package main

import (
	"fmt"
	"sync"
)

// We define the possible states as constants.
const (
	StateDocked       = "Docked"
	StateConnected    = "Connected"
	StateDisconnected = "Disconnected"
)

// Airpod
// Think of one AirPod (only left or only right)
//  1. It has state:
//     - "Docked", It means Airpod is inside the case.
//     - "Connected", It means Airpod is outside the case + connected to phone
//     - "Disconnected", It means Airpod is outside the case + not connected to phone
//  2. It has speaker channel:
//     - This is `chan byte`, whenever the Airpod should play sound, we send bytes into this channel.
//     - And another part of program can read from channel to hear the sound"
//  3. We also need a lock:
//     - Why lock? Methods like `GetState()`may be called at the same time as `Dock` or `Undock`
//     - Without lock, race condition happened.
type Airpod struct {
	mu    sync.RWMutex
	state string
	ch    chan byte
}

// AirpodCase
// The AirpodCase is the case that holds both Airpods and talks to the phone.
//  1. It must store two Airpods (left + right)
//  2. It must remember Bluetooth's is connected or not.
//  3. It must keep the channel from the phone
//     - The phone sends bytes on this channel and the case reads these bytes and forwards them to connected Airpods.
//  4. Also needs a lock(`sync.Mutex`) because:
//     - Many goroutines may call `DockLeft`, `UndockRight`, `ConnectBluetooth`, etc. at the same time.
//     -  We must protect: `btConnected`, `phoneCh`, some operations that involve both Airpods together.
type AirpodCase struct {
	mu          sync.Mutex
	left        *Airpod
	right       *Airpod
	btConnected bool
	phoneCh     chan byte
}

// --- Why this Structure design is good?
//	1. Airpod is simple and self-contained:
// 		- It knows its own state.
//		- It knows its own audio channel.
//		- It keeps its own lock for state (mu).
// 		- Using sync.RWMutex in Airpod:
//			- Many readers (GetState) can run at the same time.
//			- Writers (changing the state) use full Lock/Unlock.
//   2. AirpodCase is the manager:
// 		- It holds both AirPods.
// 		- It knows if Bluetooth is on.
// 		- It has the phoneCh from the phone.
// 		- It will later start a goroutine to read from phoneCh and send data to connected AirPods.
// 		- Using sync.Mutex in AirpodCase:
// 			- All case-level operations (ConnectBluetooth, DockLeft, UndockRight, etc.) can safely change shared fields.
// --- End

// NewAirpodCase
// Some usefully notes for start new AirpodCase struck:
//  1. You don't need to initialize Mutex in struct e.g. `mu: sync.Mutex()` because:
//     - In Go, the zero value of a `sync.Mutex` is already a valid mutex.
//     - So you don’t need to set it manually.
//  2. For Initialize the `ch` for each Airpod: `ch: make(chan byte)`:
//     - Unbuffered channels are OK, but in this problem they can easily cause blocking and weird deadlocks,
//     - especially with multiple goroutines and two AirPods.
//     - Safer to use a buffered channel,This gives some room for audio bytes without blocking immediately.
//  3. `phoneCh` should not be created here because:
//     - The problem statement says: `ConnectBluetooth(ch chan byte)` receives a channel from the phone. So:
//     - The phone creates the channel. The case stores that channel.
//     - Therefore:
//     - In `NewAirpodCase`, phoneCh should be nil.
//     - In `ConnectBluetooth`, we will set c.phoneCh = ch.
func NewAirpodCase() *AirpodCase {
	return &AirpodCase{
		// mu: zero value is fine, no need to set
		left: &Airpod{
			// mu: zero value is fine
			state: StateDocked,
			ch:    make(chan byte, 1024),
		},
		right: &Airpod{
			state: StateDocked,
			ch:    make(chan byte, 1024),
		},
		btConnected: false,
		phoneCh:     nil, // will be set in ConnectBluetooth
	}
}

func (c *AirpodCase) GetRightAirpod() *Airpod {
	return c.right
}

func (c *AirpodCase) GetLeftAirpod() *Airpod {
	return c.left
}

// GetState
// Why do we use a lock in GetState()?
//   - Because many goroutines may read the state at the same time while another goroutine is changing it.
//   - Examples of functions that change the state: `UndockLeft`, `UndockRight`, `DockLeft`, `DockRight`, `ConnectBluetooth`
//   - So at any moment, One goroutine might set `status = "Connected"`, Another goroutine might call `GetState() to read the state.
//   - Without Lock race condition is happened.
//
// Why use to `RLock()`/`RULock()`:
//   - This is a read lock, It allows multiple readers at the same time, it blocks writers and safe for raad-only operations.
//   - `Lock()` / `Unlock()`:
//   - Use for operations that change data, Only one goroutine can hold the lock, This prevents conflict when writing
func (a *Airpod) GetState() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// UndockLeft
// 1. For lock must use Airpod lock because we want to lock Airpod states, and it's belong to Airpod not AirpodCase.
//   - `c.mu` is meant to protect case-level fields like: `btConnected`, `phoneCh`
//   - The state of each Airpod should be protected bu its own `mu`.
//
// 2. From beginning, we must use lock for read or write
//   - if  read `c.left.state` before you lock anything. This is not safe if another goroutine can change it at the same time.
func (c *AirpodCase) UndockLeft() *Airpod {
	c.left.mu.Lock()
	defer c.left.mu.Unlock()
	//
	if c.left.state != StateDocked {
		return nil
	}
	// Now we know it was Docked, so we undock it.
	if c.btConnected {
		c.left.state = StateConnected
	} else {
		c.left.state = StateDisconnected
	}
	return c.left
}

func (c *AirpodCase) UndockRight() *Airpod {
	c.right.mu.Lock()
	defer c.right.mu.Unlock()
	//
	if c.right.state != StateDocked {
		return nil
	}
	//
	if c.btConnected {
		c.right.state = StateConnected
	} else {
		c.right.state = StateDisconnected
	}
	return c.right
}

func (c *AirpodCase) DockLeft() error {
	c.left.mu.Lock()
	defer c.left.mu.Unlock()
	//
	if c.left.state == StateDocked {
		return fmt.Errorf("left Airpod already is in case")
	}
	c.left.state = StateDocked
	return nil
}
func (c *AirpodCase) DockRight() error {
	c.right.mu.Lock()
	defer c.right.mu.Unlock()
	//
	if c.right.state == StateDocked {
		return fmt.Errorf("left Airpod already is in case")
	}
	c.right.state = StateDocked
	return nil
}

func (a *Airpod) GetChannel() chan byte {
	return a.ch
}

// ConnectBluetooth
// What does ConnecBluetooth need to do?
//  1. Prevent connecting twice
//     - If Bluetooth is already connected, return an error.
//     - But we must check and change `btConnected` under `c.mu` lock to avoid race conditions.
//  2. Save the phone channel
//     - We get a channel from the phone : `ch chan byte`, This is the channel where audio comes in.
//     - We must store it inside AirpodCase channel: `c.phoneCh = ch`
//  3. Change Airpods state (Disconnected to Connecte)
//     - When Bluetooth connects, Airpods that are Disconnected become connected, so:
//     - If left Airpod state is `StateDisconnected`, set to `StateConnected`
//     - Same Rule for right Airpod
//     - We must lock each Airpod `c.left.mu.Lock()`/`c.right.mu.Lock` while changing its state.
//  4. Start a goroutine to route audio
//     - A goroutine waits for bytes from `ch` (the phone). Every time it gets one byte:
//     - It check which Airpods are `Connected`
//			- Use RLock() when you get states of each Airpod.
//     - It sends that byte to their channels if the states of each Airpod is StateConnected.
func (c *AirpodCase) ConnectBluetooth(ch chan byte) error {
	c.mu.Lock()

	// 1. If already connected, return error
	if c.btConnected {
		c.mu.Unlock()
		return fmt.Errorf("bluetooth already connected")
	}

	// Mark as connected
	c.btConnected = true
	c.phoneCh = ch
	c.mu.Unlock()
	// 2. Update airpods that were disconnected
	// LEFT
	c.left.mu.Lock()
	if c.left.state == StateDisconnected {
		c.left.state = StateConnected
	}
	c.left.mu.Unlock()
	// Right
	c.right.mu.Lock()
	if c.right.state == StateDisconnected {
		c.right.state = StateConnected
	}
	c.right.mu.Unlock()
	// 3. Start audio forwarding goroutine
	go func() {
		for b := range ch {
			// Check left side state
			c.left.mu.RLock()
			leftConnected := c.left.state == StateConnected
			c.left.mu.RUnlock()
			// Check right state
			c.right.mu.RLock()
			rightConnected := c.right.state == StateConnected
			c.right.mu.RUnlock()
			// Send audio
			if leftConnected {
				c.left.ch <- b
			}
			if rightConnected {
				c.right.ch <- b
			}
		}
	}()
	return nil
}
