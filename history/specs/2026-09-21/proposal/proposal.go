//go:build ignore

// Package dna18 is CODE ON PAPER for the current-direction runtime — the successor
// to DNA-17, built to the 2026-09-20 / 2026-09-21 model: mediated actors over a
// privileged kernel, virtual ports, and explicit execution contexts.
//
// An Executable identifies what to run; its ArgContracts define accepted arguments. An
// ExecutionContext owns state and a Horizon internally and exposes an Executable for
// activation externally. Horizons contain Executable references, not unrestricted
// pointers to the contexts/services behind them. Invocation authority never grants
// the caller the target context's internal Horizon.
//
// This is a PROPOSAL for review, not an implementation. Method bodies are omitted
// and the build tag excludes the file from compilation. Where a contract is still
// open it is marked OPEN and keyed to specs/2026-09-21/carry-forward.md §22.
//
// Sources: frame-stack-and-execution-ports.md, possibility-as-process.md,
// carry-forward.md (2026-09-21); ports/mediated-actors/os-kernel-frame,
// engrams-surfaces-and-pipe-protocol, dependency-mediation (2026-09-20);
// the syscall set (2026-09-19); grips-and-invariants.md (the WHY).
//
// LAYERING (top = substrate, bottom = the world):
//
//	RuntimeState  owns Memory + Active: stored runtime state, not an authority
//	              profile or a position in the execution-context hierarchy.
//	Memory        the durable catalog (REMEMBER): engrams, execution context definitions,
//	              executable handlers. Provisioning pages from here to Active.
//	Active        the resident working set: the active execution contexts (pointers into
//	              Memory) and actual-port reservations. provision publishes both.
//	TrapController holds Executables exposing selected service operations.
//	              Invokes them through Moves, with no direct Kernel/RuntimeState access.
//	Kernel        owns RuntimeState and Scheduler; resolves execution contexts through RuntimeState.Active.
//	Scheduler     owns the Move queue and realizing flags.
//	              Egress calls Enqueue after the fence and Active reservation check.
//	              Selects work; Kernel resolves the owning ExecutionContext.
//	ExecutionContext   owns its loop: Realize interprets, commits, and dispatches one Move
//	              targeting an Executable within its Horizon. Owns its Bookkeeper.
//	Bookkeeper    owns the ExecutionContext's frame stack, root sink and dependency bookkeeping.
//	VirtualPortSpace owned by the ExecutionContext runtime: local 0,1,2,... -> reserved
//	              actual port + Executable + binding. Backings see only virtual identities.
//	Frame stack   persistent logical continuation (returnTo / waiting / phase / cont).
//	Backing       the ephemeral executable/thread reached ONLY by Moves to its Port.
//	              An ACTOR, above the line; owns adapter state only.
//	Executable    a numeric executable ID with multiple ArgContracts (profiles).
//	ArgContract   observation arguments accepted by one supported operation.
//	              Execution resolves that ID and supplies observations as arguments.
//	Kernel ops    the wire is Moves over PORTS (send -> deliver); spawn / create_port
//	              / bind / send / unbind / kill — the only crossing into the OS,
//	              invoked, never authored. There are NO pipes: the port IS the channel.
//
// OWNERSHIP (who writes what; the rest is off-limits):
//
//	RuntimeState/Memory/Active  substrate state: RuntimeState owns Memory (durable catalog) + Active
//	              (working set). Memory holds definitions; Active holds pointers to the
//	              running ones. Selected writes are exposed through restricted
//	              Executables; ownership stays with the underlying service.
//	TrapController requests selected operations through Executables. Their implementations
//	              enforce the bound caller authority; Memory and Active retain their
//	              own state operations. ExecutionContext.Realize still owns the loop.
//	Scheduler     owns queue admission and scheduling; derives routing from trusted
//	              ingress, never a payload claim. No overlapping Realize calls for
//	              one ExecutionContext; different ExecutionContexts may advance concurrently.
//	ExecutionContext   its trusted runtime loop writes its
//	              frames through its Bookkeeper and owns local mappings/bindings.
//	              Uses only its Active-reserved actual ports; cannot mint
//	              beyond its horizon — that is a new provision (the confinement edge).
//	Backing       owns adapter/native state only, behind its PORT; reaches canonical
//	              state never — only emits Moves that re-enter the one door.
//	Executable       declares its input observations through ArgContract. Dependencies
//	              are invoked by Moves through the ExecutionContext's provisioned ports.
package dna18

import "context"

// ---------------------------------------------------------------------------
// 1. Identity leaf types.
//
// Distinct newtypes prevent cross-assignment. External identities are strings
// (deterministic, human-legible); runtime-minted identities are uint64 monotonic
// (inexhaustible; a stale reference hits a dead entry, never an aliased live one —
// the reason identity is never an OS pid). See ports-as-the-location-abstraction.
// ---------------------------------------------------------------------------

type ExecutionContextID string // identity of the state/authority context in which Executables execute
type ExecutableID uint64       // numeric executable identity; knowing an ID does not provision a port
type MoveID uint64             // the identity of one Move; when the opcode is Invoke it is the invocation id
type BackingID uint64          // a live backing (process/thread) in the process table; runtime-assigned, NEVER an address
type EngramID uint64           // a compiled, memoized instruction list stored in Memory; the memo key is OPEN

// Port is a VIRTUAL number local to one ExecutionContext: 0, 1, 2, 3, ... . Zero is
// valid. A's Port(1) and B's Port(1) are unrelated. The owning ExecutionContext runtime
// resolves local mappings; a bare number carries neither global identity nor authority.
type Port uint64

// KernelPort is an actual port reserved by Active to one ExecutionContext. Its
// trusted runtime stores it in the virtual mapping; agent backings never receive it.
// Allocation is independent of local virtual numbers.
type KernelPort uint64

// FrameID is scoped UNDER an ExecutionContext to separate namespaces. Realize must still
// reject foreign frame references and duplicate local ids; the type alone does not
// authenticate ownership. Seq is monotonic within the owning ExecutionContext.
type FrameID struct {
	ContextID ExecutionContextID
	Seq       uint64
}

// HorizonGen advances when an ExecutionContext's grant SET changes. Interpret snapshots
// it at authorization; it is distinct from a binding generation.
type HorizonGen uint64

// BindingGen advances every time a Port is (re)bound to a backing. It is the fence:
// an emission carries the gen it was produced under, and a stale gen is rejected.
// This is how a disposable worker's late emission — or a superseded backing after a
// rebind — cannot take effect, while a persistent session's live emissions pass.
// (Supersedes frame-stack-and-execution-ports' "mint a new Port number per
// dispatch": same Port, new BindingGen. Also retires that doc's authGen — authority
// is now space-membership, not a carried number. carry-forward §16, possibility-as-process §18.)
type BindingGen uint64

// ---------------------------------------------------------------------------
// 2. Move / opcode — the only thing that crosses a port.
//
// The opcode is which payload pointer is non-nil (Invoke | Return | Fault). A Move
// targets an ExecutableID. The owning ExecutionContext runtime resolves its local mappings;
// naming an ID does not grant access. The scheduler checks actual-port ownership, not
// Executable identity. The ExecutionContext's trusted binding fence binds Source and
// stamps identity. Scheduler.Enqueue admission derives the owner from that bound ingress
// (or an authorized initial submission), never from payload-supplied fields.
// (os-kernel-frame: the wire is async Moves over ports, not byte pipes.)
// ---------------------------------------------------------------------------

type Move struct {
	Identity           MoveID
	ExecutionContextID ExecutionContextID // context interpreting this Move; activation-boundary correlation is OPEN
	Source             MoveSource         // bound by the ExecutionContext binding fence; zero for an admitted first Move
	Target             ExecutableID       // the Executable to invoke; unused for Return/Fault
	Invoke             *Invoke
	Return             *Return
	Fault              *Fault
}

// MoveSource records the emitting frame + binding. Bound by the trusted ExecutionContext runtime at port egress,
// never claimed by the payload. Fault targeting is validated against this (fixes
// DNA-17 issue #2). (loop-owned-authority; mediated-actors: the wiring is the authority.)
type MoveSource struct {
	Frame   FrameID
	Port    Port
	Binding BindingGen
}

type Invoke struct {
	Frame    FrameID       // minted by the owning ExecutionContext for a child
	Args     []Observation // all Executable inputs, including focus when applicable
	ReturnTo FrameID       // parent to answer; zero-Seq == the ExecutionContext's root sink
}

type Return struct {
	Answering   FrameID // correlated against the RECORDED frame, never a payload claim
	Observation Observation
}

// Fault is a failure-report payload variant that re-enters the one door, not a
// third opcode. The ExecutionContext binds its subject from MoveSource, so a backing cannot
// fault unrelated work. (DNA-17 issue #2.)
type Fault struct {
	Frame     FrameID
	FaultType FaultType
}

// FaultType is intentionally open; its fields and semantics are still to be decided.
type FaultType struct {
	// OPEN
}

// Observation is the canonical value leaf. Key names a segment author (used by the
// Codex driver to key tool arguments); Bytes is the content. A signed Observation
// is the structural provenance unit (grips invariant III). Signing is OPEN.
type Observation struct {
	Key   string
	Bytes []byte
}

// ---------------------------------------------------------------------------
// 3. ExecutionContext — the program's execution state and Horizon.
//
// ExecutionContext owns its Horizon and continuation cactus through Bookkeeper.
// Its Executable exposes start/resume as ordinary Invoke semantics. An invocation
// starts one activation; its outward Return settles that activation, not necessarily
// the persistent context. Suspension retains continuation; completion retires the
// context; failure uses Fault. Exact lifecycle observations remain OPEN.
//
// Realize runs internal Moves against this context's own Horizon. The activation
// boundary must correlate outward outcomes back to the caller without exposing or
// transferring the target context's authority. Ordinary child Returns settle child
// work, never implicitly suspend/close the context. A save request is acknowledged
// as saved only after persistence succeeds. Activation/correlation encoding is OPEN.
// ---------------------------------------------------------------------------

type ExecutionContext struct {
	Executable Executable // activation protocol; ExecutableID is distinct from ExecutionContextID
	Identity   ExecutionContextID
	Horizon    *VirtualPortSpace // runtime-owned virtual-to-actual mappings; backings see local numbers only
	Bookkeeper *Bookkeeper
	Generation HorizonGen
}

// VirtualPortSpace belongs to the ExecutionContext's trusted runtime. Active only
// knows which actual ports it has reserved for this ExecutionContext. Executable resolution,
// virtual numbering, binding generations and liveness are local bookkeeping.
// Agent backings receive virtual identities (0,1,2,...) and scoped emission sinks,
// never this map, actual port identities, or pointers to entries/bindings.
type VirtualPortSpace struct {
	entries  map[Port]*PortEntry
	nextPort Port // allocation starts at 0; no special meaning for 0/1/2
}

// PortEntry maps a local virtual port to an actual port reserved to this owner,
// plus the provisioned Executable and current binding. Only trusted runtime code
// mutates it; all actual-port use is checked against the Active reservation table.
type PortEntry struct {
	actual     KernelPort
	executable ExecutableID
	binding    *Binding // nil while unbound
	live       bool
}

// ExecutionContext runtime rebinding may replace a backing for the provisioned Executable only. Every
// (re)bind advances gen, including reuse of a virtual slot; generations must never
// reset in a way that makes an old emission valid. Exact allocation policy is OPEN.
type Binding struct {
	backing BackingID
	gen     BindingGen
	live    bool
}

// ---------------------------------------------------------------------------
// 4. Bookkeeper — owns the frame stack, root sink and dependency bookkeeping.
//
// Each ExecutionContext owns its own Bookkeeper. The stack remains local to that
// ExecutionContext, with no authority or addressing role (Moves target Executables).
// Bookkeeper operations participate in the owning ExecutionContext's state turn/lock;
// this service introduces no separate queue or independently scheduled loop.
// ---------------------------------------------------------------------------

type Bookkeeper struct {
	FrameStack *FrameStack
	RootSink   FrameID // the bottom frame; where a root result is delivered (the return-out edge — OPEN, §22.6)
}

// traceDependencies is the Bookkeeper's initial stage entry point, called by
// ExecutionContext.Realize during interpretation to trace the Move's frame dependencies.
// It reads only this ExecutionContext's stack; any decided stack changes belong to the
// same turn's commit. Exact tracing rules and result shape remain OPEN; this is a
// starting signature, not an implemented dependency traversal or multi-child join.
func (*Bookkeeper) traceDependencies(move Move) error

type FrameStack struct {
	frames  map[FrameID]*Frame // arbitrary map; keys scoped under the owning ExecutionContext
	nextSeq uint64
}

type Phase string

const (
	PhaseRunning  Phase = "running"
	PhaseWaiting  Phase = "waiting" // parked on one or more children
	PhaseReturned Phase = "returned"
	PhaseFaulted  Phase = "faulted"
)

type Frame struct {
	id       FrameID
	returnTo FrameID // the cactus edge; zero-Seq == root sink
	phase    Phase
	waiting  []FrameID // dependency set; single-child ENFORCED AT INTERPRET for now (§22.8 / issue #5)

	// the (re)dispatch spec — what to run when this frame executes:
	target ExecutableID
	args   []Observation

	cont   Continuation // externalized continuation (was DNA-17 driverState); must outlive any backing
	result Observation
	active *ExecutionRef // nil when nothing is live; virtual port 0 is valid
}

// Continuation is the frame's externalized state — enough to resume on a FRESH
// backing after the previous one exits (disposable-worker requirement). Its
// representation is OPEN: opaque bytes vs a typed contract snapshot.
type Continuation struct {
	Version uint32
	Payload []byte
}

// ExecutionRef is opaque execution correlation carried by the frame. Port is
// virtual; Binding is a fence token. The ExecutionContext runtime resolves it in its
// own space and checks mapping/liveness/generation. Scheduler ownership checks apply
// separately to the reserved actual port.
// Absence is represented by a nil *ExecutionRef, never by Port == 0.
type ExecutionRef struct {
	Port    Port
	Binding BindingGen
}

// ---------------------------------------------------------------------------
// 5. RuntimeState, Memory, Active — the substrate state.
//
// RuntimeState owns Memory (the durable catalog) and Active (the resident working
// set and port reservations). It holds no root-context identity and does not model
// ancestry or delegated authority. Kernel owns RuntimeState and uses Active to
// resolve execution contexts; each context owns Realize.
// ---------------------------------------------------------------------------

// Memory is the durable catalog — REMEMBER's store (grips invariant I): every engram,
// execution context DEFINITION, and executable handler, active or not. Provisioning pages an
// entry from here onto Active; this is what makes recurrence the resource — a warm
// execution context pages in instead of being re-derived. Registration writes here;
// registration is NOT a grant. Selected reads/writes are exposed as restricted
// Executables. (Collapses the earlier MemoryStore + ExecutableRegistry.)
type Memory struct {
	engrams           map[EngramID]*Engram                     // compiled, memoized instruction lists
	executionContexts map[ExecutionContextID]*ExecutionContext // definitions: horizon + first move + frame layout
	executables       map[ExecutableID]*Executable             // registered handlers
}

// Active is the resident working set: the execution contexts currently active. Each value
// points INTO Memory.executionContexts — a pointer to the durable definition, never a copy.
type Active struct {
	executionContexts map[ExecutionContextID]*ExecutionContext
	portReservations  map[KernelPort]ExecutionContextID // actual port -> owner; no Executable or binding knowledge
}

// provision publishes an authorized execution context and reserves its actual ports
// in one Active state change. Conflicting reservations must fail without partial
// publication. Authority decisions and local Executable bindings remain with their
// existing owners; Active owns the context/reservation records. Exact activation
// and port allocation policy remain OPEN; this is a starting declaration.
func (*Active) provision(ctx context.Context, execution *ExecutionContext, ports []KernelPort) error

// Active owns both lookups; callers receive no writable map aliases.
func (*Active) ownerOf(port KernelPort) (ExecutionContextID, bool)
func (*Active) lookup(id ExecutionContextID) (*ExecutionContext, bool)

// RuntimeState groups the catalog and active runtime records. It contains no
// rootContext field; execution-context hierarchy is a separate concern.
type RuntimeState struct {
	memory *Memory
	active *Active
}

// ---------------------------------------------------------------------------
// 5a. TrapController — selected service operations exposed as Executables.
//
// A profile is an ArgContract: it defines arguments for a supported operation,
// not a separate executable. One Executable ID can support multiple ArgContracts.
// Its program interprets the arguments behind the port boundary. The caller's
// Horizon decides reachability; trusted binding supplies caller authority.
// How an invocation selects a contract (explicit identity or argument matching)
// and how operation-level restrictions are bound remain OPEN.
// RuntimeState stays the state owner, with no unrestricted pointer passed to
// TrapController. Calls use ordinary Moves through ExecutionContext.Realize.
// ---------------------------------------------------------------------------

type TrapController struct {
	executables []Executable
}

// Example: a runtime-state executable can accept a lookup ArgContract for selected
// Memory reads and a provision ArgContract reaching Active.provision after authority
// preparation. Registration/deactivation may have further ArgContracts on that same
// executable; operations do not each require a new ExecutableID. Encoding remains OPEN.
// Registration records an Executable but grants no authority. Deactivation coordinates
// scheduling, local binding invalidation and actual-port release before reaping.
//
// Provisioning inputs travel as observations under an Executable's selected ArgContract;
// there is no separate OpenExecutionContext request or special controller Provision
// entry point. Authority preparation admits only delegated authority; activation
// bookkeeping remains Active.provision. Local mapping/binding writes remain owned
// by the ExecutionContext runtime. Context publication, port reservation and first
// Move admission must succeed together or roll back; that transaction remains OPEN.
//
// A root authority context may provision children, which may provision descendants
// only through admitted provisioning Executables and within delegated authority. Those
// contexts form an ownership hierarchy; their exposed Executables may be admitted
// into Horizons without granting access to their internal state or authority.
// Bootstrap establishes initial authority. Hierarchy storage/lifecycle remain OPEN;
// RuntimeState still holds only Memory and Active, with no root-context field.

// ---------------------------------------------------------------------------
// 5b. Kernel owns the Scheduler; ExecutionContext realizes.
//
// Port egress -> fence and stamp ownership -> Scheduler.Enqueue(move) ->
// Scheduler.Run -> supplied dispatch resolves RuntimeState.Active -> ExecutionContext.Realize(ctx, move)
// -> execute the target Executable against that context's Horizon.
// The Scheduler owns one queue of Moves. RuntimeState.Active is the only active ExecutionContext
// map; there is no second owner map or scheduling envelope around each Move.
// ---------------------------------------------------------------------------

type Kernel struct {
	runtimeState *RuntimeState // owns substrate state and resolves active execution contexts
	scheduler    *Scheduler
}

// Kernel routes actual-port ingress using RuntimeState.Active.portReservations and rejects
// use by any other owner. Active owns both context lookup and reservations. Reservation
// allocation/release never interprets local virtual numbers, Executables or bindings.
// ExecutionContext runtime maps its virtual ports onto these reserved actual ports.
// Before release/reassignment, invalidate local bindings and drain in-flight use;
// stale transport traffic must not be attributed to a new owner. The concrete
// reservation reuse/fencing mechanism remains OPEN.

type Scheduler struct {
	queue     []Move                      // admitted, detached Moves in arrival order
	realizing map[ExecutionContextID]bool // scheduler-owned; never stored on ExecutionContext
}

// Enqueue is called by trusted port egress AFTER Kernel verifies actual-port
// ownership through RuntimeState.Active.portReservations and the owning ExecutionContext's emission fence checks the bound frame and live binding/generation, commits the emission,
// and stamps ExecutionContextID and Source from the binding. A backing such as Codex
// receives only that scoped egress sink, never the Scheduler or an arbitrary-owner
// enqueue API. Authorized provisioning uses the same admission method for First;
// trusted local follow-up Moves also enter here with their owner's identity bound.
// Enqueue retains a detached copy; a sender cannot mutate queued work. Queue access
// is synchronized with Run. Admission must not block the sole scheduler on its own
// follow-up work; capacity, backpressure and shutdown policy remain OPEN.
func (*Scheduler) Enqueue(move Move) error

// DispatchMove is the narrow handoff supplied by the runtime wiring: resolve the
// selected Move's active ExecutionContext and call its Realize. It does not select
// queue work, provision missing contexts, or own Scheduler's realization tracking.
// Scheduler receives no Kernel or RuntimeState pointer.
type DispatchMove func(ctx context.Context, move Move) error

// Run takes one eligible Move from the queue, claims realizing[owner], dispatches
// it, and releases the claim on every completion/error path. It waits if no work is
// eligible, skips busy owners and preserves order within each context. It completes
// after the Realize turn, not after provider completion. Cancellation before selection
// leaves no claim. Worker concurrency and admission/dispatch errors remain OPEN.
func (*Scheduler) Run(ctx context.Context, dispatch DispatchMove) error

// loop is the PROVISIONAL location for the main loop. Final ownership and its
// receive/scheduling concurrency remain OPEN. It wires ingress admission and
// Scheduler.Run; it does not duplicate queue selection or realization bookkeeping.
func (*Kernel) loop(ctx context.Context) error

// Realize executes one Move against this context's Horizon:
//  1. Interpret: validate ownership, frame correlation and target against
//     the Horizon via local resolution of the target Executable,
//     validate Args against the selected ArgContract, delegate frame dependency
//     tracing to Bookkeeper.traceDependencies, read local state and compile an Engram. Contract matching rules remain OPEN.
//  2. Commit: apply Bookkeeper frame writes and runtime-owned binding effects
//     in the SAME ExecutionContext state turn.
//  3. Execute: release the lock, dispatch committed work asynchronously, and return
//     without waiting for provider completion. Follow-up Moves enter the Scheduler queue.
//
// Scheduler scheduling prevents overlapping Realize calls for this ExecutionContext. The
// local state turn serializes frame/emission commits with local mapping changes
// and revocation. Different ExecutionContexts share no mutable frames, mappings or
// continuations. Active reservation release coordinates with local invalidation.
// Backings receive detached inputs and a scoped emission sink, never the ExecutionContext.
// Emission admission checks the bound frame and live binding/generation, stamps
// provenance, and returns Moves through Scheduler.Enqueue. Realize validates ownership again;
// a claimed ExecutionContextID or FrameID cannot redirect work across spaces.
//
// Scheduling plus these ownership/fence rules enforce runtime state isolation.
// Provider isolation and OS containment remain separate obligations. Realize is a
// trusted scheduler entry point, not an API exposed to the agent program.
func (*ExecutionContext) Realize(ctx context.Context, move Move) error

// ---------------------------------------------------------------------------
// 6. Engram — what ExecutionContext.Realize compiles a Move into before dispatch.
//
// The serialized, ordered list of authorized instructions the OS layer runs. The
// OS runs engrams; it does not replace them (os-kernel-frame). "Recurrence is the
// resource": a warm Engram pages in instead of recompiling (grips invariant I).
// The entry representation, ordering, result binding, and memoization key are OPEN
// (§22.4). Below is the minimal shape, NOT a settled encoding.
// ---------------------------------------------------------------------------

type Engram struct {
	Identity     EngramID           // key in Memory.engrams; the memoization key is OPEN
	Move         Move               // the Move this compiled from
	Instructions []Instruction      // serialized, ordered; a consumer of an earlier result runs after it
	ContextID    ExecutionContextID // authority namespace; execution runs under this Horizon
}

// Instruction is one authorized executable entry. It is NOT a Move — Moves are
// protocol transitions; instructions are the compiled steps. The system-plane ops
// (Bookkeeper frame writes and ExecutionContext binding writes) and the execution-plane dispatch are both Instructions; the exact
// split, conditional branches, and result binding are OPEN.
type Instruction interface{ instruction() }

// System-plane instructions (ExecutionContext runtime/Bookkeeper writes at commit; the closed vocabulary):
type opMintFrame struct {
	frame Frame
	phase Phase // DECIDED by interpret: Running when authorized, Faulted when trapped
}
type opBindAndDispatch struct { // authorize a start: bind a Port to a backing (or reuse a live one) and hand it the input
	frame FrameID
	port  Port
	input DriverInput
}
type opSettleReturn struct { // mark a frame returned, record result, kill its binding
	frame FrameID
	obs   Observation
}
type opResumeParent struct { // resume a waiting parent over its EXISTING binding with a child result
	parent FrameID
	child  Observation
}
type opEmitFault struct { // collect a corrective Fault Move to re-enter the door after the lock releases
	frame     FrameID
	faultType FaultType
}
type opSettleFault struct{ frame FrameID }
type opRootResult struct { // deliver the single result to the ExecutionContext's root sink
	frame FrameID
	obs   *Observation
	err   error
}

func (opMintFrame) instruction()       {}
func (opBindAndDispatch) instruction() {}
func (opSettleReturn) instruction()    {}
func (opResumeParent) instruction()    {}
func (opEmitFault) instruction()       {}
func (opSettleFault) instruction()     {}
func (opRootResult) instruction()      {}

// ---------------------------------------------------------------------------
// 7. Executable / Backing — the execution boundary. PORTS ARE THE TRANSPORT.
//
// An Executable is an ADDRESSABLE EXECUTABLE ENDPOINT (engrams doc). Composition is not
// on the Executable — that is the Engram. A backing is reached ONLY by Moves delivered
// to its Port, and it emits ONLY Moves back to its Port. No byte pipe, no raw
// normalization, and no separate Transport crosses this boundary: there is one
// channel abstraction — the port — not ports AND pipes. Any byte framing, and any
// real subprocess a driver drives (e.g. a raw codex app-server), is PRIVATE to that
// driver, below its port, and never a runtime concept.
// ---------------------------------------------------------------------------

// Executable defines what to run. ArgContracts define accepted argument forms for
// its supported operations (profiles). For example, one Codex executable can accept
// start-thread, submit-input and tool-result arguments without separate executable
// IDs for those operations. Selection by contract identity versus argument matching
// remains OPEN. No Go method is required on an Executable; the running program
// interprets observations delivered through its port.
// An adapter is an Executable; provider protocol handling is its implementation detail.
// Invocation is conceptually execve(executableID, args): resolve the ID to executable
// code and pass observation arguments. Ports still determine reachability; the
// ExecutionContext runtime resolves Move.Target (ExecutableID) in its virtual space
// to a binding and reserved actual port before dispatch. The chosen port stays execution state, not a Move target.
// Multiple ports for the same Executable require a local selection rule (OPEN);
// ambiguous resolution must not silently choose a binding.
type Executable struct {
	ID             ExecutableID
	ArgContracts   []ArgContract
	OutputContract OutputContract
	FaultContract  FaultContract
}

// ArgContract is a profile: observations defining the arguments accepted by one
// operation of an Executable. It does not identify the program or mandate a method.
// Invoke.Args holds all supplied observations, including focus when the Executable
// accepts it. Focus is ordinary input data, not a separate protocol field or an
// authority grant; the provisioned Horizon still bounds reachability.
// Matching/validation and observation
// encoding into OS arguments remain OPEN; this list adds no separate schema type.
type ArgContract struct {
	Observations []Observation
}

// OutputContract describes expected Return observations. Lifecycle outcomes such
// as suspended/saved or completed are meaningful only at a context's outward
// activation boundary. Exact alternatives and matching rules remain OPEN, including
// encoding these observations in the existing Return.Observation value.
type OutputContract struct {
	Observations []Observation
}

// FaultContract describes declared failure outcomes. FaultType itself remains open;
// this list does not yet define categories or matching semantics.
type FaultContract struct {
	FaultTypes []FaultType
}

// Backing is the live instance bound to a Port. Deliver hands it ONE Move (with the
// frame's continuation and, on resume, the child result); the backing does work and
// EMITS Moves back to its port whenever — synchronously or long after Deliver
// returns. Async, unprompted emission (a Codex turn calling a tool mid-run) is just
// the backing sending a Move to its port; the port is always live while bound, so no
// separate persistent callback is needed — the PORT is the standing channel. The
// emit sink is scoped by the ExecutionContext runtime to its frame and binding, checked
// by the fence, and returns admitted Moves to the Scheduler queue.
// This Go interface is a private host integration sketch, not a requirement on
// executable programs. Ports and Moves are the interoperability boundary.
type Backing interface {
	Deliver(ctx context.Context, in DriverInput, emit func(Emission) error) error
}

type DriverInput struct {
	Move        Move
	State       Continuation
	ChildResult *Observation // present only on a resume step
}

// Emission is what a backing sends back over its port: emitted Moves (already in
// protocol form — the backing did any provider-specific normalization internally),
// the replacement continuation, or a failure. No raw bytes and no opaque host-command
// blobs cross here. Commit-before-send ordering (a side effect that must follow its
// commit) becomes a follow-up Move round-trip, not a Commands payload — CONFIRM.
type Emission struct {
	State   Continuation
	Moves   []Move
	Failure error
}

// ---------------------------------------------------------------------------
// 7a. Provider integration is an Executable implementation.
//
// A Codex integration has one numeric ExecutableID and multiple ArgContracts. Provider operations, when needed, are expressed in its observation
// arguments; there is no separate action field or adapter type. JSON-RPC, framing,
// thread correlation and private provider subprocesses remain inside that Executable.
// Each ExecutionContext's binding must isolate its provider state. Pooling a provider
// process requires trusted demultiplexing and provider isolation; policy remains OPEN.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// 8. The execution-plane fence — stateVia(ref).
//
// A backing reaches canonical state ONLY through the fence. Commit checks the ref
// live: is this Port still mapped in its space AND does the BindingGen match. A
// revoked (unmapped) or superseded (rebound) or settled ref rejects everything —
// one lookup covering revocation + settlement + stale-worker together. The profile
// is bound to an owning ExecutionContext and frame, with no caller-selected write target.
// Commit coordinates the local binding fence and Bookkeeper frame writes in one
// state turn; admitted Moves enter the Scheduler queue after release. A backing cannot obtain another ExecutionContext's execution profile.
// (possibility-as-process §18; handle-liveness-fence.)
// ---------------------------------------------------------------------------

type ExecutionState interface {
	Commit(ctx context.Context, ec EmissionCommit) (EmissionResult, error)
}

type EmissionCommit struct {
	State   Continuation // replace the frame's externalized continuation
	Moves   []Move       // emitted Moves; Bookkeeper mints child ids; ExecutionContext runtime stamps binding provenance
	Failure error
}

type EmissionResult struct {
	Enqueue []Move
}

// ---------------------------------------------------------------------------
// 9. Executable requests use ordinary Invoke Moves.
//
// Dependencies are requested by ExecutableID with observation arguments.
// ExecutionContext.Realize performs local Executable resolution and checks the
// selected ArgContract,
// records the edge, and dispatches. Work requiring another Executable traps for
// provisioning. No separate capability manifest or dependency-request type exists.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// 10. Trap / Fault. Trap = transfer control to preparation. Fault = an error.
// (authority-namespace.) The TrapController is an addressable Executable, not a
// callback inside the authority check. What it may do and how its outcome answers
// the original invocation are OPEN (§22.8).
// ---------------------------------------------------------------------------

// Handle receives an admitted trap and prepares corrective work OR a denial.
// Its resulting work is admitted to the Scheduler queue under the trapped ExecutionContext;
// execution remains owned by ExecutionContext.Realize.
func (*TrapController) Handle(ctx context.Context, t Trap) (Engram, error)

type Trap struct {
	Frame   FrameID
	Reason  string
	Missing []Observation // unmet input observations, if any
}

// ---------------------------------------------------------------------------
// 11. Kernel ops and the OS floor — PORTS ARE THE WIRE.
//
// The wire is async Moves over ports (send(port, move) -> deliver), NOT byte pipes:
// os-kernel-frame supersedes the 2026-09-19 read/write/close pipe framing. The
// kernel ops are spawn / create_port / bind / send / unbind / kill — invoked, never
// authored; an ExecutionContext only REQUESTS spawn/bind/kill (topology changes are authority-
// checked and attenuating). Process lifecycle (spawn/wait/signal) is real, but there
// is no PipeEnd, Transport, or read/write in the model: byte movement to a real OS
// process is PRIVATE to a driver's realization of an external port, never a runtime
// primitive. TrapController authorizes provisioning; Active owns actual-port
// reservations, checked by Kernel. ExecutionContext runtime owns virtual mappings,
// bindings, generations and liveness. ExecutionContext.Realize owns the loop and
// delegates frame bookkeeping to Bookkeeper; OSExecutor performs
// process lifecycle effects. Kernel owns the Scheduler and its Move queue. "KernelOp"
// names the primitive vocabulary, not a single implementation owner.
// ---------------------------------------------------------------------------

type KernelOp string

const (
	OpSpawn      KernelOp = "spawn"       // executable image -> a BackingID (posix_spawn-shaped; no pipe plan surfaces)
	OpCreatePort KernelOp = "create_port" // allocate a Port naming an authorized ExecutableID
	OpBind       KernelOp = "bind"        // map a backing into an ExecutionContext's port space (port -> backing)
	OpSend       KernelOp = "send"        // deliver a Move to a Port (the wire; async, re-enters the loop)
	OpUnbind     KernelOp = "unbind"      // break a binding without killing the backing
	OpKill       KernelOp = "kill"        // terminate + REVOKE — unmap the port (the revocation primitive)
)

// OSExecutor realizes execution as execve(executableID, observations): resolve the
// numeric ID to executable code and supply observations as arguments. The runtime
// does not expose paths or string argv in the Executable contract. This is the logical
// launch boundary; actual execve/spawn mechanics and argument encoding are private
// to the execution layer. The returned BackingID names the live instance, not the
// reusable Executable. Wait and Signal retain lifecycle control of that instance.
// Private host integration sketch; an executable need not implement this Go API.
// Invoking an already-bound persistent program does not launch it again.
type OSExecutor interface {
	Execve(ctx context.Context, executable ExecutableID, args []Observation) (BackingID, error)
	Wait(ctx context.Context, id BackingID) (ExitStatus, error)
	Signal(ctx context.Context, id BackingID, sig int) error
}

type ExitStatus int

// ---------------------------------------------------------------------------
// 12. THINK boundary (out of scope for this proposal; the WHY it serves).
//
// The Contract is invariant II: THINK negotiates a typed contract with the human;
// deviation surfaces as a Fault, never silent drift (grips-and-invariants). It sits
// ABOVE this substrate — the human/model negotiation that PRODUCES Engrams and the
// Horizon an ExecutionContext opens with. Declared here only to mark the seam; its full
// shape (THINK/ACT/REMEMBER, the deopt) is a separate proposal.
// ---------------------------------------------------------------------------

type Contract struct {
	Wants   string         // what the human asked for (the plan)
	Horizon []ExecutableID // the Executables it may reach; an ExecutionContext opens with this
	// Deviation is a TYPE ERROR, surfaced as Outcome{Fault}, not a skipped line.
}
