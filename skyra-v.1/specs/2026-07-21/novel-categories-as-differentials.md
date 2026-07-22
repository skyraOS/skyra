# Novel Categories as Differentials Against the Nearest Known Code

**Date:** 2026-07-21  
**Status:** RESEARCH STUB · a candidate mechanism for the one piece the two-axis notes
left open (novel-category bootstrap). Not proven, not paper-tested. It reduces the hard
problem rather than solving it, and it is offered as the shape to poke at next.  
**Addresses:**
[Two-Axis Inference](two-axis-inference.md) open problem 1 ("abduction is search, not
just scoring") · [Predicate Diagnosticity](predicate-diagnosticity.md) cold-start ·
[The Two-Bit Outcome](the-two-bit-outcome.md) (the resemblance's own context-prediction)  
**Leans on:**
[../2026-07-20/replacing-triples-with-compression.md](../2026-07-20/replacing-triples-with-compression.md)
(a differential is an MDL-native delta code)

## The problem it targets

Two-axis inference is honest that its hardest piece is uninvented: a genuinely novel
instance needs a category the heated wave cannot retrieve, because a wave returns only
what memory already indexes. The LLM must then propose the category — but a novel
category earns epistemic standing only by licensing a world-falsifiable consequence,
and it has no consequence history to estimate one from. That is where two open problems
collide:

```text
novel abduction        needs a first falsifiable consequence
predicate diagnosticity needs held-out P(fire | category-right) to select a predicate

a category with one instance has neither
```

Estimating a category-conditional firing rate from a single point is not estimation.
So a category minted from nothing is both unfalsifiable and un-diagnosable at birth.

## The proposal

Do not mint novel categories from nothing. Mint them as a **differential against the
nearest known code**.

```text
FROM SCRATCH (intractable)
    "this is a new category P"
        → invent a whole consequence Q from one instance
        → estimate diagnosticity of Q from one instance

DIFFERENTIAL (tractable)
    "this is like known code C, except it differs on dimension D"
        → inherit C's consequence edges as priors
        → stake exactly one new discriminating claim: the value of D
```

A differential category is not a fresh point in category space. It is an existing,
well-populated category plus one declared axis of departure. The claim it must
world-settle shrinks from "an entire category and its consequence" to "the single
coordinate on which this differs from C."

## Why this shrinks both hard problems at once

**The falsification address is inherited-plus-one.** C already licenses world-checkable
consequences with counted `P/N`. The differential borrows them as priors and adds one
new consequence: the observable that D predicts. The novel category therefore arrives
*with* a falsification address instead of needing to invent one — it inherits C's and
owes only D's.

**Diagnosticity no longer estimates from one point.** The only new predicate is the one
for D, and D is not estimated as a category-conditional from a single instance. It is
estimated as a **contrast against a well-populated neighbor**:

```text
base rate of D-predicate       measured over all observations (label-free, cheap)
C's behavior on D              measured over C's many settled instances
the differential's claim       "D fires here where it does not fire for C"
```

That is a two-sample contrast against a populated reference, not a one-sample estimate.
D is diagnostic *by construction* — it was chosen precisely to separate this instance
from C, so it is the discriminating coordinate the two-bit-outcome demands, not a
diffuse address the lucky hit can hide behind.

## Worked sketch

```text
nearest known code C:  normal-fridge-running
    consequence edges:  {compressor running, door closed} → cold within 10 min   P high

novel instance:        fridge that is cold-inside but warm-at-the-vent
    from scratch:       "a new appliance category" — no consequences, no data
    differential:       "like normal-fridge-running EXCEPT vent-temperature is warm"
        inherits:       cold-within-10-min consequence (as prior)
        stakes new D:    vent-temperature warm at the eligible observation
```

The differential licenses two settlements at two addresses: the inherited "still gets
cold" (does it behave like C where it should?) and the new "vent runs warm" (the
declared departure). If the world confirms both, a real sub-category has been witnessed
with a discriminating address from its first instance. But **one** inherited consequence
failing does **not** by itself refute the neighbor: a single failure is ambiguous between
"wrong neighbor" and "that consequence rule is just broken/noisy" (even a true C can fail
one consequence — the refrigerator problem). To conclude "not a variation of C, fall back
to full abduction" needs either the differential's **own** claim D to fail, or **broad**
failure across several inherited consequences — not one. A lone inherited failure updates
that consequence edge and nothing more.

## It is the MDL-native move

A differential category *is* a delta code: describe the new thing as an existing code
plus a correction. That is the cheapest possible description of something almost-known,
so it is exactly what the compression gate should prefer — recurrence pays for the
delta, and the delta is small. From-scratch categories only earn their full description
length when no near neighbor exists. Compression and falsifiability push the same way
here: both favor "known code + one declared difference" over "invent everything."

## Where it plugs into the two edge types

```text
RESEMBLANCE (differential)
    instance ──"like C, differs on D"──▶ C
        │ inherits
        ├── C's consequence edges (priors, borrowed history)
        └── one new consequence edge for D (the discriminating stake)
```

The resemblance's derived confidence still comes only from downstream world
settlement (unchanged). What differs is that a *novel* resemblance is no longer
confidence-orphaned at birth: it starts with borrowed priors and one honest new bet.

## What it does NOT solve

- **Genuine novelty with no near neighbor.** If nothing in memory is close, there is no
  C to differential against, and you are back to full from-scratch abduction. This
  mechanism covers the almost-known, which is most of experience, not the truly
  unprecedented.
- **The delta can still be diffuse.** "Differs on D" only helps if D's predicate is
  itself diagnostic. The predicate-diagnosticity discipline still applies to D; the
  differential just makes D estimable against a populated reference instead of one
  point.
- **Nearest-neighbor selection is propose-side.** Retrieval or embedding proposes "C is
  the closest known code," and that proposal is never evidence — same rule as every
  other tool. A wrong neighbor makes a bad differential that the world then refutes.
- **Promotion is open.** When does an accumulated differential stop being "C plus D" and
  become its own root code C′? Presumably when its own instances are numerous enough to
  estimate consequences directly and the delta stops being small — the MDL crossover.
  Threshold unspecified.

## Open questions

- What metric proposes the nearest neighbor without turning retrieval rank into belief?
- Can a differential inherit from more than one neighbor (a blend), or must it pick one?
- How many confirmed differentials before promotion to a root code, and what deopt path
  demotes a promoted code whose delta stopped paying?
- Does the inherited-consequence prior contaminate credit assignment (a differential
  looking good only because C's borrowed consequences carried it)? Provenance must keep
  borrowed priors distinguishable from the differential's own earned history.

## Bottom line

The two-axis notes left one hole: how a novel category gets a falsification address and
a diagnosticity estimate at birth. Differentials fill most of it by refusing to mint
from nothing — a new category is the nearest known code plus one declared, diagnostic
difference, so it inherits an address and estimates its one new claim against a
populated reference instead of a single point. It is also the cheapest description, so
compression and falsifiability agree on it.

It does not solve genuine novelty, and the delta must still be diagnostic. But it turns
"invent a category and its consequence from one instance" — hopeless — into "predict the
single coordinate on which this differs from something you already know" — falsifiable,
diagnosable, and MDL-cheap. That is the size the hardest problem needs to be shrunk to
before a mechanism can hold it.
