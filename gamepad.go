package gamepad

import (
	"context"
	"errors"
	"fmt"
	"github.com/gooseclip/pi-gamepad/hid"
	"log"
	"os"
	"strings"
	"time"
)

const (
	CrossButton    uint8 = iota // 0 - X
	CircleButton                // 1 - ()
	SquareButton                // 2 - []
	TriangleButton              // 3 - /\
	L1Button                    // 4 - L1
	R1Button                    // 5 - L2
	SelectButton                // 6 - Select
	StartButton                 // 7 - Start
	AnalogButton                // 8 - Analog
	LeftJoyButton               // 9 - Left joystick click
	RightJoyButton              // 10 - Right joystick click

	// Internal axis mapping
	dPadXAxis     = 6
	dPadYAxis     = 7
	leftJoyXAxis  = 0
	leftJoyYAxis  = 1
	rightJoyXAxis = 3
	rightJoyYAxis = 4
	l2Axis        = 2
	r2Axis        = 5
)

var ExpectedName = "Microsoft X-Box 360 pad"

type ButtonPosition int

const (
	UpPosition ButtonPosition = iota
	DownPosition
)

type ButtonEvent int

const (
	UpEvent ButtonEvent = iota
	DownEvent
	ClickEvent
	HoldEvent
)

func (e ButtonEvent) String() string {
	switch e {
	case UpEvent:
		return "Up"
	case DownEvent:
		return "Down"
	case ClickEvent:
		return "Click"
	case HoldEvent:
		return "Hold"
	}
	return "Unknown"
}

const (
	defaultClickDuration = time.Millisecond * 300
	defaultHoldDuration  = time.Millisecond * 800
)

type Gamepad struct {
	ctx           context.Context
	cancel        context.CancelFunc
	device        *hid.HID
	invertY       bool
	axisCache     map[int]int
	clickDuration time.Duration
	holdDuration  time.Duration

	// Movement
	dpadHandler     directionHandler
	leftJoyHandler  directionHandler
	rightJoyHandler directionHandler

	// Action buttons
	crossBtn    *button
	circleBtn   *button
	squareBtn   *button
	triangleBtn *button

	// Trigger buttons
	l1Btn *button
	l2Btn *button
	r1Btn *button
	r2Btn *button

	// Special function buttons
	selectBtn *button
	startBtn  *button
	analogBtn *button
	ljBtn     *button
	rjBtn     *button
}

type directionHandler func(x, y float32)

type button struct {
	handler      buttonHandler
	events       []ButtonEvent
	lastPosition ButtonPosition
	downTime     time.Time
	holdTimer    *time.Timer
}

type buttonHandler func(event ButtonEvent)

type option func(*Gamepad)

func NewGamepad(ctx context.Context, opts ...option) (*Gamepad, error) {

	ctx, cancel := context.WithCancel(ctx)

	deviceIndex := -1
	for i := 0; i < 5; i++ {
		exists := hid.DeviceExists(i)
		if exists && isGamepad(i) {
			log.Printf("Found device %v\n", i)
			deviceIndex = i
			break
		}
	}
	if deviceIndex == -1 {
		cancel()
		return nil, errors.New("cannot find device")
	}

	device := hid.Connect(ctx, deviceIndex)
	if device == nil {
		cancel()
		return nil, fmt.Errorf("Failed to connect with device: js%v", deviceIndex)
	}

	g := &Gamepad{
		ctx:           ctx,
		cancel:        cancel,
		device:        device,
		axisCache:     make(map[int]int),
		clickDuration: defaultClickDuration,
		holdDuration:  defaultHoldDuration,
	}

	for _, o := range opts {
		o(g)
	}

	// Initialize axis cache with zero values
	for i := 0; i < 8; i++ {
		g.axisCache[i] = 0
	}

	go g.handleEvents()

	log.Println("Max value is", hid.MaxValue)

	return g, nil
}

func WithInvertedY() option {
	return func(gamepad *Gamepad) {
		gamepad.invertY = true
	}
}

func WithClickDuration(duration time.Duration) option {
	return func(gamepad *Gamepad) {
		gamepad.clickDuration = duration
	}
}

func WithHoldDuration(duration time.Duration) option {
	return func(gamepad *Gamepad) {
		gamepad.holdDuration = duration
	}
}

func (g *Gamepad) Close() error {
	g.cancel()
	return nil
}

// OnDPad subscribes to dpad events
func (g *Gamepad) OnDPad(h directionHandler) {
	g.dpadHandler = h
}

// OnLeftJoystick subscribes to left joystick move events
func (g *Gamepad) OnLeftJoystick(h directionHandler) {
	g.leftJoyHandler = h
}

// OnRightJoystick subscribes to right joystick move events
func (g *Gamepad) OnRightJoystick(h directionHandler) {
	g.rightJoyHandler = h
}

// OnL1 subscribes to L1 button events
func (g *Gamepad) OnL1(h buttonHandler, events ...ButtonEvent) {
	g.l1Btn = &button{
		handler: h,
		events:  events,
	}
}

// OnR1 subscribes to R1 button events
func (g *Gamepad) OnR1(h buttonHandler, events ...ButtonEvent) {
	g.r1Btn = &button{
		handler: h,
		events:  events,
	}
}

// OnL2 subscribes to L2 button events
func (g *Gamepad) OnL2(h buttonHandler, events ...ButtonEvent) {
	g.l2Btn = &button{
		handler: h,
		events:  events,
	}
}

// OnR2 subscribes to R2 button events
func (g *Gamepad) OnR2(h buttonHandler, events ...ButtonEvent) {
	g.r2Btn = &button{
		handler: h,
		events:  events,
	}
}

// OnSelect subscribes to select button events
func (g *Gamepad) OnSelect(h buttonHandler, events ...ButtonEvent) {
	g.selectBtn = &button{
		handler: h,
		events:  events,
	}
}

// OnStart subscribes to start button events
func (g *Gamepad) OnStart(h buttonHandler, events ...ButtonEvent) {
	g.startBtn = &button{
		handler: h,
		events:  events,
	}
}

// OnAnalog subscribes to analog button events
func (g *Gamepad) OnAnalog(h buttonHandler, events ...ButtonEvent) {
	g.analogBtn = &button{
		handler: h,
		events:  events,
	}
}

// OnLJ subscribes to left joystick click events
func (g *Gamepad) OnLJ(h buttonHandler, events ...ButtonEvent) {
	g.ljBtn = &button{
		handler: h,
		events:  events,
	}
}

// OnRJ subscribes to right joystick click events
func (g *Gamepad) OnRJ(h buttonHandler, events ...ButtonEvent) {
	g.rjBtn = &button{
		handler: h,
		events:  events,
	}
}

// OnCross subscribes to X button events
func (g *Gamepad) OnCross(h buttonHandler, events ...ButtonEvent) {
	g.crossBtn = &button{
		handler: h,
		events:  events,
	}
}

// OnCircle subscribes to O button events
func (g *Gamepad) OnCircle(h buttonHandler, events ...ButtonEvent) {
	g.circleBtn = &button{
		handler: h,
		events:  events,
	}
}

// OnSquare subscribes to [] events
func (g *Gamepad) OnSquare(h buttonHandler, events ...ButtonEvent) {
	g.squareBtn = &button{
		handler: h,
		events:  events,
	}
}

// OnTriangle subscribes to /\ button events
func (g *Gamepad) OnTriangle(h buttonHandler, events ...ButtonEvent) {
	g.triangleBtn = &button{
		handler: h,
		events:  events,
	}
}

func (g *Gamepad) handleEvents() {
	for {
		select {
		case event := <-g.device.OnButton():

			var pos ButtonPosition
			if event.Value <= 0 {
				pos = UpPosition
			} else {
				pos = DownPosition
			}

			switch event.Button {
			case CrossButton:
				if err := g.processButton(g.crossBtn, pos); err != nil {
					log.Println(err)
				}
			case CircleButton:
				if err := g.processButton(g.circleBtn, pos); err != nil {
					log.Println(err)
				}
			case SquareButton:
				if err := g.processButton(g.squareBtn, pos); err != nil {
					log.Println(err)
				}
			case TriangleButton:
				if err := g.processButton(g.triangleBtn, pos); err != nil {
					log.Println(err)
				}
			case L1Button:
				if err := g.processButton(g.l1Btn, pos); err != nil {
					log.Println(err)
				}
			case R1Button:
				if err := g.processButton(g.r1Btn, pos); err != nil {
					log.Println(err)
				}
			case SelectButton:
				if err := g.processButton(g.selectBtn, pos); err != nil {
					log.Println(err)
				}
			case StartButton:
				if err := g.processButton(g.startBtn, pos); err != nil {
					log.Println(err)
				}
			case AnalogButton:
				if err := g.processButton(g.analogBtn, pos); err != nil {
					log.Println(err)
				}
			case LeftJoyButton:
				if err := g.processButton(g.ljBtn, pos); err != nil {
					log.Println(err)
				}
			case RightJoyButton:
				if err := g.processButton(g.rjBtn, pos); err != nil {
					log.Println(err)
				}
			default:
				log.Printf("Button event, button: %v, value: %v, when: %v\n", event.Button, event.Value, event.When)
			}

		case event := <-g.device.OnAxis():
			g.axisCache[int(event.Axis)] = int(event.Value)

			if event.Axis == dPadXAxis || event.Axis == dPadYAxis {
				if err := g.emitDirection(g.dpadHandler, dPadXAxis, dPadYAxis); err != nil {
					log.Println(err)
				}
				continue
			}

			if event.Axis == leftJoyXAxis || event.Axis == leftJoyYAxis {
				if err := g.emitDirection(g.leftJoyHandler, leftJoyXAxis, leftJoyYAxis); err != nil {
					log.Println(err)
				}
				continue
			}

			if event.Axis == rightJoyXAxis || event.Axis == rightJoyYAxis {
				if err := g.emitDirection(g.rightJoyHandler, rightJoyXAxis, rightJoyYAxis); err != nil {
					log.Println(err)
				}
				continue
			}

			// L2 and R2 are axis but we want them as buttons
			var pos ButtonPosition
			if event.Value <= 0 {
				pos = UpPosition
			} else {
				pos = DownPosition
			}

			if event.Axis == l2Axis {
				if err := g.processButton(g.l2Btn, pos); err != nil {
					log.Println(err)
				}
				continue
			}

			if event.Axis == r2Axis {
				if err := g.processButton(g.r2Btn, pos); err != nil {
					log.Println(err)
				}
				continue
			}
			log.Printf("Axis event, axis: %v, value: %v, when: %v\n", event.Axis, event.Value, event.When)
		}
	}
}

func isGamepad(idx int) bool {
	d, err := os.ReadFile(fmt.Sprintf("/sys/class/input/js%v/device/name", idx))
	if err != nil {
		log.Printf("Error checking device name, err: %v", err)
		return false
	}
	name := strings.TrimSpace(string(d))
	return name == ExpectedName // The name PiHut gamepad resolves as
}

func (g *Gamepad) emitDirection(handler directionHandler, xIndex, yIndex int) error {
	if handler == nil {
		return errors.New("handler not assigned")
	}

	x := g.axisCache[xIndex]
	y := g.axisCache[yIndex]

	if g.invertY {
		y = y * -1
	}
	handler(float32(x)/hid.MaxValue, float32(y)/hid.MaxValue) // TODO scale to float
	return nil
}

func includes(events []ButtonEvent, event ButtonEvent) bool {
	if events == nil {
		return true
	}
	for _, e := range events {
		if e == event {
			return true
		}
	}

	return false
}

func (g *Gamepad) processButton(btn *button, pos ButtonPosition) error {
	if btn == nil {
		return errors.New("handler not assigned")
	}

	if btn.lastPosition == pos {
		return nil // Swallow duplicate events
	}

	switch pos {
	case DownPosition:
		btn.downTime = time.Now()
		if includes(btn.events, DownEvent) {
			btn.handler(ButtonEvent(pos))
		}
		if includes(btn.events, HoldEvent) {
			if btn.holdTimer != nil {
				btn.holdTimer.Stop()
			}

			btn.holdTimer = time.AfterFunc(g.holdDuration, func() {
				btn.handler(HoldEvent)
			})
		}
	case UpPosition:
		if btn.holdTimer != nil {
			btn.holdTimer.Stop()
		}

		if includes(btn.events, UpEvent) {
			btn.handler(ButtonEvent(pos))
		}

		if includes(btn.events, ClickEvent) {
			if time.Since(btn.downTime) < g.clickDuration {
				btn.handler(ClickEvent)
			} else {
				log.Printf("Invalid click, elapsed: %v, click dur: %v\n", time.Since(btn.downTime), g.clickDuration)
			}
		}
	}

	btn.lastPosition = pos
	return nil

}
