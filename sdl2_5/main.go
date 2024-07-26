package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

// Textures, animations and character movement.
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

	app.BackgroundColor = sdl.Color{R: 40, G: 40, B: 40, A: 255}

	// Load image and create texture from it
	slimeImg, err := img.Load("assets/slime.png")
	if err != nil {
		log.Fatal("couldn't load image", err)
	}
	slimeTexture, err := app.R.CreateTextureFromSurface(slimeImg)
	if err != nil {
		log.Fatal("couldn't create texture from loaded image (surface)", err)
	}
	defer slimeTexture.Destroy()

	animIdle := Animation{
		StartPosition: sdl.Point{X: 0, Y: 0},
		FrameSize:     sdl.Point{X: 32, Y: 32},
		KeyFrames:     4,
		FrameDuration: 140,
	}
	animMoving := Animation{
		StartPosition: sdl.Point{X: 0, Y: 32},
		FrameSize:     sdl.Point{X: 32, Y: 32},
		KeyFrames:     6,
		FrameDuration: 140,
	}
	animDying := Animation{
		StartPosition: sdl.Point{X: 0, Y: 64},
		FrameSize:     sdl.Point{X: 32, Y: 32},
		KeyFrames:     5,
		FrameDuration: 100,
	}
	slime := Character{
		Position:  sdl.FPoint{X: 10, Y: 30},
		Animation: animIdle.Start(),
		MaxSpeed:  300,
	}

	running := true
	inputs := InputState{MousePosition: sdl.Point{}}
	lastFrameTime := time.Now()
	var dt float32
	for running {
		// Calculate time between frames
		dt = float32(time.Since(lastFrameTime).Seconds())
		lastFrameTime = time.Now()

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

			case *sdl.KeyboardEvent:
				// Slime movement (getting move directions)
				if e.Repeat == 0 {
					if e.Keysym.Sym == sdl.K_a {
						if e.State == sdl.PRESSED {
							slime.MovementDirection |= 1
						} else {
							mask := ^(1)
							slime.MovementDirection &= uint8(mask)
						}
					}
					if e.Keysym.Sym == sdl.K_d {
						if e.State == sdl.PRESSED {
							slime.MovementDirection |= 2
						} else {
							mask := ^(2)
							slime.MovementDirection &= uint8(mask)
						}
					}
					if e.Keysym.Sym == sdl.K_w {
						if e.State == sdl.PRESSED {
							slime.MovementDirection |= 4
						} else {
							mask := ^(4)
							slime.MovementDirection &= uint8(mask)
						}
					}
					if e.Keysym.Sym == sdl.K_s {
						if e.State == sdl.PRESSED {
							slime.MovementDirection |= 8
						} else {
							mask := ^(8)
							slime.MovementDirection &= uint8(mask)
						}
					}
				}
			}
		}

		// Drawing
		app.R.SetDrawColor(app.BackgroundColor.R, app.BackgroundColor.G, app.BackgroundColor.B, app.BackgroundColor.A)
		app.R.Clear()

		// Update animation
		animFinished := slime.Animation.Update()
		if DrawButton(app.R, app.WindowID, &inputs, &sdl.Rect{X: 5, Y: 5, W: 60, H: 20}) {
			// Start death animation
			slime.Animation = animDying.Start()
		} else if animFinished && slime.Animation == &animDying {
			// Death animation finished, start idle animation
			slime.Position.X, slime.Position.Y = 10, 30
			slime.Animation = animIdle.Start()
		} else if ((slime.MovementDirection&1 > 0) != (slime.MovementDirection&2 > 0)) ||
			((slime.MovementDirection&4 > 0) != (slime.MovementDirection&8 > 0)) {
			// Character is moving, start moving animation
			if slime.Animation == &animIdle || animFinished {
				slime.Animation = animMoving.Start()
			}
		} else if animFinished {
			// Other animations finished, start idle animation
			slime.Animation = animIdle.Start()
		}

		// Update slime position (slime movement)
		vel := sdl.FPoint{X: 0, Y: 0}
		if (slime.MovementDirection&1 > 0) != (slime.MovementDirection&2 > 0) {
			// Left and right movement
			vel.X += Tif(slime.MovementDirection&1 > 0, float32(-1), 1)
			slime.IsMovingLeft = vel.X < 0
		}
		if (slime.MovementDirection&4 > 0) != (slime.MovementDirection&8 > 0) {
			// Up and down movement
			vel.Y += Tif(slime.MovementDirection&4 > 0, float32(-1), 1)
		}
		if vel.X != 0 && vel.Y != 0 {
			NormalizeFPoint(&vel) // Normalize movement vector, without it diagonal speeds exceed max speed
		}
		if slime.Animation == &animMoving && slime.Animation.Frame >= 4 && slime.Animation.Frame <= 6 {
			// Move only when slime is in the air
			slime.Position.X += vel.X * slime.MaxSpeed * dt
			slime.Position.Y += vel.Y * slime.MaxSpeed * dt
		}

		// Draw slime
		app.R.CopyEx(slimeTexture,
			&sdl.Rect{
				X: slime.Animation.StartPosition.X + int32(slime.Animation.Frame)*slime.Animation.FrameSize.X,
				Y: slime.Animation.StartPosition.Y,
				W: slime.Animation.FrameSize.X,
				H: slime.Animation.FrameSize.Y},
			&sdl.Rect{
				X: int32(slime.Position.X), Y: int32(slime.Position.Y),
				W: slime.Animation.FrameSize.X * 4, H: slime.Animation.FrameSize.X * 4},
			0,
			&sdl.Point{X: slime.Animation.FrameSize.X / 2, Y: slime.Animation.FrameSize.Y / 2},
			Tif(slime.IsMovingLeft, sdl.FLIP_HORIZONTAL, sdl.FLIP_NONE))

		app.R.Present()
	}
}

// App struct that groups data related to sdl.Window
type App struct {
	W                *sdl.Window   // Window
	R                *sdl.Renderer // Renderer
	WindowID         uint32
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

type Character struct {
	Position          sdl.FPoint // Position
	Animation         *Animation // Current animation
	MaxSpeed          float32    // Maximum movement speed
	MovementDirection uint8      // Movement Direction
	IsMovingLeft      bool       // Is the slime moving left?
}

type Animation struct {
	StartPosition  sdl.Point // Pixel position in atlas texture
	FrameSize      sdl.Point // Size of single frame
	KeyFrames      uint16    // Number of frames
	Frame          uint16    // Current frame
	FrameDuration  uint16    // Duration of single frame in milliseconds
	FrameTimeStart int64     // Time in milliseconds when frame started
}

// Starts the animation, also returns pointer to self
func (a *Animation) Start() *Animation {
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
	return finished
}

// Ternary if statement
func Tif[T any](condition bool, vTrue, vFalse T) T {
	if condition {
		return vTrue
	}
	return vFalse
}

// Normalize length of sdl.FPoint
func NormalizeFPoint(p *sdl.FPoint) {
	l := p.X*p.X + p.Y*p.Y
	if l == 0 || l == 1 {
		return
	}
	var s float32 = 1.0 / float32(math.Sqrt(float64(l)))
	p.X *= s
	p.Y *= s
}
