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

	path, err := UploadFile(context.TODO(), node, &pinner, fh)
	fmt.Println(path)
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
// TODO: use interface instead of concrete *pinata.API
func UploadFile(ctx context.Context, node *ipfs.Node, pinner *pinata.API, f *os.File) (ipfspath.Resolved, error) {
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("uploading file %q to ipfs: %w", f.Name(), err)
	}
	path, err := node.AddAndPin(ctx, files.NewReaderStatFile(f, stat))
	if err != nil {
		return path, fmt.Errorf("uploading file %q to ipfs: %w", f.Name(), err)
	}

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
				log.Printf("error uploading file %q to ipfs: providing: %v", f.Name(), err)
			}
		}
	}()

	hash := path.Root() // FIXME: is this correct?
	log.Printf("Pinning %s (%s) containing %q", path, hash, f.Name())

	_, err = pinner.Pin(hash.String())
	if err != nil {
		return nil, fmt.Errorf("uploading file %q to ipfs: %w", f.Name(), err)
	}

	for {
		// context timeout?
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("uploading file %q to ipfs: %w", f.Name(), ctx.Err())
		default:
		}
		// keep checking if the file got successfully pinned
		pinned, err := pinner.IsPinned(hash.String())
		if err != nil {
			// FIXME: sometimes getting weird timeouts from pinata - rate limiting kicking in? so can't just return the error
			log.Printf("(retrying %q after error: %s)", f.Name(), err)
		}
		if pinned {
			return path, nil
		}
	}
}
