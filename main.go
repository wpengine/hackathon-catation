package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ipfs/go-ipfs/core"
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

	// Upload the file to IPFS
	// TODO: where do IPFS-internal temporary files get created/saved?
	node, err := core.NewNode(context.TODO(), &core.BuildCfg{
		Online: true,
		// NilRepo: true,  // ?
	})
	if err != nil {
		die(err)
	}
	defer node.Close()
	_ = coreapi.NewCoreAPI // https://pkg.go.dev/github.com/ipfs/go-ipfs@v0.6.0/core/coreapi?tab=doc#NewCoreAPI
	// https://pkg.go.dev/github.com/ipfs/go-ipfs@v0.6.0/core/node?tab=doc#BuildCfg
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
