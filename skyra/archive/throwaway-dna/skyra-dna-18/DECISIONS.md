# DNA18 bounded proof decisions

Declarations were added before implementation in `root/contracts.go`,
`protocol/contracts.go`, and `parent/contracts.go`.

- Transport: newline-delimited JSON Moves on bound duplex connections. OS pipes
  and handles stay inside the endpoint/launcher implementation. Cancellation closes
  an endpoint permanently; reconnection creates a new endpoint.
- Startup: exactly one strict JSON argument: `echo` is a trusted host executable
  path and `observations` is exactly two strings. Configuration grants no authority.
- Identity: numeric executable IDs 1 (echo), 2 (parent), 100 (context activation
  descriptor); frame sequences are local
  to the context, never reused. The activation executable descriptor belongs to the
  context; its Horizon stays private. Actual ports are host-assigned reservations.
  `Activation()` returns a detached descriptor. The external endpoint activates a
  root frame by naming an authorized target in the local Horizon; routing another
  context's activation by descriptor ID is outside this proof.
- Arguments: two observations, `operation` with bytes `echo` or `parent`, followed
  by `text`. These select accepted arguments, not another addressed operation.
- Admission: bootstrap binds an external request endpoint to one owner. It may
  submit only root Invokes with fresh nonzero request IDs. Executable output must
  correlate to a running frame on its bound endpoint. Conflicting context claims
  are rejected; absent claims are stamped from trusted provenance.
- Binding fence: a fresh private binding object plus monotonically increasing
  generation represents each replacement. Both admission and realization check
  liveness and current mapping. Queue entries pair a detached wire Move with trusted
  provenance; this is an explicit host-only envelope absent from the sketches.
- Continuations: one pending child per parent, with runtime-minted frame identity.
  The peer supplies opaque continuation bytes when yielding; the Bookkeeper retains
  them and resumes the parent with a correlated Return and those bytes.
- Faults: Go errors reject invalid admission/dispatch. The fixture wire Fault holds
  a diagnostic string. This does not settle the proposal's open FaultType taxonomy.
- Commit occurs before bound Send. Transport failure is reported, with no retries
  or delivery guarantee; retained committed state supports diagnostics only.
- Scheduler workers receive only a dispatch function. Kernel.loop provisionally
  coordinates ingress and two Run workers. No scheduler provisioning occurs.
- Host bootstrap launches the echo helper and connects the parent test peer. This
  intentionally substitutes for the proposal's OS/preparation invocations, as the
  work order permits. There is no independent OS or preparation program.
- Root proposal admission through Kernel takes precedence over shared-sketch
  comments assigning enqueue to Egress. Executables use the Move wire, not Backing
  interfaces. Horizon and frame state are private, unlike exported sketch fields.

## Verified evidence

The implementation and final checks completed on 2026-09-21:

- `gofmt -w root protocol echo parent`; subsequent `gofmt -l` produced no output.
- `go vet ./...` passed from this generation's directory.
- `go test -race -count=1 -v ./...` passed; the actual output is retained in
  [artifacts/acceptance.txt](artifacts/acceptance.txt). It covers all work-order
  acceptance scenarios, plus malformed startup and failed-launch rollback.
- `./demo.sh` built both binaries and exited successfully. Its 78-event trace is
  [artifacts/demo.jsonl](artifacts/demo.jsonl): one launch, two direct echo results,
  one parent/child round trip, three root results, binding invalidation, reservation
  release, and helper reap. Events are recorded by the live owners, not replayed.
- Integration tests additionally verify helper `Wait` completion and normal child
  observations containing `save` and `close`. Concurrency tests hold an actual
  Realize send while another context advances and use explicit cancellation barriers.

Initial sandboxed Go checks encountered cache-access restrictions. The final vet
and race-test runs used permitted Go-cache access and exited zero. No stress repeats
were needed after the final successful run. No module configuration, historical
DNA implementation, or original dated specification was changed.
