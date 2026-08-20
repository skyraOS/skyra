# skyra-dna — Implementation Specs (by cell type)

The being's cells. **One spec per cell type.** All implement `skyra.Space`; what differs is where each sits and what its `Observe`/`Express` do (`dna.md`: one type, topology differentiates — but these are distinct enough to spec separately).

## Blocker first

The whole package currently imports the **deleted `src/reality`** and does not compile. Step 1 in every spec: **port to `skyra.Space`**.

## The structure being built

```
Membrane (sensorimotor boundary — attention gate / motor; Content = root drive)
 └── MidLevel (vote tally / tie-break / grouping) — sees only groupings + crystallized understandings
      ├── Children: GROUPINGS — higher-order setpoints over the active field (the future,
      │     passed DOWN as context; replaces the parent edge). Recursive, loosely coupled.
      │     Each carries a GIST = an understanding that hasn't crystallized yet.
      ├── Agent: other / self — ONE predict-score-close pass PER GROUPING (constituents as material)
      └── cycle: observe/grouping (impulse vs active groupings, lit-set ctx; skip if no completed setpoint)
             → filter (kernel, activation × stress × alignment)
             → per-grouping CONCURRENT self/other (N≥2 waves, snapshot-in/join-out;
                   predict → grade → on closure: prediction→setpoint(n), constituents→children(n-1), gist→Understanding)
             → orchestrate the 2-D outcome (code): cascade > closure > open
                   (cascade = roll state back to solid ground ~0.2, delta→WM→Stress)
             → vote (Hopfield, pure math)
 └── Setpoints (predictions that came true — ONE node: Content = understanding (gist),
       Children = entities + responses + actions + stresses (closure snapshot). PAST only;
       no parent/future edge. setpoint(n) eligible as a constituent at n+1 — stacking.
       Created when predicted self AND other both close. Raw setpoints are BELOW MidLevel's altitude.)
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
| [sensorimotor.md](sensorimotor.md) | **the boundary system** (2026-06-25) — membrane (thin trampoline) + sensory/motor sub-actors + three-slice topology + cell-initiated crossing; **kernel returns void / Perception Invariant suspended for the build** | `[MVP]` |
| [membrane.md](membrane.md) | the sensorimotor boundary — attention gate (Observe) / motor (Express) *(boundary model superseded by `sensorimotor.md`; face-semantics still apply)* | `[MVP]` |
| [midlevel.md](midlevel.md) | MidLevel — vote tally / tie-break; agents `self` + `other`; Hopfield vote; reverse read = wave | `[MVP]` (minimal) |
| [midlevel-observe-gather.md](midlevel-observe-gather.md) | MidLevel's Observe **read-half** (2026-06-25) — gather candidate setpoints entity-outward from sensory's warm surroundings, score by overlap × recency (cell-written IndexScore), build `Children()`; inference call is the continuation | `[MVP]` |
| [midlevel-active-grouping.md](midlevel-active-grouping.md) | MidLevel's grouping pass, the **improvement** (2026-06-25) — active LLM-driven probe loop + drift-diffusion stop controller; passive single-call MVP ships first | `[improvement / post-MVP]` |
| [agent.md](../archive/agent.md) | the self/other agent (2026-06-26) — **one `AgentCell{role}`**; grade-then-predict loop (pop Prior → grade last bet → read activeGroupings → per-grouping traverse → grade-against-framing + predict → append); collapses the `SelfCell`/`OtherCell` stubs | `[MVP]` |
| [setpoint.md](setpoint.md) | confirmed predictions — content = understanding; children = entities + responses/actions/stresses | `[MVP]` (one) |
| [entity.md](entity.md) | person/place/thing — children = setpoints stamped at completion; the reverse-index hub | `[next]` |
| [episodic.md](episodic.md) | *(superseded — the binding-reality framing folded into the self/other agents)* | — |
| [temporal.md](temporal.md) | *(superseded — two-hemisphere model folded into the self/other agents)* | — |
| [action.md](../archive/action.md) | the action store — justification index | `[after]` |

> `membrane.md` **boundary model** — superseded by `sensorimotor.md` (2026-06-25): the membrane is a thin trampoline (Observe no-op, no children, fires a wave to sensory on Express); sensory/motor are separate sub-actors on their own slices; the surroundings write lives on **sensory**; the wave reaches cognition by a **forward spawn**, not as a child. The cognitive layer is unchanged.
> `coherence.md` — superseded by `midlevel.md` (2026-06-17).
> two-hemisphere model (WorkingMemory/Category) → one binding reality (2026-06-19) → **self/other agents** (2026-06-19, current).

**Conjectured / deferred, no spec yet:** Emotion (fast trigger for the recovery loop), Memory cells (trace/understanding — the base material under the binding realities).

## What dissolved from the old `binding/`

The old files don't map 1:1. **Motivation/Goal → methods** (descent/ascent of the setpoint, not cells). **Prediction/Attention → folded** (Prediction is what Coherence's Observe does; Attention is the membrane's two faces). **Prospective → folded into Goal/setpoint.** **Working memory → Coherence** (the held integrative synthesis — Baddeley's episodic buffer — not its own cell). So `binding/` gets *restructured*, not just ported.

## MVP

**membrane (sensorimotor boundary, stubbed pass-through for MVP) + one setpoint + Coherence (Predict/Resolve) + activation records.** The minimal being that traverses and lets the comparator score a *real* frame. Everything else is `[after]`.
