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

package temporal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

/*

authorization: 'Bearer ' + jwt

https://api.ipfs.temporal.cloud/api/v0/login

https://api.ipfs.temporal.cloud/api/v0/pin/add




POST https://api.temporal.cloud/v2/ipfs/public/pin/:hash

hash       IPFS Hash The specific hash to pin.
hold_time  Int       Number of months to pin the hash.
file_name  String    optional filename to name the pin with.


POST https://api.temporal.cloud/v2/auth/login

username	String	The username.
password	String	The associated password.

{
  "expire": "2018-12-21T19:31:42Z",
  "token": "eyJhbG ... "
}
*/

type Client struct {
	token string
}

/*
	bytes.NewBufferString(
		url.Values{
			"username": {username},
			"password": {password},
		}.Encode(),
*/
func New(ctx context.Context, username, password string) (*Client, error) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(map[string]string{"username": username, "password": password}); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.temporal.cloud/v2/auth/login",
		&buf,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("cache-control", "no-store,no-cache,private")
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		// todo: better error handling here
		return nil, fmt.Errorf("could not get token, %s (%d)", resp.Status, resp.StatusCode)
	}
	defer resp.Body.Close()

	var login loginResponse

	err = json.NewDecoder(resp.Body).Decode(&login)
	if err != nil {
		return nil, err
	}

	log.Printf("login worked! %v", login.Expire)
	// todo: track token expiry and refresh it
	return &Client{token: login.Token}, nil
}

type loginResponse struct {
	Expire string `json:"expire"`
	Token  string `json:"token"`
}

func (c *Client) Pin(ctx context.Context, cid string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("https://api.temporal.cloud/v2/ipfs/public/pin/%s", cid),
		bytes.NewBufferString(url.Values{"hold_time": {"6"}}.Encode()), // hold pin for 6 months by default
	)
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "Bearer "+c.token)
	req.Header.Set("cache-control", "no-store,no-cache,private")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		// todo: better error handling here
		return fmt.Errorf("could not pin hash: %s (%d)", resp.Status, resp.StatusCode)
	}

	return nil
}
