package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// Processing events.
// This example is based on previous one.
// Also very simple button was created and mouse clicks on this button are detected.
// For this example only SDL2.dll is required.

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal("SDL2 cannot be initialized", err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Hello, World!", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 640, 360, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatal("error creating a window", err)
	}
	defer window.Destroy()
	window.SetResizable(true) // Allow window resizing

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC) // Use VSync
	if err != nil {
		log.Fatal("error creating a renderer", err)
	}
	defer renderer.Destroy()

	running := true
	updateColor := time.Now()
	backgroundColor := randColor()
	mousePos := sdl.Point{}
	mousePressed, mouseClick := false, false
	buttonRect := sdl.Rect{X: 100, Y: 100, W: 60, H: 25}
	counter := 0
	for running {
		// Update background color every second
		if time.Since(updateColor).Milliseconds() >= 1000 {
			updateColor = time.Now()
			backgroundColor = randColor()
		}

		// Process events
		mouseClick = false
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			// "*sdl.QuitEvent" can be used instead of checking if event is of type "*sdl.WindowEvent"
			// and then checking what kind of window event it is (doesn't work with multiple windows)
			// case *sdl.QuitEvent:
			// 	running = false
			case *sdl.WindowEvent: // Window events
				// You can check all of the window events here: https://wiki.libsdl.org/SDL2/SDL_WindowEventID
				switch e.Event {
				case sdl.WINDOWEVENT_SHOWN:
					fmt.Println("Window displayed")
				case sdl.WINDOWEVENT_CLOSE:
					fmt.Println("Window close requested")
					running = false
				case sdl.WINDOWEVENT_ENTER:
					fmt.Println("Mouse cursor entered the window")
				case sdl.WINDOWEVENT_LEAVE:
					fmt.Println("Mouse cursor left the window")
				case sdl.WINDOWEVENT_FOCUS_GAINED:
					fmt.Println("Window got focus")
				case sdl.WINDOWEVENT_FOCUS_LOST:
					fmt.Println("Window lost focus")
				case sdl.WINDOWEVENT_EXPOSED:
					fmt.Println("Window is on top of other windows")
				case sdl.WINDOWEVENT_HIDDEN:
					fmt.Println("Window is hidden")
				case sdl.WINDOWEVENT_MINIMIZED:
					fmt.Println("Window is minimized")
				case sdl.WINDOWEVENT_MAXIMIZED:
					fmt.Println("Window is maximized")
				case sdl.WINDOWEVENT_RESIZED:
					fmt.Printf("Window is resized. New window size: (%d, %d)\n", e.Data1, e.Data2)
				case sdl.WINDOWEVENT_MOVED:
					fmt.Printf("Window position changed. New window position: (%d, %d)\n", e.Data1, e.Data2)
				}

			case *sdl.KeyboardEvent: // Keyboard events
				if e.State == sdl.RELEASED && e.Repeat == 0 {
					if e.Keysym.Sym == sdl.K_ESCAPE {
						// Esc => close window
						fmt.Println("Window close requested")
						running = false
					} else if (e.Keysym.Sym == sdl.K_RETURN && (e.Keysym.Mod&sdl.KMOD_RALT) > 0) ||
						e.Keysym.Sym == sdl.K_F11 {
						// Right Alt + Enter or F11 => toggle fullscreen
						// sdl.WINDOW_FULLSCREEN changes monitor resolution to match window resolution - true fullscreen
						// sdl.WINDOW_FULLSCREEN_DESKTOP changes window size to match monitor resolution - fullscreen borderless
						fmt.Println("Toggle fullscreen requested")
						if window.GetFlags()&sdl.WINDOW_FULLSCREEN_DESKTOP > 0 {
							window.SetFullscreen(0)
						} else {
							window.SetFullscreen(sdl.WINDOW_FULLSCREEN_DESKTOP)
						}
					}
				}

			case *sdl.MouseMotionEvent: // Mouse move event
				fmt.Printf("Mouse moved from (%d, %d) to (%d, %d). Position changed by (%d, %d)\n", mousePos.X, mousePos.Y, e.X, e.Y, e.XRel, e.YRel)
				mousePos.X, mousePos.Y = e.X, e.Y

			case *sdl.MouseButtonEvent: // Mouse button state changed
				mousePressed = e.Button == sdl.BUTTON_LEFT && e.State == sdl.PRESSED
				mouseClick = e.Button == sdl.BUTTON_LEFT && e.State == sdl.RELEASED
				if mouseClick {
					fmt.Println("Mouse left button click")
				}
			}
		}

		// Set drawing color
		renderer.SetDrawColor(backgroundColor.R, backgroundColor.G, backgroundColor.B, backgroundColor.A)
		renderer.Clear() // Clear window contents

		// Very simple button
		if mousePos.InRect(&buttonRect) {
			// If point (mouse position) is inside a rect (button rectangle) change drawing color
			if mousePressed {
				// If mouse left button is pressed set different color
				renderer.SetDrawColor(100, 100, 200, 255)
			} else {
				// Mouse is over the simple button but mouse left button is not pressed, draw with "button hovered" color
				renderer.SetDrawColor(100, 100, 100, 255)
				if mouseClick {
					// Mouse click detected, the simple button was pressed
					counter++
					fmt.Printf("Simple button was clicked %d times\n", counter)
				}
			}
		} else {
			// Mouse is outside the button, draw the button with default color
			renderer.SetDrawColor(200, 200, 200, 255)
		}
		renderer.FillRect(&buttonRect) // Draw the button

		renderer.Present() // Present new window contents
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
