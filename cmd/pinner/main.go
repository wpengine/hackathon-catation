package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	////////

	tmp1, err := http.NewRequest(http.MethodGet, "https://api.pinata.cloud/data/pinList", nil)
	if err != nil {
		panic(err)
	}
	tmp1.Header.Add("Content-Type", "application/json")
	tmp1.Header.Add("pinata_api_key", api.Key)
	tmp1.Header.Add("pinata_secret_api_key", api.Secret)
	// execute the request
	c := &http.Client{Timeout: 10 * time.Second}
	tmp2, err := c.Do(tmp1)
	if err != nil {
		panic(err)
	}
	defer tmp2.Body.Close()
	tmp3, err := ioutil.ReadAll(tmp2.Body)
	if err != nil {
		panic(err)
	}
	tmp5 := map[string]interface{}{}
	err = json.Unmarshal(tmp3, &tmp5)
	if err != nil {
		panic(err)
	}
	tmp4, err := json.MarshalIndent(tmp5, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(tmp4))
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
