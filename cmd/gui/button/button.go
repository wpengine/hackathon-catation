package button

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
)

//Button ...
type Button struct {
	selected bool
}

//Render ...
func Render(gtx layout.Context) {
	btn := Button{
		selected: true,
	}
	btn.Layout(gtx)
}

//Layout ...
func (b *Button) Layout(gtx layout.Context) layout.Dimensions {
	defer op.Push(gtx.Ops).Pop()
	for _, e := range gtx.Events(b) {
		if e, ok := e.(pointer.Event); ok {
			switch e.Type {
			case pointer.Press:
				b.selected = true
			case pointer.Release:
				b.selected = true
			}
		}
	}

	pointer.Rect(image.Rect(0, 0, 100, 100)).Add(gtx.Ops)
	pointer.InputOp{
		Tag:   b,
		Types: pointer.Press | pointer.Release,
	}.Add(gtx.Ops)

	col := color.RGBA{R: 0x80, A: 0xFF}
	if b.selected {
		col = color.RGBA{G: 0x80, A: 0xFF}
	}

	return drawButton(gtx.Ops, col)
}

func drawButton(ops *op.Ops, color color.RGBA) layout.Dimensions {
	square := f32.Rect(0, 0, 100, 100)
	paint.ColorOp{Color: color}.Add(ops)
	paint.PaintOp{Rect: square}.Add(ops)
	return layout.Dimensions{Size: image.Pt(100, 100)}
}
