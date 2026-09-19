// Package dna17 is a throwaway, exploratory generation of the Skyra runtime that
// implements the 2026-09-18 "collapse": one smart component (the Interpreter),
// one dumb serializing writer (the StateExecutor / StateOwner), and a worker
// pool, driven by a single inbound queue. The inner loop is read -> commit ->
// execute, with decode as the entry (the opcode selects a handler before any
// read), and the commit mints an fd-style Identifier per authorized execution.
//
// This file declares the structs, interfaces, and ownership comments. Bodies live
// in sibling files (state.go, interpret.go, dispatch.go, cu.go, codex.go,
// probe.go), per DNA-16's declare-before-implement discipline.
//
// Provenance / boundary: this is greenfield from the 2026-09-18 specs. It imports
// NO skyra-dna-16 package. Codex protocol details are cribbed and reimplemented in
// codex.go and never appear in the core loop or the state contracts below.
package dna17

import "context"

// ---------------------------------------------------------------------------
// Identity leaf types.
//
// Deviation from the specs: the possibility/move proposals key identity on
// uuid.UUID. This build uses string newtypes instead — deterministic IDs make the
// acceptance tests reproducible without a network dep or go.mod churn, and the
// distinct named types still prevent accidental cross-assignment. See DECISIONS.md.
// ---------------------------------------------------------------------------

type PossibilityID string
type InvocationID string
type SurfaceID string
type MoveID string

// Identifier is the fd-style token minted by the commit for one authorized
// execution. It is opaque on the wire: only the bare number travels, and scope /
// liveness live in the StateExecutor's table (see fdEntry in state.go). The
// adapter cannot fabricate one for another execution. The zero value is never a
// live fd.
type Identifier uint64

// ---------------------------------------------------------------------------
// Move / Opcode — the protocol transition between Surfaces.
//
// The opcode is which pointer is non-nil (Invoke | Return | Fault). Fault is a
// payload variant that re-enters the one queue, not a third protocol opcode; it
// shares source correlation with Return. Operations (below) are NOT Moves.
// ---------------------------------------------------------------------------

type Move struct {
	Identity      MoveID
	PossibilityID PossibilityID
	Source        MoveSource // who emitted this Move (zero for external Submit)
	Target        SurfaceID  // for Invoke: the Surface to bind; for Return/Fault: unused
	Invoke        *Invoke
	Return        *Return
	Fault         *Fault
}

// MoveSource records the emitting Surface/execution. It is bound by execute; a
// Move cannot forge its own source.
type MoveSource struct {
	SurfaceID    SurfaceID
	InvocationID InvocationID
}

type Invoke struct {
	InvocationID InvocationID
	Action       string
	Params       []Observation
	Focus        []Capability // compiled into the Surface's capability offer; must be within Horizon
	ReturnTo     InvocationID // parent invocation to answer; empty means the root result boundary
}

type Return struct {
	Answering   InvocationID // the invocation being answered; correlated against the recorded invocation
	Observation Observation
}

// Fault is the failure-report payload variant of a Move. It carries the faulted
// invocation and a reason; the interpreter settles it (root result or parent
// propagation). Trap owns any corrective-work decision.
type Fault struct {
	Invocation InvocationID
	Reason     string
}

// Observation is the canonical value leaf. Key names a segment author (used by the
// Codex driver to key tool arguments); Bytes is the content.
type Observation struct {
	Key   string
	Bytes []byte
}

// Capability names an endpoint permission: a Surface identity plus an action. It
// is not a live handler. Discovery of a Capability is not authority to use it.
type Capability struct {
	Target SurfaceID
	Action string
}

// ---------------------------------------------------------------------------
// Possibility / Horizon — pure identity-and-authority DATA.
//
// No Dispatch/Resolve methods. Authority check + Surface binding + find-or-mint of
// an execution happen in ONE StateExecutor transaction (see opAuthorizeStart).
// ---------------------------------------------------------------------------

type Possibility struct {
	Identity    PossibilityID
	FirstMoveID MoveID
	Horizon     Horizon
}

type Horizon struct {
	Generation uint64
	Grants     map[SurfaceID]Grant
}

// Grant authorizes dispatch to one Surface identity. Generation advances only when
// the grant set changes; it is distinct from any global writer revision.
type Grant struct {
	SurfaceID SurfaceID
}

// ---------------------------------------------------------------------------
// Surface / Driver — the execution boundary.
//
// A Surface owns its own native state; to touch canonical state it must request it
// through the fd-scoped ExecutionState (stateVia), never an ambient alias.
//
// The Driver shape (Ingress with a persistent output callback + Egress) is cribbed
// from DNA-16 because a real provider session (Codex thread/turn) is inherently
// long-lived and asynchronous: one Ingress returns quickly but the driver keeps
// the callback and emits many times over the invocation's life. See DECISIONS.md
// for how this strains the "identifier IS the execution" model.
// ---------------------------------------------------------------------------

type Surface interface {
	// Ingress pushes one execution step (initial invoke, or a resume carrying a
	// child result) into the driver session and returns. Asynchronous emissions
	// arrive through output, which the worker binds to the execution's fd. output
	// must never be called while the writer lock is held.
	Ingress(ctx context.Context, input DriverInput, output func(Emission, error) error) error

	// Egress is driver-owned normalization of an emission's opaque Raw payload into
	// protocol Moves. It performs no state access.
	Egress(raw string) ([]Move, error)
}

// CommandExecutor runs opaque host commands, and only after their transition has
// committed (commit-before-send).
type CommandExecutor interface {
	Execute(ctx context.Context, command HostCommand) error
}

// Transport owns framed message movement only — no protocol meaning. The
// acceptance tests use an in-process fake transport implementing this.
type Transport interface {
	Send(ctx context.Context, frame []byte) error
	Receive(ctx context.Context) ([]byte, error)
	Close() error
}

// DriverState is StateExecutor-owned storage with driver-owned meaning. Version
// zero + empty payload means "no prior state". Drivers receive detached copies and
// return complete replacements — never a pointer into canonical state.
type DriverState struct {
	Version uint32
	Payload []byte
}

// DriverInput is the detached view delivered for one driver step. ChildResult is
// present only on a resume step (the observation from the satisfied child).
type DriverInput struct {
	Move        Move
	State       DriverState
	ChildResult *Observation
}

// Emission is one asynchronous output from a driver step. Exactly one of Raw /
// Failure is meaningful for protocol output; State always carries the replacement
// driver state. Commands are opaque post-commit host instructions.
type Emission struct {
	State    DriverState
	Raw      string
	Commands []HostCommand
	Failure  error
}

type HostCommand struct {
	ID      string
	Payload []byte
}

// ---------------------------------------------------------------------------
// Operations — a closed vocabulary over canonical state, applied by the single
// writer. Operations are NOT Moves: Moves are protocol transitions between
// Surfaces; operations are requested changes over records. Request is the
// branch-free program the interpreter hands the writer.
//
// This is the "system plane": operations issued by the interpreter at commit under
// its role. The "execution plane" (operations a Surface issues during execute) is
// carried by EmissionCommit through the fd-scoped ExecutionState below.
// ---------------------------------------------------------------------------

type Operation interface{ operation() }

type Request struct {
	Move       Move
	Operations []Operation
}

// opRecordInvocation records a new invocation continuation. It authorizes nothing;
// phase is DECIDED by interpret — Running when authorized, Faulted when trapped — so
// the commit never has to write a phase and then walk it back.
type opRecordInvocation struct {
	inv    Invoke
	target SurfaceID
	poss   PossibilityID
	source MoveSource
	phase  Phase
}

// opAuthorizeStart mints the fd and the dispatch for an invocation interpret has
// ALREADY authorized. The authority read happens in interpret (the read phase); this
// op does not re-check — it only mints. Because interpret and commit are one locked
// turn, check+mint stay atomic without deferring the read into the write phase.
type opAuthorizeStart struct {
	inv InvocationID
}

// opEmitFault collects a corrective Fault Move into the commit result. The writer
// mints its id; the caller enqueues it through the one door AFTER the lock releases.
type opEmitFault struct {
	poss   PossibilityID
	inv    InvocationID
	reason string
}

// opMarkReturned marks an invocation returned, records its result, and kills its fd.
type opMarkReturned struct {
	inv InvocationID
	obs Observation
}

// opRootResult records the single root result for a root invocation.
type opRootResult struct {
	inv InvocationID
	obs Observation
}

// opResumeParent resumes a live parent with a child observation: reuse its fd and
// dispatch it. Interpret already verified the parent is live; this op only writes.
type opResumeParent struct {
	parent InvocationID
	child  Observation
}

// opMarkFaulted marks an invocation faulted and kills its fd.
type opMarkFaulted struct {
	inv InvocationID
}

// opRootFault records a root fault result for a root invocation.
type opRootFault struct {
	inv    InvocationID
	reason string
}

func (opRecordInvocation) operation() {}
func (opAuthorizeStart) operation()   {}
func (opEmitFault) operation()        {}
func (opMarkReturned) operation()     {}
func (opRootResult) operation()       {}
func (opResumeParent) operation()     {}
func (opMarkFaulted) operation()      {}
func (opRootFault) operation()        {}

// ---------------------------------------------------------------------------
// Execution-plane state access — stateVia(id).
//
// The fd is the only authority a worker holds. ExecutionState is the per-execution
// profile the fd grants: it can replace THIS execution's driver state and emit
// Moves, checked live against the fd at each request. A settled/dead fd rejects
// every request (the settlement fence). This is how a late or duplicate emission
// after a fault cannot produce a Skyra Return.
// ---------------------------------------------------------------------------

type ExecutionState interface {
	Commit(ctx context.Context, ec EmissionCommit) (EmissionResult, error)
}

// EmissionCommit is one execution-plane transaction: replace the fd's invocation
// driver state, enqueue any emitted Moves (child ids minted by the writer), or
// record a failure. State replacement and emitted effects commit together.
type EmissionCommit struct {
	State   DriverState
	Moves   []Move
	Failure error
}

type EmissionResult struct {
	Enqueue []Move
}

// ---------------------------------------------------------------------------
// Commit result / dispatch / root result.
// ---------------------------------------------------------------------------

// dispatchItem is one authorized execution: an fd plus the detached input to hand
// the bound Surface. A commit may authorize several (the current move plus any it
// unblocked); they fire concurrently.
type dispatchItem struct {
	id      Identifier
	surface SurfaceID
	input   DriverInput
}

type commitResult struct {
	authorized  []dispatchItem
	enqueue     []Move
	rootResults []RootResult
}

// RootResult is the one outcome delivered per root invocation: an observation or a
// fault. Exactly one per root.
type RootResult struct {
	Invocation  InvocationID
	Observation *Observation
	Err         error
}

// ---------------------------------------------------------------------------
// Restricted StateOwner profiles.
//
// Components receive typed profiles, never general read/write or map aliases.
// PossibilityState opens runs; SystemState is the interpreter's commit door;
// ExecutionState (above) is the fd-scoped execution door.
// ---------------------------------------------------------------------------

type PossibilityState interface {
	Open(ctx context.Context, request OpenPossibility) error
}

// OpenPossibility opens a run: the Possibility record and its first Move commit
// together or neither commits.
type OpenPossibility struct {
	Identity PossibilityID
	First    Move
	Horizon  Horizon
}

// SystemState is the interpreter's write door: apply one Request atomically.
type SystemState interface {
	Commit(ctx context.Context, request Request) (commitResult, error)
}
