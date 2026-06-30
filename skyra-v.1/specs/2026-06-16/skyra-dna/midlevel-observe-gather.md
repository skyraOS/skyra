# MidLevel — Observe / Setpoint Gather (read-half of the MVP grouping pass)

**Date:** 2026-06-25 · **Status:** read-half **BUILT** (frame assembly); write-half stubbed.
The read-half of `MidLevel.observeExternal`: how MidLevel assembles the grouping input from
sensory's warm surroundings. The **inference call** (score impulse vs. groupings) and the
**grouping write** (`setpoint → groupID`) are the **continuation** (stubbed). The active-probe /
DDM upgrade is [midlevel-active-grouping.md](midlevel-active-grouping.md).

> Pipeline context: membrane → sensory → **MidLevel** is wired (sensory's Express spawns MidLevel).
> MidLevel is *reached* but *blind* — this spec is how it reads what sensory perceived.

## The handoff — read sensory's warm state, don't snapshot

Sensory's wave already ran; its index is **warm** (`ByEntity` = the surroundings, presence
records live on its slice). MidLevel reads it **directly** — no payload snapshot, no assembly in
sensory.

- **Back-reference.** MidLevel holds a pointer to sensory, handed in at construction (sensory
  constructs MidLevel, so it passes itself: `NewMidLevelCell(id, slice, sensory)`; the pointer is
  valid and fully populated by the time MidLevel's wave runs).
- **Lock the cross-slice read.** MidLevel is on its own slice; sensory's warm state is on
  sensory's. MidLevel reads sensory's surroundings **under sensory's slice lock** (SWMR /
  lock-guarded — the same pattern as the test's `sensoryEntity` helper). This makes MidLevel
  *wait* for sensory's wave to fully finish and stays safe if the next external wave starts
  re-warming sensory. **Lock order: MidLevel reads sensory, never the reverse** (no cycle).
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
- **Cell-written IndexScore.** MidLevel writes the overlap score as **IndexScore** on each
  setpoint's record **before** the kernel's `Activate` folds it in — the prescribed division:
  *cell writes IndexScore/Alignment; kernel folds into the EMAs + threshold* (it never computes
  overlap itself).

## Frame assembly — `Parse()` direct (built)

MidLevel already **holds the gathered references** (it read them from sensory), so it assembles
the frame by calling each node's `Parse()` **directly** — *not* by routing through the kernel's
Express payload (that path only earns its keep for cross-actor result delivery; here MidLevel is
both holder and assembler).

- **`MemoryCell.Parse() → ParsedNode{Content, Timestamp, GroupID}`** — a self-description:
  structured (not formatted text), generic (the content carries itself; only the occurrence
  time is lifted out for ordering). `GroupID` empty until grouped.
- **MidLevel organizes:** sort by `Timestamp` — the MVP frame, a chronological transcript of the
  surroundings, the only structure available at cold start; group by `GroupID` with the
  understanding attached **once groupings exist** (the grading frame). Same inference call,
  richer input.

> **Cleanup target — the sectioned view (2026-06-26).** The flat timestamp transcript is the
> degenerate case (only the *Exchange* section populated). The full view is MidLevel's "conscious
> field": group by **Kind** into labeled sections — **Surroundings** (entities + presence),
> **Exchange** (utterances), **Active groupings** (each gist + its member setpoints),
> **Standalone setpoints** (ungrouped, by gist). Rules: recency-ordered but direction differs
> (Exchange ascending — it's a transcript; knowledge sections descending — recent/relevant on
> top); recency = timestamp now, the Familiarity EMA later; **capped per section** (top-K — a
> curated readout, not a dump; the gather's overlap × recency selects, this renders the
> selection); cognitive Kinds render at **gist altitude** (a setpoint as its understanding/label,
> not raw internals — MidLevel's scoped visibility); the setpoint section is internally organized
> **by grouping**. (Mirrored in the `renderFrame` doc comment.)
- **The kernel does NOT descend MidLevel's children for the frame.** Children-traversal +
  `Activate` (IndexScore fold, EMA, threshold — the *activation/scoring* half) is a **separate,
  deferred** concern, wired when the bump equation matters. The frame is a direct read.

## Cell reads Kind; the kernel does not

Filtering `entity.Children()` to setpoint-kind (and entity-kind for the involved parties) reads
`Kind` — and that's **fine**. "The wave doesn't read Kind" constrains the **kernel** and the
**uniform memory-node traversal** (entity/setpoint/understanding stay interchangeable *to the
wave*). MidLevel is a **specialized cell** building its *own* `Children()` input; reading Kind is
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

No completed setpoints → nothing to gather → **skip grouping** (per [midlevel.md](midlevel.md)).
MidLevel carries the surroundings as context; no grouping until setpoints exist.

## Wiring deltas (built)

- `NewMidLevelCell(id, slice, sensory)` — the sensory back-reference (sensory passes itself in).
- `MemoryCell.Parse() → ParsedNode` + `contentTimestamp` (occurrence time lifted from typed
  content).
- `MidLevel.observeExternal` → `assembleFrame`: lock sensory → parse entities + their children
  entity-outward → timestamp-sort → `g.frame`; `MidLevel.Frame()` exposes it.
- *Deferred (write-half):* overlap × recency scoring (IndexScore), the grouping inference call,
  the `setpoint → groupID` write, and the kernel scoring of MidLevel's children.

## Continuation (not this spec)

- **The inference call** — score impulse vs. active groupings (one call); skip if no completed
  setpoint.
- **The grouping write** — `setpoint → groupID` tags from the inference result (bottom-up; see
  working-memory.md).
- **The upgrade** — active LLM-driven probe loop + DDM stop controller
  ([midlevel-active-grouping.md](midlevel-active-grouping.md)).

## Open

- The score blend — overlap vs. recency weighting.
- Hot-entity selection — how many / how hot count as "the surroundings."
- Per-entity read cap for hubs (and whether to maintain the recency-ordered index now or later).
- One-hop (MVP) vs. limited multi-hop gather.

## Done when

- On an external wave, MidLevel reads sensory's warm surroundings **under lock**, gathers the
  involved setpoints **entity-outward**, scores them by **overlap × recency** (IndexScore
  written), and exposes them as `Children()`.
- **Bounded** — no scan-all; cold start skips cleanly.
- *(continuation)* the inference call groups the candidates.
