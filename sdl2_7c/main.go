package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Text rendering. More complicated approach.
// Firstly we render each of the characters (glyphs) into texture (glyph atlas).
// Then we use the glyph atlas like any other texture and combine parts of it into text to be rendered.
// In this approach we create glyph atlas only once and no matter what text is needed we can create it from individual glyphs.
// Text that changes often (even in every frame) can be easly rendered without any additional surface / texture creation.
// Rendering text into texture requires sdl.Renderer so we need to create the window first.
// Also simple performance metrics are implemented for comparasion to other text rendering approaches.
// This example is based on "sdl2_5".
// For this example SDL2.dll and SDL2_ttf.dll are required.
//
// Longer text on my pc:
// initialization part took: 4000-6000 us, average 5000 us
// rendering first 10 frames took (excluding first frame) took: 0-1000 us, average below 300 us
// This approach seems to be slower than "sdl2_7b" example but you don't have to manage text textures.
// In my opinion small speed decrease for big convenience increase is justiciable.

var COLOR_WHITE = sdl.Color{R: 255, G: 255, B: 255, A: 255}

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal("SDL2 cannot be initialized. ", err)
	}
	defer sdl.Quit()

	app := NewApp("Window", sdl.Point{X: 640, Y: 360})
	defer app.Cleanup()
	app.BackgroundColor = sdl.Color{R: 40, G: 40, B: 40, A: 255}
	var maxTextWidth int32 = 620 // window width - 10 px padding from left and right

	// Text to be rendered
	// text := "Hello,\nWorld!"
	// Longer text, 70 words
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam commodo nibh risus, quis vulputate arcu mattis at. Etiam massa libero, euismod in metus ut, venenatis efficitur ipsum. Etiam sed consectetur orci. Sed odio odio, aliquam in leo mattis, tincidunt rutrum ex. Proin cursus luctus magna, vel facilisis felis dapibus ac. Vivamus eu dui id ipsum hendrerit sodales vitae id nunc. Fusce vitae lacus tempor, blandit massa ut, consectetur urna. Maecenas."

	// Create and start performace analysis
	var pCounter uint8
	var pDuration time.Duration
	pTime := time.Now()

	// Initialize TTF SDL2 package
	err = ttf.Init()
	if err != nil {
		log.Fatal("SDL2 TTF cannot be initialized. ", err)
	}
	defer ttf.Quit()

	// Load font file
	font24, err := ttf.OpenFont("assets/fonts/OpenSans-SemiBold.ttf", 24)
	if err != nil {
		log.Fatal("couldn't load TTF font file. ", err)
	}
	defer font24.Close()

	// Create new font atlas
	fontAtlas24 := NewFontAtlas(font24, app.R)
	defer fontAtlas24.Atlas.Destroy()

	// Print TTF initialization and font loading time
	pDuration = time.Since(pTime)
	fmt.Printf("TTF initialization, loading font file and creating font atlas took: %d us\n", pDuration.Microseconds())

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

		// Text rendering
		pTime = time.Now() // Get start time

		// Render created text texture
		fontAtlas24.DrawTextWrapped(app.R, text, 10, 10, COLOR_WHITE, maxTextWidth)

		// Print how long it took to render one frame
		pDuration = time.Since(pTime)
		if pCounter < 10 {
			pCounter++
			fmt.Printf("Rendering text in frame %d took: %d us\n", pCounter, pDuration.Microseconds())
		}

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

// Collection of variables that create font atlas
type FontAtlas struct {
	Font         *ttf.Font               // Font used to create the atlas
	lineSkip     int32                   // Amount of pixels in vertical direction for new line
	characters   map[rune]FontAtlasGlyph // Map of rendered characters
	AtlasSize    sdl.Point               // Size of glyph atlas
	Atlas        *sdl.Texture            // Glyph atlas
	atlasSurface *sdl.Surface            // Glyph atlas surface
	lastCharPos  sdl.Point               // Position of last character
}

// Single glyph in font atlas
type FontAtlasGlyph struct {
	AtlasBounds sdl.Rect          // Position and size of glyph in atlas texture
	Metrics     *ttf.GlyphMetrics // Glyph metrics https://freetype.org/freetype2/docs/glyphs/glyphs-3.html
}

// Creates new font atlas from provided font
func NewFontAtlas(font *ttf.Font, renderer *sdl.Renderer) FontAtlas {
	var err error
	fa := FontAtlas{
		Font:        font,
		lineSkip:    int32(font.LineSkip()),
		characters:  make(map[rune]FontAtlasGlyph, 200),
		AtlasSize:   sdl.Point{X: 512, Y: 256}, // For 24 px font rendering english alphabet takes less than half of the available space, so this size for a start is enough
		lastCharPos: sdl.Point{},
	}

	// Create font atlas surface
	fa.atlasSurface, err = sdl.CreateRGBSurface(0, fa.AtlasSize.X, fa.AtlasSize.Y, 32, 0x000000FF, 0x0000FF00, 0x00FF0000, 0xFF000000)
	if err != nil {
		log.Fatal(err)
	}
	// fa.atlasSurface.SetBlendMode(sdl.BLENDMODE_BLEND)

	// Load set of standard characters
	for i := 32; i < 127; i++ {
		fa.LoadCharacter(rune(i))
	}
	fa.UpdateTexture(renderer) // Update font atlas texture

	return fa
}

// Loads new character into font atlas. Doesn't update atlas texture.
// Returns true if character was added to font atlas, otherwise false.
func (f *FontAtlas) LoadCharacter(r rune) bool {
	// If character already exists, return
	if f.characters[r].Metrics != nil {
		return false
	}

	// Render new character into surface
	s, err := f.Font.RenderGlyphBlended(r, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		log.Fatal(err)
	}
	defer s.Free()

	// Get character metrics
	m, err := f.Font.GlyphMetrics(r)
	if err != nil {
		log.Fatal(err)
	}

	// Add character info to characters map
	f.characters[r] = FontAtlasGlyph{
		AtlasBounds: sdl.Rect{
			X: f.lastCharPos.X,
			Y: f.lastCharPos.Y,
			W: s.W,
			H: s.H},
		Metrics: m,
	}

	// Copy new character surface to font atlas surface
	s.Blit(nil, f.atlasSurface, &sdl.Rect{X: f.lastCharPos.X, Y: f.lastCharPos.Y, W: s.W, H: s.H})

	// Advance atlas position, 1px as padding
	f.lastCharPos.X += s.W + 1
	if f.lastCharPos.X >= f.AtlasSize.X-int32(f.Font.Height())-2 {
		f.lastCharPos.X = 0
		f.lastCharPos.Y += f.lineSkip + 1
	}

	return true
}

// Updates atlas texture
func (f *FontAtlas) UpdateTexture(renderer *sdl.Renderer) {
	// If font atlas texture already exists, destroy it
	if f.Atlas != nil {
		f.Atlas.Destroy()
	}

	// Create font atlas texture from font atlas surface
	var err error
	f.Atlas, err = (renderer).CreateTextureFromSurface(f.atlasSurface)
	if err != nil {
		log.Fatal(err)
	}
	// f.Atlas.SetBlendMode(sdl.BLENDMODE_BLEND)
}

// Draws single character (rune) at provided position
func (f *FontAtlas) DrawCharacter(renderer *sdl.Renderer, char rune, x, y int32, color sdl.Color) int32 {
	f.Atlas.SetColorMod(color.R, color.G, color.B)
	c := f.characters[char]
	if c.Metrics == nil {
		// Character is missing in the atlas
		if f.LoadCharacter(char) {
			f.UpdateTexture(renderer)
			c = f.characters[char]
		} else {
			return 0 // Loading new character failed, just return
		}
	}
	renderer.Copy(f.Atlas, &c.AtlasBounds,
		&sdl.Rect{
			X: x,
			Y: y,
			W: c.AtlasBounds.W,
			H: c.AtlasBounds.H})
	return int32(c.Metrics.Advance)
}

// Draws left aligned text at provided position
func (f *FontAtlas) DrawText(renderer *sdl.Renderer, text string, x, y int32, color sdl.Color) {
	currX, currY := x, y
	for _, r := range text {
		if r == '\n' {
			currX = x
			currY += f.lineSkip
			continue
		}
		currX += f.DrawCharacter(renderer, r, currX, currY, color)
	}
}

// Draws left aligned text at provided position.
// The text is wrapped at white space if extends max width.
// It's slower than DrawText() method and should be used only if necessary.
// '\n' characters can be added to the text to be wrapped in DrawText() method.
func (f *FontAtlas) DrawTextWrapped(renderer *sdl.Renderer, text string, x, y int32, color sdl.Color, maxWidth int32) {
	currX, currY := x, y
	currWhiteSpaceIndex := 0
	for i, r := range text {
		if i == currWhiteSpaceIndex {
			// Find next wite space character
			currWhiteSpaceIndex = strings.Index(text[i:], " ") // Look for character suitable for line break - space
			if currWhiteSpaceIndex < 0 {
				currWhiteSpaceIndex = len(text)
			} else {
				currWhiteSpaceIndex += i + 1
			}

			// Measure text length
			w, _, _ := f.Font.SizeUTF8(text[i:currWhiteSpaceIndex])
			if (currX-x)+int32(w) >= maxWidth {
				// Current word breaks max width rule, add line break
				currX = x
				currY += f.lineSkip
			}
		}
		if r == '\n' {
			currX = x
			currY += f.lineSkip
			continue
		}
		// Just to be sure if x position is greater than max text width proceed to next line
		if (currX - x) >= maxWidth {
			currX = x
			currY += f.lineSkip
		}
		currX += f.DrawCharacter(renderer, r, currX, currY, color)
	}
}

// Draws right aligned text at provided position
func (f *FontAtlas) DrawTextRight(renderer *sdl.Renderer, text string, x, y int32, color sdl.Color) {
	var currX, currY int32
	currY = y
	lineBreak, previousLineBreak := -1, 0
	for {
		// Search for line breaks
		lineBreak = strings.Index(text[previousLineBreak:], "\n")
		if lineBreak < 0 {
			lineBreak = len(text)
		}

		// Measure the text
		t := text[previousLineBreak:lineBreak]
		w, _, err := f.Font.SizeUTF8(t)
		if err != nil {
			log.Fatal(err)
		}
		currX = x - int32(w)

		// Draw portion of the text
		f.DrawText(renderer, t, currX, currY, color)

		if len(text) == lineBreak {
			break
		}
		currY += f.lineSkip // Add line break
		previousLineBreak = lineBreak + 1
	}
}

// Draws center aligned text at provided position
func (f *FontAtlas) DrawTextCenter(renderer *sdl.Renderer, text string, x, y int32, color sdl.Color) {
	var currX, currY int32
	currY = y
	lineBreak, previousLineBreak := -1, 0
	for {
		// Search for line breaks
		lineBreak = strings.Index(text[previousLineBreak:], "\n")
		if lineBreak < 0 {
			lineBreak = len(text)
		}

		// Measure the text
		t := text[previousLineBreak:lineBreak]
		w, _, err := f.Font.SizeUTF8(t)
		if err != nil {
			log.Fatal(err)
		}
		currX = x - int32(w)/2

		// Draw portion of the text
		f.DrawText(renderer, t, currX, currY, color)

		if len(text) == lineBreak {
			break
		}
		currY += f.lineSkip // Add line break
		previousLineBreak = lineBreak + 1
	}
}
