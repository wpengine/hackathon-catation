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

package eternum

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wpengine/hackathon-catation/pup"
)

type Client struct {
	Key string
}

func New(key string) *Client {
	return &Client{Key: key}
}

func (c *Client) Fetch(ctx context.Context, filter []pup.Hash) ([]pup.NamedHash, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://www.eternum.io/api/pin/",
		nil,
	)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Token %s", c.Key))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to fetch pins: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var body listresponse

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	var m map[string]bool = nil

	if len(filter) > 0 {
		m = make(map[string]bool)
		for _, h := range filter {
			m[h] = true
		}
	}

	list := []pup.NamedHash{}
	for _, obj := range body.Results {
		if m == nil {
			list = append(list, pup.NamedHash{
				Hash: obj.Hash,
				Name: obj.Name,
				Size: obj.Size,
			})
			continue
		}
		if _, ok := m[obj.Hash]; ok {
			list = append(list, pup.NamedHash{
				Hash: obj.Hash,
				Name: obj.Name,
				Size: obj.Size,
			})
		}
	}
	return list, nil
}

type pin struct {
	Hash   string `json:"hash"`
	Active bool   `json:"active"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
}

type listresponse struct {
	Results []pin `json:"results"`
}

func (c *Client) Pin(ctx context.Context, hash pup.Hash) error {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(map[string]string{"hash": hash})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://www.eternum.io/api/pin/",
		&buf,
	)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Token %s", c.Key))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	if resp.StatusCode == http.StatusBadRequest {
		defer resp.Body.Close()

		errJSON := struct {
			NonFieldErrors []string `json:"non_field_errors"`
		}{}

		err = json.NewDecoder(resp.Body).Decode(&errJSON)
		if err != nil {
			return err
		}

		// yuck
		if errJSON.NonFieldErrors[0] == "You have already pinned an object with that hash." {
			return nil
		}
	}

	return fmt.Errorf("unable to pin hash: %d", resp.StatusCode)
}

func (c *Client) Unpin(ctx context.Context, hash pup.Hash) error {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(map[string]string{"hash": hash})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("https://www.eternum.io/api/pin/%s/", hash),
		nil,
	)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Token %s", c.Key))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return fmt.Errorf("unable to unpin hash: %d", resp.StatusCode)
}
