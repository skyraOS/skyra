# Subgraph And Causal Edge Ownership

**Status:** DECIDED · 2026-07-12 (review folded)

**Placement:** this is SUBSTRATE reform, not kernel reform. The kernel's
contract is already "traverse what Observe returns" — port names never cross
that boundary. The cell's Observe implementation is the strategy selector: it
decides which port(s) to expose for a given impulse, and the kernel walks the
surface handed back. What changes is `MemoryCell` — typed child sets replacing
the flat `Children()` slice — to align the substrate with the kernel.
`relation.go` does not move.

## Question

Where do Material/Formal structure, Efficient Cause, Final Cause, and Evidence
live without collapsing distinct explanatory contexts or making the graph
untraversable?

## Vocabulary (agreed 2026-07-12)

Older docs say "compound"; the agreed name is CausalNode. One concept — the
CausalNode and "causal space" are the same thing; the store is just its
plural, not a second design object.

| name | what it is |
|---|---|
| **CausalNode** | the base node minted per emission; owns the three process maps + the parent pointer to its birth subgraph |
| **birth frame** | the parent nodes alive at spawn + the edge records minted at spawn fixing how they were arranged |
| **material node** | a parent node in V — Entity or Term |
| **Entity** | material node born from OBSERVATION (sensory/ingest) |
| **Term** | material node born from ASSERTION (mentioned, never observed) |
| **semantic edge** | immutable directed claim in the birth frame (formal, member); probe-time: index keys + verification; never graded, never weighted |
| **weighted edge** | historic relationship occurrence (inherited, efficient, final) + its live graded state; traversal-time: amplitude + belief |
| **node registry** | the knowledge graph's home for Entities + Terms and the label lookup |
| **AitiaSet** | unchanged |

## Two Node Contracts

There are two fundamental node contracts, and every structure in this doc
belongs to exactly one:

```text
knowledge graph (semantic, probe-time)
  material nodes: Entities + Terms
  semantic edges: formal + member records
  maps: term/label lookup,
        causalNodes[nodeID]        (reverse: node → CausalNodes containing it)
        byPair / byTriple           (reverse: directed semantic edge keys)

CausalNode (weighted, traversal-time)
  three process maps: inherited, efficient, final   (weighted edges)
  parent pointer: back to the knowledge subgraph minted at its formation
```

The birth subgraph is knowledge-graph CONTENT that the CausalNode references
— the parent pointer is the only bridge between the two contracts. Terms live
on the knowledge graph, never on the CausalNode. The reverse maps live on the
knowledge graph because they are built from semantic content; the process
maps live on the CausalNode because they are graded history.

**Within the contracts:** knowledge-graph nodes come in two element types —
semantic entities and semantic edges. CausalNodes come in ONE type. What a
CausalNode "is" — material/formal, efficient, final — is not a property of
the node; it is a PERSPECTIVE from whichever CausalNode is active. If A is
active, A is the material/formal cause (the present). If A holds B in its
efficient map, then B is "the efficient cause" only with respect to active A.
When B later activates (re-anchoring), B is the material/formal cause and the
roles around it reassign through B's own maps. Cause names are read off the
active node's edges, never written into a node's kind.

## The Evidence Collapse (decided 2026-07-12)

"Skyra said x" is a cause. Evidence is not a third contract — **an evidence
node is a CausalNode at a lower level of abstraction.** When an event is
observed, it is decomposed into a knowledge subgraph and born as an
observation-level CausalNode whose maps the ingest path STAMPS at birth:

```text
material[]   -> the decomposed observation's nodes   (knowledge subgraph, V)
formal[]     -> the decomposed observation's edges   (knowledge subgraph, E)
efficient[]  -> everything alive in the praxis frame when it was emitted
final[]      -> the action that was emitted
```

The final form, said for sure:

```text
knowledge graph -> semantic entity + semantic edge
causal graph    -> causal node    + causal edge
```

Two graphs, each one node type + one edge type — four structural types in the
entire cognition.

**The provenance stamp is mandatory content, never type.** Causal nodes carry
observed-vs-COMPOSED (mirroring observed-vs-ASSERTED on semantic entities).
Same behavior, same node type, distinguished by stamp. Without it grading
goes circular: a belief graded against another belief must be visible as
belief-on-belief; a belief graded against an observation-born node is what
"evidence" MEANS. Evidence is a perspective — a CausalNode is evidence with
respect to the edge it grades — the same relational move as invariant 20.
(The two stamps might be subtly different across the graphs; that parallel is
grounds to revisit the asserted side of the knowledge graph later — ledger
item 10.)

**Two fill mechanisms, one structure.** Observation-born nodes get their maps
stamped at birth by the ingest path — recorded, not inferred. Composed nodes
get theirs via probe → compose+grade. Workers are never sent to re-derive
what ingest already stamped.

**The identity this closes:** observation-born CausalNodes become PARENTS of
midlevel's present — "michael said 'can you help'" (observation-level) is
material for "michael wants help with the skyra project" (interpretation-
level). So the inherited edge to an observation-born parent IS the evidence
relationship: grading-against-evidence and grading-the-inherited-edges are
one mechanism. The reverse index pays double — one causalNodes[] lookup
serves "what have I seen" and "what have I concluded."

**Final-cause stamping is decided only for the self.** The final cause of
skyra's own action is *the final cause of everything that is efficient within
the frame* — the action inherits its purpose from what was live in the frame
that produced it (safe: we design from skyra's perspective). Stamping final
on an observation as a direct result of skyra's emission is dangerous and
stays OPEN (ledger item 11), as does decomposition timing/depth for incoming
nodes (ledger item 12).

## Two Relationship Layers

One Aitia set contains two different relationship layers.

### Internal Material/Formal compound

The present compound is one immutable subgraph:

```text
G_present = (V, E)

V = Material Cause: participating Entity, Term, or compound-summary nodes
E = Formal Cause: directed structural relationships among those nodes
```

Its summary node is the stable key and readable anchor:

```text
PresentSubgraphID
  -> immutable causal sentence or paragraph
  -> material nodes V
  -> formal edges E
```

For example:

```text
michael[1] asks[1→2] skyra[2]

Summary P
  owns structural occurrence:
    Michael --asks--> Skyra
```

Direction belongs to the structural edge record:

```text
FromID = michael
Label  = asks
ToID   = skyra
```

`Michael asks Skyra` and `Skyra asks Michael` are therefore different formal
structures even when they contain the same Material nodes and relationship
label.

### External causal beliefs

Efficient and Final Cause relate whole compound summaries:

```text
Summary E --EfficientCause--> Summary P
Summary P --FinalCause-----> Summary F
```

The prior, present, and future summaries each anchor their own internally
complete `G = (V, E)` compounds. Efficient and Final are roles of the directed
relationships among those compounds, not intrinsic kinds of their target nodes.

## Ownership Boundary

The internal structural edges belong to the immutable compound identified by
its `SubgraphID`.

The Efficient and Final edge endpoints are compound summaries, but the causal
edge occurrences belong to the contextual `AitiaSet`.

```text
AitiaSet
  PresentSubgraphID
  EfficientCauseEdgeID
  FinalCauseEdgeID
```

The two causal edge records resolve all three compounds:

```text
EfficientCauseEdge:
  FromID = PriorSubgraphID
  ToID   = PresentSubgraphID

FinalCauseEdge:
  FromID = PresentSubgraphID
  ToID   = FutureSubgraphID
```

The set need not duplicate the prior and future IDs because they are already
the causal edges' endpoints.

## Why Causal Edges Belong To The Set

One immutable present compound may participate in several explanations:

```text
AitiaSet S1
  Deployment --Efficient--> Outage --Final--> Restoration

AitiaSet S2
  Misconfiguration --Efficient--> Outage --Final--> Investigation
```

Both sets may reuse the same immutable `Outage` compound. They must not share
the contextual causal edges or their learned strength.

Storing one fixed Efficient and Final child directly as intrinsic content of
the Outage summary would overwrite, merge, or leak explanatory contexts.
Instead:

- compound nodes may be reused;
- structural and causal edge occurrences remain contextual;
- every new Aitia set receives fresh Efficient and Final edge IDs;
- edge activation and grade history never transfer automatically.

The failure mode is the SINGULAR child — "the cause" of the compound, which a
second set would overwrite. A port of edge OCCURRENCES has no such failure:
edge cells accumulate keyed by ID, each stamped with its owning set
(`ContextID`), each with its own signed-store history. Nothing merges, nothing
leaks. Ownership does not move — the set still owns the occurrence — the
summary's port is where those occurrences are reachable from.

**Invariant: the port accumulates occurrences; it never collapses them into a
singular cause of the compound.** Read without set-context, the port is a
distribution of explanations — every explanation this compound has ever
participated in. That is the retrieval surface retrospective citation (final
closure) and re-anchoring need, so the accumulation is a feature, not a leak.

## Traversal Shape

The semantic traversal shape is:

```text
AitiaSet
  ├── PresentSummary
  │     └── StructuralEdge occurrences
  │           ├── From Material node
  │           └── To Material node
  │
  ├── EfficientCauseEdge
  │     ├── From PriorSummary
  │     └── To PresentSummary
  │
  └── FinalCauseEdge
        ├── From PresentSummary
        └── To FutureSummary
```

A structural edge must preserve explicit endpoint roles. A generic ordered
child slice is insufficient as the sole source of meaning:

```text
StructuralEdge.Children = [Michael, Skyra]
StructuralEdge.FromID    = Michael
StructuralEdge.ToID      = Skyra
```

The children expose adjacency to traversal; `FromID` and `ToID` define the
directed claim.

## Reverse Retrieval Is Derived, Not Canonical

Entities may serve as gateways into relevant memory through derived indexes:

```text
summariesByEntity[NodeID] -> SubgraphID set

summariesByDirectedEdge[
  FromID,
  Label,
  ToID
] -> SubgraphID set
```

The directed-edge key preserves order. Candidate counts may determine initial
retrieval amplitude, but counts do not define semantic direction:

```text
candidateMass[SubgraphID] += matched query component weight
```

The runtime then walks and verifies each candidate's canonical directed
subgraph. Reverse indexes find where to walk; structural edges define what the
walk means.

## Evidence Ownership

Evidence remains immutable and external to the compound and Aitia set. It
grades relationship occurrences:

```text
Evidence --grades--> InheritedEdge
Evidence --grades--> EfficientCauseEdge
Evidence --grades--> FinalCauseEdge
```

Semantic edges are NEVER graded (corrected 2026-07-12 — an earlier draft
graded StructuralEdge). You never grade a CausalNode's internals; you grade
its RELATIONSHIPS — the weighted edges. If a claim inside a CausalNode turns
out wrong, that arrives as negative evidence on the weighted edges pointing
at that node, and retain-or-requery answers it by composing a successor —
never by reaching inside the immutable original, which stays intact as "what
I believed then."

Evidence provenance may be traversed from an edge for inspection, but Evidence
does not become a Material member merely because it grades the compound.

Praxis may resolve the active set to display:

- each causal paragraph and represented subgraph;
- the Efficient and Final relationships among summaries;
- supporting, opposing, and unresolved Evidence per edge;
- signed Identity/Familiarity state per contextual edge.

## One Worker Mechanism

Every compound worker runs the same two-step loop; the only parameter is
which edge receives the grades.

**Step 1 — probe.** The worker emits a probe subgraph. The probe drives a
reverse-index lookup over the relevant maps (midlevel: material/formal
summaries; efficient/final workers: the inherited nodes' causal maps).
Candidates come back with hit counts, and traversal amplitude distributes
over them proportionally — a summary with more occurrences gets more of the
wave (invariant 9's first half; the walk still verifies exact directed
structure before a candidate's evidence counts — the second half). Under that
amplitude the worker traverses the candidates' evidence.

**Step 2 — compose + grade.** With the retrieved summaries and their evidence
in its present, the worker composes the new summary — citing candidates by
stable-id, which lands them in `inherited` — and grades: a base grade for each
evidence relationship, seeded on the RECEIVING edge so a born edge never
deadlocks at zero (the CrossBirth pattern).

The receiving edge is the parameter:

```text
midlevel (present querier)
  -> grades land on the inherited edges
     (new-summary ↔ inherited parents)

efficient worker / final worker
  -> grades land on the causal edges
     (cause-summary ↔ present-summary)
```

Candidate sourcing is GRAPH-owned: the reverse index over inherited and
causal maps supersedes the earlier "lineage history on the Slice" candidate
source. The graph owns candidacy; the Slice keeps only the active set.

## Query And Activation

```text
present worker
  -> probe, compose, persist G_present + inherited edges

efficient worker ─┐
                  ├-> probe over inherited causal maps,
final worker ─────┘   compose G_prior/G_future + fresh causal edges

all three valid
  -> atomically persist and activate AitiaSet
```

If any worker exhausts format correction, no partial Aitia set activates. The
previous active set remains intact.

## Typed Child Sets

**Decision:** durable fields preserve edge identity and direction; typed child
sets expose that topology to the kernel. References needed for traversal must
not live only in opaque fields.

The cell interface requires typed adjacency:

```go
type ChildPort string

const (
    PortFormal    ChildPort = "formal"
    PortFrom      ChildPort = "from"
    PortTo        ChildPort = "to"
    PortInherited ChildPort = "inherited"
    PortEfficient ChildPort = "efficient"
    PortFinal     ChildPort = "final"
    PortEvidence  ChildPort = "evidence"
    PortSummaries ChildPort = "summaries"
)

type ChildSets map[ChildPort]map[string]Cell
```

The inner map is keyed by child ID to prevent accidental duplicate adjacency.
An equivalent ordered representation is permitted if it preserves stable child
identity separately from presentation order.

New traversal code requests a specific port:

```go
cell.ChildrenIn(PortFormal)
cell.ChildrenIn(PortTo)
cell.ChildrenIn(PortEvidence)
```

An untyped `Children()` may remain as a compatibility and inspection method by
flattening all ports. Cognitive traversal must not rely on the flattened view
when relationship role or direction matters.

### Summary cells

A summary cell carries three causal relationship ports plus its own structure —
three relationships, three lifecycles:

```text
PresentSummary
  formal:
    ├── asks-edge
    └── investigates-edge

  inherited:
    ├── prior-cause-summary-A
    └── prior-cause-summary-B

  efficient:
    ├── efficient-cause-edge (set S1)
    └── efficient-cause-edge (set S2)

  final:
    └── final-cause-edge (set S1)
```

**Formal** contains the compound's own contextual structural edge cells.

**Inherited** contains the compounds this compound was COMPOSED FROM. The
inherited relationship is an EDGE occurrence (new-summary → parent-summary),
fresh at composition, and it is where midlevel's evidence sits: graded with
respect to the NEW summary. Evidence is not a cause node and never lives on
a cause node — it lives on the relationship between the new base node and its
inherited parents. The edge-cell split applies: **topology is birth-fixed**
(these endpoints never change — this summary was composed from that one,
forever), **grades are live** (evidence against the new summary accrues over
time). When a new cause is built from two prior causes, that provenance is
part of what the new compound IS, not part of any explanatory context — a
different relationship than Efficient or Final, and the durable home of
composition (invariant 14). No new parser syntax is needed: at Persist time,
any resolved member that is a prior summary (compound-typed) gets a fresh
inherited edge automatically — composition is already expressed through
stable-id referents.

Composition also inherits the parents' CAUSAL NODES: the efficient and final
cause summaries reachable from the inherited parents seed the new node's
candidate maps. **Node-only inheritance** — the nodes carry over as
candidates; the parents' edge occurrences never do. A regrade against the new
summary mints a NEW efficient/final edge owned by the new context, with its
own evidence and signed-store history. Edge occurrences are never shared
across contexts (invariant 6 through composition).

**Efficient / Final** accumulate contextual causal edge occurrences — stored
children, keyed by edge ID, each stamped with its owning Aitia set. They are
not intrinsic universal causes of the reused summary; the accumulate-don't-
collapse invariant above governs them. An observe strategy that wants the
ACTIVE explanation filters by the active set's ContextID; one doing history
recall reads the whole port unfiltered. Same stored children, two strategies.

### Edge cells

An edge is a traversable cell with immutable topology and mutable contextual
activation:

```text
EdgeCell
  immutable topology:
    ID
    ContextID
    FromID
    ToID
    Kind
    Label

  mutable activation:
    Identity
    Familiarity
    GradeCount
    LastGradedAt

  child sets:
    from:
      └── FromCell
    to:
      └── ToCell
    evidence:
      └── zero or more Evidence cells
```

`FromID` and `ToID` are authoritative durable roles. The corresponding child
sets make those endpoints reachable by the kernel. Construction must reject an
edge when its endpoint fields and child references disagree.

```text
edge.FromID == edge.ChildrenIn(from)[0].ID
edge.ToID   == edge.ChildrenIn(to)[0].ID
```

Fields and typed children are created together before the edge is published.
Topology and endpoint child sets are immutable after publication. Activation
changes as grades arrive.

### Entity gateways

Entities expose reverse retrieval through a separate port:

```text
michael
  summaries:
    ├── summary-7
    ├── summary-12
    └── summary-20
```

`summaries` means that the Entity participates in those compounds. It is a
derived reverse index, not a causal or Formal edge from the Entity to the
summary. Active-context filtering and exact directed-subgraph verification
still occur after candidate retrieval.

### Evidence traversal

Evidence attaches through the edge's Evidence port:

```text
StructuralEdge
  evidence:
    ├── evidence-12
    └── evidence-19

EfficientCauseEdge
  evidence:
    └── evidence-24
```

This port is a traversable provenance projection. Adding later Evidence must
not mutate the immutable topology of the edge or compound. The implementation
must therefore keep Evidence adjacency in contextual activation/provenance
state, even when `ChildrenIn(PortEvidence)` presents it as children.

## Observe Strategies

Typed ports support separate walks without inventing separate graph models.

### Walk subgraph

```text
Summary
  -> formal
  -> EdgeCell
  -> inspect from
  -> continue through to
  -> Material nodes
```

### Retrieve summaries

```text
Entity
  -> summaries
  -> candidate Summary nodes
```

### Walk inheritance

```text
Summary
  -> inherited
  -> the prior summaries this compound was composed from
```

### Walk Efficient Cause

```text
PresentSummary
  -> efficient
  -> EfficientCauseEdge   (active strategy: filter by active set's ContextID;
  -> from                  history strategy: read the port unfiltered)
  -> PriorSummary
```

### Walk Final Cause

```text
PresentSummary
  -> final
  -> FinalCauseEdge
  -> to
  -> FutureSummary
```

### Inspect Evidence

```text
EdgeCell
  -> evidence
  -> Evidence cells
```

Forward traversal does not treat `from` and `to` as equivalent fan-out. `from`
establishes the source role and `to` is the forward continuation. A reverse
observer explicitly reverses that policy.

The kernel must alternate cell role and selected port while maintaining its
visited set and amplitude budget. Cycles remain legal; typed ports determine
which relationships a given observe strategy follows.

## Review Invariants

1. One `SubgraphID` identifies one immutable Material/Formal compound.
2. Formal direction is stored explicitly as `FromID`, `Label`, and `ToID`.
3. Structural edges belong to their compound occurrence.
4. Efficient and Final edge endpoints are summary nodes.
5. Efficient and Final edge occurrences belong to an Aitia set.
6. Reusing a compound never reuses contextual causal edges or learned strength.
7. Evidence grades edges and is not automatically Material content.
8. Reverse indexes are derived retrieval aids, not the source of graph meaning.
9. Candidate overlap may allocate amplitude; exact directed structure decides
   fit.
10. A complete Aitia set activates atomically.
11. Traversable references are exposed through typed child ports.
12. Edge endpoint fields and endpoint child references must agree.
13. Untyped flattened children are not authoritative when direction matters.
14. The `inherited` port holds composition-provenance EDGE occurrences —
    minted at Persist from compound-typed members, a different relationship
    than Efficient or Final. Topology is birth-fixed; grades are live:
    midlevel's evidence sits here, graded with respect to the new summary.
15. Efficient/Final ports accumulate edge occurrences (keyed by ID, stamped
    with their owning set); they never collapse into a singular cause of the
    compound. Context filtering is an observe-strategy choice, not storage.
16. Typed ports are substrate reform (`MemoryCell`), not kernel reform: Observe
    is the strategy selector, and the kernel traverses whatever surface it
    returns. Port names never cross the kernel boundary.
17. Causal inheritance is NODE-only: the parents' efficient/final cause nodes
    seed the new compound's candidate maps; a regrade mints a NEW edge owned
    by the new context. Edge occurrences are never shared across contexts.
18. Every compound worker runs probe → compose+grade. Candidate sourcing is
    graph-owned (reverse index over inherited/causal maps); the Slice keeps
    only the active set.
19. Two node contracts, one bridge: the knowledge graph owns material nodes,
    semantic edges, and the reverse maps; the CausalNode owns the three
    process maps and the parent pointer to its birth subgraph. Semantic
    edges are never graded — grade relationships, never internals.
20. CausalNodes have ONE type. Material/formal, efficient, and final are
    perspectives read from the ACTIVE CausalNode's edges, never properties
    written into a node's kind. Re-anchoring reassigns roles by changing
    which node is active.
21. Evidence is a CausalNode at a lower level of abstraction — observed
    events are decomposed into knowledge subgraphs and born as
    observation-level CausalNodes. "Evidence" is a perspective: a CausalNode
    is evidence with respect to the edge it grades. The provenance stamp
    (observed vs composed) is mandatory content, never type.
22. Two fill mechanisms, one structure: observation-born maps are stamped at
    birth by ingest (recorded, not inferred); composed maps come from
    probe → compose+grade. The final cause of skyra's own action is the
    final cause of everything efficient within the frame; final-stamping on
    incoming/other-agent events is open (ledger 11).
