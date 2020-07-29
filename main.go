package main

import (
	"fmt"
	"os"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
)

func main() {
	fn := os.Args[1]
	fh, err := os.Open(fn)
	if err != nil {
		die(err)
	}
	defer fh.Close()

	_ = coreapi.NewCoreAPI // https://pkg.go.dev/github.com/ipfs/go-ipfs@v0.6.0/core/coreapi?tab=doc#NewCoreAPI
	_ = core.NewNode       // https://pkg.go.dev/github.com/ipfs/go-ipfs@v0.6.0/core?tab=doc#NewNode
	// https://pkg.go.dev/github.com/ipfs/go-ipfs@v0.6.0/core/node?tab=doc#BuildCfg
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
