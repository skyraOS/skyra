# Skyra: DNA-19 — the ownership split and the generational direction

**Date:** 2026-09-22  
**Reviewed build:** `skyra/src/skyra` working tree on `3f10fd4` (uncommitted: the
`root/` package split below).  
**Status:** Refactor + boundary review + design session on the permanent base.
`root/` was split from one flat `package main` into one package per
responsibility with **no behavior change**: the same 46 tests and subtests pass
under `-race`, every test's logged rejection reason is identical to the
original, `go vet` is clean and `demo.sh` runs end to end. A fine-comb review
against the boundaries drawn before the build then found where the code does not
align; findings live in [`root/improvements.md`](../src/skyra/root/improvements.md).
The session ended on a **new direction — a recursive generational state
machine** — which is proposal-stage and not built. **Part IV is the new
direction; Parts I–III are how we got there.**

## Executive summary

**The split.** `root/` was one package in which boot, the main loop, runtime
state, a context's authority, runtime growth, wires and trace all shared private
fields and locks. It is now eight packages with one-way imports:

```text
main ─▶ demos ─▶ boot ─▶ kernel ─▶ state ─▶ execution ─▶ ports ─▶ trace
                   └──▶ extension ─┘
          one owner per directory · no import cycles · locks never cross a package
```

**What the review found.** The boundaries mostly hold at the edges — every port
carries only Moves — but the **kernel's ingress is not registered to a port.**
The kernel owns no port; it reads N other actors' ingress and they meet at an
exported Go function, `Kernel.Admit`. That single gap explains the rest: a demo
injects Moves with `Admit` directly, tests do the same 20 times, and the creation
peer changes kernel state through a **callback** instead of a Move. Separately,
the Bookkeeper's `frame` mixes **persistent intent with live handles** (a
`*ports.Binding` pointer), so nothing can be checkpointed.

**Where the design went.** Working through seeding, node roles, checkpointing and
a spawn-cost experiment converged on one model:

```text
skyra invoke codex < move   →  writes frame, spawns child, emission becomes next input
skyra return       < move   →  settles frame (O_EXCL), spawns the parent's NEXT generation
      the machine = the lineage of short-lived instances, not any running process
      persistent things hold connections · ephemeral things make decisions
```

**Why.** Every step boundary becomes a durable, consistent checkpoint; there is
one writer per context *across time*; live state shrinks to one step; and the
wire format does not change — the three Move variants become the three opcodes.

---

# Part I — The ownership split (as built)

## 1. What changed

Every file under `root/` moved (via `git mv`) into the directory named for what
it owns. Private state that other packages used to touch directly is now reached
through narrow methods. Behavior, lock order and trace strings are unchanged.

## 2. Packages

| Package | Owns | Non-test lines |
|---|---|---|
| [`root/boot`](../src/skyra/root/boot) | Bringing a system up and tearing it down: startup decoding, peer processes, registration, publication, parent-key provisioning, reaping (root, codex, user, user+codex) | ~690 |
| [`root/kernel`](../src/skyra/root/kernel) | The main loop: ingress readers, admission, the Scheduler, dispatch | ~270 |
| [`root/kernel/kerneltest`](../src/skyra/root/kernel/kerneltest) | Shared fixture that drives the kernel one Move at a time | ~200 |
| [`root/state`](../src/skyra/root/state) | What exists: Memory (registered executables, retained contexts) and Active (published contexts, reserved ports) | ~200 |
| [`root/execution`](../src/skyra/root/execution) | One context's authority and progress: Horizon, Bookkeeper frames, conversations, commit/realize, detached views | ~770 |
| [`root/extension`](../src/skyra/root/extension) | Runtime growth: delegation to B, bridging, bounded creation | ~175 |
| [`root/ports`](../src/skyra/root/ports) | Wires: bindings, framed Move endpoints, helper and peer processes | ~260 |
| [`root/trace`](../src/skyra/root/trace) | The ordered transition record | ~45 |
| [`root/demos`](../src/skyra/root/demos) | Scripted scenarios that drive a booted system | ~240 |
| `root/main.go` | Picks a mode from argv; nothing else | ~40 |

`echo/`, `parent/`, `creation/`, `crosscontext/` are unchanged.

## 3. How the seams were cut

- **Locks stay private.** Operations that need Active's lock and a context's
  lock at once take a callback that runs under the context's lock:
  `ExecutionContext.Publish` (provisioning), `.Extend` (runtime install),
  `.Admit` (admission + enqueue), `.Inspect` (a locked Horizon view for grant
  checks). Lock order and atomicity match the original.
- **Read-only views.** Other packages see a context through copies —
  `Frames()`, `Parents()`, `Mappings()`, `SinkWired()` — that cannot change it.
- **Rejection lives in ports** (`ports.Reject`), the lowest package that raises
  one; kernel and execution classify it.
- **One launch path.** Three copies of peer-launch code became
  `ports.PreparePeer` + `Helper.Reap`; the codex and user boots share
  address, listener-wait and loop-start helpers.

## 4. Verification

| Check | Result |
|---|---|
| Tests + subtests (baseline vs after, `-race`) | 46 = 46, identical names; live suite skipped in both |
| Rejection reasons logged by tests | Identical (only a random temp path differs) |
| `go vet ./...` | Clean |
| `demo.sh` (baseline + extension + codex demo) | Passes; committed `artifacts/` binaries restored afterwards |

Tests moved next to the code they cover. Two execution tests
(`TestProfileAuthorityBeforeDispatch`, `TestHostParentOwnershipAndBusyFence`)
changed setup only — they now build a Codex endpoint instead of mutating another
package's private binding fields — with assertions unchanged. `live.sh` now
targets `./root/boot`; the README gained a layout table.

---

# Part II — The fine-comb review against the drawn boundaries

## 5. The Scheduler is not a bound process

`Scheduler` is an in-memory queue inside the kernel. Enqueue is a direct call
from `Kernel.Admit`; dispatch is a function argument (`Scheduler.Run(ctx,
k.Dispatch)`). The only real port is at the edge, where `Realize` sends through
a `ports.Egress`.

## 6. Underlying problem: the kernel's ingress is not registered to a port

| Where | What happens |
|---|---|
| `kernel/kernel.go:35` | `kernel.New` creates the kernel with no binding and no reserved port |
| `kernel/kernel.go:173` | `Loop(ctx, inputs)`: the entry points are whatever the caller passes |
| `kernel/kernel.go:184-209` | One reader goroutine per ingress, each calling `k.Admit` |
| `kernel/kernel.go:45` | `Attach` adds another reader — another entry point — at runtime |
| `boot/codex.go:95` (+ user boots) | Boot chooses the entry points by building the `inputs` list |
| `kernel/kernel.go:136` | `Admit` is exported: a door with no port in front of it |
| `demos/extension.go:32` | The demo injects a Move with `r.Kernel.Admit(...)`; nothing crosses a wire |
| `kernel/kerneltest/fixture.go:122`, `:156` | Tests admit by hand (20 call sites) |

## 7. The rule: Moves across ports, internal actions inside actors

Agreed during review: **anything sent over a port is a Move; Moves are the
protocol actors conform to.** Install, publish, register and mint are an actor's
**internal state transitions**, not Moves. No actor triggers another's internal
action except by a Move the owner admits. Boot reaching inside the kernel to get
it to its starting state is **accepted** — it is boot.

| Finding | Status |
|---|---|
| Creation peer mints through a Go closure into the kernel (`creation/peer.go:38`, `extension/extension.go:126`) | **Violation** — improvement 1 |
| `Kernel.Admit` exported and called by a demo and tests | **Violation** — follows from §6 |
| Boot performs kernel-internal actions from outside | **Accepted** by decision |
| Launch config via env (`SKYRA_PARENT_KEY`, `SKYRA_USER_ADDR`) | Not port traffic; out of scope of the rule |

## 8. Frames mix persistent and live state

`frame` (`execution/contracts.go:60-74`) holds persistent fields (request,
target, operation, parent, phase, result, continuation `state`) **and** a live
`*ports.Binding`; `conversation` does too. `Realize` already writes before it
executes (commit, then send), but it writes both kinds together, in memory only.
Nothing is durable. This is improvement 2 and is needed by every architecture
discussed below.

---

# Part III — The design discussion (how the direction was reached)

## 9. Seeding and node roles

Boot should do one thing: bind the kernel to the first port. The kernel is the
first actor; **frame 0 is the root.** Boot then seeds the kernel with one Move on
its birth connection, and the seed tells the node **what it is**. Three roles were
distinguished:

| Role | Owns |
|---|---|
| Kernel (node with a Scheduler) | Scheduling, bookkeeping, provisioning ports for external actors |
| Execution context | Authority scope on a kernel: Horizon, grants, contracts |
| Leaf (e.g. Codex) | Only its own stack; answers Moves; schedules nothing |

Consequence noted: bookkeeping belongs to the kernel, but today each
`ExecutionContext` owns a `Bookkeeper`.

## 10. Why not live distributed cactus stacks — checkpointing

With everything flowing through the root, the root holds **the intent of
everything intended to execute**, written before it runs. With stacks spread
across live actors, a consistent checkpoint needs a coordinated snapshot, or it
waits until an invocation recurses back to the root. That conflicts with one
writer of state, which is what keeps the system simple to reason about.
Persistent (checkpointed) state and live state are **not the same thing**;
making them the same reintroduces write-after-execute.

## 11. One root authority, possibly many Schedulers

One kernel structurally *is* the root that authority flows through; queues and
Schedulers may be many. Authority (whether a Move may happen; the ledger) stays
single-writer; a Scheduler only decides *when*. Context leases with generation
fencing (the existing binding-generation pattern) would reject stale Schedulers.
This leads toward leader/follower; consensus was deferred until a second machine
exists. **Superseded** by §14 for local execution.

## 12. The fused-invocation idea

Each invocation starts a process carrying both the executable and a kernel with
a narrowed Horizon; the parent acts as boot for the child, and the seed Move
carries the child's execution context. Findings: only the persistent half of a
context can travel; "fused" must mean co-located and piped, not shared memory,
or the executable can bypass its own kernel; write-before-execute must hold
across levels; checkpoints become a tree. The idea that survived: **centralize
the writes, not necessarily the traffic.**

## 13. Spawn-cost experiment

M1 Pro, Go 1.25, normal builds; medians over two runs.

| What | p50 | p95 |
|---|---|---|
| Spawn echo → first Return | 2.5–2.6 ms | 2.8–3.6 ms |
| Invoke → Return on a running echo | 12 µs | 18–22 µs |
| Root boot → first Return through the kernel | 2.75–2.8 ms | 3.1–3.2 ms |
| Codex boot (codex + provider + echo) → connect | 6.8 ms | 7.6–7.8 ms |

Spawning is ~200× a warm Move; the kernel's own routing adds ~0.2 ms. First-run
outliers (151 ms, 322 ms) were macOS first-launch checks on fresh binaries.

---

# Part IV — The new direction: a recursive generational state machine

## 14. Definition

A recursive generational state machine is a system where each execution takes
input, updates its state, and produces output that becomes the input and state
of a newly spawned successor. **The persistent machine is the lineage of
successive instances, not any single running process.** Each process reads the
state available to it, decides what to write or execute next, provisions child
work and continuation, then exits. When a child returns, its result activates a
**new successor** instead of waking the original. Execution is ephemeral; state,
authority and continuation form the durable thread.

Decisions:

- **The parent spawns the next program.** No central activator.
- **Single writer per context holds across time** — exactly one live generation
  per context at any moment.
- **Every step boundary is a consistent checkpoint** — this answers §10 without
  giving up the cactus shape.
- **Authority narrows going down**: a parent provisions the child's Horizon and
  where it writes.

## 15. Opcodes are Moves

One binary. The verb is the opcode and carries its own logic; the opcodes are
the protocol's existing three variants, so the wire format does not change.

```text
skyra invoke codex < move.json    write the frame, spawn codex, turn its emission into the next run
skyra return       < move.json    settle the frame, spawn the parent's next generation (continuation, result)
skyra fault        < move.json
```

## 16. Rules

1. **Move on stdin, opcode in argv** (argv is public in `ps` and size-limited).
2. **A frame's node is an inherited file descriptor**, never a path.
3. **Provenance comes from process ancestry**: only the parent, which holds the
   child's stdout, turns that output into the next run.
4. **Durable write-before-execute**: continuation and children's intents are
   flushed before anything spawns; temp-then-rename, never rewrite.
5. **Exactly one successor**: `skyra return` creates `result.json` with `O_EXCL`;
   duplicates and late returns fault.
6. **Checks run in every step**: contracts, source validation, parent
   references, the contamination fence.
7. **Nondeterminism is written before use** (time, randomness, model output), so
   a lineage replays.
8. **No secrets in the tree or in Moves.**

```text
.skyra/frames/<id>/intent.json          written by the parent before spawning
.skyra/frames/<id>/gen-<n>/state.json   one per generation, immutable
.skyra/frames/<id>/result.json          created once, O_EXCL
.skyra/frames/<id>/<child-id>/...       children nest, mirroring the cactus
```

Cost from §13: an invoke→return cycle is ≥3 spawns; with durable flushes expect
~8–15 ms — negligible for LLM-sized steps, expensive for echo-sized ones. A warm
pool of generic runners waiting on stdin brings a step near the warm path.

## 17. Persistent services, ephemeral adapters

**Persistent things hold connections; ephemeral things make decisions.** The
Codex app-server stays persistent, and so does the user's view. What a
persistent thing knows enters the tree only through an adapter step, as an
opaque handle — never the foreign state itself. Because services emit
unprompted, each splits in three:

| Part | Lifetime | Codex | User view |
|---|---|---|---|
| Service | persistent | app-server + provider | browser |
| Connection holder | persistent, relays only | holds app-server stdio, forwards events | holds the WebSocket listener, forwards messages |
| Adapter | ephemeral, one step per event | Move ↔ app-server request; contamination fence | submit → Invoke; Return → snapshot |

Today each binary is all three: `codex/contracts.go:22-34` keeps thread/turn
IDs, seen calls, pending dependency, history and retirement in the persistent
process; `user/interaction.go:9-20` keeps the one-pending `busy` state and the
request ledger there. Under the rule that decision state moves into the tree.
Accepted consequences: **handles can dangle after a restart** (the adapter must
fault, not assume), and **connection holders must never branch on content**.

## 18. How earlier layers land now

| Earlier item | Status under the new direction |
|---|---|
| DNA-19 host runtime (Kernel…Bookkeeper) | **Refactored** into responsibility packages; the reference implementation for the checks every `skyra` step must run |
| Central Scheduler + in-memory single ledger (§9, §11) | **Superseded** by parent-spawned generations and a durable tree |
| Seed Move; provenance from the wire; no secrets in Moves | **Carried** — provenance now comes from process ancestry |
| 09-21 frame stack vs execution port (DNA-19 work order item 1) | **Recast** as persistent frame record vs live handles (improvement 2) |
| Codex actor / user peer | **Split pending**: persistent service + connection holder, ephemeral adapter |
| HMAC parent references, contamination fence | **Carried**; run inside every step |

## 19. Open decisions

- Does a generation write its own node directly through its inherited handle,
  or return its writes for one writer to apply?
- Is any central Scheduler left, or is the parent's spawn the only scheduling
  decision?
- Does creation stay a separate actor emitting a mint request as a Move, or
  become an internal action of the parent generation?

## 20. Immediate work order

1. **Commit the ownership split** (Part I) as the base for everything below.
2. **Split the frame** into a serializable record and live handles; make the
   record durable (improvement 2).
3. **Remove the side doors**: unexport `Admit`, drive tests and the demo through
   a port, and make creation answer with a Move (§6, improvement 1).
4. **Prototype one generational lineage**: `skyra invoke echo` / `skyra return`
   over the `.skyra/` tree with `O_EXCL` settlement, measured against §13.
5. **Draw the adapter line in `codex/` and `user/`**: pull decision state out of
   the persistent processes.

## 21. Conceptual map (new direction)

```text
  boot ──spawn + seed──▶ gen 0 (root)                .skyra/ tree (durable, one writer per node)
                           │ writes intent ─────────▶ frames/1/intent.json
                           │ spawns
                           ▼
                   skyra invoke codex ──stdin Move──▶ adapter step (ephemeral)
                                                         │ relays via
                                                         ▼
                                     connection holder (persistent, dumb) ⇄ app-server (persistent)
                                                         │ emission
                                                         ▼
                   skyra return ──O_EXCL result.json──▶ spawns gen 1 of the parent (continuation, result)

  Guarantee line:  every step boundary ⇒ consistent checkpoint            ✅
                   live leaf state (provider threads) ⇒ in the checkpoint ❌ (faulted on restore)
```

## Primary references

### This session

- [root/improvements.md](../src/skyra/root/improvements.md) — findings, rules,
  and the full target model
- Split packages: [boot](../src/skyra/root/boot), [kernel](../src/skyra/root/kernel),
  [state](../src/skyra/root/state), [execution](../src/skyra/root/execution),
  [extension](../src/skyra/root/extension), [ports](../src/skyra/root/ports),
  [trace](../src/skyra/root/trace), [demos](../src/skyra/root/demos)
- Key seams: [kernel/kernel.go](../src/skyra/root/kernel/kernel.go),
  [execution/context.go](../src/skyra/root/execution/context.go),
  [execution/contracts.go](../src/skyra/root/execution/contracts.go),
  [state/state.go](../src/skyra/root/state/state.go),
  [extension/extension.go](../src/skyra/root/extension/extension.go),
  [creation/peer.go](../src/skyra/root/creation/peer.go)
- Adapter-line evidence: [codex/contracts.go](../src/skyra/codex/contracts.go),
  [user/interaction.go](../src/skyra/user/interaction.go),
  [user/browser.go](../src/skyra/user/browser.go)
- [README](../src/skyra/README.md) (layout table)

### Prior summary

- [DNA-19 — the permanent base and the external-binary boundary](2026-09-22-dna-19-the-permanent-base-and-the-external-binary-boundary.md)
