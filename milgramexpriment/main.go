package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Person struct {
	x, y    int
	friends []int
}
type Letter struct {
	from, to int
	money    int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	firstLine := strings.Fields(scanner.Text())
	p, _ := strconv.Atoi(firstLine[0])
	e, _ := strconv.Atoi(firstLine[1])

	people := make([]Person, p)
	nameToIdx := make(map[string]int)

	for i := 0; i < p; i++ {
		scanner.Scan()
		secondLine := strings.Fields(scanner.Text())
		name := secondLine[0]
		xComma := strings.TrimSuffix(secondLine[1], ",")
		x, _ := strconv.Atoi(xComma)
		y, _ := strconv.Atoi(secondLine[2])
		nameToIdx[name] = i
		people[i].x = x
		people[i].y = y
	}
	for i := 0; i < p; i++ {
		scanner.Scan()
		thirdLine := strings.Fields(scanner.Text())

		name := thirdLine[0]
		idx := nameToIdx[name]
		friendsIdx := make([]int, 0, len(thirdLine)-1)

		for _, friendName := range thirdLine[1:] {
			friendIdx := nameToIdx[friendName]
			friendsIdx = append(friendsIdx, friendIdx)
		}
		people[idx].friends = friendsIdx
	}
	letters := make([]Letter, e)
	for i := 0; i < e; i++ {
		scanner.Scan()
		forthLine := strings.Fields(scanner.Text())

		fromName := forthLine[0]
		toName := forthLine[1]
		money, _ := strconv.Atoi(forthLine[2])
		from := nameToIdx[fromName]
		to := nameToIdx[toName]
		letters[i] = Letter{
			from:  from,
			to:    to,
			money: money,
		}
	}
}
