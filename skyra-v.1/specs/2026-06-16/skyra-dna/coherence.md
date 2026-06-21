# Coherence — Cell Type Spec (the central cell / global workspace)

**Date:** 2026-06-16 · **Status:** skeleton (ownership explicit; body stubbed for MVP)

## File
- `../../../src/skyra/skyra-dna/coherence.go` — port + rewrite

## What it is
The central cell — the **orchestrator between the two binding realities** (categorical + episodic). The ACC/frontopolar labels dissolve into this function: it's the **global workspace** (GWT). It integrates the realities' outputs, holds the integrated result, and broadcasts it back. One cell, two phases:

- **Observe = synthesize + broadcast + hold.** Reads both realities' predictions + assumptions (categorical + episodic, self + other), produces a synthesis — **"what went right / what went wrong"** (the global alignment) — throws it down to each reality (the broadcast / prime-down), and **holds it**.
- **Express = the vote / verdict.** Watches the realities' votes. **Casts a deciding vote ONLY when their signals don't align** (the tie-break — vote-exceptional). When they align, it stays silent and the membrane's coincidence fires. On unresolved conflict → gather ("what am I missing?") or inner-voice re-entry.

## Ownership (explicit)

| Face | Owns |
|---|---|
| **Observe** | synthesize across both realities (what went right / wrong), broadcast to each, **hold the synthesis** |
| **Express** | the **voting verdict** — vote *only* to break a tie (signals don't align); else silent (the membrane fires) |

**Boundary with the membrane:** aligned votes → the **membrane** fires (coincidence, fast path); conflict → **Coherence** votes (slow path). Coherence is the center that engages *only on disagreement* — consistent with [membrane.md](membrane.md) ("the membrane never breaks ties").

## Coherence holds working memory

The synthesis Coherence holds is **the integrative piece of working memory** — Baddeley's **episodic buffer**: the transient, unified "what's happening now" bound across the realities.

- **Transient** — refreshed each cycle, decays. ⚠️ This makes Coherence's `Content` *transient*, unlike the membrane's *durable* root drive — **same slot, opposite lifetimes** (be explicit about this when building).
- **It is a sliding window — and the windowing is Coherence's, not the binding realities'.** Working memory is a window over *recent* activity that **slides** forward (carries recent cells, sheds old). It **overlaps**: a cell stays in for several slides, so it sits in multiple consecutive windows — **membership overlap, not copied storage** (the multi-context "duplication," a feature: the same event reachable from more contexts). The binding realities are durable stores; Coherence windows the recent slice of them. (Their per-reality `category-prediction-active` is an *active set* — keep-still-active + new — not a window; the windowing lives here.)
- The realities keep their **durable** memory; Coherence keeps the transient **global** synthesis.
- **This dissolves the separate WorkingMemory binding reality** — working memory is *what Coherence holds*, not its own cell. (The genome `working_memory` bind → Coherence's held synthesis.)
- **The membrane's attention gate aligns against this** — the door checks "does this catch my attention" against Coherence's held now-picture.
- **"What went wrong"** in the synthesis is the trigger for the [error-recovery loop](../../../design-docs/cognition/error-recovery-loop.md).

## Interface (`skyra.Space`)
- `Observe` → synthesize / broadcast / hold (generate+align, one inference call across the realities)
- `Express` → the vote verdict (silent on consensus; vote on conflict)
- `Content()` → the held synthesis (**transient** working memory)
- `Children()` → the two binding realities (categorical, episodic)
- fires inference via [inference.md](../inference.md) (one call)

## Changes
### 1. Port + minimal Observe/Express  `[MVP]`
- [ ] port to `skyra.Space`
- [ ] `Observe` = synthesize over the (one, for MVP) setpoint → alignment (generate+align)
- [ ] hold the synthesis on the slice (working memory)
- [ ] `Express` = minimal verdict (MVP: one reality, no conflict → pass through)
### 2. Orchestrate two realities  `[after]`
- [ ] synthesize across categorical + episodic (self + other); broadcast to each
- [ ] `Express` vote-on-conflict (tie-break); gather / inner-voice on unresolved
### 3. Two-speed: fast (membrane) / slow (Coherence audit)  `[after]`

## Design sources
- [comparator.md](../../../design-docs/cognition/comparator.md) — global workspace, vote-exceptional
- [owned-specialist-crossing.md](../../../design-docs/kernel/owned-specialist-crossing.md), [error-recovery-loop.md](../../../design-docs/cognition/error-recovery-loop.md)
- Baddeley's episodic buffer (working memory)

## Open
- when the held synthesis decays / refreshes (per-cycle? attention-weighted?)
- the raw recent-exchange buffer: also Coherence, or separate from the integrative piece?

## Done when
- **(MVP)** Coherence's `Observe` synthesizes a **real traversed frame** into an alignment and holds it (working memory); `Express` passes through on consensus. The `.1004` precondition.
