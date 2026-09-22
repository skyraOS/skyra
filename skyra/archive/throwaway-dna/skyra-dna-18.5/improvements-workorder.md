# Work order: DNA-18.5 — fold the reviewed improvements into the root experiment

Status: implemented and verified for this bounded experiment. See [README.md](README.md),
[DECISIONS.md](DECISIONS.md), and [captured acceptance results](artifacts/acceptance.txt).
Incremental successor to
[skyra-dna-18](../skyra-dna-18); build lives in this `skyra-dna-18.5/` directory.
Not a full new generation and not a decision on every open contract.
Baseline verified: DNA-18 passes `go vet ./...` and `go test -race -count=1 ./...`;
its demo produced 78 trace events and 3 root results.

## What DNA-18.5 is

DNA-18 proved the root→echo path and parent/child frames **within one context**, on
pipe transport. DNA-18.5 keeps all of that and adds exactly three things:

1. **Loop error containment** — a rejected move must not shut down unrelated contexts
   (a verified bug in `root/runtime.go`).
2. **Cross-context invocation** — the one core claim DNA-18 does *not* establish.
3. **An authority-bounded creation seam** — so the horizon is not a frozen list.

It also **honors four model corrections** from the review (below). Sockets are
explicitly deferred (see Deferred).

## Sources

- Baseline: the DNA-18 sibling dir [`../skyra-dna-18`](../skyra-dna-18) —
  [`root/runtime.go`](../skyra-dna-18/root/runtime.go),
  [`root/context.go`](../skyra-dna-18/root/context.go),
  [`root/ports.go`](../skyra-dna-18/root/ports.go), `root/*_test.go`,
  `echo/`, `parent/`, `protocol/`; and
  [`DECISIONS.md`](../skyra-dna-18/DECISIONS.md),
  [`issues.md`](../skyra-dna-18/issues.md),
  [`workorder.md`](../skyra-dna-18/workorder.md). Copy or extract from the baseline
  into this directory; do not edit the DNA-18 baseline in place.
- Review being folded in: [root improvements](../../../specs/2026-09-21/proposal/root/improvements.md)
  (both the implementation review and the pushback).
- Model: [shared proposal](../../../specs/2026-09-21/proposal/proposal.go),
  [possibility-as-process](../../../specs/2026-09-21/possibility-as-process.md),
  [carry-forward](../../../specs/2026-09-21/carry-forward.md).

Preserve the ownership split from DNA-18's work orders (Kernel / RuntimeState /
Memory / Active / Scheduler / ExecutionContext / Bookkeeper / Executable). Where the
baseline and the shared sketch differ, keep the baseline and note the discrepancy.

## Model corrections to honor (constraints, not tasks)

These come from the review's pushback and are **already accepted**. 18.5 must not
re-bake the overreaches they correct.

1. **Initial launch ≠ never invocable.** Root's bootstrap activation is by the
   launcher (not a Move). Do NOT hardcode "root can never be a `Move.Target`" or
   "never wired to a port." Whether root exposes an executable later is an authority
   decision, not a rule of rootness.
2. **Fresh authority ≠ fresh execution state.** Root authority is installed fresh at
   boot; 18.5 stays always-fresh for scope, but must NOT encode "only below-root
   contexts are restorable" as an architectural rule. Restored state must never carry
   old reservations or bindings.
3. **Authority-bounded creation (the important one).** A context's horizon is NOT a
   frozen subset of a fixed initial address list. A context holding a provisioned
   *creation capability* may mint NEW addresses (executables / child contexts) that
   were never in its initial set, as long as creation and delegation stay within the
   authority it is permitted to exercise and grant. Model creation as **invoking an
   in-horizon creation capability** (mediated, attenuating), not as re-provision-only.
   This supersedes possibility-as-process's "materialize the whole horizon up front;
   anything new = re-provision."
4. **Sink / frame / transport stay distinct.** `RootSink` (return destination),
   the bottom frame (execution tracking), and the bound `Egress` (transport) share one
   return path but are three objects. Do not collapse them. Bookkeeper owns `RootSink`.

## Tasks (core)

### 1. Loop error containment

`Kernel.loop` currently forwards any `Scheduler.Run` error to `errors` and cancels
the whole runtime on the first one; only `admit`-time rejections continue. A stale or
duplicate queued move rejected during `Realize` therefore stops unrelated contexts.

- Distinguish **expected move rejection** (record it, continue scheduling; the claim
  already releases via `defer`) from **fatal runtime failure** (endpoint/context loss,
  explicit cancellation).
- A dispatch rejection must not kill the worker or cancel the loop. Reserve
  whole-loop shutdown for explicit cancellation or failures that invalidate the runtime.
- Do NOT silently discard every dispatch error, and do NOT retry a send after a
  committed transition.
- Tests must run through the live `Kernel.loop`, not only direct `Scheduler.Run`:
  - a duplicate terminal Return is rejected with no second root result and no shutdown;
  - a binding invalidated after admission but before realization cannot settle the old
    frame from its queued output;
  - after a rejection, another context still completes and the claim is released;
  - explicit cancellation still closes endpoints and completes shutdown.

### 2. Cross-context invocation

DNA-18's context activation executable is a detached declaration; the demo creates
root frames directly against local targets. 18.5 must route an invocation to another
context **by that context's executable identity**.

- Context A invokes context B's executable through a bound port. B executes against
  **its own** horizon and returns through **its own** Bookkeeper `RootSink` to A's
  waiting invocation.
- Keep `ExecutableID` distinct from `ContextID`; keep reservations in `Active`; keep
  A and B on independent virtual-port spaces (both may use virtual port 0).
- If B is provisioned from A's authority, B's delegated horizon must contain only
  authority A is permitted to delegate (correction #3).
- Acceptance: a correlated A→B→A round trip; independent virtual-port spaces;
  rejection of an attempt to use authority outside B's horizon; and **failure
  isolation** — a rejected move in B does not stop A (depends on task 1).

### 3. Authority-bounded creation seam

Give some context a provisioned **creation capability** and exercise it.

- Holding the capability, the context requests a new executable or child context and
  receives a usable NEW address that was not in its initial horizon — minted through
  the capability, mediated, attenuating.
- Prove: the created address is usable; the created/delegated authority is a subset of
  what the creator is permitted to grant; and a creation attempt **without** the
  capability is rejected.
- This is what makes 18.5 *embody* correction #3, not just cite it. Keep it minimal —
  one creation capability, one created address — not a general provisioning API.

### 4. Re-align to the declared contracts

DNA-18 kept the core state model and ownership split, but the fixture quietly
changed a few contracts rather than just deferring a layer. Most drifts are
legitimate fixture simplifications documented in `DECISIONS.md`/`issues.md` (dropped
Engram/Backing/OSExecutor/TrapController, `Fault → Message`, one-pending-child). This
task fixes the ones that change model *meaning*, especially the two that are
undocumented:

- **`RootSink` must be a frame, not the transport.** The fixture set
  `Bookkeeper.rootSink *Egress`, collapsing the return *destination* into the
  *transport* — the exact thing model correction #4 forbids. Restore `RootSink` as a
  `FrameID` (the bottom frame / return destination); the bound `Egress` stays a
  separate object on the return *path*. Delivering a root result resolves
  destination → then transports over the Egress; the two are wired together, never
  the same field. (Silent drift; highest priority in this task.)
- **`FrameID` — decide compound vs bare, on the record.** The proposal's
  `FrameID{ContextID, Seq}` makes cross-context frame references *unrepresentable*;
  the fixture reduced it to a bare `uint64` + `Move.Context`, downgrading that to a
  runtime check. Either restore the compound type (structural isolation) or ratify
  the bare form **and** prove the cross-context isolation checks that now stand in for
  it (ties to task 2). Do not leave it as a silent reduction — record the decision in
  `DECISIONS.md`.
- **`Executable` contract shape.** The fixture reduced `ArgContract{[]Observation}`
  to `{Operation string}` and dropped `OutputContract`/`FaultContract`. Either
  restore the observation-based contracts or ratify the simpler shape explicitly;
  `OutputContract` in particular is load-bearing for the activation/return boundary
  (task 2). Record which, and why.

For every remaining documented divergence (Engram, Backing/Emission fence interface,
`VirtualPortSpace`/`ExecutionRef`, `OSExecutor`/`KernelOp`, `TrapController`,
`MoveSource` → out-of-band `queuedMove`, `Move.State`), add a one-line entry to
`DECISIONS.md` stating **"deferred layer"** vs **"contract changed"** so the gap
between the proposal and the build is explicit, not implicit. Re-alignment does not
mean building the deferred layers — it means no contract silently means something
different from the proposal.

## Deferred (NOT in 18.5 core)

- **Sockets / bind-listen-connect-accept** (review improvement #3) → a separate socket
  experiment or DNA-18.6. Keep bound `Ingress`/`Egress` + Moves; do not switch
  transport in 18.5.
- Save/resume, activation envelopes, arbitrary multi-child joins, provider/OS
  containment, fairness/backpressure, crash recovery — remain open per `issues.md`.

## Acceptance scenarios

Deterministic channels/barriers; timeouts are failure guards, not ordering. Extend
DNA-18's set (all existing scenarios must still pass) with:

| Scenario | Required evidence |
|---|---|
| Loop survives rejection | A duplicate completion or stale-binding output is rejected; no second root result; the loop keeps running and another context completes |
| Explicit shutdown | Cancellation still closes endpoints, releases claims, reaps helpers |
| Cross-context round trip | A invokes B by executable id; exactly one correlated result returns to A through B's own RootSink |
| Context isolation | A and B use independent virtual spaces; an out-of-horizon authority use in B is rejected; B's rejection does not stop A |
| Authorized creation | With the capability, a new address is minted and used; delegated authority ⊆ grantable; creation without the capability is rejected |
| RootSink distinct from transport | The return destination is a frame; the `Egress` is a separate object; a root result resolves destination then transports — neither field is the other |
| Contract decisions recorded | `FrameID` (compound vs bare) and the `Executable`/`ArgContract` shape each have a stated decision in `DECISIONS.md`; no contract silently differs from the proposal |
| Regression | Every DNA-18 acceptance scenario still passes |

## Temporary decisions, not new architecture

- No root-context field on `RuntimeState`; root stays an ordinary `Active` entry
  distinguished only by its bootstrap path.
- No `Kernel.Run`, `RootProgram.run/accept/step`, or scheduler copies of Active's maps.
- Keep `FaultType` open; a fixture error representation does not settle it.
- Record any seam that cannot be exercised without a new decision in `DECISIONS.md`
  rather than claiming the experiment proves it.

## Completion artifacts

- Updated runnable root/echo/parent programs plus the cross-context and creation peers.
- Tests for the new acceptance scenarios; a captured demo trace showing loop survival,
  a cross-context round trip, and an authorized creation.
- `DECISIONS.md` and `issues.md` updated with what 18.5 settled and what it left open.
- `go vet ./...` and `go test -race -count=1 ./...` pass in the experiment module;
  report results honestly.

Success: a rejected move no longer stops the runtime, one invocation crosses from one
execution context into another and returns correctly, and a context grows its own
address set within — and only within — the authority it was granted.
