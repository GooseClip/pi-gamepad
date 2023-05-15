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
		Input{t16
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
func (h *HID) handleEvents() {
	for {
		select {
		case <-h.ctx.Done():
			return
		case evt, ok := <-h.osEventsCh:
			if !ok {
				return
			}

			switch eventType(evt.Type) {
			case buttonEventType:
				select {
				case h.buttonCh <- buttonEvent{
					When:   toElapsed(evt.Time),
					Button: evt.Index,
					Value:  evt.Value,
				}:
				case <-time.NewTimer(time.Millisecond * 20).C:
					log.Printf("Button event dropped, index: %v", evt.Index)
				}
			case axisEventType:
				select {
				case h.axisCh <- axisEvent{
					When:  toElapsed(evt.Time),
					Axis:  evt.Index,
					Value: evt.Value,
				}:
				case <-time.NewTimer(time.Millisecond * 20).C:
					log.Printf("Axis event dropped, index: %v", evt.Index)
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
