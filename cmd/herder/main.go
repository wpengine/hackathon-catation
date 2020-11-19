package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/icza/gowut/gwu"
	ifiles "github.com/ipfs/go-ipfs-files"
	icorepath "github.com/ipfs/interface-go-ipfs-core/path"
	"golang.org/x/image/draw"

	"github.com/wpengine/hackathon-catation/cmd/uploader/ipfs"
	"github.com/wpengine/hackathon-catation/pup/pinata"
)

type config struct {
	Pinata *pinata.API
}

func main() {
	// Read config file
	cfg := readConfig()

	// Start IPFS
	node, err := ipfs.Start()
	if err != nil {
		panic(err)
	}
	defer node.Close()

	// Fetch hashes using pup/pinata/ API
	cids, err := cfg.Pinata.Fetch(nil)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: fetch thumbnails of those hashes using cmd/downloader/
	hashes := sync.Map{}
	thumbnails := make(chan string, 100)
	for _, c := range cids {
		f := &file{
			hash:     c.Hash,
			filename: c.Name,
		}
		_, loaded := hashes.LoadOrStore(c.Hash, f)
		if loaded {
			continue
		}
		// new file - start fetching it in background
		go func() {
			log.Printf("%s - starting to fetch...", f.hash)
			tree, err := node.API.Unixfs().Get(context.Background(), icorepath.New(f.hash))
			if err != nil {
				log.Printf("Could not get file with CID: %s", err)
				return
			}
			log.Printf("%s - found", f.hash)
			switch tree := tree.(type) {
			case ifiles.File:
				log.Printf("%s - is a file, thumbnailing", f.hash)
				th, err := thumbnailImage(tree, 100, 100)
				if err != nil {
					log.Printf("Could not create thumbnail of %s: %s", f.hash, err)
					return
				}
				f.contents = th
				hashes.Store(f.hash, f)
				log.Printf("%s - DONE", f.hash)
				thumbnails <- f.hash
			default:
				log.Printf("%s - is not a file, ignoring", f.hash)
			}
		}()
	}

	// Create and build a window
	win := gwu.NewWindow("main", "Herder test window")
	win.Style().SetFullWidth()
	// win.SetHAlign(gwu.HACenter)
	// win.SetCellPadding(2)

	// Build a table, each row represents one file
	t := gwu.NewTable()
	win.Add(t)
	t.SetBorder(1)
	t.SetCellPadding(2)
	t.EnsureSize(2, 2)
	t.Add(gwu.NewLabel("Thumbnail"), 0, 0)
	t.Add(gwu.NewLabel("Hash"), 0, 1)
	t.Add(gwu.NewLabel("Filename"), 0, 2)
	// FIXME: somehow sort the images (how? by hash??? :/)
	i := 0
	hashes.Range(func(_, value interface{}) bool {
		i++
		f := value.(*file)
		t.Add(gwu.NewImage("", "/hash/"+f.hash), i, 0)
		t.Add(gwu.NewLabel(f.hash), i, 1)
		t.Add(gwu.NewLabel(f.filename), i, 2)
		for j, b := range f.pinned {
			c := gwu.NewCheckBox("")
			t.Add(c, i, 3+j)
			c.SetState(b)
			// c.AddEHandler
		}
		return true // continue iterating
	})

	// Start a timer, to detect when new thumbnails are ready and show them
	s := gwu.NewTimer(1 * time.Second)
	win.Add(s)
	s.SetRepeat(true)
	s.AddEHandlerFunc(func(e gwu.Event) {
		select {
		case h := <-thumbnails:
			_ = h // TODO: only refresh specific thumbnail img
			e.MarkDirty(t)
		default:
		}
	}, gwu.ETypeStateChange)

	// Serve thumbnails over HTTP for <img src="/hash/...">
	http.Handle("/hash/", http.StripPrefix("/hash/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, ok := hashes.Load(r.URL.Path)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		f := v.(*file)
		http.ServeContent(w, r, f.filename, time.Time{}, bytes.NewReader(f.contents))
	})))

	// Create and start a GUI server (omitting error check)
	// TODO: port choice - randomize or take flag
	server := gwu.NewServer("guitest", "localhost:8081")
	server.SetText("Herder test app")
	server.AddWin(win)
	server.Start("main")
}

func readConfig() config {
	raw, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot read config.json:", err)
		fmt.Fprintln(os.Stderr, "HINT: example config.json:")
		v, _ := json.MarshalIndent(config{
			Pinata: &pinata.API{Key: "", Secret: ""},
		}, "", "  ")
		fmt.Fprintln(os.Stderr, string(v))
		os.Exit(1)
	}
	cfg := config{}
	err = json.Unmarshal(raw, &cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot decode config.json:", err)
		os.Exit(1)
	}
	return cfg
}

type file struct {
	// thumbnail image.Image
	contents []byte
	filename string
	hash     string
	pinned   []bool
}

func thumbnailImage(r io.Reader, maxw, maxh int) ([]byte, error) {
	src, typ, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("parsing image for thumbnail: %w", err)
	}

	// Calculate thumbnail size
	srcb := src.Bounds()
	if srcb.Dx() > maxw || srcb.Dy() > maxh {
		sx := float64(maxw) / float64(srcb.Dx())
		sy := float64(maxh) / float64(srcb.Dy())
		scale := sx
		if sy < sx {
			scale = sy
		}
		maxw = int(float64(srcb.Dx()) * scale)
		maxh = int(float64(srcb.Dy()) * scale)
	} else {
		maxw, maxh = srcb.Dx(), srcb.Dy()
	}

	// Render the thumbnail
	dst := image.NewRGBA(image.Rect(0, 0, maxw, maxh))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, srcb, draw.Src, nil)

	// Encode the thumbnail back to original format
	buf := bytes.NewBuffer(nil)
	switch typ {
	case "png":
		err = png.Encode(buf, dst)
		if err != nil {
			return nil, fmt.Errorf("encoding png thumbnail: %w", err)
		}
		return buf.Bytes(), nil
	default:
		err = jpeg.Encode(buf, dst, nil)
		if err != nil {
			return nil, fmt.Errorf("encoding jpeg thumbnail: %w", err)
		}
		return buf.Bytes(), nil
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
