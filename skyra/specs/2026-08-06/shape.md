# The Shape

**Date:** 2026-08-06
**Status: SETTLED.** Merges `2026-08-03/graph.md` and `2026-08-03/families.md`. The first led
with a rule that is now false; the second was revised twice. This is what survived both.

> **One family of state. Composition decides kind. EXPECT is the only writer.**

---

## Three structural states, no type tag

```text
entries beneath, no Act    a COMPOSITE     an EXPECTATION   (a directory)
no entries, an Act         a LEAF          an ACTION        (a file)
no entries, no Act         a LEAF          a RAW IMPULSE    (what arrived)
```

Unix has one inode and decides by what it contains. Same here, and one better: **not even a mode
bit is stored.** Kind is read off composition, so it cannot disagree with the structure.

This is `2026-07-27/expectations.md`'s *"the typing is derived, not declared"* arriving from the
other end — and the collapse that doc predicted, which it deferred for one reason:

> *children of a node must run sequentially when they are actions and concurrently when they are
> expectations… Deferred, not forgotten.*

That condition came back as **the parent's ordering policy** — sequential or concurrent, declared
by the parent — which is the same doc's *"a flag the parent declares"* and what answers kernel
§13 Q3.

### What falls out

**No empty directories.** Every expectation carries its own perceive, and a perceive is an action,
so every directory holds at least one file.

**Directories are never executed.** A directory's content is its entries. An expectation acts only
through the actions it contains.

**The stack anchors only at directories.** You cannot `cd` into a file. An action is something you
run from where you are standing, not somewhere you can stand.

**`Q(E,A)` has an edge.** The directory entry *is* the expectation → action edge, so efficacy has
a home without actions becoming a second node type.

**A file never becomes a directory.** Composition-derived typing would otherwise let an entry
under an action silently promote it. The runtime refuses the growth: an action that needs
decomposing was never an action.

## Every expectation carries its own perceive

> **An expectation cannot be constructed without one.**

`2026-07-27/expectations.md`'s falsifiability invariant, promoted from a check to a constitutive
rule — the same move the kernel made. Do not enforce a rule; remove the ability to break it. There
is no unfalsifiable expectation to detect because there is no way to make one.

### Graded, not settled

Vocabulary, and load-bearing. The old model **settled** — a bet resolved once, terminally. This one
**grades**, continuously; the evidence set is append-only and nothing reaches a final state.

```text
SETTLE   terminal. the record closes.        belongs to the old model.
GRADE    continuous. the record never closes.
```

## EXPECT is the only writer of state

The interpretation node is **gone**. A stored reading is a claim about the world, and a claim about
the world is state — so storing one made the model a writer of state.

`dnav3/observation.go` had already found this and could only mark it:

> INTERPRETED is **a model output sitting on the WORLD'S side of the boundary, wearing the world's
> authority.** LAW 1 does not catch it, because it is not an expression.

`Fidelity()` returning `0.3` was a discount on the problem. Removing the node is the fix.

```text
EXPECT      writes NODES.   structure. the only writer of state.
INTERPRET   writes EDGES.   grades. it may not create anything.
```

Checkable rather than remembered: anything calling `AddExpectation` is the expectation worker, and
nothing else may.

**What the model may produce is two things, neither an assertion:** a scalar on an edge, or a
proposal that becomes an expectation — with a perceive. So a reading like *"michael wants water"*
cannot be filed as a fact. **Every claim about the world is falsifiable by construction**, because
creating one is the only way to state one.

## Two edge kinds

```text
SUPPORT     [-1,1]   impulse → expectation.        ASSIGNED by grading. signed, because
                                                   evidence can bear against.
CONFIDENCE  [0..1]   child → parent expectation.   AUTHORED by the proposer: "I am x% sure
                                                   this child completes this parent."
```

Different writers, different ranges, different jobs. INTERPRET writes only the first, EXPECT only
the second.

`T`, `G`, and `P` fold over **support** alone, so confidence has no path into the arithmetic —
which is `2026-07-27/trust.md`'s *confidence routes, never weights* made structural rather than
remembered.

### Direction: store down, fold up

`2026-07-27/FLAGS.md` flagged two arrow conventions as an unresolved contradiction. They are not
two conventions; they are two operations:

```text
STORE   the COMPOSITE points at its CONSTITUENTS, and owns those edges.
        "an expectation owns every edge it points at, and never one that points at it."

FOLD    contributions run CONSTITUENT → COMPOSITE, on ascent. nothing is stored this way.
```

The kernel already performs both — `Perturb` descends the stored edges, `Observe` folds back up
them. Drawing them as one arrow is what made the direction feel undecidable.

## Append-only

```text
ACTION EDGE     write-once
SUPPORT VALUE   immutable once written
EVIDENCE SET    append-only
```

Nothing is removed or rewritten, which is what lets an execution walk the graph while learning
happens: a stack cannot be standing on something that ceased to exist. It is also why a plan is
never *ejected* — a better one is added beside it and activation moves.

---

## What killing the interpretation node cost

Two debts, both worth the writer rule, both recorded so they are not rediscovered.

**Toy 2 came back to `T`.** With an impulse plane, ten copied records inflated one interpretation's
fan-in and never touched the fold. Now they are ten support edges, and *"distinct is not
necessarily independent"* is a defect in the grounding arithmetic again — back where
`2026-07-27/trust.md` left it.

**Re-scoring lost its home, and got one back.** `2026-07-27/workers.md` needs an impulse
re-presented once the field grows a target for it, or it sits at ≈0 forever. A second interpretation
node was a clean place for a second reading. Now a second reading is a second support edge between
the same pair — which `O(E)` being a *set of distinct observations* forbids, since it is the same
impulse counted twice.

> **Store two edges. Fold one.**

Both readings persist; support is immutable, so neither is rewritten and nothing is destroyed. But
only one enters `T`/`G`, because two edges from one impulse is Toy 2's inflated fan-in arriving
again — and this doc already records that Toy 2 came back.

```text
THE FIRST EDGE    evidence about the WORLD.    folds.
THE SECOND EDGE   a record about the GRADER.   does not fold.
```

The divergence between them is the learning signal: it is the only place the system can see its own
reading of one thing change. Discarding either reading destroys it, which is why *newest wins* and
*first is superseded* were both worse.

```text
STILL OPEN   which edge folds when there are more than two, and whether "first" is the
             right choice or merely the simplest.
```

## The caveat that keeps confidence safe

Confidence is a whole edge species with no blind assessor — the exemption `FLAGS.md` names,
enlarged. It stays tolerable **only because support edges grade it**: authored confidence versus
assigned support is `actions.md`'s calibration signal, and both numbers now sit on the same
expectation, so the comparison is a local read.

> **If calibration is never computed, confidence edges are unfalsifiable decoration — a number the
> model writes that nothing ever checks.**

The separation is not what makes this safe. What makes it safe is that the separation makes the
comparison possible.
