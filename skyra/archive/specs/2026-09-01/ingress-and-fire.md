# Scheduler owns scheduler state; the CTU owns execution

> **PARTLY SUPERSEDED (2026-09-06):** [engram-loading-and-fault-return.md](engram-loading-and-fault-return.md)
> places scheduler inside the kernel and permits kernel installation of executable backing into CTU
> memory. Involuntary faults return once to scheduler, which routes kernel recovery. The lifecycle
> single-writer rule, claim-before-dispatch, and separation of execution from scheduling still hold.
> Conflicting kernel-installation restrictions and unresolved fault-routing statements below are historical.

**Date:** 2026-09-05
**Status: DECISION — reflected in the engine proposal on 2026-09-06; algorithms remain stubbed.** Captures the discussion of table ownership and the CTU
execution boundary. Supersedes the ownership and separate Work/InFlight representation in
`the-reactor.md`, `../2026-08-27/register-lifecycle-and-transition.md`, and the execution-engine proposal
in `../../src/skyra/skyra-ctu-engine/engine.go` — including its `Op.Fire(host, move)` direct table mutation,
now revised so Fire RETURNS an event (see "Execution returns an event"). Other decisions remain in force
unless revised here.

## The decision

Scheduler interprets scheduling events and is the single writer of scheduler state: **one map from MoveID to
State**. It records lifecycle transitions and hands the CTU an executable block containing only ready moves.
The CTU owns execution contexts, operands, registers, and results. Its execution boundary is
**Fire(list of fds)**. Completion and fault events return through scheduler.

This refines the first version of this decision: "only scheduler writes state" means SCHEDULER state, not
all state in the machine. Contexts and observations are not values in the scheduler map.

```text
incoming invocation ─┐
completion / fault ──┴─→ SCHEDULER ── ready, issued batch ──→ CTU.Fire(fds)
                           ↑                                  │
                           └──────── result / fault events ────┘
```

Scheduler names the component that owns lifecycle state, admission, and claims. Ingress refers only to
its event-entry boundary (Apply), not to the component itself. Local completions, remote results, new invocations, and fault resolutions all enter
through this owner. This is a logical ownership boundary; its concrete placement and API are still open.

## One scheduler map: MoveID → State

Work and InFlight represent the same move at different points in its lifetime. Keep one entry per MoveID,
with its lifecycle state as the value:

```go
map[MoveID]State
```

The stable MoveID is the key, not the whole executable Move. The map contains neither a Context nor an
Observation nor the executable move payload. Executable moves are resolved separately for Fire; the
descriptor storage and lookup interface remain open.

Claiming and completing update this one entry. The ordinary lifecycle is `ready → claimed → completed`;
these names are provisional, and fault/waiting states have not yet been specified.

| Operation | Meaning in the unified table | Writer |
|---|---|---|
| Admit ready work | Establish an entry eligible for execution | Scheduler |
| Claim / issue | Record the entry as issued before handing it to Fire | Scheduler |
| Complete | Mark the entry completed after execution reports completion; retain the entry | Scheduler |
| Retire | Remove a completed lifecycle record under a separate retention policy | Scheduler |
| Fault / resolution | Record the event and arrange any eligible re-drive | Scheduler |

Ready and issued remain distinct conditions. This explicitly reverses the prior rule that those conditions
must be represented by membership in separate tables. Internal indexes may support selection, but do not
become independent sources of truth.

Completion is not deletion. This supersedes the first version's "retire the pending entry on completion."
When completed records can be removed is a separate decision. Consequently, map emptiness is not the test
for quiescence: a completed run can still have lifecycle records in the map.

**Still open:** whether moves with missing operands already have waiting entries in this table, or remain
outside it until ready. Nothing in this decision permits those moves into a Fire block.

## What makes a move ready

There are two ordinary sources of newly ready work:

1. An invocation introduces moves whose required inputs are already available.
2. A completion lands an input that makes a dependent move ready.

Scheduler handles the scheduling consequences of both. Because the CTU owns the operands, it must expose
enough information to establish readiness or emit a ready event. Which mechanism supplies this information,
and who maintains the dependency index, remain open. Scheduler need not store a Context to receive that signal.
The CTU does not select or claim work from the scheduler map; Fire receives only ready moves.

## Fire receives an executable block

The proposal call shape is `CTU.Fire(fds []MoveID)`: a stable batch of identities for ready, already-claimed
moves. It is a selection from the scheduler's one authoritative map, not a second move map. The CTU treats
the slice as read-only; its backing storage must remain stable while in use. The mutable scheduler map is
not handed to execution. An immutable view could replace the slice later without changing this ownership.

For each fd, `CTU.Fire` calls `Memory.Resolve(fd)` to obtain its executable move and resident frame, constructs
the execution context, calls `MMU.Check` for any declared access, and invokes `Op.Fire` only when authorized.
Memory and MMU are structs contained by the CTU. Memory owns resident contexts and resolution; MMU owns checking;
the CTU owns batch execution. `MMU.Check(horizon, target, action)` remains opcode-agnostic and returns a
verdict without executing anything. This replaces the former CTU `resolve` and `Check` helpers; `step` is
retired. Descriptor storage, lookup failure handling, check dependencies, and event delivery remain open.

- Scheduler records each move as issued before handing it off, preventing a second admission from selecting
  the same outstanding work.
- Fire dispatches the supplied moves without waiting for the entire batch to finish.
- A batch implies neither an execution order nor a completion barrier. Its moves can complete independently.
- A dependent move can be released as soon as its own inputs arrive; unrelated batch members need not finish.
- The executor manages its execution state and reports completion or fault events. It does not mutate the
  scheduler map.

The handling of a failed handoff after issuance remains an explicit implementation question. Recording
issuance alone does not establish crash recovery or exactly-once external effects.

## Execution returns an event; the event drives scheduler state

This is the resolution of the standing tension with the self-executing-opcode decision (D11), which had
`Op.Fire(host, move)` mutate the scheduling tables directly. Fire is retained as the transition, but it
does not write the scheduler map. **Fire runs the divergent execution in CTU-owned state and RETURNS an
event; scheduler applies that event to the map.** The opcode still drives the state — it drives it by
choosing which event to emit, not by performing the write.

So the return of Fire becomes an event, not a table mutation:

```go
// The closed, opcode-agnostic vocabulary scheduler switches on. Fire returns one of these.
type Event interface{ isEvent() }
type Issued  struct{ Move MoveID }                      // dispatched; now awaiting a child or the driver.
type Landed  struct{ Fd MoveID; Value c.Observation }   // a value arrived at Fd; wake dependents reading Fd.
type Faulted struct{ Move MoveID; Vector Kind }          // trap to the kernel/human; arrange re-drive.

// Fire's return type becomes Event (was a direct scheduler-map write).
// One opcode; distinct from CTU.Fire(fds), the batch dispatch boundary.
Fire(x *ExecCtx) Event
```

Two axes, deliberately not crossed:

```text
OPCODE   the EXECUTION axis — open, growing (Call/Exec/Ret/Trap/…). polymorphic behind Fire. switched once,
         where behavior genuinely diverges (materialize a child vs. touch a driver vs. land a value).
EVENT    the SCHEDULER axis — closed, small (Issued/Landed/Faulted). scheduler switches on THIS, authored once.
```

The scheduler must not switch on the opcode. Different opcodes collapse onto the same event — Call and Exec
both merely become `Issued`; the scheduler does not care that one materializes a child and the other hits a
driver, because that difference is pure execution and lives behind Fire. Switching scheduler on the opcode
would re-encode each opcode's meaning a second time (once in Fire, once in the scheduler switch) and at every
transition point — the exact enum/payload desync D11 exists to prevent. Switching on the closed event set
keeps the transition authored once in Fire (which event to emit) and enacted once in scheduler (`Apply`), with
one apply site rather than a per-stage opcode switch:

```go
func (s *scheduler) Apply(e Event) {   // scheduler: the single writer, one site, closed switch.
    switch e := e.(type) {
    case Issued:  // acknowledge dispatch; claim already recorded BEFORE Fire; never regress Completed
    case Landed:  // mark completed; admit dependents whose readiness is now established
    case Faulted: // park; arrange any eligible re-drive
    }
}
```

**The carrier form — what Fire returns is DATA, not a function.** Two acceptable shapes, one rejected:

```text
EVENT AS DATA (plain switch)      Fire returns a concrete event; scheduler `Apply` switches on the closed set and
                                  writes. the event→write mapping is authored once, in scheduler.
EVENT WITH METHOD (Apply)         the event carries `Apply(w SchedulerWriter)`; dispatch is by type (no central
                                  switch), but the event is still a concrete data value and scheduler still owns
                                  the pen — the method may only touch the map through the narrow SchedulerWriter
                                  it is handed. a style choice against the plain switch; both keep every property.
REJECTED — WRITEBACK CLOSURE      Fire returns `func(*scheduler)`. rejected: (1) it re-inverts the two axes — the
                                  opcode re-authors scheduler-write logic inside the closure and must close over
                                  the scheduler's mutation API, re-coupling execution to scheduler internals;
                                  (2) a closure is opaque — it cannot cross the Egress/Ingress barrier (remote
                                  completions must arrive as data), cannot enter the REMEMBER trace (which is
                                  walked and compiled), and cannot be logged/replayed/unit-tested; (3) the data
                                  form is required for the remote and trace paths regardless, so a local closure
                                  is only a redundant SECOND representation of the same transition.
```

The dividing line: Fire knows *what happened* (only execution can); scheduler knows the *scheduling
consequence* (only the scheduler owns readiness/admission). The event is the interface between those two
distinct pieces of knowledge — not a lossy re-decode of one into the other. Data on both sides keeps the
barrier, the trace, and replay open; a function forecloses all three. Plain-switch vs. Apply-method is left
open as style; the writeback closure is closed out.

**Still open:** the event vocabulary's completeness (whether Issued/Landed/Faulted suffice, or a blocked/
waiting event is needed for missing-operand moves), and the `Faulted` payload — the retry identity and the
operand/context state a re-drive must carry. These are the event-shape questions; the seam itself is settled.

## Contexts and registers belong to the CTU

The scheduler map tracks lifecycle state. The CTU manages the Contexts needed to execute moves, including
operand access, register landing, results, and value release. Registers retain produced observations until
consumption; their lifetime is distinct from the scheduler record of the producing move.

The logical flow is:

```text
CTU execution → result available in CTU-owned state → completion event
scheduler      → mark completed → admit dependents whose readiness is established → issue ready batch
```

Scheduler does not write registers or own their lifetime. The result-publication/event ordering, exact
consume/free point, operand lifetime across faults, and duplicate-result handling remain to be worked.
This decision establishes ownership without prescribing a locking or transaction implementation.

## Interpretation and authority

Scheduling-event interpretation and the preparation of executable batches live at scheduler. Opcode execution
lives behind Fire. This revises the engine proposal where `Op.Fire(host, move)` directly mutates scheduling
maps: execution reports events to change scheduler state, while retaining ownership of its own contexts.

Authorization must still be valid when an action touches its target. Admission-time interpretation does not
remove the per-access authority requirement: a queued capability can expire before execution. The precise
placement of verification relative to scheduler and Fire is open; this decision does not authorize treating
an earlier admission check as indefinitely valid.

Kernel authority remains distinct from scheduler ownership of scheduler state. Scheduler does not acquire the
right to mint capabilities or widen a workflow horizon. Fault handling can involve the kernel or human;
its resolution returns through scheduler to update scheduler state and arrange any reissue. CTU-owned
execution state is prepared through the execution boundary, not written by scheduler.

## Why this boundary

Keep the CTU loop small by separating scheduling from execution and making execution compositional.
Scheduler-state interfaces should expose permitted lifecycle operations; opcodes report events rather than
reconstruct scheduling rules using unrestricted map writes. The interface methods are a next design step.

The settled split is:

```text
SCHEDULER   own MoveID → State; interpret scheduling events; record claims/completions; submit ready batches
CTU       own Contexts/operands/registers/results; Fire(fds); report readiness information and execution events
```

## Questions left for the next pass

- The exact State vocabulary and the home of not-yet-ready moves.
- Completed-record retention and retirement.
- The scheduler event interface, delivery mechanism, and serialization discipline.
- The readiness query/event boundary and ownership of dependency indexing.
- Executable-move lookup for batches and operand access within CTU-owned contexts.
- Fault parking, retry identity, duplicate events, and recovery from failed handoff.
- Consumption and value release, including fan-out and re-drive.
- Completion/quiescence detection with retained completed entries; an empty Fire batch alone is not completion.
- The exact authority-check boundary and the state-interface methods available to each component.
- The event vocabulary's completeness (a blocked/waiting event for missing-operand moves?) and the `Faulted`
  payload — the retry identity and the operand/context state a re-drive must carry.

The engine proposal now reflects these boundaries in its types and stubs. Runtime behavior is not implemented.
