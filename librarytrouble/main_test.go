package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSample(t *testing.T) {
	lib := NewLibrary(100)
	res := lib.AddBook("golestan")
	assert.Equal(t, "OK", res)

	res = lib.BorrowBook("Golestan", "Ali")
	assert.Equal(t, "OK", res)

	res = lib.ReturnBook("golestan")
	assert.Equal(t, "OK", res)
}
