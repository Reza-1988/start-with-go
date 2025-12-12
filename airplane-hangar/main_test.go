package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var idGen int

func nextID() int {
	idGen++
	return idGen
}

func CreateOldNest(aircraftCount int) *Nest {
	nest := &Nest{
		ID:        nextID(),
		Aircrafts: []*Aircraft{},
	}
	for i := 0; i < aircraftCount; i++ {
		a := &Aircraft{
			ID:     nextID(),
			Cabin:  &Cabin{ID: nextID()},
			Wheels: []*Wheel{{ID: nextID()}, {ID: nextID()}, {ID: nextID()}},
			Wings: []*Wing{
				{ID: nextID(), Engine: &Engine{ID: nextID()}},
				{ID: nextID(), Engine: &Engine{ID: nextID()}},
			},
		}
		nest.Aircrafts = append(nest.Aircrafts, a)
	}
	return nest
}

func TestSimple(t *testing.T) {
	idGen = 0
	oldNest := CreateOldNest(1)
	newNest := CreateOldNest(2)

	newNest.Aircrafts[1].Cabin = nil

	newNest.Aircrafts[0].Wings[0].Engine = nil
	newNest.Aircrafts[1].Wings[0] = nil
	newNest.Aircrafts[1].Wings[1] = nil

	newNest.Aircrafts[0].Wheels[0] = nil
	newNest.Aircrafts[0].Wheels[1] = nil
	newNest.Aircrafts[1].Wheels[0] = nil
	newNest.Aircrafts[1].Wheels[1] = nil

	Repair(oldNest, newNest)

	assert.Nil(t, oldNest.Aircrafts[0].Cabin)
	assert.NotNil(t, newNest.Aircrafts[1].Cabin)

	assert.Nil(t, oldNest.Aircrafts[0].Wheels[0])
	assert.Nil(t, oldNest.Aircrafts[0].Wheels[1])
	assert.Nil(t, oldNest.Aircrafts[0].Wheels[2])

	assert.NotNil(t, newNest.Aircrafts[0].Wheels[0])
	assert.NotNil(t, newNest.Aircrafts[0].Wheels[1])
	assert.NotNil(t, newNest.Aircrafts[1].Wheels[0])
	assert.Nil(t, newNest.Aircrafts[1].Wheels[1])

	// at first we solve wings problem for all aircrafts
	assert.NotNil(t, newNest.Aircrafts[0].Wings[0])
	assert.NotNil(t, newNest.Aircrafts[1].Wings[0])
	assert.NotNil(t, newNest.Aircrafts[1].Wings[1])
	// and then there is no engine for this one
	assert.Nil(t, newNest.Aircrafts[0].Wings[0].Engine)
}
