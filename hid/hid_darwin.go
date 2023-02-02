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
func Connect(c context.Context) (d *HID) {
	// Initialize a new Context.
	ctx := gousb.NewContext()

	// Open any device with a given VID/PID using a convenience function.
	dev, err := ctx.OpenDeviceWithVIDPID(0x045e, 0x028e)
	if err != nil {
		log.Fatalf("Could not open a device: %v", err)
	}

	fmt.Println("Opened device")

	// Switch the configuration to #1
	cfg, err := dev.Config(1)
	if err != nil {
		log.Fatalf("%s.Config(1): %v", dev, err)
	}

	intf, err := cfg.Interface(0, 0)
	if err != nil {
		log.Fatalf("%s.Interface(0, 0): %v", cfg, err)
	}

	// Claim the default interface using a convenience function.
	// The default interface is always #0 alt #0 in the currently active
	// config.
	intf, done, err := dev.DefaultInterface()
	if err != nil {
		panic(err)
	}
	defer done()

	in, err := intf.InEndpoint(1)
	if err != nil {
		panic(err)
	}

	d = NewHID(c)

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
	return d
}

type osEvent struct {
	Time  uint32 // ms since firstTimestamp
	Value int16
	Type  uint8
	Index uint8
}

type input struct {
}

// Header 0014
func readDeviceInput(in *gousb.InEndpoint, c chan osEvent) {
	buf := make([]byte, in.Desc.MaxPacketSize)
	for {

		readBytes, err := in.Read(buf)
		if err != nil {
			log.Printf("Read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if readBytes == 0 {
			log.Fatalf("IN endpoint returned 0 bytes of data.")
		}

		// First 2 bytes are header
		fmt.Printf("Buf: %x\n", buf)

		var eventType uint8
		// byte 2 MSB
		b2msb := buf[2] >> 4
		// Start = 1
		if b2msb == 1 {
			fmt.Println("Start")
			eventType = ButtonEventType
		}
		// Select = 2
		if b2msb == 2 {
			fmt.Println("Select")
			eventType = ButtonEventType
		}
		// LJ = 4
		if b2msb == 4 {
			fmt.Println("LJ")
			eventType = ButtonEventType
		}
		// RJ = 8
		if b2msb == 8 {
			fmt.Println("RJ")
			eventType = ButtonEventType
		}

		// byte 3 - DPAD
		b2lsb := buf[2] & 0xf
		// Left
		if b2lsb == 4 {
			fmt.Println("Left DPAD")
			eventType = AxisEventType
		}
		// Right
		if b2lsb == 8 {
			fmt.Println("Right DPAD")
			eventType = AxisEventType
		}
		// Up
		if b2lsb == 1 {
			fmt.Println("Up DPAD")
			eventType = AxisEventType
		}
		// Down
		if b2lsb == 2 {
			fmt.Println("Down DPAD")
			eventType = AxisEventType
		}

		// byte 4 MSB - Actions
		b3msb := buf[3] >> 4
		// X
		if b3msb == 1 {
			fmt.Println("X")
			eventType = ButtonEventType

		}
		// O
		if b3msb == 2 {
			fmt.Println("O")
			eventType = ButtonEventType
		}
		// []
		if b3msb == 4 {
			fmt.Println("[]")
			eventType = ButtonEventType
		}
		// /\
		if b3msb == 8 {
			fmt.Println("/\\")
			eventType = ButtonEventType
		}

		// byte 5 LSB - Top triggers + Analog
		b3lsb := buf[3] & 0xf
		// L1
		if b3lsb == 1 {
			fmt.Println("L1")
			eventType = ButtonEventType
		}
		// R1
		if b3lsb == 2 {
			fmt.Println("R1")
			eventType = ButtonEventType
		}
		// Analog
		if b3lsb == 4 {
			fmt.Println("Analog")
			eventType = ButtonEventType
		}

		// byte 4 - L2
		b4 := buf[4]
		if b4 > 0 {
			fmt.Println("L2")
			eventType = AxisEventType // L2 treated as axis
		}

		// byte 5 - R2
		b5 := buf[5]
		if b5 > 0 {
			fmt.Println("R2")
			eventType = AxisEventType // R2 treated as axis
		}

		// byte 6 + 7
		b67 := int16(binary.LittleEndian.Uint16(buf[6:8]))
		if b67 != 0 {
			fmt.Printf("Left joy X: %v\n", b67)
			eventType = AxisEventType
		}

		// byte 8 + 9
		b89 := int16(binary.LittleEndian.Uint16(buf[8:10]))
		if b89 != 0 {
			fmt.Printf("Left joy Y: %v\n", b89)
			eventType = AxisEventType
		}

		// byte 10 + 11
		b1011 := int16(binary.LittleEndian.Uint16(buf[10:12]))
		if b1011 != 0 {
			fmt.Printf("Right joy X: %v\n", b1011)
			eventType = AxisEventType
		}

		// byte 11 + 12
		b1213 := int16(binary.LittleEndian.Uint16(buf[12:14]))
		if b1213 != 0 {
			fmt.Printf("Right joy Y: %v\n", b1213)
			eventType = AxisEventType
		}

		time.Sleep(time.Millisecond * 200)

		//ev := osEvent{
		//	Time:  uint32(time.Since(firstTimestamp).Milliseconds()),
		//	Value: 0,
		//	Type:  eventType,
		//	Index: 0,
		//}
		log.Println("Event type", eventType)
		//c <- ev
	}
}

// Compat with Linux drivers, just return the already computed milliseconds
func toElapsed(m uint32) time.Duration {
	return time.Duration(m) * time.Millisecond
}
