package main

import "gopkg.in/check.v1"

package main

func FindTreasure(ch chan interface{}) {

	// We get the three initial numbers:
	_ := (<- ch).(int) // maxQuery
	n := (<- ch).(int)
	m := (<- ch).(int)


	// ask function to ask the judge:
	// We send a point (x, y) and get a string response.
	ask := func(x, y int) string {
		ch <- x
		ch <- y
		ansInterface := <- ch
		ans, _ := ansInterface.(string)
		return ans
	}

	// Starting point: (1,1) – must be inside the ground because n,m >= 1
	curX, curY := 1, 1
	// The first response is always "Go!", it doesn't matter to the algorithm; it just needs to be read.
	_ := ask(curX, curY)

	// --- Step 1: Find the treasure x (tx) with fixed y
	Lx, Rx := 1, n
	yFixed := curY // 1

	for Lx < Rx {
		mid := (Lx + Rx) / 2
		// First we go to (mid, yFixed) to make this point "previous".
		if curX != mid || curY != yFixed {
			_ := ask(mid, yFixed) // // We don't use the answer for comparison
			curX, curY = mid, yFixed
		}
		// Now we move to (mid+1, yFixed)
		// The answer to this move tells us what f(mid+1) is relative to f(mid).
		ans := ask(mid+1, yFixed)
		curX, curY = mid+1, yFixed
		// // means the new point is closer, minimum on the right [mid+1..Rx]
		if ans == "$s$s$" {
			Lx = mid + 1
		} else if ans == "s.s.s" {
			Rx = mid
		} else {

		}
	}



}

