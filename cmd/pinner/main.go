package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/wpengine/hackathon-catation/cmd/pinner/pinata"
)

func main() {
	api := pinata.API{
		Key:    os.Getenv("PINATA_API_KEY"),
		Secret: os.Getenv("PINATA_SECRET_API_KEY"),
	}

	resp, err := api.Pin(os.Args[1])
	if err != nil {
		die(err)
	}

	// format output
	s, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		die(err)
	}

	fmt.Println(string(s))

	// Wait until verified successful pin
	for {
		fmt.Fprintf(os.Stderr, ".")
		done, err := api.IsPinned(os.Args[1])
		if err != nil {
			panic(err)
		}
		if done {
			log.Println("pinned!")
			break
		}
	}
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
