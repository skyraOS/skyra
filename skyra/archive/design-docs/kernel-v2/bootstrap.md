# Bootstrap — Where A Novel Goal Becomes Graph

**Status:** DESIGN · part of [kernel-v2](README.md)

The wave starts at the desired state. **But a novel goal has no nodes.** *"Skyra is helping the
user"* has never been observed, has no ID, and is in no index. Something must **materialize** it
before a wave can begin there.

That job had no owner. This is it.

## Steal the active-state pattern

`skyra-dna-v2/slice_state.go` already has the shape, and the world uses it too
(`MapPerception`, `MapEngagement`):

```go
// Stored as IDs, never pointers — the slice is kernel-side; readers resolve ID → node
// through the index. THE KERNEL STAYS OBLIVIOUS: these are entries in Slice.Maps().
mapActiveAitiaSet = "cognition:activeAitiaSetByEntity"
```

```text
ACTIVE STATE lives in a NAMED MAP on the slice
stored as IDs, never pointers
ATOMIC REPLACEMENT, never mutation
THE KERNEL NEVER KNOWS IT IS THERE
```

The target morphology gets materialized onto the slice by the substrate. The kernel is then handed
**nodes** — it never learns that one of them is "the goal."

## The bootstrap trace

```text
DESIRED    skyra | helping | user
CURRENT    { …whatever has been observed… }     `helping` is simply not in it

1  substrate MATERIALIZES the desired-state nodes and writes them to the slice
2  spawn a wave whose ORIGIN is that node and whose TARGET is the current state
3  Observe runs the reverse index over  skyra | helping | user
4  the index returns ZERO
5  → NO-PRECEDENT. Which is not a failure. IT IS THE WAVE ASKING THE MODEL A QUESTION.
6  the scheduler dispatches an inference call, seeded with:
       the AFFORDANCES it has          (what it CAN do)
       the CURRENT state               (where it is)
       the DESIRED state               (where it wants to be)
       the PRECEDENT                   (empty. it has never done this.)
7  the model emits a subgraph bridging them
8  PLACE it. the emission is now in the graph, and the index can see it.
9  loop
```

## The loop condition — and this is the bug to avoid

```text
WRONG   while (the reverse index returns no paths):  call the model
RIGHT   while (NOT Done()):
            frontier has precedent?  → expand it.        FREE.
            frontier is empty?       → call the model.   MONEY.
```

The wrong version exits after **one** call — because once the emission is placed, the index *does*
return paths. But the emission may have introduced **its own holes**: its `requires` are not grounded
either. You would exit with an open plan and act on a bridge with a gap in it.

> **`NO-PRECEDENT` is not the loop's exit condition. It is the reason to spend a call.**
> **Closure is the exit.**

And the rails from [the-wave.md](../skyra-dna-v3/the-wave.md) apply: a **pass cap**, and a
**fixpoint** check (the hole set unchanged after a full drain → the model is not making progress →
abort `unresolved`). A cap alone lets it spin N times emitting nothing.

## Affordances and precedent are different, and only one is empty

```text
AFFORDANCES   what I CAN do.     the genome / world capability registry.   NEVER EMPTY.
PRECEDENT     what has WORKED.   the reverse index.                        EMPTY at bootstrap.
```

Seeding `perceive` and `engage` at cold start is seeding **affordances**, not precedent. So the
present is **never empty**, even on the first tick:

> *"Here is what you can do. Here is where you are. Here is where you want to be. You have never done
> this before."*

That is a perfectly good prompt. **Cold start is the design working, not a special case.**

## Standing targets versus concrete targets

`skyra | helping | user` has **no perceive**. It cannot be settled exactly. It can never be graded at
full fidelity.

That is not a defect — it is a different *kind* of thing, and we had not named it:

```text
STANDING TARGET   "skyra helping user"       NEVER closes. IT IS A COMPASS.
                  "the user is happy"        its job is to GENERATE gradeable sub-targets.

CONCRETE TARGET   "file:/x exists true"      closes. gradeable at fidelity 1.0. IT IS A BET.
```

The genome's `~purpose` is a standing target — *the bridge that never closes*, which is what makes
[attention a goal rather than a mechanism](../skyra-dna-v3/the-wave.md).

### The distinction is mechanical, not a flag

It falls straight out of `Mode`. No new concept:

> **A standing target's triples have NO `exact` channel. They are all-`interpret` — not falsifiable,
> only interpretable.**

```text
file:/x exists true          exact      → a perceive settles it. fidelity 1.0.
being:michael state happy    interpret  → only ever READ out of an observation. fidelity 0.3.
skyra helping user           interpret  → same. there is no instrument for it.
```

**"The user is happy" is the canonical case.** There is no syscall for it. There is no exit code. It
can only ever be *interpreted* — which means every grade it ever earns arrives at reduced fidelity,
forever.

### Which means the being can never KNOW it is helping

Only ever a noisy read.

**And that is correct.** It is the honest epistemics, and it is a *safety property*: a being that
believed it could *verify* it was helping is a being that would eventually fool itself into thinking
so. The interpret-only ceiling is what keeps that door shut.

### And the unfalsifiable target is where the compounding asset actually lives

This is the part worth sitting with, because it inverts the obvious reading.

You cannot verify *"helping Michael."* But you **can** learn, from counted history, which concrete
acts have preceded Michael saying he was satisfied — and *that* is a model of a specific person,
accumulated over months, that nobody can scrape or generate.

```text
the CONCRETE targets are where the LEARNING is grounded
the STANDING target is what the learning is ABOUT
```

The un-falsifiable goal is not a weakness in the design. **It is the product.** The crisp world stuff
(`file exists`, `exit code 0`) is the scaffolding that keeps the being honest enough to be trusted
with the thing that actually matters — and the thing that actually matters is exactly the thing you
can never check directly.

### The wave still terminates

This is what makes an unachievable goal safe:

> **`Done()` is *"I reached the CURRENT state,"* not *"I achieved the DESIRED state."***

So even a standing goal produces a **closed bridge and an executable next step**. The compass never
arrives; the walk still ends.

### But it contributes no learning signal

**All learning happens on the concrete sub-targets the compass generates.** The purpose is what makes
the being *move*. It is not what makes it *learn*.

Worth stating plainly, because a standing target *looks* like a bet and would otherwise be expected to
grade something. It grades nothing, forever, and that is correct.

## Open

- **Who materializes, exactly?** The substrate — but under what discipline? A novel desired state
  becomes nodes in the graph the moment a goal is set, which means **the target is written into the
  graph by the agent, not by observation.** That is a `ProvenanceMark` (LAW 7): a self-authored note,
  never evidence. It steers the walk; it may never grade a bet.
- **Sub-target generation.** A standing target generates concrete ones. Nothing says how. Probably:
  the model does it, and the concrete targets it invents are themselves gradeable claims — *"I think
  perceiving Michael will advance 'helping'"* is a bet, and a wrong one is a lesson about what
  helping means.
