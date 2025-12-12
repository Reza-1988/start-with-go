package smart_thermostat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestControllerCreation(t *testing.T) {
	sc := NewSystemController()
	assert.NotNil(t, sc, "SystemController should be created successfully")
	assert.IsType(t, sc.Rooms, make(map[string]*Room), "Type Hint!")
}
func TestGetCurrentTemperature(t *testing.T) {
	room := &Room{
		ID: "123",
		Thermostat: Thermostat{
			CurrentTemperature: 23, // int
			TargetTemperature:  27, // int
		},
	}

	assert.Equal(
		t,
		room.Thermostat.CurrentTemperature,
		room.GetCurrentTemperature(),
	)
}
