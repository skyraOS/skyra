# Think — the interface to potential

**Date:** 2026-08-13
**Status: PROPOSAL.** This proposes the semantic interface and system boundary of `Think`. It does
not propose concrete types, storage, transports, drivers, or an implementation order.

This supersedes the earlier seven-peer-method surface now retained only in the carried semantic
notes. Where it conflicts with
[`reach-boundaries-proposal.md`](../reach/reach-boundaries-proposal.md), this document supersedes
that proposal's claim that Reach itself
returns a bounded ranked set. The bounded set belongs to Think; Reach realizes one decided member.

> **Think is the universal interface by which a possibility asks the system what it could do next.
> It returns a small bounded set of expectation-bearing observations. Think creates potential;
> Decide chooses potential; Reach makes the chosen composition live; Remember determines what
> survives.**

---

## The interface

The caller supplies three dimensions:

```text
THINK(
    entity[],
    expectation[],
    observation[],
)
```

The runtime supplies two bounds around every call:

```text
FROM        namespace × requesting engram × runtime
HORIZON     the maximum outward authority available to this thought
```

The result is:

```text
A SMALL RANKED SET OF OBSERVATIONS
    ordinarily no more than five
    each positive observation carries a candidate expectation
    each candidate is potential which may be selected by Decide
```

In one line:

```text
ENTITY[] × EXPECTATION[] × OBSERVATION[]  within FROM × HORIZON
    → BOUNDED RANKED EXPECTATION-OBSERVATIONS
```

There is no `address` operand.

```text
ENTITY          WHO might interpret.
EXPECTATION     WHAT might be interpreted.
OBSERVATION     WITH WHAT, or in response to what need.

ADDRESS         absent. volatile resolution behind the kernel boundary.
```

An address is not semantic potential the caller should bind into an expectation. It is obtained by
composable behavior beneath Think and resolved when Decide attempts to make one candidate live.
Keeping it behind the boundary lets an expectation survive movement without embedding a stale
endpoint in what the behavior is.

---

## The possibility-facing surface

For a possibility, the universal composition is:

```text
THINK       create potential.
DECIDE      select potential and request that it become live.
REMEMBER    retain returned causality.
```

Semantic composition is embodied by composable behavior below Think:

```text
REACH PLANNING       behavior which thinks about composition with an entity.
RECOGNITION          behavior which thinks about present evidence against experience.
DISCOVERY            behavior which constructs the diagnosis that known potential is insufficient.
```

`Orient` is separate: it is a kernel-owned method which returns a bounded structural observation
of the caller's live position. Think may consume that observation, but does not own or implement
Orient.

Actual Reach is different from reach planning. It is the hidden realization phase after Decide has
selected one complete composition. Think may propose how an entity could be reached; it does not
contact it.

`Approach` remains the outside-in admission boundary and is not settled by this proposal. The claim
here is about operations exposed to a possibility already inside the system.

---

## The three caller-supplied dimensions

The arguments are constraints on thought, not a complete invocation and not three mandatory
values.

### `entity[]` — who

The entities about which the caller wants Think to reason.

```text
EMPTY             propose without a caller-chosen interpreter.
ONE               hold this entity fixed.
SEVERAL           resolve only within this candidate set.
```

An entity constrains who may interpret a resulting expectation. It does not say where or how that
entity can be contacted.

### `expectation[]` — what

The existing or proposed behaviors about which the caller wants Think to reason.

```text
EMPTY             propose what behavior might address the observations.
ONE               reason about, refine, translate, or compose this behavior.
SEVERAL           reason only within this candidate behavior set.
```

An expectation supplied here remains behavior-as-material for the duration of thought. Think does
not open it merely because it was supplied.

### `observation[]` — with what and why

The evidence, need, or material which gives this thought meaning.

```text
EMPTY             ask what potential is available under the other constraints and the Horizon.
ONE               reason from this observation.
SEVERAL           reason from their composition, with membership and order preserved.
```

Observations are the only caller-supplied semantic material. They may guide what candidates mean;
they cannot widen Horizon or rewrite From.

### Empty, bounded, and fixed

Each dimension has the same three semantic states:

```text
EMPTY             unresolved. Think may propose freely within From and Horizon.
CANDIDATE SET     constrained. Think may propose only within the supplied set.
SINGLE VALUE      fixed. Think may compose around it but may not substitute it.
```

This is why the operands are sets even when the ordinary call supplies one item. Cardinality says
how much of that dimension remains open without requiring a second constraint language.

The maximally empty call asks:

> **What potential is available from here, for this namespace and requesting engram, within this
> Horizon?**

It is still bounded by substrate-supplied origin and authority. Empty does not mean global.

---

## `From` and `Horizon`

Neither is caller-authored input.

```text
FROM
    NAMESPACE           whose memory and authority is thinking.
    REQUESTING ENGRAM   which durable behavior requested the thought.
    RUNTIME             where "here" is for this thought.

HORIZON
    the outer boundary of entities, namespaces, credentials, hosts, authority domains,
    effects, and resource ceilings this thought may consider or provision downward.
```

The runtime fills `From`. The current possibility receives Horizon from above. Think may reason
within them but cannot alter either one.

```text
FRAME       what the possibility currently knows and can name.
FROM        whose thought this is, which behavior requested it, and where here is.
HORIZON     how far outward this thought may legitimately propose action.
```

Horizon is not a second frame and carries no application values. It can only narrow down a causal
chain:

```text
CHILD HORIZON ⊆ PARENT HORIZON
```

A candidate expectation may require less than the available Horizon. It may never require more.

### Working note: a failed child contract can supply a focused frame

When a child observation returns and fails the parser owned by its parent, the parent may Think
inside a frame focused on that exact dependency. The failure already supplies the relevant semantic
material:

```text
REQUESTING ENGRAM    the parent which owns the failed child contract
EXPECTATION          the return contract its continuation requires
OBSERVATION          the actual child observation which did not satisfy it
ENTITY               fixed only if the correction must preserve the same child entity
HORIZON              inherited and no wider than the failed composition's Horizon
```

This is not a new Think signature. It is a particularly constrained use of the existing dimensions.
The runtime may further limit the available behaviors and observations to those relevant to the
failed dependency.

If a proposed correction encounters a deeper child mismatch, that child contract's parent may
receive another, narrower invocation. Think can therefore participate in progressive causal zoom
without becoming ambient search:

```text
failed return
    ↓ focused Think
candidate corrective behavior
    ↓ Decide → Reach
deeper failed return
    ↓ more focused Think
```

At no point may a repair invocation widen `From`, Horizon, or the observational frame merely because
the current continuation failed.

---

## What Think returns

Think returns observations because its conclusions must compose with the same causal vocabulary as
everything else.

A positive result says, in substance:

```text
Given these entities, expectations, and observations,
from this namespace, requesting engram, and runtime,
within this Horizon,
this expectation is a candidate next behavior.
```

The expectation is inside an observation because it is still potential. Think has not installed,
opened, executed, or made it durable.

### Candidate observations, not executable handles

Each candidate observation must contain enough semantic material for the caller to decide on the
proposed composition. It may identify:

```text
the proposed expectation
the entity held fixed or proposed to interpret it
the observations it is proposed to receive
the narrower Horizon the composition would require
the evidence from which the proposal was formed
```

It does not contain an address, transport, live relationship, or claim that contact has already
been established.

### Bounded plurality

Think ordinarily returns no more than five observations.

```text
ONE RESULT          leaves no local selection when alternatives were available.
UNBOUNDED RESULTS   lets system-wide potential flood one frame.
BOUNDED RESULTS     lets Think rank globally while the caller selects locally.
```

Five is an operating bound, not a metaphysical constant. The invariant is a small ranked set.

### When no candidate exists

“Nothing fits” is a constructed conclusion, not an empty lookup result. Think returns one
composable diagnostic observation rather than silently returning nothing.

The diagnosis distinguishes at least:

```text
NO KNOWN POTENTIAL       nothing considered could satisfy the supplied constraints.
BEYOND YOUR HORIZON      relevant potential exists only outside the authority available here.
```

It must not expose addresses, endpoint failures, or hidden strategy names merely to explain that
conclusion.

---

## Syscalls create potential

Think does not do the proposed work.

```text
THINK       creates bounded semantic potential.
DECIDE      chooses and binds one potential composition.
REACH       attempts to make that exact composition live.
POSSIBILITY the successfully realized composition.
REMEMBER    determines whether its causal result and behavior become durable.
```

Before Decide, a returned expectation is data about possible behavior. After Decide, that same
potential becomes the definition of a requested possibility. It is still not live until Reach
succeeds.

```text
EXPECTATION-OBSERVATION
        ↓ Decide selects it
BOUND COMPOSITION
        ↓ internal Reach realizes it
POSSIBILITY
```

This gives the syscall boundary one rule:

> **A syscall may enlarge the live space of possible next actions. It may not silently turn one of
> them into actual causal work.**

---

## Executing the observation preserves causality

The caller must decide on the candidate observation, not extract and copy its expectation text.

```text
observations A, B
      ↓
Think opens hidden thinking behavior T
      ↓
candidate observation C carrying expectation X
      ↓
parent Decides on C
      ↓
internal Reach realizes X as possibility P
      ↓
P returns observation D
```

The causal path is:

```text
A, B → T → C → Decide → P → D
```

Copying `X` out of `C` and opening it as unrelated text would sever the edge through the thought
that authored or selected it. The system would retain the behavior while losing why it exists.

The load-bearing rule is:

> **When a thought-derived expectation is executed, the observation which proposed it remains a
> constituent of that execution.**

This closes the unattributed-authoring hole: authored expectations no longer appear between two
causal events without an observation naming the behavior which produced them.

---

## Think is one interface over hidden composable behaviors

The top-level possibility sees `Think`. It does not select the behaviors by which Think constructs
its answer.

Conceptually:

```text
think
├── think-reach-entity
├── think-recognize-observation
├── think-discover-novelty
└── further thinking behaviors which earn their place through use
```

These names describe possible internal decompositions, not a fixed required child list.

The subject and supplied constraints let Think select the relevant behavior:

```text
ENTITY FIXED, EXPECTATION OPEN
    think-reach-entity may propose expectations through which this engram could compose with the
    entity.

OBSERVATIONS FIXED, EXPECTATION OPEN
    recognition or discovery behavior may propose expectations applicable to that evidence.

EXPECTATION FIXED, ENTITY OPEN
    interpretation behavior may propose entities capable of making that expectation live.

LIVE EXECUTION AS OBSERVATION
    kernel.Orient first reports bounded live topology; Think may then propose expectations for
    responding to that structural observation.
```

Hidden does not mean absent from causality:

```text
HIDDEN FROM CALLER SELECTION     the caller cannot reach around Think and name its private child.
VISIBLE IN THE CAUSAL CONE       the behavior which produced a candidate remains attributable.
```

This preserves two learning perspectives:

```text
THINK LEARNS ACROSS CALLERS
    which internal behavior produces useful potential for classes of request.

THE CALLER LEARNS LOCALLY
    which candidate expectation serves this engram's particular purpose.
```

---

## Reach behind Think and Decide

`think-reach-entity` does not reach the entity. It performs a provisional runthrough over remembered
experience and the current bounds.

```text
think(entity E)
      ↓
think-reach-entity
      ↓
bounded expectation-observations describing possible compositions with E
```

Those observations may express that a composition is plausible from this `From` and within this
Horizon. They do not prove the entity is reachable now. Current reachability can only be learned by
attempting the chosen composition.

The caller then decides:

```text
candidate observation C
      ↓ Decide
chosen expectation + entity + inputs + narrowed Horizon
      ↓ Reach, behind the kernel boundary
possibility | beyond current reach | beyond your horizon
```

Therefore:

```text
THINK-REACH-ENTITY     understands and proposes potential compositions.
DECIDE                 selects exactly one composition.
REACH                  resolves volatile location and attempts to realize it.
```

Reach cannot substitute another entity or expectation. That would make a different possibility
from the one Decide selected.

---

## The copy-and-replace relationship

Thinking about a new behavior has the semantic shape of taking known behavior as material and
provisionally replacing or refining its expectation.

```text
KNOWN ENGRAM OR EXPECTATION
        ↓ thought, translation, or refinement
CANDIDATE OBSERVATION CARRYING A DRAFT EXPECTATION
```

The original durable engram is not edited by thought. The candidate is not yet an engram. It is a
provisional observation in live causal work.

If Decide executes it and its result is ultimately Remembered, the retained causal cone contains
both the prior behavior and the thought which produced the replacement. If it is not Remembered,
the candidate remains unrealized or provisional work and disappears with the live branch.

```text
THINKING IS FREE        candidate behavior may be proposed without durable mutation.
DECIDING HAS CAUSALITY  one candidate becomes live work if Reach succeeds.
REMEMBERING HAS WEIGHT  only retained behavior joins durable experience.
```

---

## Boundaries

Think may:

```text
interpret the supplied semantic material
read experience made available within the namespace and Horizon
select and compose hidden thinking behaviors
author, translate, refine, and rank candidate expectations
return a small ranked set of attributable observations
```

Think may not:

```text
open a candidate possibility
contact an entity
resolve or expose an address
perform the work described by a candidate expectation
widen Horizon
rewrite From
make provisional behavior durable
return an unbounded candidate population
hide which internal behavior produced a candidate from the causal record
```

---

## Rejected shapes

```text
THINK RETURNS AN ADDRESS
    Address is volatile mechanism behind the kernel boundary. Exposing it freezes WHERE into
    semantic potential and makes callers understand transports.

THINK RETURNS A POSSIBILITY
    That silently combines proposal, selection, and realization. Decide loses its subject.

THINK EXECUTES EACH CANDIDATE TO TEST IT
    A runthrough is provisional reasoning over experience, not contact or application work.
    Executing candidates creates causal work and effects before the caller has selected one.

THE CALLER NAMES THINK'S CHILD
    That exposes kernel decomposition as application vocabulary and prevents Think from learning
    which internal strategy serves which kind of request.

THE CANDIDATE IS ONLY EXPECTATION TEXT
    Copying the text can discard the observation and behavior which produced it. The expectation
    must remain inside a composable observation until Decide.

HORIZON TRAVELS AS AN OBSERVATION
    Authority would become caller-authored data. Horizon is supplied and narrowed by the runtime,
    outside the semantic operands.

AN EMPTY RESULT MEANS NOTHING EXISTS
    Absence cannot distinguish no match, no authority, and failure to think. Think must construct a
    bounded diagnosis.
```

---

## What remains open

```text
the canonical content of an expectation-bearing observation
whether a candidate may propose more than one entity
how a candidate declares the narrower Horizon it would require without turning authority into data
how provisional expectation replacement is represented during live work
how Think distinguishes recognition, refinement, and discovery subjects
whether the ordinary bound is always five or configurable beneath a fixed maximum
which diagnoses besides no known potential and beyond your horizon are necessary
how the maximally empty call is seeded at boot
```

None changes the proposed boundary:

```text
THE CALLER supplies entities, expectations, and observations.
THE RUNTIME supplies From and Horizon.
THINK returns bounded expectation-observations and performs no proposed work.
DECIDE selects an observation and preserves it as a cause.
REACH attempts to make that exact composition live without exposing address.
REMEMBER determines what becomes durable.
```

---

## Acceptance test

The proposal holds only if:

```text
1  Address is absent from the interface and from candidate expectations.
2  Every positive result is an observation carrying potential behavior, not live work.
3  The result set is ranked and bounded, ordinarily to five.
4  Empty, candidate-set, and fixed forms of all three operands have distinct meanings.
5  From is runtime-supplied as namespace × requesting engram × runtime.
6  Horizon is runtime-supplied, separate from the frame, and can only narrow.
7  The caller cannot directly select Think's internal behaviors.
8  Internal thinking behavior remains visible in the causal cone.
9  Think performs no contact, application work, world effect, or durable mutation.
10 Decide on a candidate preserves the candidate observation as a constituent.
11 Reach can fail without changing which composition Decide selected.
12 No-candidate and beyond-Horizon results are explicit composable diagnoses.
13 The same interface can think about a known entity, expectation, observation, or any constrained
   composition of the three.
```

If any one fails, Think has either become a lookup too weak to construct meaning or an executor too
powerful to remain the interface to potential.
