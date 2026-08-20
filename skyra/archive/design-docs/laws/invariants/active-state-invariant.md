# The Active State Invariant

## Status

**Foundational.** This invariant defines what the kernel traverses and where active state lives.

## The Invariant

> **The kernel traverses inactive Space and realizes active state into that Space's Slice.**

In short:

> The kernel traverses disk space and brings it into active space.

## What It Means

- **Space is durable structure.** A Space lives on disk: DNA, organism, world, memory, binding topology, and anything else that can be traversed.
- **Slice is active record.** Slice is active space: activation records, traversal count, current/prior tree, thresholds, and whatever else is real about that Space right now.
- **The kernel never traverses the Slice.** The Slice is not the graph. The kernel traverses Space and writes the active subset it discovers into the Slice.
- **Every Space owns its Slice.** The active record belongs to the Space whose structure is being traversed. An empty Slice means nothing is active yet; it is not a different kind of object.

## The Rule

For any wave:

```text
kernel traverses TraversedSpace
kernel uses TraversedSpace.Slice as the active record
kernel writes what became active into that Slice
```

The Slice is not passed around as the thing being traversed. It is the record of what the traversal made active.

## Boundary Crossing

When traversal encounters `Content()` that is itself a Space, the kernel has found another actor boundary.

That does not make the parent wave continue through the child actor's active state. It starts an independent wave over the contained Space:

```text
parent wave traverses parent Space
Content Space found
spawn child wave over child Space
child wave writes into child Space's Slice
parent wave continues over parent Space
```

The parent wave does not wait for the child wave and does not receive the child wave's result. Each wave returns through its own origin by starting a perception wave there, as stated by the [Perception Invariant](perception-invariant.md).

## Why It Matters

This keeps the runtime from confusing the map with the activation record.

The world, an organism, a memory graph, or a being's interior can all be mostly dormant. The kernel walks the dormant structure and pages the relevant parts into active space. Active state is built by traversal; it is not the substrate being traversed.

This is the rule that lets a finite wave move through a much larger world without requiring the whole world to be active first.
