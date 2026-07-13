# Cause And Effect Research

**Status:** CURRENT MODEL INDEX · 2026-07-12 · decisions folded in below

This directory owns the current query-by-subgraph layered cognitive model. It
supersedes the fixed four-node Schema/Aitia/Maxim/Evidence model. The minimal
Aitia-set loop is the current implementation direction; later decisions in the
ownership document override conflicting earlier drafts in this directory.

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
5. [Implementation Phases](./implementation-phases.md) — the current eight-phase
   build plan: Phases 0–7 are complete through atomic Efficient/Final expansion,
   continuous grading, AIMD, zero-strength amplitude semantics, and read-only
   Praxis/history inspection.
6. [Subgraph And Causal Edge Ownership](./subgraph-and-causal-edge-ownership.md)
   (DECIDED) — internal Formal edges, contextual Efficient/Final edges, the
   inherited/efficient/final summary ports, reverse retrieval, Evidence
   attachment, and typed child traversal as substrate reform.

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
- **Local graded-belief budgets — not one big frame.** Efficient and Final each
  own an AIMD lineage so a negative Efficient grade changes only its budget.
  Material/Formal is structural and ungraded, so it has no grade-driven AIMD
  lineage. Grading and provenance stay local.
- **Composition is role-uniform and free.** A summary node is a node, so
  previous compounds are legal referents in new sentences (`[k=compound:id]`) —
  present compounds compose from present compounds, efficient from prior
  efficients, same for final; role lives on the edge/position, never the node,
  so the old same-CauseKind composition constraint dissolves. ~~V0 reuse
  candidates come from the lineage history on the Slice~~ REVISED 2026-07-12:
  reuse candidates are **graph-owned** — the probe's reverse-index over
  inherited/causal maps (see the worker-mechanism bullet below); the Slice
  keeps only the active set. (Offer-only vs instructed reuse: ledger item 9.)
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
- **Typed child traversal.** Durable edge fields preserve identity, kind, and
  `FromID → ToID`; typed child ports expose Formal, Inherited, Efficient,
  Final, endpoint, Evidence, and reverse-summary adjacency to observe
  strategies. A flattened `Children()` view is compatibility-only when
  direction matters. This is SUBSTRATE reform (`MemoryCell`), not kernel
  reform — Observe is the strategy selector; port names never cross the kernel
  boundary.
- **Three causal ports on a summary, three lifecycles.** `inherited` =
  immutable composition-provenance EDGES: the compounds this compound was
  built FROM, minted at Persist from compound-typed members. Inherited edges
  record ancestry only; they have no signed grade, EMA, or ordinary evidence
  attachment. `efficient`/`final` = accumulating contextual
  edge occurrences, each stamped with its owning Aitia set. **The port
  accumulates occurrences; it never collapses them into a singular cause.**
  Unfiltered, the port reads as every explanation the compound has
  participated in — the retrieval surface for retrospective citation and
  re-anchoring. Active-context filtering is an observe-strategy choice, not
  storage.
- **Node-only causal inheritance.** Composition seeds the new compound's
  candidate maps with the parents' efficient/final cause NODES; the parents'
  edge occurrences never carry over. A regrade against the new summary mints
  a NEW edge owned by the new context, with its own evidence and signed-store
  history — edge occurrences are never shared across contexts.
- **Two node contracts, agreed vocabulary.** Knowledge graph (semantic,
  probe-time): Entities + Terms, semantic edges, the label lookup, and the
  reverse maps. CausalNode (weighted, traversal-time): the three process maps
  (inherited/efficient/final) + the parent pointer to its birth subgraph —
  the only bridge. CausalNodes have ONE type: material/formal, efficient,
  final are perspectives from the ACTIVE node, never node kinds. Semantic
  edges are never graded — grade relationships, never internals. Full
  vocabulary table in subgraph-and-causal-edge-ownership.md. ("Compound" in
  older docs = CausalNode; code renamed 2026-07-12.)
- **The Evidence collapse — two graphs, four types.** Evidence is not a third
  contract: an evidence node is a CausalNode at a lower level of abstraction
  (observed events decompose into knowledge subgraphs and are born as
  observation-level CausalNodes with maps STAMPED at birth: material/formal =
  the decomposition, efficient = what was alive in the frame, final = the
  emitted action). Final form: knowledge graph = semantic entity + semantic
  edge; causal graph = causal node + causal edge. Provenance stamps are
  mandatory content, never type (observed/composed ∥ observed/asserted).
  The inherited edge to an observation-born parent IS the evidence
  relationship. Self-action final = the final of everything efficient in the
  frame; other-agent final stamping + decomposition timing open (ledger
  10–12).
- **One retrieval mechanism, two worker outcomes.** Every worker may probe by
  subgraph, use the reverse index, normalize candidate hit mass into traversal
  amplitude, and verify exact directed structure. The present worker then
  chooses `retain`, `compose`, or `requery`; any inherited edges created by
  composition are ungraded ancestry. Efficient and Final workers each emit a
  complete prior/future subgraph and receive a fresh contextual belief edge.
  Only those two edges accept continuous grades or explicit no-grade results.
  Candidate sourcing is GRAPH-owned; the Slice keeps only the active set.

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
- An unresolved claim produces no grade; no-grade is not zero and is not defeat.
- Praxis inspection follows typed ports deterministically, including zero-
  strength history; ordinary cognitive traversal remains magnitude-weighted.
