package pup

import "context"

type Hash = string

type NamedHash struct {
	Hash
	Name string // optional; can be filename or path [?]
	Size int64  // optional; in bytes [TODO: 0 or empty?]
}

type Pup interface {
	// Fetch retrieves a list of pinned hashes. If filter is non-empty, the
	// returned list will contain only hashes from the filter list.
	Fetch(ctx context.Context, filter []Hash) ([]NamedHash, error)
	Pin(ctx context.Context, hash Hash) error
	//Unpin(ctx context.Context, hash Hash) error
}
