package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	bitlyShortenEndpoint = "https://api-ssl.bitly.com/v4/shorten"
)

type (
	bitlyHashResponse struct {
		CreatedAt      string        `json:"created_at"`
		ID             string        `json:"id"`
		Link           string        `json:"link"`
		CustomBitlinks []interface{} `json:"custom_bitlinks"`
		LongURL        string        `json:"long_url"`
		Archived       bool          `json:"archived"`
		Tags           []interface{} `json:"tags"`
		Deeplinks      []interface{} `json:"deeplinks"`
		References     struct {
			Group string `json:"group"`
		} `json:"references"`
	}
)

func main() {
	hash := os.Args[1]

	jsonPayload, err := json.Marshal(map[string]string{
		"long_url": hashToURL(hash),
	})

	if err != nil {
		die(err)
	}

	req, err := http.NewRequest(http.MethodPost, bitlyShortenEndpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		die(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", ("Bearer " + os.Getenv("BITLY_API_KEY")))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		die(err)
	}
	defer resp.Body.Close()

	// Parse respone
	// TODO - check for error codes
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		die(err)
	}
	var responseObject bitlyHashResponse
	json.Unmarshal(responseData, &responseObject)
	fmt.Println(responseObject.Link)
}

func hashToURL(hash string) string {
	return "http://ipfs.io/ipfs/" + hash
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
