# World Assertions & the Policy Chain — the Active-Inference Reading

**Date:** 2026-06-29 · **Status:** **design direction** (the structured-upgrade path). The MVP
runs on free-text predictions + LLM-judged alignment (what's built); this is the evolution that
makes grading **falsifiable graph overlap** instead of an LLM opinion. The cells don't change —
the *representation* (world-assertion edges) and the *grading* (overlap, with the action as the
hinge) do.

## The frame — one representation for facts and predictions

A **world assertion** is a **present-tense, independently falsifiable subject-relation-object
triple** — which *is* a graph edge:

```
User VIEWING RestaurantList
Cart CONTAINS Item
Payment AUTHORIZED
```

This is the constructive form of the
[prediction-falsifiability invariant](../laws/invariants/prediction-falsifiability-invariant.md):
a prediction is a **claimed edge**; it is falsified when the observed graph doesn't contain it.

**Observed vs. Expected** — the crucial split:
- **Observed** assertions — what sensory wrote (facts). The world *already contains* them.
- **Expected** assertions — what self/other emit. **Not observations — hypotheses about state.**
- **Alignment = overlap( observed graph , expected graph )** — graph **containment**, not an LLM
  vibe-check. Predictive coding, made literal (generative model vs. sensory input; error =
  mismatch).

## The stack, as active inference

| Cell | Role (active inference) | Emits |
|---|---|---|
| **sensory** | observation | observed world-assertions |
| **mid-level** (grouping) | the generative model / framer / assembler | frames *now → end*; plugs the relevant edges + entities into the grouping (its children) and traverses. **Not a query engine** — an assembler. |
| **other** | the **goal** (preferred end-state) | the expected world *at the horizon* — where this is headed (theory of mind of the other's trajectory) |
| **self** | the **policy** (the bridge) | the **goal-conditioned** action that moves now → end |
| **motor** | execute | the action |

- **Intent is dropped** — it collapses into the **grouping** (the grouping *is* the reusable
  situation-pattern; its constituents are the steps).
- **Horizon = grouping depth.** A shallow grouping → near-term expected state (next observation);
  a deep/recursive grouping → the far goal. The recursion *is* the planning horizon.

## The three-index policy chain

Self is not a guesser; it is a **bridge**, and the bridge is three indexes:

```
1.  self's latent entities/edges  →  action                  (CONVERGENCE — the intersection)
2.  action                        →  falsifiable observation  (the gradeable HINGE)
3.  falsifiable observation       →  other's entities/edges   (goal-graph membership)
```

- **Index 1 is a convergence, not a stored edge.** The key is a *set* (self's structure); the
  action is **where the threads intersect/overlap** — the same IndexScore/overlap mechanism as
  the gather, and the membrane-motor coincidence-fire (aligned threads converge → fire). **One
  convergence hop, then stored lookups** (same shape as the rest of the traversal).
- **Index 2 is the linchpin** — the falsifiability invariant made *structural*: "this action
  should produce **this** checkable world-state." It is the **only** place the loop becomes
  gradeable. Drop it and you can act but never know if it worked.
- **Index 3** places the action's effect inside other's goal-graph.

**Bidirectional — this is what makes self a *bridge*, not a guesser:**
- **Forward (reactive):** self → action → expected-obs → is it in other's goal-graph?
- **Backward (planning):** other's goal → the obs that satisfies it → the action that produces
  it → the self-structure that implies it. **Self back-chains from the goal.**

Concretely (FoodOrdering): other's goal `Order SUBMITTED` → required obs `Checkout COMPLETED` →
action `submit` → self-structure `{Cart, Item, Payment}` converging.

## Grading — falls out of indexes 2 + 3 (the two channels)

```
self → action → { expected obs }   ⊆?   { other's expected world }
```

- **Index 2 graded** — did the action actually produce its expected observation? → **self's
  policy effectiveness**.
- **Index 3 / overlap** — does that observation sit inside other's expected world? → **did we
  reach the goal**.

The action is good **iff its falsifiable effect is a subset of other's goal-graph.** Alignment as
**graph containment** — the observed/expected overlap, with the action as the testable hinge in
the middle. This is the error-recovery loop's two channels (alignment vs. system-failure),
grounded.

## Decomposition — how the observed graph gets built from text

An impulse is raw text; the observed graph is structured assertions. Between them sits a
**decomposition stage** — small models extract triples and link entities.

- **Async + buffered.** Decomposition is slow (several model calls + entity resolution); it must
  **not block the cognitive loop**. Producer/consumer: the decomposer fills the observed graph;
  cognition runs on whatever's resolved; late edges arrive next cycle. (Impulses wait in a buffer
  to be fully decomposed.)
- **Entity resolution is the floor** — the #1 seam in
  [../../captains-log/observations/02-ancestor-failure-modes.md](../../../captains-log/observations/02-ancestor-failure-modes.md)
  (relevance-correctness ≤ resolution-correctness). **Preserve low-confidence ambiguity** —
  forcing a hard link too early corrupts the graph permanently. Keep ambiguous links as
  **low-weight candidates**; let evidence resolve them over time ("similarity may gather, only
  judgment may groove").

## Relationships

- [../laws/invariants/prediction-falsifiability-invariant.md](../laws/invariants/prediction-falsifiability-invariant.md)
  — the invariant; this doc is its constructive form (the world-assertion = present-tense S-R-O
  graph edge, graded by overlap).
- [comparator.md](comparator.md) — alignment; here it becomes **graph overlap** with the action
  as the hinge.
- [self-and-other-modeling.md](self-and-other-modeling.md) — self/other; here **other = goal**,
  **self = policy**.
- [../../specs/2026-06-16/skyra-dna/agent.md](../../specs/2026-06-16/skyra-dna/agent.md) — the
  agents' predict step (where the expected assertions are emitted).
