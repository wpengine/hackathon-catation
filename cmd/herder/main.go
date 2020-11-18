package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func main() {
	// Initialize
	window := app.NewWindow(
		app.Title("Catation Herder"),
	)
	ui := ui{
		pins: []pinUI{
			{
				thumbnail: nil, // FIXME
				filename:  "",
				hash:      "asdfasdfasdfsdf",
				pinned: []widget.Bool{
					{Value: true},
					{Value: false},
					{Value: false},
				},
			},
			{
				thumbnail: nil, // FIXME
				filename:  "",
				hash:      "hjkkhjhjkhkjhkjhjk",
				pinned: []widget.Bool{
					{Value: false},
					{Value: true},
					{Value: true},
				},
			},
		},
		pinsLayout: layout.List{Axis: layout.Vertical},
		theme:      material.NewTheme(gofont.Collection()),
	}
	// ui.images = findImages(".")

	// Start handling events
	go func() {
		var ops op.Ops
		// FIXME(akavel): do we need all stuff below? can we simplify this somehow?
		for e := range window.Events() {
			switch e := e.(type) {
			case key.Event:
				if e.Name == key.NameEscape {
					os.Exit(0)
				}
			case system.DestroyEvent:
				if e.Err != nil {
					log.Fatal(e.Err)
				}
				os.Exit(0)
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				ui.render(gtx)
				e.Frame(gtx.Ops)
			}
		}
		os.Exit(0)
	}()

	// Pass control to gioui framework
	app.Main()
}

type ui struct {
	pins       []pinUI
	pinsLayout layout.List

	theme *material.Theme
}

type pinUI struct {
	thumbnail image.Image
	filename  string
	hash      string
	pinned    []widget.Bool
}

func (ui *ui) render(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X // TODO(akavel): do we need this?
	th := ui.theme

	return ui.pinsLayout.Layout(
		gtx, len(ui.pins),
		func(gtx layout.Context, i int) layout.Dimensions {
			row := []layout.FlexChild{
				// layout.Flexed(1, xwidget.Image{Src: paint.NewImageOp(ui.pins[i].thumbnail)}.Layout),
				layout.Rigid(material.Label(th, unit.Dp(10), ui.pins[i].hash).Layout),
			}
			for j := range ui.pins[i].pinned {
				row = append(row,
					layout.Rigid(material.CheckBox(th, &ui.pins[i].pinned[j], "").Layout))
			}
			return layout.Flex{}.Layout(gtx, row...)
		},
	)
}

func fetchPins(basedir string) (pins []pinUI) {
	filepath.Walk(basedir, func(path string, info os.FileInfo, err error) error {
		switch filepath.Ext(path) {
		case ".jpg", ".jpeg", ".png":
			// ok
		default:
			return nil
		}

		img, err := readImage(path)
		if err != nil {
			log.Println("error", err)
			return nil
		}

		pins = append(pins, pinUI{
			thumbnail: img,
			filename:  path,
			hash:      "asdfasdfasdf",
			pinned: []widget.Bool{
				{Value: true},
				{Value: false},
				{Value: false},
			},
		})
		return nil
	})
	return
}

func readImage(path string) (image.Image, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("parsing image: %w", err)
	}
	defer fh.Close()

	img, typ, err := image.Decode(fh)
	if err != nil {
		return nil, fmt.Errorf("parsing image %q: %w", path, err)
	}

	switch typ {
	case "png", "jpeg":
		return img, nil
	default:
		return nil, fmt.Errorf("parsing image %q: unsupported type %q", path, typ)
	}
}
