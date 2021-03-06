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

package pinata

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
		return nil, fmt.Errorf("pinata: adding hash %q: %w", hash, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("pinata_api_key", api.Key)
	req.Header.Add("pinata_secret_api_key", api.Secret)

	// execute the request
	// TODO: use context.Context instead of raw timeout ?
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pinata: adding hash %q: %w", hash, err)
	}
	defer resp.Body.Close()

	// FIXME: if response is failed because e.g. missing API keys, return meaningful error instead of empty + nil

	// parse response
	var r PinResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return &r, fmt.Errorf("pinata: adding hash %q: decoding response: %w", hash, err)
	}
	return &r, nil
}

func (api *API) IsPinned(hash string) (bool, error) {
	// TODO: use some metadata, otherwise this is very ineffective and currently limited to 1000 pins

	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.pinata.cloud/data/pinList"+
			"?status=pinned",
		nil)
	if err != nil {
		return false, fmt.Errorf("pinata: querying hash %q: %w", hash, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("pinata_api_key", api.Key)
	req.Header.Add("pinata_secret_api_key", api.Secret)

	// execute the request
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return false, fmt.Errorf("pinata: querying hash %q: %w", hash, err)
	}
	defer resp.Body.Close()

	// FIXME: if response is failed because e.g. missing API keys, return meaningful error instead of empty + nil

	// parse response
	var r struct {
		Rows []struct {
			IPFSPinHash string `json:"ipfs_pin_hash"`
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return false, fmt.Errorf("pinata: querying hash %q: decoding response: %w", hash, err)
	}

	for _, row := range r.Rows {
		if row.IPFSPinHash == hash {
			return true, nil
		}
	}
	return false, nil
}
