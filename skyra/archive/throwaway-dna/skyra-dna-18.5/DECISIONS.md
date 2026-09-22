# DNA-18.5 decisions

DNA-18.5 copies DNA-18 into this directory and preserves its owner split and pipe
transport. The dated proposals and DNA-18 baseline are unchanged. This is a bounded
experiment using the existing repository module, not a production runtime.

## Contracts decided for this build

- **RootSink restored:** `Bookkeeper.rootSink` is a `FrameID`. Local frame zero
  identifies the separately stored `Bookkeeper.bottom` frame; invocation sequences
  begin at one. Bottom is outside the invocation-frame map. Root settlement checks
  and resolves that destination before committing and sending through the separate
  `ExecutionContext.rootEgress`. The trace records `root sink resolved` before
  `transition committed` and `root result sent`. B has its own sink and transport.
- **FrameID ratified as bare:** this build retains `uint64`, qualified by the
  authenticated owning context rather than restoring `{ContextID, Seq}`. This is
  a **contract changed** from structural identity to runtime isolation. Admission
  checks claimed context, source binding, frame binding and correlation; realization
  repeats them. Foreign context claims, wrong binding generations and duplicate
  terminal returns cannot select another context's frame. The activation bridge
  retains A's correlation, allocates B-local request identities, and explicitly
  translates B's result back to A's frame. It never imports A's frame into B.
  Existing forged-owner/virtual-space tests and the new cross-context test exercise
  this choice, including overlapping local frame numbers.
- **Executable/ArgContract ratified:** `Operation string` remains the argument
  profile. Invoke still carries observations: exactly `operation` and `text`.
  This is a **contract changed** from observation-based profile declarations;
  arbitrary observation schema matching is deferred. `OutputContract{Key}` is
  added and enforced before settlement; an empty key means `text`, while creation
  returns `executable`. Activation accepts echo arguments and text results.
  `FaultContract{}` explicitly leaves diagnostic faults unrestricted; this is a
  **contract changed**, not a settled FaultType or declared-outcome taxonomy.
- **Error containment:** pre-commit Move validation failures are typed rejections,
  traced with context/Move/frame and consumed by the worker after claim release.
  A well-formed JSON Move with invalid variant/identity is also recoverable.
  Unknown errors, missing active contexts, broken sink wiring, unexpected endpoint
  loss, undecodable streams and failed committed sends remain fatal to this small
  runtime. Intentional supersession closes the old ingress without failing the
  loop. There is no retry after commit. Cancellation retires Active entries,
  invalidates endpoints and releases reservations; launcher shutdown reaps peers.
- **Cross-context activation:** B is executable 101, distinct from context `B`.
  A owns a bound endpoint targeting 101; a serial pipe peer translates its Invoke
  into a B-local echo activation. B schedules and resolves echo in its own horizon,
  settles through its own Bookkeeper sink and returns to A's waiting frame. A and B
  both use virtual zero, with different actual reservations. This is one fixed
  activation profile, not a general envelope or concurrent activation multiplexer.
- **Delegation and creation:** the launcher's immutable grant policy permits only
  independent echo behavior. B gets echo and plumbing endpoints, not A's parent,
  activation or creation authority. Capability 200 is installed in A's horizon;
  invoking `create` with text `1` asks for an echo executable. The trusted mediator
  verifies live capability and grant authority, then repeats those checks under
  the publication lock. Memory chooses a fresh unused ID starting at 1000; Active
  owns the new actual reservation, and the creator's local mapping grows. The
  one minted executable accepts only echo and cannot delegate or create. No caller
  chooses its context, reservation or virtual port. This deliberately implements
  one capability, one successful mint and one fixed port slot, not a general API.
  Denied templates return a correlated diagnostic Fault; a context without the
  capability is rejected before reaching its peer. A failed send is never retried.
- **Bootstrap corrections:** root authority is installed fresh; no root field,
  never-invocable rule or below-root-only restoration rule is introduced. Whether
  any context exposes its executable is a provisioning decision. Persistence is
  unimplemented; restored state must not restore old bindings or reservations.
  New addresses can be minted through a provisioned capability: the initial
  address set is not the lifetime authority boundary.

## Remaining proposal differences, explicitly classified

- Engram: **deferred layer** — direct frame commits stand in for compilation,
  instruction serialization and memoization.
- Backing/Emission fence interface: **deferred layer** — bound Move peers and
  repeated binding/frame checks implement the bounded fence without the Go interfaces.
- VirtualPortSpace/ExecutionRef: **deferred layer** — local mappings and private,
  never-reused binding objects carry owner, generation and liveness.
- OSExecutor/KernelOp: **deferred layer** — trusted launcher/creation plumbing uses
  pipes and starts/reaps helpers; the general addressable OS service is absent.
- TrapController: **deferred layer** — denied authority is diagnosed; general
  corrective preparation and resumption are absent.
- MoveSource → host-only queuedMove: **contract changed** — authenticated binding
  provenance travels outside executable-controlled JSON, beside a detached Move.
- Move.State: **contract changed** — opaque serialized continuation bytes travel
  directly on child Invoke and parent resume instead of a separate Emission.
- Fault → diagnostic Message: **contract changed** — string diagnostics do not
  settle FaultType, failure recovery or corrective control transfer.
- One-pending-child: **contract changed** — a second outstanding child is rejected;
  arbitrary multi-child dependency joins remain open.
- Enqueue ownership: **contract changed** relative to the shared sketch's Egress
  comment — Kernel admits/enqueues, following the root proposal and DNA-18.
- Public sketch state: **contract changed** — owner maps, horizons and frames stay
  private; Activation returns a detached declaration.

## Evidence

See `artifacts/acceptance.txt`, `artifacts/vet.txt` and `artifacts/demo.jsonl` for
captured verification. README maps acceptance scenarios to tests. Demo events come
from live owners; scheduling order between goroutines is not fixed.

Final verification on 2026-09-21: `go vet ./...` exited zero (empty vet log),
`go test -race -count=1 -v ./...` exited zero, and `./demo.sh` exited zero.
The captured demo contains 177 monotonically numbered events, seven sink deliveries
(three baseline results, three extension results in A/root, one result in B), and
two helper reaps. The final race run required approved Go-cache access after a
sandbox cache read was denied. No test failure was suppressed.
