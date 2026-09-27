//go:build ignore

// Package main is CODE ON PAPER for the root program's entrypoint and protocol.
// This is a separate executable boundary, not a new name for RuntimeState.
//
// Shared model names (Kernel, ExecutionContext, Move, Observation, ExecutableID,
// FrameID, Port, ...) refer to ../proposal.go. Core ownership structs and method
// declarations are repeated below for review; remaining types are not copied or
// imported from that build-ignored sketch. This file is not a standalone
// Go build; shared wire-package extraction and method bodies come later.
//
// The root executable runs within the root authority ExecutionContext. Bootstrap
// installs its initial authority; descendants receive delegated subsets. The root
// program is executable behavior, its ExecutionContext is execution state, and its
// Kernel coordinates that state and scheduled work. RuntimeState still contains
// only Memory and Active, with no root-context identity field.
package main

import "context"

// ---------------------------------------------------------------------------
// 1. Entrypoint and the bootstrap floor.
// ---------------------------------------------------------------------------

// main is the logical program entrypoint. It receives serialized startup arguments
// as strings, decodes them into observations inside the running program, acquires
// inherited port endpoints, and calls boot followed by Kernel.loop (provisional). A concrete Go build
// would use a thin func main() wrapper reading OS argv; this signature describes
// the startup contract rather than Go's required entrypoint signature.
// Arg encoding, launcher wiring and exit-status translation remain OPEN. The
// launcher must supply real authority-bearing connections, not merely claimed
// executable IDs in argv. No recursive provisioning is needed to start this floor.
func main(args []string)

// RootProgram holds startup configuration and bound endpoints; runtime work belongs
// to Kernel and its component owners. Its execution context is
// registered in kernel.runtimeState.active like other active contexts. No second
// context registry is introduced. Actual-port reservations stay on Active; queue
// and realization tracking stay on Scheduler.
type RootProgram struct {
	kernel    *Kernel
	execution *ExecutionContext
	ingress   []*Ingress
	egress    []*Egress
	boot      Bootstrap
}

// Bootstrap is decoded root-program input, NOT authority by itself. These IDs name
// executables already admitted by trusted bootstrap into the root context's Horizon.
// Live endpoints belong to RootProgram. Their roles are established by trusted
// endpoint binding, not duplicated as port identifiers in startup configuration.
// No particular numeric port (including 0) is assigned a universal meaning here.
type Bootstrap struct {
	Arguments   []Observation
	OS          ExecutableID // program providing launch/bind/lifecycle operations
	Preparation ExecutableID // persistent trap/preparation program
}

// Ingress and Egress are private, port-bound endpoints acquired at startup or
// established by trusted binding. Each represents exactly one port for its lifetime.
// Callers cannot choose or override a port on an individual Receive/Send call.
// Actual OS handles and framing stay private; this is not an interface foreign
// executable programs must implement.
//
// The endpoint binding must retain authenticated ownership and generation/liveness
// information. Every operation checks that binding; replacing a connection creates
// a new endpoint rather than silently retargeting an existing one. Concrete binding
// and transport storage remain OPEN; port alone is not sufficient authority.
type Ingress struct {
	port Port // fixed by trusted binding; not supplied in a received Move
	// OPEN: private transport and authenticated owner/liveness binding.
}

// Receive returns only the Move. Provenance comes from this endpoint's binding,
// never from a port field supplied by the sender. EOF/failure behavior remains OPEN.
func (*Ingress) Receive(ctx context.Context) (Move, error)

type Egress struct {
	port Port // fixed destination; a Move cannot redirect this endpoint
	// OPEN: private transport and authenticated owner/liveness binding.
}

func (*Egress) Send(ctx context.Context, move Move) error

// boot establishes the root program before accepting ordinary invocations:
//
//  1. Validate decoded observations against the root executable's argument contract
//     and attach the trusted endpoints, whose roles were established at binding.
//  2. Create Kernel{RuntimeState{Memory, Active}, Scheduler}; construct the root
//     ExecutionContext with its Bookkeeper and authorized bootstrap connections.
//  3. Publish that context and reserve actual ports through Active.provision.
//     Populate the context's local virtual mappings; no registry moves into Scheduler.
//  4. Invoke the OS executable, through its inherited port, to launch/bind the
//     preparation executable if it is not already live. Launch arguments are
//     observations; the root does not call a provider's Go method.
//  5. Await its correlated Return before declaring the preparation path ready.
//     Make that executable reachable when provisioning subsequent contexts.
//  6. Enable ordinary invocation ingress. A required boot failure produces an
//     outward Fault and unwinds resources acquired by this boot attempt.
//
// A minimal bootstrap event pump must handle launch responses before readiness;
// boot cannot block the only receiver while waiting for one of its own Returns.
// Exact ready/rollback observations, executable IDs and installation policy are OPEN.
func boot(ctx context.Context, args []Observation, ingress []*Ingress, egress []*Egress) (*RootProgram, error)

// ---------------------------------------------------------------------------
// 2. Runtime owners and their interfaces — no RootProgram run/accept/step loop.
//
// These are the shared proposal's ownership structs made visible at the entrypoint.
// They describe the same objects, not additional state stores or executable boundaries.
// ---------------------------------------------------------------------------

type Kernel struct {
	runtimeState *RuntimeState
	scheduler    *Scheduler
}

type RuntimeState struct {
	memory *Memory
	active *Active
}

type Active struct {
	executionContexts map[ExecutionContextID]*ExecutionContext
	portReservations  map[KernelPort]ExecutionContextID
}

type Scheduler struct {
	queue     []Move
	realizing map[ExecutionContextID]bool
}

// loop is only the current place to sketch the main loop; its final owner is OPEN.
// It connects bound ingress Receive/admission with Scheduler.Run. The scheduler
// owns taking queued work and tracking realization; loop must not repeat that work.
// How receives and multiple context turns progress concurrently is also OPEN.
func (*Kernel) loop(ctx context.Context) error

// admit owns routing/admission, not the root program. The source is the bound
// endpoint used by the trusted receive pump, never a sender-selected port. Kernel
// uses Active.ownerOf on the actual connection behind source, then coordinates
// the context's binding fence/emission commit before Scheduler.Enqueue. Payload
// identity alone cannot authorize work. Exact external admission policy is OPEN.
func (*Kernel) admit(ctx context.Context, source *Ingress, move Move) error

// Active owns context and reservation lookup. Callers receive no writable map aliases.
func (*Active) ownerOf(port KernelPort) (ExecutionContextID, bool)
func (*Active) lookup(id ExecutionContextID) (*ExecutionContext, bool)

// provision owns publication and port reservation; authority preparation precedes
// this operation. Conflicts fail without partial publication.
func (*Active) provision(ctx context.Context, execution *ExecutionContext, ports []KernelPort) error

// Enqueue accepts admitted Moves and retains detached copies. It does not resolve
// runtime state or execute programs.
func (*Scheduler) Enqueue(move Move) error

// DispatchMove is a narrow function supplied by runtime wiring. It resolves the
// selected context through Active.lookup and calls ExecutionContext.Realize; it
// rejects an inactive/missing context without provisioning it. It carries no queue
// policy. Scheduler needs no pointer to Kernel or RuntimeState.
type DispatchMove func(ctx context.Context, move Move) error

// Run takes one eligible queued Move, claims realizing[owner], calls dispatch, and
// releases the claim on success or failure. It waits when nothing is eligible and
// preserves admission order per context while skipping busy contexts. Completion
// means the Realize turn returned, not that the external provider finished.
func (*Scheduler) Run(ctx context.Context, dispatch DispatchMove) error

// Realize owns interpretation, Bookkeeper dependency tracing, commit and dispatch
// against its Horizon. It returns after dispatch rather than waiting for external
// completion. New work/returns arrive through bound ingress and re-enter the queue.
func (*ExecutionContext) Realize(ctx context.Context, move Move) error

// Bookkeeper owns the cactus of continuations and parent/child correlation.
func (*Bookkeeper) traceDependencies(move Move) error

// The concrete flow through those owners is:
//   Ingress.Receive -> Kernel.admit -> Active.ownerOf + context emission fence
//   -> Scheduler.Enqueue -> Scheduler.Run -> supplied dispatch -> Active.lookup
//   -> ExecutionContext.Realize -> Bookkeeper.traceDependencies -> commit/dispatch
//   -> return to Scheduler.Run, which releases its claim. Executable output returns
//   through Ingress.Receive. Kernel.loop is the provisional coordinator, not a
//   decision about final main-loop ownership.
//
// Missing preparation is another executable invocation. Child work parks a recorded
// continuation; a correlated Return resumes it. Neither Kernel nor RootProgram owns
// the dependency graph, and no host-language stack waits for the child to finish.
// Queue capacity/fairness and coordination with activation/deactivation remain OPEN.

// ---------------------------------------------------------------------------
// 3. Contracts are separate programs; executables have argument contracts.
// ---------------------------------------------------------------------------

// This root executable has a boot ArgContract and an invocation protocol. Programs
// reached through OS and Preparation have their own argument contracts. Different
// operations on one executable need not create different executable identities.
// Contract selection (explicit identity or argument matching) remains OPEN.
//
// If a contract is realized by an independent validation/preparation program, root
// invokes that program through a port and waits via its continuation graph. It does
// not call an in-process validate/Realize interface on foreign code. The contract
// program's result must distinguish prepared arguments, missing inputs and failure;
// exact observations and association with ArgContract descriptors remain OPEN.
//
// A live bound executable receives invocations without another spawn. Launching
// creates a backing; binding connects it; invocation supplies observations. These
// transitions must remain distinct even if one OS operation performs spawn+execve.

// ---------------------------------------------------------------------------
// 4. Activation outcomes and the root boundary.
// ---------------------------------------------------------------------------

// finish handles an OUTWARD outcome for a recorded activation, never an arbitrary
// child Return. It correlates against Bookkeeper before calling Send on the Egress
// bound for the activation's outward result on RootProgram. The call carries only
// the Move, not a destination port; result routing follows the recorded activation.
//
// Return observations may describe suspension/saved state or completion. Fault
// carries the open FaultType. Saving is acknowledged only after persistence succeeds;
// closing invalidates bindings and coordinates Active release before acknowledgment.
// A suspended context may outlive the activation that returned. Ordinary child
// observations cannot instruct the runtime to save or close the enclosing context.
//
// Activation identity, result envelope, persistence executable, cancellation, and
// close-versus-parent/descendant lifecycle rules remain OPEN. This entrypoint does
// not interpret free-form text as a lifecycle command.
func (p *RootProgram) finish(ctx context.Context, activation FrameID, outcome Move) error

// The next executable proposal should describe the OS program's port protocol:
// launch request -> backing/connection result -> invocation delivery -> exit/fault.
// That supplies the first real handoff needed by boot without expanding this root
// file into the implementations of OS, preparation, provider or contract programs.
