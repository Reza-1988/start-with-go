package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	line := scanner.Text()
	p, _ := strconv.Atoi(line)

	var tax float64

	if p <= 100 {
		tax = float64(p) * 0.05
	} else if p <= 500 {
		tax = 100*0.05 + float64(p-100)*0.1
	} else if p <= 1000 {
		tax = 100*0.05 + 400*0.1 + float64(p-500)*0.15
	} else {
		tax = 100*0.05 + 400*0.1 + 500*0.15 + float64(p-1000)*0.2
	}

	fmt.Println(int(math.Floor(tax)))
}
