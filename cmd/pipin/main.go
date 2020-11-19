package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// This package is needed so that all the preloaded plugins are loaded automatically

func main() {
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
	r.Use(newAuthMiddleware(*token))

	log.Printf("Starting HTTP API on %s...", *addr)
	log.Fatal(http.ListenAndServe(*addr, r))
}
