package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	config "github.com/ipfs/go-ipfs-config"
	libp2p "github.com/ipfs/go-ipfs/core/node/libp2p"
	iface "github.com/ipfs/interface-go-ipfs-core"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"

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
	r.Use(authMiddleware(*token))

	log.Printf("Starting HTTP API on %s...", *addr)
	log.Fatal(http.ListenAndServe(*addr, r))
}

func authMiddleware(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if "Bearer "+token != r.Header.Get("authorization") {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

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
	fmt.Fprintf(w, "pin create %q", vars["hash"])
}

func (api *API) pinStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Fprintf(w, "pin status %q", vars["hash"])
}

func openRepo(ctx context.Context, path string) (iface.CoreAPI, error) {
	plugins, err := loader.NewPluginLoader(filepath.Join(path, "plugins"))
	if err != nil {
		return nil, fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return nil, fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return nil, fmt.Errorf("error initializing plugins: %s", err)
	}

	cfg, err := config.Init(ioutil.Discard, 2048)
	if err != nil {
		return nil, err
	}

	err = fsrepo.Init(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init node: %s", err)
	}

	repo, err := fsrepo.Open(path)
	if err != nil {
		return nil, err
	}

	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTOption,
		Repo:    repo,
	}

	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, err
	}

	return coreapi.NewCoreAPI(node)
}
