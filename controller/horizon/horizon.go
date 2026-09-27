// Package horizon owns authority: what a frame runs, what it may reach, and
// the OS sandbox that authority renders to. It imports only package contract.
package horizon

import (
	"fmt"
	"slices"
	"strings"

	"skyra/controller/contract"
)

// Surface is a registered executable: what may run, pinned by content hash,
// with the contracts the runtime checks its input and its reply against.
type Surface struct {
	Path   string           `json:"path"`
	SHA256 string           `json:"sha256"`
	Input  *contract.Schema `json:"input"`           // what invoking it must fit
	Output *contract.Schema `json:"output"`          // what its reply must fit: the reply capability's input contract
	Reach  []string         `json:"reach,omitempty"` // what it may be granted, if its parent holds it. Empty: a leaf
	// Connect lists the unix sockets it may connect to: an adapter reaching
	// its actor host (next/15). The operator's registry grants them.
	Connect []string `json:"connect,omitempty"`
}

// Leaf reports whether the surface can only reply.
func (s Surface) Leaf() bool { return len(s.Reach) == 0 }

// Horizon is everything a step knows besides the bytes. It is the release
// (D6): the child's authority, sent by the parent after it verifies the child's
// commit. The data travels as JSON. The frame stack travels beside it as
// directory fds over SCM_RIGHTS, and those fds are the capability.
//
// Reach is the registry as this frame sees it: the surfaces its emissions may
// invoke. There is no global registry lookup. Genesis holds the root reach,
// and every child's reach is a subset of its parent's.
//
// Every Horizon also holds reply: the right to discharge Frame's slot, once.
// Resume is the continuation waiting on that slot, as it was when it emitted:
// its Stamp is the round that opened the slot. The operator's Horizon has no
// Resume.
type Horizon struct {
	Stamp   string             `json:"stamp"`             // this frame's latest commit; the predecessor of whatever is emitted under it
	Surface *Surface           `json:"surface,omitempty"` // what this frame runs; nil for the operator
	Reach   map[string]Surface `json:"reach"`             // what emissions under this Horizon may invoke
	Frame   string             `json:"frame"`             // the frame whose slot this Horizon's reply discharges
	Resume  *Horizon           `json:"resume,omitempty"`  // the continuation waiting on Frame's slot
}

// Depth is the number of frames in the stack that travels with this Horizon.
func (h Horizon) Depth() int {
	n := 1
	for r := h.Resume; r != nil; r = r.Resume {
		n++
	}
	return n
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
	for _, sock := range h.Surface.Connect {
		fmt.Fprintf(&b, "(allow network-outbound (remote unix-socket (path-literal %q)))\n", sock)
	}
	fmt.Fprintf(&b, "(deny file-read* file-write* (subpath %q))\n", floor)
	return b.String()
}
