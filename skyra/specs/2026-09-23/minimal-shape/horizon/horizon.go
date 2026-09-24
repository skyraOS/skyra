// Package horizon owns authority: what a frame runs, what it may reach, and
// the OS sandbox that authority renders to. It imports nothing from this module.
package horizon

import (
	"fmt"
	"slices"
	"strings"
)

// Surface is a registered executable: what may run, pinned by content hash.
type Surface struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Leaf   bool   `json:"leaf"`
}

// Horizon is everything a step knows besides the bytes. It is the release
// (D6): the child's authority, sent by the parent after it verifies the child's
// commit. The data travels as JSON. The frame's node travels beside it as an
// fd over SCM_RIGHTS, and that fd is the capability.
//
// Reach is the registry as this frame sees it: the surfaces its emissions may
// invoke. There is no global registry lookup. Genesis holds the root reach,
// and every child's reach is a subset of its parent's.
type Horizon struct {
	Stamp   string             `json:"stamp"`             // this frame's commit; the predecessor of whatever is emitted under it
	Surface *Surface           `json:"surface,omitempty"` // what this frame runs
	Reach   map[string]Surface `json:"reach"`             // what emissions under this Horizon may invoke
	Settles string             `json:"settles,omitempty"` // set when the frame settles a parent frame
}

// Names lists what the Horizon reaches, sorted.
func (h Horizon) Names() []string {
	var n []string
	for k := range h.Reach {
		n = append(n, k)
	}
	slices.Sort(n)
	return n
}

// Render turns a Horizon into a Seatbelt profile (D7). Everything is denied
// unless listed, and floor (the .skyra/ directory) is always denied, whatever
// the Horizon says. The floor is a parameter, not a Horizon field, so no
// Horizon can grant it.
func Render(h Horizon, floor string) string {
	var b strings.Builder
	b.WriteString("(version 1)\n(deny default)\n")
	fmt.Fprintf(&b, "(allow process-exec (literal %q))\n", h.Surface.Path)
	b.WriteString("(allow file-read*)\n(allow sysctl-read)\n(allow mach-lookup)\n")
	fmt.Fprintf(&b, "(deny file-read* file-write* (subpath %q))\n", floor)
	return b.String()
}
