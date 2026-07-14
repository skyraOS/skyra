# Transport — Budget, Cost, Affordability

**Status:** DESIGN · part of [kernel-v2](README.md)
**Lifts from:** `specs/2026-07-08/proportional-amplitude-allocation.md` (un-superseded)

## The arithmetic — this is the entire kernel

```text
child_budget = (parent_budget − parent_cost) × weight
descend iff  child_budget ≥ child_cost              ← affordability
```

Two lines. No decay factor, no threshold, no normalization, no notion of what the budget measures.

## The budget axis is polymorphic

The kernel does not know the units. Which is why the same kernel does three different jobs:

```text
INFERENCE CALLS   the workflow wave    how much planning can be afforded
CONTEXT TOKENS    the present wave     how much of the graph fits in the window
HOPS / MS         a perceive           how far into the world a reach can go
```

And it means the same arithmetic **is** gRPC deadline propagation: a request carries a deadline, each
hop spends some, passes the remainder down, and refuses a call it cannot afford to finish.

> **The present is what fit in the budget. Attention is affordability.**

That is not a metaphor. It is the mechanic.

## Proportional allocation — the correction

The bug the 07-08 spec found: **child amplitude was split EQUALLY, then scored by edge strength.**

```go
childAmp := (node.Amplitude * amplitudeDecay) / float64(len(candidates))   // equal split
score    := cand.activation * childAmp                                     // strength only ranks
```

So edge strength decided whether a child *survived* but not how much energy it *received*. A
strong edge and a weak one got the same wave.

**The rule:**

```text
available = parent_budget − parent_cost
claim_i   = |weight_i|
sum       = Σ claim_i

child_budget_i = available × (claim_i / sum)
```

Budget follows the edges that actually carry the signal.

```text
weights  A=0.44  B=0.73  C=0.22   available=90

EQUAL (wrong)        A=30    B=30    C=30
PROPORTIONAL         A=28.2  B=46.8  C=14.1     ← B is the strong edge; B gets the wave
```

## Signed magnitude: polarity is meaning, not transport

**Allocate by |weight|. Carry the sign separately.**

A strongly *negative* relationship is strongly *relevant* — it must not become invisible to
traversal just because its sign is negative. "This absolutely does not work" is information worth
spending budget on.

```text
child_budget_i = available × (|w_i| / Σ|w|)     ← MAGNITUDE drives transport
child_polarity = sign(w_i)                       ← SIGN is meaning, carried, never spent
```

The kernel never interprets the sign. It hands it to the cell and forgets it.

## Conserve versus broadcast — which is AND versus OR

The cell chooses whether its weights sum to 1, and **the kernel does not normalize**:

```text
weights SUM TO 1   CONSERVE    a bounded frame, split among alternatives   → an OR node
weights ALL 1      BROADCAST   flood; every child gets the full remainder  → an AND node
```

This was written to separate cognition (bounded) from space (flood). It turns out to be the
distinction between **disjunction and conjunction**:

```text
a triple      is an OR node    ANY prior expression could ground it. alternatives COMPETE.
              → weights sum to 1

an expression is an AND node   ALL of its requires must be grounded. every child is NEEDED.
              → weights all 1
```

**The kernel needs no new concept.** It multiplies budget by weight. It has no idea it is doing
backward chaining.

## Cost

```go
Cost(origin Space, impulse any) int
```

Stamped on the node, read by the kernel, in the budget's own units. A cell that is expensive to bring
into the present (a large file, a long transcript) costs more tokens. A world hop costs 1. A remote
call costs milliseconds.

**Affordability is the only gate.** There is no threshold, no cutoff, no relevance score in the
kernel — a child is descended into if and only if its share of the remaining budget covers its cost.

## What the kernel must never learn

```text
alignment · probability · relevance · decay · grounded · a grade · a target · AND · OR
```

If any of these appears in kernel code, the kernel has stopped being a mechanism and become
choreography — and choreography is what the next model absorbs.
