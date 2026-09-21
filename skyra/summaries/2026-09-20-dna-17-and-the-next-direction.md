# Skyra: DNA-17 and the next direction

**Date:** 2026-09-21 (updated; originally 2026-09-20)  
**Reviewed build:** `skyra/src/skyra/skyra-dna-17` at `264ef78`.  
**Status:** Conceptual and implementation handoff. DNA-17 is the executable
reference — an *exploratory, throwaway* generation of the 2026-09-18 "collapse,"
proven only against a fake in-process App Server (no live inference, no real
`codex` binary, no OS containment, no network). The direction then advanced in
two moves: `specs/2026-09-19` named the runtime→OS **execution handoff** and a
revised Surface/Possibility model; `specs/2026-09-20` and `specs/2026-09-21` then
converged everything onto one model — **mediated actors over a privileged kernel**,
with **ports** as identity and the **Possibility as a process**. All of it is
proposal-stage and explicitly "rejectable." **Part III is the current direction;
Part II is the intermediate layer it builds on.** The single precedence index is
[`specs/2026-09-21/carry-forward.md`](../specs/2026-09-21/carry-forward.md) — read
it to know which statement to build against.

## Executive summary

DNA-17 proved the *collapsed* runtime loop across one writer and a worker pool:

```text
Interpret (one door)
  → decode (opcode jump table)
  → handler: read + decide (authority, correlation, resume)  [interpret]
  → StateExecutor.InterpretTurn: one atomic locked turn       [commit]
  → Dispatch: concurrent workers, one per fd                  [execute]
  → Surface.Ingress + persistent output callback
  → emissions committed through stateVia(fd)                  [execution plane]
  → re-enter by the one door
```

**What held:** the request/response spine. `Invoke → mint fd → dispatch → Return`
`→ settle → root result` is clean. Because one writer holds the lock across
read+commit (`InterpretTurn`), the interpret path needs **no OCC/revision** —
nothing goes stale. The **fd settlement fence** gave the post-fault rejection
proof *structurally and for free*: a late agent message hits `stateVia(deadFd)`,
is rejected, enqueues no Move, so no Skyra Return can appear.

**What strained:** the persistent, asynchronous provider session. The collapse's
mental model is "one identifier == one execution == one `Execute` that emits
Moves." A Codex parent invocation is a long-lived thread/turn that is *pushed to*
several times and *emits unprompted* (turn/started, item/tool/call,
turn/completed) with no Skyra Move driving each emission. That forced three
deviations (fd reused per session not per dispatch; a persistent callback, not
`range Emissions()`; resume = re-dispatch to the live fd) and an asymmetric
settlement point (Return settles at interpret; Failure kills the fd at
emission-commit, then re-enters a Fault).

**Where the specs went next.** `specs/2026-09-19` took that finding and pushed on
two fronts — the runtime→OS **execution-handoff seam** (turn a dispatch into a real
forked Codex process behind a closed syscall set) and a **Surface/Possibility**
model (explicit interpret/commit profiles, parent-owned child contracts,
per-Possibility isolation) — with a standing instruction to **fix seven open
correctness issues first**.

**Then the model converged (Part III).** The 09-20 Possibility investigation and
the 09-21 consolidation resolve those fronts into one shape:

```text
Mediated actors over a privileged kernel.
  · Only a Move crosses a pipe.
  · The loop IS the kernel — sole writer of every actor's inbound port,
    so authority is delivered by TOPOLOGY, not re-checked per dispatch.
  · Identity is a PORT (a name the runtime owns), never a PID.
  · A Possibility IS a process: its port space IS its Horizon; it parents
    the program, requests spawns (the kernel binds), and confines its children.
  · The fd table splits: a persistent FRAME STACK + a disposable EXECUTION PORT.
  · A Move COMPILES (one atomic turn) into an ENGRAM the OS layer dispatches.
```

Three 09-19 dispositions survive into that model unchanged: the main loop is
**not** a composable Surface (it *holds* authority; a Surface *consumes* it);
`spawn` is a **kernel mechanism, not an address**; and the invocation graph is a
**cactus per Possibility** — now literally one process per sealed tree.

---

# Part I — What DNA-17 implements today

## 1. Purpose and status

DNA-17 greenfields the 2026-09-18 collapse: one smart Interpreter, one dumb
serializing `StateExecutor` (the sole writer), a worker pool, one inbound queue,
`read → commit → execute`, with fd-style identifiers minted at commit. It imports
**no** `skyra-dna-16` package; `codex.go` reimplements the Codex protocol cribbed
from DNA-16's `codex/`.

`go build`, `go vet`, and `go test -race` all pass (25 runs, deterministic). It
does **not** prove live inference, real process containment, durability, or OS
sandboxing — all are in-process fakes by construction.

## 2. File outline

| File | Responsibility |
|---|---|
| `contracts.go` | Structs, interfaces, ownership comments; the closed Operation vocabulary. |
| `cu.go` | `CU`: the one door (`Interpret`/`Submit`), the interpreter loop (`run`/`step`), `deliverRoot`, `Close`. |
| `interpret.go` | `decode` + `interpretInvoke`/`interpretReturn`/`interpretFault`; `authorize`; `invokeTemplate`. |
| `state.go` | `StateExecutor` — the sole writer: `InterpretTurn`, `applyLocked`, the fd table, `stateVia`, `Open`, `snapshot`, id minting. |
| `dispatch.go` | `Dispatch`/`launch` (one worker per fd); `makeOutput` (the persistent execution-plane callback); `commitFailure`. |
| `codex.go` | Codex App Server driver (protocol, thread/turn lifecycle, tool mapping, contamination) — reimplemented, in-process. |
| `probe.go` | Echo child Surface + deterministic root, for the loop proof. |
| `codex_accept_test.go` | Full dynamic-tool round trip, native-contamination rejection, echo-loop sanity, unauthorized-target trap. Fake channel transport, barrier-synchronized. |

## 3. Canonical protocol

```go
type Move struct {
    Identity      MoveID
    PossibilityID PossibilityID
    Source        MoveSource // bound by execute; a Move cannot forge its source
    Target        SurfaceID  // Invoke: the Surface to bind; Return/Fault: unused
    Invoke        *Invoke
    Return        *Return
    Fault         *Fault
}
```

The **opcode is which pointer is non-nil**. `Fault` is a payload variant that
re-enters the one queue — not a third protocol opcode; it shares source
correlation with `Return`.

```go
type Invoke struct {
    InvocationID InvocationID
    Action       string
    Params       []Observation
    Focus        []Capability // must be ⊆ the Possibility's Horizon
    ReturnTo     InvocationID // parent to answer; empty == the root-result boundary
}
type Return struct { Answering InvocationID; Observation Observation }
type Fault  struct { Invocation InvocationID; Reason string }
```

`InvocationID` is stable for one invoked execution. `ReturnTo` names the parent to
answer without the child knowing the parent Surface. A `Return` correlates against
the **recorded** invocation, never a payload claim.

**Possibility / Horizon are pure identity-and-authority data** (no Dispatch/Resolve
methods):

```go
type Possibility struct { Identity PossibilityID; FirstMoveID MoveID; Horizon Horizon }
type Horizon     struct { Generation uint64; Grants map[SurfaceID]Grant }
```

Authority check + Surface binding + find-or-mint of an execution all happen in one
`StateExecutor` transaction.

## 4. The Identifier (fd) and the two planes

`type Identifier uint64` is the fd-style token the **commit** mints per authorized
execution. Only the bare number travels; **scope/generation/liveness live in the
writer's `fds` table**, never in the adapter's hand. The zero value is never live.

- **System plane** — Operations the interpreter issues at commit, a closed set:
  `opRecordInvocation`, `opAuthorizeStart`, `opEmitFault`, `opMarkReturned`,
  `opRootResult`, `opResumeParent`, `opMarkFaulted`, `opRootFault`. A
  `Request = {Move, []Operation}` is the branch-free program interpret hands the
  writer. **Operations are NOT Moves.**
- **Execution plane** — a Surface touches canonical state only through
  `stateVia(fd) → ExecutionState.Commit(EmissionCommit{State, Moves, Failure})`,
  checked live against the fd every time. A dead fd rejects every request.

## 5. The loop (real signatures)

```go
func (cu *CU) Interpret(move Move)                                   // the one door
func decode(move Move) (func(readView, Move) (Request, error), error)
func interpretInvoke(view readView, move Move) (Request, error)      // read + decide
func (s *StateExecutor) InterpretTurn(ctx, move, decide) (commitResult, error) // lock: read+commit
func (s *StateExecutor) applyLocked(request Request) (commitResult, error)     // the commit
func (cu *CU) Dispatch(items []dispatchItem)                         // concurrent fan-out
func (cu *CU) makeOutput(id Identifier, s Surface) func(Emission, error) error // persistent cb
func (s *StateExecutor) stateVia(id Identifier) ExecutionState       // fd-scoped door
```

`step` decodes, runs `InterpretTurn` (the whole read+commit under one lock hold),
then — **after the lock releases** — dispatches authorized executions, re-enters
any corrective/propagated Moves through the door, and delivers root results.
`rootOnce` (a `sync.Map`) enforces exactly one root result per root invocation.
No driver/transport/Surface code ever runs under the writer lock.

## 6. The Surface / Driver boundary

```go
type Surface interface {
    Ingress(ctx, input DriverInput, output func(Emission, error) error) error
    Egress(raw string) ([]Move, error) // driver-owned normalization; no state access
}
```

`Ingress` pushes one step (initial invoke, or a resume carrying a child result)
and returns; **asynchronous emissions arrive later through `output`**, which the
worker binds to the fd. `output` must never run under the writer lock. Native
session state lives in the opaque `DriverState` (writer-owned storage, driver-owned
meaning), never in the loop. `commit-before-send`: host commands run only after the
transition commits.

## 7. The finding (from DECISIONS.md)

- **fd is per-invocation-session, reused across resume** — not "minted on resume."
  The session answering the tool call is the same live session that emitted the
  child; a fresh fd would leave two live fds for one session and orphan the async
  `turn/completed`. So resume re-dispatches on the parent's existing fd. The fd
  cleanly scopes *state access + emission correlation*; it does **not** subsume the
  provider session identity, which stays keyed by `InvocationID` in the driver's
  `live` map and opaque `DriverState`.
- **The worker binds a persistent callback**, not `call Execute; range Emissions()`.
- **Resume happens inside interpret** (`interpretReturn → opResumeParent`), not via
  a second `decode` trip.
- **Settlement is asymmetric:** `Return` settles at interpret (uniform); `Failure`
  kills the fd at emission-commit (the fence needs it dead before later frames),
  then re-enters a `Fault` so the root result stays uniform.

> **Superseded by Part III.** This finding is the whole reason the fd table splits.
> `fd`-per-session was `fd` doing two jobs — *logical continuation* and *live
> execution authority* — at once. Part III separates them (a persistent frame + a
> per-dispatch execution port), so resume mints a **fresh** port and the
> session-identity tension disappears.

## 8. Deliberate deviations & stubs

**Deviations:** string IDs (not `uuid.UUID`, for deterministic tests); the inbound
queue is a buffered channel (records live in state; the channel only transports
pending Moves); the writer is a mutex, not a request-channel goroutine; a small
hand-picked Operation set; the two-mailbox / Supervisor / custody / from→to world
was **deliberately not built** (per the 09-18 landmine warning); `Focus` rides on
the `Invoke`.

**Stubs:** unbounded worker pool (goroutine-per-dispatch); minimal Trap (a pre-exec
authority failure enqueues a Fault; no Trap Surface, no corrective vocabulary);
lightly-exercised non-root fault propagation; no worker monitoring/timeouts; no
Horizon revocation; no durability/outbox/crash recovery; no real Codex
launch/containment; `revision` exists but is never consulted.

## 9. Open correctness issues (the review — fix first)

`specs/2026-09-19/issues.md` records **seven open issues**; fix #1–#3 first.

| # | Sev | Issue |
|---|---|---|
| 1 | High | **Invocation IDs can overwrite executions.** Generated child IDs share the caller namespace and are inserted without a collision check; a child can overwrite its parent and point `ReturnTo` at itself. Also reused across possibilities. |
| 2 | High | **An emitted Fault can target another invocation.** `Fault.Invocation` is trusted, not bound to the emitting fd/possibility — a Surface can fault unrelated work. |
| 3 | High | **Interpreter can deadlock on its own inbound queue.** It synchronously enqueues corrective Moves into the same bounded channel it alone drains. |
| 4 | Med | **Turn-notification ordering can stop the driver.** `turn/started` before the `turn/start` response desyncs response semantics; the receive loop `failAll`s. |
| 5 | Med | **Child returns don't validate the parent's dependency.** `waitingOn` holds one ID; any child return resumes the parent, even the wrong one. |
| 6 | Med | **Late faults overwrite successful terminal state.** `interpretFault` isn't idempotent; a post-Return Fault flips `returned → faulted`. |
| 7 | Med | **Submission can succeed when execution is impossible.** `Open` ignores context/shutdown; empty id/action passes submit but fails later, silently. |

The in-tree `src/skyra/skyra-dna-17/issues.md` carries further **design
observations** (#8 routing async emissions between resumes to the right session;
#9 async emissions have no owning Move; #11 `rootOnce` as belt-and-suspenders on
the settlement fence; #12 `OpenPossibility` atomicity).

---

# Part II — The intermediate layer (`specs/2026-09-19`)

> **Read Part II as the bridge, not the destination.** Its **execution-handoff /
> syscall** work (§11–§13) still holds and is carried forward wholesale — it is the
> OS floor Part III's kernel stands on. Its **Surface/Possibility struct** model
> (§10, the compound `Surface`) is **superseded** by Part III: a Surface becomes a
> bare addressable endpoint, composition moves to the Engram, and the Possibility
> becomes a process. Where the two disagree, Part III wins.

These specs aim to replace the in-process fake with a **real runtime→OS execution
handoff**, and to firm up the **Surface / Possibility / contract** model. Two axes:

## 10. Explicit interpret/commit phases and Surface profiles

[`improvements.md`](../specs/2026-09-19/improvements.md): `step` must call
**interpret** and **commit** as *distinct* functions — not both buried in
`InterpretTurn`. **Absolutely no canonical-state reads past interpret.** Lock what
the turn needs before interpret, hold through commit, release, then execute. A
caller arrives with its fd; interpret resolves the fd's *predefined profile*, which
determines the state functions it may use.

[`surface.md`](../specs/2026-09-19/surface.md): the revised Surface —

```go
type Surface struct {
    Identity       uuid.UUID
    Contracts      []ContractPair   // parent owns input+return contracts for children
    Dependencies   []*Surface       // was Capabilities
    Driver         Driver
    IngressProfile Profile          // interpret→commit for input normalization
    EgressProfile  Profile          // a NEW interpret→commit turn for output
}
```

Egress is its own interpret/commit turn — it does not read canonical state inside
the original endpoint execution. Profiles are carried for dispatch but never grant
endpoint code direct state access. Open: concrete profile signatures, how
`InputID`/`OutputID` select and correlate a call and its return, how higher
Surfaces compose leaves.

[`surfaceID-moveID-fd.md`](../specs/2026-09-19/surfaceID-moveID-fd.md) fixes the
identity vocabulary: **SurfaceID** = the durable registered *address*; **MoveID** =
one call to it; **RequestID** = the committed authorization of an execution step;
**transport handle** = the live pipe binding reused by many calls — distinct from
all three.

## 11. The execution-handoff syscall boundary (`claude-syscall/`)

Attributed to Claude, every point rejectable. The through-line: DNA-17 proved the
logical loop on a fake transport; the runtime→OS seam is what's being named.

- [`syscall-boundary.md`](../specs/2026-09-19/claude-syscall/syscall-boundary.md) —
  a syscall is a **protocol primitive, invoked-not-authored** by app code; the
  syscall set is the execution-plane answer to the closed operation vocabulary.
- [`the-handoff-set.md`](../specs/2026-09-19/claude-syscall/the-handoff-set.md) —
  the Codex round-trip needs exactly six leaves: **`spawn`, `write`, `read`,
  `close`, `wait`, `signal`**. The unit `read`/`write`/`close` act on is a **pipe
  end**, so stdin/stdout/stderr are three pipes, not a stderr special case. The
  **receive loop is a driver pump, not a syscall** — it loops on `read` and crosses
  egress; it cannot be a syscall, only call one.
- [`spawn-not-fork.md`](../specs/2026-09-19/claude-syscall/spawn-not-fork.md) —
  collapse `fork-and-configure-child` into a **`spawn-child`** (`posix_spawn`-shaped)
  syscall; `fork`/`dup2`/`execve` become OS-adapter detail. Deletes the `fork(pid)`
  disclaimers and the phantom "non-launcher parent" requirement. The fd plan is a
  **set of pipes** (wiring is data).
- [`handle-liveness-fence.md`](../specs/2026-09-19/claude-syscall/handle-liveness-fence.md) —
  OS handle liveness is the DNA-17 fd settlement fence one layer down; reuse it.
- [`two-axes-one-seam.md`](../specs/2026-09-19/claude-syscall/two-axes-one-seam.md) —
  `resolve-runtime-process` is DNA-17's driver `live` map promoted to canonical
  state; the fd and the MoveID↔OS-process binding are one seam seen from two layers.

## 12. Runtime syscall Surfaces and the OS floor

[`syscall/skyra-syscall-surface.md`](../specs/2026-09-19/syscall/skyra-syscall-surface.md):
runtime Surfaces resolve authority/resources and *describe* operations; the OS
adapter implements them. A runtime Surface need not be one OS syscall; a PID/fd is
distinct from a MoveID/RequestID/handle. `resolve-runtime-process(moveID)` is an
**interpret-phase canonical read** (no OS crossing). "No reads past interpret"
concerns canonical state — OS reads (draining stdout) are execution effects and
remain necessary; runtime commit and OS side effects are **not one transaction**.
Initial Go implementation: `os/exec` (`exec.Command`/`Cmd.Start`,
`Stdin/Stdout/StderrPipe`) — *not* manual forking of a multithreaded Go runtime.

[`mac-syscall-surface.md`](../specs/2026-09-19/syscall/mac-syscall-surface.md) and
[`os-syscall-reference.md`](../specs/2026-09-19/syscall/os-syscall-reference.md)
map the macOS facilities (`posix_spawn`/`fork`, `execve`, `pipe`, `dup2`/`fcntl`,
`read`/`write`/`close`, `waitpid`, `kill`, `kqueue`/`poll`).
[`what-exists-os.md`](../specs/2026-09-19/transport/research/what-exists-os.md)
separates the three questions: *what exists* (resource) / *how we refer to it*
(handle: fd, PID, port) / *what we can do* (syscall).

[`start-process/`](../specs/2026-09-19/syscall/start-process/proposal.md) holds the
`create-child-process` / `ensure-child-process` / `attach-child-process` contracts
and the dependency map — now candidates to **defer** (see §14).

## 13. Codex and stdio Surfaces

[`codex/codex-integration.md`](../specs/2026-09-19/codex/codex-integration.md):
five top-level Surfaces — `codex.connect` (initialize→initialized, connection
handle), `codex.health` (read-only `thread/loaded/list` probe), `codex.start`
(`thread/start`→`turn/start`), `codex.continue` (`thread/resume`→`turn/start`),
`codex.fork` (`thread/fork`). One long-lived connection serves many start/continue
calls. [`codex-endpoints.md`](../specs/2026-09-19/codex/codex-endpoints.md) is the
JSON-RPC snapshot for installed `codex-cli 0.154.0` (RPC ids ≠ MoveIDs;
threadId/turnId are separate identities).

[`stdio/stdio-surface.md`](../specs/2026-09-19/stdio/stdio-surface.md): a stdio
Surface owns bidirectional framed byte movement (`attach`/`send`/`close`) with a
**persistent receive loop** — no per-RPC competing read. It transports opaque
frames; Codex Surfaces own JSON-RPC/thread/turn/tool meaning. Launch and
termination stay the launcher's responsibility.

## 14. Architectural dispositions carried into the build

- **The main loop is not a Surface.** [`main-loop-as-surfaces.md`](../specs/2026-09-19/main-loop-as-surfaces.md)
  is **parked**: the `interpret→commit→execute→return` phase shape is genuine and
  kept as shared discipline, but `step()` stays a privileged driver — the loop
  *holds* authority; a Surface *consumes* it. The honest model is a sandwich: main
  loop on top, syscalls on the bottom, composable Surfaces between.
- **spawn, the invocation cactus, and per-Possibility isolation.**
  [`invocation-cactus-and-possibility-isolation.md`](../specs/2026-09-19/invocation-cactus-and-possibility-isolation.md)
  ties four faces together: (1) `spawn` is **not addressable** (no SurfaceID) — a
  mechanism that *mints* the child's addressable capability; lifecycle wrappers
  deferred until a workflow must address creation. (2) `spawn` is a **first-class
  system operation**, an execute-phase peer of `read`/`write` — not an
  interpret-phase canonical read. (3) The invocation graph is a **cactus** rooted at
  `rootOnce` / the primordial space; process *resolution* walks it toward the root;
  process *lifetime* is a separate liveness concern. (4) **One cactus per
  Possibility** — a sealed forest, which makes issues #1 and #2 *unrepresentable*
  (scope `InvocationID` under `PossibilityID`; a fd/Fault can only name a frame in
  its own tree) rather than merely checked.

---

# Part III — The current direction (`specs/2026-09-20` → `2026-09-21`)

The 09-19 finding pushed the model past "collapse." The 09-20 Possibility
investigation and the 09-21 consolidation converge on **mediated actors over a
privileged kernel**. The precedence index is
[`carry-forward.md`](../specs/2026-09-21/carry-forward.md); the three 09-21 docs
([`frame-stack-and-execution-ports.md`](../specs/2026-09-21/frame-stack-and-execution-ports.md),
[`possibility-as-process.md`](../specs/2026-09-21/possibility-as-process.md),
`carry-forward.md`) state the current contracts.

## 15. The spine — mediated actors over a kernel

Only a **Move** crosses a pipe. The loop is the **kernel** — a distinguished actor,
not a peer — and the **sole writer of every actor's inbound port**, so an
unauthorized Move *cannot reach* its target: nothing but the loop can put a Move on
that port. **Authority is delivered by topology, not by a per-hop check** — DNA-17's
settlement fence generalized. Above the line: Surfaces (own adapter state,
late-bind their dependencies through mediation). Below the line: the loop,
StateOwner, the authority checker, the tracer — substrate, statically wired,
**not** traced. A Move **compiles** (THINK, one atomic turn) into an **Engram** (the
serialized authorized instruction list) the OS layer dispatches. Two rules keep it
honest: authorization stays **one atomic compile turn**, never a peer-hop chain
(preserves the old atomicity, C3); and **control state is substrate, adapter state
is the actor's** (C8). Topology ops (`spawn`/`bind`/`create_port`/`kill`) are kernel
ops a process only *requests*, authorized as an **attenuating** subset of what it
already holds.
[`mediated-actors-over-a-kernel.md`](../specs/2026-09-20/possibility/claude-input/mediated-actors-over-a-kernel.md),
[`os-kernel-frame-reconciliation.md`](../specs/2026-09-20/possibility/claude-input/os-kernel-frame-reconciliation.md).

## 16. Ports as identity — and the fd table splits in two

Identity **cannot be a PID** (OS-owned, ephemeral, recycled). It must be a name the
runtime owns: a **port**, an index in its own table. Three levels: **Port** (stable)
/ **Binding** (`port → backing`, kernel-owned, rebindable) / **Backing**
(ephemeral, never the identity). DNA-17's `Identifier` + `fdEntry` (scope/
generation/liveness in the writer) **is already a port table**, scoped to one
execution.

The fd table was doing **two jobs** — *logical continuation* and *live execution
authority* — so it splits:

| Was fused in the fd table | Becomes |
|---|---|
| `returnTo`, `waitingOn`, `phase`, `driverState` | **Frame stack** — persistent logical continuation; an arbitrary `map[FrameID]*Frame` for now. Outlives every backing. |
| the live `fd` + `surface` + liveness | **Execution port** — one disposable request/response handle per **dispatch**; the fence. |

Resume mints a **fresh** port (fd-per-dispatch), so a late emission from the
pre-resume backing hits its own dead port and is rejected unambiguously —
DNA-17 issue #8 dissolved.
[`frame-stack-and-execution-ports.md`](../specs/2026-09-21/frame-stack-and-execution-ports.md).

## 17. The Possibility is a process (port space = Horizon)

The port table is **per-process**, not one global map: a process reaches only what
is mapped into *its* space — so a process's **port space *is* its Horizon**. That
collapses five things into one object:

> **Possibility = process = port space = Horizon = cactus root.**

It parents the program, **requests** spawns (the kernel binds, attenuating), and
confines its children. The payoff: **authority is checked once, at provision/bind
time; execution against a provisioned port is free** — no per-dispatch check.
Two guardrails: (1) the Possibility *requests*, the **kernel binds** — if a process
minted its own ports, confinement would be fiction; (2) the Possibility wears two
hats at two ranks — its **port space/frames are substrate**, its **running program
is a backing** (the C4 fix, generalized).
[`possibility-as-process.md`](../specs/2026-09-21/possibility-as-process.md).

## 18. The fence and revocation merge

Revocation = unmap the port = `kill`. So the settlement fence and the
authority/revocation check become **one control-plane lookup** — *"is this port
still mapped and live in its space?"* — answering both "was this revoked" and "is
this settled." That is the single liveness check `handle-liveness-fence.md` wanted;
it closes **C11** and DNA-17 issue #3 (Horizon revocation) together and **retires
the `authGen` generation** (authority is space-membership, not a carried number).

## 19. Same address, sealed backings — the Codex case

Two Possibilities both holding a `codex` port is **not** a cross-tree edge — the
address is not the resource:

| Layer | Codex reading | Shared? |
|---|---|---|
| Port protocol / address | "codex app server" — a *kind* of endpoint | Yes — a type/role, not a resource |
| Binding | this Possibility's `port → thread` | No — A binds `thread_A`, B binds `thread_B` |
| Backing | the **thread** — the live stateful cut | No — **the isolation unit** |
| Provider resource | the app-server **process** hosting many threads | Pooled below the line; addressable by no Possibility |

The **thread** is the confinement unit, not the OS process, so a shared app-server
is safe *pooling* (open decision D-B dissolves). Two obligations, stated not
assumed: **kernel-owned thread demux** (promote DNA-17's private `threads` map to
canonical control state — `two-axes-one-seam`), and a **provider that seals
threads**.

## 20. Dependency mediation; Surfaces, Engrams, the pipe

**Inject the verb, not the noun.** Hand a Surface the *right to request* a
dependency (a typed token redeemed at the runtime), never a live handle — a static
handle *is* "the Surface executes directly," and is the least traceable thing there
is. **Tracing ⟺ mediation ⟺ late-binding** are one property. The **Profile** is the
bridge: a compile-time manifest of what a Surface may ask for (static, checkable)
plus request-time resolution (`Profile request ⊆ Horizon → admit or trap`); it is
the declared-operation contract (C7). The tracer is **not in the graph** — the trust
boundary terminates the recursion. Composition-wise: a **Surface** is an
*addressable executable endpoint* (an int), an **Engram** is the *serialized
composition* (`codex start-thread` is an Engram, not a Surface), and **anything
executable runs behind a pipe speaking the Move protocol** — external code need not
implement a Go in-process interface. Work discovered mid-execution (a Codex tool
call) is the **normal** path: a Surface emitting a new request against its Profile
and suspending.
[`dependency-mediation-and-the-trust-boundary.md`](../specs/2026-09-20/possibility/claude-input/dependency-mediation-and-the-trust-boundary.md),
[`engrams-surfaces-and-pipe-protocol.md`](../specs/2026-09-20/possibility/engrams-surfaces-and-pipe-protocol.md).

## 21. How the earlier layers land now

| Earlier item | Status under Part III |
|---|---|
| DNA-17 finding (fd per session) | **Superseded** — frame/port split; resume mints a fresh port. |
| Issue #1 (id overwrite) | **Cross-possibility part structural** (scope `InvocationID` under `PossibilityID`); same-namespace duplicate still needs an in-loop check. |
| Issue #2 (fault scoping) | **Structural** — a Fault resolves only within its own Possibility; emitted Moves bind to the emitting port. |
| Issue #3 (queue self-enqueue deadlock) | **Still open, still fix-first** — independent of the model. |
| Issue #5 (single `waitingOn`) | Becomes a dependency set (`waiting []FrameID`); single-child enforced at interpret for now. |
| 09-19 compound `Surface` struct | **Superseded** — Surface = endpoint; Engram carries composition. |
| `main-loop-as-surfaces` | **Still parked** — loop holds authority, Surface consumes it. |
| Syscall boundary / `spawn`-not-fork / handle-liveness / six leaves | **Carried wholesale** — now framed as kernel ops below the confinement line. |
| Codex `0.154.0` facts (new-thread-per-tool-change, native contamination) | **Carried** — the only empirically-checked content. |

## 22. Open decisions (current)

From `carry-forward.md`'s live list, deduped to what actually gates a build:

1. **Unbound-port semantics** — fault (start here) vs buffer-until-rebind (durable
   mailbox, later, bounded). Interacts with the supervisor restart window.
2. **Resource-port generation policy** — bump `gen` on decommission→reassign to a
   *new logical resource*, not on same-resource supervisor restart. (Execution ports
   sidestep this — never rebound.)
3. **Kernel ↔ Possibility process placement** *(new)* — one kernel hosting N
   Possibility port spaces, vs. a kernel-per-host with Possibilities as real OS
   children. The two-ranks rule names it but does not pin it to processes.
4. **The Engram contract** — entry representation, ordering, conditional branches,
   result binding, whole-list vs successive-ready-portions.
5. **Read-modify-write atomicity** (C3) and **authority scope of generated Moves**
   (C2 — loop inspection vs target execution).
6. **Return-out-of-a-Possibility edge** (D-A) — the one sanctioned `rootOnce` join;
   and whether a Move can open a child Possibility or only `Submit` can.
7. **Provider-resource lifetime** — who owns the live session, later notifications,
   and correlation after a worker exits; is the executable unit a whole Codex turn
   or each protocol step?
8. Trap/fault recovery contract; Profile placement/validation; multi-child join.

## 23. Immediate work order (current)

1. **Fix issue #3 (queue self-enqueue deadlock)** — real regardless of model;
   internal follow-up work must not require the sole consumer to block on its own
   full queue.
2. **Split the fd table** against DNA-17: `invocationRecord`/`fdEntry` → `Frame`
   (arbitrary map) + `portEntry`; make the port table **per-Possibility** (port
   space = Horizon); retire `authGen` for the "mapped-and-live" lookup. Touch points:
   `opAuthorizeStart` (mint port), `executionProfile.Commit` (fence by port),
   `interpretReturn`→resume (mint fresh port).
3. **Land the interpret/commit phase split** and the **verb-injection Profile**
   shape (declare the right to ask; resolve at request time against Horizon).
4. **Build the real handoff behind the six syscalls as kernel ops** with `os/exec`:
   spawn Codex, wire pipes, run the receive pump, and enforce **kernel-owned thread
   demux**; repeat the DNA-17 dynamic-tool and contamination proofs against a
   **real** `codex app-server`; then layer containment (version pin, isolated
   `CODEX_HOME`, no inherited env).
5. **Resolve D-A (return-out edge) and kernel↔Possibility placement (§22.3, §22.6)**
   before generalizing beyond a single runtime-spawned Codex.

## 24. Conceptual map (current direction)

```text
                         KERNEL (the loop) — sole writer of ports; below the line
                          · compiles Move → Engram in one atomic turn
                          · owns port tables, bindings, thread demux
                          · spawn / bind / create_port / kill (requested, attenuating)
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        ▼                           ▼                            ▼
  Possibility A               Possibility B                Possibility C
  = process = port space = Horizon = cactus root  ...  (sealed; no cross-edges)
        │  frame stack (persistent)          each port bound to its OWN backing
        │  ports mapped in = its authority   (thread_A ≠ thread_B, same address)
        ▼
   program (a backing, above the line)
        │  emits a Move: "I need capability X"
        ▼
   request → kernel authorizes ONCE at bind (⊆ Horizon) → maps a port in
        │  thereafter: execute against the port FREE (authority by topology)
        ▼
   dispatch → backing over a pipe → emission → "port still mapped & live?"
                                                └─ no ⇒ rejected (fence = revocation)

  Provider processes (e.g. one codex app-server) are POOLED below the line and
  demuxed by thread → port → Possibility; no Possibility can name them.
```

### DNA-17 loop as-built (for reference — Part I)

```text
  CU one door → interpreter loop → InterpretTurn (read+commit, one lock) →
  Dispatch (worker per fd) → Surface.Ingress + persistent callback →
  emission → Egress → stateVia(fd).Commit → (dead fd ⇒ rejected: the fence)
```

## Primary references

### DNA-17 (executable reference)

- [DECISIONS.md — shape, deviations, the finding](../src/skyra/skyra-dna-17/DECISIONS.md)
- [contracts.go — protocol, operations, profiles](../src/skyra/skyra-dna-17/contracts.go),
  [cu.go](../src/skyra/skyra-dna-17/cu.go),
  [interpret.go](../src/skyra/skyra-dna-17/interpret.go),
  [state.go](../src/skyra/skyra-dna-17/state.go),
  [dispatch.go](../src/skyra/skyra-dna-17/dispatch.go),
  [codex.go](../src/skyra/skyra-dna-17/codex.go)
- [Correctness review — seven open issues](../specs/2026-09-19/issues.md);
  [in-tree design issues (#8–#12)](../src/skyra/skyra-dna-17/issues.md)

### The next direction (`specs/2026-09-19`)

- Phases & Surfaces: [improvements.md](../specs/2026-09-19/improvements.md),
  [surface.md](../specs/2026-09-19/surface.md),
  [surfaceID-moveID-fd.md](../specs/2026-09-19/surfaceID-moveID-fd.md)
- The syscall boundary: [claude-syscall/README.md](../specs/2026-09-19/claude-syscall/README.md)
  and its five notes (boundary, handoff-set, spawn-not-fork, handle-liveness-fence,
  two-axes-one-seam)
- Runtime→OS: [skyra-syscall-surface.md](../specs/2026-09-19/syscall/skyra-syscall-surface.md),
  [mac-syscall-surface.md](../specs/2026-09-19/syscall/mac-syscall-surface.md),
  [os-syscall-reference.md](../specs/2026-09-19/syscall/os-syscall-reference.md),
  [what-exists-os.md](../specs/2026-09-19/transport/research/what-exists-os.md),
  [start-process/proposal.md](../specs/2026-09-19/syscall/start-process/proposal.md)
- Provider surfaces: [codex-integration.md](../specs/2026-09-19/codex/codex-integration.md),
  [codex-endpoints.md](../specs/2026-09-19/codex/codex-endpoints.md),
  [stdio-surface.md](../specs/2026-09-19/stdio/stdio-surface.md)
- Dispositions: [main-loop-as-surfaces.md (parked)](../specs/2026-09-19/main-loop-as-surfaces.md),
  [invocation-cactus-and-possibility-isolation.md](../specs/2026-09-19/invocation-cactus-and-possibility-isolation.md)

### The current direction (`specs/2026-09-20` → `2026-09-21`)

- Precedence index: [carry-forward.md](../specs/2026-09-21/carry-forward.md)
- 09-21 contracts: [frame-stack-and-execution-ports.md](../specs/2026-09-21/frame-stack-and-execution-ports.md),
  [possibility-as-process.md](../specs/2026-09-21/possibility-as-process.md)
- The kernel model: [mediated-actors-over-a-kernel.md](../specs/2026-09-20/possibility/claude-input/mediated-actors-over-a-kernel.md),
  [os-kernel-frame-reconciliation.md](../specs/2026-09-20/possibility/claude-input/os-kernel-frame-reconciliation.md)
- Ports & mediation: [ports-as-the-location-abstraction.md](../specs/2026-09-20/possibility/claude-input/ports-as-the-location-abstraction.md),
  [dependency-mediation-and-the-trust-boundary.md](../specs/2026-09-20/possibility/claude-input/dependency-mediation-and-the-trust-boundary.md)
- Composition & findings: [engrams-surfaces-and-pipe-protocol.md](../specs/2026-09-20/possibility/engrams-surfaces-and-pipe-protocol.md),
  [authority-namespace-and-execution-context.md](../specs/2026-09-20/possibility/authority-namespace-and-execution-context.md),
  [loop-owned-authority-and-move-serialization.md](../specs/2026-09-20/possibility/loop-owned-authority-and-move-serialization.md),
  [consistency-review.md](../specs/2026-09-20/possibility/consistency-review.md),
  [takeaways.md](../specs/2026-09-20/possibility/takeaways.md)

### Prior summary

- [DNA-16 and the current direction](2026-09-12-dna-16-and-current-direction.md)
