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
    world := Boot(genome)   // builds the durable kernel; registration hands each being's
                            // slice reference to the kernel and its environment reference to the being
    for impulse := range world.Signals() {
        world.Receive(impulse)  // routes to the target and runs a wave on the ONE persistent kernel
    }
}
```

The kernel is **durable** — built once at boot, it holds the slice references (and the active state it builds into them) across waves. It is **not** recreated per signal; the active state a wave builds would evaporate every time. `Realize(origin, traversed, impulse)` runs a wave; the impulse and origin pass through the call chain as parameters — the kernel doesn't store them. A being acts outward with `world.Act(self, content)`, which runs a wave on that same persistent kernel.

The world is both:
- **An active process** — it collects signals at its boundary and writes them to the `Signals()` channel
- **A passive Reality** — traversable by the kernel like anything else

These aren't contradictory. The boundary is the ears. The Reality interface is the body the kernel moves through.

## Relation Struct

```go
type Relation struct {
    ActiveState map[string]Slice   // references to agent-owned slices (the registry)
    Amplitude   float64            // initial wave amplitude, reset per Realize
    wg          sync.WaitGroup     // tracks the current wave tree (it + its cross-actor sub-waves)
}
```

The kernel holds **references** to agent-owned slices, populated at registration — it builds active state into them but does not own them. The per-wave state (amplitude, the wait group) is fresh per `Realize`; the impulse and origin pass as parameters, not stored. `Observe(origin, impulse)` / `Express(origin, impulse)` thread the wave's owner so a cell can tell a self-wave from an external one. Express returns its output (never a dispatch trigger).

```go
func (r *Relation) Realize(origin, traversed Reality, impulse any) any {
    // DFS traversal
    // payload passes through Observe/Express
    // content type check crosses an actor boundary → spawns a goroutine
    // each actor delivers its result to its own origin (no return-bubble across actors)
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

Content is a Reality → the kernel crosses an actor boundary. Because that reality has its own slice (its own frame and proper-time clock), the crossing **spawns a goroutine** running the same function on the inner topology — not a synchronous call. The **origin is threaded unchanged**, so the inner actor reads the wave as external and its result returns to the wave's original starter (the [Perception Invariant](../invariants/perception-invariant.md)); nothing bubbles back up the call stack. The outer traversal is the world; the inner is the agent; same function, concurrent. A response from the inner actor is a separate wave back, not a return value. (See [Two Recursions](activation-and-traversal.md#two-recursions): traverse in-plane on the explicit stack, cross an actor boundary by spawning.)

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

## Wave Shape — Ownership

Every wave has a shape — an aperture (a heading and a width), and how its signal divides and attenuates as it travels. The question is who owns each part. Four parties, each owning exactly what it alone knows.

- **Attention owns shape** — the aperture: heading and width. This is the being's orientation, its stance. It is not computed per-act; it is attentional state, set by what the being last looked at and holds focus on. Width is the thing the old model mistook for a primitive property: focused attention is narrow, diffuse attention is wide. Width sets how signal divides across forward edges, which sets reach — narrow penetrates, wide surveys. Attention is durable (the stance persists across traversals); on emission it imprints the aperture onto the Relation as transient wave state, the same prior-and-learner pattern the Thread Reality uses (see [activation-and-traversal.md](activation-and-traversal.md)). Cells read it off the Relation; on ascent attention updates from what the traversal proved. One field, read from both directions — see [*Attention Is One Field, Read Twice*](#attention-is-one-field-read-twice) below.
- **Primitive owns verb + effect** — the verb (perceive, commit, sustain, emit, receive) and the boundary effect (cross into the target, emit to the thread). Pure execution. No geometry. `look` is *perceive* through the current aperture; `engage` is *commit* through it. The narrow-vs-wide was never the primitive's — it was the aperture's.
- **Cell owns adjacency resolution** — given the aperture off the Relation, which of *my* edges are forward / in-cone. Only the cell knows its own neighbors, so only the cell can turn a cone into edges. The kernel never learns geometry; the cell resolves it.
- **Kernel owns transport** — carry the Relation (which carries the aperture), hand it to each cell, traverse whatever the cell returns as survivors. Never interprets shape, never knows geometry.

Attenuation belongs to none of them — it is wave physics, the signal weakening as it propagates (the Energy on the Relation). Independent of verb and aperture both.

The aperture that reaches the kernel is intent, not resolved edges. If it were resolved, grid geometry would be back in the kernel and the affordances split (see [world-dna.md](../world/world-dna.md)) would be undone. The kernel relays the aperture; the cell plus the aperture produce the survivors.

### Attention Is One Field, Read Twice

The aperture is not two things — an inbound filter and a separate outbound heading. It is one field, heading and width, that the membrane's two faces read in opposite directions:

- **Sensory (Observe, descent)** reads it as a *containment test*. A signal arrives by propagation; attention asks whether it falls within the cone — aligned with the heading, inside the width. Within → it penetrates the membrane and reaches the binding realities. Outside → filtered. This is the resonance filter.
- **Motor (Express, ascent)** reads it as the *emission aperture*. The act fires through the same cone — same heading, same width. The wave launches along the field.

Inbound it gates; outbound it aims. The sensory/motor asymmetry is the *direction of the read*, not two mechanisms. (The membrane itself stays pure routing — the gating lives in the Attention cell the membrane routes into, not in the membrane.)

Two consequences fall out:

- **The efference copy is free.** Because motor and sensory share the field, the being predicts its own action's sensory return: the predicted reafference is "what falls within my own aperture," so it subtracts cleanly (von Holst). This is the mechanism [episodic-encoding.md](../cognition/episodic-encoding.md) assumes when it splits Observe into efference (self-signal) and exafference (other-signal) modes — the shared field is the efference copy.
- **You can only act where you attend.** One field means tracking one thing while acting on another is impossible in a single pass — the being must reorient, which serializes. That is the attention bottleneck, and it is correct: no precise action toward what you are not attending to.

### Shape Is Data, Not Kernel Code

The kernel carries the aperture; it never runs it. So the vocabulary grows in two independent directions without touching the kernel: a new **verb** is a new primitive (perceive / commit / … — pure execution), and a new **aperture** is an attentional state (a focused sweep, a wide scan). Both are data the kernel relays. The kernel's wave mechanics are frozen; verbs and apertures grow above it. Same extensibility as a new Reality dropping into the graph without touching the Relation.

### The Primitive Collapsed to Execution

A primitive used to read as two halves — shape and effect. Shape migrated to attention. What remains is verb + effect: what the being does, and what happens when the wave lands. "Primitives just execute" is the precise statement of that collapse. The aperture they execute through is attention's, read off the Relation; the primitive contributes none of it.

### Direction Is the Third Relative Quantity

Distance is activation strength relative to the observer. Time is proper time per being. Direction is heading relative to the being's own attention — there is no global north, and no per-act bearing computation. "Toward michael" is just where the being is already oriented, because it looked at michael and held focus. Three relativities, one substrate; all three live in the being's frame, not the world's.

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

**Supersedes steps 9–11.** A being's `Express` produces *output*, not a dispatch signal. Outward action is an explicit `world.Act(self, content)` call the motor makes when it decides to act, and the wave's result is delivered back to the origin (`origin.observe`) on completion. Every wave delivers its result to its origin — internal or external, self-started or not. This is the **[Perception Invariant](../invariants/perception-invariant.md)**, foundational and load-bearing for all of v.1. The signature, delivery, and goroutine spawn are implemented (A1/A2); the `world.Act` entry point itself is A3 (build the World), so the return-based steps 9–11 above are obsolete.

## What the Kernel Does NOT Know

- What primitives mean — it dispatches them through the world, doesn't interpret them
- Where signals come from or where primitive effects land
- The payload's shape — it passes it, never reads it
- The wave's shape — it carries the aperture on the Relation and hands it to cells, never resolves it into edges
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
- The wave's shape (the aperture), carried on the Relation — relayed to cells, never interpreted
- How to dispatch: reads Express return value, calls `world.Primitives[req.Primitive](cell, req.Payload)`

## The Payload

The payload has no struct in this spec. It is opaque above the membrane: main plumbs it, the kernel passes it, intermediate cells ignore it. Its shape is endpoint knowledge — minted at the world boundary, read by the cells it concerns. Passed by reference, it is also the cell-to-cell medium: anything a cell wants to leave for another cell goes into the payload, not into kernel state.

## Open

- **Engage navigation — resolved:** the momentum-vs-homing fork dissolved once direction moved to attention. Engage does not navigate toward a target; it commits along the aperture attention already established (look aims, focus holds, engage commits). The heading is set upstream, in attention, so there is nothing to re-lock — momentum by construction. Remaining sub-question: the **capture basin** — a being on the grid radiates a weak presence field so a near-miss resonates partially rather than all-or-nothing, turning hit/miss into a gradient (clean / glancing / missed) that reads out how well the being was oriented.
- **Broadcast signals:** an impulse with no target — does the world route it to all agents? To nearby agents on the grid? Under the shape model this is just a primitive with spread = all edges; the question is whether it earns a name.
- **Multiple worlds:** the personal network means multiple worlds. Each world has its own grid and its own primitives. Worlds connect at their boundaries. Future — but the World struct should not assume singleton.
