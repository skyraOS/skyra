# Skyra: A Traversal Runtime for Agentic Cognition

## Thesis

A traversal runtime is a better foundation for agents than an orchestration pipeline.

Most agent frameworks coordinate behavior from the outside: assemble context, call a model, parse output, route tools, update memory. Skyra moves the load-bearing work into a substrate. A Relation moves through Reality graphs. Each being interprets that wave through its own Slice, Brain, and DNA.

The goal is not to remove structure. The goal is to move structure into stable mechanics:

- traversal
- activation
- attention
- binding fields
- prediction
- stress
- memory formation
- motor convergence

The runtime should stay small. Behavior should grow from topology, activation records, and binding realities.

## Core Split

Skyra has three core runtime contracts:

```go
type ActivationRecord interface {
    Activate(visited bool, traversalCount int) float64
}

type IRelation interface {
    ID() string
    Realize(beingID string, brID string, root Reality)
    GetSlice(beingID string) *Slice
}

type Reality interface {
    ID() string
    NewActivation() ActivationRecord
    Observe(r *Relation)
    Express(r *Relation)
    Children() map[string]Reality
}
```

These are separate things:

- **Reality** — the inhabitant. A cell, being, place, tool, memory, door, binding reality, or action leaf.
- **Relation** — the wave. The DNA-independent traversal carrier.
- **ActivationRecord** — DNA-owned activation state and scoring behavior.

The Relation traverses Realities. It does not become them.

## Kernel Mechanics vs DNA Behavior

The traversal kernel owns mechanics:

- start traversal
- call `Observe` on descent
- call `Express` on ascent
- call `ActivationRecord.Activate`
- apply threshold
- apply energy
- maintain DFS stack
- build the traversal tree
- preserve `Prior`
- construct `Current`
- expose auditable records

DNA owns behavior:

- how activation is computed
- what bump and decay mean
- what a score means
- what `Observe` does
- what `Express` does
- what children exist
- how competence selects phase behavior
- how logging explains local meaning

The kernel does not decide whether a being is wise. It guarantees the traversal was mechanically valid.

## Relation

The Relation is the current wave.

It carries current-wave properties such as:

- impulse
- energy
- traversal identity
- references needed to traverse

It does not own durable being state. Ownership is separate from reachability. The Relation may hold references to Realities so traversal can occur, but durable state belongs to the being, its Slice, and its Brain.

The same Relation mechanics can pass through different DNA boundaries. Each being interprets the wave locally.

## Slice

The Slice is durable per-being activation state. It is the being's universal memory manager / active working set.

It owns:

- activation records
- traversal count
- thresholds
- surroundings / attention state
- prior and current traversal trees
- stress signal
- per-binding-reality traversal pairs

The Slice is not a temporary object on the Relation. It survives across Relations.

The useful analogy is memory management:

- global store / world state = backing store
- Slice = per-being working set
- activation record = activity metadata
- attention = currently active mapped subset
- retrieval = page-in from external archive when detail is missing

## Brain

The Brain holds specialized durable binding realities. The Slice is universal. The Brain is specialized.

Examples:

- Episodic
- Goal
- Motivation
- Action
- Emotion
- Prediction
- Coherence
- Prospective
- WorkingMemory

Binding realities do not own the traversal kernel. They interpret the Slice, write fields, record deltas, and contribute pressure signals.

## One Traversal

The current traversal kernel is iterative DFS.

```
Begin traversal
  shift Current -> Prior
  clear Current and stack
  root.Observe
  score root children
  push root node

Loop
  if top has unvisited survivor and Relation has energy:
      spend energy
      child.Observe
      score child children
      push child node
  else:
      pop top node
      top.Cell.Express
      attach popped node to parent or Current root
```

This gives O(n) tree construction:

```
push each visited node once
pop each visited node once
attach to parent in O(1)
```

The traversal itself creates the audit record. There is no second pass to reconstruct the tree.

## Prior and Current

The Slice keeps separate traversal buffers:

- **Prior** — last completed traversal tree. Read-only.
- **Current** — traversal tree being built now.
- **Stack** — DFS state machine for the current traversal.

Binding realities read `Prior`. The current traversal writes `Current`. Nothing reads a half-built tree.

This gives snapshot isolation:

```
Prior + Relation + Current + Delta = decision record
```

## Activation

Activation is being-relative. A cell should not own "how strongly this being remembers me."

The kernel calls:

```go
record.Activate(visited, traversalCount)
```

The DNA decides what that means.

The general model remains:

```
score = identity * familiarity * recency * Π(binding_fields) * Π(contextual_fields)
```

But that equation belongs behind `ActivationRecord`. The kernel should not know whether a DNA uses EMA, recency, stress, trust, or something else.

## Competence

Competence is the pattern for phase-specific behavior inside a stable cell.

The cell's role stays stable. Its behavior changes by phase.

Examples:

- Membrane Observe: `efference` or `exafference`
- Episodic Observe: score overlap or deposit frame
- Motivation Observe: passive dependency check or verification result
- Action Express: fire primitive or record outcome

This preserves uniformity without splitting a cell into artificial subtypes.

## Membrane

The Membrane is the being boundary.

It faces both directions:

- inward to Brain / Slice traversal
- outward to world propagation

Sensory prediction lives in Membrane `Observe`.

Observe has two core modes:

- **efference** — self-signal. Motor just fired. Receive prediction payload and prime binding realities.
- **exafference** — other-signal. External input arrived. Compare against primed prediction. Surprise passes through.

Prediction is not a separate pipeline stage. Prediction is sensory competence.

## World, Surroundings, and Attention

World state and being attention are different.

- **WorldState / Surroundings** — what exists or propagated to the being from the world.
- **Slice.Attention** — what this being currently attends to after resonance and prediction filtering.

The reception pipeline:

```
Express
  -> Propagate
    -> Surroundings
      -> Resonance
        -> Prediction gate
          -> Attention
            -> Binding realities
```

Episodic reads Attention, not raw world state.

## Resonance

Reception has two sides:

- sender determines force
- receiver determines resonance

```
impulse_strength = force * resonance
```

Force is outward energy. Resonance is membrane reception.

Resonance can come from:

- primitive alignment, such as listening or looking toward the source
- goal relevance, such as hearing your name or a dependency cue
- semantic match to active goals

The world should not become a smart router. The world propagates. The being resonates.

## Propagation

Propagation is outward traversal through the world graph.

Inward traversal:

```
Relation energy descends through a being's inner graph.
```

Outward propagation:

```
Expression force spreads through the world graph.
```

The mechanics are structurally similar: score, threshold, spend energy, traverse survivors, record what was reached. The map changes. The traversal kernel should remain stable.

## Episodic Memory

Episodic does not store memories as transcript blobs. It stores indexes over lived frames.

An episode is a sequence of frames:

```
Episode
  Frame 0: cells A, B, C co-present
  Frame 1: cells A, D, E co-present
  Frame 2: cells D, F co-present
```

Each frame is an encoding pulse: what was co-present after prediction filtered it.

Episodic maintains sparse inverted indexes:

```
cellID -> []episodeID
conjunctionHash -> []{ episodeID, frameID }
placeID -> []episodeID
deviceID -> []episodeID
toolID -> []episodeID
goalID -> []episodeID
```

The current context is the query. People, places, devices, tools, goals, and actions are Realities in Attention. They participate in episodic overlap.

This makes context-dependent memory substrate-native:

```
current Attention
  -> sparse episode lookup
  -> frame overlap
  -> binding fields
  -> activation
```

## Motivation

Motivation tracks whether the reason for a goal still holds. Goal tracks progress. Motivation tracks validity.

Motivation maintains dependencies:

```
goalID -> []dependency cell IDs
```

If a dependency is present in Attention, Motivation can verify it passively. If it fades from Attention, Motivation does not assume failure. It marks uncertainty and may raise an epistemic goal to re-ground the dependency.

This is a built-in hallucination pressure brake:

```
stale assumption fades
  -> dependency uncertainty rises
    -> epistemic goal forms
      -> look / retrieve / inspect
        -> dependency re-grounded or invalidated
```

## Stress

Stress is not unresolved persistence. A long-running goal can be calm.

Stress is persistent concern plus instability:

```
stress = persistence * instability
```

Persistence:

- active across traversals
- survived episode boundaries
- carried forward by Motivation

Instability:

- goal error trend
- dependency uncertainty
- prediction error
- action failure
- field disagreement
- load pressure
- energy pressure
- concurrency pressure
- episode churn

Stress lives on the Slice. Binding realities contribute components. Stress changes memory formation, compression pressure, salience, and episode segmentation sensitivity.

High stress lowers the segmentation threshold:

```
effective_segmentation_threshold = base_threshold / (1 + stress)
```

This produces smaller, denser episode fragments during unstable periods. Later relief can compress fragments into understanding.

## Retrieval, Notes, and RAG

RAG is not the cognitive substrate. Retrieval is an external support primitive.

The substrate remembers enough to know when detail is missing:

- dependency uncertainty
- weak episodic overlap
- stress
- failed prediction
- blocked goal

Then a being can learn to use:

- notes
- archives
- transcript lookup
- vector RAG
- graph RAG
- search
- filesystem

External retrieval re-grounds missing detail. It does not replace memory.

## Act Service

External tools and retrieval systems live behind leaf boundaries.

```
Runtime = cognition
Act service = consequence / external evidence
Pipe = boundary
```

An action leaf serializes a block call:

```json
{
  "block": "search",
  "params": {
    "query": "..."
  }
}
```

The act service executes and returns JSON. The runtime never handles credentials. The being only reaches what its topology exposes.

This is where existing systems belong:

- Composio
- Nango
- n8n
- OpenClaw
- Hermes
- Windmill
- NocoBase
- local RAG
- local notes
- filesystem archives

## Concurrency

Concurrent traversal should not corrupt Slice state.

Future shape:

- each Relation gets private run state: stack, Current, Delta
- durable Slice state is updated at commit
- commit is short and explicit
- inference can use a coalescing gate

If Relation A reaches inference and Relation B arrives, B can be buffered. It may be visible to A's inference as "arrived during inference window," but it should not mutate A's activation tree. After A commits, buffered Relations replay.

## Swarms

Swarms should emerge from learned action topology, not from a manager bolted on top.

A large task can learn concurrent motor patterns:

```
goal pattern
+ dependency graph
+ action history
+ stress/load profile
=> fire multiple action leaves concurrently
```

Concurrency becomes a learned reflex:

- one primitive for small tasks
- sequence for medium tasks
- concurrent set for larger tasks
- recruited beings/specialists for still larger tasks

## Test Rooms

The architecture should be proven in worlds, not argued abstractly.

Rooms:

- **Escape Room** — does the whole loop work?
- **Maze Room** — can the being maintain orientation across depth?
- **Social Room** — can beings coordinate without orchestration?
- **Decay Room** — can the being prioritize under scarcity?
- **Repeat Room** — does repeated experience become cheaper?
- **Distractor Room** — can goal and motivation prevent context bloat?

Progression:

```
1 room -> basic observe / activate / act / update
2 rooms -> context split
3 rooms -> distinct episodes and goals
cross-room puzzles -> episodic transfer without flooding
```

## What This Solves

| Problem | Skyra's structural answer |
|---|---|
| Context window overload | Activation decides what reaches inference at all |
| Memory bolted on | Memory participates through Attention, Episodic, and Slice activation |
| Stale assumptions | Motivation dependency checks trigger re-grounding |
| Tool overload | Tools are reachable action leaves, not a flat prompt list |
| Orchestration bloat | Traversal, activation, and motor convergence carry sequencing |
| Poor auditability | Prior + Relation + Current + Delta explains decisions |
| RAG pretending to be memory | RAG becomes a learned retrieval primitive |
| Multi-agent manager overhead | Concurrent action patterns can emerge from reflex |

## Implementation Status

Current direction:

- `Reality` and `IRelation` are separate contracts.
- `ActivationRecord` is DNA-owned.
- `Relation.Realize` is an iterative DFS traversal kernel.
- `Slice` owns durable per-being activation records and traversal trees.
- `Node` records traversal/audit structure.
- Binding realities are being rebuilt around the new kernel.
- Membrane is the next major cell to update.

The architecture is not proven yet. The next proof is small: one room, one being, a few context cells, deterministic activation, and observable traversal records.
