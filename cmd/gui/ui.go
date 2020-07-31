package main

import (
	"image"
	"image/color"
	"os/exec"

	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/wpengine/hackathon-catation/cmd/uploader/pinata"
)

var theme *material.Theme

type UI struct {
	imageList       *layout.List
	imageInfos      []imageInfo
	buttonClickable *widget.Clickable
}

type imageInfo struct {
	path             string
	imgData          image.Image
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
		buttonClickable: &widget.Clickable{},
	}
}

func (u *UI) layout(gtx layout.Context) {
	for range u.buttonClickable.Clicks() {
		var paths []string
		for _, imgInfo := range u.imageInfos {
			if imgInfo.checkboxSelected.Value {
				paths = append(paths, imgInfo.path)
			}
		}

		link := pinata.Upload(paths)
		_ = exec.Command("open", link).Start()
	}

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
	l := u.imageList
	return l.Layout(gtx, len(u.imageInfos), func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Flex{}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return widget.Image{Src: paint.NewImageOp(u.imageInfos[index].imgData)}.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.CheckBox(theme, u.imageInfos[index].checkboxSelected, u.imageInfos[index].path).Layout(gtx)
			}),
		)
	})
}
