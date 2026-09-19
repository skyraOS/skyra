// Package protocol defines the shared Invoke/Return wire values and scheduler events.
package protocol

import (
	"context"
	"crypto/rand"
	"encoding/hex"

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

func (Invoke) opcode() {}

func (Return) opcode() {}

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

// CompiledRegister represents compiled expectations owned by Scheduler alongside
// live continuation state. Its contents and continuation-instruction representation
// remain OPEN; this placeholder must not be treated as an empty satisfied contract.
// It is not a second live register and does not travel with Invoke.
type CompiledRegister struct{}

// Assessment describes contract interpretation, not dispatch lifecycle. Missing
// observations mean Waiting, not Mismatch. A boolean contract matcher can be stubbed
// after required inputs are present. Zero is unspecified, never Satisfied.
type Assessment uint8

const (
	Waiting Assessment = iota + 1
	Satisfied
	Mismatch
)

// Context is the invocation's execution lifetime, inherited by its children. A
// sibling's or returning transport's context never replaces the parent's context.
// Result marks the source invocation as answered, independently of its dispatch
// State. This is what makes late acknowledgements and duplicate returns harmless.
// Result is execution bookkeeping, not the engram's current work artifact.
type Continuation struct {
	Move    Move
	State   State
	Context context.Context
	Result  *Observation
	// Nil means no compiled expectations, not contract failure or satisfaction.
	// The declaration-path decision remains separate from interpretation.
	CompiledRegister *CompiledRegister
	// Reserve one parent resumption in the local single-wave probe. The parent
	// resumes its existing Invoke with an updated register, not a new Return
	// invocation. Repeated fan-out waves remain outside this proposal.
	Resumed bool
}

// OpcodeEvent is one write protocol with distinct payloads, not one ambiguous
// bag of fields. MoveBatch is validated admission; DispatchResult acknowledges
// access/handoff; failures cannot masquerade as invokable Moves.
type OpcodeEvent interface{ opcodeEvent() }

// Source is nil only for CTU.Submit. For driver output it is bound by execute.
// Admit the entire batch under one lock BEFORE selecting any child. The first
// callback can otherwise complete a child before a sibling has been recorded.
// Local probe: one emission is either a batch of Invokes or exactly one Return, not a
// mixture. Later streaming fan-out needs a separate declaration-closure decision.
type MoveBatch struct {
	Source *InvocationID
	Moves  []Move
}

func (MoveBatch) opcodeEvent() {}

type DispatchResult struct {
	Invocation InvocationID
	Authorized bool
	Err        error
}

func (DispatchResult) opcodeEvent() {}

// AdmissionFailure describes an external rejected submission. It must not mutate
// a resident entry merely because the input claimed that entry's InvocationID.
type AdmissionFailure struct {
	Move Move
	Err  error
}

func (AdmissionFailure) opcodeEvent() {}

// ExecutionFailure is bound to the known source by execute: transport, parse, or
// normalized-output validation failed. It needs no invented child Move/ID. A
// callback reporting error must not be mistaken for a second execution failure.
// Automatic retry is outside the local probe; fault the source and report failure.
type ExecutionFailure struct {
	Source InvocationID
	Err    error
}

func (ExecutionFailure) opcodeEvent() {}

// RootResult reports a root return or fatal root/workflow execution failure.
// Exactly one notification attempt per root; callback errors propagate, with no
// automatic replay. Source Result/Faulted bookkeeping prevents duplicate attempts.
// This is root outcome, not global process quiescence or remote cancellation proof.
type RootResult struct {
	Invocation  InvocationID
	Observation *Observation
	Err         error
}

// crypto/rand 16 bytes encoded as lowercase hex; random failure propagates. No
// package-global counter or scheduler side table. Called for new Invoke only;
// Return uses the captured invocation as its target, without allocating an ID.
func NewInvocationID() (InvocationID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return InvocationID{}, err
	}
	return InvocationID{Value: hex.EncodeToString(b[:])}, nil
}
