package main

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"
)

// A simple struct to hold a point on the screen:
//
//		  X represents the row.
//	   Y represents the column.
type point struct {
	X int
	Y int
}

// These are the strings that the judge on the channel returns to the FindTreasure function
const (
	firstTimeMsg    = "Go!"            // A message that is returned only for the first query.
	closerMsg       = "$s$s$"          // This means that the new point is closer to the treasure than the previous point.
	furtherMsg      = "s.s.s"          // This means that the new point is further away from the treasure than the previous one.
	sameAsBeforeMsg = "same as before" // This means that the distance from the new point to the treasure is equal to the previous distance.
	outOfRangeMsg   = "invalid point"  // Error text used when the point is outside the bounds of the terrain.
)

// Note: In FindTreasure, you must read these same strings and compare them with these same values
// (you only consume these messages as a client, you do not send them).

// validatePoint checks whether the proposed point is inside the terrain or not:
// @param boarder is land dimensions
func validatePoint(boarder, pt point) error {
	if pt.X <= 0 || pt.X > boarder.X {
		return errors.New(outOfRangeMsg)
	}
	if pt.Y <= 0 || pt.Y > boarder.Y {
		return errors.New(outOfRangeMsg)
	}
	return nil
}

// calculateDistance calculates the Manhattan distance between the treasure point and the proposed point
// This distance is what the judge uses to determine whether the new point is closer or farther away.
func calculateDistance(goal, pt point) int {
	dist := math.Abs(float64(goal.X-pt.X)) + math.Abs(float64(goal.Y-pt.Y))
	return int(dist)
}

// findMaxQuery Calculates a maximum number of queries allowed (maxQuery) for an x × y page.
// This number is large enough for a good algorithm to find the treasure in this number of questions.
func findMaxQuery(x, y int) int {
	maxQuery := int(math.Log2(float64(x))+math.Log2(float64(y))+2) * 3
	// println("max query is : ", maxQuery)
	return maxQuery
}

// It is the judge/simulator that plays the role of the examination.
// @param goal, The point where the treasure is located.
// @param border, The dimensions of the terrain (maximum X and Y).
// @param maxQuery, The maximum number of queries allowed.
// @param res chan error,  The channel that sends the final test result (nil means successful).
func judge(goal, border point, maxQuery int, res chan error) {
	lastDistance := int(-1)
	ch := make(chan interface{})
	// Run the FindTreasure function (you) in a separate goroutine and give it the same channel.
	// From this point on, judge and FindTreasure talk to each other on ch.
	go FindTreasure(ch)

	// maxQuery
	ch <- maxQuery

	// first msg
	ch <- border.X
	ch <- border.Y

	// This loop handles one query from FindTreasure at a time.
	for i := 0; i < maxQuery; i++ {
		pt := new(point)
		xInterface := <-ch
		yInterface := <-ch

		// Type conversion and type checking, This means that the FindTreasure has written something else (e.g., string) on the channel.
		var xOk, yOk bool
		pt.X, xOk = xInterface.(int)
		pt.Y, yOk = yInterface.(int)
		if !xOk || !yOk {
			res <- errors.New("cast error, you must write two integers for x and y!")
			return
		}

		// If the FindTreasure's proposed point is exactly equal to the goal coordinates:
		//The value nil (no error) is sent to res: i.e. the test passes.
		//And judge ends.
		if *pt == goal {
			res <- nil
			return
		}

		// Checking whether you are inside/outside the land boundary
		if err := validatePoint(border, *pt); err != nil {
			res <- err
			return
		}

		// Calculate the new distance and send the appropriate message.
		// The distance from Manhattan to the new point is calculated as the treasure.
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

	// If the ring is finished and the treasure is not found
	res <- fmt.Errorf("max query number reached, max query number is: %d", maxQuery)
	// If the for loop i := 0; i < maxQuery is executed completely and none of the questions have a suggested point equal to goal:
	// This means that the FindTreasure has asked too many questions and has not yet found the treasure.
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
