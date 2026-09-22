# Skyra DNA 18

An isolated root-boundary experiment using the existing repository Go module and
standard library. The echo is a separate persistent process; the deterministic
parent is a serialized-Move test peer. Proposal snapshots remain build-ignored.

Run these commands from `skyra/src/skyra/skyra-dna-18`:

```sh
# Build and run boot, two echo invocations, and one parent/child round trip.
./demo.sh

# Run all isolation/rejection scenarios, integration tests, and race checks.
go test -race -v ./...

# Required static check.
go vet ./...

# Formatting.
gofmt -w root protocol echo parent
```

Both demonstration commands exit nonzero on failure. `demo.sh` builds these two
binaries; the equivalent explicit build and invocation commands are:

```sh
go build -o artifacts/echo ./echo
go build -o artifacts/root ./root
./artifacts/root '{"echo":"./artifacts/echo","observations":["hello","again"]}'
```

Startup is exactly one JSON argument with `echo` (trusted host executable path)
and `observations` (exactly two strings). Unknown fields, trailing JSON, and malformed
input fail before context publication or process launch. Inside the process the
strings become `operation` and `text` observations. JSON encodes observation byte
arrays as base64. The startup path is host plumbing, never an executable identity.

To reproduce the captured evidence:

```sh
./demo.sh > artifacts/demo.jsonl
go test -race -count=1 -v ./... > artifacts/acceptance.txt
```

The trace numbers actual events under a shared trace lock. Owner, context, Move,
and frame fields show admission, scheduling, commits, sends, and teardown. Event
order between goroutines can vary: a receiver may log reception before the sender
logs its successful Send. Tests assert causal constraints, including publication
before invocation and commit before result. All three root results are verified;
the last operation returns `parent:child` after one recorded child fork.

`root/contracts.go` declares owner state; implementation is in separate files.
`protocol/` owns the wire vocabulary. `root/ports.go` contains bound endpoint and
bootstrap transport plumbing. Actual port reservations never enter the wire.
`echo/` has the real helper main; `parent/` owns the deterministic parent peer.

| Acceptance evidence | Test |
|---|---|
| Startup, real process, one launch/two calls, causal trace, parent resume | `TestProcessDemoAndLifecycle` |
| Atomic reservation conflict | `TestReservationConflictAtomic` |
| Forged owner and two contexts sharing virtual port 0 | `TestForgedOwnershipAndVirtualPortIsolation` |
| Fixed endpoint destination; malformed contract never reaches helper | `TestFixedDestinationAndContractRejection` |
| Same-context serialization and other-context progress | `TestSameContextSerializationAndDifferentContextProgress` |
| Missing context and dispatch failure release claims | `TestDispatchFailureReleasesClaim` |
| Stale admission and already-queued output; duplicate completion | `TestStaleOutputAndDuplicateCompletion` |
| One pending child; forged correlation; retained continuation; one resume | `TestOnePendingChildAndCorrelation` |
| `save` and `close` child observations remain data; helper reaped | `TestProcessDemoAndLifecycle` |
| Blocked receive, idle scheduler, and claimed turn cancellation | `TestCancellationAndShutdown` |
| Invalid startup has no publication or launch effects | `TestMalformedStartup` |

Boundary tests use channels, gated connections, and completion signals. Deadlines
only detect failure. Stale-output tests deliberately retain decoded output across
rebinding to exercise the receive/admission and queue/realization races.

See [DECISIONS.md](DECISIONS.md) for temporary contracts and proposal discrepancies,
[issues.md](issues.md) for unproved boundaries, and [workorder.md](workorder.md) for
the original acceptance requirements. This proves logical routing and local state
separation; host bootstrap substitutes for the independent OS/preparation programs.
