package main

import "fmt"

type Car struct {
	speed   int
	battery int
}

func NewCar(speed, battery int) *Car {
	c := &Car{speed: speed, battery: battery}
	return c
}
func GetSpeed(car *Car) int {
	s := car.speed
	return s
}
func GetBattery(car *Car) int {
	b := car.battery
	return b
}
func ChargeCar(car *Car, minutes int) {

	car.battery += minutes / 2
	if car.battery >= 100 {
		car.battery = 100
	}
}
func TryFinish(car *Car, distance int) string {
	batteryNeed := distance / 2
	if car.battery < batteryNeed {
		car.battery = 0
		return ""
	}
	car.battery -= distance / 2
	t := float64(distance) / float64(car.speed)
	return fmt.Sprintf("%.2f", t)
}
