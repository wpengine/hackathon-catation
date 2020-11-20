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

package bitly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type API struct {
	Key string
}

func (api *API) Shorten(url string) (string, error) {
	payload, err := json.Marshal(map[string]string{
		"long_url": url,
	})
	if err != nil {
		// Logic bug, should never happen
		panic(err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api-ssl.bitly.com/v4/shorten",
		bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("bitly: shortening %q: %w", url, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+api.Key)

	// Send request
	// TODO: add timeout (ideally, use context.Context)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("bitly: shortening %q: %w", url, err)
	}
	defer resp.Body.Close()

	// Parse respone
	// FIXME: check for error codes
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("bitly: shortening %q: reading response: %w", url, err)
	}
	var r struct {
		Link string `json:"link"`
	}
	err = json.Unmarshal(responseData, &r)
	if err != nil {
		return "", fmt.Errorf("bitly: shortening %q: parsing response: %w", url, err)
	}
	return r.Link, nil
}
