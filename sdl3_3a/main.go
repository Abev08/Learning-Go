package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jupiterrider/purego-sdl3/sdl"
	"github.com/jupiterrider/purego-sdl3/ttf"
)

// Text rendering.
// To render text:
// 1. Initialize SDL and SDL_ttf libraries
// 2. Load .ttf font file
// 3. Create window and renderer
// 4. Create ttf.TextEngine, I have chosen to use RendererTextEngine
// 5. Create ttf.Text struct from string text
// 6. Optionally set text wrap width and color
// 7. Draw the text
// The created ttf.Text struct can be reused in future frames (create once use multiple times).
// Also it can be modified by SDL_ttf library.
//
// Also simple performance metrics are implemented to compare with sdl2_7x examples.
// On my pc:
// - initialization part took around 500 us, which is faster than in sdl2_7x examples,
// - rendering first frame took around 15000 us, which is way slower than in sdl2_7x examples,
// - rendering frames 1-10 took around 0 us, which is way faster than in sdl2_7x examples.
//
// Calling ttf.UpdateText() on ttf.Text structure after it's creation reduces first draw call duration from ~15000 us to ~0 us.

func main() {
	// Initialize SDL3
	if !sdl.Init(sdl.InitVideo) {
		log.Fatalln(sdl.GetError())
	}
	defer sdl.Quit()

	// Create and start performance analysis
	var pCounter uint8
	var pDuration time.Duration
	pTime := time.Now()

	// Initialize SDL3_ttf
	if !ttf.Init() {
		log.Fatalln(sdl.GetError())
	}
	defer ttf.Quit()

	// Load .ttf font file
	font := ttf.OpenFont("assets/fonts/OpenSans-SemiBold.ttf", 24)
	if font == nil {
		log.Fatalln("couldn't load TTF font file.", sdl.GetError())
	}
	defer ttf.CloseFont(font)

	// Print TTF initialization and font loading time
	pDuration = time.Since(pTime)
	fmt.Printf("TTF initialization and loading font file took: %d us\n", pDuration.Microseconds())

	// Get SDL version
	version := make([]int32, 3)
	version[0], version[1], version[2] = sdl.GetVersion()
	log.Printf("SDL %d.%d.%d\n", version[0], version[1], version[2])
	version[0], version[1], version[2] = ttf.Version()
	log.Printf("SDL_tff %d.%d.%d\n", version[0], version[1], version[2])
	version[0], version[1], version[2] = ttf.GetFreeTypeVersion()
	log.Printf("FreeType %d.%d.%d\n", version[0], version[1], version[2])

	// Print TTF initialization and font loading time
	pDuration = time.Since(pTime)
	window := Window{BackgroundColor: sdl.Color{R: 40, G: 40, B: 40, A: 255}}
	defer window.Destroy()

	// Create window and renderer
	if !sdl.CreateWindowAndRenderer("Text rendering", 1280, 720, sdl.WindowResizable, &window.W, &window.R) {
		log.Fatalln(sdl.GetError())
	}
	sdl.SetRenderVSync(window.R, 2) // 0 - off, 1 - vsync, 2 - vsync/2, 3 - vsync/3, 4 - vsync/4

	window.TextEngine = ttf.CreateRendererTextEngine(window.R)
	if font == nil {
		log.Fatalln("couldn't create ttf.TextEngine.", sdl.GetError())
	}
	// Text to be rendered, 70 words
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam commodo nibh risus, quis vulputate arcu mattis at. Etiam massa libero, euismod in metus ut, venenatis efficitur ipsum. Etiam sed consectetur orci. Sed odio odio, aliquam in leo mattis, tincidunt rutrum ex. Proin cursus luctus magna, vel facilisis felis dapibus ac. Vivamus eu dui id ipsum hendrerit sodales vitae id nunc. Fusce vitae lacus tempor, blandit massa ut, consectetur urna. Maecenas."
	var maxTextWidth int32 = 1260 // window width - 10 px padding from left and right

	window.Text = ttf.CreateText(window.TextEngine, font, text, 0)
	if window.Text == nil {
		log.Fatalln("couldn't create ttf.Text.", sdl.GetError())
	}
	if !ttf.SetTextWrapWidth(window.Text, maxTextWidth) {
		log.Println("couldn't set text wrap width. ", sdl.GetError())
	}
	if !ttf.SetTextColor(window.Text, 220, 100, 220, 255) {
		log.Println("couldn't set text color. ", sdl.GetError())
	}
	ttf.UpdateText(window.Text)

	// Main loop
	running := true
	for running {
		// Process events
		event := sdl.Event{}
		for sdl.PollEvent(&event) {
			switch event.Type() {
			case sdl.EventQuit:
				log.Println("Window close requested")
				running = false

			case sdl.EventWindowResized:
				e := event.Window()
				// Change text wrap width on window resize - automatically fit text inside the window
				if !ttf.SetTextWrapWidth(window.Text, e.Data1-20) {
					log.Println("couldn't set text wrap width. ", sdl.GetError())
				}
			}
		}

		sdl.SetRenderDrawColor(window.R, window.BackgroundColor.R, window.BackgroundColor.G, window.BackgroundColor.B, window.BackgroundColor.A)
		sdl.RenderClear(window.R)

		// Render text
		pTime = time.Now() // Get start time
		ttf.DrawRendererText(window.Text, 10, 10)
		pDuration = time.Since(pTime)
		if pCounter < 10 {
			pCounter++
			fmt.Printf("Rendering text in frame %d took: %d us\n", pCounter, pDuration.Microseconds())
		}

		// ... more render operations ...

		sdl.RenderPresent(window.R)
	}
}

// A struct to encapsulate window related stuff
type Window struct {
	W          *sdl.Window     // The window
	R          *sdl.Renderer   // Renderer to the window
	TextEngine *ttf.TextEngine // TextEngine to create ttf.Text structs

	ID              sdl.WindowID // Window ID
	BackgroundColor sdl.Color    // Window background color

	Text *ttf.Text // Text to be rendered by the renderer
}

// Destroy the window and stuff used by it
func (w *Window) Destroy() {
	if w.Text != nil {
		ttf.DestroyText(w.Text)
	}
	if w.TextEngine != nil {
		ttf.DestroyRendererTextEngine(w.TextEngine)
	}
	if w.R != nil {
		sdl.DestroyRenderer(w.R)
	}
	if w.W != nil {
		sdl.DestroyWindow(w.W)
	}
}
