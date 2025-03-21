package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Create window and renderer.
// Enable VSync on created renderer.
// In the main loop:
// - process events (a lot of them, just for testing),
// - change background color every 1 sec,

func main() {
	// Get SDL version
	sdlVersion := make([]int32, 3)
	sdlVersion[0], sdlVersion[1], sdlVersion[2] = sdl.GetVersion()
	log.Printf("SDL %d.%d.%d\n", sdlVersion[0], sdlVersion[1], sdlVersion[2])

	// Initialize SDL3
	if !sdl.Init(sdl.InitVideo) {
		log.Fatalln(sdl.GetError())
	}
	defer sdl.Quit()

	// Enable VSync, it has to be called before creating the renderer
	// if !sdl.SetHint(sdl.HintRenderVsync, "1") {
	// 	log.Println(sdl.GetError())
	// }

	// Create window and renderer
	var window *sdl.Window
	var renderer *sdl.Renderer
	if !sdl.CreateWindowAndRenderer("Hello, World!", 1280, 720, sdl.WindowResizable, &window, &renderer) {
		log.Fatalln(sdl.GetError())
	}
	defer sdl.DestroyRenderer(renderer)
	defer sdl.DestroyWindow(window)

	// This can be used any time after creating renderer, not like sdl.SetHint(sdl.HintRenderVsync, "1")
	sdl.SetRenderVSync(renderer, 2) // 0 - off, 1 - vsync, 2 - vsync/2, 3 - vsync/3, 4 - vsync/4
	
	// Some variables
	mousePos := sdl.FPoint{}
	updateColor := time.Now()
	backgroundColor := getRandColor()
	buttonClicks := 0
	var mousePressed, mouseClicked bool

	// Main loop
	running := true
	for running {
		mouseClicked = false // Reset user input state

		// Process events
		event := sdl.Event{}
		for sdl.PollEvent(&event) {
			switch event.Type() {
			case sdl.EventQuit:
				log.Println("Window close requested")
				running = false
			case sdl.EventWindowShown:
				log.Println("Window displayed")
			case sdl.EventWindowMouseEnter:
				log.Println("Mouse cursor entered the window")
			case sdl.EventWindowMouseLeave:
				log.Println("Mouse cursor left the window")
			case sdl.EventWindowFocusGained:
				log.Println("Window got focus")
			case sdl.EventWindowFocusLost:
				log.Println("Window lost focus")
			case sdl.EventWindowExposed:
				log.Println("Window is on top of other windows")
			case sdl.EventWindowHidden:
				log.Println("Window is hidden")
			case sdl.EventWindowMinimized:
				log.Println("Window is minimized")
			case sdl.EventWindowMaximized:
				log.Println("Window is maximized")
			case sdl.EventWindowResized:
				// This one seems that is not implemented correctly, yet
				e := event.Motion()
				log.Printf("Window is resized. New window size: (%d, %d)\n", e.Which, e.State)
			case sdl.EventWindowMoved:
				// This one seems that is not implemented correctly, yet
				e := event.Motion()
				log.Printf("Window position changed. New window position: (%d, %d)\n", e.Which, e.State)
			case sdl.EventMouseButtonDown:
				mousePressed = true
			case sdl.EventMouseButtonUp:
				mousePressed = false
				e := event.Button()
				if e.Clicks == 2 {
					log.Printf("Double clicked mouse button '%d' at position (X: %v, Y: %v)\n", e.Button, e.X, e.Y)
				} else {
					log.Printf("Clicked mouse button '%d' at position (X: %v, Y: %v)\n", e.Button, e.X, e.Y)
				}
				mouseClicked = true
			case sdl.EventMouseMotion:
				e := event.Motion()
				mousePos.X, mousePos.Y = e.X, e.Y
				// log.Printf("Cursor moved to position (X: %v, Y: %v), relative to previous position (Xdiff: %v, Ydiff: %v)\n", e.X, e.Y, e.Xrel, e.Yrel)
			}
		}

		// Update background color every second
		if time.Since(updateColor).Milliseconds() >= 1000 {
			updateColor = time.Now()
			backgroundColor = getRandColor()
		}

		sdl.SetRenderDrawColor(renderer, backgroundColor.R, backgroundColor.G, backgroundColor.B, backgroundColor.A)
		sdl.RenderClear(renderer) // Clear the screen

		// Create simple "button"
		buttonRect := sdl.FRect{X: 50, Y: 100, W: 120, H: 60}
		if sdl.PointInRectFloat(mousePos, buttonRect) {
			// If point (mouse position) is inside a rect (button rectangle) change drawing color
			if mousePressed {
				// If mouse left button is pressed set different color
				sdl.SetRenderDrawColor(renderer, 100, 100, 200, 255)
			} else {
				// Mouse is over the simple button but mouse left button is not pressed, draw with "button hovered" color
				sdl.SetRenderDrawColor(renderer, 100, 100, 100, 255)
				if mouseClicked {
					// Mouse click detected, the simple button was pressed
					buttonClicks++
					fmt.Printf("Simple button was clicked %d times\n", buttonClicks)
				}
			}
		} else {
			// Mouse is outside the button, draw the button with default color
			sdl.SetRenderDrawColor(renderer, 200, 200, 200, 255)
		}
		sdl.RenderFillRect(renderer, &buttonRect) // Draw the button

		// ... more render operations ...

		sdl.RenderPresent(renderer) // Present new window contents
	}
}

// Get random number in range 0-255
func randByte() uint8 {
	return uint8(rand.Uint32N(256))
}

// Get random color
func getRandColor() sdl.Color {
	return sdl.Color{R: randByte(), G: randByte(), B: randByte(), A: 255}
}
