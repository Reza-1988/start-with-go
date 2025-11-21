package main

import (
	"bufio"
	"fmt"
	"os"
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

func parseFlight(line string) Flight {

	layout := "Mon, Jan 2, 2006 3:04 PM"
	sepIdx := strings.Index(line, " => ")
	left := strings.TrimSpace(line[:sepIdx])
	right := strings.TrimSpace(line[sepIdx+4:])

	// from Tehran
	if strings.Contains(left, "Tehran(") {
		parts := strings.SplitN(left, " ", 2)
		plane := parts[0]
		rest := parts[1]
		openIdx := strings.Index(rest, "(")
		closeIdx := strings.Index(rest, ")")
		fromCity := strings.TrimSpace(rest[:openIdx])
		timeStr := strings.TrimSpace(rest[openIdx+1 : closeIdx])
		tehranTime, _ := time.Parse(layout, timeStr)
		toCity := strings.TrimSpace(right)
		return Flight{
			Plane:      plane,
			From:       fromCity,
			To:         toCity,
			TehranTime: tehranTime,
			FromTehran: true,
		}
		// to Tehran
	} else {

		parts := strings.SplitN(left, " ", 2)
		plane := parts[0]
		fromCity := parts[1]
		openIdx := strings.Index(right, "(")
		closeIdx := strings.Index(right, ")")
		toCity := strings.TrimSpace(right[:openIdx])
		timeStr := strings.TrimSpace(right[openIdx+1 : closeIdx])
		tehranTime, _ := time.Parse(layout, timeStr)
		return Flight{
			Plane:      plane,
			From:       fromCity,
			To:         toCity,
			TehranTime: tehranTime,
			FromTehran: false,
		}

	}
}
func main() {

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	p, _ := strconv.Atoi(scanner.Text())
	planeNameSpeed := make(map[string]int)

	for i := 0; i < p; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		speed, _ := strconv.Atoi(parts[1])
		planeNameSpeed[parts[0]] = speed
	}
	//
	scanner.Scan()
	c, _ := strconv.Atoi(scanner.Text())
	cityNameDistance := make(map[string]int)

	for i := 0; i < c; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		distance, _ := strconv.Atoi(parts[1])
		cityNameDistance[parts[0]] = distance
	}
	//
	scanner.Scan()
	f, _ := strconv.Atoi(scanner.Text())
	flights := make([]Flight, 0, f)
	for i := 0; i < f; i++ {
		scanner.Scan()
		line := scanner.Text()
		flight := parseFlight(line)
		flights = append(flights, flight)
	}
	scanner.Scan()
	todo := scanner.Text()

	if todo == "admin" {
		type Interval struct {
			Start time.Time
			End   time.Time
		}

		intervals := make([]Interval, 0, len(flights))

		for _, fl := range flights {
			var start, end time.Time
			if fl.FromTehran {
				start = fl.TehranTime.Add(-5 * time.Minute)
				end = fl.TehranTime.Add(5 * time.Minute)
			} else {
				start = fl.TehranTime.Add(-10 * time.Minute)
				end = fl.TehranTime.Add(5 * time.Minute)
			}
			intervals = append(intervals, Interval{Start: start, End: end})
		}
		// with AI help
		type Event struct {
			Time  time.Time
			Delta int
		}

		events := make([]Event, 0, len(intervals)*2)
		for _, iv := range intervals {
			events = append(events, Event{Time: iv.Start, Delta: +1})
			events = append(events, Event{Time: iv.End, Delta: -1})
		}

		sort.Slice(events, func(i, j int) bool {
			if events[i].Time.Equal(events[j].Time) {
				return events[i].Delta < events[j].Delta
			}
			return events[i].Time.Before(events[j].Time)
		})
		//
		current := 0
		maxRunways := 0
		for _, ev := range events {
			current += ev.Delta
			if current > maxRunways {
				maxRunways = current
			}
		}

		fmt.Println(maxRunways)
		return
	}

}
