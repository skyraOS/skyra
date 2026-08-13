# Distance

**Date:** 2026-08-13
**Status: WORKING NOTE.** Proposed, not built and not measured. Companion to
[`reach-abstraction.md`](reach-abstraction.md), which establishes what the operating system drops
and therefore what this has to recover.

> **Distance is the observable cost of crossing from where an invocation is oriented to an entity.**

```text
ORIENT  ×  ENDPOINT  ×  EXPERIENCE  →  DISTANCE
```

---

## The three terms, and why each is load-bearing

```text
ORIENT       the FROM.   where this invocation currently sits.        primitive 7
ENDPOINT     the TO.     a candidate address, held by a host.         Where · decision 4
EXPERIENCE   the COST.   what crossing has actually cost before.      Where · decision 5
```

Drop any one and it fails differently:

```text
NO ORIENT       distance is a RELATION and needs two operands. An address tells you where the
                thing is; whether that is near is unanswerable without knowing where you are.

NO ENDPOINT     nothing to cross to.

NO EXPERIENCE   the cost would have to be DECLARED — and a declared topology is wrong the moment
                a laptop leaves the house. Topology is where declaring is most tempting, because
                it feels static. It is not.
```

---

## Distance is the resolver's output, not a kernel primitive

Cost is not one number. A fast-but-flaky endpoint and a slow-but-reliable one are not comparable on
one axis without a weighting, and choosing that weighting is a judgment.

```text
THE KERNEL SUPPLIES    orient (live topology) · endpoints (host-held) · experience (history)
THE RESOLVER RANKS     and the ranking IS distance
```

Which is `Where` decision 7 — *the kernel routes; the resolver ranks*. So distance is **not an
eighth primitive**. It is what a `reach` engram computes, which makes it learnable: different reach
strategies can weigh latency against reliability differently and compete on their own records.

---

## What is actually measurable

The operating system reports none of this. All of it is obtained by timing the call and recording
the outcome — see `reach-abstraction.md` §4.

```text
LATENCY        wall-clock from reach to handle.
RELIABILITY    the fraction of reaches that produced a handle at all.
DOMAIN         does it fail when I fail. Not measured per-reach — inferred from CORRELATED
               FAILURE across many, per Who · decision 4.
```

Three axes, and they do not collapse into a scalar on their own. That is the reason ranking is
semantic and lives in an engram.

### Measure at reach, not at return

```text
REACH DISTANCE   time to obtain a handle.                    ← this is distance
WORK DISTANCE    time to obtain an answer.                   ← this is the WORK
```

Reaching Claude takes milliseconds; the inference takes seconds. Conflating them would make a hard
question look like a far entity, and the resolver would learn to avoid an endpoint for being asked
difficult things.

**A failed reach is a measurement too.** `Where` decision 5 already says fallback is the learning
signal rather than error handling — this is the same observation read as cost.

---

## It is per-invocation, because `Orient` is live

Not a static edge weight between a host and an entity. A fresh evaluation each time, because the
*from* moved.

```text
A TOPOLOGY TABLE   one number per pair. correct until something moves, then silently wrong.
DISTANCE           computed at the moment of crossing, from where you actually are.
```

Same host, same entity, different distance at 9am and at the coffee shop — **with nothing to update,
because nothing was stored.**

---

## Derived, never stored

```go
// ByReturn — "The credit walk's entry point. Derived, rebuildable, deletable."
```

Distance belongs to that class. A stored distance is a declaration with a timestamp on it, wrong the
moment anything moves. Computed from experience at each reach, it cannot go stale — it can only be
**out of date**, which the next observation fixes.

That is the whole difference between this and a topology map, and it is why the third term is
`Experience` rather than `Topology`.

---

## Cold start

An endpoint never tried has no experience, so it has no distance. It might be the fastest one
available.

The design's existing posture applies without modification:

```text
credit.md   "Exploration costs nothing permanent — a possibility may try five things, use one,
             and only the one it used forms a child."

world.md    "COLD START IS THE WHOLE COST — send Python to a TV, it fails immediately, one
             record, no model call. Compare a wrong declaration, silently believed forever."
```

So an untried endpoint is tried, and one bad attempt is the entire cost of finding out. The
alternative — assuming its distance from something structural about the address — is a declaration
in disguise.

---

## Open

```text
ONE NUMBER OR THREE      the resolver must rank, which requires an ordering. Whether it collapses
                         latency and reliability itself, or the frame carries all three and the
                         weighting is per-strategy, is undecided.

HOW MUCH IS ENOUGH       one sample is noise. Nothing yet says how many reaches constitute
                         experience, or how old an observation stops counting.

DOMAIN IS NOT A COST     it is a correlation, and it is the one axis that cannot be measured
                         from a single crossing. It may not belong in this signature at all.

DOES DISTANCE FEED       a fast endpoint is not the right one if it reaches the wrong entity.
SELECTION, OR ONLY       distance probably ranks addresses WITHIN an already-decided entity,
RESOLUTION               and never influences which entity is chosen. Unconfirmed.
```
