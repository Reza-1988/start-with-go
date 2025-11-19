package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type Person struct {
	x, y    int
	friends []int
}

func chooseNextFriend(current, dst int, people []Person) int {
	minDist := math.MaxInt64
	nearestFriend := -1
	for _, friendIdx := range people[current].friends {
		dx := people[friendIdx].x - people[dst].x
		dy := people[friendIdx].y - people[dst].y
		dist := dx*dx + dy*dy
		if dist < minDist {
			minDist = dist
			nearestFriend = friendIdx
		}
	}
	return nearestFriend
}

func milgramSimulator(people []Person, src, dst, money int) (bool, int) {
	current := src
	intermediates := 0

	if src == dst {
		return true, 0
	}
	for money > 0 {
		if current == dst {
			return true, intermediates
		}
		next := chooseNextFriend(current, dst, people)
		money--
		current = next
		if current != src && current != dst {
			intermediates++
		}
	}
	if current == dst {
		return true, intermediates
	}
	return false, intermediates
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

	deliveredCount := 0
	sumIntermediates := 0

	for i := 0; i < e; i++ {
		scanner.Scan()
		forthLine := strings.Fields(scanner.Text())

		fromName := forthLine[0]
		toName := forthLine[1]
		money, _ := strconv.Atoi(forthLine[2])

		from := nameToIdx[fromName]
		to := nameToIdx[toName]

		delivered, intermediates := milgramSimulator(people, from, to, money)

		if delivered {
			deliveredCount++
			sumIntermediates += intermediates
		}
	}
	if deliveredCount*2 > e {
		avg := float64(sumIntermediates) / float64(deliveredCount)
		fmt.Printf("%.2f\n", avg)
	} else {
		fmt.Println("This world isn't small!")
	}
}
