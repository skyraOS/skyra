# 02 · The Entity Boundary

**Date:** 2026-08-11
**Status: CHANGE.** What the kernel calls into. Distilled from `../../interface-revision.md`.

> **The kernel must HOLD the crossing. It still may not READ it.**

---

## The change

```go
// BEFORE
type Interpreter interface {
    Step(x Expectation, f Frame) (Emission, error)
    Monitor(o *Observation)
}

// AFTER
type Interpreter interface {
    Assemble(x Expectation, f Frame) (Rendering, error)   // ENTITY-bounded: how much it can take
    Project(r Rendering, h HostID) (Delivery, error)      // HOST-bounded: what form it can realize
    Step(d Delivery) (Emission, error)                    // still the only PRODUCER
    Monitor(o *Observation)                               // unchanged. returns nothing.
}

type Rendering interface{ isRendering() }   // opaque. no method the kernel can call.
type Delivery  interface{ isDelivery()  }
```

**Marshalling stays entity-side.** What moves is who **calls** it, and therefore who **holds** the
result.

## Why — three things the design already requires and cannot do

```text
THE MONITOR INVARIANT    interfaces.md: "assemble(frame) == assemble(frame₀) + monitor* … BUILD
IS SELF-REPORTED         BOTH AND COMPARE." with assembly behind the interface the only place that
                         comparison can go is INSIDE THE ENTITY. the build has the entity grading
                         itself, and reported "the invariant held" — meaning the entity said so.
                         A SELF-REPORT AS GROUNDING, which the set refuses everywhere else.

RECOVERY IS SPECIFIED    kernel.md: "a host dies mid-invocation, THE KERNEL RE-ASSEMBLES in full
AND UNAVAILABLE          from the engram and the returns it already has." it cannot. assembly is
                         not reachable from it.

PROJECTION CANNOT        world.md makes fail-loudly the single implementation requirement
FAIL LOUDLY              everything rests on. projection happens inside the adapter, so the kernel
                         cannot enforce the one thing it is told to enforce.
```

## Why the types are opaque

Seeing is not reading.

```text
THE KERNEL ALREADY   Observation.Content passes through it constantly and it interprets none of
HOLDS CONTENT        it. NO SEMANTICS means it makes no judgement, not that no bytes cross it.

WHAT ACTUALLY        two lines reading content to truncate it for a log. holding an opaque blob
BREACHED THE RULE    is fine; reaching into it is not.
```

So the crossing types carry no methods the kernel can call. The incapacity is structural rather than
a policy of restraint.

---

## Struct changes that go with it

```go
// entity/codex.Interpreter
- base      map[ids.ObservationID]bool   // the invariant check moves to the kernel
- delta     map[ids.ObservationID]bool
- Divergent []string
- Run       *trace.Run

// entity/runtime.Entity
- Run *trace.Run
```

The entity stops holding its own verification state and its own trace handle. Both were symptoms of
the kernel being blind.

## What it buys

```text
THE INVARIANT BECOMES     the kernel holds both forms and compares them. no entity grades itself.
CHECKABLE

RECOVERY BECOMES          exactly what kernel.md already describes.
POSSIBLE

REPLAY STOPS DEPENDING    the exact bytes sent to an entity currently exist only in a trace file
ON SOMETHING OUTSIDE      whose package comment says it is not part of the design. kernel-held,
THE DESIGN                re-running means re-sending the identical rendering.

PROJECTION IS             the kernel calls it, so the kernel sees it error.
ENFORCEABLE

TRACING GETS A KEY        instrumentation currently runs from three independent handles with no
                          shared schema — kernel logs topology, adapter dumps prompts, nothing
                          joins them, and concurrent possibilities on one engram are
                          indistinguishable. with the crossing kernel-held the span is available.
```

## What it costs

**Two methods to four**, in the surface the design keeps smallest. Real, and it should be agreed to
rather than inherited.

Against it: nothing here adds a *concept*. Assemble and project are already named, already
specified, already bounded by different owners. The build left both inside a black box; the growth is
the cost of them being visible.

---

## Rejected

```text
FIXING IT IN THE TRACE   the trace is outside the design and says so. making it the place the
FACILITY                 crossing is observable makes an out-of-band facility load-bearing for
                         recovery and replay.

THE KERNEL DOING THE     assemble is ENTITY-bounded — how much it can take — and a kernel that
ASSEMBLY ITSELF          builds prompts knows about models. it must CALL, not implement.

Rendering AS A string    the kernel could then read it, and would.

COLLAPSING Assemble      different owners, different bounds. the entity says how much it can take;
AND Project              the host says what form it can realize.
```

## The general failure this is an instance of

> **What the kernel must HOLD is a different question from what it may READ. The set only ever
> answers the second.**

A boundary drawn for one good reason quietly took something else with it. The kernel is forbidden
from reading content, so assembly went behind the interface — and recovery, replay, honest
verification and correlated tracing went with it, none of which required reading anything.
