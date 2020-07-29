package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
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
	// generate payload from input
	var (
		hash    = os.Args[1]
		payload = pinByHashPayload{Hash: hash}
	)

	jsonPayload, err := json.Marshal(&payload)
	if err != nil {
		die(err)
	}

	req, err := http.NewRequest(http.MethodPost, pinByHashEndpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		die(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(pinataAPIKeyHeader, os.Getenv("PINATA_API_KEY"))
	req.Header.Add(pinataSecretAPIKeyHeader, os.Getenv("PINATA_SECRET_API_KEY"))

	// make request
	c := &http.Client{Timeout: 10 * time.Second}

	resp, err := c.Do(req)
	if err != nil {
		die(err)
	}
	defer resp.Body.Close()

	// parse response
	var r pinByHashResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		die(err)
	}

	// format output
	s, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		die(err)
	}

	fmt.Println(string(s))
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
