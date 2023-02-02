package hid

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type osEvent struct {
	Time  uint32
	Value int16
	Type  uint8
	Index uint8
}

// Linux spec: https://www.kernel.org/doc/Documentation/input/joystick-api.txt

const MaxValue = 1<<15 - 1

var lastTimestamp uint32

func deviceExists(index int) bool {
	_, err := os.Stat(fmt.Sprintf("/dev/input/js%v", index))
	return err == nil
}

func isGamepad(idx int) (driverName, bool) {
	d, err := os.ReadFile(fmt.Sprintf("/sys/class/input/js%v/device/name", idx))
	if err != nil {
		log.Printf("Error checking device name, err: %v", err)
		return "", false
	}
	name := strings.TrimSpace(string(d))
	for k, _ := range DriverMapping {
		if name == string(k) {
			return k, true
		}
	}
	return "", false
}

// Connect to device by index found in /dev/input/js*
func Connect(ctx context.Context) (*HID, error) {

	var driver driverName
	deviceIndex := -1
	for i := 0; i < 5; i++ {
		exists := deviceExists(i)
		if exists {
			if n, ok := isGamepad(i); ok {
				driver = n
				deviceIndex = i
				break
			}
		}
	}
	if deviceIndex == -1 {
		return nil, errors.New("cannot find device")
	}

	r, e := os.OpenFile(fmt.Sprintf("/dev/input/js%v", deviceIndex), os.O_RDWR, 0)
	if e != nil {
		return nil, e
	}
	d := newHID(ctx)
	d.Driver = driver

	// Clean up on context done
	go func() {
		<-ctx.Done()
		_ = r.Close()
	}()

	// Start reading from /dev/input device
	go readDeviceInput(r, d.osEventsCh)

	// Read initial events from gamepad
	d.mapInitalEvents()
	return d, nil
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
