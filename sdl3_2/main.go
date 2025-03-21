package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/jupiterrider/purego-sdl3/img"
	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Create two windows (use windows structs to keep stuff related to the window together).
// Using multiple windows means:
// - each window has to be created individually,
// - each window has to has own renderer,
// - processing events has to check from which window the event comes,
// - if using textures each window has to has own texture atlas.
// Also draw two different textures on both of the windows (same position and size for simplicity).
// One of the textures will switch between drawing part of it (single sprite) or entire texture (atlas),
// clicking the button will switch between drawing modes.
// From texture atlas create animation struct that will allow animating the sprite.

var windows = make([]Window, 2)
var inputsState = InputState{}

func main() {
	// Get SDL version
	version := make([]int32, 3)
	version[0], version[1], version[2] = sdl.GetVersion()
	log.Printf("SDL %d.%d.%d\n", version[0], version[1], version[2])
	version[0], version[1], version[2] = img.Version()
	log.Printf("SDL_image %d.%d.%d\n", version[0], version[1], version[2])

	// Initialize SDL3
	if !sdl.Init(sdl.InitVideo) {
		log.Fatalln(sdl.GetError())
	}
	defer sdl.Quit()

	defer Cleanup() // Defer cleaning up the windows

	// Create window and renderer
	if !sdl.CreateWindowAndRenderer("1st window", 1280, 720, sdl.WindowResizable, &windows[0].W, &windows[0].R) {
		log.Fatalln(sdl.GetError())
	}
	if !sdl.CreateWindowAndRenderer("2nd window", 1280, 720, sdl.WindowResizable, &windows[1].W, &windows[1].R) {
		log.Fatalln(sdl.GetError())
	}

	// FIXME SDL3 Go bindings are missing sdl.GetWindowID(), for now assume WindowID 3 and 4
	// for i := range windows {
	// 	w := &windows[i]
	// 	w.ID = sdl.GetWindowID(w.W)
	// }
	windows[0].ID = 3
	windows[1].ID = 4

	// Enable VSync. This can be used any time after creating renderer, not like sdl.SetHint(sdl.HintRenderVsync, "1")
	for i := range windows {
		sdl.SetRenderVSync(windows[i].R, 2) // 0 - off, 1 - vsync, 2 - vsync/2, 3 - vsync/3, 4 - vsync/4
	}

	// Some variables
	for i := range windows {
		windows[i].BackgroundColor = getRandColor()
	}

	// Load image into surface (CPU pixel data) and create texture (GPU pixel data) from loaded surface
	peepoHappySurface := img.Load("assets/peepoHappy-4x.png")
	defer sdl.DestroySurface(peepoHappySurface)
	peepoHeySurface := img.Load("assets/peepoHey-4x.png")
	for i := range windows {
		w := &windows[i]
		w.peepoHappyTexture = sdl.CreateTextureFromSurface(w.R, peepoHappySurface)
		w.peepoHeyTexture = sdl.CreateTextureFromSurface(w.R, peepoHeySurface)

		frameDur := 100
		if i == 1 {
			frameDur = 50 // Twice the speed for 2nd window
		}

		w.peepoHey = &Animation{
			StartPosition: sdl.Point{X: 0, Y: 0},
			FrameSize:     sdl.Point{X: 128, Y: 128},
			KeyFrames:     6,
			FrameDuration: uint16(frameDur),
		}
	}
	var peepoHappyScale int32 = 2

	// Main loop
	running := true
	updateColor := time.Now()
	for running {
		inputsState.MouseLeftClicked = false

		// Process events
		event := sdl.Event{}
		for sdl.PollEvent(&event) {
			switch event.Type() {
			case sdl.EventQuit:
				fallthrough
			case sdl.EventWindowCloseRequested:
				log.Println("Window close requested")
				running = false

			case sdl.EventMouseMotion:
				e := event.Motion()
				inputsState.MouseOnWindowID = e.WindowID
				inputsState.MousePosition.X, inputsState.MousePosition.Y = e.X, e.Y

			case sdl.EventMouseButtonDown:
				e := event.Button()
				inputsState.MouseOnWindowID = e.WindowID
				inputsState.MouseLeftPressed = e.Button == uint8(sdl.ButtonLeft)
			case sdl.EventMouseButtonUp:
				e := event.Button()
				inputsState.MouseOnWindowID = e.WindowID
				inputsState.MouseLeftPressed = false
				inputsState.MouseLeftClicked = e.Button == uint8(sdl.ButtonLeft)

			case sdl.EventWindowMouseEnter:
				e := event.Window()
				// "Autofocus" the window when mouse enters it
				for i := range windows {
					w := &windows[i]
					if e.WindowID == w.ID {
						// FIXME SDL3 Go bindings are missing sdl.RaiseWindow()
						// sdl.RaiseWindow(w.W)
					}
				}
			}
		}

		// Update background color every second
		if time.Since(updateColor).Milliseconds() >= 1000 {
			updateColor = time.Now()
			for i := range windows {
				windows[i].BackgroundColor = getRandColor()
			}
		}

		// Draw all of the windows
		for i := range windows {
			w := &windows[i]
			sdl.SetRenderDrawColor(w.R, w.BackgroundColor.R, w.BackgroundColor.G, w.BackgroundColor.B, w.BackgroundColor.A)
			sdl.RenderClear(w.R) // Clear the screen

			// Draw simple "button"
			if drawButton(w) {
				w.buttonClicks++
				fmt.Printf("Button on window %d was clicked %d times\n", (i + 1), w.buttonClicks)
			}

			// Draw peepoHappy texture
			sdl.RenderTexture(w.R, w.peepoHappyTexture, nil,
				&sdl.FRect{X: 200, Y: 100, W: float32(w.peepoHappyTexture.W * peepoHappyScale), H: float32(w.peepoHappyTexture.H * peepoHappyScale)})
			// Draw peepoHey texture
			if w.buttonClicks%2 == 0 {
				// Draw single frame (2nd frame - offset by 128 pixels)
				sdl.RenderTexture(w.R, w.peepoHeyTexture,
					&sdl.FRect{X: 128, Y: 0, W: 128, H: 128},
					&sdl.FRect{X: 100, Y: 400, W: 128, H: float32(w.peepoHeyTexture.H)})
			} else {
				// Draw entire atlas
				sdl.RenderTexture(w.R, w.peepoHeyTexture,
					nil,
					&sdl.FRect{X: 100, Y: 400, W: float32(w.peepoHeyTexture.W), H: float32(w.peepoHeyTexture.H)})
			}

			// Draw peepoHey animation
			if !w.peepoHey.Started {
				w.peepoHey.Start()
			} else {
				w.peepoHey.Update()
			}
			sdl.RenderTexture(w.R, w.peepoHeyTexture,
				&sdl.FRect{
					X: float32(w.peepoHey.StartPosition.X + w.peepoHey.FrameSize.X*int32(w.peepoHey.Frame)), Y: float32(w.peepoHey.StartPosition.Y),
					W: float32(w.peepoHey.FrameSize.X), H: float32(w.peepoHey.FrameSize.Y)},
				&sdl.FRect{X: 700, Y: 100, W: float32(w.peepoHey.FrameSize.X), H: float32(w.peepoHey.FrameSize.Y)})

			// ... more render operations ...

			sdl.RenderPresent(w.R) // Present new window contents
		}
	}
}

// Cleans up after the app (destroys windows and renderers, etc.)
func Cleanup() {
	for i := range windows {
		w := &windows[i]
		if w.peepoHappyTexture != nil {
			sdl.DestroyTexture(w.peepoHappyTexture)
		}
		if w.R != nil {
			sdl.DestroyRenderer(w.R)
		}
		if w.W != nil {
			sdl.DestroyWindow(w.W)
		}
	}
}

// A struct to encapsulate window related stuff
type Window struct {
	W *sdl.Window   // The window
	R *sdl.Renderer // Renderer to the window

	ID              sdl.WindowID // Window ID
	BackgroundColor sdl.Color

	buttonClicks      int
	peepoHappyTexture *sdl.Texture
	peepoHeyTexture   *sdl.Texture
	peepoHey          *Animation
}

// Get random number in range 0-255
func randByte() uint8 {
	return uint8(rand.Uint32N(256))
}

// Get random color
func getRandColor() sdl.Color {
	return sdl.Color{R: randByte(), G: randByte(), B: randByte(), A: 255}
}

// Input devices state
type InputState struct {
	MouseOnWindowID                    sdl.WindowID
	MousePosition                      sdl.FPoint
	MouseLeftPressed, MouseLeftClicked bool
}

// Draws a button on provided window. Returns true if the button was clicked, otherwise false
func drawButton(w *Window) bool {
	clicked := false
	buttonRect := sdl.FRect{X: 50, Y: 100, W: 120, H: 60}
	if inputsState.MouseOnWindowID == w.ID && sdl.PointInRectFloat(inputsState.MousePosition, buttonRect) {
		// If point (mouse position) is inside a rect (button rectangle) change drawing color
		if inputsState.MouseLeftPressed {
			// If mouse left button is pressed set different color
			sdl.SetRenderDrawColor(w.R, 100, 100, 200, 255)
		} else {
			// Mouse is over the simple button but mouse left button is not pressed, draw with "button hovered" color
			sdl.SetRenderDrawColor(w.R, 100, 100, 100, 255)
			if inputsState.MouseLeftClicked {
				// Mouse click detected, the simple button was pressed
				clicked = true
			}
		}
	} else {
		// Mouse is outside the button, draw the button with default color
		sdl.SetRenderDrawColor(w.R, 200, 200, 200, 255)
	}
	sdl.RenderFillRect(w.R, &buttonRect) // Draw the button

	return clicked
}

type Animation struct {
	Started        bool      // Was the animation started?
	StartPosition  sdl.Point // Pixel position in atlas texture
	FrameSize      sdl.Point // Size of single frame
	KeyFrames      uint16    // Number of frames
	Frame          uint16    // Current frame
	FrameDuration  uint16    // Duration of single frame in milliseconds
	FrameTimeStart int64     // Time in milliseconds when frame started
}

// Starts the animation, also returns pointer to self
func (a *Animation) Start() *Animation {
	a.Started = true
	a.Frame = 0
	a.FrameTimeStart = time.Now().UnixMilli()
	return a
}

// Updates state of the animation
func (a *Animation) Update() bool {
	finished := false
	if time.Now().UnixMilli()-a.FrameTimeStart >= int64(a.FrameDuration) {
		a.Frame++
		if a.Frame >= a.KeyFrames {
			finished = true
			a.Frame = a.Frame % a.KeyFrames
		}
		a.FrameTimeStart = time.Now().UnixMilli()
	}

	if finished {
		a.Started = false
	}
	return finished
}
