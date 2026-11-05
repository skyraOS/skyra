// Package mission is the trusted bridge from Mission contracts and human-root
// observations to the existing worker protocol. No worker instruction is added.
package mission

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	root "skyra/controller/core/humanroot"
	typed "skyra/controller/core/mission"
	"skyra/controller/runtime/registry"
)

type Catalogue struct {
	Typed       typed.Catalogue
	entries     map[string]registry.Entry
	surfaces    map[string]boundary.Surface
	authorities map[string]boundary.Authority
	source      string
}

// Load pins actual binaries, schemas and resource scopes using the operator's
// registry. Only the explicitly offered task targets become proposed tools.
func Load(file string, targets []string) (*Catalogue, error) {
	file, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	surfaces, ceiling, err := registry.Load(file)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var entries map[string]registry.Entry
	if err = json.Unmarshal(b, &entries); err != nil {
		return nil, err
	}
	c := &Catalogue{entries: entries, surfaces: surfaces, source: file, authorities: map[string]boundary.Authority{}}
	for name, entry := range entries {
		a, err := registry.EntryAuthority(file, entry)
		if err != nil {
			return nil, err
		}
		c.authorities[name] = a
	}
	var descriptors []typed.Descriptor
	for i, target := range targets {
		if slices.Contains(targets[:i], target) {
			return nil, fmt.Errorf("duplicate task target %s", target)
		}
		surface, ok := surfaces[target]
		if !ok || surface.Path == "" {
			return nil, fmt.Errorf("task target %s is not a registered executable", target)
		}
		if surface.BundlePath != "" {
			return nil, fmt.Errorf("bundle-backed task %s requires a scoped bundle adapter", target)
		}
		scope, err := c.scope(target, map[string]bool{})
		if err != nil {
			return nil, err
		}
		for _, schema := range surface.All() {
			identity, err := typed.Identify("sha256:"+surface.SHA256, schema)
			if err != nil {
				return nil, err
			}
			descriptors = append(descriptors, typed.Descriptor{Target: target, Contract: identity, Pin: boundary.Program(surface), Authority: scope})
		}
	}
	t, err := typed.NewCatalogue(ceiling, descriptors)
	if err != nil {
		return nil, err
	}
	c.Typed = t
	return c, nil
}
func (c *Catalogue) scope(name string, seen map[string]bool) (boundary.Authority, error) {
	if seen[name] {
		return boundary.Authority{}, nil
	}
	seen[name] = true
	_, ok := c.entries[name]
	if !ok {
		return boundary.Authority{}, fmt.Errorf("missing dependency %s", name)
	}
	a := c.authorities[name]
	for _, dep := range c.surfaces[name].Uses {
		scope, err := c.scope(dep, seen)
		if err != nil {
			return a, err
		}
		a, err = registry.CombineAuthority(a, scope)
		if err != nil {
			return a, err
		}
	}
	return a, nil
}

func absolute(dir, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(dir, path)
}

// Prepare produces an operator-owned registry with precisely the selected task
// contracts and declared dependencies. Provider operations and the apprentice
// are infrastructure, not additional task contracts offered to the model.
func (c *Catalogue) Prepare(dir, actor string, resolved typed.Resolved) (string, json.RawMessage, error) {
	selected := map[string][]*contract.Schema{}
	for _, d := range resolved.Contracts {
		held, ok := c.Typed.Lookup(d.Ref)
		if !ok || held.Target != d.Target || held.Contract != d.Contract || held.Pin != d.Pin {
			return "", nil, fmt.Errorf("resolved capability is not in the pinned catalogue")
		}
		var schema contract.Schema
		if err := json.Unmarshal([]byte(d.Contract.Shape), &schema); err != nil {
			return "", nil, err
		}
		if !slices.ContainsFunc(selected[d.Target], func(s *contract.Schema) bool { return contract.Same(s, &schema) }) {
			selected[d.Target] = append(selected[d.Target], &schema)
		}
	}
	count := map[typed.Contract]bool{}
	for _, d := range resolved.Contracts {
		count[d.Contract] = true
	}
	kind := "act"
	if len(count) > 1 {
		kind = "delegate"
	}
	if kind != resolved.Kind {
		return "", nil, fmt.Errorf("worker kind differs from the selected contract types")
	}
	if len(selected) == 0 {
		return "", nil, fmt.Errorf("cannot prepare an empty task grant")
	}
	include := map[string]bool{}
	var add func(string) error
	add = func(name string) error {
		if include[name] {
			return nil
		}
		s, ok := c.surfaces[name]
		if !ok {
			return fmt.Errorf("unregistered infrastructure %s", name)
		}
		include[name] = true
		for _, dep := range s.Uses {
			if err := add(dep); err != nil {
				return err
			}
		}
		return nil
	}
	for name := range selected {
		if err := add(name); err != nil {
			return "", nil, err
		}
	}
	if err := add(actor); err != nil {
		return "", nil, err
	}
	if err := add("apprentice"); err != nil {
		return "", nil, err
	}
	if _, task := selected[actor]; task {
		return "", nil, fmt.Errorf("provider cannot also be a task capability")
	}
	if _, task := selected["apprentice"]; task {
		return "", nil, fmt.Errorf("apprentice cannot also be a task capability")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", nil, err
	}
	projected := map[string]registry.Entry{}
	var expectedAuthority boundary.Authority
	sourceDir := filepath.Dir(c.source)
	for name := range include {
		entry := c.entries[name]
		surface := c.surfaces[name]
		// Recheck before publishing: a catalogue is not permission to run a changed
		// file at the same path. Kernel boot pins and verifies these again.
		if surface.Path != "" {
			b, err := os.ReadFile(surface.Path)
			if err != nil {
				return "", nil, err
			}
			if sum(b) != surface.SHA256 {
				return "", nil, fmt.Errorf("registered binary %s changed", name)
			}
		}
		entry.Path = surface.Path
		entry.Contract = absolute(sourceDir, entry.Contract)
		entry.Bundle = absolute(sourceDir, entry.Bundle)
		authority := c.authorities[name]
		var err error
		expectedAuthority, err = registry.CombineAuthority(expectedAuthority, authority)
		if err != nil {
			return "", nil, err
		}
		entry.Authority = authority
		entry.Connect = nil
		schemas := surface.All()
		if chosen, ok := selected[name]; ok {
			schemas = chosen
		}
		if surface.BundlePath != "" && name != "apprentice" {
			return "", nil, fmt.Errorf("bundle-backed infrastructure requires a scoped bundle adapter")
		}

		if name == "apprentice" {
			entry.Bundle = "" // Launch only needs the ordinary tree contract.
			for _, schema := range surface.All() {
				if schema.Title == "tree" {
					schemas = []*contract.Schema{schema}
					break
				}
			}
			if len(schemas) != 1 || schemas[0].Title != "tree" {
				return "", nil, fmt.Errorf("apprentice has no tree contract")
			}
		}
		if len(schemas) > 0 {
			declaration := registry.Declared{Input: schemas[0], Contracts: schemas[1:], Continuation: surface.Continuation}
			b, _ := json.Marshal(declaration)
			path := filepath.Join(dir, name+"-contract.json")
			if err = os.WriteFile(path, b, 0600); err != nil {
				return "", nil, err
			}
			entry.Contract = path
		}
		projected[name] = entry
	}
	path := filepath.Join(dir, "registry.json")
	b, _ := json.Marshal(projected)
	if err := os.WriteFile(path, b, 0600); err != nil {
		return "", nil, err
	}
	pinned, authority, err := registry.Load(path)
	if err != nil {
		return "", nil, err
	}
	if !c.Typed.Ceiling().Allows(authority) || !expectedAuthority.Allows(authority) || !authority.Allows(expectedAuthority) {
		return "", nil, fmt.Errorf("prepared registry exceeds the human's authority ceiling")
	}
	for name, schemas := range selected {
		s := pinned[name]
		if boundary.Program(s) != boundary.Program(c.surfaces[name]) || len(s.All()) != len(schemas) {
			return "", nil, fmt.Errorf("projection changed %s's identity or contract set", name)
		}
		for _, schema := range schemas {
			if !s.Declares(schema) {
				return "", nil, fmt.Errorf("projection dropped a selected contract")
			}
		}
	}
	names := make([]string, 0, len(selected))
	for name := range selected {
		names = append(names, name)
	}
	slices.Sort(names)
	payload, _ := json.Marshal(map[string]any{"tree": map[string]any{"goal": resolved.Spec.Goal, "kind": resolved.Kind, "actor": actor, "reach": names}})
	return path, payload, nil
}

// CapabilityRefs exposes held catalogue references; it conveys no new grant.
func (c *Catalogue) CapabilityRefs() []root.Capability {
	out := []root.Capability{}
	for _, d := range c.Typed.Entries() {
		out = append(out, d.Ref)
	}
	return out
}
