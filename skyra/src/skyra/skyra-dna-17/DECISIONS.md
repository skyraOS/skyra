# DNA-17 — decisions, shape, deviations, stubs

Date: 2026-09-19. Status: exploratory, throwaway generation. **All proof here is
against a mocked/fake in-process App Server. There is NO live inference, no real
`codex` binary, no OS containment, no network.**

DNA-17 greenfields the 2026-09-18 "collapse" (one smart Interpreter, one dumb
serializing StateExecutor, a worker pool, one inbound queue, read → commit →
execute, fd-style identifiers minted at commit). It imports **no** `skyra-dna-16`
package; the Codex protocol logic in `codex.go` is a reimplementation cribbed from
DNA-16's `codex/`.

## What was built

- **Core loop** (`cu.go`, `interpret.go`, `state.go`, `dispatch.go`):
  `Interpret` (one door) → interpreter goroutine → `decode` (opcode jump table) →
  `handler` read+decide → `StateExecutor.InterpretTurn` (one atomic turn) →
  `Dispatch` (concurrent workers) → `Surface.Ingress` → emissions committed through
  `stateVia(id)` → re-enter by the door.
- **StateExecutor** is the sole writer behind restricted profiles
  (`PossibilityState.Open`, `SystemState`-style `InterpretTurn`, `ExecutionState`
  from `stateVia`). Components get typed reads (`readView`, `invSnapshot`) that
  return detached copies; no map aliases escape.
- **fd table** (`fds map[Identifier]*fdEntry`): scope/generation/liveness live in
  the writer; the adapter holds only the opaque `Identifier`. Every execution-plane
  request is checked live against it.
- **Codex driver** (`codex.go`) + **Echo child Surface** (`probe.go`).
- **Acceptance tests** (`codex_accept_test.go`): the full dynamic-tool round trip
  and the native-contamination rejection proof, plus an echo-loop sanity test and
  an unauthorized-target trap test. Fake channel-based transport; deterministic
  barriers, no timing sleeps for ordering.

`go build`, `go vet`, and `go test -race` all pass (25 runs, deterministic).

## The loop shape (real signatures)

```go
func (cu *CU) Interpret(move Move)                                  // the one door
func decode(move Move) (func(readView, Move) (Request, error), error)
func interpretInvoke(_ readView, move Move) (Request, error)        // handler.read+decide
func interpretReturn(view readView, move Move) (Request, error)
func (s *StateExecutor) InterpretTurn(ctx, move, decide) (commitResult, error) // read+commit, one lock hold
func (s *StateExecutor) applyLocked(request Request) (commitResult, error)     // the commit
func (cu *CU) Dispatch(items []dispatchItem)                        // concurrent fan-out
func (cu *CU) launch(item dispatchItem)                             // one worker
func (s *StateExecutor) stateVia(id Identifier) ExecutionState      // fd-scoped access
func (e *executionProfile) Commit(ctx, EmissionCommit) (EmissionResult, error) // execution plane
```

`Request = {Move, []Operation}` with a closed operation set
(`opRecordInvocation`, `opAuthorizeStart`, `opSettleReturn`, `opSettleFault`).
Operations are applied by the writer; they are **not** Moves.

## Where the collapsed model held, and where it strained (the finding)

**It held for the request/response spine.** Invoke → mint fd → dispatch → Return
→ settle → root result is clean. `read → commit → execute` with one writer needed
no OCC/revision on the interpret path: `InterpretTurn` takes the writer lock once
and runs decide (read) + applyLocked (commit) inside it, so nothing goes stale.
The fd settlement fence gave the **rejection proof structurally and for free**: a
post-fault agent message hits `stateVia(deadFd).Commit`, is rejected, enqueues no
Move, so no Skyra Return can appear. That is the nicest result of the experiment.

**It strained on the persistent, asynchronous provider session.** The spec's
mental model is "one identifier == one execution == one `Surface.Execute` that
emits Moves." A Codex parent invocation is not that. It is a long-lived thread/turn
that:

1. is *pushed to* several times (initial invoke, then a resume carrying each child
   result), and
2. *emits unprompted*, asynchronously (turn/started, item/tool/call,
   turn/completed) with no Skyra Move driving each emission.

Two concrete deviations followed:

- **fd is minted once per invocation and reused across resume, not "minted on
  resume."** `minimal-loop-v2.md` says the resumed waiting move gets a freshly
  minted identifier at the commit that unblocks it. But the Codex session that must
  answer the tool call is the *same* live session that emitted the child. If resume
  minted a new fd, the async `turn/completed` emission that arrives later would have
  to correlate to *which* fd, and we'd have two live fds for one session. So
  `settleReturnLocked` **reuses the parent's existing live fd** for the resume
  dispatch (`dispatchItem{id: parent.fd, ...}`). The fd is therefore
  fd-per-invocation-session, not fd-per-dispatch. This is the central tension: the
  fd cleanly scopes *state access and emission correlation for a session*, but it
  does **not** subsume the provider session's identity — that still lives keyed by
  `InvocationID` inside the driver's `live` map and its opaque `DriverState`,
  orthogonal to the fd.

- **The worker is not "call Execute, drain Emissions()."** Because emissions are
  async and outlive the `Ingress` call, the worker binds a *persistent* output
  callback (`makeOutput`) to the fd and returns; the driver's own receive goroutine
  invokes that callback many times later. `minimal-loop-v2.md`'s `work()` sketch
  (call `surface.Execute`, then `range surface.Emissions()`) assumes emissions are
  bounded by one Execute call. For a real driver they are not. DNA-16 already made
  this choice (Ingress + persistent callback); DNA-17 keeps it and overlays the fd.

- **Resume is "re-dispatch to the live fd," not "re-enqueue."** Consistent with the
  spec's "resume happens inside interpret." `interpretReturn` → `opSettleReturn`
  resolves the parent in the same turn and emits a resume `dispatchItem`; there is
  no second trip through `decode`.

- **Settlement point is asymmetric.** A `Return` settles the invocation at
  *interpret* (uniform, protocol-driven). A `Failure` emission kills the fd
  immediately at *emission-commit* (the fence needs the fd dead before later frames
  arrive) and then re-enters a `Fault` Move so the interpreter still produces the
  single root result uniformly. Documented because it is a real seam, not an
  accident.

## Other deviations from the specs

- **String IDs, not `uuid.UUID`.** The move/possibility proposals key on
  `uuid.UUID`; DNA-17 uses distinct string newtypes (`InvocationID`, `SurfaceID`,
  …). Deterministic IDs make the tests reproducible with no dependency added to the
  shared `go.mod`. Type distinctness is preserved.
- **The inbound queue is a buffered channel, not state.** The spec says "the
  canonical Move and its unfinished-work record live in state, not in the queue."
  The *records* (invocations, waiting graph, fds) do live in `StateExecutor`; the
  channel is only a transport for pending Moves. Single-process simplification.
- **The writer is a mutex, not a request-channel goroutine loop.** Both serialize.
  The "one turn" guarantee is met by holding the lock across read+commit in
  `InterpretTurn`. No driver/transport/Surface code ever runs under the lock.
- **Operation vocabulary is a small hand-picked set**, not the fully split
  system/execution taxonomy the spec leaves open. `Request.Operations` is honored
  at the interpret→commit seam; the execution plane uses `EmissionCommit` rather
  than an `[]Operation` (a documented shortcut).
- **The ported `MessageBroker/proposal.go` + `tu/proposal.go` two-mailbox,
  Supervisor, custody, from/to world was NOT built** — per the 09-18 README's
  landmine warning. One queue, one Interpreter, one writer, no custody bounce.
- **Focus lives on the `Invoke`** (and is checked ⊆ Horizon at authorize), rather
  than in a separate invocation-state record. Possibility invariant 8 says Focus
  belongs to invocation/execution state; carrying it on the Invoke that opens the
  invocation is the pragmatic realization.

## What is stubbed / minimal (explicitly)

- **Worker pool is unbounded** (goroutine per dispatch, tracked by a WaitGroup).
  No capacity gating, no "capacity-before-take." Fine for the two paths; not a pool.
- **Trap** is minimal: a pre-execution authority failure enqueues a Fault Move
  (`trapLocked`); there is no distinct Trap Surface, no corrective-work vocabulary.
- **Fault propagation to a non-root parent** enqueues a parent Fault Move but is
  lightly exercised (both acceptance paths fault at the root).
- **No independent worker monitoring / timeouts / hung-worker detection.**
- **No Horizon revocation mid-execution, no `ReplaceHorizon`.**
- **No durability / crash recovery / outbox.** In-process only.
- **Codex launch/containment (`launch.go`, process host, version pin, config
  overrides) is NOT ported.** DNA-17 uses only the in-process fake transport. Real
  process containment is unproven by definition here.
- **`revision` exists but is never consulted** (kept for honesty / future
  concurrency).
