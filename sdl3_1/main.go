package main

import (
	"log"
	"math/rand/v2"
	"time"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

func main() {
	sdlVersion := make([]int32, 3)
	sdlVersion[0], sdlVersion[1], sdlVersion[2] = sdl.GetVersion()
	log.Printf("SDL %d.%d.%d\n", sdlVersion[0], sdlVersion[1], sdlVersion[2])

	// Initialize SDL3
	if !sdl.Init(sdl.InitVideo) {
		log.Fatalln(sdl.GetError())
	}
	defer sdl.Quit()

	// Enable VSync, it has to be called before creating the renderer
	if !sdl.SetHint(sdl.HintRenderVsync, "1") {
		log.Println(sdl.GetError())
	}

	// Create window and renderer
	var window *sdl.Window
	var renderer *sdl.Renderer
	if !sdl.CreateWindowAndRenderer("Hello, World!", 1280, 720, sdl.WindowResizable, &window, &renderer) {
		log.Fatalln(sdl.GetError())
	}
	defer sdl.DestroyRenderer(renderer)
	defer sdl.DestroyWindow(window)

	// Main loop
	running := true
	updateColor := time.Now()
	setRandColor(renderer)
	for running {
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
			case sdl.EventMouseButtonUp:
				e := event.Button()
				if e.Clicks == 2 {
					log.Printf("Double clicked mouse button '%d' at position (X: %v, Y: %v)\n", e.Button, e.X, e.Y)
				} else {
					log.Printf("Clicked mouse button '%d' at position (X: %v, Y: %v)\n", e.Button, e.X, e.Y)
				}
			case sdl.EventMouseMotion:
				// e := event.Motion()
				// log.Printf("Cursor moved to position (X: %v, Y: %v), relative to previous position (Xdiff: %v, Ydiff: %v)\n", e.X, e.Y, e.Xrel, e.Yrel)
			}
		}

		// Update background color every second
		if time.Since(updateColor).Milliseconds() >= 1000 {
			updateColor = time.Now()
			setRandColor(renderer)
		}

		sdl.RenderClear(renderer) // Clear the screen
		// ... more render operations ...
		sdl.RenderPresent(renderer) // Present new window contents
	}
}

// Get random number in range 0-255
func randByte() uint8 {
	return uint8(rand.Uint32N(256))
}

// Set renderer draw random color
func setRandColor(renderer *sdl.Renderer) {
	sdl.SetRenderDrawColor(renderer, randByte(), randByte(), randByte(), 255)
}
