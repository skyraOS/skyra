# Energy

**Date:** 2026-08-14
**Status: PROPOSAL.** Names the dissipative laws the design already runs on, establishes the field
they are measured in, and proposes energy as a bound provisioned downward.

> **Energy is dissipative, not conserved. The field is TOKENS. Every other cost is energy of a
> different type, and the design has already made architectural decisions on energy grounds without
> naming the unit.**

---

## This directory

```text
distance.md   MEASUREMENT.  what a crossing and a remembered result COST.
              retrospective · feeds ranking · C(r) · the namespace cost account

README.md     LAW AND BUDGET.  what may be spent, and by whom.
              prospective · feeds bounding and eviction
```

[`distance.md`](distance.md) **controls everything about measurement**, including `C(r)`, which
quantities are measurable, whether unlike quantities may be combined, and who owns the cost account.
Where this document touches those, it defers.

## What this respects

[`../../2026-08-09/possibilities.md`](../../2026-08-09/possibilities.md) rules conservation out, and
this does not reopen it:

> *"Nothing is conserved — N in, 1 out, and information is destroyed on purpose… If the metaphor
> becomes energy it must be **dissipative** … conservation is a law this design does not have, and
> assuming it sends someone looking for where the missing observations went."*

`2026-08-08/primitives.md` says the same; `2026-07-22/time.md` a third time. **Conservation and
accounting are different things** — a dissipative system accounts exactly, because dissipation is
what is being accounted for. Nothing below asserts a conserved quantity.

## The field is tokens

The dominant cost is inference, and the design already treats it that way.

```text
skyra-dna-v5/frame/frame.go

    "The frame is also a BUDGET boundary. Everything a possibility sees ships to the entity's
     host on every step, so what it may see is simultaneously the learning mechanism and the
     per-step transfer cost, and the two pull against each other."
```

Two places show the unit was already load-bearing:

```text
graph.Disperse    "It is O(n · depth) BY DESIGN … IT COSTS NO MODEL CALL."
                  the walk is defended as free. that is only an argument if model calls are the
                  currency and other compute rounds to zero.

host.StorageOps   exists because "spawn alone is expressively sufficient and COST-CATASTROPHIC:
                  'read a file' becoming 'spawn a reader' is a process per read."
                  an entire capability class added on energy grounds.
```

**An architectural decision has already been made on this field.** Naming it changes nothing that
happened; it makes the next one arguable instead of intuitive.

```text
DOMINANT TYPE     tokens. what most executions actually spend.
NOT A WEIGHTING   naming the dominant type is not an exchange rate. distance.md's ruling stands
                  in full: "the kernel must not invent that weighting."
```

### Mass and energy

Frame size is a *static* property; a crossing moves it.

```text
MASS      the frame. what a possibility weighs at rest.
ENERGY    mass moved across a boundary.
COST      Σ mass, over crossings.
```

This derives a result rather than decorating one.
[`../continuation-language/surfaces-and-solidification.md`](../continuation-language/surfaces-and-solidification.md)
observes that primitive mode is quadratic because `Returned` accumulates — **mass grows, and it is
paid at every crossing.** Solidification does not reduce mass; it reduces crossings.

```text
⚠️ MASS IS CONSERVED IN PHYSICS AND IS NOT CONSERVED HERE. Frame mass is destroyed at every
   commit — N in, 1 out. If the word is used, that has to be said in the same breath, or someone
   will reason from conservation of mass and go looking for where the observations went. Which is
   the exact failure 08-09 warns about, arriving through a borrowed word.
```

### The tik is the clock

A **tik** is one movement of the kernel over a possibility: the entity is handed a frame and returns
an emission. `skyra-dna-v5` calls this `Interpreter.Step` today — **tik is the name for it as a unit
of time**, rather than as a loop iteration.

```text
ONE TIK  =  ONE CROSSING  =  ONE FRAME SHIPPED
```

Which is what makes it the natural clock: a chain's time base should be the thing it actually
spends. [`../syscall-read-example/outcomes/When-outcomes.md`](../syscall-read-example/outcomes/When-outcomes.md)
§1 already refused wall clock as the basis for causal **when** — *"WHEN is causal position, not
wall-clock time."* A tik is that position, counted.

### The derived quantities

```text
ENERGY           Σ tokens over the chain             TOTAL.   what it cost.
TIKS             crossings.                          the chain's own clock.
MEAN MASS        tokens / tik                        MEAN.    frame bloat.
POWER            tokens / second                     RATE.    intensity.
USEFUL FRACTION  energy that rose / energy spent     RATIO.   waste.
```

`energy / tiks` collapses to mean mass **by construction** — a tik costs exactly the mass it ships.
So power cannot be derived from the tik clock; it needs wall time, which `distance.md` already
admits for measurement (*"LATENCY — wall-clock from reach to handle"*) even where `When-outcomes.md`
refuses it for causal position.

**Mean mass** has the most direct claim on an existing tension. `frame.go` says what a possibility
may see is *"simultaneously the learning mechanism and the per-step transfer cost, and the two pull
against each other."* Mean mass is that pull, as a number.

**Useful fraction** is the one the budget exists for. A stalled head and a working loop spend energy
at similar rates, so neither power nor mean mass separates them — the ratio does, because a stalled
head trends toward zero. It is 08-09's `used / seen` compression ratio denominated in tokens, and
its warning carries over intact:

> **Low is not bad.** It is exploration that discarded correctly, or a bad query, and it means
> something only differentially.

### The types are not interchangeable

```text
TOKENS       inference. frame shipped in, emission out.       THE PRIMARY TYPE.
WALL CLOCK   a group is a barrier; a slow member holds it.
PROCESS      spawn. the reason StorageOps exists.
STORAGE      read and write.
TRANSPORT    reaching a host that is not this one.
```

All energy, and provably not one axis:

> **Concurrency is the only operation that trades tokens for time.** N children in a group cost the
> same tokens as sequential and less wall clock.

An operation that moves one and not the other is a proof the two do not collapse. `distance.md`
reaches the same conclusion from measurement — *"three axes, and they do not collapse into a scalar
on their own"* — and separately splits time in two: *"REACH, time to obtain a handle ← execution
distance. WORK, time to obtain an answer ← not distance at all."*

```text
⚠️ TIME IS NOT POTENTIAL ENERGY. Potential is storable; time is not — it passes whether or not it
   is spent. And POTENTIAL IS ALREADY TAKEN: primitives.md has Think "create potential" as bounded
   ranked candidates, Decide convert one, and Reach make it live. That is the potential/kinetic
   pair in this design, and it already behaves correctly.
```

## The laws that already exist

Three, none previously named as such:

```text
AUTHORITY      Reach: "a parent gives a SUBSET … monotonic narrowing is the whole rule:
               A SELECT MAY DROP, IT CAN NEVER MINT."   Horizon: "can only narrow."
                                                          DISSIPATIVE. LAW EXISTS.

PROVENANCE     "nothing may appear from nowhere" — every value traces to a parameter or a
               child return.
                                                          NON-MINTING. LAW EXISTS.

OBSERVATIONS   N in, 1 out, destroyed on purpose. used/seen is "a compression ratio …
               a gauge, not a verdict."
                                                          DISSIPATIVE. LAW EXISTS (08-09).

ENERGY         measured by distance.md. NOT BOUNDED BY ANYTHING.
                                                          NO LAW.
```

Each of the three has one shape: **something enters at a boundary, narrows or is destroyed on the
way down, and cannot be minted from inside.** Energy is the quantity that most obviously behaves
this way and the only one with no law written for it.

### Credit is the deliberate anti-law, and that is consistent

```text
graph.Disperse    one commit → an Acceptance at EVERY accepting ancestor. O(n · depth) BY DESIGN.
```

Credit multiplies on purpose; the repetition is what populates the hop profile. That is consistent
here for one reason: **credit is a record, not energy.** Records may multiply because multiplying
them is free — which is precisely what *"it costs no model call"* asserts.

Worth noting where the one permitted multiplication led.
[`../semantics/second-kind-of-credit.md`](../semantics/second-kind-of-credit.md) records the
consequence — *"one evidential origin wearing forty records"* — so the single quantity allowed to
multiply is also the one that produced a problem nobody has closed.

## Provisioning: energy narrows downward

Proposed, and it is the shape authority already has.

```text
A PARENT PROVISIONS a subset of its energy to a child position.
A CHILD MAY SPEND OR SUBDIVIDE IT. IT MAY NOT MINT.
```

Energy enters at Approach, where `From` and Horizon are already established, and where
`distance.md` already puts the cost account: *"the durable subject of acquisition cost is the
namespace that paid for the work."* Nothing inside a chain can create more.

This is the half `distance.md` does not cover. Measurement says what was spent; a budget says what
**may** be spent, and only the second can bound anything.

### What it dissolves

```text
LIFETIME BOUNDS   ../continuation-language/conditions-and-loops.md records that a condition-gated
                  loop makes lifetime DATA-DEPENDENT while MaxSteps and MaxDepth are fixed, so
                  evicting a healthy long loop and evicting a stalled head become one event.
                  A BUDGET SEPARATES THEM: a working loop spends and produces observations that
                  rise; a stalled head spends and produces nothing. Eviction becomes exhaustion
                  rather than an arbitrary constant.

THE VERIFIER      ../misc/os-parallels.md §16 asks whether the kernel should admit a continuation
                  it cannot prove terminates. METERING IS THE OTHER ANSWER, and the one
                  proportionate to the stakes: a program that cannot outrun its budget needs no
                  termination proof. eBPF carries both — a verifier AND an instruction limit — and
                  only the second survives a callee whose behavior is authored elsewhere.
```

This is already the design's instinct: *"a RUNAWAY is depth or breadth past a bound, and it is an
EVICTION decision."* The answer was always operational. Energy makes the bound principled rather
than a constant someone chose.

## Open

```text
WHERE ENERGY ENTERS     Approach is the natural boundary, but nothing gates Approach today
                        (2026-08-09/questions.md §5, and kernel.Ingress says so plainly). A budget
                        on an ungated admission point is a budget anyone can open.

WHAT A LOOP EARNS       a session-long loop cannot be provisioned up front. Whether energy is
                        replenished at a boundary, drawn continuously, or a loop is exempt is
                        unspecified.

IS ENERGY READ AT       it measures what was spent, which is the shape second-kind-of-credit.md
SELECTION               refuses for credit — "a tag read at selection time." Cheap-therefore-good
                        is that tag under another name. STRONG PRIOR: read by reflection and by
                        eviction, never by selection.

BUDGET VERSUS THE       distance.md leaves open whether cost is a second durable append path with
COST ACCOUNT            the namespace as subject. A provisioned budget needs the same append and
                        may be the same mechanism. Deciding one probably decides the other.

REPLAY AND CACHING      a same-question hit costs no energy. Whether that is recorded as a saving,
                        and against which invocation, is unspecified.
```
