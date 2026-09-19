# Proposal 3 — scheduler transition loop with confirmed placement

Status: consolidated proposal recording accepted decisions; not implemented and
not an implementation order. This supersedes
[proposal 2](../../../../../specs/2026-09-09/loop%20improvements/loop%20proposals/2/loop-proposal.md) as the current proposal. The
[placement-gated continuation algorithm](placement-gated-continuation-algorithm.md)
is the companion specification for continuation and handoff state, feedback association, and
processing order. Concrete Go interfaces, runtime opcodes, and wire encoding
remain implementation work.

## Execution progression belongs to transition()

`transition()` owns progression from an incoming event through the next handoff.
It acquires the scheduler mutex at entry and owns state application, contract
assessment, execution eligibility, and selection/reservation within that lock.
Application helpers run under the existing lock; they neither acquire it again
nor introduce another state-transition entry point.

The internal `write()` helper recognizes the event family and delegates to the
appropriate application helper. It returns affected continuations, an optional
host notification, and any processing error. It does not select, dispatch, notify
the host, or own another execution loop. Selection stays visible in `transition()`.

Absorb `run(selected)` into `transition()`, including its saved-context handling,
iteration over reserved snapshots, resident root-state checks, dispatch calls,
and error propagation. Release the mutex before dispatch and host notification.
Immediate and delayed feedback can reenter `transition()`.

Before each next handoff, check resident root fault/completion under the mutex,
then unlock before dispatch. A synchronous fault from one child is processed
before considering the next selected child. Suppression does not claim to cancel
a handoff already proceeding after unlock.

## Two event families

| Family | Meaning and interpretation owner |
| --- | --- |
| `runtimeEvent` (`dispatchEvent`) | Runtime placement feedback and delayed execution, transport, parse, or normalization faults. The runtime-event branch owns interpretation and associated runtime state application. |
| `continuationEvent` (Move) | Continuation transitions with Invoke and Return opcodes. The continuation-opcode branch owns the Horizon check and opcode state application. |

`runtimeEvent` is the name for the family previously called `dispatchEvent`;
there are two families, not three. A combined distributed envelope can carry
both families without introducing another opcode family.

A dispatch callback carrying `runtimeEvent` (`dispatchEvent`) does not perform
a Horizon authorization check. Dispatch results contain no authorization flag or
verdict. Source/handoff binding is still required; it is separate from continuation
authorization. Placement success
confirms the handoff; it does not establish an observation or successful execution.
A continuation Return contributes an observation from an existing invocation.
It is never a runtime placement acknowledgement.

Runtime feedback enters `transition()` through the bound dispatch callback,
including failures reported after the original dispatch call has returned. A
driver or transport can originate an error; the runtime establishes its source
association and delivers a runtime event without manufacturing a Move or Return.
Concrete fault opcode names and payloads can be specified within this ownership
boundary during implementation.

A Go function error is not another wire family. Execution faults retain their
source. Processing and notification errors propagate without being replayed as
execution failures. Rejected external admission must not mutate resident work
merely because the rejected input claimed its invocation identity.

## Confirm placement before processing continuation output

Continuation output cannot advance execution until the runtime has confirmed
successful placement of the handoff that produced it. Receive early output into
the producer's resident `Continuation.Handoff`, but defer its Invoke/Return
effects until confirmation.
This gate concerns output from dispatched work; external root admission has no
producing handoff to acknowledge and retains its validated root-admission path.

The producing invocation's resident `Continuation.Handoff` owns:

- The handoff sequence and placement status: pending, succeeded, or failed.
- Pending continuation output, retaining the whole batch and declaration order.
- Whether that handoff's output has been consumed.

`Handoff` is a sibling of `Move` on the resident `Continuation`.
`Move.Invoke.Register` remains the semantic continuation frame: its contract,
accepted child declarations, accepted child observations, and declaration order.
Received output belongs to `Handoff`; only accepted declarations and observations
enter the appropriate register through opcode application.

Both records remain scheduler-owned and protected by the same mutex. This split
introduces no additional state store, manager, or lock. Drivers receive a detached
Move with its semantic Register, without placement receipts, pending output, or
consumption flags. The runtime binds the invocation ID and handoff sequence through
the callback or transport envelope; that association is not driver-owned state.

Parent and child InvocationIDs are distinct; they share a PossibilityID. Match
feedback to the producing InvocationID and handoff sequence. A parent emitting
children is the producer of that declaration; each admitted child has its own
invocation identity. A returning child is the producer of its Return, and its
stored parent link locates the receiving parent entry.

The same parent's initial dispatch and resumption reuse its InvocationID.
Runtime-bound handoff sequences distinguish initial `0` and resumed `1`. Stale
initial feedback cannot satisfy the resumed handoff's gate. This association
does not add retries or create another invocation.

When returning both runtime feedback and continuation output over the wire, wrap
the associated portions together and process the runtime portion first. When
portions arrive separately, match them through the resident handoff record.
Placement success releases pending output; failure discards it and applies the runtime fault;
unconfirmed placement leaves it pending. The wrapper itself is not proof of a
valid source: the runtime must establish the bound association.

An immediate local output callback must be able to buffer output and return
acceptance without waiting for placement feedback. Current ingress can produce
output before returning its handoff acknowledgement; waiting in that callback
would deadlock the acknowledgement path.

The [companion algorithm](placement-gated-continuation-algorithm.md) specifies
the full sequence and duplicate/stale-feedback requirements. Pending-output
storage is runtime bookkeeping, not semantic admission of children or acceptance
of a returned observation.

## Horizon check and stamp

For continuation output released by the placement gate, enter
`continuationOpcode` under the scheduler lock. At its top, before branching into
Invoke and Return, check the trusted Horizon against current time. Attach the
runtime-generated check result and checked-at timestamp to the Move before
Invoke/Return state application. Root admission also passes through this check.
Authorization is required for both **Invoke and Return**. Each opcode handler
gates its state changes on its runtime-generated stamp. A Return does not bypass
authorization because its producing Invoke was previously authorized or its
placement succeeded.

The event family determines this check, not delivery through a callback. Processing
a runtimeEvent does not check authorization; if that event releases buffered
Invoke/Return output, each released Move enters the continuation authorization
path before its opcode effects. A combined envelope follows the same rule: process
the runtime portion first, then authorize its eligible continuation Moves.

The check reads trusted grant data and current time. It does not call drivers,
dispatch, host callbacks, or scheduler methods that acquire the mutex. The
arbitrary callback behavior of v14's `Authority.Check` is not moved under the
lock unchanged.

Return carries no Horizon in v14; resolve its trusted Horizon from the bound
resident source. Checking Return authorization is required; only the precise
opcode response to an expired stamp remains open, including how the Return
handler treats it.

This is the authorization point. A gap between authorization and execution is
accepted; dispatch neither rechecks Horizon nor reports authorization. A successful
check at `t=1`, expiry at `t=3`, and execution at `t=4` does not by itself violate
this timing guarantee. AWS likewise accepts separation between request acceptance
and later work, as illustrated by
[asynchronous Lambda queuing](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html)
and [S3 downloads continuing after URL expiry](https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-presigned-url.html#PresignedUrl-Expiration).
These examples do not establish an AWS-wide maximum delay or automatic rollback.

## Dispatch and selection

Dispatch interprets the Move's routing information and places it as specified.
It does not choose the destination, determine execution eligibility, or perform
Horizon authorization. Scheduler event application does not take over routing.

`affected` is a temporary sequence or set of resident continuations to reconsider,
not another authoritative state store or wire field.

| Applied input | Selection candidates |
| --- | --- |
| New root Invoke | Newly admitted root. |
| Child Invoke batch released through the placement gate | Newly admitted children in declaration order. |
| First accepted child Return | Its stored parent. |
| Root Return | None; prepare the host notification. |
| Identical duplicate Return | None. |
| Placement success | None by itself; released continuation output can produce candidates through its application. |
| Unconfirmed placement with buffered output | None. |
| Rejection or runtime fault | None; apply the rejection or fault handling. |

Validate a complete child batch before installing any child continuation or
parent-register entry. Install all entries before selecting any child. Source
binding, independent child identity, inherited grants, and order remain intact.

Contract interpretation is a pure contract-to-contract JSON check, separate from
event recognition, Horizon checks, endpoint validation, and routing. Assessment
does not perform state transitions or dispatch. Compiled matching remains a stub;
the accepted local completeness rule permits one parent resumption when all its
declared children have observations. It does not assert compiled satisfaction.

## Intended composition

Illustrative ownership and order, not concrete signatures:

```text
transition(input)
  lock
    write(input)
      resolve bound source/handoff for dispatched-work feedback
      process a combined envelope's runtime portion first
      runtime-event branch
        interpret placement feedback or delayed faults without Horizon authorization
        update Continuation.Handoff; release or discard pending output as appropriate
      continuation-event branch
        hold producer output until its placement is confirmed
        for eligible output, or validated external root admission:
          continuationOpcode: check Horizon, stamp, branch Invoke/Return
          apply opcode state changes gated on the stamp
      return affected continuations, optional notification, processing error
    if no processing error:
      assess affected continuations
      select/reserve work and initialize pending Continuation.Handoff records
  unlock
  propagate processing error or notify host when required
  dispatch reserved snapshots
    check resident root state under lock before each handoff
    unlock before external calls
    bound feedback reenters transition()
```

## Change from current DNA v14

Consolidating `run` into `transition()` centralizes composition and ownership.
Current v14 already has one mutation/selection/reservation path; its `run(selected)`
does not select work again. Removing that helper alone adds no correctness guarantee.

The accepted behavioral changes are the continuation authorization point and stamp,
dispatch feedback without authorization, and placement-gated continuation processing
with continuation-owned Handoff records for placement, pending output, and consumption.
In v14, early output can advance execution before placement acknowledgement. In this proposal,
placement success may release buffered output for processing and selection.

The scheduler remains the sole owner of resident state. Preserve saved contexts,
detached snapshots, terminal Result/Faulted facts, one parent resumption, and one
root-notification attempt with callback-error propagation. No driver, dispatch,
or host call executes under the scheduler mutex.

## Remaining implementation detail and scope

Define concrete runtime opcodes, envelopes, Handoff fields, stamp representation,
and individual fault state effects within these accepted boundaries. Settle the
Invoke/Return response to the stamp, including expiry. Endpoint validation must
retain its existing guarantees without violating external-call lock boundaries.

Return remains applied directly; only Invoke is dispatched. Placement confirmation
for a producer does not turn its continuation Return into another dispatched
invocation. Notification-producing transitions do not also select work under the
current local rules; an extension requiring both needs an explicit decision.

Retries, repeated execution waves, persistence, retirement, compiled matching,
production authority verification, and concrete network transport are outside
this proposal. Describing the distributed envelope does not implement transport.

## Verification

Preserve the immediate, reversed, concurrent, source-binding, duplicate-return,
fault-suppression, context, snapshot-isolation, and callback-error proofs. Add the
[algorithm's ordering and handoff tests](placement-gated-continuation-algorithm.md#difference-from-current-v14-and-verification),
verify authorization and the Horizon stamp precede both Invoke and Return effects.
Verify runtimeEvent callbacks do not perform Horizon authorization, while output
released by placement success does pass through that check. Directly test scheduler
batch rejection without partial admission. Verify driver-facing Moves include the
semantic Register without exposing or permitting mutation of resident Handoff
bookkeeping. Build, vet, and run the scoped race tests when implementation exists; this document does not claim those changes are
implemented or verified.

Current code references: [scheduler](../../../skyra-dna-v14/scheduler/scheduler.go),
[controller](../../../skyra-dna-v14/controller/controller.go),
[protocol](../../../skyra-dna-v14/protocol/protocol.go),
and [CTU](../../../skyra-dna-v14/ctu/ctu.go).
