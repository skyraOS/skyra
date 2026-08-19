# Surfaces and solidification

**Date:** 2026-08-14
**Status: PROPOSAL.** Two surfaces over one semantics, when each is used, and how a flow becomes a
continuation. This does not change the primitives or the language; it says where the language
begins and what precedes it.

> **The continuation language cannot create its own child positions. Discovery and origination are
> primitive; composition is the language. A flow solidifies into a continuation when what happened
> is projected as structure.**

---

## The bootstrap gap

[`continuation-language.md`](continuation-language.md) has the runtime supply the frame, including
the `children` block. On an engram that has never run, that block is empty:

```text
TripResult planTrip(TripRequest request, Traveler traveler, Location origin)

children { }          ← nothing has been composed here yet

returns { plan: TravelPlan }
```

With no declared positions the language can express exactly one thing: a return composed from
parameters. There is no syntax for Think or Decide — those occur *beneath* a method call, and a
method call requires a position that already exists.

> **The language consumes structure. Something else must create it.**

This is not only the leaf question. It applies to every engram on its first execution, including
ones that will become deep orchestrators.

## Three acts, two surfaces

```text
DISCOVER    create a position that does not exist yet     PRIMITIVE   Think → Decide → Reach
ORIGINATE   produce content from judgment                 PRIMITIVE   the crossing
COMPOSE     over positions that already exist             LANGUAGE
```

Two of the three acts are primitive. The language covers the one act where a compiled artifact buys
something: a reconstructible dependency graph, parent-owned parsers, and derived invocation order.

Leaf and orchestrator stay structural, as
[`engram-contracts-and-composition.md`](engram-contracts-and-composition.md) requires — *"the same
engram may behave as a leaf along one path and invoke children along another."* Neither becomes a
declared mode.

### The language has no origination syntax

`continuation-language.md` names the danger it must avoid: a leaf construct *"must not allow an
orchestrator to smuggle uncaused values into an otherwise composed result."* That danger is
unavoidable while origination is a language construct, because anything a leaf can write an
orchestrator can write.

So the language has no such construct.

```text
ORCHESTRATOR   has a body. every value traces to a parameter or a child return.
LEAF           has NO body. the entity's interpretation of the expectation produces the observation.
```

A leaf's content is not uncaused. It is caused by the crossing, which is where
[`../semantics/the-crossing.md`](../semantics/the-crossing.md) already puts it: *"the kernel must
HOLD the crossing. It still may not READ it."* Keeping origination outside the language lets
*nothing may appear from nowhere* hold without exception.

## Two modes, not two systems

[`../overnightBuild/stops.md`](../overnightBuild/stops.md) rejected "code entities skip the parse
path" because two classes of entity would break the kernel's hold on what produced a decision. Two
*modes of one engram's life* is a different thing, but only under one condition:

> **The continuation language desugars to exactly the primitives — nothing more, nothing less.**

Primitive mode is then the same thing written longhand, not an escape hatch. The test is
mechanical and worth keeping as an acceptance criterion:

```text
EVERY LANGUAGE CONSTRUCT      has a primitive sequence it compiles to.
EVERY PRIMITIVE SEQUENCE      has a language form, or is DISCOVER / ORIGINATE and deliberately has none.
```

If the language can express something the primitives cannot, or a composition has no surface form,
there are two systems wearing one name.

## Solidification

```text
RUN 1        primitive. exploratory. Think → Decide → Reach → Remember.
REMEMBER     the retained cone is projected — and its shape IS the continuation.
RUN 2+       the compiled continuation replays. Think only fills positions.
```

The retained cone is already a dependency graph: which observations fed which invocations, in what
order. That is structurally the same object the compiled continuation needs, so the continuation is
compiled from what happened rather than authored from nothing.

This gives *solidifying a flow* a literal meaning: **the flow becoming a continuation.**
[`../reach/age.md`](../reach/age.md) then versions exactly that solidified structure, which is why
the age triggers are a position's boundary and not its body.

It also explains a feature of the existing examples. Every worked example in
`continuation-language.md` is a straight-line DAG, and a continuation derived from a single trace
*would* be straight-line. The document is describing the first-draft form. Branching arrives when a
later execution diverges, and by the age rule a body change does not increment the version — so the
structure grows without churning its identity.

## Which surface, and when

Not size. Size is a proxy for the real distinction:

```text
DECLARED DEPENDENCE      I know I need f(x), then g(f(x)).           → LANGUAGE. one emission.
DISCOVERED DEPENDENCE    I do not know what to do until I see f(x).  → PRIMITIVE. the round trip is the point.
```

In primitive mode a data dependency *forces* a crossing: the parent must see the value before it can
emit the next selection. In the language a data dependency is a name —
`findFlights(origin, destination, dates)` composes with `destination` without the parent ever
observing it.

This is why the repair loop is irreducibly primitive. A parser returns `0`, the parent wakes and
**looks**, and only then Thinks. That cannot be compiled ahead of time because the input to the
decision does not exist yet.

## The cost this is answering

The per-step transfer cost is already named in the code as a first-class tension —
`skyra-dna-v5/frame/frame.go`:

> *"The frame is also a BUDGET boundary. Everything a possibility sees ships to the entity's host on
> every step, so what it may see is simultaneously the learning mechanism and the per-step transfer
> cost, and the two pull against each other."*

And the frame grows. `Returned` accumulates, so step *k* ships *k−1* prior returns:

```text
PRIMITIVE, DEPTH N     N crossings of GROWING size.   quadratic in transferred context.
LANGUAGE               one emission.
```

The amplification only applies to *sequentially dependent* work — independent children already
emit as one group, since arity is the whole concurrency declaration. But sequential dependence is
exactly the case the language removes the crossings from.

### Success costs no crossing

The deeper payoff is not batching. In language mode the parent-owned parser runs in the kernel
against the returned observation, so a satisfied contract needs no entity at all.

```text
SATISFIED RETURN     kernel parses. continuation advances. NO CROSSING.
UNSATISFIED RETURN   parent wakes. focused frame. entity judgment required.
```

> **The entity is consulted on novelty and repair, and on nothing else.**

`MaxSteps` exists in the current kernel because every step is a crossing. A solidified flow has no
step loop on the happy path.

## What this settles

```text
WHY THE LANGUAGE NEEDS A PREDECESSOR   it cannot declare the positions it consumes.
WHERE ORIGINATION LIVES                the crossing. the language has no construct for it.
WHAT SOLIDIFICATION IS                 the retained cone projected as a continuation.
WHICH SURFACE                          declared dependence → language. discovered → primitive.
WHAT KEEPS THEM ONE SYSTEM             exact desugaring, checkable in both directions.
```

## Open

```text
REFRESHING A CONTINUATION      compile-from-cone creates a continuation correctly, but a cone
                               records only the path that executed. Re-deriving from a later cone
                               would delete branches that did not run that time. This argues for
                               deriving ONCE and amending thereafter, but the amendment mechanism
                               is unspecified. FLAGGED, NOT DECIDED.

WHAT RE-OPENS A SOLIDIFIED     once a flow compiles, the parent stops selecting, and a parser
STRUCTURE                      returning 0 becomes the only signal that its structure is wrong. A
                               position that is unnecessary but always satisfies its contract is
                               invisible to that signal. Whether anything deliberately re-opens
                               solidified structure is unaddressed. FLAGGED, NOT DECIDED.

WHO COMPILES                   [`../coherence/README.md`](../coherence/README.md) names a
                               "continuation compiler" as the owner of invalid source programs, but
                               no primitive corresponds to it. `primitives.md` absorbs COMPILE into
                               Think, and Think returns ranked candidates rather than artifacts.

THE FIRST FRAME                what an engram's frame contains before any position exists, and
                               whether the empty `children` block is supplied at all or the engram
                               is simply opened in primitive mode.

MIXED PATHS                    an engram that composes on one path and originates on another must
                               drop to the crossing for the second. Whether that is one possibility
                               changing surface mid-life, or two, is unspecified.
```
