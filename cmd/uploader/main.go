package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ipfs/go-datastore"
	config "github.com/ipfs/go-ipfs-config"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/bootstrap"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/repo"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

func main() {
	// Open the file that we want to add to IPFS
	fn := os.Args[1]
	fh, err := os.Open(fn)
	if err != nil {
		die(err)
	}
	defer fh.Close()

	// Upload the file to IPFS...

	// We have to create a repo explicitly to be able to tweak config options
	repo, err := defaultRepo(datastore.NewMapDatastore())
	if err != nil {
		die(err)
	}
	// Source: https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#autorelay
	// via: https://discuss.ipfs.io/t/how-to-connect-to-a-node-behind-nat/5270
	repo.C.Swarm.EnableRelayHop = false
	repo.C.Swarm.EnableAutoRelay = true

	// TODO: where do IPFS-internal temporary files get created/saved?
	node, err := core.NewNode(context.TODO(), &core.BuildCfg{
		// NilRepo: true,  // ?
		Repo:   repo,
		Online: true,
	})
	if err != nil {
		die(err)
	}
	defer node.Close()

	// FIXME: do we need this?
	// WIP: trying to resolve NAT issues
	err = node.Bootstrap(bootstrap.DefaultBootstrapConfig)
	if err != nil {
		die(err)
	}

	api, err := coreapi.NewCoreAPI(node)
	if err != nil {
		die(err)
	}
	stat, err := fh.Stat()
	if err != nil {
		die(err)
	}
	path, err := api.Unixfs().Add(context.TODO(), files.NewReaderStatFile(fh, stat))
	if err != nil {
		die(err)
	}
	fmt.Println(path)

	// try to make sure the file is pinned and visible
	log.Println("pinning...")
	err = api.Pin().Add(context.TODO(), path)
	if err != nil {
		panic(err)
	}
	log.Println("providing...")
	err = api.Dht().Provide(context.TODO(), path)
	if err != nil {
		panic(err)
	}

	log.Println("finding providers...")
	providersChan, err := api.Dht().FindProviders(context.TODO(), path)
	if err != nil {
		die(err)
	}
	for p := range providersChan {
		fmt.Println(p)
	}

	os.Stderr.WriteString("Press enter to continue: ")
	os.Stdin.Read([]byte("tmp"))

	r, err := api.Object().Data(context.TODO(), path)
	if err != nil {
		die(err)
	}
	io.Copy(os.Stdout, r)
}

// FIXME: copied from: https://github.com/ipfs/go-ipfs/blob/5b28704e505eb9a65c1ef8d2336da95af8e828c8/core/node/builder.go#L125-L151
func defaultRepo(dstore repo.Datastore) (*repo.Mock, error) {
	c := config.Config{}
	priv, pub, err := p2pcrypto.GenerateKeyPairWithReader(p2pcrypto.RSA, 2048, rand.Reader)
	if err != nil {
		return nil, err
	}

	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return nil, err
	}

	privkeyb, err := priv.Bytes()
	if err != nil {
		return nil, err
	}

	c.Bootstrap = config.DefaultBootstrapAddresses
	c.Addresses.Swarm = []string{"/ip4/0.0.0.0/tcp/4001", "/ip4/0.0.0.0/udp/4001/quic"}
	c.Identity.PeerID = pid.Pretty()
	c.Identity.PrivKey = base64.StdEncoding.EncodeToString(privkeyb)

	return &repo.Mock{
		D: dstore,
		C: c,
	}, nil
}

func die(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, "error:", fmt.Sprint(msg...))
	os.Exit(1)
}
