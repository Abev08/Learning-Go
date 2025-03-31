package main

import (
	"fmt"
	"log"
	"slices"

	"github.com/jupiterrider/purego-sdl3/sdl"
)

// Splines:
// - Linear,
// - from Quadratic Bézier curves,
// - from Cubic Bézier curves,
// - Catmull-Rom.

const (
	POINT_RECT_SIZE      float32 = 10 // Rectangle width and height for marking anchor points
	POINT_RECT_SIZE_HALF         = POINT_RECT_SIZE / 2
	MAX_STEP_SIZE        float32 = 15  // Maximum step size when creating curve segments
	MAX_STEPS            float32 = 100 // Maximum number of steps when creating curve segments
)

var window = Window{
	BackgroundColor: sdl.Color{R: 40, G: 40, B: 40, A: 255},
	Points:          make([]Anchor, 0, 10),
	PointsVisible:   true, HandlesVisible: true, LinesVisible: true,
}
var inputsState = InputState{KeyState: make(map[sdl.Keycode]bool)}
var BezierQuadratic []sdl.FPoint
var BezierQuadraticStrings []sdl.FPoint
var BezierCubic []sdl.FPoint
var BezierCubicStrings []sdl.FPoint
var CatmullRom []sdl.FPoint
var running = true

func main() {
	// Get SDL version
	version := make([]int32, 3)
	version[0], version[1], version[2] = sdl.GetVersion()
	log.Printf("SDL %d.%d.%d\n", version[0], version[1], version[2])

	// Initialize SDL3
	if !sdl.Init(sdl.InitVideo) {
		log.Fatalln(sdl.GetError())
	}
	defer sdl.Quit()

	// Create window and renderer
	defer Cleanup() // Defer cleaning up the window
	if !sdl.CreateWindowAndRenderer("Splines", 1280, 720, sdl.WindowResizable, &window.W, &window.R) {
		log.Fatalln(sdl.GetError())
	}

	// FIXME SDL3 Go bindings are missing sdl.GetWindowID(), for now assume WindowID = 3
	// window.ID = sdl.GetWindowID(w.W)
	window.ID = 3

	// Enable VSync. This can be used any time after creating renderer, not like sdl.SetHint(sdl.HintRenderVsync, "1")
	sdl.SetRenderVSync(window.R, 1) // 0 - off, 1 - vsync, 2 - vsync/2, 3 - vsync/3, 4 - vsync/4

	// Main loop
	for running {
		// Process events
		inputsState.NewFrame()
		inputsState.ProcessEvents()

		window.Update()
		window.Draw()
	}
}

// Cleans up after the app (destroys windows and renderers, etc.)
func Cleanup() {
	if window.R != nil {
		sdl.DestroyRenderer(window.R)
	}
	if window.W != nil {
		sdl.DestroyWindow(window.W)
	}
}

// A struct to encapsulate window related stuff
type Window struct {
	W *sdl.Window   // The window
	R *sdl.Renderer // Renderer to the window

	ID              sdl.WindowID // Window ID
	BackgroundColor sdl.Color

	Points                                 []Anchor
	HoldingAnchorID, HoldingAnchorHandleID int

	PointsVisible, HandlesVisible bool
	LinesVisible                  bool
	BezierQuadraticVisible        bool
	BezierCubicVisible            bool
	BezierStringsVisible          bool
	CatmullRomVisible             bool
}

func (w *Window) Update() {
	if inputsState.MouseOnWindowID != w.ID {
		return
	}

	if inputsState.MouseLeftClicked && inputsState.KeyState[sdl.KeycodeLShift] {
		// Add new point at mouse position
		w.Points = append(w.Points, NewAnchor(inputsState.MousePosition))
		UpdateBezierQuadratic()
		UpdateBezierCubic()
		UpdateCatmullRom()
	} else if w.HoldingAnchorID >= 0 {
		if !inputsState.MouseLeftPressed {
			w.HoldingAnchorID = -1
		} else {
			w.Points[w.HoldingAnchorID].SetPosition(inputsState.MousePosition)
			UpdateBezierQuadratic()
			UpdateBezierCubic()
			UpdateCatmullRom()
		}
	} else if w.HoldingAnchorHandleID >= 0 {
		if !inputsState.MouseRightPressed {
			w.HoldingAnchorHandleID = -1
		} else {
			w.Points[w.HoldingAnchorHandleID].SetHandlePosition(inputsState.MousePosition)
			UpdateBezierQuadratic()
			UpdateBezierCubic()
			UpdateCatmullRom()
		}
	} else if inputsState.MouseLeftJustPressed && w.HoldingAnchorID == -1 {
		for i := range w.Points {
			p := &w.Points[i]
			if sdl.PointInRectFloat(inputsState.MousePosition, p.Rect) {
				w.HoldingAnchorID = i
				break
			}
		}
	} else if inputsState.MouseRightJustPressed && w.HoldingAnchorHandleID == -1 && w.HandlesVisible && !w.LinesVisible && !w.CatmullRomVisible {
		for i := range w.Points {
			p := &w.Points[i]
			if sdl.PointInRectFloat(inputsState.MousePosition, p.HandleRect) {
				if inputsState.KeyState[sdl.KeycodeLShift] {
					// Reset handle position
					p.SetHandlePosition(sdl.FPoint{X: p.Center.X, Y: p.Center.Y})
				} else {
					w.HoldingAnchorHandleID = i
				}
				break
			}
		}
	} else if inputsState.MouseRightClicked && inputsState.KeyState[sdl.KeycodeLShift] {
		for i := range w.Points {
			p := &w.Points[i]
			if sdl.PointInRectFloat(inputsState.MousePosition, p.Rect) {
				w.Points = slices.Delete(w.Points, i, i+1)
				UpdateBezierQuadratic()
				UpdateBezierCubic()
				UpdateCatmullRom()
				break
			}
		}
	}
}

func (w *Window) Draw() {
	sdl.SetRenderDrawColor(w.R, w.BackgroundColor.R, w.BackgroundColor.G, w.BackgroundColor.B, w.BackgroundColor.A)
	sdl.RenderClear(w.R) // Clear the screen

	// Helper text
	sdl.SetRenderDrawColor(w.R, 255, 255, 255, 255)
	sdl.RenderDebugText(w.R, 10, 10, "LShift + LClick for new point")
	sdl.RenderDebugText(w.R, 10, 20, "LShift + RClick to delete a point")
	sdl.RenderDebugText(w.R, 10, 30, "LClick to move a point")
	sdl.RenderDebugText(w.R, 10, 40, "RClick to move point handle")

	// Few checkboxes to control what is displayed
	{
		cbRect := sdl.FRect{X: 300, Y: 10, W: 16, H: 16}
		w.PointsVisible = DrawCheckBox(w, cbRect, w.PointsVisible, "Points")
		cbRect.Y += cbRect.H + 5
		w.HandlesVisible = DrawCheckBox(w, cbRect, w.HandlesVisible, "Handles")
		cbRect.X, cbRect.Y = 400, 10
		if DrawCheckBox(w, cbRect, w.LinesVisible, "Lines") != w.LinesVisible {
			w.LinesVisible = true
			w.BezierQuadraticVisible = false
			w.BezierCubicVisible = false
			w.CatmullRomVisible = false
		}
		cbRect.X, cbRect.Y = 500, 10
		if DrawCheckBox(w, cbRect, w.BezierQuadraticVisible, "Bezier Quadratic") != w.BezierQuadraticVisible {
			w.LinesVisible = false
			w.BezierQuadraticVisible = true
			w.BezierCubicVisible = false
			w.CatmullRomVisible = false
		}
		cbRect.Y += cbRect.H + 5
		w.BezierStringsVisible = DrawCheckBox(w, cbRect, w.BezierStringsVisible, "Bezier strings")
		cbRect.X, cbRect.Y = 700, 10
		if DrawCheckBox(w, cbRect, w.BezierCubicVisible, "Bezier Cubic") != w.BezierCubicVisible {
			w.LinesVisible = false
			w.BezierQuadraticVisible = false
			w.BezierCubicVisible = true
			w.CatmullRomVisible = false
		}
		cbRect.X, cbRect.Y = 900, 10
		if DrawCheckBox(w, cbRect, w.CatmullRomVisible, "Catmull-Rom") != w.CatmullRomVisible {
			w.LinesVisible = false
			w.BezierQuadraticVisible = false
			w.BezierCubicVisible = false
			w.CatmullRomVisible = true
		}
	}

	// Bezier quadratic
	if w.BezierQuadraticVisible {
		if w.BezierStringsVisible {
			sdl.SetRenderDrawColor(w.R, 255, 0, 0, 255)
			for i := 0; i < len(BezierQuadraticStrings)-1; i += 2 {
				p0, p1 := &BezierQuadraticStrings[i], &BezierQuadraticStrings[i+1]
				sdl.RenderLine(w.R, p0.X, p0.Y, p1.X, p1.Y)
			}
		}
		sdl.SetRenderDrawColor(w.R, 255, 255, 0, 255)
		for i := range len(BezierQuadratic) - 1 {
			p0, p1 := &BezierQuadratic[i], &BezierQuadratic[i+1]
			sdl.RenderLine(w.R, p0.X, p0.Y, p1.X, p1.Y)
		}
	}

	// Bezier cubic
	if w.BezierCubicVisible {
		if w.BezierStringsVisible {
			sdl.SetRenderDrawColor(w.R, 255, 0, 0, 255)
			for i := 0; i < len(BezierCubicStrings)-2; i += 3 {
				p0, p1, p2 := &BezierCubicStrings[i], &BezierCubicStrings[i+1], &BezierCubicStrings[i+2]
				sdl.RenderLine(w.R, p0.X, p0.Y, p1.X, p1.Y)
				sdl.RenderLine(w.R, p1.X, p1.Y, p2.X, p2.Y)
			}
		}
		sdl.SetRenderDrawColor(w.R, 255, 255, 0, 255)
		for i := range len(BezierCubic) - 1 {
			p0, p1 := &BezierCubic[i], &BezierCubic[i+1]
			sdl.RenderLine(w.R, p0.X, p0.Y, p1.X, p1.Y)
		}
	}

	// Line
	if w.LinesVisible {
		sdl.SetRenderDrawColor(w.R, 0, 255, 0, 255)
		for i := range len(w.Points) - 1 {
			p0, p1 := &w.Points[i], &w.Points[i+1]
			sdl.RenderLine(w.R, p0.Center.X, p0.Center.Y, p1.Center.X, p1.Center.Y)
		}
	}

	// Catmull-Rom
	if w.CatmullRomVisible {
		sdl.SetRenderDrawColor(w.R, 255, 255, 0, 255)
		// for i := 0; i < len(CatmullRom)-1; i += 2 {
		// 	p0, p1 := &CatmullRom[i], &CatmullRom[i+1]
		// 	sdl.RenderLine(w.R, p0.X, p0.Y, p1.X, p1.Y)
		// }
		for i := range len(CatmullRom) - 1 {
			p0, p1 := &CatmullRom[i], &CatmullRom[i+1]
			sdl.RenderLine(w.R, p0.X, p0.Y, p1.X, p1.Y)
		}
	}

	// Anchor points handles
	if w.HandlesVisible && !w.LinesVisible && !w.CatmullRomVisible {
		sdl.SetRenderDrawColor(w.R, 0, 255, 255, 255)
		for i := range w.Points {
			p := &w.Points[i]
			sdl.RenderRect(w.R, &p.HandleRect)
			if w.HoldingAnchorHandleID >= 0 {
				sdl.RenderLine(w.R, p.HandleCenter.X, p.HandleCenter.Y, p.Center.X, p.Center.Y)
			}
		}
	}
	// Anchor points
	if w.PointsVisible {
		sdl.SetRenderDrawColor(w.R, 255, 255, 255, 255)
		for i := range w.Points {
			p := &w.Points[i]
			sdl.RenderRect(w.R, &p.Rect)
		}
	}

	sdl.RenderPresent(w.R)
}

type Anchor struct {
	Center       sdl.FPoint
	Rect         sdl.FRect
	HandleCenter sdl.FPoint
	HandleRect   sdl.FRect
}

// Creates new anchor point at provided position
func NewAnchor(pos sdl.FPoint) Anchor {
	a := Anchor{}
	a.Rect.W, a.Rect.H = POINT_RECT_SIZE, POINT_RECT_SIZE
	a.HandleRect.W, a.HandleRect.H = POINT_RECT_SIZE, POINT_RECT_SIZE
	a.SetPosition(pos)
	a.SetHandlePosition(pos)
	return a
}

// Sets anchor position to provided one, anchor handle is also moved
func (a *Anchor) SetPosition(pos sdl.FPoint) {
	handleX, handleY := a.HandleCenter.X-a.Center.X, a.HandleCenter.Y-a.Center.Y
	a.Center.X, a.Center.Y = pos.X, pos.Y
	a.Rect.X, a.Rect.Y = pos.X-POINT_RECT_SIZE_HALF, pos.Y-POINT_RECT_SIZE_HALF
	a.SetHandlePosition(sdl.FPoint{X: a.Center.X + handleX, Y: a.Center.Y + handleY})
}

// Sets anchor handle to provided position
func (a *Anchor) SetHandlePosition(pos sdl.FPoint) {
	a.HandleCenter.X, a.HandleCenter.Y = pos.X, pos.Y
	a.HandleRect.X, a.HandleRect.Y = pos.X-POINT_RECT_SIZE_HALF, pos.Y-POINT_RECT_SIZE_HALF
}

// Input devices state
type InputState struct {
	MouseOnWindowID                                             sdl.WindowID
	MousePosition                                               sdl.FPoint
	MouseLeftPressed, MouseLeftJustPressed, MouseLeftClicked    bool
	MouseRightPressed, MouseRightJustPressed, MouseRightClicked bool
	KeyState                                                    map[sdl.Keycode]bool
}

func (is *InputState) NewFrame() {
	is.MouseLeftJustPressed = false
	is.MouseLeftClicked = false
	is.MouseRightJustPressed = false
	is.MouseRightClicked = false
}

func (is *InputState) ProcessEvents() {
	event := sdl.Event{}
	for sdl.PollEvent(&event) {
		switch event.Type() {
		case sdl.EventQuit:
			fallthrough
		case sdl.EventWindowCloseRequested:
			log.Println("Window close requested")
			running = false

		case sdl.EventKeyDown:
			fallthrough
		case sdl.EventKeyUp:
			e := event.Key()
			is.MouseOnWindowID = e.WindowID
			is.KeyState[e.Key] = e.Down

		case sdl.EventMouseMotion:
			e := event.Motion()
			is.MouseOnWindowID = e.WindowID
			is.MousePosition.X, is.MousePosition.Y = e.X, e.Y

		case sdl.EventMouseButtonDown:
			fallthrough
		case sdl.EventMouseButtonUp:
			e := event.Button()
			is.MouseOnWindowID = e.WindowID
			switch e.Button {
			case uint8(sdl.ButtonLeft):
				is.MouseLeftPressed = e.Down
				is.MouseLeftJustPressed = e.Down
				is.MouseLeftClicked = !e.Down
			case uint8(sdl.ButtonRight):
				is.MouseRightPressed = e.Down
				is.MouseRightJustPressed = e.Down
				is.MouseRightClicked = !e.Down
			}
		}
	}
}

func abs(val float32) float32 {
	if val < 0 {
		return -val
	}
	return val
}

func UpdateBezierQuadratic() {
	BezierQuadratic = slices.Delete(BezierQuadratic, 0, len(BezierQuadratic))
	BezierQuadraticStrings = slices.Delete(BezierQuadraticStrings, 0, len(BezierQuadraticStrings))

	for i := range len(window.Points) - 1 {
		p := &window.Points[i]
		pNext := &window.Points[i+1]

		diffHP := sdl.FPoint{X: p.HandleCenter.X - p.Center.X, Y: p.HandleCenter.Y - p.Center.Y}
		diffPNextH := sdl.FPoint{X: pNext.Center.X - p.HandleCenter.X, Y: pNext.Center.Y - p.HandleCenter.Y}

		// Calculate amount of steps required for a smooth curve
		var steps float32 = 5
		var stepHP, stepPNextH sdl.FPoint
		for {
			stepHP.X, stepHP.Y = diffHP.X/steps, diffHP.Y/steps
			stepPNextH.X, stepPNextH.Y = diffPNextH.X/steps, diffPNextH.Y/steps

			if steps >= MAX_STEPS {
				fmt.Println("Reached maximum number of steps!")
				break
			} else if abs(stepHP.X) > MAX_STEP_SIZE || abs(stepHP.Y) > MAX_STEP_SIZE ||
				abs(stepPNextH.X) > MAX_STEP_SIZE || abs(stepPNextH.Y) > MAX_STEP_SIZE {
				steps++
			} else {
				break
			}
		}

		// Create Bezier line points
		for j := range int(steps) {
			s := float32(j)
			m1 := sdl.FPoint{X: p.Center.X + stepHP.X*s, Y: p.Center.Y + stepHP.Y*s}
			m2 := sdl.FPoint{X: p.HandleCenter.X + stepPNextH.X*s, Y: p.HandleCenter.Y + stepPNextH.Y*s}

			if j > 0 {
				BezierQuadraticStrings = append(BezierQuadraticStrings, []sdl.FPoint{m1, m2}...)
			}

			diffMM := sdl.FPoint{X: m2.X - m1.X, Y: m2.Y - m1.Y}
			stepMM := sdl.FPoint{X: diffMM.X / steps, Y: diffMM.Y / steps}
			BezierQuadratic = append(BezierQuadratic, sdl.FPoint{X: m1.X + stepMM.X*s, Y: m1.Y + stepMM.Y*s})
		}
		BezierQuadratic = append(BezierQuadratic, pNext.Center) // Add last point
	}
}

func UpdateBezierCubic() {
	BezierCubic = slices.Delete(BezierCubic, 0, len(BezierCubic))
	BezierCubicStrings = slices.Delete(BezierCubicStrings, 0, len(BezierCubicStrings))

	for i := range len(window.Points) - 1 {
		p := &window.Points[i]
		pNext := &window.Points[i+1]

		diffHP := sdl.FPoint{X: p.HandleCenter.X - p.Center.X, Y: p.HandleCenter.Y - p.Center.Y}
		diffHNextPNext := sdl.FPoint{X: pNext.Center.X - pNext.HandleCenter.X, Y: pNext.Center.Y - pNext.HandleCenter.Y}
		diffHNextH := sdl.FPoint{X: pNext.Center.X + diffHNextPNext.X - p.HandleCenter.X, Y: pNext.Center.Y + diffHNextPNext.Y - p.HandleCenter.Y}

		// Calculate amount of steps required for a smooth curve
		var steps float32 = 5
		var stepHP, stepHNextH, stepHNextPNext sdl.FPoint
		for {
			stepHP.X, stepHP.Y = diffHP.X/steps, diffHP.Y/steps
			stepHNextH.X, stepHNextH.Y = diffHNextH.X/steps, diffHNextH.Y/steps
			stepHNextPNext.X, stepHNextPNext.Y = diffHNextPNext.X/steps, diffHNextPNext.Y/steps

			if steps >= MAX_STEPS {
				fmt.Println("Reached maximum number of steps!")
				break
			} else if abs(stepHP.X) > MAX_STEP_SIZE || abs(stepHP.Y) > MAX_STEP_SIZE ||
				abs(stepHNextH.X) > MAX_STEP_SIZE || abs(stepHNextH.Y) > MAX_STEP_SIZE ||
				abs(stepHNextPNext.X) > MAX_STEP_SIZE || abs(stepHNextPNext.Y) > MAX_STEP_SIZE {
				steps++
			} else {
				break
			}
		}

		// Create Bezier line points
		for j := range int(steps) {
			s := float32(j)
			m1 := sdl.FPoint{X: p.Center.X + stepHP.X*s, Y: p.Center.Y + stepHP.Y*s}
			m2 := sdl.FPoint{X: p.HandleCenter.X + stepHNextH.X*s, Y: p.HandleCenter.Y + stepHNextH.Y*s}
			m3 := sdl.FPoint{X: pNext.Center.X + diffHNextPNext.X - stepHNextPNext.X*s, Y: pNext.Center.Y + diffHNextPNext.Y - stepHNextPNext.Y*s}

			BezierCubicStrings = append(BezierCubicStrings, []sdl.FPoint{m1, m2, m3}...)

			diffM2M1 := sdl.FPoint{X: m2.X - m1.X, Y: m2.Y - m1.Y}
			diffM3M2 := sdl.FPoint{X: m3.X - m2.X, Y: m3.Y - m2.Y}

			stepM2M1 := sdl.FPoint{X: diffM2M1.X / steps, Y: diffM2M1.Y / steps}
			stepM3M2 := sdl.FPoint{X: diffM3M2.X / steps, Y: diffM3M2.Y / steps}

			m21 := sdl.FPoint{X: m1.X + stepM2M1.X*s, Y: m1.Y + stepM2M1.Y*s}
			m32 := sdl.FPoint{X: m2.X + stepM3M2.X*s, Y: m2.Y + stepM3M2.Y*s}

			diffM := sdl.FPoint{X: m32.X - m21.X, Y: m32.Y - m21.Y}
			stepM := sdl.FPoint{X: diffM.X / steps, Y: diffM.Y / steps}
			BezierCubic = append(BezierCubic, sdl.FPoint{X: m21.X + stepM.X*s, Y: m21.Y + stepM.Y*s})
		}
		BezierCubic = append(BezierCubic, pNext.Center) // Add last point
	}
}

func UpdateCatmullRom() {
	CatmullRom = slices.Delete(CatmullRom, 0, len(CatmullRom))
	if len(window.Points) < 2 {
		return
	}

	for i := range len(window.Points) - 1 {
		p := window.Points[i].Center
		pNext := window.Points[i+1].Center
		pHandle, pNextHandle := sdl.FPoint{}, sdl.FPoint{}

		if i > 0 {
			pPrev := window.Points[i-1].Center
			pHandle.X, pHandle.Y = (pNext.X-pPrev.X)/2, (pNext.Y-pPrev.Y)/2
		} else {
			pHandle.X, pHandle.Y = pNext.X-p.X, pNext.Y-p.Y
		}

		if i < len(window.Points)-2 {
			pNextNext := window.Points[i+2].Center
			pNextHandle.X, pNextHandle.Y = -(pNextNext.X-p.X)/2, -(pNextNext.Y-p.Y)/2
		} else {
			pNextHandle.X, pNextHandle.Y = -(pNext.X - p.X), -(pNext.Y - p.Y)
		}

		// Calculate amount of steps required for a smooth curve
		var steps float32 = 5
		var stepHP, stepHNextPNext sdl.FPoint
		for {
			stepHP.X, stepHP.Y = pHandle.X/steps, pHandle.Y/steps
			stepHNextPNext.X, stepHNextPNext.Y = pNextHandle.X/steps, pNextHandle.Y/steps

			if steps >= MAX_STEPS {
				fmt.Println("Reached maximum number of steps!")
				break
			} else if abs(stepHP.X) > MAX_STEP_SIZE || abs(stepHP.Y) > MAX_STEP_SIZE ||
				abs(stepHNextPNext.X) > MAX_STEP_SIZE || abs(stepHNextPNext.Y) > MAX_STEP_SIZE {
				steps++
			} else {
				break
			}
		}

		// Create line points
		sStep := 1 / steps
		addLastPoint := true
		for s := float32(0); s <= 1; s += sStep {
			ss := s * s
			sss := ss * s
			a1 := 2*sss - 3*ss + 1
			a2 := sss - 2*ss + s
			a3 := -2*sss + 3*ss
			a4 := sss - ss

			pS := sdl.FPoint{
				X: a1*p.X + a2*pHandle.X + a3*pNext.X - a4*pNextHandle.X,
				Y: a1*p.Y + a2*pHandle.Y + a3*pNext.Y - a4*pNextHandle.Y,
			}
			CatmullRom = append(CatmullRom, pS)

			if s == 1 {
				addLastPoint = false
			}
		}
		if addLastPoint {
			// Sometimes due to floating point inaccuracy last point is skipped
			CatmullRom = append(CatmullRom, pNext)
		}
	}
}

// Checkbox
func DrawCheckBox(window *Window, rect sdl.FRect, checked bool, text string) bool {
	if inputsState.MouseOnWindowID == window.ID && sdl.PointInRectFloat(inputsState.MousePosition, rect) {
		// Mouse over background
		sdl.SetRenderDrawColor(window.R, 120, 120, 120, 255)
		sdl.RenderFillRect(window.R, &rect)

		if inputsState.MouseLeftClicked {
			checked = !checked
		}
	}

	// Frame
	sdl.SetRenderDrawColor(window.R, 255, 255, 255, 255)
	sdl.RenderRect(window.R, &rect)
	rect.X, rect.Y = rect.X+2, rect.Y+2
	rect.W, rect.H = rect.W-4, rect.H-4
	if checked {
		sdl.RenderFillRect(window.R, &rect)
	}

	// Text
	sdl.RenderDebugText(window.R, rect.X+rect.W+6, rect.Y+3, text)

	return checked
}
