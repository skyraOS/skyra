# Being / Slice / Relation Split

## The Tension

Binding realities (Thread, ActionHistory) currently sit on the Relation — the shared medium. But their durable state (member maps, weights) is the being's history, not the current signal. The being's private topology is on the wrong struct. Relationships on Skyra is redundant with the binding realities' member maps.

## The Split

Three layers, clean ownership:

### The Being — the membrane. Sensory and motor are the same cell.

The being is the boundary between inside and outside. Observe is sensory (reads the world, routes into binding realities). Express is motor (reads binding reality votes, coincidence detection, acts on the world). Same cell, two faces, already in the recursion as the top frame.

```go
type Skyra struct {
    // ... existing fields (Content, weights, DNA)
    // Observe = sensory (descent — routes impulse into brain)
    // Express = motor (ascent — reads votes, gates action)
    Brain map[string]Reality  // the being's binding realities
}
```

The Brain map holds binding realities. Binding realities are cells. Cells have members that are cells. All the way down. Same type everywhere. No separate cell index. You find things by traversing through activation.

A memory that lives in multiple binding realities isn't stored in multiple places. Same Reality instance referenced by multiple members maps. Thread points to it. Goal points to it. Same pointer. Different activation context depending on which binding reality you entered through.

### Three Structural Roles — All Emergent

Nothing is declared. What a cell does emerges from what it accumulated and where it sits in the traversal.

1. **The being (membrane)** — the top frame. Observe is sensory (reads the world, routes into binding realities). Express is motor (reads binding reality votes, coincidence detection, acts on the world). Not a cell type — it's the boundary.
2. **Binding realities** — cells with members. Indexes. They score their members against the impulse and surface what activates. The gate.
3. **Memories** — cells with content. Fresh ones have strong context bindings (thread, goal, emotional state). Old ones lost their bindings through consolidation and repeated retrieval — context-free knowledge. Traces and understandings are the same thing at different points on the binding decay curve.

A cell reads the Relation to determine its role. Position in the traversal determines what it does, not a type declaration.

### Slice — on the Relation. Transient. Per-being. Universal structure.

The physics of traversal. Same shape for every being regardless of DNA. How the signal propagates is universal — what makes beings different is their Brain and DNA, not how traversal works.

```go
type Slice struct {
    // traversal state
    TrainOfThought     map[string]Reality
    TrainOfThoughtKeys []string
    PreviousActions    []Reality
    Signal             float64
    Load               int
    Forming            *Skyra
    Surroundings       map[string]float64
    LastSeen           Reality
    Depth              int
    Trace              []string

    // fields — imprinted by this being's sensory/binding realities during traversal
    Fields map[string]map[string]float64  // nodeID → bindingReality → weight

    // goal and intent — reflected here transiently
    Goal   string
    Intent IntentGraph
}
```

Fields move from the Relation to the Slice. Skyra's binding realities imprint Skyra's Fields on Skyra's Slice. Builder's binding realities imprint Builder's Fields on Builder's Slice. Same Relation, different fields per being.

### Relation — the wave. Holds references.

The Relation carries Slices (per-being transient state) and references to each being's Brain (so binding realities can be reached during traversal). The Relation doesn't own any being's state. It's the medium.

```go
type Relation struct {
    // identity
    id      string
    Origin  string
    Impulse string

    // references — not owned, just pointers
    Brains map[string]map[string]Reality  // beingID → brain reference
    Slices map[string]*Slice              // beingID → transient traversal state

    // world
    World map[string]Reality

    // limits
    MaxLoad  int
    MaxDepth int
    Cancelled bool
}
```

## What Moves Where

| Currently | From | To |
|---|---|---|
| `Relation.Fields` | Relation | Slice |
| `Relation.Threads` | Relation | Skyra.Brain |
| `Relation.ActionHistories` | Relation | Skyra.Brain |
| `Skyra.Relationships` | Skyra | removed — Brain is the topology |
| `Relation.Brains` | (new) | Relation — reference to each being's Brain |

## How Traversal Works

**Entry:** Impulse arrives. Relation enters the being. TrainOfThought is empty. The being's Observe (sensory) reads the raw signal.

**Indexing:** The being routes into binding realities. Each binding reality scores its members against the impulse. Nodes above threshold across multiple binding realities form the subgraph — the intersection of what's relevant across every index. The impulse pulls a subgraph out of the being's topology through activation. Different impulse, different subgraph, same being.

**Descent:** The activated subgraph traverses. Each cell reads the Relation — TrainOfThought is now populated, so competence shifts. Cells contribute content, get picked up, shape activation for the next level. Signal attenuates.

**Ascent:** Binding realities read traversal evidence from the Slice → update their own durable weights in the Brain.

**Cell deposit:** When signals from multiple binding fields converge (high co-activation product), a trace forms. The cell gets added as a member to every binding reality that was active when it formed. Forming stays on the Slice as working state. On deposit, the formed cell moves to the Brain's binding realities as a durable member.

## Why This Is Right

- The being owns its history (Brain). The wave carries the working state (Slice). The Relation is just the medium.
- The being can't own its own traversal — traversal is physics, universal, same for everyone.
- Fields are per-being because different binding realities produce different imprints. They belong on the Slice, not the Relation.
- Memories live in binding realities. Descent targets come from binding reality member maps scored by activation. The binding reality IS the gate.
- The split mirrors biology: neuron morphology + synaptic weights (Brain) vs signal propagation mechanics (Slice) vs the signal itself (Relation).

## What Makes Beings Different vs Universal

**Different per being (Brain + DNA):**
- Which binding realities they have
- What members those hold and at what weights
- ObserveFns, ExpressFns, Primitives — behavior

**Universal (Slice):**
- TrainOfThought, Signal, Load, Surroundings
- Fields structure
- How traversal accumulates and attenuates
