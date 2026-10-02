//go:build !js

package replay

import (
	"skyra/controller/core/horizon"
	"skyra/controller/runtime/tree"
)

// dir is the tree on disk as a Store.
type dir string

func (d dir) Get(kind, hash string) ([]byte, error) { return tree.Get(string(d), kind, hash) }

// Replay walks from head back to genesis, then forward, and returns the
// number of records in the chain.
func Replay(root, head string) (int, error) {
	return Lineage(dir(root), head)
}

// HorizonAt derives, from the tree alone, the Horizon the record at stamp
// released: the one whatever it runs, or whatever is emitted under it, runs
// under. Every record on the way is re-derived, as Replay does. Authority is
// never taken from anyone who holds the work (next/24, next/20 finding 1).
func HorizonAt(root, stamp string) (horizon.Horizon, error) {
	h, _, err := walk(dir(root), stamp)
	return h, err
}
