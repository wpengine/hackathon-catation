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
	"image"
	"image/color"
	"os/exec"

	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/wpengine/hackathon-catation/cmd/uploader/pinata"
)

type UI struct {
	images []imageRow

	// Widgets
	imageList    *layout.List
	uploadButton *widget.Clickable
	theme        *material.Theme
}

type imageRow struct {
	path     string
	contents image.Image

	// Widgets
	selected widget.Bool
}

func newUI() *UI {
	return &UI{
		imageList: &layout.List{
			Axis: layout.Vertical,
		},
		uploadButton: &widget.Clickable{},
		theme:        material.NewTheme(gofont.Collection()),
	}
}

func (u *UI) layout(gtx layout.Context) {
	for range u.uploadButton.Clicks() {
		var paths []string
		for _, img := range u.images {
			if img.selected.Value {
				paths = append(paths, img.path)
			}
		}

		link := pinata.Upload(paths)
		_ = exec.Command("open", link).Start()
	}

	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(u.renderHeading),
		layout.Rigid(u.renderUploadButton),
		layout.Rigid(u.renderImages),
	)
}

func (u *UI) renderHeading(gtx layout.Context) layout.Dimensions {
	l := material.H3(u.theme, "Select Photos To Share")
	l.Color = color.RGBA{127, 0, 0, 255} // maroon
	l.Alignment = text.Middle
	return l.Layout(gtx)
}

func (u *UI) renderUploadButton(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(10)).Layout(gtx,
		material.Button(u.theme, u.uploadButton, "Upload").Layout)
}

func (u *UI) renderImages(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	l := u.imageList
	return l.Layout(gtx, len(u.images), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.Flex{}.Layout(gtx,
			layout.Flexed(1, widget.Image{Src: paint.NewImageOp(u.images[i].contents)}.Layout),
			layout.Rigid(material.CheckBox(u.theme, &u.images[i].selected, u.images[i].path).Layout),
		)
	})
}
