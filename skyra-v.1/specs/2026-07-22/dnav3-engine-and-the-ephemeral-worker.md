# The dnav3 Engine — the Triple-Nested Four-Loop

**Date:** 2026-07-22
**Status:** DESCRIPTIVE. Behavior read off `src/skyra/cognitive-agent/skyra-dna-v3` on branch
`slice-based-dispatch` at `9693df3` plus uncommitted work in the tree (prediction fusion,
memory selection, and the `Closures` metric landed in `being.go` at 05:15 on the same day).
No code changed by this doc. The correspondence section is a PROPOSAL and is marked as
such.

**Line references are to that working tree and will drift.** They were verified at write
time; treat a mismatch as staleness in this doc, not in the code.

**Reads on:** `src/skyra/cognitive-agent/skyra-dna-v3/doc.go` (the laws),
[slice-based-dispatch](../2026-07-03/slice-based-dispatch.md) and
[dispatch-model](../2026-07-03/dispatch-model.md) (the v2 pattern this proposes to reuse),
[traversal-implementation-plan](../../design-docs/kernel/traversal-implementation-plan.md)
(Phase 3 is the collapse this enables),
[descent-ascent](../../design-docs/kernel/descent-ascent.md),
[perception-invariant](../../design-docs/laws/invariants/perception-invariant.md).

---

## Why this doc exists

v3's engine is three nested `for` loops in `being.go`, with three hand-set bounds and its
loop-carried state in local variables. It works, and the behavior is correct. But the
structure is invisible: you cannot point at the loop, you cannot inspect a run mid-flight,
and you cannot mount a second world into it.

v2 already solved this shape once, with the ephemeral-worker pattern: a durable genome, a
disposable per-wave transcript, a job board holding the coordination state, and a static
transition table doing the routing. That pattern was archived along with v2's cognition —
but the cognition is what was wrong with v2, not the dispatch.

This doc writes down what v3 actually does, in a form that makes the reuse decidable.

---

## 1. The engine

`Being.Run` is three nested loops. Each level is the same four-position cycle at a
different granularity.

```text
Run(target, maxPasses)                                    being.go:370
│
└─ for cycle < maxReconcileCycles (6)          ── L1 RECONCILE ── target granularity
   │                                                      being.go:383
   ├─ plan(target, maxPasses, c)                          being.go:425
   │  └─ for pass < maxPasses (6, caller-supplied)  ── L2 CLOSURE ── plan granularity
   │     │                                               being.go:429
   │     └─ present → Infer → Parse → Validate → Stamp → Place
   │
   └─ express(plan, c)                                    being.go:509
      └─ for round < maxExpressRounds (8)      ── L3 ASCENT ── act granularity
         │                                                being.go:516
         ├─ for e in plan
         │  └─ for a in e.Runnable(resolve)  →  World.Do(a)
         └─ Settle()
```

Three bounds: **6 / 6 / 8**. All three are magic numbers, and two of them are named
constants whose comments admit they are rails rather than semantics
(`maxReconcileCycles`: "a being that cannot reach the target in this many cycles has hit a
wall"; `maxExpressRounds`: "a plan needing more rounds than this is almost certainly
cyclic").

### The four positions, per level

The cycle in `doc.go` is `observation → state change → action → observation`. It appears
at all three levels, unchanged in shape:

| | L1 RECONCILE | L2 CLOSURE | L3 ASCENT |
|---|---|---|---|
| **observation** | `targetMet(target)` | `present(target, holes)` | `Obs.Resolve` on each `Requires` |
| **state change** | `plan()` returns a closed plan | `Infer` emits a framing + bets | `Runnable` selects the fireable acts |
| **action** | `express()` | `Parse` → `Validate` → `Stamp` → `Place` | `World.Do(a)` |
| **observation** | `targetMet(target)` | `Closed(plan)` → holes | effects appended, then `Settle()` |
| **exits on** | target observed true | zero holes | nothing fired this round |
| **rail** | 6 cycles | 6 passes + fixpoint on `len(holes)` | 8 rounds + `fired` set |

L1's closing observation is L1's opening observation on the next turn. That is the loop
being circular rather than terminating, and it is why `targetMet` is checked twice per
`Run` — once at the top of the cycle, once after the loop exits.

### What each level costs

**L2 is the only level that spends money.** One `Infer` call per pass, and nothing else in
the system calls a model. L1 and L3 are arithmetic and syscalls.

That asymmetry is the thesis: `PassesToClosure` counts L2 iterations, and the whole design
is aimed at making L1 and L3 do enough work that L2 runs fewer times.

**Note (same-day change).** `PassesToClosure` is now appended **only when the target was
met**, alongside a new `Metrics.Closures` counter — so the curve is over *successful* runs
and `Closures/Cycles` is a separate closure rate. Previously every `Run` appended,
successful or not. This is the right shape (a run that never closed has no
passes-to-closure), and it means the pre-change and post-change curves are not directly
comparable. Any conversion regression test must be built against the new definition.

### Level detail

**L1 — RECONCILE** (`being.go:383`). The watcher. Plan, act, observe, ask whether the
target is met; if not, re-plan against what was just learned. This is what lets the being
discover a prerequisite rather than merely learn a probability about failing: a bare build
closes structurally, fails, reports `build:/ missing app.src`, and the next pass sees it.
A single-shot `Run` would plan once, fail once, and stop.

**L2 — CLOSURE** (`being.go:429`). Emit until the plan's `Requires` are all grounded —
either observed true, or produced by an `expect` of another act in the same plan.
Grounding is set membership (`Closed`, `being.go:339`): no graded overlap, no threshold,
no judge. The loop carries `plan []*Expression` across passes and accumulates; a pass that
adds nothing trips the fixpoint rail.

**L3 — ASCENT** (`being.go:516`). Fire → settle → fire. This is a loop for a reason
recorded in the code: an act guarded on `file:/x exists true` cannot fire in the same
round as the write that creates the file, because **the write does not observe the file**.
The world reports `written`; existence must be perceived. That is LAW 1 working — so the
rhythm `fire → bets mature → Settle (go and look) → guards resolve → fire the unblocked`
is forced by the laws, not chosen for convenience.

---

## 2. Loop-carried state — the part that wants a Board

Everything the three loops carry lives in local variables and dies with the call. This is
the table that motivates the whole reuse proposal.

| State | Where it lives now | What it does | v2 counterpart |
|---|---|---|---|
| `plan []*Expression` | local to `plan()` | the accumulating plan across L2 passes | `Row` working set (`SetEpisode`/`Episode`) |
| `lastHoles int` | local to `plan()` | fixpoint rail — stop if a pass made no progress | `TryReframe` (the one-reframe rail) |
| `fired map[string]bool` | local to `express()` | an act fires ONCE; not idempotent, must not replay | `TryFireMotor` (the atomic once-only join) |
| `c *Cycle` | passed by pointer through all three | Passes, Acts, Rejects — the run's trace | `Row.history []Span` (append-only) |
| `holes []Triple` | recomputed each pass | the L2 condition | a worker's `Condition` string |
| `b.tick` | `Being` field | subjective clock; `Within` maturity counts in it | — (no counterpart; see §5c) |
| `b.currentSpan` | `Being` field | occurrence identity for grading eligibility | `Row.ImpulseID` |
| `b.currentBornFrom` | `Being` field | the evidence that caused this emission | `Row.Prev` (the reafference chain) |

Two of these — `fired` and `lastHoles` — are **rails that v2 implemented as board
primitives with exactly the same semantics**, arrived at independently. `fired` is
`TryFireMotor`: check-and-set, true to exactly one caller. `lastHoles` is `TryReframe`:
allowed the first time, suppressed after. That convergence is the strongest single
argument that the pattern fits.

---

## 3. What v2's ephemeral worker actually was

Four parts (`archive/_skyra-dna-v2/skyra-dna-v2/`):

**`DNA`** — the durable genome. Read-only. A registry of competences keyed
`namespace:gene`. Never copied.

**`EphemeralRoot`** (`ephemeral.go`) — the transcript. mRNA. A per-wave `skyra.Space`
created at wave start and discarded at completion. It references the DNA, selects one
observe gene and one express gene as the kernel's bookends, and holds **all** per-wave
mutable state — never on the DNA, never on the impulse. Because it is unique per wave and
thrown away, concurrent waves cannot clobber each other's state. Parallelism the way cells
do it: keep the genome read-only, spin off disposable transcripts.

**`Board`** (`board.go`) — one `Row` per in-flight impulse, keyed by impulse ID. Holds the
join set (`expects`/`done`), the rails (`TryReframe`, `TryFireMotor`), the append-only
`[]Span` trace, the causal link to the impulse that provoked this one, and the working-set
handoffs between steps. Board-level lock guards only row creation; each row carries its
own lock, never held across inference.

**`workflow.go`** — the static transition table. `(step, condition) → Route{Next | Reframe
| Done | Wait}`. Config, not runtime data. A missing entry is an inert edge.

And the engine that runs them, `Orchestrator` (`membrane.go`): per wave it is a
`membraneRoot`. `Observe` transcribes the worker for this step and exposes it as a child;
the kernel descends into it; `Express` reads the worker's reported `Condition`, appends the
span, looks up `routeFor(step, cond)`, and dispatches the next step by **re-invoking
itself** — `go o.kernel.Realize(m, m, impulse)`.

> **Iteration in v2 is a chain of waves, not a `for`. The loop structure is data.**

That is the sentence this whole document is pointing at.

---

## 4. The correspondence (PROPOSAL)

| v3 today | v2 ephemeral worker |
|---|---|
| three nested `for` loops | one orchestrator re-invoking itself through a transition table |
| `maxReconcileCycles=6`, `maxPasses=6`, `maxExpressRounds=8` | traversal budget + affordability (`traverse.go`) |
| `holes`, `Runnable`, `targetMet` as `if`s | `Condition` strings routed by `transition` |
| local vars per call | `Row` (durable per-run) + `EphemeralRoot` (per-wave scratch) |
| `b.Infer(b.System, present)` inline in `plan()` | one gene on one transcript |
| `b.World.Do(a)` inline in `express()` | a leaf worker whose Express is the outcome |
| `c *Cycle` counters | `Row.history []Span`, flushed on done |
| `EpisodeID` / `SpanID` | `Row.ImpulseID` / `Row.Prev` |
| — | a run you can inspect while it is in flight |

The payoff beyond legibility: **a device-place becomes expressible.** `Realm` is flat —
`Do` and `Perceive`, two methods — and cannot represent "the Mac is a place containing
spaces containing affordances" without smuggling hierarchy into target strings. A worker
tree can. Mounting a partition of a device is adding a child; relocation is re-rooting.

---

## 5. What does NOT map

This is the part that decides whether the reuse is cheap or a rewrite.

### a) The grader must not become a routed STEP — LAW 4

**Not: the grader must not be a worker.** A gene is `func(*EphemeralRoot, skyra.Space,
any)`. Nothing in the mechanism requires a model — `genes_ingest.go` is a pure gene with no
`Infer` call anywhere in it. "Worker" and "inference call" are independent, and conflating
them is a mistake this doc made in its first draft.

What *is* true historically: v2's two grade genes did call the model — `genes_final.go:44`,
`genes_efficient.go:35`. That was v2's **judge**, and it is exactly what v3 was built to
delete. But that was a property of those two implementations, not of the gene contract.

The concern that survives is about **call site, not worker-ness**.

Today `grade(o)` is invoked unconditionally from inside `Observe()` — `being.go:184` and
`being.go:194`. Every observation write grades. No route is consulted, no condition string
is matched, no table entry has to exist.

v2's routing semantics are the opposite: `routeFor` returns `ok=false` on a missing entry,
and the orchestrator treats that as an **inert edge — nothing dispatched**. That is correct
for cognition steps (an unhandled condition should stall, not improvise) and *silently
wrong* for grading. A missing or mistyped condition string on the grade path is an
ungraded observation, with no error and no trace entry. LAW 5's "every bet gets settled"
would come to depend on table completeness.

> **The rule.** `grade`, `gradeInterpretations`, and `compareExactExpectation` may be
> workers. They may not be **routed steps**. They are folds the observation write path
> invokes directly — reached because an observation landed, never because a table said so.
> A worker may *cause* an observation; it may not be *scheduled to adjudicate* one.

Workers do not have to be dispatched by the transition table to be workers. That
distinction is what keeps the pattern reusable here.

Two structural hygiene notes, cheap to honour at build time:

- `Infer` sits on `EphemeralRoot` itself (`ephemeral.go:60`), so every transcript carries
  the seam whether or not it uses it. Grading workers should be transcribed onto a type
  that does not have the method — capability, not discipline.
- Grading must remain unconditional on the write path, so it needs no `Condition` and
  should report none. A worker that returns no condition cannot be routed by construction.

### b) Void `Realize` — the problem is COMPLETION, not ordering

The ordering objection dissolves; a different one is real and silent.

**Ordering is fine.** `express()` looks like it needs a synchronous return:

```go
out := b.World.Do(a)          // being.go:528
e.Complete(a.ID, b.tick)      // arm predictions BEFORE appending effects
for _, t := range out.Effects { b.observeInSpan(t, ProvenanceRaw, e.SpanID) }
```

But v2's asynchrony is **between steps, not within a wave**. `membraneRoot.Observe` builds
the worker as a child; the kernel descends; `Express` runs post-order and reads
`m.worker.Condition` *synchronously*. Only the next step is spawned with `go`.

So a local act maps directly: the act is a child Space, its `Express` yields the `Outcome`,
the parent folds it post-order, and arm-then-append keeps its relative order inside that
fold. The ordering becomes structural — outbound edge arms, return edge observes — which is
*better* than enforcing it by statement order.

**The real problem is that completion stops being guaranteed.**

```text
World.Do returns  →  e.Complete(actID, tick)  →  x.Arm(tick)
                                              →  MaturesAt = tick + Within

Matured()  requires  x.Armed        expression.go:289
Settle()   skips     !Matured       being.go:315
```

Under a synchronous `World.Do`, **every act completes — because the function returned.**
That guarantee is doing invisible work, and nothing in v3 states it.

Under void dispatch a remote act can dangle. Its expects are never armed, so they never
mature, so `Settle` never sweeps them, so they are never graded. The bet stays open
forever and **nothing notices**, because `Settle` cannot distinguish "not yet armed" from
"not yet due" — both are simply skipped. That is a direct LAW 5 break ("every bet gets
settled"), and it is silent, which is the worst property a break can have.

The fix is already specified — in `ephemeral-world-workers.md`, this project's own doc:
`Deadline` on the Invocation, and `indeterminate` as a terminal status, described there as
*"not failure. Cognition may need to perceive the target again before retrying."*

v3 carries `indeterminate` in the Outcome vocabulary (`world.go:35`) but **nothing produces
it, and there is no deadline anywhere in the package.** That gap costs nothing while the
world is an in-process filesystem whose calls cannot fail to return. It stops being free
the moment a world lives on the other end of a socket.

**Therefore:** an act needs a terminal outcome under all schedules — success, failure, or
`indeterminate` on deadline expiry — and `Complete` must be reachable from the deadline
path, not only from the success path. A bet whose act timed out should arm and settle
against `indeterminate`, not hang.

### c) The tick has no definition, only a location

`b.tick` is a plain `int` on `Being`. It advances in exactly two places, both inside
`express`:

```go
b.tick++;  out := b.World.Do(a)   // being.go:527 — once per act fired
b.tick++;  b.Settle()             // being.go:541 — once per round, before the sweep
```

`Arm` stamps `MaturesAt = tick + x.Within` (`expression.go:259`). So **`Within` is
denominated in ticks, and one tick is currently one world-touching event** — an act firing,
or a settle sweep. An L3 round costs `1 + (acts fired)` ticks. When the model emits
`within: 3` it is saying *"check this about three world-events from now."*

That meaning is nowhere written down. It is an accident of where the two `b.tick++`
statements happen to sit, and it evaporates the moment `express` does.

Wave dispatch has no `express` loop, so the increment needs a new home, and the candidates
are not interchangeable:

| tick site | consequence |
|---|---|
| **per wave / per step** (incl. pure routing hops) | tick inflates. A bet armed at 10 with `Within: 3` matures at 13 — under a 5-steps-per-round engine that is **less than one round later**. It is perceived before the acts that would make it true have fired, comes back false, and grades the being for being wrong about something that had not happened yet. |
| **per act completion + per settle sweep** | identical to today; survives the conversion unchanged. |
| **per traversal depth** | a different quantity entirely; `Within` would no longer mean anything the model could reason about. |

The reason this is not a detail: **every edge EMA in the expression graph was learned from
bets that matured under the current rate.** Change the rate and different bets settle at
different moments → different grades → both `GradeSum` and `PassesToClosure` move. The
engine would have been silently re-parameterized while appearing merely to have been
refactored.

So what is needed is not a number but a **definition**, pinned to a semantic event rather
than to a dispatch event:

> **SUPERSEDED — 2026-07-22.** The invariant proposed here ("the tick advances when the
> being touches the world") is **not** the project's rule. `local-relativity-invariant.md`
> is Foundational and already settles it: *time is proper time — each Slice carries its own
> count, bumped once per wave.* Not motor events, not world-touching. See
> [time](time.md) for the full analysis and what remains open.

Corollary worth noting: under §5b, an act that terminates as `indeterminate` on deadline
has still *touched the world*, so it ticks. Only that way does a timed-out act let its
dependent bets mature rather than stranding them.

v2 has no counterpart to steal here — `Span.At` is a trace index, not a clock.

### d) The Board is flat; v3's nesting is three-deep

`Row` is one record per in-flight impulse. v3 has three concurrent scopes — a target, a
plan, an act — each with its own rails and its own trace. Either `Row` grows nested scopes,
or there are three row types with parent links, or the three levels collapse into one
transition table where the level is part of the step name.

The third is probably right and is the one that recovers the budget story: if all three
levels are steps in one table, the three bounds (6/6/8) become one budget attenuating
through depth, which is exactly `traverse.go`'s arithmetic and exactly Phase 3 of the
traversal plan.

### e) v3 is single-threaded by design

`being.go`'s own header: *"Single-threaded, unbudgeted, no kernel — the wave is an
OPTIMIZATION for a graph that is large, and at this scale brute force is correct."*

v2's board has per-row mutexes because waves are `go`-spawned. Adopting the pattern means
adopting concurrency, and `b.Metrics`, `b.observationSeq`, `b.expressionSeq`, `b.tick`,
`Obs`, and `Expr` are all currently unguarded. `skyra.Slice` has `Lock`/`Unlock` for this,
and nothing in v3 uses a Slice yet.

This is not a reason to refuse — it is a reason to price it.

### f) v2's fan-out join has no v3 counterpart

v2's `Declare`/`Wait`/`TryFireMotor` machinery joins two *causes* (final + efficient
predict) before motor collapse. v3 has no such fan-out: its parallelism is `Runnable`, and
acts join on **observed world state** via `Requires`, not on branch completion.

So roughly half of `Row` — `expects`, `done`, `Declare`, `key(episode, agent)` — is v2
cognition machinery with nothing to do in v3. Take the rails; leave the join.

---

## 6. What this doc does not decide

1. **Proper time** (§5c, superseded) — the rule is `Slice.Count` per the Foundational
   local-relativity invariant. What is open is that dnav3 has no Slice. See [time](time.md).
2. **Act deadlines and who produces `indeterminate`** (§5b) — free today, load-bearing
   the moment a world is remote.
3. Whether the three levels become one transition table or three (§5d).
4. Whether `Realm` survives as the leaf contract under a Space tree, or is replaced.
5. Whether conversion happens before or after macAgentPort exists — i.e. whether the
   sandbox filesystem is the thing carried onto the new engine first, or the Mac is.
6. Whether the thesis measurement (`PassesToClosure`, `Closures/Cycles`, `GradeSum`) is
   preserved bit-for-bit across the conversion. **It should be**: L2 passes are countable
   under either engine, and a conversion that changes the numbers has changed the
   experiment, not the implementation. This is the regression test for the whole
   exercise — and per §1 it must be pinned against the post-05:15 metric definition, on a
   recorded baseline run, *before* any conversion starts. There is nothing to compare
   against afterwards otherwise.
