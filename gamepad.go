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

//var KnownNames []string = []string{"Microsoft X-Box 360 pad", "SHANWAN Android Gamepad"}

const (
	InputTypeButton int = iota
	InputTypeAxis
)

type Input struct {
	Type  int
	Value uint8
}

type InputMapping map[Input]Resolved

var DriverMapping = map[hid.DriverName]InputMapping{
	// Ubuntu 22.04 arm64
	"Microsoft X-Box 360 pad": {
		//CrossButton:               0,
		Input{InputTypeButton, 0}: CrossButton,
		//CircleButton:              1,
		Input{InputTypeButton, 1}: CircleButton,
		//SquareButton:              2,
		Input{InputTypeButton, 2}: SquareButton,
		//TriangleButton:            3,
		Input{InputTypeButton, 3}: TriangleButton,
		//L1Button:                  4,
		Input{InputTypeButton, 4}: L1Button,
		//R1Button:                  5,
		Input{InputTypeButton, 5}: R1Button,
		//SelectButton:              6,
		Input{InputTypeButton, 6}: SelectButton,
		//StartButton:               7,
		Input{InputTypeButton, 7}: StartButton,
		//AnalogButton:              8,
		Input{InputTypeButton, 8}: AnalogButton,
		//LeftJoyButton:             9,
		Input{InputTypeButton, 9}: LeftJoyButton,
		//RightJoyButton:            10,
		Input{InputTypeButton, 10}: RightJoyButton,
		//DPadXAxis:                 6,
		Input{InputTypeAxis, 6}: DPadXAxis,
		//DPadYAxis:                 7,
		Input{InputTypeAxis, 7}: DPadYAxis,
		//LeftJoyXAxis:              0,
		Input{InputTypeAxis, 0}: LeftJoyXAxis,
		//LeftJoyYAxis:              1,
		Input{InputTypeAxis, 1}: LeftJoyYAxis,
		//RightJoyXAxis:             3,
		Input{InputTypeAxis, 3}: RightJoyXAxis,
		//RightJoyYAxis:             4,
		Input{InputTypeAxis, 4}: RightJoyYAxis,
		//L2Axis:                    2,
		Input{InputTypeAxis, 2}: L2Axis,
		//R2Axis:                    5,
		Input{InputTypeAxis, 5}: R2Axis,
	},

	// Raspberry pi 4 - Ubuntu 22.10
	"SHANWAN Android Gamepad": {
		//CrossButton:               0,
		Input{InputTypeButton, 0}: CrossButton,
		//CircleButton:              1,
		Input{InputTypeButton, 1}: CircleButton,
		//SquareButton:              2,
		Input{InputTypeButton, 3}: SquareButton,
		//TriangleButton:            3,
		Input{InputTypeButton, 4}: TriangleButton,
		//L1Button:                  4,
		Input{InputTypeButton, 6}: L1Button,
		//R1Button:                  5,
		Input{InputTypeButton, 7}: R1Button,
		//SelectButton:              6,
		Input{InputTypeButton, 10}: SelectButton,
		//StartButton:               7,
		Input{InputTypeButton, 11}: StartButton,
		//AnalogButton:              8,
		Input{InputTypeButton, 12}: AnalogButton,
		//LeftJoyButton:             9,
		Input{InputTypeButton, 13}: LeftJoyButton,
		//RightJoyButton:            10,
		Input{InputTypeButton, 14}: RightJoyButton,
		//DPadXAxis:                 6,
		Input{InputTypeAxis, 6}: DPadXAxis,
		//DPadYAxis:                 7,
		Input{InputTypeAxis, 7}: DPadYAxis,
		//LeftJoyXAxis:              0,
		Input{InputTypeAxis, 0}: LeftJoyXAxis,
		//LeftJoyYAxis:              1,
		Input{InputTypeAxis, 1}: LeftJoyYAxis,
		//RightJoyXAxis:             3,
		Input{InputTypeAxis, 2}: RightJoyXAxis,
		//RightJoyYAxis:             4,
		Input{InputTypeAxis, 3}: RightJoyYAxis,
		//L2Axis:                    2,
		Input{InputTypeAxis, 5}: L2Axis,
		//R2Axis:                    5,
		Input{InputTypeAxis, 4}: R2Axis,
	},
}

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
	axisCache     map[Resolved]int
	clickDuration time.Duration
	holdDuration  time.Duration
	inputMapping  InputMapping
	debug         bool

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
	device := hid.Connect(ctx)
	if device == nil {
		cancel()
		return nil, fmt.Errorf("failed to connect with device")
	}

	g := &Gamepad{
		ctx:           ctx,
		cancel:        cancel,
		device:        device,
		axisCache:     make(map[Resolved]int),
		clickDuration: defaultClickDuration,
		holdDuration:  defaultHoldDuration,
		inputMapping:  DriverMapping[device.Driver],
	}

	for _, o := range opts {
		o(g)
	}

	// Initialize axis cache with zero values
	for i := DPadXAxis; i <= R2Axis; i++ {
		g.axisCache[i] = 0
	}

	go g.handleEvents()

	return g, nil
}

func WithDebug() option {
	return func(gamepad *Gamepad) {
		gamepad.debug = true
	}
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

func (g *Gamepad) debugLn(s string) {
	if g.debug {
		log.Println(s)
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

			resolved, ok := g.inputMapping[Input{
				Type:  InputTypeButton,
				Value: event.Button,
			}]
			if !ok {
				g.debugLn(fmt.Sprintf("Button unknown: %v\n", event.Button))
				continue
			}

			if g.debug {
				g.debugLn(fmt.Sprintf("Button, input: %v, resolved as: %v\n", event.Button, resolved))
			}

			switch resolved {
			case CrossButton:
				if err := g.processButton(g.crossBtn, pos); err != nil {
					g.debugLn(err.Error())
				}
			case CircleButton:
				if err := g.processButton(g.circleBtn, pos); err != nil {
					g.debugLn(err.Error())
				}
			case SquareButton:
				if err := g.processButton(g.squareBtn, pos); err != nil {
					g.debugLn(err.Error())
				}
			case TriangleButton:
				if err := g.processButton(g.triangleBtn, pos); err != nil {
					g.debugLn(err.Error())
				}
			case L1Button:
				if err := g.processButton(g.l1Btn, pos); err != nil {
					g.debugLn(err.Error())
				}
			case R1Button:
				if err := g.processButton(g.r1Btn, pos); err != nil {
					g.debugLn(err.Error())
				}
			case SelectButton:
				if err := g.processButton(g.selectBtn, pos); err != nil {
					g.debugLn(err.Error())
				}
			case StartButton:
				if err := g.processButton(g.startBtn, pos); err != nil {
					g.debugLn(err.Error())
				}
			case AnalogButton:
				if err := g.processButton(g.analogBtn, pos); err != nil {
					g.debugLn(err.Error())
				}
			case LeftJoyButton:
				if err := g.processButton(g.ljBtn, pos); err != nil {
					g.debugLn(err.Error())
				}
			case RightJoyButton:
				if err := g.processButton(g.rjBtn, pos); err != nil {
					g.debugLn(err.Error())
				}
			default:
				g.debugLn(fmt.Sprintf("Button event, button: %v, value: %v, when: %v\n", event.Button, event.Value, event.When))

			}

		case event := <-g.device.OnAxis():
			resolved, ok := g.inputMapping[Input{
				Type:  InputTypeAxis,
				Value: event.Axis,
			}]
			if !ok {
				g.debugLn(fmt.Sprintf("Button unknown: %v\n", event.Axis))
				continue
			}

			g.debugLn(fmt.Sprintf("Axis, input: %v, resolved as: %v\n", event.Axis, resolved))

			g.axisCache[resolved] = int(event.Value)

			if resolved == DPadXAxis || resolved == DPadYAxis {
				if err := g.emitDirection(g.dpadHandler, DPadXAxis, DPadYAxis); err != nil {
					g.debugLn(err.Error())
				}
				continue
			}

			if resolved == LeftJoyXAxis || resolved == LeftJoyYAxis {
				if err := g.emitDirection(g.leftJoyHandler, LeftJoyXAxis, LeftJoyYAxis); err != nil {
					g.debugLn(err.Error())
				}
				continue
			}

			if resolved == RightJoyXAxis || resolved == RightJoyYAxis {
				if err := g.emitDirection(g.rightJoyHandler, RightJoyXAxis, RightJoyYAxis); err != nil {
					if g.debug {
						g.debugLn(err.Error())
					}
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

			if resolved == L2Axis {
				if err := g.processButton(g.l2Btn, pos); err != nil {
					g.debugLn(err.Error())
				}
				continue
			}

			if resolved == R2Axis {
				if err := g.processButton(g.r2Btn, pos); err != nil {
					g.debugLn(err.Error())
				}
				continue
			}

			g.debugLn(fmt.Sprintf("Axis event, axis: %v, value: %v, when: %v\n", event.Axis, event.Value, event.When))
		}
	}
}

func isGamepad(idx int) (hid.DriverName, bool) {
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

func (g *Gamepad) emitDirection(handler directionHandler, xIndex, yIndex Resolved) error {
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
				g.debugLn(fmt.Sprintf("Invalid click, elapsed: %v, click dur: %v\n", time.Since(btn.downTime), g.clickDuration))
			}
		}
	}

	btn.lastPosition = pos
	return nil

}
