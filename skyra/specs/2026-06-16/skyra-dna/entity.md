# Entity — Cell Type Spec

**Date:** 2026-06-20 · **Status:** in spec (next to build)

## File
- `../../../src/skyra/skyra-dna/entity.go` — **new**

## What it is
A **person / place / thing** — a cell (`skyra.Space`) with a stable **ID** and a **type**. Entities are nodes in the same graph as setpoints. Their `Children()` are the **setpoints stamped on them** at completion — which makes the entity the routing hub for associative recall. See [entities-and-surroundings.md](../../../design-docs/cognition/entities-and-surroundings.md).

## Interface (`skyra.Space`)
- `ID()` → the **stable entity id** (the resolution target)
- `Content()` → the entity descriptor / type (person / place / thing)
- `Children()` → the **setpoints stamped on this entity** (the reverse index, as graph edges, keyed by edge identity / map key)
- `Observe` → intake (stub for MVP)
- `Express` → contributes to the wave (stub for MVP)
- `NewActivation()` → edge-owned activation record

## How it's used
- **Surroundings map** (on the slice, updated on every `perceive`) holds the currently active entities. **Written by the membrane's `Observe`** (sensory intake): resolve the perceive result to entity IDs, deposit on the slice.
- **Reverse read = a wave fired from the surroundings entities** — it spreads `understanding → entity → understanding → entity …`, bounded by amplitude. Not a hash lookup; retrieval as traversal.
- **Overlap → priority.** A setpoint touched by more of the active surroundings entities scores higher — that overlap *is* the `IndexScore` term in the bump (`IndexScore × Alignment × stress`). Computed **once per cycle** (one-shot count / wave convergence), feeding a single bump — not N bumps for N entity-waves (that would over-groove the EMAs).
- **Write-time invariant:** on setpoint completion, append the setpoint to each involved entity's `Children()`. The reverse read depends on this append.
- **Decay automatic:** each `entity → setpoint` edge is edge-owned activation; decayed edges fall below threshold and the wave skips them (no cleanup).
- **Salience from branching:** a hub entity (many setpoint-children) divides amplitude thin; a rare entity spreads focused.

## Changes
### 1. Entity cell  `[next]`
- [ ] implement `skyra.Space`; stable `ID`, type ∈ {person, place, thing}
- [ ] `Children()` = stamped setpoints (edge-keyed)
### 2. Resolver / registry  `[next]`
- [ ] `descriptor → stable entity-id` (exact-match first; fuzzy/embedding later)
### 3. Surroundings map  `[next]`
- [ ] on the slice; updated on `perceive`; the active entity set (+ stress)
- [ ] **written by the membrane's `Observe`** (sensory intake) — resolved.

## Open
- **Entity resolution** is the real work (`perceived descriptor → stable id`)
- Physical entities (place/device/tool) via the world/`perceive`; cognitive (person/goal) from the runtime — same map

## Done when
- An `Entity` implements `skyra.Space`; setpoints append themselves to involved entities' children at completion; a wave fired from the surroundings entities surfaces the relevant setpoints (associative recall) bounded by amplitude.
