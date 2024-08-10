package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// Geometry rendering.
// Rendering custom shapes from raw vertex data using build in SDL RenderGeometry() method.
// It's an alternative to using textures and rendering images containing custom shapes.
// RenderGeometry() method can also use optional Texture to cover the shape in provided Texture.
// For this example SDL2.dll is required.

const TwoPi = math.Pi * 2

var COLOR_WHITE = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var COLOR_RED = sdl.Color{R: 255, G: 0, B: 0, A: 255}
var COLOR_GREEN = sdl.Color{R: 0, G: 255, B: 0, A: 255}
var COLOR_BLUE = sdl.Color{R: 0, G: 0, B: 255, A: 255}
var BlinkState = false

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal("SDL2 cannot be initialized. ", err)
	}
	defer sdl.Quit()

	app := NewApp("Window", sdl.Point{X: 640, Y: 360})
	defer app.Cleanup()
	app.BackgroundColor = sdl.Color{R: 40, G: 40, B: 40, A: 255}

	// Start blink loop
	go UpdateBlinkState()

	// Create arrow geometry
	// Numbers in drawing below represent vertices
	// 3 triangles needs to be rendered to create arrow shape:
	// 0-2-1, 4-3-5, 4-5-6
	// It will require 9 numbers in indices array and 7 vertices
	//       1
	//     / |
	//   /   3-----4
	// 0     |     |
	//   \   5-----6
	//     \ |
	//       2
	arrowPos := sdl.FPoint{X: 400, Y: 100}
	arrow := Geometry{Vertices: []sdl.Vertex{
		{Position: sdl.FPoint{X: arrowPos.X, Y: arrowPos.Y + 20}, Color: COLOR_WHITE},
		{Position: sdl.FPoint{X: arrowPos.X + 20, Y: arrowPos.Y}, Color: COLOR_WHITE},
		{Position: sdl.FPoint{X: arrowPos.X + 20, Y: arrowPos.Y + 40}, Color: COLOR_WHITE},
		{Position: sdl.FPoint{X: arrowPos.X + 20, Y: arrowPos.Y + 12}, Color: COLOR_WHITE},
		{Position: sdl.FPoint{X: arrowPos.X + 45, Y: arrowPos.Y + 12}, Color: COLOR_WHITE},
		{Position: sdl.FPoint{X: arrowPos.X + 20, Y: arrowPos.Y + 28}, Color: COLOR_WHITE},
		{Position: sdl.FPoint{X: arrowPos.X + 45, Y: arrowPos.Y + 28}, Color: COLOR_WHITE},
	}, Indices: []int32{
		2, 1, 0,
		4, 3, 5,
		4, 5, 6,
	}}

	// Create circle geometry
	circlePos := sdl.FPoint{X: 500, Y: 220}
	circleRadius := 40.0
	step := math.Pi / 10 // Amount of triangles per half circle, experiment with it and see what happens to the circle
	circle := Geometry{Vertices: make([]sdl.Vertex, 0), Indices: make([]int32, 0)}
	circle.Vertices = append(circle.Vertices, sdl.Vertex{Position: circlePos, Color: COLOR_WHITE}) // Add center point
	// Create points around the edge
	for angle := 0.0; angle <= TwoPi; angle += step {
		// Calculate x and y offset from the circle center
		x := float32(math.Cos(angle)) * float32(circleRadius)
		y := float32(math.Sin(angle)) * float32(circleRadius)
		// Create new vertex and append it
		circle.Vertices = append(circle.Vertices, sdl.Vertex{Position: sdl.FPoint{X: circlePos.X + x, Y: circlePos.Y + y}, Color: COLOR_WHITE})
		if angle > 0 {
			// Skip first point, add indices to create new triangle (connecting the vertices to form a triangle)
			lastIndex := int32(len(circle.Vertices))
			circle.Indices = append(circle.Indices, lastIndex-2) // Previous point
			circle.Indices = append(circle.Indices, 0)           // Center point
			circle.Indices = append(circle.Indices, lastIndex-1) // Current point
			if angle+step > TwoPi {
				// Last point was added - connect last tirangle
				circle.Indices = append(circle.Indices, lastIndex-1) // Current point
				circle.Indices = append(circle.Indices, 0)           // Center point
				circle.Indices = append(circle.Indices, 1)           // First point
			}
		}
	}

	previousBlinkState := BlinkState
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

		// Check BlinkState
		if BlinkState != previousBlinkState {
			circle.SetColor(Tif(BlinkState, COLOR_WHITE, COLOR_GREEN))
		}
		previousBlinkState = BlinkState

		// Drawing
		app.R.SetDrawColor(app.BackgroundColor.R, app.BackgroundColor.G, app.BackgroundColor.B, app.BackgroundColor.A)
		app.R.Clear()

		// Simple triangle, simplest way to use RenderGeometry()
		// verticies and indices arrays should be declared outside of the main loop to reuse them
		verticies := []sdl.Vertex{
			{Position: sdl.FPoint{X: 10, Y: 160}, Color: COLOR_RED},
			{Position: sdl.FPoint{X: 410, Y: 310}, Color: COLOR_GREEN},
			{Position: sdl.FPoint{X: 210, Y: 10}, Color: COLOR_BLUE},
		}
		indices := []int32{0, 1, 2}
		app.R.RenderGeometry(nil, verticies, indices)

		// Arrow, The arrow is declared outside of the main loop
		app.R.RenderGeometry(nil, arrow.Vertices, arrow.Indices)

		// Circle, The circle is declared outside of the main loop
		app.R.RenderGeometry(nil, circle.Vertices, circle.Indices)

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

// Updates BlinkState every second (changing it's value)
func UpdateBlinkState() {
	for {
		BlinkState = !BlinkState
		time.Sleep(time.Second)
	}
}

// Simple geometry struct that encapsulates vertices and indices arrays
type Geometry struct {
	Vertices []sdl.Vertex
	Indices  []int32
}

// Sets color for entire geometry
func (g *Geometry) SetColor(color sdl.Color) {
	for i := range g.Vertices {
		g.Vertices[i].Color = color
	}
}
