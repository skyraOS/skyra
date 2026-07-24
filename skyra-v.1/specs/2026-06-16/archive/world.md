# World — Implementation Spec

**Date:** 2026-06-16 · **Status:** skeleton — fill in during execution

## Files

- `../../src/skyra/distributed-agent/world/space.go` — `World`, `Point`
- `../../src/skyra/distributed-agent/world/slice.go` — `Slice`

## Design sources

- [world-kernel-coupling.md](../../design-docs/kernel/world-kernel-coupling.md) — Space owns firing, the two crossing modes
- [kernel-change-plan.md](../../design-docs/kernel/kernel-change-plan.md) — §2 (A3: `world.observe`)
- [slice-and-world.md](../../design-docs/world/slice-and-world.md) — the recursion break
- [owned-specialist-crossing.md](../../design-docs/kernel/owned-specialist-crossing.md) — Space owns firing

## The world's job: facilitate the crossings the kernel shed

The kernel is now pure transport — it deleted the actor crossing ([kernel.md](kernel.md) #1). **The World takes it over.** It is the natural home because it already holds the grid (where beings are placed) and the kernel (`world.New(kernel)`). Crucially this keeps **Space-owns-firing**: the firing logic now lives in a *Space* (the Point/World, owning its own crossing by competence), not in the kernel deciding. A Point's competence is "fire the being I contain" — the kernel just walks it.

## Changes

### 1. Place a being  `[MVP]`
- `Space.Place(space, x, y)` already exists.
- [ ] confirm a membrane can be placed and reached (`pointOf` → `Realize`)
- **Note:** the MVP roots the wave **directly at the being** (`kernel.Realize(being, being, signal)`), so the crossing below is *not* exercised yet. #1 is the only MVP item here.

### 2. Point checks `Executables`, fires the crossing  `[after]`
Replaces the kernel's old Content-crossing.
- [ ] `Point.Executables()` = its occupant being(s). (Apply the `Content`/`Executables` split: `Content()` = terrain/data; `Executables()` = the beings to fire. Reconcile `perceive`/`engage`, which currently read `Content()` for the occupant — perception reports "who's here" from `Executables`.)
- [ ] When a Point is traversed (in its own `Observe`/`Express`), it checks its `Executables` and fires each occupant's wave via the World's kernel — origin per crossing mode (**peer:** origin = the being; **owned:** origin = parent).
- The **kernel never reads `Executables`; the Point (world code) does.** Pure-transport kernel preserved.

### 3. `world.observe` — the outward-action facilitator (A3)  `[after]`
- [ ] `world.observe(origin, content)`: **origin != nil** → a being is acting → relay to the kernel (fire a wave); **origin == nil** → being walked as a realm → intake, **don't relay** (the recursion break).
- [ ] wire the environment ref at registration (chain: being → world → kernel)

## In-plane vs. crossing — what actually crosses

- **Inner being structure** (binding realities, self/other sub-realities) is **in-plane** for now — one being, one slice, traversed as `Children`. No inner crossing, no firing.
- **Crossings are peer-level:** grid Point → being, being → being. *These* are what the world handles (#2).
- **Concurrency caveat:** making Coherence's specialists run on separate slices (parallel inference) would turn the inner fan-out into a crossing too — **deferred**. The membrane-vs-world environment pin defers with it (see Open).

## Dependencies

- #2/#3 consume the pure-transport kernel + the `Executables` interface split ([kernel.md](kernel.md) #2, `[after]`).
- [boot.md](boot.md) uses placement (#1).

## Open

- **Environment-layer pin:** with the inner structure in-plane, the only crossings are peer-level (world-handled) — so the **membrane may never need to be a facilitator at all.** Revisit only if/when inner concurrency is wanted.

## Done when

- **(MVP):** a being can be placed and reached (or rooted directly for the first run).
- **(after):** the world fires the beings its Points contain (the crossing the kernel shed), and a being acts outward via `world.observe` with the result returning to its origin.
