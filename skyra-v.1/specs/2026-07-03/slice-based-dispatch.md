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

## The orchestrator — membrane as the workflow engine

Routing is **not** distributed across each cell's `Express` (worker-fires-its-own-next).
It is **centralized** in an orchestrator, and **the join forces this**: no cell's Express
can fire motor on "self AND other done," because self's Express doesn't know other
finished. *Someone* must watch the board and dispatch when a step's dependencies land.
That someone is the orchestrator. The moment the join is taken seriously, distributed
routing is off the table.

**Membrane is the orchestrator** (the boundary is already where waves enter/exit). The
shape is a **workflow engine** (Temporal / Step-Functions): a central dispatcher reading
durable workflow state and firing the next activity.

| Workflow engine | Here |
|---|---|
| the engine / dispatcher | **membrane** |
| workflow state (per run) | the **impulse node** on the coordination slice (job board), keyed by impulseID |
| an activity / task | an **ephemeral cell** (one step of work) |
| the workflow definition | the **transition rules** — live on the orchestrator, not scattered per-cell |
| run id | the **impulseID** (= processID; the impulse stays pure) |

### The orchestration step (one wave = one bracketed step)

```
wave arrives at membrane (impulse carries impulseID X)
  membrane.Observe: look up X on the board → build the ephemeral worker for the
                    next step, hand it down as membrane's child
  kernel descends → worker does its ONE step (its selected observe/express),
                    reports its parse-condition + marks completion on X's board row
  membrane.Express: read X's board row (condition + workflow def) → decide next:
                    - a step ready       → spawn the next membrane-bracketed wave
                    - a fan-out          → spawn TWO (self, other) membrane-bracketed waves
                    - join satisfied     → fire motor (self AND other done)
                    - done               → flip X pending→done, flush hook
```

Membrane is the **always-fires bookend** the transition helper needed: every step is
`membrane(build) → worker(do) → membrane(route)`. The transition rule is a helper the
orchestrator calls on Express — `(reported condition, workflow def) → next`.

### Why it's not a bottleneck

Membrane's work is a **board read + a spawn — no inference**. The 30s cost stays in the
workers, which run concurrently. The orchestrator is the conductor, not a player: a few
nanoseconds of coordination per step buys the join and a single routing point. The
extra hop (`worker → membrane → next`) is the price, and it's cheap.

### Source of truth

Step/process state may be *mirrored* on the ephemeral cell (its working context) but the
**board row (keyed by impulseID) is the source of truth**. If they diverge, the board
wins — one clock.

### Note for later — two roles, co-located

The **orchestrator** role (workflow engine) and the **sensorimotor-membrane** role
(attention gate, motor emission) are conceptually distinct — the-engine vs. the-boundary.
Co-locating on membrane is fine to start; if membrane accretes too much, splitting the
engine out from the boundary is the natural seam.

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
   verdict move onto the per-wave root. Un-skip `TestConcurrentEngagesRapidFire`
   (scenario 4) — green with no locks. *(This is substrate 3/3.)*
2. **Competence maps on the base cell** — `observeFN`/`expressFN` maps; the ephemeral
   root selects and delegates. Behavior unchanged (one entry per map to start).
3. **The coordination slice (job board)** — impulse node per impulseID with
   `expects`/`done` + the append-only history column; impulseID propagates as the
   correlation ID (impulse stays pure). No behavior change yet — just the board exists
   and steps write to it.
4. **Membrane becomes the orchestrator** — its Observe builds the ephemeral worker for
   the next step and hands it down; its Express reads the board (`condition + workflow
   def`) and dispatches. Fold the existing routing into the workflow definition one edge
   at a time (sensory→midlevel→…), verifying `cmd/validate` is unchanged after each.
5. **The join** — MidLevel declares the expected grouping set; agents mark completion;
   membrane fires motor per impulse when the join is satisfied. Replaces the hardcoded
   `self→motor` edge with the board read.
6. **Delete the successor pointers** — once the orchestrator + board drive everything,
   the struct fields (`o.self`, `s.motor`, …) and per-cell branch logic are dead; remove
   them. The transition rules now live only in the workflow definition.

## Open

- **Who selects the competence for the worker** — the **orchestrator (membrane)** picks
  the next step's competence from the workflow definition and builds the worker in its
  Observe. (Supersedes the earlier "spawner writes selection onto the impulse" — the
  orchestrator owns selection now.)
- **Per-wave mutable home across the bookend** — the ephemeral root writes its parse-
  condition in Observe and reports it out; confirm the root (not a value copy) is the
  mutable home, and that its report reaches the orchestrator's Express (via the board row
  or the returned Node).
- **The workflow definition's form** — a data table `(step, condition) → next-step(s)`
  on the orchestrator. Includes fan-out edges (one step → two) and join edges (two
  done → one). Small and concrete for these ~6 steps; not a general engine.
- **The condition vocabulary** — the finite set a worker reports: `match`/`none`
  (grade), `proposal`, `yes`/`partial`/`no` (exec), the framing verbs. This is the
  dispatch-model parse-conditions, now the orchestrator's routing input.
- **Coordination-slice keying detail** — impulse-keyed, grouping-nested. Pin the record
  shape (`expects`/`done` sets + append-only history column, row-locked appends / lock-
  free reads) and cleanup (done nodes stay as linked history; GC'd from the *pending*
  view — decay like everything else?).
- **Orchestrator re-entry cost** — every step round-trips through membrane. Confirmed
  cheap (board read + spawn, no inference), but watch the spawn count under long
  navigation-heavy loops.
- **Reframe/self-initiated impulseID inheritance** — reframes inherit; the world's reply
  mints a new linked node. Pin where the link edge is written (on the action node → its
  reafference).
- **Kernel owns the registry/transition, or DNA?** Lean DNA (kernel stays mechanism;
  routing/competence is policy) — the kernel still just delivers waves and discards its
  frame.

## Done when

- No cell holds a successor pointer; **membrane (the orchestrator)** builds the worker
  and dispatches the next step from the workflow definition + the board.
- Per-wave routing state is on the ephemeral root, not cell fields;
  `TestConcurrentEngagesRapidFire` (scenario 4) un-skips and is race-clean, no added
  locks.
- Motor fires per **impulseID** on the impulse node's join (self AND other complete),
  not off a hardcoded `self→motor` edge; one action per external impulse even when it
  cohered multiple groupings.
- Adding a route (a new reframe source, a vote branch) is a table row, not spawn code.
- The live two-stroke behavior (reframe on miss, verdict-on-wave, C rule) is unchanged —
  same `cmd/validate` trace, now driven by data.
