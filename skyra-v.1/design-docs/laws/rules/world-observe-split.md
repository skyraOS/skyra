# The World Observe Split

## Status

**Rule.** This is an operational rule for the World Space. It protects the kernel from recursive relay.

## The Rule

> **`world.Observe(self, impulse)` facilitates action. `world.Observe(nil, impulse)` participates in traversal.**

Short form:

> Non-nil self invokes the world. Nil self means the world is being traversed.

## What It Means

The world is both an environment and a Space. Those roles must not collapse.

When `self != nil`, a being is acting through the layer below it. The world is being invoked as an environment doorway, so it may call the kernel — threading `self` as the **owner** but the being's **location** as the traversed root:

```go
kernel.Realize(self, pointOf(self), impulse)
```

`origin` (owner) and `traversed` (root) are independent axes: the act is owned by the being but roots at its point, so it propagates across world-space while the result still returns to the being. See the [Actor Frame Invariant](../invariants/actor-frame-invariant.md).

When `self == nil`, the kernel is traversing the world as world-space. The world must not call `kernel.Realize`; it only performs intake / propagation behavior for the current traversal.

The kernel produces this branch during traversal:

```go
traversed.Observe(nil, impulse)
child.Observe(nil, impulse)
```

Return perception is different:

```go
Realize(origin, origin, resultImpulse)
```

## Why It Exists

This split is the recursion break.

If `world.Observe(nil, impulse)` relays to the kernel, then kernel traversal of the world re-enters `world.Observe`, which starts another kernel traversal, recursively.

The presence of `self` means "facilitate an actor's act."

The absence of `self` means "I am being traversed."

## Part of a Universal Rule

This split is not special to the world. The kernel calls `Observe(nil, impulse)` on **every** Space it traverses — interior cells included — and reserves a named origin for invocation and return perception. The world's two cases are one instance of the runtime-wide law:

> **nil origin = being traversed; named origin = being invoked. Return perception is a new wave rooted at the origin.**

Facilitation (`world.Observe(self, …)`) is the *invoked* case; an interior cell walked with `nil` is the *traversed* case; a being's reafference is the return perception wave rooted at that being. Same rule everywhere — see the [Perception Invariant](../invariants/perception-invariant.md).

## Latent Edge (not yet live)

"Named origin → facilitate" assumes the world is not currently using `World.Observe(world, …)` as a return-perception entry. Return perception should root a new wave at the origin, not call `Observe` as a callback. Safe while the world owns no waves; revisit before it does.
