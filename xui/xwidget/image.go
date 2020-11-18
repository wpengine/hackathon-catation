// FIXME: what license (?) based on: ~/go/pkg/mod/gioui.org@v0.0.0-20200726090339-83673ecb203f/widget/image.go

package xwidget

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

// Image is a widget that displays an image.
type Image struct {
	// Src is the image to display.
	Src paint.ImageOp
	// Scale is the ratio of image pixels to
	// dps. If Scale is zero Image falls back to
	// a scale that match a standard 72 DPI.
	Scale float32
}

func (im Image) Layout(gtx layout.Context) layout.Dimensions {
	scale := im.Scale
	if scale == 0 {
		scale = 160.0 / 72.0
	}
	size := im.Src.Rect.Size()
	wf, hf := float32(size.X), float32(size.Y)
	w, h := gtx.Px(unit.Dp(wf*scale)), gtx.Px(unit.Dp(hf*scale))
	cs := gtx.Constraints
	d := cs.Constrain(image.Pt(w, h))
	stack := op.Push(gtx.Ops)
	clip.Rect(image.Rectangle{Max: d}).Add(gtx.Ops)

	// FIXME(akavel): rescaling, but I don't really know what I'm doing here with `ops` :/
	resize := float32(d.X) / float32(w)
	if resizey := float32(d.Y) / float32(h); resizey < resize {
		resize = resizey
	}
	d = image.Pt(int(float32(w)*resize), int(float32(h)*resize))
	op.Affine((&f32.Affine2D{}).Scale(f32.Pt(0, 0), f32.Pt(resize, resize))).Add(gtx.Ops)

	im.Src.Add(gtx.Ops)
	paint.PaintOp{Rect: f32.Rectangle{Max: f32.Point{X: float32(w), Y: float32(h)}}}.Add(gtx.Ops)
	stack.Pop()
	return layout.Dimensions{Size: d}
}
