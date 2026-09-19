# Work order — DNA 16 driver host and transport

**Status: W0–W7 complete within the local implementation scope.** The runtime,
transport, process host, Codex protocol adapter, containment path, conservative
recovery rule, and mocked integrated proof are implemented. Production credential
provisioning, live inference, and OS sandbox installation remain explicitly
deferred.

## Objective

Build the smallest DNA16 boundary that can host a long-lived external driver such
as Codex App Server while preserving Skyra ownership:

```text
kernel       owns authority and possibility-scoped host construction
scheduler    owns continuation lifetime and opaque DriverState storage
dispatcher   resolves the stable entity/address binding
driver       owns capability and private protocol meaning
transport    owns framed message movement only
```

Prove that a Codex dynamic tool call becomes a child Skyra invocation and returns
to its stable parent invocation without placing Codex threads, turns, JSON-RPC, or
native-tool semantics in the scheduler.

## Authority and scope

Read these local documents as the working set:

1. [`contracts.go`](contracts.go): the DNA15 declaration baseline plus the
   implemented DNA16 driver-transition, `Transport`, and `DriverState` additions.
2. [Driver host, transport, and opaque state proposal](proposals/driver-host-and-transport/proposal.md): ownership, invariants, and the intentionally open transition boundary.
3. [Codex driver design](../../../specs/2026-09-11/codex-driver/README.md): the provider-specific pressure test and prototype behavior.
4. [Driver composition and opaque state](../../../specs/2026-09-10/driver-composition-and-state.md): the prior ingress, egress, transport, state, and restore decisions.
5. [Dispatch and operand ownership](../../../specs/2026-09-11/dispatch-operand-ownership.md): the dispatcher/scheduler/driver interpretation split.

Keep implementation changes inside `skyra-dna-16` until an explicit integration
item says otherwise. Treat DNA15 and the dated specs as read-only references while
performing this work. Preserve unrelated user changes.

This work order covers the driver-host boundary. It does not automatically
authorize the separate continuation-language migration from `Action + Params` to
ordered `Operands`, the final entity/address/capability encoding, or production
authentication policy.

## Declare the surface before implementing it

Keep structs, interfaces, supporting types, and ownership comments in
`contracts.go`. Put method bodies and concrete implementations in separate files.
Do not add a declaration merely to make an implementation compile; record and
review the missing surface as its own W0 decision first.

The following declaration items were settled before their dependent bodies:

| Surface item | Required decision |
| --- | --- |
| Transport | Retain an optional opaque-frame `Send`/`Receive`/`Close` seam; do not embed it in every `Driver`. |
| Driver state | Retain a versioned opaque payload owned as storage by Scheduler and interpreted only by the selected driver. |
| Driver input | Define the detached invocation view and previous `DriverState` delivered for one driver transition. |
| Driver event | Define correlation without exposing provider sessions, threads, turns, or JSON-RPC identifiers to scheduler semantics. |
| Driver transition | Define complete replacement state and requested runtime effects without giving the driver a Scheduler reference. |
| Host command | Define how committed transitions request outbound frames, interruption, or host lifecycle work. |
| Atomicity | Choose outbox, acknowledgement, or reconciliation semantics for a crash between state commit and external send. |
| Signals | Define completion, failure, availability, and contamination without converting provider errors into continuation-language opcodes. |

`Driver.Ingress` now receives `DriverInput` and emits `DriverEmission`; `Egress`
retains driver-owned normalization. The controller converts emissions into bound
`DriverTransition` events. The scheduler applies state, Moves, and failure as one
transition, and commands execute only after a successful commit.

## Ownership rules

- Scheduler is the sole owner of resident `Continuation.DriverState`. Clone its
  payload on admission, replacement, dispatch views, and snapshots.
- Driver state is not the semantic `Register`; it cannot satisfy continuation
  contracts or grant authority.
- Drivers receive detached values and return complete replacements. They never
  retain pointers into scheduler state.
- Transport owns framing and I/O only. It does not parse provider messages, emit
  Skyra `Move`s, select capabilities, retry semantic work, or decide authority.
- Driver host owns live process/connection resources. No process, pipe, socket,
  goroutine, context, mutex, callback, or writer enters `DriverState`.
- Kernel constructs each host under possibility-derived authority. Possibilities
  with different authority do not share a host process.
- Dispatcher resolves the stable entity/address binding. Transport and provider
  identifiers do not enter the public address.
- Driver interprets capability and operand meaning and converts private protocol
  events into declared transition effects.
- Never call a driver, transport, process, validator, or host callback while the
  scheduler mutex is held.

## Transport rules

- One `[]byte` value is one complete frame; empty frames are invalid.
- Concurrent sends must not interleave bytes.
- `Close` is idempotent and unblocks an outstanding `Receive`.
- `Send` and `Receive` reject nil contexts and honor cancellation where the
  underlying mechanism permits it; closure remains the guaranteed escape from a
  blocked read.
- Returned frame bytes do not alias an internal reusable buffer.
- Framing errors and EOF are explicit delivery failures. Transport never invents
  provider responses or retries frames on its own.
- JSONL is the first concrete transport, not the universal wire format.

## Driver-state and transition rules

- Version zero plus an empty payload means no prior driver state.
- Nonempty state must have a nonzero version.
- The owning driver validates and migrates versions. Unknown, malformed, or
  incompatible state faults the affected invocation explicitly.
- A transition returns a full replacement, not a mutation patch.
- Replacement state and accepted continuation effects commit together or neither
  commits.
- Duplicate and late driver events cannot reopen terminal work.
- Provider event correlation is driver-owned; scheduler-visible identity remains
  `InvocationID` and handoff sequence.
- An outbound host command is not considered delivered merely because its state
  transition committed. The selected atomicity rule must represent that gap.

## Codex-specific boundary rules

- One Codex App Server process belongs to one possibility.
- One active Codex-backed `InvocationID` owns one thread and at most one turn in
  the first implementation.
- Focus is compiled into the thread's complete `dynamicTools` snapshot.
- `item/tool/call` is the only accepted effect-bearing model-to-Skyra request.
- A dynamic tool call creates a child `InvocationID`; its parent invocation and
  Codex thread remain stable while waiting.
- Native/default effect-bearing items contaminate and fault the invocation from
  `item/started`. A later agent message cannot clear contamination.
- Native item notification is detection, not proof that the hosted effect was
  prevented. Configuration exclusion, version admission, and the kernel process
  boundary remain required.
- Unknown notifications and server requests fail closed. Never invent approval or
  user-input responses.
- Codex protocol fields remain inside the Codex driver's opaque payload.

## Work sequence

| Item | Status | Result |
| --- | --- | --- |
| W0 — contract closure | Complete | `DriverInput`, `DriverEmission`, `DriverTransition`, `HostCommand`, atomic failure, and commit-before-send semantics declared. |
| W1 — transport proof | Complete | `JSONLTransport` implements framed serialized writes, detached reads, idempotent close, and close-unblocks-receive. |
| W2 — resident driver state | Complete | Scheduler clones snapshots and atomically replaces or rolls back `DriverState` with semantic effects. |
| W3 — driver host | Complete locally | `ProcessHost` directly starts one explicitly configured executable and owns stdio transport, stderr, and shutdown; a live App Server handshake passes. |
| W4 — Codex driver port | Complete locally | Thread/turn lifecycle, Focus tools, child invocation/return, and final Return use opaque state. |
| W5 — containment | Complete locally | Native/unknown items contaminate and fault; version-pinned strict launch overrides disable measured native surfaces. OS policy installation is deferred. |
| W6 — recovery proof | Complete | A restored live Codex state without provable request reconciliation faults explicitly and never resets. |
| W7 — integrated proof | Complete with fake App Server | Real DNA16 boot/scheduler/controller/driver calls execute the dynamic tool round trip and native rejection path under the race detector. |

Keep each item independently reviewable. A passing isolated transport test does not
claim durable driver correctness; a passing mocked Codex lifecycle does not claim
process containment.

## First acceptance path

```text
boot possibility and construct restricted Codex host
  initialize App Server transport
  admit parent Invoke A and commit initial DriverState
  start Thread A with Focus-derived dynamicTools
  start Turn A
  receive dynamic ToolCall C
  atomically commit waiting state and child Invoke B
  run child B through the ordinary Skyra dispatcher
  apply child Return B to parent A
  commit resumed state and answer ToolCall C
  receive clean final agent message and successful turn completion
  atomically commit terminal driver state and parent Return A
  deliver one root result
```

The corresponding rejection proof injects a native effect-bearing
`item/started`, commits contamination, faults the invocation, and proves that a
later agent message and `turn/completed` cannot produce a Skyra Return.

## Deferred work

- Multiple simultaneous Codex tool calls.
- Steering, compaction, thread forking, and repeated Focus refresh.
- Reusing or deleting terminal provider threads.
- Production login/account provisioning and multi-user credential isolation.
- General driver-host pooling across possibilities.
- A universal process manager, request router, or transport registry.
- Durable persistence of an entire possibility across Skyra process loss.
- Retries whose correctness depends on provider idempotency guarantees.
- The `Action + Params` to ordered `Operands` migration.
- Final entity/address/capability registry-key encoding.

Deferred work must not be hidden inside generic payloads or implementation-local
side tables.

## Proof and checks

Use deterministic channels or explicit barriers; do not use timing sleeps for
concurrency and shutdown proofs. Required checks for each completed item are:

```sh
go build ./src/skyra/skyra-dna-16/...
go vet ./src/skyra/skyra-dna-16/...
go test -race ./src/skyra/skyra-dna-16/...
```

The integrated proof must additionally establish:

- transport frames remain intact under concurrent sends;
- shutdown releases all blocked reads and driver-host goroutines;
- snapshots cannot mutate resident driver-state bytes;
- driver-state replacement and continuation effects cannot be observed separately;
- stale, duplicate, or mismatched provider events have no semantic effects;
- host loss never silently resets a live invocation to new state;
- only Focus-derived dynamic tools become child Invokes;
- native or unknown effect-bearing items prevent final Return propagation; and
- no Codex identifier or protocol type appears in scheduler contracts.

## Implementation report

Implementation records are maintained in:

- `DECISIONS.md` for settled W0 choices, item status, and deviations;
- `issues.md` for unresolved correctness questions and blocking failures; and
- focused tests beside each implementation rather than one mock-only end-to-end
  test.

Do not mark an item complete because declarations compile. Record the exact tests
run and distinguish mocked protocol proof, live App Server proof, and OS
containment proof.
