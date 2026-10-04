// Package checkpoint owns the user adapter's recovery slot. The slot contains
// only the invocation anchor, its contract and the accepted frontier. It is
// not a transcript, phase store or provider mapping database.
package checkpoint

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"syscall"
)

type Contract struct {
	Target string          `json:"target"`
	Schema json.RawMessage `json:"schema"`
}

type Checkpoint struct {
	RootInvocation string   `json:"root_invocation"`
	Leaves         []string `json:"leaves"`
	Contract       Contract `json:"contract"`
}

// Evidence is the versioned result delivered by system.strand. The adapter
// accepts this only as a returned syscall step, never from human input.
type Evidence struct {
	Version  int               `json:"version"`
	Root     string            `json:"root"`
	Frontier []string          `json:"frontier"`
	Records  map[string]Record `json:"records"`
	Opaque   map[string]string `json:"opaque"`
	Covered  map[string]string `json:"covered"`
}

type Record struct {
	Predecessor string   `json:"predecessor"`
	Deps        []string `json:"deps"`
}

var ErrConflict = errors.New("checkpoint changed; reconstruct the combined frontier before retrying")

type Store struct{ Dir string }

func validStamp(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func validID(id string) bool {
	return id != "" && filepath.Base(id) == id && id != "." && id != ".." && id[0] != '.'
}
func (s Store) path(id string) (string, error) {
	if !validID(id) {
		return "", fmt.Errorf("invalid conversation reference")
	}
	return filepath.Join(s.Dir, id+".json"), nil
}

func (c Checkpoint) Key() string {
	b, _ := json.Marshal(c)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func (c *Checkpoint) validate() error {
	if !validStamp(c.RootInvocation) || len(c.Leaves) == 0 || c.Contract.Target == "" || !json.Valid(c.Contract.Schema) {
		return fmt.Errorf("invalid checkpoint")
	}
	for _, leaf := range c.Leaves {
		if !validStamp(leaf) {
			return fmt.Errorf("invalid checkpoint leaf")
		}
	}
	slices.Sort(c.Leaves)
	c.Leaves = slices.Compact(c.Leaves)
	return nil
}
func (s Store) Read(id string) (Checkpoint, error) {
	p, e := s.path(id)
	if e != nil {
		return Checkpoint{}, e
	}
	return read(p)
}
func read(p string) (Checkpoint, error) {
	var c Checkpoint
	b, e := os.ReadFile(p)
	if e != nil {
		return c, e
	}
	if e = json.Unmarshal(b, &c); e != nil {
		return c, e
	}
	return c, c.validate()
}

// Initialize is exclusive. A repeated initialization must identify the same
// anchor and contract, and never resets an already advanced frontier.
func (s Store) Initialize(id string, root string, contract Contract) (Checkpoint, error) {
	initial := Checkpoint{RootInvocation: root, Leaves: []string{root}, Contract: contract}
	if e := initial.validate(); e != nil {
		return Checkpoint{}, e
	}
	return s.locked(id, func(p string) (Checkpoint, error) {
		old, e := read(p)
		if e == nil {
			if old.RootInvocation != root || !sameContract(old.Contract, contract) {
				return old, ErrConflict
			}
			return old, nil
		}
		if !os.IsNotExist(e) {
			return old, e
		}
		return initial, publish(p, initial)
	})
}
func sameContract(a, b Contract) bool {
	var av, bv any
	if json.Unmarshal(a.Schema, &av) != nil || json.Unmarshal(b.Schema, &bv) != nil {
		return false
	}
	x, _ := json.Marshal(av)
	y, _ := json.Marshal(bv)
	return a.Target == b.Target && string(x) == string(y)
}

// Advance performs compare-and-update under a process-safe lock. A stale
// child cannot restore a leaf removed by a winning join. Conflict requires a
// fresh syscall over the union; it never silently drops concurrent progress.
func (s Store) Advance(id string, expected Checkpoint, evidence Evidence) (Checkpoint, error) {
	if e := expected.validate(); e != nil {
		return Checkpoint{}, e
	}
	return s.locked(id, func(p string) (Checkpoint, error) {
		current, e := read(p)
		if e != nil {
			return current, e
		}
		if current.Key() != expected.Key() {
			return current, ErrConflict
		}
		if evidence.Version != 1 || evidence.Root != current.RootInvocation || len(evidence.Frontier) == 0 {
			return current, fmt.Errorf("invalid reconstruction evidence")
		}
		if _, ok := evidence.Records[current.RootInvocation]; !ok {
			return current, fmt.Errorf("evidence omits the anchor")
		}
		for _, leaf := range append(slices.Clone(current.Leaves), evidence.Frontier...) {
			if !validStamp(leaf) {
				return current, fmt.Errorf("invalid evidence leaf")
			}
			if _, ok := evidence.Records[leaf]; !ok {
				if _, ok := evidence.Opaque[leaf]; !ok {
					return current, fmt.Errorf("evidence omits a retained leaf")
				}
			}
		}
		// Even malformed service output must not cross the retained root or
		// create cycles. system.strand checks runtime acceptance; the slot
		// independently checks the structural evidence it consumes.
		state := map[string]int{}
		var validate func(string) error
		validate = func(stamp string) error {
			if stamp == current.RootInvocation {
				return nil
			}
			if state[stamp] == 1 {
				return fmt.Errorf("cyclic reconstruction evidence")
			}
			if state[stamp] == 2 {
				return nil
			}
			state[stamp] = 1
			if boundary, ok := evidence.Opaque[stamp]; ok {
				if !validStamp(boundary) || boundary == stamp {
					return fmt.Errorf("invalid opacity boundary")
				}
				if err := validate(boundary); err != nil {
					return err
				}
			} else {
				r, ok := evidence.Records[stamp]
				if !ok || !validStamp(r.Predecessor) {
					return fmt.Errorf("evidence leaves the invocation")
				}
				if err := validate(r.Predecessor); err != nil {
					return err
				}
				for _, dep := range r.Deps {
					if err := validate(dep); err != nil {
						return err
					}
				}
			}
			state[stamp] = 2
			return nil
		}
		for _, leaf := range append(slices.Clone(current.Leaves), evidence.Frontier...) {
			if err := validate(leaf); err != nil {
				return current, err
			}
		}
		leaves := append(slices.Clone(current.Leaves), evidence.Frontier...)
		slices.Sort(leaves)
		leaves = slices.Compact(leaves)
		covered := map[string]bool{}
		for old, newer := range evidence.Covered {
			if !validStamp(old) || !validStamp(newer) || old == newer {
				return current, fmt.Errorf("invalid frontier coverage")
			}
			if err := validate(newer); err != nil {
				return current, err
			}
			covered[old] = true
		}
		var cover func(string, map[string]bool)
		cover = func(stamp string, seen map[string]bool) {
			if seen[stamp] {
				return
			}
			seen[stamp] = true
			if boundary, ok := evidence.Opaque[stamp]; ok {
				covered[boundary] = true
				cover(boundary, seen)
				return
			}
			r, ok := evidence.Records[stamp]
			if !ok {
				return
			}
			for _, parent := range append([]string{r.Predecessor}, r.Deps...) {
				if parent == "" {
					continue
				}
				covered[parent] = true
				cover(parent, seen)
			}
		}
		for _, leaf := range leaves {
			cover(leaf, map[string]bool{})
		}
		kept := []string{}
		for _, leaf := range leaves {
			if !covered[leaf] {
				kept = append(kept, leaf)
			}
		}
		if len(kept) == 0 {
			return current, fmt.Errorf("cyclic evidence frontier")
		}
		current.Leaves = kept
		return current, publish(p, current)
	})
}

func (s Store) locked(id string, fn func(string) (Checkpoint, error)) (Checkpoint, error) {
	p, e := s.path(id)
	if e != nil {
		return Checkpoint{}, e
	}
	if e = os.MkdirAll(s.Dir, 0700); e != nil {
		return Checkpoint{}, e
	}
	f, e := os.OpenFile(p+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return Checkpoint{}, e
	}
	defer f.Close()
	if e = syscall.Flock(int(f.Fd()), syscall.LOCK_EX); e != nil {
		return Checkpoint{}, e
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn(p)
}
func publish(p string, c Checkpoint) error {
	b, e := json.Marshal(c)
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(p), ".checkpoint-")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = os.Rename(f.Name(), p); e != nil {
		return e
	}
	dir, e := os.Open(filepath.Dir(p))
	if e != nil {
		return e
	}
	defer dir.Close()
	return dir.Sync()
}
