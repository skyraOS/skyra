# Differentiation — structure from overload

**Date:** 2026-08-06
**Status: WORKING TOWARD.** Not built, and two things it depends on do not exist yet. Written
now because the shape is finally mechanical rather than decorative.

> Entered `2026-08-06` intact. Still not built, and still honest about it.

> **An expectation splits when it no longer fits in the frame it was allocated.**

---

## Why this one is not theatre

The version that would be: *"when a node gets big, split it"* — a tuned constant, an arbitrary
partition, more structure and no more information.

Two things make this different.

**The threshold is not a number. It is the budget.** A node carries an observation token count,
and the frame it can be assembled into is already allocated by the descent —
`child_budget = (parent_budget − cost) × weight`, from the same arithmetic
[behaviors.md](behaviors.md) uses. Differentiation fires when:

```text
cost > allocated budget
```

No constant to tune, and the trigger moves correctly on its own if the context window changes.

**The payoff is sharper retrieval, not tidier structure.** After a split, an impulse activates one
child rather than the whole accumulation, so the assembled frame gets *more relevant* — not merely
smaller.

> **Differentiation is how the graph stays assemblable as it grows.**

Without it, a long-lived expectation eventually cannot be brought into any frame at all, and the
being goes blind to its most-used knowledge. That is a failure that arrives *because* things went
well, which is the kind worth building a mechanism for.

## It is EXPECT's second trigger

Same worker, and both triggers are comparisons:

```text
children closed, self not closed     →  EXPECT   replan
observation load exceeds budget      →  EXPECT   DIFFERENTIATE
```

Mechanical trigger, boundary work — *what are the two things here* is a partition and needs
judgement. The same shape as everything else: the exact pass decides *that* it should happen, the
model decides *what*.

## Append-only makes the split additive

Support is immutable and edges are write-once, so history cannot be redistributed:

```text
OLD evidence   stays attached to the parent. nothing moves.
NEW evidence   lands on the children.
```

**This is what stops it re-splitting forever.** A naive version reads current load, finds the old
evidence still there, and fires again every turn. Here the fix falls out of append-only:

> **A differentiated expectation stops being a direct attachment point.**

Its load is frozen at the moment of the split, so the trigger cannot re-fire on it. New grounding
has to land on a child.

### And the design already has a name for the result

[shape.md](shape.md):

> *an expectation with no actions of its own is grounded entirely through its children — strict
> inheritance.*

Differentiation is the **transition** that turns a directly-graded expectation into an inheriting
one. Nothing new is introduced; a state the model already describes acquires the event that
produces it.

## The risk, and the measurement it needs

**A bad partition degrades retrieval silently.** Split the wrong way and the evidence for one
thing scatters across two children, so neither activates strongly and the frame gets *worse* —
with nothing failing. Nothing closes wrong. Nothing errors.

The signal, logged from the first split:

```text
ACTIVATION SPREAD ACROSS SIBLINGS

  concentrated   one child lights up          → the split found a real seam
  flat           both light weakly            → the split cut through something
```

Cheap, and **not recoverable after the fact** — same category as the goal-degeneration detectors
in `2026-07-27/workers.md`, and it should be logged for the same reason.

## What has to exist first

```text
COST         no Node carries a token count. the trigger has nothing to compare, and this is
             also what ASSEMBLE is waiting on. one mechanism unblocks both.

THE BOUNDARY the partition is a model call, and nothing spends the outbox yet.
```

Both are already the critical path for other reasons, so differentiation adds no new dependency —
it becomes available roughly when assembly does.

---

## Open

```text
TWO CHILDREN OR N     "splits into two" is the cell-division framing, but the partition is a
                      clustering and clusterings are not naturally binary. forcing two means
                      repeated splits to reach three — workable, but the tree shape then
                      records split HISTORY rather than structure.

THE PARENT'S PERCEIVE  every expectation has one, so a differentiated parent keeps its own
                      even though it is now grounded through children. whether that is
                      redundancy or a genuinely different question is undecided — and it
                      decides whether a parent can still be wrong on its own terms.

WHAT PARTITIONS       the model is asked to split, but on what basis. dnav3 has machinery
                      that clusters observations (density_pockets.go, segmentation, classes)
                      built against the old model — worth reading before inventing.

DEPTH AND ATTENUATION a heavily-differentiated region gets deep, and depth attenuates
                      contribution. so the more a region is used, the quieter each part of it
                      becomes. that may be correct — or it may punish exactly the knowledge
                      the being uses most.
```
