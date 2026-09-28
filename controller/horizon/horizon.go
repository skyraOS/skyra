// Package horizon owns authority: what a frame runs, what it may reach, and
// the OS sandbox that authority renders to. It imports only package contract.
package horizon

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"skyra/controller/contract"
)

// PathGrant names one OS path. Literal grants only that path; subpath grants
// the path and its descendants. Registry loading resolves paths before they
// enter a Horizon, so Render never consults the filesystem.
type PathGrant struct {
	Path  string `json:"path"`
	Scope string `json:"scope"` // literal or subpath
}

// Resource describes an inherited descriptor that travels beside a Horizon's
// frame stack. It is transport metadata, not authority to invoke a surface.
type Resource struct {
	Kind        string `json:"kind"`
	Owner       string `json:"owner"`
	Incarnation string `json:"incarnation"`
}

// Authority is the OS authority a surface can exercise directly. Reach is
// separate: a continuation can delegate a registered tool's authority without
// gaining that tool's file or process permissions itself.
type Authority struct {
	Exec    []PathGrant `json:"exec,omitempty"`
	Read    []PathGrant `json:"read,omitempty"`
	Write   []PathGrant `json:"write,omitempty"`
	Mach    []string    `json:"mach,omitempty"`
	Connect []string    `json:"connect,omitempty"` // outbound Unix socket paths
}

// Allows reports whether every direct grant in requested is contained in a
// ceiling. It compares paths by components, never by string prefix.
func (ceiling Authority) Allows(requested Authority) bool {
	return pathsAllow(ceiling.Exec, requested.Exec) &&
		pathsAllow(ceiling.Read, requested.Read) &&
		pathsAllow(ceiling.Write, requested.Write) &&
		stringsAllow(ceiling.Mach, requested.Mach) && stringsAllow(ceiling.Connect, requested.Connect)
}

func pathsAllow(ceiling, requested []PathGrant) bool {
	for _, want := range requested {
		if want.Scope != "literal" && want.Scope != "subpath" {
			return false
		}
		found := false
		for _, have := range ceiling {
			if have.Scope == "literal" && want.Scope == "literal" && have.Path == want.Path {
				found = true
			}
			if have.Scope == "subpath" && filepath.IsAbs(have.Path) && filepath.IsAbs(want.Path) {
				rel, err := filepath.Rel(have.Path, want.Path)
				if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					found = true
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func stringsAllow(ceiling, requested []string) bool {
	for _, want := range requested {
		if !slices.Contains(ceiling, want) {
			return false
		}
	}
	return true
}

// Surface is a registered executable: what may run, pinned by content hash,
// with the contracts the runtime checks its input and its reply against.
type Surface struct {
	Path               string           `json:"path"`
	SHA256             string           `json:"sha256"`
	Input              *contract.Schema `json:"input"`                          // what invoking it must fit
	Output             *contract.Schema `json:"output"`                         // what its reply must fit: the reply capability's input contract
	Reach              []string         `json:"reach,omitempty"`                // what it may be granted, if its parent holds it. Empty: a leaf
	Authority          Authority        `json:"authority,omitempty"`            // direct OS grants, pinned in the registry
	Resident           bool             `json:"resident,omitempty"`             // launch and retain a stream connection in custody
	Connection         bool             `json:"connection,omitempty"`           // receive the caller's resident connection as fd 3
	CloseConnection    bool             `json:"close_connection,omitempty"`     // release that connection after this operation
	Args               []string         `json:"args,omitempty"`                 // fixed arguments for a resident target
	ResourceOwnerDepth int              `json:"resource_owner_depth,omitempty"` // ancestor frame identifying the resident connection
	ResidentProfile    string           `json:"resident_profile,omitempty"`     // named runtime-rendered profile template
	ProxyHosts         []string         `json:"proxy_hosts,omitempty"`          // exact HTTPS CONNECT names for this resident
	ProxyPath          string           `json:"proxy_path,omitempty"`           // pinned local proxy executable for this profile
	ProxySHA256        string           `json:"proxy_sha256,omitempty"`
	// Delegate attenuates the listed children's direct grants. An absent entry
	// passes the full grant already in this frame's Reach; it never adds one.
	Delegate map[string]Authority `json:"delegate,omitempty"`
	// Host, when set, makes the surface async (next/16): its invocations are
	// parked with the actor host on this socket, which holds custody until an
	// outside event completes them. There is no binary to run.
	Host string `json:"host,omitempty"`
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
	b.WriteString("(allow sysctl-read)\n")
	// These are loader and operating-system reads, not surface grants. In
	// particular, no user home or workspace is included here.
	for _, path := range []string{"/System", "/usr/lib", "/private/etc"} {
		fmt.Fprintf(&b, "(allow file-read* (subpath %q))\n", path)
	}
	for _, path := range []string{"/", "/dev/null", "/dev/random", "/dev/urandom", h.Surface.Path} {
		fmt.Fprintf(&b, "(allow file-read* (literal %q))\n", path)
	}
	for _, grant := range h.Surface.Authority.Exec {
		renderPath(&b, "process-exec", grant)
	}
	for _, grant := range h.Surface.Authority.Read {
		renderPath(&b, "file-read*", grant)
	}
	for _, grant := range h.Surface.Authority.Write {
		renderPath(&b, "file-write*", grant)
	}
	for _, service := range h.Surface.Authority.Mach {
		fmt.Fprintf(&b, "(allow mach-lookup (global-name %q))\n", service)
	}
	for _, sock := range h.Surface.Authority.Connect {
		fmt.Fprintf(&b, "(allow network-outbound (remote unix-socket (path-literal %q)))\n", sock)
	}
	fmt.Fprintf(&b, "(deny file-read* file-write* (subpath %q))\n", floor)
	return b.String()
}

func renderPath(b *strings.Builder, operation string, grant PathGrant) {
	if !filepath.IsAbs(grant.Path) || (grant.Scope != "literal" && grant.Scope != "subpath") {
		return // invalid input grants nothing; registry rejects it at genesis
	}
	fmt.Fprintf(b, "(allow %s (%s %q))\n", operation, grant.Scope, grant.Path)
}
