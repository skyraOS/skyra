# Traces — Stigmergy

**Status:** DESIGN DIRECTION. **NOT BUILT, and not to be built yet.**
Part of [world-agent-v2](README.md).

> **Do not build this until the being can be wrong about something.** It is a whole subsystem,
> downstream of a core loop that does not exist. Written down so it is not lost, and so the hazard
> below is recorded before anyone is tempted.

## The idea

Each wave leaves a **trace on the space it is operating over.**

```text
the world holds a representation of the TARGET STATE
    (because the agent gave it the target state)
              ↓
the wave STARTS AT THE TARGET and walks BACK toward the origin on OBSERVE
              ↓
OBSERVE leaves traces
              ↓
EXPRESS decays those traces, in the reverse direction, as the tick runs backward
              ↓
a decaying gradient, laid down by descent and worn away by ascent
```

And **perceive is fully allowed to actually perceive** — it is not forced to. It reads the traces
where it stands.

## What it actually is

This is **stigmergy** — pheromone trails, ant colony optimization. Agents coordinate by modifying a
shared environment rather than by messaging each other.

And notice what it *is* in this system's own terms:

> **The potential field from `scratchpad/2026-07-13/README.md` — materialized on the world instead
> of computed in cognition.**

The field math that got parked is not dead. It **relocated**. Which buys real things:

- **The world does the computation.** The agent does not hold the field in memory — it reads the
  local gradient where it stands.
- **Decay is free.** It is a function of tick, not of thought.
- **It is naturally distributed.** On a device-world, the trace lives on the device.
- **Descent lays the bridge down; the trace IS the bridge.**

## The hazard — read this before building it

> **The agent writes a trace onto the world. Later, the agent perceives the trace. Its own output has
> become "observation."**

That is **laundering** — the being writing something down, walking away, coming back, and treating
its own handwriting as evidence about the world. `LAW 1` does not catch it (a trace is not an
expression) and `LAW 2` does not catch it (it *did* come back through perception).

Same hazard for putting the target state in the world: that is a **projection** (mine) sitting in the
**observation channel** (theirs).

### The fix, and it is narrow

Ants already solved this: **a pheromone is not a fact about the world. It is a record of traffic.**

```text
LEGITIMATE   "a wave of mine passed here at tick N"
             → TRUE. a fact about my own history. perceiving it is memory, not fiction.

FORBIDDEN    "there is a trace toward the kitchen, therefore food is in the kitchen"
             → NO. the trace says I WENT THAT WAY. it says nothing about food.
```

> **A mark may be perceived as a fact about itself. It may never be evidence for a claim about
> anything else.**

A trace can grade a claim *about traces*. It can **steer where you look**. It can **never settle what
you claimed about the world**. It is a **search aid**, not **evidence**.

Mechanically: traces are `ProvenanceMark`, and a mark's `Fidelity()` toward any non-mark claim is
**zero**.

## The clock

Marks decay on **the tick of the agent that laid them** — not the world's.

A trace records *"I searched here."* That is a fact about the agent's history, and its relevance
decays with the agent's experience. If the being has not thought, its search has not gone stale.

And staleness protection does not come from decay at all — **it comes from perception.** If the being
needed to know, it looked.

*(This reverses the original sketch, which said world-tick. The mark/fact distinction is what settles
it: a mark belongs to its author, so it is measured on its author's clock.)*

## Open

- Two cognitive agents in one space: who may read whose marks? Reading another agent's trace is
  reading a true fact about *that agent's* history — legitimate, and potentially very useful (this is
  how a colony coordinates). But it is also how one agent could **poison another's search** by laying
  misleading traces. Adversarial stigmergy is a real failure mode and there is no answer here yet.
- Does the trace carry *what the wave was for*, or only that it passed? The first is far more useful
  and far more dangerous.
- Interaction with the reverse index. Both are "how have I gotten here before." One is in the
  expression graph, the other is smeared on the world. Are they the same mechanism at two
  granularities, or genuinely different? **Suspect the former** — and if so, one of them should not
  exist.
