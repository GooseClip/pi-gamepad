# Pi Gamepad

Simplified event-driven gamepad library for interacting with a Raspberry Pi - tested using the PiHut gamepad.
> NOTE: Linux only

#### Example

```
	gamepad, err := NewGamepad(context.Background(), WithInvertedY())
	if err != nil {
		panic(err)
	}

	// Handle left joystick
	gamepad.OnLeftJoystick(func(x, y float32) {
		log.Printf("x: %v, y: %v\n", x, y)
	})
```

#### Using a different gamepad type
By default NewGamepad attempts to find a device by the name "Microsoft X-Box 360 pad" (this is what the PiHut gamepad reports it's name as).
To override this behaviour you can modify gamepad.ExpectedName = "YOUR DEVICE NAME HERE
> NOTE: You can find your device name in `/sys/class/input/js${DEVICE_INDEX}/device/name`

<img src="https://cdn.shopify.com/s/files/1/0176/3274/products/raspberry-pi-compatible-wireless-gamepad-controller-the-pi-hut-102347-22608519185_1000x.jpg?v=1646248693" width="250"/>

[PiHut link](https://thepihut.com/products/raspberry-pi-compatible-wireless-gamepad-controller)

----

##### Inspired by: https://github.com/splace/joysticks