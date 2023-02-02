package hid

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/google/gousb"
	"log"
	"time"
)

const MaxValue = 1<<15 - 1

var firstTimestamp time.Time

// Connect to device by index found in /dev/input/js*
func Connect(c context.Context) (*HID, error) {
	// Initialize a new Context.
	ctx := gousb.NewContext()

	// Open any device with a given VID/PID using a convenience function.
	dev, err := ctx.OpenDeviceWithVIDPID(0x045e, 0x028e)
	if err != nil {
		return nil, fmt.Errorf("could not open a device: %v", err)
	}

	log.Printf("Opened device: %v", dev)

	// Switch the configuration to #1
	cfg, err := dev.Config(1)
	if err != nil {
		return nil, fmt.Errorf("invalid config number for device: %v", err)
	}

	intf, err := cfg.Interface(0, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid interface number for device: %v", err)
	}

	in, err := intf.InEndpoint(1)
	if err != nil {
		return nil, fmt.Errorf("invalid input endpoint for device: %v", err)
	}

	d := newHID(c)
	d.Driver = "MacOS"

	// Clean up on context done
	go func() {
		<-c.Done()
		intf.Close()
		_ = cfg.Close()
		_ = dev.Close()
		_ = ctx.Close()
	}()

	// Start reading from /dev/input device
	go readDeviceInput(in, d.osEventsCh)

	// Read initial events from gamepad
	firstTimestamp = time.Now()
	return d, nil
}

type osEvent struct {
	Time  uint32 // ms since firstTimestamp
	Value int16
	Type  uint8
	Index uint8
}

type cache struct {
	startBtn  bool
	selectBtn bool
	ljBtn     bool
	rjBtn     bool
	lpadAxis  bool
	rpadAxis  bool
	upadAxis  bool
	dpadAxis  bool
	xBtn      bool
	oBtn      bool
	sBtn      bool
	tBtn      bool
	l1Btn     bool
	r1Btn     bool
	analogBtn bool
	l2Axis    bool
	r2Axis    bool
	ljXAxis   bool
	ljYAxis   bool
	rjXAxis   bool
	rjYAxis   bool
}

func (c *cache) buttonEdge(b byte, want byte, cache *bool) (eventType, uint8) {
	if b == want {
		if !*cache {
			// Rising edge
			*cache = true
			return buttonEventType, 1
		}
	} else if *cache {
		// Falling edge
		*cache = false
		return buttonEventType, 0
	}

	return invalidEventType, 0
}

func (c *cache) axisEdge(val int16, cache *bool) (eventType, int16) {
	if val != 0 {
		//if !*cache {
		// Rising edge
		*cache = true
		return axisEventType, val
		//}
	} else if *cache {
		// Falling edge
		*cache = false
		return axisEventType, 0
	}

	return invalidEventType, 0
}

func emit(ch chan osEvent, eventType eventType, index uint8, value int16) {
	ev := osEvent{
		Time:  uint32(time.Since(firstTimestamp).Milliseconds()),
		Value: value,
		Type:  uint8(eventType),
		Index: index,
	}

	ch <- ev
}

// Values which will be used to map in gamepad.go
const (
	startButtonIndex = iota + 1
	selectButtonIndex
	ljButtonIndex
	rjButtonIndex
	dpadXAxisIndex
	dpadYAxisIndex
	crossButtonIndex
	circleButtonIndex
	squareButtonIndex
	triangleButtonIndex
	l1ButtonIndex
	r1ButtonIndex
	analogButtonIndex
	l2AxisIndex
	r2AxisIndex
	ljxAxisIndex
	ljyAxisIndex
	rjxAxisIndex
	rjyAxisIndex
)

func readDeviceInput(in *gousb.InEndpoint, ch chan osEvent) {
	c := cache{}
	buf := make([]byte, in.Desc.MaxPacketSize)
	for {

		readBytes, err := in.Read(buf)
		if err != nil {
			log.Fatalf("Read error: %v", err)
		}

		if readBytes == 0 {
			log.Fatalf("Device returned 0 bytes of data.")
		}

		// byte 2 MSB
		b2msb := buf[2] >> 4
		// Start = 1
		if ev, v := c.buttonEdge(b2msb, 1, &c.startBtn); ev != invalidEventType {
			emit(ch, ev, startButtonIndex, int16(v))
		}

		// Select = 2
		if ev, v := c.buttonEdge(b2msb, 2, &c.selectBtn); ev != invalidEventType {
			emit(ch, ev, selectButtonIndex, int16(v))
		}

		// LJ = 4
		if ev, v := c.buttonEdge(b2msb, 4, &c.ljBtn); ev != invalidEventType {
			emit(ch, ev, ljButtonIndex, int16(v))
		}

		// RJ = 8
		if ev, v := c.buttonEdge(b2msb, 8, &c.rjBtn); ev != invalidEventType {
			emit(ch, ev, rjButtonIndex, int16(v))
		}

		// byte 3 - DPAD
		b2lsb := buf[2] & 0xf

		// Left
		if ev, v := c.buttonEdge(b2lsb, 4, &c.lpadAxis); ev != invalidEventType {
			if v == 1 {
				emit(ch, axisEventType, dpadXAxisIndex, -MaxValue)
			} else {
				emit(ch, axisEventType, dpadXAxisIndex, 0)
			}
		}

		// Right
		if ev, v := c.buttonEdge(b2lsb, 8, &c.rpadAxis); ev != invalidEventType {
			if v == 1 {
				emit(ch, axisEventType, dpadXAxisIndex, MaxValue)
			} else {
				emit(ch, axisEventType, dpadXAxisIndex, 0)
			}
		}

		// Up
		if ev, v := c.buttonEdge(b2lsb, 1, &c.upadAxis); ev != invalidEventType {
			if v == 1 {
				emit(ch, axisEventType, dpadYAxisIndex, MaxValue)
			} else {
				emit(ch, axisEventType, dpadYAxisIndex, 0)
			}
		}

		// Down
		if ev, v := c.buttonEdge(b2lsb, 2, &c.dpadAxis); ev != invalidEventType {
			if v == 1 {
				emit(ch, axisEventType, dpadYAxisIndex, -MaxValue)
			} else {
				emit(ch, axisEventType, dpadYAxisIndex, 0)
			}
		}

		// byte 4 MSB - Actions
		b3msb := buf[3] >> 4
		// X
		if ev, v := c.buttonEdge(b3msb, 1, &c.xBtn); ev != invalidEventType {
			emit(ch, ev, crossButtonIndex, int16(v))
		}

		// O
		if ev, v := c.buttonEdge(b3msb, 2, &c.oBtn); ev != invalidEventType {
			emit(ch, ev, circleButtonIndex, int16(v))
		}

		// []
		if ev, v := c.buttonEdge(b3msb, 4, &c.sBtn); ev != invalidEventType {
			emit(ch, ev, squareButtonIndex, int16(v))
		}

		// /\
		if ev, v := c.buttonEdge(b3msb, 8, &c.tBtn); ev != invalidEventType {
			emit(ch, ev, triangleButtonIndex, int16(v))
		}

		// byte 5 LSB - Top triggers + Analog
		b3lsb := buf[3] & 0xf
		// L1
		if ev, v := c.buttonEdge(b3lsb, 1, &c.l1Btn); ev != invalidEventType {
			emit(ch, ev, l1ButtonIndex, int16(v))
		}

		// R1
		if ev, v := c.buttonEdge(b3lsb, 2, &c.r1Btn); ev != invalidEventType {
			emit(ch, ev, r1ButtonIndex, int16(v))
		}

		// Analog
		if ev, v := c.buttonEdge(b3lsb, 4, &c.analogBtn); ev != invalidEventType {
			emit(ch, ev, analogButtonIndex, int16(v))
		}

		// byte 4 - L2
		b4 := buf[4]
		if ev, v := c.buttonEdge(b4, 255, &c.l2Axis); ev != invalidEventType {
			if v > 0 {
				emit(ch, axisEventType, l2AxisIndex, MaxValue)
			} else {
				emit(ch, axisEventType, l2AxisIndex, 0)
			}
		}

		// byte 5 - R2
		b5 := buf[5]
		if ev, v := c.buttonEdge(b5, 255, &c.r2Axis); ev != invalidEventType {
			if v > 0 {
				emit(ch, axisEventType, r2AxisIndex, MaxValue)
			} else {
				emit(ch, axisEventType, r2AxisIndex, 0)
			}
		}

		// byte 6 + 7
		b67 := int16(binary.LittleEndian.Uint16(buf[6:8]))
		if ev, v := c.axisEdge(b67, &c.ljXAxis); ev != invalidEventType {
			emit(ch, ev, ljxAxisIndex, v)
		}

		// byte 8 + 9
		b89 := int16(binary.LittleEndian.Uint16(buf[8:10]))
		if ev, v := c.axisEdge(b89, &c.ljYAxis); ev != invalidEventType {
			emit(ch, ev, ljyAxisIndex, v)
		}

		// byte 10 + 11
		b1011 := int16(binary.LittleEndian.Uint16(buf[10:12]))
		if ev, v := c.axisEdge(b1011, &c.rjXAxis); ev != invalidEventType {
			emit(ch, ev, rjxAxisIndex, v)
		}

		// byte 11 + 12
		b1213 := int16(binary.LittleEndian.Uint16(buf[12:14]))
		if ev, v := c.axisEdge(b1213, &c.rjYAxis); ev != invalidEventType {
			emit(ch, ev, rjyAxisIndex, v)
		}
	}
}

// Compat with Linux drivers, just return the already computed milliseconds
func toElapsed(m uint32) time.Duration {
	return time.Duration(m) * time.Millisecond
}
