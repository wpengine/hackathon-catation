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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wpengine/hackathon-catation/pup"
)

// FIXME: unify with cmd/pinner/pinata/
// FIXME[LATER]: unify with cmd/uploader/pinata/ ?

type API struct {
	Key, Secret string
}

func New(key, secret string) *API {
	return &API{Key: key, Secret: secret}
}

func (api *API) Fetch(ctx context.Context, filter []pup.Hash) ([]pup.NamedHash, error) {
	// TODO: use some metadata, otherwise this func is very ineffective and currently limited to 1000 pins (TODO: first, check if they didn't publish some newer API)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.pinata.cloud/data/pinList?status=pinned",
		nil,
	)
	if err != nil {
		// Logic bug, should never happen
		panic(fmt.Errorf("pinata: building fetch request: %w", err))
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("pinata_api_key", api.Key)
	req.Header.Add("pinata_secret_api_key", api.Secret)

	// execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pinata: fetching: %w", err)
	}
	defer resp.Body.Close()

	// resp.Body = ioutil.NopCloser(io.TeeReader(resp.Body, os.Stderr))
	// FIXME: if response is failed because e.g. missing API keys, return meaningful error instead of empty + nil

	// parse response
	var rows struct {
		Rows []struct {
			Hash     string `json:"ipfs_pin_hash"`
			Size     int64
			Metadata struct {
				Name string
			}
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("pinata: decoding fetched response: %w", err)
	}

	// Prepare filter
	m := map[string]bool{}
	for _, h := range filter {
		m[h] = true
	}
	if len(filter) == 0 {
		m = nil
	}

	// Convert to output format & filter hashes if needed
	list := []pup.NamedHash{}
	for _, row := range rows.Rows {
		if m == nil || m[row.Hash] {
			list = append(list, pup.NamedHash{
				Hash: row.Hash,
				Name: row.Metadata.Name,
				Size: row.Size,
			})
		}
	}
	return list, nil
}

func (api *API) Pin(ctx context.Context, hash pup.Hash) error {
	payload, err := json.Marshal(map[string]string{
		"hashToPin": hash,
	})
	if err != nil {
		// Logic bug, should never happen
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.pinata.cloud/pinning/pinByHash",
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("pinata: adding hash %q: %w", hash, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("pinata_api_key", api.Key)
	req.Header.Add("pinata_secret_api_key", api.Secret)

	// execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("pinata: adding hash %q: %w", hash, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pinata: pin call returned HTTP code %d", resp.StatusCode)
	}

	return nil
}

func (api *API) Unpin(ctx context.Context, hash pup.Hash) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("https://api.pinata.cloud/pinning/unpin/%s", hash),
		nil,
	)
	if err != nil {
		return fmt.Errorf("pinata: removing hash %q: %w", hash, err)
	}

	req.Header.Add("pinata_api_key", api.Key)
	req.Header.Add("pinata_secret_api_key", api.Secret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("pinata: removing hash %q: %w", hash, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pinata: unpin call returned HTTP code %d", resp.StatusCode)
	}

	return nil
}

/*
func (api *API) isPinned(ctx context.Context, hash string) (bool, error) {
	// TODO: use some metadata, otherwise this is very ineffective and currently limited to 1000 pins

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.pinata.cloud/data/pinList?status=pinned",
		nil,
	)
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
*/
