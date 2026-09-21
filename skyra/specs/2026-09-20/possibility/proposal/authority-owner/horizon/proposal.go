//go:build ignore

// Package horizon contains discussion declarations, not runtime code.
package horizon

import "github.com/google/uuid"

// Horizon starts from the September 18 and DNA-17 authority record.
// Proposed role: the admission and executable-resolution boundary for every
// Move in its Possibility, including returned outcomes and state operations.
// Gate reads/decisions happen in interpret; commit applies the prepared writes.
type Horizon struct {
	Generation uint64
	Grants     map[uuid.UUID]Grant

	// NEW: registered address of the TrapController Surface. This is a
	// reference, not a live callback or permission to execute it unchecked.
	// Controller provisioning and the gate's method signatures remain open.
	TrapControllerID uuid.UUID
}

type Grant struct {
	SurfaceID uuid.UUID
}

// TODO: define the trap/fault gate's input, admitted binding, and trap outcome.
// Whether bindings live here or in a separate runtime registry is unresolved.
// Authority checking and executable resolution must form one interpretation
// decision whose reads remain valid through commit.
