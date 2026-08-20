# The Plane Slice Invariant

## Status

**Foundational.** This invariant defines what a Slice belongs to and how `Children()` differs from `Content()`.

## The Invariant

> **`Children()` stay on the same slice plane. `Content()` may cross to another slice plane.**

In full:

> Every `Space` belongs to a slice plane. `Space.Slice()` returns the active-state record for the plane that Space belongs to, not a unique per-node Slice. Traversal through `Children()` remains in-plane and writes into the same Slice. Encountering a `Space` in `Content()` is a possible plane boundary; if crossed, the kernel starts an independent wave over that contained Space and writes into that Space's own plane Slice.

## What It Means

- **Slice is shared plane state.** A world point returns the world Slice. A memory cell inside a being returns that being's Slice. A binding reality inside a being returns that being's Slice.
- **Children are in-plane.** `Children()` are part of the same traversed plane. They share the same Slice as the parent Space.
- **Content may be out-of-plane.** `Content()` can hold ordinary data, nil, or another `Space`. If it holds another `Space`, that Space may belong to another plane with another Slice.
- **Actor boundary is slice boundary.** Crossing into `Content()` Space starts a new wave because the kernel has found a different plane of active state.
- **The kernel does not own slices.** It asks the traversed Space for `Slice()` and writes active state there.
- **No per-point slice explosion.** A million world points can all return the same world Slice. They are separate Spaces in one plane, not a million actors.

## Runtime Rule

```text
Children traversal:
  parent.Slice() == child.Slice()
  same wave
  same stack
  same active record

Content crossing:
  parent.Slice() may differ from contentSpace.Slice()
  new wave
  new goroutine
  new active record
```

## Short Form

> Children share the plane. Content can open another plane.
