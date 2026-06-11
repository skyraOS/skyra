# Activation & Traversal Mechanics

## The Problem

The cell sees what's in front of it. The Slice keeps score. The Relation carries energy. Who does what, when, and where does state live?

## The Architecture

Two contracts. The wave and the inhabitant.

**IRelation** — the wave + shared substrate. Drives traversal, holds per-being state. This is the kernel. It does not know how scoring works.

**Reality** — the inhabitant. Cells implement this. The kernel traverses these. Each cell type defines its own activation struct and scoring equation.

They meet at the interface but they don't own each other. The Relation is mechanism. The Reality is identity.

## The Loop

1. **Cell.Observe(Relation)** — cell receives the Relation, does internal work
2. **Kernel reads Children()** — gets the cell's visible neighbors by reference
3. **Kernel calls Activate(true)** — on each present child's activation record
4. **Kernel calls Activate(false)** — on absent children (previously seen, not in current report)
5. **Kernel filters by threshold** — above threshold = survivor
6. **Kernel checks energy** — enough to descend?
7. **Survivors get pushed onto the stack** — iterative DFS continues

The cell is the sensor. The activation record is the scorer. The kernel is the driver.

**Note:** Attention is a separate concept from activation records. Attention is the being's aperture — heading and width — one field the membrane reads from both directions: sensory gates incoming signal against it, motor fires through it (see [world-kernel-coupling.md](world-kernel-coupling.md#attention-is-one-field-read-twice)). Activation records are the scoring system. They are not the same thing.

## The Interfaces

Three interfaces. Two contracts and one extension point.

### IRelation — the wave

```go
type IRelation interface {
    ID() string
    Realize(beingID string, brID string, root Reality)
    GetSlice(beingID string) *Slice
}
```

- **Realize** — drives the full iterative DFS. The one entry point.
- **GetSlice** — the being reads its own cognition. Attention, Prior trees, Records.

The Relation is both the wave that moves and the substrate that remembers. It carries energy for one traversal but holds Slices that persist across all traversals. The relationship between the world and the being.

### Reality — the inhabitant

```go
type Reality interface {
    ID() string
    NewActivation() ActivationRecord
    Observe(r *Relation)
    Express(r *Relation)
    Children() map[string]Reality
}
```

- **NewActivation** — factory. Returns whatever scoring struct this cell type's DNA defines. The kernel stores it, calls it, never downcasts.
- **Observe** — called on descent. The cell does its internal work.
- **Express** — called on ascent. The cell does its output work.
- **Children** — exposes visible neighbors to the kernel.

### ActivationRecord — the scoring contract

```go
type ActivationRecord interface {
    Activate(visited bool, traversalCount int) float64
}
```

One method. The kernel calls it. The cell defines what it means. DNA injects the factory via `NewActivation()` on the Reality interface — the genome wires in the activation at boot, the cell carries it, the kernel scores it.

Each cell type implements `NewActivation()` — a factory that produces whatever scoring struct that cell type needs. The kernel stores it, calls it, reads the score. Never downcasts.

## BaseActivation

A default implementation for embedding:

```go
type BaseActivation struct {
    Identity       float64  // EMA α=0.02 — the groove
    Familiarity    float64  // EMA α=0.1 — recent relationship
    LocalTraversal int      // per-cell visit count
    LastTraversed  int      // global count when last bumped
}

func (b *BaseActivation) Activate(visited bool, traversalCount int) float64 {
    // bump or decay EMAs, compute recency, return product
}
```

Binding realities embed this and extend:

```go
type EpisodicActivation struct {
    BaseActivation
    Overlap float64
}

func (e *EpisodicActivation) Activate(visited bool, traversalCount int) float64 {
    base := e.BaseActivation.Activate(visited, traversalCount)
    return base * e.Overlap
}
```

Each binding reality defines what activation means for its own cells. No shared BindingFields map. No external writes. The cell owns its scoring.

## Bump and Decay

The kernel calls `Activate(true)` for present children, `Activate(false)` for absent ones. What happens inside is the activation record's business.

The BaseActivation implementation:
- **Visited = true** → bump Identity EMA toward 1.0, bump Familiarity EMA toward 1.0, increment LocalTraversal, update LastTraversed
- **Visited = false** → decay Identity EMA toward 0.0, decay Familiarity EMA toward 0.0

Other implementations can do whatever they want. The kernel doesn't care.

## Threshold

The Slice holds a threshold value (DNA-defined). After the kernel calls Activate on all children, only those returning a score above threshold become survivors for descent.

## Dependencies

What depends on what. If something accumulates across traversals, the Slice owns it. If something exists only for the current wave, the Relation owns it.

### What the Slice owns:
- **TraversalCount** — monotonic global clock, increments once per BeginTraversal
- **Records** — `map[string]ActivationRecord` — persistent per-cell scoring state
- **LastChildren** — `map[string][]string` — what children each cell reported last time (for decay)
- **Attention** — what's around the being (placeholder)
- **Threshold** — DNA-defined minimum score to descend
- **Stack** — DFS state machine (ephemeral per traversal)
- **Trees** — `map[string]*TraversalPair` — current/prior trees per binding reality

### What the Relation owns:
- **Energy** — the wave's remaining budget. Each descent step costs 1.
- **Slices** — `map[string]*Slice` — one per being

### What the ActivationRecord owns:
- Whatever internal state it needs for scoring (EMAs, counters, custom fields)
- Updated only through the `Activate` method
- Persists across traversals via the Slice's Records map

### Traversal itself
- **Depends on:** score > threshold AND energy > 0
- **Therefore:** both Slice (score) and Relation (energy) must pass for descent

### Order of operations per cell:
1. Kernel calls cell.Observe(Relation)
2. Kernel reads cell.Children()
3. Kernel looks up or creates ActivationRecord for each child
4. Kernel calls Activate(true) on present, Activate(false) on absent
5. Kernel filters by threshold
6. Kernel checks energy
7. Survivors pushed to stack

## The Node

The kernel's internal traversal frame:

```go
type Node struct {
    CellID    string
    Cell      Reality
    Record    ActivationRecord
    Score     float64
    Children  []*Node    // all children — above and below threshold
    Survivors []Reality  // above threshold — queued for descent
    NextChild int        // index into Survivors for iterative DFS
}
```

This is ephemeral traversal state. The Node exists for one traversal. The Record inside it persists across traversals (stored in Slice.Records).

## Iterative DFS

No recursion. No suspended stack frames. The kernel drives the loop explicitly.

```
push root
loop:
    top = stack peek
    if top has unvisited survivors AND energy > 0:
        child = next survivor
        child.Observe(r)
        score children, push child node
        energy--
    else:
        pop top
        top.Cell.Express(r)
        append top to parent's Children (or to tree root if stack empty)
```

The Relation sees all the cells' methods. The Relation calls those methods in a specific order that traverses the interior. The cell doesn't drive traversal — the kernel does. If the cell's implementation doesn't traverse correctly, that's the cell's problem. Like Linux — no concessions.

## Tree Reconstruction

The tree builds on ascent. When a node pops, its parent is always the current top of stack (guaranteed by sequential DFS — no interleaving).

```
[A]         → A sees B, C. Pushes B first
[A, B]      → B sees D, E. Pushes D
[A, B, D]   → D has no survivors. Express pops D, appends to B
[A, B]      → B pushes E
[A, B, E]   → E pops, appends to B
[A, B]      → B done. Express pops B (with children D,E), appends to A
[A, C]      → C sees F, G, H. Pushes F
[A, C, F]   → F pops, appends to C
[A, C, G]   → G pops, appends to C
[A, C, H]   → H pops, appends to C
[A, C]      → C done. Express pops C (with children F,G,H), appends to A
[A]         → A has full tree
```

### Double Buffer (TraversalPair)

Each binding reality gets its own current/prior tree pair:

```go
type TraversalPair struct {
    Current []*Node
    Prior   []*Node
}

Trees map[string]*TraversalPair  // keyed by binding reality ID
```

At the start of every traversal: `Prior = Current`, `Current = nil`, stack cleared.

Nothing reads a half-built tree. Binding realities read Prior (last complete traversal). Current builds without interference.

### Complexity

O(n) push + O(n) pop = O(n). Theoretical minimum for tree construction.

### Parallel Future

If traversal ever goes parallel, the stack model breaks — interleaving destroys parent lookup. Would need separate stacks per subtree. Not now.

## Two Recursions

This is the single hardest thing to hold about the kernel, and once you see it the whole system reads cleanly: the kernel runs **two different recursions, using two different mechanisms.** Keep them separate and it makes sense; conflate them and nothing does.

**1. In-plane traversal — the explicit stack, synchronous.** A wave moving across *one* actor's topology (its own binding realities, memories — reached through `Children()`) is the iterative DFS above, driven by a manual stack (`[]*Node`). No recursion here, no goroutines — one wave, one stack. The kernel owns every frame so it can apply policy at each step (visit-once, energy, threshold, scoring) and build the actor's result on the ascent. This is the wave spreading through a graph, and it is sequential.

**2. Cross-actor crossing — goroutine spawn, asynchronous.** When the wave hits a cell whose `Content` is itself a Reality, that is not a sub-structure of the current computation — it is **another actor.** The tell is in the code: the crossing looks up *that reality's own slice* in `ActiveState`, and a separate slice means a separate frame, a separate proper-time clock, a separate being. Actors are concurrent, so the crossing **spawns a goroutine** (`go realize(inner, inner, impulse)`) — it does not call down the Go stack. The inner actor runs its own in-plane DFS on its own goroutine and delivers its result to *its own* origin. The outer wave does not wait.

So the shape is: **an explicit stack-machine inside each actor, and a goroutine per actor.** The goroutine tree *is* the actor tree. Within an actor, one stack walks the topology sequentially; cross an actor boundary and a new goroutine opens with its own stack — concurrently.

### Why the split

Each recursion is matched to the mechanism that fits it:

- **In-plane is large, policy-heavy, and one being's own thought** → explicit stack, sequential. A plane can hold a huge topology, so you want a heap-allocated stack; and the kernel must own the frame state (survivors, NextChild, score) to gate and score each step.
- **Cross-actor is a boundary between concurrent beings** → goroutine spawn, async. Crossing into another actor synchronously would mean one being blocking inside another's interior and ripping out its response — wrong. The other being is its own goroutine; it processes on its own clock and answers when it answers.

### Delivery to origin is the relay

There is **no return bubbling up a Go stack across actors.** Each goroutine delivers its result to *its own* origin — that is the [Perception Invariant](../invariants/perception-invariant.md), and it is the only relay. The explicit stack delivers in-plane (it builds the actor's own result on the ascent); the goroutine delivers cross-actor (to the inner being). A response from one being to another is not a return value — it is a **separate wave back**, another goroutine. Async message passing between concurrent actors.

That this is async is not a compromise, it is more correct: it respects being boundaries (no blocking inside another being), it gives the concurrency temporal binding wants (many staggered waves over the shared substrate), and it matches the physics — active sensing has propagation delay, so a perception returning a cycle later *is* the reafference loop, not a bug.

### The two-word vocabulary

- **traverse** a plane → the explicit stack (one actor, one topology, sequential).
- **cross** an actor boundary → a goroutine spawn (async; the inner actor delivers to its own origin).

Sync exists only *inside* an actor; crossing an actor boundary is always a spawn. Whether something is sync or async is therefore a **placement decision** — make it a `Child` (in-plane, part of my thought, sequential) or put it behind `Content` as its own actor (cross, concurrent). The genome decides by where it puts the reality. (The "Parallel Future" caveat above still holds: a *single* plane's DFS stays sequential — the concurrency is across actors, never within one stack.)

## What the Kernel Does NOT Know

- What scoring means
- How many EMAs exist or what their alphas are
- What recency is
- What binding fields are
- How decay works internally
- What emotion, episodic overlap, or motivation relevance are
- What a wave's shape is — it carries the aperture on the Relation and traverses the survivors the cell returns; it never interprets shape or resolves direction. Shape is attention's, not the kernel's (see [world-kernel-coupling.md](world-kernel-coupling.md#wave-shape--ownership))

The kernel is mechanism. The cell is policy. DNA is the distro.
