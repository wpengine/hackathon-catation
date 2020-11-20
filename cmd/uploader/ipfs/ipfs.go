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

package ipfs

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	config "github.com/ipfs/go-ipfs-config"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/bootstrap"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/repo"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	ipfspath "github.com/ipfs/interface-go-ipfs-core/path"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

type Node struct {
	node *core.IpfsNode
	API  coreiface.CoreAPI
}

func Start() (*Node, error) {
	// We have to create a repo explicitly to be able to tweak config options
	repo, err := defaultRepo(sync.MutexWrap(datastore.NewMapDatastore()))
	if err != nil {
		return nil, fmt.Errorf("starting ipfs: repo initialization: %w", err)
	}
	// Source: https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#autorelay
	// via: https://discuss.ipfs.io/t/how-to-connect-to-a-node-behind-nat/5270
	repo.C.Swarm.EnableRelayHop = false
	repo.C.Swarm.EnableAutoRelay = true

	node, err := core.NewNode(context.TODO(), &core.BuildCfg{
		Repo:   repo,
		Online: true,
	})
	if err != nil {
		return nil, fmt.Errorf("starting ipfs: creating node: %w", err)
	}

	err = node.Bootstrap(bootstrap.DefaultBootstrapConfig)
	if err != nil {
		return nil, fmt.Errorf("starting ipfs: bootstrapping: %w", err)
	}

	api, err := coreapi.NewCoreAPI(node)
	if err != nil {
		return nil, fmt.Errorf("starting ipfs: accessing API: %w", err)
	}

	return &Node{node: node, API: api}, nil
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

func (n *Node) Close() error {
	return n.node.Close()
}

func (n *Node) AddAndPin(ctx context.Context, file files.Node) (ipfspath.Resolved, error) {
	path, err := n.API.Unixfs().Add(ctx, file)
	if err != nil {
		return path, fmt.Errorf("adding to ipfs: %w", err)
	}
	err = n.API.Pin().Add(ctx, path)
	if err != nil {
		return path, fmt.Errorf("pinning in ipfs: %w", err)
	}
	return path, nil
}

func (n *Node) Provide(ctx context.Context, path ipfspath.Path) error {
	err := n.API.Dht().Provide(ctx, path)
	if err != nil {
		return fmt.Errorf("providing %s to ipfs: %w", path, err)
	}
	return nil
}
