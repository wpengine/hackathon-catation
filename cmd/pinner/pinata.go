package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type API struct {
	Key, Secret string
}

type PinResponse struct {
	ID       string `json:"id"`
	IPFSHash string `json:"ipfsHash"`
	Status   string `json:"status"`
	Name     string `json:"name"`
}

func (api *API) Pin(hash string) (*PinResponse, error) {
	payload, err := json.Marshal(map[string]string{
		"hashToPin": hash,
	})
	if err != nil {
		// Logic bug, should never happen
		panic(err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.pinata.cloud/pinning/pinByHash",
		bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("pinata: hash %q: %w", hash, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("pinata_api_key", api.Key)
	req.Header.Add("pinata_secret_api_key", api.Secret)

	// execute the request
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pinata: hash %q: %w", hash, err)
	}
	defer resp.Body.Close()

	// FIXME: if response is failed because e.g. missing API keys, return meaningful error instead of empty + nil

	// parse response
	var r PinResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return &r, fmt.Errorf("pinata: hash %q: %w", hash, err)
	}
	return &r, nil
}
