//go:build ignore

// Package possibility contains proposed contracts, not a runtime implementation.
// Move is defined in the sibling move proposal; package wiring remains provisional.
package possibility

import (
	"context"

	"github.com/google/uuid"
)

// Possibility is the canonical identity-and-authority record for one run. A run
// may contain many invocation identities. Lifecycle, execution, and native
// resource state are owned elsewhere.
type Possibility struct {
	Identity    uuid.UUID
	FirstMoveID uuid.UUID
	Horizon     Horizon
}

type PossibilityMap map[uuid.UUID]Possibility

// Horizon is the authority installed for one Possibility. Generation changes
// only when the grant set changes; it is distinct from StateOwner's global
// revision used for optimistic concurrency control.
type Horizon struct {
	Generation uint64
	Grants     map[uuid.UUID]Grant
}

// Grant authorizes dispatch to one independently addressable Surface. It names
// semantic identity, not a handler, transport, process, credential, or driver.
// The explicit record leaves room for later constraints without changing the
// shape of Horizon.
type Grant struct {
	SurfaceID uuid.UUID
}

// PossibilitySnapshot is detached from canonical StateOwner data. In
// particular, Horizon.Grants must be copied before the snapshot is returned.
type PossibilitySnapshot struct {
	Revision    uint64
	Possibility Possibility
}

// OpenPossibility is committed atomically with First: the Possibility record,
// first Move record, Interpretation custody, and Interpretation mailbox entry
// either all commit or none commit.
type OpenPossibility struct {
	Identity uuid.UUID
	First    Move
	Horizon  Horizon
}

// ReplaceHorizon replaces the complete grant set. ExpectedGeneration prevents
// one authority update from silently overwriting another. StateOwner copies and
// validates Grants before committing and increments Horizon.Generation.
type ReplaceHorizon struct {
	PossibilityID      uuid.UUID
	ExpectedGeneration uint64
	Grants             []Grant
}

// PossibilityState is a restricted StateOwner profile. It does not expose
// lifecycle operations or general map access.
type PossibilityState interface {
	Open(
		ctx context.Context,
		request OpenPossibility,
	) (PossibilitySnapshot, error)

	ReplaceHorizon(
		ctx context.Context,
		request ReplaceHorizon,
	) (PossibilitySnapshot, error)
}
