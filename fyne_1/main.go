package main

import (
	"fmt"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	var app = app.New()
	var window = app.NewWindow("Hello, World!")

	window.SetContent(container.NewVBox(
		widget.NewLabel("Hi!"),
		widget.NewButton("Click me", func() {
			fmt.Println("clicked")
		}),
	))

	window.ShowAndRun()
}
