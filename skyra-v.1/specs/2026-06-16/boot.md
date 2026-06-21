# Boot — Implementation Spec

**Date:** 2026-06-16 · **Status:** skeleton — fill in during execution

## Files

- `../../main.go` — currently a stub (no being placed)
- `../../genome.skyra` — the declaration

## Design sources

- [kernel-change-plan.md](../../design-docs/kernel/kernel-change-plan.md) — §2 (A3 boot/registration), the loop
- [setpoints.md](../../design-docs/cognition/setpoints.md) — root drive in the genome
- [world-kernel-coupling.md](../../design-docs/kernel/world-kernel-coupling.md) — `Signals()` / `Receive()` loop

## Changes

### 1. Genome: the root drive  `[MVP]`
- [ ] add a `~drive "..."` / `~serves <name>` field to the `being` line ([setpoints DNA Placement](../../design-docs/cognition/setpoints.md#dna-placement)). Prototype value: `be helpful to the user`.

### 2. Boot wiring  `[MVP]`
- [ ] build the **durable** kernel (once); build the world
- [ ] register the being: install the root drive into the chosen home (membrane `Content`), assign its slice plane, wire the environment ref down `[after]`
- [ ] **inject the inference method into every cell** at registration (the capability all cells carry — see [inference.md](inference.md) #2; same pattern as `NewActivation`)
- [ ] place the being on the grid

### 3. The run loop  `[MVP]`
- [ ] feed one signal (`"hi"`) → root a wave at the being → traverse → Coherence scores a real frame
- [ ] `[after]` the full `for sig := range world.Signals() { world.Receive(sig) }` loop on the one persistent kernel

## Dependencies
- **last** — needs kernel, world, skyra-dna, inference

## Done when
- **(MVP / the milestone):** `go run` boots a being, feeds a signal, and the comparator scores a **real traversed frame**. This is the captain's-log first priority and the `.1004` precondition — the point where the architecture stops being argued and starts being measured.
