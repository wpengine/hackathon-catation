package pipin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/wpengine/hackathon-catation/pup"
)

type Client struct {
	UseTLS bool
	Host   string // TODO(akavel): replace this + above with BaseURL ?
	Token  string
}

func New(useTLS bool, host, token string) *Client {
	return &Client{
		UseTLS: useTLS,
		Host:   host,
		Token:  token,
	}
}

func (c *Client) endpoint(path string, args ...interface{}) *url.URL {
	scheme := "http"
	if c.UseTLS {
		scheme = "https"
	}

	return &url.URL{
		Scheme: scheme,
		Host:   c.Host,
		Path:   fmt.Sprintf(path, args...),
	}
}

func (c *Client) Fetch(ctx context.Context, filter []pup.Hash) ([]pup.NamedHash, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.endpoint("pins").String(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", c.Token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// todo: better errors
		return nil, fmt.Errorf("unable to list pins")
	}
	defer resp.Body.Close()

	hashes := []string{}

	err = json.NewDecoder(resp.Body).Decode(&hashes)
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
	for _, hash := range hashes {
		if m == nil {
			list = append(list, pup.NamedHash{Hash: hash})
			continue
		}
		if _, ok := m[hash]; ok {
			list = append(list, pup.NamedHash{Hash: hash})
		}
	}

	return list, nil
}

func (c *Client) Pin(ctx context.Context, hash pup.Hash) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint("pin/%s", hash).String(),
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", c.Token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		// todo: better errors
		return fmt.Errorf("unable to pin hash")
	}

	return nil
}

func (c *Client) Unpin(ctx context.Context, hash pup.Hash) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.endpoint("pin/%s", hash).String(),
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", c.Token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		// todo: better errors
		return fmt.Errorf("unable to unpin hash")
	}

	return nil
}
