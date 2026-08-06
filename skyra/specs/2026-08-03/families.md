# Families — one family of state, two edge kinds

**Date:** 2026-08-05
**Status: SPEC. REVISED 08-06** — the interpretation node is gone. See *What changed* below.

> **The expectation family is the only state. Impulses are input; grades are edges.**

---

## What changed — 08-06

The interpretation node is **killed.** An interpretation is a claim about the world, and a claim
about the world is state — so storing one made the model a writer of state.

`dnav3/observation.go` had already found this and could only mark it:

> INTERPRETED is **a model output sitting on the WORLD'S side of the boundary, wearing the
> world's authority.** LAW 1 does not catch it, because it is not an expression. It is the one
> place model output crosses into the observation graph.

`Fidelity()` returning `0.3` was a discount on the problem, not a fix. Removing the node is the
fix.

> **THE EXPECTATION WORKER IS THE ONLY WRITER OF STATE.**

What the model may now produce is exactly two things, and neither is an assertion:

```text
A SCALAR ON AN EDGE      "this bears −0.7 on that."   a grade, not a claim.
A PROPOSAL               which EXPECT turns into an EXPECTATION — with its perceive.
```

So a reading like *"michael wants water"* cannot be filed as a fact. It can only become an
expectation, and expectations are unconstructible without a perceive. **Every claim about the
world is now falsifiable by construction**, because creating one is the only way to state one.

## The shape

```text
Raw Impulse ──support [-1,1]──▶  parent expectation
                                          │ confidence [0..1]
Raw Impulse ──support [-1,1]──▶           ▼
                                     expectation
                                          │ confidence [0..1]
                                          ▼
                                       action
```

## One family of state, and one kind of input

```text
STATE    expectation | action     durable. differentiated by composition. EXPECT writes it.
INPUT    raw impulse              what arrived. not state. graded, never asserted.
```

The expectation family still differentiates by composition, unchanged:

```text
has entries beneath it   → EXPECTATION   (a directory)
has none                 → ACTION        (a file, a leaf)
```

The impulse family no longer exists as a family — killing the composite left only leaves. A raw
impulse is what arrived; it is cited by the edges it grounds and is not a node in the tree.

## Who writes what

```text
EXPECT     writes NODES.   structure. the only writer of state.
INTERPRET   writes EDGES.   grades. it may not create anything.
```

That division is the whole rule, and it is checkable: a behavior that calls `AddExpectation` is
the expectation worker, and nothing else may.

## Two edge kinds, and the split is what protects the arithmetic

```text
SUPPORT     [-1,1]   impulse → expectation.        ASSIGNED by grading. signed, because
                                                   evidence can bear against.
CONFIDENCE  [0..1]   child → parent expectation.   AUTHORED by the proposer: "I am x% sure
                                                   this child completes this parent."
```

Different writers, different ranges, different jobs — and INTERPRET may write only the first
while EXPECT writes only the second.

This is what `2026-07-27/trust.md` spends a section forbidding, made impossible instead:

> `c(o,E) = relationship × confidence × …` **DO NOT.** Weighting `T` by confidence would
> systematically shrink the mass of unfamiliar contradictory observations.

`T`, `G`, and `P` fold over support edges. Confidence lives on a different edge species and
**there is no path by which it can enter that fold**. Not a rule anyone remembers — a different
edge set. Same move as the kernel having no method that would let it classify.

## Direction: store down, fold up

`2026-07-27/FLAGS.md` flags exactly this as unresolved — *"two arrow conventions, and a
self-contradicting citation"* — with `README.md` pointing an expectation at its evidence and
`trust.md` pointing observations at the expectation. The flag says one convention must be chosen
and the other named explicitly.

Both are right. They are different operations:

```text
STORE   the COMPOSITE points at its CONSTITUENTS, and owns those edges.
        parent expectation ──▶ child expectations
        "an expectation owns every edge it points at, and never one that points at it."

FOLD    contributions run CONSTITUENT → COMPOSITE, on ascent.
        this is literally what Observe does. nothing is stored in this direction.
```

Drawing them as one arrow is what makes the direction feel undecidable. It is not a choice
between two conventions; it is one stored direction and one computed direction, and the kernel
already performs both — `Perturb` descends the stored edges, `Observe` folds back up them.

It is also consistent with authorship. The parent's proposer authors *"I am x% confident this
child completes me"* at the moment it creates the child, so the parent owns that edge — stored
downward, evaluated upward.

## What killing the node COSTS

Two things the interpretation node was buying, now given back. Both worth losing for the writer
rule, and both should be recorded as debts rather than discovered later.

**Toy 2 comes back to `T`.** With an impulse plane, ten copied records inflated one
interpretation's *fan-in* and never touched the vertical fold — a small local problem. With
impulses grading expectations directly, ten copies are ten support edges, and *"distinct is not
necessarily independent"* is once again a defect in the grounding arithmetic itself. Back where
`2026-07-27/trust.md` left it.

**Re-scoring loses its home.** `2026-07-27/workers.md` needs an impulse re-presented once the
field grows a target for it, or it sits at ≈0 forever. A second interpretation node was a clean
place to put a second reading. Now a second reading means a second support edge between the same
pair — which `O(E)` being a *set of distinct observations* is supposed to forbid, since it is the
same impulse counted twice.

```text
NEEDED   a rule for what a second grading of the same impulse means. newest wins, or
         both count, or the first is superseded. support is immutable, so it cannot be
         a rewrite.
```

## The caveat that makes this safe

Confidence is now a whole edge species, which enlarges an exemption `FLAGS.md` already names:

> Nothing grades confidence. It is the one number in the system with no blind assessor, which is
> tolerable only because it routes rather than settles.

It stays tolerable **only because support edges grade it.** Authored confidence versus assigned
support is `actions.md`'s calibration signal:

```text
assigned − predicted        →  did the claim know its own contribution?
```

and both numbers now sit on the same expectation — assigned support beneath it, authored
confidence above it — so the comparison is a local read.

> **If calibration is never computed, confidence edges are unfalsifiable decoration — a number
> the model writes that nothing ever checks.**

The separation is not what makes this safe. What makes it safe is that the separation makes the
comparison possible. Build the comparison early.

---

## Open

```text
DOES IT TERMINATE AT ROOT   confidence edges run up toward the standing root. a standing
                            expectation never closes, so confidence accumulating there never
                            resolves — consistent, but nothing says what reading it has.

RAW IMPULSE CONTENT         questions.md Q10, and narrower: only STRUCTURED impulses — world
                            returns and boundary answers — need content comparable to a
                            perceive. raw external payloads are residual by definition and
                            need none.

SECOND GRADING              see What killing the node costs. re-scoring has no home.

TWO NUMBERS ON ONE EDGE     the vertical edge carries authored confidence; the calibration
                            signal needs the assigned value too. whether both live on one edge
                            or the assigned value is derived at the join is not decided.

WHAT EXPECT WRITES         EXPECT creates a child expectation and authors its confidence edge.
                            it is now also the ONLY writer of nodes, which makes that rule
                            checkable: nothing else may call AddExpectation.
                            whether it also writes the support edge from the interpretation
                            that licensed it, or that is a separate pass, is open.
```
