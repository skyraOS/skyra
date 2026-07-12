# Minimal Aitia Set

**Status:** MINIMAL CANDIDATE · implementation target

## Purpose

The first system does only four things:

```text
query by subgraph
  -> set the active Aitia set
  -> grade each incoming impulse
  -> retain the set or requery
```

Maxim, Effect, prediction horizons, action selection, and Praxis learning are
not required for this loop.

## Aitia Collapse

There are no independent Material, Formal, Efficient, or Final Aitia nodes.
Material and Formal Cause collapse into one present subgraph:

```text
G_present = (V, E)

V = Material Cause
E = Formal Cause
```

Efficient and Final Cause are active belief relationships between this present
subgraph and other subgraphs:

```text
G_prior  --EfficientCause--> G_present
G_present --FinalCause-----> G_future
```

`G_prior` and `G_future` are also hylomorphic subgraphs. If either is later
queried as the present situation, its nodes become Material Cause and its edges
become Formal Cause from that new standpoint.

## Minimal Aitia Set

Version zero uses exactly one active present subgraph, one active Efficient
Cause, and one active Final Cause:

```text
AitiaSet
  PresentSubgraphID
  EfficientCauseEdgeID
  FinalCauseEdgeID
```

This is an active explanatory set, not four cause records. The set and all
subgraph contents are immutable after birth. Graph adjacency and belief strength
remain edge-local.

> **Decided 2026-07-12 — summary-node realization.** Each compound is one
> **summary node** (carrying the sentence) with its subgraph behind it — the
> grouping→gist shape recurring. So the set's three references resolve to three
> summary nodes: `PresentSubgraphID` → the present compound's summary node, and
> the Efficient/Final edges run **compound-summary-node → compound-summary-node**
> (`G_prior → G_present`, `G_present → G_future`), each with the target's
> subgraph hanging behind its summary node.

Changing the present subgraph or either active causal relationship creates or
reuses an immutable Aitia set and installs a fresh contextual edge. Nodes may be
reused. Edges and their learned strength are never inherited into a new context.

## Query Contract

> **Decided 2026-07-12 — three workers, not one big frame.** The present,
> Efficient, and Final compounds are emitted by **three separate workers with
> three AIMD lineages**, not one call. midlevel is the present-compound querier;
> the efficient/final workers **fan out in parallel** (both read only G_present
> + the evidence snapshot; neither consumes the other's output — the same
> data-dependency rule that sequenced other→self) and **join at the atomic
> set-commit** via the board's declare/mark/atomic-fire machinery. Budget
> attribution stays local: a contradicted efficient story taxes the efficient
> lineage only. This is also the **universal querier contract**: every cognitive
> worker emits (sentence, subgraph) in the same format.

Each query worker receives its available input and its lineage's
`QueryTokenBudget`. Per worker, it emits:

```text
present worker:   one sentence + present subgraph G(V,E)
efficient worker: one sentence + prior compound + EfficientCause edge
final worker:     one sentence + future compound + FinalCause edge
```

The worker chooses the firmest specificity rung it can support within the
budget. A smaller budget means: remove the least-supported distinctions and
emit a complete, coarser account. It must not abbreviate or truncate.

Every sentence referent must map to a node, every asserted relationship must
map to an edge, and every emitted graph element must contribute to the sentence.
**Format direction (decided):** the worker emits one **annotated sentence** and
the subgraph is parsed from it, so the correspondence rules are enforced by the
shared parser, not a second judge. Exact syntax: ledger item 1.

## Active State

Active state belongs to the mutable Slice, never to an immutable Aitia set:

```text
Slice.ActiveAitiaSetByEntity[EntityID] = AitiaSetID
```

Activation commits atomically after a valid query result is stored. Requerying
replaces the Slice reference; it does not mutate the previous set. A bounded
history may retain prior active-set selections.

## Incoming Impulse And Evidence

Each incoming world impulse is preserved as immutable Evidence with nonnegative
mass. Evidence contains no `support`, `oppose`, or `defeat` category.

The grading episode relates that Evidence to the active set's claims:

```text
Evidence
  -> grade relevant structural edges inside G_present
  -> grade the Efficient Cause edge when relevant
  -> grade the Final Cause edge when relevant
```

Irrelevant or unresolved relationships receive no grade. No grade is not a
grade of zero.

Evidence is not stored inside the immutable Aitia set. Provenance records link
the Evidence and each relationship it graded. The Praxis screen may traverse
those records to display supporting, opposing, and unresolved Evidence for the
active set without owning or changing it.

## Edge Learning And Defeat

Existing signed edge mechanics carry over unchanged:

```text
-1 <= grade <= 1

Identity    += 0.02 * (grade - Identity)
Familiarity += 0.10 * (grade - Familiarity)
edgeStrength = (Identity + Familiarity) / 2
```

Polarity expresses resonance or tension. Absolute magnitude controls ordinary
proportional traversal. Defeat belongs to a contextual edge, never to an
immutable subgraph or Aitia set.

An edge that trends to zero receives no ordinary traversal allocation. Its
target remains available through another path, where reaching it creates a new
contextual edge rather than resurrecting the defeated edge.

## Retain Or Requery

After grading, the substrate computes fit for the active Aitia set:

```text
supported:
  retain the active set
  QueryTokenBudget = min(B_max, B + alpha)

contradicted:
  QueryTokenBudget = max(B_min, beta * B)
  query the same available evidence at the smaller budget
  atomically install the replacement Aitia set if valid

unknown or irrelevant:
  retain the active set
  leave QueryTokenBudget unchanged
```

Only correctness earns tokens. Incorrectness removes budget and therefore the
right to make equally detailed claims. Initial tuning candidates remain
`alpha = 10 tokens` and `beta = 0.5`.

The substrate owns budget updates, retries, activation, and any cascade. The LLM
sees only the current query, evidence, and permitted budget.

## Traversal Budget Is Separate

```text
TraversalAmplitude = context budget controlling graph reach
QueryTokenBudget   = earned budget controlling query specificity
edgeStrength       = learned contextual relationship weight
```

None of these values may be substituted for another.

## Minimal Runtime Loop

```text
bootstrap or receive Entity/Situation
  -> query Aitia set within QueryTokenBudget
  -> persist immutable subgraphs, set, and fresh edges
  -> activate set on Slice

for each incoming impulse
  -> persist immutable Evidence
  -> grade relevant active-set relationships
  -> update edge EMAs
  -> classify active-set fit
  -> update QueryTokenBudget
  -> retain or requery
```

## Praxis Screen, Version Zero

Praxis is read-only in this implementation. It may display:

- the active present subgraph and its sentence;
- the active Efficient and Final Cause beliefs;
- Evidence provenance for each displayed relationship;
- signed grades and current EMA strengths;
- the current earned `QueryTokenBudget`;
- prior active sets retained in history.

Praxis does not yet select actions, create Maxims, predict Effects, or grade
Skyra's interventions.

## Deferred

```text
Maxim and action-guiding principles
Effect and explicit prediction machines
prediction horizons and Effect/Evidence matching
Praxis action selection and learning
reward, utility, and emitted-impulse grading
multiple simultaneously active Efficient or Final beliefs
```

These may later consume the active Aitia set. They are not prerequisites for
the minimal cognitive loop.

## Required Invariants

1. Material and Formal Cause exist as `V` and `E` of one present subgraph.
2. Version zero activates exactly one Efficient and one Final Cause relationship.
3. Every durable node and Aitia set is immutable.
4. Reuse preserves nodes but always creates fresh contextual edges.
5. Evidence is immutable, nonnegative-mass, polarity-free, and external to the set.
6. No grade is distinct from grade zero.
7. Defeat belongs to edges and does not delete reachable content.
8. Only verified correctness earns query tokens.
9. Incorrectness multiplicatively reduces the query token budget.
10. Active state lives on the Slice.
11. Traversal amplitude and query token budget remain independent.
12. Maxim, Effect, and Praxis learning remain deferred.
13. Efficient and Final are ONE belief-edge mechanism read in opposite temporal
    directions; only their grading semantics differ (records vs orientation).
14. Composition is role-uniform: summary nodes are legal referents in new
    compounds; role lives on the edge/position, never the node.

## Still To Settle Before Code

Moved to the open ledger —
[minimal-loop-uncertainties.md](./minimal-loop-uncertainties.md) — resolved one
at a time, each decision recorded in place (the perceive-build method).

