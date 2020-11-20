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
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/wpengine/hackathon-catation/internal"
)

// This package is needed so that all the preloaded plugins are loaded automatically

func main() {
	internal.PrintGPLBanner("pipin", "2020")

	var (
		token    *string = flag.String("token", "change-me", "HTTP auth token")
		repoPath *string = flag.String("repo", "./repo", "IPFS repository path")
		addr     *string = flag.String("addr", ":9229", "Address to bind HTTP API on")
	)
	flag.Parse()

	ctx := context.Background()

	log.Println("Starting IPFS node...")

	ipfs, err := openRepo(ctx, *repoPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	api := &API{ipfs}

	r := mux.NewRouter()
	r.HandleFunc("/pins", api.pinListHandler).Methods("GET")
	r.HandleFunc("/pin/{hash}", api.pinCreateHandler).Methods("POST")
	r.HandleFunc("/pin/{hash}", api.pinStatusHandler).Methods("GET")
	r.HandleFunc("/pin/{hash}", api.pinRemoveHandler).Methods("DELETE")
	r.Use(newAuthMiddleware(*token))

	log.Printf("Starting HTTP API on %s...", *addr)
	log.Fatal(http.ListenAndServe(*addr, r))
}
