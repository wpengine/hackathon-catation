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

package main

import (
	"context"
	"encoding/json"
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

	if err = json.NewEncoder(w).Encode(pins); err != nil {
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
		return
	}

	err = api.ipfs.Pin().Add(context.Background(), testCID)
	if err != nil {
		log.Printf("Could not pin file with CID: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode([]string{vars["hash"]}); err != nil {
		log.Printf("error encoding to json: %v", err)
	}
}

func (api *API) pinStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, pinned, err := api.ipfs.Pin().IsPinned(context.Background(), icorepath.New(vars["hash"]))
	if err != nil {
		log.Printf("Could not check pin status: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	if err = json.NewEncoder(w).Encode(map[string]interface{}{"pinned": pinned}); err != nil {
		log.Printf("error encoding to json: %v", err)
	}
}

func (api *API) pinRemoveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := api.ipfs.Pin().Rm(context.Background(), icorepath.New(vars["hash"]))
	if err != nil {
		log.Printf("Could not delete pin: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	if _, err = w.Write([]byte(`{"pinned": false}\n`)); err != nil {
		log.Printf("error writing response: %v", err)
	}
}
