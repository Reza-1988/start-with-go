package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSample(t *testing.T) {
	gotHour, gotMinute, gotSecond := ExtractTimeUnits(3600)
	assert.Equal(t, 1, gotHour)
	assert.Equal(t, 0, gotMinute)
	assert.Equal(t, 0, gotSecond)
	got := ConvertToDigitalFormat(2, 20, 2)
	assert.Equal(t, "02:20:02", got)
}
