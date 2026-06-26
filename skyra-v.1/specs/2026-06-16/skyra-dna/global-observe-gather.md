# Global — Observe / Setpoint Gather (read-half of the MVP grouping pass)

**Date:** 2026-06-25 · **Status:** in spec — **start of a spec.** The MVP read-half of
`Global.observeExternal`: how Global assembles the candidate setpoints (the grouping input)
from sensory's warm surroundings. The **inference call** (score impulse vs. groupings) and the
**grouping write** (`setpoint → groupID`) are the **continuation** (stubbed for now). The
active-probe / DDM upgrade is [global-active-grouping.md](global-active-grouping.md).

> Pipeline context: membrane → sensory → **Global** is wired (sensory's Express spawns Global).
> Global is *reached* but *blind* — this spec is how it reads what sensory perceived.

## The handoff — read sensory's warm state, don't snapshot

Sensory's wave already ran; its index is **warm** (`ByEntity` = the surroundings, presence
records live on its slice). Global reads it **directly** — no payload snapshot, no assembly in
sensory.

- **Back-reference.** Global holds a pointer to sensory, handed in at construction (sensory
  constructs Global, so it passes itself: `NewGlobalCell(id, slice, sensory)`; the pointer is
  valid and fully populated by the time Global's wave runs).
- **Lock the cross-slice read.** Global is on its own slice; sensory's warm state is on
  sensory's. Global reads sensory's surroundings **under sensory's slice lock** (SWMR /
  lock-guarded — the same pattern as the test's `sensoryEntity` helper). This makes Global
  *wait* for sensory's wave to fully finish and stays safe if the next external wave starts
  re-warming sensory. **Lock order: Global reads sensory, never the reverse** (no cycle).
- **Read the index, not the tree.** Read `ByEntity` (the durable warm surroundings), not
  `Tree().Current` — the tree is ephemeral / double-buffered and may be mid-rebuild on the next
  wave.

## The gather — entity-outward, bounded

The relevance question is "which setpoints does the current situation bear on" — answered
**entity-outward**, never by scanning all setpoints:

- Take the **hot surroundings entities** (from `sensory.ByEntity`, ranked by presence/recency).
- For each, read its **stamped setpoints** via the reverse index (`entity.Children()`,
  Kind-filtered to setpoints).
- **Bounded by Σ(hot entities' setpoints)**, *not* by the total graph. Never scan-all-setpoints
  ("every setpoint that has ever lived").

## The score — overlap × recency (IndexScore, cell-written)

- **Overlap (the discriminator).** A setpoint's relevance = **how many hot entities reach it** —
  the intersection of multiple present entities. A setpoint touched by several hot entities
  outranks a single-entity one. Computed as a **one-shot count per cycle**, NOT N separate bumps
  (N bumps would over-groove the durable EMAs and conflate "relevant to a lot *now*" with
  "consistently relevant *over time*" — activation doc).
- **× recency.** The `entity → setpoint` **edge activation** (recently-active edges first).
- **Cell-written IndexScore.** Global writes the overlap score as **IndexScore** on each
  setpoint's record **before** the kernel's `Activate` folds it in — the prescribed division:
  *cell writes IndexScore/Alignment; kernel folds into the EMAs + threshold* (it never computes
  overlap itself).

## Build `Children()` — the dynamic map

- `Global.Children()` returns a **fresh map of the scored candidate setpoints**, computed *this
  cycle* by reading the surroundings — the canonical "author the map-reading `Children()`; the
  kernel loops" pattern (working-memory.md).
- The kernel then traverses + thresholds the survivors. **No kernel change.**

## Cell reads Kind; the kernel does not

Filtering `entity.Children()` to setpoint-kind (and entity-kind for the involved parties) reads
`Kind` — and that's **fine**. "The wave doesn't read Kind" constrains the **kernel** and the
**uniform memory-node traversal** (entity/setpoint/understanding stay interchangeable *to the
wave*). Global is a **specialized cell** building its *own* `Children()` input; reading Kind is
**cell policy, not kernel mechanism.**

## Hub saturation — the known seam (#1)

Even entity-outward, a long-lived **hub entity** can have a huge stamped-setpoint set, so reading
all of its children is the cost (ancestor-failure-modes #1: "within-hub discrimination; the real
discriminator is *intersection* of multiple present entities").

- **MVP:** gather + threshold. The **overlap** (multi-entity intersection) and **recency** cuts
  *discriminate* relevance; bounded by the graph, not unbounded exploration.
- **Later:** a per-entity **recency-ordered index** so a hub's top-K is read without scanning all
  its children. The **multi-hop** associative recall (reaching *absent*-but-related entities) is
  the active-probe loop's job, not the MVP's. **One-hop bounded gather now.**

## Cold start

No completed setpoints → nothing to gather → **skip grouping** (per [global.md](global.md)).
Global carries the surroundings as context; no grouping until setpoints exist.

## Wiring deltas

- `NewGlobalCell(id, slice, sensory)` — the sensory back-reference.
- `Global.Observe` (`observeExternal`): lock sensory → gather hot entities → their setpoints →
  score (overlap × recency, write IndexScore) → build the candidate map.
- `Global.Children()` returns the gathered map.

## Continuation (not this spec)

- **The inference call** — score impulse vs. active groupings (one call); skip if no completed
  setpoint.
- **The grouping write** — `setpoint → groupID` tags from the inference result (bottom-up; see
  working-memory.md).
- **The upgrade** — active LLM-driven probe loop + DDM stop controller
  ([global-active-grouping.md](global-active-grouping.md)).

## Open

- The score blend — overlap vs. recency weighting.
- Hot-entity selection — how many / how hot count as "the surroundings."
- Per-entity read cap for hubs (and whether to maintain the recency-ordered index now or later).
- One-hop (MVP) vs. limited multi-hop gather.

## Done when

- On an external wave, Global reads sensory's warm surroundings **under lock**, gathers the
  involved setpoints **entity-outward**, scores them by **overlap × recency** (IndexScore
  written), and exposes them as `Children()`.
- **Bounded** — no scan-all; cold start skips cleanly.
- *(continuation)* the inference call groups the candidates.
