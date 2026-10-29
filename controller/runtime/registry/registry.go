// Package registry reads the operator's registry: which surfaces exist, where
// their binaries are, and the contracts they declare. Only the operator reads
// it, once, at boot, to build genesis: the root reach and its authority. There
// is no global registry after that.
//
// A surface carries no authority (next/25). An entry's "authority" is a grant
// the operator brings into genesis with it; every grant in the file becomes
// the root Boundary's one authority, which everything under it shares.
package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	"skyra/controller/core/record"
)

// Entry is one surface in the registry file. Relative paths are relative to
// the file.
type Entry struct {
	Uses      []string           `json:"uses,omitempty"` // the actors it runs through, kept when it is narrowed to
	Path      string             `json:"path"`
	Contract  string             `json:"contract"`
	Connect   []string           `json:"connect,omitempty"`   // unix sockets it may reach: an adapter's actor host
	Host      string             `json:"host,omitempty"`      // async (next/16): served by the actor host on this socket; no binary
	Authority boundary.Authority `json:"authority,omitempty"` // a grant this entry brings into the root Boundary
	State     bool               `json:"state,omitempty"`     // the actor keeps state of its own, in .skyra/<name>/ (next/32)
	Bundle    string             `json:"bundle,omitempty"`    // versioned authored procedure bundle, provisioned before genesis
}

// Declared is what a program's contract file says about it: its input
// contract, and how it reads its input. Nothing about what it returns, and
// nothing about what it may reach.
type Declared struct {
	Input        *contract.Schema   `json:"input"`
	Contracts    []*contract.Schema `json:"contracts,omitempty"` // its other contracts: one program, several entry points (next/32)
	Continuation bool               `json:"continuation,omitempty"`
}

// Load reads the registry and every contract it names, pins each binary by
// content hash, and returns the root Boundary: its reach, and the authority everything
// under it shares.
func Load(file string) (map[string]boundary.Surface, boundary.Authority, error) {
	var entries map[string]Entry
	if err := strict(file, &entries); err != nil {
		return nil, boundary.Authority{}, fmt.Errorf("registry: %w", err)
	}
	return load(file, entries)
}

func load(file string, entries map[string]Entry) (map[string]boundary.Surface, boundary.Authority, error) {
	var root boundary.Authority
	dir := filepath.Dir(file)
	reach := map[string]boundary.Surface{}
	for name, e := range entries {
		if name == "system" || strings.HasPrefix(name, "system.") {
			return nil, root, fmt.Errorf("%s: the Boundary's syscalls are the runtime's, never registered", name)
		}
		var d Declared
		if err := strict(resolve(dir, e.Contract), &d); err != nil {
			return nil, root, fmt.Errorf("%s contract: %w", name, err)
		}
		if err := d.Input.Check(); err != nil {
			return nil, root, fmt.Errorf("%s input contract: %w", name, err)
		}
		for i, c := range d.Contracts {
			if err := c.Check(); err != nil {
				return nil, root, fmt.Errorf("%s contract %d: %w", name, i+1, err)
			}
		}
		// Contracts enter genesis in their normal form (42), so a root
		// Boundary does not depend on the order a file lists required fields.
		d.Input = contract.Normal(d.Input)
		for i, c := range d.Contracts {
			d.Contracts[i] = contract.Normal(c)
		}
		var bundlePath, bundleSHA string
		if e.Bundle != "" {
			if !e.State || !d.Continuation || e.Host != "" {
				return nil, root, fmt.Errorf("%s: a bundle needs a stateful continuation", name)
			}
			var err error
			bundlePath, err = filepath.Abs(resolve(dir, e.Bundle))
			if err != nil {
				return nil, root, err
			}
			bundlePath, err = filepath.EvalSymlinks(bundlePath)
			if err != nil {
				return nil, root, err
			}
			b, err := os.ReadFile(bundlePath)
			if err != nil {
				return nil, root, err
			}
			var bundle struct {
				Version    int
				Procedures []struct{ Contract *contract.Schema }
			}
			if err := json.Unmarshal(b, &bundle); err != nil {
				return nil, root, err
			}
			if bundle.Version != 1 || len(bundle.Procedures) == 0 {
				return nil, root, fmt.Errorf("%s: invalid procedure bundle", name)
			}
			bundleContracts := map[string]bool{}
			for _, p := range bundle.Procedures {
				if err := p.Contract.Check(); err != nil {
					return nil, root, err
				}
				key := string(contract.Canonical(contract.Normal(p.Contract)))
				if bundleContracts[key] {
					return nil, root, fmt.Errorf("%s: bundle repeats a contract", name)
				}
				bundleContracts[key] = true
				// A bundle may implement the actor's primary contract itself.
				// It is already in reach, so do not advertise it twice.
				if contract.Same(p.Contract, d.Input) {
					continue
				}
				for _, held := range d.Contracts {
					if contract.Same(held, p.Contract) {
						return nil, root, fmt.Errorf("%s: bundle repeats a declared contract", name)
					}
				}
				d.Contracts = append(d.Contracts, contract.Normal(p.Contract))
			}
			bundleSHA = record.Sum(b)
		}
		if e.Host != "" {
			if e.Path != "" || len(e.Connect) > 0 || d.Continuation || len(e.Authority.Read)+len(e.Authority.Write)+len(e.Authority.Mach)+len(e.Authority.Connect) > 0 || e.Authority.Workspace != "" {
				return nil, root, fmt.Errorf("%s: an async surface cannot carry executable authority", name)
			}
			reach[name] = boundary.Surface{Host: e.Host, Input: d.Input}
			continue
		}
		if e.Path == "" {
			return nil, root, fmt.Errorf("%s: missing executable path", name)
		}
		path, err := filepath.EvalSymlinks(resolve(dir, e.Path))
		if err != nil {
			return nil, root, err
		}
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, root, err
		}
		authority, err := normalizeAuthority(dir, e.Authority)
		if err != nil {
			return nil, root, fmt.Errorf("%s authority: %w", name, err)
		}
		if len(e.Connect) > 0 {
			legacy := boundary.Authority{Connect: e.Connect}
			legacy, err = normalizeAuthority(dir, legacy)
			if err != nil {
				return nil, root, fmt.Errorf("%s connect: %w", name, err)
			}
			authority.Connect = append(authority.Connect, legacy.Connect...)
		}
		if authority.Workspace != "" && root.Workspace != "" && authority.Workspace != root.Workspace {
			return nil, root, fmt.Errorf("%s: a second workspace %s; the root has %s", name, authority.Workspace, root.Workspace)
		}
		root = union(root, authority)
		bin, err := os.ReadFile(path)
		if err != nil {
			return nil, root, fmt.Errorf("%s: %w", name, err)
		}
		var state string
		if e.State {
			if !boundary.ValidState(name) {
				return nil, root, fmt.Errorf("%s: a name that keeps state is one plain lowercase segment, none of the kernel's", name)
			}
			state = name
		}
		reach[name] = boundary.Surface{Uses: e.Uses, Path: path, SHA256: record.Sum(bin), Input: d.Input, Contracts: d.Contracts, Continuation: d.Continuation, State: state, BundlePath: bundlePath, BundleSHA256: bundleSHA}
	}
	if !root.Covers() {
		return nil, root, fmt.Errorf("the workspace %s is under no read grant", root.Workspace)
	}
	for name, s := range reach {
		for _, used := range s.Uses {
			if _, ok := reach[used]; !ok || used == name {
				return nil, root, fmt.Errorf("%s: uses %q, which is not another entry", name, used)
			}
		}
	}
	return reach, root, nil
}

// union adds a grant to the root authority, keeping each grant once and
// the result in a stable order, so genesis is the same for the same file.
// The order means nothing: the workspace is named, not the first read grant.
func union(a, b boundary.Authority) boundary.Authority {
	grants := func(x, y []boundary.PathGrant) []boundary.PathGrant {
		for _, g := range y {
			if !slices.Contains(x, g) {
				x = append(x, g)
			}
		}
		slices.SortFunc(x, func(p, q boundary.PathGrant) int { return strings.Compare(p.Scope+" "+p.Path, q.Scope+" "+q.Path) })
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
	workspace := a.Workspace
	if workspace == "" {
		workspace = b.Workspace
	}
	return boundary.Authority{Read: grants(a.Read, b.Read), Write: grants(a.Write, b.Write), Mach: names(a.Mach, b.Mach), Connect: names(a.Connect, b.Connect), Workspace: workspace}
}

func normalizeAuthority(dir string, a boundary.Authority) (boundary.Authority, error) {
	var err error
	if a.Read, err = normalizePaths(dir, a.Read); err != nil {
		return a, err
	}
	if a.Write, err = normalizePaths(dir, a.Write); err != nil {
		return a, err
	}
	if a.Workspace != "" {
		paths, err := normalizePaths(dir, []boundary.PathGrant{{Path: a.Workspace, Scope: "subpath"}})
		if err != nil {
			return a, err
		}
		a.Workspace = paths[0].Path
	}
	for i, sock := range a.Connect {
		paths, err := normalizePaths(dir, []boundary.PathGrant{{Path: sock, Scope: "literal"}})
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

func normalizePaths(dir string, paths []boundary.PathGrant) ([]boundary.PathGrant, error) {
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
