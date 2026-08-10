# Traversal — does the two-phase interface survive?

**Date:** 2026-08-08
**Status: OPEN.** The question has been asked and nothing has been argued. This doc holds what the
split is currently carrying, so that whatever replaces it is checked against the real list.

> **`Perturb` / `Observe` is not one decision. It is at least four, and they were made
> separately.**

---

## The question

Whether the traversal interface stays two methods, and if not, what happens to the work those two
methods are currently doing for free.

Nothing is argued here yet. What follows is the inventory.

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

**The spine test is the one flagged to watch** — kernel.md says it fails if the kernel ever gains
the ability to branch. Any replacement interface should be checked against it first, because it is
the test most likely to be quietly weakened by a change that looks like a simplification.

## The thing that must not be re-imported

Whatever replaces it, the constraint from `kernel.md` §2 is the floor:

> The kernel cannot enumerate children, interpret grouping, select or order a frontier, compare
> magnitudes, read a cost, or infer meaning.

Safe by incapacity, never by restraint. An interface that hands the kernel more to look at is a
worse interface no matter how much simpler it reads, and `locality.md` already names the specific
regression to watch for — a driver that rebuilds the tree is enumerating composition, Law 7
inverted.

---

## To fill in

```text
THE ACTUAL ARGUMENT    not written. what is wrong with the split, in its own terms, before
                       anything is proposed to replace it.

WHAT REPLACES IT       nothing proposed.

WHICH OF THE SIX       a replacement does not have to keep all of them. it has to say which
SURVIVE                ones it drops and what pays for them.
```
