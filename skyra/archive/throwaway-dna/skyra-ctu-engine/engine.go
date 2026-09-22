// Package engine is a partial implementation probe, not a working runtime.
//
// AUTHORITY: specs/2026-09-01/register-frame.md, focus-and-move.md, and
// engram-loading-and-fault-return.md refine ingress-and-fire.md. The earlier
// three-peer topology and v12 scalar-register frames are superseded here.
//
// KERNEL contains Scheduler and provisions through shared Memory. MEMORY owns
// resident structures, registers, and value lifetimes. CTU executes using Memory.
// Scheduler owns one MoveID → State map. CTU.Fire dispatches claimed batches. Op.Fire
// executes one operation and returns a DATA event. Scheduler interprets events,
// never opcodes. One work/result ingress is planned to split on opcode: Call
// provisions an fd entry; Ret reaches the existing return behavior. Move is the
// universal wire envelope and ingress input; receiver/result signature remain OPEN. Landed reports prior publication, never a second write.
// Fire(batch) is the claimed-work handoff. Algorithm stubs panic explicitly rather
// than report progress.
//
// LATEST LOOP DECISION: CTU evaluates readiness and executes ready entries. Scheduler
// marks handed-off work Pending, receives emitted Moves through opcode ingress, and
// publishes Ret observations through Memory. Issued is removed. The remaining Claim /
// Claimed / Landed / Event surface below is the earlier probe, pending replacement;
// it is not the contract for the next run. Fault and waiting-work transport are OPEN.
package engine

import (
	"bytes"
	"encoding/json"

	c "skyra-v1/src/skyra/skyra-dna-v12/causal"
)

// MoveID canonically represents Action + ordered observation Params. Fd uses this
// identity directly to address Memory.Fds; there is no separate register table.
// Encoding, lookup scope, and sharing for identical invocations remain OPEN.
// Target is outside this identity; actions must be unambiguous within that scope.
// This replaces v12's Possibility/Writes identity only within the engine proposal.
type MoveID struct{ Value string }

// Contract describes the returned observation's structure. Action and Params live
// on invoke, not here. RawMessage is an opaque carrier, not a schema validator.
// Schema discovery, dialect, validation, and unknown-schema representation are OPEN.
type Contract struct {
	Schema json.RawMessage
}

// Register is one execution's return frame, keyed by its fd in Memory. It collects
// many child observations to fulfill one return contract. Scheduling state is separate.
type Register struct {
	Contract Contract
	Children []ChildObservation
}

// ChildObservation identifies a child execution and its one returned value.
// The slice retains declaration order: a landing fills the matching fd entry,
// never appends in completion order. Non-nil means arrived, not necessarily valid.
type ChildObservation struct {
	Fd    MoveID
	Value *c.Observation
}

// EngramID addresses immutable executable backing, whether resident or persisted.
// The digest encoding and durable Engram/Body struct are still OPEN; do not reuse
// v12 Body as if its scalar-register contract implemented the new specification.
type EngramID struct{ Value string }

// PHASES & OWNERSHIP
//
// ADMIT     scheduler records Ready using readiness information from CTU.
// CLAIM     scheduler records Claimed BEFORE handing a batch to CTU.Fire.
// EXECUTE   CTU invokes Op.Fire; execution has no scheduler-map access.
// LANDED    report a result; scheduler records Completed and retains the entry.
// FAULTED   report a fault; scheduler parks/arranges resolution (state OPEN).
// CONSUME   CTU uses Memory's operand-release operations (lifetime rules OPEN).
// RETIRE    scheduler removes completed lifecycle records under a policy (OPEN).
// REMEMBER  persist execution structure/type feedback, distinct from completion.
//
// Duplicate returns MUST NOT regress Completed. Retry correlation, duplicates,
// and failed handoff need a transition discipline. Map emptiness is NOT quiescence.
// Materialize reissues the SAME MoveID; no separate Attempt ID is adopted.
// Batches have no ordering or completion barrier; moves complete independently.

// State contains scheduling lifecycle only. Names are provisional. Zero is
// unspecified/absent, not Ready. Waiting/fault states and retry transitions are OPEN.
type State uint8

const (
	Ready State = iota + 1
	Claimed
	Completed
)

// Scheduler owns scheduling lifecycle state. The private map contains neither executable
// payloads nor Contexts nor Observations. Descriptor lookup, dependency-index
// ownership, and readiness/event delivery APIs remain OPEN.
type Scheduler struct {
	states map[MoveID]State
}

// Admit establishes Ready after validating readiness information. OPEN: its query/
// event ABI and the home of not-yet-ready moves. This is not an unchecked assertion
// of readiness that an untrusted caller may use to bypass admission.
func (s *Scheduler) Admit(fds []MoveID) { panic("PROPOSAL: readiness/admission") }

// Claim records Claimed before dispatch and returns only successfully claimed fds.
// OPEN: serialization and recovery if the following handoff fails.
func (s *Scheduler) Claim(fds []MoveID) []MoveID {
	var claimed []MoveID
	for _, fd := range fds {
		if s.states[fd] == Ready {
			s.states[fd] = Claimed
			claimed = append(claimed, fd)
		}
	}
	return claimed
}

// Apply is execution-event ingress: one closed switch, no opcode knowledge and no
// register writes. The plain switch is the sketch choice; event.Apply(writer) is
// still an alternative. This single-writer probe handles trusted dispatch/completion
// events only; dependent admission and fault recovery remain blocked. The caller
// must establish publication before delivering Landed, as required by its contract.
func (s *Scheduler) Apply(e Event) {
	switch e := e.(type) {
	case Landed:
		if s.states[e.Fd] != Claimed && s.states[e.Fd] != Completed {
			panic("engine: Landed requires a claimed or completed move")
		}
		s.states[e.Fd] = Completed
	case Faulted:
		panic("PROPOSAL: fault parking and resolution")
	default:
		panic("PROPOSAL: invalid execution event")
	}
}

// Retire removes completed records separately from completion. It neither frees
// Memory's registers nor persists an engram. Eligibility/retention remains OPEN.
func (s *Scheduler) Retire(id MoveID) { panic("PROPOSAL: completed-record retention") }

// Event is a closed scheduling vocabulary, independent of extensible opcodes.
// Encoding/delivery remain OPEN: a Go interface is not itself a wire codec.
type Event interface{ isEvent() }

// Landed reports a completed return after Ret.Fire has published through Memory.
// Scheduler records completion and coordinates readiness; it does not route/write
// child observations. Root-result delivery and asynchronous adapters remain OPEN.
type Landed struct {
	Fd    MoveID
	Value c.Observation
}

// Faulted is retained pending reconciliation with the Move-return path. Scheduler parks the
// move and routes kernel recovery; CTU does not independently invoke a handler.
// Missing names the required engram for Materialize; nil for other fault kinds.
// This optional payload is provisional: validation/correlation remain OPEN. A
// fault neither fills a parent's child entry nor grants additional authority.
type Faulted struct {
	Move    MoveID
	Vector  Kind
	Missing *EngramID
}

func (Landed) isEvent()  {}
func (Faulted) isEvent() {}

// Op is one execution behavior. Its immediate event need not be the final result:
// The Event result is an older surface pending reconciliation with emitted Moves.
type Op interface{ Fire(x *ExecCtx) Event }

// ExecCtx binds an operation to its CTU, move, and Memory-owned return frame.
// This provisional carrier has no scheduler reference; narrower execution
// interfaces can replace Host once their methods are settled.
type ExecCtx struct {
	Host  *CTU
	Move  *Move
	Frame *Register // the Register on the Fd entry handed to CTU.Fire.
}

// Touching declares an access; MMU checks actions, not opcode kinds (D13).
type Touching interface{ Touches() (c.Target, c.Name) }

type invoke struct {
	Target c.Target
	Action c.Name          // sole model-visible method name; 160-token cap, enforcement OPEN.
	Params []c.Observation // actual inputs, in order; resolve references before deriving MoveID.
}

func (i invoke) Touches() (c.Target, c.Name) { return i.Target, i.Action }

// Call invokes a child locally or through Egress. Dispatch and completion are
// separate events. It never writes scheduler state.
type Call struct{ invoke }

func (o Call) Fire(x *ExecCtx) Event { panic("PROPOSAL: child invocation") }

// Exec touches a driver after authority checks. CTU owns the asynchronous adapter;
// the synchronous Driver has no scheduling responsibility.
type Exec struct{ invoke }

func (o Exec) Fire(x *ExecCtx) Event { panic("PROPOSAL: marshal and driver dispatch") }

// Ret carries the proposed output Observation. Fire checks required child
// observations and x.Frame.Contract, then uses x.Move.ParentFd and x.Move.Fd to
// publish through Memory.Land before emitting Landed. Nil ParentFd routes to the
// root result destination (OPEN). Output construction and validation remain OPEN:
// the schema constrains values but cannot construct them. No contract copy lives here.
type Ret struct{ Value c.Observation }

func (o Ret) Fire(x *ExecCtx) Event { panic("PROPOSAL: result landing") }

// Trap is retained as an explicit-opcode proposal only. Its remaining role is OPEN.
// An involuntary fault already returns Faulted; CTU must not also fire a Trap to
// independently start the same recovery. Any reissue goes through Scheduler.
type Trap struct {
	Vector  Kind
	Faulted *Move
}

func (o Trap) Fire(x *ExecCtx) Event { panic("PROPOSAL: explicit trap role and event delivery") }

type Kind int

const (
	Materialize Kind = iota // missing backing, distinct from missing operands.
	Remember                // persist execution structure/type feedback.
	Confirm                 // effect gate requiring kernel/human resolution.
)

// Move is the universal wire envelope; its Op selects ingress behavior. Registers
// stay resident in Memory. Op discriminator/payload encoding remains OPEN.
// Move carries execution identity and its parent return address. Fd identifies
// this move's register and its entry in the parent's Children; ParentFd identifies
// the parent register. Nil ParentFd denotes a root. Both survive materialization.
type Move struct {
	Op       Op
	Root     c.PossibilityID
	Fd       MoveID
	ParentFd *MoveID // parent register; nil for a root move.
	Horizon  c.Horizon
	Focus    []c.Capability // selected existing Horizon grants; no new authority.
	Engram   *EngramID      // backing reference; nil before a declaration has backing.
}

// CTU executes using shared Memory and MMU. Kernel holds the same Memory pointer.
// No scheduler map, claim loop, ready queue, or involuntary-fault handler table lives here.
// Dependency-index ownership is OPEN, so the former FlipIndex field is omitted.
// Execution-state synchronization is not implied by scheduler single-writer ownership.
type CTU struct {
	Self   c.Target
	Memory *Memory
	MMU    MMU
	Local  map[c.Target]Driver
	Out    Egress
}

// Fire executes a batch of already-claimed entries from Memory.Fds. Kernel
// provisions the entries; Scheduler claims their MoveIDs before handoff. The batch
// holds references to Memory-owned entries, not a second authoritative table.
// For each entry, check declared access and construct ExecCtx from its Move and
// Register before invoking Move.Op.Fire. No executable lookup occurs in Fire.
// Batch/entry lifetimes and event delivery remain OPEN; no completion barrier.
func (ct *CTU) Fire(fds []*Fd) { panic("PROPOSAL: batch dispatch and event delivery") }

// Fd is the resident executable and its live return frame. Move.Fd is its key in
// Memory.Fds. Kernel provisions both together; Memory owns their storage.
type Fd struct {
	Move     Move
	Register *Register
}

// Memory owns one fd table keyed directly by canonical MoveID. There is no
// separate register table or InvocationKey-to-fd translation. Kernel and CTU
// share this instance. Installation/selection methods and lifetime rules remain
// to be specified; the main loop must not bypass them through direct map access.
// Scope and sharing for identical action/params, including multiple parents, are OPEN.
type Memory struct {
	Fds map[MoveID]*Fd
}

// Resolve is retained as a provisional pre-handoff lookup seam, not a CTU dispatch
// step. The fd entry now provides a home for both results. Installation, selection,
// and the unknown-fd/missing-backing result remain to be specified.
func (r *Memory) Resolve(fd MoveID) (*Move, *Register) {
	panic("PROPOSAL: fd resolution and miss result")
}

// Land fills the childFd entry in Fds[parentFd].Register, preserving declared order.
// Ret.Fire supplies both addresses from its Move and emits Landed after publication.
// In this serial probe, identical repeats succeed, conflicts and ambiguous or
// absent destinations fail without mutation. Copy the bytes on first publication
// so producer buffer reuse cannot change the stored value. This does not validate
// contracts or establish readiness; Ret.Fire must do that before publication.
// Synchronization and consumer-side mutation/lifetime rules remain OPEN.
// Root returns have no parent and require a separate result destination.
func (r *Memory) Land(parentFd MoveID, childFd MoveID, value c.Observation) c.Outcome {
	if r == nil || r.Fds[parentFd] == nil || r.Fds[parentFd].Register == nil {
		return c.Failure
	}
	frame := r.Fds[parentFd].Register
	index := -1
	for i := range frame.Children {
		if frame.Children[i].Fd == childFd {
			if index != -1 {
				return c.Failure
			}
			index = i
		}
	}
	if index == -1 {
		return c.Failure
	}
	if old := frame.Children[index].Value; old != nil {
		if old.Segment == value.Segment && bytes.Equal(old.Content.Bytes, value.Content.Bytes) {
			return c.Success
		}
		return c.Failure
	}
	value.Content.Bytes = bytes.Clone(value.Content.Bytes)
	frame.Children[index].Value = &value
	return c.Success
}

// MMU owns access checking. Verification dependencies and indexing remain OPEN;
// it contains no authority-minting or opcode execution behavior.
type MMU struct{}

// Check only checks access and returns a verdict. It never invokes Op.Fire,
// dispatches a driver, or changes scheduler state. Fire handles the verdict before
// execution. Authority must hold at the actual access. Denial/fault event delivery
// remains OPEN; it must not silently fall through or invent a Materialize fault.
// CTU obtains (target, action) through Touching; MMU remains opcode-agnostic.
func (m *MMU) Check(h c.Horizon, target c.Target, action c.Name) Verdict {
	panic("PROPOSAL: access checks")
}

// REMEMBER is kernel persistence work, distinct from scheduler-record retirement.

// Driver is synchronous. CTU must preserve both its observation and Outcome when
// adapting completion. The old partial adapter discarded Outcome and is removed
// until failure/fault delivery is specified.
type Driver interface {
	Decl() DriverDecl
	Perform(action c.Name, in []c.Observation) (c.Observation, c.Outcome)
}
type DriverDecl struct {
	Target  c.Target
	Actions []c.Name
}
type Verdict int

const (
	Authorized Verdict = iota
	Denied
	NeedsConfirm
	NotResident
)

// Egress carries portable moves. The receiving ingress splits on opcode: incoming
// Ret reaches return behavior before Memory publication and the Landed notification.
// The receiving adapter/signature remains OPEN; Landed itself never republishes.
type Egress interface {
	Egress(from c.Target, mv Move) c.Outcome
}

// Handler is a kernel recovery seam. Scheduler routes Faulted here; the handler
// may install backing through shared Memory but cannot recursively execute the
// faulted move or mutate scheduling maps. Resolution notification is still OPEN.
type Handler interface{ Handle(f FaultFrame) c.Outcome }

// FaultFrame retains the move, its return frame, and authority during recovery.
// Missing is the exact engram requested by Materialize. No Context result channel
// is needed: kernel provisioning now uses shared Memory's installation interface.
type FaultFrame struct {
	Vector   Kind
	Move     *Move
	Register *Register
	Under    c.Horizon
	Missing  *EngramID
}

// Kernel coordinates and provisions. Scheduler alone writes lifecycle state;
// Memory owns resident structures; CTU executes. FS is the SAME filesystem Driver
// mechanism used for ordinary file work, supplied at boot with kernel authority.
// Its concrete actions/argument encoding are OPEN; there is no EngramStore/Pager
// adapter between Kernel.Load and the driver.
type Kernel struct {
	Sched       Scheduler
	Memory      *Memory
	FS          Driver
	StorageRoot string           // configurable application-data directory; default resolution OPEN.
	Traps       map[Kind]Handler // kernel recovery handlers, not CTU-side dispatch.
}

// Load resolves one exact EngramID to a file under StorageRoot, calls FS directly,
// decodes/verifies the contents, and installs the engram through Memory. Only after
// installation and readiness may Scheduler release waiting moves. Missing files,
// I/O failures, or invalid contents must not report successful installation.
// The Outcome signature is provisional: detailed error results, engram encoding,
// Memory.Install's durable argument type, shared-load dedup, and wakeup remain OPEN.
// There is no directory search, live-frame page-in, or silent compile on load failure.
// Filesystem access and loading code must be available without paging an engram in.
func (k *Kernel) Load(id EngramID) c.Outcome {
	panic("PROPOSAL: exact-ID filesystem load, verify, and Memory installation")
}

// Bound creates workflow authority. Scheduling and execution cannot widen a
// horizon on their own. Signing, expiry, and confirm policy remain OPEN.
func (k *Kernel) Bound(root EngramID) c.Horizon {
	panic("PROPOSAL: mint horizon at provisioning")
}

// Boot installs recovery handlers inside the kernel. Runtime separately wires
// shared Memory. Actual fault/event routing and readiness delivery remain OPEN.
func (k *Kernel) Boot() {
	if k.Traps == nil {
		k.Traps = make(map[Kind]Handler)
	}
	k.Traps[Materialize] = materializeVec{k}
	k.Traps[Confirm] = confirmVec{k}
	k.Traps[Remember] = rememberVec{k}
}

// Materialize is routed by Scheduler from Faulted. Load Missing, install backing,
// and notify readiness before the same MoveID is reclaimed. Recovery failure is
// accounted for explicitly. No direct CTU handler call or recursive re-drive.
type materializeVec struct{ k *Kernel }

func (h materializeVec) Handle(f FaultFrame) c.Outcome {
	panic("PROPOSAL: Kernel.Load and recovery notification")
}

// Confirm parks until human/kernel resolution. The confirmation policy and
// corrective-observation binding remain OPEN; a denial cannot widen authority.
type confirmVec struct{ k *Kernel }

func (h confirmVec) Handle(f FaultFrame) c.Outcome {
	panic("PROPOSAL: confirmation and recovery notification")
}

// Remember compiles discovered contracts and reusable structure for persistence.
// It must not persist live child fds as compiled dependency identities or conflate
// trace lifetime with value release. Engram/Body encoding remains OPEN.
type rememberVec struct{ k *Kernel }

func (h rememberVec) Handle(f FaultFrame) c.Outcome {
	panic("PROPOSAL: compile remembered structure and persist through FS")
}

// Runtime wires one shared Memory into Kernel and CTU. Scheduler belongs to Kernel,
// not a third independent peer. CTU returns execution events to Kernel.Sched.Apply;
// new invocations and asynchronous recovery notifications also drive coordination.
// Transport, worker layout, and lifecycle algorithms remain OPEN.
type Runtime struct {
	Memory *Memory
	CTU    *CTU
	Kernel *Kernel
}

// Boot wires the declared runtime Memory, allocating it if omitted. Supply Kernel
// and CTU before calling. Runtime.Memory is the authoritative boot-time reference;
// callers with preloaded memory provide it there. This does not start an event loop,
// select a storage directory, install a filesystem driver, or implement execution.
func (r *Runtime) Boot() {
	if r.Kernel == nil || r.CTU == nil {
		panic("engine: Runtime.Boot requires Kernel and CTU")
	}
	if r.Memory == nil {
		r.Memory = &Memory{}
	}
	r.Kernel.Memory = r.Memory
	r.CTU.Memory = r.Memory
	r.Kernel.Boot()
}
