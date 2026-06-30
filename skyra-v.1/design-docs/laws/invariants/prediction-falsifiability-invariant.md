# The Prediction-Falsifiability Invariant

**Date:** 2026-06-28

## Status

**Foundational for cognition.** Every cognitive prediction the being makes — other's next
move, self's next move, a grouping's gist read as an expectation, a setpoint's satisfied-when —
must obey this. A prediction that does not cannot be **graded**, and an ungradeable prediction
breaks the learning loop (no alignment → no closure → no setpoint → no learning).

## The Invariant

> **Every cognitive prediction must be representable as a present-tense, independently
> falsifiable world assertion.**

## What It Means

- **World assertion** — a prediction must reduce to a claim about the *world* (a state or event
  the next observation can confirm or deny), not an abstract intention or feeling. "other will
  ask about the architecture" reduces to "the other's next message is about the architecture" —
  a fact about an incoming observation.
- **Present-tense** — when the prediction resolves, it is checkable *now*, against what just
  happened. It becomes true or false in the present moment of the reafference, not in some
  unbounded, never-arriving future.
- **Independently falsifiable** — the world can prove it wrong *on its own*, without the system
  declaring it true. (**Observed, not asserted** — closure is read off external evidence, never
  claimed. See [comparator.md](../../cognition/comparator.md), [setpoints.md](../../cognition/setpoints.md).)
  A prediction you cannot be *wrong* about is not a prediction.
- **Representable as** — it need not be *phrased* this way; it must be *reducible* to one, so the
  comparator has something to grade against.

## Why It Must Hold

The grade/closure loop runs entirely on it. The comparator grades a prediction by comparing it
to what came back; a prediction that is not falsifiable against the world has **nothing to
compare against** → no alignment, no episodic closure, no crystallization into a setpoint, no
signal for either learning channel. Falsifiability is the **precondition** for every downstream
cognitive mechanism.

This is Popper applied to the being: a claim that cannot be refuted carries no information. **A
prediction that can't be wrong is a vibe, not cognition.** Enforcing falsifiability at the point
of prediction is what keeps the whole architecture from drifting into unfalsifiable
self-congratulation (the failure that lets a system "act without consequence and never learn").

## Where It Bites

- **The agents' predict step** ([../../../specs/2026-06-16/skyra-dna/agent.md](../../../specs/2026-06-16/skyra-dna/agent.md))
  — the `<prediction>` an agent emits must reduce to a checkable claim about the next
  observation, or next cycle's grade step has nothing to close.
- **A setpoint's satisfied-when** ([setpoints.md](../../cognition/setpoints.md)) — the closure
  hook must be **observable AND falsifiable** (the same constraint, one layer up).
- **The comparator** ([comparator.md](../../cognition/comparator.md)) — alignment is meaningful
  only against a falsifiable target; otherwise the −1..1 score is noise.

## Relationships

- [world-assertions-and-policy.md](../../cognition/world-assertions-and-policy.md) — the
  **constructive form**: a prediction is a present-tense S-R-O graph edge (an *expected* world
  assertion), graded by overlap with the *observed* graph; the action→observation edge is the
  gradeable hinge.
- [setpoints.md](../../cognition/setpoints.md) — observable AND falsifiable; the satisfied-when hook.
- [comparator.md](../../cognition/comparator.md) — alignment, observed-not-asserted.
- [error-recovery-loop.md](../../cognition/error-recovery-loop.md) — grading/closure depends on it.
- [perception-invariant.md](perception-invariant.md) — the reafference that *delivers* the
  evidence a falsifiable prediction is graded against.
