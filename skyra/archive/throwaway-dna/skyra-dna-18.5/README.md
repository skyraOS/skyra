# Skyra DNA 18.5

A separate successor to DNA-18, using the existing repository Go module and only
standard-library dependencies. The work order is
[improvements-workorder.md](improvements-workorder.md).

From this directory:

```sh
./demo.sh > artifacts/demo.jsonl
go vet ./... > artifacts/vet.txt 2>&1
go test -race -count=1 -v ./... > artifacts/acceptance.txt 2>&1
```

`demo.sh` builds `artifacts/echo` and `artifacts/root`, then supplies one strict JSON
startup argument with the trusted echo path and two observation strings. It first
runs the inherited two echo calls and parent/child demonstration. A second fresh
bootstrap adds B and the creation capability, rejects B's unauthorized creation
request, completes root→B→root, mints a new executable, and invokes it. Both real
helper processes are reaped. One monotonically numbered trace spans both runs;
the second run intentionally reuses context names with fresh execution state and
fresh authority. Both commands exit nonzero on failure.

```sh
go build -o artifacts/echo ./echo
go build -o artifacts/root ./root
./artifacts/root '{"echo":"./artifacts/echo","observations":["hello","again"]}'
```

The new `crosscontext/` peer translates only Moves between independently bound
ports. `creation/` contains the capability peer and minted echo peer. Trusted
mediation and topology setup live in `root/extension.go`; the execution owners
remain Kernel, RuntimeState, Memory, Active, Scheduler, ExecutionContext and
Bookkeeper. The bound transport remains JSON Moves over pipes.

| Evidence | Test |
|---|---|
| Duplicate/stale queued rejection in live loop; another context completes | `TestLiveLoopRejectionContainment` |
| Explicit cancellation closes claimed send, endpoints, reservations | `TestLiveLoopCancellationClosesClaimedTurn` |
| Fatal transport/context failure remains visible; no send retry | `TestLiveLoopFatalFailures` |
| Correlated A→B→A; virtual zero isolation; B rejection does not stop A | `TestCrossContextAndCreation` |
| New address minted and used; denied template/capability; one-mint bound | `TestCrossContextAndCreation` |
| Ungrantable child delegation rejected before publication | `TestDelegationDeniedWithoutGrant` |
| Frame sink distinct from transport; bad output cannot settle | `TestRootSinkAndOutputContract` |
| Both demos, trace ordering, helper launch/reap, lifecycle text remains data | `TestProcessDemoAndLifecycle` |
| All inherited DNA-18 boundary scenarios | `boundaries_test.go`, `demo_test.go` |

Tests use queued work, channels, connection gates and scheduler idle barriers.
Timeouts are failure guards, never ordering delays. All inherited acceptance
scenarios remain, including forged ownership, stale provenance, parent correlation,
reservation conflicts, cancellation, detached moves and startup rollback.

See [DECISIONS.md](DECISIONS.md) for explicit contract decisions and
[issues.md](issues.md) for the proof's limits. The copied proposal snapshots remain
build-ignored. This experiment does not implement sockets or production recovery.
