# Observable Causality

**Date:** 2026-08-13
**Status: WORKING NOTE.** This states what Skyra can prove about an accepted computation, what it
can test, and which questions it deliberately refuses because their answers live inside an entity.

> **Skyra does not ask an entity to explain what caused its answer. The substrate controls the
> conditions, records the choices, observes the outputs, and retains the causal cone accepted at
> the outer boundary.**

---

## The correction

Ordinary model attribution begins after the answer. A model produces a result and is then asked
which sources influenced it. That account is self-reported, may be incomplete, and cannot establish
what the model actually received.

Skyra has a different object. Causality is established at invocation, before the child runs:

```text
PARENT DECIDES
    which engram will run
    which entity will interpret it
    which observations it will receive
    which visible observation prompted the choice
    which reach and grants travel downward
            ↓
SUBSTRATE CONSTRUCTS THE FRAME
            ↓
ENTITY INTERPRETS EXACTLY THAT FRAME
            ↓
PARENT REMEMBERS OR DOES NOT REMEMBER THE RESULT
            ↓
OUTERMOST REMEMBER RETAINS ONE CAUSAL CONE
```

The entity does not author that account. It makes choices and emits outputs inside conditions the
substrate created.

---

## What the record proves

For one invocation, the substrate can prove:

```text
this expectation
interpreted by this entity
received these observations
under this parent decision
and produced this emission
```

For an accepted branch, it can additionally prove:

```text
these child invocations occurred
in this causal and emitted order
these results were remembered by their parents
this result was remembered at the outer boundary
this backward-reachable causal cone became durable
```

This is not a model's explanation of its reasoning. It is the execution record of an isolated
computational system.

---

## 1 · Influence is observable exposure

If an observation was in the frame, it influenced the computation in the sense available to the
runtime: it was part of the entity's actual input.

```text
IN FRAME       participated in the execution conditions.
NOT IN FRAME   could not participate through the observation channel.
```

The question *by how much did this observation alter hidden activations?* belongs inside the model.
Only the model's builder may be able to measure it, and Skyra cannot recover it from outside.

That does not weaken the causal record. It defines its proper boundary:

```text
SKYRA RECORDS     exposure · choice · order · output · remembrance · acceptance
SKYRA IGNORES     attention weight · hidden activation · private chain of thought
```

The first set is observable and structurally enforceable. The second is inaccessible and is not
required by the mechanism.

---

## 2 · Necessity is experimentally testable

One invocation identifies the candidate causal set. Because the exact expectation, entity, and
frame were recorded, necessity can be tested by intervention.

```text
E × N × {A, B, C}  →  R     accepted

E × N × {A, C}     →  R₁
E × N × {A, B}     →  R₂
E × N × {B, C}     →  R₃
```

Removing one observation and re-running asks an observable question:

```text
WITHOUT B
    did the entity make the same decisions?
    did it return the same material or shape?
    did the branch still close?
    did the outer perspective still remember the result?
```

If removing `B` prevents an acceptable result under the tested conditions, `B` was necessary to
that result in those conditions.

The original record does not need to contain a model-authored necessity claim. It preserves the
conditions required to perform the test.

Necessity is therefore neither guessed from prose nor read from attention. It is approached by
controlled replay and ablation over recorded invocations.

---

## 3 · Understanding is not a system property

*Did the entity understand the observation?* has no stable structural answer and no role in the
runtime.

Skyra can record:

```text
WHAT IT RECEIVED
WHAT IT SELECTED
WHAT IT RETURNED
WHAT ITS PARENT REMEMBERED
HOW FAR THAT RESULT SURVIVED
WHETHER THE OUTER PERSPECTIVE ACCEPTED IT
```

If the result was accepted, the entity's interpretation was sufficient for that local purpose.
Nothing is gained by adding an unverifiable internal condition called understanding.

This is the same refusal that removed entity capability declarations. The system records what
happened when an entity was tried; it does not ask the entity what it believes it can do or whether
it believes it understood.

---

## 4 · Correctness is local to acceptance

The graph does not certify one global truth. It witnesses the verdict rendered by a perspective.

```text
CHILD RESULT REMEMBERED BY PARENT    correct enough for that parent, in that invocation.
RESULT SURVIVES SEVERAL HOPS         accepted at causal distance.
RESULT REMEMBERED BY THE USER        correct for that namespace at that event.
```

The outermost caller is the final authority for its branch. If the user remembers a result, the
system can prove that the user accepted it under the recorded conditions.

If the user accepts something factually wrong, the record remains truthful:

> **This principal accepted this result, produced by this causal cone, at this time.**

A later world observation may contradict it. A later invocation may produce a different result. A
different parent may refuse it. Those are new local verdicts, not evidence that the earlier graph
lied about what its owner accepted.

The terminus limit follows cleanly: a closed causal system cannot contradict its terminal authority
without a new observation from outside. That is not an inability to record correctness. It is the
boundary of whose correctness the record can witness.

---

## 5 · There is no uncited observational influence

An observation can affect an invocation only by being placed in its frame.

```text
PARENT HOLDS A SET OF OBSERVATIONS
            ↓ chooses a subset
SUBSTRATE BINDS THAT SUBSET TO THE CHILD
            ↓
CHILD CAN OBSERVE THAT SUBSET AND NOTHING ELSE FROM THE GRAPH
```

An observation outside the frame is categorically outside that invocation's observation channel.
It is not an uncited influence. It is absent.

This depends on the entity boundary holding:

```text
NO BACK-CHANNEL
    everything the entity acts on for this invocation arrived through the rendering and delivery
    the substrate supplied, or through world handles explicitly present in that delivery.
```

Hidden continuity, ambient graph access, or undeclared context would not be a subtle attribution
problem. It would be a violation of the entity contract because the frame would no longer describe
the computation's inputs.

---

## 6 · Citation begins at invocation

Citation is not primarily a reference the child writes into its answer. The causal statement is
made when the parent invokes the child.

```text
DECIDE / SELECT
    ENGRAM          what behavior to run
    ENTITY          who interprets it
    OBSERVATIONS    what it may reason over
    CITE            which visible observation prompted this choice
    ACCESS          what authority is provisioned downward
```

The observations bound as parameters isolate the child's causal input. The per-member cite records
why the parent reached for that child. Both are parent choices recorded by the substrate before the
child runs.

The other edge occurs on the way back:

```text
INVOCATION EDGE      these observations were supplied as causes of this computation.

REMEMBRANCE EDGE     this returned observation was retained by the receiving perspective.
```

The invocation edge establishes the input conditions. The remembrance edge establishes which
result entered the parent's live causal history and, eventually, which result survived outer
acceptance.

The child cannot rewrite either fact after the event:

- It cannot add an observation that was not placed in its frame.
- It cannot erase an observation the substrate placed there.
- It cannot decide whether its parent remembers the return.
- It cannot decide whether the outermost caller remembers the branch.

`skyra-dna-v5` derives citations from `Ref`s carried in committed content. That is an earlier
approximation of retained data flow. The proposed live causal model is stronger: invocation,
selection, ordering, translation, revision, and remembrance are substrate-held causal events.

The model may emit the choices that request some of those events. It does not author the durable
account of whether they occurred or were retained.

---

## Two claims that must remain separate

```text
CAUSAL PARTICIPATION
    observation B was supplied to this invocation and therefore participated in its conditions.

COUNTERFACTUAL NECESSITY
    without observation B, no result accepted by this perspective was produced under the tested
    conditions.
```

The first follows directly from the frame record. The second is established through intervention.
Neither requires access to hidden reasoning.

Likewise:

```text
LOCAL CORRECTNESS
    this perspective accepted the result.

GLOBAL OR TIMELESS TRUTH
    the result can never be contradicted by another perspective or later observation.
```

Skyra records the first. It does not claim the second exists.

---

## What the substrate must preserve

For these claims to hold, the live and durable records must preserve:

```text
EXPECTATION          the program interpreted.
ENTITY               the interpreter's identity.
FRAME MEMBERSHIP     the exact observations supplied.
SELECTION CAUSE      the visible observation prompting each invocation.
PARENTAGE            who opened whom.
SEQUENCE             order within one possibility.
EMITTED POSITION     deterministic position inside a concurrent group.
REMEMBRANCE          which returned observation each parent retained.
REVISION             which draft governed each invocation.
OUTER ACCEPTANCE     which causal cone became durable.
```

And the entity boundary must preserve:

```text
ISOLATION            no graph access outside the frame.
NO HIDDEN CONTINUITY no private memory of this caller between invocations.
OPAQUE CROSSING      the kernel may hold delivery artifacts without interpreting their meaning.
LOUD FAILURE         missing reach or unsupported projection cannot masquerade as successful work.
```

Without isolation, out-of-frame influence becomes possible. Without exact frame membership,
necessity cannot be tested. Without remembrance edges, exposure and retention collapse into one
ambiguous signal. Without outer acceptance, there is no subject for the final verdict.

---

## The claim, precisely

> **Skyra can prove the observable causal conditions of an accepted computation: which expectation
> was interpreted, by which entity, with exactly which observations in scope; which children were
> selected with which inputs; which results were remembered; and which causal cone the outer
> perspective accepted.**

From that record it can:

```text
ESTABLISH EXPOSURE       structurally, from frame membership.
EXCLUDE OTHER INPUTS     categorically, through isolation.
TEST NECESSITY           experimentally, through replay and ablation.
RECORD CORRECTNESS       locally, as a perspective's acceptance.
MEASURE VALIDATION       by causal distance through successive remembrances.
IGNORE INTERNAL FICTIONS attention · understanding · private explanations of why.
```

This is stronger than source attribution and narrower than omniscience. It is experimental
causality over an isolated computational system, grounded by the substrate rather than reported by
the entity.
