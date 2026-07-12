# Cause And Effect Research

**Status:** CANDIDATE REVISION · 2026-07-12 · decisions folded in (below)

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
- **Three workers, three budgets — not one big frame.** AIMD lineage per
  compound so budget attribution stays local: a contradicted efficient story
  taxes the efficient lineage, not the present compound's specificity. Grading
  and provenance stay local for free.
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
  composition-provenance EDGES: the compounds this compound was built FROM,
  minted at Persist from compound-typed members — a different relationship
  than Efficient or Final, and the durable home of composition. Topology
  birth-fixed, grades live: midlevel's evidence sits on the inherited edges,
  graded with respect to the NEW summary (evidence is not a cause node and
  never lives on a cause node). `efficient`/`final` = accumulating contextual
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
- **One worker mechanism: probe → compose+grade.** Step 1: emit a probe
  subgraph → reverse-index lookup → candidates with hit counts → traversal
  amplitude distributed proportionally → traverse candidates' evidence.
  Step 2: compose the new summary (citing candidates lands them in
  `inherited`) + seed a base grade per evidence relationship on the RECEIVING
  edge (the CrossBirth pattern). The receiving edge is the only parameter:
  midlevel grades the inherited edges; efficient/final workers grade the
  causal edges. Candidate sourcing is GRAPH-owned — the reverse index over
  inherited/causal maps supersedes the earlier "lineage history on the Slice"
  source; the Slice keeps only the active set.

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
