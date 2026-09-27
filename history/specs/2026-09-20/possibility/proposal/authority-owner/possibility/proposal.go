//go:build ignore

// Package possibility contains discussion declarations, not runtime code.
// Horizon is in ../horizon; supporting value types are in the reference snapshot.
// Cross-directory package wiring and identity conversions remain provisional.
package possibility

import "github.com/google/uuid"

// Possibility starts from the September 18 identity-and-authority record.
// NEW: the invocation map belongs logically to this Possibility. StateOwner
// still owns canonical storage and writes; this map is not exposed by alias.
type Possibility struct {
	Identity    uuid.UUID
	FirstMoveID uuid.UUID
	Horizon     Horizon
	Invocations map[uuid.UUID]InvocationFrame
}

type PossibilityMap map[uuid.UUID]Possibility

// InvocationFrame adapts DNA-17's invocationRecord into a Possibility-local
// frame. ReturnTo is local to this Possibility; nil denotes the root boundary.
// The map is an index; ReturnTo supplies the call-stack relationship. With
// concurrent children, the structure branches instead of having one stack top.
type InvocationFrame struct {
	Identity    uuid.UUID
	SurfaceID   uuid.UUID
	Action      string
	Params      []Observation
	Focus       []Capability
	ReturnTo    uuid.UUID
	DriverState DriverState
	Phase       Phase
	Result      Observation

	// Carried from DNA-17 to expose what needs revision. WaitingOn currently
	// holds one child, and FD currently names a session retained across resumes.
	// Neither is the final multi-child/disposable-worker design.
	WaitingOn uuid.UUID
	FD        Identifier
}

type Phase string

const (
	PhaseRunning  Phase = "running"
	PhaseWaiting  Phase = "waiting"
	PhaseReturned Phase = "returned"
	PhaseFaulted  Phase = "faulted"
)
