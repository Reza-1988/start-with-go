package main

import "math"

type FilterFunc func(int) bool
type MapperFunc func(int) int

func IsSquare(x int) bool {
	if x < 0 {
		return false
	}
	r := int(math.Sqrt(float64(x)))
	return r*r == x
}

func IsPalindrome(x int) bool {
	original := x
	reversed := 0
	sign := 1
	if x < 0 {
		sign = -1
		x = -x
	}

	for x != 0 {
		last := x % 10
		reversed = reversed*10 + last
		x = x / 10
	}
	if sign == -1 {
		reversed = -reversed
	}
	if reversed != original {
		return false
	}
	return true
}

func Abs(num int) int {
	if num < 0 {
		return -num
	}
	return num
}

func Cube(num int) int {

	return num * num * num
}

func Filter(input []int, f FilterFunc) []int {
	result := make([]int, 0, len(input))
	for _, x := range input {
		if f(x) {
			result = append(result, x)
		}
	}
	return result
}

func Map(input []int, m MapperFunc) []int {
	result := make([]int, len(input))
	for i, x := range input {
		result[i] = m(x)
	}
	return result
}
