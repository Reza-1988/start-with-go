package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	clothesColour := make(map[string][]string)

	for i := 0; i < 5; i++ {
		scanner.Scan()
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valuesPart := strings.TrimSpace(parts[1])
		values := strings.Fields(valuesPart)
		clothesColour[key] = values
	}

	scanner.Scan()
	season := strings.TrimSpace(scanner.Text())

	switch season {
	case "SPRING":
		for _, shirt := range clothesColour["SHIRT"] {
			for _, pants := range clothesColour["PANTS"] {
				fmt.Printf("SHIRT: %s PANTS: %s\n", shirt, pants)

				for _, coat := range clothesColour["COAT"] {
					fmt.Printf("COAT: %s SHIRT: %s PANTS: %s\n", coat, shirt, pants)
				}

				for _, cap := range clothesColour["CAP"] {
					fmt.Printf("SHIRT: %s PANTS: %s CAP: %s\n", shirt, pants, cap)
				}

				for _, coat := range clothesColour["COAT"] {
					for _, cap := range clothesColour["CAP"] {
						fmt.Printf("COAT: %s SHIRT: %s PANTS: %s CAP: %s\n", coat, shirt, pants, cap)
					}
				}
			}
		}

	case "SUMMER":
		for _, shirt := range clothesColour["SHIRT"] {
			for _, pants := range clothesColour["PANTS"] {
				for _, cap := range clothesColour["CAP"] {
					fmt.Printf("SHIRT: %s PANTS: %s CAP: %s\n", shirt, pants, cap)
				}
			}
		}

	case "FALL":
		for _, shirt := range clothesColour["SHIRT"] {
			for _, pants := range clothesColour["PANTS"] {
				fmt.Printf("SHIRT: %s PANTS: %s\n", shirt, pants)

				validCoats := []string{}
				for _, coat := range clothesColour["COAT"] {
					if coat != "yellow" && coat != "orange" {
						validCoats = append(validCoats, coat)
					}
				}

				for _, coat := range validCoats {
					fmt.Printf("COAT: %s SHIRT: %s PANTS: %s\n", coat, shirt, pants)
				}

				for _, cap := range clothesColour["CAP"] {
					fmt.Printf("SHIRT: %s PANTS: %s CAP: %s\n", shirt, pants, cap)
				}

				for _, coat := range validCoats {
					for _, cap := range clothesColour["CAP"] {
						fmt.Printf("COAT: %s SHIRT: %s PANTS: %s CAP: %s\n", coat, shirt, pants, cap)
					}
				}
			}
		}

	case "WINTER":
		for _, shirt := range clothesColour["SHIRT"] {
			for _, pants := range clothesColour["PANTS"] {

				for _, coat := range clothesColour["COAT"] {
					fmt.Printf("COAT: %s SHIRT: %s PANTS: %s\n", coat, shirt, pants)
				}

				for _, jacket := range clothesColour["JACKET"] {
					fmt.Printf("SHIRT: %s PANTS: %s JACKET: %s\n", shirt, pants, jacket)
				}
			}
		}
	}
}
