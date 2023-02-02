package hid

import (
	"context"
	"log"
	"time"
)

type DriverName string

type HID struct {
	ctx        context.Context
	osEventsCh chan osEvent
	buttonCh   chan ButtonEvent
	axisCh     chan AxisEvent
	Driver     DriverName
}

type ButtonEvent struct {
	When   time.Duration
	Button uint8
	Value  int16
}

type AxisEvent struct {
	When  time.Duration
	Axis  uint8
	Value int16
}

const (
	ButtonEventType uint8 = 1
	AxisEventType   uint8 = 2
)

func NewHID(ctx context.Context) *HID {
	h := &HID{
		ctx:        ctx,
		osEventsCh: make(chan osEvent),
		buttonCh:   make(chan ButtonEvent),
		axisCh:     make(chan AxisEvent),
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

			switch evt.Type {
			case ButtonEventType:
				select {
				case h.buttonCh <- ButtonEvent{
					When:   toElapsed(evt.Time),
					Button: evt.Index,
					Value:  evt.Value,
				}:
				case <-time.NewTimer(time.Millisecond * 20).C:
					log.Printf("Button event dropped, index: %v", evt.Index)
				}
			case AxisEventType:
				select {
				case h.axisCh <- AxisEvent{
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

func (h *HID) OnButton() <-chan ButtonEvent {
	return h.buttonCh
}

func (h *HID) OnAxis() <-chan AxisEvent {
	return h.axisCh
}
