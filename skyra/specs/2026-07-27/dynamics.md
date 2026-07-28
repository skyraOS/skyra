# Runtime Dynamics

**Date:** 2026-07-27

> **Dynamics is not a third predictor. It is a controller.**

Global control values — stress, context depth, attention breadth, exploration — that the rest of
the loop keys on. They are functions of the **measured state transition**, not of any single
expectation or action, and they shape how the next loop executes.

```text
interpretation   LEARNS     observation → expectation
policy           LEARNS     expectation → action
dynamics         CONTROLS   derived from the transition, modulates the next pass
```

Two learners and one controller. Dynamics needs no grading arm of its own, because it never
claims anything.

---

## The slot already exists

`spacetime/space.go`:

> *"stress (the being's global neuromodulator) is deliberately NOT here. The kernel is mechanism
> and must not know about cognition signals. Stress is DNA state — held on the being's substrate
> and read by cells via a capability assertion, never through this kernel contract."*

and `world/slice.go` carries `stress float64 // global stress (neuromodulator); mechanics wired
later`. The hole was cut and left empty. This fills it.

## Why transitions, and not levels

Transition characteristics are diagnostics of **what cognition is doing**, never of whether it
is right. That makes them the wrong thing to grade on and the right thing to steer on — which
is exactly the role they are licensed for:

> *it may inform when-to-think; it may not move an edge.*

## Inputs — all of them already defined

Nothing new has to be measured. The controller reads `T`, `G`, `|O|`, coherence, and their
deltas.

```text
‖ΔP‖ large              surprise            → widen attention, deepen context
BORN count high         novelty             → raise exploration
coherence falling       contested           → look more before acting
‖ΔP‖ shrinking          converging          → lower exploration, commit
T rising, P flat        consolidating       → cheaper loop, less looking
```

## The trap, and the fix

A small `ΔP` is **converged** or **disengaged**, and movement alone cannot tell them apart. A
controller that reads "change is small → reduce exploration" would go quiet exactly when the
being had stopped engaging — backwards, and self-reinforcing.

Trust disambiguates it:

```text
converged     small ΔP · HIGH trust · high coherence
disengaged    small ΔP · LOW trust
```

Which is why [trust](trust.md) is unsigned mass rather than signed position. A controller
reading only direction cannot make this distinction.

---

## Two requirements

**Bound it, and make it slow.** Control values derived from a transition and modulating the next
transition form a feedback loop. Unbounded, it runs away:

```text
narrow attention → fewer observations → higher stress → narrower attention
```

Homeostatic bounds, and a time constant slower than the loop so it tracks the trend rather than
chasing noise.

**Log every control value, per wave.** Dynamics modulate what surfaces, and anything touching
retrieval breaks replay unless its inputs are in the record. Without the log there is no
answering *why did this run go that way* — the control values are the answer and they are gone
by the time you ask.
