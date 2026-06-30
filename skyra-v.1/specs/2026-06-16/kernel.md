# Kernel — Implementation Spec

**Date:** 2026-06-16 · **Status:** skeleton — fill in during execution

## Files

- `../../src/skyra/relation/relation.go` — the kernel (`traverse`)
- `../../src/skyra/relation/relation_test.go` — remove the crossing test
- `../../src/skyra/space.go` — the interface (only for the `[after]` Executables split)

## Design sources

- [owned-specialist-crossing.md](../../design-docs/kernel/owned-specialist-crossing.md) — "Pending revision — Space owns firing", §7
- [world-kernel-coupling.md](../../design-docs/kernel/world-kernel-coupling.md)
- [kernel-change-plan.md](../../design-docs/kernel/kernel-change-plan.md)
- [self-and-other-modeling.md](../../design-docs/cognition/self-and-other-modeling.md) — the Content/Executables split surfaced here

## Changes

### 1. Remove the actor crossing  `[MVP]`

**The only kernel-logic change for the MVP.** The kernel becomes pure transport (in-plane DFS + deliver-to-origin); it never reads `Content`. In `traverse()`, on the ascent/pop branch, delete these 7 lines (~86–92):

```go
// boundary crossing — Content is a Space → another actor.
// Spawn an independent wave whose origin is the inner actor.
// Return-perception waves do not cross: perceiving a result is
// inward and must not re-stimulate the actors in the topology.
if inner, ok := top.Cell.Content().(skyra.Space); ok && !isReturnImpulse(impulse) {
	go r.realize(inner, inner, impulse)
}
```

- [ ] delete the crossing spawn (lines ~86–92)
- [ ] update the `traverse` doc-comment (lines ~42–44, "A cell whose Content is a Space is another actor…") → "pure transport, no crossing"
- [ ] **remove the crossing test** in `relation_test.go` (~lines 300–365, "A Content Space is another actor: the crossing spawns an independent wave") — it exercises the spawn that no longer exists, so it won't compile/pass

**Consequence:** with the crossing gone, the kernel can't get a wave into *any* actor (no peer-being, no fan-out) until `world.observe` exists ([world.md](world.md), `[after]`). That's fine for the MVP — **boot roots the wave directly at the being** (`kernel.Realize(being, being, signal)`) and traverses its binding realities in-plane. No crossing needed.

### 1b. `origin` = the wave owner, both faces  `[DONE]`

`Observe` and `Express` both take `origin`, and it means **one thing in both: the wave owner** (who the wave returns to), always non-nil during traversal. The old "origin is nil during traversal" convention is gone — it existed only because `Observe` was overloaded as the outward-fire entry (`World.Observe(self)`). That overload is removed:

- [x] kernel passes the wave owner to `Observe` (was `nil`)
- [x] `Point.Observe`'s `origin != nil` guard removed (Observe is always traversal intake now)
- [x] `World.Observe` stays the **entry** (the run goes through Observe) but delegates wave-rooting to a **`World.Emit(owner, impulse)`** helper — the seed of the `world.observe` facilitator. A being's *own* Observe is pure intake; it acts outward by calling `World.Observe`/`Emit`, not through its own Observe.
- [x] interface doc + tests updated (`TestTraversalObserveUsesOwner`)

**The sender is separate from the owner.** "Who sent this signal" (for entity resolution / the surroundings map) rides on **`impulse.Source`**, not on `origin`. origin is "who I answer to"; Source is "who/where this came from." (Impulse-struct shape to be settled later.)

### 2. Space owns firing — `Agents()` on the interface  `[after]`

Not needed for the MVP kernel traversal. Part of the firing rewrite.

- File: `space.go`
- [ ] add `Agents() map[string]Space` to `skyra.Space` — every cell declares its dispatchable sub-agents (most return empty map)
- [ ] separate from `Content() any` (kernel-blind data — what the cell *is*) and `Children()` (traversed in-plane)
- [ ] cells dispatch agents via their environment (`world.observe`), not the kernel traversal
- **Note:** previously called `Executables()` — renamed to `Agents()` (sub-agents are beings, same `Space` interface)

### 3. Owned-specialist map-crossing — DO NOT BUILD

The `map[string]Space` Content crossing (owned-specialist-crossing §1) is **superseded** by `Agents()` + Space-owns-firing. MidLevel dispatches its agents via its environment. Skip the kernel map-crossing entirely.

### 4. Slice — `Result` introspection field  `[after]`

- [ ] confirm `Result()`/`SetResult()` present (they are in `slice.go`); not on the MVP path (kernel-change-plan A1 remainder).

## Graph topology — one map, two edge directions

`Children() map[string]Space` on `skyra.Space` is the **single map** the kernel walks. For setpoint nodes it holds **past edges only** (Understandings, Stresses, Responses, Actions — the confirmed past). There is **no parent/future edge**: the future is not stored per-cell — it lives as **MidLevel's grouping index** over the active field and is passed down as context (see [midlevel.md](skyra-dna/midlevel.md)). No `Parents()` method.

The kernel handles this as-is:
- Iterates the flat map (one direction now — past)
- Visited-once rail (`onPath`) prevents cycles from back-links
- Amplitude divides across candidates; deeper/quieter cells survive only on leftover amplitude
- New cells appended mid-cycle don't appear until the next traversal (`scoreChildren` already ran)

## Edge-owned activation — key off the map key  `[DONE]`

**Activation is owned by the edge, not the node** (the synaptic model) — A→B and C→B are independent weights, not one number on B. This needs **no new interface and no edge-node objects**. Records still live on the slice; the kernel just keys them off the **map key** the cell hands it, not the child's own `ID()`:

```go
// was:  records[cell.ID()]            — node-keyed: one record per cell
// now:  recordFor(key, cell, sl)      — keyed off the map key the cell provided
func recordFor(key string, cell skyra.Space, sl skyra.Slice) skyra.ActivationRecord {
	rec := records[key]
	...
}
```

- [x] `recordFor` takes an explicit `key`; root keys by `cell.ID()`, `scoreChildren` keys by the parent's **map key** for the child
- [x] a cell achieves edge-ownership by handing the kernel **edge-scoped map keys** (`"A→B"` vs `"C→B"`) — the records then differ automatically
- [x] **world-safe / oblivious:** the world's map keys already equal its point IDs (`children[p.ID()] = p`), so node-keying is preserved there with zero change; the kernel never learns what an "edge" means — it keys a record off whatever identity it was handed
- [x] verified no-op against existing `relation_test.go` (all current maps key by cell ID)

**The slice stays the home for records** (no global-table retirement needed). The cell owns its `Children()` *adjacency* (which edges exist); the slice owns the edge-keyed *weights*. Footprint is `O(actual edges)` because edge keys only exist for edges a cell actually declares.

### Transitive reach is a path, not an edge

A only declares edges to its **direct** neighbors. A 2-hop cell (A→B→C) is reached *through* B; A's excitation of C is the **path composition** `ab ⊗ bc`, computed at traversal time by the existing amplitude attenuation (`activation × amplitude-that-arrived`). Not stored. A-specific because the amplitude started at A. No new machinery — `amplitudeDecay` already does this (spend-then-spill is emergent: direct edges spend the loud amplitude, leftover attenuates to potentials).

### Path → edge crystallization (Hebbian)  `[after]`  — DNA-side, not kernel

A potential (path-value) that grades above the keep-threshold (the same `0.1` that gates adding stimulus to children) **crystallizes into a direct edge**: the source cell adds the target to its own `Children()` adjacency (a DNA-side write — the kernel doesn't do this), and the edge-keyed record gets seeded on the slice. Next traversal the kernel reaches it directly. Edges that stop earning grades decay below threshold and fall away — the graph forms *and* prunes. **History** (scored potentials that didn't crystallize, including negatives) = edge-keyed records on the slice with **no** corresponding adjacency: they persist with their value but the kernel never walks them (it only walks `Children()`).

## Not in scope (existing gaps, deferred)

- **Decay** (`Activate(false)` on absent children) is unwired — deferred per kernel-change-plan; irrelevant to a single-traversal MVP.

## Dependencies

- **#1 has no dependencies — it's a deletion.**
- #2 blocks on [world.md](world.md) (the environment ref to fire through) and the environment-layer pin (below).

## Open

- **Environment-layer pin:** membrane vs. world as a cell's environment ([owned-specialist-crossing Open](../../design-docs/kernel/owned-specialist-crossing.md#open-questions)). Decide before #2 — does **not** block the MVP.

## Done when

- **(MVP):** the crossing is gone, the two comments updated, the crossing test removed; the kernel is pure transport, and a being rooted directly (`Realize(being, being, signal)`) traverses its binding realities in-plane and returns its result to origin.
- **(after):** `Executables()` on the interface; cells fire via `world.observe`.
