package main

import (
	"context"
	. "github.com/gooseclip/pi-gamepad"
	"log"
)

func main() {
	gamepad, err := NewGamepad(context.Background())
	if err != nil {
		panic(err)
	}

	// Handle navigation
	gamepad.OnDPad(func(x, y float32) {
		log.Printf("DPad, x: %v, y: %v\n", x, y)
	})

	gamepad.OnLeftJoystick(func(x, y float32) {
		log.Printf("Left Joystick, x: %v, y: %v\n", x, y)
	})

	gamepad.OnRightJoystick(func(x, y float32) {
		log.Printf("Right Joystick, x: %v, y: %v\n", x, y)
	})

	// Handle triggers
	gamepad.OnL1(func(event ButtonEvent) {
		log.Printf("L1 %v\n", event)
	})
	gamepad.OnR1(func(event ButtonEvent) {
		log.Printf("R1 %v\n", event)
	})

	gamepad.OnL2(func(event ButtonEvent) {
		log.Printf("L2 %v\n", event)
	})

	gamepad.OnR2(func(event ButtonEvent) {
		log.Printf("R2 %v\n", event)
	})

	// Handle action buttons
	gamepad.OnCross(func(event ButtonEvent) {
		log.Printf("Cross %v\n", event)
	})

	gamepad.OnCircle(func(event ButtonEvent) {
		log.Printf("Circle %v\n", event)
	})

	gamepad.OnSquare(func(event ButtonEvent) {
		log.Printf("Square %v\n", event)
	})

	gamepad.OnTriangle(func(event ButtonEvent) {
		log.Printf("Triangle %v\n", event)
	})

	// Handle special buttons
	gamepad.OnSelect(func(event ButtonEvent) {
		log.Printf("Select %v\n", event)
	})

	gamepad.OnStart(func(event ButtonEvent) {
		log.Printf("Start %v\n", event)
	})

	gamepad.OnAnalog(func(event ButtonEvent) {
		log.Printf("Analog %v\n", event)
	})

	gamepad.OnLJ(func(event ButtonEvent) {
		log.Printf("Left Joystick Click %v\n", event)
	})

	gamepad.OnRJ(func(event ButtonEvent) {
		log.Printf("Right Joystick Click %v\n", event)
	})

	<-make(chan struct{})
}
