# Sensory Split (was Brain Split) — Implementation Plan

## Phase 1: Add Brain to Skyra

Add `Brain map[string]Reality` to the Skyra struct. The brain is binding realities. Binding realities are cells. A Thread isn't a container that holds cells — it IS a cell whose members are other cells. All the way down. Same type everywhere. No separate cell index. No separate lookup. You find things by traversing the brain the same way you find things by traversing anything else — activation.

A memory that lives in multiple binding realities isn't stored in multiple places. It's the same Reality instance referenced by multiple members maps. Thread points to it. Goal points to it. Same pointer. Different activation context depending on which binding reality you entered through.

Initialize in `Init()` and `Create()`.

At bootstrap, thread and action history get added to `skyra.Brain` instead of `skyra.Relationships`.

```go
skyra.Brain["thread:main"] = thread
skyra.Brain["action:history"] = actionHistory
```

The Relation gets a `Brains` reference map so it can find any being's brain during traversal:

```go
Brains map[string]map[string]Reality  // beingID → brain
```

At world creation: `rel.Brains["skyra"] = skyra.Brain`

## Phase 2: Move Fields from Relation to Slice

Move `Fields map[string]map[string]float64` from the Relation struct to the Slice struct. Initialize it in `Impress()` and `GetSlice()`.

Update writers:
- `Thread.Observe` — writes to `sl.Fields` instead of `r.Fields`. Needs the beingID to get the right slice.
- `ActionHistory.Observe` — same.

Update readers:
- `Skyra.Activation` — reads from `sl.Fields` instead of `rel.Fields`. Needs to look up the slice for the being that owns this cell.

## Phase 3: Move Threads/ActionHistories from Relation to Brain

Remove `Threads map[string]*Thread` and `ActionHistories map[string]*ActionHistory` from the Relation struct.

Everything that currently reads `r.Threads` or `r.ActionHistories` instead looks up `r.Brains[beingID]` and finds the thread/action history there.

Update:
- `commands.go` — /createThread, /threads look up brain
- `primitives.go` — Listen looks up brain
- `present.go` — renders from brain
- `motor.go` — looks up action history from brain
- `relation.go` — snapshot methods iterate brain
- `main.go` — bootstrap puts them on brain, not relation

## Phase 4: Remove Relationships from Skyra

Remove `Relationships map[string]Reality` from the struct. Remove from `Init()` and `Create()`.

Descent in `Realize()` changes from iterating `s.Relationships` to iterating the being's brain's binding realities and their members. The binding realities provide the descent targets.

```go
// old: iterate everything unconditionally
for name, rel := range s.Relationships {
    r = rel.Realize(r)
}

// new: iterate brain's binding realities, score their members
brain := r.Brains[s.BeingID]
for _, bindingReality := range brain {
    // the binding reality itself fires (imprints fields)
    bindingReality.Realize(r)
    // then its members are the descent targets
    // scored by activation before traversal
}
```

All cells have relationships. What those relationships are depends on what the cell is. What the cell is depends on the order of the traversal — the cell reads the Relation to determine its role.

TrainOfThought empty → the cell is an entry point, an index. It routes into binding realities. TrainOfThought populated → the cell is deeper, it contributes content. Same type everywhere. Position in the traversal determines role, not a type declaration.

The impulse is the query. The binding realities are the indexes. Each binding reality scores its members against the impulse. Nodes above threshold across multiple binding realities form the subgraph — the intersection of what's relevant across every index. The train of thought is built by the impulse pulling a subgraph out of the brain through activation. Different impulse, different subgraph, same brain.

Descent becomes:
```go
// binding realities are the first layer — indexes
brain := r.Brains[s.BeingID]
for _, bindingReality := range brain {
    // binding reality scores its members, surfaces what activates
    bindingReality.Realize(r)
}
// train of thought now holds the activated subgraph
// deeper traversal follows the subgraph's own relationships
```

## Phase 5: Cell Formation

Cell formation is triggered by signal convergence — see "Cell Deposit" section below. Forming stays on the Slice as working state. On deposit, the formed cell gets added to the Brain's binding realities as a durable member of every binding reality that was active when it formed.

## Phase 6: Fix Look, Listen, Compress, Debug

- `PrimitiveLook` — scores brain members instead of Relationships
- `PrimitiveListen` — adds thread to brain instead of Relationships
- `ExpressCompress` — stores output into the active thread's members instead of Relationships
- `debugview` — walks brain instead of Relationships
- `commands.go` — /createThread adds to brain

## Cell Deposit — Signal Convergence, Not Placement

Cell deposit is not a separate mechanism per binding reality. It's the same physics everywhere — when signals from multiple binding fields converge on the same point during traversal, a trace forms there.

The hippocampus doesn't have a special "store episodic memory" mechanism. It has a connectivity pattern that makes it where co-occurring signals from different cortical areas meet. The trace forms there because that's where the signals intersected. The amygdala doesn't have a special "store emotional memory" mechanism. Threat signals and sensory signals converge there. The trace forms at the intersection.

In Skyra: high co-activation across binding fields = trace formation. Low co-activation = signal passes through without leaving a trace. The threshold for deposit is the product of the binding fields being strong enough together.

The cell doesn't get "placed" into a region. It forms at the intersection of the active fields and inherits membership in all of them. A cell born during a conversation about the goal while emotionally aroused belongs to the thread, the goal, and the emotional binding reality — because all three fields were high when it formed.

```
deposit trigger:
    read Slice.Fields for the current node
    compute product across all active binding fields
    if product > deposit_threshold:
        form cell
        add as member to every binding reality that was active
```

Forming stays on the Slice as working state (it's transient — part of the current traversal). On deposit, the formed cell gets added to the Brain's binding realities as a durable member. The Slice builds it, the Sensory keeps it.

## Resolved

1. **Generic member access on binding realities.** Binding realities are cells. Cells have members that are cells. Thread.Members becomes `map[string]Reality` — same type all the way down. Weights live on the member instances themselves (IdentityWeight, FamiliarityWeight, TraversalCount). The binding reality doesn't need a separate weight map — it reads activation from the member. No distinction between a binding reality and a cell. They're the same thing.

2. **Descent target selection.** By activation. Each binding reality scores its members using the activation equation. Members above threshold get traversed. Members below don't. This is the gate that was missing from the old Relationships iteration. The binding reality IS the gate.
