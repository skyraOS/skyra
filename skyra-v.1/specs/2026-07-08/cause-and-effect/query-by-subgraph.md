# Query By Subgraph

**Status:** ADOPTED FOUNDATION / PARTIALLY SUPERSEDED · 2026-07-12

> The query-by-subgraph contract and Material/Formal collapse are adopted.
> Later sections treating Evidence as a separate type, Effect as part of the
> minimal loop, or causal-node roles as intrinsic are superseded by
> [Subgraph And Causal Edge Ownership](./subgraph-and-causal-edge-ownership.md).
> Efficient and Final remain perspectival weighted edges between one-type
> `CausalNode`s.

## Decision Under Study

Material Cause and Formal Cause are one queryable hylomorphic compound:

```text
G = (V, E)

V = material: the participating entities, observations, constraints, and terms
E = form: the relationships that organize those nodes into this situation
```

The compound is primary. Matter and form remain analytically distinguishable:
the same nodes can carry different form, and the same form can organize
different nodes.

## Causal Bounds Are Beliefs About Other Subgraphs

Efficient Cause and Final Cause do not become extra contents inside `G`.
They are belief edges from the current query to other subgraphs:

```text
G_prior  --EfficientCause belief--> G_current
G_current --FinalCause belief-----> G_future
```

- Efficient Cause is an inferred prior compound believed to have produced or
  constrained the current compound.
- Final Cause is an intended or anticipated future compound believed to orient
  the current compound.
- When either target becomes the standpoint of a later query, it is itself a
  Material/Formal compound `G = (V, E)`. Cause is therefore a relational role,
  not an intrinsic node type.

Here, **unbounded belief** means that Efficient and Final beliefs are not
collapsed into the present query and may point across arbitrary abstraction or
temporal distance. It does not mean unlimited traversal or token cost.

## Query Result Contract

The LLM chooses the firmest specificity rung warranted by the available input.
It may answer concretely or abstractly, but it must emit both:

1. one concise sentence expressing the query;
2. one subgraph that expresses exactly that sentence.

The correspondence is strict:

- every material referent in the sentence maps to one or more nodes;
- every asserted relationship maps to one or more edges;
- every emitted node and edge contributes to the sentence;
- unsupported detail is forbidden;
- inability to justify a fine rung causes retreat to a coarser supported rung,
  not invention and not automatic abstention.

> **Decided 2026-07-12 — correspondence is free by construction.** The worker
> emits one **annotated sentence** (e.g. `michael[n1] asked[e1:n1→n2] skyra[n2]
> for help[n3]`) and the subgraph is **parsed from it** by one shared parser.
> Dual emission followed by an LLM correspondence check is rejected — the
> parser is the validator. Exact annotation syntax is ledger item 1
> ([minimal-loop-uncertainties.md](./minimal-loop-uncertainties.md)).

The serialized sentence and subgraph must fit `QueryTokenBudget`. The budget is
a control over representational specificity, not graph traversal amplitude.
The implementation must test the assumption that reducing this budget reliably
induces useful abstraction rather than merely truncation.

## Requery And Persistence

When grading shows that the current Material/Formal query is losing fit, the
substrate requests a new query over the same available evidence with a smaller
`QueryTokenBudget`. The LLM performs only the local rewrite; it is not told
whether the event belongs to a larger cascade.

The new sentence and subgraph are new immutable nodes. Previous nodes remain in
history. Reuse means reusing an immutable node through a new edge, never
mutating or reusing the old edge.

Efficient and Final belief targets survive a Material/Formal requery. Their
edges continue to carry their own grades and EMAs until evidence drives their
support toward zero. Requerying the present compound does not silently defeat
its causal bounds.

> **Tension to reconcile (ledger item 6):** the paragraph above lets
> Efficient/Final edges *keep* their EMAs across a present-compound requery, but
> the minimal-set rule says a new AitiaSet installs **fresh contextual edges**
> and "edges and their learned strength are never inherited into a new
> context." A requery mints a new set — so do the surviving belief edges keep
> their state (requery ≠ new context) or re-mint fresh (requery = new context)?
> See [minimal-loop-uncertainties.md](./minimal-loop-uncertainties.md).

## Verification Semantics

Resolution has at least three states:

```text
supported    observed information agrees with the claim
contradicted observed information disagrees with the claim
unknown      records are absent, ambiguous, or the future has not resolved
```

`unknown` holds the controller and must not be converted to contradiction.
Efficient claims can be checked against records only when those records are
sufficient. A predicted Effect can be checked after its horizon. Final Cause is
purpose or orientation and is not automatically falsified merely because a
predicted Effect did not occur. If Final Cause is later redefined as a future
prediction, that semantic change must be explicit.

> **Invariant (2026-07-12) — one edge mechanism, two grading rules.** Efficient
> and Final are ONE belief-edge mechanism read in opposite temporal directions
> from the anchor (re-anchoring already makes the roles relational). The entire
> difference lives in grading semantics: **efficient edges grade against
> records** (retrodiction — did the past actually look like this); **final
> edges grade against orientation** (is the unfolding present still bending
> toward the target — defeated when the trajectory stops pointing at it, not
> when a clock runs out). Do not build a separate final-cause machine.

> **CANDIDATE (2026-07-12) — final closure by retrospective citation.** A final
> edge closes **positively** when a later efficient compound cites this
> compound's lineage: when the anticipated future eventually arrives and is
> queried as a present, its efficient worker either reuses/references this
> present compound (compounds are legal referents — the composition principle)
> or it does not. If it does, the future, once actual, explained itself
> *through* this present — purpose verified by retrospective explanatory
> participation, not by occurrence. Substrate-checkable (a lineage-reference
> walk, no Effect-matching, no horizon timers); Effect stays deferred. Marked
> candidate: must survive a live run before it hardens into the grading rules.

## Effect Boundary

This draft reserves **Effect** for a predicted observable consequence. That
keeps three concepts separate:

```text
Final Cause = what future or purpose orients this situation
Effect      = what consequence this model predicts
Evidence    = what was subsequently observed
```

This separation permits forward verification without turning purpose into a
prediction or turning observation into belief.

## Open Implementation Questions

- What canonical graph serialization makes sentence/subgraph correspondence
  mechanically checkable?
- Is the budget scoped per query lineage, relation, worker, or some combination?
- What grading signal measures Material/Formal fit without grading verbosity?
- When does a requery reuse a prior immutable subgraph rather than create one?
- Does a lower token budget produce abstraction consistently across models?
