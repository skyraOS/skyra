// Package boundary owns authority: what a frame runs, what it may reach, and
// the OS sandbox that authority renders to. It imports only package contract.
package boundary

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"skyra/controller/core/contract"
)

// PathGrant names one OS path. Literal grants only that path; subpath grants
// the path and its descendants. Registry loading resolves paths before they
// enter a Boundary, so Render never consults the filesystem.
type PathGrant struct {
	Path  string `json:"path"`
	Scope string `json:"scope"` // literal or subpath
}

// Authority is the OS authority a Boundary grants to everything that runs
// under it (next/24). Exec is not among it: a profile allows exec of the one
// program being run and nothing else, so no grant can let a target start
// another program without the runtime recording it.
//
// Scratch is a directory everything under the Boundary may read and write: a
// place to put work, not a place to run it. Nothing written there can be
// executed, by the same rule that holds everywhere: a profile allows exec of
// the program being run and nothing else. It is its own field, not a read
// and a write grant, so a narrowing can drop it by name.
//
// Workspace is where a relative read resolves. It is information, not
// authority: a read grant must cover it, and the profile enforces the grant.
// It is named, never inferred from the order of the read grants, which the
// registry sorts.
type Authority struct {
	Read      []PathGrant `json:"read,omitempty"`
	Write     []PathGrant `json:"write,omitempty"`
	Mach      []string    `json:"mach,omitempty"`
	Connect   []string    `json:"connect,omitempty"`   // outbound Unix socket paths
	Scratch   string      `json:"scratch,omitempty"`   // an absolute directory, read and write
	Workspace string      `json:"workspace,omitempty"` // where a relative read resolves; under a read grant
}

// Allows reports whether every direct grant in requested is contained in a
// ceiling. It compares paths by components, never by string prefix.
func (ceiling Authority) Allows(requested Authority) bool {
	return pathsAllow(ceiling.Read, requested.Read) &&
		pathsAllow(ceiling.Write, requested.Write) &&
		stringsAllow(ceiling.Mach, requested.Mach) && stringsAllow(ceiling.Connect, requested.Connect) &&
		(requested.Scratch == "" || requested.Scratch == ceiling.Scratch) &&
		(requested.Workspace == "" || requested.Workspace == ceiling.Workspace)
}

// Covers reports whether the read grants cover the workspace, if there is one.
func (a Authority) Covers() bool {
	return a.Workspace == "" || pathsAllow(a.Read, []PathGrant{{Path: a.Workspace, Scope: "subpath"}})
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

// Surface is a registered executable: what a program is, not what it may do
// (next/25). Its identity (the binary, pinned by content hash) and its
// contracts. What it may reach and touch belongs to the Boundary it runs under.
// AgentConfig identifies operator-approved agents and their private transport operations.
type AgentConfig struct {
	Operations []string `json:"operations"`
}

// ToolBinding identifies executable contracts, rather than prompt labels alone.
type ToolBinding struct {
	Target    string             `json:"target"`
	Contracts []*contract.Schema `json:"contracts"`
}

type Surface struct {
	WorkTool     bool               `json:"work_tool,omitempty"`
	Agent        *AgentConfig       `json:"agent,omitempty"`
	Path         string             `json:"path"`
	SHA256       string             `json:"sha256"`
	Input        *contract.Schema   `json:"input"`                  // what invoking it must fit: its registered contract
	Contracts    []*contract.Schema `json:"contracts,omitempty"`    // its other contracts: a syscall's, or one the actor added to itself (next/32)
	Continuation bool               `json:"continuation,omitempty"` // reads a Continuation on stdin; otherwise its bare payload
	System       string             `json:"system,omitempty"`       // a syscall facet, implemented by the runtime (next/26): no binary, never registrable
	// State names the actor's own directory under the floor (next/32,
	// decision 24): the one part of .skyra/ its sandbox may read and write.
	// An actor can only see its own state.
	State string `json:"state,omitempty"`
	// Bundle pins operator-provisioned procedure definitions. Bootstrap passes
	// it to the actor's installer before genesis; targets gain no file grant.
	BundlePath   string `json:"bundle_path,omitempty"`
	BundleSHA256 string `json:"bundle_sha256,omitempty"`
	// Host, when set, makes the surface async (next/16): its invocations are
	// parked with the actor host on this socket, which holds custody until an
	// outside event completes them. There is no binary to run.
	Host string `json:"host,omitempty"`
}

// All is every contract of a surface, the registered one first. A contract
// is identified by (target, schema): an invoke names the target, and the
// payload selects the one contract it fits (next/32, decision 20).
func (s Surface) All() []*contract.Schema {
	var all []*contract.Schema
	if s.Input != nil {
		all = append(all, s.Input)
	}
	return append(all, s.Contracts...)
}

// Boundary is everything a step knows besides the bytes. It is the release
// (D6), sent by the parent after it verifies the child's commit. The data
// travels as JSON. The frame stack travels beside it as directory fds over
// SCM_RIGHTS.
//
// Reach and Authority are the Boundary proper (next/24): what emissions under
// it may invoke, and the OS grants everything under it runs with. Every call
// shares its caller's; genesis holds the root's. There is no
// global registry lookup.
//
// The rest places a step in the stack: Frame is the slot its return
// discharges, and Resume the continuation waiting on it, as it was when it
// emitted (its Stamp is the round that opened the slot). The operator's
// Boundary has no Resume.
type Boundary struct {
	WorkTools *map[string]ToolBinding `json:"work_tools,omitempty"`
	Stamp     string                  `json:"stamp"`             // this frame's latest commit; the predecessor of whatever is emitted under it
	Surface   *Surface                `json:"surface,omitempty"` // what this frame runs; nil for the operator
	Reach     map[string]Surface      `json:"reach"`             // what emissions under this Boundary may invoke
	Authority Authority               `json:"authority"`         // the OS grants everything under this Boundary runs with
	Refs      map[string]Grant        `json:"refs,omitempty"`    // references received from system.narrow: each names a narrowed Boundary for one actor (next/26)
	Name      string                  `json:"name,omitempty"`    // the name in reach Surface was invoked through (for a reference, its actor): who an extend-reach extends
	Origin    string                  `json:"origin,omitempty"`  // the invoke that created this Boundary by a narrowing; "" for the root's. Frames that share it share one Boundary
	Forked    bool                    `json:"forked,omitempty"`  // this frame, or one below it in this Boundary, is one of several siblings (fan-out)
	Frame     string                  `json:"frame"`             // the frame whose slot a return discharges
	Resume    *Boundary               `json:"resume,omitempty"`  // the continuation waiting on Frame's slot
}

// Grant is what a reference stands for: an actor in reach, and the narrowed
// Boundary it runs under when the reference is invoked.
type Grant struct {
	WorkTools *map[string]ToolBinding `json:"work_tools,omitempty"`
	Actor     string                  `json:"actor"`
	Reach     []string                `json:"reach"`
	Authority Authority               `json:"authority"`
}

// Depth is the number of frames in the stack that travels with this Boundary.
func (h Boundary) Depth() int {
	n := 1
	for r := h.Resume; r != nil; r = r.Resume {
		n++
	}
	return n
}

// Names lists what the Boundary reaches, sorted.
func (h Boundary) Names() []string {
	var n []string
	for k := range h.Reach {
		n = append(n, k)
	}
	slices.Sort(n)
	return n
}

// Render turns a Boundary into a Seatbelt profile (D7): the Boundary's grants,
// the same for everything under it, plus the exclusions no Boundary can grant
// or remove (next/24). Everything is denied unless listed. The floor (the
// .skyra/ directory) is always denied but for the running actor's own state
// directory (next/32), and exec is allowed for the one program being run and
// nothing else: the runtime's verified copy of it (43 D5), never the path it
// was registered at. The floor is a parameter, not a Boundary field, so no
// Boundary can grant it.
func Render(h Boundary, floor string) string {
	var b strings.Builder
	b.WriteString("(version 1)\n(deny default)\n")
	exe := Verified(h, floor)
	if exe != "" {
		fmt.Fprintf(&b, "(allow process-exec (literal %q))\n", exe)
	}
	b.WriteString("(allow sysctl-read)\n")
	// These are loader and operating-system reads, not surface grants. In
	// particular, no user home or workspace is included here.
	for _, path := range []string{"/System", "/usr/lib", "/private/etc"} {
		fmt.Fprintf(&b, "(allow file-read* (subpath %q))\n", path)
	}
	for _, path := range []string{"/", "/dev/null", "/dev/random", "/dev/urandom"} {
		fmt.Fprintf(&b, "(allow file-read* (literal %q))\n", path)
	}
	for _, grant := range h.Authority.Read {
		renderPath(&b, "file-read*", grant)
	}
	for _, grant := range h.Authority.Write {
		renderPath(&b, "file-write*", grant)
	}
	if s := h.Authority.Scratch; s != "" {
		renderPath(&b, "file-read*", PathGrant{Path: s, Scope: "subpath"})
		renderPath(&b, "file-write*", PathGrant{Path: s, Scope: "subpath"})
	}
	for _, service := range h.Authority.Mach {
		fmt.Fprintf(&b, "(allow mach-lookup (global-name %q))\n", service)
	}
	for _, sock := range h.Authority.Connect {
		fmt.Fprintf(&b, "(allow network-outbound (remote unix-socket (path-literal %q)))\n", sock)
	}
	fmt.Fprintf(&b, "(deny file-read* file-write* (subpath %q))\n", floor)
	if dir := StateDir(h, floor); dir != "" {
		// After the floor's deny, so it wins for this one directory.
		fmt.Fprintf(&b, "(allow file-read* file-write* (subpath %q))\n", dir)
	}
	if exe != "" {
		// After the floor's deny, so the copy the runtime checked can be read to run.
		fmt.Fprintf(&b, "(allow file-read* (literal %q))\n", exe)
	}
	return b.String()
}

// Verified is where the runtime keeps the copy of h's program it checked and
// runs (43 D5): under the floor, which only the runtime can write, named by
// the hash of its bytes. It is "" if the surface has no well-formed hash.
func Verified(h Boundary, floor string) string {
	if h.Surface == nil || len(h.Surface.SHA256) != 64 || strings.Trim(h.Surface.SHA256, "0123456789abcdef") != "" {
		return ""
	}
	return filepath.Join(floor, "surfaces", h.Surface.SHA256)
}

// StateDir is the running actor's own directory under the floor, or "".
func StateDir(h Boundary, floor string) string {
	if h.Surface == nil || !ValidState(h.Surface.State) {
		return ""
	}
	return filepath.Join(floor, h.Surface.State)
}

// ValidState reports whether a name may be an actor's state directory: one
// plain path segment that is none of the kernel's own directories.
func ValidState(name string) bool {
	if name == "" || len(name) > 64 || slices.Contains([]string{"records", "inputs", "boundaries", "frames", "acceptances", "surfaces"}, name) {
		return false
	}
	for i, c := range name {
		if !(c >= 'a' && c <= 'z' || i > 0 && (c >= '0' && c <= '9' || c == '-' || c == '_')) {
			return false
		}
	}
	return true
}

func renderPath(b *strings.Builder, operation string, grant PathGrant) {
	if !filepath.IsAbs(grant.Path) || (grant.Scope != "literal" && grant.Scope != "subpath") {
		return // invalid input grants nothing; registry rejects it at genesis
	}
	fmt.Fprintf(b, "(allow %s (%s %q))\n", operation, grant.Scope, grant.Path)
}

// Identity is a surface's identity: the hash of its canonical encoding. It
// covers a binary (its sha256) and a host surface alike, and any change to a
// registration is a new identity (next/30). Names are labels, never part of it.
func Identity(s Surface) string {
	b, _ := json.Marshal(s)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
