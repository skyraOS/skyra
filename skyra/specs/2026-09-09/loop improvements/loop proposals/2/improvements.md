# Proposal 2 — improvements to work through

Status: review checklist with accepted decisions recorded below; not an implementation order.
These items assess [the proposal](loop-proposal.md) against the invariants and
current DNA v14 behavior. No runtime changes are requested by this document.

All five review items are resolved. The accepted decisions are consolidated in
[proposal 3](../3/loop-proposal.md). Verification checkboxes below remain pending
implementation.

The proposal preserves the central rules: atomic child admission, source-bound
output, scheduler-owned state, one parent resumption, detached snapshots, saved
contexts, and one root-notification attempt. The decisions and their resolutions
are recorded below.

## 1. Specify authorization timing — resolved

- [x] Establish the authorization point before continuation state application.

**Accepted decision:** after acquiring the scheduler lock, check the trusted
Horizon against current time at the top of `continuationOpcode`, before branching
into `Invoke` and `Return`. Attach the runtime-generated check result and checked-at
timestamp to the Move before writing it to resident state. Each opcode handler
gates its state changes on that stamp and decides how to handle the result.

```text
transition(input)
  lock
    continuationOpcode(move)
      check trusted Horizon against current time
      attach check result and checked-at timestamp
      branch into Invoke or Return
        gate state application on the stamp
    assess and reserve eligible work
  unlock
  dispatch reserved work
```

**Guarantee:** the stamp records the authorization decision at this defined check
point. A gap between authorization and execution is accepted. This decision does
not require another check immediately before dispatch and does not promise that
execution begins or finishes before Horizon expiry. For example, a successful
check at `t=1`, expiry at `t=3`, and execution at `t=4` does not by itself violate
the timing guarantee.

**AWS evidence:** AWS accepts a gap between authorization and execution. Lambda
[accepts asynchronous events into a queue for later execution](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html).
S3 explicitly [allows a download accepted before URL expiry to continue after expiry](https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-presigned-url.html#PresignedUrl-Expiration).
These examples support a defined request-acceptance boundary; they do not specify
an AWS-wide maximum authorization-to-execution delay or promise automatic rollback
of effects after expiry.

**Current v14:** `ContinuationController.Dispatch` calls `Authority.Check` for
each dispatched invocation. Preserving that precise placement is not a requirement
of the accepted design. The earlier review incorrectly elevated this implementation
choice into a required invariant; that requirement is withdrawn.

**Accepted event boundary:** dispatch results contain no authorization flag or
verdict. Authorization denial is handled in the continuation-opcode path using
the check stamp, not carried as dispatch feedback. References to existing v14
authorization behavior describe the code being replaced, not the proposed envelope.

**Separate opcode-policy detail:** define how each handler responds to the stamp,
including an expired Return. Current Return carries no Horizon, so its trusted
Horizon must be resolved from the bound resident source. These details do not
reopen the settled timing/placement decision.

## 2. Specify the authorization lock boundary — resolved

- [x] Specify the check implementation's contract so it is safe under the
  scheduler mutex at the placement accepted in item 1.

**Accepted decision:** the Horizon check reads trusted grant data and current
time and produces the check stamp. It does not call drivers, dispatch, host
callbacks, or scheduler methods that acquire the lock. This is the check's
contract at the top of `continuationOpcode` under the existing scheduler lock;
no additional mechanism is required.

**Invariant to preserve:** moving authorization must not prevent valid scheduler
reentry or make policy code deadlock against scheduler observation.

**Current v14:** `Authority.Check` is an arbitrary injected Go method, invoked
outside the scheduler mutex. A policy can inspect a CTU snapshot.

**Adversarial sequence:** `transition()` holds the mutex while assessing
eligibility, calls the existing authority, and that authority calls `Snapshot`.
The snapshot waits for the mutex held by its caller.

**Resolution:** use the constrained check contract above. The arbitrary callback
behavior of current v14 is not carried under the lock unchanged.

## 3. Define source-bound asynchronous fault delivery — resolved

- [x] Place delayed-fault interpretation and state application in the runtime-event
  branch at the event-family split within `transition()`.

**Accepted decision:** there are two event families: `runtimeEvent` (the family
previously called `dispatchEvent`) and `continuationEvent` (Move). Runtime feedback
enters `transition()` through the bound dispatch callback. The runtime-event branch
owns interpretation and state application for placement feedback and delayed
execution, transport, parse, or normalization faults. It contains no authorization
result. Authorization remains in the continuation-opcode check-and-stamp path.

An error may originate in a driver or transport, but the runtime binds its source
and delivers a runtime event. It does not manufacture a continuation Return.

**Invariant to preserve:** an execution failure faults the actual producing
invocation and its workflow. An error processing or reporting that failure must
propagate without being treated as another execution failure.

**Current v14:** the bound output callback writes
`ExecutionFailure{Source, Err}`. The error returned by that write is a separate
processing/reporting outcome. Rejected external admission uses a separate path
that cannot fault resident work merely because the input claimed its identity.

**Adversarial sequence:** ingress accepts work and returns nil. Later, its output
callback reports a transport failure. There is no original dispatch stack left
through which to return that execution failure. The bound source must still reach
the scheduler, and a host-notification error must not replay the fault.

**Resolution scope:** identifying this callback path and interpretation owner is
sufficient for the structural proposal. Concrete opcode names, payloads, and
individual fault state effects can remain open until implementation. Source binding,
rejected-admission isolation, and callback-error propagation remain required
invariants; leaving opcode details open does not reopen the ownership issue.

## 4. Settle the meaning of late dispatch feedback — resolved

- [x] Gate continuation processing on the corresponding runtime placement-success
  message, with receipt and pending-output state in the producer's register.

**Accepted algorithm:** see
[proposal 3 — placement-gated continuation algorithm](../3/placement-gated-continuation-algorithm.md)
for the full sequence. Runtime feedback is processed first; early continuation
output is held without applying its Invoke/Return effects. Success releases it,
failure discards it and applies the runtime fault, and missing confirmation keeps
it pending. Immediate output callbacks accept buffering without waiting for the
placement acknowledgement.

Parent and child InvocationIDs remain distinct; they share a PossibilityID.
Feedback matches the producer's InvocationID and runtime-bound handoff sequence
(initial `0`, resumed `1`), held in its register. Stale initial feedback cannot
satisfy the resumed handoff. Combined distributed envelopes carry the associated
runtime and continuation portions and process the runtime portion first.

**Invariant to preserve:** late feedback cannot erase a result, reopen faulted
work, or reserve another parent resumption.

**Current v14:** initial execution and parent resumption share an invocation ID.
Successful acknowledgements stop updating dispatch state once `Resumed` is true.
This also prevents a successful acknowledgement of the resumed handoff from
marking it `InFlight`. Returned or faulted invocations ignore dispatch feedback.

**Adversarial sequence:** the initial root handoff emits children synchronously;
the children return and the root resumes before the initial ingress call returns.
The resumed handoff remains pending when the initial acknowledgement arrives.
Invocation identity alone does not distinguish these two handoffs.

**Resolution:** replace v14's early-output progression with the accepted placement
gate and handoff association. This is a behavioral change with additional
register-owned state. Concrete types and runtime opcode details remain
implementation work within the settled boundary; retries remain out of scope.

## 5. State what scheduling consolidation changes — resolved

- [x] Describe absorbing `run` into `transition()` as an ownership/composition
  change, and identify any intended behavioral change separately.

**Invariant to preserve:** one authoritative mutation and reservation path, with
resident root-state checks before subsequent handoffs and external calls outside
the mutex.

**Current v14:** `Write` already mutates, selects, and reserves under the mutex.
`run(selected)` dispatches saved snapshots without selecting again. It checks root
fault/completion before each handoff and propagates callback errors.

**Accepted clarification:** absorbing `run` centralizes composition; it does not
remove an existing second selection path. The authorization placement/stamp and
placement-gated continuation algorithm are behavioral changes, documented
separately in [proposal 3](../3/loop-proposal.md#change-from-current-dna-v14).
Preserve ordering and lock boundaries. Suppression does not cancel a handoff
already proceeding after unlock.

## Verification after decisions are settled

Retain the existing immediate, reversed, concurrent, duplicate-return,
fault-suppression, context, snapshot, and callback-error proofs. Add focused checks
for the decisions above:

- [ ] The Horizon check and runtime-generated stamp precede Invoke/Return state
  application; both handlers consume that decision.
- [ ] A check at `t=1` followed by expiry at `t=3` and execution at `t=4` is not
  rejected solely because time passed after acceptance. No per-dispatch recheck
  is required by the accepted timing decision.
- [ ] The check obeys the lock-safe contract settled in item 2.
- [ ] A delayed execution failure after successful ingress acknowledgement faults
  the correct source/root; a notification error propagates without replay.
- [ ] Verify the accepted algorithm's output-before-success, success-before-output,
  combined-envelope, placement-failure, and duplicate/stale-handoff cases. Initial
  feedback cannot satisfy the resumed handoff or cause another resumption.
- [ ] Direct scheduler rejection of an invalid child batch leaves no partial
  continuation or parent-register installation. The current integration test
  rejects an invalid endpoint in the controller before scheduler admission.

Compiled matching, retries, repeated execution waves, persistence, and retirement
remain outside this review. Their absence is not a defect in this structural
proposal.

## Current code references

- [Scheduler state application, selection, and dispatch](../../../../../src/skyra/skyra-dna-v14/scheduler/scheduler.go)
- [Controller authorization, output binding, and fault reporting](../../../../../src/skyra/skyra-dna-v14/controller/controller.go)
- [Protocol event and continuation types](../../../../../src/skyra/skyra-dna-v14/protocol/protocol.go)
- [CTU admission](../../../../../src/skyra/skyra-dna-v14/ctu/ctu.go)
- [Integration tests](../../../../../src/skyra/skyra-dna-v14/ctu/ctu_test.go)
