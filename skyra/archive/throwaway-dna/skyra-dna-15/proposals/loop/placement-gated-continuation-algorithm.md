# Placement-gated continuation algorithm

Status: accepted algorithm, not implemented. This document develops the placement
and late-feedback decision from [proposal 2](../../../../../specs/2026-09-09/loop%20improvements/loop%20proposals/2/loop-proposal.md) and resolves
[improvement 4](../../../../../specs/2026-09-09/loop%20improvements/loop%20proposals/2/improvements.md#4-settle-the-meaning-of-late-dispatch-feedback--resolved).
It is the companion algorithm to the consolidated [proposal 3](loop-proposal.md).
Concrete Go types and wire encoding remain implementation work.

## Invariant and ownership

Continuation output cannot advance execution until the runtime has confirmed
successful placement of the handoff that produced it. Receiving output before
that confirmation is permitted; applying its Invoke/Return effects is gated.

The scheduler holds the placement receipt and pending continuation output in the
producing invocation's resident `Continuation.Handoff`. Drivers, callbacks, and
wire envelopes do not become additional authoritative state stores. `transition()` owns state
application, selection, and reservation under its mutex. Dispatch and host calls
remain outside that mutex.

There are two event families:

| Family | Responsibility |
| --- | --- |
| `runtimeEvent` (`dispatchEvent`) | Placement feedback and delayed runtime faults, interpreted by the runtime-event branch. |
| `continuationEvent` (Move) | Invoke/Return transitions, processed through the Horizon check-and-stamp path once placement is confirmed. |

Placement success confirms the handoff. It carries no authorization verdict.
Horizon authorization is required for both Invoke and Return in
`continuationOpcode`. A dispatch callback carrying `runtimeEvent` (`dispatchEvent`)
does not perform this authorization check or carry an authorization verdict.
Source/handoff binding remains required and is separate from authorization.

This boundary follows the event family, not the callback or envelope delivering
it. When runtime feedback releases buffered continuation output, each released
Invoke or Return must pass the Horizon check before its opcode effects. Placement
success and prior Invoke authorization do not exempt a Return from that check.

## Identity and resident state

A parent and its children have distinct InvocationIDs and share a PossibilityID.
Match feedback to the invocation that produced it, not merely to the possibility.

- For a parent emitting child Invokes, the source is the parent's InvocationID.
  Each admitted child receives its own independent InvocationID.
- For a child emitting Return, the source is the child's InvocationID. Its stored
  parent link identifies where the resulting observation belongs.
- Initial dispatch and resumption of the same parent reuse that parent's
  InvocationID. A runtime-bound handoff sequence distinguishes initial `0` from
  resumed `1`; this does not create another invocation or introduce retries.

The feedback association is therefore the producing InvocationID plus its handoff
sequence. The runtime establishes this association when dispatching. A distributed
response echoes the bound association; untrusted continuation output cannot choose
which handoff it satisfies.

Conceptual resident layout:

```text
Continuation                         scheduler-owned resident state
  Move.Invoke.Register               semantic continuation frame
    Contract
    Children                         accepted declarations and observations
    Order

  Handoff                            runtime execution bookkeeping
    Sequence
    Placement                        pending | succeeded | failed
    PendingOutput                    whole batch in declaration order
    OutputConsumed

  Context
  State
  Result
  Resumed
  CompiledRegister
```

These are conceptual fields, not final public Go declarations. `Handoff` is a
sibling of `Move` on the resident continuation; it is not part of `Invoke.Register`.
The register describes declared child work and accepted observations. The handoff
record describes the current execution handoff and output awaiting processing.
Received output belongs to `Handoff`; accepted child declarations and observations
enter the appropriate register only through opcode application.

Both records remain authoritative scheduler state under the same mutex, with no
additional state store, manager, or lock. Initial execution and the one permitted
resumption each initialize a new handoff record with the corresponding sequence.
Feedback must match that sequence before it can affect the current placement gate.

The driver-facing projection is a detached Move containing the semantic Register.
It excludes Handoff bookkeeping, including placement receipts, pending output,
and consumption flags. Invocation ID and handoff sequence are bound through the
runtime callback or transport envelope, separately from the driver-facing frame.

## Algorithm

1. **Reserve the handoff under the lock.** Record the handoff sequence and pending
   placement in the producer's `Continuation.Handoff` before dispatch. Initialize
   its pending-output and consumption state for this handoff. Select a detached
   Move with its semantic Register and retain the saved execution context. This
   gives immediate feedback a resident record to match.

2. **Dispatch after unlocking.** Bind the runtime feedback callback or distributed
   response envelope to the producer and handoff. Preserve the existing resident
   root-state check before each next handoff and suppress work after observed root
   completion or fault.

3. **Receive feedback through `transition()`.** Acquire the lock and resolve the
   bound invocation and handoff. For distributed delivery, wrap associated runtime
   feedback and continuation output together in a response envelope when returning
   both. The wrapper preserves the two event families; it is not a third opcode
   family. Process its runtime-event portion first. If the two portions arrive
   separately, the resident Handoff retains the state needed to join them.

4. **Apply runtime feedback and gate output.** The runtime-event branch interprets
   placement feedback and updates the matching `Continuation.Handoff`:

   | Placement | State and continuation action |
   | --- | --- |
   | Pending | Hold arriving output; do not apply its Invoke/Return effects. |
   | Succeeded | Record success and release pending output for continuation processing. Later output can proceed through the same gate. |
   | Failed | Record failure, discard pending output, and apply the associated workflow fault. |

   No placement confirmation means no continuation progress. Delayed execution or
   transport faults remain runtime events owned by this branch.

5. **Process eligible continuation output under the lock.** Enter
   `continuationOpcode`, check the trusted Horizon against current time, attach the
   runtime-generated check result and timestamp, and branch into Invoke/Return.
   Authorization is required for both opcodes; resolve Return's trusted Horizon
   from its bound resident source. Each handler gates its state changes on that
   stamp. The pending-output buffer
   does not constitute semantic acceptance; this is the authorization point before
   Invoke/Return state application. Preserve whole-batch validation and atomic child
   admission. Accepted declarations and observations update the appropriate semantic
   Register. Record output consumption in `Continuation.Handoff` so duplicate delivery
   cannot apply it twice. Preserve identical-return idempotence and rejection of
   conflicting observations without overwriting accepted results.

6. **Select and reserve affected work under the same lock.** Placement success by
   itself contributes no observation and selects no continuation. Applying output
   released by that success can produce affected continuations and select work.
   Keep one resumption of the existing parent under the single-wave rules.

7. **Unlock before external work.** Dispatch reserved work or deliver the prepared
   host notification. Preserve one notification attempt and propagate processing
   or notification errors without replaying them as execution failures.

## Immediate callbacks and late feedback

In current v14, a driver can call its output callback inside `Ingress`, while the
runtime reports placement after `Ingress` returns. The output callback must be able
to store pending output and return acceptance without waiting for confirmation.
Waiting there would prevent the very acknowledgement required to release it.

With the gate, the initial parent's child declaration cannot advance until its
initial placement success is processed. The parent therefore cannot resume before
that confirmation. On resumption, a new handoff sequence prevents duplicated or
stale initial feedback from satisfying or altering the resumed handoff's placement
gate. Late feedback cannot erase Result, reopen Faulted work, or cause another
resumption. Detailed runtime-fault opcodes remain owned by the runtime-event branch.

## Difference from current v14 and verification

This is a behavioral change: v14 currently applies continuation output before
placement acknowledgement. The new gate also adds a separate continuation-owned
Handoff record and permits a placement-success event to release buffered continuation processing.
It is more than renaming or extracting scheduler helpers.

Verify a later implementation with these sequences:

- Output arrives first: no child admission or Return application until placement
  success, and the immediate callback returns without deadlock.
- Placement success arrives first: later output passes the matching gate once.
- A combined envelope is processed runtime-first regardless of its wire layout.
- Both Invoke and Return pass the Horizon check before opcode effects; Return uses
  the trusted Horizon of its bound resident source.
- A runtimeEvent dispatch callback does not perform Horizon authorization. Any
  continuation output it releases is checked through the continuation-opcode path.
- Placement fails after output is buffered: no continuation effects are applied,
  and the root receives one failure-notification attempt.
- Duplicate feedback/output and stale initial-handoff feedback cannot satisfy the
  resumed handoff or repeat continuation effects.
- Concurrent child completion preserves independent invocation identity, parent
  linkage, declaration order, and one parent resumption.
- Driver-facing Moves expose the semantic Register without resident Handoff
  bookkeeping; buffering output leaves semantic child declarations and observations
  unchanged until opcode acceptance.
- Callback errors, saved contexts, detached snapshots, and atomic batch rejection
  retain their existing guarantees.

Retries, repeated execution waves, persistence, retirement, and compiled matching
remain outside this algorithm.
