package main

import (
	"context"
	"fmt"
	"log"
	"os"

	files "github.com/ipfs/go-ipfs-files"
	ipfspath "github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/wpengine/hackathon-catation/cmd/pinner/pinata"
	"github.com/wpengine/hackathon-catation/cmd/uploader/ipfs"
)

func main() {
	// Open the file that we want to add to IPFS
	fn := os.Args[1]
	fh, err := os.Open(fn)
	if err != nil {
		die(err)
	}
	defer fh.Close()

	// Upload the file to IPFS...

	node, err := ipfs.Start()
	if err != nil {
		panic(err)
	}
	defer node.Close()

	pinner := pinata.API{
		Key:    os.Getenv("PINATA_API_KEY"),
		Secret: os.Getenv("PINATA_SECRET_API_KEY"),
	}

	path, err := AddFile(context.TODO(), node, fh)
	if err != nil {
		die(err)
	}
	log.Printf("Pinning %s containing %q", path, fh.Name())

	err = Pin(context.TODO(), node, &pinner, path)
	if err != nil {
		die(err)
	}
	log.Printf("UPLOAD SUCCESSFUL! ---> %s", path)
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
