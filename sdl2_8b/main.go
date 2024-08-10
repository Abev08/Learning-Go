package main

import (
	"fmt"
	"log"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

// Geometry rendering. Using Textures.
// Using Texture on arrow geometry created in sdl_8a example.
// Also created new simple rectangle geometry and used color tint on texture used in geometry rendering.
// For this example SDL2.dll and SDL2_image.dll are required.

var COLOR_WHITE = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var COLOR_RED = sdl.Color{R: 255, G: 0, B: 0, A: 255}
var COLOR_GREEN = sdl.Color{R: 0, G: 255, B: 0, A: 255}
var COLOR_BLUE = sdl.Color{R: 0, G: 0, B: 255, A: 255}

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal("SDL2 cannot be initialized. ", err)
	}
	defer sdl.Quit()

	app := NewApp("Window", sdl.Point{X: 640, Y: 360})
	defer app.Cleanup()
	app.BackgroundColor = sdl.Color{R: 40, G: 40, B: 40, A: 255}

	// Load image and create texture from it
	atlasImage, err := img.Load("assets/atlas.png")
	if err != nil {
		log.Fatal("couldn't load image. ", err)
	}
	defer atlasImage.Free() // Defer freeing the image
	atlasTexture, err := app.R.CreateTextureFromSurface(atlasImage)
	if err != nil {
		log.Fatal("couldn't create texture from loaded image (surface). ", err)
	}
	defer atlasTexture.Destroy() // Defer destroying the texture

	// Create arrow
	arrow1 := NewArrow(50, 50, 128, 96,
		sdl.FRect{
			X: 176.0 / float32(atlasImage.W), Y: 336.0 / float32(atlasImage.H), // (176,336) is position of top left pixel of the texture in the atlas
			W: 16.0 / float32(atlasImage.W), H: 16.0 / float32(atlasImage.H)}, // (16,16) is  size of the texture in the atlas
		COLOR_WHITE)
	// Create arrow but override texture color
	arrow2 := NewArrow(330, 160, -128, 96, // Negative width/height allows geometry mirroring in y/x axis
		sdl.FRect{
			X: 0.0, Y: 0.0, // (0,0) is position of top left pixel of the texture in the atlas
			W: 16.0 / float32(atlasImage.W), H: 16.0 / float32(atlasImage.H)}, // (16,16) is  size of the texture in the atlas
		COLOR_GREEN) // By "default" the texture in top left corner is brown, but when rendering it will be tinted green
	// Create rectangle
	rect := NewRectangle(300, 20, 64, 64,
		sdl.FRect{
			X: 128.0 / float32(atlasImage.W), Y: 80.0 / float32(atlasImage.H), // (128,80) is position of top left pixel of the texture in the atlas
			W: 16.0 / float32(atlasImage.W), H: 16.0 / float32(atlasImage.H)}, // (16,16) is  size of the texture in the atlas
		COLOR_WHITE)

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

		// Render geometry structs
		app.R.RenderGeometry(atlasTexture, arrow1.Vertices, arrow1.Indices)
		app.R.RenderGeometry(atlasTexture, arrow2.Vertices, arrow2.Indices)
		app.R.RenderGeometry(atlasTexture, rect.Vertices, rect.Indices)

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
		log.Fatal("error creating a window. ", err)
	}
	app.W.SetResizable(true) // Allow window resizing

	app.WindowID, err = app.W.GetID() // Get window ID
	if err != nil {
		log.Fatal("couldn't receive window ID. ", err)
	}

	app.R, err = sdl.CreateRenderer(app.W, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC) // Use VSync
	if err != nil {
		log.Fatal("error creating a renderer. ", err)
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

// Create new arrow geometry.
// uv is normalized texture coordinates and size
func NewArrow(x, y, w, h float32, uv sdl.FRect, color sdl.Color) Geometry {
	//       1
	//     / |
	//   /   3-----4
	// 0     |     |
	//   \   5-----6
	//     \ |
	//       2
	return Geometry{Vertices: []sdl.Vertex{
		{Position: sdl.FPoint{X: x, Y: y + h*0.5}, Color: color, TexCoord: sdl.FPoint{X: uv.X, Y: uv.Y + uv.H*0.5}},
		{Position: sdl.FPoint{X: x + w*0.5, Y: y}, Color: color, TexCoord: sdl.FPoint{X: uv.X + uv.W*0.5, Y: uv.Y}},
		{Position: sdl.FPoint{X: x + w*0.5, Y: y + h}, Color: color, TexCoord: sdl.FPoint{X: uv.X + uv.W*0.5, Y: uv.Y + uv.H}},
		{Position: sdl.FPoint{X: x + w*0.5, Y: y + h*0.3}, Color: color, TexCoord: sdl.FPoint{X: uv.X + uv.W*0.5, Y: uv.Y + uv.H*0.3}},
		{Position: sdl.FPoint{X: x + w, Y: y + h*0.3}, Color: color, TexCoord: sdl.FPoint{X: uv.X + uv.W, Y: uv.Y + uv.H*0.3}},
		{Position: sdl.FPoint{X: x + w*0.5, Y: y + h*0.7}, Color: color, TexCoord: sdl.FPoint{X: uv.X + uv.W*0.5, Y: uv.Y + uv.H*0.7}},
		{Position: sdl.FPoint{X: x + w, Y: y + h*0.7}, Color: color, TexCoord: sdl.FPoint{X: uv.X + uv.W, Y: uv.Y + uv.H*0.7}},
	}, Indices: []int32{
		2, 1, 0,
		4, 3, 5,
		4, 5, 6,
	}}
}

// Create new rectangle geometry.
// uv is normalized texture coordinates and size
func NewRectangle(x, y, w, h float32, uv sdl.FRect, color sdl.Color) Geometry {
	// 0----------1
	// |          |
	// |          |
	// 2----------3
	return Geometry{Vertices: []sdl.Vertex{
		{Position: sdl.FPoint{X: x, Y: y}, Color: color, TexCoord: sdl.FPoint{X: uv.X, Y: uv.Y}},
		{Position: sdl.FPoint{X: x + w, Y: y}, Color: color, TexCoord: sdl.FPoint{X: uv.X + uv.W, Y: uv.Y}},
		{Position: sdl.FPoint{X: x, Y: y + h}, Color: color, TexCoord: sdl.FPoint{X: uv.X, Y: uv.Y + uv.H}},
		{Position: sdl.FPoint{X: x + w, Y: y + h}, Color: color, TexCoord: sdl.FPoint{X: uv.X + uv.W, Y: uv.Y + uv.H}},
	}, Indices: []int32{
		1, 0, 2,
		1, 2, 3,
	}}
}
