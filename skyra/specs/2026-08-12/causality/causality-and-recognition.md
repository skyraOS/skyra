# Causality and Recognition

**Date:** 2026-08-13
**Status: WORKING NOTE.** This records a direction to investigate. It is not yet a type proposal or
work order.

**Coherence position:** [`observable-causality.md`](observable-causality.md) is the narrower account
of what the substrate can prove. Where this document discusses primitive surface, the later
[`../continuation-language/primitives.md`](../continuation-language/primitives.md) controls.

---

## The correction

Creating a candidate is free. Making behavior durable is not.

An entity may emit an emission, an observation, or a proposed expectation into the live execution.
That act does not by itself write durable state. The material belongs to the ephemeral possibility
and may disappear completely.

An engram is different. It is compiled behavior that has been remembered and appended into
durable structure.

```text
EMIT                 free. ephemeral. part of what is happening.
REMEMBER             selective closure. retains what contributed.
ENGRAM               durable behavior produced by that retention.
```

This makes the system look like two programs sitting on top of one another:

```text
THE ENGRAM           the program being interpreted — WHAT to do.
THE ENTITY           the program interpreting it — HOW it becomes an emission.

Engram × Entity × Frame  →  Emission
```

Resolving the entity below an engram is itself behavior and may therefore be an engram. A resolver
can select an entity capable of interpreting or compiling a new behavior. If the resulting work is
rejected, it is thrown away. If remembered, the contributing work is appended and becomes reusable.

The compiler analogy is therefore not only about parsing text. It is about the boundary between a
free, provisional execution and remembered, durable behavior.

---

## The two laws of remembrance

> **Something can occur without being remembered. But something cannot be remembered without
> having occurred.**

These are two separate laws:

```text
1  OCCURRENCE DOES NOT IMPLY REMEMBRANCE
   Something may happen in live execution and then evaporate.

2  REMEMBRANCE IMPLIES OCCURRENCE
   Nothing may enter memory unless the substrate witnessed it occur.
```

In compact form:

```text
REMEMBERED ⊆ OCCURRED
```

The first law permits exploration. Possibilities may be opened, observations produced, drafts
revised, endpoints tried, and effects staged without forcing any of them into durable state.

The second law grounds memory. `Remember` may select from witnessed causal events; it may not
originate an event, invent a result, or write a retrospective account of work that did not happen.
Every durable remembrance must resolve to an occurrence held in the live causal record.

Together they define remembrance as selective retention of reality rather than narration:

```text
OCCUR            live, witnessed, possibly ephemeral.
REMEMBER         select an occurrence and carry its causal path upward.
OUTER REMEMBER   project recursively remembered occurrences into durable state.
```

### Remembrance is selective; cost is not

> **Causality cannot be created by remembrance; energy spent creating possibilities cannot be
> unspent. Remembrance determines which causal work survives, and efficiency measures the cost of
> that selection.**

For a remembered result `r`:

> **C(r) = total physical/economic cost required to obtain remembered result r.**

An occurrence may be excluded from durable memory without becoming physically free. A failed
endpoint, discarded model attempt, or abandoned draft may be absent from the remembered causal cone
while its time, energy, bandwidth, compute, or monetary charge was still paid on the path to `r`.

```text
REMEMBERED(r) ⊆ OCCURRED(r)
COST BASIS(r) may include unremembered occurrences required to obtain r
```

The durable content projection may therefore forget an occurrence while cost accounting preserves
that the occurrence was paid for. The two records have different purposes and must not be silently
collapsed.

This gives the three quantities different subjects:

```text
OCCURRED WORK       every witnessed possibility and crossing that physically ran.
REMEMBERED WORK     the causal subset selected to survive.
C(r)                the physical/economic cost required to obtain remembered result r.
EFFICIENCY          the cost of producing and selecting that surviving causal work.
```

Efficiency cannot be inferred from the compactness of the durable cone. A small remembered result
may have required a large field of expensive possibilities. Forgetting those possibilities removes
their content from memory; it does not refund their cost.

The cost account belongs to the namespace on whose behalf the work occurred. Whether the kernel
must append that cost directly when no causal chain is remembered remains open. This is deliberately
separate from `Remember`: causal content and incurred cost have different subjects.

---

## Two semantic distinctions exposed by this investigation

```text
1  REMEMBER         Which events in this live execution become durable?

2  RECOGNIZE        Have I seen this observation, or something like it, before?
```

They operate on different things and at different moments.

`Remember` operates at every parent boundary. It unifies handing a result upward with the operation
currently wearing the name `Accept`. A parent remembers a child's observation into its live causal
history. When the outermost caller remembers, that same operation closes the branch, retains what
contributed, appends it to the graph, and materializes reached effects.

`Recognize` operates while the branch is still live. It receives an observation remembered into a
possibility and arranges for it to be translated into a form that can be compared with prior
experience. Its products remain ephemeral unless a later `Remember` reaches them.

Only the kernel holds the live causal structure on which they operate, but that does not make both
peer kernel methods. `Remember` is a causal form. Recognition is specialized behavior composed
beneath Think. Only their structural coordination belongs to the kernel:

```text
REMEMBER     the kernel attaches a result to its parent causally. at the outer boundary it walks
             those remembrances and appends the retained projection. no semantic judgement is
             required.

RECOGNIZE    Think may compose recognition behavior and the kernel records its causal result.
             recognition behavior interprets meaning; the kernel does not.
```

The signature is deliberately not proposed here. An inner possibility and the external caller have
different concrete shapes, and the important decision is that they perform the same operation:
they remember a result from below.

Internal commit, upward handoff, and external acceptance were three names around one recursive act.
The unified name is `Remember`.

`Recognize` may contain two different kinds of work without giving semantic judgement to the
kernel:

```text
TRANSLATE     invoke behavior that turns the remembered observation into a comparable key.
LOOK UP       kernel arithmetic: has this key, or this explicitly produced class key, been seen?
DISPATCH      open the automatic behavior associated with the recognized key.
```

Exact key equality is structural. "Something like this" remains semantic only until a translator
has made the likeness explicit as a key the kernel can compare.

---

## Method one: Remember preserves causality

The system must preserve order while execution is happening because causality inside an engram
matters. A frame cannot be only the latest expectation plus unordered collections of observations.
It needs an append-only ephemeral account of what happened and what caused what.

```text
e1  received observation A
e2  interpreted A under expectation X              after e1
e3  selected translator T                          after e1 + e2
e4  parent remembered translator observation B     after e3
e5  revised expectation X → X'                     after e4
e6  selected behavior P                            after e4 + e5
e7  parent remembered behavior observation C       after e6
e8  committed C                                    after e7
```

All of these events live in the kernel while the possibility remains open. None of them needs to be
written durably as it happens.

### Order is not enough by itself

A single sequence number preserves order within one possibility, but selections may run children
concurrently. Arrival order is a scheduling accident and must not become meaning.

The live record therefore needs to preserve:

```text
SEQUENCE             order within one possibility.
CAUSES               explicit predecessor edges between events.
EMITTED POSITION     deterministic order within a concurrent selection group.
```

The result is a causal graph rather than one global timeline.

An illustrative shape, not a settled type:

```go
type Event struct {
    ID       EventID
    After    []EventID
    Kind     EventKind
    Engram   EngramID
    Inputs   []ObservationID
    Output   ObservationID
}
```

The kernel may hold and connect these identifiers without interpreting observation content.

### Remember is a backward walk

Each inner `Remember` creates a causal edge from a parent to the observation it chose from below.
When the outermost caller remembers, begin with that chosen observation and walk backward through
those dependencies:

```text
remembered C
   ← remembered behavior observation
   ← selection of P
   ← translated observation B
   ← remembered translator observation
   ← original observation A
```

Only that causal cone becomes durable. Expectations, observations, revisions, selections, and
observations that did not contribute to what was remembered are not appended.

This is better understood as projection than mutation:

```text
EPHEMERAL CAUSAL RECORD
          ↓ Remember projects the retained causal cone
DURABLE GRAPH RECORD
```

The kernel does not semantically judge which material was useful. It follows structural causal
edges from the remembered result. Everything unreachable from that result evaporates with the live
possibility.

This generalizes the existing citation walk. A citation is one kind of causal edge, but revisions,
translations, and selection decisions may also need explicit causal representation if `Remember`
must retain them.

---

## Method two: Recognize

When an observation is remembered into a receiving possibility, the kernel may call `Recognize`.
Recognition requires translating or parsing the observation into a representation that can be
compared with something known.

```text
remembered observation
        ↓
translation / parsing
        ↓
canonical or comparable representation
        ↓
recognition
        ↓
automatic behavior previously learned for this kind of thing
```

For example, these may be successive representations of the same evidence:

```text
"disk is 91% full"
{"filesystem_usage": 0.91}
[metric: filesystem-usage, band: critical]
```

Recognition can occur at several strengths:

```text
IDENTICAL       I have seen this canonical material before.
SIMILAR         I have seen this structural shape or partition before.
FAMILIAR        a learned translator places it under a known key or class.
```

The existing system already contains primitive forms of this:

```text
Content.Shape()       structural similarity independent of values.
Fingerprint()         comparable material independent of observation identity.
Partitions()          comparable invocations grouped by input and output shape.
```

These are beginnings, not the whole recognition mechanism.

### Translation is behavior

The kernel must not decide what an observation means. Translation is semantic work and must itself
be performed by an engram interpreted by an entity.

```text
incoming observation
       ↓
translator engram
       ↓
canonical observation
       ↓
recognized behavior engram
       ↓
result
```

Because translation is ordinary behavior, it remains observable and correctable. If a translation
causes a result that is later remembered, its invocation is inside the retained causal cone. If it
misclassifies the observation and its result is not remembered, that pairing receives no durable
support.

### Recognition must not become a cache

Recognition means knowing which behavior to execute, not substituting an old observation for a new
invocation.

```text
CORRECT      "I have seen something like this, so I know what behavior to run."

INCORRECT    "I have seen something like this, so the old answer is the current answer."
```

The second interpretation destroys sensitivity to a changing world. Historical observations are
evidence about behavior, never replacements for performing that behavior now.

---

## Automatic behavior as compilation

Repeated interpretation can produce specialized reusable behavior:

```text
FIRST ENCOUNTER

observation → general interpretation → behavior → result → remembered


LATER ENCOUNTER

observation → translation → recognized class → automatic behavior → result
```

The automatic behavior is compiled in the meaningful sense: an expensive or general interpretation
has become a stable procedure for a recognized class of input.

It is not durable merely because an entity emitted it. During the live branch it is provisional.
It becomes an engram only if `Remember` reaches the causal events that produced it and appends those
events into durable state.

```text
ENTITY EMITS          a provisional program or result.
FRAME RECORDS         its position and causal dependencies.
PARENT USES           the proposal, translation, or result.
REMEMBER REACHES      the contributing causal chain.
REMEMBER APPENDS      the retained behavior as durable structure.
```

This permits compiler entities and compiler strategies to compete like any other interpreters.
Different entities may translate or compile the same material differently. Their results can be
tried, remembered, rejected, and selected using accumulated history.

---

## The emerging loop

```text
LIVE EXECUTION
    append causally linked events to ephemeral state

REMEMBER
    receive an observation into the parent's causal state; at the outer boundary make its cone durable

RECOGNIZE
    produce a comparable representation through ordinary behavior

MATCH
    find exact, similar, or learned classes from prior remembered experience

SPECIALIZATION
    invoke reusable automatic behavior associated with that recognition

OUTERMOST REMEMBER
    project the recursively remembered cone into the durable graph and close the branch
```

The division of responsibility is:

> **The kernel remembers structure and orchestrates recognition. Engrams interpret, translate, and
> recognize meaning.**

---

## Questions this creates

These are not answered here:

```text
CAUSAL EDGE SURFACE       Which events need explicit edges beyond observation citations?

EXPECTATION REVISIONS     How is a revision represented so a remembered result retains the exact
                          expectation under which it was produced?

CONCURRENT GROUPS         What is the minimal representation that preserves emitted order without
                          preserving accidental arrival order?

CANONICAL FORM            Is there one base observation format, several translators, or a hierarchy
                          of progressively more specific representations?

RECOGNITION THRESHOLD     What distinguishes exact repetition from "something like" repetition,
                          and which part is structural versus interpreted?

COMPILED ARTIFACT         What provisional representation becomes an engram at closure?

IDENTITY AND OWNERSHIP    At what exact point does remembered behavior receive an EngramID, a name,
                          and one durable parent?

EVICTION                  What becomes of an open causal record when execution is evicted before
                          Remember?
```

The next useful investigation is not merely how to parse textual emissions. It is how a live,
causally ordered interpretation becomes recognized automatic behavior without writing failed or
irrelevant exploration into durable state.
