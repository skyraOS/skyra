# Settlement Set — A → set of O, keyed to v3

**Date:** 2026-07-20 (revised same day — first draft was keyed to v2 vocabulary;
corrected against skyra-dna-v3, which already computes grades and already draws the
raw/interpreted line)
**Status:** Ground decision. Revises edge kind 6 of
[observe-mechanism-over-spine](../2026-07-19/observe-mechanism-over-spine.md).
**Parents:** [../../design-docs/skyra-dna-v3/scoring.md](../../design-docs/skyra-dna-v3/scoring.md) ·
[../../design-docs/skyra-dna-v3/the-cycle.md](../../design-docs/skyra-dna-v3/the-cycle.md) ·
[../2026-07-14/closure-duration-objective.md](../2026-07-14/closure-duration-objective.md)

## The rule that governs everything here

> **No LLM grades observations. The difference is the grade — computed, never judged.**
> (v3's founding line. v2's ceiling was precisely that free-text predictions could only
> be graded by a judge, so the learning signal bottomed out in an opinion.)

A graded observation has two kinds of content, and they are not settled the same way:

```text
DETERMINISTIC   the world's returns. "skyra put the bottle in the fridge" —
                witnessed, decomposed into RAW triples, resolves expects by
                EXACT MATCH at full fidelity. no judge anywhere.

INTERPRETED     the model's reading of the world's returns. "michael wants
                water" — an interpreted triple: marked, confidence-carrying,
                DerivedFrom-pointing, resolves at capped fidelity (~0.3),
                and its own grounding loop is OPEN (tabled in observation.go,
                marking required now precisely so it can be added later).
```

The observe mechanism's bookend discipline is the same line: the deterministic part is
the mechanical middle's food; the interpreted part exists only at ingest (the
decomposer) and collapse — and the collapse commit is itself an interpreted claim
awaiting contradiction.

## The decision, restated in v3 vocabulary

The first draft said "an action's prediction can have many observations, so kind 6 is
`A → {(O, grade, t), ...}`." In v3 terms that conflated two mechanisms that scoring.md
keeps separate:

```text
an EXPECT resolves ONCE     (triple, confidence, within) → one observation → one
                            computed grade → done. never re-graded.

the EDGE is what repeats    P(W | E), an EMA over counted outcomes across
                            instances. accumulation lives HERE, not on the expect.
```

So the settlement **set** decomposes into three legitimate pluralities, none of which
is "one bet graded many times":

1. **Many expects per action.** One action birth-stamps several triples ("bottle in
   fridge", "michael satisfied", "door closed") — each resolves once against its own
   observation. This is the per-action set, and it is already v3-native (a joint is an
   edge whose Source is a set; cardinality does the work).
2. **Edge-level accumulation across instances.** "Garage 2/2" is not two grades on one
   action — it is one edge's EMA fed by two resolved instances (P, N, plastic floor).
   The class statistics of episode-as-region ARE edge EMAs grouped by class.
3. **State-holding = re-issued expects, not re-graded ones.** "The bottle lives in the
   kitchen fridge" is a durable expectation maintained by repeated bets, each with a
   `within` deadline, each resolving once. The renewal profile
   (closure-duration-objective) is the *cadence of re-betting*, read off the edge.
   **`TruthUnknown` never moves the EMA** — the bottle unobserved is not the bottle
   missing; O11 (bottle out, in use, in the garage) resolves nothing unless a bet was
   due.

## What survives from the first draft

- **Append-only.** Resolutions never rewrite; `ResolvedBy` names the evidence forever
  (LAW 2 as a query: every grade points at the observation that produced it).
- **Long-range edges.** An observation resolving an old bet arcs back across spans —
  the spine's cross-span stitching, and a traversal surface for the wave.
- **Late evidence updates the counted layer** (edge EMAs → class statistics) without
  touching span membership or the spine.
- **Born counting over resolutions, not actions** — which in v3 vocabulary is just:
  read the edge's (P, N), which already counts resolved instances.

## What the correction kills

- Any reading of kind 6 where an LLM adjudicates whether an observation settled a
  prediction. Attribution is exact-match against birth-stamped triples, or it is not
  kind 6.
- "Retrospective attribution" as a settlement mechanism. "Maybe the fertilizer did it"
  is an interpreted causal claim — it may enter the graph as an interpreted triple
  with DerivedFrom, capped fidelity, its own eventual edge — never as a witnessed
  settlement.
- The first draft's "grade vocabulary (+/−/suspended/renewing)" open item: dissolved.
  Outcomes are {0,1} at the edge; "suspended" is just an unresolved bet within its
  window; "renewing" is re-betting cadence. No new vocabulary needed.

## Keying note for this session's docs

The paper tests (#1–#4) and episode-as-region were drafted while reading dnav2;
their mechanics survive re-keying but their vocabulary shifts:

```text
session term                      v3 term
"resolution act on the spine"     state-change node (interpretation, has interior)
"birth-stamped Final causes"      the expect: (triple, confidence, within) at emission
"settlement +/−"                  computed grade: log(p_claim) − log(p̂_edge)
"class statistics"                edge EMAs (P, N) grouped by episode class
"aitia layer" (first draft)       interpreted triples + their eventual edges
dnav2 edit/new LLM judge          does not exist in v3; segmentation is mechanical
                                  (paper test #4); the decomposer at ingest is the
                                  one open interpretation site
```

The spine's `O → A → O` succession reads through v3's loop as
`observation → state change → action → observation` — the paper tests folded the
state-change node into their "resolution acts." Un-folding it changes no arithmetic
but places the open interpretation portion exactly where v3 puts it: on the
`observation → state change` edge, "the one edge with the entire cognitive act
inside it."
