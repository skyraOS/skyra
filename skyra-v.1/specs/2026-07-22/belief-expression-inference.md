# Belief and Expression Inference — Two Presents, Two Rights

**Date:** 2026-07-23  
**Status:** CORE DECISION. The information boundaries and ambiguity rights are
settled. Exact rendering, variance arithmetic, and budget allocation remain open.

**Reads on:** [tokenizer](tokenizer.md) for the stable sensory alphabet,
[v4-data-model](v4-data-model.md) for the observation/belief/action spine, and
[workers](workers.md) for the earlier worker inventory this decision refines.

---

## 1. The two calls

There are two inference calls with different inputs, outputs, and epistemic
rights:

```text
heated Observations + completed Actions
    ↓ BELIEF INFERENCE
Beliefs
heated Beliefs + grounding Observations
    ↓ EXPRESSION / ACTION INFERENCE
Engage Actions + Expectations
```

They are not two stages of one prompt. They must remain separate because the
belief call interprets what happened, while the expression call decides what to
do.

```text
BELIEF INFERENCE       observations/actions → beliefs
EXPRESSION INFERENCE   beliefs/observations → actions/expectations
```

The belief call is desire-blind. The expression call is allowed to be
goal-directed.

### The creator cannot see prior creations of its own type

Both calls obey one traversal law:

> **The worker responsible for creating a type cannot see previously created
> concrete instances of that type.**

```text
BELIEF WORKER
    sees      Observations + completed Actions
    creates   Beliefs
    hides     prior concrete Beliefs

ACTION WORKER
    sees      Beliefs + Observations
    creates   Actions
    hides     prior concrete Actions
```

This is a concrete-visibility rule, not a traversal prohibition. Same-kind
segmented nodes may still conduct heat and route `Perturb` toward their grounding
or consequences. During `Observe`, their concrete counterparts are transparent
and only eligible grouped results remain in the realized Space.

The worker carries this right as its
[`CreationPhase`](perturb-observe-traversal.md#6a-creation-phase-controls-concrete-visibility)
through the entire traversal.

### Position in the concrete/segmented cycle

These calls cross fields between durable concrete forms; they do not erase the
distinction between a whole record and its segmented token representation:

```text
ConcreteObservation → SegmentedObservation
SegmentedObservation → ConcreteBeliefs       // belief inference

ConcreteBeliefs → SegmentedBeliefs
SegmentedBeliefs → ConcreteActions           // expression inference

ConcreteActions → SegmentedActions
SegmentedActions → ConcreteObservations      // execution + world perception
```

The `Concrete → Segmented` transitions are deterministic decomposition. The
`Segmented → Concrete` transitions contextualize and collapse field material
into a new bounded record or set of records.

The complete ordered token sequence remains on the concrete parent. Segmented
tokens retain parent and position, so traversal may address them independently
without losing the boundary of the Observation, Belief, or Action they came
from. See [tokenizer](tokenizer.md#the-concretesegmented-cycle) for the
representation contract.

---

## 2. Core ambiguity right

When belief inference encounters ambiguity, the model has exactly three legal
responses:

```text
RESOLVE      choose the reading supported by the available context
HOLD         preserve multiple readings or entity candidates
DISREGARD    emit no belief from this ambiguity
```

This right is core. The substrate supplies candidates and evidence; it does not
force every ambiguity to collapse.

### Resolve

The model emits one belief naming the selected reading or entity. It cites the
Observation token spans and Actions that licensed the resolution.

Resolution is provisional interpretation, not truth. The belief still earns or
loses support through its own independently addressable evidence.

### Hold

The model emits multiple competing beliefs, or one explicit unresolved candidate
set, when the present does not discriminate.

```text
Observation span "Mike"
    ├── candidate entity:mike-1
    └── candidate entity:mike-2
```

Holding ambiguity is a successful inference outcome. It is not a parser failure
and must not be coerced into a single entity merely to simplify downstream work.

The expression call can see the held fork and choose an information-producing
action such as asking, perceiving, or waiting for discriminating evidence.

### Disregard

The model emits no belief for the ambiguous material because it is irrelevant,
unsupported, or not worth the representational cost.

Disregard never deletes the Observation or Action. The authoritative token
sequence and event remain durable and may be reconsidered when a later
perturbation heats the region differently.

```text
disregarded now ≠ erased
disregarded now ≠ false
```

### The authority boundary

The model has the right to choose the representational response. It does not have
the right to declare the chosen belief true.

```text
model authority       resolve · hold · disregard
world authority       settle · defeat · leave unknown
```

No ambiguity is resolved at grading time.

### Entity-resolution mechanism is deferred

The resolve/hold/disregard right is settled, but a dedicated entity-resolution
mechanism is not.

V0 does not add:

- an entity-resolution field layer;
- a durable token-to-entity map;
- an alias registry;
- a typed candidate-resolution store;
- a reusable hard-link created from one inference call.

Belief inference receives heated token spans directly. It may emit provisional
belief content that refers to an entity, preserve plural readings, or disregard
the material, but the runtime does not yet promote that output into a separate
entity-resolution substrate.

This deferral does not apply to typed world addresses. Those remain exact and
world-given.

---

## 3. Belief-inference present

The belief inference call sees the heated region's Observations and completed
Actions.

```text
HEAT selects the region.

Within that region:
    Observations provide what crossed the boundary.
    Actions provide what occurred around or produced those observations.
```

The call may receive:

- ordered token sequences from heated Observations;
- exact Observation IDs and token-span addresses;
- completed historical Actions in the heated region;
- whether each Action was MINE or THE WORLD'S;
- exact entity candidates produced by token/alias indexes;
- relevant addresses, primitives, source, count, and wall stamps.

The call must not receive:

- previously created concrete Beliefs, including held interpretations;
- the current target;
- open Expectations;
- prediction confidences;
- the answer planning wants;
- pending or merely proposed Actions;
- downstream action success presented as evidence that an interpretation was
  correct.

Completed Actions are visible as historical context, not as backward credit.
They remain explicitly typed MINE or THE WORLD'S so the model cannot confuse what
the being did with what the world reported.

This is the desire-blind boundary:

> Belief inference may see what happened and what was done. It may not see what
> answer would satisfy the current plan.

---

## 4. Belief-inference output

The output is a set of Beliefs tied to exact evidence:

```go
type BeliefProposal struct {
    Triples    []Triple
    Evidence   []TokenSpan
    Actions    []ActID
    Confidence float64
}
```

The shape above is illustrative; the v4 struct pass owns the final fields.

Required properties:

1. every resolved reading cites the token spans that licensed it;
2. referenced Actions must be completed historical events;
3. competing readings remain independently identifiable;
4. a held ambiguity is representable without inventing a winner;
5. disregarded material produces no Belief but remains in the authoritative
   perturbation and Observation stores;
6. the output cannot settle itself;
7. the output cannot mutate the source Observation or Action.

Belief inference is compression over the heated observation/action region. The
source material remains reachable.

Prior segmented Beliefs may have routed the traversal to that region. Their
concrete propositions do not enter the call and cannot justify the new Beliefs.

---

## 5. Expression-inference present

Expression inference uses heated Beliefs as grounding headers. Beneath each
header it receives the highest-variance Observations in that Belief's region,
including observations produced after prior Actions. Prior concrete Actions are
not shown.

```text
BELIEF HEADER
    the grounded claim or held ambiguity
    its heat, support, and settlement state

    HIGH-VARIANCE OBSERVATIONS
        exact token spans that remain surprising, disputed, sparse, or
        poorly explained
```

The present therefore has two axes:

```text
HEAT       which Belief regions matter now
VARIANCE   which raw Observations beneath them are worth showing
```

Heat without variance repeatedly shows the model what the system already knows.
Variance without heat spends context on uncertainty unrelated to the current
situation. The expression present uses both:

> Heat selects the grounding headers. Variance selects the detail placed under
> them.

### The ascent builds the prompt

The expression present is not assembled by a separate query after traversal.
Post-order ascent is the assembly.

This is the `Observe` phase of
[Perturb and Observe](perturb-observe-traversal.md). `Perturb` traversed only
segmented field material on descent. During stack unwind, `Observe` receives the
concrete results returned by its children and realizes the concrete counterpart
of the current segmented node. The old name `Express` no longer names this
ascent phase.

The perturb phase has already built the authoritative heated tree on the Slice.
It may contain segmented prior Actions for routing, but `ActionPhase` prevents
their concrete realization:

```text
Belief
    ├── Observation candidates
    ├── Observation candidates
    └── segmented prior Action
            └── resulting Observation candidates
```

During ascent, the Observation layer returns only its highest-leverage members
to the owning Belief node. Prior Action nodes are transparent and pass eligible
Observation results upward. "Leverage" is the selection quantity that
combines the value of exposing this unresolved detail with its cost; its exact
formula remains open (§10).

```go
type PromptDetail struct {
    SourceID ObservationID
    Leverage float64
    Cost     int
    Payload  TokenSequence
}
```

The type above is illustrative. The invariant is the return shape:

```text
Observation Observe on ascent
    → highest-leverage detail, or NOTHING
```

The Belief node receives the surviving child details:

```text
NO surviving Observations
    → return NOTHING
    → do not add this Belief to the prompt payload

ONE OR MORE surviving Observations
    → add the Belief as the grounding header
    → place the surviving details beneath it
    → return the completed prompt block upward
```

Candidate fold:

```go
// Private behavior inside a BeliefAssemblySpace. The kernel never receives
// this detail collection.
func observeBelief(b Belief, children []PromptDetail) *PromptBlock {
    selected := highestLeverage(children)
    if len(selected) == 0 {
        return nil
    }
    return &PromptBlock{
        BeliefID: b.ID,
        Header:   renderBelief(b),
        Details:  selected,
    }
}
```

The Belief header is therefore conditional. Heat makes the region traversable;
surviving leverage makes it visible to expression inference.

The root receives completed `PromptBlock`s and renders them into the payload log.
Rendering may put the header before its details even though the details returned
first in post-order. The structured block is authoritative; string order is a
final presentation operation.

```text
post-order return                     rendered prompt

details ──▶ belief fold               [BELIEF header]
                 │                      observation detail
                 └──▶ PromptBlock
```

This removes a separate gather/sort/render pass. The prompt is the value returned
by the observed tree.

### The ascent collapses superposition

The perturbation tree is the live superposition of everything the event reached:
all heated Belief regions, all eligible Observation detail beneath them, and any
concretely hidden prior Action routing nodes. It is complete and authoritative
for that perturbation.

The ascent collapses that superposition into one bounded expression present:

```text
AUTHORITATIVE PERTURBATION TREE
    all reached candidates remain on the Slice
        ↓ post-order leverage cut
OBSERVED PROMPT TREE
    only grounding headers with surviving detail
        ↓ render
EXPRESSION INFERENCE PAYLOAD
```

"Collapse" here is an attention operation, not a truth operation.

- an omitted branch is not false;
- an omitted Observation is not erased;
- an omitted Action did not fail to occur;
- an omitted Belief is not defeated;
- a different later perturbation may collapse the same durable region
  differently.

The Slice retains what was traversed. The ascent records what this inference call
was allowed to see.

### The payload log is part of the experiment

The completed prompt tree must remain inspectable as the exact basis of the
expression call:

```go
type PromptBlock struct {
    BeliefID BeliefID
    Header   TokenSequence
    Details  []PromptDetail
}

type ExpressionPayload struct {
    PerturbationID string
    Blocks         []PromptBlock
    Budget         int
    Spent          int
}
```

Again, final structs remain open. What is settled is that the payload records:

- which Beliefs became grounding headers;
- which Observations survived beneath each;
- their leverage and cost at the cut;
- the perturbation tree from which they ascended;
- the exact bounded payload shown to inference.

Without that record, the system cannot later explain why an Action was proposed
or distribute credit over the actual grounding frame.

### Candidate rendering

```text
[BELIEF b17 heat=.82 support=9 settled=mixed]
Michael requested water.

  [OBSERVATIONS · highest variance]
  o41[3:7]  "grab my water"
  o44[0:5]  "no, I wanted still"
```

The Belief header tells the model what structure currently grounds the region.
The children expose where that structure is weakest or where an action can gain
information.

---

## 6. Expression-inference rights

Expression inference may:

- act from a resolved Belief;
- ask or perceive to discriminate a held ambiguity;
- decline to act when the grounding is insufficient;
- produce an Action that deliberately gathers information;
- emit independently falsifiable Expectations for every Action;
- use the observed consequences and grounded Beliefs reached through prior
  Action routes.

It may not:

- receive a prior concrete Action or Expression as an example to imitate;
- rewrite the grounding Beliefs;
- collapse a held ambiguity merely by acting as if one branch were true;
- treat its own proposed Action as an Observation;
- omit falsification addresses for the Actions it emits;
- turn high variance into falsehood merely because the region is uncertain.

If expression discovers that the grounding is inadequate, its legitimate output
is an information-producing Action, not an unrecorded reinterpretation.

---

## 7. Prior Actions are visible only to belief inference

The same-kind exclusion produces an asymmetric boundary:

```text
BELIEF INFERENCE
    completed prior Actions are visible historical context

EXPRESSION / ACTION INFERENCE
    prior concrete Actions are hidden
    their grounding Beliefs and resulting Observations may be visible
```

An Action can still conduct heat during an action-phase perturbation. Its
segmented representation may route traversal toward what was believed before it
and what the world reported afterward. During ascent the Action node is
transparent: those eligible concrete Beliefs and Observations continue upward,
but the prior concrete Action does not.

This allows consequences to inform action creation without turning the action
worker into an imitation engine.

Pending and proposed Actions are not completed history and never enter the
belief present as though they happened. Newly emitted Actions also cannot be fed
back into the same creation batch.

---

## 8. Relationship to perturb and observe

The authoritative perturbation tree records everything reached by the event.
Post-order observation cuts that tree into the heated region available to
cognition.

The two inference presents are projections over that observed region. For
expression inference, the projection is constructed directly by ascent:

```text
AUTHORITATIVE PERTURBATION TREE
    ↓ ActionPhase hides prior Actions
    ↓ post-order Observation leverage cut
BELIEF HEADERS WITH SURVIVING DETAIL
    ↓ root render of returned PromptBlocks
EXPRESSION INFERENCE PRESENT
```

Disregarding an ambiguity changes only the belief projection. It does not cut the
source event out of the authoritative perturbation tree.

---

## 9. Measurements

The boundary should make these quantities observable:

```text
ambiguities resolved / held / disregarded
held ambiguities later discriminated
resolved ambiguities later defeated
belief-call tokens spent per admitted Belief
expression-call tokens spent per settled Expectation
heat admitted per grounding header
variance admitted per Observation detail
information-producing Actions / total Actions
same-kind prior concrete records admitted / call    MUST BE 0
```

A rising forced-resolution rate is a warning that the representation cannot hold
ambiguity. A rising disregard rate with stable later performance may mean the
belief call is learning not to spend representation on noise.

---

## 10. What remains open

1. The exact variance measure for Observations in the action present.
2. Whether heat and variance receive fixed context shares or compete through one
   affordability calculation.
3. The wire format for `resolve`, `hold`, and `disregard`.
4. Whether a held ambiguity is represented as plural Beliefs or a typed candidate
   set referenced by Beliefs.
5. How much completed Action history may enter belief inference before it becomes
   an indirect desire leak.
6. Whether expression and expectation generation remain one model call or two
   workers over one emission.
7. The exact leverage function for Observation details: variance, diagnosticity,
   expected information gain, consequence reach, cost, or a measured
   combination.
8. Whether leverage selection is entirely local at each Belief node or receives a
   root allocation so Belief regions compete under one context budget.
9. Entity resolution as a dedicated mechanism and data shape. V0 deliberately
   carries unresolved token spans into belief inference instead.

---

## 11. One-paragraph statement

Belief inference sees the heated region's Observations and completed Actions and
has the explicit right to resolve an ambiguity, preserve it, or disregard it.
That right chooses a provisional representation; only later world evidence can
settle it. Prior concrete Beliefs are hidden from that creation call. Expression
inference is separately grounded by heated Beliefs, rendered as headers, with
the highest-variance Observations placed beneath each header. Prior concrete
Actions are hidden even when their segmented nodes routed the traversal. Heat
determines which regions matter; variance determines which unresolved details
are worth paying to show. The belief call remains blind to desire, while the
expression call may act, ask, or perceive from the grounded and explicitly
uncertain structure it receives.
