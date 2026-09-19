# Proposed surface after the overnight probe

> [Latest checkpoint](checkpoint.md): opcode-specific readiness, the stubbed parent-contract boundary,
> and the egress routing question supersede older descriptions here where they differ.

**Date:** 2026-09-06  
**Status:** DESIGN REVIEW before the next implementation handoff. Accepted distinctions are recorded
first; unresolved interfaces follow. The earlier broad proposal is retained below as unapproved ideas.

## Latest decided loop: Pending, Move in / Move out

CTU evaluates readiness and executes what is ready from the handed-off entries. Scheduler coordinates
work and publishes returned observations through Memory. It does not first ask CTU to certify readiness
and then dispatch the same work back to it.

- **Pending replaces Claimed:** handed-off work is awaiting its return; the name does not assert readiness.
- **Issued is removed:** marking work Pending before Fire already establishes the handoff state.
- **Landed is superseded:** a Move carrying Ret supplies the result to scheduler ingress. Scheduler-side
  return processing publishes through Memory; CTU does not publish and then emit a separate Landed event.

```text
Scheduler marks entries Pending → CTU.Fire(batch)
  → CTU executes ready work and emits Moves
  → scheduler ingress splits on opcode
      Call → resolve/provision
      Ret  → publish observation through Memory
  → continue
```

Still open: Fire's return/delivery signature, how unready entries are retained or reported and awakened,
involuntary-fault representation, and completion/duplicate handling. No new result wrapper is selected.

This supersedes the earlier claim-before-readiness and publication-before-Landed descriptions below.
Issued has been removed from code. The remaining Claim/Claimed/Landed/Event probe surface awaits
reconciliation with these decisions before the next implementation handoff.

## Accepted revision: kernel hands executable fd entries to CTU

The accepted handoff now uses the following shape, reflected in `engine.go`:

```go
type Fd struct {
    Move     Move
    Register *Register
}

type Memory struct {
    Fds map[MoveID]*Fd
}

func (ct *CTU) Fire(fds []*Fd)
```

Kernel provisions Move and Register together through Memory. Scheduler claims the entries' MoveIDs,
then the claimed entries are handed to CTU.Fire. This batch references entries in the one Memory-owned
fd table. It is not the entire table and does not transfer scheduling ownership to CTU.

CTU executes the entries it receives: check access, construct ExecCtx using `&entry.Move` and
`entry.Register`, then invoke the move's Op.Fire. Resolve is no longer part of CTU dispatch; any
lookup needed to prepare the batch happens before handoff. The main loop still coordinates through
specified methods. Installation/selection methods and the event return connection remain open.

This revision replaces the earlier proposed Execution wrapper, private executions table, and
ID-based CTU.Fire below. It does not approve the other added structures, methods, validator profile,
or recovery extension. Those remain suggestions for discussion, not implementation requirements.

## Decided: Move is the universal wire envelope

One ingress receives a Move and dispatches on its opcode. The conceptual entry is
`Ingress(move Move)`; its receiver and result signature remain to be specified.

- `Move{Op: Call{...}}` carries the invocation; kernel provisions its resident Move/Register entry.
- `Move{Op: Ret{Value: observation}}` carries the return; ingress routes it to existing return behavior.
- Other opcodes require explicit ingress semantics before being supported.

Move crosses the wire. Fd holds the resident Move and Register. Registers and their live observation
state stay in Memory; they are not part of the wire envelope. No separate call/return envelope is needed.
Issued, Landed, and Faulted remain internal execution events for the scheduler.

Wire encoding still needs an explicit opcode discriminator and payload encoding for Go's Op interface.
This is a serialization decision, not a new conceptual envelope or a claim that remote transport is
implemented. The source of a new register's contract and child requirements remains a provisioning seam.

## Decided: one identity, resident lookup, then durable resolution

**MoveID and EngramID name the same identity.** The arriving Move carries the ID needed to resolve
its executable. It does not carry a register. This supersedes the earlier distinction between a
separate invocation ID and durable backing ID. The canonical action + ordered observation params
rule remains the starting point; encoding and how remembered revisions use that identity still need
resolution.

```text
Move → ingress splits on opcode
  Call → scheduler looks up MoveID in Memory
    resident → use the fd entry
    absent   → kernel resolves the same ID through the filesystem driver
                 found → provision executable + register in Memory
                 no remembered executable → declaration supplies executable behavior;
                                             kernel provisions the register and executable
  eligible entry → scheduler claims → CTU.Fire(batch)
```

Scheduler coordinates lookup and eligibility. On a Memory miss, kernel calls the filesystem driver
and installs its returned fd through Memory. CTU executes the claimed entries handed to it.

**Decided: Resolve lives on the filesystem driver and returns the existing Fd entity.**

```go
// Method on the filesystem driver.
Resolve(id MoveID) (*Fd, error)
```

The key is the same MoveID carried by the arriving Move. The returned Fd contains Move and Register;
no new resolution-result wrapper is needed. The method is not Scheduler.Resolve or a kernel-owned
decode operation. Kernel coordinates the call and installation; the filesystem driver resolves the
stored executable into the returned fd. There is no intervening store/pager layer.

The concrete absence/error convention and how returned data binds to the incoming Move's live
context remain open. A successful lookup does not itself install the fd, claim work, or dispatch it.
This supersedes the older kernel-owned read/decode/verify description. Kernel.Load, if retained as a
coordination method, delegates resolution to the filesystem driver.

**Creating a register is distinct from obtaining executable behavior.** The kernel can create the
live register when no remembered executable exists, but an empty register alone is not executable.
The declaration path must supply behavior before dispatch. Distinguish absence of remembered backing
from I/O failure or invalid stored content; those errors must not silently count as successful
resolution or automatically trigger declaration.

The entry retains the incoming execution's return address and authority; resolving remembered behavior
must not substitute a previously stored live parent/register. Reuse and multiple-parent behavior still
need a policy. A resident entry also does not, by itself, establish readiness or authorize re-execution.

**Code reconciliation pending:** engine.go still declares separate MoveID/EngramID types and an optional
Move.Engram reference. Those represent the superseded distinction. Consolidate them when settling the
resolution surface; do not treat the old separate fields as a requirement to add another identity.
Kernel.FS is still typed as the generic Driver, which has only Decl/Perform. The filesystem-specific
Resolve method must be exposed there before implementation; its interface name/composition is not
chosen by this decision. The older Memory.Resolve stub describes resident lookup, not this filesystem
operation, and its name/signature also needs reconciliation.

## Accepted distinction: ingress, dispatch, and completion

**Calling CTU.Fire(batch) is the handoff of claimed work.** No additional dispatch abstraction is
needed. One ingress brings work or results into kernel coordination and splits on opcode: Call and
Ret take their respective paths below. It does not insert unclaimed work straight into an execution
batch. The payload is decided: Move. The exact receiver and result signature remain open.

| Incoming item | Responsibility |
|---|---|
| Call | Resolve the arriving MoveID in Memory, then through kernel filesystem resolution or declaration if absent; provision as needed and establish readiness before claim. |
| Ret | Route the incoming observation to the existing return behavior for the returning move and its parent register, then reconsider parent readiness after publication. |


```text
Ingress → opcode split
  Call → resolve/provision fd entry → readiness → claim → Fire(batch)
  Ret  → existing return behavior → Memory publication → Landed
                                                       → parent readiness → claim → Fire(batch)
```

Ret.Fire continues to define return publication through Memory. Ingress must route to that behavior,
not implement another version of landing. An incoming return is not a new child invocation and must
not overwrite the existing move/register. How ingress supplies the execution context for that return
is still open. The Move input is decided; transport and scheduling policy for processing incoming
returns remain open.

**Landed is a post-publication event, not another ingress opcode.** It notifies the scheduler without
publishing the observation again. Scheduler handles lifecycle events, not an opcode switch. The call/return distinction belongs at
work/result ingress; Issued, Landed, and Faulted remain the execution-event vocabulary. CTU faults
still return through that event path to kernel recovery.

## Remaining issues before the next handoff

These are questions to settle together, not prescribed implementations. Resolve each with an agreed
surface or an explicit scope deferral before freezing the next work order. The earlier proposal below
offers possibilities; none of its additional types are implicitly required by this checklist.

| Open seam | Decision needed |
|---|---|
| Ingress | Move input is decided. Specify the ingress receiver/result, how Ret obtains the right move/frame, which other opcodes it accepts, and the opcode wire encoding. |
| Resolution, provisioning, and batch preparation | Filesystem Resolve(MoveID) (*Fd, error) is decided. Define absence/error conventions, the declaration handoff, and how resolved behavior supplies register requirements; Memory methods to install and obtain claimed entries. Fire itself is the handoff. Define what happens if preparation or handoff cannot complete. |
| Readiness and waiting | CTU evaluates and executes ready work. Define how unready entries wait and become eligible after scheduler publication without a busy loop. |
| CTU return delivery | Move in / Move out is decided; Issued and Landed are removed from the target model. Define Fire output transport, internal fault representation, and the Pending lifecycle. |
| Return computation and contracts | How child observations become the proposed return value; the minimum schema/validation semantics and representation of an unknown contract. No new computation opcode or validator interface is accepted yet. |
| Root result and loop entry | How the caller starts the local loop, receives the root observation or failure, and knows execution has finished. No ready work alone does not establish completion. |
| Non-success outcomes | How driver Outcome and observation are preserved, how access/validation errors are reported, and what scheduler transitions handle them without calling every failure a materialization fault. |
| Identity and reuse | MoveID = EngramID is decided. Consolidate the old type/field split; settle canonical encoding, remembered revisions, lookup scope, and repeated identical calls or multiple parents. Distinguish fixture IDs from an implemented identity rule. |
| Authority | What the minimal local run needs from MMU and provisioned capabilities; where verification dependencies come from if driver access is included. |
| Ownership and lifetime | How long handed-off entries and return context remain valid, who may mutate them, and what completion/cleanup retains. Confirm the initial serial/concurrent scope and duplicate-event policy. |

**Separate scope decisions:** implementing durable Engram/Body storage/loading and declaration, parked/recovery transitions, confirmation, REMEMBER, remote transport, and multicore execution
remain open. The current minimal work order defers these. Reaffirm that scope at handoff; do not silently
expand the local loop to solve them all.

The next handoff should freeze only the agreed types and methods needed for the chosen local path.
Implementation algorithms and incidental choices remain the implementer's responsibility.

## Earlier proposal — retained for review, not the accepted surface

This proposal supplies the missing state and handoffs identified in
[boundary-crossings.md](boundary-crossings.md). Its acceptance target is a serial, preloaded local
execution tree that evaluates two child operations, publishes their observations in declaration
order, computes a parent observation, and returns the published root through Runtime.Run.

The first implementation owns one execution tree per Runtime. All children are required. Calls
with multiple parents, repeated effects under the same fd, dynamic Call, remote execution,
concurrent access, durable loading, and automatic retries remain unsupported. Reject those inputs;
do not approximate them. The separate recovery extension below is not required for this first loop.

## Inventory: what is new and what changes

Every declaration below is labeled NEW, CHANGED, or UNCHANGED. Method declarations are signatures,
not Go implementations. Names from `c` refer to the existing v12 causal vocabulary.

| Kind | New declarations |
|---|---|
| Structs, required for the local loop | Execution, Resolution, Readiness, Compose, StepFailed |
| Interfaces, required for the local loop | ReadinessSource, ContractValidator, ReturnBuilder, CapabilityVerifier |
| Enum types | ResolutionStatus, ReadinessStatus |
| Structs, recovery extension only | RecoveryResult |
| Interfaces, recovery extension only | RecoveryService |

| Existing declaration | Proposed change |
|---|---|
| Memory | Replace the public Registers map with a private fd → Execution map; add a root fd. |
| Scheduler | Keep the sole lifecycle map; add a ReadinessSource dependency. |
| CTU | Add a ContractValidator dependency; Fire gains an event callback and error result. |
| MMU | Add a capability verifier and injected clock. |
| State | Add Waiting and Failed; Parked is reserved for the recovery extension. |
| Event | Extend its closed implementations with StepFailed. |
| Runtime | Add Run; keep its existing fields and shared-Memory boot ownership. |
| Existing methods | Resolve, Admit, Apply, and Fire change signatures; others are added as labeled below. |

MoveID, EngramID, Move, Register, ChildObservation, Contract, ExecCtx, Ret, Exec, Driver,
Issued, Landed, Faulted, Op, Touching, Verdict, Kernel, and Runtime field shapes remain unchanged.
Call and Trap remain explicit unsupported operations in the local profile. No new Attempt ID,
fd translation table, scheduler payload map, or durable Body type is introduced.

## 1. Memory owns executable descriptors and publication

```go
// NEW struct: the live execution record, not a durable engram.
type Execution struct {
    Move   Move
    Frame  Register
    Output *c.Observation // nil until validated publication, including for root
}

// CHANGED struct: one authoritative table, still keyed directly by MoveID.
type Memory struct {
    executions map[MoveID]*Execution
    root       MoveID
}

// NEW enum type and constants.
type ResolutionStatus uint8
const (
    UnknownMove ResolutionStatus = iota
    Resident
    MissingBacking
)

// NEW struct: a tagged result, not another storage record.
type Resolution struct {
    Status  ResolutionStatus
    Move    *Move
    Frame   *Register
    Missing *EngramID
}

// NEW methods.
func (m *Memory) InstallLocal(root MoveID, entries []Execution) error
func (m *Memory) FDs() []MoveID
func (m *Memory) Dependents(fd MoveID) []MoveID
func (m *Memory) Publish(fd MoveID, value c.Observation) error
func (m *Memory) RootResult() (c.Observation, bool)
func (m *Memory) Release() error

// CHANGED signature.
func (m *Memory) Resolve(fd MoveID) Resolution

// UNCHANGED signature; becomes a storage sub-operation of Publish.
func (m *Memory) Land(parentFd, childFd MoveID, value c.Observation) c.Outcome
```

`InstallLocal` is boot-time installation into empty Memory. It validates the entire input before
making any record visible. Require one root with nil ParentFd; one connected, acyclic tree; unique
fds; non-nil supported Ops; reciprocal ParentFd/Children relationships; no repeated child entry;
and no pre-populated child values or Output. Every non-root has exactly one parent. Reject missing
frames, disconnected records, replacement of an installed graph, and unsupported operation modes.
A non-nil Op is the explicit resident executable for this local profile; an Engram pointer alone is
not evidence that executable backing is installed.

Ownership transfers through copies of descriptors, slices, schema bytes, capability signature bytes,
and observation bytes. Ops/builders are trusted immutable program objects; callers must not mutate
captured state or perform effects in a ReturnBuilder. A future durable decoder is a separate concern.
`Resolve` returns read-only snapshots, not mutable storage pointers. UnknownMove carries no pointers;
Resident carries Move and Frame; MissingBacking carries the retained descriptor/frame and the exact
missing EngramID, with no executable Op. MissingBacking becomes constructible only with the recovery
extension; InstallLocal rejects unresolved entries.

`Dependents` uses the stored ParentFd: zero or one result in this tree profile. No dependency index
is needed. `FDs` returns a deterministic copy, never the table itself.

`Publish` is the common storage boundary for root and child returns. After CTU validation it checks
that the source exists, and that the destination is consistent with the installed tree. For a child,
it uses Land to fill the parent's declared entry; for a root, there is no parent write. In the same
serial operation it records the source's Output. Validate every possible failure before either
write. An identical repeat succeeds; conflicting repeats fail without mutation. Both destinations
own their bytes. Output is needed to distinguish root publication, detect repeated returns, and
retain source completion evidence; it is not a scheduler state flag.

Raw Land preserves the probe's storage semantics, but cannot establish a complete return by itself.
Production operations use Publish, not Land directly. Public Land is retained for compatibility;
callers outside Memory must not use it as a runtime publication path. A later encapsulation change
could make it private. RootResult returns a copy of the published root observation and an availability
flag. Root failure is returned by Run, not disguised as an observation.

Memory remains available for inspection after Run, including failure. Release explicitly discards the
entire tree after the caller has copied the result and decided to discard diagnostics. It is not
per-child consumption or durable persistence. Call it only through Runtime.Close after execution has
returned; the first implementation has no eviction while running.

**Identity scope:** one tree per Memory avoids cross-run fd collisions. The caller supplies canonical
MoveIDs under the existing Action + ordered Params rule; this proposal does not change its encoding.
InstallLocal rejects duplicate fds instead of silently sharing executions. Root and internal Compose
moves also need declared invocation identities supplied by the preloader. A public invocation compiler
still requires an agreed canonical encoding and action namespace; the local proof uses explicit
address fixtures and makes no claim to implement that compiler.

## 2. CTU determines readiness; Scheduler owns lifecycle

```go
// NEW enum type and constants.
type ReadinessStatus uint8
const (
    InvalidWork ReadinessStatus = iota
    WaitingForChildren
    Runnable
    NeedsBacking
)

// NEW struct: ephemeral query result.
type Readiness struct {
    Status  ReadinessStatus
    Missing *EngramID // only NeedsBacking
    Err     error     // only InvalidWork
}

// NEW interface, implemented by CTU using Memory and contract validation.
type ReadinessSource interface {
    Inspect(fd MoveID) Readiness
    Dependents(fd MoveID) []MoveID
}

// CHANGED struct.
type Scheduler struct {
    states map[MoveID]State
    Work   ReadinessSource
}

// NEW State constants, alongside Ready, Claimed, Completed.
// Waiting: known work with unsatisfied required child inputs.
// Failed: terminal operation/validation failure.
const (
    Waiting State = Completed + 1 + iota
    Failed
)

// NEW CTU methods; Dependents delegates to Memory.
func (ct *CTU) Inspect(fd MoveID) Readiness
func (ct *CTU) Dependents(fd MoveID) []MoveID

// CHANGED signatures.
func (s *Scheduler) Admit(fds []MoveID) error
func (s *Scheduler) Apply(e Event) error

// NEW methods.
func (s *Scheduler) ReadyFDs(limit int) []MoveID
func (s *Scheduler) Quiescent() bool
func (s *Scheduler) Reset() error

// UNCHANGED signatures.
func (s *Scheduler) Claim(fds []MoveID) []MoveID
func (s *Scheduler) Retire(id MoveID)
```

Inspect resolves the descriptor and validates the supported return contract. Unknown fd or invalid
contract is InvalidWork; absent backing is NeedsBacking. Otherwise every declared child must have
arrived and satisfy that child's installed return contract. Missing arrivals produce WaitingForChildren;
invalid arrivals produce InvalidWork. Zero-child operations are runnable once their contract is supported.
Checking availability never executes an Op, driver, or ReturnBuilder. Authority is rechecked at execution.

Admit consults Work itself. It records Waiting or Ready only from those results, never from a caller's
assertion. It does not regress Claimed, Completed, or Failed. In the local profile NeedsBacking is an
explicit unsupported error; it does not claim or execute the move. InvalidWork marks that fd Failed and
returns the diagnostic. Partial initial admission is retained on error and Run stops before dispatch.

ReadyFDs returns a deterministic bounded snapshot. Claim remains the existing single-writer transition
before dispatch. On Landed, Apply records Completed and rechecks that fd's dependents through Admit.
On StepFailed it records Failed and returns an error to terminate this run. Unrelated outstanding work
is retained for diagnosis, not reported as completed. Issued never creates a claim or regresses completion.

Quiescent means every installed/admitted lifecycle record is terminal (Completed or Failed), not merely
that the map is empty or no fd is ready. Run calls initial Admit for all installed fds before using it.
Retire only removes Completed records after Quiescent; it is not used during Run. Reset clears records
only during explicit post-run cleanup. Runtime is single-use until Close; no re-entry or retry of a
partially failed run is permitted by this profile.

## 3. Explicit dispatch event transport and failure preservation

```go
// CHANGED method: serial dispatch with synchronous per-operation event delivery.
func (ct *CTU) Fire(fds []MoveID, emit func(Event) error) error

// NEW struct and NEW implementation of the existing Event interface.
type StepFailed struct {
    Fd          MoveID
    Observation c.Observation
    Outcome     c.Outcome
    Err         error
}
func (StepFailed) isEvent() {}
// NEW methods: allow Run's error result to preserve the complete failure event.
func (e StepFailed) Error() string
func (e StepFailed) Unwrap() error
```

Fire resolves each claimed fd, rechecks readiness and declared access via Touching/MMU, invokes its
Op, and immediately calls emit with the resulting event. No scheduler reference is added to CTU or
ExecCtx. A nil callback is rejected before any execution. The kernel-owned Run supplies Sched.Apply;
this callback is trusted in-process ingress, not an externally writable completion API.

A local operation must return a final Landed or StepFailed. Issued-only completion, Faulted recovery,
and remote delivery are explicitly unsupported in the first loop. They cannot leave Run spinning.
A callback error stops the batch and returns to Run. Already-published output and remaining claims
are retained; there is no automatic re-drive after a possibly executed effect. A process crash or
panicking driver is outside the first loop's recovery guarantee.

Exec checks its driver's declaration and access before Perform. Success passes the observation to
Ret.Fire using the same ExecCtx. Failure, NotYet, and Fault all produce StepFailed preserving the
original Observation and Outcome, with a diagnostic that distinguishes an actual driver failure from
an unsupported wait/recovery result. Neither Fault nor NotYet is invented into Materialize. The local
runner terminates on those results because the Driver interface supplies no wakeup or fault identity.
Validation or authorization failure also uses StepFailed, with Outcome Failure and a diagnostic.
If validation rejects a successful driver's output, preserve that output and its original Success
Outcome in StepFailed; Err describes the return-stage failure. Event kind, not Outcome alone, determines
scheduler failure. Apply returns StepFailed as its error, preserving these fields for `errors.As`;
Unwrap returns Err. Exec must wrap a Ret validation/publication failure with the original driver's
Observation and Outcome rather than discarding them.

## 4. Construct, validate, and publish returns

```go
// NEW interfaces.
type ContractValidator interface {
    Check(contract Contract) error
    Validate(contract Contract, value c.Observation) error
}

type ReturnBuilder interface {
    Build(children []ChildObservation) (c.Observation, error)
}

// NEW struct and implementation of existing Op.
type Compose struct {
    Builder ReturnBuilder
}
func (o Compose) Fire(x *ExecCtx) Event

// CHANGED struct: existing fields retained, one dependency added.
type CTU struct {
    Self      c.Target
    Memory    *Memory
    MMU       MMU
    Local     map[c.Target]Driver
    Out       Egress
    Contracts ContractValidator // NEW field
}

// UNCHANGED signature; newly specified behavior.
func (o Ret) Fire(x *ExecCtx) Event
```

Compose receives the ordered child observations from the resolved frame, calls a pure ReturnBuilder,
then delegates the constructed observation to Ret.Fire. Builders receive defensive copies; they do
not traverse Memory, schedule children, or capture an executable graph. This adds the missing explicit
computation step without treating a schema as an output program. Ret remains useful for constant leaves;
Exec supplies driver results through the same return path.

Ret verifies that all required children have arrived and are valid, validates Value against the source
frame's Contract, calls Memory.Publish, and only then returns Landed. Missing inputs, invalid schema,
invalid output, or publication conflict must never emit Landed. Return errors become StepFailed.
Use the same checks as Inspect; readiness is not permission to skip return validation.

For the first implementation, configure one explicit contract profile: Content.Bytes is a complete
JSON value, and Schema is a JSON object containing exactly one `type` property with value `string`
or `boolean`. Reject absent schemas, extra schema keywords, unknown types, trailing JSON, and invalid
JSON. Segment remains observation metadata and is not constrained by this profile. This deliberately
small validator is not advertised as general JSON Schema support. A subsequent validator can implement
a broader documented dialect through the same interface. Boot requires a validator; nil does not mean
accept everything. This profile is a proposed local choice, not a change to the existing specs.

## 5. Capability verification has explicit dependencies

```go
// NEW interface: checks cryptographic authenticity against provisioned trust material.
type CapabilityVerifier interface {
    Verify(cap c.Capability) error
}

// CHANGED struct; time is from the standard library.
type MMU struct {
    Verifier CapabilityVerifier
    Now      func() time.Time
}

// UNCHANGED signature.
func (m *MMU) Check(h c.Horizon, target c.Target, action c.Name) Verdict
```

MMU requires exact target/action match, a future ExpiresAt, and successful signature verification for
at least one matching grant. Treat ExpiresAt as Unix seconds in this proposed profile. An expired or
invalid matching grant does not prevent another valid grant from authorizing access. Missing dependencies
are boot errors and checks fail closed. No match yields Denied; the local profile does not manufacture
NeedsConfirm or NotResident from an invalid grant. Kernel boot supplies verification dependencies;
workflow input cannot replace them. Verifier owns its key source and canonical signature preimage;
a production verifier requires that protocol to be specified. Tests may inject a deterministic verifier
but must label it as a fixture. A pure Ret/Compose demonstration needs no driver access or grants.

## 6. A complete orchestration path through methods

```go
// NEW methods on existing Runtime.
func (r *Runtime) Run(root MoveID, entries []Execution) (c.Observation, error)
func (r *Runtime) Close() error
```

Boot retains its shared-Memory wiring and additionally wires Scheduler.Work to CTU, checks required
dependencies, and rejects conflicting preconfiguration. Run requires an empty, unused runtime and
installs the graph itself; callers do not mutate Memory or scheduler maps. Existing probe boot fixtures
would need migration to InstallLocal if this replacement surface is accepted.

The proposed Run body is expressible with these methods:

```text
Boot
Memory.InstallLocal(root, entries)
Scheduler.Admit(Memory.FDs())
repeat:
    candidates = Scheduler.ReadyFDs(batchLimit)
    if candidates is empty:
        value, published = Memory.RootResult()
        if published and Scheduler.Quiescent(): return value, nil
        return explicit stalled/unsupported-recovery error
    claimed = Scheduler.Claim(candidates)
    CTU.Fire(claimed, Scheduler.Apply)
    on error: return the error, retaining state and published evidence
```

Run does not inspect frames, mutate states, call Op.Fire, or construct root observations itself.
All parent wakeups originate in Apply after publication. A bounded nonzero batchLimit is a local
constant, not persistent architectural state. Successful Run retains records and observations;
Close, called after Run returns, invokes Scheduler.Reset and Memory.Release. The serial API forbids
concurrent/reentrant Run, Close, and component mutation. This gives an explicit lifetime policy without
adding per-value reference counts. Close after an errored run explicitly abandons outstanding work;
it does not certify that an external effect was rolled back.

### Proposed acceptance test

Install root A as Compose with a builder that compares two JSON document strings and returns a JSON
boolean. Install B and C as constant Ret leaves, each with a JSON-string contract. A declares Children
in order [B, C] and has a JSON-boolean contract. Choose fixture fd ordering so ReadyFDs dispatches C
before B. Run must publish C into entry 1, B into entry 0, admit A only after both are valid, execute
Compose, validate and publish A, then return that actual root observation.

Check equal and unequal documents, reversed leaf completion, missing/malformed contracts, invalid
child output, duplicate/conflicting publication, invalid graph wiring, unclaimed completion, unknown fd,
and preserved driver non-success outcomes. Verify that denied/expired/invalid-signature access never
calls Perform. Assert root unavailability before publication, publication before Landed, retained
completion after late Issued, and no second effect execution after callback failure. No test should
seed Ready or call Ret directly and describe that as a successful Run.

## 7. Recovery extension — separate from the preloaded loop

The new resolution result makes a missing engram distinguishable, but does not by itself implement
loading. These additional declarations propose the control boundary once a durable format is agreed:

```go
// NEW struct, extension only.
type RecoveryResult struct {
    Fd  MoveID
    Err error
}

// NEW interface, extension only; implemented by Kernel.
type RecoveryService interface {
    Recover(f FaultFrame) RecoveryResult
}

// NEW Memory method, extension only: bind decoded resident behavior to an existing descriptor.
func (m *Memory) Bind(fd MoveID, backing EngramID, op Op) error

// NEW Scheduler method and dependency, extension only.
func (s *Scheduler) Resume(result RecoveryResult) error
// Scheduler gains Recovery RecoveryService and State gains Parked.
```

On Faulted, Apply first records Parked, constructs FaultFrame from the retained Memory resolution
through an expanded ReadinessSource recovery query, and calls Kernel.Recover through RecoveryService.
That query is an additional proposed interface method:

```go
// NEW method on ReadinessSource and CTU, extension only.
RecoveryFrame(event Faulted) (FaultFrame, error)
```

The query checks the event fd and missing ID against the retained descriptor before exposing a frame.
Kernel loads and verifies the exact engram, then binds its decoded executable through Memory. Bind
requires matching backing identity and cannot change fd, parent, frame, inputs, or Horizon. Resume
accepts results only for Parked work; failure marks Failed, success rechecks readiness before making
it claimable. Reissue preserves the same fd. No CTU-side handler or recursive Fire is involved.
For this serial extension, recovery returns synchronously through Apply; asynchronous recovery needs
an explicit queued delivery contract and stale-result discipline before it can use Resume safely.

Kernel.Load's existing Outcome-only signature cannot retain detailed decode/I/O failure causes.
The extension proposes changing it to `Load(id EngramID) error`, with typed/wrapped errors for
missing object, I/O failure, invalid content, and installation failure. The durable Engram/Body,
decoder, filesystem action encoding, and shared-backing installation API still require their own
proposal; Bind alone is not an engram store. Until those are defined, Recover returns an explicit
unsupported error. Confirm, REMEMBER, shared-load deduplication, and general Driver Fault/NotYet
resumption likewise remain unsupported. This extension specifies the missing ownership and return
path without claiming that durable loading is ready to implement.
