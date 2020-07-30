package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"

	files "github.com/ipfs/go-ipfs-files"
	ipfspath "github.com/ipfs/interface-go-ipfs-core/path"

	"github.com/wpengine/hackathon-catation/cmd/builder/build"
	"github.com/wpengine/hackathon-catation/cmd/pinner/pinata"
	"github.com/wpengine/hackathon-catation/cmd/shortener/bitly"
	"github.com/wpengine/hackathon-catation/cmd/uploader/ipfs"
)

func main() {
	// TODO: check if this can help cleanup something: https://github.com/ipfs/go-ipfs/blob/master/docs/examples/go-ipfs-as-a-library/README.md

	if len(os.Args) <= 1 || os.Args[1] == "--help" {
		fmt.Printf("Usage: %s IMAGE_PATH...\n", os.Args[0])
		os.Exit(2)
	}

	// Verify that required API keys are configured
	pinner := pinata.API{
		Key:    os.Getenv("PINATA_API_KEY"),
		Secret: os.Getenv("PINATA_SECRET_API_KEY"),
	}
	if pinner.Key == "" || pinner.Secret == "" {
		die("please set pinata API key env variables PINATA_API_KEY and PINATA_SECRET_API_KEY to proper values (see http://pinata.cloud)")
	}
	shortener := bitly.API{
		Key: os.Getenv("BITLY_API_KEY"),
	}
	if shortener.Key == "" {
		die("please set bitly API key env variable BITLY_API_KEY to proper value (see http://bitly.com)")
	}

	// Initialize IPFS
	node, err := ipfs.Start()
	if err != nil {
		panic(err)
	}
	defer node.Close()

	// Upload the files to IPFS...
	var (
		images   = os.Args[1:]
		hashes   = make([]string, len(images))
		wgAdd    sync.WaitGroup
		wgPin    sync.WaitGroup
		progress int32
	)
	for i, fn := range images {
		wgAdd.Add(1)
		wgPin.Add(1)
		go func(i int, fn string) {
			// Open the file that we want to add to IPFS
			fh, err := os.Open(fn)
			if err != nil {
				die(err)
			}
			defer fh.Close()

			ipfsPath, err := AddFile(context.TODO(), node, fh)
			if err != nil {
				die(err)
			}
			hashes[i] = ipfsPath.Root().String()
			wgAdd.Done()

			err = Pin(context.TODO(), node, &pinner, ipfsPath)
			if err != nil {
				die(err)
			}
			log.Printf("pinned %d/%d: %s",
				atomic.AddInt32(&progress, 1), len(images),
				fh.Name())
			wgPin.Done()
		}(i, fn)
	}

	wgAdd.Wait()

	indexHTML, err := build.IndexHTML(hashes...)
	if err != nil {
		die(err)
	}
	pathIndex, err := AddIndexHTML(context.TODO(), node, indexHTML)
	if err != nil {
		die(err)
	}
	log.Println("index.html -->", pathIndex)

	log.Printf("Pinning %s containing %q", pathIndex, "index.html")
	err = Pin(context.TODO(), node, &pinner, pathIndex)
	if err != nil {
		die(err)
	}
	log.Printf("UPLOAD SUCCESSFUL! ---> %s", pathIndex)

	// URL shortener
	hash := pathIndex.Root().String()
	link, err := shortener.Shorten("http://ipfs.io/ipfs/" + hash)
	if err != nil {
		die(err)
	}

	log.Println("Waiting for remaining images...")
	wgPin.Wait()

	fmt.Printf(`

>>>>>
>>>>>     %s
>>>>>
`, link)
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}

// TODO: use interface instead of concrete *ipfs.Node
func AddFile(ctx context.Context, node *ipfs.Node, f *os.File) (ipfspath.Resolved, error) {
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("adding file %q to ipfs: %w", f.Name(), err)
	}
	path, err := node.AddAndPin(ctx, files.NewReaderStatFile(f, stat))
	if err != nil {
		return path, fmt.Errorf("adding file %q to ipfs: %w", f.Name(), err)
	}
	return path, nil
}

// TODO: use interface instead of concrete *ipfs.Node
func AddIndexHTML(ctx context.Context, node *ipfs.Node, contents []byte) (ipfspath.Resolved, error) {
	path, err := node.AddAndPin(ctx, files.NewMapDirectory(map[string]files.Node{
		"index.html": files.NewBytesFile(contents),
	}))
	if err != nil {
		return path, fmt.Errorf("adding index.html (%d B) to ipfs: %w", len(contents), err)
	}
	return path, nil
}

// TODO: use interface instead of concrete *ipfs.Node
// TODO: use interface instead of concrete *pinata.API
func Pin(ctx context.Context, node *ipfs.Node, pinner *pinata.API, path ipfspath.Resolved) error {
	subctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Provide in background, until upload succeeds
	go func() {
		for {
			if subctx.Err() != nil {
				// cancelled, quit silently
				return
			}
			// keep providing the path...
			err := node.Provide(subctx, path)
			if subctx.Err() != nil {
				// cancelled, quit silently
				return
			}
			if err != nil {
				log.Printf("error uploading %s to ipfs: providing: %v", path, err)
			}
		}
	}()

	hash := path.Root() // FIXME: is this correct?

	_, err := pinner.Pin(hash.String())
	if err != nil {
		return fmt.Errorf("pinning %q: %w", path, err)
	}

	for {
		// context timeout?
		select {
		case <-ctx.Done():
			return fmt.Errorf("pinning %q: %w", path, ctx.Err())
		default:
		}
		// keep checking if the file got successfully pinned
		pinned, err := pinner.IsPinned(hash.String())
		if err != nil {
			// FIXME: sometimes getting weird timeouts from pinata - rate limiting kicking in? so can't just return the error
			log.Printf("(retrying after error: %s)", err)
		}
		if pinned {
			return nil
		}
	}
}
