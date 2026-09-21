//go:build ignore

// Package surfaceregistry contains discussion declarations, not runtime code.
// Surface is the struct in ../../agent-surface-worker/surface. Package wiring remains provisional.
package surfaceregistry

import "github.com/google/uuid"

// SurfaceRegistry extracts the Surface table from DNA-17's StateExecutor and
// the earlier StateOwner userState. Access is serialized by the owning
// StateOwner's turn; this component has no independent writer or lock.
type SurfaceRegistry struct {
	// Keys equal Surface.Identity. Surface.Driver is the current handler slot.
	// Private storage only: callers receive neither this map nor mutable aliases.
	surfaces map[uuid.UUID]*Surface
}

// TODO: restricted registration/read profiles and the authorized execution
// binding. Do not expose unrestricted handler lookup to executing Surfaces.
// Replacement, retirement, per-Possibility binding, and recovery are unresolved.
