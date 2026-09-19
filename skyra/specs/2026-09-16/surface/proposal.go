//go:build ignore

// Package surface contains proposed contracts, not a runtime implementation.
// Observation, Move, DriverEmission, and StateKey await their owner proposals.
package surface

import (
	"context"

	"github.com/google/uuid"
)

type Surface struct {
	Identity     uuid.UUID
	Contracts    []ContractPair
	Capabilities []*Surface
	Driver       Driver
	State        StateReader
}

type ContractPair struct {
	Identity uuid.UUID
	Input    InputContract
	Output   OutputContract
}

type InputContract struct {
	SurfaceID    uuid.UUID
	Observations []Observation
}

type OutputContract struct {
	SurfaceID   uuid.UUID
	Observation Observation
}

type Driver interface {
	Ingress(
		ctx context.Context,
		move Move,
		emit func(DriverEmission) error,
	) error

	Egress(emission DriverEmission) Move
}

// StateReader: ownership boundaries are uncertain; we will revisit them.
type StateReader interface {
	Snapshot(ctx context.Context) (map[StateKey]Observation, error)
}
