package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// Multiple windows.
// This example is based on previous one and is slightly more difficult.
// Using multiple windows means:
// - each window has to be created individually,
// - each window has to has own renderer,
// - processing events has to check from which window the event comes,
// - if using textures each window has to has own texture atlas,
// I'll use structs to simplify things.
// For this example only SDL2.dll is required.

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal("SDL2 cannot be initialized", err)
	}
	defer sdl.Quit()

	app1 := NewApp("Window 1", sdl.Point{X: 640, Y: 360})
	defer app1.Cleanup()
	app2 := NewApp("Window 2", sdl.Point{X: 640, Y: 360})
	defer app2.Cleanup()

	app1.BackgrouundColor = randColor()
	app2.BackgrouundColor = randColor()

	running := true
	updateColor := time.Now()
	inputs := InputState{MousePosition: sdl.Point{}}
	b1Rect := sdl.Rect{X: 100, Y: 100, W: 60, H: 25}
	b2Rect := sdl.Rect{X: 100, Y: 100, W: 60, H: 25}
	b1Counter, b2Counter := 0, 0
	for running {
		// Update background color every second
		if time.Since(updateColor).Milliseconds() >= 1000 {
			updateColor = time.Now()
			app1.BackgrouundColor = randColor()
			app2.BackgrouundColor = randColor()
		}

		// Process events
		inputs.NewFrame()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.WindowEvent:
				switch e.Event {
				case sdl.WINDOWEVENT_CLOSE:
					fmt.Println("Window close requested")
					running = false
				case sdl.WINDOWEVENT_ENTER:
					// Focus window on mouse enter (not really working when mouse is inside the window while it opens, but good enough)
					switch e.WindowID {
					case app1.WindowID:
						app1.W.Raise()
					case app2.WindowID:
						app2.W.Raise()
					}
				}

			case *sdl.MouseMotionEvent: // Mouse move event
				inputs.MouseOnWindowID = e.WindowID
				inputs.MousePosition.X, inputs.MousePosition.Y = e.X, e.Y

			case *sdl.MouseButtonEvent: // Mouse button state changed
				inputs.MouseLeftPressed = e.Button == sdl.BUTTON_LEFT && e.State == sdl.PRESSED
				inputs.MouseLeftClicked = e.Button == sdl.BUTTON_LEFT && e.State == sdl.RELEASED
				inputs.MouseRightPressed = e.Button == sdl.BUTTON_RIGHT && e.State == sdl.PRESSED
				inputs.MouseRightClicked = e.Button == sdl.BUTTON_RIGHT && e.State == sdl.RELEASED
			}
		}

		// Window 1 draw
		app1.R.SetDrawColor(app1.BackgrouundColor.R, app1.BackgrouundColor.G, app1.BackgrouundColor.B, app1.BackgrouundColor.A)
		app1.R.Clear()
		if DrawButton(app1.R, app1.WindowID, &inputs, &b1Rect) {
			b1Counter++
			fmt.Printf("Button on window 1 was clicked %d times\n", b1Counter)
		}
		app1.R.Present()

		// Window 2 draw
		app2.R.SetDrawColor(app2.BackgrouundColor.R, app2.BackgrouundColor.G, app2.BackgrouundColor.B, app2.BackgrouundColor.A)
		app2.R.Clear()
		if DrawButton(app2.R, app2.WindowID, &inputs, &b2Rect) {
			b2Counter++
			fmt.Printf("Button on window 2 was clicked %d times\n", b2Counter)
		}
		app2.R.Present()
	}
}

// Get random number in range [0,256)
func randByte() uint8 {
	return uint8(rand.Uint32N(256))
}

// Get random color
func randColor() sdl.Color {
	return sdl.Color{R: randByte(), G: randByte(), B: randByte(), A: 255}
}

// App struct that groups data related to sdl.Window
type App struct {
	W                *sdl.Window   // Window
	R                *sdl.Renderer // Renderer
	WindowID         uint32
	BackgrouundColor sdl.Color
}

// Cleans up after the app
func (app *App) Cleanup() {
	if app.R != nil {
		app.R.Destroy()
	}
	if app.W != nil {
		app.W.Destroy()
	}
}

// Creates new app (window and renderer)
func NewApp(title string, size sdl.Point) App {
	var err error
	app := App{}

	app.W, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, size.X, size.Y, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatal("error creating a window", err)
	}
	app.W.SetResizable(true) // Allow window resizing

	app.WindowID, err = app.W.GetID() // Get window ID
	if err != nil {
		log.Fatal("couldn't receive window ID", err)
	}

	app.R, err = sdl.CreateRenderer(app.W, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC) // Use VSync
	if err != nil {
		log.Fatal("error creating a renderer", err)
	}

	return app
}

// Input devices state
type InputState struct {
	MouseOnWindowID uint32
	MousePosition   sdl.Point
	MouseLeftPressed, MouseLeftClicked,
	MouseRightPressed, MouseRightClicked bool
}

// Reset input states for new frame
func (i *InputState) NewFrame() {
	i.MouseLeftClicked = false
	i.MouseRightClicked = false
}

// Draws a button on provided renderer. Returns true if the button was clicked, otherwise false
func DrawButton(renderer *sdl.Renderer, windowID uint32, inputs *InputState, rect *sdl.Rect) bool {
	clicked := false

	if windowID == inputs.MouseOnWindowID && inputs.MousePosition.InRect(rect) {
		if inputs.MouseLeftPressed {
			renderer.SetDrawColor(100, 100, 200, 255)
		} else {
			renderer.SetDrawColor(100, 100, 100, 255)
			if inputs.MouseLeftClicked {
				clicked = true
			}
		}
	} else {
		renderer.SetDrawColor(200, 200, 200, 255)
	}
	renderer.FillRect(rect)

	return clicked
}
