package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	pinataAPIKeyHeader       = "pinata_api_key"
	pinataSecretAPIKeyHeader = "pinata_secret_api_key"

	pinByHashEndpoint = "https://api.pinata.cloud/pinning/pinByHash"
)

type (
	pinByHashPayload struct {
		Hash string `json:"hashToPin"`
	}

	pinByHashResponse struct {
		ID       string `json:"id"`
		IPFSHash string `json:"ipfsHash"`
		Status   string `json:"status"`
		Name     string `json:"name"`
	}
)

func main() {
	api := API{
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
			fmt.Fprintln(os.Stderr, "pinned!")
			break
		}
	}
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
