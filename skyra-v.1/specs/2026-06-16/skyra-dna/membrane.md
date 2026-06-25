# Membrane — Cell Type Spec

> **⚠️ Boundary model superseded (2026-06-25) — see [sensorimotor.md](sensorimotor.md).**
> The membrane is now a **thin trampoline** (Observe no-op, **no children**, fires a wave to
> sensory on Express); **sensory/motor are separate sub-actors on their own slices**, and the
> **surroundings write lives on sensory**, not here. The wave reaches cognition by a **forward
> spawn**, not by descending into Global as a child. The Observe/Express *face-semantics* below
> (attention gate / motor) still describe the boundary conceptually, but their **placement has
> moved to the sub-actors**, and the kernel currently **returns void** (Perception Invariant
> suspended for the build).

**Date:** 2026-06-16 · **Status:** skeleton (implementation stubbed for MVP; **ownership is explicit**)

## File
- `../../../src/skyra/skyra-dna/membrane.go` — currently imports the deleted `src/reality`; port to `skyra.Space`

## What it is
The being's **sensorimotor boundary** — *not* pure routing. It owns the one attention field, read from both faces ([world-kernel-coupling.md](../../../design-docs/kernel/world-kernel-coupling.md#attention-is-one-field-read-twice): "Attention is one field, read twice"):

- **Observe = the sensory attention gate.** *"Does this catch my attention?"* — an alignment check against working memory (and the active aperture/goal). Caught → **pass up** (traverse the binding realities). Not → **discard** (don't propagate). The inbound containment test.
- **Express = the motor.** Reads the returning votes; **above threshold AND agree → fire the action.** The outbound emission — basal-ganglia coincidence detection (aligned votes fire reflexively, no center needed).

> Supersedes the earlier "Observe = no-op pure routing." The membrane does real work at both faces. The old cell-types "gating lives in a separate Attention cell" line is **consolidated into the membrane** — attention *is* the membrane's two faces.

## Ownership (explicit)

| Face | Owns | Mechanism |
|---|---|---|
| **Observe** | the inbound **attention / relevance gate** | alignment vs. working memory → pass-up or discard |
| **Express** | the **motor** — coincidence + action | above-threshold-and-agree → fire; **conflict → defer to Coherence** |

**Not the membrane's:** global Predict/Resolve and the **conflict tie-break** — those are Coherence. The membrane fires on *consensus*; Coherence (ACC) engages *only on disagreement*. So:
- **Membrane Express = basal ganglia** (coincidence + fire — the fast path).
- **Coherence Resolve = ACC** (tie-break — the slow path).

This **places the basal-ganglia convergence** the voting section left unnamed ([owned-specialist §6](../../../design-docs/kernel/owned-specialist-crossing.md#6-the-voting-set)): it's the membrane's Express.

## Observe also writes the surroundings map (sensory intake)

The sensory face does more than gate — it's where perception **lands**. On a `perceive` result, the membrane's `Observe`:
1. **resolves** the perceived signal to stable **entity IDs** (the resolver lives here — perceived descriptor → stable id; without resolution the same person shows up as a new entity every time and overlap never accumulates),
2. **deposits** those entities into the **surroundings map** on the slice (the active-memory write).

This is "the sensory part receives signals and puts them in the right place in active memory" — and "the right place" *is* the entity resolution. Matches `episodic-encoding.md` ("working memory IS the attention map"; place/device/tool flow through Observe like keyboard input). Closes the open question of who owns the surroundings write: **the membrane's Observe.** See [entities-and-surroundings.md](../../../design-docs/cognition/entities-and-surroundings.md).

## Interface (`skyra.Space`)
- `Children()` → the brain (Global at the top)
- `Content()` → the **root drive**
- `Observe(nil, impulse)` → the attention gate + **sensory surroundings write** (gate is a pass-through stub for MVP; the surroundings write is real)
- `Express(...)` → the motor (**minimal/no-op stub for MVP**)
- `NewActivation()` → the membrane's activation record

## Changes
### 1. Port + stub  `[MVP]`
- [ ] implement `skyra.Space`; drop `src/reality`
- [ ] `Observe` = **pass-through stub** for the gate; ownership documented above
- [ ] `Express` = minimal (no action production yet)
### 2. Observe: surroundings write  `[next]`
- [ ] on `perceive` result: resolve to entity IDs (exact-match resolver first) → deposit into the surroundings map on the slice
### 3. Observe: the attention gate  `[after]`
- [ ] alignment check vs. working memory → pass-up or discard
- [ ] **efficiency:** cheap resonance-first (embedding similarity, the project's 0.55 threshold), escalate to an inference call only when borderline — it runs on *every* input
### 4. Express: the motor  `[after]`
- [ ] coincidence detection (above-threshold-and-agree) → produce / spawn the action
- [ ] **conflict → Global** (the membrane never breaks ties)

## Design sources
- [world-kernel-coupling.md](../../../design-docs/kernel/world-kernel-coupling.md#attention-is-one-field-read-twice) — Attention is one field, read twice
- [cell-types.md](../../../design-docs/being/cell-types.md) — Membrane (the pure-routing line is superseded here)
- [comparator.md](../../../design-docs/cognition/comparator.md) — basal-ganglia coincidence; Coherence tie-break

## Open
- root-drive home: membrane `Content` vs. a designated root setpoint
- the gate's aperture: working memory only, or also the active goal/setpoint?
- **entity resolution** in the surroundings write (exact-match first; fuzzy/embedding later) — the real work

## Done when
- **(MVP)** membrane implements `skyra.Space`, passes signals through (stub gate), and the kernel descends into Global — with the Observe/Express ownership documented even though stubbed.
- **(next)** Observe writes the surroundings map: resolve `perceive` results to entity IDs and deposit on the slice.
- **(after)** Observe gates by attention (alignment vs. working memory); Express fires actions on agreeing votes and defers conflict to Global.
