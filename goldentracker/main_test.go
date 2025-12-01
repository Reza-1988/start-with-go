package main

import (
	"fmt"
	"math"
	"testing"
	"time"
)

package main

import (
"errors"
"fmt"
"math"
"testing"
"time"
)

type point struct {
	X int
	Y int
}

const (
	firstTimeMsg    = "Go!"
	closerMsg       = "$s$s$"
	furtherMsg      = "s.s.s"
	sameAsBeforeMsg = "same as before"
	outOfRangeMsg   = "invalid point"
)

func validatePoint(boarder, pt point) error {
	if pt.X <= 0 || pt.X > boarder.X {
		return errors.New(outOfRangeMsg)
	}
	if pt.Y <= 0 || pt.Y > boarder.Y {
		return errors.New(outOfRangeMsg)
	}
	return nil
}

func calculateDistance(goal, pt point) int {
	dist := math.Abs(float64(goal.X-pt.X)) + math.Abs(float64(goal.Y-pt.Y))
	return int(dist)
}

func findMaxQuery(x, y int) int {
	maxQuery := int(math.Log2(float64(x)) + math.Log2(float64(y)) + 2) * 3
	// println("max query is : ", maxQuery)
	return maxQuery
}

func judge(goal, border point, maxQuery int, res chan error) {
	lastDistance := int(-1)
	ch := make(chan interface{})
	go FindTreasure(ch)

	// maxQuery
	ch <- maxQuery

	// first msg
	ch <- border.X
	ch <- border.Y



	for i := 0; i < maxQuery; i++ {
		pt := new(point)
		xInterface := <-ch
		yInterface := <-ch

		var xOk, yOk bool
		pt.X, xOk = xInterface.(int)
		pt.Y, yOk = yInterface.(int)
		if !xOk || !yOk {
			res <- errors.New("cast error, you must write two integers for x and y!")
			return
		}

		if *pt == goal {
			res <- nil
			return
		}

		if err := validatePoint(border, *pt); err != nil {
			res <- err
			return
		}

		newDistance := calculateDistance(goal, *pt)
		if lastDistance == -1 {
			ch <- firstTimeMsg
		} else if newDistance < lastDistance {
			ch <- closerMsg
		} else if newDistance > lastDistance {
			ch <- furtherMsg
		} else {
			ch <- sameAsBeforeMsg
		}
		lastDistance = newDistance
	}

	res <- fmt.Errorf("max query number reached, max query number is: %d", maxQuery)
}

func TestSimple(t *testing.T) {
	goal := point{
		X: 120,
		Y: 120,
	}
	boarder := point{
		X: 220,
		Y: 220,
	}
	maxQuery := findMaxQuery(boarder.X, boarder.Y)
	res := make(chan error)

	go judge(goal, boarder, maxQuery, res)

	select {
	case <-time.After(2 * time.Second):
		t.Errorf("timeout reached!")

	case err := <-res:
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	fmt.Println("Test passed successfully!")
}

func TestSmall(t *testing.T) {
	goal := point{
		X: 2,
		Y: 3,
	}
	boarder := point{
		X: 3,
		Y: 3,
	}
	maxQuery := findMaxQuery(boarder.X, boarder.Y)
	res := make(chan error)

	go judge(goal, boarder, maxQuery, res)

	select {
	case <-time.After(2 * time.Second):
		t.Errorf("timeout reached!")

	case err := <-res:
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	fmt.Println("Test passed successfully!")
}
