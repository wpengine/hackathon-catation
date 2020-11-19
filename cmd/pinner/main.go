package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/wpengine/hackathon-catation/cmd/pinner/pinata"
	"github.com/wpengine/hackathon-catation/cmd/pinner/temporal"
)

func main() {
	pAPI := pinata.API{
		Key:    os.Getenv("PINATA_API_KEY"),
		Secret: os.Getenv("PINATA_SECRET_API_KEY"),
	}

	tAPI, err := temporal.New(context.Background(), os.Getenv("TEMPORAL_USERNAME"), os.Getenv("TEMPORAL_PASSWORD"))
	if err != nil {
		die("unable to connect to temporal: ", err)
	}

	if err = tAPI.Pin(context.Background(), os.Args[1]); err != nil {
		die("unable to pin to temporal ", err)
		os.Exit(1)
	}

	resp, err := pAPI.Pin(os.Args[1])
	if err != nil {
		die("unable to pin to pinata ", err)
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
		done, err := pAPI.IsPinned(os.Args[1])
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
