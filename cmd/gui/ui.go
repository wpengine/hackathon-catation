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

	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var theme *material.Theme

type UI struct {
	imageList        *layout.List
	buttonClickable  *widget.Clickable
	checkboxSelected *widget.Bool
}

func init() {
	theme = material.NewTheme(gofont.Collection())
}

func newUI() *UI {
	return &UI{
		imageList: &layout.List{
			Axis: layout.Vertical,
		},
		buttonClickable:  &widget.Clickable{},
		checkboxSelected: &widget.Bool{},
	}
}

func (u *UI) layout(gtx layout.Context) {
	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return u.renderHeading(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return u.renderUploadButton(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return u.renderImages(gtx)
		}),
	)
}

func (u *UI) renderHeading(gtx layout.Context) layout.Dimensions {
	l := material.H1(theme, "Select Photos To Share")
	l.Color = color.RGBA{127, 0, 0, 255} // maroon
	l.Alignment = text.Middle
	return l.Layout(gtx)
}

func (u *UI) renderUploadButton(gtx layout.Context) layout.Dimensions {
	return material.Button(theme, u.buttonClickable, "Upload").Layout(gtx)
}

func (u *UI) renderImages(gtx layout.Context) layout.Dimensions {
	images, err := cwdImages()
	if err != nil {
		return layout.Dimensions{}
	}

	l := u.imageList
	return l.Layout(gtx, len(images), func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Flex{}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return widget.Image{Src: paint.NewImageOp(images[index])}.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.CheckBox(theme, u.checkboxSelected, "label").Layout(gtx)
			}),
		)
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
