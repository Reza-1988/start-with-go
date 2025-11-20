package main

import (
	"bufio"
	"os"
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

}
