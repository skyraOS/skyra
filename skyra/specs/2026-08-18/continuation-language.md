# Continuation language

**Date:** 2026-08-19
**Status: WORKING PROPOSAL.** Rewritten against the causal-transition-unit model. Supersedes
`2026-08-12/continuation-language/continuation-language.md`, which carried three claims that did
not survive — the desugaring, and the two operations it both called Remember. What it got right
is carried forward, and much of its prose with it.

> **The body writes ordinary-looking typed composition code. The runtime supplies the frame in
> which that code is legal. Beneath the syntax, EVERY LINE IS A MOVE.**

---

## What changed from the 08-12 proposal

```text
RETIRED   "every child call is Decide → Reach → Possibility → Remember."
          A child call is ONE MOVE with a request header. Reach is a stage inside it, not a step.

RETIRED   "typed assignment is local Remember" and "return is outward Remember."
          REMEMBER IS ROOT-ONLY AND BULK. Neither of these is Remember, and both are still real.
          §4 gives them their actual names, and neither needs a new primitive.

ADDED     SEQUENCE IS A DEPENDENCY, NOT A SCHEDULE. The 08-12 document listed this as open —
          "source layout alone must not accidentally make scheduler timing semantic." It is
          settled, and the answer changes what a body means.

ADDED     THE BODY DOES NOT AUTHOR ITS CHILD SIGNATURES. THINK does, and the available set
          SLIDES as execution proceeds.
```

Everything else in that document holds and is restated here.

---

## What the language compiles

```text
ENGRAM = CHILD CONTRACTS + COMPILED CONTINUATION
```

The continuation language is the authored representation of that shape. Its declarations state
the contracts; its body states how observations flow through them.

It is not a general-purpose programming language. Its purpose is narrower:

> **Given observations already in scope and child behaviours already made available, define the
> causal order in which those children are composed and which returned observations leave the
> engram.**

---

## 1 · The runtime supplies the frame

The body does not invent the environment it writes in. Three language-visible pieces:

```text
METHOD SIGNATURE     typed observations available to this invocation
CHILD SIGNATURES     child positions available from here, and their contracts
RETURN CONTRACT      typed observations the parent expects this engram to return
```

The frame is also bounded by a **Horizon**, which is not an ordinary language value. The
continuation cannot declare, rewrite, or widen it — and under capability passing it does not
appear in the body at all. It rides in the envelope of every MOVE the body emits.

### The child signatures come from THINK, and they SLIDE

This is the change with the largest consequence for the surface.

```text
08-12   the frame's `children` block was fixed for the invocation, and the engram "owned the
        contracts for its own child positions."
NOW     THINK authors signatures. The block holds AT MOST FIVE potentials, and THINK re-fires
        on every MOVE that leaves resolved ground.
```

So the set of invocable positions is **a sliding window over the frontier**, not a fixed
declaration. A body cannot name a child that is not currently in the window, and the window
refills ahead of the execution pointer.

Three authorships, and none of them is the body's own choice of interface:

```text
SIGNATURE        authored by THINK.       NEITHER THE CALLER NOR THE CALLEE.
BODY             authored by the CHILD.   it fills a shape it did not choose.
RETURN CONTRACT  NEGOTIATED.              crystallized at REMEMBER.
```

---

## 2 · Example frame

```text
TripResult planTrip(
    TripRequest request,
    Traveler traveler,
    Location origin
)

children {
    Destination   getDestination(TripRequest request);
    TravelDates   getTravelDates(TripRequest request);
    FlightOptions findFlights(Location origin, Destination destination, TravelDates dates);
    HotelOptions  findHotels(Destination destination, TravelDates dates);
    Itinerary     buildItinerary(Traveler traveler, FlightOptions flights, HotelOptions hotels);
}

returns {
    plan:    TravelPlan,
    flights: FlightOptions
}
```

It does not say that any child has run, that any endpoint is reachable, or that any outward
return has been satisfied. **It describes the legal space in which the continuation may be
compiled** — and, now, only the part of that space currently on the frontier.

## 3 · Example continuation

```text
TripResult planTrip(
    TripRequest request,
    Traveler traveler,
    Location origin
) {
    Destination   destination = getDestination(request);
    TravelDates   dates       = getTravelDates(request);

    FlightOptions flights     = findFlights(origin, destination, dates);
    HotelOptions  hotels      = findHotels(destination, dates);

    Itinerary     itinerary   = buildItinerary(traveler, flights, hotels);

    return { plan: itinerary.plan, flights: flights };
}
```

The syntax resembles an ordinary typed method. The semantics are causal.

---

## 4 · The three surface forms, and what each is

This is the section the old document got wrong. All three are MOVE.

### A child call is a REQUEST

```text
getDestination(request)
```

One MOVE, request header. The envelope carries the target and the horizon; the payload carries
the header, a correlation, and the bound inputs. The fabric routes and may refuse; it does not
look inside.

```text
target    the resolved child position          ENVELOPE
horizon   inherited, narrowed only by BOUND    ENVELOPE
payload   request · correlation · inputs       OPAQUE
```

Reach is not a statement anyone writes. It is a stage inside the MOVE — the walk that resolves
the target, invisible when it succeeds and a fault when it does not.

### Typed assignment is the DESTINATION OPERAND — not an act of its own

```text
Destination destination = getDestination(request);
```

The old document called this *local Remember*, which conflated **binding** with **retaining**.
They are different, and separating them is most of what changed:

```text
destination      THE DESTINATION REGISTER of that request. Where the response lands.
Destination      ITS FRAME-RESIDENT CONTRACT, loaded by ACTIVATE with the body.
the landing      performed at the RECOGNIZE port when the response arrives, in the PARENT'S
                 frame, against that contract.
```

So there is no third operation. **A typed assignment names a register and its contract.** The
child does not return a variable named `destination`; it returns an observation, RECOGNIZE
checks it against the contract at that position, and it lands.

Binding is not retention. **Nothing here is durable, and nothing here is Remember.**

### Return is a RESPONSE

```text
return { plan: ..., flights: ... };
```

One MOVE, response header, back to whoever sent the request. It selects observations already
bound in this continuation and presents them under the supplied return contract. It cannot
originate a value that was never supplied or returned.

### And REMEMBER is none of the above

```text
REMEMBER IS AVAILABLE ONLY AT THE ROOT. It is bulk and traced: it tears down every possibility
in the arena and writes only the cone the remembered observation depended on.
```

Which means the durable consequence of everything a body did is decided **somewhere else, after
it is gone.** A body cannot retain, cannot request retention, and cannot observe whether it was
retained. Contribution is the criterion and it is computed.

---

## 5 · Sequence is a DEPENDENCY, not a schedule

```text
WRITING `A; B` DOES NOT ORDER THEM. B follows A only if B DEPENDS ON A.
```

This is where the language will differ from every imperative language a reader brings habits
from. In C or Python `;` means *after*. Here it means *next in the text*, and the machine is
free to run both at once if neither reads the other.

```text
PROGRAM ORDER      what the body says.                      A SPECIFICATION.
EXECUTION ORDER    any order the dependencies allow.        DATAFLOW. fires on readiness.
RETIREMENT         the watermark sweeps forward in order.   restores the illusion.
```

In the example, `findFlights` and `findHotels` have no dependency on one another, so **both
fire as soon as `destination` and `dates` have landed.** Concurrency needs no keyword, no group
form, and no scheduler hint — it is the absence of a dependency.

The cost, and it is deliberate: **ordering for effect alone must be written as a dependency.**
There is no other way to say it, which puts ordering where it can be seen rather than inferred
from layout.

---

## 6 · Nothing may appear from nowhere

Every value must trace to exactly one available causal source:

```text
METHOD PARAMETER    an observation passed into this frame
CHILD ASSIGNMENT    an observation returned by a child and bound locally
```

No unbound literals, no ambient graph reads, no hidden memory of the caller. The dependency
graph is therefore reconstructible from names and contracts alone:

```text
TripRequest ──▶ Destination
            └─▶ TravelDates

Location + Destination + TravelDates ──▶ FlightOptions
Destination + TravelDates            ──▶ HotelOptions
Traveler + FlightOptions + HotelOptions ──▶ Itinerary
```

This rule is now enforced **upstream as well**: THINK draws input types only from observations
in scope, so a signature naming something unavailable cannot be proposed in the first place.

The corresponding operand rule: **no immediates.** A value carried inside an instruction has no
provenance, which is exactly what this forbids. The one exception is a leaf, where the crossing
is the origination point.

---

## 7 · Types are contracts grounded by provenance

```text
NOT TYPING   child returns {"type": "Destination"}
TYPING       the substrate knows which engram and invocation produced the observation, and the
             parent-owned contract determines whether that observation satisfies Destination HERE
```

```text
Y ⊨ Destination
```

The same observation may satisfy a contract in one parent and not another. A type name does not
create a global class hierarchy; it names a contract meaningful in the continuation that owns it.

Two additions from the current model:

```text
A TYPE IS THE RETURN TYPE OF A METHOD SIGNATURE — provisionally. Every type is a method header;
nothing is inert and everything nameable is callable. THERE ARE NO DATA TYPES.

NEW TYPES ENTER ONLY THROUGH REMEMBER. A signature that survives to be remembered crystallizes
into the vocabulary; one that does not leaves nothing. Same rule as contracts and engrams.
```

---

## 8 · Contract mismatch is structural failure

```text
Dataset data = getDataset(request);        // realized child returns Y, and Y ⊭ Dataset
```

No binding named `data` is created. The continuation has no valid next state through that
statement and cannot proceed along dependencies requiring `data`. **That is the failure. A
universal semantic `Error` object is not required.**

An explanatory observation may exist — but it *reports on* the mismatch, it does not create it.

The occurrence must be preserved: which position was selected, which target was decided, whether
the walk resolved, which observation actually returned, which contract it did not satisfy, and
which dependencies could no longer proceed.

### Three failure boundaries, and they arrive differently now

```text
1  COMPILE-TIME INCOMPATIBILITY   a name, argument, position, or return field violates the
                                  supplied contracts. NO MOVE IS EMITTED.
2  UNRESOLVED OR REFUSED          the composition is valid, but the walk did not resolve, or the
                                  horizon refused. CUSTODY NEVER LEFT — the body still holds the
                                  work and may retry, repair, or choose again. RECOVERABLE.
3  RETURN MISMATCH                custody left and came back wrong. NOT recoverable in place;
                                  the only ways on are a further request or escalation.
```

The line between 2 and 3 is **whether custody crossed**, which is what MOVE's success code
reports. That is also the line between a condition you branch on and a fault that unwinds.

---

## 9 · Conditions and loops

Carried from `2026-08-12/continuation-language/conditions-and-loops.md`, which stands.

> **Failure and control flow are independent axes. The contract asks whether a return may be
> consumed at all; the condition asks what it says.**

```text
A CONDITION      reads a returned observation that ALREADY SATISFIED its contract, and chooses
                 among continuations. It cannot rescue a mismatch — that is the other axis.
A LOOP           is repeated occurrence at ONE position. Not a new position per iteration, and
                 not a new contract.
THE DOT          `itinerary.plan` is a PATH INTO an observation, not a message send. It resolves
                 within what already landed and emits no MOVE.
```

Neither adds a primitive. Both are ordinary reads over things that have already crossed.

**AND/OR gates are deferred to this language and are not needed yet.** The current rule settles
the question after the fact: if the remembered observation has no dependency on a potential,
that potential is not recorded. All five contributing is AND in effect; one contributing is OR
in effect. A gate would be a declaration *about* the work; the trace is a record *of* it.

---

## 10 · Leaves

A leaf has no child signatures and still has a continuation:

```text
Result leaf(Input input)
children { }
returns  { result: Result }
```

ACTIVATE loads its contracts and no code; the observation comes from the crossing. **Leaf is a
position, not a kind** — it means nothing has been invoked under the current expression, and it
is the one place raw content may enter without deriving from a parameter or a child return.

How leaf origination appears in the surface remains open. It must not let a composing body
smuggle uncaused values into an otherwise composed result.

---

## 11 · What the language may not express

```text
undeclared child names                    ambient addresses or transports
arbitrary graph lookup                    unbounded world access
values with no observation provenance     authority wider than Horizon
self-declared type satisfaction           durable mutation because code was compiled
a jump to an arbitrary point in a body    retention of its own results
```

The last two are new, and both are **inexpressible rather than forbidden**, which is the form
this design prefers wherever it can get it. There is no operand that names a position inside a
body, and there is no verb a non-root can utter that retains anything.

---

## 12 · Acceptance test

```text
1   Every name in the body traces to a method parameter or a child return.
2   Every child call names a position currently supplied in the frame.
3   Every argument satisfies that position's contract.
4   Every binding is created only after a returned observation satisfies its contract.
5   Every outward field satisfies the supplied return contract.
6   EVERY CHILD CALL IS ONE MOVE WITH A REQUEST HEADER. Reach is a stage, never a statement.
7   A TYPED ASSIGNMENT NAMES A DESTINATION REGISTER AND ITS CONTRACT. It retains nothing.
8   RETURN IS ONE MOVE WITH A RESPONSE HEADER. It is not Remember.
9   Type claims cannot be forged by observation content.
10  Contract mismatch blocks dependent continuation without requiring an Error value.
11  The dependency graph is reconstructible from names and contracts.
12  Every child return is checked at the RECOGNIZE port, in the parent's frame.
13  Address, transport, ambient graph access, and Horizon mutation are inexpressible.
14  The body cannot change its own signature, its child signatures, or its return contract.
15  A leaf fires terminal behaviour without inventing a declared leaf kind.
16  Compilation creates provisional behaviour, not durable history.
17  TWO STATEMENTS WITH NO DEPENDENCY BETWEEN THEM MAY EXECUTE IN EITHER ORDER OR AT ONCE.
```

If any fails, this has ceased to be a typed continuation over observable causes and has become
either an unrestricted programming language or decorative syntax over hidden runtime behaviour.

---

## 13 · Open

```text
CONCRETE SYNTAX AND        `researchingCompilersWithJVM/` is the upstream research. The parser
THE PARSER                 builds the AST; the compiler walks it emitting MOVEs.

THE SLIDING WINDOW IN      the `children` block is now a frontier, not a declaration. What a body
THE SURFACE                sees when it refills mid-execution, and whether that is visible in the
                           syntax at all, is unwritten.

ESCALATION AFTER A         the 08-12 "progressive causal zoom" assumed per-frame lifetimes and
MISMATCH                   waking a suspended parent. Under bulk teardown a failed possibility is
                           not woken — it contributes nothing and goes with everything else. WHAT
                           REPLACES ZOOM IS UNRESOLVED, and it is the largest gap in this document.

RETRY                      whether it reuses a position or creates a new occurrence.
LEAF ORIGINATION           how it is written without weakening provenance.
NEGOTIATION                what it means mechanically, given one request and exactly one response.
```
