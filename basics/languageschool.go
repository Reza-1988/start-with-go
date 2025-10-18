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

	for i := 0; i < n; i++ {
		scanner.Scan()
		teacher := scanner.Text()

		scanner.Scan()
		scores := strings.Fields(scanner.Text())

		sum := 0
		for _, s := range scores {
			num, _ := strconv.Atoi(s)
			sum += num
		}
		avg := float64(sum) / float64(len(scores))

		var result string
		if avg >= 80 {
			result = "Excellent"
		} else if avg >= 60 && avg < 80 {
			result = "Good"
		} else if avg >= 40 && avg < 60 {
			result = "Very Good"
		} else {
			result = "Fair"
		}
		fmt.Println(teacher, result)
	}

}
