//go:build ignore

// Package move contains proposed contracts, not a runtime implementation.
// Observation awaits its owner proposal. Interpretation belongs to the Interpreter.
package move

import "github.com/google/uuid"

type Move struct {
	Identity      uuid.UUID
	PossibilityID uuid.UUID
	Source        MoveSource
	Target        uuid.UUID
	Invoke        *Invoke
	Return        *Return
}

type MoveSource struct {
	SurfaceID    uuid.UUID
	InvocationID uuid.UUID
}

type Invoke struct {
	InvocationID uuid.UUID
	ContractID   uuid.UUID
	Observations []Observation
}

type Return struct {
	ReturnID    uuid.UUID
	ContractID  uuid.UUID
	Observation Observation
}
