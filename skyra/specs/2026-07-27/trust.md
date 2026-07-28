# Trust

**Date:** 2026-07-27
**Status: FINAL.**

> **Trust is accumulated observational magnitude, scaled by the expectation edges it passes
> through.**

It measures how much independent observation ultimately grounds an expectation. It is
**unbounded**, and it is a **ranking** signal — attention, retrieval, what to look at next. It
is not a truth value and may not settle anything.

---

## The invariant

```text
SUPPORT     stored. on edges only.        signed value s(x,y)
MAGNITUDE   stored nowhere.               derived at evaluation, discarded when it ends
```

Expectations store no scalar. Every magnitude in the system is a property of the support graph,
computed on demand.

## The graph

```text
Observation ── support ──▶ Expectation
Expectation ── support ──▶ Expectation
```

**Observations are the only source of grounding.** Expectations neither create nor store it;
they only scale what reaches them from beneath.

These edges are owned by the expectation they ground — *an expectation owns every edge it points
at, and never one that points at it* ([README](README.md)). So "what grounds `E`" is a field;
"what `E` grounds" is a query.

## Definition

Let `O(E)` be the set of **distinct** observations that ground `E`, and let `c(o,E)` be the
signed contribution of observation `o` after applying every edge scalar along its path.

```text
T(E) = Σ  |c(o,E)|        over o ∈ O(E)
```

**Trust is unsigned.** It measures how much independent observation bears on `E`, not which way
it leans. A well-refuted claim is exactly as worth attending to as a well-supported one, and
contested material — evidence pulling both ways — is the most worth attending to of all. Signed
trust would sort both to the bottom. Direction lives in [truth](truth.md); mass lives here.

`c` is the product of the edge scalars on the path, beginning with the observation's own edge:

```text
O ──0.5──▶ A ──0.7──▶ B

c(O,B) = 0.5 × 0.7 = 0.35
```

**When an observation reaches `E` by more than one route, it is counted once, at its strongest
route — greatest absolute magnitude, sign preserved:**

```text
c(o,E) = ∏ s(e) along p*      where  p* = argmax over paths p of | ∏ s(e) |
```

Note the absolute value in the selection. A signed `max` would pick `−0.2` over `−0.8` — the
*weaker* grounding — because it reads "largest" as "most positive."

Distinct **observations**, never distinct paths. Summing routes would count one observation
twice, which is the thing `O(E)` being a set exists to prevent.

## Worked

```text
O₁ ──0.2──▶ A
O₂ ──0.5──▶ A
O₃ ──0.3──▶ A
A  ──0.7──▶ B

T(A) = 0.2 + 0.5 + 0.3           = 1.0
T(B) = 0.7(0.2) + 0.7(0.5) + 0.7(0.3) = 0.7
```

The graph stores four numbers — `0.2, 0.5, 0.3, 0.7`. Nothing else is written.

## Two properties that fall out

**Ungrounded chains are worth nothing.** Because expectation edges *multiply* rather than add,
an expectation with no observation beneath it contributes `0` however strong its support edges
are. A chain of expectations propping each other up has `T = 0` throughout. Rumination is
suppressed arithmetically, with no rail required.

**Depth attenuates for free.** With edge values in `[-1,1]`, every hop shrinks the contribution
(`0.5 × 0.7 = 0.35`). Grounding weakens with distance from observation and needs no decay
constant. This is also why the edge range is load-bearing: values outside `[-1,1]` would let
chains amplify without bound.

## What Trust is and is not

```text
IS      unbounded · ordinal · grows with the amount of independent grounding
        a RANKING signal — attention, retrieval, what to look at next

IS NOT  a probability · a confidence · a position on any scale
        it may never close a claim or settle anything
```

`T` conflates two axes deliberately: how strongly each observation supports, and how many there
are. Anything that needs those separated must not read `T`.

## Two things to state, neither open

**Sign propagates through multiplication, and it is correct.** Negative evidence for a negative
cause is positive evidence for the effect:

```text
A = "my phone is low on battery"
    o₁  charger plugged in            −0.2
    o₂  charging for 300 ticks        −0.3
    o₃  "your phone looks full"       −0.7

A ──(−0.6)──▶ B = "I'll have my phone on me in an hour"

c(o₁,B) = −0.2 × −0.6 = +0.12
c(o₂,B) = −0.3 × −0.6 = +0.18
c(o₃,B) = −0.7 × −0.6 = +0.42
```

Evidence that the phone is *not* low, through a link saying low battery *hurts* availability,
supports availability. Multiplication gets this for nothing.

**Note the distinction:** *negatively grounded* flips sign through a negative scalar.
*Ungrounded* (`c = 0`) propagates nothing through any scalar. Those are different, and only the
first produces the flip.

**Acyclicity is a precondition, and it is free.** `T` is undefined on a cycle. Support edges are
written once at creation and point backward — you cannot cite what does not yet exist — so no
cycle can form. Stated because it is load-bearing and otherwise invisible.

## Open — distinct is not necessarily independent

`O(E)` is a set of distinct observation **records**. Two records may describe one underlying
fact, and trust would count both.

The fix is not "dedupe repeats," because repetition is sometimes real evidence — *"I looked five
times and it was still absent"* is stronger than looking once, and that sustained-absence
property is one worth keeping. The failure is narrower: **correlated evidence counted as
independent.**

Nothing in this set defines observational independence. One candidate, recorded not adopted:
independence is **measurable rather than declarable** — sources that reliably co-occur are not
independent, and the amount of co-occurrence is exactly what a coupling statistic measures.
Until something decides this, `T` overstates on correlated sources.

---

## Truth

A normalization of this same grounding, and **TENTATIVE** — the normalizer is not settled.
See [truth](truth.md).

What is not tentative is what that quantity is for: **the change in it over time, not the
level.** The direction of a truth delta is available without settling the normalizer at all,
since `sign(Δf(T)) = sign(ΔT)` for any monotone `f`.
