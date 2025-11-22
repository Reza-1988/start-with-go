package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
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
	// Airbus-320 Tehran(Mon, Apr 7, 2025 7:36 PM) => Bandar-abbas
	reFromTehran = regexp.MustCompile(`^(\S+)\s+Tehran\((.+)\)\s+=>\s+(\S+)\s*$`)
	// Boeing-737 Tabriz => Tehran(Sun, Apr 6, 2025 3:24 PM)
	reToTehran = regexp.MustCompile(`^(\S+)\s+(\S+)\s+=>\s+Tehran\((.+)\)\s*$`)
)

func parseFlight(line string) Flight {
	line = strings.TrimSpace(line)

	if m := reFromTehran.FindStringSubmatch(line); m != nil {
		plane := m[1]
		timeStr := m[2]
		toCity := m[3]

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
	// with AI
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

		maxRunways := 0

		for i := 0; i < len(intervals); i++ {
			current := 0
			for j := 0; j < len(intervals); j++ {
				if intervals[i].Start.Before(intervals[j].End) &&
					intervals[j].Start.Before(intervals[i].End) {
					current++
				}
			}
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
