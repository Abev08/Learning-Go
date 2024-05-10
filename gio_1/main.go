package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func main() {
	window := new(app.Window)

	// Func for minimum fps
	go func() {
		for {
			window.Invalidate()
			time.Sleep(time.Second)
		}
	}()

	go func() {
		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}

func run(window *app.Window) error {
	theme := material.NewTheme()
	var ops op.Ops

	var button widget.Clickable
	var labelButton widget.Clickable
	var buttonClickCounter int

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceEnd,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					for labelButton.Clicked(gtx) {
						fmt.Println("Label clicked")
					}

					// Define an large label with an appropriate text:
					title := material.H1(theme, "Hello, Gio")

					// Change the color of the label.
					maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
					title.Color = maroon

					// Change the position of the label.
					title.Alignment = text.Middle

					// Set label clicable area to the title size
					labelButton.Layout(gtx, title.Layout)

					// Draw the label to the graphics context.
					return title.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					for button.Clicked(gtx) {
						buttonClickCounter++
						fmt.Printf("Button clicked %d times\n", buttonClickCounter)
					}
					return material.Button(theme, &button, "Click me").Layout(gtx)
				}),
			)

			// Pass the drawing operations to the GPU.
			e.Frame(gtx.Ops)
		}
	}
}
