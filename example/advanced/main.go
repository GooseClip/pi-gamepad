package main

import (
	"context"
	. "github.com/gooseclip/pi-gamepad"
	"log"
	"time"
)

func main() {
	gamepad, err := NewGamepad(
		context.Background(),
		WithDebug(),                     // Use to log debug messages
		WithInvertedY(),                 // Invert y axis
		WithClickDuration(time.Second),  // Change the click window - default 300ms
		WithHoldDuration(time.Second*3), // Change the hold window - default 800ms
	)
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
	}) // Default handle all events for this button

	gamepad.OnR1(func(event ButtonEvent) {
		switch event {
		case UpEvent:
		case DownEvent:
		case ClickEvent:
		case HoldEvent:
		}
		log.Printf("R1 %v\n", event)
	}, ClickEvent, UpEvent, DownEvent, HoldEvent) // Explicitly handle all events for this button

	gamepad.OnL2(func(event ButtonEvent) {
		log.Printf("L2 %v\n", event)
	}, HoldEvent) // Only handle hold events for this button

	gamepad.OnR2(func(event ButtonEvent) {
		log.Printf("R2 %v\n", event)
	}, HoldEvent) // Only handle hold events for this button

	// Handle action buttons
	gamepad.OnCross(func(event ButtonEvent) {
		log.Printf("Cross %v\n", event)
	}, ClickEvent) // Only handle click events for this button

	gamepad.OnCircle(func(event ButtonEvent) {
		log.Printf("Circle %v\n", event)
	}, ClickEvent) // Only handle click events for this button

	gamepad.OnSquare(func(event ButtonEvent) {
		log.Printf("Square %v\n", event)
	}, ClickEvent) // Only handle click events for this button

	gamepad.OnTriangle(func(event ButtonEvent) {
		log.Printf("Triangle %v\n", event)
	}, ClickEvent) // Only handle click events for this button

	// Handle special buttons
	gamepad.OnSelect(func(event ButtonEvent) {
		log.Printf("Select %v\n", event)
	}, UpEvent) // Only handle up events for this button

	gamepad.OnStart(func(event ButtonEvent) {
		log.Printf("Start %v\n", event)
	}, UpEvent) // Only handle up events for this button

	gamepad.OnAnalog(func(event ButtonEvent) {
		log.Printf("Analog %v\n", event)
	}, UpEvent) // Only handle up events for this button

	gamepad.OnLJ(func(event ButtonEvent) {
		log.Printf("Left Joystick Click %v\n", event)
	}, DownEvent) // Only handle down events for this button

	gamepad.OnRJ(func(event ButtonEvent) {
		log.Printf("Right Joystick Click %v\n", event)
	}, DownEvent) // Only handle down events for this button

	<-make(chan struct{})
}
