# Episodic Binding Reality — Cell Type Spec

**Date:** 2026-06-16 · **Status:** skeleton · **`[after]`** (not on the MVP path)

## Files
- `../../../src/skyra/skyra-dna/episodic.go` — port + restructure
- self/other sub-realities — 2 hard-separate stores under this reality

## What it is
The **distributed** memory reality — the activation **field**. Energy spreads across setpoints by weight: the more active a setpoint, the more energy it carries in this field. **Wide / global / ambient** — the diffuse background of "what's lit up right now."

Three registers, same character: Tulving **episodic memory** (distributed, associative, pattern-completion); wave aperture **wide** (*surveys*); hemisphere **global**.

## Access pattern: distributed by weight
- Many setpoints active at once, each weighted by activation. **Breadth / spreading activation.**
- **Fast, contextual, decaying** (recency-weighted, short-lived).
- Provides the **weighted field**; its peak feeds the Temporal reality's spotlight (see coupling).

## Sub-realities (self/other — structural)
- `episodic-self`, `episodic-other` — two genuinely separate stores (the ToM split, enforced structurally; [self-and-other-modeling.md](../../../design-docs/cognition/self-and-other-modeling.md)).

## Coupling with Temporal
**Episodic = the field; Temporal = the spotlight on the field's peak.** Episodic maintains the weighted background; the most-active setpoint is what Temporal focuses on and explores deeply. See [temporal.md](temporal.md) for the move-vs-hold dynamic.

## Field activation — how a setpoint gets its weight (the salience the spotlight reads)
Salvaged from the retired `episodic-memory-summary.md` (deleted). Two bottom-up retrieval mechanisms set each setpoint's weight in the field:

- **person/place/thing → reverse lookup.** The active entities (agent/user, place, thing/tool) are an **inverted index** into the frames/setpoints that contain them: `active entity → reverse lookup → candidate setpoints get weight`. (Location-based memory.)
- **stress → state-dependent recall.** Current **stressful setpoints raise other stress-congruent setpoints *and their history*** (the frames/episodes they were part of). Stress can activate **independently** of any entity match (pure state-dependent recall).

**Combination = product across active dimensions.** `weight = product(entity-match × stress-congruence × …)` over whatever dimensions are active — more dimensions matching = stronger; a single dimension fires alone (product of one — so stress captures without an entity match). This is the salience equation.

Top-down then modulates it: **Coherence bumps/decays each setpoint vs. working memory** (its held synthesis). Bottom-up product-salience + top-down Coherence bias = the full field weighting the spotlight reads.

## Durable field — the window over it is Coherence's
This reality is the **durable** distributed field (long-term, weighted by activation). The **sliding window** over recent activity — the transient, overlapping working-memory window — is **[Coherence's](coherence.md)**, *not* this reality's. The old doc's "one frame in N episodes" (membership overlap, the multi-context "duplication," a feature not bloat) is that window's overlap, and it lives in Coherence's working memory. This reality just holds the field Coherence windows; it does not slide.

## Retrieval notes (salvaged before retiring the old doc)
- **Two retrieval modes, one mechanism:** *associative* (first wave — current state activates broadly, the activation gradient chooses what returns) and *direct* (re-descent — name specific keys from the first wave; only the intersection lights up). The re-descent is the error-recovery loop's deeper passes.
- **Suppress, don't delete:** three-layer noise filtering — amplitude (mechanical, no inference), binding-reality (inferential, traceable), suppressed-not-removed (low activation, resurfaceable on a state change). Same non-destructive curation as the comparator.
- **Shared store, per-reality keys:** frames live in one durable store; each reality has its own typed key, entity keys (person/place/thing) are shared, all resolve to frame IDs. Frames are born transient on the slice and graduate to the store when the traversal closes.
- threshold-descent (descend until sufficient) → [error-recovery-loop.md](../../../design-docs/cognition/error-recovery-loop.md); competence-on-the-wave → `comparator.md`.

## Changes
### 1. Port + restructure  `[after]`
- [ ] port `episodic.go` to `skyra.Space`; the distributed-field access (keep all survivors, weighted)
- [ ] `self`/`other` sub-realities
- [ ] fast / decaying activation dynamics
### 2. Voting  `[after]`
- [ ] generates a distributed self-model with confidence/alignment → votes

## Design sources
- [self-and-other-modeling.md](../../../design-docs/cognition/self-and-other-modeling.md), [setpoints.md](../../../design-docs/cognition/setpoints.md), [cell-types.md](../../../design-docs/being/cell-types.md) — Episodic

## Open
- decay rate / field dynamics
- how the field's peak is read out for Temporal's spotlight

## Done when
- **(after)** the episodic field — distributed weighted activation across setpoints, self + other — compiles and traverses; its peak is readable.
