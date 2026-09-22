//go:build ignore

// Package tu contains proposed contracts, not a runtime implementation.
// Move belongs to ../move/proposal.go. Operation variants and package wiring
// remain provisional; Operation must be a closed set of typed system operations.
package tu

import (
	"context"

	"github.com/google/uuid"
)

type Component uint8

const (
	Interpretation Component = iota + 1
	Supervisor
)

// CustodyRecord answers only which component currently owns the Move.
// State owns storage and all reads and writes; custody has no separate read API.
type CustodyRecord struct {
	MoveID    uuid.UUID
	Component Component
}

// Request carries a Move and the current stage's pending runtime operations.
// Each stage submits its own operations through its restricted profile; already
// committed operations are not carried forward as pending work for the next stage.
// Application dependency semantics belong to Interpreter, not Operation.
// Handoff transfers exclusive ownership or copies all mutable data, including
// nested Move payloads and operation arguments; passing this struct by value
// alone does not detach its slices or pointers.
type Request struct {
	Move       Move
	Operations []Operation
}

// TransitionUnit is a restricted StateOwner profile, not a separate state owner
// or runtime service. Interpreter and Supervisor propose typed system operations
// with respect to a required Move; StateOwner validates and commits them together.
// Boot binds private adapters to component identities and permitted operations.
// TU is stateless: it forwards the whole Request and bound routing context through
// MessageBroker's request mailbox. StateOwner holds canonical transition records.
//
// Validation of revision, authority, canonical Move identity, correlation, and
// ownership occurs in the same serialized turn as the commit. Related changes,
// including continuation records, custody, and mailbox placement, are atomic.
// External execution and waits remain outside that transaction.
//
// from and to describe the permitted pipeline route, injected at boot and
// validated against the caller's bound profile. Interpreter forwards only after
// the carried Move's application dependencies are satisfied and its runtime
// operations are committed together with Supervisor mailbox placement and custody.
// Supervisor commits its own execution preparation before releasing execution.
// Each stage is one transaction: all required runtime operations commit before
// work advances. There is no transaction or staging area spanning both stages.
// StateOwner validates the legal transition against the canonical record for the
// current processing attempt. Execution registration requires that attempt's
// committed Interpreter handoff; an old handoff cannot authorize new execution.
// Exact processing-attempt and execution-release records remain to be declared.
//
// Introducing an emitted Move is distinct from transferring an existing Move:
// new Moves have no prior owner to validate. The operation variants must express
// this distinction and preserve source execution correlation. Exact introduction
// semantics and operation fields remain to be declared.
//
// At each stage, success acknowledges commit, not merely request enqueueing.
// A stale revision at commit is a known non-commit. A lost or
// cancelled acknowledgment does not prove non-commit. Commit identity and
// acknowledgment reconciliation belong behind Transition, in the TU/StateOwner
// request protocol; individual services neither define nor supply commit IDs.
// Repeating the same committed request must not apply its changes again. Move
// identity alone cannot identify a commit because a Move may cause multiple
// transitions. Exact internal request correlation remains to be defined.
//
// Snapshot-dependent operations must retain their original revision assumptions
// through the profile/request protocol. Reading the latest revision at commit
// time cannot establish freshness. That internal contract remains open; services
// do not supply a separate CommitMetadata argument.
type TransitionUnit interface {
	Transition(
		ctx context.Context,
		request Request,
		from Component,
		to Component,
	) error
}
