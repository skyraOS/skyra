# Builder change order — consolidate scheduling and record ownership

## Make one scheduler dispatch path

Keep state mutation, affected-work selection, and handoff reservation together under
Scheduler's mutex in `Write`. Release the lock before calling CC, drivers, or host callbacks.

Replace the unused `run(move Move)` with `run(selected []Continuation) error`. Have `Write`
pass its selected snapshots to that method after unlocking. Move the dispatch loop currently
inside `Write` into `run`; remove the alternate selection/dispatch implementation.

`selectMoves` remains private and runs under the scheduler lock. `run` must not select or
reserve the same work again. Preserve the existing check that suppresses remaining selected
work after the root has faulted or completed. Any lock needed for that check must be released
before external calls. Preserve each snapshot's saved context and callback error propagation.

The intended responsibility split is:

```text
Write: update + select + reserve under lock
  → unlock
  → run(selected): dispatch snapshots through CC
       → callbacks re-enter Write
```

## Keep execution reporting close to ingress

Retain the implementation's division of outcome reporting:

- `Dispatch` reports lookup or authorization failures through the scheduler callback.
- After successful authorization, `execute` owns the ingress handoff outcome report.
- Output callbacks independently report normalized Moves or source-bound execution failures.

Update the proposal comments to match this ownership. `execute` no longer returns an ingress
error for `Dispatch` to report a second time; its returned error indicates processing/reporting
failure. A callback write error must not be reclassified as a failed driver handoff.

Keep these guarantees explicit: success means authorized and accepted, not completed work;
each handoff outcome is reported at most once; a failure to deliver a callback is surfaced;
and a late acknowledgement cannot regress a returned or faulted invocation. Do not add retries.

## Retain and document the useful additions

Keep `rootLocked` and `faultLocked` on Scheduler, with their caller-holds-lock requirements
documented. Keep the private cloning, dependency-checking, completeness, and fixture-parsing
helpers. Keep the test-only `testDriver` fixture and its existing two function fields.

These existing additions, and the `run` signature change above, are explicitly authorized by
this order. Record them as surface changes rather than claiming the builder added nothing.
No additional runtime structs, fields, or public interfaces are requested.

Move method-specific rationale alongside the corresponding implementations where practical;
avoid leaving contradictory or unattached method comments in `proposal.go`. Update the v14
README/improvement notes where they still describe implemented methods as stubs.

## Verify and report

Preserve the immediate, delayed/reversed, and concurrent echo proofs. Check that the single
dispatch path preserves fault suppression, source binding, and callback-error propagation.
Add or adjust focused tests only where existing tests do not cover the changed behavior.

From the `skyra` module directory:

```sh
go build ./src/skyra/skyra-dna-v14/...
go vet ./src/skyra/skyra-dna-v14/...
go test -race ./src/skyra/skyra-dna-v14/...
```

Keep changes in v14 and preserve unrelated work. Report what changed, the checks run, and any
remaining gap. Compiled interpretation and general retries remain outside this change order.
