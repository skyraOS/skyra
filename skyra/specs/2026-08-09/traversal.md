# Traversal — does the two-phase interface survive?

**Date:** 2026-08-08 · carried into 08-09
**Status: CHECKLIST.** The inventory was written before anything replaced the split. It is kept as
the **bill** — what the two methods were carrying, and what pays for each item now.

> **`Perturb` / `Observe` is not one decision. It is at least four, and they were made
> separately.**

**All bare doc references below are to `2026-08-06`** — `phases.md`, `behaviors.md`, `locality.md`,
and that set's `kernel.md`. This set now has a `kernel.md` of its own and they are not the same
document.

---

## The question, and what answered it

Whether the traversal interface stays two methods, and if not, what happens to the work those two
methods were doing for free.

**It is not two methods, and it is not methods.** `Perturb` was `Observe` with `realized = nil` — a
name for the first iteration of a loop already iterating. What replaced it is a **grammar**: a
possibility emits `SELECT` or `COMMIT` and nothing else parses. The inventory below is what that has
to pay for, settled item by item at the end.

## What the split is load-bearing for

Each of these was derived somewhere else and landed on the split afterward. That is why the split
feels bigger than an interface choice — it is the join point for four independent arguments.

```text
1  EVICTION SAFETY      descent writes nothing, so a stack cut off mid-descent is a no-op and
                        needs no rollback. mid-ascent leaves a PREFIX, and append-only makes a
                        prefix consistent.                              phases.md

2  NOTHING FIRES        a file's primitive runs in Observe, so an evicted wave cannot have
   INSIDE A WAVE        acted. locality.md upgrades this from a rule to geometry once firing
                        moves between invocations.                      behaviors.md · locality.md

3  THE API SPLIT        Perturb is handed a Reader and has no writer to call. the phase rule
                        cannot be broken by forgetting it.              phases.md

4  ORDERING WITHOUT     post-order fires the inner Observe first, so worker order is nesting
   A SCHEDULER          depth. "not the scheduler's job and must not become it."
                                                                        behaviors.md
```

And two more that are structural rather than safety:

```text
5  THE RETURN PATH      Observe returning a Space is what deleted Node, TraversalPair,
   IS THE TREE          Survivors, NextChild, and the tree double-buffer. the kernel's whole
                        claim to existence.                             kernel.md §3 · locality.md

6  THE ONE SHAPE        every node asks two questions, and the answer to the second IS what a
                        phase would have been. there is no dispatch table because the split
                        already carries the dispatch.                   behaviors.md
```

## Where the split is doing nothing

Worth naming, because a collapse argument should get credit for these rather than have to defend
against them.

```text
LEAF NODES        Perturb returns nil immediately. an action is a file. the read phase is a
                  formality there, and EXHAUST/FINISH being separate is a fairness concession
                  to the stepper, not a cognitive distinction.

WALK              Perturb takes one member and returns the rest; Observe folds into ephemeral
                  accumulation. it writes nothing durable in EITHER phase, so the read/write
                  justification does not apply to it at all.

ITERATION         Perturb ITERATES, which already means one frame runs several logical phases.
                  so "two phases" was never the real count — the frame is a pipeline, and the
                  interface names two of its moments and not the others.
```

That last one is the sharpest observation available before any argument is made: **the interface
says two, the mechanism does N.** A frame is asked repeatedly and pops when it promotes nothing,
which is a loop, and `Perturb`/`Observe` is a naming of the loop body's edges rather than a
description of it.

## The tests that would have to be re-earned

`kernel_test.go`, and each is a claim rather than coverage — so each is a thing to preserve, not a
file to update.

```text
descent then post-order ascent            the return path IS the tree
nil Perturb is base case, still observes   termination is a fixed point, not a verdict
transcript is the spine                    depth and parentage
nil realization still records the hop      the two artifacts keep different authority
grouped members never cross the kernel     3 members run · 1 hop recorded
only the boundary reaches the concrete     the invariant, as method reachability
concurrent waves share no state            nothing shared to corrupt
```

---

## The bill, settled

Five of the six are kept and one is deliberately dropped.

```text
1  EVICTION SAFETY     KEPT, AND CHEAPER. every durable write is caused by a commit, so a
                       possibility cut off mid-exploration wrote nothing. the phase rule
                       engineered this with a permission split; now it is just where the
                       information lives.                                       credit.md

2  NOTHING FIRES       KEPT. same mechanism as 1 — there is no write to leave behind, so
   MID-DESCENT         an evicted possibility leaves no trace at all.           credit.md

3  THE API SPLIT       DROPPED, and replaced by something stronger. it was "the phase rule
                       cannot be broken by forgetting it." now the rule is not expressible:
                       incapacity is a PARSING property, not a handed-in Reader. two forms
                       parse and there is no third.                             grammar.md

4  ORDERING WITHOUT    KEPT, and made explicit. ordering was nesting depth; it is now
   A SCHEDULER         invocation cardinality. the runtime does not find, select, or order a
                       frontier — it executes the group already emitted.        grammar.md

5  THE RETURN PATH     KEPT. the commit is singular and a returning child hands its parent
   IS THE TREE         (observation, possibility) — exactly what forms the child engram. no
                       Node, no TraversalPair, no double-buffer.    grammar.md · credit.md

6  THE ONE SHAPE       KEPT, and widened. one universal signature, and `Step` is the only
                       method that produces anything. no dispatch table, and nothing
                       per-callee to discover.                                  grammar.md
```

**The dropped one is the honest cost**, and it is not free: an API split is checkable by a compiler,
where a grammar is checked by a parser at runtime. What is bought is that the constraint now covers
everything a possibility could emit rather than the two methods someone remembered to split.

## The tests are not migrated

`kernel_test.go` belongs to `skyra-dna-v4`, which stays and runs. Nothing here migrates — the fork is
by version, not by rewrite. **The claims those tests encode still apply as claims**, and the spine
test is the one to re-earn first: `2026-08-06/kernel.md` flags that it fails if the kernel ever gains
the ability to branch, and it remains the test most likely to be quietly weakened by something that
reads as a simplification.

## The floor still holds

The constraint from `2026-08-06/kernel.md` §2, which no replacement was allowed to re-import:

> The kernel cannot enumerate children, interpret grouping, select or order a frontier, compare
> magnitudes, read a cost, or infer meaning.

**Carried intact.** See `kernel.md` in *this* set, where it becomes *the kernel holds execution
topology and no semantics* — it may know that a chain runs and where its head is, and it may read
nothing. Safe by incapacity, never by restraint.

`locality.md` names the regression to watch: a driver that rebuilds the tree is enumerating
composition, Law 7 inverted. An interface that hands the kernel more to look at is a worse interface
however much simpler it reads.
