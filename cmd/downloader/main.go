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
	"context"
	"flag"
	"fmt"
	"os"

	files "github.com/ipfs/go-ipfs-files"
	icorepath "github.com/ipfs/interface-go-ipfs-core/path"

	"github.com/wpengine/hackathon-catation/cmd/uploader/ipfs"
	"github.com/wpengine/hackathon-catation/internal"
)

func main() {
	internal.PrintGPLBanner("catation", "2020")

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
