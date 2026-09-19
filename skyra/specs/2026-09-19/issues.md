# DNA-17 review issues

Date: 2026-09-19  
Status: Open  
Reviewed revision: `264ef78` (`dna-17: move ALL reads into interpret; add ownership-boundary display`)  
Scope: `skyra/src/skyra/skyra-dna-17`

The reviewed version compiles. Its four existing acceptance tests pass with the
race detector, and `go vet` passes. The behavioral reproductions below were run
against an unmodified copy of the package in a temporary directory, with additional
review tests. Those tests assert the observed bugs; their passing does not mean
the bugs are fixed. No live provider or OS containment was tested.

All seven issues remain open. Fix invocation identity, fault scoping, and queue
deadlock first.

## 1. High — Invocation IDs can overwrite existing executions

**Location:** [state.go](../../src/skyra/skyra-dna-17/state.go),
`applyLocked` / `opRecordInvocation` and `mintInvIDLocked`;
[interpret.go](../../src/skyra/skyra-dna-17/interpret.go), `interpretInvoke`.

Invocation records are inserted without checking for an existing ID. Generated
child IDs (`inv-1`, `inv-2`, etc.) share the namespace of caller-supplied IDs and
do not check for collisions. Repeated submissions with the same invocation ID
can also overwrite records across possibilities.

**Reproduction:** Start a root named `inv-1`, then commit a child Invoke from its
fd. The first generated child is also `inv-1`. Interpreting that child overwrites
the parent record, makes its `ReturnTo` point to itself, and leaves the original
fd live.

**Impact:** Continuations and results can be miscorrelated. An old fd can address
a replacement record, undermining execution isolation.

**Required behavior:** Reject conflicting invocation identities before writing
records; define duplicate/replay handling; ensure generated IDs cannot collide
with submitted or pending IDs. Test parent/child collisions, duplicate Invokes,
and reuse across possibilities.

## 2. High — An emitted Fault can target another invocation

**Location:** [state.go](../../src/skyra/skyra-dna-17/state.go),
`executionProfile.Commit`; [interpret.go](../../src/skyra/skyra-dna-17/interpret.go),
`interpretFault`.

Emission commits bind Invoke and Return identities to the emitting fd, but do not
validate or replace `Fault.Invocation`. The fault interpreter trusts that field
without checking the emitting invocation or possibility.

**Reproduction:** Create two live invocations. Commit a Move containing
`Fault{Invocation: "victim"}` through the other invocation's fd. Interpreting the
committed Move faults the victim.

**Impact:** A Surface's normalized output can terminate unrelated work despite
having only its own execution authority. The bundled normalizers do not currently
emit Fault Moves, but the execution boundary accepts them.

**Required behavior:** Reject execution-emitted Fault Moves or bind them to the
emitting invocation. Keep interpreter-generated parent fault propagation explicit
and distinguishable from execution-originated claims. Test foreign-invocation and
foreign-possibility targets.

## 3. High — The interpreter can deadlock on its own inbound queue

**Location:** [cu.go](../../src/skyra/skyra-dna-17/cu.go), `step` and `Interpret`.

The interpreter synchronously enqueues corrective/propagated Moves into the same
bounded channel it alone drains. Another producer can fill the slot released by
the interpreter before it enqueues a corrective fault.

**Reproduction:** Fill the inbound queue, let the interpreter take an unauthorized
Invoke, and refill the released slot before its commit completes. The interpreter
then blocks sending the resulting Fault into the full queue.

**Impact:** All interpretation stops until shutdown cancels the blocked send.
Increasing channel capacity only postpones the condition.

**Required behavior:** Internal follow-up work must not require the sole consumer
to block on its own full queue. Preserve ordering and backpressure with a suitable
pending-work mechanism. Add a saturation test covering corrective faults and
parent fault propagation.

## 4. Medium — Turn notification ordering can stop the provider driver

**Location:** [codex.go](../../src/skyra/skyra-dna-17/codex.go), `updateTurn`,
`handleResponse`, and `receive`.

`turn/started` changes the invocation phase to `running`. If that notification
arrives before the `turn/start` response, response handling no longer recognizes
the outstanding request because it selects response semantics using the current
invocation phase.

**Reproduction:** With the fake transport, deliver `turn/started` before the
matching `turn/start` response. The response produces
`unexpected Codex response in phase "running"`; the receive loop invokes
`failAll` and exits.

**Impact:** This ordering stops the shared driver and faults other active
invocations. Existing acceptance tests cover only response-first ordering.
The reproduction establishes local behavior; live-server ordering guarantees
were not verified in this review.

**Required behavior:** Correlate responses with the outstanding request's method
and expected response type, and reconcile matching lifecycle notifications without
regressing state. Test both event orders, or establish and enforce a documented
transport ordering constraint.

## 5. Medium — Child returns do not validate the parent's dependency

**Location:** [interpret.go](../../src/skyra/skyra-dna-17/interpret.go),
`interpretReturn`; [state.go](../../src/skyra/skyra-dna-17/state.go),
`executionProfile.Commit` and `opResumeParent`.

Every child return resumes a live parent without checking that the parent is
waiting for that child. `waitingOn` holds one ID, but emission commits accept
multiple child Invokes and overwrite that ID.

**Reproduction:** Commit two child Invokes from one parent. The parent's
`waitingOn` names the second child. Return the first child: the parent resumes and
its wait for the second child is cleared.

**Impact:** The parent can resume prematurely or receive concurrent resumes with
results that do not match its pending continuation.

**Required behavior:** For the current single-child model, reject overlapping
child invocations and require the returning child to match the waiting dependency.
If multiple children are supported later, track and satisfy those dependencies
explicitly. Test both batched and successive overlapping child emissions.

## 6. Medium — Late faults overwrite successful terminal state

**Location:** [interpret.go](../../src/skyra/skyra-dna-17/interpret.go),
`interpretFault`.

Unlike `interpretReturn`, fault interpretation does not guard against an invocation
that already returned. It unconditionally schedules `opMarkFaulted`.

**Reproduction:** Interpret a root Return, then a Fault for the same invocation.
The canonical phase changes from `returned` to `faulted`. At the CU boundary,
`rootOnce` suppresses the second root notification but does not prevent the state
change.

**Impact:** Canonical state contradicts the successful result already delivered.
A late fault for a returned child can also propagate into a still-live parent.

**Required behavior:** Preserve terminal outcomes and make fault handling
idempotent. Account for the existing failure path, which premarks an invocation
faulted before enqueueing its first Fault; simply dropping every already-faulted
record would lose legitimate fault delivery. Test late faults after success and
duplicate faults after failure.

## 7. Medium — Submission can succeed when execution is impossible

**Location:** [state.go](../../src/skyra/skyra-dna-17/state.go), `Open`;
[cu.go](../../src/skyra/skyra-dna-17/cu.go), `Submit`, `Interpret`, and `step`.

`Open` ignores its context and does not check runtime shutdown. Submission also
does not validate the invocation ID or action before recording the possibility;
those errors arise later and `step` silently discards them.

**Reproduction:** Close the CU, then submit a new possibility. `Submit` returns
`nil` even though the interpreter has exited. Separately, an empty invocation ID
or action passes submission validation but fails subsequent interpretation.

**Impact:** Callers receive successful acceptance for work that will never produce
a root result. Possibility records can remain without an executable invocation.

**Required behavior:** Reject submissions after shutdown and honor canceled
submission contexts. Validate the first Invoke before accepting the possibility,
and define how asynchronous interpretation failures reach callers. Test closed,
canceled, and malformed submissions, including queue backpressure.
