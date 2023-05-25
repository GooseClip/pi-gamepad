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
		WithDebug(),
		WithInvertedY(),
		WithClickDuration(time.Second),
		WithHoldDuration(time.Second*3),
	)
	if err != nil {
		panic(err)
	}

	gamepad.OnDPad(func(x, y float32) {
		log.Printf("DPad, x: %v, y: %v\n", x, y)
	})

	gamepad.OnLeftJoystick(func(x, y float32) {
		log.Printf("Left Joystick, x: %v, y: %v\n", x, y)
	})

	gamepad.OnRightJoystick(func(x, y float32) {
		log.Printf("Right Joystick, x: %v, y: %v\n", x, y)
	})

	gamepad.OnL1(func(event ButtonEvent) {
		log.Printf("L1 %v\n", event)
	})

	gamepad.OnR1(func(event ButtonEvent) {
		switch event {
		case UpEvent:
		case DownEvent:
		case ClickEvent:
		case HoldEvent:
		}
		log.Printf("R1 %v\n", event)
	}, ClickEvent, UpEvent, DownEvent, HoldEvent)

	gamepad.OnL2(func(event ButtonEvent) {
		log.Printf("L2 %v\n", event)
	}, HoldEvent)

	gamepad.OnR2(func(event ButtonEvent) {
		log.Printf("R2 %v\n", event)
	}, HoldEvent)

	gamepad.OnCross(func(event ButtonEvent) {
		log.Printf("Cross %v\n", event)
	}, ClickEvent)

	gamepad.OnCircle(func(event ButtonEvent) {
		log.Printf("Circle %v\n", event)
	}, ClickEvent)

	gamepad.OnSquare(func(event ButtonEvent) {
		log.Printf("Square %v\n", event)
	}, ClickEvent)

	gamepad.OnTriangle(func(event ButtonEvent) {
		log.Printf("Triangle %v\n", event)
	}, ClickEvent)

	gamepad.OnSelect(func(event ButtonEvent) {
		log.Printf("Select %v\n", event)
	}, UpEvent)

	gamepad.OnStart(func(event ButtonEvent) {
		log.Printf("Start %v\n", event)
	}, UpEvent)

	gamepad.OnAnalog(func(event ButtonEvent) {
		log.Printf("Analog %v\n", event)
	}, UpEvent)

	gamepad.OnLJ(func(event ButtonEvent) {
		log.Printf("Left Joystick Click %v\n", event)
	}, DownEvent)

	gamepad.OnRJ(func(event ButtonEvent) {
		log.Printf("Right Joystick Click %v\n", event)
	}, DownEvent)

	<-make(chan struct{})
}
