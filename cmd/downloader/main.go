package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	files "github.com/ipfs/go-ipfs-files"
	icorepath "github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/wpengine/hackathon-catation/cmd/uploader/ipfs"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: downloader <cid> <destination>\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Download a CID from IPFS to a destination path.\n")
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	node, err := ipfs.Start()
	if err != nil {
		panic(err)
	}
	defer node.Close()

	cid := flag.Arg(0)
	output := flag.Arg(1)

	fmt.Printf("Fetching CID: %v\n", cid)
	testCID := icorepath.New(cid)

	rootNode, err := node.API.Unixfs().Get(context.Background(), testCID)
	if err != nil {
		panic(fmt.Errorf("Could not get file with CID: %s", err))
	}

	err = files.WriteTo(rootNode, output)
	if err != nil {
		panic(fmt.Errorf("Could not write out the fetched CID: %s", err))
	}

	fmt.Printf("Wrote the file to %s\n", output)

}
