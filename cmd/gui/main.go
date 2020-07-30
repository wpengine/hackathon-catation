package main

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		w := app.NewWindow(
			app.Title("Catation"),
		)
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func Layout(gtx layout.Context) {
	//renderHeading(gtx)
	renderImages(gtx)

}

func renderHeading(gtx layout.Context) {
	th := material.NewTheme(gofont.Collection())
	l := material.H1(th, "Select Photos To Share")
	l.Color = color.RGBA{127, 0, 0, 255} // maroon
	l.Alignment = text.Middle
	l.Layout(gtx)
}

func renderImages(gtx layout.Context) layout.Dimensions {
	images, err := cwdImages()
	if err != nil {
		return layout.Dimensions{}
	}

	var list layout.List
	return list.Layout(gtx, len(images), func(gtx layout.Context, index int) layout.Dimensions {
		return widget.Image{Src: paint.NewImageOp(images[index])}.Layout(gtx)
	})
}

func cwdImages() ([]image.Image, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	paths, err := listImagePaths(dir)
	if err != nil {
		return nil, err
	}

	var images []image.Image
	for _, path := range paths {
		img, err := parseImage(path)
		if err != nil {
			log.Println("error parsing image", err)
			continue
		}
		images = append(images, img)
	}

	return images, nil
}

func parseImage(path string) (image.Image, error) {
	fh, err := os.Open(path)
	if err != nil {
		log.Printf("error opening image %s", path)
		return nil, err
	}

	defer fh.Close()

	_, imgType, err := image.Decode(fh)
	if err != nil {
		log.Printf("invalid image format %s", path)
		return nil, err
	}

	fh.Seek(0, 0) // reset to beginning of file to avoid EOF

	switch imgType {
	case "png":
		return png.Decode(fh)
	case "jpeg":
		return jpeg.Decode(fh)
	default:
		return nil, errors.New("unsupported image type")
	}
}

func listImagePaths(dir string) (paths []string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		switch filepath.Ext(path) {
		case ".jpg", ".jpeg", ".png":
			paths = append(paths, path)
		}
		return nil
	})

	return
}
