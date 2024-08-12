package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// Snake. Simple game.
// A/D or Left/Right arrows change movement direction.
// R reset.
// For this example SDL2.dll is required.

var COLOR_WHITE = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var COLOR_RED = sdl.Color{R: 255, G: 0, B: 0, A: 255}
var COLOR_GREEN = sdl.Color{R: 0, G: 255, B: 0, A: 255}
var COLOR_BLUE = sdl.Color{R: 0, G: 0, B: 255, A: 255}
var COLOR_ORANGE = sdl.Color{R: 255, G: 165, B: 0, A: 255}

const GRID_SIZE int32 = 20            // Size of grid cell
const SNAKE_BODY_SIZE = GRID_SIZE - 2 // Size of snake body part

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal("SDL2 cannot be initialized. ", err)
	}
	defer sdl.Quit()

	app := NewApp("Window", sdl.Point{X: 600, Y: 400}) // Window size should be multiples of GRID_SIZE
	defer app.Cleanup()
	app.BackgroundColor = sdl.Color{R: 40, G: 40, B: 40, A: 255}

	startPos := sdl.Point{X: GRID_SIZE * 8, Y: GRID_SIZE * 5} // Snake starting position
	snake := NewSnake(startPos.X, startPos.Y)                 // Create new snake
	var apple *sdl.Rect = nil                                 // Pointer to apple rect

	running := true
	inputs := InputState{MousePosition: sdl.Point{}}
	moveTimer := time.Now()
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

			case *sdl.KeyboardEvent:
				if e.Repeat == 0 && e.State == sdl.RELEASED {
					if e.Keysym.Sym == sdl.K_a || e.Keysym.Sym == sdl.K_LEFT {
						// Turn left
						snake.Turn = 'L'
					} else if e.Keysym.Sym == sdl.K_d || e.Keysym.Sym == sdl.K_RIGHT {
						// Turn right
						snake.Turn = 'R'
					} else if e.Keysym.Sym == sdl.K_r {
						// Reset
						snake = NewSnake(startPos.X, startPos.Y)
						apple = nil
						moveTimer = time.Now()
					}
				}
			}
		}

		// If apple doesn't exist, create new one
		if apple == nil {
			wrongLocationCounter := 0 // Counter how many wrong locations are generated - to prevent endless loop
			var wrongLocation bool
			for {
				if wrongLocationCounter >= 1000 {
					// Couldn't create new apple, win?
					snake.IsAlive = false
					fmt.Println("WIN! Score:", snake.Score)
					break
				}
				wrongLocation = false
				// Create new random location for the apple
				x := (rand.Int31n(app.WindowSize.X/GRID_SIZE) * GRID_SIZE) + 1
				y := (rand.Int31n(app.WindowSize.Y/GRID_SIZE) * GRID_SIZE) + 1
				// Check if any of the body parts of the snake are present in created location
				for _, b := range snake.Body {
					if b.X == x && b.Y == y {
						wrongLocation = true
						wrongLocationCounter++
						break
					}
				}
				if wrongLocation {
					continue // Wrong apple location found, try different one
				}
				// Created good location, create an apple there
				apple = &sdl.Rect{X: x, Y: y, W: SNAKE_BODY_SIZE, H: SNAKE_BODY_SIZE}
				break
			}
		}

		// Update the snake (move, collide, eat apples, etc.)
		if snake.IsAlive && time.Since(moveTimer) >= snake.MoveCooldown {
			moveTimer = time.Now()
			var prevPos *sdl.Point = nil // Pointer to position of previous body part
			for i := range snake.Body {
				b := &snake.Body[i]
				p := &snake.PreviousBodyPosition[i]
				p.X, p.Y = b.X, b.Y // Update previous body part position
				if prevPos == nil {
					// Head part move it towards moving direction
					// Turning the head
					if snake.Turn == 'L' {
						switch snake.Direction {
						case 'N':
							snake.Direction = 'E'
						case 'S':
							snake.Direction = 'W'
						case 'W':
							snake.Direction = 'N'
						case 'E':
							snake.Direction = 'S'
						}
					} else if snake.Turn == 'R' {
						switch snake.Direction {
						case 'N':
							snake.Direction = 'W'
						case 'S':
							snake.Direction = 'E'
						case 'W':
							snake.Direction = 'S'
						case 'E':
							snake.Direction = 'N'
						}
					}
					snake.Turn = ' ' // Reset turn state

					// Move the head
					switch snake.Direction {
					case 'N':
						b.Y += GRID_SIZE
					case 'S':
						b.Y -= GRID_SIZE
					case 'W':
						b.X -= GRID_SIZE
					case 'E':
						b.X += GRID_SIZE
					}

					// Cap the position to window size
					if b.X < 0 {
						b.X = app.WindowSize.X + b.X
					} else if b.X > app.WindowSize.X {
						b.X -= app.WindowSize.X
					} else if b.Y < 0 {
						b.Y = app.WindowSize.Y + b.Y
					} else if b.Y > app.WindowSize.Y {
						b.Y -= app.WindowSize.Y
					}

					// Check if head collides with an apple
					if b.X == apple.X && b.Y == apple.Y {
						// Consume the apple
						apple = nil
						snake.Score++
						snake.CreateNewBodyPart()
						if snake.MoveCooldown > time.Millisecond*150 {
							snake.MoveCooldown -= time.Millisecond * 10
						}
					}

					// Check if head collides with other body parts
					for i := 1; i < len(snake.Body); i++ {
						p := snake.Body[i]
						if b.X == p.X && b.Y == p.Y {
							// Position of the head is the same as one of the body parts
							// The snake collided with itself
							snake.IsAlive = false
							fmt.Println("LOST! Score:", snake.Score)
							break
						}
					}
				} else {
					// Other body part - move it to position of previous body part
					b.X, b.Y = prevPos.X, prevPos.Y
				}
				prevPos = p // Remember position of previous body part
			}
		}

		// Drawing
		app.R.SetDrawColor(app.BackgroundColor.R, app.BackgroundColor.G, app.BackgroundColor.B, app.BackgroundColor.A)
		app.R.Clear()

		// Draw the snake
		for i, b := range snake.Body {
			if !snake.IsAlive {
				app.R.SetDrawColor(COLOR_BLUE.R, COLOR_BLUE.G, COLOR_BLUE.B, COLOR_BLUE.A)
			} else if i == 0 {
				app.R.SetDrawColor(COLOR_ORANGE.R, COLOR_ORANGE.G, COLOR_ORANGE.B, COLOR_ORANGE.A)
				app.R.FillRect(&b)
				app.R.SetDrawColor(COLOR_GREEN.R, COLOR_GREEN.G, COLOR_GREEN.B, COLOR_GREEN.A)
				continue
			}
			app.R.FillRect(&b)
		}

		// Draw the apple
		if apple != nil {
			app.R.SetDrawColor(COLOR_RED.R, COLOR_RED.G, COLOR_RED.B, COLOR_RED.A)
			app.R.FillRect(apple)
		}

		app.R.Present()
	}
}

// App struct that groups data related to sdl.Window
type App struct {
	W               *sdl.Window   // Window
	R               *sdl.Renderer // Renderer
	WindowSize      sdl.Point     // Size of the window
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
	app.WindowSize = size

	app.W, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, app.WindowSize.X, app.WindowSize.Y, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatal("error creating a window. ", err)
	}
	// app.W.SetResizable(true) // Allow window resizing

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

type Snake struct {
	Body                 []sdl.Rect    // Parts of snake body
	PreviousBodyPosition []sdl.Point   // Previous body positions
	Direction            rune          // Movement direction (North, South, West, East)
	Turn                 rune          // Turn Left/Right
	MoveCooldown         time.Duration // Time between move actions
	IsAlive              bool          // Is the snake alive?
	Score                uint32        // Amount of eaten apples
}

// Creates new Snake struct
func NewSnake(x, y int32) Snake {
	padding := (GRID_SIZE - SNAKE_BODY_SIZE) / 2
	s := Snake{
		Body: []sdl.Rect{
			{X: x + padding, Y: y + padding, W: SNAKE_BODY_SIZE, H: SNAKE_BODY_SIZE},
			{X: x + padding - GRID_SIZE, Y: y + padding, W: SNAKE_BODY_SIZE, H: SNAKE_BODY_SIZE},
			{X: x + padding - GRID_SIZE*2, Y: y + padding, W: SNAKE_BODY_SIZE, H: SNAKE_BODY_SIZE},
		},
		PreviousBodyPosition: make([]sdl.Point, 3),
		Direction:            'E',
		MoveCooldown:         time.Millisecond * 500,
		IsAlive:              true,
		Score:                0,
	}
	s.PreviousBodyPosition[0] = sdl.Point{X: s.Body[0].X, Y: s.Body[0].Y}
	s.PreviousBodyPosition[1] = sdl.Point{X: s.Body[1].X, Y: s.Body[1].Y}
	s.PreviousBodyPosition[2] = sdl.Point{X: s.Body[2].X, Y: s.Body[2].Y}
	return s
}

// Creates new body part at the end of the snake
func (s *Snake) CreateNewBodyPart() {
	lastIndex := len(s.Body) - 1
	prevBody := s.Body[lastIndex]
	s.Body = append(s.Body, sdl.Rect{X: prevBody.X, Y: prevBody.Y, W: SNAKE_BODY_SIZE, H: SNAKE_BODY_SIZE})
	s.PreviousBodyPosition = append(s.PreviousBodyPosition, sdl.Point{X: prevBody.X, Y: prevBody.Y})
}
