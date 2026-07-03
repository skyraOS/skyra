# Slice-Based Dispatch — Routing off the Slice, not off Hardcoded Pointers

**Date:** 2026-07-03 · **Status:** in spec (the refactor that makes the dispatch model real) · **Parents:** [dispatch-model.md](dispatch-model.md) (the transition tables this makes executable), [two-stroke-cycle.md](two-stroke-cycle.md) (scenario 4 / [B], the fan-out join), [memory-as-world.md](../../design-docs/cognition/memory-as-world.md) (affordances), [comparator.md](../../design-docs/cognition/comparator.md) (**competence-on-the-wave** — the old pattern this revives for routing)

> **One-line thesis:** the hardcoded successor chain is a *deviation* from a pattern the
> design already had — **competence-on-the-wave** (dumb cells, smart payload). Slice-based
> dispatch = that pattern applied to routing, carried by a **per-wave ephemeral root**,
> with cross-wave joins on an **impulse-keyed coordination slice**.

## The gap this closes

The dispatch-model spec describes routing as `cell reads slice-state →
parse-condition → fires the selected affordance`. **The build does not do this.**
Audited on `accumulation-build` (2026-07-03), the actual routing is:

- **Hardcoded successor pointers.** Each cell holds a direct reference to its
  successor(s) and spawns it in `Express`: membrane→`m.sensory`, sensory→`s.midLevel`,
  MidLevel→`g.other`, other→`o.self` (or `o.midLevel` on reframe), self→`s.motor`
  (or `s.midLevel` on reframe). The topology is wired into the structs at
  construction; nothing is *selected*.
- **Per-wave routing state on CELL FIELDS.** The go/no-go and branch decisions read
  mutable fields on the long-lived cell object — `o.reframed`, `o.verdict`,
  `s.plannedThisWave`. Set in `Observe`, read in `Express` of the same wave.

Two problems fall out:

1. **Not concurrency-safe (scenario 4).** Two waves through the *same* cell clobber
   its per-wave fields (`o.reframed` etc.) — proved by the race detector. This is
   why "substrate 3/3" is *not* a locking problem: the state is on a shared cell,
   so no lock granularity fixes it. It has to *move*.
2. **Not the affordance model.** "Which cell fires next" is fixed wiring, not a
   function of `(parse-condition, slice-state)`. The transition tables in
   [dispatch-model.md](dispatch-model.md) are documentation, not executable.

## The model — competence-on-the-wave meets the ephemeral root

The refactor is **not a new invention** — it is the **old competence pattern**
(comparator.md: "competence travels on the wave — dumb cells, smart payload") applied
to routing, realized through a **per-wave ephemeral root**. The hardcoded chain is the
deviation from a pattern the design already had.

### The competence pattern (recap — the precedent)

A parent launches a wave with its **method attached as payload**; generic cells run
that payload against their own content; the kernel never reads it. It **brackets** the
cell's work (runs first on Observe = context primes, last on Express = audit after the
vote — the onion shape), and it **nests** (a cell throws its own competence to its
members — fractal). Dispatch is just competence that says *where to go next*.

### The ephemeral root — the bookend carrier

A per-wave **ephemeral root**: a `Space` that is a thin decorator over a base DNA cell,
created at wave start, **discarded when the wave completes**.

- `Slice()` / `Children()` / `NewActivation()` → **delegate to the base cell** (the
  real topology + durable substrate stay put; the root is transparent for descent).
- `Observe()` → look up the **selected** `observeFN` from the base cell's map, run it.
- `Express()` → look up the **selected** `expressFN`, run it (the bookend + the routing
  decision).

The base cell holds the **menu** — `observeFN map[key]method`, `expressFN
map[key]method` (the affordances/competences, shared read-only during a wave). The
ephemeral root holds the **choice** for this wave + any per-wave state. **Kernel: zero
changes** — it already treats `traversed` as the bookend and already discards its
per-wave `Node` frame; this just makes the root cell ephemeral too.

**This is substrate 3/3.** Per-wave mutable state (`reframed`, verdict, the routing
decision) lives on the ephemeral root — unique per wave, thrown away — not on the shared
base cell. The scenario-4 race dissolves **by construction**, no locks: two concurrent
waves have two ephemeral roots. The base cell exposes only immutable maps (concurrent
reads, safe) + its slice (already locked).

### The three state homes (the whole architecture)

Pulling apart what used to be jammed into "the slice":

| Home | Job | Scope | Concurrency |
|---|---|---|---|
| **Private per-actor slice** | the kernel builds its tree + the records/EMAs | per cell, durable | the substrate (built) — untouched, separate from dispatch |
| **The impulse, in-flight** | **competence** — where this wave goes next | per-wave | safe by construction (each wave owns its impulse) |
| **The shared coordination slice** | **the join** — impulse nodes + their completion state | one board, cross-wave | the "noSQL board" — locked writes, the emergencies/surprises/join home |

**The dividing line:** the impulse carries what **one wave** does next; the
coordination slice carries what **multiple waves** join on. Forward routing rides the
impulse (single-wave); the fan-out join lives on the board (multi-wave) — because **no
single impulse can express "wait for two waves."**

### The impulse has two faces

1. **In-flight** (riding a wave): carries **competence** — the `(observeFN, expressFN)`
   selection and where to go next. Per-wave forward routing. Replaces the hardcoded
   chain. The ephemeral root reads this to run its bookend and to decide `Express`'s
   spawn.
2. **As a node** (on the coordination slice): the durable **transaction record** — its
   declared grouping set, its self/other completion, linked to its causal chain. Same
   impulseID threads the two faces.

### The join — impulse-keyed, not grouping-keyed

The workqueue is keyed by **impulseID**, because the **impulse is the coherence/action
boundary** while the grouping is only the prediction unit. One external impulse
activates N groupings (independent self/other waves MidLevel cohered on); MidLevel
coheres them into **one** action. So motor fires once per *impulse*, not per grouping.

- **impulseID propagates down the whole chain as a correlation ID** — every spawned
  wave carries the *originating* impulseID unchanged to motor. (This is a
  distributed-tracing trace-ID; path 2 gets cross-machine join for free — the
  correlation ID is transport-agnostic.)
- **MidLevel declares the expected set** so the join knows when it's complete: it
  registers `X → expects {G1, G2}` on the board; the agents mark `X.G1.self`,
  `X.G1.other`, …; when X's declared set is satisfied → cohere/vote → motor fires one
  action for X. Board is impulse-keyed, grouping-nested:

```
impulse X → { expects: {G1, G2}, done: {G1.self, G1.other, G2.other}, → pending }
             (motor fires when done ⊇ expects × {self, other})
```

- **The workqueue is not a separate structure** — it *is* the pending impulse nodes on
  the coordination slice. "In flight" = impulse nodes whose join isn't satisfied; motor
  reads pending nodes, fires on completion, the node flips to done and stays as history.

### Impulse boundaries — new external event = new impulse node

- **Reframes / surprises inherit** the impulseID (still resolving the same external
  event, mid-transaction).
- **The world's reply (reafference, stroke 2) is a NEW external event → a new impulse
  node** on the slice, **linked** to its causal predecessor (the action that provoked
  it). Continuity lives in the *topology* (a walkable chain of impulse nodes — the
  reafference loop as graph), while coordination stays scoped per node. Clean
  boundaries, preserved history, no transaction that lives forever. (Conversation-as-
  topology, notes.md.)

## What this buys

- **Scenario 4 / substrate 3/3 by construction.** Per-wave state on the ephemeral root
  can't be clobbered by a concurrent wave — each wave has its own root. The
  N-same-cell-wave race is gone with **no locking**. (A trigger coalescer is still
  wanted for *cognitive* dedup — not running N full pipelines for N rapid triggers —
  but that is a higher-level concern on the coordination slice, separate from
  data-race safety.)
- **The dispatch model becomes real, and it's the old pattern.** The transition tables
  execute instead of documenting; adding a route is a table row, not spawn code. And it
  is competence-on-the-wave — so it inherits the bracketing (Observe-first/Express-last)
  and the nesting/fractal property for free, neither of which a bespoke router would
  have given us.
- **The motor join is expressible.** "Fire after self AND other" has a home (the
  impulse node's completion row) instead of being smuggled into a hardcoded `self→motor`
  edge. Cross-wave coordination finally has a place that isn't a cell field.
- **Affordances as the vocabulary** ([memory-as-world.md](../../design-docs/cognition/memory-as-world.md)):
  the being's "what can fire" is the base cell's competence maps; routing selects among
  them. The inner-world primitives (match/edit/new) and the pipeline routing become the
  same dispatch shape.
- **Path 2 alignment (free).** impulseID-as-correlation-ID + linked impulse-node chains
  is exactly a distributed trace; the join works the same whether a wave's action stayed
  local or crossed to another machine.

## Scope / risk

A **real refactor**, not additive — it introduces the ephemeral root + competence maps,
the coordination slice, and rewrites construction/routing. Do it *after* the current
working model is preserved (committed + pushed — done). Spread-before-collapse:
the transition table + competence set is small (5 cells, a handful of conditions/
methods); build it concrete for these cells, let generality emerge if a second being
ever needs different wiring.

## Migration (increment order)

1. **Ephemeral root + per-wave state off cell fields** — `reframed`/`plannedThisWave`/
   verdict move onto the per-wave root (or the in-flight impulse it reads). Cells set it
   in Observe, read in Express. Un-skip `TestConcurrentEngagesRapidFire` (scenario 4) —
   green with no locks. *(This is substrate 3/3.)*
2. **Competence maps on the base cell** — `observeFN`/`expressFN` maps; the ephemeral
   root selects and delegates. Behavior unchanged (one entry per map to start).
3. **Affordance registry** — `role → base cell`; keep the current routing but resolve
   the next affordance through the registry (mechanical).
4. **Transition table** — replace per-cell branch logic with `(role, condition) → next`
   data; `Express` (the selected `expressFN`) becomes generic.
5. **The coordination slice + impulse-node join** — impulseID propagates as a
   correlation ID; MidLevel declares the expected grouping set; agents mark completion;
   motor fires per impulse on join. Replaces the hardcoded `self→motor` edge with the
   board read.
6. **Delete the successor pointers** — once transitions + the join drive everything, the
   struct fields (`o.self`, `s.motor`, …) are dead; remove them.

## Open

- **Who selects the competence / makes the choice for the root** — the spawning cell
  (it knows what to hand down), the kernel at Realize, or read off the in-flight impulse.
  Lean: the spawner writes the selection onto the impulse; the kernel builds the
  ephemeral root from `(base cell, impulse's selection)`. Confirm the kernel seam.
- **Impulse mutability across the bookend** — the ephemeral root writes the parse-
  condition in Observe and reads it in Express; confirm the per-wave root (or a pointer
  the impulse carries) is the mutable home, not a value copy.
- **Coordination-slice keying detail** — impulse-keyed, grouping-nested (above). Pin the
  record shape (`expects`/`done` sets) and cleanup (done nodes stay as linked history;
  when are they GC'd from the *pending* view — decay like everything else?).
- **Reframe/self-initiated impulseID inheritance** — reframes inherit; the world's reply
  mints a new linked node. Pin where the link edge is written (on the action node → its
  reafference).
- **Kernel owns the registry/transition, or DNA?** Lean DNA (kernel stays mechanism;
  routing/competence is policy) — the kernel still just delivers waves and discards its
  frame.

## Done when

- No cell holds a successor pointer; the ephemeral root's `Express` fires whatever the
  transition selects from the registry.
- Per-wave routing state is on the ephemeral root, not cell fields;
  `TestConcurrentEngagesRapidFire` (scenario 4) un-skips and is race-clean, no added
  locks.
- Motor fires per **impulseID** on the impulse node's join (self AND other complete),
  not off a hardcoded `self→motor` edge; one action per external impulse even when it
  cohered multiple groupings.
- Adding a route (a new reframe source, a vote branch) is a table row, not spawn code.
- The live two-stroke behavior (reframe on miss, verdict-on-wave, C rule) is unchanged —
  same `cmd/validate` trace, now driven by data.
