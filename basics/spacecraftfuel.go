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

	names := make([]string, n)
	counts := make([]int, n)

	for i := 0; i < n; i++ {
		scanner.Scan()
		line := scanner.Text()
		parts := strings.Fields(line)

		names[i] = parts[0]

		fuels := make([]int, 0, len(parts)-1)
		for _, part := range parts[1:] {
			num, _ := strconv.Atoi(part)
			fuels = append(fuels, num)
		}

		if len(fuels) < 3 {
			counts[i] = 0
			continue
		}

		total := 0
		for start := 0; start <= len(fuels)-3; start++ {
			for end := start + 2; end < len(fuels); end++ {
				d := fuels[start+1] - fuels[start]
				isok := true
				for k := start + 2; k <= end; k++ {
					if fuels[k]-fuels[k-1] != d {
						isok = false
						break
					}
				}
				if isok {
					total++
				}
			}
		}
		counts[i] = total
	}
	for i := 0; i < n; i++ {
		for j := 0; j+1 < n; j++ {
			if counts[j] < counts[j+1] ||
				(counts[j] == counts[j+1] && names[j] > names[j+1]) {
				counts[j], counts[j+1] = counts[j+1], counts[j]
				names[j], names[j+1] = names[j+1], names[j]
			}
		}
	}
	for i := 0; i < n; i++ {
		fmt.Println(names[i], counts[i])
	}
}
