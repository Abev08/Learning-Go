package main

import (
	"log"
	"math/rand/v2"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// Initializing SDL2 window.
// SDL2 has external requirements before it can be build: https://github.com/veandco/go-sdl2?tab=readme-ov-file#requirements
// For example if "sdl.INIT_EVERYTHING" is not recognized it's a tip that some requirements are missing.
// If you followed my steps how to install GCC (MSYS2) from main README place "include" and "lib" folders in "C:\msys64\ucrt64\" directory.
// After restarting vscode SDL2 source files will be compiled - it may take long time.
// The same with first go build with SDL2 packages.
// SDL2 also requires to place it's .dll files into directory with compiled .exe. The files are in "bin" directories in downloaded .zip.
// For this example only SDL2.dll is required.
//
// In this example we are not processing any SDL2 events.
// The window won't respond to any events (clicking, closing, dragging, focus change, etc.).

func main() {
	// Initialize SDL2 library
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal("SDL2 cannot be initialized", err)
	}
	defer sdl.Quit() // Defer cleaning up

	// Create the window
	window, err := sdl.CreateWindow("Hello, World!", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 640, 360, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatal("error creating a window", err)
	}
	defer window.Destroy() // Defer cleaning up

	// Create renderer that can render window contents
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC) // Use VSync
	if err != nil {
		log.Fatal("error creating a renderer", err)
	}
	defer renderer.Destroy() // Defer cleaning up

	// Main loop
	updateColor := time.Now()
	backgroundColor := randColor()
	for {
		// Update background color every second
		if time.Since(updateColor).Milliseconds() >= 1000 {
			updateColor = time.Now()
			backgroundColor = randColor()
		}

		// Set drawing color
		renderer.SetDrawColor(backgroundColor.R, backgroundColor.G, backgroundColor.B, backgroundColor.A)
		renderer.Clear() // Clear window contents
		// More draw operations goes between Clear() and Present()
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
