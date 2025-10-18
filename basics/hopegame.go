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
	line := scanner.Text()

	parts := strings.Fields(line)

	p, _ := strconv.Atoi(parts[0])
	q, _ := strconv.Atoi(parts[1])

	for i := 1; i <= q; i++ {
		if i%p == 0 {
			count := i / p
			for j := 0; j < count; j++ {
				fmt.Print("Hope ")
			}
			fmt.Println()
		} else {
			fmt.Println(i)
		}
	}
}
