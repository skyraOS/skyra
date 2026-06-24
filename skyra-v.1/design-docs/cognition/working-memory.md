# Working Memory — the Lit-Up Subgraph + the Index Set

**Date:** 2026-06-22 (supersedes the 2026-06-21 "uncrystallized buffer" framing)

## Core definition

**Working memory is not a separate buffer or store. It is the lit-up
(high-recency), traversal-reachable subgraph — what's currently active.** There is
one graph; WM is a *view* over it, produced by traversal, weighted by recency. No
container, no second store, no copy.

This is Cowan's model done right — "the activated subset of long-term memory in the
focus of attention." Recent = lit, because the fast EMA (Familiarity) just bumped it.
The *focus* mechanism is the **traversal entrypoint chain** (below): you don't scan
the graph for high activation, you traverse out from the current surroundings and
what you reach *because it's lit* is, by definition, working memory.

> This reverses the earlier framing ("WM = the uncrystallized buffer"). The reason:
> a crystallized setpoint you're actively using is obviously in working memory, so
> WM-ness can't be "uncrystallized." WM is **lit**, regardless of crystallization.

## Two orthogonal axes — lit and crystallized

The earlier doc conflated these. They are independent:

| | **lit** (WM — high recency) | **not lit** (dim) |
|---|---|---|
| **crystallized** (durable graph node, has BirthAlignment) | a known thing I'm thinking about now | dormant memory |
| **uncrystallized** (just-appended artifact, no stamp) | the live exchange / open prediction | a faded artifact, about to be forgotten |

- **lit (WM)** = a *signal level* — recency activation. Decays.
- **crystallized** = a *status* — proven/durable, set at closure (the `BirthAlignment`
  stamp). Orthogonal to lit-ness; a crystallized node is in WM only if it's lit.

## Two quantities — excitation (decides view) and `completed` (a readout)

These were both being called "activation"; they are different and have different
homes. **Excitation decides what's lit; `completed` is a readout on the lit set — it
does not drive lit-ness.**

1. **Excitation / activation** — *what's in view.* The bump
   (`IndexScore × Alignment × stress × intrinsic` into the EMAs). `Alignment` (per-edge,
   node↔setpoint match) lives here. This is the **sole gate** on view.
2. **`completed` (notDone)** — *how done a lit setpoint is.* A **signed −1..1 inference
   judgment** on each lit setpoint (a judgment call, not arithmetic). Transitory:
   meaningful only while lit. Triggers closure — `+threshold → crystallize`,
   `supporting setpoint < 0 → cascade`, between → in-progress.
3. **Crystallization (proven)** — *durable.* A status stamp (`BirthAlignment`), set at
   closure. Orthogonal to both (a crystallized node is in WM only if lit).

**`completed` is gated by excitation, not a cause of it.** A setpoint that loses
excitation drops **out of view → "not important at the moment"**, regardless of its
`completed`. There is **no tension force keeping open loops alive** — an open
prediction is lit only while it's genuinely excited (recent/relevant via its
grouping). When the context moves on, it dims and fades like anything else.
Neglected intentions **self-prune**; open loops don't accumulate into infinite
never-closed goals all screaming. (This corrects an earlier "tension keeps it lit"
framing — `completed` is downstream of excitation, never a driver.)

### A lit prediction/setpoint leaves view three ways

- **crystallizes** — its own `completed` crosses `+threshold` → converts to a
  **setpoint** (+ `BirthAlignment`).
- **cascades** — a **supporting** setpoint is judged `completed < 0` → rolls back, the
  failed span converts to a **Stress** (see [error-recovery-loop.md](error-recovery-loop.md#cascade--a-failure-wave-de-crystallization)).
- **fades** — loses excitation (context moved on) → out of view, plain decay.

**The only marker of "open":** a **prediction-`Kind` node that's still lit and hasn't
been converted.** The grading pass judges lit prediction-kind nodes' `completed`;
closure flips the `Kind` (prediction → setpoint / stress). That `Kind` is the only
thing distinguishing an open prediction from a recent-but-closed setpoint.

## No two-pool model

There is **one** activation record per edge (Identity α=0.02 = the groove/LTM;
Familiarity α=0.1 = the recency that *constitutes* WM). The earlier "WM fast pool vs
graph slow pool" is gone — same record, the fast EMA is the WM read, the slow EMA is
the durable read. Recording recent artifacts still happens (so the wave can reach
them); "buffer as a separate structure" does not.

## Three layers (don't conflate)

| Layer | Stateful? | Home |
|---|---|---|
| **The graph** (all nodes, crystallized + uncrystallized artifacts) | yes | the slice |
| **WM** — the lit, traversal-reachable subgraph | a *view*, not a store | (emergent) |
| **The frame** — the kernel's tree this round | **no** (rebuilt) | Global's map |

The frame is the **readout** (top activations this round); WM is the **lit view** the
traversal produces; the graph is the substrate. Persistence/decay live on the nodes'
EMAs. **Frame = hard refresh; node activation = graded decay.**

## The index set

WM is a set of per-slice indexes — the transient lookups each cell type maintains.
The **`groupID` is the spine** that joins observe ↔ self ↔ other; the **action** keys
the justification. All forward-only (reverse direction is the wave / entity overlap).

| Slice | Index | Role |
|---|---|---|
| **Sensory** (membrane Observe) | `entity → last-known state` | presence / location |
| **Sensory** | `entity → recent utterance/action (from/to)` | the observed exchange |
| **Self** | `groupID → newPrediction` | generation (my next move) |
| **Self** | `incoming stimulus → prediction` | grading / audit (reafference) |
| **Other** | `groupID → newPrediction` | generation (their expected move) |
| **Other** | `incoming stimulus → prediction` | grading / audit (their response) |
| **Global (observe)** | `groupID → setpoints` | grouping membership (**shifts each round**) |
| **Global (express)** | `action → { selfPrediction, otherPrediction }` | justification |
| *(on the node)* | `grade / alignment` | cell-parsed, reachable via the indexes above |

### The grouping spine
Given a `groupID` you assemble its whole frame — its setpoints (observe) + its
self/other predictions (self/other) — by joining on the id. Given an `action` you
trace its justification (express). The `incoming → prediction` indexes are the
**grading audit** — *which evidence tested which prediction* — without storing the
grade value redundantly (the grade lives on the prediction node, cell-parsed).

### Loose coupling — at the `grouping → setpoints` joint
Predictions are **not** hard-bound to setpoints. `grouping → setpoints` is the index
that *churns* each round (A,B,C → B,C,D as Global re-groups); `grouping → prediction`
stays put. So a prediction made under grouping G automatically applies to **G's
current setpoints**, whatever they are now. On closure the prediction becomes its
own new setpoint, stamped with the present surroundings — so it never needed a fixed
setpoint binding.

**Requires:** groupings **persist** (stable id, decaying, with shifting membership).
That's what keeps `grouping → prediction` stable while `grouping → setpoints` churns.

### Self vs. other — symmetric
"incoming" means different things per side:
- **Self**: incoming = my action's **reafference** → which self-prediction it closes.
- **Other**: incoming = their **response** → which other-prediction it tests.

The from/to entity tags on the sensory exchange index also carry self/other (skyra-
said = self, michael-said = other), so the observed stream is split for free.

## The traversal chain (the entrypoint structure)

The indexes above realize one layered chain — the structure WM traversal walks.
**Surroundings (entity) is the base layer**; the chain runs:

```
entity ──(overlap hop)──▶ grouping ──▶ { setpoints,            ← Global-observe: groupID→setpoints
                                          self-prediction,      ┐ merged: grouping's children
                                          other-prediction }    ┘ are all three kinds
   self-prediction  ──▶ action-response      ("did my intent execute")
   other-prediction ──▶ lastSaid             (their response — closes back to the base layer)
```

Plus the entity's own value: `entity → {location, lastSaid, recency-activation}`.

**The merge.** `groupID → setpoints` (observe) and `grouping → prediction` (self/other)
are keyed by the same thing, so a grouping's `Children()` is **one row of three
kinds** — setpoints + self-pred + other-pred — not three separate indexes. The
grouping node *is* the spine join point.

**The leaf closes the loop — the prediction's child is its own evidence:**
- **other-prediction → lastSaid** — did they respond as predicted (and `lastSaid` is
  already in the surroundings, so the leaf points back to the base layer).
- **self-prediction → action-response** — did my intent execute.
The comparator at each prediction node reads `prediction` (content) vs `child`
(evidence) → alignment. Traversal *delivers* the comparison.

**One overlap hop, stored lookups below it.** `entity → grouping` is **not** a stored
edge (the entity doesn't keep a list of its groupings) — `entity.Children()` returns
**candidate groupings**, surfaced by surroundings-overlap (recency × IndexScore).
Everything below (grouping → {setpoints, preds}, prediction → response) is **stored
adjacency**. So: one overlap hop at the top, stored lookups the rest of the way.

**Per-cell append ownership.** Each cell appends into the shared chain at its own
node; the merged `grouping.Children()` is *assembled* by reading the owners' appends
for that groupID:
- Global(observe) appends `grouping → setpoint`
- self appends `grouping → self-prediction`
- other appends `grouping → other-prediction`
One writer per kind (no write conflict), one reader at traversal (the union).

**Each cell's entrypoint.** WM is per-cell-relative over the one shared chain: the
membrane enters at the surroundings, Global at groupings, self/other at their
predictions — each traverses *outward from its node* by activation, and what it
reaches is *its* working-memory view.

## Traversal — the kernel's DFS over dynamic `Children()`

WM is **not** read by a separate replay pass or a hand-rolled loop. It's traversed by
the **kernel's existing DFS**, where each cell's `Children()` is **computed
dynamically by reading its region's map**. The kernel calls `Children()` at each hop,
scores the row by activation, descends the survivors — that *is* the loop. You author
the map-reading `Children()`; the kernel loops.

```
Global.Children()    → reads the grouping map      → active groupings
grouping.Children()  → reads groupID→setpoints      → its setpoints
setpoint.Children()  → its understandings
action / response    → read action→preds, etc.
```

**One slice** holds the several region maps (sensory entity map, self/other prediction
maps, Global's grouping map) + the durable graph — "reading the other maps" is reading
distinct maps on the same plane. One base cell (Global / the membrane root) is where
the wave starts.

**Two hop kinds, both handled by the kernel's activation scoring:**
- **Stored-adjacency** — `Children()` reads a real index (`grouping→setpoints`,
  `setpoint→understanding`, `action→predictions`) and returns the linked row.
- **Overlap** — no stored index (e.g. `entity→grouping`); `Children()` returns the
  candidate set and the kernel surfaces the relevant ones by IndexScore (surroundings
  overlap).

Either way it bottoms out in "score children by activation, descend survivors" — **no
kernel change.** Edge-keying is preserved: `Children()` builds each row keyed by edge
identity so activation stays edge-owned.

**Two decay pools, split at crystallization:**
- WM entries (uncrystallized) carry **fast-decay** activation (recency).
- Crystallized nodes carry the **slow EMAs** (the durable groove).
- Crystallization promotes fast→slow (+ the `BirthAlignment` stamp). The pools don't
  mix — WM recency never contaminates the durable groove.

So "recent" = high fast-decay activation, surfaced by the *same* wave that surfaces
durable memories by their slow activation. No replay needed for selection.

**Ordering:** selection is by activation (relevance + recency), not sequence. If the
assembled *text* frame needs chronological readability, keep a lightweight
`seq`/timestamp per entry used **only** to sort the selected set at text-assembly
time — the kernel stays order-blind.

## Capacity & eviction

The lit view is bounded even though the graph isn't:

- **Token cap (~50k, TBD) on the readout.** The *frame* (what's assembled for an
  inference call) is a token budget, not an item count. Assembly is a knapsack:
  highest-activation lit nodes that fit. The cap is on the readout, not on what the
  graph stores. Lines up with the ~100k recall ceiling from the captain's log.
- **Below-0 → removed.** A negative `Alignment` bump (the node was contradicted)
  drives activation below zero → **evicted**. For an uncrystallized artifact that's
  true forgetting; a crystallized node going negative is the cascade/stress path.
- **Neglected uncrystallized artifacts fade** (decay) — the open-prediction GC above.
- **The durable graph keeps** (crystallized nodes suppress toward 0 but persist as
  history; no cap). So the *graph* is durable; the *lit view over it* is bounded and
  forgetful — correct biology.

Crystallized vs. uncrystallized is the **`BirthAlignment` stamp**, not lit-ness: an
uncrystallized artifact has no stamp (intrinsic neutral); crystallization stamps it
(closure). Both can be lit (in WM) or dim — see "Two orthogonal axes" above.

## Concurrency

With **one slice per being**, the traversal is a **single wave** that the kernel runs
under the slice lock — so reads and writes within a being are **serialized by
construction**, one wave at a time. self/other are sub-traversals on that one slice
(not separate-slice concurrent agents), so there's no intra-being concurrent map
access to worry about. Appends from `perceive` land between waves (or under the lock).

(If we ever split a being across slices — separate-slice self/other for true
concurrency — the rule would be: Go maps allow unlimited concurrent **reads** with no
concurrent writes, so the discipline becomes **build/freeze → concurrent reads →
serialize writes at the join** (snapshot-in / join-out). Not needed for the one-slice
design; noted in case concurrency is reintroduced.)

## Open

- The token-cap value (~50k) — find by experiment.
- Reverse directions (`prediction → action`, `prediction → grouping`) are
  un-indexed; add only if a flow needs them.
- Whether `incoming → prediction` is always-on or audit-only (it's the auditability
  index — keep if the trace earns its cost).

## Source / related

- `entities-and-surroundings.md` — the surroundings map + entity-indexing + the wave
- `groupings-and-understandings.md` — groupings, gists, the consolidation pipeline
- `activation-and-traversal.md` — the bump equation, EMAs, BirthAlignment
- `error-recovery-loop.md` — the justification index (action → predictions), the cascade
- specs: `specs/2026-06-16/skyra-dna/{global,membrane,setpoint}.md`
