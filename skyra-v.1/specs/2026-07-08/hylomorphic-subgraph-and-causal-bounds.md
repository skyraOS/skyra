# Hylomorphic Subgraph And Causal Bounds

**Date:** 2026-07-11 · **Status:** RESEARCH DRAFT / CANDIDATE MODEL REVISION
**Related:** [Schema / Aitia / Maxim / Evidence](./schema-aitia-maxim-evidence.md),
[Aitia Maxim Telos Schema Praxis](./aitia-maxim-telos-schema-praxis.md)

This document develops a candidate revision to the four-cause representation.
It does not yet supersede the accepted evidence-side model. Its central claim is
that Material Cause and Formal Cause should not be represented as two coequal
objects. Together they are one present-tense compound: a bounded subgraph whose
nodes are its matter and whose edges are its form.

## Thesis

Begin with Aristotle's four causes:

```text
Material Cause
Formal Cause
Efficient Cause
Final Cause
```

The candidate model collapses Material and Formal Cause into one primary object:

```text
G = (V, E)

V = material cause
    the components present in the compound

E = formal cause
    the arrangement that makes those components this compound
```

This is not intended as a simplification that discards either cause. It treats
their unity as primary. Matter and form are analytically distinguishable, but in
the present object they occur together as one hylomorphic compound.

```text
Material + Formal
  -> one bounded subgraph
  -> one present compound
```

## The Hylomorphic Compound

Consider a letter:

```text
material:
  words
  tokens
  phrases
  feelings expressed

formal:
  greeting
  body
  signoff
  ordering and relations among phrases
  author/addressee role arrangement
```

The letter is not first a bag of components and then a separately existing form.
It is the graph made by those components under that arrangement.

```text
LetterGraph = (letter nodes, letter edges)
```

The decomposition remains available as an analytical operation:

```text
same nodes, different edges:
  same matter, different form

isomorphic edge structure, relabeled nodes:
  same form, different matter

different nodes and different edges:
  different compound
```

The analytical distinction is preserved. The compound is ontologically and
computationally primary.

## The Three Bounds

Once Material and Formal Cause collapse into the present subgraph, three bounds
appear around it. They can be referenced by the present frame but cannot be
contained as parts of the same present standpoint.

```text
backward bound:
  Efficient Cause / inferred past

forward bound:
  Final Cause / imagined future

standpoint bound:
  occupant / irreducible point of view
```

“Cannot be contained” means they must not be collapsed into the central
present-tense subgraph. They may have graph representations of their own and may
be reached through pointers from the current frame.

## Backward Bound: Efficient Cause

Efficient Cause is not another component co-present with the present compound.
It is a backward-pointing causal hypothesis about what produced the compound.

```text
present frame
  --efficient-cause pointer-->
inferred prior frame
```

Any statement of Efficient Cause is reconstructive. The past frame is not
available in the present merely because the present bears its consequences.
Several different histories may produce the same current subgraph.

```text
Past A ─┐
Past B ─┼─> same present compound
Past C ─┘
```

Collapsing the inferred past into the present would erase the uncertainty that
makes causal inference necessary. Prior frames therefore remain separate,
pointer-connected structures with independently graded relationships.

## Forward Bound: Final Cause

Final Cause is the mirror orientation of Efficient Cause. It points toward a
future frame that does not yet exist and may never exist.

```text
present frame
  --final-cause pointer-->
imagined future frame
```

For the letter to Grandma:

```text
grandma reads the letter
the relationship is strengthened
the intended feeling is received
```

These are not parts of the present letter. They are possible future frames
toward which the present compound is directed.

Efficient and Final Cause therefore share one structural family with opposite
orientation:

```text
Efficient Cause:
  present -> hypothesized past

Final Cause:
  present -> imagined future
```

Both are external frame references. Neither is collapsed into the present
compound.

## Efficient And Final Cause Are Roles Of Other Compounds

An Efficient or Final Cause is not an intrinsically different kind of object.
The target of either pointer is another hylomorphic subgraph with its own
Material and Formal aspects:

```text
prior compound G(t-1):
  V(t-1) = its Material Cause
  E(t-1) = its Formal Cause

present compound G(t):
  V(t) = its Material Cause
  E(t) = its Formal Cause

possible future compound G(t+1):
  V(t+1) = its Material Cause
  E(t+1) = its Formal Cause
```

Relative to the present compound:

```text
G(t-1) plays the role of Efficient Cause
G(t+1) plays the role of Final Cause
```

Internally, however, both `G(t-1)` and `G(t+1)` remain ordinary hylomorphic
compounds. Efficient and Final Cause are therefore relational roles between
subgraphs, not intrinsic node kinds contained inside them.

```text
                 produces / explains
G(t-1) ─────────────────────────────────> G(t)
  │                                        │
  │ its own V/E                            │ its own V/E
  │ Material/Formal                        │ Material/Formal
  ▼                                        ▼
compound                                 compound

G(t) ───────────── directed-toward ─────────────> G(t+1)
                                                    │
                                                    │ its own V/E
                                                    │ Material/Formal
                                                    ▼
                                                  compound
```

The same immutable subgraph may play different causal roles from different
Frames:

```text
from its own Frame:
  present hylomorphic compound

from a later Frame:
  Efficient Cause / inferred producer

from an earlier Frame:
  Final Cause / imagined end
```

Role belongs to the directed edge and viewing Frame, not to the target
subgraph's immutable content.

This yields two useful directional pairs:

```text
current --explained-by--------> prior
prior   --efficient-cause-of--> current

current --directed-toward-----> future
future  --final-cause-of-------> current
```

The reverse edge in each pair is not assumed to have the same strength. As with
cause/effect prediction and explanation, opposite directions answer different
questions and carry independent contextual activation.

An imagined Final Cause may later be realized as a present compound. Realization
does not mutate the imagined Frame; it creates or identifies an actual present
Frame and records a graded relationship between the imagined and realized
subgraphs.

## Standpoint Bound: Occupancy

Perspective roles can be represented inside the hylomorphic subgraph:

```text
author slot
composer slot
addressee slot
observer slot
```

These slots are part of form. A letter is partly constituted by the arrangement
of who addresses whom.

Actual occupation of a role is different. The standpoint from which a frame is
experienced cannot be reduced to another object inside that same frame.

```text
shared letter subgraph:
  contains author-slot and addressee-slot

author frame:
  occupies author-slot
  points to shared letter subgraph

grandma frame:
  occupies addressee-slot
  points to shared letter subgraph

observer frame:
  occupies observer-slot
  points to shared letter subgraph
```

The perspective slots coexist in the shared graph. The acts of occupying them
remain separate frames.

```text
the eye is not in the visual field
```

## Governing Collapse Rule

> Collapse what shares a standpoint. Never collapse the standpoint into what it
> points at.

Material and Formal Cause share the here-and-now of one compound, so they merge
as the node and edge dimensions of one subgraph.

Efficient Cause, Final Cause, and occupancy are defined by their distance from
the present compound:

```text
Efficient Cause:
  exterior through temporal distance into inferred past

Final Cause:
  exterior through temporal distance into imagined future

Occupancy:
  exterior through perspectival irreducibility
```

Collapsing any of these into the present subgraph destroys the relation that
constitutes it.

## Frames And Pointers

A first conceptual representation is:

```go
type Frame struct {
    ID string

    // Present hylomorphic compound.
    CompoundSubgraphID string

    // External temporal bounds.
    EfficientCauseFrameIDs []string
    FinalCauseFrameIDs     []string

    // Occupancy remains frame-relative.
    OccupantEntityID string
    OccupiedSlotID   string
}
```

This is illustrative rather than an implementation commitment. Canonical graph
storage should still use immutable nodes and separate edge records.

The important topology is:

```text
                     inferred past frames
                              ▲
                              │ efficient-cause pointers
                              │
occupant frame ───────> present hylomorphic subgraph
                              │
                              │ final-cause pointers
                              ▼
                     imagined future frames
```

Perspective slots exist inside the subgraph. Occupancy edges originate from
separate frames.

## Two Degrees Of Camera Motion

A frame viewing the compound has two degrees of freedom.

### Vertical Motion: Zoom

Zoom changes descriptive resolution while retaining the same standpoint and
compound lineage.

```text
tokens
  -> words
      -> sentences
          -> letter
              -> gesture of care
```

Material cause stacks by level. What counts as a component at one resolution may
be a compound at a finer resolution.

Zoom therefore requires explicit relations among descriptions at different
levels:

```text
coarser-than
finer-than
composes
abstracts
```

Whether zoomed descriptions are the same immutable subgraph, linked subgraphs,
or separate Frames remains an integration question.

### The Firmest Specific Rung

> Find the most specific rung you can stand on firmly, and stand on it.

This is the underused third option between unsupported precision and total
abstention.

```text
unsupported precision:
  descend beyond the Evidence and invent detail

premature abstention:
  say nothing because the finest desired description is unsupported

firmest specific rung:
  choose the finest resolution fully grounded by the present frame
```

The LLM worker is allowed to choose its descriptive ground. It should attempt
the most specific causal statement or prediction available, then move to a
coarser rung whenever the next finer claim would require an unsupported
assumption.

```text
specific claim firmly supported:
  emit it

specific claim partly supported:
  move one rung coarser

coarser claim firmly supported:
  emit the coarser claim

no grounded rung:
  emit unresolved
```

Example:

```text
too specific:
  Michael avoids the email because he fears rejection.

firmly grounded:
  Michael avoids the email fearing wrong wording.

too vague:
  Michael has a problem.
```

The rule applies to both causal synthesis and Effect prediction:

```text
Cause worker:
  emit the most specific cause supported by observed structure

Effect worker:
  emit the most specific falsifiable consequence supported by the active frame
```

This is not permission to turn subjective model confidence into truth. The
chosen rung remains a hypothesis represented by immutable content and graded
contextual edges. Later Evidence may create resonance, tension, or residual
error that forces a different rung.

Prompt contract:

```text
Choose the most specific resolution fully supported by the supplied frame and
Evidence. If a detail requires an assumption not present in the input, move one
resolution rung coarser and try again. Do not invent detail to satisfy the
requested specificity. Emit unresolved only when no supported rung remains.
```

Numeric confidence should not be embedded in immutable node content. The worker
chooses a rung; edge grading and accumulated Evidence determine whether that
ground remains stable.

### Horizontal Motion: Pan

Pan changes the simulated anchor of “I.” It does not transfer actual occupancy.

```text
my frame
  -> render another occupant's frame
  -> inside my current frame
```

The movement is asymmetric. A system can simulate Grandma's standpoint from the
author's frame, but it does not become Grandma or acquire Grandma's irreducible
occupancy.

This yields a working definition:

> Empathy is rendering another's frame, at a chosen resolution, inside one's own
> frame.

The rendering should preserve provenance:

```text
simulating frame
simulated occupant
occupied perspective slot
chosen resolution
evidence used for the simulation
```

## Candidate Final Shape

```text
                            ┌─────────────────────┐
                            │ inferred past frame │
                            └──────────▲──────────┘
                                       │ efficient pointer
                                       │
┌────────────────┐        ┌────────────┴────────────┐
│ occupant frame │───────>│ hylomorphic subgraph   │
│                │occupies│ V = Material Cause     │
└────────────────┘ a slot │ E = Formal Cause       │
                          │ perspective slots      │
                          └────────────┬────────────┘
                                       │ final pointer
                                       ▼
                            ┌─────────────────────┐
                            │ imagined future    │
                            │ frame              │
                            └─────────────────────┘

camera:
  vertical = zoom through descriptive resolution
  horizontal = simulate another standpoint without transferring occupancy
```

## Consequences For The Current Schema Model

If accepted, this model revises the current invariant that Schema is exactly
four co-present Aitia nodes.

Candidate replacement:

```text
Schema/Frame
  present:
    one hylomorphic subgraph
      nodes = Material Cause
      edges = Formal Cause

  backward bounds:
    1:* efficient-cause roles pointing to hylomorphic prior Frames

  forward bounds:
    1:* final-cause roles pointing to hylomorphic future Frames

  standpoint:
    one occupant in the viewing Frame
    1:* perspective slots inside the compound
```

This would change Aitia from “one of four cause-field nodes” toward either:

```text
the causal relation between Frames
the worker that infers temporal pointers
or the complete bounded causal explanation
```

That naming decision must be settled before this draft supersedes the accepted
Schema/Aitia model.

## Relationship To Effect

Final Cause and predicted Effect must remain distinct:

```text
Final Cause:
  imagined future toward which the present is directed

Effect:
  falsifiable consequence predicted from the present causal structure
```

An Effect may point toward or away from a Final Cause. Evidence grades whether
the Effect occurred; later Praxis/Telos mechanics may grade progress toward the
Final Cause.

The candidate cause/effect structure is:

```text
present compound / Aitia
  --predicts--> Effect

Effect
  --explained-by--> present compound / Aitia

Effect
  --evidence--> observed result receipt
```

Forward prediction and reverse explanation are separate contextual edges. They
must not share activation state because `effect given cause` and `cause given
effect` are different relationships.

## Candidate Invariants

```text
Material and Formal Cause form one present hylomorphic subgraph.
Material is the node set; Formal is the edge structure.
Efficient Cause is a relational role played by an inferred prior compound.
Final Cause is a relational role played by an imagined future compound.
Every prior/future target remains its own Material/Formal hylomorphic subgraph.
Causal role belongs to the edge and viewing Frame, not immutable node content.
Past and future Frames are never collapsed into the present compound.
Perspective slots belong to the compound's form.
Actual occupancy belongs to a separate viewing Frame.
One Frame has one irreducible occupant.
Zoom changes resolution without changing occupancy.
LLM workers choose the finest descriptive rung fully grounded by current Evidence.
Unsupported precision backs off to a coarser rung before returning unresolved.
Pan simulates another occupancy without transferring it.
All durable nodes remain immutable; graph change creates nodes and edges.
Effect remains a falsifiable prediction, not Final Cause or Maxim.
```

## Integration Questions

1. Is `Schema` the viewing Frame, the central compound, or the complete bounded
   structure containing compound plus pointers?
2. What does `Aitia` name after the collapse: a causal pointer, a causal worker,
   or the complete explanation?
3. Are Efficient/Final pointers directed edges to whole Frames, Effects, or both?
4. How is uncertainty over several possible Efficient Cause histories graded?
5. How are several possible Final Cause futures represented without confusing
   them with predicted Effects?
6. Does every zoom level create a separate immutable subgraph, or do levels share
   nodes through abstraction/composition edges?
7. What exact edge records occupancy of a perspective slot?
8. How is a simulated pan represented and prevented from masquerading as actual
   occupancy?
9. Does Cause Evidence grade the present compound, temporal pointers, or both?
10. Does this model replace the accepted four-Aitia Schema immediately, or only
    after an implementation comparison?

Until these questions are settled, this document is research context and must
not silently change the accepted evidence-side implementation contract.

## Origin Example: Note To Grandma

The model emerged from analyzing a note to Grandma:

```text
present compound:
  the note as words-under-arrangement

efficient bound:
  the inferred history that led the author to write

final bound:
  Grandma reads; the relationship may strengthen

standpoint bound:
  author, Grandma, and observer occupy separate Frames

perspective slots:
  author/addressee roles inside the note's form

zoom:
  token -> phrase -> note -> gesture of love

pan:
  author simulates Grandma reading without becoming Grandma
```

The example is small, but the structure generalizes: one present compound, two
temporal bounds, one irreducible occupant per Frame, and a camera that can zoom
freely while panning only through simulation.
