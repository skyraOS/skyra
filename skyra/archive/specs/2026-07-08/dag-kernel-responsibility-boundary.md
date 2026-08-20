# DAG / Kernel Responsibility Boundary

**Date:** 2026-07-08 · **Status:** PARTIALLY SUPERSEDED RUNTIME DRAFT
**Parents:** [slice-based-dispatch.md](../2026-07-03/slice-based-dispatch.md),
[directed-affordance-invocation.md](./directed-affordance-invocation.md)

> **Supersession notice (2026-07-12):** The DAG-versus-kernel boundary,
> ephemeral-worker lifecycle, Board routing, and substrate-owned cascades remain
> relevant. The Maxim → Aitia → Entity grading order and all coupling to
> `schema-aitia-maxim-evidence.md` are superseded by the
> [layered graph](./cause-and-effect/subgraph-and-causal-edge-ownership.md) and
> its probe → compose+grade worker mechanism.

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

The scheduler/board should route by expressed worker results. Routing should not
be a side effect of doing all work in Observe.

---

## Membrane / Worker Lifecycle

The membrane is the impulse ingress and work-registration boundary. The scheduler
owns next-wave decisions.

Target lifecycle:

```text
1. impulse arrives at membrane

2. membrane Observe:
   read board/slice
   if impulse is not tracked:
     append work item / mark impulse in-flight

3. scheduler:
   read board/slice
   decide next step
   create/select ephemeral worker node

4. membrane Express:
   return Spacetime for the selected worker, or nil

5. Time kernel:
   Realize(st Spacetime)
   one wave records on the worker's slice

6. worker Observe:
   read the active graph roots it needs
   return edge-keyed traversal surface

7. kernel:
   traverses the exposed graph roots

8. worker Express:
   perform/commit the work
   append affordance invocation/result records
   update board / report result
   return next Spacetime value(s) or nil

9. scheduler/board:
   continue, wait, fan out, or finish/flush the impulse
```

The world/external caller already starts the membrane wave. The membrane does not
need to recursively spawn a wave on itself just to continue the same impulse. It
can finish the current wave by returning a `Spacetime` value for the next worker,
or nil if no work is ready.

---

## Ephemeral Workers And Slices

Current code may expose an ephemeral worker as a child for the current kernel
wave:

```text
membraneRoot
  -> ephemeralWorker
```

This is not a durable memory edge.

It is a per-wave traversal surface, stored on the ephemeral membrane transcript.
That keeps the kernel graph-native without writing worker nodes into the durable
`MemoryCell` graph.

However, this is only a transitional cheat. The cleaner target is one wave / one
slice:

```text
membrane wave:
  records on membrane slice
  Express returns Spacetime{Traversed: worker}

worker wave:
  records on worker namespace slice
  Observe returns active graph roots
  Express commits work and returns next Spacetime/nil
```

The important boundary remains:

```text
ephemeral worker relation = scheduling / per-wave state
MemoryCell children       = durable graph state
```

Do not store ephemeral workers as durable `MemoryCell` children. Prefer returning
`Spacetime` for the worker over descending into the worker from the membrane wave.

---

## Observe / Express Contract

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

Target kernel interface:

```go
Cost(origin, impulse) int
Observe(origin, impulse) map[string]Space
Express(origin, impulse) any
```

For ephemeral workers, the target contract is:

```text
worker Observe:
  read the corresponding active graph roots
  return edge-keyed spaces to the kernel

worker Express:
  perform/commit the affordance work
  append invocation/result records
  update board
  return next Spacetime value(s) or nil
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

This is the main migration away from the current implementation, where `Observe`
does all substantive work and `EphemeralRoot.Children()` is nil.

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

## Grading Episodes And Cascades

One logical grading episode may contain many kernel traversals. It is not one
recursive worker call and not one giant cross-slice wave.

```text
GradingEpisode
  -> grade Maxim
  -> mount local Maxim repair when required
  -> regrade against the same Evidence snapshot
  -> compute residual error
  -> stop or mount Aitia grading/repair
  -> stop or mount Entity/Schema grading/repair
```

Canonical worker order is sequential:

```text
Maxim -> Aitia -> Entity
```

Each stage reads committed results from the prior stage. A stage cannot consume
its own partial writes. Every traversal remains one wave recording on one Slice.

The Board owns episode state:

```go
type GradingEpisodeState struct {
    EpisodeID         string
    EvidenceSnapshot  []string
    OriginalAmplitude float64
    ResidualMagnitude float64
    CurrentStage      string
    AttemptsByStage   map[string]int
    TriedStructureIDs map[string]bool
}
```

Exact field types may change, but ownership does not. Workers receive only local
structure, the stable Evidence snapshot, incoming magnitude/error, and allowed
local operations.

```text
worker:
  perform local grade or reorganization
  Express committed result

Scheduler/Board:
  measure improvement
  compute residual error
  mount retry, next stage, or finish
```

The LLM is never told that a cascade is occurring and never chooses a higher
layer. The substrate applies the canonical principle:

> Each layer gets one or more chances to absorb prediction error locally. Only
> residual error propagates upward, and the substrate—not the LLM—decides whether
> that becomes a cascade.

Termination requirements:

```text
one stable Evidence snapshot per episode
accept reorganization only when fit improves
cap attempts at every stage
stop below residual threshold
escalate when local repair cannot absorb enough error
workers never recursively schedule themselves
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

## Settled Worker Details

```text
a cognitive affordance may be a leaf worker
leaf Observe returns no traversal surface
leaf Express commits/reports its result

Condition is Board/work-item state or an Express result
Condition is not a mutable routing side channel on EphemeralRoot

trace spans are appended from expressed/committed worker results
the orchestrator does not infer completion by peeking at root mutation
```

Moving existing gene code from `Observe` to `Express` is an incremental migration
task, not an architectural decision. New code follows the target contract;
legacy code may move one worker at a time while tests preserve behavior.
