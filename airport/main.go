package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Flight struct {
	Plane      string
	From       string
	To         string
	TehranTime time.Time
	FromTehran bool
}

const timeLayout = "Mon, Jan 2, 2006 3:04 PM"

var (
	// Regex for flights departing FROM Tehran:
	// Airbus-320 Tehran(Mon, Apr 7, 2025 7:36 PM) => Bandar-abbas
	// Pattern: "<Plane> Tehran(<DateTime>) => <City>"
	// Capturing groups:
	//   m[1] = plane name  (e.g., "Airbus-320")
	//   m[2] = datetime inside parentheses
	//   m[3] = destination city
	//
	// Explanation:
	// ^(\S+)               → plane name (non-space characters)
	// \s+                 → one or more spaces
	// Tehran\(            → literal "Tehran("
	// (.+)                → full datetime (can contain spaces, commas, AM/PM)
	// \)                  → closing parenthesis
	// \s+=>\s+            → separator " => "
	// (\S+)               → destination city name (non-space)
	// \s*$                → optional trailing spaces, end of line

	reFromTehran = regexp.MustCompile(`^(\S+)\s+Tehran\((.+)\)\s+=>\s+(\S+)\s*$`)
	// Regex for flights arriving TO Tehran:
	// Boeing-737 Tabriz => Tehran(Sun, Apr 6, 2025 3:24 PM)
	// Pattern: "<Plane> <City> => Tehran(<DateTime>)"
	// Capturing groups:
	//   m[1] = plane name
	//   m[2] = origin city
	//   m[3] = datetime inside parentheses
	//
	// Explanation:
	// ^(\S+)               → plane name
	// \s+(\S+)            → origin city
	// \s+=>\s+            → separator " => "
	// Tehran\(            → literal "Tehran("
	// (.+)                → datetime
	// \)\s*$              → closing parenthesis + end of line

	reToTehran = regexp.MustCompile(`^(\S+)\s+(\S+)\s+=>\s+Tehran\((.+)\)\s*$`)
)

// parseFlight parses one line of input and determines whether the flight
// is:
//  1. Departing from Tehran:   "<Plane> Tehran(<Time>) => <City>"
//  2. Arriving to Tehran:      "<Plane> <City> => Tehran(<Time>)"
//
// It uses regex to extract:
//   - plane name
//   - city (origin or destination)
//   - the exact timestamp string inside parentheses
//
// FindStringSubmatch returns:
//
//	m[0] = full matched string
//	m[1], m[2], ... = captured groups inside parentheses
//
// Based on which regex matches, we construct a Flight object and mark
// whether it is an outbound flight (FromTehran = true) or inbound.
func parseFlight(line string) Flight {
	line = strings.TrimSpace(line)

	if m := reFromTehran.FindStringSubmatch(line); m != nil {
		plane := m[1]
		timeStr := m[2]
		toCity := m[3]

		// Convert the extracted datetime string into time.Time.
		// The layout must match EXACTLY the format used in input:
		//   "Mon, Jan 2, 2006 3:04 PM"
		tehranTime, _ := time.Parse(timeLayout, timeStr)

		return Flight{
			Plane:      plane,
			From:       "Tehran",
			To:         toCity,
			TehranTime: tehranTime,
			FromTehran: true,
		}
	}

	if m := reToTehran.FindStringSubmatch(line); m != nil {
		plane := m[1]
		fromCity := m[2]
		timeStr := m[3]

		tehranTime, _ := time.Parse(timeLayout, timeStr)

		return Flight{
			Plane:      plane,
			From:       fromCity,
			To:         "Tehran",
			TehranTime: tehranTime,
			FromTehran: false,
		}
	}
	return Flight{}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	pLine := strings.TrimSpace(scanner.Text())
	p, _ := strconv.Atoi(pLine)

	planeNameSpeed := make(map[string]int, p)
	for i := 0; i < p; i++ {
		scanner.Scan()
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		speed, _ := strconv.Atoi(parts[1])
		planeNameSpeed[parts[0]] = speed
	}

	scanner.Scan()
	cLine := strings.TrimSpace(scanner.Text())
	c, _ := strconv.Atoi(cLine)

	cityNameDistance := make(map[string]int, c)
	for i := 0; i < c; i++ {
		scanner.Scan()
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		distance, _ := strconv.Atoi(parts[1])
		cityNameDistance[parts[0]] = distance
	}

	scanner.Scan()
	fLine := strings.TrimSpace(scanner.Text())
	f, _ := strconv.Atoi(fLine)

	flights := make([]Flight, 0, f)
	for i := 0; i < f; i++ {
		scanner.Scan()
		line := scanner.Text()
		fl := parseFlight(line)
		flights = append(flights, fl)
	}

	scanner.Scan()
	todo := strings.TrimSpace(scanner.Text())

	// Using sweep-line algorithm to calculate the minimum number of runways needed.
	// Each flight occupies the runway during a specific time interval in Tehran.
	// We convert every interval into two timeline events:
	//   - start time  → +1  (runway becomes occupied)
	//   - end time    → -1  (runway becomes free)
	// After sorting the events by time, we scan through them and track
	// the maximum number of simultaneously occupied runways.

	if todo == "admin" {
		// A timeline event representing a change in required runways.
		type Event struct {
			t     time.Time // moment on the timeline
			delta int       // +1 = runway needed, -1 = runway released
		}

		// Collect all events (two per flight).
		events := make([]Event, 0, len(flights)*2)

		for _, fl := range flights {
			// Skip invalid or unparsable flights.
			if fl.TehranTime.IsZero() {
				continue
			}

			// Compute the exact interval during which the runway is occupied.
			// Outbound flights (from Tehran):   [-5 min, +5 min]
			// Inbound flights (to Tehran):     [-10 min, +5 min]
			var start, end time.Time
			if fl.FromTehran {
				start = fl.TehranTime.Add(-5 * time.Minute)
				end = fl.TehranTime.Add(5 * time.Minute)
			} else {
				start = fl.TehranTime.Add(-10 * time.Minute)
				end = fl.TehranTime.Add(5 * time.Minute)
			}

			// Convert the interval into two events.
			events = append(events,
				Event{t: start, delta: +1}, // runway becomes occupied
				Event{t: end, delta: -1},   // runway becomes free
			)
		}

		// Sort all events on the timeline:
		// 1) Primary key: time (earlier events first)
		// 2) Tie-breaker: if two events happen at the same time,
		//    process +1 (runway becomes occupied) BEFORE -1 (runway becomes free).
		//    This ensures that if one flight starts using the runway at the exact
		//    moment another one stops, they are both counted as overlapping.
		sort.Slice(events, func(i, j int) bool {
			if events[i].t.Equal(events[j].t) {
				return events[i].delta > events[j].delta // +1 before -1
			}
			return events[i].t.Before(events[j].t)
		})

		// Sweep over the sorted events and track how many runways are in use.
		// 'current' is the number of occupied runways at the current moment.
		// 'maxRunways' is the maximum value of 'current' over the whole timeline.
		current := 0
		maxRunways := 0

		for _, e := range events {
			current += e.delta
			if current > maxRunways {
				maxRunways = current
			}
		}

		fmt.Println(maxRunways)
		return
	}

	city := todo

	for _, fl := range flights {
		if fl.From == "Tehran" && fl.To == city {
			distance := cityNameDistance[city]
			speed := planeNameSpeed[fl.Plane]

			estimatedSec := 3600.0 * float64(distance) / float64(speed)
			duration := time.Duration(estimatedSec) * time.Second

			arrivalTime := fl.TehranTime.Add(duration)

			fmt.Printf("%s Tehran => %s(%s)\n",
				fl.Plane, city, arrivalTime.Format(timeLayout))

		} else if fl.From == city && fl.To == "Tehran" {
			distance := cityNameDistance[city]
			speed := planeNameSpeed[fl.Plane]

			estimatedSec := 3600.0 * float64(distance) / float64(speed)
			duration := time.Duration(estimatedSec) * time.Second

			departTime := fl.TehranTime.Add(-duration)

			fmt.Printf("%s %s(%s) => Tehran\n",
				fl.Plane, city, departTime.Format(timeLayout))
		}
	}
}
