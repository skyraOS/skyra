# Skyra: the runtime era, from DNA 7 to 19

**Date:** 2026-09-22
**Reviewed:**
- `skyra/archive/throwaway-dna/`: `skyra-dna-v7` to `v14`, `skyra-ctu-engine`, and
  `skyra-dna-15` to `18.5`;
- the live DNA-19 code in `skyra/src/skyra` (flattened from `skyra-dna-19`,
  with the uncommitted `root/` package split in the working tree).

**Span:** 2026-08-20 (v7 was set up the day v2–v6 were archived) to 2026-09-22
(`0b3ec15`, "promote DNA-19 to permanent base; archive throwaways").

**Scope:**
- The throwaways are about 25.6k lines of Go, of which about 5.7k are tests. I read
  every generation's README or DECISIONS record, the core source of v7 and DNA-17,
  and the key interfaces and implementations of the rest.
- I then checked specific claims against the code: authentication, signing, stubs,
  and panics.
- I did not re-read the design specs under `skyra/specs/`. DNA-16, 17 and 19 already
  have their own detailed summaries (09-12, 09-20 and two on 09-22), so this document
  covers the arc and points to those for depth.

**Health:**

| code | status |
|---|---|
| **DNA-19** (`src/skyra`) | `go build`, `go vet` and `go test ./...` pass (38 test functions plus many subtests). I did not run the opt-in live Codex suite. |
| **The throwaways** | do not build as they sit: they import `skyra-v1/src/skyra/skyra-dna-*`, which moved when they were archived |

Each throwaway's record says it passed build, vet and race tests at the time. There
are about 113 test functions across them.

**Dates:** git history for v12–16 and the ctu-engine is squashed into one snapshot
commit (`40a0340`, 09-19). I dated those generations from their own README and
DECISIONS headers.

---

## 1. What this era is

The agent era (v2–v6) asked **how a being learns**. This era sets that aside and asks
something narrower and more mechanical: **what is the execution substrate, and who
owns what inside it?**

- **The same thing gets rebuilt over and over:** a runtime that carries a
  request out to something untrusted and brings a checked result back.
- **The cognitive vocabulary disappears:** there is no memory graph, no grading, and
  no expectations.
- **Language moves from cognition to machine design,** in two waves:
  - **CPU vocabulary** (v7–v14): CTU, reorder buffers, register files, fds and
    capabilities;
  - **ownership vocabulary** (15–19): controller, scheduler, StateExecutor, kernel,
    owner, and "who is the sole writer of what".

**The method is as important as the architecture.** From v7 onward, every generation
is a deliberately **throwaway, boundary-finding build**:
- a fixed set of interfaces is declared;
- an overnight implementer must build the loop *without changing that declared
  surface*;
- the implementer reports every place the surface failed to hold.

The paper trail repeats each time:

| file | role |
|---|---|
| `overnight.md` / `workorder.md` | the job |
| `boundary-crossings.md` | where the interfaces broke |
| `claude-ideas.md` | which parts of the shape were Claude's proposals, each attributed and rejectable |
| `DECISIONS.md` | what was actually decided |

The top-level README says it outright: *"Do not add structs or fields… If a boundary
prevents progress, report the missing capability."* A generation that **fails to
close its loop** under that rule still produces its main output: the list of
interfaces that didn't hold.

## 2. The arc

```
PHASE 1 — THE CTU AS A CPU  (v7–v11, 08-20 → 08-26)
v7   08-22  MOVE = custody transfer; GREET/RECOGNIZE ports; kernel Activate/Remember/Define   RUNS
v8   08-23  the possibility PARTITIONED across CTUs; ed25519-signed segments; gather-REMEMBER   RUNS
v9   08-23  possibility = REORDER BUFFER; edge-triggered DRIVE; loops unfold by tag           RUNS
v10  08-24  THINK's acquire pipeline; faults cross to a privilege principal; kernel MINTs     RUNS (fresh build)
v11  08-24  Body (compiled) vs Context (live); COLD path (kernel) vs HOT path (capability)   RUNS (hot path)

PHASE 2 — THE COLD PATH AND THE LOOP THAT WOULDN'T CLOSE  (v12–ctu-engine, 08-27 → 09-07)
v12  08-27  cold path: COMPILE → BOUND → MINT → REMEMBER; engram disk tier                  INTERFACES ONLY
v13  ~09-01 direct loop probe, frozen surface                                                BLOCKED (echo proof not met)
v14  ~09-01 single-wave local Invoke/Return; protocol/driver/controller/scheduler/ctu split  RUNS
ctu  09-06  fd = MoveID; one Memory fd table; Fire(batch)                                    PARTIAL (panics on unbuilt paths)

PHASE 3 — OWNERSHIP, THEN A REAL PROCESS BOUNDARY  (15 → 19, 09-07 → 09-22)
15   ~09-08 transition loop + kernel-owned registration; placement-gated continuations      RUNS
16   09-12  long-lived external driver (Codex App Server) behind opaque DriverState          RUNS (fake peer)
17   09-19  "the collapse": one Interpreter, one sole-writer StateExecutor, worker pool     RUNS (fake peer)
18   ~09-20 root boundary: a REAL echo process, serialized Moves over pipes                 RUNS (real process)
18.5 ~09-21 cross-context activation + capability-mediated creation of a new executable    RUNS
19   09-22  Codex moved OUT into its own binary; HMAC parent refs; kept as permanent base   RUNS (+ live Codex)
```

## 3. Phase 1: the CTU as a CPU (v7–v11)

This phase grows out of the CTU pivot at the end of the agent era (v6's "eight
primitives over one transport verb", 08-19). The idea is to **treat the runtime as a
processor**, where the "CTU" (causal transition unit) is the CPU and every hard
question maps onto a hardware analogy.

### v7: the quadrant, as packages (08-22)
**The four quadrants:**

| | system space (trusted) | user space (untrusted) |
|---|---|---|
| **privileged** | the kernel | |
| **unprivileged** | the CTU (dispatcher, interpreter, scheduler) and the drivers, which form the trusted host table | entities, the untrusted entity table, whose output is never trusted |

**`MOVE` is custody transfer,** the only transport. Its envelope carries a header,
target, horizon, optional system instruction and correlation; its payload is opaque
to everything except the entity.

**A run goes:** request → **GREET** (the request port) → kernel **Activate** →
**Interpret** into the set of eligible resources → **Schedule** → the entity runs →
the response arrives at **RECOGNIZE** (the response port), which checks the returned
observation against the contract stored in the parent frame.

**The kernel has three verbs:**
- `Activate` loads a name from durable storage into a live possibility;
- `Remember` is root-only: it commits the cone and tears down the arena;
- `Define` authenticates a PIN into an account plus its root horizon.

The macOS driver writes the graph with a flushed, atomic replace.

**What didn't hold** (`changes/boundary-crossings.md`):
- Six crossings returned zero values on failure, so every call gained a three-way
  `Outcome`: Success, Failure or NotYet.
- `Greet` had no return address.
- A `Target` could reach a runtime but not an instruction.
- The correlated contract had no home, so an `Expect` ledger was added.
- Resource trust booleans were turned into a **sum type** (host, entity or kernel
  resource), because which dispatch table answered *is* the trust decision.

### v8: the distributed possibility (08-23)
- A possibility is **split across CTUs**. When it hits a dependency it can't resolve
  locally, the CTU forwards the **slice** instead of parking the parent.
- A MOVE is the cut between parts.
- Each CTU **signs** (ed25519) the provenance segment it produced.
- `Remember` gathers the content-addressed cone home through `Fabric.Fetch`.

### v9: the possibility as a reorder buffer (08-23)
- The possibility becomes a graph of Moves over a **register file**, with an
  in-flight cut index and a retirement watermark.
- The CTU is a nearly stateless core, and execution is **edge-triggered**: GREET,
  RECOGNIZE and local completion each feed one engine, `Drive`.
- The Slice is dissolved: a cut carries a narrow child Possibility.
- The kernel gains **Mint**.
- Loops unfold by tag and stop when a SWITCH move stops the growth.
- The demo runs a four-move DAG across two CTUs with fan-out, two parked cuts, a join,
  a gather-Remember, and two tagged loop iterations.

### v10: can THINK acquire capabilities securely? (08-24, a fresh build that imports nothing)
- `acquire` is a possibility with three children: discover, test and admit.
- THINK and acquire hold **no permission to ask and never mint**. First contact with a
  new endpoint **faults** to a separate *privilege* principal, which decides under the
  user's policy (known endpoints are allowed automatically; new ones wake the user).
  On a grant it traps to the **kernel to mint**. In short, privilege decides and the
  kernel creates.
- A grant that has to wait parks like any cut.
- The capability it produces is an **entity**, so its replies are **checked** against
  the recovered `ConnectionContract`.

### v11: cold path and hot path (08-24 to 08-26)
**Split:** v10 had fused the **Body** (compiled, static, content-addressed) with the
**Context** (live, per run); v11 separates them. `Materialize` is the loader between
the two.

**The two paths:**
- **COLD:** a first request goes through the kernel. THINK compiles a Body, the
  kernel provisions it and mints a capability, it runs once, and it is remembered.
- **HOT:** later requests present the held capability, which the CTU **verifies
  offline** and runs with no call to the kernel.

The model is described as *"OAuth + macaroons at the request level."* The parent move's
`MoveID` becomes the **fd**, and attenuation becomes literal removal: a cut carries
only the capabilities it matches.

## 4. Phase 2: the cold path, and the loop that wouldn't close (v12 to ctu-engine)

This is the least productive stretch of the era, and it is honest about it.

- **v12 (08-27): interfaces only.** It proposes the cold path as
  COMPILE → **BOUND** (the kernel creates the whole workflow's horizon) → MINT →
  materialize and run → REMEMBER, with an `engram/` disk tier and a thin `boot/`.
  There are no function bodies outside a few in `ctu/` and no tests. Its `gaps.md`
  (09-01) concludes it *"cannot be implemented literally while the current surfaces
  are both frozen and exhaustive."*
- **v13 (~09-01): blocked.** It builds the registry, controller validation and
  dispatch reporting, and CTU wiring within the frozen surface.
  - `Scheduler.Write` rejects everything.
  - `selectMoves` **panics** ("eligibility and lifecycle require decisions").
  - The echo proof is never demonstrated.

  Its boundary report is precise about why: `Write`/`run` have no carrier for
  `context`, there are no fixture drivers, and there is no rule for routing a return.
- **v14 (~09-01 to 09-06): the first loop that runs after v11.** It is a single-wave
  local Invoke/Return runtime, split into `protocol`, `driver`, `controller`,
  `scheduler`, `ctu` and `probe`.
  - `Write` updates, selects and reserves work under the scheduler lock; `run`
    dispatches after unlocking, and callbacks re-enter `Write`.
  - A root declares two echo children, they return out of order, and the parent
    resumes exactly once.

  This package shape survives all the way to DNA-19.
- **skyra-ctu-engine (09-06 to 09-07): a partial probe.** It proposes that
  **the fd is the MoveID, and the MoveID is the EngramID**:
  - there is one Memory fd table;
  - a register is a return frame that collects ordered child observations;
  - `Fire(batch)` is the handoff.

  Boot, publishing children in order, and claim/complete transitions all run, but
  admission, retirement and fault parking are `panic("PROPOSAL: …")`.
  `surface-proposal.md` lists the changes it would need; none were applied.

## 5. Phase 3: ownership, then a real process boundary (15 to 19)

The generation names drop the `v`, and the question changes from "what is the
machine?" to **"who is the sole owner of each piece of state, and what is anyone else
allowed to see?"**

### DNA-15 (~09-08)
- It combines v14's transition loop with **kernel-owned registration**.
- **Transition** handles event application, assessment and reservation under one
  mutex; dispatch, validators and notification all run outside it.
- It adds source and sequence binding, placement-gated handoffs, envelopes that put
  the runtime first, and authorization stamps.
- The failure rules are approved and written down. For example, a denied output is
  consumed with no effects, and its producer and root fault with exactly one
  notification.
- `issues.md` records **two real races**, both reproduced and fixed with permanent
  regression tests. One is "buffered child output repeated in an envelope can strand
  admitted children".

### DNA-16 (09-12): see [the 09-12 summary](2026-09-12-dna-16-and-current-direction.md)
- It hosts a **long-lived external driver** (the Codex App Server) without leaking
  its protocol into the core.
- Codex threads, turns and pending requests live in a versioned, opaque
  `DriverState` that the scheduler stores but can't read.
- Codex dynamic tool calls become **ordinary child Invokes**.
- A native tool call that has effects **contaminates** the session and faults it.
- It adds JSONL transport, a process host, and a rule that recreating a host with live
  state faults rather than guessing.
- **Only a fake App Server peer was used; there was no live inference.**

### DNA-17 (09-19): see [the 09-20 summary](2026-09-20-dna-17-and-the-next-direction.md)
This is "the collapse", greenfield with no imports from DNA-16:
- **one door:** `Interpret`;
- **decode:** an opcode jump table;
- **the handler reads and decides**;
- **`StateExecutor.InterpretTurn`:** one atomic turn under the **sole writer**'s lock;
- **concurrent `Dispatch` workers**;
- **re-entry through the same door.**

It uses fd-style `Identifier`s minted at commit, whose scope and liveness live only in
the writer's table. The finding: the request/response spine held, and the fd
settlement fence rejected contaminated sessions *structurally*. Proof was against a
fake peer only.

### DNA-18 (~09-20): the root boundary becomes real
- **The echo is a separate, persistent OS process,** and Moves are serialized as
  JSON over pipes.
- Startup takes exactly one strict JSON argument.
- The owners that DNA-19 inherits appear here: Kernel, RuntimeState, Memory,
  Active, Scheduler, ExecutionContext and Bookkeeper.
- There are 11 acceptance scenarios, for example:
  - forged ownership;
  - two contexts sharing "virtual port 0" in isolation;
  - stale output and duplicate completion;
  - one pending child with correlation;
  - cancellation;
  - malformed startup with no side effects.

### DNA-18.5 (~09-21)
It adds **cross-context activation** (root → B → root, through B's own horizon) and
**capability-mediated creation**: a new executable is minted and then invoked, with
denied templates and a limit of one mint. It also adds a root sink that is distinct
from egress, so bad output can't settle.

### DNA-19 (09-22): see [the two 09-22 summaries](2026-09-22-dna-19-the-permanent-base-and-the-external-binary-boundary.md)
- **The untrusted provider moves out of the runtime into its own `codex` binary.**
  The two binaries share only `protocol/` (wire types plus declarative profile
  validation), and neither imports the other.
- **Authority is an HMAC-signed parent reference** (boot namespace, owning context,
  sequence, binding generation). A bare sequence or a provider thread ID grants
  nothing.
- **It is kept.** It becomes the permanent base.
- **A `user/` executable** (browser over WebSocket → user → root → codex/echo) is then
  added as a second peer registered with root.
- **The working tree splits `root/`** into eight packages with one-way imports:
  boot, kernel, state, execution, extension, ports, trace and demos.
- **An opt-in live run** against the real `codex app-server` passed.

## 6. Ideas that lasted, and ideas that were dropped

**Survived into DNA-19:**
- **Custody and correlation.** Every request carries a correlation ID; the response
  is admitted only against the frame that expects it; late or duplicate returns are
  rejected. This runs from v7's `Expect` ledger to DNA-19's Bookkeeper frames.
- **Trust comes from registration, not from a flag.** v7's host, entity and kernel
  resource types become DNA-19's "a binding contains only its admitted profiles".
- **Output from untrusted code is checked; output from trusted code is believed.**
  This distinction holds from v7 to the end.
- **Only an owner can create authority.** Kernel `Define` (v7), then `Mint` (v9),
  then "privilege decides, kernel creates" (v10), then `MintParent` with HMAC refs
  (DNA-19).
- **The fd is the execution identity, and its liveness lives with the writer**
  (v11, then the ctu-engine, then DNA-17, then DNA-19).
- **Opaque driver state plus a contamination fence** (DNA-16, then DNA-17, then the
  real `codex` binary).
- **The v14 package split:** protocol, driver, controller, scheduler and CTU.

**Dropped along the way:**
- **the CPU metaphor itself:** reorder buffers, register files, retirement watermarks,
  and loops unfolded by tag;
- **distributing a possibility across CTUs** (v8) and gathering and signing its cone;
- **the cold/hot split and macaroon capabilities** (v11–v12);
- **THINK as a compiler of Bodies;**
- **the engram disk tier** (v12).

DNA-19 has **no persistence and no THINK**: it is a transport and authority runtime
with Codex as a peer.

## 7. Things I found reading the code

These come from the code alone; beyond DNA-19's test suite I didn't run anything.

1. **v11's "verified offline" capability is never actually verified.**
   - `Capability` carries a `Sig`, and the README's second proof target is "CTU
     VERIFIES offline".
   - The only `Verify` implementation is `VerifyFunc`, and the only place it is built
     is in `ctu_test.go`, as a function that **always returns `true`**.
   - No signature is ever checked.

   The hot path is proven to *skip the kernel*, not to be *authenticated*. v8 and v9
   do real ed25519 signing, so this is a step back that the v11 README doesn't
   mention.
2. **v7 authenticates by using the PIN as a map key.** `Define` does
   `s.accounts[pin.Secret]`: the plaintext secret is the lookup key, with no hashing
   and no constant-time comparison. It's a throwaway, but it is the only
   authentication code before DNA-19's HMAC references.
3. **Phase 2 produced almost no running code for about ten days.**
   - v12 has no bodies.
   - v13 panics in `selectMoves`.
   - The ctu-engine panics on admission, retirement and fault parking.

   The "freeze the surface, report the gaps" rule was working as designed, but the
   surfaces were being frozen before they were decided. v12's `gaps.md` says so
   directly. The loop only closed again when v14 relaxed that rule.
4. **Until DNA-18, every "Codex" proof was against a fake.** DNA-16 and 17 state this
   clearly. The first real OS process boundary is DNA-18's echo, and the first real
   inference is DNA-19's opt-in live suite.
5. **The same gaps stay open in every generation:**
   - save/resume persistence;
   - crash recovery;
   - reliable delivery;
   - fairness and backpressure;
   - joins over arbitrary numbers of children;
   - OS containment of the provider.

   DNA-18, 18.5 and 19 all list them again. The runtime is correct inside one process
   lifetime and makes no promises across restarts.
6. **Nothing archived builds anymore.** The throwaways still import
   `skyra-v1/src/skyra/skyra-dna-*`, and those paths were moved into
   `archive/throwaway-dna/`. Reproducing any generation's proof means rewriting its
   imports first.

## 8. Where this leaves things

Across 13 throwaway generations the architecture shrank:
- from a distributed, signing, capability-minting CPU with a compiler in front (v8–v12)
- to a single host runtime with a strict owner for every piece of state, a signed
  parent reference as the only authority, and every untrusted thing (the model
  provider, the browser user) in **its own process speaking only Moves** (DNA-19).

The work that remains is what the agent era was about: persistence, a THINK or
planning layer, and learning. They were deliberately kept out of the runtime so it
could be made correct first. The [DNA-19 ownership summary](2026-09-22-dna-19-ownership-split-and-the-generational-direction.md)
proposes adding them back as a recursive generational state machine *on top of* this
base rather than inside it.
