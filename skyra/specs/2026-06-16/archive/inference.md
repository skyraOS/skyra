# Inference — Implementation Spec

**Date:** 2026-06-16 · **Status:** skeleton — fill in during execution

## Files
- `../../src/inference/inference.go` — `Call(system, present)`
- `../../src/inference/*.go` — providers (anthropic, claude, deepseek, embed)

## What it is
Inference is a **capability, not a cell** — a **method every cell carries, injected at boot** (same pattern as the activation factory `NewActivation`, "set by genome at boot"). No cell imports the `inference` package directly. The cell **owns** the decision: *when* to fire (inside its `Observe`/`Express`) and *what* to send (its composed prompt-fragment + present). Inference is the wired-in method it calls.

## Design sources
- [comparator.md](../../design-docs/cognition/comparator.md) — the `<comparison>`/`<alignment>` output
- [owned-specialist-crossing.md](../../design-docs/kernel/owned-specialist-crossing.md) — §7 atomic generate+align, composition boundary
- [setpoints.md](../../design-docs/cognition/setpoints.md) — alignment = self-confidence
- [dna.md](../../design-docs/being/dna.md) — "system prompt composed from DNA fields"

## Changes

### 1. Generate + align in ONE call (shared trunk)  `[MVP]`
- A single call emits the cells/predictions **and** their `<alignment>` in one structured output. **No second scoring pass** (a disconnected critic drifts).
- [ ] a frame-call that returns predictions + alignment together
- [ ] tolerant parser for `<comparison>`/`<alignment>` (the .1001 runs showed format drift — tolerate the bare `<alignment>` variant)

### 2. Inject the method at boot  `[MVP]`
- [ ] cells carry an inference reference, wired at registration (see [boot.md](boot.md)) — like `NewActivation`. Home: a cell field or reached via the slice (small detail to settle in execution).
- [ ] cells call it from `Observe`/`Express`; they never `import` the provider package

### 3. Composable prompt-fragments  `[after]`
- Assemble the frame's system prompt from the **active operations** ([dna.md](../../design-docs/being/dna.md): "prompt composed from DNA fields"). Operations are fragments — data, not code.
- [ ] fragment registry + assembler
- [ ] composition boundary: couple-for-consistency (generate+align one call) vs. fan-out-for-independence; cap fragments to avoid the mega-call ([§7](../../design-docs/kernel/owned-specialist-crossing.md))

### 4. Confidence  `[after]`
- self-side: **confidence = alignment** (no separate field). other-side: epistemic confidence (separate quantity).

## Dependencies
- the injection is wired in [boot.md](boot.md)
- consumed by [skyra-dna/](skyra-dna/README.md) (membrane gate, Coherence Predict/Resolve, setpoints all fire it)
- existing `Call` (OpenRouter) is the substrate — extend, don't replace

## Open
- the method's home: a field on every cell vs. reached through the slice
- prompt-fragment registry shape; mega-call guard threshold

## Done when
- **(MVP)** a cell, from its `Observe`/`Express`, fires the injected inference method and gets graded predictions (cells + `<alignment>`) back in one call — the comparator running on a *real* traversed present.
