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

### One DNA — competences as namespaced genes (NOT per-agent cells)

There is **one canonical base cell: the DNA.** The v1 dispatcher cells
(`sensory`, `midlevel`, `other`, `self`, `motor`) **collapse into it** — they are **not
separate objects**, they are **competence namespaces** (keys) on the one DNA's competence
map:

```
DNA.competences = {
  sensory:intake,
  other:grade,   other:predict,
  self:grade,    self:predict,
  midlevel:frame,
  motor:collapse,
}
```

`other:grade` and `self:predict` are just genes. The DNA holds **all the durable indexes**
(the entity index, the grouping spine, the action index — the long-term graph) **and all
the competences.** One genome. The "cells" were never separate organisms — they are
*expression contexts* of the one DNA, and the signal (impulse + workflow step) selects
which gene expresses. `other` vs `self` = the genome expressing an other-model gene vs a
self-model gene.

### The ephemeral root — the transcript (mRNA)

A per-wave **ephemeral root**: a `Space` that is a thin transcript over the DNA, created
at wave start, **discarded when the wave completes** (mRNA: a transcript of *one selected
gene*, not a copy of the genome — reference + select, never copy).

- `Observe()` / `Express()` → run the **selected competence** (`root.competence`); the
  bookends.
- `Children()` / `NewActivation()` → delegate to the DNA (transparent for descent).
- `Slice()` → the relevant **namespace's active-state slice** (below), *not* a per-wave
  slice and *not* the DNA — the kernel builds its tree there.
- **per-wave mutable state lives HERE** (the routing decision, `reframed`, the verdict,
  the parse-condition, the impulseID it carries). **The impulse stays pure** — a clean
  external signal + correlation ID; it does **not** carry routing state. Cross-actor
  handoff (e.g. a verdict to the next step) is **stamped on the next root at creation**,
  or lives on the coordination board — never on the impulse.

**Kernel: zero changes** — it already treats `traversed` as the bookend and discards its
per-wave `Node` frame; this just makes the root ephemeral too.

**This is substrate 3/3.** Per-wave state on the transcript — unique per wave, thrown
away — not on a shared object. Scenario 4 dissolves **by construction** (two concurrent
waves = two transcripts), the way cells do parallelism: keep the DNA read-only, spin off
disposable mRNA. **The invariant that makes this hold: competences are stateless
transforms; ALL durable state lives on the DNA. The moment a competence stashes durable
state on itself, the protection leaks** — this is what lets behavior churn freely (swap
competences, reorder the pipeline, ditch the orchestrator) without ever endangering the
accumulated experience.

### The four things (the whole architecture)

| Thing | Count | Job |
|---|---|---|
| **the DNA / nucleus** | one | the **master graph** (all indexes) + competences (namespaced genes). Held *on the membrane* — the membrane is the nucleus. |
| **the coordination board** | one | impulse-node rows (the join + trace) |
| **the orchestrator / membrane** | one | holds the nucleus; reads workflow def + board → **transcribes workers, handing each its working-set references** → spawns |
| **ephemeral workers (transcripts)** | many, transient | each carries a **working set** (handed references into the master graph) + per-wave state; runs one gene; discarded |

### The nucleus model — indexes on the membrane, workers get handed a working set

(Resolved 2026-07-03.) **The indexes ARE the graph, and the graph lives on the
membrane — the nucleus.** There is *not* a set of per-namespace slices held elsewhere.
Instead:

- **The membrane (nucleus) holds the one master graph** (entity index, grouping spine,
  action index; later the crystallized layer).
- **Transcription hands the worker references** to the specific nodes it needs — its
  **working set** ("its own special index it does work on"). The worker operates on those
  handed references; **writes flow back through to the master graph** (shared by
  reference). This is the mRNA discipline: the transcript doesn't reach back into the
  nucleus, it **carries out what it needs.**
- **The entity index rides on *every* worker** (it's traversed every wave). Other indexes
  are handed per-step (e.g. `midlevel:frame` just produced the grouping → the fan-out's
  predict workers are handed that grouping).
- So "the worker's active-state slice" = its handed working set. It is still **read live,
  never snapshotted** — a competence reads *what is active now*, not "what was active when
  the impulse entered." (If state moved since, you *want* the newer value; a frozen view
  would be the bug. This *is* the lagged-reliance model — "self reads other's *latest*
  distribution" = read-active-state — and the "active state is the last observation"
  invariant.) What remains is only **data-race safety** (RWMutex + per-record locks on the
  master graph), *not* semantic snapshotting.

**Working memory vs. durable, on the nucleus.** The master graph holds both the
**working memory** (active groupings, predictions, current entities — churns, and is
almost everything at cold-start) and the **durable crystallized layer** (setpoints,
understandings, stamped entities — empty at cold-start, grows with accumulation). The
"all durable state on the DNA" invariant means the *crystallized* layer; the protection
it buys is against **code churn** (swap competences freely), not against working memory
churning — that churn *is* cognition.

### The impulse has two faces (impulse stays pure)

1. **In-flight** (riding a wave): a **pure external signal + correlation ID** (the
   impulseID). It does *not* carry routing state — that's on the transcript.
2. **As a node** (on the coordination board): the durable **transaction record** — its
   declared grouping set, its self/other completion, its append-only trace, linked to its
   causal chain. Same impulseID threads the two faces.

### The impulse has two faces (impulse stays pure)

1. **In-flight** (riding a wave): a **pure external signal + correlation ID** (the
   impulseID). It does *not* carry routing state — that's on the transcript.
2. **As a node** (on the coordination board): the durable **transaction record** — its
   declared grouping set, its self/other completion, its append-only trace, linked to its
   causal chain. Same impulseID threads the two faces.

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
| an activity / task | an **ephemeral root** (transcript — one step of work) |
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

### Two tables (different kinds — don't conflate)

- **The workflow definition — STATIC.** The routing DAG `(step, condition) → next`.
  Fixed, small, defines the process shape. A **Go map on the orchestrator** — config, the
  schema. **Not in the slice** (static routing rules in runtime data is a smell). (Far
  future: a being rewriting its own definition — the "programmable brain wiring" seed;
  not now.)
- **The job board — DYNAMIC.** One row per in-flight impulse, keyed by impulseID. The
  noSQL-esk table **in the coordination slice** — the runtime state.

### The workflow definition (the corrected DAG)

**Grade fronts the stroke** — the grade is the *intake*, and its outcome routes. Sensory
forks **three ways** on what the impulse *is*; each of `other`/`self` is a `{grade,
predict}` pair of competences (grade before framing, predict in the fan-out):

| step | condition | → next |
|---|---|---|
| sensory | **external** (another agent) | other:grade |
| sensory | **reafference** (world reply to my action) | self:grade |
| sensory | **internal** (reframe) | midlevel |
| other:grade | all preds < 0 (total miss) | midlevel (reframe) |
| other:grade | held / no-prior (cold start) | midlevel (frame) |
| self:grade | exec = no (hard fail) | midlevel (reframe) |
| self:grade | exec ok / partial | midlevel (frame) |
| midlevel | framed | **fan-out** {other:predict, self:predict} + declare `expects` |
| midlevel | empty | done |
| other:predict | done | mark `other` done |
| self:predict | done | mark `self` done |
| **join** | `done ⊇ expects` | motor |
| motor | emitted | done → flush |

- **The gist is the traversal seed** (2026-07-03). `midlevel:frame` produces a gist; the
  predict genes (`other:predict`, `self:predict`) **traverse from the gist's entities**
  through their handed entity index — the warm graph *pulls toward* related nodes
  (associative recall). Cold-start it's trivial (empty graph, nothing to pull); it enriches
  as experience accumulates. This *is* the thing the accumulation experiment tests — wire
  it and watch ("we'll see").
- **Two symmetric grade lanes** = stroke 1 (evidence) vs stroke 2 (reafference):
  external→`other:grade` (did the *other's* move come true), reafference→`self:grade`
  (did *my* action execute). **Motor emitting is what creates the reafference impulse**
  that fronts `self:grade` — closing the loop. Perceive-returns are reafference too.
- **The reafference loop is the self-driven navigation loop** (move→perceive→move…),
  bounded by `midlevel:empty` (nothing left to act on). The workflow *predicts* the maze
  behavior we already observed — good sign.
- "all preds < 0" is the **graded** total-miss (every branch's alignment negative), a
  threshold on the distribution — richer than the current binary `<match>none</match>`;
  a small behavior refinement, noted.

**The condition vocabulary** — the finite alphabet the orchestrator routes on (what a
competence reports): `external · reafference · internal · all-preds<0 · held · no-prior ·
exec=no · exec-ok · framed · empty · predicted · emitted`.

### The job board row (schema)

```
impulseID   string
status      pending | done
expects     set{(grouping, agent)}   // midlevel declares; MVP = {(G,self),(G,other)}
done        set{(grouping, agent)}   // agents mark
reframed    bool                     // the one-reframe rail
motorFired  bool                     // the atomic join guard (below)
history     []span  (append-only)    // the trace — flush on done, lock-free reads
prev        impulseID                // causal link (reafference chain)
```

- **MVP collapses the grouping nesting:** cold-start = one grouping/impulse, so `expects
  = {(G,self),(G,other)}` and the join is "both agents done." The set-shaped schema
  already supports warm N-grouping; we only populate one for now.
- **The join fire must be atomic.** Both agents complete concurrently, both return to the
  orchestrator, both see `done ⊇ expects`. So "fire motor" is a **check-and-set under the
  row lock**: whoever sees `satisfied AND !motorFired` sets `motorFired` and spawns motor;
  the other sees it set and does nothing. Same row lock as the append — no new machinery.
  Without it, motor fires twice.

### Why it's not a bottleneck

Membrane's work is a **board read + a spawn — no inference**. The 30s cost stays in the
workers, which run concurrently. The orchestrator is the conductor, not a player: a few
nanoseconds of coordination per step buys the join and a single routing point. The
extra hop (`worker → membrane → next`) is the price, and it's cheap.

### Source of truth

Step/process state may be *mirrored* on the ephemeral root (its working context) but the
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
  the being's "what can fire" is the DNA's competence map (namespaced genes); routing selects among
  them. The inner-world primitives (match/edit/new) and the pipeline routing become the
  same dispatch shape.
- **Path 2 alignment (free).** impulseID-as-correlation-ID + linked impulse-node chains
  is exactly a distributed trace; the join works the same whether a wave's action stayed
  local or crossed to another machine.

## Scope / risk / build approach

A **big shift**, not an in-place refactor. Decisions (2026-07-03):

- **Fresh package `src/skyra/skyra-dna-v2/`.** The v1 dir stays as reference (patterns,
  method bodies) and as the **acceptance oracle** — v2 is not "done" until it produces
  the same `cmd/validate` live trace (reframe on the pivot, verdict on the match, one
  action per impulse). Keep v1 wired to `cmd/tui`/`cmd/validate`; v2 gets `cmd/validate-v2`.
- **Port, don't rewrite** — the rewrite graveyard is real ("rebuild for months, never
  match the old behavior"). *Rebuilt in v2:* the structural shell (DNA, ephemeral root,
  orchestrator, board, workflow def). *Ported near-verbatim as competence bodies:* the
  grading judge prompts, the reframe rail, motor's retry-with-correction, the self/other
  predict logic — validated live, re-signatured to `(root, origin, impulse)`. *Imported
  unchanged:* `MemoryCell`, `parse.go`, the inference seam, the world, `Route`/`Prediction`.
- **The guard through all of it:** competences are **stateless transforms; all durable
  state on the DNA.** This is the whole point — it lets behavior churn (which we *will*
  do, even after the first accumulation run) without endangering the accumulated
  experience. That protection is *why* this refactor is worth doing before the long
  experiment, not scaffolding-for-later.
- Spread-before-collapse: the workflow def + competence set is small (a handful of
  steps/conditions); build it concrete, let generality emerge only if a second being
  needs different wiring.

## Migration

### The structural shell — DONE (in `src/skyra/skyra-dna-v2/`, tested w/ stubs, -race clean)

- **Transcript core** (`d4834b0`) — `DNA` (competences + the nucleus's indexes),
  `EphemeralRoot` (mRNA), `Competence`. 64 concurrent transcripts isolate per-wave state
  by construction (substrate 3/3 at the mechanism level).
- **The board** (`0bf0edc`) — rows by impulseID, atomic join (motor fires once under the
  fan-out race), one-reframe rail, append-only trace, causal chain.
- **The orchestrator + workflow** (`264b9dd`) — membrane as the workflow engine; per-wave
  `membraneRoot` transcribes the worker as its child (model A within), re-invokes itself
  between steps (model B between); the static workflow DAG drives it. Stub competences
  prove the full pipeline routes, motor fires once, the reframe rail holds.

### Porting the competences — the remaining work (two phases)

**Phase 1 — cold-start turn 1 (real inference, real action, no grading).** On turn 1
there is no prior bet, so grading is trivial. Port each gene's cold path:
- `sensory:intake` — resolve the sender → entity index (the handed working set, on the
  nucleus); classify external/reafference/internal → turn 1 = `external`.
- `other:grade` — no prior distribution → `no-prior` (**no inference**).
- `midlevel:frame` — assemble the frame from the entity index → **infer the gist** →
  create the grouping + seed its tokens → `framed`.
- `other:predict` — **traverse from the gist** (handed entity index) → **infer the
  distribution** → append to the grouping → `predicted`.
- `self:predict` — read other's distribution → **infer the plan** → append → `predicted`.
- `motor:collapse` — read self's plan → **infer the concrete primitive** (with retry) →
  emit → `emitted`.

Checkpoint: `cmd/validate-v2` turn 1 ≈ v1 turn 1 (gist + distribution + action, ~4 real
calls). This de-risks the whole skeleton against real Codex with the simplest cognition.

**Phase 2 — grading + reframe (turn 2+).** Port the real bodies (already validated in v1):
`other:grade` (the adversarial judge → `held`/`all-preds<0`), `self:grade` (reafference
execution), the reframe conditions. The orchestrator *already* routes reframe/join — these
just report the real conditions. Checkpoint: full 3-turn `cmd/validate-v2` ≈ v1 (reframe
on the pivot, verdict on the match).

**Then:** wire `cmd/validate-v2`; keep v1 as the acceptance oracle; flip the default and
retire v1 only once v2 matches the live trace.

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
