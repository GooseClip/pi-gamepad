package main

import (
	"context"
	. "github.com/mousybusiness/pi-gamepad"
	"log"
	"os"
	"os/signal"
)

func main() {
	gamepad, err := NewGamepad(context.Background(), WithInvertedY())
	if err != nil {
		panic(err)
	}

	// Handle left joystick
	gamepad.OnLeftJoystick(func(x, y float32) {
		log.Printf("x: %v, y: %v\n", x, y)
	})

	// Handle click events from X button
	gamepad.OnCross(func(_ ButtonEvent) {
		log.Printf("x button clicked\n")
	}, ClickEvent)

	// Handle hold events from L2
	gamepad.OnL2(func(_ ButtonEvent) {

	}, HoldEvent)

	// Handle all events by default
	gamepad.OnSquare(func(event ButtonEvent) {
		log.Printf("[] button event: %v\n", event)
	})

	// Block until exit signal is received
	block := make(chan os.Signal, 1)
	signal.Notify(block, os.Interrupt)
	<-block
}
