package main

import "fmt"

// Do Not Change This File

var tikCount int
var takCount int

func SayTik() {
	tikCount++
	fmt.Print("Tik")
}

func SayTak() {
	takCount++
	fmt.Print("Tak")
}
