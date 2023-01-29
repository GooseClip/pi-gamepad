package main

import (
	"context"
	"github.com/gooseclip/pi-gamepad"
	"log"
)

func main() {

	// Setup a custom device name and configure inputs.
	// Note: Inputs can be discovered by first setting up an empty mapping and using WithDebug
	// to determine the input mapping.
	gamepad.DriverMapping["My Custom Driver Name"] = gamepad.InputMapping{
		//DPadXAxis:                 3,
		gamepad.Input{Type: gamepad.InputTypeAxis, Value: 2}: gamepad.DPadXAxis,
		//DPadYAxis:                 2,
		gamepad.Input{Type: gamepad.InputTypeAxis, Value: 3}: gamepad.DPadYAxis,
	}

	gp, err := gamepad.NewGamepad(context.Background())
	if err != nil {
		panic(err)
	}

	// Handle navigation
	gp.OnDPad(func(x, y float32) {
		log.Printf("DPad, x: %v, y: %v\n", x, y)
	})

	<-make(chan struct{})
}
