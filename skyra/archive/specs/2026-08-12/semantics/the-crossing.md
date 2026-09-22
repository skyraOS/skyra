# The Interface, Revised

**Date:** 2026-08-11
**Status: PROPOSAL.** Revises the `Interpreter` boundary in `2026-08-09/interfaces.md`.

> **The kernel must HOLD the crossing. It still may not READ it.**

> **CARRIED FORWARD** from `specs/2026-08-11/semantics/interface-revision.md`.
>
> **Status here: UNAFFECTED** by the 08-12/08-13 work. Renamed for what it claims rather than
> when it was written. Nothing in the seven primitives, the entity criteria, or the reach work
> touches it — what the kernel must HOLD is still a different question from what it may READ.

---

## The defect

```go
Step(x Expectation, f Frame) (Emission, error)
```

Frame in, emission out. **The kernel hands over a frame and gets back a decision, and never sees
what produced it.** Assembly happens inside the adapter; the raw response is parsed inside the
adapter. Three things the design already requires become impossible, and none of them needed the
kernel to read anything.

### The monitor invariant is self-reported

`interfaces.md` states it and says how to check it:

> `assemble(frame) == assemble(frame₀) + monitor*` … **build both and compare.**

Because assembly lives behind the interface, the only place that comparison can go is **inside the
entity**. So the build has the entity grading itself on whether it holds state the kernel can
reconstruct, and the build's report — *"the invariant held on every step of every run"* — means the
entity said so.

**That is a self-report as grounding**, which the set refuses everywhere else. If the kernel calls
assemble it holds both forms and compares them itself. Honest witness, same as everywhere.

### Recovery is specified and unavailable

`kernel.md`, on the invariant paying for itself:

> a host dies mid-invocation, **the kernel re-assembles in full** from the engram and the returns it
> already has, and the possibility continues.

The kernel cannot. Assembly is not reachable from it. **The spec specified the right thing and the
implementation put it in the wrong place.**

### Projection cannot fail loudly

`world.md` makes this the single implementation requirement everything rests on — an unsupported
modality must error rather than silently drop, because otherwise the invocation "succeeds," nobody
perceived anything, and the record reads as success.

Projection also happens inside the adapter, so **the kernel cannot enforce the one thing it is told
to enforce.**

---

## Seeing is not reading

The objection this will keep meeting, and it dissolves on inspection.

```text
THE KERNEL ALREADY HOLDS    Observation.Content passes through it constantly and it interprets
CONTENT                     none of it. NO SEMANTICS means it makes no judgement, not that no
                            bytes cross it.

WHAT ACTUALLY BREACHED      two lines that READ content to truncate it for a log —
THE RULE                    shortExpr(Expectation.Expression) and oneLine(Content.Render()).
                            holding an opaque blob is fine; reaching into it is not.
```

So the crossing types should carry **no methods the kernel can call**. The incapacity becomes
structural rather than a policy of restraint, which is the form this set wants:

```go
type Rendering interface{ isRendering() }   // opaque. the kernel holds it and cannot look inside.
type Delivery  interface{ isDelivery()  }
```

---

## The shape

```go
type Interpreter interface {
    // ASSEMBLE — engram state → what this entity reads. Bounded by the ENTITY:
    // how much it can take. The kernel calls it, so the kernel can compare the
    // full form against the delta form, and can re-assemble after a host dies.
    Assemble(x Expectation, f Frame) (Rendering, error)

    // PROJECT — rendering → the host's medium. Bounded by the HOST: what form it
    // can realize. MUST FAIL LOUDLY. The kernel calls it, so the kernel can hold
    // it to that.
    Project(r Rendering, h HostID) (Delivery, error)

    // STEP — still the only method that PRODUCES anything.
    Step(d Delivery) (Emission, error)

    // MONITOR — the delta form of assemble. Returns nothing, by constraint.
    Monitor(o *Observation)
}
```

**Marshalling stays entity-side.** *Marshalling is what interpreting is* is untouched — a model's
assemble builds a prompt, python's builds stdin and argv, a human's builds a message. What moves is
**who calls it**, and therefore who holds the result.

**`Step` remains the only producer.** Assemble and project are transformations the kernel drives;
only `Step` yields an emission. The two-forms claim is unaffected.

---

## What it buys

```text
THE INVARIANT BECOMES     the kernel holds assemble(frame) and assemble(frame₀)+monitor* and
CHECKABLE                 compares them. no entity grades itself.

RECOVERY BECOMES          exactly what kernel.md already describes. re-assemble in full from the
POSSIBLE                  engram and the returns, continue the possibility elsewhere.

REPLAY STOPS DEPENDING    the exact bytes sent to an entity currently exist only in a trace file
ON SOMETHING OUTSIDE      whose own package comment says it is not part of the design. held by the
THE DESIGN                kernel, re-running a possibility means re-sending the identical
                          rendering — determinism is checkable and two runs are diffable.

PROJECTION IS             the kernel calls it, so the kernel sees it error.
ENFORCEABLE

TRACING GETS A REAL       instrumentation currently runs from three independent handles with no
CORRELATION KEY           shared schema — the kernel logs topology, the adapter dumps prompts, and
                          nothing joins them. concurrent possibilities on one engram are
                          indistinguishable in the trace. with the crossing kernel-held, the span
                          is available and the join is exact.
```

---

## What it costs

**The `Interpreter` interface grows from two methods to four**, in the one surface the design keeps
smallest. That is real and should be agreed to rather than inherited.

Held against it: nothing here adds a *concept*. Assemble and project are already named, already
specified, already bounded by different owners. What the build did was leave both inside a black box,
and the growth is the cost of them being visible.

---

## The general shape of the failure

Twice now the same thing: **a boundary drawn for one good reason quietly took something else with
it.** The kernel is forbidden from reading content, so assembly went behind the interface — and
recovery, replay, honest verification and correlated tracing went with it, none of which required
reading anything.

> **What the kernel must HOLD is a different question from what it may READ.
> The set only ever answers the second.**

---

## Rejected

```text
FIXING IT IN THE TRACE   the trace is outside the design and says so. making it the place the
FACILITY                 crossing is observable makes an out-of-band facility load-bearing for
                         recovery and replay, which is worse than the defect.

THE KERNEL DOING THE     assemble is bounded by the ENTITY — how much it can take — and a kernel
ASSEMBLY ITSELF          that builds prompts knows about models. it must CALL, not implement.

Rendering AS A string    the kernel could then read it, and would. an opaque type makes the
                         incapacity structural, which is the difference between this and a policy
                         of restraint.

COLLAPSING Assemble      they have different owners and different bounds — the entity says how
AND Project              much it can take, the host says what form it can realize. world.md
                         separates them for that reason and the reason survives.
```
