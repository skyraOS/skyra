//go:build ignore

// Package surface contains proposed contracts, not a runtime implementation.
// Historical starting point: the later engrams-surfaces-and-pipe-protocol.md
// refinement defines Surface as an addressable executable endpoint and Engram
// as composition. External executables speak the protocol over pipes; this Go
// interface is not their required implementation contract. Fields below await revision.
// Structs come from specs/2026-09-19/surface.md. Driver is retained from the
// September 18 proposal; DNA-17's different interface is in ../../authority-owner/reference.
// Profile, Observation, Move, and DriverEmission await package/type wiring.
package surface

import (
	"context"

	"github.com/google/uuid"
)

type Surface struct {
	Identity       uuid.UUID
	Contracts      []ContractPair
	Dependencies   []*Surface
	Driver         Driver
	IngressProfile Profile
	EgressProfile  Profile
}

type ContractPair struct {
	SurfaceID uuid.UUID
	Input     InputContract
	Output    OutputContract
}

type InputContract struct {
	InputID      uuid.UUID
	Observations []Observation
}

type OutputContract struct {
	OutputID    uuid.UUID
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

// Profile's concrete signatures remain open in the September 19 proposal.
// Each profile has interpret (permitted reads and preparation) and commit
// (prepared writes). Ingress and egress run as separate state turns.
