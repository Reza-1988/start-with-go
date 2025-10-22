package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	s := scanner.Text()

	temp := ""
	total := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			temp += string(r)
		} else if temp != "" {
			num, _ := strconv.Atoi(temp)
			total += num
			temp = ""
		}
	}
	if temp != "" {
		num, _ := strconv.Atoi(temp)
		total += num
	}
	totalStr := strconv.Itoa(total)
	length := len(totalStr)

	if total == 0 {
		fmt.Println("YES")
		return
	}
	powerSum := 0
	number := total
	for number > 0 {
		t := number % 10
		pow := 1
		for i := 0; i < length; i++ {
			pow *= t
		}
		powerSum += pow
		number = number / 10
	}
	if powerSum == total {
		fmt.Println("YES")
	} else {
		fmt.Println("NO")
	}
}
