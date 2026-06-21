# skyra-dna — Implementation Specs (by cell type)

The being's cells. **One spec per cell type.** All implement `skyra.Space`; what differs is where each sits and what its `Observe`/`Express` do (`dna.md`: one type, topology differentiates — but these are distinct enough to spec separately).

## Blocker first

The whole package currently imports the **deleted `src/reality`** and does not compile. Step 1 in every spec: **port to `skyra.Space`**.

## The structure being built

```
Membrane (sensorimotor boundary — attention gate / motor; Content = root drive)
 └── Global (vote tally / tie-break / grouping) — sees only groupings + crystallized understandings
      ├── Children: GROUPINGS — higher-order setpoints over the active field (the future,
      │     passed DOWN as context; replaces the parent edge). Recursive, loosely coupled.
      │     Each carries a GIST = an understanding that hasn't crystallized yet.
      ├── Agent: other / self — ONE predict-score-close pass PER GROUPING (constituents as material)
      └── cycle: observe/grouping (impulse vs active groupings, buffer ctx; skip if no completed setpoint)
             → filter (kernel, activation × stress × alignment)
             → per-grouping CONCURRENT self/other (N≥2 waves, snapshot-in/join-out;
                   predict → grade → on closure: prediction→setpoint(n), constituents→children(n-1), gist→Understanding)
             → orchestrate the 2-D outcome (code): cascade > closure > open
                   (cascade = roll state back to solid ground ~0.2, delta→WM→Stress)
             → vote (Hopfield, pure math)
 └── Setpoints (predictions that came true — ONE node: Content = understanding (gist),
       Children = entities + responses + actions + stresses (closure snapshot). PAST only;
       no parent/future edge. setpoint(n) eligible as a constituent at n+1 — stacking.
       Created when predicted self AND other both close. Raw setpoints are BELOW Global's altitude.)
 └── Entities (person/place/thing) — Children = setpoints stamped on them at completion.
       The reverse-index hub: a wave from the surroundings spreads understanding↔entity (associative
       recall), bounded by amplitude. No hash, no scan. (entities-and-surroundings.md)
 Surroundings map — on the slice; the active entities, updated on every `perceive`.
 Action store (actions ⨯ justifying-assumptions) — read by the self-recovery pass
```

> Pipeline detail: [../../design-docs/cognition/groupings-and-understandings.md](../../../design-docs/cognition/groupings-and-understandings.md) — gist = uncrystallized understanding; grouping/gist and episode-closure are the same pipeline at two altitudes.

## Cell types

| Spec | Cell type | MVP? |
|---|---|---|
| [membrane.md](membrane.md) | the sensorimotor boundary — attention gate (Observe) / motor (Express) | `[MVP]` |
| [global.md](global.md) | Global — vote tally / tie-break; agents `self` + `other`; Hopfield vote; reverse read = wave | `[MVP]` (minimal) |
| [setpoint.md](setpoint.md) | confirmed predictions — content = understanding; children = entities + responses/actions/stresses | `[MVP]` (one) |
| [entity.md](entity.md) | person/place/thing — children = setpoints stamped at completion; the reverse-index hub | `[next]` |
| [episodic.md](episodic.md) | *(superseded — the binding-reality framing folded into the self/other agents)* | — |
| [temporal.md](temporal.md) | *(superseded — two-hemisphere model folded into the self/other agents)* | — |
| [action.md](action.md) | the action store — justification index | `[after]` |

> `coherence.md` — superseded by `global.md` (2026-06-17).
> two-hemisphere model (WorkingMemory/Category) → one binding reality (2026-06-19) → **self/other agents** (2026-06-19, current).

**Conjectured / deferred, no spec yet:** Emotion (fast trigger for the recovery loop), Memory cells (trace/understanding — the base material under the binding realities).

## What dissolved from the old `binding/`

The old files don't map 1:1. **Motivation/Goal → methods** (descent/ascent of the setpoint, not cells). **Prediction/Attention → folded** (Prediction is what Coherence's Observe does; Attention is the membrane's two faces). **Prospective → folded into Goal/setpoint.** **Working memory → Coherence** (the held integrative synthesis — Baddeley's episodic buffer — not its own cell). So `binding/` gets *restructured*, not just ported.

## MVP

**membrane (sensorimotor boundary, stubbed pass-through for MVP) + one setpoint + Coherence (Predict/Resolve) + activation records.** The minimal being that traverses and lets the comparator score a *real* frame. Everything else is `[after]`.
