# Conditions and loops

**Date:** 2026-08-14
**Status: PROPOSAL.** Branching and iteration in the continuation language, the read path they
require, and why neither adds a primitive.
[`continuation-language.md`](continuation-language.md) lists both as open — *"the syntax for
explicit concurrency… branching and alternate continuations… whether loops are admitted, bounded, or
compiled as repeated positions."* This answers the branching and loop halves.

> **Failure and control flow are independent axes. The parser asks whether a return may be consumed
> at all; the condition asks what it says. A loop reads the second and only touches the first if the
> child actually breaks.**

---

## Two axes

```text
PARSER      does this return satisfy the contract?      1/0.   FAILURE AXIS.
CONDITION   what does the return say?                          CONTROL-FLOW AXIS.
```

They are orthogonal, and collapsing them is the error worth naming. A loop that ends because its
condition went false has not failed:

```text
{ seated: false }

    the observation is well-formed
    the return contract is satisfied
    the parser returns 1
    the condition is false, so the loop exits

NOTHING IS BROKEN.
```

If the condition rode on the parser, *"the user stood up"* and *"the child is broken"* would arrive
as the same signal. The loop is the case that makes the separation necessary, and it retroactively
justifies keeping admissibility distinct from transition selection everywhere else.

## A loop is repeated occurrence at one position

```text
execute child → return observation → evaluate condition → execute child
```

That is the whole mechanic. The loop is an instruction to the kernel to keep occupying one child
position while the condition holds.

```text
DESUGARS TO     repeated Decide at one position, with the condition selecting whether to repeat.
ADDS            nothing. every step is an existing primitive.
```

This satisfies the desugaring test in
[`surfaces-and-solidification.md`](surfaces-and-solidification.md) — the language may not express
anything the primitives cannot.

### Retry is the degenerate case

`continuation-language.md` also asks *"whether retry reuses or creates a child position
occurrence."* Under this reading it is the same shape with a different gate:

```text
LOOP     one position, re-occupied, gated by a condition over the child's return.
RETRY    one position, re-occupied, gated by the repair frame after a parser returns 0.
```

Both accumulate history at one position, which is where the design already wants learning to land.

## The condition is parent-authored

The predicate is written by the parent, in its continuation, and evaluated by the kernel over the
returned observation's content.

This does not breach *the kernel holds the crossing but may not read it*. The kernel already
traverses content structurally and always has:

```text
skyra-dna-v5/kernel/kernel.go    runGroup calls ret.Content.Reach() on EVERY return
skyra-dna-v5/content/content.go  Cites() scans content for Refs; Shape() folds it structurally
```

None of these has ever been treated as a breach, because the prohibition is on **interpretation**,
not traversal. And the predicate was authored by the parent, so the kernel is not deciding what
matters — it is evaluating something the parent already decided.

```text
THE PARENT      authors the predicate.        SEMANTICS.
THE KERNEL      evaluates it over content.    MECHANISM.
```

## The read path: the dot is a path, not a message send

Conditions require reading *into* an observation, which is the first construct in the language that
is not whole-observation dataflow. Every value in a body until now is a whole observation bound to a
name.

The notation already exists in the system. `content.collectShape` builds dotted paths
(`path + "." + k`), so `plan.total` already means *this location in content* everywhere else.

The language should use the dot the same way:

```text
flights.isEmpty()      REJECTED. the observation answers a question about itself —
                       the child self-certifying, one level down.

flights.items.length   the parent reads a location and decides what it means.
```

> **The parent writes the manipulation of the child's return.** The observation stays inert data.
> "Empty" is a judgment, and judgments belong to the consumer.

This is the ownership rule from
[`engram-contracts-and-composition.md`](engram-contracts-and-composition.md) applied one level
further in. It already governs the input contract, the return contract, and the parser; the
predicate is the same rule reaching content.

### No getters — the return contract is the read surface

A declared accessor would be a second statement of something the contract already makes. `returns {
plan: TravelPlan }` names what comes back, so the readable locations are the contract's locations.
Reading outside it is the same class as ambient access, which the language already refuses.

The strictness comes for free, and it buys a guarantee from ordering alone:

> **Field access cannot fail.** The parser runs before any read, and contract satisfaction
> guarantees those locations exist. No missing-field case, no null, no defensive read.

## What this costs

Nothing per iteration. The entity emits the continuation once and the kernel drives the loop:

```text
N ITERATIONS     one emission. zero crossings, as long as each parse lands.
```

Which is the same property `surfaces-and-solidification.md` states for straight-line composition —
*success costs no crossing* — extended over time. A session-long loop costs one emission plus
whatever the body itself costs.

## Impulses are not required

`skyra-dna-v5/kernel/possibility.go` holds a provisional subscription mechanism:

> *"Notices that a referent behind a handle already in frame has moved… an impulse is a notice that
> a referent MOVED, and the referent is still moved. The next assemble re-reads current world state.
> It is a re-read, not a lost event."*

It is tempting here and it is not needed. **A polling loop re-observes by executing** — the child's
return carries the condition, so there is no ambient world read and no second channel. *Nothing may
appear from nowhere* holds unchanged.

Impulses would matter only for a loop that **waits without executing**. Recorded so the mechanism is
not reached for by default, and so the case that would justify it is stated.

## What this settles

```text
BRANCHING              a condition over a returned observation's content, authored by the parent.
LOOPS                  repeated occurrence at one child position, gated by that condition.
RETRY                  the same shape, gated by the repair frame instead.
THE READ PATH          dotted paths into content, bounded by the return contract.
WHY IT ADDS NOTHING    both desugar to repeated Decide; the condition is evaluated, not interpreted.
```

## Open

```text
ONE CONSTRUCT OR TWO       `if` evaluates a condition once; `while` re-evaluates and re-occupies.
                           Whether they share one surface form is unspecified.

A LOOP THAT WAITS          a loop that blocks on a world change rather than polling needs the
                           impulse mechanism and a way to express "wait" without occupying a
                           position. Not proposed here.

LIFETIME BOUNDS            a condition-gated loop makes possibility lifetime DATA-DEPENDENT, where
                           MaxSteps and MaxDepth are fixed. Eviction of a legitimately long-running
                           loop and eviction of a stalled head become the same event under the
                           current bounds, and they are not the same thing. This is one of the
                           three decisions the 08-12 work order fenced off ("DO NOT CHANGE when a
                           possibility dies") and it must be opened deliberately.

STAGED EFFECTS OVER        Root is the transaction boundary and nested transactions were refused by
A LONG-LIVED ROOT          name, so a long root already holds staged effects for its duration. A
                           long-running loop makes that visible rather than introducing it — but it
                           makes it visible at a scale the model has not been tested at.

WRITING THE PREDICATE      what operators the predicate language admits, and what keeps it from
                           becoming an unrestricted expression language inside a continuation whose
                           whole claim is that it is boring.
```
