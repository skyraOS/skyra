package protocol

import (
v13 "skyra-v1/src/skyra/skyra-dna-v13"
)

// Reuse settled leaf values, not v13's blocked component implementations.
// Engram identity remains a durable-storage concern, outside the loop wire.
type EngramID = v13.EngramID

type InvocationID = v13.InvocationID

type PossibilityID = v13.PossibilityID

type Observation = v13.Observation

type Action = v13.Action

type EntityID = v13.EntityID

type HostID = v13.HostID

type Contract = v13.Contract

type State = v13.State

const (
	Dispatched = v13.Dispatched
	InFlight   = v13.InFlight
	Faulted    = v13.Faulted
)

// Move is a transition envelope, not necessarily a new invocation. Invoke starts
// work; Return contributes the observation from existing work. This supersedes
// the earlier rule that every Move, including a return, had a fresh InvocationID.
type Move struct {
	Root   PossibilityID
	Target Target
	Op     Opcode
	Stamp  AuthorizationStamp
}

// For Invoke, endpoint fields route execution; Invocation may name an existing
// addressed continuation, or be zero for an endpoint-only destination. It is not
// a substitute for Invoke.InvocationID. For Return, Invocation MUST identify the
// invocation being answered. CC binds it from execute's captured source identity;
// endpoint fields are not a request to execute that driver again.
type Target struct {
	Invocation InvocationID
	Entity     EntityID
	Address    string
	Host       HostID
}

type Opcode interface{ opcode() }

// Invoke carries all execution inputs and authority. Scheduler owns its resident
// Register and sends a detached snapshot; the driver adapts it to its endpoint.
// ReturnFd names the parent receiving this invocation's result. Zero denotes the
// host/root result boundary in the local proposal. No MoveID exists on the wire.
type Invoke struct {
	InvocationID InvocationID
	Action       Action
	Params       []Observation
	ReturnFd     InvocationID
	Register     Register
	Horizon      Horizon
	Focus        []Capability
}

// Return carries only the resulting observation. No fresh ID, InReplyTo, register,
// grant, or parent route is needed: Move.Target.Invocation locates the source and
// the stored Invoke.ReturnFd locates its parent. Arbitrary external Returns are
// not admitted; execute establishes their source-bound route.
type Return struct {
	Observation Observation
}

// Capability target remains an endpoint permission, rather than an exact live
// invocation identity. Keeping that distinction avoids accidentally making a
// newly allocated InvocationID invalidate otherwise matching endpoint grants.
type Capability struct {
	Target    Target
	Action    Action
	ExpiresAt int64
	Sig       []byte
}

type Horizon struct{ Capabilities []Capability }

// Register remains invocation -> invocation/observation. Order explicitly retains
// child declaration order, which the v13 map alone lost. No lifecycle lives here.
type Register struct {
	Contract Contract
	Children map[InvocationID]RegisterEntry
	Order    []InvocationID
}

type RegisterEntry struct {
	Invocation  ChildInvocation
	Observation *ChildObservation
}

// The recorded child Invoke has an empty Register; its authoritative frame lives on its
// independent continuation's Move. This avoids recursively embedding snapshots.
type ChildInvocation struct {
	InvocationID InvocationID
	Move         Move
}

type ChildObservation struct {
	InvocationID InvocationID
	Observation  Observation
}

// OpcodeEvent transports two families: MoveBatch continuation output/admission
// and runtime placement/execution feedback. Envelope composes them runtime-first.
// AdmissionFailure carries a local processing error, not an execution fault.
type OpcodeEvent interface{ opcodeEvent() }

// Source is nil only for CTU.Submit. For driver output it is bound by execute.
// Admit the entire batch under one lock BEFORE selecting any child. The first
// callback can otherwise complete a child before a sibling has been recorded.
// Local probe: one emission is either a batch of Invokes or exactly one Return, not a
// mixture. Later streaming fan-out needs a separate declaration-closure decision.
type MoveBatch struct {
	Source   *InvocationID
	Sequence uint64
	Moves    []Move
}

type DispatchResult struct {
	Invocation InvocationID
	Sequence   uint64
	Err        error
}

// AdmissionFailure describes an external rejected submission. It must not mutate
// a resident entry merely because the input claimed that entry's InvocationID.
type AdmissionFailure struct {
	Move Move
	Err  error
}

// ExecutionFailure is bound to the known source by execute: transport, parse, or
// normalized-output validation failed. It needs no invented child Move/ID. A
// callback reporting error must not be mistaken for a second execution failure.
// Current-handoff faults fault active workflows, including after a source Return.
// Older-handoff and terminal-workflow feedback cannot reopen work.
type ExecutionFailure struct {
	Source   InvocationID
	Sequence uint64
	Err      error
}

// RootResult reports a root return or fatal root/workflow execution failure.
// Exactly one notification attempt per root; callback errors propagate, with no
// automatic replay. Source Result/Faulted bookkeeping prevents duplicate attempts.
// This is root outcome, not global process quiescence or remote cancellation proof.
type RootResult struct {
	Invocation  InvocationID
	Observation *Observation
	Err         error
}

// Runtime-produced stamp; inbound stamps are overwritten. CheckedAt is Unix nano.
// The local bounded policy matches endpoint/action grants, with Unix-second expiry;
// signatures are trusted fixture data, not production cryptographic verification.
type AuthorizationStamp struct {
	Allowed   bool
	CheckedAt int64
}

// Envelope contains the two event families; runtime is applied first. Runtime must
// be DispatchResult or ExecutionFailure, never another envelope or MoveBatch.
type Envelope struct {
	Runtime      OpcodeEvent
	Continuation *MoveBatch
}

