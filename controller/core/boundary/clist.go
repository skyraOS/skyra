package boundary

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"

	"skyra/controller/core/contract"
)

// CList is contracts by actor (48): for each actor it names, the contracts a
// narrowing keeps. A contract is (actor, schema), so keeping an actor is not
// keeping all of it.
type CList map[string][]*contract.Schema

// Hash names a Boundary by what it reaches (48): one line per actor in reach,
// "reach\tprogram\tname\n", and one per contract of it,
// "contract\tprogram\tname\n", sorted bytewise, then its authority, which
// still travels on the Boundary. Equal capabilities, equal Boundary; a
// Boundary is never updated, so a proposal or a narrowing makes another one.
func (h Boundary) Hash() string {
	var lines []string
	for name, s := range h.Reach {
		program := Program(s)
		lines = append(lines, "reach\t"+program+"\t"+name+"\n")
		for _, c := range s.All() {
			b, _ := json.Marshal(contract.Normal(c))
			sum := sha256.Sum256(b)
			lines = append(lines, hex.EncodeToString(sum[:])+"\t"+program+"\t"+name+"\n")
		}
	}
	slices.Sort(lines)
	lines = slices.Compact(lines)
	authority, _ := json.Marshal(h.Authority)
	sum := sha256.Sum256([]byte(strings.Join(lines, "") + string(authority)))
	return hex.EncodeToString(sum[:])
}

// Keep is s with only the contracts in kept: what a narrowing leaves of it,
// the registered one first if it is kept.
func Keep(s Surface, kept []*contract.Schema) Surface {
	in := func(c *contract.Schema) bool {
		return slices.ContainsFunc(kept, func(k *contract.Schema) bool { return contract.Same(k, c) })
	}
	if s.Input != nil && !in(s.Input) {
		s.Input = nil
	}
	var rest []*contract.Schema
	for _, c := range s.Contracts {
		if in(c) {
			rest = append(rest, c)
		}
	}
	s.Contracts = rest
	return s
}

// Declares reports whether s may be invoked through contract c.
func (s Surface) Declares(c *contract.Schema) bool {
	return slices.ContainsFunc(s.All(), func(d *contract.Schema) bool { return contract.Same(d, c) })
}
