# Slice, World, Organism — How They Relate

## The Analogy

- **World (grid)** — the full 1000x1000 grid of Points. Each Point is a Space with bidirectional pointers to its neighbors. The world is an executable like everything else — most of it is dormant (untraversed). Only the points the kernel visits are active. The world follows the same pattern as any organism: full state on disk, active subset in the slice.
- **Organism (Skyra DNA)** — the executable image on disk. Everything the being could be. All memories, all binding realities, all weights. Durable. Not everything is active at once. Exists independently — the kernel doesn't own it, the world doesn't own it.
- **Slice** — the running process. What's real about the being RIGHT NOW. The active subset. **Owned by the agent** — it is the agent's active mind. The kernel holds a *reference* to it (handed over at registration) so it can build active state into it. The slice never lives on the world or the kernel. Does NOT hold a reference back to the organism.
- **DNA** — the text segment. The compiled instructions. What the organism can do. Read-only. Doesn't change. Multiple Slices can share the same DNA.
- **Kernel (Relation)** — the OS. Manages traversal through the heap. Scores, filters, **builds** what gets paged in. Doesn't live on the grid — passes through it. Holds *references* to the slices (a registry, populated at registration). Durable — it persists across waves; it is not recreated per signal.

## Who Holds What

### The world holds the executables (the filesystem)

```
World.Executables: organism ID → organism
```

The full set of organisms that exist. The disk. What's available to be loaded. The world holds this — and the world is itself an executable. No special case. Placement puts an executable on a Point; the registry says it exists at all.

### The kernel references the slices (the process table)

```
Relation.ActiveState: being ID → *Slice   // references to agent-owned slices
```

A registry of *references*, populated at registration. **The agent owns its slice; the kernel only holds a reference** — so it can realize active state into it. One slice per active being, owned by that being. The agent hands the reference to the kernel at boot; the kernel never mints slices. The kernel is the durable runtime — it holds these references across waves, the same way the OS holds the process table (the process doesn't carry its own PCB, and the OS doesn't own the process).

When traversal crosses into an executable, the kernel looks up its slice reference in ActiveState, binds the frame — records, threshold, count — and **builds** active state into it as it traverses. The mapping goes one direction: the kernel references the slice; the slice doesn't point back, and the organism doesn't know it has one.

## The Traversal

```
Relation traverses grid
    → hits Point
        → finds the executable in Point.Content
            → Membrane.Observe (the front door)
                → kernel looks up the executable's Slice in ActiveState
                    → traversal runs in that slice's frame — records, threshold, count
                        → kernel descends into binding realities, memories, etc.
                            → Express on ascent
            → Membrane.Express (the exit)
    → continues to next Point
```

The membrane is the boundary between world space and the organism's interior. The Slice is the kernel's bookkeeping about what's active behind that boundary — it never sits on the world. The kernel doesn't know it walked into a being — it just sees more Spaces to traverse. ActiveState is how the kernel knows which frame to bind.

## Ownership

Nothing owns anything else. Everything exists independently. The kernel manages access.

- **World** — exists independently. The grid. Hardware. Holds the executables — and is itself an executable.
- **Organisms** — exist independently. On disk. The executables.
- **Slices** — **owned by their agent** (its active mind, RAM). The kernel holds a reference, handed over at registration. Never on the world. Don't know their origin.
- **Kernel (Relation)** — durable, not recreated per signal. Holds *references* to slices (a registry) and builds active state into them by traversing. Traverses the world. Binds the frame at each membrane crossing. Manages access between disk and RAM.

## Active State Is Built, Not Traversed

The kernel does not traverse active state — it **builds** it. The wave walks the **passive** Space (the organism on disk, the full topology reachable through `Children`/`Content`) and realizes the active subset into the slice as it goes. Active state is the *product* of a traversal, not an input to it.

This is why the slices live with their agents and the world holds none. A wave that "hit active state on the world" could only traverse what was already lit — and at boot nothing is lit. The wave has to walk the disk and build the active set. The world is simply more disk to traverse, not a store of pre-lit state.

## How a Being Acts

A being is situated: it holds a reference to **what it's standing on** — its environment, the layer below it — set at registration alongside its slice. To **think**, a being traverses *itself*: `realize(self, self, impulse)` over its own `Children`. To **act**, it calls into its environment:

```
world.observe(self, content)   →   realize(self, pointOf(self), content)
```

`observe` is the universal "receive a signal" — one verb everywhere, different competence per Space. The world's `observe` is a **facilitator**: it just relays to the kernel. A memory cell's `observe` absorbs; the origin's `observe` perceives the result. There is no new verb and no collision — `world.observe` (the world receiving the act) and `self.observe` (the being receiving the result) are the same uniform method on different objects. The being passed *itself* as the owner, so the kernel's callback (`self.observe(result)`) lands on it. The chain is **being → world → kernel → slices**. You act *down* (push on the layer below); you perceive *up* (the result returns to your origin — the [Perception Invariant](../laws/invariants/perception-invariant.md)).

The world is a **doorway, not a destination**: it relays the act to the kernel and is never itself the traversed realm. But the realm is not the being either — it is the being's **location**. So the relay threads `self` as the **owner** and `pointOf(self)` as the **root**: `realize(self, pointOf(self), content)` — neither `realize(self, self)` (that is *thought*, the being walking its own interior) nor `realize(self, world)` (the world-as-whole is not traversed). The act roots at the being's point and propagates across world-space from there; the result still returns to `self` (the [Perception Invariant](../laws/invariants/perception-invariant.md)). Owner and perceiver are the being; the root is the being's location. See the [Actor Frame Invariant](../laws/invariants/actor-frame-invariant.md).

### Why it can't recurse

`world.observe` would loop if relaying were unconditional — the kernel walks the world as space too, and walking it would re-call `observe`, which would re-relay. The break is that the world **splits its behavior on the presence of the origin label**:

```go
func (w *World) Observe(origin skyra.Space, impulse any) {
    if origin != nil {
        // a being is acting through me — facilitate, rooting the act at its location
        w.kernel.Realize(origin, w.pointOf(origin), impulse)
    } else {
        // I'm being walked as a realm — intake / propagate, never facilitate
        ...
    }
}
```

- A being acting calls `world.observe(self, content)` — origin **present** → relay to the kernel.
- The kernel walking the world as space calls `world.observe(nil, content)` — origin **absent** → the intake branch, no relay.

The self label is what arms facilitation; the traversal path strips it, disarming it. The recursion break lives entirely in the world's competence — no kernel special-case, no discriminator to invent. It is the same identity-presence test everywhere: a Space acts on a wave it owns; `nil` origin means "I'm being traversed, not invoked."

## Three Layers of State

- **Organism** — durable. Everything the being has accumulated. All memories, all weights, all binding realities. Persists across traversals. Disk.
- **Slice** — active. What's lit up right now. The active activation records, the current tree, what survived threshold this pass. Changes every traversal. RAM.
- **DNA** — the code. How the organism processes. Observe/Express functions, competence, cell types. Doesn't change. The text segment.

## The World Is an Executable

The world follows the same pattern as everything else. No special case.

```
World.Executables["world"]   → the world itself
Relation.ActiveState["world"] → the world's slice
```

The world's Point is itself — every other executable lives ON a point. The world IS the points.

Most of the grid is dormant. Points exist, neighbors are wired, but nothing is real about them until the kernel traverses them. The active subset — what the kernel has visited, what has content, what participates in wave propagation — that's the world's slice, held in ActiveState like everyone else's. The boundary between active and dormant shifts every traversal, the same way the observable universe's horizon shifts over time.

The world having a slice is what lets the kernel traverse space the same way it traverses a mind. A wave starts at a Point (from a being's Express), the world's frame scores the propagation, and only active space participates — dormant points don't. The world is an executable like everything else, not a background assumption.

## Space Is Local

There is no global map. An agent doesn't have a position relative to anything else until it observes. Activation strength IS distance — 1.0 is inside you, 0.0 is unreachable. The geometry emerges from traversal, not from coordinates.

The grid provides structure (adjacency, bidirectional pointers) but distance is subjective. Two agents on the same Point experience different proximities to the rest of the world because their weights are different.
