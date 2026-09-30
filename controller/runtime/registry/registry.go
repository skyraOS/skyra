// Package registry reads the operator's registry: which surfaces exist, where
// their binaries are, and the contracts they declare. Only the operator reads
// it, once, to build the room: genesis's root reach and its authority. There
// is no global registry after that.
//
// A surface carries no authority (next/25). An entry's "authority" is a grant
// the operator brings into the room with it; every grant in the file becomes
// the room's one authority, which everything under the root Horizon shares.
package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// Entry is one surface in the registry file. Relative paths are relative to
// the file.
type Entry struct {
	Path               string            `json:"path"`
	Contract           string            `json:"contract"`
	Connect            []string          `json:"connect,omitempty"`   // unix sockets it may reach: an adapter's actor host
	Host               string            `json:"host,omitempty"`      // async (next/16): served by the actor host on this socket; no binary
	Authority          horizon.Authority `json:"authority,omitempty"` // a grant this entry brings into the room
	Resident           bool              `json:"resident,omitempty"`
	Connection         bool              `json:"connection,omitempty"`
	CloseConnection    bool              `json:"close_connection,omitempty"`
	Args               []string          `json:"args,omitempty"`
	ResourceOwnerDepth int               `json:"resource_owner_depth,omitempty"`
	ResidentProfile    string            `json:"resident_profile,omitempty"`
	ProxyHosts         []string          `json:"proxy_hosts,omitempty"`
	ProxyPath          string            `json:"proxy_path,omitempty"`
}

// Declared is what a program's contract file says about it: its contracts,
// and how it reads its input. Nothing about what it may reach.
type Declared struct {
	Input        *contract.Schema `json:"input"`
	Output       *contract.Schema `json:"output"`
	Continuation bool             `json:"continuation,omitempty"`
}

// Load reads the registry and every contract it names, pins each binary by
// content hash, and returns the room: its reach, and the authority everything
// under it shares.
func Load(file string) (map[string]horizon.Surface, horizon.Authority, error) {
	var room horizon.Authority
	var entries map[string]Entry
	if err := strict(file, &entries); err != nil {
		return nil, room, fmt.Errorf("registry: %w", err)
	}
	dir := filepath.Dir(file)
	reach := map[string]horizon.Surface{}
	for name, e := range entries {
		var d Declared
		if err := strict(resolve(dir, e.Contract), &d); err != nil {
			return nil, room, fmt.Errorf("%s contract: %w", name, err)
		}
		if err := d.Input.Check(); err != nil {
			return nil, room, fmt.Errorf("%s input contract: %w", name, err)
		}
		if err := d.Output.Check(); err != nil {
			return nil, room, fmt.Errorf("%s output contract: %w", name, err)
		}
		if e.Host != "" {
			if e.Path != "" || e.Resident || e.Connection || len(e.Args) > 0 || len(e.Connect) > 0 || d.Continuation || len(e.Authority.Read)+len(e.Authority.Write)+len(e.Authority.Mach)+len(e.Authority.Connect) > 0 {
				return nil, room, fmt.Errorf("%s: an async surface cannot carry executable authority", name)
			}
			reach[name] = horizon.Surface{Host: e.Host, Input: d.Input, Output: d.Output}
			continue
		}
		if e.Path == "" {
			return nil, room, fmt.Errorf("%s: missing executable path", name)
		}
		if e.Resident && e.Connection || e.CloseConnection && !e.Connection || !e.Resident && (len(e.Args) != 0 || e.ResidentProfile != "" || len(e.ProxyHosts) != 0 || e.ProxyPath != "") {
			return nil, room, fmt.Errorf("%s: invalid resident or connection configuration", name)
		}
		if e.ResidentProfile != "" && (e.ResidentProfile != "darwin-appserver-v1" || len(e.ProxyHosts) == 0 || e.ProxyPath == "") {
			return nil, room, fmt.Errorf("%s: unknown resident profile or no proxy hosts", name)
		}
		for _, host := range e.ProxyHosts {
			if host == "" || host != strings.ToLower(host) || strings.ContainsAny(host, "/:@ \n\r\x00") {
				return nil, room, fmt.Errorf("%s: invalid proxy host %q", name, host)
			}
		}
		if e.ResourceOwnerDepth < 0 || e.ResourceOwnerDepth > 8 || e.ResourceOwnerDepth > 0 && !e.Resident && !e.Connection {
			return nil, room, fmt.Errorf("%s: invalid resource owner depth", name)
		}
		path, err := filepath.EvalSymlinks(resolve(dir, e.Path))
		if err != nil {
			return nil, room, err
		}
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, room, err
		}
		authority, err := normalizeAuthority(dir, e.Authority)
		if err != nil {
			return nil, room, fmt.Errorf("%s authority: %w", name, err)
		}
		if len(e.Connect) > 0 {
			legacy := horizon.Authority{Connect: e.Connect}
			legacy, err = normalizeAuthority(dir, legacy)
			if err != nil {
				return nil, room, fmt.Errorf("%s connect: %w", name, err)
			}
			authority.Connect = append(authority.Connect, legacy.Connect...)
		}
		room = union(room, authority)
		bin, err := os.ReadFile(path)
		if err != nil {
			return nil, room, fmt.Errorf("%s: %w", name, err)
		}
		var proxyPath, proxySHA string
		if e.ProxyPath != "" {
			proxyPath, err = filepath.EvalSymlinks(resolve(dir, e.ProxyPath))
			if err != nil {
				return nil, room, err
			}
			var proxy []byte
			proxy, err = os.ReadFile(proxyPath)
			if err != nil {
				return nil, room, err
			}
			proxySHA = record.Sum(proxy)
		}
		reach[name] = horizon.Surface{Path: path, SHA256: record.Sum(bin), Input: d.Input, Output: d.Output, Continuation: d.Continuation, Resident: e.Resident, Connection: e.Connection, CloseConnection: e.CloseConnection, Args: e.Args, ResourceOwnerDepth: e.ResourceOwnerDepth, ResidentProfile: e.ResidentProfile, ProxyHosts: e.ProxyHosts, ProxyPath: proxyPath, ProxySHA256: proxySHA}
	}
	return reach, room, nil
}

// union adds a grant to the room's authority, keeping each grant once and
// the result in a stable order, so genesis is the same for the same file.
func union(a, b horizon.Authority) horizon.Authority {
	grants := func(x, y []horizon.PathGrant) []horizon.PathGrant {
		for _, g := range y {
			if !slices.Contains(x, g) {
				x = append(x, g)
			}
		}
		slices.SortFunc(x, func(p, q horizon.PathGrant) int { return strings.Compare(p.Scope+" "+p.Path, q.Scope+" "+q.Path) })
		return x
	}
	names := func(x, y []string) []string {
		for _, n := range y {
			if !slices.Contains(x, n) {
				x = append(x, n)
			}
		}
		slices.Sort(x)
		return x
	}
	return horizon.Authority{Read: grants(a.Read, b.Read), Write: grants(a.Write, b.Write), Mach: names(a.Mach, b.Mach), Connect: names(a.Connect, b.Connect)}
}

func normalizeAuthority(dir string, a horizon.Authority) (horizon.Authority, error) {
	var err error
	if a.Read, err = normalizePaths(dir, a.Read); err != nil {
		return a, err
	}
	if a.Write, err = normalizePaths(dir, a.Write); err != nil {
		return a, err
	}
	for i, sock := range a.Connect {
		paths, err := normalizePaths(dir, []horizon.PathGrant{{Path: sock, Scope: "literal"}})
		if err != nil {
			return a, err
		}
		a.Connect[i] = paths[0].Path
	}
	for _, name := range a.Mach {
		if name == "" || strings.ContainsAny(name, "\n\r\x00") {
			return a, fmt.Errorf("invalid mach service %q", name)
		}
	}
	return a, nil
}

func normalizePaths(dir string, paths []horizon.PathGrant) ([]horizon.PathGrant, error) {
	for i, grant := range paths {
		if grant.Path == "" || (grant.Scope != "literal" && grant.Scope != "subpath") {
			return nil, fmt.Errorf("invalid path grant %+v", grant)
		}
		absolute, err := filepath.Abs(resolve(dir, grant.Path))
		if err != nil {
			return nil, err
		}
		// A write target may not exist yet. Resolve the nearest existing parent,
		// then append the remaining components, so symlink aliases cannot be used
		// to defeat component-wise containment checks.
		parent, tail := absolute, ""
		for {
			resolved, err := filepath.EvalSymlinks(parent)
			if err == nil {
				grant.Path = filepath.Join(resolved, tail)
				break
			}
			if !os.IsNotExist(err) || parent == filepath.Dir(parent) {
				return nil, err
			}
			tail = filepath.Join(filepath.Base(parent), tail)
			parent = filepath.Dir(parent)
		}
		paths[i] = grant
	}
	return paths, nil
}

func strict(file string, v any) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func resolve(dir, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(dir, p)
}
