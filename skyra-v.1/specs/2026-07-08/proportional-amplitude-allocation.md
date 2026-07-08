# Proportional Amplitude Allocation

**Date:** 2026-07-08 · **Status:** DRAFT
**Parents:** [kernel.md](../2026-06-16/kernel.md),
[activation-and-traversal.md](../../design-docs/kernel/activation-and-traversal.md),
[traversal-mechanics.md](../../design-docs/kernel/traversal-mechanics.md)

This spec records a traversal correction: child amplitude should not be split
equally across candidate children. The traversal should allocate amplitude
proportionally across the open edge-keyed children exposed by `Observe`.

The kernel owns generic transport math. `Observe` owns policy: which children are
exposed, what edge state exists, and what cognition-specific concepts such as
alignment, probability, relevance, or decay mean.

---

## Problem

Current traversal divides amplitude equally across all candidate children, then
uses each child's activation as a score:

```go
childAmp := (node.Amplitude * amplitudeDecay) / float64(len(candidates))
score := cand.activation * childAmp
```

That means edge strength affects whether a child survives, but it does not decide
how much wave energy that child receives.

Example:

```text
remaining amplitude = 100
edge products:
  A = 0.44
  B = 0.73
  C = 0.22
```

Current behavior with `amplitudeDecay = 0.9`:

```text
available = 100 * 0.9 = 90
childAmp = 90 / 3 = 30

A score = 0.44 * 30 = 13.2
B score = 0.73 * 30 = 21.9
C score = 0.22 * 30 = 6.6
```

This preserves equal energy distribution even though B is the strongest edge.

---

## Target Rule

Amplitude should split according to normalized positive transport products for
the open children:

```text
available = parentAmplitude * amplitudeDecay
sum       = Σ childEdgeProduct

childAmplitude = available * (childEdgeProduct / sum)
childScore     = childAmplitude
```

Using the same example:

```text
available = 100 * 0.9 = 90
sum       = 0.44 + 0.73 + 0.22 = 1.39

A amplitude = 90 * (0.44 / 1.39) = 28.49
B amplitude = 90 * (0.73 / 1.39) = 47.27
C amplitude = 90 * (0.22 / 1.39) = 14.24
```

The stronger edge receives more of the remaining traversal energy.

---

## Semantics

The edge product is the child's transport claim on the parent's remaining wave
energy.

```text
edgeProduct = generic active edge product for that parent->child edge
```

The kernel should not treat fanout as equal unless all open candidate edge
products are equal.

The kernel should not know the cognitive meaning of the product. It only sees the
edge-keyed traversal surface and the active edge values available on the slice.
The DNA/space layer decides which children are exposed and what the edge product
means.

This preserves the conservation intuition:

```text
sum(childAmplitude) = parentAmplitude * amplitudeDecay
```

The threshold should apply to the allocated child amplitude/score after
normalization.

---

## Zero And Negative Products

Open design details:

1. If all child edge products are `<= 0`, no child receives amplitude.
2. Negative products should probably not participate in the positive amplitude
   allocation pool.
3. A negative edge may still need to be recorded as inhibitory evidence, but it
   should not receive positive traversal energy from this pass.

Initial conservative rule:

```text
positiveCandidates = children where edgeProduct > 0
sum = Σ positive edgeProduct
```

Only positive candidates participate in proportional allocation.

---

## Merge Interaction

This split rule pairs with a convergence requirement:

```text
split:
  parent amplitude divides proportionally by edge product

convergence:
  multiple paths can point at the same durable node
```

For the first implementation, keep the traversal's strict visited-once rule.
Already-visited nodes are not re-entered during the same traversal pass.

Instead of sending amplitude into an already-visited child, remove that child from
the allocation pool and redistribute the available amplitude across still-open
children.

This avoids:

```text
- trapped amplitude,
- double charging token/count decay,
- re-entry races,
- late propagation queues,
- and cycle explosions.
```

It does not fully represent convergence yet. It is the stable first
implementation.

---

## Blocked Convergence

If a child was skipped only because it was already visited, record that blocked
convergence as traversal metadata/debug evidence.

Example:

```text
X -> A edge 0.70  // already visited
X -> B edge 0.20  // open
X -> C edge 0.10  // open
```

Do not allocate amplitude to `A` during this pass. Instead normalize over open
children only:

```text
open sum = 0.20 + 0.10 = 0.30

B amp = available * (0.20 / 0.30)
C amp = available * (0.10 / 0.30)
```

Record the blocked convergence:

```text
blocked:
  from: X
  to: A
  edgeProduct: 0.70
  reason: already-visited
```

This keeps traversal easy to reason about while leaving evidence for a later,
richer merge design.

---

## Future Merge Model

The richer future model may allow multiple active branches to point at the same
durable child and sum incoming branch amplitudes.

Example continuation:

```text
root -> A = 28.49
root -> B = 47.27
root -> C = 14.24

A -> D
B -> D
C -> D
```

`D` should not be treated as three unrelated visits. The traversal should merge
the contributions:

```text
D incoming amplitude = 28.49 + 47.27 + 14.24
D incoming amplitude = 90.00
```

If decay is applied on the hop into `D`, then:

```text
D amplitude = 90.00 * 0.9 = 81.00
```

The exact placement of decay must be consistent, but the topological rule is
stable:

```text
split by normalized edge product
merge by durable target node
split again from the merged node
```

Then if `D` has seven children, `D`'s merged amplitude becomes the parent
amplitude for the next proportional split:

```text
D available = D.amplitude * amplitudeDecay
D child amp = D available * (childEdgeProduct / sum(D child edge products))
```

The traversal result must also preserve provenance. Merging amplitude must not
erase how the node was reached:

```text
D reachedBy:
  root -> A -> D : 28.49
  root -> B -> D : 47.27
  root -> C -> D : 14.24

D mergedAmplitude:
  90.00 before hop decay
  81.00 after hop decay
```

This is a shift from path/tree DFS to wavefront propagation over a graph:

```text
current shape:
  recursively walk path branches
  produce a traversal tree

target shape:
  expand a frontier
  collect outgoing contributions
  merge contributions by durable target node
  traverse each merged target once per wavefront
```

Conceptual data structure:

```text
currentFrontier: map[nodeID]frontierNode
nextFrontier:    map[nodeID]frontierNode

frontierNode:
  cell
  mergedAmplitude
  incomingPaths[]
```

Each depth step:

```text
for each node in currentFrontier:
  compute positive child edge products
  split node.mergedAmplitude proportionally
  append contribution to nextFrontier[child.ID]

after all parents expand:
  sum contributions per child
  observe/traverse each merged child once
```

This makes convergence first-class graph evidence instead of an accident of DFS
ordering.

This future model is not the first implementation target. First implement:

```text
strict visited-once
proportional split over open children
blocked convergence records for already-visited children
```

---

## Migration Notes

Current implementation lives in:

```text
src/skyra/relation/relation.go
  scoreChildren
```

The current behavior is:

```text
equal amplitude split across candidates
child score = edge activation * equal child amplitude
```

Target behavior:

```text
measure all candidate edge products
drop already-visited children from the allocation pool
record dropped children as blocked convergence
normalize positive products across still-open children
allocate child amplitude proportionally
use allocated child amplitude as child score/survival value
descend with that allocated child amplitude
```

Tests should cover:

```text
parent amplitude 100, decay 0.9, edge products .44/.73/.22
expected child amplitudes: 28.49 / 47.27 / 14.24
```
