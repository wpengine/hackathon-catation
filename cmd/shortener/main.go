package main

import (
	"fmt"
	"os"

	"github.com/wpengine/hackathon-catation/cmd/shortener/bitly"
)

func main() {
	hash := os.Args[1]

	bitly := bitly.API{
		Key: os.Getenv("BITLY_API_KEY"),
	}

	link, err := bitly.Shorten("http://ipfs.io/ipfs/" + hash)
	if err != nil {
		die(err)
	}

	fmt.Println(link)
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
