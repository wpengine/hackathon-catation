package main

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
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
	ui.imageInfos = getCWDImageInfos()

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

func getCWDImageInfos() []imageInfo {
	images, err := cwdImages()
	if err != nil {
		return nil
	}

	var imgInfos []imageInfo
	for path, data := range images {
		imgInfos = append(imgInfos, imageInfo{
			path:    path,
			imgData: data,
			checkboxSelected: &widget.Bool{
				Value: false,
			},
		})
	}

	return imgInfos
}

func cwdImages() (map[string]image.Image, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	paths, err := listImagePaths(dir)
	if err != nil {
		return nil, err
	}

	images := make(map[string]image.Image)
	for _, path := range paths {
		img, err := parseImage(path)
		if err != nil {
			log.Println("error parsing image", err)
			continue
		}
		images[path] = img
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

	_, _ = fh.Seek(0, 0) // reset to beginning of file to avoid EOF

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
