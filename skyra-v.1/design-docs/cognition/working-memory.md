# Working Memory — the Uncrystallized Buffer + the Index Set

**Date:** 2026-06-21

## Core definition

**Working memory is everything that hasn't crystallized yet.** The boundary *is*
crystallization:

- **Crystallized → the graph.** Setpoints, understandings, entities — durable nodes
  with a `BirthAlignment` stamp. The traversal lights some of them up, and they're
  *in the frame* as context — but being activated does **not** make them WM. They're
  consolidated memory that happens to be lit.
- **Uncrystallized → working memory.** The raw recent stream — what I said, what the
  other said, the predictions I'm holding — none of which has become a graph node.
  A transient, decaying buffer on the slice.

This is the hippocampal fast-buffer / Baddeley episodic-buffer (recent experience
held *before* consolidation), **not** Cowan's "activated subset of LTM." We have an
explicit consolidation event (crystallization), so the buffer is the right model:
crystallization = consolidation; "has a `BirthAlignment` stamp" = "is durable."

## Three layers (don't conflate)

| Layer | Stateful? | Decays? | Home |
|---|---|---|---|
| **Durable graph** (crystallized) | yes, long | slowly (Identity) | the graph |
| **WM** — the uncrystallized buffer + the groupings | yes | yes (like everything) | the slices |
| **The frame** — the kernel's tree this round | **no** (rebuilt) | n/a | Global's map |

The frame can be non-stateful *because* the decay/persistence lives one layer down
(WM + the EMAs). The frame is the **readout** — the top activations the kernel
surfaced this round; WM is the **stateful substrate** the kernel reads to produce it.
Rule of thumb: **frame = hard refresh; WM = graded decay.** (A grouping or in-flight
prediction that goes quiet for one round should *attenuate*, not be hard-deleted, so
it can recover.)

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

## Capacity & eviction — WM is lossy (the graph is not)

WM is bounded; the durable graph is not. This asymmetry is deliberate.

- **Token cap (~50k, TBD).** WM's job is to fill an inference frame, so capacity is a
  **token budget**, not an item count. Assembly is a knapsack: highest-activation
  entries that fit. Lines up with the ~100k recall ceiling from the captain's log.
- **Below-0 → removed.** A negative `Alignment` bump (the entry was contradicted)
  drives activation below zero → **evicted** (true forgetting).
- **Contrast with the graph:** the graph *keeps* (suppresses; negatives persist as
  history; no cap). WM *forgets* (token-capped; below-0 evicted). WM is lossy, LTM is
  durable — correct biology.

Activation reuses the same machinery (entity-indexed, Identity/Familiarity EMAs,
decay wave). The one difference from a crystallized node: a WM entry has **no
`BirthAlignment`** yet (intrinsic defaults to neutral; it's stamped *at*
crystallization). "Has a birth stamp" = "is crystallized" = "left WM for the graph."

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
