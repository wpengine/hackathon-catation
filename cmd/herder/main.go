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
	"github.com/wpengine/hackathon-catation/internal"
	"github.com/wpengine/hackathon-catation/pup"
	"github.com/wpengine/hackathon-catation/pup/eternum"
	"github.com/wpengine/hackathon-catation/pup/pinata"
	"github.com/wpengine/hackathon-catation/pup/pipin"
)

type config struct {
	Pinata  *pinata.API
	Pipin   *pipin.Client
	Eternum *eternum.Client
}

func main() {
	internal.PrintGPLBanner("herder", "2020")

	trigger := make(chan struct{}, 1)
	trigger <- struct{}{}
	TRIGGER := func() {
		select {
		case trigger <- struct{}{}:
		default:
		}
	}
	{
		tick := time.NewTicker(60 * time.Second)
		go func() {
			for range tick.C {
				TRIGGER()
			}
		}()
	}

	// Read config file
	cfg := readConfig()

	// Start IPFS
	node, err := ipfs.Start()
	if err != nil {
		panic(err)
	}
	defer node.Close()

	// Initialize Pup plugins
	type pupColumn struct {
		i    int
		name string
		pup.Pup
	}
	pups := []pupColumn{}
	if cfg.Pipin != nil {
		pups = append(pups, pupColumn{len(pups), "pipin", cfg.Pipin})
	}
	if cfg.Pinata != nil {
		pups = append(pups, pupColumn{len(pups), "pinata", cfg.Pinata})
	}
	if cfg.Eternum != nil {
		pups = append(pups, pupColumn{len(pups), "eternum", cfg.Eternum})
	}

	// In a background loop, start fetching hashes from pups, to be fed into
	// the GUI table.
	//
	// rowChange is a message describing how the GUI should toggle a checkbox
	// for a particular file's row
	type rowChange struct {
		*file        // basic data of the row (esp. in case it needs to be newly added)
		ipup    int  // which pup's checkbox to change
		checked bool // to what state should the pup's checkbox be changed
	}
	rowChanges := make(chan rowChange, 100)
	go func() {
		hashes := map[string]*file{}
		// Infinite loop, iterating over all pups
		for {
			<-trigger
			for ii := 0; ii < 2; ii++ {
				for _, p := range pups {
					ctx := context.Background()
					ctx, release := context.WithTimeout(ctx, 10*time.Second)
					// TODO: protect against panic
					cids, err := p.Fetch(ctx, nil)
					release()
					if err != nil {
						log.Printf("Cannot fetch from %q: %s", p.name, err)
						continue
					}
					log.Printf("Fetched %v items from %q", len(cids), p.name)
					fetched := map[string]bool{}
					for _, c := range cids {
						fetched[c.Hash] = true
					}
					// Un-check all hashes not in fetched
					for _, f := range hashes {
						// log.Printf("%q TEST %s fetched? %v", p.name, f.hash, fetched[f.hash])
						if !fetched[f.hash] {
							// log.Printf("FETCH UNPIN %s @ %v %q", f.hash, p.i, p.name)
							rowChanges <- rowChange{f, p.i, false}
						}
					}
					// Add missing hashes
					for _, c := range cids {
						f := &file{
							hash:     c.Hash,
							filename: c.Name,
							pinned:   make([]bool, len(pups)),
						}
						hashes[f.hash] = f // TODO: do we need this Map? seems unused elsewhere
						rowChanges <- rowChange{f, p.i, true}
					}
				}
				// time.Sleep(1 * time.Second)
				time.Sleep(2 * time.Second)
			}
		}
	}()

	// Create and build a window
	win := gwu.NewWindow("main", "Herder test window")
	win.Style().SetFullWidth()
	win.Add(gwu.NewHTML(`<h1>Catation Forever!</h1>`))
	// win.SetHAlign(gwu.HACenter)
	// win.SetCellPadding(2)

	// Start building a table, each row will represent one file
	rowsByHash := map[string]struct {
		y        int
		statuses []gwu.Panel
	}{}
	t := gwu.NewTable()
	win.Add(t)
	t.SetBorder(1)
	t.SetCellPadding(2)
	t.EnsureSize(2, 2)
	t.Add(gwu.NewLabel("Thumbnail"), 0, 0)
	t.Add(gwu.NewLabel("Hash"), 0, 1)
	t.Add(gwu.NewLabel("Filename"), 0, 2)
	for _, p := range pups {
		t.Add(gwu.NewLabel(p.name), 0, 3+p.i)
	}
	// Every second, if there are new rows fetched, add them to the table
	thumbnailsByHash := sync.Map{}       // map[string][]byte
	thumbnails := make(chan string, 100) // TODO: rename&refactor, e.g.: thumbnails fetched
	{
		s := gwu.NewTimer(1 * time.Second)
		win.Add(s)
		s.SetRepeat(true)
		s.AddEHandlerFunc(func(e gwu.Event) {
			// log.Println("TIMER(rowChanges) begin")
			// defer log.Println("TIMER(rowChanges) end")
			for i := 0; i < 100; i++ { // make sure we don't loop forever in event handler
				select {
				case f := <-rowChanges:
					r := rowsByHash[f.hash]

					// Do we need to add a new row?
					if r.statuses == nil {
						r.y = len(rowsByHash) + 1
						log.Printf("new row: %v @ %v", f.hash, r.y)
						t.Add(gwu.NewImage("", "/hash/"+f.hash), r.y, 0)
						t.Add(gwu.NewLabel(f.hash), r.y, 1)
						t.Add(gwu.NewLabel(f.filename), r.y, 2)
						for _, p := range pups {
							p := p // capture the loop variable for use in closures

							cell := gwu.NewHorizontalPanel()
							t.Add(cell, r.y, 3+p.i)
							cell.SetCellPadding(5)
							r.statuses = append(r.statuses, cell)

							// c := gwu.NewCheckBox("")
							// cell.Add(c)
							// c.SetEnabled(false) // read-only, showing current status in pup

							add := gwu.NewButton("📌")
							cell.Add(add)
							add.AddEHandlerFunc(func(e gwu.Event) {
								ctx, release := context.WithTimeout(context.Background(), 2*time.Second)
								defer release()
								// TODO: make it more async & faster
								// log.Printf("--- check %s / %q = %v ---", f.hash, p.name, v)
								err := p.Pup.Pin(ctx, f.hash)
								if err != nil {
									log.Printf("%s.Pin error: %s", p.name, err)
									return
								}
								log.Printf("%s.Pin success", p.name)
								TRIGGER()
							}, gwu.ETypeClick)

							rm := gwu.NewButton("🗑")
							cell.Add(rm)
							rm.AddEHandlerFunc(func(e gwu.Event) {
								ctx, release := context.WithTimeout(context.Background(), 2*time.Second)
								defer release()
								// TODO: make it more async & faster
								// log.Printf("--- check %s / %q = %v ---", f.hash, p.name, v)
								err := p.Pup.Unpin(ctx, f.hash)
								if err != nil {
									log.Printf("%s.Unpin error: %s", p.name, err)
									return
								}
								log.Printf("%s.Unpin success", p.name)
								TRIGGER()
							}, gwu.ETypeClick)
						}
						e.MarkDirty(t)
						go fetchThumbnail(thumbnails, &thumbnailsByHash, node, f.hash)
					}

					// Change the status of a checkbox
					if f.checked {
						r.statuses[f.ipup].Style().SetBackground("#00ff00")
					} else {
						r.statuses[f.ipup].Style().SetBackground("#ffffff")
					}
					rowsByHash[f.hash] = r
					// e.MarkDirty(r.statuses[f.ipup])
					e.MarkDirty(t)

				default:
					return
				}
			}
		}, gwu.ETypeStateChange)
	}

	// FIXME: somehow sort the images (how? by hash??? :/)

	// Start a timer, to detect when new thumbnails are ready and show them
	s := gwu.NewTimer(1 * time.Second)
	win.Add(s)
	s.SetRepeat(true)
	s.AddEHandlerFunc(func(e gwu.Event) {
		// log.Println("TIMER(thumbnails) begin")
		// defer log.Println("TIMER(thumbnails) end")
		select {
		case h := <-thumbnails:
			log.Printf("THUMB %s", h)
			_ = h // TODO: only refresh specific thumbnail img
			e.MarkDirty(t)
		default:
		}
	}, gwu.ETypeStateChange)

	// Serve thumbnails over HTTP for <img src="/hash/...">
	http.Handle("/hash/", http.StripPrefix("/hash/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, ok := thumbnailsByHash.Load(r.URL.Path)
		log.Printf("GET thumb=%q ok=%v", r.URL.Path, ok)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		th := v.([]byte)
		// TODO: if filename is somehow known, try using it instead of "" below
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(th))
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
		fmt.Fprintln(os.Stderr, "HINT: example config.json (not all entries are required!):")
		v, _ := json.MarshalIndent(config{
			Pinata:  &pinata.API{},
			Pipin:   &pipin.Client{},
			Eternum: &eternum.Client{},
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
	// row      int
	// contents []byte
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

func fetchThumbnail(fetched chan<- string, thumbnailsByHash *sync.Map, node *ipfs.Node, hash string) {
	log.Printf("%s - starting to fetch...", hash)
	tree, err := node.API.Unixfs().Get(context.Background(), icorepath.New(hash))
	if err != nil {
		log.Printf("Could not get file with CID: %s", err)
		return
	}
	log.Printf("%s - found", hash)
	switch tree := tree.(type) {
	case ifiles.File:
		log.Printf("%s - is a file, thumbnailing", hash)
		th, err := thumbnailImage(tree, 100, 100)
		if err != nil {
			log.Printf("Could not create thumbnail of %s: %s", hash, err)
			return
		}
		thumbnailsByHash.Store(hash, th)
		log.Printf("%s - DONE", hash)
		fetched <- hash
	default:
		log.Printf("%s - is not a file, ignoring", hash)
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
