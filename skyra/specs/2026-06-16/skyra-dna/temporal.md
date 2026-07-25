# Temporal Binding Reality — Cell Type Spec

**Date:** 2026-06-16 · **Status:** skeleton · **`[after]`** (not on the MVP path)

## ⚠️ Naming — open (pin before building)
This is the **focused / deep / single-active** reality — called **"categorical"** earlier, **"temporal"** now. **The function is locked; the name is not.** "Temporal" fits if you read it as the **throughline** (the goal held *persistently over time*); "categorical/semantic" fits if you read it as the **atemporal goal structure**. Those characterize the same reality slightly differently — pick the name to match the reading. Function is load-bearing; the name is a label.

## Files
- `../../../src/skyra/skyra-dna/temporal.go` — **new** (or rename `categorical`)
- self/other sub-realities — 2 hard-separate stores

## What it is
The **focused** memory reality — the **spotlight**. Deep within a **single** active setpoint; **only one active at a time** (the attention bottleneck — focal attention is unitary). **Narrow / local / focal.**

Three registers, same character: Tulving **semantic/structural** (hierarchical, deep); wave aperture **narrow** (*penetrates*); hemisphere **local**; attention **focal**.

## Access pattern: one, deep
- One **root** setpoint active (one topic/goal); its substructure (the subgoals) is the depth. "One active" is one *root*, not one node.
- **Winner-take-all** at each level → one path, deep.
- **Slow, stable, persistent** (the throughline held across the session).

## Sub-realities (self/other — structural)
- `temporal-self`, `temporal-other` — two genuinely separate stores.

## Coupling with Episodic
**Temporal = the spotlight; Episodic = the field.** The spotlight selects the field's peak and goes deep. **The spotlight moves** when a new peak captures it (bottom-up — fed by the membrane attention gate) **or holds** when top-down goal persistence resists the distractor. The **move-vs-hold** control is the open question (focal-attention control / "should I switch goals").

## Changes
### 1. Port + restructure  `[after]`
- [ ] new `temporal.go` (or rename `categorical`); the focused-deep access (winner-take-all, one active)
- [ ] `self`/`other` sub-realities
- [ ] slow / stable activation dynamics ("same structure ≠ same temperature" vs. Episodic)
### 2. Voting  `[after]`
- [ ] generates a focused/deep self-model with confidence/alignment → votes

## Design sources
- [self-and-other-modeling.md](../../../design-docs/cognition/self-and-other-modeling.md), [setpoints.md](../../../design-docs/cognition/setpoints.md)

## Open
- **the name** (temporal vs. categorical vs. semantic vs. focal)
- **move-vs-hold:** what moves the spotlight — bottom-up capture vs. top-down persistence

## Done when
- **(after)** the focused reality — one active setpoint, deep, self + other — compiles and traverses; the spotlight selects the field's peak and goes deep.
