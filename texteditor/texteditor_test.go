package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test1(t *testing.T) {
	e := NewTwoStackEditor("my awesome code")
	e.CursorRight()
	e.CursorRight()
	e.CursorLeft()

	text := e.WholeText()
	assert.Equal(t, "my awesome code", text)

	e.InsertRune('آ')
	text = e.WholeText()
	assert.Equal(t, "mآy awesome code", text)

	r, _ := e.RightStack.Top()
	assert.Equal(t, 'y', r)

	l, _ := e.LeftStack.Top()
	assert.Equal(t, 'آ', l)

	e.DeleteBackward()
	e.DeleteForward()
	e.InsertString("abc")
}
