package scheduler

import (
	"context"
	"skyra-v1/src/skyra/skyra-dna-16/controller"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
	"sync"
)

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
	Move        protocol.Move
	Handoff     Handoff
	State       protocol.State
	DriverState protocol.DriverState
	Context     context.Context
	Result      *protocol.Observation
	// Nil means no compiled expectations, not contract failure or satisfaction.
	// The declaration-path decision remains separate from interpretation.
	CompiledRegister *CompiledRegister
	// Reserve one parent resumption in the local single-wave probe. The parent
	// resumes its existing Invoke with an updated register, not a new Return
	// invocation. Repeated fan-out waves remain outside this proposal.
	Resumed bool
}

// Scheduler owns ALL live continuation state. The mutex protects admission,
// register updates, result recording, and selection. Controller is injected at
// startup. Complete is a host callback, called outside mu, not another state store.
type Scheduler struct {
	mu            sync.Mutex
	controller    controller.Controller
	continuations map[protocol.InvocationID]Continuation
	complete      func(protocol.RootResult) error
}

// W0: resident placement state is separate from semantic Register. Sequence starts
// at zero and increments only for the single parent resumption.
type Placement uint8

const (
	PlacementPending Placement = iota
	PlacementSucceeded
	PlacementFailed
)

type Handoff struct {
	Sequence  uint64
	Placement Placement
	// Rejection records the consumed batch rejection; the workflow faults once, without retry.
	Rejection error
}
