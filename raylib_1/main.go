package main

import (
	"fmt"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Very simple GUI app with raylib package.
// Just a window with a text (label) and a button that you can click.
// The button is clickable and on clicks it increases a counter and prints a message to console.
// TTF font is used for on screen font rendering.
// On Windows raylib requires GCC to build.

var BackgroundColor = rl.NewColor(20, 20, 20, 255)
var Font rl.Font
var WindowSize = rl.NewVector2(640, 480)
var count = 0

func main() {
	rl.InitWindow(int32(WindowSize.X), int32(WindowSize.Y), "Test")
	defer cleanup()
	rl.SetTargetFPS(30)                       // Limit FPS to 30
	rl.SetWindowState(rl.FlagWindowResizable) // Allow window to be resized

	// Load font
	Font = rl.LoadFontEx("assets/fonts/OpenSans-SemiBold.ttf", 24, nil, 250)

	// Main loop
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(BackgroundColor)

		drawUi()

		rl.EndDrawing()
	}
}

func cleanup() {
	rl.UnloadFont(Font)
	rl.CloseWindow()
}

func drawUi() {
	// Draw text in the middle of top row
	text := "Hello, World!"
	textSize := rl.MeasureTextEx(Font, text, float32(Font.BaseSize), 1)
	rl.DrawTextEx(Font, text, rl.NewVector2(WindowSize.X/2-textSize.X/2, 10), float32(Font.BaseSize), 1, rl.White)

	// Draw button from raygui package
	if gui.Button(rl.NewRectangle(10, 40, 120, 30), "Click me") {
		count++
		fmt.Printf("Clicked %d times\n", count)
	}

	// Draw font atlas just because we can
	rl.DrawTexture(Font.Texture, 2, 100, rl.White)
}
