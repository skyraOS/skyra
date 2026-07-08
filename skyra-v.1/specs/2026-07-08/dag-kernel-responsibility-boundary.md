# DAG / Kernel Responsibility Boundary

**Date:** 2026-07-08 · **Status:** DRAFT
**Parents:** [slice-based-dispatch.md](../2026-07-03/slice-based-dispatch.md),
[directed-affordance-invocation.md](./directed-affordance-invocation.md)

This spec records a structural mismatch in the current dnav2 DAG implementation.
It is not blocking the immediate affordance-record unification, but it should be
fixed before the workflow model hardens.

---

## Problem

The current DAG/orchestrator is doing too much work that should belong to the
main kernel traversal model.

Today, most dnav2 genes do their substantive work in `Observe` on the
`EphemeralRoot`. The orchestrator then reads `root.Condition` after the wave and
routes to the next step.

That makes the DAG feel like the active executor:

```text
DAG creates worker
worker Observe performs the cognitive work
orchestrator reads condition
DAG routes next worker
```

This is operationally fine for the current build, but it muddies the intended
model.

---

## Intended Boundary

The DAG should not be the main executor of cognition.

The DAG's job is to select and mount the next ephemeral worker:

```text
DAG:
  choose step
  create ephemeral node
  attach the proper observe method
  attach the proper express method
  hand the node to the kernel
```

The kernel's job is to run the node through the normal traversal lifecycle:

```text
kernel:
  Observe = query / activation / intake strategy
  traverse children according to activation
  Express = perform/commit/report the affordance result
```

The worker's `Express` method is where the worker should do or commit the main
affordance work. The worker's `Observe` method should prepare/query/activate the
local traversal field.

Put differently:

```text
Observe asks: what is available / active here?
Express says: what happened after traversal?
```

---

## Current Mismatch

Current dnav2 pipeline genes mostly treat `Observe` as the work body:

```text
sensory:intake Observe:
  classify and record

midlevel:frame Observe:
  infer/create/edit gist

other:predict Observe:
  infer and append predictions

self:predict Observe:
  infer and append self proposal

motor:collapse Observe:
  infer concrete primitive and emit
```

`Express` is usually nil, and the orchestrator reads the condition directly from
the worker after the kernel wave.

That means the kernel is being used mostly as a bookend runner, not as the owner
of the observe/traverse/express process.

---

## Desired Shape

Each DAG step is an affordance invocation mounted as an ephemeral node.

The DAG creates the node. The kernel runs it. The node's methods implement the
kernel contract:

```text
Ephemeral worker
  Observe:
    read impulse / current graph
    expose candidate children or activation strategy
    prepare local state

  Children:
    expose query/traversal surface when the affordance needs one

  Express:
    commit the result
    append the affordance invocation record
    set/report condition
```

The orchestrator may still route by `(step, condition)`, but the condition should
be the worker's expressed result, not a side effect of doing all work in Observe.

---

## Membrane / Worker Lifecycle

The membrane is the impulse/work tracker and scheduling boundary.

Target lifecycle:

```text
1. impulse arrives at membrane

2. membrane Observe:
   read board/slice
   if impulse is not tracked:
     append work item / mark impulse in-flight

3. DAG/scheduler:
   read board/slice
   decide next step
   create ephemeral worker node

4. membrane exposes the ephemeral worker to the kernel

5. worker Observe:
   read the active graph roots it needs
   prepare the worker's traversal surface

6. worker Children:
   expose the prepared durable graph roots to the kernel

7. kernel:
   traverses the exposed graph roots

8. worker Express:
   perform/commit the work
   append affordance invocation/result records
   update/report condition

9. membrane Express:
   inspect the board
   if the impulse is complete:
     finish/flush
   otherwise:
     leave the work item in-flight for the next scheduling opportunity
```

The world/external caller already starts the membrane wave. The membrane does not
need to recursively spawn a wave on itself just to continue the same impulse. It
can finish the current wave after updating the board; the next scheduling
opportunity can be driven by the next membrane entry or an explicit scheduler
tick.

---

## Ephemeral Children Are Not Durable Graph Edges

The membrane may expose the ephemeral worker as a child for the current kernel
wave:

```text
membraneRoot
  -> ephemeralWorker
```

This is not a durable memory edge.

It is a per-wave traversal surface, stored on the ephemeral membrane transcript.
That keeps the kernel graph-native without writing worker nodes into the durable
`MemoryCell` graph.

Current code already does this shape:

```text
membrane Observe:
  create worker

membrane Children:
  return worker
```

That part is an acceptable cheat. The important boundary is:

```text
ephemeral worker child edge = per-wave traversal surface
MemoryCell children         = durable graph state
```

Do not store ephemeral workers as durable `MemoryCell` children.

---

## Observe / Children / Express Contract

Current kernel interface:

```go
Observe(origin, impulse)
Children() map[string]Space
Express(origin, impulse) any
```

`Observe` does not currently return the next child directly. Instead:

```text
Observe prepares local/per-wave state
Children returns the traversal surface prepared by Observe
Express commits results after traversal
```

For ephemeral workers, the target contract is:

```text
worker Observe:
  read the corresponding active graph roots
  cache them on the worker transcript

worker Children:
  expose those roots to the kernel

worker Express:
  perform/commit the affordance work
  append invocation/result records
  report condition
```

Worker root examples:

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

This is the main migration away from the current implementation, where
`EphemeralRoot.Children()` is nil and most genes do all substantive work directly
inside `Observe`.

---

## Why This Matters

This matters because the affordance-record model wants every meaningful step to
be a directed invocation:

```text
affordanceRecord
  -> from
  -> to
```

If the DAG is treated as the executor, records will tend to be centralized in the
orchestrator or in ad hoc Observe bodies.

If the kernel lifecycle owns execution, each ephemeral worker can record its own
invocation in the same place it expresses its result.

That gives a cleaner invariant:

```text
The affordance that expresses the result records the invocation.
```

---

## Migration Direction

Do not fix this as part of the first data-model pass unless it becomes necessary.

First pass:

1. Unify world-facing records (`engage`, `perceive`, motor-emitted actions) under
   the directed from/to record model.
2. Keep the current Observe-centered gene bodies if that keeps the change small.
3. Avoid baking the current Observe-centered shape into new abstractions.

Later pass:

1. Give each DAG step an explicit `Observe` and `Express` pair.
2. Move commit/report/record-writing behavior toward `Express`.
3. Let `Observe` become query/activation preparation.
4. Let the kernel traversal semantics matter again for ephemeral workers.
5. Keep the DAG as route selection and mounting, not as the effective executor.

---

## Open Questions

1. Does every cognitive affordance need children/traversal, or can some remain
   leaf workers whose `Express` commits immediately?
2. Should `Condition` remain a mutable field on `EphemeralRoot`, or become the
   returned value of `Express`?
3. Should the orchestrator append trace spans from `Express` results rather than
   reading `root.Condition` directly?
4. How much of the current gene code should move from Observe to Express during
   the affordance-record migration?
