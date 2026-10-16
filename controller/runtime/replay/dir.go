//go:build !js

package replay

import (
	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

// dir is the tree on disk as a Store.
type dir string

func (d dir) Get(kind, hash string) ([]byte, error) { return tree.Get(string(d), kind, hash) }

// Replay walks from head back to genesis, then forward, and returns the
// number of records re-derived: the chain, and every sibling a resolve names.
// Then it checks the whole tree: no round was resumed twice (ResumedOnce).
func Replay(root, head string) (int, error) {
	n, err := Lineage(dir(root), head)
	if err != nil {
		return n, err
	}
	stamps, err := tree.Stamps(root)
	if err != nil {
		return n, err
	}
	records := make(map[string]record.Record, len(stamps))
	for _, s := range stamps {
		if records[s], err = tree.Load(root, s); err != nil {
			return n, err
		}
	}
	return n, ResumedOnce(records)
}

// HorizonAt derives, from the tree alone, the Horizon the record at stamp
// released: the one whatever it runs, or whatever is emitted under it, runs
// under. Every record on the way is re-derived, as Replay does. Authority is
// never taken from anyone who holds the work (next/24, next/20 finding 1).
func HorizonAt(root, stamp string) (horizon.Horizon, error) {
	w := walker{s: dir(root), seen: map[string]horizon.Horizon{}}
	return w.at(stamp)
}

// NewWalker is a Walker over a tree.
func NewWalker(root string) *Walker {
	return &Walker{walker{s: dir(root), seen: map[string]horizon.Horizon{}}}
}
