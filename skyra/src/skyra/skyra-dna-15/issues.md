# DNA 15 inspection findings

Both findings below have been reproduced and fixed. Permanent regression tests
are in [issues_test.go](issues_test.go). The original trigger descriptions are
retained below as the defect record. Scoped build, vet, and fresh race tests pass:

```sh
go build ./src/skyra/skyra-dna-15/...
go vet ./src/skyra/skyra-dna-15/...
go test -race -count=1 ./src/skyra/skyra-dna-15/...
```

## 1. Buffered child output repeated in an envelope can strand admitted children

Relevant code: [envelope processing](transition.go), specifically `write`,
`runtime`, `continuation`, and the selection condition in `transition`.

### Trigger sequence

1. Dispatch a parent and leave placement pending.
2. Receive a valid child Invoke batch through the bound callback. It is buffered
   in the parent's Handoff without child admission.
3. Receive an Envelope containing matching placement success and the same child
   batch, with the same producer identity and handoff sequence.
4. The envelope's runtime portion releases the buffered batch, installs the child
   continuations and parent entries, and marks the output consumed. It returns
   those children as affected work.
5. `write` then processes the envelope's continuation portion. `continuation`
   rejects the repeated Invoke batch with "handoff output already consumed".
6. `transition` sees the processing error and skips selection/reservation, despite
   the children already having been installed by the runtime portion.

### Consequence

The parent has admitted children that have not been dispatched. Repeated placement
success does not release them again because placement is already recorded. The
workflow can remain waiting without progressing or reporting a root failure.

### Required behavior

An identical repeated batch must not prevent work admitted by placement release
from being selected. Process the handoff's output once while preserving the work
produced by that processing. Do not globally ignore processing errors or treat
conflicting output as an identical duplicate to address this case.

### Regression proof

Add a focused scheduler test for output-first followed by a matching combined
placement/output envelope carrying the identical child batch. Assert one child
admission per ID, one initial dispatch per child, and eventual parent resumption
and root completion through the normal execution path. Existing placement-order
tests cover combined delivery without an already-buffered child batch.

## 2. Buffered Return can be accepted after its saved context is cancelled

Relevant code: [output context check](controller.go) in `execute`,
[placement release](transition.go) in `runtime`, and
[Return application](scheduler.go) in `continuationOpcode`.

### Trigger sequence

1. Submit an invocation with a cancellable context and valid authorization.
2. Its driver emits a valid Return while that context is still active. Output
   validation succeeds and the Return is buffered because placement is pending.
3. Cancel the saved execution context before placement is confirmed.
4. Deliver successful placement feedback for the matching handoff.
5. Runtime processing releases the buffered Return using the saved context.
6. `continuationOpcode` authorizes and applies the Return without checking that
   context. Its context checks occur after the Return branch, on Invoke paths.

A local driver can exercise this ordering by emitting output, cancelling the
context, and then returning a successful ingress acknowledgement.

### Consequence

The cancelled invocation can acquire a successful Result. A root Return can notify
the host of success after cancellation was already observable before the buffered
output's semantic application. A child Return can install an observation despite
that same cancellation window.

### Required behavior

Check the saved execution context when buffered Return output becomes eligible
for application, before recording its successful effects. Use the established
cancellation/failure handling and notification rules. This is a context check at
continuation processing, not Horizon authorization on runtime placement feedback.
It does not require an autonomous cancellation watcher while a driver is silent.

### Regression proof

Add a deterministic real-driver test: emit a Return, confirm it is buffered,
cancel the context, then complete the handoff successfully. Assert that no
successful Result or root-success notification is produced and that cancellation
is handled through the normal failure path. Cover a child Return as well so its
parent register cannot acquire the cancelled observation.

## Review scope

The principal ownership boundaries appear aligned with the proposals: resident
Handoff state is separate from semantic Register inputs; scheduler transition owns
progression; both Invoke and Return pass authorization; runtime feedback does not
perform authorization; registration and resolution use separate capabilities.
These findings concern the two specific event-ordering paths above, not a request
to redesign those boundaries or expand the runtime's scope.

## Resolution

1. Envelope validation detects matching resident pending output before any effects.
   A conflicting continuation portion rejects the envelope without releasing or
   admitting the buffered batch. For identical output, runtime placement release
   applies the batch once and its affected children proceed to normal selection.
   The repeated continuation portion is not reapplied. No errors are globally ignored.
   `TestBufferedChildrenRepeatedInEnvelopeCompleteLoop` runs the real driver and
   callback path through two child executions, one root resumption and completion.
   `TestConflictingBufferedEnvelopeHasNoPartialEffects` verifies rejection and that
   subsequent valid placement still dispatches the original child.

2. continuationOpcode resolves the resident saved context and checks it before
   Invoke or Return effects, including on placement release. Cancellation uses the
   existing consumed-rejection/fault/notification path.
   `TestCancelledBufferedReturnFaultsBeforeEffects` covers real root and child
   drivers, cancellation after buffering, no successful results/observations, one
   notification attempt, sibling suppression, and propagation of cancellation plus
   host-notification errors.

No new state, types or method signatures were required.
