# The Kernel Recording Invariant

## Status

**Foundational.** This invariant defines who is allowed to write wave-produced active state into a Slice.

## The Invariant

> **DNA computes. Kernel records.**

In full:

> During a kernel wave, `Observe` and `Express` may compute local meaning and return local output, but they do not mutate traversal-owned Slice state. The kernel is the mechanism that writes activation records, traversal trees, result maps, and other wave-produced active state into the traversed Slice.

## What It Means

- **Space is durable behavior and structure.** A Space defines topology, content, observation behavior, and expression behavior.
- **Slice is active state.** A Slice records what became active during traversal.
- **The kernel manages Slice mutation.** During a wave, the kernel decides when and where traversal-produced active state is written.
- **The traversed plane owns the Slice.** The kernel does not own slices as durable state. It asks the traversed Space for `Slice()` and writes the wave's active state there.
- **`Observe` performs local intake.** It may update the Space's local traversal-facing behavior, such as preparing same-plane children, but it does not record result maps for the wave.
- **`Express` returns output.** It computes what this Space expresses for the current impulse and returns that output to the kernel.
- **The kernel records expression output.** If the impulse names a result map, the kernel writes the returned output into the traversed Slice at the current Space's ID.

Short form:

```text
Space/DNA computes local output.
Kernel writes wave state into Slice.
Slice belongs to the traversed plane.
```

## World Example

For a world action:

```text
Point.Express(origin, impulse) returns PointResult
kernel records PointResult into worldSlice.Maps()[impulse.ResultMap][pointID]
```

`PointResult` may include source metadata such as point ID and coordinates. That metadata is part of the expressed output being recorded; it is not a recording instruction and it does not change the kernel's return origin.

So a point does not write directly into the world Slice's perception or engagement map. It only expresses what is true at that point. The kernel records that expression as part of the wave.

The world may register maps such as `perception` and `engagement` on its Slice. Registration prepares recording addresses. It does not make points responsible for writing those maps.

## What It Forbids

- **No DNA-owned wave recording.** A Space's `Express` must not write its own output into the Slice result maps.
- **No hidden aggregation in points.** A point may report its content, neighbors, and primitive result, but it must not decide how the wave aggregates those reports.
- **No kernel primitive semantics.** The kernel may read the impulse's recording address, such as `ResultMap`, but it does not interpret the primitive's meaning.
- **No slice ownership in the kernel.** The kernel manages mutation during traversal; it does not become a durable registry of all slices.
- **No bypassed active-state path on the kernel route.** If a wave goes through the kernel, wave-produced active state is recorded by the kernel.

## Enforcement Boundary

This invariant is guaranteed on the kernel path. Code that mutates a Slice outside a kernel wave is outside the runtime mechanism and must be treated as explicit developer-owned policy.

Off-kernel mutation is not forbidden globally. It must be named and owned by the code that does it. It is not the default wave recording path.

## Relationships

- [active-state-invariant.md](active-state-invariant.md) — the kernel traverses Space and realizes active state into the Slice.
- [plane-slice-invariant.md](plane-slice-invariant.md) — many Spaces can share one plane Slice.
- [perception-invariant.md](perception-invariant.md) — wave results return through the origin as perception waves.
