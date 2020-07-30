package main

import (
	"context"
	"fmt"
	"log"
	"os"

	files "github.com/ipfs/go-ipfs-files"
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

	stat, err := fh.Stat()
	if err != nil {
		die(err)
	}
	path, err := node.AddAndPin(context.TODO(), files.NewReaderStatFile(fh, stat))
	if err != nil {
		die(err)
	}
	fmt.Println(path)

	// try to make sure the file is pinned and visible
	log.Println("providing...")
	err = node.API.Dht().Provide(context.TODO(), path)
	if err != nil {
		panic(err)
	}

	log.Println("finding providers...")
	providersChan, err := node.API.Dht().FindProviders(context.TODO(), path)
	if err != nil {
		die(err)
	}
	for p := range providersChan {
		fmt.Println(p)
	}

	os.Stderr.WriteString("Press enter to continue: ")
	os.Stdin.Read([]byte("tmp"))

	// r, err := node.API.Object().Data(context.TODO(), path)
	// if err != nil {
	// 	die(err)
	// }
	// io.Copy(os.Stdout, r)
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
