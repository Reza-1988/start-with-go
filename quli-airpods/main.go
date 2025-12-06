package main

import "sync"

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
// 	1. It must store two Airpods (left + right)
// 	2. It must remember Bluetooth's is connected or not.
// 	3. It must keep the channel from the phone
//     - The phone sends bytes on this channel and the case reads these bytes and forwards them to connected Airpods.
// 	4. Also needs a lock(`sync.Mutex`) because:
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
//	1. You don't need to initialize Mutex in struct e.g. `mu: sync.Mutex()` because:
//  	- In Go, the zero value of a `sync.Mutex` is already a valid mutex.
// 		- So you don’t need to set it manually.
// 	2. For Initialize the `ch` for each Airpod: `ch: make(chan byte)`:
// 		- Unbuffered channels are OK, but in this problem they can easily cause blocking and weird deadlocks,
//		- especially with multiple goroutines and two AirPods.
// 		- Safer to use a buffered channel,This gives some room for audio bytes without blocking immediately.
// 	3. `phoneCh` should not be created here because:
// 		- The problem statement says: `ConnectBluetooth(ch chan byte)` receives a channel from the phone. So:
// 			- The phone creates the channel. The case stores that channel.
// 			- Therefore:
// 				- In `NewAirpodCase`, phoneCh should be nil.
// 				- In `ConnectBluetooth`, we will set c.phoneCh = ch.
func NewAirpodCase() *AirpodCase {
	return &AirpodCase{
		left: &Airpod{
			state: "Docked",
			ch:    make(chan byte, 1024),
		},
		right: &Airpod{
			state: StateDocked,
			ch:    make(chan byte, 1024),
		},
		btConnected: false,
		phoneCh:     nil,
	}
}

func (a *AirpodCase) GetRightAirpod() *Airpod {
	return nil
}

func (a *AirpodCase) GetLeftAirpod() *Airpod {
	return nil

}
func (a *Airpod) GetState() string {
	return ""
}

func (a *AirpodCase) UndockLeft() *Airpod {
	return nil
}

func (a *AirpodCase) UndockRight() *Airpod {
	return nil
}

func (a *AirpodCase) DockLeft() error {
	return nil

}
func (a *AirpodCase) DockRight() error {
	return nil
}

func (a *Airpod) GetChannel() chan byte {
	return nil
}

func (c *AirpodCase) ConnectBluetooth(ch chan byte) error {
	return nil
}
