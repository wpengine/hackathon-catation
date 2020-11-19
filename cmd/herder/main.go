package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dchest/uniuri"
	"github.com/icza/gowut/gwu"
)

func main() {
	// Create and build a window
	win := gwu.NewWindow("main", "Herder test window")
	win.Style().SetFullWidth()
	// win.SetHAlign(gwu.HACenter)
	// win.SetCellPadding(2)

	// Build a table, each row represents one pin
	t := gwu.NewTable()
	win.Add(t)
	t.SetBorder(1)
	t.SetCellPadding(2)
	t.EnsureSize(2, 2)
	t.Add(gwu.NewLabel("Thumbnail"), 0, 0)
	t.Add(gwu.NewLabel("Hash"), 0, 1)
	t.Add(gwu.NewLabel("Filename"), 0, 2)
	files := fetchImages(".")
	hashes := map[string]*file{}
	for i, f := range files {
		t.Add(gwu.NewImage("", "/hash/"+f.hash), 1+i, 0)
		hashes[f.hash] = &files[i]
		t.Add(gwu.NewLabel(f.hash), 1+i, 1)
		t.Add(gwu.NewLabel(f.filename), 1+i, 2)
		for j, b := range f.pinned {
			c := gwu.NewCheckBox("")
			t.Add(c, 1+i, 3+j)
			c.SetState(b)
			// c.AddEHandler
		}
	}

	http.Handle("/hash/", http.StripPrefix("/hash/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := hashes[r.URL.Path]
		if f == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeContent(w, r, f.filename, time.Time{}, bytes.NewReader(f.contents))
	})))

	// Create and start a GUI server (omitting error check)
	// TODO: port choice - randomize or take flag
	server := gwu.NewServer("guitest", "localhost:8081")
	server.SetText("Herder test app")
	server.AddWin(win)
	server.Start("main")
}

type file struct {
	// thumbnail image.Image
	contents []byte
	filename string
	hash     string
	pinned   []bool
}

func fetchImages(basedir string) (files []file) {
	filepath.Walk(basedir, func(path string, info os.FileInfo, err error) error {
		switch filepath.Ext(path) {
		case ".jpg", ".jpeg", ".png":
			// ok
		default:
			return nil
		}

		data, err := readImage(path)
		if err != nil {
			log.Println("error", err)
			return nil
		}

		files = append(files, file{
			// thumbnail: img,
			contents: data,
			filename: path,
			hash:     uniuri.New(),
			pinned: []bool{
				true,
				false,
				false,
			},
		})
		return nil
	})
	return
}

func readImage(path string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	_, typ, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("parsing image %q: %w", path, err)
	}

	switch typ {
	case "png", "jpeg":
		return buf, nil
	default:
		return nil, fmt.Errorf("parsing image %q: unsupported type %q", path, typ)
	}
}

/*
import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
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
		pins:       fetchPins("."),
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
				layout.Rigid(material.Label(th, unit.Dp(10), ui.pins[i].filename).Layout),
			}
			for j := range ui.pins[i].pinned {
				row = append(row,
					layout.Rigid(material.CheckBox(th, &ui.pins[i].pinned[j], "").Layout))
			}
			return layout.Flex{}.Layout(gtx, row...)
		},
	)
}

*/
