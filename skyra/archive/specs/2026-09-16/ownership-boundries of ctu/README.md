# CTU ownership: sequential interpretation and supervised execution

Updated: 2026-09-16.

Status: selected direction for the initial working demo; detailed loop contracts
remain open. This replaces the earlier Dependency Gate / Scheduler decomposition
within this exploration. It does not change DNA-16 or silently replace the
[existing CTU loop proposal](../../2026-09-12/ctu/minimal-loop.md). The other September 12
contracts still require reconciliation before implementation.

## Central pattern

```text
new Move / Surface outcome / corrective Move
    -> State Owner: interpretation mailbox
    -> Interpreter: one Move at a time
        -> Invoke / Return semantics and dependency checks
        -> State Owner: commit changes + transfer to supervisor mailbox
    -> Supervisor: dispatch + concurrent workers
        -> Surface ingress / native work / egress
        -> emitted Move returns to interpretation mailbox

all canonical state operations -> State Owner request loop
unresolved dependencies / execution trouble -> Trap Surface via ordinary Move path
```

The CTU has three main services: Interpretation, Supervisor, and State Owner.
There is no separate Dependency Gate or Admission service. Admission was only a
mailbox for Interpretation; that mailbox now belongs to State Owner.

## Responsibilities

| Service or role | Responsibility |
|---|---|
| State Owner | Owns both Move mailboxes and all canonical runtime records: Move/continuation state, custody, and worker execution/progress/retry bookkeeping. Serializes restricted state requests and acknowledges after commit. |
| Interpreter | Consumes one Move at a time; applies Invoke/Return semantics, checks dependencies, and requests state changes and execution or Trap work. |
| Supervisor | Resolves dispatch bindings and runs the supervisor/worker pattern with concurrent bounded computations. Reports progress and outcomes through its state profile. |
| Trap Controller | A registered Surface whose ingress handles troublesome Moves and whose capabilities describe corrective work. It is not an additional custody owner. |
| Surface | Owns its contracts, ingress/egress, capabilities, and Surface-native state. Emits one normalized Move per egress result. |

Interpreter and Supervisor are stateless with respect to authoritative runtime
records. Their locals, goroutines, active calls, and live worker resources still
exist. Canonical worker state lives in State Owner; Surface-native state remains
with the Surface. Dispatch and supervision are roles within Supervisor, not new
state-owning services.

## Sequential interpretation owns dependency semantics

The public conceptual entry is `Interpret(move)`. Invoke and Return handlers
live inside that interpretation phase. They determine both the meaning of the
transition and its prerequisites.

```text
Interpret(move)
    -> select Invoke or Return handler
    -> inspect current state and declared dependencies
    -> determine and commit opcode-specific changes
    -> release eligible work for dispatch, or request Trap handling
```

For example, Invoke establishes invocation/dependency relationships; Return
records a result observation against the appropriate relationship. Parent and
child updates that belong together must commit atomically. Their exact field
contracts and ordering relative to prerequisite checks remain to be specified.

Invoke and Return both target an address and use the same dispatch mechanism.
A Return is not replaced by a graph-only update or a special dispatcher parent
route. The addressed Surface interprets the delivered payload.

Two dependency paths remain:

- **Target resolution:** the addressed Surface must be available through an
  authorized handler.
- **Input availability:** the required observations must be present for that
  Surface to interpret.

A known dependency already in flight does not need another corrective invocation.
An unresolved requirement may need Trap to obtain corrective work. Interpretation
uses declared requirements; it does not invent procedure intent or predict all
inputs a Surface may discover during execution.

Horizon remains the authorized resolution boundary. A readiness check does not
itself grant authority or guarantee that a later binding remains valid after
revocation. Dispatch must preserve the Horizon contract.

An unresolved Move cannot be discarded. Its authoritative record, waiting
ownership, and notification when observations change must be defined in the loop
contract. This rewrite does not introduce a second waiting map in a Gate.

## Recursive Surface identity and observation contracts

A Surface carries explicit contract declarations alongside its ingress and
egress. Each input contract is bound to that Surface's identity, with a one-to-one
mapping to the expected return contract:

```text
(Surface identity, input contract) -> expected return contract
```

Input contracts describe the observations accepted for the interaction; return
contracts describe the expected output observations. Contracts are unique with
their Surface identity, not reusable payload schemas detached from an endpoint.
Two Surfaces may accept structurally similar observations, but those are distinct
contracts because the Surface identity is part of the contract identity. Exact
encoding, versioning, and the contracts-list type remain open.

Capabilities remain Surfaces, rather than separate capability descriptors:

```text
Surface
    Identity
    Contracts: input contract -> expected return contract
    Local state
    Ingress / Egress
    Capabilities: []Surface
        -> child Surface with the same recursive shape
```

The contract declarations live on the child Surface. Its parent reads them to
establish what inputs the child accepts and what observations it expects back.
The child's ingress interprets input and its egress expresses output; the child
does not need to inspect its own return contract to decide whether its parent's
dependency is satisfied. Return interpretation enforces the recorded caller-side
expectation. Generic ingress/egress Go signatures alone do not expose these
observation contracts, so an explicit contracts list is needed.

Every capability Surface has its own globally unique, stable identity. It can be
registered and addressed independently: a Move can target it directly without
invoking or traversing its discovery parent first. The contract pairing uses the
capability Surface's own identity, not its parent's identity or discovery path.
Listing the same capability under multiple parents does not create new identities.

```text
Parent.Capabilities() -> discovery of a child Surface
Move.Target = child's identity -> Horizon -> child's authorized handler
```

Independent addressability does not grant authority. Direct capability invocation
still passes through the invocation's provisioned Horizon. The exact relationship
between identity and Address, global registration scope, and collision validation
must be reconciled with the Surface and Boot contracts.

### Return validation

When interpreting a Return, compare its observation and correlation against the
specific recorded invocation:

1. Did it originate from the Surface identity that was invoked?
2. Does it correspond to the input contract used for that invocation?
3. Does its return contract match the expected contract, and does the observation
   satisfy that contract?

A mismatch enters the interrupt path through Trap and must not satisfy the
waiting dependency. Matching some historical invocation with the same contracts
is insufficient: repeated calls require execution-specific correlation. Source
identity is bound by the actual execution path, not trusted from a payload claim.

The expected input/return pairing is recorded when the request is admitted.
Later changes to a Surface's advertised contracts do not silently change the
expectations of an existing invocation. Exact Return fields and how that recorded
pairing is referenced still need the Move contract.

### Trap is a Surface using the same contracts

Trap Controller is registered as a Surface. A trap Move carries the troublesome
original Move and its triggering event as payload observations. Trap ingress
interprets that payload and performs Trap-local bookkeeping. Its registered
capability children are themselves Surfaces with their own identities and
input/return contract pairings.

The caller can therefore declare the expected Trap response using the same
contract mechanism. Trap uses its kernel-provided interrupt/fault handlers to
describe corrective Moves, which re-enter the ordinary queue. The exact trap
payload and child capability set remain to be defined; this is not permission
for Trap to author arbitrary corrective work or bypass Horizon.

## Mailboxes, writes, and dispatch ordering

State Owner owns exactly two Move mailboxes: Interpretation and Supervisor.
Each entry carries one Move. The separate state-request channel is infrastructure
for serializing state access, not a third Move mailbox.

Mailbox insertion and the corresponding custody change commit in one State
Owner operation. A transfer does not first change ownership and then call a
receiver that might reject the Move: committing the destination mailbox entry
is the handoff. If the operation fails before commit, the current owner retains
responsibility. Initial submission records Interpretation ownership and its
mailbox entry together.

Mailboxes preserve FIFO insertion order. Concurrent producers are ordered by
the State Owner's committed insertions, not by when they began their requests.
Exactly one Interpreter consumes its mailbox and finishes interpreting the current
Move before taking the next. It waits when empty and resumes when work arrives;
there is no periodic drain or parallel interpretation. Supervisor consumes its
mailbox and runs supervised workers concurrently, so completion order may differ.

Taking a Move for processing does not erase its canonical record or change its
component owner. Queued and actively processing Moves both belong to that service.
Canonical work tracking must retain unfinished work after mailbox removal.

The earlier buffered `chan Move` admission proposal is superseded. Channels may
carry state requests and wakeups, but an independent channel send plus a custody
map write would again be two separate operations. The mailbox contents themselves
are State Owner state. Waiting for an empty mailbox must not block the State Owner
request loop; exact notification, capacity, cancellation, and shutdown mechanics
remain to be defined.

Interpretation requests related graph changes and the supervisor mailbox handoff
in one commit before execution starts. Surface effects occur outside this
transaction. Atomic state changes do not make external computation atomic.
Incoming responses must be recorded even if the work they introduce cannot run
yet. Submission transfers exclusive use of payload data or copies it so producers
cannot mutate canonical inputs afterward.

## What the Move write proves

The address continuation Move map records:

> A source invocation tried to execute this Move.

The source can be a parent emitting an Invoke or a child emitting a Return. The
record establishes the requested transition and its relationships, not Surface
acceptance, actual execution, or successful completion. Initial external ingress
and its source attribution still need a boot/admission contract.

Opcode interpretation determines how that request updates the graph. Subsequent
acceptance records and outcome Moves establish what happened afterward. A failed
handoff does not erase the recorded request or silently authorize a retry.

## Dispatch binds; supervised Execute accepts

Dispatch resolves Move.Target through the provisioned Horizon and binds the
Move to the selected Surface's ingress and egress, with its declared execution
and retry policy. Invoke and Return share this lookup and handoff path.

`Execute(move, ingress, egress, policy)` is a conceptual supervised-execution
boundary within Supervisor, not a finalized Go signature. Acceptance returns
promptly once supervised work is registered. Component custody already belongs
to Supervisor through its mailbox handoff. Actual execution and egress output continue separately.
The method does not wait to return the eventual outcome Move.

There may be many dispatch invocations. For the same execution request, finding
an existing execution record or registering a new one must be one atomic operation through
the state owner. Two dispatchers must not both observe absence and launch duplicate
workers. Different requests to the same Surface remain independent.

Dispatch retains only the local information needed to complete its handoff. Once
the supervisor accepts, Dispatch may exit. If no owner accepts, failure propagates
without abandoning the work. Canonical registration and acceptance feedback go
through the sole Writer; Dispatcher and workers do not mutate its tables.

## Acceptance is not output

The local handoff rule remains:

```text
State Owner commits destination mailbox + custody -> returns success
    -> sender releases responsibility
```

A function return is sufficient acknowledgment between adjacent components:
`nil` means accepted; error means no transfer. A receiver must not return an
acceptance error while leaving the same work accepted or scheduled. Removing an
accepted Move from a sender's outgoing work is local cleanup, not a second
canonical request write.

There are distinct acceptance boundaries:

- **Mailbox acceptance:** State Owner has committed the Move and its component ownership.
- **Supervisor acceptance:** worker registration is recorded and supervised work
  can proceed; Dispatch may return. This does not change component custody.
- **Surface acceptance:** ingress returns “I got it.” This does not imply that
  the Surface has emitted a response or completed the invocation.

Their exact call signatures and writer interactions remain open. In particular,
queue acceptance must not be described as a completed graph write unless that
write has actually occurred.

Surface outcome Moves arrive on a separate output path and re-enter the queue.
A child Invoke may hand further work back without completing its source
invocation. Execution output, Surface acceptance, and invocation completion
must not be conflated.

## One writer, request records, and custody records

The Writer role now belongs to a centralized State Owner. This extends the
earlier custody-only service: components submit both read and write requests to
one state-access boundary. Its request channel has one consumer that processes
requests sequentially.

```text
Interpreter / Trap / supervision / custody transfer
    -> state-request channel
    -> State Owner
        -> read: return detached snapshot and revision
        -> write: validate, commit, advance revision, acknowledge
```

The State Owner owns authoritative runtime maps. Components do not read those
maps directly or receive mutable aliases into them. A single writer alone would
not make concurrent unsynchronized map reads safe. Reads through the State Owner
observe committed state; snapshots must detach nested mutable data as needed.
Canonical worker state also belongs to State Owner. Surface-native resources
and temporary worker locals remain with their runtime owners.

Write requests express declared operations, such as recording a Move or
transferring custody, rather than arbitrary callbacks with unrestricted mutation
access. Interpretation and Trap retain their semantics; the State Owner validates
and applies their proposed changes. Related changes can commit in one request.
Successful write acknowledgment follows commit. Surface execution stays outside
this loop, and the State Owner must not wait on work requiring another state
request to complete.

The state-request channel serializes state access; the two State Owner mailboxes
retain Moves for Interpretation and Supervisor. This is still one writer, not one writer per table. Exact request
types, replies, capacity, and shutdown coordination remain to be defined.

### Revision-based reconciliation

Sequential interpretation and sequential Trap processing do not alone prevent
their separate requests or supervisor feedback from interleaving. Even with all
reads and writes serialized, a detached snapshot may become stale before its
caller proposes a write.

Use a monotonically increasing revision assigned by the State Owner, not a
wall-clock timestamp, to reconcile state-dependent proposals. This is optimistic
concurrency control. For the demo, start with one global revision:

```text
read -> snapshot at revision 42
component interprets snapshot and proposes changes based on revision 42
State Owner processes proposal:
    current revision == 42
        -> validate and atomically commit changes
        -> advance revision and return success
    current revision != 42
        -> reject as stale without applying changes
        -> caller reads current state and re-evaluates
```

Revision checking and commit occur in the same State Owner turn. A successful
state-changing transaction advances the revision; reads do not. State-independent
operations need not carry a snapshot revision merely for form's sake, but still
pass their operation-specific validation. Unrelated writes may invalidate a
proposal under the global scheme; per-record revisions are a possible later
refinement, not required for the initial demo.

Reconciliation must precede the release of external execution. Re-evaluating a
stale proposal does not authorize replaying a Surface effect or creating duplicate
corrective work. A committed decision does not make later execution outcomes
atomic with its write or replace execution-time authority requirements.

This closes the stale-snapshot gap; it does not replace custody acceptance,
duplicate-handoff handling, or response correlation. Retry bounds and response
types for stale proposals remain loop implementation details.

### Component-specific state access

The latest refinement replaces general Read/Write access with boot-injected
InterpretationState, SupervisorState, and CustodyState profiles. MessageBroker
owns the state and serializes all profile requests through the same State Owner.
Earlier read/write descriptions above describe internal operations, not public
component authority. Exact profile methods remain to be declared.

Each component's constructor requires its typed profile. Distinct private adapter
implementations restrict exposed methods; components do not dynamically assume
profiles or receive the full broker. Compile-time checks cover interface method
compatibility, while boot and broker validation enforce runtime identity and
permitted transitions. See the
[MessageBroker proposal](../MessageBroker/README.md).

### Component custody and boot wiring

Custody answers exactly one question: which component currently owns this Move?

```text
MoveID -> Interpretation | Supervisor
```

There is no pending-acceptance status. A Move queued for Interpretation belongs
to Interpretation; one queued for Supervisor belongs to Supervisor. State Owner
owns storage, not custody of every Move. Trap is a Surface, not a custody component.
Worker identity, Surface identity, response acknowledgments, and retry policy
remain separate execution facts.

CustodyState is the restricted state profile for the transfer protocol. It does
not require a separate custody service, queue, or writable map. The custody
proposal directory remains the home for that protocol and its record types.
Boot injects component identity, permitted destinations, and restricted adapters.

The transfer request retains its explicit source and destination:

```go
Transfer(ctx context.Context, moveID uuid.UUID,
    fromComponent Component, toComponent Component) error
```

State Owner validates the expected owner and commits the destination mailbox
entry and custody change together. Successful return acknowledges that commit.
Interpretation's graph changes and associated handoff must be expressible as one
operation, not a graph commit followed by an unrelated Transfer call. The exact
profile method for that combined operation remains to be proposed.

See the [custody profile](../tu/proposal.go) and
[MessageBroker proposal](../MessageBroker/README.md).

### One outer handoff loop

Interpretation and Supervisor both request transfers through their injected state
profiles. The outer mechanics are the same in both directions:

```text
Interpretation mailbox -> interpret -> transfer to Supervisor mailbox
Supervisor mailbox -> supervised worker execution -> transfer emitted Move to Interpretation mailbox
```

Each handoff submits a Move to the destination mailbox, records its custody, and
acknowledges the commit. State Owner performs mailbox placement and ownership
recording atomically. Components continue according to that acknowledgment.

Interpretation normally transfers the existing Move; Supervisor normally submits
a newly emitted Move. The state operation preserves that identity distinction:
an emitted Move receives its own record and Interpretation custody, while the
original Move retains its separate execution record. This does not require a
different outer loop. Exact method parameters for accepting new Moves remain to
be defined; the existing MoveID-only transfer signature does not yet express that.

### Request history and execution feedback remain separate

| Record | Meaning |
|---|---|
| Address continuation Move map | The source requested a transition; records relationships and observations. |
| Component custody record | Interpretation or Supervisor currently owns this Move. |
| Execution feedback / acknowledgment state | State Owner records Surface acceptance, worker progress, retries, and correlated responses for Supervisor. |

The earlier writeback ack-table discussion concerns execution feedback, not
component custody. It must not introduce pending acceptance, Surface identities,
or invocation completion into CustodyRecord. Supervisor's registry and ack
bookkeeping still need reconciliation with each other, independently of the
two-owner custody map.

The execution-side writeback requirements remain: correlate to a specific
handoff, record a qualifying response before settling its outstanding feedback,
and prevent late or duplicate acknowledgments from reopening a settled handoff.
A child Invoke can be a response without completing the source invocation.
These are execution semantics, not extra custody-transfer states.

## Supervision, retries, and Trap

The execution wrapper supervises workers rather than requiring a crashed or
hung worker to detect its own failure. Surface-specific policy can describe
progress expectations, deadlines, and permitted retries. The detailed monitor
shape and whether it is shared or per execution remain open.

Acceptance timeout and response timeout are different conditions. Neither proves
that native effects did not occur. Repetition requires a safe operation or a
stable identity with suitable duplicate suppression/reconciliation. Supervisors
must not retry arbitrary effects solely because a timer expired.

When a handoff fails, or supervision reaches a condition it cannot safely resolve,
report the original Move, correlation, and failure context to Trap. An internal
failure report is not a corrective Move authored by Dispatcher or supervisor.
The current owner retains responsibility until Trap accepts the report; failure
to transfer cannot silently discard the request.

Trap uses one kernel-provided interrupt/fault interface for both unresolved
dependencies and execution trouble. Its handlers describe corrective Moves;
they do not perform the correction directly. Trap emits only those handler-
provided Moves back into the queue. Normal interpretation, authorization, and
execution perform the corrective work.

Trap tracks whether corrective work is already outstanding and whether its
result satisfies the requested requirement. Its bookkeeping cannot independently
mutate authoritative graph/custody state. The relationship between any Trap-local
indexes and Writer-owned facts still needs definition.

A correlated result is not automatically a satisfied dependency. Opcode handlers
must use the declared satisfaction rule, with Trap supplying correlated corrective
outcomes rather than a competing definition of readiness.

## Invariants retained for the demo

1. One State Owner serializes all authoritative reads and writes; components
   use detached snapshots and explicit requests rather than direct map access.
2. Invoke and Return semantics live in their interpretation handlers.
3. Every dispatched Move has a committed request record.
4. Dispatch is shared across opcodes and respects Horizon authority.
5. External calls and native execution occur outside write transactions.
6. Component handoff commits destination mailbox placement and custody together.
7. Output is bound to its actual source and execution; Surfaces cannot forge
   another invocation's provenance or mutate resident graph values.
8. Acceptance, response writeback, and invocation completion are different facts.
9. Corrective work returns through the same queue and normal Move path.
10. No standalone Gate or persistent Dispatcher table duplicates canonical state.
11. Snapshot-dependent writes reconcile their revision atomically at commit;
    stale decisions are re-evaluated before external execution is released.

## Reconciliation with the September 12 proposals

- [CTU minimal loop](../../2026-09-12/ctu/minimal-loop.md): replace the monolithic trace/select
  loop, Fired map, synchronous Execute returning Moves, and direct fault
  construction with this composition. Preserve atomic graph writes, detached
  inputs, source binding, and the continuing Possibility.
- [Surface](../../2026-09-12/surface/proposal.md): retain Surface-owned ingress/egress and local
  state. Add Surface-bound input/return contract declarations and globally unique,
  independently addressable capability identities while retaining recursive
  child Surfaces. Define ingress acceptance separately from outcome emission,
  reconcile the supervised wrapper, and retain Horizon enforcement.
- [Possibility](../../2026-09-12/possibility/proposal.md): preserve provisioned authority and
  lifecycle ownership. Reconcile Dispatch returning outcomes with acceptance
  plus asynchronous output and supervision.
- Move, Kernel, and Boot proposals remain missing. They must define exact opcode
  changes, correlation, interrupt handlers, construction, and root lifecycle.

This note replaces its own earlier Gate/Scheduler ownership sketch. It does not
silently promote all supporting types to accepted contracts or implement the loop.

## Next: define the loop details

Resolve these in order rather than adding further runtime components:

1. **Mailbox operations:** FIFO take/wakeup, capacity, unfinished-work tracking,
   and atomic custody plus mailbox commits through restricted profiles.
2. **Interpretation transaction:** exact Invoke/Return mutations, dependency
   assessment, source binding, recorded contract pairings and Return validation,
   and recording responses that introduce blocked work.
3. **Unresolved work:** where it remains recorded, Trap acceptance, and how new
   observations cause reconsideration without lost wakeups or duplicate execution.
4. **Execution registration:** identities, atomic find-or-create, capacity
   admission, and the relation between registry and writeback ack records.
5. **Acceptance and output ordering:** supervisor ack, surfaceAck, early output,
   duplicate feedback, settlement evidence, and write failures after acceptance.
6. **Supervision and termination:** policy hooks, safe retries, cancellation,
   final outcome, late output, resource cleanup, and failures in the Trap path.

No empty queue or empty ack table alone proves the Possibility has completed.
In-process custody records also do not provide durability across loss of the
whole process; crash recovery remains a separate contract.
