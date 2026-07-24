# distributed-agent-v2

**Status:** DESIGN — nothing built. Companion to [skyra-dna-v3](../skyra-dna-v3/README.md).
**Supersedes in spirit:** `design-docs/world/*` (v1 world), and folds in
`distributed systems design docs/ephemeral-world-workers.md`.

## What a distributed-agent is

> **An agent in this runtime is a series of spaces that belong to the same underlying thing — a
> bounded group of specialized spaces.**

That is the definition, and it is **not** "a thing that thinks."

```text
skyra-agent    COGNITION — nucleus, expression graph, grades.   it MOVES.
distributed-agent    SPACES    — topology, state, affordances.        it IS MOVED OVER.
device-agent   SPACES    — a bounded group of device spaces.    it IS MOUNTED.
```

There is not one agent. There are many. **Right now exactly one of them has cognition.**

The skyra-agent moves across the world and mounts other distributed-agents. World-agents are strictly
spaces that cognitive agents move over, do work on, and mount.

## The one law that keeps this honest

> **The world REPORTS. It does not GRADE.**

A world worker returns an `Outcome` — `succeeded / blocked / denied / stale / unavailable /
cancelled / partial / indeterminate`. That is a statement about **what happened**. It carries no
opinion about whether that was good, and **it has never seen the agent's prediction.**

```text
cognition makes the bet     (private. the world never sees it.)
world says what happened    (independent. no stake.)
cognition diffs             (the grade.)
```

**And that is exactly why the world is trustworthy.** It has no stake in the bet. It cannot be
flattered and it cannot be gamed, *because it does not know what answer was hoped for.* An LLM
judge — which sees the prediction it is grading — can never have that property. **Independence is
the entire value.**

Corollary: `denied` is **a peer agent refusing**, not a permission check the skyra-agent runs on
itself. Same for `blocked` and `unavailable`. Those are another agent's answers, which is what makes
a mount a *crossing* rather than an API call.

## The world holds two different kinds of thing

This is the central v2 addition, and getting it wrong collapses the grounding story.

```text
FACTS    what is TRUE.                          file exists. door is open. process running.
         → provenance: raw.  MAY GRADE A BET.

MARKS    what agents have WRITTEN.              traces, target state, notes.
         → self-authored.   MAY NEVER GRADE A BET (about anything but itself).
```

The world is not only *what is true* — it is also **a place agents can write to**. A whiteboard in
the room. Anyone can read what is on it; it is not thereby a fact about the room.

### The mark rule

Without this, an agent writes a trace, walks away, comes back, perceives its own handwriting, and
treats it as evidence about the world. That is **laundering** — the fiction problem wearing a lab
coat. LAW 1 does not catch it (a mark is not an expression) and LAW 2 does not catch it (it *did*
come back through perception).

The rule is narrow and exact:

```text
LEGITIMATE   "a wave of mine passed here at tick N"
             → TRUE. a fact about my own history. perceiving it is MEMORY, not fiction.

FORBIDDEN    "there is a trace toward the kitchen, therefore food is in the kitchen"
             → NO. the trace says I WENT THAT WAY. it says nothing about food.
```

> **LAW — A mark may be perceived as a fact about itself. It may never be evidence for a claim
> about anything else.**

A mark can grade a claim *about marks*. It can steer where you look. **It can never settle what you
claimed about the world.**

### Marks decay on their author's clock

A trace records *"I searched here."* That is a fact about the **agent's history**, not the world's.
Its relevance decays with **that agent's** experience.

```text
AGENT TICK    subjective. advances on COGNITION.   → bet maturity (`within`), MARK DECAY
WORLD TICK    objective.  advances on THE WORLD.   → world-process timing
```

If the being has not thought, its search has not gone stale — a search is a thing *it did*, not a
thing that *happened*. And staleness protection does not come from decay anyway: **it comes from
perception.** If the being needed to know, it looked. If it did not look, it did not need to know.

So: **marks decay on the tick of the agent that laid them.** In a shared world, another agent's
marks decay on *that* agent's tick. The world needs no clock for this at all.

## The return of a wave is the world's report — not the agent's output

Worth stating because it is easy to get backwards.

```text
DID THE WORLD AGREE?   "did you do the thing you said you did"
                       → returns ON THE WAVE. an Outcome. RAW. It crossed the boundary.
                       → failure means the ACTION failed. the message did not send.

DID THE AGENT AGREE?   "did the other agent respond the way you predicted"
                       → returns as a NEW IMPULSE. interpreted.
                       → failure means the MODEL OF THEM failed. Michael did not want lunch.
```

Same code path, different edges, **completely different lessons**. One act produces **two expects**
(see [expression-grammar.md](../skyra-dna-v3/expression-grammar.md)).

## The pull model — the space holds impulses

The space **holds impulses until an agent perceives them.** They are not pushed.

This is consistent with the cycle: *you only observe what you acted to observe.* It makes perceive a
real, chosen, costly act rather than an interrupt, and it is why the observation log is already a
fact about the being (see [the-cycle.md](../skyra-dna-v3/the-cycle.md)).

**And the being does not go deaf**, because attention is not a mechanism — **attention is a goal.**
The genome bootstraps a standing target that requires perceiving, so the being hears Michael not
because an interrupt fired, but because *knowing what Michael is doing is something it wants*, and
the plan to get it includes looking. No push channel. No peripheral-vision hack. Just a bridge that
never closes.

## The parts

| doc | what it settles |
|---|---|
| [traces.md](traces.md) | stigmergy: the potential field, materialized on the world. **NOT BUILT.** |

Still to write (mostly by lifting from `distributed systems design docs/`):

- **the invocation/outcome contract** — `ephemeral-world-workers.md` is already right; move it here
- **the mount** — the membrane as the agent↔world seat; relocation and projection as one primitive
- **device-agents** — a device is a bounded group of spaces; the genome already had this in v.05
- **perceive scope** — what a perceive is authoritative *about*

## Open

- Does a distributed-agent ever get cognition? Everything written assumes **no** — the correspondence maps
  DNA → *place capability registry*, not DNA → DNA. If that changes, a mount becomes two learning
  beings meeting rather than one being visiting a place that answers, and much of this is wrong.
- Shared worlds: if two cognitive agents leave marks in the same space, whose tick governs whose
  decay is answered above — but **who may read whose marks** is not.
