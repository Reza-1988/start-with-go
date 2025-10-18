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
	n, _ := strconv.Atoi(scanner.Text())

	var guide = make(map[string]string)
	for i := 0; i < n; i++ {
		scanner.Scan()
		line := scanner.Text()
		parts := strings.Fields(line)
		guide[parts[1]] = parts[0]
	}
	scanner.Scan()
	q, _ := strconv.Atoi(scanner.Text())
	countries := make([]string, 0, q)
	for i := 0; i < q; i++ {
		scanner.Scan()
		line := scanner.Text()
		prefix := line[:3]
		if v, k := guide[prefix]; k {
			countries = append(countries, v)
		} else {
			countries = append(countries, "Invalid Number")

		}
	}
	for i := 0; i < q; i++ {
		fmt.Println(countries[i])
	}
}
