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
	Unpin(ctx context.Context, hash Hash) error
}
