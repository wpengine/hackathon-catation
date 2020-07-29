package main

import (
	"context"
	"fmt"
	"io"
	"os"

	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/bootstrap"
	"github.com/ipfs/go-ipfs/core/coreapi"
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

	// TODO: where do IPFS-internal temporary files get created/saved?
	node, err := core.NewNode(context.TODO(), &core.BuildCfg{
		Online: true, // TODO: should we disable Online and drop Bootstrap call?
		// NilRepo: true,  // ?
	})
	if err != nil {
		die(err)
	}
	defer node.Close()

	// FIXME: do we need this?
	// WIP: trying to resolve NAT issues
	err = node.Bootstrap(bootstrap.DefaultBootstrapConfig)
	if err != nil {
		die(err)
	}
	// TODO: node.Bootstrap() ? // https://pkg.go.dev/github.com/ipfs/go-ipfs@v0.6.0/core?tab=doc#IpfsNode.Bootstrap

	api, err := coreapi.NewCoreAPI(node)
	if err != nil {
		die(err)
	}
	stat, err := fh.Stat()
	if err != nil {
		die(err)
	}
	path, err := api.Unixfs().Add(context.TODO(), files.NewReaderStatFile(fh, stat))
	if err != nil {
		die(err)
	}
	fmt.Println(path)

	os.Stderr.WriteString("Press enter to continue: ")
	os.Stdin.Read([]byte("tmp"))

	r, err := api.Object().Data(context.TODO(), path)
	if err != nil {
		die(err)
	}
	io.Copy(os.Stdout, r)

}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
