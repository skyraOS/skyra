# Slice-Based Dispatch — Routing off the Slice, not off Hardcoded Pointers

**Date:** 2026-07-03 · **Status:** in spec (the refactor that makes the dispatch model real) · **Parents:** [dispatch-model.md](dispatch-model.md) (the transition tables this makes executable), [two-stroke-cycle.md](two-stroke-cycle.md) (scenario 4 / [B]), [memory-as-world.md](../../design-docs/cognition/memory-as-world.md) (affordances)

## The gap this closes

The dispatch-model spec describes routing as `cell reads slice-state →
parse-condition → fires the selected affordance`. **The build does not do this.**
Audited on `accumulation-build` (2026-07-03), the actual routing is:

- **Hardcoded successor pointers.** Each cell holds a direct reference to its
  successor(s) and spawns it in `Express`: membrane→`m.sensory`, sensory→`s.midLevel`,
  MidLevel→`g.other`, other→`o.self` (or `o.midLevel` on reframe), self→`s.motor`
  (or `s.midLevel` on reframe). The topology is wired into the structs at
  construction; nothing is *selected*.
- **Per-wave routing state on CELL FIELDS.** The go/no-go and branch decisions read
  mutable fields on the long-lived cell object — `o.reframed`, `o.verdict`,
  `s.plannedThisWave`. Set in `Observe`, read in `Express` of the same wave.

Two problems fall out:

1. **Not concurrency-safe (scenario 4).** Two waves through the *same* cell clobber
   its per-wave fields (`o.reframed` etc.) — proved by the race detector. This is
   why "substrate 3/3" is *not* a locking problem: the state is on a shared cell,
   so no lock granularity fixes it. It has to *move*.
2. **Not the affordance model.** "Which cell fires next" is fixed wiring, not a
   function of `(parse-condition, slice-state)`. The transition tables in
   [dispatch-model.md](dispatch-model.md) are documentation, not executable.

## The target

**Per-wave routing state lives off the cell; `Express` reads it and fires the
affordance the state selects — the affordance is looked up, not hardcoded.**

Three moving parts:

### 1. The wave carries its own routing state (not the cell)

Per-wave decision state (`reframed`, `plannedThisWave`, the verdict) is exactly
that — *per wave*. Its home is the wave, never a field on a cell shared across
waves. Two candidate homes, decide in build:

- **On the impulse** (lean). The impulse *is* the per-wave object; `Route` already
  carries forward-flowing routing (verdict, reframe-reason). Extend it to carry the
  wave's dispatch decision. `Observe` writes the decision onto the impulse; `Express`
  reads it. Because each wave has its own impulse, there is no cross-wave clobber —
  the scenario-4 race disappears **by construction**, no lock. (Requires the impulse
  reach Express as a mutable per-wave value — a pointer, or a per-wave scratch the
  kernel threads through both faces.)
- **On the slice, keyed by wave id.** A per-wave routing record on the slice under a
  wave key the kernel assigns; cells read/write their decision under that key;
  cleaned up when the wave completes. Literally "off the slice," but needs a wave-id
  + lifecycle. Heavier than the impulse option.

**Lean: the impulse.** It is already per-wave, already carries `Route`, and needs no
new id/lifecycle. "Slice-based" then means *slice/impulse state*, not cell fields —
the three-state-classes principle applied to routing (Routes ride the impulse).

### 2. The affordance registry — cells looked up, not pointered

"Fire the self affordance" becomes a **lookup by role**, not `o.self`. The being
holds a registry `role → cell` (`sensory`, `midlevel`, `other`, `self`, `motor`).
`Express` resolves the next affordance's *name* (from the transition) against the
registry and spawns it. Removing the hardcoded successor pointers is what lets the
topology be *data*.

### 3. The transition becomes data — Express is generic

Each cell's `Express` collapses to one shape:

```
Express(impulse):
    next := transition(myRole, parseConditionThisWave(impulse))   // data lookup
    if next == none { return }                                    // go/no-go
    spawn(registry[next], carryForward(impulse))                  // fire the selected affordance
```

The per-cell hardcoded `go kernel.Realize(o.self, …)` disappears. The transition
table from [dispatch-model.md](dispatch-model.md) (`affordance × parse-condition →
next`) becomes the actual data structure `transition` reads. `reframed` → `next =
midlevel`; `!planned` → `next = none`; normal → `next = successor` — all rows in the
table, no cell fields.

## What this buys

- **Scenario 4 for free (the substrate's 3/3).** Per-wave state on the impulse can't
  be clobbered by a concurrent wave — each has its own. The N-same-cell-wave race is
  gone with no locking. (The trigger coalescer is still wanted for *cognitive* dedup
  — not running N full pipelines for N rapid triggers — but that's a separate,
  higher-level concern from data-race safety.)
- **The dispatch model becomes real.** The transition tables execute instead of
  documenting; adding a route is a table row, not new spawn code in a cell.
- **Affordances as the vocabulary** ([memory-as-world.md](../../design-docs/cognition/memory-as-world.md)):
  the being's "what can fire" is the registry; routing selects among affordances by
  `(role, parse-condition)`. The inner-world primitives (match/edit/new) and the
  pipeline routing become the same dispatch shape.

## Scope / risk

This is a **real refactor**, not additive — it rewrites every cell's `Express` and
the construction wiring (registry instead of successor pointers). Do it *after* the
current working model is preserved (committed + pushed — done). Spread-before-collapse
caveat: the transition table is small (5 cells, a handful of conditions); build it as
a concrete table for these cells, not a general framework, and let the generality
emerge if a second being ever needs different wiring.

## Migration (increment order)

1. **Move per-wave state to the impulse** — `reframed`/`plannedThisWave`/verdict off
   cell fields onto `Route` (or a per-wave dispatch struct on the impulse). Cells set
   it in Observe, read in Express. Race-test scenario 4 — should go green with no
   locks. *(This alone is substrate 3/3.)*
2. **Affordance registry** — the being exposes `role → cell`; keep the hardcoded
   spawns but route through the registry (mechanical, no behavior change).
3. **Transition table** — replace the per-cell branch logic with a data lookup
   `(role, condition) → next`. `Express` becomes generic.
4. **Delete the successor pointers** — once transitions drive everything, the struct
   fields (`o.self`, `s.motor`, …) are dead; remove them.

## Open

- **Impulse vs slice for the per-wave state** — lean impulse (per-wave by nature, no
  id/lifecycle); confirm the impulse reaches Express as a mutable per-wave value in
  the kernel (today it's passed to both faces — verify identity/mutability).
- **parseConditionThisWave** — where the condition is recorded. Natural: the cell's
  Observe writes the parse-condition (match/none/proposal/exec-grade) onto the
  impulse's dispatch state; Express reads it. This is the dispatch-model's
  "parse-condition drives the transition" made literal.
- **Carry-forward vs fresh impulse** — some spawns forward the impulse (verdict-on-
  wave), some mint a fresh one (reframe, the `act` trigger). The transition needs to
  say which. Encode per-edge.
- **Whether the kernel should own the registry/transition** or the DNA. Lean DNA (the
  kernel stays mechanism; routing is cognition/policy) — the kernel still just
  delivers waves.

## Done when

- No cell holds a successor pointer; `Express` spawns whatever the transition table
  selects from the registry.
- Per-wave routing state is on the impulse, not cell fields; `TestConcurrentEngagesRapidFire`
  (scenario 4) un-skips and is race-clean with no added locks.
- Adding a route (e.g. a new reframe source, a vote branch) is a table row, not new
  spawn code.
- The live two-stroke behavior (reframe on miss, verdict-on-wave, C rule) is
  unchanged — same `cmd/validate` trace, now driven by data.
