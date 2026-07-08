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
Observe returns children/spaces to traverse.
Observe returns nil when there is no traversal surface.
Express returns a wave/result/condition/nil after traversal.
```

This removes the hidden-state dependency between `Observe` and `Children`.

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

## Membrane Example

Current code creates an ephemeral worker in membrane `Observe` and exposes it via
membrane `Children`.

Target:

```text
membrane Observe:
  read board/slice
  if impulse not tracked:
    append work item
  ask DAG/scheduler for next step
  create ephemeral worker node
  return {worker.ID: worker}
```

The worker is still a node. The kernel is not handed a function call. The worker
is just an ephemeral traversal surface for this wave.

No durable edge is written:

```text
membrane -> worker
```

is a per-wave traversal relation, not a `MemoryCell` child edge.

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
  report condition or emit wave
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
Express returns a wave or nil.
```

Examples:

```text
MemoryCell Express:
  return content or nil

motor Express:
  emit/return world impulse
  append invocation record

sensory Express:
  append result record
  report condition

membrane Express:
  if impulse complete: return nil / finish
  if more work remains: leave board item in-flight
```

The exact type of "wave" may remain `any` in the immediate migration, but the
semantics should be clear: `Express` is the outward/completion side of the node.

---

## Migration Plan

This is a large interface change. Do not do it incidentally while fixing the
affordance record model.

Possible migration:

```text
1. Add an adapter/helper conceptually named ObserveChildren.
2. Update the kernel internally to prefer returned observe children when present.
3. Migrate ephemeral workers first.
4. Migrate MemoryCell to return durable Children from Observe.
5. Remove Space.Children from the kernel contract.
6. Keep MemoryCell.Children as a durable graph API.
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
score returned children
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
2. Update every Space implementation.
3. Update relation.traverse to capture Observe return.
4. Update scoreChildren to accept returned children.
5. Remove Space.Children from the kernel interface.
6. Keep MemoryCell.Children as durable graph API.
```

Example target implementations:

```go
func (m *MemoryCell) Observe(origin skyra.Space, impulse any) map[string]skyra.Space {
    return m.Children()
}
```

```go
func (m *membraneRoot) Observe(origin skyra.Space, impulse any) map[string]skyra.Space {
    // create/select ephemeral worker
    return map[string]skyra.Space{m.worker.ID(): m.worker}
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
