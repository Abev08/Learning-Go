package main

import (
	"fmt"
	"log"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Text rendering. Simplest approach.
// Every frame text is converted into sdl.Surface, rendered into sdl.Texture and then rendered on screen.
// This approach is the slowest, uses a lot of resources and queues a lot of work for garbage collector.
// But it allows to update displayed text in every frame, which may be required in some applications.
// Also simple performance metrics are implemented for comparasion to other text rendering approaches.
// This example is based on "sdl2_5".
// For this example SDL2.dll and SDL2_ttf.dll are required.
//
// Longer text on my pc:
// initialization part took: 100-1000 us, average 300 us
// rendering first 10 frames took (excluding first frame) took: 500-1500 us, average 500 us

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal("SDL2 cannot be initialized. ", err)
	}
	defer sdl.Quit()

	// Text to be rendered
	// text := "Hello,\nWorld!"
	// Longer text, 70 words
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam commodo nibh risus, quis vulputate arcu mattis at. Etiam massa libero, euismod in metus ut, venenatis efficitur ipsum. Etiam sed consectetur orci. Sed odio odio, aliquam in leo mattis, tincidunt rutrum ex. Proin cursus luctus magna, vel facilisis felis dapibus ac. Vivamus eu dui id ipsum hendrerit sodales vitae id nunc. Fusce vitae lacus tempor, blandit massa ut, consectetur urna. Maecenas."

	// Create and start performace analysis
	var pCounter uint8
	var pDuration time.Duration
	pTime := time.Now()

	// Initialize TTF SDL2 package
	err = ttf.Init()
	if err != nil {
		log.Fatal("SDL2 TTF cannot be initialized. ", err)
	}
	defer ttf.Quit()

	// Load font file
	font24, err := ttf.OpenFont("assets/fonts/OpenSans-SemiBold.ttf", 24)
	if err != nil {
		log.Fatal("couldn't load TTF font file. ", err)
	}
	defer font24.Close()

	// Print TTF initialization and font loading time
	pDuration = time.Since(pTime)
	fmt.Printf("TTF initialization and loading font file took: %d us\n", pDuration.Microseconds())

	app := NewApp("Window", sdl.Point{X: 640, Y: 360})
	defer app.Cleanup()
	app.BackgroundColor = sdl.Color{R: 40, G: 40, B: 40, A: 255}
	maxTextWidth := 620 // window width - 10 px padding from left and right

	running := true
	inputs := InputState{MousePosition: sdl.Point{}}
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
		app.R.SetDrawColor(app.BackgroundColor.R, app.BackgroundColor.G, app.BackgroundColor.B, app.BackgroundColor.A)
		app.R.Clear()

		// Text rendering
		pTime = time.Now() // Get start time

		// Create sdl.Surface
		s, err := font24.RenderUTF8BlendedWrapped(text, sdl.Color{R: 255, G: 255, B: 255, A: 255}, maxTextWidth)
		if err != nil {
			log.Fatal(err)
		}
		// defer s.Free() // Defer doesn't work because we do not leave the main loop

		// Create sdl.Texture from created surface
		t, err := app.R.CreateTextureFromSurface(s)
		if err != nil {
			log.Fatal(err)
		}
		// defer t.Destroy() // Defer doesn't work because we do not leave the main loop

		// Render created text texture
		app.R.Copy(t, nil, &sdl.Rect{X: 10, Y: 10, W: s.W, H: s.H})

		// Manual release of resources is required
		t.Destroy()
		s.Free()

		// Print how long it took to render one frame
		pDuration = time.Since(pTime)
		if pCounter < 10 {
			pCounter++
			fmt.Printf("Rendering text in frame %d took: %d us\n", pCounter, pDuration.Microseconds())
		}

		app.R.Present()
	}
}

// App struct that groups data related to sdl.Window
type App struct {
	W               *sdl.Window   // Window
	R               *sdl.Renderer // Renderer
	WindowID        uint32
	BackgroundColor sdl.Color
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

// Ternary if statement
func Tif[T any](condition bool, vTrue, vFalse T) T {
	if condition {
		return vTrue
	}
	return vFalse
}
