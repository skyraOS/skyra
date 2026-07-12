# Cause And Effect Research

**Status:** CANDIDATE REVISION · 2026-07-12 · four decisions folded in (below)

This directory develops the query-by-subgraph cause-and-effect model. The
minimal Aitia-set loop is the current implementation candidate. It does not yet
supersede the accepted four-node Schema model.

## Documents

1. [Minimal Aitia Set](./minimal-aitia-set.md) defines the current narrow
   implementation candidate: query, activate, grade, and retain or requery.
2. [Query By Subgraph](./query-by-subgraph.md) defines the hylomorphic query,
   causal belief bounds, emitted sentence/subgraph contract, and persistence
   rules.
3. [AIMD Abstraction Controller](./aimd-abstraction-controller.md) defines how
   verification changes the token budget and therefore the permitted rung of
   specificity.
4. [Minimal Loop Uncertainties](./minimal-loop-uncertainties.md) — the open
   ledger: remaining decisions, resolved one at a time, each recorded in place.
5. [Implementation Phases](./implementation-phases.md) — the build plan: six
   phases, each landing live before the next; ledger items resolved
   just-in-time per phase; built on dnav2, kernel reform kept a separate track.

## Decided 2026-07-12

- **Universal querier contract.** Every cognitive worker emits (sentence,
  subgraph) — midlevel is the present-compound querier (which it always was);
  efficient/final workers emit prior/future compounds; later workers (predict,
  propose) adopt the same shape. One emission format everywhere → one shared
  validator.
- **Annotated sentence as the format direction.** The worker emits one annotated
  sentence (`michael[n1] asked[e1:n1→n2] skyra[n2] …`); the subgraph is PARSED
  from it, so sentence↔subgraph correspondence is checked by code, not a second
  judge. Exact syntax is ledger item 1.
- **Summary-node shape.** Each compound = one summary node (the sentence) with
  its subgraph behind it — the grouping→gist pattern recurring for causes. The
  AitiaSet references three summary nodes: present, efficient, final.
- **Efficient ∥ Final — parallel, join at set-commit.** Rule (already load-bearing
  in the pipeline): sequence when one worker consumes the other's output;
  parallelize when both consume only the shared snapshot. Efficient and Final
  both read G_present + the evidence snapshot and write disjoint compounds →
  fan-out in parallel, join via the board's declare/mark/atomic-fire machinery,
  commit the AitiaSet atomically. (A lagged efficient→final read can be added
  later — how self reads other's latest — without re-architecting.)
- **Three workers, three budgets — not one big frame.** AIMD lineage per
  compound so budget attribution stays local: a contradicted efficient story
  taxes the efficient lineage, not the present compound's specificity. Grading
  and provenance stay local for free.
- **Composition is role-uniform and free.** A summary node is a node, so
  previous compounds are legal referents in new sentences (`[k=compound:id]`) —
  present compounds compose from present compounds, efficient from prior
  efficients, same for final; role lives on the edge/position, never the node,
  so the old same-CauseKind composition constraint dissolves. V0 reuse
  candidates come from the **lineage history on the Slice** (bounded
  active-state history), not open-graph recall. (Offer-only vs instructed
  reuse: ledger item 9.)
- **One edge mechanism, two grading rules (invariant).** Efficient and Final
  are ONE belief-edge mechanism read in opposite temporal directions from the
  anchor; the difference lives ONLY in grading semantics — efficient edges
  grade against records (retrodiction), final edges grade against orientation
  (is the unfolding present still bending toward the target). Never build a
  separate final-cause machine.
- **Candidate final-closure rule** (candidate, not decided — must survive a
  live run): *a final edge closes positively when a later efficient compound
  cites its target's lineage* — the anticipated future, once actual and queried
  as a present, explains itself through this compound. Purpose graded by
  retrospective explanatory participation, not occurrence. See
  query-by-subgraph.md verification semantics.

## Central Shape

```text
                  inferred prior subgraph
                     Efficient belief
                            |
                            v
Entity/Situation -> Query G = (V, E) <-grades- Evidence/Impulse
                            |
                            v
                     Final belief
                  imagined future subgraph
```

Within the query, `V` is Material Cause and `E` is Formal Cause. Efficient and
Final Cause point to other subgraphs. Those targets may themselves be queried
as Material/Formal compounds when the traversal is re-anchored on them.

Maxim and Effect are deferred. They may later consume the active Aitia set but
are not required by the minimal loop.

## Explicit Separations

- `TraversalAmplitude` spends context on graph reach.
- `QueryTokenBudget` bounds the detail of one generated query.
- Edge strength records belief support and tension.
- A budget reduction requests a coarser description; it does not mutate nodes.
- An unresolved claim is pending or unknown, not defeated.
