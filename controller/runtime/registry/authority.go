package registry

import (
	"fmt"
	"path/filepath"
	"slices"

	"skyra/controller/core/boundary"
)

// EntryAuthority resolves one registered entry's grants with the same rules as
// Load. It is used by the trusted Mission catalogue, never by model payloads.
func EntryAuthority(file string, e Entry) (boundary.Authority, error) {
	dir := filepath.Dir(file)
	a, err := normalizeAuthority(dir, e.Authority)
	if err != nil {
		return a, err
	}
	if len(e.Connect) > 0 {
		legacy, err := normalizeAuthority(dir, boundary.Authority{Connect: e.Connect})
		if err != nil {
			return a, err
		}
		a = union(a, legacy)
	}
	return a, nil
}

// CombineAuthority unions only already-normalized held grants. Conflicting
// workspace/scratch roots are refused rather than silently choosing one.
func CombineAuthority(a, b boundary.Authority) (boundary.Authority, error) {
	// union sorts and appends into its first argument. Keep held snapshots
	// immutable, including when their slices have spare capacity.
	a.Read = slices.Clone(a.Read)
	a.Write = slices.Clone(a.Write)
	a.Mach = slices.Clone(a.Mach)
	a.Connect = slices.Clone(a.Connect)
	if a.Workspace != "" && b.Workspace != "" && a.Workspace != b.Workspace {
		return a, fmt.Errorf("conflicting workspaces")
	}
	if a.Scratch != "" && b.Scratch != "" && a.Scratch != b.Scratch {
		return a, fmt.Errorf("conflicting scratch grants")
	}
	scratch := a.Scratch
	if scratch == "" {
		scratch = b.Scratch
	}
	out := union(a, b)
	out.Scratch = scratch
	return out, nil
}
