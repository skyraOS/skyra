# DNA 15 implementation record

## Completed items

The user requested W0–W4 implementation and subsequently approved the proposed
failure rules. All required local paths are implemented; no policy seams remain
blocked within this work order.

- W0: contracts.go declares resident Handoff/Placement, source sequences on both
  event families, runtime-first Envelope, AuthorizationStamp, boot inputs, kernel
  and separate registry adapters, public CTU, composition signatures and fixtures.
  The approved failure-policy revision removes ErrPolicyBlocked and retains
  Handoff.Rejection as the original diagnostic for a consumed rejected batch.
  Contract revisions preceded dependent bodies.
- W1: canonical-address and nil/typed-nil checks, atomic append-only registration,
  non-embedding resolver/registrar adapters, registration-only kernel initialization,
  separately supplied ordinary root, and public Submit/Snapshot implemented.
- W2: transition owns event application, assessment and reservation under one mutex.
  write delegates application only. Dispatch, validators and notification execute
  outside the mutex. Detached snapshots include pending output; saved execution
  contexts persist through children and root resumption.
- W3: source/sequence binding, placement buffering/release/discard, runtime-first
  envelopes, bounded Invoke/Return authorization and stamps, atomic child batches,
  one parent resumption and the approved failure outcomes implemented.
- W4: real boot/Submit/Ingress/Egress/bound-callback/notification proof passes for
  immediate, reversed and concurrent completion. Distinct output values verify
  declaration order; counters verify two root and two child executions.

## Approved failure rules

Denied or expired resident output has no successful opcode effects. The batch is
consumed, pending output cleared and its original rejection retained. Its producer
and root fault, reserving one notification attempt. The processing error and any
host-notification error propagate through errors.Join. Duplicate delivery cannot
retry effects or notification. External admission rejection only returns an error;
it never faults existing work merely because the input claims its identity.

Runtime faults are matched to the producing invocation and handoff sequence.
Older-handoff feedback is ignored. Current-handoff execution faults fault an active
workflow, including when the producing child already returned. Accepted observations
remain recorded. A late execution fault does not rewrite a successful placement
receipt. Terminal workflows cannot reopen or attempt another notification.

## Verification

From the skyra module directory, all required commands passed:

```sh
go build ./src/skyra/skyra-dna-15/...
go vet ./src/skyra/skyra-dna-15/...
go test -race ./src/skyra/skyra-dna-15/...
```

The starting scoped build passed. Scoped tests passed during composition and after
the placement/boot and final failure-policy changes. Tests use deterministic
coordination without timing sleeps. No unrelated package was repaired or changed.

Coverage includes placement orders, immediate callback acceptance, failed-placement
discard, stale feedback across resumption, whole-batch rejection, denied Invoke and
Return effects, source rebinding, identical/conflicting Returns, context/snapshot
isolation, callbacks outside the scheduler mutex, sibling suppression, notification
error propagation without replay, concurrent registry lookup/installation, typed
nils, failed boot, capability separation, rejection consumption and late faults
before/after source Return and after terminal root completion.

## Explicitly permitted stubs and scope

Fixed local kernel installation and trusted grants remain fixture policy. Signatures
are not verified; admission Horizons are trusted host inputs. Authorization matches
endpoint/action and Unix-second expiry with bounded grant/focus counts. Compiled
matching returns Waiting; observed child completeness alone permits one resumption.
These are not production authority verification or compiled contract satisfaction.

Persistence, retirement, repeated execution waves, retries, production policy,
network transport and dynamic loading remain outside the work order.

## issues.md correction plan

The two reported event-ordering defects were reproduced. The fixes use existing
W0 declarations and signatures; no new resident state or component surface is
needed. Envelope validation compares matching pending output before applying
placement, rejects conflicting envelopes without partial effects, and applies an
identical buffered batch only through runtime release. continuationOpcode checks
the resident source's saved context before either opcode's successful effects.
Permanent regressions cover real execution, conflicts, and root/child cancellation.

Both issues.md corrections are implemented and verified. The permanent
issues_test.go regressions exercise the actual controller/driver callback path for
the duplicate envelope and cancelled root/child Returns, plus conflicting-envelope
atomicity. Fresh scoped race tests, build and vet pass. The earlier completion
assessment missed these ordering cases; they are now part of the required checks.
