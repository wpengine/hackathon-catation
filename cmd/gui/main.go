package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

func main() {
	go loop()
	app.Main()
}

func loop() {
	window := app.NewWindow(
		app.Title("Catation"),
	)
	ui := newUI()
	ui.images = findImages(".")

	var ops op.Ops
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
			ui.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
	os.Exit(0)
}

func findImages(basedir string) []imageRow {
	var images []imageRow

	err := filepath.Walk(basedir, func(path string, info os.FileInfo, err error) error {
		switch filepath.Ext(path) {
		case ".jpg", ".jpeg", ".png":
			// ok
		default:
			return nil
		}

		img, err := parseImage(path)
		if err != nil {
			log.Println("error", err)
			return nil
		}

		images = append(images, imageRow{
			path:     path,
			contents: img,
			selected: &widget.Bool{},
		})
		return nil
	})

	if err != nil {
		log.Println("error listing images:", err)
	}
	return images
}

func parseImage(path string) (image.Image, error) {
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
