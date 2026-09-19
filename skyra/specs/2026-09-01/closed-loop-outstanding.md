# Closing the loop — outstanding decisions

> **Current review:** the [engine proposal checklist](../../src/skyra/skyra-ctu-engine/surface-proposal.md#remaining-issues-before-the-next-handoff)
> tracks the next handoff. Accepted updates: one fd table holds Move + Register; Fire receives claimed
> entries; one ingress splits on Call/Ret; Ret publishes before Landed. The older trace and identity
> examples below are historical where they conflict with those decisions.

**Date:** 2026-09-06
**Status: FOCUS / PARTLY DECIDED.** The next area of focus is the **fault path**. Ownership and the
involuntary-fault return route now follow [engram-loading-and-fault-return.md](engram-loading-and-fault-return.md):
scheduler belongs to the kernel, the kernel may install executable backing into CTU memory, and faults
return through the execution-event transport. This revises [ingress-and-fire.md](ingress-and-fire.md).
The [engine proposal](../../src/skyra/skyra-ctu-engine/engine.go) has not yet been reconciled;
interfaces and algorithms remain to be implemented.

## The boundary we have

```text
Scheduler selects ready fds → records Claimed → CTU.Fire(fds)
    → Memory.Resolve(fd) → MMU.Check(access) → Op.Fire(execCtx)
    → execution event → Scheduler.Apply → next eligible batch
```

Scheduler, inside the kernel, alone writes `MoveID → State`. Kernel coordinates and provisions based on
CTU returns, new invocations, and recovery notifications. Shared Memory owns resident engrams, execution
frames, and its internal structures; kernel and CTU both use its defined operations. Kernel requests
installation/reclamation; CTU reads operands and writes execution results through Memory. MMU checks
declared access and returns a verdict. CTU coordinates execution; opcodes perform their particular behavior.
The fd batch is stable and read-only by contract, not a shared mutable scheduler map.

Claiming precedes dispatch. Issued acknowledges that dispatch; it cannot establish the original claim or
regress a completed entry. Completed records remain until a separate retirement decision.

## Next focus: the fault path

Work one move through a missing-backing fault, recovery, and completion. The intended ownership sequence is:

```text
claimed fd → missing EngramID → Faulted returned to Scheduler → pending recovery
    → Scheduler routes kernel loading → kernel installs engram through Memory interface
    → installation/readiness established → Scheduler makes the same MoveID eligible and claims it again
    → CTU.Fire → access check → execution → completion
```

The return route is decided: CTU reports Faulted once using the execution-return transport, and Scheduler
routes recovery inside the kernel. CTU does not also invoke a handler independently. Faulted does not
complete the move or fill its consumer's result register. No separate Attempt identity is adopted;
the same MoveID is reissued after installation. Event/state encoding and resolution notifications remain
open. A handler must not bypass Scheduler by recursively executing the faulted move.

| Question | What must be pinned down |
|---|---|
| Where is the fault detected? | Memory.Resolve currently returns only move/frame pointers. Define how lookup misses are reported, and how they relate to MMU.NotResident. |
| What identifies the fault? | Materialize must identify the blocked MoveID and missing EngramID. Design the concrete payload and retain the original execution inputs; no separate Attempt ID yet. |
| Where does the move wait? | Define its state in the single scheduler map while recovery is outstanding; it must not be selectable for ordinary execution. |
| Who invokes the handler? | Decided: Scheduler receives Faulted and routes kernel recovery; CTU does not independently invoke the same handler. The remaining role of explicit Trap opcodes is open. |
| How does resolution return? | Define success, still-pending, and failed recovery; kernel installs through Memory's interface, then Scheduler changes eligibility. |
| When is retry eligible? | Establish that backing and operands are usable before re-claiming, and check authority again before access. |
| How are duplicates handled? | A late event from an earlier attempt must not complete or regress a newer attempt. Concurrent faults for the same backing need a materialization discipline. |

Begin with Materialize. Then trace Confirm through the same boundary: human acceptance/refusal is a
resolution to represent, not permission for the executor to mutate scheduling state. Missing or invalid
authority must not silently become a request to load backing or widen a horizon.

## Keep the non-success cases distinct

| Case | Meaning | Required treatment |
|---|---|---|
| Missing operand | A producer has not supplied the value yet | Establish readiness later; no Materialize merely because a dependency is unfinished |
| Missing backing | Required executable/frame backing is not resident | Fault, resolve, install, and retry |
| Denied access | Authority does not permit the access | Explicit denial path; never execute or silently grant authority |
| Confirmation required | An authorized effect requires a gate | Park until the kernel/human resolution permits the next transition |
| Execution failure | The operation reports failure | Preserve the driver Outcome and error observation; define their event representation |

Only ready moves enter Fire. A fault discovered after admission is a runtime event; it does not make
missing-operand work eligible for execution. Exact fault, waiting, and terminal state names remain open.

## Artifact 1: legal transition table

Complete this table before implementing Scheduler.Apply. It includes scheduler operations as well as
execution events: Claim is a scheduler operation BEFORE Fire, not a returned event. Bracketed terms are
descriptive placeholders, not additions to the State/Event enums. References to the current attempt
describe event validity, not a proposed Attempt ID; retry correlation is still open.

| Operation / event | Admissible current state | Resulting state | Duplicate or unexpected arrival |
|---|---|---|---|
| Admit with readiness established | Absent, or [waiting] if adopted | Ready | Must not reset Claimed/Completed; exact rejection/reporting policy open |
| Claim | Ready | Claimed | Only the successful claim enters a Fire batch; other states cannot be dispatched by this claim |
| Issued for the current attempt | Claimed | Claimed | Acknowledgement only; late Issued cannot regress Completed or overwrite recovery state |
| Landed for the current attempt, publication established | Claimed | Completed | No second completion or dependent admission; conflicting results and stale attempts need an explicit policy |
| Faulted for the current attempt | Claimed | [pending recovery] | Must not launch duplicate recovery; stale/terminal-state handling open |
| [Recovery installed / ready notification] | [pending recovery] | Ready once readiness is established | Must not authorize duplicate retry; notification correlation and validation open |
| [Recovery still pending] | [pending recovery] | [pending recovery] | No ordinary dispatch; identify what event can wake the move |
| [Recovery failed] | [pending recovery] | [terminal failure or explicit retry policy] | No automatic Ready transition; failure representation and retry policy open |
| Retire | Completed, subject to retention eligibility | Absent | Must not remove active work; behavior for already-retired records open |

For each open cell, specify whether the event is ignored, rejected, or surfaced as a protocol fault. Avoid
default state assignments that allow an unexpected event to manufacture readiness. Terminal failure
retirement also needs a rule once its representation is chosen.

## Artifact 2: fault-frame lifetime rule

A parked move must retain enough information to recover without reconstructing intent or losing its inputs.
The scheduler map remains lifecycle-only. Retained execution material belongs to CTU/Memory or its backing
store; a kernel handler borrows the information needed for recovery.

| Material | Why recovery needs it | Owner / release question |
|---|---|---|
| MoveID and required EngramID | Associate the blocked execution with its missing durable backing | Scheduler validates lifecycle correlation; no separate Attempt ID adopted; stale-event handling remains open |
| Executable descriptor or resolvable reference | Identify the operation to retry, including when its body is absent | CTU/Memory or backing store; retain recoverability until retry or terminal disposition |
| Operand values or stable references | Retry with the intended inputs | CTU/Memory; do not consume/free the only usable inputs while recovery can still need them |
| Return fd and workflow identity | Route the eventual result to the original consumer | Execution descriptor/frame; preserve through recovery and result publication |
| Horizon and access declaration | Recheck authority on retry | Execution descriptor; retaining a capability does not extend expiry or grant permission |
| Fault vector and recovery request | Let the handler repair the specific missing condition | Fault handoff; specify who retains it while the handler or human is pending |
| Installed backing / recovered frame | Make the next execution attempt resolvable | Kernel installs executable backing through Memory's defined interface; Memory maintains residency consistency and lifetime |

Pin down the release points for recovery success, recovery failure, and final completion. Successful
recovery does not itself consume operands. Terminal failure must account for retained execution resources;
recording a terminal scheduler state alone does not release them. Consumer-held result values and retained
completed scheduler records have independent lifetimes. Exact pinning/refcount and reclamation APIs remain
open; this table records the obligations they must satisfy.

## Artifact 3: worked Materialize fault trace

Use fd `M = MoveID{Possibility: P, Writes: result}` and required engram `E`. Readiness has been established,
but E is absent when execution resolves M. The fault happens BEFORE any external effect. M is preserved
through recovery and reissue; this trace introduces no separate Attempt identity.

| Step | Operation / event | Scheduler state for M | Owner and observable consequence |
|---|---|---|---|
| 1 | Admit M | Ready | Scheduler accepts readiness evidence; CTU/Memory retains execution inputs |
| 2 | Claim M | Claimed | Scheduler records the claim before handing off the fd |
| 3 | CTU.Fire([M]); Memory.Resolve misses | Claimed | CTU detects missing backing; no opcode effect executes; miss-result API open |
| 4 | Return Faulted(M, missing E) | [pending recovery] | Existing execution-return path delivers the fault to Scheduler; M is parked, with no completion or consumer value |
| 5 | Scheduler routes Materialize recovery | [pending recovery] | Kernel.Load(E) derives the exact path, calls the filesystem driver directly, and decodes/verifies E; CTU does not independently invoke a second handler |
| 6 | Load succeeds; E installed | [pending recovery] | Kernel installs E through Memory's defined interface before Scheduler releases M |
| 7 | Recovery readiness notification | Ready | Scheduler validates correlation and readiness; event shape remains open |
| 8 | Reclaim M; CTU.Fire([M]) | Claimed | Scheduler reclaims the same MoveID; Memory resolves; MMU rechecks access before Op.Fire |
| 9 | Execution dispatch acknowledgement, if applicable | Claimed | Issued acknowledges dispatch; later completion arrives independently |
| 10 | Result publication established; Landed | Completed | CTU owns the result; Scheduler records completion and admits only ready dependents |
| 11 | Retention eligibility reached; Retire | Absent | Scheduler removes its record; execution values follow their own release rules |

**Unsuccessful recovery branch:** replace steps 6–11 with:

| Step | Operation / event | Scheduler state for M | Owner and observable consequence |
|---|---|---|---|
| 6F | Handler cannot obtain required backing | [pending recovery] | Kernel reports failure without executing the original operation |
| 7F | Recovery-failure notification | [terminal failure or explicit retry policy] | Scheduler accounts for failure; it does not silently re-admit M |
| 8F | Failure delivered and retained resources reconciled | Representation open | Define how the caller receives the error and how CTU/Memory releases or retains retry material |

At step 5, an empty ready selection means the scheduler is waiting for recovery, not finished. At steps 8–9,
it may be waiting for execution. Record the expected wake source for each parked state. Completion must
account for the root outcome and outstanding work/recovery; lack of ready work alone proves neither.

## Other boundaries needed to close the loop

- **Fd resolution:** locate the executable descriptor and its frame, including root, return, and retry
  correlation. Preserve the reference to EngramID even when backing is absent; distinguish an unknown
  MoveID from a known move with missing engram backing. The current context map alone is not that lookup.
  Do not introduce a second scheduler map or a separate Attempt ID without a concrete need.
- **Readiness:** decide whether Memory/CTU supplies a query or ready events, and who owns dependency indexing.
  Scheduler must have evidence of readiness without storing execution contexts.
- **Event delivery:** distinguish Op.Fire's immediate return from later asynchronous completion. Choose the
  delivery/serialization discipline, handoff failure handling, and backpressure without a batch barrier.
  Faulted uses that same return path; its recovery meaning remains distinct from Landed completion.
- **Result publication:** decide whether Landed reports an already-installed value or precedes CTU
  installation. Dependents cannot run before publication, and Scheduler.Apply never writes registers.
- **Lifetime and completion:** preserve operands needed for retry; define consume/free and fan-out behavior.
  Decide when the root result establishes completion and when records may retire. Neither an empty batch
  nor map emptiness is a sufficient completion test under retained lifecycle records.

These questions may be resolved while tracing the fault path. They are not a request to design every
distributed or multicore mechanism before a local loop can run.

## Evidence that the local loop is closed

Use a tiny local workflow and inspect explicit states/events rather than timing assumptions:

1. One ready move is claimed, resolved, checked, executed, and completed. Its result is available and its
   completed scheduler record remains until retirement.
2. The same move encounters missing backing before its effect, faults, is recovered, and completes after
   re-claim. No direct handler re-entry bypasses Scheduler and no effect fires during the unresolved fault.
3. Two independent moves run; one faults while the other completes. A consumer becomes eligible only when
   its own required values are published. There is no batch completion barrier.
4. Repeated/late notifications cannot cause duplicate admission or regress completion. A recovery failure
   or denied access produces an explicit accounted-for outcome rather than leaving an unexplained claim.

Passing these traces would establish a local execution/recovery loop. It would not establish crash-safe
exactly-once external effects, distributed recovery, or the full REMEMBER compiler. Those remain separate
claims to implement and evaluate.
