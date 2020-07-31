package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
)

type App struct {
	w  *app.Window
	ui *UI
}

func main() {
	go func() {
		a := &App{
			w: app.NewWindow(
				app.Title("Catation"),
			),
			ui: newUI(),
		}

		if err := loop(a); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(a *App) error {
	var ops op.Ops
	for {
		e := <-a.w.Events()
		switch e := e.(type) {
		case key.Event:
			if e.Name == key.NameEscape {
				os.Exit(0)
			}
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			a.ui.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}
