package main

import (
	"fmt"
	"strings"
	"unicode"
)

func translator(word string, dict map[string]string) string {
	if word == "" {
		return ""
	}
	//
	i := len(word)
	for i > 0 {
		c := word[i-1]
		if c == '.' || c == ',' || c == '!' || c == '?' {
			i--
		} else {
			break
		}
	}

	base := word[:i]
	punctuation := word[i:]

	if base == "" {

		return word
	}
	//
	runes := []rune(base)
	first := runes[0]
	isUpper := first >= 'A' && first <= 'Z'

	lowerBase := strings.ToLower(base)

	trans, ok := dict[lowerBase]
	if !ok {

		return word
	}
	//
	trRunes := []rune(trans)
	if isUpper && len(trRunes) > 0 {
		trRunes[0] = unicode.ToUpper(trRunes[0])
	}
	transFinal := string(trRunes)

	return transFinal + punctuation
}

func main() {

	var num int
	_, err := fmt.Scan(&num)
	if err != nil {
		fmt.Println("Error1", err)
		return
	}

	dict := make(map[string]string, num)

	for i := 0; i < num; i++ {
		var left, arrow, right string
		_, err = fmt.Scan(&left, &arrow, &right)
		if err != nil {
			fmt.Println("Error2", err)
		}
		if arrow != "->" {
			continue
		}
		dict[left] = right
	}

	var words []string
	for {
		var w string
		_, err = fmt.Scan(&w)
		if err != nil {
			break
		}
		words = append(words, translator(w, dict))
	}

	res := strings.Join(words, " ")
	fmt.Println(res)

}
