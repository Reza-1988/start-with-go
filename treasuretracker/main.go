package main

func FindTreasure(ch chan interface{}) {
	maxQueryInterface := <-ch
	nInterface := <-ch
	mInterface := <-ch

	// We get the three initial numbers:
	_ = (maxQueryInterface).(int) // maxQuery
	n := (nInterface).(int)
	m := (mInterface).(int)

	// ask function to ask the judge:
	// We send a point (x, y) and get a string response.
	ask := func(x, y int) string {
		ch <- x
		ch <- y
		ansInterface := <-ch
		ans, _ := ansInterface.(string)
		return ans
	}

	// Starting point: (1,1) – must be inside the ground because n,m >= 1
	curX, curY := 1, 1
	// The first response is always "Go!", it doesn't matter to the algorithm; it just needs to be read.
	_ = ask(curX, curY)

	// --- Step 1: Find the treasure x (gx) with fixed y
	Lx, Rx := 1, n
	yFixed := curY // 1

	for Lx < Rx {
		mid := (Lx + Rx) / 2
		// First we go to (mid, yFixed) to make this point "previous".
		if curX != mid || curY != yFixed {
			_ = ask(mid, yFixed) // // We don't use the answer for comparison
			curX, curY = mid, yFixed
		}
		// Now we move to (mid+1, yFixed)
		// The answer to this move tells us what f(mid+1) is relative to f(mid).
		ans := ask(mid+1, yFixed)
		curX, curY = mid+1, yFixed
		// means the new point is closer, minimum on the right [mid+1..Rx]
		if ans == "$s$s$" {
			Lx = mid + 1
			// means the new point is further away, minimum in [Lx..mid]
		} else if ans == "s.s.s" {
			Rx = mid
			// "same as before" is not really seen in the logic of Manhattan distance for two adjacent points,
			// but just in case, a simple interval compaction:
		} else {
			if mid+1-Lx > Rx-mid {
				Lx = mid
			} else {
				Rx = mid + 1
			}
		}
	}

	gx := Lx // This is the treasure x

	// --- Step 2: Find y treasure (gy) with x fixed = gx
	Ly, Ry := 1, m
	xFixed := gx

	for Ly < Ry {
		mid := (Lx + Ry) / 2

		if curX != xFixed || curY != mid {
			_ = ask(xFixed, mid)
			curX, curY = xFixed, mid+1
		}

		ans := ask(xFixed, mid+1)
		curX, curY = xFixed, mid+1
		if ans == "$s$s$" {
			Ly = mid + 1
		} else if ans == "s.s.s" {
			Ry = mid
		} else {
			if mid+1-Ly > Ry-mid {
				Ly = mid
			} else {
				Ry = mid + 1
			}
		}
	}
	gy := Ly
	// --- Step 3: Declare the final answer ----------

	// Now that we are sure that (gx, gy) is the minimum distance point (i.e. the treasure),
	// We send this point to the judge once again.
	// If the judge is the goal, he immediately passes the test.
	_ = ask(gx, gy)

	// After this, judge will no longer send anything on ch and our goroutine will block,
	// but it doesn't matter for the test because judge will send the result on the res channel and the test will be finished.

}
