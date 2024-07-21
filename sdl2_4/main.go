package main

import (
	"fmt"
	"log"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

// Textures, static image.
// This example is based on previous one.
// For this example SDL2.dll and SDL2_image.dll are required.

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal("SDL2 cannot be initialized", err)
	}
	defer sdl.Quit()

	app := NewApp("Window", sdl.Point{X: 800, Y: 450})
	defer app.Cleanup()

	app.BackgrouundColor = sdl.Color{R: 40, G: 40, B: 40, A: 255}

	// Load image and create texture from it
	slime, err := img.Load("assets/slime.png")
	if err != nil {
		log.Fatal("couldn't load image", err)
	}
	slimeTexture, err := app.R.CreateTextureFromSurface(slime)
	if err != nil {
		log.Fatal("couldn't create texture from loaded image (surface)", err)
	}
	defer slimeTexture.Destroy()

	running := true
	inputs := InputState{MousePosition: sdl.Point{}}
	drawEntireAtlas := false
	for running {
		// Process events
		inputs.NewFrame()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.WindowEvent:
				switch e.Event {
				case sdl.WINDOWEVENT_CLOSE:
					fmt.Println("Window close requested")
					running = false
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

		// Drawing
		app.R.SetDrawColor(app.BackgrouundColor.R, app.BackgrouundColor.G, app.BackgrouundColor.B, app.BackgrouundColor.A)
		app.R.Clear()

		// Draw entire atlas button toggle
		if DrawButton(app.R, app.WindowID, &inputs, &sdl.Rect{X: 5, Y: 5, W: 60, H: 20}) {
			drawEntireAtlas = !drawEntireAtlas
		}

		// Draw slime or entire atlas
		if drawEntireAtlas {
			// Draw entire atlas, also increase texture size x4
			app.R.Copy(slimeTexture, &sdl.Rect{X: 0, Y: 0, W: slime.W, H: slime.H}, &sdl.Rect{X: 10, Y: 30, W: slime.W * 4, H: slime.H * 4})
		} else {
			// Draw portion of slime texture, also increase texture size x4
			app.R.Copy(slimeTexture, &sdl.Rect{X: 0, Y: 0, W: 32, H: 32}, &sdl.Rect{X: 10, Y: 30, W: 128, H: 128})
		}

		app.R.Present()
	}
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
