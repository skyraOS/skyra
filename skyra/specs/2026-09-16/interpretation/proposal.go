//go:build ignore

// Package interpretation contains proposed contracts, not a runtime implementation.
// Move is defined in ../move/proposal.go. Component and CustodyState are defined
// in ../tu/proposal.go; package wiring remains provisional.
package interpretation

import "context"

// Interpreter reads a Move, validates its Invoke/Return selection, and applies
// the corresponding semantics through separate internal handlers. It requests
// state changes; the State Owner validates and commits those requests.
// Canonical state and its input mailbox belong to State Owner. One consumer
// finishes each interpretation before taking the next Move.
// Move, Invoke, and Return remain data-only types.
type Interpreter struct {
	state   InterpretationState
	custody CustodyState

	self Component
	next Component
}

// New stores boot-injected dependencies and fixes the Interpreter's role.
// Construction performs no state reads and starts no goroutines. Boot supplies
// distinct private adapters bound to Interpretation, with transfer permission
// to Supervisor; StateOwner validates those bindings on every operation.
func New(state InterpretationState, custody CustodyState) *Interpreter {
	return &Interpreter{
		state:   state,
		custody: custody,
		self:    Interpretation,
		next:    Supervisor,
	}
}

// Run processes one Move and completes its handoff before taking another.
// Boot starts Run once after wiring the components and supervises its return.
func (i *Interpreter) Run(ctx context.Context) error {
	for {
		move, err := i.state.Take(ctx, i.self)
		if err != nil {
			return err
		}

		if err := i.interpretAndTransfer(ctx, move); err != nil {
			return err
		}
	}
}

// InterpretationState is the restricted profile injected at boot. This method
// set remains incomplete until the continuation records define the typed reads
// needed by Invoke/Return interpretation.
type InterpretationState interface {
	Take(ctx context.Context, self Component) (Move, error)
}

// interpretAndTransfer remains the helper described in ../ctu/minimal-loop.md:
// read detached state, interpret or normalize a fault, then Transfer the result.
// ErrStaleRevision means nothing committed: re-read and reinterpret the same
// Move. A committed handoff permits the next Take; other errors propagate.
// Result and the typed read contracts remain to be declared.
