# Proposal 2 — scheduler transition loop

Superseded by [proposal 3](../3/loop-proposal.md), which consolidates the accepted
review decisions and the placement-gated continuation algorithm.

Status: working proposal, not an implementation order. This revision records the
agreed ownership and event distinctions following review of
[proposal 1](../1/write-loop.md). Concrete interfaces and dispatch-event opcodes
remain to be specified. This document makes no runtime changes.

The late-feedback issue is resolved by the accepted
[placement-gated continuation algorithm in proposal 3](../3/placement-gated-continuation-algorithm.md).
That algorithm defines the register-owned receipt, pending output, handoff association,
and runtime-first processing that supersede v14's processing of output before
placement acknowledgement.

## transition() owns execution progression

`transition()` owns the progression from an incoming event through the next handoff.
Remove `run` and absorb its iteration over selected snapshots, fault/completion
suppression, saved-context handling, dispatch calls, and error propagation into
`transition()`. Subsequent helper composition must preserve this ownership.

`transition()` acquires the lock at the top and owns it around state application,
contract assessment, execution eligibility checks, and selection/reservation.
State-application and selection helpers
execute within that lock scope and do not acquire their own locks or introduce
another state-transition entry point.

Release the lock before dispatch or host notification. Immediate dispatch
feedback must be able to enter `transition()`; ownership of the execution loop does not
mean holding the mutex across dispatch.

For a selected batch, `transition()` checks resident state before each next handoff under
its lock. If the first child faults synchronously, that fault passes through
`transition()` before it considers the next child. Already-selected work is suppressed
when the root is observed faulted or complete. The resident-state decision belongs
to `transition()`, not a separate `run` loop. This does not claim cancellation of a
handoff already proceeding after the lock is released.

## Two event families and wire formats

| Event family | Meaning |
| --- | --- |
| `continuationEvent` | Requested continuation work from a driver; retains the `Invoke` and `Return` opcodes. |
| `runtimeEvent` (`dispatchEvent`) | Runtime feedback, including placement results and delayed execution/transport/output-processing faults; has its own opcodes and wire format. Contains no authorization result. |

These are distinct opcode families. Their concrete Go representations and the
individual dispatch-event opcodes remain open; the distinction is settled.
`runtimeEvent` names the runtime-feedback family previously called `dispatchEvent`;
these names do not introduce a third family. `continuationEvent` carries Moves.
Dispatch results do not carry an authorization flag or verdict. Authorization
belongs to the continuation-opcode path described below; an authorization denial
is handled there, not reported as dispatch feedback.

`transition()` calls its internal `write()` helper under the existing lock.
`write()` recognizes the event family and delegates to its application helpers.
It returns affected continuations, an optional notification, and any processing
error to `transition()`. It does not acquire a lock, select or dispatch work,
notify the host, or provide another state-transition entry point.
A continuation `Return` contributes an observation from existing work. It is
never a dispatch acknowledgement. Runtime feedback must retain its association
with the invocation whose placement or execution it describes.
Successful placement does not establish a returned observation or successful
execution. Continuation output can arrive before dispatch acknowledgement, but
must remain pending in the producer's register until the matching handoff's
placement success is processed. Combined distributed response envelopes process
the runtime-event portion before their continuation Moves. See the
[accepted algorithm](../3/placement-gated-continuation-algorithm.md).

Runtime feedback enters `transition()` through the bound dispatch callback,
including delayed failures after the original dispatch call has returned.
At the event-family split, the runtime-event branch owns interpretation of that
feedback and application of its associated runtime state changes. A driver or
transport may originate an error; the runtime binds it to the actual invocation
and delivers it as a runtime event, not a continuation Move or Return.

This establishes the fault-delivery ownership boundary. Concrete runtime opcode
names, payloads, and individual state effects can be specified during implementation;
they are not required to settle this structural decision. Preserve source binding,
one root-notification attempt, and propagation of notification/processing errors
without replaying them as execution failures.

A Go function error is a fault to handle, not a third wire format. Fault handling
preserves its source and applies the appropriate state change without manufacturing
a continuation request. Rejected admission must not mutate resident work merely
because the rejected input claimed its identity. Processing or notification errors
must propagate without being replayed as another execution failure.

## Dispatch places the Move

Dispatch interprets the Move wire format and places the Move as specified. It does
not choose a destination or decide execution eligibility. Scheduler event
application records dispatch feedback; it does not acquire dispatch-routing
semantics.

Under the `transition()` lock, `continuationOpcode` checks the trusted Horizon
against current time before branching into `Invoke` and `Return`. It attaches a
runtime-generated check result and checked-at timestamp to the Move before state
application. The opcode handlers gate their state changes on that stamp. For a
Return, the trusted Horizon comes from its bound resident source. The precise
opcode response to an expired stamp remains to be specified.

The Horizon check reads trusted grant data and current time and produces the
stamp. It does not call drivers, dispatch, host callbacks, or scheduler methods
that acquire the lock. This constrained contract makes the accepted check safe
within the existing lock scope; no additional mechanism is required.

The authorization point is this continuation check. A gap between authorization
and execution is accepted; dispatch does not recheck Horizon or report authorization.
Execution eligibility and reservation remain owned by `transition()`. See
[the accepted timing decision](improvements.md#1-specify-authorization-timing--resolved).

## Keep affected and central selection

`affected` is the temporary set or sequence of resident continuations that must
be reconsidered after applying an event. It is not another authoritative state
store or a wire field.

Application helpers update state; `write()` passes affected continuations, an
optional host notification, and any processing error back to `transition()`.
Selection remains visible in the
`transition()` core: it assesses those continuations and reserves eligible execution.

| Applied input | Selection candidates under the current local semantics |
| --- | --- |
| New root Invoke | Newly admitted root. |
| Child Invoke batch | All newly admitted children, in batch order. |
| First accepted child Return | Its stored parent. |
| Root Return | None; prepare a host notification. |
| Identical duplicate Return | None. |
| Placement-success acknowledgement | None by itself; it may release pending continuation output, whose application produces candidates according to the rows above. |
| Rejection or execution fault | None; handle rejection or fault as appropriate. |

Preserve the complete child batch and its bound source through event application.
Validate the entire batch before installing any continuation or parent register
entry, then select only after all entries are installed.

## Contract interpretation

Interpretation is a pure contract-to-contract JSON check. It is separate from
recognizing event families, checking Horizon, validating endpoints, and placing
Moves. Remove proposal 1's generic `protocol.Interpret(input)` event classifier.

Contract interpretation is used where continuation eligibility requires contract
assessment. Its result does not itself perform state transitions or dispatch.
The existing compiled-contract implementation remains a stub; this structural
proposal does not define compiled matching or replace the accepted local
completeness rule with a requirement for a satisfied compiled contract.

## Intended composition

Illustrative structure only; this does not establish concrete Go signatures:

```text
transition(input)
  lock at the top
    write(input)
      split on runtimeEvent (dispatchEvent) or continuationEvent (Move)
      resolve producer/handoff; process combined envelope runtime portion first
      hold continuation output in the register until matching placement success
      continuationOpcode: check Horizon and stamp before Invoke/Return branching
        apply the opcode gated on the stamp
      runtime opcode helper: interpret placement results or delayed faults
        apply associated runtime state changes without authorization data
        release pending output on placement success; discard it on placement failure
      return affected continuations, optional notification, and processing error

    if no processing error:
      assess contracts and execution eligibility for affected continuations
      reserve eligible execution
  unlock

  propagate processing error or notify host when required
  place reserved Moves through dispatch
    keep sequencing and synchronized resident-state checks owned by transition()
    release the lock before each dispatch call
    bound callback delivers runtime feedback into transition()
      runtime-event branch owns delayed-fault interpretation and state application
```

The concrete composition must accommodate immediate reentry as well as delayed
feedback. It must not introduce a second scheduling path or move authoritative
continuation state into a driver, dispatch component, or snapshot.

## Invariants to carry through implementation

- Admit child batches atomically before selecting any child.
- Bind continuation output to its actual source; retain independent child
  identities, stored parent links, inherited grants, and declaration order.
- Record a continuation Return on the existing invocation; distinguish identical
  duplicates from conflicting observations without overwriting accepted results.
- Resume the existing parent once under the current single-wave rules. Preserve
  reservation before handoff and protect `Result`, `Faulted`, and `Resumed` facts
  against late dispatch acknowledgement.
- Keep saved execution contexts and detached dispatch snapshots.
- Gate continuation effects on matching runtime placement success. Keep receipts,
  pending output, and handoff sequence in the producer's resident register as
  specified by the [accepted algorithm](../3/placement-gated-continuation-algorithm.md).
- Keep driver, dispatch, and host calls outside the scheduler mutex.
- Preserve one root-notification attempt and callback-error propagation, without
  automatic replay.
- Keep fault/completion suppression decisions inside `transition()` as it progresses
  through selected work.

## Remaining specification work

Define the concrete event envelopes and dispatch-event opcode set and their state
effects within the accepted producer-invocation/handoff association. Settle lifecycle state names
without losing the distinction between reservation before handoff and placement
feedback afterward. The runtime-event branch owns fault interpretation and state
application; its concrete opcode details and helper composition remain implementation
work within that settled boundary.

The current runtime applies continuation `Return` directly and dispatches only
`Invoke`. Separate event families do not by themselves change that behavior.
Any extension to placement of continuation `Return` requires an explicit decision.
The accepted placement gate changes when continuation output may be applied;
placement success may now release buffered output for processing and selection.
Current notification-producing transitions do not also select work;
any extension that needs both must define that behavior.

## Code reference and verification

The current implementation is
[`scheduler.go`](../../../../../src/skyra/skyra-dna-v14/scheduler/scheduler.go).
Existing controller, protocol, and CTU code must be considered when translating
the event boundaries; the change is not only an extraction of scheduler helpers.

Verify a later implementation with immediate, reversed, concurrent,
duplicate-return, fault-suppression, saved-context, snapshot-isolation, and
callback-error tests, including the race detector. Directly verify scheduler batch
rejection without partial admission as well as controller-side validation.
Compiled matching, retries, repeated execution waves, persistence, and retirement
remain outside this proposal.
