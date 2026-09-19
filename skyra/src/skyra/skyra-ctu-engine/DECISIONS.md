# Overnight probe — 2026-09-06

> Historical probe report. The subsequent accepted [fd-entry handoff](surface-proposal.md#accepted-revision-kernel-hands-executable-fd-entries-to-ctu)
> gives Move and Register a shared storage home and changes Fire to accept entries. Installation
> and execution remain unimplemented; the findings below describe the original frozen surface.

The probe is complete; the runtime is not. The frozen surface cannot preload executable Moves,
validate admission, or deliver CTU events. No graph produces a root observation. Independent
publication and lifecycle behavior is implemented and tested without changing any type or signature.

## Choices and assumptions

- Serial use only, matching the small local scope. No locks, goroutines, new structures, side tables,
  or payload encodings for missing runtime state were added.
- `Memory.Land` fills exactly one declared child entry in the existing parent frame. Missing/nil
  frames, unknown children, and repeated child addresses return `Failure` without mutation. It does
  not allocate a missing frame, append children, or infer a materialization fault.
- First publication copies content bytes and retains Segment. Empty observations still arrive.
  An identical repeat succeeds without replacement; different bytes or Segment fail. Nil and empty
  byte slices compare equally. These are local probe policies, not settled distributed retry rules.
  Consumers must treat published values as read-only; public pointers cannot enforce that.
- Publication is storage, not validation or readiness. Land cannot verify the returning child's
  contract from its arguments. Ret remains a stub instead of silently accepting unknown schemas.
- `Claim` takes only Ready entries, records Claimed before returning, preserves request order, and
  deduplicates through the existing state map. Its returned slice does not alias the request.
- `Apply` accepts trusted Issued/Landed events only for Claimed or Completed entries. Completion is
  retained and duplicates/late Issued cannot regress it. Invalid transitions panic without mutation.
  Landed assumes prior publication as its interface requires; Scheduler cannot independently verify
  it or compare duplicate values, and it does not admit dependents. Faulted remains explicitly blocked.
- Unit tests seed Ready states to isolate transitions. No production path does so; Admit remains
  blocked because accepting arbitrary fds would bypass its required readiness validation.
- MoveID strings in tests are opaque address fixtures, not a canonical identity implementation.

## Evidence and reproduction

From the `skyra` module directory, all required checks passed:

```sh
go build ./src/skyra/skyra-ctu-engine/...
go vet ./src/skyra/skyra-ctu-engine/...
go test ./src/skyra/skyra-ctu-engine/...
```

`engine_test.go` retains the boot test. `probe_test.go` checks reverse-order child arrivals,
producer-buffer isolation, repeat/conflict handling, absent/ambiguous routing, claims, retained
completion, and rejected lifecycle events. `TestLocalLoopBlockedAtAdmissionResolutionAndDelivery`
boots a resident root frame and demonstrates explicit failures at Admit, Resolve, Fire, and Ret.
It is boundary evidence, not an end-to-end success. Direct Land calls test storage only.

The smallest proposed execution example is a parent comparing two child documents whose returns
arrive in reverse order. Its storage portion runs in tests. Executable preloading, invoking the
children, constructing/validating the comparison, and delivering a root result remain blocked.

Load, authority minting/verification, recovery handlers, Trap, Call, Exec, retirement, and REMEMBER
remain explicitly unimplemented. Driver dispatch was not partially enabled: its Outcome has no
lossless general mapping into the current event vocabulary. See the boundary report for proposed
surface changes; none were applied. All edits are confined to this engine directory.
