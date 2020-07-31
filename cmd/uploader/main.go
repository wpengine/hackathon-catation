package main

import (
	"fmt"
	"os"

	"github.com/wpengine/hackathon-catation/cmd/uploader/pinata"
)

func main() {
	// TODO: check if this can help cleanup something: https://github.com/ipfs/go-ipfs/blob/master/docs/examples/go-ipfs-as-a-library/README.md

	if len(os.Args) <= 1 || os.Args[1] == "--help" {
		fmt.Printf("Usage: %s IMAGE_PATH...\n", os.Args[0])
		os.Exit(2)
	}

	pinata.Upload(os.Args[1:])
}
