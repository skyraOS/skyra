# A3 World Observe / Slice Result Stub

## Status

**Superseded — done.** This stub named the A3 work before it was designed. That work is now built and specified; the file is kept only as a breadcrumb to where each piece landed.

## What Was A3, and Where It Landed

### `world.observe` / environment / minimal world-being action path — DONE

A being acts by calling into the world it stands on; the world facilitates a kernel wave rooted at the being's location.

- entry + recursion break: [World Observe Split](../laws/rules/world-observe-split.md) — `world.Observe(self, …)` facilitates, `world.Observe(nil, …)` participates.
- owner vs. root: [Actor Frame Invariant](../laws/invariants/actor-frame-invariant.md) — `Realize(self, pointOf(self), content)`; owner stays the actor, root becomes its location.
- full design + code shape: [world-observe.md](../world/world-observe.md). Implemented in `src/skyra/world/space.go` (`World.Observe`, `pointOf`, ephemeral points, conducting space).

### Slice result — DONE (as result maps, not a single `Result()`)

The wave result is recorded by the kernel into the traversed Slice; DNA never writes it.

- who records: [Kernel Recording Invariant](../laws/invariants/kernel-recording-invariant.md) — `Express` returns output; the kernel records it into `Slice.Maps()[impulse.ResultMap][cellID]`. `ResultMap` is a recording address, not primitive semantics.
- how the result returns to the actor: [Perception Invariant](../laws/invariants/perception-invariant.md) — a return perception wave rooted at the origin (`Realize(origin, origin, …)`), **not** an `origin.Observe` callback (that callback is retired).
- concurrency: return perception is serialized at the origin's Slice; mailbox / scheduled-delivery semantics are a later concern.

## Non-Goals (still hold)

- Do not reintroduce `ActiveState` on the kernel.
- Do not make the kernel durable, or a registry of all slices.
- Do not make parent waves wait for child actor waves.
- Do not make `Realize` return a result — it returns void.
