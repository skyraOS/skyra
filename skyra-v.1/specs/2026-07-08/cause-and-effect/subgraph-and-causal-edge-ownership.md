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
Evidence --grades--> StructuralEdge
Evidence --grades--> EfficientCauseEdge
Evidence --grades--> FinalCauseEdge
```

Evidence provenance may be traversed from an edge for inspection, but Evidence
does not become a Material member merely because it grades the compound.

Praxis may resolve the active set to display:

- each causal paragraph and represented subgraph;
- the Efficient and Final relationships among summaries;
- supporting, opposing, and unresolved Evidence per edge;
- signed Identity/Familiarity state per contextual edge.

## Query And Activation

```text
present worker
  -> emit and persist G_present

efficient worker ─┐
                  ├-> emit G_prior/G_future and fresh causal edges
final worker ─────┘

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

**Inherited** contains the compounds this compound was COMPOSED FROM.
Birth-fixed, immutable, intrinsic: when a new cause is built from two prior
causes, that provenance is part of what the new compound IS, not part of any
explanatory context — a different relationship than Efficient or Final. It is
the durable home of composition (invariant 14). No new parser syntax is
needed: at Persist time, any resolved member that is a prior summary
(compound-typed) lands in the inherited port automatically — composition is
already expressed through stable-id referents; the port makes it traversable
as its own relationship instead of burying it in undifferentiated membership.
Storing it cannot leak because it never changes after birth.

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
14. The `inherited` port is birth-fixed composition provenance — populated at
    Persist from compound-typed members, immutable after, and a different
    relationship than Efficient or Final.
15. Efficient/Final ports accumulate edge occurrences (keyed by ID, stamped
    with their owning set); they never collapse into a singular cause of the
    compound. Context filtering is an observe-strategy choice, not storage.
16. Typed ports are substrate reform (`MemoryCell`), not kernel reform: Observe
    is the strategy selector, and the kernel traverses whatever surface it
    returns. Port names never cross the kernel boundary.
