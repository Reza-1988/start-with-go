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
	line1 := scanner.Text()
	scanner.Scan()
	line2 := scanner.Text()
	parts := strings.Fields(line1)
	w, _ := strconv.Atoi(parts[0])
	sh, _ := strconv.Atoi(parts[1])
	p, _ := strconv.ParseFloat(line2, 64)

	everyDaySalary := float64(w)*(1-p) + (float64(w)-float64(sh))*p
	wholeYearSalary := 365.0 * everyDaySalary
	fmt.Printf("%.2f\n", wholeYearSalary)
}
