# kernel/03 · Think

**Package:** `kernel`

The universal possibility-facing interface. It creates potential and performs none of it.

---

## 1 · The signature

```go
func (k *Kernel) Think(f Frame, a ThinkArgs) ([]Candidate, error)

type ThinkArgs struct {
    Entities     []ids.EntityID
    Expectations []ExpectationRef
    Observations []observation.Observation
}
```

```text
ENTITY[] × EXPECTATION[] × OBSERVATION[]   within FROM × HORIZON
    → BOUNDED RANKED EXPECTATION-OBSERVATIONS
```

`From` and `Horizon` arrive inside `Frame` and are not members of `ThinkArgs`. That separation is
structural, not cosmetic: it is what makes *"the caller supplies three dimensions, the runtime
supplies two bounds"* true in the type rather than in the documentation.

**There is no `address` operand**, per P12 and the rejected shape *THINK RETURNS AN ADDRESS*.

## 2 · Why all three are slices

Cardinality carries meaning, which is why the ordinary one-item call still passes a slice:

```text
EMPTY            unresolved. Think may propose freely within From and Horizon.
CANDIDATE SET    constrained. Think may propose only within the supplied set.
SINGLE VALUE     fixed. Think may compose around it but may not substitute it.
```

> *"This is why the operands are sets even when the ordinary call supplies one item. Cardinality
> says how much of that dimension remains open without requiring a second constraint language."*

A `*ids.EntityID` with `nil` for "unconstrained" would express two of the three states and lose the
candidate-set case, which is the one that makes bounded proposal possible.

The maximally empty call — all three empty — asks *what potential is available from here, for this
namespace and requesting engram, within this Horizon?* **Empty does not mean global**; it is still
bounded by `From` and `Horizon`.

---

## 3 · What comes back

```go
type Candidate struct {
    Observation observation.Observation   // Provenance.Origin == OriginThink
    Rank        int
}
```

**The candidate is an observation, not a handle.** This is the load-bearing choice in the whole
Think surface:

> *"The expectation is inside an observation because it is still potential. Think has not installed,
> opened, executed, or made it durable."*

The candidate observation's content may identify the proposed expectation, the entity, the
observations it would receive, the narrower Horizon it would require, and the evidence it was formed
from. It may **not** contain an address, a transport, a live relationship, or a claim that contact
has been established.

`Candidate` has **no methods**. Nothing on it runs, opens, or resolves anything. P7 becomes
structural: there is no way to execute a candidate except by handing it to `Decide`.

### Bounded plurality

```text
ONE RESULT          leaves no local selection when alternatives were available.
UNBOUNDED RESULTS   lets system-wide potential flood one frame.
BOUNDED RESULTS     lets Think rank globally while the caller selects locally.
```

Ordinarily no more than five. Five is an operating bound, not a constant — the invariant is *a small
ranked set*. Whether the bound is fixed or configurable beneath a maximum is open.

### The empty case is not an empty slice

```text
"NOTHING FITS" IS A CONSTRUCTED CONCLUSION, NOT AN EMPTY LOOKUP RESULT.
```

Think returns one composable diagnostic observation rather than silently returning nothing,
because absence cannot distinguish *no match*, *no authority*, and *failure to think*:

```go
type Diagnosis uint8

const (
    NoKnownPotential  Diagnosis = iota   // nothing considered could satisfy the constraints
    BeyondYourHorizon                    // relevant potential exists only outside this authority
)
```

A diagnosis must not expose addresses, endpoint failures, or hidden strategy names merely to explain
itself.

```text
FOR REVIEW    modelling the diagnosis as a Candidate with a Diagnosis field, versus a separate
              return, versus a distinguished observation kind. The source document says Think
              "returns one composable diagnostic observation," which argues for it being an
              observation and therefore a Candidate. That makes `[]Candidate` never empty on a
              successful call, which is a strong and checkable invariant — recorded as such and
              not settled.
```

---

## 4 · Hidden behaviors are hidden from selection, not from causality

```text
think
├── think-reach-entity
├── think-recognize-observation
├── think-discover-novelty
└── further thinking behaviors which earn their place through use
```

These are possible internal decompositions, not a required child list, and **none of them is
nameable in any signature above**:

```text
HIDDEN FROM CALLER SELECTION   the caller cannot reach around Think and name its private child.
VISIBLE IN THE CAUSAL CONE     the behavior which produced a candidate remains attributable.
```

The second is what `Provenance.Engram` on the candidate observation carries. The caller cannot
*select* the thinking behavior but can *see* which one authored a candidate after the fact — which
is the precondition for the two-sided learning the document requires:

```text
THINK LEARNS ACROSS CALLERS   which internal behavior produces useful potential for classes of
                              request.
THE CALLER LEARNS LOCALLY     which candidate expectation serves this engram's purpose.
```

**The regress stops at Think itself.** Calling Think does not require first thinking about which
thinker to call.

---

## 5 · Deciding on the observation, not on its text

The one rule that constrains how `Decide` may be typed:

> *"When a thought-derived expectation is executed, the observation which proposed it remains a
> constituent of that execution."*

```text
A, B → T → C → Decide → P → D
```

Copying `X` out of `C` and opening it as unrelated text would sever the edge through the thought
that authored it. The system would retain the behavior while losing why it exists — the
unattributed-authoring hole.

**In the type surface this is why `Decide` takes a `Candidate` and not an `ExpectationRef`.** There
is no signature through which a caller can pass the expectation alone.

---

## 6 · What Think may not do

Restated because each maps to something absent from the types above:

```text
open a candidate possibility        Candidate has no methods.
contact an entity                   no transport is reachable from ThinkArgs or Candidate.
resolve or expose an address        no Address type exists in this package.
widen Horizon                       Horizon is in Frame, opaque, and narrows only.
rewrite From                        From has no setters.
make provisional behavior durable   only Remember writes. see 06-remember.md.
return an unbounded population      []Candidate, bounded by contract, ordinarily five.
hide which behavior produced a      Provenance is set by the substrate on every candidate.
  candidate
```

---

## 7 · Reconciliation with the runtime

```text
discover        →  THINK. and the rename is a widening: discover was a lookup, Think searches
                   AND reasons. RECALL is the flat lookup inside it; DISCOVER is reserved for
                   novelty. (../../../semantics/terminology.md addendum.)

graph.Summary   →  the current runtime's Summary exists "because something had to render a child
                   set into a prompt." Under Think that rendering is entity-side marshalling and
                   the type does not survive here — which agrees with ../../README.md's own hit list.
```
