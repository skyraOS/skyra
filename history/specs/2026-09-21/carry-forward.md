# Carry-forward — what still sticks in the current meta

Date: 2026-09-21
Status: Consolidation index. Not a new decision; a filter over the 2026-09-19 and
2026-09-20 specs saying what survives into the current model, what is
parked/superseded, and which open contracts are still live. Answers
[consistency-review C12](../2026-09-20/possibility/consistency-review.md) (no single
current contract supersedes the checkpoints). Reading it does not replace the
source docs; it says which of them to still build against.

## The current meta in one place

The model both dated folders converge on:

> **Mediated actors over a kernel.** A Move is the only thing that crosses a pipe.
> The loop is the kernel — a distinguished actor, not a peer — and the sole writer
> of every actor's inbound port, so authority is delivered by **topology, not by a
> check**. Above the line: Surfaces (executables that parse Moves, own adapter
> state, late-bind their dependencies through mediation). Below the line: the loop,
> StateOwner, the authority checker, the tracer — substrate, statically wired, not
> traced. A Move **compiles** (THINK, one atomic turn) into an **Engram** — the
> serialized authorized instruction list — which the **kernel/OS layer** then
> dispatches to backings over ports.

Everything below is the detail that survives under that spine.

---

## What sticks

### Identity & location

- **Identity is a runtime-owned name, never a PID.** PIDs are OS-owned, ephemeral,
  recycled. Erlang/Mach/DNS all refuse the instance handle as identity.
  → [ports-as-the-location-abstraction](../2026-09-20/possibility/claude-input/ports-as-the-location-abstraction.md)
- **Three levels: Port (stable) / Binding (kernel-owned, rebindable) / Backing
  (ephemeral).** `Port : Process :: Profile : Invocation :: Engram : dispatch-run`.
- **Port space = Horizon.** The port table is **per-process**, not one global map:
  a process reaches only what is mapped into *its* space. "Isn't mapped in → can't
  be reached" and "the horizon's complement is what it can't touch" are the same
  statement. Object-capability confinement, made structural.
  → [os-kernel-frame-reconciliation](../2026-09-20/possibility/claude-input/os-kernel-frame-reconciliation.md)
- **The backing/process id is monotonic and non-addressing.** Only ports address;
  the process-table id is runtime-assigned, never the OS pid, never an address.
- **The fd table splits in two.** DNA-17's fused table becomes a persistent **frame
  stack** (logical continuation; arbitrary map now) and a disposable **execution
  port** (the request/response unit; the fence).
  → [frame-stack-and-execution-ports](frame-stack-and-execution-ports.md)
- **A Possibility is a process** whose port space *is* its Horizon: it parents the
  program, requests spawns (the kernel binds, attenuating), and confines its
  children. Authority is checked **once at bind**, not per dispatch; revocation is
  `kill`. Same address ≠ same resource — two Possibilities holding a `codex` port
  bind different threads. → [possibility-as-process](possibility-as-process.md)
- **You already built the port table.** DNA-17's `Identifier` + `fdEntry`
  (scope/generation/liveness in the writer) is a port table scoped to one
  execution; the move is to generalize it.

### Kernel & authority

- **Everything leaving the loop for execution is an AuthorizedMove; not everything
  entering it is.** The loop owner authorizes its own preparation (e.g. reading a
  Horizon to check a Move) without a separate authority service, and inspection
  authority ≠ target-execution authority.
  → [loop-owned-authority-and-move-serialization](../2026-09-20/possibility/loop-owned-authority-and-move-serialization.md)
- **Authority is delivered by topology.** The loop is the sole writer of any actor's
  inbound port; an unauthorized Move *cannot reach* the target because nothing but
  the loop can put a Move on that port. This is DNA-17's settlement fence
  generalized. → [mediated-actors-over-a-kernel](../2026-09-20/possibility/claude-input/mediated-actors-over-a-kernel.md)
- **Authority stays one atomic compile turn — never a hop through peer actors.**
  Distributing authorization across actor-hops re-opens the regress and loses C3
  atomicity.
- **Control state is substrate; adapter state is the actor's.** The port table,
  Horizons, bindings are not "just another actor's private state." This split is
  what terminates the recursion and answers C8.
- **Topology manipulation is itself capability-checked and attenuating.**
  `spawn / bind / create_port / kill` are kernel ops a process only *requests*; a
  parent delegates only a subset of what it holds; broader authority is minted by
  the kernel, never a peer. Without this, "reachability = authority" leaks the
  moment any process orchestrates.
- **`kill` is the revocation primitive.** Authority disappears with the process;
  lifecycle *is* the authority lifecycle.

### Dependency mediation

- **Inject the verb, not the noun.** Hand a Surface the *right to request* a
  dependency (a typed token redeemed at the runtime), never a live handle. A static
  handle is the least traceable thing there is and *is* "the Surface executes
  directly." → [dependency-mediation-and-the-trust-boundary](../2026-09-20/possibility/claude-input/dependency-mediation-and-the-trust-boundary.md)
- **Tracing ⟺ mediation ⟺ late-binding.** They are one property, not a tension.
- **The Profile is the bridge.** Declaration-time (static, typed manifest of what a
  Surface may ask for) and request-time (dynamic: Profile request ⊆ Horizon → admit
  or trap) are the two states, both true at once. The Profile is the declared-
  operation contract (answers C7).
- **The tracer is not in the graph.** Terminate by a trust boundary, not a base case
  inside it. If a component must exist before *any* dependency resolves → substrate,
  static, untraced. If *some* computations need it and others don't → Surface,
  mediated, traced.

### Execution model

- **Only a Move crosses a pipe.** Anything executable — including model-authored
  behavior not in the runtime binary — runs behind a pipe and speaks the Skyra Move
  protocol. That protocol is the interoperability boundary.
  → [engrams-surfaces-and-pipe-protocol](../2026-09-20/possibility/engrams-surfaces-and-pipe-protocol.md)
- **Surface = an addressable executable endpoint (an int).** Not a compound struct
  carrying Dependencies + Contracts + Driver.
- **Engram = the serialized composition.** `codex start-thread` is an Engram, not a
  Surface. The Engram is what a Move compiles into before dispatch; the OS runs
  engrams, it does not replace them.
- **The wire is async Moves, not RPC.** Real sessions emit asynchronously and
  unprompted (a turn starts, calls a tool mid-turn, resumes). The unit is a Move
  that re-enters the loop; correlation and resume are protocol, not return values.
  DNA-17 proved `send → response` does not hold.
- **Persistent parent, disposable backings.** The parent keeps the logical
  invocation and its externalized continuation; a backing is a request/response
  shell that may exit after reporting "needs a child result." Three lifetimes:
  logical invocation / worker execution / provider resource.
  → [disposable-workers-state-turns-and-execution-policy](../2026-09-19/disposable-workers-state-turns-and-execution-policy.md),
  [persistent-parent-and-separate-pools](../2026-09-19/persistent-parent-and-separate-pools.md)
- **Work discovered mid-execution is the normal path, not a special case.** A Codex
  turn spawning tool work is a Surface emitting a new request against its Profile
  and suspending; it re-enters the one door. The "no unresolved prerequisite reaches
  the executor" rule governs *pre-invocation* readiness only.
- **The receive pump is a Surface, not a syscall.** It loops on `read`, frames
  bytes, and crosses back through egress (interpret→commit) per frame. It outlives
  any one dispatch — DNA-17's persistent `makeOutput` callback bound to the handle.

### Syscall / OS boundary

- **Syscalls are protocol primitives from a closed set — invoked, never authored,
  by workflow code, and consuming authority rather than minting it.** This is the
  execution-plane analogue of the system-plane `op*` vocabulary (closes half of
  DNA-17 issue #1). → [syscall-boundary](../2026-09-19/claude-syscall/syscall-boundary.md)
- **The handoff needs six leaves: `spawn`, `read`, `write`, `close`, `wait`,
  `signal`.** → [the-handoff-set](../2026-09-19/claude-syscall/the-handoff-set.md)
- **`spawn`, not fork.** `posix_spawn`-shaped: image creation with a prepared fd
  plan. fork/dup2/close-on-exec/execve are OS-adapter detail *below* the boundary.
  → [spawn-not-fork](../2026-09-19/claude-syscall/spawn-not-fork.md)
- **The fd plan is a set of pipes.** A pipe is the unit; stdin/stdout/stderr are
  three pipes wired to child fds 0/1/2 by convention. "stdio transport" is a wiring
  policy, not a type; stderr is just a third pipe.
- **OS handle liveness is the port fence one layer down.** Mint process/transport
  handles with generation + liveness in canonical state, checked live at execute
  time; a write to a dead/reused handle is structurally rejected.
  → [handle-liveness-fence](../2026-09-19/claude-syscall/handle-liveness-fence.md)
- **`resolve-runtime-process` is a canonical-state read, not an OS crossing** — the
  driver `live` map promoted to protocol-owned state (two-axes-one-seam).

### State discipline

- **Interpret is the read phase; commit is the write phase; absolutely no canonical
  reads past interpret.** `step` calls them as visibly distinct functions; hold the
  turn's locks across both, release before execute.
  → [improvements](../2026-09-19/improvements.md)
- **Reads and writes stay in one consistent state turn** — not a read pool and a
  write pool. The prepared Request goes straight from interpret to commit, never
  parked in an independent queue where its assumptions go stale.
- **Preparation terminates in serialized executable operations.** A primitive state
  read/write becomes an executable entry; it does not re-enter the whole preparation
  cycle to run. StateOwner is a Surface executing state ops and emitting Moves.

### Isolation

- **One cactus per Possibility — a sealed forest.** `ReturnTo` is a call stack;
  concurrent fds make each run a cactus; scope `FrameID` **under** `PossibilityID`
  and cross-possibility collisions/faults become *unrepresentable*, not checked.
  → [invocation-cactus-and-possibility-isolation](../2026-09-19/invocation-cactus-and-possibility-isolation.md)
- **Process resolution is a walk toward the root; a bound process's lifetime is
  ownership/liveness, not push/pop.** The frame holds a reference; the resource
  lives on its own fence.

### Trap / fault

- **Trap = transfer control to preparation. Fault = an error.** An invocation whose
  required state is unavailable is trapped, not dispatched unprepared. TrapController
  is an addressable Surface, not a callback inside the authority check.
  → [authority-namespace-and-execution-context](../2026-09-20/possibility/authority-namespace-and-execution-context.md)

### Codex grounding (facts, not direction)

- Four/five `codex.*` top-level Surfaces over App Server JSON-RPC; `initialize →
  initialized` handshake; `thread/start` then `turn/start`; tool call answered by
  the RPC response, **not** `thread/resume`.
  → [codex-integration](../2026-09-19/codex/codex-integration.md),
  [codex-endpoints](../2026-09-19/codex/codex-endpoints.md)
- **New thread per dynamic-tool change** — 0.154.0 has `dynamicTools` only on
  `thread/start`; resume/settings have no such field.
- Native effect-bearing items (`commandExecution`, `fileChange`, `mcpToolCall`, …)
  contaminate a turn → fault; the fence blocks a later clean message from producing
  a Return. (DNA-17's rejection proof.)

---

## Parked or superseded — do NOT carry

- **`step()` as composable Surface** — parked. The loop *holds* authority; a Surface
  *consumes* it. Keep the phase shape as shared discipline only.
  → [main-loop-as-surfaces](../2026-09-19/main-loop-as-surfaces.md)
- **StateOwner as whole-loop host / god-object** — corrected. It is a provisioned
  state *service* (a Surface); coordination is a different rank. The three-role
  folders (authority-owner / execution-broker / agent-surface-worker) are
  responsibility boundaries to test, **not** three processes and not a settled
  architecture. → [takeaways](../2026-09-20/possibility/takeaways.md)
- **One fd per invocation, reused across resume** — superseded by the frame/port
  split (per-dispatch execution ports; resume mints a fresh port).
- **Single `waitingOn` slot** — superseded by a dependency set (`waiting []FrameID`),
  with single-child still enforced at interpret for now.
- **Synchronous `send → receive` / RPC framing** — superseded by async Moves.
- **Static instance injection of dependencies** — rejected; inject the verb.
- **`fork-and-configure-child` / `setup-child-pipes` as surfaces** — folded below the
  `spawn` boundary; the non-launcher-parent "cooperating fork" section and the
  `fork(pid)` disclaimers are dropped.
- **Surface as a compound struct (Dependencies + Contracts + native Driver)** —
  superseded: Surface is an endpoint; Engram carries composition; external code need
  not implement a Go in-process interface.
- **Execution context carrying an invocation stack / Horizon copy / resource bag** —
  superseded: the execution context is serialized Moves + handlers, ideally one
  handler list. Frames and preparation bookkeeping stay with the coordinator.
- **"Issue #1 is fully structural"** — overstated: scoping `InvocationID` under
  `PossibilityID` kills *cross*-possibility collisions, not *same*-possibility
  duplicate IDs. Keep an in-namespace collision check.

Identity encoding is still unsettled (DNA-17 string ids + uint64 Identifier vs the
proposals' UUIDs). Do not equate MoveID, invocation identity, execution handle
(port), and OS fd — they are separate identities on separate lifetimes.

---

## Live open contracts (deduped across both folders)

1. **Unbound-port semantics** — fault (start here) vs buffer-until-rebind (durable
   mailbox, later, bounded). Interacts with the supervisor restart window.
2. **Resource-port generation policy** — bump on decommission→reassign to a *new
   logical resource*, not on same-resource supervisor restart. (Execution ports
   sidestep this entirely.)
3. **Engram contract** — entry representation, encoding, ordering, conditional
   branches, result binding, and whether the executor gets the whole list or
   successive ready portions.
4. **Atomicity of related read/modify/write groups** (C3) — ordered serialization
   does not by itself serialize a read-modify-write group; define atomic groups and
   conditional-write validation.
5. **Authority scope of generated Moves** (C2) — loop inspection (loop's authority)
   vs target execution (the Possibility's Horizon); how trusted caller identity
   reaches the checker.
6. **Return-out-of-a-Possibility edge** (D-A) — the one sanctioned `rootOnce` join to
   a write-once sink in the opener; and whether a Move can open a child Possibility
   or only `Submit` can.
7. **Shared process vs per-possibility isolation** (D-B) — per-possibility spawn
   (clean) vs one shared backing demuxed to the owning Possibility (cheap, makes
   emission-routing load-bearing).
8. **Provider-resource port** (C10 remainder) — who owns the live provider session,
   later notifications, and correlation after a worker exits; is the executable unit
   a whole Codex turn or each protocol step?
9. **Trap/fault recovery** — what TrapController receives, may do, how its outcome
   answers the original invocation, and what happens if it fails or is denied.
10. **Profile placement** — where bounded-unit read/write profiles are declared,
    validated, and stored; how a Profile request maps to Horizon.
11. **Multi-child join** — how interpret identifies and authorizes all moves a
    satisfied dependency unblocks in one commit.
12. **Interpreter self-enqueue deadlock** (issue #3) — internal follow-up work must
    not require the sole consumer to block on its own full queue.
13. **OpenPossibility / shutdown honesty** (issue #7) — reject submissions after
    shutdown; validate the first Invoke before recording the Possibility.

---

## Precedence (when two docs disagree, the later wins)

| Topic | Current source of truth |
|---|---|
| Identity / location | ports-as-the-location-abstraction + frame-stack-and-execution-ports |
| Possibility model | possibility-as-process |
| Kernel & authority | mediated-actors-over-a-kernel + loop-owned-authority-and-move-serialization |
| Port space = Horizon; topology rules | os-kernel-frame-reconciliation |
| Dependency handling | dependency-mediation-and-the-trust-boundary |
| Surface / Engram / pipe | engrams-surfaces-and-pipe-protocol |
| Execution context shape | authority-namespace-and-execution-context |
| Syscall boundary | claude-syscall/* |
| State phases | improvements |
| Isolation | invocation-cactus-and-possibility-isolation |
| Findings ledger | consistency-review |

The `proposal/` code-on-paper and `reference/dna17.go` are historical source
material, not the current object layout. Where a `proposal.go` struct still shows
the fused fd table, single `waitingOn`, or compound Surface, read it through this
index, not literally.

## Related

- [Frame stack and execution ports](frame-stack-and-execution-ports.md)
- [Consistency review](../2026-09-20/possibility/consistency-review.md)
- [Possibility investigation entry](../2026-09-20/possibility/README.md)
- DNA-17 issues
