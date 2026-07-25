# Perturb and Observe — Segmented Descent, Concrete Ascent

**Date:** 2026-07-23  
**Status:** CORE DECISION. The phase meanings, executable-Space return, and
direction of dataflow are settled.

**Interface companion:** [kernel-interface](kernel-alteration.md) specifies the
exact kernel boundary. Both documents use the same canonical names and
dataflow.

**Reads on:** [tokenizer](tokenizer.md#the-concretesegmented-cycle) for concrete
and segmented representations, [traversal-spine](../2026-07-22/traversal-spine.md) for the
field layers, and
[belief-expression-inference](belief-expression-inference.md#the-ascent-builds-the-prompt)
for prompt construction during ascent.

---

## 1. Decision

One wave has two complementary phases:

```text
PERTURB
    traverse the segmented field forward

    SegmentedObservation
        ↓
    SegmentedBelief
        ↓
    SegmentedAction

OBSERVE
    unwind the same stack through concrete counterparts

    ConcreteAction
        ↑
    ConcreteBelief
        ↑
    ConcreteObservation
```

Perturbation expands a concrete disturbance through the distributed,
individually addressable representation. Observation collapses the traversed
region into bounded concrete realizations.

> **Perturb traverses the segmented representation. Observe concretely realizes
> what that traversal reached.**

There is no third `Express` traversal phase in the core interface. What the old
interface called `Observe` becomes `Perturb`. What it called `Express` becomes
`Observe`.

---

## 2. Perturb traverses only segmented material

The kernel does not build a concrete tree during descent. It asks the current
segmented Space what executable continuation should run next:

```text
Perturb(segmented node, opaque impulse)
    → one ephemeral executable Space
    → nil means the base case
```

The returned Space privately carries the selected segmented material, its
grouping, and the behavior governing that grouping. The durable graph may
contain more edges; the continuation contains only what this perturbation
reached under its heat, phase, allocation, and affordability policy.

All cognitive arithmetic remains space-owned:

```text
field propagation
heat allocation
affordability
frontier ordering
phase rules
```

The kernel only preserves lineage, visited-once behavior, descent order, and
stack order. It never receives or interprets a child collection.

---

## 3. The Slice is the authoritative perturbation

The Slice records the complete segmented traversal as an immutable transcript of
the ephemeral Spaces:

```text
which grouped segmented members were reached
under which ephemeral Space boundaries
in which perturbation
with which space-owned amplitude/evidence phase
under which worker CreationPhase
with which BehaviorID and version
```

That record is authoritative for the statement:

> This is what the traversal was during perturbation.

Observation does not mutate the Slice into the concrete result and does not
remove branches from it. A branch that later returns nothing was still genuinely
perturbed.

```text
AUTHORITATIVE SLICE
    complete segmented traversal
        ↓ post-order Observe
DERIVED CONCRETE RETURN
    bounded realization of that traversal
```

The Slice preserves the superposition. Observe collapses it for this use.
Collapse is selection and realization, not truth or deletion.

---

## 4. Observe is a post-order concrete fold

When `Perturb` returns nil, descent stops. The existing call stack then unwinds.
No second walk is required.

At each visited Space, `Observe` receives:

- the Space that was perturbed;
- the original opaque impulse;
- the one realized Space returned by executing its continuation.

It returns:

- its concrete counterpart;
- one concrete Space containing the selected realization; or
- nothing.

Illustrative contract:

```go
type Space interface {
    Perturb(origin Space, impulse any) Space
    Observe(origin Space, impulse any, realized Space) Space
}
```

The required dataflow is:

```text
ephemeral executable Space
    ↓ execute
one realized Space
    ↑
Observe(current Space, realized Space)
    ↑
one concrete Space or nil
```

The concrete result flows back up the stack. The kernel does not independently
reconstruct it from the Slice.

---

## 5. The two trees have different authority

```text
PERTURBATION TREE
    segmented
    complete for the wave
    stored on the Slice
    records everything reached

OBSERVED TREE
    concrete
    post-order return value
    may be cut or compressed
    records what was realized
```

An observed Space may return nothing. A grouping Space may omit low-leverage
members, compose several internal results, or return nothing itself. None of
those choices rewrites the perturbation transcript.

If the exact observed tree grounds an inference call, it is retained in the
payload log. That log records the concrete collapse; the Slice records the
segmented traversal from which it was collapsed.

Both records name the `CreationPhase`. The payload log must make the visibility
law auditable: the count of admitted prior concrete records matching its creation
type is always zero.

---

## 6. Direction is structural, not reverse causality

The domain's causal direction remains:

```text
Observation → Belief → Action
```

The return direction is:

```text
ConcreteAction → ConcreteBelief → ConcreteObservation
```

The second line describes stack unwinding and value composition. It does not
claim that an Action causes its grounding Belief backward in time or that a
Belief manufactures its source Observation.

```text
causal direction        Observation → Belief → Action
perturb direction       segmented Observation → Belief → Action
fold direction          concrete Action → Belief → Observation
```

The distinction is load-bearing: reverse return flow is how a forward traversal
is concretely assembled.

---

## 6a. Creation phase controls concrete visibility

Every worker carries the type it is creating as a `CreationPhase` for the whole
traversal:

```go
type CreationPhase uint8

const (
    BeliefPhase CreationPhase = iota
    ActionPhase
)
```

The name is deliberately specific. `CreationPhase` is not the field amplitude's
observed/projected evidence phase. It is the worker's epistemic phase: which
kind of durable thing this wave is allowed to create.

The governing law is:

> **A worker creating a type may perturb prior instances of that type as
> segmented routing structure, but may not observe their concrete
> representations.**

This makes traversal visibility phase-dependent:

| Creation phase | Concrete Observations | Concrete Beliefs | Concrete Actions | Concrete Expectations |
|---|---:|---:|---:|---:|
| `BeliefPhase` | visible | hidden | visible when completed | **hidden** |
| `ActionPhase` | visible | visible | hidden | **hidden** |

Therefore:

```text
BELIEF PRESENT = Observations + completed Actions
ACTION PRESENT = Beliefs + Observations
```

### Expectations are a phase of the Action layer — DECIDED 2026-07-24

An Expectation is **belief-like in content** — `doc.go`'s four causes put the *state change*
(final cause) between observation and action, and an Expectation is that final cause stated as
a falsifiable address. That ontology is real and is not what the layer encodes.

**The layer encodes creation and traversal structure, not ontology.** By creation, an
Expectation is inseparable from its act:

> LAW 5 — *every act appears in at least one expect, and every expect carries a `within`, or
> the emission is rejected.*

An act without an expect is rejected; an expect without acts has no `By`. They are **one
emission with two faces** — what I do, and what I claim results. So the Expectation is a phase
of the Action layer:

```text
L0  tokenized observations
L1  tokenized beliefs
L2  tokenized actions        ← two faces: the ACT, and its EXPECTATION
```

**Why this placement and not the alternatives:**

```text
L1-phase (belief phase)   ActionPhase would create an L1 record while needing to SEE L1
                          records — the audit in §5 then needs sub-type discrimination
                          inside L1. Rejected.

own field (L3)            clean, but adds a type and misorders the causal chain: an
                          expectation is not "after" an action, it is emitted WITH it.
                          Rejected.

L2-phase                  CreationPhase maps 1:1 to layers. BeliefPhase → L1,
                          ActionPhase → L2 (acts + expects). §5's audit stays trivial:
                          "zero prior concrete L2 in an ActionPhase payload." CHOSEN.
```

### The uniform visibility rule

With expectations on L2, the law needs no carve-outs:

> **Same-kind hidden; other-kind visible when done.**

- `ActionPhase` hides **all** prior concrete L2 — acts and expectations alike. Same-kind, one
  rule, no sub-type branch. This is what prevents the action worker becoming an imitation
  engine for either its prior acts *or* its prior bets.
- `BeliefPhase` hides L1 (same-kind) and sees L2 acts **when completed** — historical context,
  never backward credit.

### Expectations are hidden from BeliefPhase even when settled

**Open expectations** are hidden by LAW 6 — the belief call must not see the open bets, their
confidences, or the target, or it resolves ambiguity toward the answer it wants.

**Settled expectations are also hidden**, which is the less obvious half:

```text
a settled bet adds nothing the settling OBSERVATION does not already carry
showing it hands the belief call a prediction AND its confidence — the exact desire leak
calibration is supposed to arrive as AGGREGATED precedent (edge P, N, base rates),
    never as a readable prior bet
```

So the Expectation column is uniformly **hidden**, and the property that follows is worth
stating as a law in its own right:

> **No inference call ever reads a concrete Expectation.** It is written by the action call,
> and read only by the settler and the grader — both deterministic, neither a model.

That makes LAW 4 (*"the model cannot touch the grading path at all"*) structural rather than
remembered.

**Revisitable.** The settled-expectations ruling rests on the claim that a settled bet is
informationally redundant against its settling observation. If a case appears where the belief
call demonstrably needs the prior claim — not its aggregate statistics — this cell is the one
to reopen, and only this cell.

The same-kind node is not removed from perturbation. On ascent it becomes
transparent:

```text
same kind as CreationPhase
    perturb normally
    do not realize its concrete counterpart
    preserve eligible grouped results in the realized Space
```

Example:

```text
Segmented prior Belief
    ├── Segmented Observation
    └── Segmented completed Action

BeliefPhase Observe
    Concrete prior Belief       HIDDEN
    Concrete Observation        returned upward
    Concrete completed Action   returned upward
```

The prior Belief can conduct heat toward its grounding without becoming evidence
for another Belief merely because it already exists.

Likewise, an Action worker may traverse prior segmented Actions to reach their
grounding Beliefs and resulting Observations, but the prior concrete Actions
themselves cannot enter the action-inference payload.

The rule applies to durable records that existed before the inference call. A
set of outputs emitted by one call is one creation batch; none of those outputs
may be fed back as grounding during that same call.

This phase boundary prevents:

```text
Beliefs justifying new Beliefs
Actions imitating prior Actions
same-kind records becoming self-authenticating context
```

while preserving the learned routing value of the complete segmented field.

---

## 7. Relationship to inference payloads

The creation phase determines which concrete tree ascent may assemble.

```text
BELIEF PHASE
    segmented Beliefs route but remain concretely transparent
    concrete Observations and completed Actions return upward
    root assembles the belief-inference present

ACTION PHASE
    segmented Actions route but remain concretely transparent
    concrete Beliefs and Observations return upward
    Beliefs become grounding headers over surviving Observation detail
    root assembles the action-inference present
```

Within the phase-legal types, `Observe` may still preserve only the
highest-leverage concrete details or nothing. A transparent same-kind member
omits its own concrete representation while preserving eligible grouped results
inside the realized Space.

Thus the ascent simultaneously:

1. collapses the perturbed superposition;
2. realizes concrete counterparts;
3. cuts the result to the available context;
4. builds the exact inference payload.

The prompt is not gathered by querying the Slice again after traversal. It is the
concrete value produced by observing the perturbed stack.

---

## 8. Kernel pseudocode

```go
func execute(current Space, origin Space, impulse any, slice *Slice) Space {
    next := current.Perturb(origin, impulse)
    slice.RecordPerturbation(current, next)

    var realized Space
    if next != nil {
        realized = execute(next, origin, impulse, slice)
    }

    return current.Observe(origin, impulse, realized)
}
```

This pseudocode states structure, not concurrency or storage mechanics. A
grouping Space owns its members and their sequence, fan-out, join, or cut
behavior. It may return further continuation Spaces while the kernel continues
to see one executable value at a time.

---

## 9. Interface laws

1. `Perturb` returns exactly one executable ephemeral Space or nil.
2. Nil from `Perturb` is the descent base case.
3. The returned Space owns its grouped members and behavior.
4. The Slice records every executable Space before it disappears.
5. `Observe` receives exactly one realized Space or nil.
6. `Observe` returns exactly one concrete Space or nil.
7. Omitting a concrete result never deletes its segmented perturbation record.
8. The kernel performs no cognitive arithmetic.
9. The kernel does not rebuild the concrete result from the Slice.
10. Reverse return order must never be interpreted as reverse causal authority.
11. The worker carries one `CreationPhase` for the whole traversal.
12. Same-kind prior nodes remain perturbable but are concretely transparent
    during `Observe`.
13. A call's new outputs cannot re-enter the grounding of that same creation
    batch.

---

## 10. One-paragraph statement

Skyra perturbs only the segmented Observation–Belief–Action field. Each
`Perturb` returns one ephemeral Space containing grouped data and the behavior
governing its execution. The Slice durably records that continuation before it
disappears. When descent reaches nil, the same stack unwinds and `Observe`
receives one realized Space, applies creation-phase visibility, and returns one
concrete Space or nil. That result is the bounded, phase-legal collapse of the
authoritative perturbation transcript and may directly become the logged
inference payload. The kernel only executes Spaces and does no cognitive math.
