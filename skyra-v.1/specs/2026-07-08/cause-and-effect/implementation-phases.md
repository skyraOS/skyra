# Minimal Loop — Implementation Phases

**Date:** 2026-07-12 · **Status:** PARTIALLY SUPERSEDED BUILD PLAN — each phase lands live (TUI + real
inference), committed green, before the next begins (the perceive-build
discipline). Ledger items resolve **just-in-time**: each phase names the
decisions it needs; nothing else blocks.
**Parents:** [minimal-aitia-set.md](./minimal-aitia-set.md),
[minimal-loop-uncertainties.md](./minimal-loop-uncertainties.md), the dnav2
pipeline (the substrate this builds ON, per the isomorphism: midlevel→querier,
gist→summary node, edit/new→retain-requery, activeGistID→ActiveAitiaSetByEntity,
settle+other:grade→evidence grading, ingest→evidence persistence).

> Phase 0/1 history remains accurate. Replan Phase 2 onward against the decided
> layered graph before implementation. In particular, Evidence no longer lands
> as a separate node contract; observation-born `CausalNode`s are stamped by
> ingest, semantic edges are never graded, and grades land on weighted
> inherited/Efficient/Final causal edges. Typed child ports and graph-owned
> reverse retrieval are also now required substrate work.

> **Confirmed 2026-07-12:** both structural calls below (build inside dnav2;
> kernel reform as a separate, non-prerequisite track) are accepted.

## Build-on, not fork

The minimal loop is built **inside dnav2** — new genes + a new signed-EMA edge
store — reusing the orchestrator/board (declare/mark/atomic-fire join), the
ephemeral-worker machinery, the inference seam + logging, the world, and the
TUI. No v3 package; port-don't-rewrite held before and holds here.

**The kernel reform (Spacetime / Cost / Observe-returns-surface / Scheduler) is
a separate track, NOT a prerequisite.** V0 reaches its active set by Slice
pointer, not traversal; frame assembly stays on the mini-kernel. Proportional
allocation and history-recall (ledger item 8) join when recall-over-history is
needed — that is the seam where the two tracks meet, and it is *after* this
plan.

---

## Phase 0 — The parser (the mouth everything speaks through)

**Resolve first:** ledger item 1 (annotated-sentence syntax — proposal on the
table).

Build, pure code, no inference:

- the annotated-sentence **parser**: sentence → (summary node, V, E), with the
  `[k=compound:id]` composition/reuse form and post-parse label resolution
  (token-web exact-match discipline)
- the **validator** = the parser (correspondence free by construction);
  retry-with-correction protocol for parse failures (bounded, no AIMD tax)
- the **immutable persistence shapes**: compound = summary node + subgraph
  members; the signed edge store (`Identity += .02·(g−I)`,
  `Familiarity += .10·(g−F)`, strength = mean, lazy Familiarity decay) — new
  store beside the existing ActivationStore, not a mutation of it
- golden tests: sentences ↔ graphs, malformed inputs, reuse resolution

**Done when:** a corpus of annotated sentences round-trips deterministically.

---

## Phase 1 — The present querier (midlevel becomes what it always was)

**Resolve first:** ledger item 7 (plurality — lean: one situation-anchored set).

- `midlevel:frame` emits the annotated sentence; parse → present compound;
  persist immutable; **activate on the Slice** (`ActiveAitiaSetByEntity`, set
  carrying only the present reference — causal edges arrive Phase 4, which item
  5's summary-node-only lean already licenses)
- the sentence IS the gist text — renderers/frames keep working unchanged
- retain/requery replaces edit/new mechanically (same operator seam), but
  without grading yet: every turn requeries-or-retains on the framing op's own
  judgment, as today

**Live checkpoint:** converse in the TUI; the framing is now a compound; logs
show sentence + parsed subgraph per turn.

---

## Phase 2 — Evidence grading + retain-or-requery (the loop closes)

**Resolve first:** ledger items 2 (aggregation), 3 (defeat floor), 6 (requery =
lineage context, lean A). **Pin at build:** Evidence mass definition for v0
(candidate: mass = 1 per impulse; token-count later if experiments need it —
this is the gap flagged on the big set, now owned here).

- incoming impulses persist as immutable **Evidence** (the ingest/interaction
  paths already land them; add mass + provenance records)
- a **grading worker** grades the present compound's structural edges against
  new Evidence (adversarial-judge pattern, batch like the settle); signed EMAs
  update; no grade ≠ zero
- substrate computes **set fit** (supported/contradicted/unknown per item 2)
  and drives **retain-or-requery**; requery = new immutable compound + atomic
  re-activation; prior compounds retained (the gist-sequence behavior,
  formalized)

**Live checkpoint:** a contradiction in conversation triggers a visible
requery; supported evidence visibly retains and strengthens edges.

---

## Phase 3 — The AIMD controller (specificity gets earned)

**Resolve first:** ledger item 4 (B₀ = B_max/2 lean; expressibility floor).

- substrate-owned budget on the present lineage; wired into the querier's
  prompt; supported → +α, contradicted → ×β, unknown → hold
- the **required measurements** from the AIMD doc are this phase's deliverable:
  budget trajectories, and the truncation-vs-abstraction question answered with
  data (comparator-style judges grade completeness of coarse requeries)

**Live checkpoint:** watch the budget move over a session; confirm reduced
budgets produce coarser-but-complete sentences, not fragments.

---

## Phase 4 — Efficient ∥ Final fan-out + the full AitiaSet

**Resolve first:** ledger items 5 (summary-node-only allowed at low budget) and
9 (offer-only composition).

- two new workers on the existing fan-out machinery: midlevel commits the
  present → board declares {efficient, final} → both run in parallel (shared
  snapshot, disjoint writes) → join → **atomic AitiaSet commit** → activate
- three AIMD lineages (controller from Phase 3, instantiated per lineage)
- composition candidates: the lineage history's prior efficient/final summary
  nodes, offered in-frame (offer-only)
- evidence grading (Phase 2 worker) extends to the causal edges "when
  relevant" — records-grading for efficient, orientation-grading for final
  (invariant 13)

**Live checkpoint:** full sets visible; a wrong efficient story gets
contradicted by records and requeries at reduced budget without touching the
present compound's budget.

---

## Phase 5 — Closure candidate + the Praxis screen

**Resolve during:** the final-closure CANDIDATE (query-by-subgraph.md) — this
phase is its harden-or-drop run.

- the **closure checker**: on each new efficient compound, a lineage-reference
  walk — does it cite a compound some prior final edge pointed at? If yes, that
  final edge grades positively (closure by retrospective citation)
- the **read-only Praxis screen** (debugview precedent): active set + sentence,
  causal beliefs, Evidence provenance per relationship, signed grades/EMAs,
  current budgets, prior sets in history

**Live checkpoint:** run long enough sessions that at least one final edge gets
the chance to close by citation; keep or drop the candidate on what we see.

---

## Phase → ledger map

| Phase | Resolves | Pins at build |
|---|---|---|
| 0 | 1 (syntax) | parse-retry bounds |
| 1 | 7 (plurality) | set shape with absent causal refs |
| 2 | 2, 3, 6 | Evidence mass v0 definition |
| 3 | 4 | α, β, B_min/B_max starting values |
| 4 | 5, 9 | relevance gate for causal-edge grading |
| 5 | closure candidate | Praxis screen scope |

Item 8 (birth activation) stays parked — it bites at the history-recall seam,
which is the kernel-reform track's territory, after Phase 5.

## What "done" means

The minimal runtime loop from minimal-aitia-set.md runs live end-to-end:
bootstrap → query → activate; impulse → Evidence → grade → EMAs → fit → budget
→ retain/requery — with all three lineages, and every claim Skyra holds
inspectable as a sentence whose structure is a graph.
