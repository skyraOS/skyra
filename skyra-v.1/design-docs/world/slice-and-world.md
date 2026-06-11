# Slice, World, Organism — How They Relate

## The Analogy

- **World (grid)** — the full 1000x1000 grid of Points. Each Point is a Reality with bidirectional pointers to its neighbors. The world is an executable like everything else — most of it is dormant (untraversed). Only the points the kernel visits are active. The world follows the same pattern as any organism: full state on disk, active subset in the slice.
- **Organism (Skyra DNA)** — the executable image on disk. Everything the being could be. All memories, all binding realities, all weights. Durable. Not everything is active at once. Exists independently — the kernel doesn't own it, the world doesn't own it.
- **Slice** — the running process. What's real about the being RIGHT NOW. The active subset. Lives as a reference in the kernel's ActiveState map — never on the world. Does NOT hold a reference back to the organism. Doesn't know where it came from. Just active state.
- **DNA** — the text segment. The compiled instructions. What the organism can do. Read-only. Doesn't change. Multiple Slices can share the same DNA.
- **Kernel (Relation)** — the OS. Manages traversal through the heap. Scores, filters, decides what gets paged in. Doesn't live on the grid — passes through it. Holds the active state: a map of slices, keyed by being.

## Who Holds What

### The world holds the executables (the filesystem)

```
World.Executables: organism ID → organism
```

The full set of organisms that exist. The disk. What's available to be loaded. The world holds this — and the world is itself an executable. No special case. Placement puts an executable on a Point; the registry says it exists at all.

### The kernel holds active state (the process table)

```
Relation.ActiveState: being ID → Slice
```

The kernel's bookkeeping about what's running. One slice per active being. The Slice doesn't know which organism it came from. The organism doesn't know it has a slice. The kernel holds the mapping, the same way the OS holds the process table — the process doesn't carry its own PCB.

When traversal crosses into an executable, the kernel looks up its slice in ActiveState and binds the frame — records, threshold, count. The executable doesn't do the lookup. The kernel does. The mapping goes one direction from the kernel's perspective. The loaded state doesn't point back to its source.

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

The membrane is the boundary between world space and the organism's interior. The Slice is the kernel's bookkeeping about what's active behind that boundary — it never sits on the world. The kernel doesn't know it walked into a being — it just sees more Realities to traverse. ActiveState is how the kernel knows which frame to bind.

## Ownership

Nothing owns anything else. Everything exists independently. The kernel manages access.

- **World** — exists independently. The grid. Hardware. Holds the executables — and is itself an executable.
- **Organisms** — exist independently. On disk. The executables.
- **Slices** — live as references in the kernel's ActiveState map. Active state in RAM. Never on the world. Don't know their origin.
- **Kernel (Relation)** — holds ActiveState (being ID → Slice). Traverses the world. Binds the frame at each membrane crossing. Manages access between disk and RAM.

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
