package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	library := make(map[int]string)

	for i := 0; i < n; i++ {
		scanner.Scan()
		line := scanner.Text()
		parts := strings.Fields(line)

		var command, title string
		var isbn int

		if len(parts) >= 2 {
			command = parts[0]
			isbn, _ = strconv.Atoi(parts[1])
			if len(parts) > 2 {
				title = strings.Join(parts[2:], " ")
			}
		}

		if command == "ADD" {
			library[isbn] = title
		} else if command == "REMOVE" {
			delete(library, isbn)
		}
	}
	isbns := make([]int, 0, len(library))
	titles := make([]string, 0, len(library))
	for key, value := range library {
		isbns = append(isbns, key)
		titles = append(titles, value)
	}
	for i := 0; i < len(isbns); i++ {
		for j := 0; j+1 < len(isbns); j++ {
			if titles[j] > titles[j+1] ||
				(titles[j] == titles[j+1] && isbns[j] > isbns[j+1]) {

				titles[j], titles[j+1] = titles[j+1], titles[j]
				isbns[j], isbns[j+1] = isbns[j+1], isbns[j]
			}
		}
	}
	for i := 0; i < len(isbns); i++ {
		fmt.Println(isbns[i])
	}
}
