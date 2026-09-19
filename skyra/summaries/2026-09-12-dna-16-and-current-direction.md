# Skyra: DNA-16 and the current direction

**Date:** 2026-09-12  
**Reconciled:** 2026-09-16 against the current ownership, struct, and minimal-loop proposals.  
**Status:** Conceptual and implementation handoff. DNA-16 remains the executable
reference. The September 16 proposals describe the current direction, not an
implemented replacement. September 12 supplies retained Surface/Possibility/Horizon
background where it has not been superseded; exact contracts still need reconciliation.

## Executive summary

DNA-16 proved a local continuation loop across real component boundaries:

```text
Move → CTU → Scheduler → Continuation Controller
     → address-resolved Driver → DriverEmission
     → normalized []Move → Scheduler transition
```

Its integrated fake Codex App Server proof turns a dynamic tool call into a
child Skyra invocation, returns the child result to the stable parent and Codex
thread, and converts a clean final message into a Return. A separate real-process
handshake performed no inference. Provider protocol details remain opaque to the
scheduler.

The current CTU direction has three main services:

```text
StateOwner    sole owner of canonical state and both Move mailboxes
Interpreter   sequential Invoke/Return interpretation and dependency semantics
Supervisor    dispatch, bounded concurrent workers, progress and retry supervision
```

Both processing components use `Take(ctx, self)`, with identity and permitted
destination injected at boot. Interpreter transfers work to Supervisor; Supervisor
transfers emitted Moves back to Interpretation. StateOwner commits destination
mailbox placement and custody together before acknowledging. Transferring an
existing Move and introducing an emitted Move use the same outer mechanics while
preserving their distinct identities.

There is no standalone Admission service or Dependency Gate. Custody is a state
profile and protocol, not another service or Move owner. Its bounded ownership
set is `Interpretation | Supervisor`, with no pending-acceptance status.

The continuation map is the map of Moves. StateOwner also holds possibilities,
custody, worker records, system profiles, and user-registered Surfaces such as
Codex. Surface-native state remains with its Surface. Canonical worker state lives
in StateOwner; Interpreter and Supervisor retain only their processing machinery
and temporary working data.

Surfaces have globally unique identities, recursive independently addressable
capabilities, and Surface-bound input/output contract pairings. Move has Invoke
and Return data; interpretation belongs to Interpreter. A Return carries one
observation, and egress produces one Move per emission. Both opcodes share the
same authorized dispatch path through the Possibility's Horizon.

Kernel holds Trap and SysCall Surfaces. Move-level failures become ordinary
failure-report Moves for Trap; Trap determines corrective work. State proposals
reconcile revisions before commit. A stale Supervisor handoff must not rerun
external execution. Independent monitoring handles workers that fail to return.

The latest loop is a conceptual sketch. Exact state methods, request/result
records, worker lifecycle, response correlation, shutdown, and Horizon enforcement
remain to be settled before an implementation work order replaces DNA-16.

---

# Part I — What DNA-16 implements today

## 1. Purpose and status

DNA-16 implements the driver-host and transport work order. W0–W7 are marked
complete within the local scope:

- runtime transition contracts;
- opaque driver state;
- framed JSONL transport;
- one-process host;
- Codex App Server driver;
- native-tool detection and launch containment;
- conservative recovery behavior;
- integrated fake-App-Server proof.

It is not a production-runtime claim. It does not prove live authenticated
inference, production credential isolation, OS sandbox installation, durable
recovery across Skyra process loss, or safe provider request replay.

## 2. Package outline

| Package | Implemented responsibility |
|---|---|
| `protocol` | Canonical runtime values and scheduler events. |
| `ctu` | Boot composition and public `Submit`/`Snapshot` API. |
| `scheduler` | Resident continuations, registers, atomic transitions, selection, faults, and root completion. |
| `controller` | Validation, address resolution, dispatch, driver callbacks, and normalization. |
| `driver` | Driver/transport contracts and address registry. |
| `kernel` | Boot-time endpoint registration through the privileged registrar. |
| `host` | Concrete OS process and JSONL stream ownership. |
| `codex` | Codex App Server protocol, lifecycle, tool mapping, and launch policy. |
| `probe` | Deterministic root and echo drivers for the continuation-loop proof. |
| `internal/support` | Detached-copy and register helpers. |
| root `api.go` | Compatibility facade over the package-owned contracts. |

This split made ownership explicit without yet solving the universal surface
model.

## 3. Canonical protocol

### Move and target

```go
type Move struct {
	Root   PossibilityID
	Target Target
	Op     Opcode
	Stamp  AuthorizationStamp
}

type Target struct {
	Invocation InvocationID
	Entity     EntityID
	Address    string
	Host       HostID
}
```

`Root` correlates a workflow. `Target` currently combines live continuation
routing with endpoint placement. `Stamp` records the runtime authorization
verdict. The final semantic surface path is not represented yet.

### Invoke and Return

```go
type Invoke struct {
	InvocationID InvocationID
	Action       Action
	Params       []Observation
	ReturnFd     InvocationID
	Register     Register
	Horizon      Horizon
	Focus        []Capability
}

type Return struct {
	Observation Observation
}
```

`InvocationID` is stable for the lifetime of one invoked executable. A child
tool call gets a new invocation ID; `ReturnFd` names its stable parent. A Return
contributes to an existing invocation rather than creating another thread-like
identity.

This corrected the older formulation that every Move, including a Return, must
have a fresh invocation ID.

### Register

```go
type Register struct {
	Contract Contract
	Children map[InvocationID]RegisterEntry
	Order    []InvocationID
}
```

The register stores declared child invocations and their observations. The map
provides correlation; `Order` preserves declaration order. Stored child Moves
have empty nested registers, preventing recursive continuation snapshots.

### Capability, Horizon, and Focus

The compiled capability names a target and Action plus expiry and signature. A
Horizon is a capability list. Focus is checked as a subset of Horizon and sent
to the driver as the advertised tool set.

Horizon still travels on every DNA-16 Invoke, and child admission verifies that
it equals the parent's Horizon. This is working probe behavior, not final
authority placement.

### Runtime events

DNA-16 separates continuation data from placement/execution feedback:

- `MoveBatch` admits one atomic set of normalized moves.
- `DispatchResult` records endpoint placement success or failure.
- `AdmissionFailure` rejects external input without mutating a claimed
  invocation.
- `ExecutionFailure` faults a known source and handoff sequence.
- `DriverTransition` atomically replaces driver state and applies moves,
  commands, or failure.
- `RootResult` reports the one root observation or terminal fault.

Handoff sequence numbers make stale driver feedback harmless. They are runtime
correlation, not continuation-language meaning.

## 4. CTU boot and public boundary

The implemented CTU interface is:

```go
type CTU interface {
	Submit(ctx context.Context, move Move) error
	Snapshot(id InvocationID) (Continuation, bool)
}
```

`NewCTU` currently:

```text
creates DriverRegistry
registers kernel at its configured address
constructs a read-only Resolver
constructs ContinuationController
constructs Scheduler
gives privileged Registrar to Kernel.Initialize
resolves and verifies the declared root endpoint
publishes the CTU
```

`Submit` accepts one root Invoke with an empty frame, assigns an invocation ID
when absent, validates it through the controller, and admits it as a MoveBatch.
`Snapshot` returns a detached continuation copy.

The current kernel embeds Driver and exposes `Initialize(Registrar)`. The
fixture kernel registers one root plus a fixed local endpoint list. Its own
ingress and egress reject execution.

## 5. Registry and current driver boundary

```go
type Driver interface {
	Ingress(context.Context, DriverInput, outputCallback) error
	Egress(raw string) ([]Move, error)
}

type DriverRegistration struct {
	Driver Driver
	Input  InputValidator
}
```

`DriverRegistry` maps one canonical string address to one registration.
`Resolver` is the read door held by the controller. `Registrar` is the
privileged write door held by the kernel during initialization. Unexported
adapter types prevent a Resolver from being asserted back into a Registrar.

`InputValidator` is a first-class DNA-16 endpoint component with schema
declaration and Move validation. The current direction removes it as a peer
abstraction: driver ingress owns malformed operand interpretation and egress
normalizes the error.

## 6. Continuation Controller

The controller is stateless across invocations. It currently:

1. validates broad Move shape;
2. resolves `Target.Address`;
3. calls the endpoint validator;
4. records successful placement before driver execution;
5. calls ingress outside the scheduler lock;
6. serializes that ingress call's output callbacks;
7. normalizes raw driver output through the same driver's egress;
8. assigns child invocation IDs and binds root, `ReturnFd`, register, and
   inherited Horizon;
9. validates all normalized Moves before one atomic transition;
10. executes host commands only after successful commit.

The controller, not the provider, creates live runtime correlation. A driver can
request child work semantically but cannot forge its runtime identity or parent
route.

## 7. Scheduler

The scheduler is the sole owner and writer of resident continuations:

```go
type Continuation struct {
	Move             Move
	Handoff          Handoff
	State            State
	DriverState      DriverState
	Context          context.Context
	Result           *Observation
	CompiledRegister *CompiledRegister
	Resumed          bool
}
```

Its one transition door mutates state under a mutex. Driver calls and root
completion callbacks happen after the lock is released.

The implementation currently combines:

- admission and invocation uniqueness;
- parent/child linkage;
- register mutation and ordered observations;
- Horizon checks and authorization stamps;
- placement and handoff lifecycle;
- opaque driver-state replacement;
- dispatch selection;
- one parent resumption;
- workflow fault propagation;
- one root-result notification attempt.

A `DriverTransition` is transactional. The scheduler snapshots its resident
map, replaces the source state, applies all moves, and restores the old map if an
effect is rejected.

The general continuation interpreter is not implemented. `CompiledRegister`
is a placeholder and `interpret` returns `Waiting`. The local proof supports
one child declaration wave and one parent resumption after all child slots fill.
Multiple waves, partial joins, contract mismatch, cancellation, and general
terminal laws remain open.

## 8. Opaque state and post-commit commands

```go
type DriverState struct {
	Version uint32
	Payload []byte
}

type DriverInput struct {
	Move  Move
	State DriverState
}
```

The scheduler owns storage and copying; the selected driver owns payload meaning.
Version zero plus empty bytes is the canonical empty state. Nonempty state
requires a nonzero version.

`DriverEmission` carries replacement state, optional raw output, host commands,
or an atomic failure. Failed emissions cannot also carry output or commands.

`HostCommand` makes state commit happen before an external send. It does not
yet implement a durable outbox or exactly-once delivery across a crash.

## 9. Host and transport

`JSONLTransport` implements `Send`, `Receive`, and `Close` over complete
byte frames. It serializes concurrent writes, appends one newline, detaches
received bytes, rejects empty frames, and closes idempotently to release blocked
reads. It understands neither Codex nor Moves.

`ProcessHost` owns:

- `exec.Cmd` and the OS process;
- stdin/stdout pipes and JSONL transport construction;
- captured stderr;
- wait/exit status;
- deterministic close and kill behavior.

`ProcessSpec` supplies an explicit path, arguments, working directory, and
environment. There is no shell expansion.

This implementation proved resource ownership. Its special host/transport
abstractions are now candidates to move behind recursive surfaces. In the new
direction, authorized handlers belong to Horizon and native state belongs to the
Surface that owns it.

## 10. Codex App Server

DNA-16 uses long-lived Codex App Server over stdio JSONL, not one-shot
`codex exec`. The driver initializes the server and runs one receive loop.
The local policy gives one Codex process to one possibility. One active Skyra
invocation owns one Codex thread and at most one turn in the first build.

Opaque versioned Codex state records:

- phase;
- thread and turn IDs;
- rendered input;
- dynamic tool name → Skyra capability mapping;
- pending tool request and call IDs;
- final text;
- contamination.

The scheduler stores these bytes and understands none of them. Live callbacks and
request-routing maps remain private to the driver.

The state machine is:

```text
startingThread
  → startingTurn
  → running
  → waitingForChild
  → running
  → completed

unsafe or unknown effect
  → faulted + contaminated
```

On first ingress, Focus becomes a complete `dynamicTools` snapshot with opaque
names such as `cap_000`. A dynamic `item/tool/call` becomes a child Invoke.
When its observation fills the parent's register, the driver answers the pending
Codex call and continues the same turn. A clean completed agent message becomes
the final Return.

Known native effects—including shell execution, file change, MCP, web, image,
and collaboration tools—contaminate and fault from `item/started`. Unknown
effect-bearing events fail closed. A later message cannot clear contamination.

Detection is not prevention. Launch policy also pins the Codex version, uses
strict configuration, never-approve/read-only settings, isolates `CODEX_HOME`,
provides an explicit executable search path, disables measured native features,
and inherits no user environment. Production OS containment remains deferred.

If restored Codex state claims live work but the matching host/request state
cannot be proven, DNA-16 faults. It does not silently create a new thread.

## 11. Probe and verification

The probe root emits two child Invokes in one batch, waits for their observations,
preserves declaration order, combines them, and Returns. Echo drivers may return
immediately or later, proving asynchronous and out-of-order completion.

The Codex fake-server path proves:

```text
parent Invoke A
  → Codex thread and turn
  → dynamic tool call
  → child Invoke B
  → normal Skyra dispatch
  → child Return B
  → parent A resumes
  → response enters the same Codex turn
  → final agent message
  → Return A
  → one root result
```

The recorded DNA-16 checks are:

```text
go build ./src/skyra/skyra-dna-16/...       PASS
go vet ./src/skyra/skyra-dna-16/...         PASS
go test -race ./src/skyra/skyra-dna-16/...  PASS
```

An opt-in real-process test completed the pinned Codex App Server initialize
handshake without performing inference.

## 12. DNA-16 findings and migration constraints

1. `InvocationID` is stable for one execution lifetime.
2. A tool call is a new child invocation with its own identity.
3. `ReturnFd` lets a child return without knowing its parent surface.
4. DNA-16 separates scheduler-owned storage from driver-owned state meaning;
   September 12 instead places provider state locally on its owning Surface.
5. DNA-16 commits driver-state replacement and semantic effects atomically. The
   current proposal commits related graph changes and mailbox/custody handoffs
   together, but does not make Surface-native state or external effects part of
   that transaction.
6. Provider IDs and protocols terminate at the driver boundary.
7. Focus can compile into a provider's dynamic tool registry.
8. Hosted native effects fail closed and cannot be laundered by later output.
9. OS process and transport resources never belong in continuation state.
10. Driver calls and callbacks remain outside the scheduler transaction.

## 13. Pressures exposed by DNA-16

- Scheduler combines storage, authorization, interpretation, lifecycle, and
  selection.
- Continuation interpretation is still stubbed.
- Only one child fan-out/resumption wave exists.
- `Action + Params` remains the application shape.
- Target is still flat entity/address/host fields.
- Horizon is accepted from each Invoke.
- Input validation is a separate first-class component.
- Transport and ProcessHost are peer abstractions beside Driver.
- `DriverInput` carries state alongside Move.
- Host commands are not a durable outbox.
- Production auth, live inference, safe replay, and OS containment are absent.

These pressures produced the current conceptual direction.

---

# Part II — The current direction

## 14. Proposal authority and implementation status

The [September 16 ownership notes](<../specs/2026-09-16/ownership-boundries of ctu/README.md>)
and [minimal loop](../specs/2026-09-16/ctu/minimal-loop.md) capture the current
composition. The same dated space contains Surface, Move, state/profile, custody,
and interpretation declarations. These Go proposals are excluded from normal
builds and deliberately leave supporting types unresolved.

The September 12 monolithic write/trace/select loop, binary Fired map, batch result
queue, and synchronous Execute-returning-Moves model are earlier proposals. They
must not be treated as the latest CTU contract. September 12 Possibility and Horizon
remain inputs to reconcile, especially lifecycle, authority, and handler binding.

DNA-16 remains the executable reference. Its recorded verification in Part I is
historical evidence; this summary update does not claim a fresh runtime test run.

## 15. Move data and interpretation

The current [Move proposal](../specs/2026-09-16/move/proposal.go) declares:

```go
type Move struct {
    Identity      uuid.UUID
    PossibilityID uuid.UUID
    Source        MoveSource
    Target        uuid.UUID
    Invoke        *Invoke
    Return        *Return
}

type MoveSource struct {
    SurfaceID    uuid.UUID
    InvocationID uuid.UUID
}

type Invoke struct {
    InvocationID uuid.UUID
    ContractID   uuid.UUID
    Observations []Observation
}

type Return struct {
    ReturnID    uuid.UUID
    ContractID  uuid.UUID
    Observation Observation
}
```

Move, Invoke, and Return are data. There is no separate opcode wrapper or
interpretation method on these structs. Interpreter validates the Invoke/Return
selection and applies their distinct semantics through internal handlers. The Go
pointer fields alone do not enforce exactly one variant.

Both variants target an address and share dispatch. Return is not merely a graph
update that bypasses Surface delivery. The continuation map records:

> A source invocation tried to execute this Move.

The source may be a parent invoking or a child returning. This record does not
prove execution started, Surface acceptance occurred, or the invocation completed.
Exact continuation relationships, return correlation, and transaction fields are
still to be defined. Fault reports use ordinary Moves, not a third opcode.

## 16. Recursive Surfaces and observation contracts

The latest [Surface proposal](../specs/2026-09-16/surface/proposal.go) uses a struct:

```go
type Surface struct {
    Identity     uuid.UUID
    Contracts    []ContractPair
    Capabilities []*Surface
    Driver       Driver
    State        StateReader
}

type ContractPair struct {
    Identity uuid.UUID
    Input    InputContract
    Output   OutputContract
}

type InputContract struct {
    SurfaceID    uuid.UUID
    Observations []Observation
}

type OutputContract struct {
    SurfaceID   uuid.UUID
    Observation Observation
}
```

Surface and contract identities are UUIDs. Each Surface-bound input contract maps
one-to-one to its expected output contract. Structurally similar inputs on two
Surfaces are distinct contracts because Surface identity is part of the binding.
A Return supplies one observation.

Capabilities are recursive Surfaces with globally unique identities. They can be
registered and invoked independently of their discovery parent. Discovery does
not grant execution authority. Parents know their children's input/output
contracts; a child does not decide whether its output satisfies the parent's
waiting dependency.

Return interpretation checks the actual originating Surface identity, the input
contract recorded for the specific invocation, and the expected return contract
and observation. Matching an unrelated historical invocation is insufficient.
Mismatches enter Trap handling and do not satisfy the dependency.

Surface-native resources and session state remain local. `StateReader` ownership
boundaries are explicitly unresolved in the proposal; registering a Surface in
StateOwner does not move all its native state into runtime tables.

## 17. Possibility and authorized dispatch

The retained [Possibility direction](../specs/2026-09-12/possibility/proposal.md)
identifies an invocation and its first recorded Move and carries a provisioned
Horizon. The latest StateOwner includes a `PossibilityMap`, whose exact shape is
not yet declared.

Horizon owns authorized handler resolution. A Move may reach a Surface only
through its Possibility's provisioned Horizon; discovery or a known Surface UUID
does not itself authorize execution. Invoke and Return share this dispatch path.

Handler types, identity/address mapping, revocation versus concurrent dispatch,
and the boundary preventing direct unauthorized driver calls remain open. The
older `Dispatch(ctx, move) ([]Move, error)` signature needs reconciliation with
single-Move egress, supervision, and asynchronous output.

## 18. StateOwner, system state, and user Surfaces

The [state proposal](../specs/2026-09-16/MessageBroker/proposal.go) now declares
StateOwner as the single owner behind the MessageBroker profiles:

```go
type StateOwner struct {
    requests chan stateRequest
    system   systemState
    user     userState
}

type systemState struct {
    revision uint64
    profiles SystemProfiles

    continuations ContinuationMap // The map of Moves.
    possibilities PossibilityMap
    custody       map[uuid.UUID]Component
    workers       map[uuid.UUID]WorkerRecord

    interpretation mailbox
    supervisor     mailbox
}

type userState struct {
    surfaces     map[uuid.UUID]*Surface
    userSurfaces UserSurfaces
}

type mailbox struct {
    queued []uuid.UUID
}

type SystemProfiles struct {
    Interpretation InterpretationState
    Supervisor     SupervisorState
    Custody        CustodyState
    TrapSurface    TrapSurfaceState
}

type UserSurfaces struct {
    SysCallSurface SysCallSurfaceState
}

type Kernel struct {
    TrapSurface    *Surface
    SysCallSurface *Surface
}
```

There is no duplicate `moves` map: continuations already hold the Moves.
`userState` means user-registered Surfaces, such as Codex, and their associated
access declarations. It is not a second writer. System and user state use the
same StateOwner request loop.

Boot injects restricted typed adapters. Components receive operation-specific
interfaces rather than general Read/Write authority or mutable map references.
Distinct private adapter types prevent recovery of broader broker authority by
type assertion. Profiles are wired at boot, not dynamically assumed. Interface
compatibility is checked at compile time; identity and permitted operations still
require runtime validation. Despite its renamed container, `UserSurfaces` currently
holds the SysCallSurface state interface; concrete Surface references are held by
Kernel and the registry.

Reads return detached snapshots. Snapshot-dependent writes carry a StateOwner
revision; checking the revision and committing happen in the same serialized
operation. The initial direction uses one global revision. Exact request types,
map records, profile methods, and persistence contracts remain provisional.

## 19. Two mailboxes and one Take method

StateOwner owns exactly two Move mailboxes: Interpretation and Supervisor. Its
state-request channel serializes operations; it is not a third Move mailbox.
There is no independent Admission service or authoritative admission channel.

Both components call the same operation:

```go
move, err := state.Take(ctx, self)
```

Boot injects `self` and `next`:

| Component | self | next |
|---|---|---|
| Interpreter | Interpretation | Supervisor |
| Supervisor | Supervisor | Interpretation |

StateOwner chooses the mailbox and checks the caller binding. Mailboxes preserve
committed FIFO insertion order. Concurrent producers have no predefined ordering
before insertion, and execution completion may occur out of order.

An empty Take waits without blocking StateOwner's request loop. One Interpreter
finishes the current Move and its handoff before taking another. Supervisor
reserves worker capacity before Take, keeping capacity policy outside the shared
mailbox method. Taking work does not erase its canonical record; unfinished work
must remain tracked after removal from the queue.

## 20. Custody and the outer loop

Custody answers only:

```text
MoveID -> Interpretation | Supervisor
```

Queued and actively processing work belong to the same component. Custody has no
pending-acceptance status, Surface identity, worker identity, or retry policy.
StateOwner owns storage but never becomes the Move's custodian. Trap is a Surface,
not another custody component. CustodyState is a profile/protocol, not a separate
runtime service.

The outer loop is:

```text
Interpretation mailbox
    -> Take(ctx, self)
    -> interpret Invoke/Return and prepare changes
    -> Transfer(result, self, next)
Supervisor mailbox
    -> Take(ctx, self)
    -> supervised worker executes
    -> Transfer(result, self, next)
Interpretation mailbox
```

Transfer commits related state changes, destination mailbox placement, and custody
together, then acknowledges. There is no separate post-commit component handler
that must accept the Move to complete the handoff. External Surface effects are
outside the state transaction.

Interpretation generally transfers an existing Move. Supervisor generally submits
a newly emitted Move. The state operation preserves that distinction and the
original execution record; it does not require different outer loops.

The existing custody Go proposal still takes `moveID, fromComponent, toComponent`.
The latest loop requires a result carrying the Move and associated changes. That
signature is explicitly awaiting reconciliation, not already implemented.

## 21. Interpretation, dependencies, and state commits

The separate Dependency Gate/Scheduler exploration has been replaced by sequential
interpretation. Invoke and Return handlers know their own prerequisites and graph
semantics. There are still two dependency paths: an authorized target Surface must
resolve, and its required input observations must be available.

Interpretation reads detached state and proposes changes without committing or
running external effects. Transfer commits those changes and the handoff together.
A stale revision commits nothing: Interpreter reads current state and re-evaluates
the same Move before taking another. Known in-flight work and an unresolved
requirement remain semantically different; Trap handles corrective work where
needed. Waiting records and wakeup details remain open.

The earlier binary Fired map is not the latest state model. Worker registration,
progress, acceptance, and response bookkeeping now belong to StateOwner through
Supervisor's profile. Their exact record shapes and parent-resumption rules still
need definition; the redesign does not erase the need for execution correlation.

The older Interpreter interface returns only `error`; the current loop uses a
result proposal plus failure information. These signatures need reconciliation.

## 22. Supervisor, execution, and response acknowledgment

Supervisor owns dispatch and the bounded worker pattern. It registers workers
through StateOwner before launching, monitors progress independently, and applies
configured policy. Canonical worker records live in StateOwner. Live goroutines,
local variables, calls, and native resources are not eliminated by describing the
processing services as stateless.

The current Surface Driver remains callback-based:

```go
type Driver interface {
    Ingress(context.Context, Move, func(DriverEmission) error) error
    Egress(DriverEmission) Move
}
```

Ingress acceptance means the Surface received the work. It is separate from
emitted outcome Moves and invocation completion. A child Invoke may be an outcome
without completing its source invocation. The conceptual supervised Execute
wrapper and its relationship to callback ingress still need an exact contract.

Execution feedback/ack tracking is separate from component custody. It correlates
Surface acceptance and subsequent responses, records a qualifying response before
settlement, and prevents late/duplicate feedback from reopening settled work.
The worker registry and earlier writeback-ack table need reconciliation.

If a completed execution's handoff sees a stale revision, Supervisor rebuilds only
the state proposal from the retained outcome. It preserves emitted identity and
correlation and never reruns Execute solely to repair a stale write.

## 23. Trap, faults, cancellation, and stalled workers

Kernel holds TrapSurface and SysCallSurface. TrapSurface uses the system state
profile; SysCallSurface's state interface is in UserSurfaces. Exact syscall
operations and boot/lifecycle methods remain undeclared.

Trap is a registered Surface with ingress bookkeeping and registered capability
children. A trap Move carries the troublesome Move and triggering event. Kernel
interrupt/fault handlers describe corrective Moves; ordinary interpretation and
execution perform the correction. Trap tracks outstanding corrective work and
whether its result satisfies the requirement.

The tightened minimal loop distinguishes:

- **Move-level failures:** normalize to ordinary failure-report Moves targeting
  the kernel Trap Surface. Fault is payload, not another opcode. FaultResult
  records the original Move and correlation; it does not choose corrective work.
- **Stale state proposals:** re-evaluate the proposal after a known non-commit.
  Interpreter may re-interpret; Supervisor must retain the execution outcome.
- **Component shutdown and infrastructure failure:** propagate out of the loop
  or worker wrapper without erasing unfinished records. Cancellation of the
  supplied component context is separate from a Surface-local timeout.

Panic recovery covers the current goroutine's Move handling, not arbitrary child
goroutines or a whole-process failure. Fault construction itself can fail; there
is no promise that every failure advances the Move.

Independent Supervisor monitoring handles workers that never return. A configured
deadline or lost progress is recorded and reported through the normal Move path.
A timeout does not prove the worker stopped or external effects did not occur.
Late outcomes must reconcile against the recorded report, capacity must reflect
actual worker lifecycle, and any retry requires the configured safe policy.

Cancellation while awaiting a transfer acknowledgment does not prove non-commit.
Ambiguous outcomes need operation-identity reconciliation before resubmission.
Exact identities, monitor mechanics, shutdown, and recovery remain open. Empty
mailboxes or ack tables alone do not prove Possibility completion; in-memory
custody does not provide durability across process loss.

## 24. Ownership and DNA-16 migration

| Owner | Current direction |
|---|---|
| StateOwner | Serialized canonical state access; continuation/possibility maps, both mailboxes, custody, worker records, and user Surface registration. |
| Interpreter | Sequential opcode semantics, dependency checks, and state/handoff proposals. |
| Supervisor | Authorized dispatch, bounded workers, progress monitoring, and configured retry behavior. |
| CustodyState | Restricted atomic mailbox/custody operations; no independent service or map. |
| Horizon | Authorized live handler resolution; detailed enforcement still open. |
| Kernel / Trap | Trap and SysCall Surface references; corrective-work decisions through Trap handlers. |
| Surface | Contracts, ingress/egress, recursive capabilities, and local/native state. |

| DNA-16 | Current treatment |
|---|---|
| Combined scheduler state/selection/interpretation | StateOwner storage, sequential Interpreter, supervised concurrent execution. |
| MoveBatch and private feedback families | One Move per mailbox entry/egress result; atomic related state changes remain necessary. |
| Resident continuations | ContinuationMap is the Move map, with separate custody and worker records. |
| Placement feedback | Separate component custody from Surface acceptance and response tracking. |
| Scheduler-stored opaque DriverState | Provider-native state remains Surface-local; runtime worker state remains StateOwner-owned. |
| Driver-state/effect transaction | State changes and mailbox handoff commit together; native effects remain outside the transaction. |
| String address registration and peer validators | UUID Surface identity, explicit contract pairs, recursive independent capabilities, authorized handler lookup. |
| Action + Params and ReturnFd | Explicit Invoke/Return data, observation contracts, and source fields; exact correlation laws still open. |
| Codex adapter and dynamic tool proof | Retained migration pressure test; userState can register Codex as a Surface. |

## 25. Immediate contract work

1. Define restricted state profile methods, Take wakeups, capacity, request replies,
   and Transfer result/transaction types. Reconcile the older custody and Interpreter
   signatures with the minimal loop.
2. Define continuation and possibility records, opcode updates, return correlation,
   contract snapshots, parent resumption, and response settlement.
3. Define worker registration, outcome retention, independent monitoring, retry
   policy, late-output reconciliation, and ambiguous handoff acknowledgment handling.
4. Reconcile supervised Execute with callback ingress and one-Move egress; preserve
   acceptance versus output versus completion.
5. Specify Trap/SysCall contracts, user Surface registration, boot wiring, Horizon
   enforcement and revocation, and lifecycle/shutdown behavior.
6. Reconcile Surface-native state and effects with state commits, then authorize
   implementation and repeat the Codex vertical proof.

## 26. Current conceptual map

```text
                       StateOwner request loop
                   systemState       userState
                 continuations       registered Surfaces
                 possibilities       UserSurfaces
                 custody/workers
                 SystemProfiles
                        │
              owns both Move mailboxes
                        │
          ┌─────────────┴────────────────┐
          ▼                              ▼
 Interpretation mailbox           Supervisor mailbox
          │                              │
     Take(self)                     Take(self)
          │                      after reserving capacity
          ▼                              ▼
     Interpreter                     Supervisor
 one Move at a time             dispatch + bounded workers
 Invoke/Return semantics        authorized Horizon -> Surface
          │                              │
          └── Transfer -> Supervisor     └── Transfer -> Interpretation

Transfers commit via StateOwner: related state + mailbox + custody, then ack.
Trap receives failure/dependency reports as ordinary Moves and describes corrections.
Surface-native state remains local; canonical worker state belongs to StateOwner.
```

## Primary references

### Current September 16 proposals

- [Ownership boundaries](<../specs/2026-09-16/ownership-boundries of ctu/README.md>)
- [Minimal CTU loop](../specs/2026-09-16/ctu/minimal-loop.md)
- [StateOwner, profiles, user Surfaces, and Kernel structs](../specs/2026-09-16/MessageBroker/proposal.go)
- [State access and mailbox notes](../specs/2026-09-16/MessageBroker/README.md)
- [Surface and observation contracts](../specs/2026-09-16/surface/proposal.go)
- [Move, Invoke, and Return](../specs/2026-09-16/move/proposal.go)
- [Custody profile — signature awaiting reconciliation](../specs/2026-09-16/tu/proposal.go)
- [Interpreter interface — signature awaiting reconciliation](../specs/2026-09-16/interpretation/proposal.go)

### Retained September 12 background and superseded loop

- [September 12 contract set](../specs/2026-09-12/README.md)
- [Earlier Surface proposal](../specs/2026-09-12/surface/proposal.md)
- [Possibility and Horizon — retained direction, provisional types](../specs/2026-09-12/possibility/proposal.md)
- [Earlier CTU loop — superseded by September 16 exploration](../specs/2026-09-12/ctu/minimal-loop.md)

### Executable reference

- [DNA-16 work order](../src/skyra/skyra-dna-16/workorder.md)
- [DNA-16 decisions and recorded verification](../src/skyra/skyra-dna-16/DECISIONS.md)
- [DNA-16 deferred issues](../src/skyra/skyra-dna-16/issues.md)

### Historical inputs and provider pressure test

These inform reconciliation; they do not override the current proposals.

- [Dispatch and operand ownership](../specs/2026-09-11/dispatch-operand-ownership.md)
- [Ordered operands](../specs/2026-09-11/continuation-language-operands.md)
- [Earlier Horizon design](../specs/2026-09-11/horizon.md)
- [Earlier runtime decomposition](../specs/2026-09-11/system-boundary-cleanup-priorities.md)
- [Earlier universal Surface](../specs/2026-09-11/host-driver/universal-surface.md)
- [Observe, express, realize](../specs/2026-09-11/host-driver/observe-express-realize.md)
- [Kernel assumption](../specs/2026-09-11/host-driver/ctu-kernel-proposal/kernel-assumption.md)
- [Earlier Possibility boundary](../specs/2026-09-11/host-driver/ctu-kernel-proposal/possibility-boundary.md)
- [Codex driver pressure test](../specs/2026-09-11/codex-driver/README.md)
