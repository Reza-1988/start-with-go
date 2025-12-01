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
		// Make sure the judge compares (mid+1) with (mid).
		// The judge always compares the new query with the *previous* query.
		// So if we are not already at (mid, yFixed), we must move there first
		// to make it the "previous" point before checking (mid+1, yFixed).
		if curX != mid || curY != yFixed {
			_ = ask(mid, yFixed) // // We don't use the answer for comparison
			curX, curY = mid, yFixed
		}
		// We are inside the loop with a current search interval [Lx, Rx] on the x-axis.
		// We have already chosen:
		//     mid := (Lx + Rx) / 2
		// and (mid, yFixed) is now the "previous" point (we just asked it before).

		// Now we move to (mid+1, yFixed).
		// The judge will compare the distance of (mid+1, yFixed) to the distance of (mid, yFixed).
		// So the answer tells us whether f(mid+1) < f(mid), f(mid+1) > f(mid), or they are equal.
		// In other words: it tells us which side of the V-shaped distance function we are on.
		ans := ask(mid+1, yFixed)
		curX, curY = mid+1, yFixed

		// Case 1: "$s$s$"  → the new point (mid+1) is CLOSER than (mid).
		// That means: distance(mid+1) < distance(mid).
		// In a V-shaped function, this can only happen if the minimum (treasure x)
		// is on the RIGHT side of mid (or at mid+1, mid+2, ...).
		// So we can safely discard the LEFT part [Lx..mid], because the treasure
		// cannot be there anymore. The new search interval becomes [mid+1..Rx].
		if ans == "$s$s$" {
			Lx = mid + 1

			// Case 2: "s.s.s" → the new point (mid+1) is FURTHER than (mid).
			// That means: distance(mid+1) > distance(mid).
			// This means we are going away from the minimum when we move to the right,
			// so the minimum must be on the LEFT side (or exactly at mid).
			// Therefore, we can discard the RIGHT part (mid+1..Rx),
			// and our new search interval becomes [Lx..mid].
		} else if ans == "s.s.s" {
			Rx = mid

			// Case 3: "same as before" → distance(mid+1) == distance(mid).
			// For two adjacent points in a Manhattan-distance V-shape, this is unusual,
			// but we handle it just in case.
			// Here we simply "shrink" the interval from the longer side to keep progress:
			// - If the left side [Lx..mid] is longer, move Lx closer to mid.
			// - Otherwise, move Rx closer to mid+1.
			// This keeps reducing [Lx..Rx] until we eventually find a single point.
		} else {
			if mid+1-Lx > Rx-mid {
				// Left side is longer → move Lx towards mid
				Lx = mid
			} else {
				// Right side is longer → move Rx towards mid+1
				Rx = mid + 1
			}
		}
	// When the loop ends we have Lx == Rx, meaning the search interval
	// has narrowed down to a single x value. This unique point is the
	// minimum of the V-shaped distance function, so Lx is exactly the
	// treasure's x-coordinate.
	gx := Lx

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
