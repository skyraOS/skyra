//go:build ignore

// Package supervisor contains proposed contracts, not a runtime implementation.
// Move, Component, and CustodyState belong to the sibling Move and custody
// proposals. WorkerPool, PreparedExecution, and worker helpers remain provisional.
// SupervisorComponent denotes custody.Supervisor in the eventual package wiring.
package supervisor

import (
	"context"

	"github.com/google/uuid"
)

// Supervisor owns dispatch, bounded concurrent workers, and independent
// monitoring. StateOwner owns canonical worker and outcome records.
type Supervisor struct {
	state   SupervisorState
	custody CustodyState

	self Component
	next Component

	workers *WorkerPool
}

// New stores boot-injected restricted adapters and live worker machinery.
// Boot binds the adapters to Supervisor with transfer permission to
// Interpretation. Construction starts no processing or goroutines.
func New(state SupervisorState, custody CustodyState, workers *WorkerPool) *Supervisor {
	return &Supervisor{
		state:   state,
		custody: custody,
		self:    SupervisorComponent,
		next:    Interpretation,
		workers: workers,
	}
}

// Run is the conceptual capacity-and-admission loop. Boot starts it after wiring.
// Shutdown coordination with live workers remains to be specified.
func (s *Supervisor) Run(ctx context.Context) error {
	for {
		// Naming/decomposition remains provisional: reconcile Reserve here with
		// PrepareExecution below when defining the atomic profile operations.
		slot, err := s.workers.Reserve(ctx)
		if err != nil {
			return err
		}

		move, err := s.state.Take(ctx, s.self)
		if err != nil {
			slot.Release()
			return err
		}

		// startWorker takes responsibility for the reservation on entry. It
		// prepares/registers execution before launch and releases the slot if
		// no worker starts. Otherwise the pool retains it until actual exit.
		if err := s.startWorker(ctx, slot, move); err != nil {
			return err
		}
	}
}

// SupervisorState is the restricted StateOwner profile injected at boot.
//
// Design note: profile methods should be decomposed forms of the operations in
// the loop, stripped to their bare atomic state operations. Choose boundaries by
// asking which related changes must commit together. Keep execution and waits
// outside StateOwner's serialized turn.
//
// Reserve and PrepareExecution currently have overlapping conceptual names.
// This is acceptable for the sketch, but the final loop and profile method names
// must match for the same operation. If execution preparation is called Reserve,
// use Reserve consistently in the loop and profile. Reconcile the decomposition
// rather than preserving differently named wrappers for the same operation.
// Currently pool Reserve waits for local capacity before Take; PrepareExecution
// below authorizes and registers a particular Move after Take. Preserve those
// ordering requirements while settling the atomic operations and their names.
type SupervisorState interface {
	Take(ctx context.Context, self Component) (Move, error)

	// Atomically checks authority, resolves the Surface/contract binding, and
	// finds or registers the execution. Existing work must not launch again.
	PrepareExecution(
		ctx context.Context,
		request PrepareExecution,
	) (PreparedExecution, error)

	// Typed atomic operations for acceptance, progress, retained outcomes,
	// monitoring events, and worker exit remain to be defined.
}

// StateOwner derives Possibility, target, and contract from the canonical Move.
type PrepareExecution struct {
	MoveID uuid.UUID
}

// PreparedExecution must distinguish existing execution from new registration
// and carry an opaque authorized binding. Exact fields await execution identity
// and binding contracts; do not substitute any or expose a raw *Surface.
