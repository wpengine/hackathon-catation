package pinata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wpengine/hackathon-catation/pup"
)

// FIXME: unify with cmd/pinner/pinata/
// FIXME[LATER]: unify with cmd/uploader/pinata/ ?

type API struct {
	Key, Secret string
}

func (api *API) Fetch(filter []pup.Hash) ([]pup.NamedHash, error) {
	// TODO: use some metadata, otherwise this func is very ineffective and currently limited to 1000 pins (TODO: first, check if they didn't publish some newer API)

	req, err := http.NewRequest(http.MethodGet,
		"https://api.pinata.cloud/data/pinList?status=pinned", nil)
	if err != nil {
		// Logic bug, should never happen
		panic(fmt.Errorf("pinata: building fetch request: %w", err))
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("pinata_api_key", api.Key)
	req.Header.Add("pinata_secret_api_key", api.Secret)

	// execute the request
	// TODO: [LATER] configurable timeout - or rather, pass Context as Fetch argument
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
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
