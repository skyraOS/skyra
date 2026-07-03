# Dispatch Model — Affordances, Parse-Conditions, and Slice-State

**Date:** 2026-07-03 · **Status:** in spec (the structure the build hangs on) · **Parents:** [two-stroke-cycle.md](two-stroke-cycle.md) (the dynamics this makes buildable), [memory-as-world.md](../../design-docs/cognition/memory-as-world.md) (the parked affordance table, now recast) · **Blocks:** the concurrent fan-out build

## Why this exists

The two-stroke and concurrency specs pin the *dynamics* (when things fire, what
races, what overwrites). But you build against *structure*, and the structure was
still prose. This spec is the missing piece: **the transition table** — every
inference call, the discrete conditions its response can parse into, and the exact
state each condition writes. With this, the pipeline is a state machine you can
implement without inventing the data model mid-build.

## The model — three layers

- **Affordance = a cell.** The being's "what can I do" vocabulary *is* its cells:
  `sensory`, `MidLevel`, `other`, `self`, `motor`. An affordance is an actor that
  can fire, not a verb.
- **Parse-condition = a discrete outcome parsed from an inference response.** The
  branch point. `match` / `edit` / `new` (framing), a graded `match=N|none`, a
  motor emission kind, a `yes|partial|no` execution verdict — these are *conditions*,
  not affordances. They are what the LLM's answer **means**, reduced to a finite set.
- **State — two homes, one transition.** A parse-condition drives a transition that
  writes **slice-state** (persistent, on the owning cell's slice — the graph deposits
  + the phase/pointers) and **cell-state** (the cell's own transient mode for this
  wave) as warranted.

**The per-cell loop:**

```
read slice-state  →  (maybe) fire inference  →  parse response → CONDITION
     →  apply(slice delta, cell delta)  →  spawn next affordance
```

The kernel delivers waves; the *cell* is a small state machine keyed on
parse-conditions. This is the concrete form of "the cell reads the slice to decide
which method fires" — the method is chosen by (incoming slice-state × parse-condition).

## Where state lives (resolving the tension write-time-locking created)

Two facts must both hold: (a) predictions/plans hang off the **grouping** (the spine
— reachable from the grouping node), and (b) cells write their **own** slice
(write-time locking, single-writer). Resolution, straight from
[sensorimotor.md](../2026-06-16/skyra-dna/sensorimotor.md)'s "shared by reference"
rule:

- **The node is shared by reference.** A prediction/plan node is a heap `MemoryCell`
  the agent creates and appends as a child of the grouping. Identity is singular;
  many indexes can point at it.
- **The write is owner-local.** The *append* (mutating the grouping's children map)
  and the *activation record* land under a single writer's lock. Since the grouping's
  children map is the contested structure, **the grouping node carries its own mutex**
  (finer than the slice) — appends to a grouping lock *that grouping*, not the whole
  plane. This is the "lock the cell we append to, not the slice" decision made
  concrete: **the lock granularity is the node, not the slice.**
- **Cross-cell reads take that node-lock briefly** to snapshot (microseconds), never
  across an inference call.

So "slice-state" below means: state reachable from a cell's slice (its indexes +
the grouping spine it publishes), with **per-node locks** on the contested maps.

## The kernel/slice concurrency model (settled 2026-07-03)

This is the prerequisite the whole fan-out sits on. It came out of a walk that
corrected an earlier over-strong line ("the kernel stops locking entirely" — no;
see below). The resolution is an **ownership** fix, and the mechanics fall out of it.

### The bug being fixed

`relation.go` `traverse` holds `traversed.Slice()`'s lock across the **entire wave**,
including any inference a dispatcher cell fires inside `Observe`. That conflates two
unrelated things: *"I am mutating a map"* (nanoseconds) and *"I am thinking"* (~30s
of codex). The wave-wide lock exists to protect the **tree the kernel builds**, not
the thinking — but because the tree-build is threaded through the DFS, the lock ends
up wrapping the inference too.

### The ownership cut (the real fix)

The kernel should not implicitly own the slice it builds on. **The cell hands the
kernel a slice to build the tree on** (an explicit parameter, decoupled from
`traversed.Slice()`); concurrency = hand different slices. Sorting the current
`Slice` state by owner:

| State | Owner | Notes |
|---|---|---|
| **Records** (edge-owned activation — the EMAs, the synaptic weights) | the cell, **durable** | the learned memory; **shared by reference** across a cell's concurrent slices (one being = one set of synapses, not one per wave) |
| **Maps** (indexes, the grouping spine) | the cell, durable | shared by reference likewise |
| **Threshold** | the cell, durable | config |
| **Count** (the being's proper-time clock, stamped on nodes) | the cell, durable | ticked once per wave → the increment must be **atomic/locked** (concurrent waves must not lose a tick). NOT a global "total traversals ever" — that count is killed unless something needs it. |
| **Tree** (`TraversalPair`, the DFS build + Prior/Current double-buffer) | **the slice handed in this wave** | stays a slice concept (so `Prior`/`Current` is ready when recall makes it load-bearing — it is **not** vestigial, just not-yet-exercised). Two concurrent waves = two slices = two trees, no collision. |
| **Stack** | the kernel, transient | already a local in `traverse` |

So: two concurrent waves on one cell get **sibling slices** — each its own Tree,
both pointing at the **same Records/Maps by reference**. The tree stops colliding
because it's per-slice; the weights stay singular because they're shared.

### Locking (two concerns, both cheap)

Records is a shared `map[string]ActivationRecord`. Two levels need protection, and
the naive "lock the whole map" is what would cost us — the fix is finer:

1. **The record's value** (the EMA bump) — **a mutex per record.** Two waves bumping
   *different* edges never contend; two waves bumping the *same* edge serialize for
   nanoseconds (a few float multiply-adds).
2. **The map's structure** (adding a new edge key) — **an `RWMutex` on the map.**
   A per-record lock can't cover this: Go panics if one goroutine adds a key while
   another reads the map. Reads (bumping existing edges, every wave) take the shared
   read-lock → concurrent, no contention. The exclusive lock fires **only on edge
   birth** (Hebbian crystallization — rare). Common path fully concurrent; expensive
   lock only on the occasional new synapse.

**Why the cost is unmeasurable:** lock hold time is nanoseconds (arithmetic);
inference is seconds. Nine orders of magnitude apart. The *only* performance crime
is holding a lock across a think — which is the very thing this cut removes. As long
as every lock wraps arithmetic and never inference, the locking is free noise.

### What this corrects

- The two-stroke spec's "the kernel stops locking **entirely**" is too strong. The
  kernel keeps a lock — but **only around its own bookkeeping** (the Count tick, the
  Records/Maps touches), released around `Observe`/`Express`. Inference always runs
  lockless.
- **`activation-and-traversal.md` ownership table is revised:** it currently reads
  "the Slice owns TraversalCount, Records, Stack, Trees." New: **Records/Maps/
  Threshold/Count durable on the cell** (Count locked, global-count killed); **Stack
  the kernel's**; **Tree on the slice handed in per wave** (sibling slices share
  Records by reference). Update that doc as part of the build.

### Watch-item (correctness, not speed — parked)

If two overlapping waves bump the **same** edge, the per-record lock makes it *safe*
but maybe not *right*: the standing "one bump per cycle, not per wave" note
(entities-and-surroundings.md — an edge should groove once per frame) means two
concurrent bumps could double-count the groove and distort learning. This is a
cognition-correctness question, separate from the data race. **Parked as a watch-item**
— surface it if learning looks distorted under concurrency; do not design against it
pre-emptively.

## The inference-call inventory (ground truth — the six)

| # | Affordance | Call (system prompt) | Parse fn | Conditions it yields |
|---|---|---|---|---|
| 1 | **MidLevel** | framing (`coldStartSystem` → Stage C `<frame>`) | `ParseLabel` → `ParseFrame` | `match` / `edit` / `new` / *parse-fail* |
| 2 | **other** | predict (`otherSystem`) | `ParsePredictions` | `distribution` / *empty* / *parse-fail* |
| 3 | **other** | grade (`graderSystem`) | `ParseMatch` | `matched(N)` / `none` / *parse-fail* |
| 4 | **self** | predict (`selfSystem`) | `ParsePrediction` | `proposal` / *parse-fail* |
| 5 | **self** | grade (`execGraderSystem`) | `ParseExecGrade` | `yes` / `partial` / `no` / *parse-fail* |
| 6 | **motor** | collapse (`motorSystem`) | `ParseMotorEmission` | `engage(ext)` / `engage(self)` / `move` / `perceive` / `wait` / *parse-fail* |

*(Parse-fail is a real condition everywhere — it drives retry-with-correction, then
hold. Not an error to swallow; a branch.)*

## The transition tables (per affordance)

Read each row as: **condition → (slice delta · cell delta · next affordance).**

### MidLevel — the framing machine

| Condition | Slice delta | Cell delta | Next |
|---|---|---|---|
| `new` | create grouping anchor + first framing node; publish to grouping index | active-framing = new | fan out other ∥ self |
| `edit` | new framing node on the **same** grouping (old decays) | active-framing = edited | fan out other ∥ self |
| `match` | bump matched framing's edge (re-elect; no node) | active-framing = matched | fan out other ∥ self |
| `reframe`-triggered (from a grade `none` / exec `no`) | same as edit/new but **failure summary in the present** | active-framing = reframed | fan out other ∥ self |
| parse-fail | none (last framing stands) | — | retry → hold |

Cold-start (scenario 1) is the `new` row with no priors. MidLevel's slice-state:
`{grouping index, active-framing pointer, pending-trigger slot [B, deferred]}`.

### other — grade-then-predict (two calls, gated)

| Wave kind | Condition | Slice delta | Cell delta | Next |
|---|---|---|---|---|
| external engage (stroke 1) | *grade* `matched(N)` | **claim** the bet (test-and-set `Graded` under the bet's node-lock), write score = stored p(N); matched path stands | verdict = matched | **spawn self, verdict on the impulse** (no re-predict — fresh distribution waits for stroke 2) |
| external engage (stroke 1) | *grade* `none` | claim bet, write score 0 | verdict = miss | **spawn MidLevel `reframe`** (failure summary on impulse) |
| self-initiated / reafference (stroke 2) | — (grade gate closed) | — | — | *predict* → `distribution`: append new bets to grouping (per-cycle key), stamped `About` [E, deferred] |
| predict | `distribution` | append bet nodes | — | (deposit only; self reads lagged) |
| predict | *empty* / parse-fail | none | — | retry → hold |

other's slice-state: `{its bet nodes on the grouping, claim markers on each}`.
The claim marker is *the* concurrency primitive here (double-grade guard, scenario add A).

### self — grade-execution-then-plan

| Wave kind | Condition | Slice delta | Cell delta | Next |
|---|---|---|---|---|
| stroke 2 (motor emitted) | *exec grade* `yes/partial/no` | claim the realized proposal, write score | planned-this-wave = **false** | if `no` → spawn MidLevel `reframe`; else nothing |
| stroke 1 (has verdict from other) | *plan* `proposal` | append proposal node (per-cycle key) | planned-this-wave = **true** | spawn motor |
| stroke 1, carry-over (verdict ≥ bar, plan unfinished) | re-affirm (MVP: model says keep) | mark standing plan re-affirmed | planned-this-wave = **true** | spawn motor |
| any | parse-fail | none | planned-this-wave = false | retry → hold |

self's slice-state: `{standing plan node, planned-this-wave bool}`. The bool is the
**C wiring rule** made concrete: *motor is spawned only when planned-this-wave is
true* — the first slice-state-dispatch.

### motor — collapse-and-emit

| Condition | Slice delta | Cell delta | Next (world) |
|---|---|---|---|
| `engage(ext)` | append action (consumes proposal); record outgoing to sensory | pending = action | emit engage → world (spatial) |
| `engage(self)` | append action; **inner-note** path | pending = action; inner-note-origin flag | emit engage → world (**identity-direct**, private) |
| `move` / `perceive` / `wait` | append action (consumes proposal) | pending = action | emit primitive → world |
| parse-fail | none | — | retry-with-correction → hold |
| — (relay rail) | if this wave *was* triggered by an inner-note, the `engage(self)` condition is **unavailable** | — | fall back to perceive / hold |

motor's slice-state: `{ByEntity action index (built), pending action, inner-note-origin flag}`.
The consumed-proposal guard (built in Stage A) is the "append action → proposal now
has a child → readProposal skips it" chain.

### sensory — intake (no inference; condition = wave kind)

sensory's "parse" is the impulse discriminator, not an LLM call. Conditions:
`external-engage` (resolve sender, record, presence bump, → MidLevel) /
`self-engage` (inner-note: record on self entity, no presence bump) /
`perceive|move|bump return` (ingest, self-initiated → MidLevel) /
`self-wave` (skip). Already largely built; the `self-engage` inner-note row is the
new one (relay valve).

## The scenarios as table traversals (proof the tables are complete)

- **1 — cold start:** MidLevel`new` → other`predict:distribution` ∥ self (no verdict → `plan:proposal`, planned=true) → motor`emit`. No claim, no buffer.
- **2 — match:** other`grade:matched` (claim, verdict on wave) → self`plan` (verdict≥bar → revise-or-keep, planned=true) → motor`emit`. Sequential via the wave, not a lock.
- **3 — miss:** other`grade:none` → MidLevel`reframe` → other`predict` ∥ self`plan` → motor. Grade fronts the routing.

Every arrow above is a table row. If a scenario needs a transition not in a table,
the table is incomplete — that's the test.

## What this pins (the four missing artifacts, now concrete)

1. **Affordance set** = the five cells. (No general verb table — instance-level, per
   spread-before-collapse; the `memory-as-world` general table waits for instance two.)
2. **Slice-state struct, per cell** — enumerated in each table's caption. These are
   the fields to add: `active-framing`, `planned-this-wave`, `inner-note-origin`,
   claim markers, pending-trigger slot [B].
3. **Routing envelope** — Routes ride the impulse: the grade **verdict** and the
   **reframe summary** need a typed slot. `Impulse.Payload any` overloaded is fragile;
   add a typed `Route` field (verdict | reframe-reason | trigger-cause) — small struct,
   decided here so it isn't improvised mid-build.
4. **Cross-cell read contract** — `snapshotLatest(grouping) → node` under the
   grouping's node-lock; returns a value copy, never a live reference held across a
   think. One method, used by self (reads other's bet), motor (reads self's plan).

## Open

- **Route envelope shape** — one typed field vs a small tagged union on the impulse. Lean: tagged struct `Route{Kind, Verdict, Summary}`, nil on plain waves.
- **Lock granularity — settled** (see "The kernel/slice concurrency model"): RWMutex on the Records map (concurrent reads, exclusive only on edge-birth) + a mutex per record (nanosecond serialization on same-edge collisions). Sibling slices share Records by reference; Tree is per-slice. The grouping's children map follows the same pattern (append under a short structural lock).
- **planned-this-wave lifetime** — reset at wave entry; confirm it's wave-scoped cell-state, not persisted.
- **Where `active-framing` lives** — MidLevel slice pointer vs derived (highest-activation framing child). Lean: derived (no pointer to keep coherent), computed by `currentFraming()` — but cache per wave.
- **[B] pending-trigger slot / [E] About stamp / [F] staleness** — still deferred to the scenario-4 pass; their table rows are stubbed above where they'd attach.

## Done when

- Each of the five cells has a written slice-state struct and a transition table with no scenario needing an off-table move.
- The `Route` envelope is a typed impulse field; verdict and reframe-reason travel on it.
- `snapshotLatest` is the one cross-cell read; no cell holds another's node-lock across inference.
- Scenarios 1–3 (and, after its pass, 4) are each a clean sequence of table rows.
