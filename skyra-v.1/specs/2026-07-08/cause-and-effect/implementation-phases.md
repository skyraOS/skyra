# Cause And Effect — Implementation Phases

**Date:** 2026-07-12 · **Status:** CURRENT BUILD PLAN

Each phase lands live, passes its tests, and is committed green before the next
phase begins. Decisions from
[minimal-loop-uncertainties.md](./minimal-loop-uncertainties.md) are resolved
just in time: a phase may not invent an unresolved mechanism merely to make its
implementation convenient.

**Canonical architecture:**
[subgraph-and-causal-edge-ownership.md](./subgraph-and-causal-edge-ownership.md).
That document overrides older Compound, separate-Evidence, and fixed
Material/Formal/Efficient/Final node models.

## Build boundary

Build inside `dnav2`; do not fork a v3 package. Reuse the orchestrator/board,
ephemeral workers, inference seam, logging, world, and TUI.

Typed child ports are substrate work, not kernel reform. A `MemoryCell` observe
strategy selects which typed ports to expose; the generic kernel continues to
traverse only the returned surface and never learns causal vocabulary.

The larger Spacetime/Cost/Scheduler reform remains a separate track. This plan
uses the current Slice to hold active context. The two tracks meet later when
history recall requires proportional traversal beyond the active context.

## Delivery rule

Every phase must include:

1. durable runtime shapes rather than temporary compatibility concepts;
2. focused unit and integration tests;
3. a live TUI or trace checkpoint proving the new path is exercised; and
4. no silent fallback that commits malformed or structurally incomplete model
   output.

---

## Phase 0 — Annotated subgraph parser — COMPLETE

The parser is the mouth every cognitive worker speaks through.

Delivered:

- annotated causal sentence parsing into a summary plus directed semantic
  subgraph;
- shared ordinals and stable-reference resolution;
- normalized-label Entity/Term reuse;
- transactional validation and persistence;
- bounded retry-with-correction for malformed output, with no AIMD tax;
- parser failures surfaced to the worker rather than converted into plain
  summaries; and
- golden tests for valid, malformed, reused, and multi-clause structures.

The parser preserves complete directed relationships. For example,
`michael[1] asked[1->2] skyra[2]` is not equivalent to the reversed sentence.

**Gate:** a corpus of annotated sentences round-trips deterministically and a
failed parse commits no partial graph.

---

## Phase 1 — Present-subgraph querier — COMPLETE

The former midlevel framing worker became the present-subgraph querier.

Delivered:

- the worker emits one readable causal sentence carrying its represented
  semantic subgraph;
- persistence creates an immutable semantic birth subgraph and its `CausalNode`
  anchor;
- the emitted sentence is the readable summary and the `SubgraphID` key;
- active present context is retained on the Slice; and
- the existing TUI and inference path can converse through the new format.

The current code is a Phase 1 foundation, not the completed causal loop. Legacy
storage literals and generic membership relationships may remain only where
explicitly identified for migration in Phase 2.

**Gate:** a live turn emits, parses, persists, and activates a directed present
subgraph with inspectable identity.

---

## Phase 2 — Traversable two-layer substrate — COMPLETE

Establish the permanent graph topology before grading it.

Build:

- typed child ports on substrate cells, including the minimum useful surfaces:
  `formal`, `from`, `to`, `summaries`, `inherited`, `efficient`, `final`, and
  the evidence/provenance projection where required by observation;
- immutable semantic edge records preserving `(FromID, Label, ToID)`;
- durable bridges between every `CausalNode` and its immutable semantic birth
  subgraph;
- graph-owned reverse indexes from semantic nodes and complete directed edge
  signatures to the `CausalNode`s whose birth subgraphs contain them;
- context-stamped weighted causal edge records with immutable topology and
  mutable activation state; and
- observe strategies that expose typed ports without leaking port names into
  the generic kernel.

The entity `summaries` port is a retrieval index, not a causal assertion.
Candidate occurrence counts may seed amplitude, but exact directed-subgraph
verification determines structural fit.

Migration work in this phase removes or confines legacy `compound` naming and
prevents generic member edges from standing in for inherited causal edges.

**Resolve at build:** none of the epistemic thresholds. This phase establishes
identity, ownership, adjacency, and traversal only.

**Gate:** given a directed query subgraph, the runtime can retrieve candidate
`CausalNode`s through reverse indexes, reject directionally incompatible
candidates, and traverse their typed causal ports.

**Landed 2026-07-12:** `MemoryCell` now owns authoritative typed child ports
with a compatibility-only flattened surface. Present persistence publishes
traversable semantic `EdgeCell`s, durable Material/Formal birth-subgraph
bridges, material-node `summaries` projections, graph-owned node/pair/triple
reverse indexes, exact directed-subgraph verification, and fresh weighted
Inherited/Efficient/Final edge construction backed by the shared signed store.
Tests prove endpoint field/port agreement, reverse-direction rejection,
node reuse with fresh edge identity, and isolation of contextual activation.
The Evidence port is vocabulary only until Phase 3 supplies observation-born
`CausalNode`s that can legally occupy it.

---

## Phase 3 — Observation-born causal nodes — COMPLETE

Make incoming world impulses part of the same causal graph instead of creating
a separate Evidence structure.

Build:

- one light LLM pass that decomposes each incoming utterance into the universal
  annotated sentence plus immutable semantic subgraph; observed structure is
  primary and speaker/text-only metadata is not a sufficient decomposition;
- one observation-stamped `CausalNode` anchored to that subgraph;
- mandatory observed/composed provenance on causal nodes and observed/asserted
  provenance on semantic nodes;
- birth stamping of relationships known from the ingest frame rather than
  re-inferring them with workers;
- immutable inherited ancestry edges from observation-born nodes into composed
  present interpretations; and
- inspectable provenance showing which observation supports which inherited
  relationship.

An observation-born `CausalNode` is Evidence only relationally. Once
decomposition commits, it becomes the shared input for two paths:
Material/Formal receives a structural retain/compose/requery decision, while
populated Efficient and Final beliefs receive independent signed grades.
Material/Formal is never graded and owns no EMA. There is no separate Evidence
node contract or Evidence persistence store.

```text
impulse
  -> light LLM decomposition
  -> persist observation-born CausalNode
  -> read active set from Slice
  -> structurally retain/compose/requery Material/Formal
  -> independently grade populated Efficient and Final beliefs
```

**Resolved:** ledger item 12. Utterances decompose eagerly at ingest through a
light, bounded LLM call before any grading worker runs. The emitted subgraph may
choose its warranted abstraction, but it must represent the observation's
structure rather than cutting down to transport metadata. Ledger item 11 is
required only for any incoming Final-cause stamp; the safe default is to leave
that stamp absent until its open half is decided.

**Gate:** an incoming utterance becomes a queryable observation-born
`CausalNode`, and a composed present node can cite it through a fresh inherited
edge without inventing a separate Evidence entity.

**Landed 2026-07-12:** external utterances receive a light structural LLM pass
with transactional parse/correction. Valid output persists through the same
semantic/CausalNode topology with mandatory observed/composed and
observed/asserted provenance. The observation ID rides its impulse's board row,
so concurrent turns cannot cross-link Evidence. Efficient/Final assessments run
only when those belief positions exist. Failed decomposition commits no
observation graph; malformed assessments receive bounded correction and commit
no result when exhausted.

The superseded present `CauseAssessment` was removed in Phase 4. Material/Formal
gets a structural operation, never an EMA. Efficient/Final assessment and EMA
integration remains with their causal-belief phase.

---

## Phase 4 — Present structural update loop — COMPLETE

Close the Material/Formal loop without inventing belief strength for the
present representation.

Build the sequence:

```text
semantic query subgraph
  -> reverse-index candidates
  -> exact directed-subgraph verification
  -> proportional amplitude allocation
  -> choose retain | compose | requery
  -> select or mint immutable present CausalNode
  -> fresh inherited edges only when composition occurs
  -> atomically retain or replace active PresentID
```

Semantic edges are immutable and unweighted; they are never graded.
Material/Formal itself is also ungraded: it is retained, composed, or replaced
as a whole immutable subgraph.

Ordinary impulse processing is not automatically composition. The
Material/Formal worker chooses the structural operation. Multiple inherited
edges belong only to a freshly composed representation and are never targets
of ordinary signed grading.

The first implementation retains strict visited-once traversal. More elaborate
same-wave convergence remains deferred.

**Resolved:** the exact retain/compose/requery contract and ledger item 9's
offer-only rule. Ledger item 6 moves to Phase 6 because no Efficient/Final
belief edges exist yet for Phase 4 to preserve or replace.

**Gate:** a fitting observation retains the active present node, an extension
can compose a successor with explicit ancestry, and a conflict can requery and
atomically replace `PresentID`. None writes a Material/Formal EMA.

**Landed 2026-07-12:** the live midlevel path accepts exactly `<retain/>`,
`<compose>...</compose>`, or `<requery>...</requery>`. Retain preserves the
active set and `PresentID`; compose must cite at least one `CausalNode` stable
ID offered by the current reverse-index pass and persists explicit inherited
ancestry; requery must be standalone and creates no inherited ancestry.
Reverse-index occurrence counts are normalized into candidate amplitudes before
the offers enter the structural present. Compose/requery output passes the
universal annotated parser before any persistence. Invalid operations receive bounded
correction; exhaustion leaves grouping, set, and present state unchanged.
Observation provenance attaches to composition edges without grading them. The
legacy `<new>/<edit>` framing protocol is no longer accepted by the live path.

---

## Phase 5 — AIMD specificity controller — COMPLETE

Make warranted specificity adaptive without conflating tokens, amplitude, and
edge strength.

Build:

- a substrate-owned `QueryTokenBudget` controller primitive for graded causal
  belief lineages;
- positive continuous grades earn tokens additively in proportion to grade;
- negative continuous grades reduce budget multiplicatively in proportion to
  tension;
- grade zero and no-grade hold the budget;
- parser-format retries do not change the budget; and
- traces for budget, outcome, emitted size, parse completion, and judged
  abstraction level.

`QueryTokenBudget` controls how specifically a worker may describe a subgraph.
Traversal amplitude controls how far or strongly a wave travels. Edge strength
records learned resonance or tension. They are three different quantities.

**Resolve before/during:** ledger item 4, initial value and bounds. AIMD alpha,
beta, and bounds are declared tuning parameters, not architectural questions.

**Gate:** controller tests preserve continuous grade magnitude, distinguish
zero from no-grade, and show that reduced budgets produce coarser complete
claims rather than truncated sentences. Live lineage integration occurs with
Efficient/Final.

**Landed 2026-07-12:** `QueryBudgetStore` owns one independent controller per
emitting LINEAGE — the Efficient and Final worker voices (corrected 2026-07-12
from per-edge keying: the budget is a property of the querying machinery, which
persists across sets, not of a contextual edge, which dies with its set).
Grades arrive per-edge and update the owning lineage's controller; each history
event records the sample's edge as provenance. Continuous positive grade
magnitude earns additive tokens; continuous negative magnitude applies
proportional multiplicative decrease; explicit grade zero and no-grade both
hold budget but remain distinct immutable history events. Invalid
grades/configuration are rejected without state. Budgets clamp to configured
bounds, initialize lazily at `B0`, expose current state and defensive history
snapshots, and are safe under concurrent lineage updates. DNA owns the store.
Default tuning is `32/128/256`, `alpha=10`, `beta=0.5`. Parser-format retries
never call the controller. The loop is closed: each expansion reads its
lineage's current budget into the worker's system prompt, and the substrate
verifies emission size (whitespace-field token proxy) — an over-budget emission
is a contract failure returned for bounded correction, producing a coarser
complete claim rather than a truncated sentence.

---

## Phase 6 — Efficient and Final causal expansion — COMPLETE

Extend the proven present loop with two additional perspectival relationships,
not two new node types.

Build:

- Efficient and Final workers using the active present birth subgraph as their
  shared query context;
- parallel fan-out with board declaration/mark/join and atomic active-set
  commit;
- each worker emitting the same universal `(sentence, subgraph)` contract;
- weighted contextual Efficient and Final causal edges between `CausalNode`s;
- independent lineage-keyed AIMD budgets for the Efficient and Final voices;
- the same compose/reuse and fresh-edge rules established in Phase 4; and
- edge-specific grading semantics: Efficient against records, Final against
  orientation toward the predicted or intended future.

Cause labels are roles read from edges relative to the active node. A prior or
future node remains an ordinary immutable `CausalNode` and may later become the
active Material/Formal present through re-anchoring.

**Resolved:** item 2's categorical fit is removed; item 3 defines defeat as
exactly zero strength with no threshold; item 5 requires full subgraphs; item 6
requires fresh edges and fresh signed state for every new Aitia set while
allowing node reuse (amended 2026-07-12: the lineage-keyed query budget
persists across sets by design — there is no edge-keyed budget to transfer).

**Gate:** one active set exposes a present node plus independently graded
Efficient and Final relationships. Failure in one lineage does not erase or
tax the others.

**Landed 2026-07-12:** Efficient and Final workers consume the same immutable
present snapshot and start in parallel. Each emits one complete annotated
subgraph with bounded correction and graph-owned candidate offers; neither
result persists unless both validate. `MintFullSet` atomically publishes one
PresentID plus fresh Efficient/Final edges. Worker failure preserves the prior
active set. Incoming observations grade the two active belief edges in parallel:
continuous grades update the exact signed EMA and the owning lineage's budget;
`<no-grade/>` updates neither EMA nor budget while remaining an instrumented
provenance event. Evidence is traversable from both edge cells. New sets may
reuse nodes but never edges or signed state; lineage budgets persist across
sets by design. `AllocateAmplitude` distributes remaining amplitude by absolute
signed strength; exact zero and absent alignment receive no excitation. Budget
enforcement is substrate-side verification with bounded correction, not
provider-side truncation: each worker reads its lineage's current budget in its
system prompt and an over-budget emission is re-corrected onto coarser ground.

---

## Phase 7 — Inspection, history, and hardening — COMPLETE

Make the substrate observable enough to evaluate rather than merely admire.

Build:

- a read-only Praxis/debug screen showing active birth subgraphs, directed
  semantic structure, causal relationships, provenance, grades, Identity,
  Familiarity, amplitude allocation, and token-budget histories;
- context-filtered and unfiltered views of accumulating causal edge
  occurrences;
- history/provenance walks through typed ports; and
- an experimental closure checker for the candidate rule that a prior Final
  target gains support when later reality cites it retrospectively through an
  Efficient explanation.

The closure rule is a candidate to test, not an assumption to bury in the
storage model. Maxim, Effect, and action-producing Praxis mechanics remain
deferred.

**Decision (2026-07-12):** inspection and provenance use deterministic typed-
port walks that do not consult edge strength. Cognitive recall remains weighted;
no artificial birth strength or exploration floor was introduced. The live TUI
opens the read-only Praxis screen with `ctrl+p`. The Final-closure checker is a
read-only candidate reporter and cannot grade edges or alter topology; keep or
reject the candidate based on observed runs.

**Gate:** every active belief and its support path can be inspected; history
recall does not make fresh zero-strength relationships permanently unreachable;
and closure experiments can be evaluated without changing graph topology.

---

## Phase-to-ledger map

| Phase | Ledger decisions | Main durable result |
|---|---|---|
| 0 — complete | 1 — decided | universal annotated parser |
| 1 — complete | 7 — decided | active present CausalNode |
| 2 — complete | none | typed two-layer topology and reverse retrieval |
| 3 — complete | 12; limited 11 | observation-born CausalNodes |
| 4 — complete | structural contract; 13 decided | Material/Formal structural loop |
| 5 — complete | 4 — decided | continuous AIMD controller primitive |
| 6 — complete | 2, 3, 5, 6 — decided | Efficient/Final fan-out and grading |
| 7 — complete | 8 — decided; closure experiment | inspectable history and hardening |

Ledger item 10 remains a non-blocking revisit after the Evidence collapse has
been exercised in production-shaped sessions.

## What V0 means

V0 is complete when this runs live end to end:

```text
impulse
  -> observed semantic subgraph + observation-born CausalNode
  -> query present by subgraph
  -> retrieve, compose/reuse, and activate
  -> expand Efficient and Final relationships
  -> grade fresh contextual causal edges
  -> update signed EMAs
  -> allocate amplitude
  -> retain or requery
  -> adjust each lineage's token budget
```

Every claim Skyra holds must remain inspectable as a readable sentence, an
immutable semantic birth subgraph, and a set of contextual weighted causal
relationships. No action-side reward or Praxis emission mechanism is required
for this evidence-side V0.
