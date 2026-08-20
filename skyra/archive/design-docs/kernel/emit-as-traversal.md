# Emit as Traversal — the Motor Gathers World Cells as its Children

**Status:** Parked direction (sidebar — come back to this). Not a build step yet. It
revises the current motor-emit mechanism in
[slice-based-dispatch.md](../../specs/2026-07-03/dispatch-model.md) and is the concrete
realization of [world-kernel-coupling.md](world-kernel-coupling.md)'s "the world is a
passive Reality, traversable like anything else" and
[operator-runtime-and-interrupts.md](../../archive/design-docs/operator-runtime-and-interrupts.md)'s
"shared substrate — one law across memory, devices, places, world."

## The claim

**The motor ephemeral unit is not special.** Every ephemeral unit does the same thing:
it gathers `Children()` and hands them to the kernel; the kernel traverses them. The
motor should do *exactly that* — the only difference is **which children it gathers**.

For the motor, the children are **the cells of the world** — the point Skyra stands on,
the target agent, the surrounding points. The kernel then does **raw traversal** into
them, and that traversal **is the emit**:

- reaching the target agent's cell **is** the engage,
- sweeping the surrounding point-cells **is** the perceive,
- entering the destination point-cell **is** the move.

No `Emit` callback. No bespoke world-primitive dispatch. The same `Children() → kernel
traverses the survivors` contract the rest of the system already runs on.

## The shape stays uniform

```
membrane (orchestrator, reading live state off the slices)
    └─ ephemeral unit  (the worker for this step)
         └─ raw traversal  (kernel walks the unit's Children())
```

This is the identical bracket for a cognition step and a motor step. A cognition unit's
children are cognition cells (or none — inert). A motor unit's children are **world
cells**. The kernel does not know or care which — it scores children, gates by
threshold, descends survivors. Emit, perceive, and act collapse into the one traversal,
differentiated only by which world cells the unit gathers under its framing.

## What this replaces (the current mechanism)

Today the motor breaks the uniformity:

```
motor:collapse → DNA.Emit(imp) → (callback) → realm.Observe(being, imp)
    → world spins up its OWN kernel: w.kernel.Realize(...)   // a separate, re-rooted wave
```

So action runs on a *different* path than thought (a callback + a re-root into a second
kernel) instead of on traversal. That is the "two mechanisms, two kernels" seam this note
removes: motor stops calling out and instead **exposes world cells as its children**, and
the wave simply continues into them.

## The consequence: it points at one kernel

If the motor hands world cells to the **being's** kernel, a single wave crosses
interior→world under the uniform `Children()` contract — which is exactly what makes the
**second kernel unnecessary**. The actor-boundary crossing (goroutine spawn — see
[activation-and-traversal.md](activation-and-traversal.md#two-recursions)) is still
available as a *placement* choice when the world is treated as a separate concurrent
actor, but it is the **same traversal law**, reached the same way (a child whose content
is another actor), never a side-channel `Emit`.

> Registered, not scheduled: this note establishes *why* the two-kernel split is
> incidental (it exists only because emit is a callback today, not because traversal
> requires it). Collapsing it is downstream work, not this note.

## Why it composes with the motor changes

[prediction-motor-and-closure.md](../../archive/design-docs/prediction-motor-and-closure.md) has
motor operating **inside a bound** — continuing locally over a coherent slice of the
world rather than re-running the full brain per act. If the motor's children *are* world
cells, the **bound is literally the world-cell neighborhood the unit is licensed to
traverse**. "Motor operates inside the bound" becomes "the unit traverses that
neighborhood"; a surprise that breaks the bound is a world cell that grades out of the
frame. The bound and the traversal are the same object seen from two sides — which is why
this is cleaner than what we have, and why it lands *with* the motor changes rather than
against them.

## What does NOT change

- **The world keeps its affordance semantics.** What "reaching michael's cell" *does*
  (deliver an utterance, block if out of range, teleport at the maze exit) still lives
  behind the world boundary. This note changes the **entry** into those affordances
  (traversal of gathered children) not their meaning.
- **The perception invariant holds.** The traversal delivers its result to origin as
  always; a perceive's surroundings return to the being. (How the being *internalizes*
  that return is the separate perceive-internalization spec, next.)
- **The membrane is still the orchestrator** reading live state off the namespace slices
  to decide which unit runs and what children it gathers.

## To reconcile when we pick this up

- **Which world cells the motor gathers, and how** — the target agent + a radius of
  points is the obvious set, but "gather the relevant world cells under the current
  framing" is the same relevance question the cognition units already answer by
  activation. Likely: the world cells become children with activation records, scored by
  the same bump, so "which of the world is in reach" falls out of traversal rather than a
  hardcoded radius.
- **In-plane vs. crossing** — is the world a sub-structure of this wave (one kernel,
  sequential) or a separate actor (spawn)? This is the placement decision from the Two
  Recursions, and it is *where the second-kernel question actually gets decided*.
- **Return/reafference** — the perceive return currently dies at the bridge and the
  sensory stub; unifying emit with traversal does not by itself fix internalization (see
  the perceive-internalization spec).
