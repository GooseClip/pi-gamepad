//go:build linux
// +build linux

package hid

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

// Linux spec: https://www.kernel.org/doc/Documentation/input/joystick-api.txt
type osEvent struct {
	Time  uint32
	Value int16
	Type  uint8
	Index uint8
}

const MaxValue = 1<<15 - 1

var lastTimestamp uint32

func DeviceExists(index int) bool {
	_, err := os.Stat(fmt.Sprintf("/dev/input/js%v", index))
	return err == nil
}

// Connect to device by index found in /dev/input/js*
func Connect(ctx context.Context, index int) (d *HID) {
	r, e := os.OpenFile(fmt.Sprintf("/dev/input/js%v", index), os.O_RDWR, 0)
	if e != nil {
		return nil
	}
	d = NewHID(ctx)

	// Clean up on context done
	go func() {
		<-ctx.Done()
		_ = r.Close()
	}()

	// Start reading from /dev/input device
	go readDeviceInput(r, d.osEventsCh)

	// Read initial events from gamepad
	d.mapInitalEvents()
	return d
}

func (h *HID) mapInitalEvents() {
	for {
		evt, ok := <-h.osEventsCh
		if !ok {
			return
		}

		switch evt.Type {
		case 0x81:
			lastTimestamp = evt.Time
		case 0x82:
			lastTimestamp = evt.Time
		default:
			// Receiving the first non 0x81 or 0x82 event is our signal that populating is done. Forward this event as a real event.
			go func() { h.osEventsCh <- evt }()
			return
		}
	}
}

func readDeviceInput(r io.Reader, c chan osEvent) {
	var evt osEvent
	for {
		if binary.Read(r, binary.LittleEndian, &evt) != nil {
			close(c)
			return
		}
		c <- evt
	}
}

func toElapsed(m uint32) time.Duration {
	return time.Duration(m-lastTimestamp) * time.Millisecond
}
