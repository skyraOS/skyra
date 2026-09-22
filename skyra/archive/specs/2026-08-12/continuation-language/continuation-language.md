# Continuation language

**Date:** 2026-08-13
**Status: WORKING PROPOSAL.** A small typed, LLM-facing language for compiling an engram's child
contracts and continuation. This document defines the language's semantic surface. It does not
choose a parser, serialization, compiler implementation, or host language.

> **The LLM writes ordinary-looking typed composition code. The runtime supplies the frame in
> which that code is legal. Beneath the syntax, every child call is
> Decide → Reach → Possibility → Remember.**

This proposal builds on:

- [`engram-contracts-and-composition.md`](engram-contracts-and-composition.md), for parent-owned
  child contracts and contextual typing;
- [`think-interface-proposal.md`](think-interface-proposal.md), for bounded potential expressed as
  expectation-bearing observations; and
- [`primitives.md`](primitives.md), for the semantic roles of Think, Decide, Reach, and Remember.

---

## What the language compiles

An engram has two fundamental parts:

```text
ENGRAM

1  CHILD CONTRACTS
   the behavior positions available beneath this engram
   and the typed observations they may receive and must return

2  COMPILED CONTINUATION
   the ordered composition of those child positions
```

In one line:

```text
ENGRAM = CHILD CONTRACTS + COMPILED CONTINUATION
```

The continuation language is the authored representation of that shape. Its declarations state
the contracts. Its method body states how observations flow through them.

It is not a general-purpose programming language. Its purpose is narrower:

> **Given observations already in scope and child behaviors already made available, define the
> causal order in which those children are composed and which returned observations leave the
> engram.**

---

## The runtime supplies the frame

The LLM does not invent the environment in which it writes. The runtime supplies a frame with
three language-visible pieces:

```text
METHOD SIGNATURE
    typed observations available to this invocation

CHILD SIGNATURES
    child behaviors available from this position and their contracts

RETURN CONTRACT
    typed observations the parent expects this engram to return
```

The frame may also be bounded by runtime-supplied `From` and `Horizon`, but those are not ordinary
language values:

```text
FROM        namespace × requesting engram × runtime
HORIZON     the outer bound on what this invocation may see outward or cause
```

The continuation cannot declare, rewrite, or widen them.

### The frame combines two ownership perspectives

The runtime assembles one frame, but it does not own all three contracts:

```text
METHOD SIGNATURE      supplied from the parent-owned input contract for this child position.

RETURN CONTRACT       supplied from the parent-owned return contract for this child position.

CHILD SIGNATURES      owned by the current engram for its own child positions.
```

This distinction is load-bearing. The parent owns what it can supply and consume because its
continuation depends on the result. The current engram owns the contracts for the children on which
its own continuation depends. The runtime renders those contracts together; it does not author
them.

---

## Example frame

```text
TripResult planTrip(
    TripRequest request,
    Traveler traveler,
    Location origin
)

children {
    Destination getDestination(
        TripRequest request
    );

    TravelDates getTravelDates(
        TripRequest request
    );

    FlightOptions findFlights(
        Location origin,
        Destination destination,
        TravelDates dates
    );

    HotelOptions findHotels(
        Destination destination,
        TravelDates dates
    );

    Itinerary buildItinerary(
        Traveler traveler,
        FlightOptions flights,
        HotelOptions hotels
    );

    TravelPlan createTravelPlan(
        Itinerary itinerary
    );
}

returns {
    plan: TravelPlan,
    flights: FlightOptions
}
```

This says:

```text
ALREADY OBSERVABLE
    request · traveler · origin

AVAILABLE CHILD POSITIONS
    getDestination · getTravelDates · findFlights · findHotels
    buildItinerary · createTravelPlan

REQUIRED OUTWARD OBSERVATIONS
    plan: TravelPlan · flights: FlightOptions
```

It does not say that any child has run, that any endpoint is reachable, or that any outward return
has been satisfied. It describes the legal space in which the continuation may be compiled.

---

## Example continuation

```text
TripResult planTrip(
    TripRequest request,
    Traveler traveler,
    Location origin
) {
    Destination destination = getDestination(request);
    TravelDates dates = getTravelDates(request);

    FlightOptions flights = findFlights(
        origin,
        destination,
        dates
    );

    HotelOptions hotels = findHotels(
        destination,
        dates
    );

    Itinerary itinerary = buildItinerary(
        traveler,
        flights,
        hotels
    );

    TravelPlan plan = createTravelPlan(itinerary);

    return {
        plan: plan,
        flights: flights
    };
}
```

The syntax resembles an ordinary typed method. The semantics are causal rather than local-memory
semantics.

---

## Method signatures

```text
TripResult planTrip(
    TripRequest request,
    Traveler traveler,
    Location origin
)
```

The signature defines what is already observable inside this invocation.

```text
TYPE             parent-owned observation contract
LOCAL NAME       name by which this continuation refers to that observation
```

The method parameters are not arbitrary values copied into a process. They are observations the
parent actually supplied to this frame. Nothing omitted from the signature becomes available merely
because it exists elsewhere in the graph or world.

```text
TripRequest request
│           │
│           └── local name owned by this continuation
└── contract the supplied observation is known to satisfy
```

The method's nominal return name, `TripResult`, identifies the outward contract as a whole. The
`returns` block gives its consumable members.

---

## Child signatures

```text
Destination getDestination(
    TripRequest request
);
```

A child signature is a parent-owned contract for one child position in this engram.

```text
CHILD NAME         the position this continuation may invoke
INPUT CONTRACT     which typed observations may be bound to it
RETURN CONTRACT    which typed observation this continuation can consume from it
```

It is not a declaration made by a globally authoritative child. It says:

> **At this position, this engram can compose with behavior accepting these observations and can
> continue from a returned observation satisfying this contract.**

The same underlying behavior may be used under a different contract by another parent. No
persistent type edge is required between this declaration and one particular child engram.

---

## Child invocation

```text
Destination destination = getDestination(request);
```

One line represents a full causal transition:

```text
the continuation chooses the child position and inputs
    ↓ DECIDE
the runtime attempts to realize the selected executable and destination
    ↓ REACH
the child possibility becomes live
    ↓ POSSIBILITY
the child returns an observation
    ↓ CONTRACT SATISFACTION
the parent remembers that observation locally as `destination`
    ↓ REMEMBER
the continuation may proceed
```

The surface call expresses the decision. Reach is not another statement the LLM writes. It is the
runtime's attempt to realize that decision.

The child name refers to a contract position, not a network address or globally fixed callee.
Think and prior composition may have supplied compatible executable and reach observations for the
position; the continuation language consumes the resulting position without exposing volatile
mechanism.

---

## Typed assignment is local Remember

```text
Destination destination = getDestination(request);
```

The typed assignment has three participants:

```text
Destination        expected observation contract
destination        local name chosen by the parent continuation
returned object    observation produced by the realized child possibility
```

The child does not return a variable named `destination`. It returns an observation. The parent
tests that observation against its return contract and, when satisfied, remembers it under a name
meaningful inside its own continuation.

```text
RETURNED OBSERVATION Y
        ↓ Y ⊨ Destination
LOCAL REMEMBRANCE `destination`
```

The local name is not durable identity and is not a type tag placed into observation content. It is
a binding in this compiled continuation.

---

## Return is outward Remember

```text
return {
    plan: plan,
    flights: flights
};
```

The return statement selects observations already remembered in this continuation and presents them
according to the supplied outward return contract.

```text
LOCAL NAME       plan
OUTWARD FIELD    plan
CONTRACT         TravelPlan
```

The statement cannot originate a value which was never supplied or returned. It composes known
observations into the outward shape expected by the parent.

Semantically, this is the engram's outward Remember:

```text
local remembered observations
        ↓ select fields satisfying the return contract
returned observation
        ↓ parent may Remember it
parent continuation
```

Inner return remains ephemeral. Only an unbroken chain of remembrance reaching the outer boundary
makes the retained causal cone durable.

---

## Nothing may appear from nowhere

Every value used by the continuation must be traceable to exactly one available causal source:

```text
METHOD PARAMETER
    an observation passed into this frame by the parent

CHILD ASSIGNMENT
    an observation returned by a child possibility and remembered locally
```

There are no unbound literals standing in for observations, no ambient graph reads, and no hidden
entity memory of this caller.

The compiler can therefore reconstruct the observation dependency graph directly from names and
contracts:

```text
TripRequest
   ├──▶ Destination
   └──▶ TravelDates

Location + Destination + TravelDates
   └──▶ FlightOptions

Destination + TravelDates
   └──▶ HotelOptions

Traveler + FlightOptions + HotelOptions
   └──▶ Itinerary

Itinerary
   └──▶ TravelPlan
```

The outward return contract determines which paths must ultimately complete:

```text
TravelPlan       required as `plan`
FlightOptions    required as `flights`
```

Any invocation whose return cannot contribute to a future dependency or the outward return is
observable as unused exploration rather than silently becoming part of the result.

---

## Types are contracts grounded by provenance

The familiar type names in the language are not self-declared tags.

```text
NOT TYPING
    child returns {"type": "Destination"}

TYPING
    the substrate knows which engram and invocation produced the observation;
    the parent-owned contract determines whether that known observation satisfies Destination here
```

The language makes contextual contract satisfaction readable:

```text
Y ⊨ Destination
```

Where:

```text
Y              observation with substrate-known provenance
Destination    contract owned by this parent position
⊨              this continuation has a valid transition for Y here
```

The same observation may satisfy a contract in one parent and not another. A type name therefore
does not create a global class hierarchy. It names a contract meaningful in the continuation that
owns it.

The compiler checks name and contract flow. The substrate grounds producer identity at runtime.
Neither trusts a child's content to certify itself.

---

## Contract mismatch is structural failure

Suppose the parent compiled:

```text
Dataset data = getDataset(request);
```

The realized child returns `Observation<Y>`, but:

```text
Y ⊭ Dataset
```

Then no binding named `data` is created. The continuation has no valid next state through that
statement and cannot proceed along dependencies requiring `data`.

That is the substrate-level failure. A universal semantic `Error` object is not required.

An explanatory observation may still exist:

```text
{"explanation": "the source returned tabular metadata rather than a dataset"}
```

But the explanation reports on the mismatch. It does not create it.

The runtime must preserve the structural occurrence:

```text
which child position was selected
which executable and destination were decided
whether Reach created the child possibility
which observation actually returned
which parent-owned contract it did not satisfy
which continuation dependencies could no longer proceed
```

---

## Three distinct failure boundaries

The language makes three failures separable:

```text
1  COMPILE-TIME INCOMPATIBILITY
   a name, argument, child position, or return field violates the supplied contracts.
   no decision is emitted.

2  UNREALIZABLE DECISION
   the composition is valid, but Reach cannot establish it now or Horizon excludes it.
   no child possibility exists.

3  RETURN MISMATCH
   the child possibility ran, but the returned observation does not satisfy the assignment type.
   no local remembrance is bound and the dependent continuation cannot proceed.
```

All may produce explanatory observations. None depends on an `Error` value to be a failure.

---

## Order and dependency

The method body defines causal order.

```text
STATEMENT ORDER       which decision occurs before another where order is explicit.
NAME DEPENDENCY       which invocation cannot occur until required observations exist.
RETURN DEPENDENCY     which remembered observations must exist before outward return.
```

In the example:

```text
findFlights cannot be decided before destination and dates are remembered.
findHotels cannot be decided before destination and dates are remembered.
buildItinerary cannot be decided before flights and hotels are remembered.
createTravelPlan cannot be decided before itinerary is remembered.
```

`findFlights` and `findHotels` have no data dependency on one another. Whether independent calls
run concurrently must be explicit in the eventual language semantics; source layout alone must not
accidentally make scheduler timing semantic.

The current proposal commits only to this:

> **The compiled continuation preserves declared order and observation dependencies. Completion
> order may never rewrite either one.**

The syntax for explicit concurrency, branching, retry, and alternate continuation remains open.

---

## Think and continuation compilation

Think creates bounded potential as expectation-bearing observations. A continuation may be authored
or revised by deciding on such an observation while preserving it as a cause.

```text
FRAME CONTRACTS
      ↓
Think proposes a continuation expectation
      ↓
candidate observation C
      ↓ Decide on C
compile continuation under the supplied frame
      ↓
provisional engram behavior
```

The compiler must not accept expectation text detached from the observation that proposed it. Doing
so would preserve the code while discarding the causal path by which it was authored.

Once compiled, the continuation itself invokes only child positions supplied in its frame. It does
not call Think implicitly whenever it encounters an unfamiliar name. New potential must enter
through an explicit available behavior and remain observable in the causal record. When an actual
child return later fails its parent-owned parser, the runtime may wake that parent into a narrower
frame where inspection and new thought are causal behavior rather than implicit model action.

---

## Working note: execution has direction

The continuation language describes behavior on the way down. It selects child positions, binds
observations, and establishes the dependency chain that execution will attempt. Contract truth is
then realized on the way up, when children return observations to the parents that own their
contracts.

```text
DOWN
    typed program composes children and dependencies
    engrams may be discovered, selected, or revised
    ↓
LEAF
    terminal behavior fires once the activated flow reaches it
    ↓
UP
    returned observations cross parent-owned parsers
    satisfying observations advance compiled continuations
    outward Remember carries the result farther upward
```

Contracts constrain the downward program, but a runtime return mismatch is necessarily discovered
on the way up. A parent cannot know that a particular invocation returned what it needs until the
observation crosses that boundary.

---

## The parent owns the parser for each child

The parent-owned return contract is also a parser for that child position. It performs one local,
binary judgment over the observation that actually returned:

```text
child returns Observation<Y>
        ↓
parent parser for this child position
        │
        ├── 1   Y satisfies the expected contract; continue
        └── 0   no continuation exists for Y here; wake the parent
```

The producer cannot make this judgment because it does not own the consumer's continuation. No
global parser must walk the entire causal tail either. Each parent parses one child boundary, then
produces an observation which its own parent parses recursively.

```text
leaf produces A
    ↓ parent parses A and produces B
    ↓ parent parses B and produces C
    ↓ parent parses C
```

In this sense, the recursion is the parser. The complete causal type is established by the chain of
local parent-owned judgments.

Operationally, that boundary runs through the kernel's `Recognize` method. The parent owns the
contract because its continuation defines what the child position must return. The kernel owns the
protocol: it delivers the observation, calls the relevant entity and engram methods, and records
the result. Those methods compose the judgment beneath the protocol.

> **The kernel owns the protocol. Entities and engrams compose under it.**

---

## Inspection follows the failed dependency

`inspect()` is not an operation the LLM elects to place speculatively in every continuation. A
parser result of `0` identifies the exact parent, child position, expected contract, and returned
observation involved in the failure. That parent wakes, and inspection begins from that dependency.

The runtime may supply a new, more tightly bounded frame:

```text
EXPECTED       the contract owned for this child position
RECEIVED       the observation which failed that parser
FAILED CHILD   the exact child position which returned it
AVAILABLE      only observations and behaviors relevant to satisfying the dependency
AUTHORITY      inherited From and a Horizon no wider than the failed invocation's
```

Within this frame, the engram may compose behavior to rectify its own failed expectation. If that
behavior encounters a deeper type mismatch, the same rule applies recursively:

```text
continuation
    ↓ type failure at child A
bounded frame A
    ↓ type failure at child B
bounded frame B
    ↓ type failure at child C
bounded frame C
    ↓ resolved observation
resume and unwind upward
```

This is a progressive causal zoom, not ambient graph access. Every new frame must remain centered
on the failed dependency and must not gain unrelated observations, children, or authority. When a
satisfying observation is produced, the suspended continuation may resume. If the dependency
cannot be resolved within the narrowed frame and Horizon, the continuation remains structurally
unable to proceed.

The ownership rule is therefore:

> **The engram which defines a child contract owns both the parser for that return and the behavior
> used when the parser cannot continue.**

---

## The language remains intentionally boring

```text
SIGNATURE       observations already available
CHILD METHODS   behavior positions which may be invoked
TYPES           parent-owned contracts constraining composition
ASSIGNMENT      local Remember of a child's returned observation
RETURN          outward Remember according to the supplied contract
METHOD BODY     compiled continuation and observation dependency order
TYPE CHECK      parent-owned binary parsing of an observation returned by a child
INSPECT         dependency-following behavior after a failed return wakes that parent
```

The syntax should feel familiar because novelty belongs in the semantics, not punctuation.

Underneath:

```text
METHOD CALL
    Decide → Reach → Possibility

TYPED ASSIGNMENT
    contract satisfaction → local Remember

RETURN
    outward Remember

WHOLE BODY
    causal continuation over known observations
```

---

## What the language may express

The minimal language must express:

```text
typed method parameters
declared child contracts
typed child invocations
local names for returned observations
dependencies through those names
outward return composition
```

It may not express:

```text
undeclared child names
ambient addresses or transports
arbitrary graph lookup
unbounded world access
values with no observation provenance
authority wider than Horizon
self-declared type satisfaction
durable mutation merely because code was compiled
```

---

## Leaves

A leaf engram has no child signatures but still has a compiled continuation:

```text
Result leaf(Input input)

children {
    // none
}

returns {
    result: Result
}
```

Its terminal behavior produces an observation according to the outward contract without invoking a
child. Exactly how leaf origination appears in the continuation language remains open; it must not
allow an orchestrator to smuggle uncaused values into an otherwise composed result.

Leaf and orchestrator remain structural:

```text
LEAF            this execution path invokes no child.
ORCHESTRATOR    this execution path invokes at least one child.
```

They are not declarations on the engram.

---

## What remains open

```text
the concrete syntax and parser
the representation of an observation contract behind a familiar type name
the syntax for explicit concurrency
branching and alternate continuations after an unsatisfied return
retry and whether it reuses or creates a child position occurrence
whether multiple return contracts may govern one child position
the precise Inspection observation contract and what candidate information it may reveal
how provisional continuation revision appears while an engram remains live
how leaf terminal origination is expressed without weakening causal provenance
whether loops are admitted, bounded, or compiled as repeated positions
how a language-level child name is paired with Think-produced executable and reach observations
```

None changes the core semantics:

```text
THE RUNTIME supplies the frame.
THE ENGRAM owns its child contracts and continuation.
THE PARENT owns this engram's input and return contracts.
THE LLM composes only typed observations available through those contracts.
METHOD CALLS cause Decide → Reach → Possibility.
ASSIGNMENTS and RETURN are Remember at two boundaries.
```

---

## Acceptance test

The proposal holds only if:

```text
1  Every name used in the body traces to a method parameter or child return.
2  Every child call names a child position supplied in the frame.
3  Every argument satisfies that position's parent-owned input contract.
4  Every assignment is created only after a returned observation satisfies its contract.
5  Every outward field satisfies the supplied return contract.
6  Method calls preserve Decide → Reach → Possibility as distinct events.
7  Typed assignment and outward return preserve local and outward Remember respectively.
8  Type claims cannot be forged by observation content.
9  Contract mismatch blocks dependent continuation without requiring an Error value.
10 The compiled dependency graph can be reconstructed from names and contracts.
11 Every child return is parsed by the parent which owns that child contract.
12 A failed return wakes that parent and bounds inspection to the exact failed dependency.
13 Address, transport, ambient graph access, and Horizon mutation are inexpressible.
14 The LLM cannot change the supplied method signature, child signatures, or return contract.
15 A reached leaf fires terminal behavior without inventing a declared leaf kind.
16 Compilation creates provisional behavior, not durable history.
```

If any one fails, the language has ceased to be a typed continuation over observable causes and has
become either an unrestricted programming language or a decorative syntax over hidden runtime
behavior.
