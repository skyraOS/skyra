//go:build ignore

// Package stateowner contains discussion declarations, not runtime code.
// Referenced sibling types and unresolved record types are conceptual, as in
// the previous dated proposals. No empty-interface substitutes are introduced.
// Surface is the struct in ../../agent-surface-worker/surface; wiring is provisional.
package stateowner

import (
	"sync"

	"github.com/google/uuid"
)

// StateOwner is a Surface conforming to the Skyra protocol. It receives Moves
// through its provisioned binding and emits Moves like any other Surface.
// Embedding supplies Identity, Contracts, Dependencies, Driver, IngressProfile,
// and EgressProfile from the common proposed Surface shape.
// The private storage/pool grouping below is retained from the earlier draft;
// ownership and coordinator placement remain provisional.
type StateOwner struct {
	Surface

	mu       sync.Mutex // held across interpret and commit
	turns    StateTurnPool
	state    canonicalState
	registry SurfaceRegistry
}

// StateTurnPool queues inputs to interpretation, not already-prepared Requests.
// Start with one worker. Increasing this count needs a concurrency design;
// it is not implied by having multiple external execution workers.
type StateTurnPool struct {
	inbound chan StateTurnInput
	workers int // initial value: 1
}

type canonicalState struct {
	possibilities map[uuid.UUID]Possibility // each owns its invocation frames
	moves         map[uuid.UUID]Move        // includes FirstMoveID targets
	executions    map[uuid.UUID]ExecutionRecord
}

// StateTurnInput must carry a Move and trusted runtime-bound origin/context;
// its exact declaration is open. ExecutionRecord's key and relationship to
// invocation identity, attempts, and fds are also open, not newly equated here.
// Request is an internal prepared interpret-to-commit artifact, not a second
// public invocation protocol. StateOwner's public requests and outcomes are Moves;
// emitted Moves return through runtime interpretation and authority.
// Restricted profiles and method signatures are the next discussion step.
