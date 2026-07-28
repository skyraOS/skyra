# Truth

**Date:** 2026-07-27
**Status: TENTATIVE.** The normalizer is not settled — see *Why this is tentative* below. The
part that is not tentative is what the quantity is *for*: the change in it over time.

**Reads on:** [trust](trust.md) — Truth is a normalization of the same grounding.

---

## Definition

Truth is the normalized form of the same grounding. It is **bounded**, and it represents the
average support per independent observation.

Trust is unsigned mass; truth needs direction, so it normalizes the **signed** sum:

```text
G(E) = Σ  c(o,E)          net grounding — SIGNED
T(E) = Σ |c(o,E)|         trust — UNSIGNED mass

P(E) = G(E) / |O(E)|
```

where `|O(E)|` is the number of **distinct** grounding observations beneath `E`.

## What P measures — closure

An expectation is a state **change**, not a state ([actions](actions.md)). So its position is
not "how true is this condition" but **how much of the change has happened**:

```text
P → +1   the change has CLOSED
P →  0   nothing has moved
P → −1   the change is REVERSING
```

The signed range stops being an arbitrary choice: negative is not "no evidence," it is the world
going the other way, and that is a distinct and reportable thing.

This is also why the *change* in `P` is the signal rather than its level. It isn't that a
derivative is taken of something static — **the object is already a change**, and `P` says how
far along it is.

## Worked

```text
O₁ ──0.2──▶ A
O₂ ──0.5──▶ A
O₃ ──0.3──▶ A
A  ──0.7──▶ B

T(A) = 0.2 + 0.5 + 0.3   = 1.0     |O(A)| = 3     P(A) = 1.0 / 3 = 0.333
T(B) = 0.7 × 1.0         = 0.7     |O(B)| = 3     P(B) = 0.7 / 3 ≈ 0.233
```

**Note `|O(B)| = 3`, not 1.** The observation set beneath `B` is the union of what grounds
everything beneath it. `B` is not grounded by `A`; it is grounded by `O₁ O₂ O₃`, reached
through `A`.

The graph stores four numbers — `0.2, 0.5, 0.3, 0.7`. Both Trust and Truth are derived at
evaluation and never persisted.

## Three quantities, and what falls out of them

```text
T = Σ|c|      MASS       how much observation bears on E, either way.   ranks.
G = Σ c       DIRECTION  which way it leans, net.
|O|           COUNT      how many independent observations.
P = G / |O|   POSITION   the average.
```

`T = |G|` only when every contribution shares a sign — as in the example above. Otherwise
`T > |G|`, and the gap between them is itself informative:

```text
coherence = |G| / T  ∈ [0,1]     1 = all evidence agrees · 0 = it perfectly cancels
```

That resolves an ambiguity `P` alone cannot. A claim sitting at the midpoint may be **contested**
(`T` high, coherence low) or **unexamined** (`T` low). Nothing new is stored — it is a ratio of
two derived quantities.

---

## What Truth is for — the change, not the level

> **A truth value on its own says little. The signal is the CHANGE in truth over time.**

A level is a snapshot with no comparison point. A change is inherently comparative — it says
something happened. This is why the normalizer being unsettled is survivable: **the direction of
a truth delta is normalizer-independent.**

```text
for any monotone f,   sign( Δf(T) ) = sign( ΔT )
```

So the *sign* of a truth change is available now. Only its *magnitude* waits.

Two traps to carry into whatever the delta becomes:

```text
MOVEMENT ≠ CORRECTNESS   a small delta is CONVERGED or DISENGAGED, and movement alone
                         cannot tell them apart. it needs settlement to be read.

THE BASIS MOVES          expectations are created and destroyed, so a delta over a changing
                         node set is not well defined until BORN / DIED / MOVED are separated.
```

And `T` has its own delta, which is a different signal:

```text
ΔT rising, P flat       accumulating CONSISTENT evidence — the picture is holding
ΔT rising, P moving     the picture is CHANGING
```

That pair is what distinguishes converged from disengaged, which the first trap says a truth
delta alone cannot do.

---

## Why this is tentative

**1 · The mean is not monotone in evidence.**

```text
one strong observation                    T = 0.9              |O| = 1     P = 0.900
add nine weak but POSITIVE observations   T = 0.9 + 9(0.05)    |O| = 10    P = 0.135
                                            = 1.35
```

Nine additional *supporting* observations and truth falls from 0.900 to 0.135. Any contribution
below the current mean drags it down even when positive, so the mean **punishes breadth**. A
claim with one strong and many weak grounds reads as barely supported.

**2 · The scale is unpinned.** `P(A) = 0.333` from three positive observations is coherent on a
`[-1,1]` scale centred at **0** ("moderately supported") and wrong on a `[0,1]` scale centred at
**0.5** ("leaning false"). These are different scales and nothing yet says which.

**One candidate replacement, recorded not adopted:** squash the **signed** sum.

```text
tanh(G)   for a 0-centred scale on [-1,1]
σ(G)      for a 0.5-centred scale on [0,1]
```

Bounded, monotone in signed evidence, needs no complement or partition, and leaves the graph and
the derivation untouched — only the normalizer changes. It saturates at high evidence, which is a
display concern rather than a storage one, since `P` is derived either way.

> **It must take `G`, not `T`.** An earlier draft here said `σ(T)`, written before trust became
> unsigned mass. `σ` of an always-positive quantity is always above the midpoint, so strong
> refutation would read as strong support. **Direction can only come from the signed quantity.**
