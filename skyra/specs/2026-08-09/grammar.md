# The Grammar

**Date:** 2026-08-09
**Status: SPEC.** Not built.

> **Two forms — SELECT and COMMIT. Everything else is a name.**

---

## The two forms

A possibility emits. It does not spawn, and it has no API.

```text
<select>
  <engramName cite=obsID entity=entID host=hostID> obsID, obsID </engram>
  <engramName cite=obsID entity=entID>             obsID         </engram>
</select>

<commit><observationID></commit>          an orchestrator NAMES
<commit>{ … }</commit>                    a leaf ORIGINATES
```

```text
SELECT   name one or more engrams, bind observations to each.   AN INVOCATION GROUP.
COMMIT   the value up, AND THE ONLY DURABLE WRITE.
```

That is the entire action vocabulary. **Select IS invoke** — there is no separate call form, which
makes *"selection happens at the call site"* literal rather than descriptive.

Because this is a grammar rather than an interface, the incapacity is a **parsing** property, not an
API property. A possibility cannot do anything but emit these two forms, and nothing else is
expressible.

## Invocation cardinality

> **Every selection awaits. Concurrency is expressed by how many possibilities are selected at once,
> not by a blocking mode.**

```text
CARDINALITY = 1    one causal step.
CARDINALITY > 1    a concurrent frontier chosen by the possibility.
```

```text
SELECT { A }  ·  await  ·  SELECT { B }        A → B, and B may cite A.

SELECT { A, B }  ·  await both                 neither needs the other's return to begin.
```

**The runtime does not find, select, or order the frontier.** It executes the group the possibility
already emitted. There is no `async`, no `await`, no execution flag — arity is the whole declaration.

### The frame boundary

Every member receives observations from the frame **as it existed when the group was emitted**.
Members cannot see one another's returns. When the whole group completes, its returns enter the
caller's frame and may shape the next selection.

That makes invocation history a partial order for free:

```text
A then B, and B cites A         a CONFIRMED dependency.
A then B, but B never cites A   possible unnecessary serialization.
A and B in one group            declared return-independence AT THAT POINT.
```

The last claim is narrow. Concurrent members may share inputs and contribute to the same eventual
return. They are independent only in that neither needs the other's return to begin.

### It is what makes irreducibly joint action representable

Two simultaneous releases are a group of two — separate possibilities, separate engrams, separate
credit, executed together. The alternative was hiding both inside one body, which works and destroys
the ability to learn about either one.

### The cost, accepted deliberately

**A group is a barrier.** One slow or trapped member delays the next selection even if the others
have returned enough to continue. Whether partial completion is ever necessary is open, and it must
not be smuggled in as an execution flag without a case that demands it —
[questions.md](questions.md).

## Every emission cites its sources

A select member has **two positions, and they are different roles**:

```text
<engramName cite=obsID> obsID, obsID </engram>
 ^^^^^^^^^^             ^^^^^^^^^^^^
 WHO                    WHAT
```

An observation that supplied the *name* is never bound as an argument and never composed, so without
a citation it falls through as unused — and that is the observation that made the invocation possible
at all.

`cite` fixes it in the grammar rather than in the metric. It also buys something larger: **capability
provenance is exact.** *How did you come to be able to name this?* is answerable for every invocation
the system ever made. Derived attribution cannot do that — two observations carrying the same ID are
indistinguishable, and you cannot tell which grant was exercised.

A name may also come from durable structure rather than an observation, and the distinction is a
signal:

```text
CITED FROM AN OBSERVATION   I had to discover this before I could act.
CITED FROM DURABLE          I already had the structure.
```

A possibility that must discover before every action has internalized nothing, and the ratio between
the two measures how well-structured it has become.

### `cite` is per MEMBER, never per group

```text
PER MEMBER   { A cite=obs7, B cite=obs9 }   two independent facts.
PER GROUP    cite=obs7 for { A, B }         ONE CAUSE FOR N SELECTIONS.
```

**Grouping is a scheduling fact and must not become a shared cause.** Selection credit is exact
precisely because a selection has one cause; a group-level cite reintroduces sibling blindness at the
one place the design has it solved.

It also breaks two things downstream: `(Q → P)` pairs are what recur and therefore carry statistics,
and a *set* does not recur the same way — and pruning is supposed to be finer than the engram, so
`(Q → P)` can go while `(R → P)` stays. With a group-level cite, only whole groups are prunable.

Same possibilities, same concurrency; each keeps its own reason for being there.

## The five questions an emission answers

```text
WHO     the possibility.                    named.
WHAT    the observations.                   bound as arguments.
WHERE   the entity.                         named.
WHEN    the tick of the expectation.        local.
WHY     what prompted it.                   the citation.
```

**`WHERE`.** Entity sits in every signature identically, which is why it never needed discovering —
the caller supplies it per *invocation*, so the same possibility runs in different places without
being a different possibility. It was never about the callee.

**`WHEN` makes time per-cell**, which lands proper time — ownerless since the Slice died — and keeps
it consistent with locality: no global clock, no component holding a model of the whole. **The cost
is that events cannot be globally ordered.** Cross-cell causality is reconstructible only through
citation chains, never through timestamps.

**`WHY` is split**, and collapsing it would lose the half that matters:

```text
WHO ASKED       the invoking possibility.   structural. free — the runtime knows the emitter.
WHAT PROMPTED   the observation cited.      emitted. the capability trail.
```

## Everything is by ID, so capability is scoping

A possibility holds **names, not content**. It can only name what is in frame: what it was passed, plus
what its own selects returned.

> **Capability is name resolution. You can select what you can name; you can name what you were
> given.**

Reach does not need its own type — it needs a scope. A possibility cannot address what was never put in
its frame.

**Reach must not travel in the data channel.** A returned entity stays type-distinct from a returned
observation:

```text
DELEGATION   reach arriving from an entity I CHOSE to invoke.   open() handing back a descriptor.
INJECTION    reach arriving in the data channel.                whoever can talk to me picks what
                                                                I can touch.
```

Same bytes, different provenance, and only the type keeps them apart.

## Identity

A durable engram owns its ID. An observation needs no separate minting — its identity is
**which engram produced it, in which span, at which position**. Engram and span together name the
opening; position names the emission within it.

Provenance is therefore intrinsic to the ID rather than a field beside it, and a returning child
hands its parent `(observation, engram)` — which is exactly what the parent needs to attach it. The
possibility itself is gone by then, and nothing durable should have needed it.

## One universal signature

```text
pos "pos-name" (obs obs[], entity) {
    // collection behaviour — the while loop
    // return behaviour — one shot
}
```

**Every possibility has this signature.** Not one per callee — one, universally. So there is no
per-possibility contract to discover, nothing for discovery to return beyond the name, and no
compatibility check before invoking.

> **Type safety is replaced by learning.** Anything may be invoked with anything; whether that was
> useful is what `used/seen` and credit measure.

---

## Rejected

```text
AN ENTITY[] PARAMETER     capability as a second parameter list. scoping does it better —
                          reach is a scope, not a type.

PER-CALLEE SIGNATURES     discovery returning parameter shapes. unnecessary once the signature
                          is universal.

A SEPARATE INVOKE FORM    select already is one.
```
