# Space / Kernel Interface Reform

**Date:** 2026-07-08 · **Status:** DRAFT
**Parents:** [dag-kernel-responsibility-boundary.md](./dag-kernel-responsibility-boundary.md),
[directed-affordance-invocation.md](./directed-affordance-invocation.md),
[proportional-amplitude-allocation.md](./proportional-amplitude-allocation.md)

This spec records a target cleanup to the `skyra.Space` contract and kernel
traversal lifecycle.

---

## Problem

Current `Space` splits traversal discovery across `Observe` and `Children`:

```go
Observe(origin Space, impulse any)
Children() map[string]Space
Express(origin Space, impulse any) any
```

That means `Observe` must mutate hidden state if it wants to change what the
kernel traverses, and the kernel later asks `Children()` for the traversal
surface.

Current mental model:

```text
Observe:
  prepare children somewhere

Children:
  kernel asks what Observe prepared

Express:
  return content/result
```

This works, but it blurs two different concepts:

```text
durable graph children
per-wave traversal surface
```

For `MemoryCell`, `Children()` means durable graph edges. For an ephemeral worker,
children are not durable; they are the active roots selected for this wave.

---

## Target Interface

Make `Observe` return the traversal surface directly:

```go
type Space interface {
    ID() string
    Slice() Slice
    NewActivation() ActivationRecord
    Cost(origin Space, impulse any) int

    Observe(origin Space, impulse any) map[string]Space
    Express(origin Space, impulse any) any

    Agents() map[string]Space
    Content() any
}
```

Target mental model:

```text
Observe = what should the kernel look at next?
Express = what should happen after traversal?
```

Rules:

```text
Observe returns edge-keyed children/spaces to traverse.
Observe returns nil when there is no traversal surface.
Express returns a wave/result/condition/nil after traversal.
```

This removes the hidden-state dependency between `Observe` and `Children`.

The returned map is still intentionally simple:

```go
map[string]Space
```

The map key is the edge identity. The value is the target node.

```text
edgeKey -> target Space
```

This preserves the current edge-owned activation model without introducing an
edge wrapper object. The edge is represented by the key; the child is represented
by the value.

If a node needs to express weighting/selection policy, it does so by deciding
which edge-keyed children to return from `Observe`. The kernel should not know the
cognitive meaning of weights, relevance, alignment, probability, or decay.

---

## Durable Children Versus Traversal Surface

Durable graph children should remain a property of graph node types such as
`MemoryCell`, but they should not necessarily be the kernel's universal child
interface.

Target distinction:

```text
MemoryCell.Children()
  = durable graph topology

Space.Observe(...)
  = per-wave traversal surface
```

For a passive durable memory node, `Observe` may simply return its durable
children:

```text
MemoryCell.Observe(...) -> MemoryCell.Children()
```

For an ephemeral worker, `Observe` returns active roots selected for this wave:

```text
midlevel worker Observe:
  return active grouping/gist roots

self/other prediction Observe:
  return active gist and relevant result graph

motor Observe:
  return group/gist/self proposal/other proposal roots
```

This keeps durable topology and ephemeral scheduling concerns separate.

---

## Cost, Budget, And Amplitude

Each wave starts with an amplitude/budget.

That starting amount is a kernel concern because it defines the traversal's
transport envelope:

```text
Realize(spacetime)
```

The wave-start struct should be named `Spacetime`:

```go
type Spacetime struct {
    Origin    Space
    Traversed Space
    Impulse   Impulse
    Amplitude float64
}

type Time interface {
    ID() string
    Realize(st Spacetime)
}
```

`Spacetime` is the concrete event of a wave moving through a traversed space under
the time kernel. It is not an interface.

Amplitude belongs on `Spacetime`, not on `Impulse`.

Target split:

```text
Impulse:
  domain/event envelope

Spacetime:
  kernel wave-start metadata
  origin
  traversed space
  initial amplitude/budget
```

The principle is:

```text
one wave starts with one traversal budget/amplitude
```

Token/read cost belongs on the node, not the edge:

```go
Cost(origin Space, impulse any) int
```

Reason:

```text
edge = route/allocation from parent to child
node = content that costs tokens/read budget to include
```

So a rendered memory node can charge token cost, while a world point can charge
hop cost, and an ephemeral worker can charge zero or a small fixed cost.

Affordability remains a generic kernel check:

```text
if incoming budget/amplitude can pay target.Cost:
  traverse target
else:
  skip target
```

But the policy that decides which children are returned, which edges are exposed,
and what cognitive decay/weighting means belongs in `Observe`, not in the kernel.

---

## Kernel Policy Boundary

The kernel should own transport mechanics:

```text
- one wave starts with a budget/amplitude,
- one wave records on one slice,
- call Observe,
- traverse returned edge-keyed spaces,
- enforce visited-once,
- apply generic affordability,
- call Express post-order,
- record the traversal result tree and traversed edge keys.
```

The kernel should not own cognition-specific weighting mechanics:

```text
- alignment,
- probability,
- relevance,
- prompt token importance,
- cognitive decay policy,
- why one edge is stronger than another.
```

Those belong to `Observe` implementations and the DNA/space layer.

---

## Membrane Example

Current code creates an ephemeral worker in membrane `Observe` and exposes it via
membrane `Children`.

That works as a transitional cheat, but it mixes slices. The membrane wave records
on the membrane slice, while the worker should record on its own namespace slice.
The cleaner target is one wave / one slice.

Target membrane role:

```text
membrane Observe:
  read board/slice
  if impulse not tracked:
    append work item

membrane Express:
  ask scheduler for next worker wave
  return Spacetime or nil
```

The worker is still a node. The kernel is not handed a function call. But the
worker should usually run in its own wave on its own slice:

```text
membrane wave:
  slice = membrane slice
  Express returns Spacetime{Traversed: worker, Amplitude: ...}

worker wave:
  slice = worker namespace slice
  Observe returns active graph roots
  Express commits work and returns next Spacetime/nil
```

No durable edge is written:

```text
membrane -> worker
```

is a per-wave traversal relation, not a `MemoryCell` child edge.

Under the final model, even that per-wave membrane -> worker child relation may go
away. The membrane can finish by returning a `Spacetime` value for the worker
instead of descending into the worker in the same wave.

---

## Scheduler

Orchestration should move out of the membrane and into a scheduler/board helper.

The scheduler owns workflow decisions:

```text
given:
  board/work item
  impulse id
  last worker condition/result

return:
  next Spacetime value(s) or nil
```

Conceptual API:

```go
type Spacetime struct {
    Origin    Space
    Traversed Space
    Impulse   Impulse
    Amplitude float64
}

type Scheduler interface {
    Next(board *Board, impulseID string, result any) []Spacetime
}
```

This gives every `Express` the same shape:

```text
commit local work
update board
ask scheduler for next wave(s)
return wave(s) or nil
```

Examples:

```text
membrane Express:
  returns wave to sensory worker

sensory Express:
  returns wave to midlevel or grade worker

midlevel Express:
  returns waves to self/other workers

self/other Express:
  returns motor wave if join satisfied, else nil

motor Express:
  returns world wave
```

The membrane becomes ingress and board registration, not the central
orchestrator.

---

## Worker Example

Target worker lifecycle:

```text
worker Observe:
  read corresponding active graph roots
  return those roots to the kernel

kernel:
  traverse returned roots

worker Express:
  perform/commit work
  append affordance invocation/result records
  update board
  ask scheduler for next wave
  return wave or nil
```

Worker roots:

```text
midlevel:
  active groupID / active gist

self/other grading:
  active gistID plus prior proposal/result records

self/other prediction:
  active gistID and relevant exchange/result graph

motor:
  groupID
  gistID
  self proposal/process
  other proposal/process
```

---

## Express Return

Target mental model:

```text
Express returns Spacetime value(s) or nil.
```

Examples:

```text
MemoryCell Express:
  return content or nil

motor Express:
  return world Spacetime
  append invocation record

sensory Express:
  append result record
  update board condition

membrane Express:
  register/advance work item
  return next worker Spacetime or nil
```

The exact return type may remain `any` during the mechanical migration, but the
target semantic type is `Spacetime`: `Express` is the outward/completion side of
the node and may return the next spacetime wave(s).

The kernel can stay generic by treating returned `Spacetime` values as output; a
scheduler or caller may decide whether to immediately call `Time.Realize` for
those values.

---

## Migration Plan

This is a large interface change. Do not do it incidentally while fixing the
affordance record model.

Possible migration:

```text
1. Add an adapter/helper conceptually named ObserveChildren.
2. Add node Cost.
3. Update the kernel internally to prefer returned observe children when present.
4. Add wave budget/amplitude as explicit start-of-wave state.
5. Migrate MemoryCell to return durable Children from Observe.
6. Move membrane orchestration into scheduler/board helper.
7. Migrate workers to run as their own waves on their own slices.
8. Remove Space.Children from the kernel contract.
9. Keep MemoryCell.Children as a durable graph API.
```

Compatibility phase:

```text
old:
  Observe(...)
  Children()

bridge:
  ObserveChildren(...) map[string]Space

new:
  Observe(...) map[string]Space
  Cost(...)
```

The final API should make the kernel's traversal question explicit:

```text
node, what should I traverse from here for this impulse?
```

---

## Kernel Code Shape

The central kernel change is in `src/skyra/relation/relation.go`.

Current traversal flow:

```go
traversed.Observe(origin, impulse)

rootNode := r.rootNode(traversed, sl, amplitude)
r.scoreChildren(rootNode, sl)
```

`scoreChildren` currently reads children from the node:

```go
func (r *Relation) scoreChildren(node *skyra.Node, sl skyra.Slice) {
    for id, child := range node.Cell.Children() {
        ...
    }
}
```

Target traversal flow:

```go
children := traversed.Observe(origin, impulse)

rootNode := r.rootNode(traversed, sl, amplitude)
r.scoreChildren(rootNode, sl, children)
```

And during descent:

```go
children := childNode.Cell.Observe(origin, impulse)
r.scoreChildren(childNode, sl, children)
```

`scoreChildren` should take the traversal surface returned by `Observe`:

```go
func (r *Relation) scoreChildren(
    node *skyra.Node,
    sl skyra.Slice,
    children map[string]skyra.Space,
) {
    for id, child := range children {
        ...
    }
}
```

So the kernel lifecycle becomes:

```text
children := node.Observe(...)
score/filter returned edge-keyed children
descend into surviving children
out := node.Express(...)
```

The traversal tree can remain:

```go
type Node struct {
    Survivors []*Node
    Children  []*Node
}
```

But `Survivors` are built from `Observe`'s returned map, not from
`node.Cell.Children()`.

---

## Implementation Impact

Mechanical impact:

```text
1. Change Space.Observe signature.
2. Add Space.Cost or equivalent traversal-cost hook.
3. Update every Space implementation.
4. Update relation.traverse to capture Observe return.
5. Change Time.Realize to accept Spacetime.
6. Move Amplitude from Impulse to Spacetime.
7. Update scoreChildren to accept returned children.
8. Preserve one wave / one slice.
9. Move cross-slice workflow to scheduler-returned Spacetime values.
10. Remove Space.Children from the kernel interface.
11. Keep MemoryCell.Children as durable graph API.
```

Example target implementations:

```go
func (m *MemoryCell) Observe(origin skyra.Space, impulse any) map[string]skyra.Space {
    return m.Children()
}
```

```go
func (m *membraneRoot) Observe(origin skyra.Space, impulse any) map[string]skyra.Space {
    // ensure/track work item on the board
    return nil
}

func (m *membraneRoot) Express(origin skyra.Space, impulse any) any {
    // ask scheduler for the next worker wave
    return m.scheduler.Next(...)
}
```

```go
func (e *EphemeralRoot) Observe(origin skyra.Space, impulse any) map[string]skyra.Space {
    // run observe gene, select active roots
    return e.ActiveRoots
}
```

This makes the hidden-state bridge explicit. Nodes no longer need to mutate
private state in `Observe` and reveal it later through `Children`; they answer the
kernel's traversal question directly.

---

## Mini-Kernel Collapse Checklist

The current `src/skyra/traverse` package is a prototype traversal substrate used
by frame assembly and settle. It should collapse into the real `Time` kernel once
the `Space` contract can carry the missing mechanics.

### 1. Node Cost

Settled direction:

```go
Space.Cost(origin Space, impulse any) int
```

Token/read cost belongs on the node because cost is about the content being
included/traversed, not about the edge that reached it.

### 2. Affordability

Settled direction:

Once `Cost` exists on `Space`, the kernel can own generic affordability:

```text
remaining = currentAmplitude - node.Cost()
childAmplitude = ...

if childAmplitude >= child.Cost():
  traverse child
else:
  skip child
```

The kernel does not know why a child was exposed or what cognitive policy produced
its allocation. It only checks whether the traversal can pay the target node's
cost.

### 3. Slice Read Access

Settled direction:

The traversed `Space` owns the slice for the wave:

```text
slice = st.Traversed.Slice()
```

The kernel writes the traversal result tree to that slice. Callers such as frame
assembly should read the result tree from the traversed space's slice after
`Time.Realize(st)`.

The current `Slice.Tree()` exposes mutable internals. A cleaner future API should
add a read/snapshot method:

```go
type Slice interface {
    ...
    ReadCurrentTree() []*Node
}
```

or:

```go
type SliceSnapshot struct {
    Tree   TraversalPair
    Result any
    Maps   map[string]any
    Count  int
}

Read() SliceSnapshot
```

Exact API is open, but the direction is settled: result tree access belongs on
the slice, because one wave records on one slice.

### 3a. Active State History

Traversal tree, traversal trace, and traversal totals should be retained and
discarded together.

This is a recurring pattern:

```text
metadata about traversal done by cells belongs with the active state produced by
that traversal
```

Do not store tree history, edge trace history, and aggregate sums as unrelated
side channels. Store them as one bounded active-state snapshot.

Target shape:

```go
type ActiveStateSnapshot struct {
    Tick   int
    RootID string

    Tree   []*Node
    Trace  TraversalTrace
    Totals TraversalTotals
}

type TraversalTrace struct {
    Edges map[string]TraversedEdge
    Nodes map[string]TraversedNode
}

type TraversalTotals struct {
    Amplitude float64
    Score     float64
    Cost      int
}
```

Example edge trace:

```go
type TraversedEdge struct {
    EdgeKey   string
    FromID    string
    ToID      string
    Amplitude float64
    Score     float64
    Cost      int
}
```

Use a bounded history, implemented as a ring buffer or bounded deque:

```go
type ActiveStateHistory struct {
    capacity int
    items    []ActiveStateSnapshot
}
```

Public shape:

```go
type Slice interface {
    ...
    ActiveState() *ActiveStateHistory
}
```

Core operations:

```text
Append(snapshot)
Current()
Prior()
Recent(n)
```

When the history exceeds capacity, discard the oldest snapshot:

```text
drop oldest ActiveStateSnapshot
```

That drops the tree, edge trace, node trace, and aggregate sums together. This
keeps active-state retention coherent and avoids separate cleanup paths.

Do not use a raw doubly linked list unless arbitrary middle deletion becomes
necessary. The desired behavior is append-newest/drop-oldest, which fits a ring
buffer/deque.

### 4. Edge Eligibility

Open and related to settle.

The mini-kernel has:

```go
Result.Edges() map[string]float64
```

This means:

```text
which edges actually participated in this traversal?
```

Settle/credit use this eligibility set to decide which edges can be crossed,
graded, or strengthened.

The real kernel needs an equivalent over its `skyra.Node` tree:

```go
Edges(tree []*Node) map[string]float64
```

or a first-class edge field on `Node`:

```go
type Node struct {
    EdgeKey string
    ...
}
```

This is not a cognition policy. It is traversal accounting. But what consumes the
eligibility set is cognition policy, especially settle.

Under the active-state history model, edge eligibility is a derived view over the
current `ActiveStateSnapshot.Trace.Edges`:

```text
eligible edges = currentSnapshot.Trace.Edges
```

The kernel records traversal trace facts. Grading/settle decides what those facts
mean.

### 5. Frame Adapter

Settled direction:

The mini-kernel currently uses `frameNode` to adapt `MemoryCell` into:

```text
Cost
Frontier
Express
```

After the `Space` interface grows `Cost` and `Observe` returns an edge-keyed
traversal surface, this adapter should shrink or disappear.

Options:

```text
1. MemoryCell implements the needed behavior directly.
2. Keep a lightweight FrameView wrapper only for rendering-specific behavior.
```

The default target is to remove the separate `traverse.Node` adapter layer unless
frame rendering genuinely needs a view wrapper.

### 6. Frame Rendering Reads Slice Tree

Settled direction:

Frame assembly should run the real `Time` kernel, then read the result tree from
the traversed root's slice:

```text
Time.Realize(st)
tree := st.Traversed.Slice().ReadCurrentTree()
render tree
```

Each worker/root reads its own slice in `Express` or in the caller that requested
the frame.

### 7. Settle / Coherence Loop

Open and related to edge eligibility.

Settle is not generic kernel behavior. It is grading/coherence policy.

Current settle depends on the mini-kernel because it needs:

```text
- the lit frame,
- the traversed edge eligibility set,
- the active gist/framing,
- and repeated re-traversal after pruning.
```

Target direction:

```text
Board:
  tracks whether a gist/process needs settle, is settling, or is settled

Gist:
  owns the durable framing being settled

Settle worker:
  runs as a cognitive worker/process
  reads the real kernel traversal tree/eligible edges
  writes edge grades/coherence results
```

Potential graph shape:

```text
gist
  -> settle invocation
      -> settle result
```

Do not bake settle into the generic `Time` kernel. The kernel should expose the
traversal accounting settle needs; settle itself remains DNA/cognition policy.

### Summary

Settled direction:

```text
1. Cost on Space.
2. Generic affordability in kernel.
3. Result tree read from the traversed Space's slice.
5. Collapse or shrink frameNode adapter.
6. Frame rendering reads the real kernel tree.
```

Still open and linked:

```text
4. Edge eligibility accounting.
7. Settle/coherence consumption of eligible edges.
```
