# World–Kernel Coupling

## The Split

**Passive state** — the topology. Grid, points, agents, binding realities, memory cells. Exists whether or not a signal is passing through. Data at rest.

**Active state** — the Relation mid-traversal. Built by the kernel as it calls Observe/Express on each cell it visits. Exists only during a wave. Data in flight.

The kernel traverses passive state and brings it into active state. The world collects signals and hands them to the kernel. They are coupled — the world lives on the Relation as a field. The Relation is the wave and the substrate. It already holds energy, slices, records. The world belongs there too. The kernel has the world, so when an agent's Express produces primitives on the ascent, the kernel can resolve them without the agent knowing anything about the world.

## Three Roles

**World** — the where. Owns the grid. Collects signals at its boundary and hands them to the runtime. Knows what primitives mean and how to execute them. Implements Reality.

**Kernel (Relation)** — the how. The Relation's `Realize()`. Iterative DFS through whatever topology it's pointed at. Scores children, filters by threshold, descends into survivors, calls Express on ascent. Holds the world as a field — the only thing that touches both the world and the agents. When an agent emits a primitive, the kernel knows who's expressing (it's the cell it just called Express on) and has the world to resolve execution.

**Agent** — the who. Membrane with binding realities inside. Pure topology — only wakes up when the kernel reaches it during traversal. Never holds a reference to the kernel or the world. Implements Reality.

## Not Physics: The Boundary

How signals get produced and how primitive execution reaches the outside are not physics. They live behind the world's boundary, outside this spec. The world exposes exactly two surfaces to the runtime: `Signals()` (payloads in) and primitive execution (effects out). What implements them is the world's private business — not a concern of the kernel, main, or this document.

## The Loop

```go
func main() {
    world := Boot(genome)
    for impulse := range world.Signals() {
        rel := NewRelation(world)
        rel.Realize(world, impulse)
    }
}
```

The Relation takes the world at construction. `Realize()` takes a root Reality and a payload. The impulse passes through the call chain as a parameter — the kernel never stores it.

The world is both:
- **An active process** — it collects signals at its boundary and writes them to the `Signals()` channel
- **A passive Reality** — traversable by the kernel like anything else

These aren't contradictory. The boundary is the ears. The Reality interface is the body the kernel moves through.

## Relation Struct

```go
type Relation struct {
    World     *World
    Energy    float64
    Records   map[string]ActivationRecord
    Count     int
}
```

The kernel is stateless. No impulse field, no request accumulator. The payload enters as a parameter to `Realize()` and passes through the call chain — `Observe(rel, payload)`, `Express(rel, payload)`. Express returns values rather than depositing them on the Relation. The kernel reads the return on the ascent and has the world to resolve execution.

```go
func (r *Relation) Realize(root Reality, payload any) any {
    // DFS traversal
    // payload passes through Observe/Express
    // content type check recurses: r.Realize(innerReality, payload)
    // Express return value bubbles back up
}
```

## PrimitiveRequest

```go
type PrimitiveRequest struct {
    Primitive string  // speak, look, listen, engage
    Payload   string  // the content
}
```

No routing. No `from` or `target`. The primitive defines what happens. `speak` means emit to the current thread. `look` means scan what's around me. The world resolves the rest from context — the kernel tells it which cell is expressing, the world knows where that cell is on the grid.

## World Struct

```go
type World struct {
    id          string
    Grid        [GridSize][GridSize]*Point
    Executables map[string]Reality  // organism ID → organism. the filesystem
    Primitives  map[string]func(cell Reality, payload string)
    signals     chan any
}
```

**Executables** — organism ID → organism. The filesystem. The world holds what exists; the kernel holds what's running (`ActiveState`: being ID → Slice). The world is itself an executable — no special case. Placement puts an executable on a Point; the registry says it exists at all.

**Primitives** — speak, look, listen, engage. The world's affordances. Agents return `PrimitiveRequest{Primitive, Payload}` from Express. The kernel reads the return value on the ascent and calls `world.Primitives[req.Primitive](cell, req.Payload)` — the world handles execution. How an effect reaches anything outside the world is behind the boundary.

**Signals()** — returns the channel. Main loop reads from it. Each signal becomes a Relation.

### World as Reality

```go
func (w *World) Observe(rel *Relation, payload any) {
    // read the payload (impulse)
    // determine which point(s) it targets
    // route inward — the kernel's DFS handles the rest
}

func (w *World) Express(rel *Relation, payload any) any {
    // world's Express handles primitive dispatch
    // reads return values from agent traversals
    // dispatches each through w.Primitives
    // returns aggregate result
}

func (w *World) Children() map[string]Reality {
    // returns occupied points — points with Content
    // the kernel only descends into points that have agents
    // empty grid points are not traversal targets
}
```

The world's `Children()` doesn't return all million points. It returns the occupied subset — the points where agents live. The grid topology (neighbors, distance, adjacency) is used internally by the world when it executes primitives, not by the kernel during traversal.

## Content Type Check Rule

Every cell can have a `Content` field. The kernel applies one rule, inside the traversal loop — not at the front, but after processing each cell:

```go
if r, ok := cell.Content.(Reality); ok {
    result := rel.Realize(r, payload)  // full traversal into the inner topology
}
```

Content is a Reality → the kernel calls Realize on itself with that Reality as a new root and the same payload. Full DFS — descent, scoring, tree building, ascent. The kernel recurses into itself. The outer traversal is the world. The inner traversal is the agent. Same function. The return value bubbles back up.

Content is anything else (string, state, nil) → the kernel ignores it. Not its business. The cell reads its own Content during Observe/Express. Skyra's Content might be a string — her durable state, her context. The kernel never touches it.

This is the boundary crossing. The grid point holds Skyra as Content. The kernel hits the point, processes it normally, checks Content, sees a Reality, and calls Realize with Skyra as root. Skyra's Children() returns her binding realities. The kernel traverses those, builds the inner tree, ascends back, and returns to the grid point's traversal. One type check bridges the world topology and the agent topology — without the kernel knowing what a grid is, what an agent is, or that a boundary was crossed.

## Point Changes

```go
type Point struct {
    id        string
    X, Y      int
    Neighbors []*Point
    Content   any
}
```

`Content` stays `any`. A point's Content might be a Reality (an agent) or might be something else (terrain, data, nothing). The kernel's content type check handles both cases. The point doesn't decide — the kernel checks.

Points are the world's internal data structure for spatial relationships. The kernel sees them through the world's `Children()`. The world uses points to execute primitives (distance, adjacency, propagation).

## Primitive Flow

```
1. Skyra's Express returns PrimitiveRequest{speak, "hello"}
2. Kernel reads the return value on ascent — it knows the calling cell from the stack
3. Kernel calls world.Primitives["speak"](callingCell, "hello")
4. World looks up the calling cell's position on the grid
5. World resolves what "speak" means in context (emit to current thread)
6. World executes it at its boundary
```

No routing in the request. The primitive *is* the routing. The world resolves the rest from the calling cell's position and the current traversal context.

## Wave Shape — Three-Way Ownership

Every wave is born from a primitive. There are no shapeless waves — an inbound signal addressed to a being is an engage *toward* it; a being's motor emitting look is a look *from* its point. Because every wave has a launching primitive, every wave has one author for its shape.

Three parties, each owning exactly what it alone knows:

- **Primitive owns shape** — spread (how many forward edges the wave considers), how signal divides across them, attenuation per step, the initial heading. The *what*. Topology-agnostic intent, set once at emission. Rides on the Relation like Energy does — transient wave physics, gone when the wave ends.
- **Cell owns adjacency resolution** — given a heading off the Relation, which of *my* edges are forward. The *where*. Only the cell knows its own neighbors, so only the cell can turn a cone into edges. The kernel never learns geometry; the cell resolves it.
- **Kernel owns transport and composition** — carry the shape, hand it to each cell, traverse whatever the cell returns as forward survivors. The *how-it-moves*. Never interprets shape, never knows geometry.

The config that reaches the kernel is intent, not resolved edges. If it were resolved, grid geometry would be back in the kernel and the affordances split (see [world-dna.md](../world/world-dna.md)) would be undone. The kernel relays the descriptor; the cell plus the descriptor produce the survivors.

### Shape Is Data, Not Kernel Code

Because shape is a descriptor the kernel carries rather than logic the kernel runs, a new wave shape is a new primitive with new config values — broadcast is spread = all edges, a whisper is spread = 1 with fast attenuation — and the kernel never changes. Same extensibility as a new Reality dropping into the graph without touching the Relation. The kernel's wave mechanics are frozen; the vocabulary of waves grows above it.

### A Primitive Has Two Halves

**Shape** steers traversal and is consumed by cells on the descent. **Effect** is what the world does at the boundary when the wave lands — cross into the target, emit to the thread — and fires on the ascent through the world's primitive function. Same primitive, two faces, used at opposite ends of one traversal.

### Direction Is the Third Relative Quantity

Distance is activation strength relative to the observer. Time is proper time per being. Direction is heading relative to emission — there is no global north. "Toward michael" resolves to a bearing relative to the *emitting* being's point; another being engaging the same michael launches a different absolute direction for the same relative intent. Three relativities, one substrate.

## Signal Flow

```
1. A signal arrives at the world boundary
2. The world writes it to world.signals
3. Main loop reads from world.Signals()
4. Main loop creates Relation with the world
5. rel.Realize(world, impulse) — kernel starts from the world with the impulse as payload
6. World.Observe receives the payload, routes to the target agent
7. Kernel descends into agent's topology, payload passes through
8. Agent binding realities fire (episodic, goal, emotion...)
9. Agent Express returns PrimitiveRequest
10. Kernel ascends — reads return value, dispatches through world.Primitives
11. Loop waits for next signal
```

## What the Kernel Does NOT Know

- What primitives mean — it dispatches them through the world, doesn't interpret them
- Where signals come from or where primitive effects land
- The payload's shape — it passes it, never reads it
- The wave's shape — it carries the shape descriptor on the Relation and hands it to cells, never resolves it into edges
- The grid's spatial layout

## What Agents Do NOT Know

- That they're on a grid
- Where other agents are
- How primitives execute
- That the kernel exists
- How signals enter the system

## What the World Knows

- The grid and who's where
- What primitives mean and how to execute them
- How its own cells resolve a heading into forward edges — adjacency is the cell's, the geometry is the world's
- What sits at its own boundary — the runtime and kernel never do

## What the Kernel Knows

- The world (it's a field on the Relation)
- Which cell it just called Express on (it's on the stack)
- The wave's shape descriptor, carried on the Relation — relayed to cells, never interpreted
- How to dispatch: reads Express return value, calls `world.Primitives[req.Primitive](cell, req.Payload)`

## The Payload

The payload has no struct in this spec. It is opaque above the membrane: main plumbs it, the kernel passes it, intermediate cells ignore it. Its shape is endpoint knowledge — minted at the world boundary, read by the cells it concerns. Passed by reference, it is also the cell-to-cell medium: anything a cell wants to leave for another cell goes into the payload, not into kernel state.

## Open

- **Engage navigation:** for a target not on one of the 8 straight lines from the being, does the ray conserve pure momentum (fixed forward neighbor, travels straight, may approximate) or re-lock each step (greedy seek toward the target)? `look` is unambiguously momentum — no target to seek. `engage` is the fork: launched-and-conserved, or homing.
- **Broadcast signals:** an impulse with no target — does the world route it to all agents? To nearby agents on the grid? Under the shape model this is just a primitive with spread = all edges; the question is whether it earns a name.
- **Multiple worlds:** the personal network means multiple worlds. Each world has its own grid and its own primitives. Worlds connect at their boundaries. Future — but the World struct should not assume singleton.
