# durable/03 · The Crossing

**Package:** `entity`

The interpreter boundary. The kernel must HOLD the crossing. It still may not READ it.

---

## 1 · The interface

```go
type Interpreter interface {
    // ASSEMBLE — engram state → what this entity reads. Bounded by the ENTITY:
    // how much it can take. The kernel calls it, so the kernel can compare the
    // full form against the delta form, and can re-assemble after a host dies.
    Assemble(c contract.Child, f Frame) (Rendering, error)

    // PROJECT — rendering → the host's medium. Bounded by the HOST: what form it
    // can realize. MUST FAIL LOUDLY.
    Project(r Rendering, h ids.HostID) (Delivery, error)

    // STEP — still the only method that PRODUCES anything.
    Step(d Delivery) (grammar.Emission, error)

    // MONITOR — the delta form of assemble. Returns nothing, by constraint.
    Monitor(o *observation.Observation)
}

type Rendering interface{ isRendering() }   // opaque. the kernel holds it, cannot look inside.
type Delivery  interface{ isDelivery()  }
```

Carried from `../../../semantics/the-crossing.md`, which is marked **UNAFFECTED** by the 08-12/13 work:
*"what the kernel must HOLD is still a different question from what it may READ."*

One change: `Assemble` takes a `contract.Child` rather than an `Expectation`, following
`01-engram.md` §2. What an entity is asked to interpret is now the parent-owned contract at the
position, not a free-floating expectation on the child.

---

## 2 · Why four methods and not two

The defect it fixes:

```go
Step(x Expectation, f Frame) (Emission, error)   // the runtime today
```

Frame in, emission out. **The kernel hands over a frame and gets back a decision, and never sees
what produced it.** Three things the design already requires become impossible, and none of them
needed the kernel to read anything:

```text
THE MONITOR INVARIANT     "assemble(frame) == assemble(frame₀) + monitor* … build both and
IS SELF-REPORTED          compare." with assembly behind the interface, the only place that
                          comparison can go is INSIDE THE ENTITY — the entity grading itself.
                          That is a self-report as grounding, which the set refuses everywhere
                          else.

RECOVERY IS SPECIFIED     kernel.md: a host dies mid-invocation, "the kernel re-assembles in full
AND UNAVAILABLE           from the engram and the returns it already has." The kernel cannot.
                          Assembly is not reachable from it.

PROJECTION CANNOT FAIL    world.md makes this the single implementation requirement everything
LOUDLY                    rests on — an unsupported modality must error rather than silently
                          drop. Projection happens inside the adapter, so the kernel cannot
                          enforce the one thing it is told to enforce.
```

**Marshalling stays entity-side.** *Marshalling is what interpreting is* is untouched — a model's
assemble builds a prompt, python's builds stdin and argv, a human's builds a message. What moves is
**who calls it**, and therefore who holds the result.

```text
THE COST, STATED PLAINLY: the Interpreter interface grows from two methods to four, in the one
surface the design keeps smallest. Nothing here adds a CONCEPT — assemble and project are already
named, already specified, already bounded by different owners. The growth is the cost of them
being visible.
```

---

## 3 · Seeing is not reading

The objection this keeps meeting, and it dissolves on inspection:

```text
THE KERNEL ALREADY HOLDS    Observation.Content passes through it constantly and it interprets
CONTENT                     none of it. NO SEMANTICS means it makes no judgement, not that no
                            bytes cross it.

WHAT ACTUALLY BREACHED      two lines that READ content to truncate it for a log —
THE RULE                    shortExpr(Expectation.Expression) and oneLine(Content.Render()).
```

So the crossing types carry **no methods the kernel can call**. `Rendering` and `Delivery` are
opaque interfaces with unexported markers, which makes the incapacity structural rather than a
policy of restraint. P16.

```text
REJECTED   Rendering AS A string — the kernel could then read it, and would.
```

**And this is where `../contracts/03-parser.md` §3 collides with it.** If the kernel is to parse
returns without a crossing, it must evaluate something about the returned observation. Provenance it
may read, because it set it. Content it may not. Shape sits between, and whether computing it counts
as reading is the unresolved question. That is H1, and this document is the reason it is hard.

---

## 4 · Step returns text again

`../../../overnightBuild/stops.md` stopped E1/E2 on exactly this signature:

> `Step(d Delivery) (Raw, error)` … **the kernel calls `grammar.Parse`**, not the adapter.

and hit:

```text
entity/runtime/runtime.go:244   byIndex    "reach": content.EngramRef{…}
entity/runtime/runtime.go:278   byRecency  "reach": content.EngramRef{…}
entity/runtime/runtime.go:312   create     "reach": content.EngramRef{…}

All three resolvers commit content containing a content.EngramRef. THE GRAMMAR HAS NO SURFACE FORM
FOR ONE.
```

**Under provenance typing that stop dissolves** — see `../contracts/01-observation.md` §4. With
reach carried as `Provenance` rather than as a content value, there is nothing for `grammar.Parse`
to need a surface form for, and `Step` can return raw text with the kernel doing the parse.

The signature above still returns `grammar.Emission` rather than `Raw`, because that is the shape
the crossing document specifies and this outline is not deciding the stop on its own. But the
blocker it recorded is gone, and whoever picks E1 back up should know that.

```text
THE STOP'S OTHER FINDING STILL STANDS: stops.md also rejected "code entities skip the parse path,"
because two classes of entity would break the kernel's hold on what produced a decision.
../../../continuation-language/surfaces-and-solidification.md agrees and adds the condition under
which two MODES are acceptable where two CLASSES were not:

    THE CONTINUATION LANGUAGE DESUGARS TO EXACTLY THE PRIMITIVES — nothing more, nothing less.
```

---

## 5 · What the crossing must preserve

`../../../causality/observable-causality.md` states the entity-boundary requirements the whole causal
claim rests on:

```text
ISOLATION             no graph access outside the frame.
NO HIDDEN CONTINUITY  no private memory of this caller between invocations.
OPAQUE CROSSING       the kernel may hold delivery artifacts without interpreting their meaning.
LOUD FAILURE          missing reach or unsupported projection cannot masquerade as successful work.
```

> *"Hidden continuity, ambient graph access, or undeclared context would not be a subtle attribution
> problem. It would be a violation of the entity contract because the frame would no longer describe
> the computation's inputs."*

The four map onto the four methods: `Assemble` bounds what the entity reads, `Project` is where loud
failure is enforced, `Step` is the only producer, and `Monitor` returning nothing is what keeps the
delta form from becoming a back-channel.

---

## 6 · What holding the crossing buys

```text
THE INVARIANT BECOMES     the kernel holds assemble(frame) and assemble(frame₀)+monitor* and
CHECKABLE                 compares them. no entity grades itself.

RECOVERY BECOMES          re-assemble in full from the engram and the returns, continue the
POSSIBLE                  possibility elsewhere.

REPLAY STOPS DEPENDING    the exact bytes sent to an entity currently exist only in a trace file
ON SOMETHING OUTSIDE      whose own package comment says it is not part of the design.
THE DESIGN

PROJECTION IS             the kernel calls it, so the kernel sees it error.
ENFORCEABLE

TRACING GETS A REAL       with the crossing kernel-held, the span is available and the join is
CORRELATION KEY           exact.
```

The third and fifth together retire `trace.Run`, which `../../09-trace.md` places outside the design
by its own admission. Replay and correlation stop needing an out-of-band facility, which is the one
argument for keeping it.

**And replay is not a convenience here.** `observable-causality.md` §2 makes necessity testable *by
intervention* — re-running with an observation removed — which requires that the exact conditions be
reconstructible. That capability is what holding the crossing provides, and it is why this document
sits in `durable/` rather than in `kernel/`.
