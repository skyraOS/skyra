package boundary

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"

	"skyra/controller/core/contract"
)

// CList is a Boundary's capabilities by contract (48): for each actor it
// names, the contracts an invocation under the Boundary may select. A
// contract is (actor, schema), so holding an actor is not holding all of it.
type CList map[string][]*contract.Schema

// Holds reports whether an invocation under h may select contract c of the
// actor named name. Selection runs first against everything the actor
// declares, so a restriction never makes an ambiguous payload valid; this is
// the check after it.
func (h Boundary) Holds(name string, c *contract.Schema) bool {
	s, ok := h.Reach[name]
	if !ok {
		return false
	}
	return slices.ContainsFunc(h.held(name, s), func(held *contract.Schema) bool { return contract.Same(held, c) })
}

// HeldOf is every contract of name an invocation under h may select.
func (h Boundary) HeldOf(name string) []*contract.Schema {
	s, ok := h.Reach[name]
	if !ok {
		return nil
	}
	return h.held(name, s)
}

func (h Boundary) held(name string, s Surface) []*contract.Schema {
	if h.Held == nil {
		return s.All()
	}
	return (*h.Held)[name]
}

// Effective is h's c-list written out: what Held names of what is in reach,
// or, for a Boundary never narrowed by contract, every contract of it.
func (h Boundary) Effective() CList {
	out := CList{}
	for name, s := range h.Reach {
		if held := h.held(name, s); len(held) > 0 {
			out[name] = held
		}
	}
	return out
}

// Hash names a Boundary by what it holds (48): one line per held contract,
// "contract\tactor\tname\n", and one per actor in reach, "reach\tactor\tname\n"
// (what it declares decides which contract a payload selects), sorted
// bytewise, then its authority, which still travels on the Boundary. Equal
// capabilities, equal Boundary; a Boundary is never updated, so an export or
// a narrowing makes another one.
func (h Boundary) Hash() string {
	var lines []string
	for name, s := range h.Reach {
		lines = append(lines, "reach\t"+Identity(s)+"\t"+name+"\n")
	}
	for name, held := range h.Effective() {
		actor := Identity(h.Reach[name])
		for _, c := range held {
			b, _ := json.Marshal(contract.Normal(c))
			sum := sha256.Sum256(b)
			lines = append(lines, hex.EncodeToString(sum[:])+"\t"+actor+"\t"+name+"\n")
		}
	}
	slices.Sort(lines)
	lines = slices.Compact(lines)
	authority, _ := json.Marshal(h.Authority)
	sum := sha256.Sum256([]byte(strings.Join(lines, "") + string(authority)))
	return hex.EncodeToString(sum[:])
}
