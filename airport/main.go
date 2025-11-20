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
	for i := 0; i < f; i++ {
		scanner.Scan()

	}

}
