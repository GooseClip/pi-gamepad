package hid

import (
	"context"
	"log"
	"time"
)

type Resolved int

const (
	// Buttons
	CrossButton    Resolved = iota //  X
	CircleButton                   // ()
	SquareButton                   // []
	TriangleButton                 // /\
	L1Button
	R1Button
	SelectButton
	StartButton
	AnalogButton
	LeftJoyButton
	RightJoyButton
	// Axis
	DPadXAxis
	DPadYAxis
	LeftJoyXAxis
	LeftJoyYAxis
	RightJoyXAxis
	RightJoyYAxis
	L2Axis
	R2Axis
)

const (
	InputTypeButton int = iota
	InputTypeAxis
)

type Input struct {
	Type  int
	Value uint8
}

type InputMapping map[Input]Resolved

type driverName string

var DriverMapping = map[driverName]InputMapping{
	// Ubuntu 22.04 arm64
	"Microsoft X-Box 360 pad": {
		Input{InputTypeButton, 0}:  CrossButton,
		Input{InputTypeButton, 1}:  CircleButton,
		Input{InputTypeButton, 2}:  SquareButton,
		Input{InputTypeButton, 3}:  TriangleButton,
		Input{InputTypeButton, 4}:  L1Button,
		Input{InputTypeButton, 5}:  R1Button,
		Input{InputTypeButton, 6}:  SelectButton,
		Input{InputTypeButton, 7}:  StartButton,
		Input{InputTypeButton, 8}:  AnalogButton,
		Input{InputTypeButton, 9}:  LeftJoyButton,
		Input{InputTypeButton, 10}: RightJoyButton,
		Input{InputTypeAxis, 6}:    DPadXAxis,
		Input{InputTypeAxis, 7}:    DPadYAxis,
		Input{InputTypeAxis, 0}:    LeftJoyXAxis,
		Input{InputTypeAxis, 1}:    LeftJoyYAxis,
		Input{InputTypeAxis, 3}:    RightJoyXAxis,
		Input{InputTypeAxis, 4}:    RightJoyYAxis,
		Input{InputTypeAxis, 2}:    L2Axis,
		Input{InputTypeAxis, 5}:    R2Axis,
	},

	// Raspberry pi 4 - Ubuntu 22.10
	"SHANWAN Android Gamepad": {
		Input{InputTypeButton, 0}:  CrossButton,
		Input{InputTypeButton, 1}:  CircleButton,
		Input{InputTypeButton, 3}:  SquareButton,
		Input{InputTypeButton, 4}:  TriangleButton,
		Input{InputTypeButton, 6}:  L1Button,
		Input{InputTypeButton, 7}:  R1Button,
		Input{InputTypeButton, 10}: SelectButton,
		Input{InputTypeButton, 11}: StartButton,
		Input{InputTypeButton, 12}: AnalogButton,
		Input{InputTypeButton, 13}: LeftJoyButton,
		Input{InputTypeButton, 14}: RightJoyButton,
		Input{InputTypeAxis, 6}:    DPadXAxis,
		Input{InputTypeAxis, 7}:    DPadYAxis,
		Input{InputTypeAxis, 0}:    LeftJoyXAxis,
		Input{InputTypeAxis, 1}:    LeftJoyYAxis,
		Input{InputTypeAxis, 2}:    RightJoyXAxis,
		Input{InputTypeAxis, 3}:    RightJoyYAxis,
		Input{InputTypeAxis, 5}:    L2Axis,
		Input{InputTypeAxis, 4}:    R2Axis,
	},

	// MacOS arm64 and amd64
	"MacOS": {
		Input{InputTypeButton, 7}:  CrossButton,
		Input{InputTypeButton, 8}:  CircleButton,
		Input{InputTypeButton, 9}:  SquareButton,
		Input{InputTypeButton, 10}: TriangleButton,
		Input{InputTypeButton, 11}: L1Button,
		Input{InputTypeButton, 12}: R1Button,
		Input{InputTypeButton, 2}:  SelectButton,
		Input{InputTypeButton, 1}:  StartButton,
		Input{InputTypeButton, 13}: AnalogButton,
		Input{InputTypeButton, 3}:  LeftJoyButton,
		Input{InputTypeButton, 4}:  RightJoyButton,
		Input{InputTypeAxis, 5}:    DPadXAxis,
		Input{InputTypeAxis, 6}:    DPadYAxis,
		Input{InputTypeAxis, 16}:   LeftJoyXAxis,
		Input{InputTypeAxis, 17}:   LeftJoyYAxis,
		Input{InputTypeAxis, 18}:   RightJoyXAxis,
		Input{InputTypeAxis, 19}:   RightJoyYAxis,
		Input{InputTypeAxis, 14}:   L2Axis,
		Input{InputTypeAxis, 15}:   R2Axis,
	},
}

type HID struct {
	ctx        context.Context
	osEventsCh chan osEvent
	buttonCh   chan buttonEvent
	axisCh     chan axisEvent
	Driver     driverName
}

type buttonEvent struct {
	When   time.Duration
	Button uint8
	Value  int16
}

type axisEvent struct {
	When  time.Duration
	Axis  uint8
	Value int16
}

type eventType uint8

const (
	invalidEventType eventType = iota
	buttonEventType
	axisEventType
)

func newHID(ctx context.Context) *HID {
	h := &HID{
		ctx:        ctx,
		osEventsCh: make(chan osEvent),
		buttonCh:   make(chan buttonEvent),
		axisCh:     make(chan axisEvent),
	}
	go h.handleEvents()
	return h
}

// handleEvents waits on the HID.OSEvents channel (so is blocking), then puts any events matching onto any registered channel(s).
// This function handles HID events
func (h *HID) handleEvents() {
	for {
		select {
		// If the context is done, return
		case <-h.ctx.Done():
			return
		// If there is an event on the OS events channel
		case evt, ok := <-h.osEventsCh:
			// If the channel is closed, return
			if !ok {
				return
			}

			// Determine the type of event
			switch eventType(evt.Type) {
			// If it's a button event
			case buttonEventType:
				// Try to send the event to the button channel
				select {
				// If the event is successfully sent
				case h.buttonCh <- buttonEvent{
					When:   toElapsed(evt.Time),
					Button: evt.Index,
					Value:  evt.Value,
				}:
				// If the button channel is full, drop the event and log a message
				case <-time.NewTimer(time.Millisecond * 20).C:
					log.Printf("Button event dropped, index: %v", evt.Index)
				}
			// If it's an axis event
			case axisEventType:
				// Try to send the event to the axis channel
				select {
				// If the event is successfully sent
				case h.axisCh <- axisEvent{
					When:  toElapsed(evt.Time),
					Axis:  evt.Index,
					Value: evt.Value,
				}:
				// If the axis channel is full, drop the event and log a message
				case <-time.NewTimer(time.Millisecond * 20).C:
					log.Printf("Axis event dropped, index: %v", evt.Index)
				}
			}
		}
	}
}
}

func (h *HID) OnButton() <-chan buttonEvent {
	return h.buttonCh
}

func (h *HID) OnAxis() <-chan axisEvent {
	return h.axisCh
}
