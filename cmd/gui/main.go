// Copyright (C) 2020  WPEngine
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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

	"github.com/wpengine/hackathon-catation/internal"
)

func main() {
	internal.PrintGPLBanner("catation", "2020")

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
