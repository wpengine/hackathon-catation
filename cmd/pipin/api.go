package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	iface "github.com/ipfs/interface-go-ipfs-core"
	icorepath "github.com/ipfs/interface-go-ipfs-core/path"

	"github.com/gorilla/mux"
)

// API is the public interface over HTTP
type API struct {
	ipfs iface.CoreAPI
}

type pinResponse struct {
	Hash string `json:"hash"`
	Path string `json:"path"`
}

func (api *API) pinListHandler(w http.ResponseWriter, r *http.Request) {
	pinchan, err := api.ipfs.Pin().Ls(context.Background())
	if err != nil {
		log.Printf("error fetching pins: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pins := []string{}
	for pin := range pinchan {
		pins = append(pins, pin.Path().Cid().String())
	}

	err = json.NewEncoder(w).Encode(pins)
	if err != nil {
		log.Printf("error encoding to json: %v", err)
	}
}

func (api *API) pinCreateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	testCID := icorepath.New(vars["hash"])

	_, err := api.ipfs.Unixfs().Get(context.Background(), testCID)
	if err != nil {
		log.Printf("Could not get file with CID: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	err = api.ipfs.Pin().Add(context.Background(), testCID)
	if err != nil {
		log.Printf("Could not pin file with CID: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode([]string{vars["hash"]})
}

func (api *API) pinStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Fprintf(w, "pin status %q", vars["hash"])
}
