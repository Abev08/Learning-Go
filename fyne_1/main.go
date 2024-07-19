package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Very simple GUI app with Fyne package.
// Just a window with a text and a button that you can click.
// The button click is detected and message to console is printed.
// Fyne has prerequisites to compile: https://docs.fyne.io/started/
// Error message "... build constraints exclude all Go files in ..." means that some of the prerequisites are missing.
// On Windows it requires GCC (GNU Compiler Collection).
// It may take very long to compile...

func main() {
	var app = app.New()
	var window = app.NewWindow("Hello, World!")
	var counter = 0
	window.Resize(fyne.NewSize(640, 360))

	window.SetContent(container.NewVBox(
		widget.NewLabel("Hi!"),
		widget.NewButton("Click me", func() {
			counter++
			fmt.Printf("Button clicked %d times\n", counter)
		}),
	))

	window.ShowAndRun()
}
