# 03 · The Entity Boundary

**Status: WORK ORDER.** The largest item. Do it last, in its own commits, so it can be dropped
whole without losing `01` and `02`.

> **This changes how an entity is called. It must not change when possibilities live and die.**
> `kernel.go` is touched at the interpreter call sites and nowhere else. `commit` still closes.
> `graph.Apply` is still called from the commit path.

---

## The defect being fixed

```go
Step(x graph.Expectation, f frame.Frame) (grammar.Emission, error)
```

Frame in, emission out. **The kernel hands over a frame and gets back a decision, and never sees what
produced it.** Assembly happens inside the adapter and the raw response is parsed inside the adapter,
which makes three things impossible that never required the kernel to read anything:

```text
THE MONITOR INVARIANT   assemble(frame) == assemble(frame₀) + monitor* can only be checked inside
IS SELF-REPORTED        the entity, so the entity grades itself on whether it holds state the
                        kernel can reconstruct. A SELF-REPORT AS GROUNDING.

RECOVERY IS             a host dies mid-invocation and the kernel is supposed to re-assemble in
UNAVAILABLE             full and continue. It cannot: assembly is not reachable from it.

REPLAY DEPENDS ON       the exact bytes sent to an entity exist only in a trace file whose own
SOMETHING OUTSIDE       package comment says it is not part of the design.
THE DESIGN
```

**Seeing is not reading.** The kernel already holds `Observation.Content` constantly and interprets
none of it. What breached the rule was two lines that *read* content to shorten it for a log.

---

## E1 · The interface

**File:** `entity/entity.go`

```go
// BEFORE
type Interpreter interface {
    Step(x graph.Expectation, f frame.Frame) (grammar.Emission, error)
    Monitor(o *graph.Observation)
}

// AFTER
type Interpreter interface {
    // ASSEMBLE — engram state → what this entity reads. Bounded by the ENTITY: how much
    // it can take. The kernel calls it, so the kernel can compare the full form against
    // the delta form and can re-assemble after a host dies.
    Assemble(x graph.Expectation, f frame.Frame) (Rendering, error)

    // PROJECT — rendering → the host's medium. Bounded by the HOST: what form it can
    // realize. MUST FAIL LOUDLY on an unsupported form.
    Project(r Rendering, h ids.HostID) (Delivery, error)

    // STEP — still the only method that PRODUCES anything.
    Step(d Delivery) (Raw, error)

    // MONITOR — the delta form of assemble. Returns nothing, by constraint: a return value
    // would be a third emission channel.
    Monitor(o *graph.Observation)
}

// Opaque. The kernel holds them and cannot look inside. Each adapter defines its own
// unexported implementation.
type Rendering interface{ isRendering() }
type Delivery  interface{ isDelivery()  }

type Raw string   // the entity's response, unparsed.
```

**Marshalling stays entity-side.** *Marshalling is what interpreting is* — a model's assemble builds
a prompt, python's builds stdin and argv, a human's builds a message. What moves is **who calls it**,
and therefore who holds the result.

---

## E2 · The kernel parses

**File:** `kernel/kernel.go`, and `entity/codex`.

`Step` now returns `Raw`. **The kernel calls `grammar.Parse`**, not the adapter.

```text
THIS IS NOT THE KERNEL READING CONTENT.  the grammar is the kernel's OWN vocabulary and it
already type-switches on the result. what it must not do is interpret what is INSIDE a commit.
```

It also supplies something the trace sink needs and could not previously get: the response itself.
Note that `grammar.ErrNoEmission` already carries `Raw` on **failure** — the kernel held the response
when parsing failed and not when it succeeded, and this removes that asymmetry.

The loop becomes:

```text
Assemble → Project → Step → Parse → switch on Select | Commit
```

with the `Select` and `Commit` branches **unchanged**.

---

## E3 · The trace sink

Replaces the three independent `*trace.Run` handles.

```go
// entity — handed to the kernel, never reaches into state.
type Events interface {
    Opened(span ids.SpanID, e ids.EngramID, ent ids.EntityID, depth int)
    Crossed(span ids.SpanID, e ids.EngramID, ent ids.EntityID, r Rendering, raw Raw, err error)
    Emitted(span ids.SpanID, e ids.EngramID, em grammar.Emission)
    Closed(span ids.SpanID, e ids.EngramID, out ids.ObservationID)
    Refused(span ids.SpanID, e ids.EngramID, why error)
}
```

```go
// REMOVE
kernel.Kernel          Run *trace.Run
entity/codex           Run *trace.Run
entity/runtime.Entity  Run *trace.Run

// ADD
kernel.Kernel          Events Events    // NIL-SAFE. tests pass nothing and nothing traces.
```

**`span` on every method is the correlation key the current trace lacks** — concurrent possibilities
on one engram are indistinguishable without it.

**Delete these**, which is what the sink makes possible:

```text
kernel/kernel.go   indent · shortExpr · oneLine        34 lines
kernel/kernel.go   9 Run.Logf calls
```

`shortExpr` reads `Expectation.Expression` and `oneLine` reads `Observation.Content`. **They are the
only two places the kernel reads content**, and they exist only because the kernel wanted to say
something readable in a log and had no other channel. All rendering and truncation moves behind
`Events`.

`trace.Run` itself is fine and stays — it becomes an implementation of `Events`. Its placement was
the problem, not the package.

---

## C1 · `graph.Summary` moves to the adapter

**Files:** `graph/history.go`, `frame/frame.go`, `entity/codex/assemble.go`

```go
// REMOVE from graph
type Summary struct { … }
func Summarize(name string, e *Engram, d content.Deref) Summary
```

`Summary` is **pure presentation** — it exists because something had to render a child set into a
prompt, which is marshalling, which is what `Assemble` is for. A presentation type in the **durable**
package is misplaced, and it currently leaks across `frame.View`, which exists to stop exactly that
kind of reach.

```go
// frame.View — currently
Summary(name string) (graph.Summary, bool)
```

Give `View` something structural instead — the child engram with `Cites` stripped, which is the
requirement `View` actually has — and let `entity/codex/assemble.go` build whatever prose shape it
wants from that, locally.

```text
THE REQUIREMENT ON View, UNCHANGED:  an entity must not be able to reach
Invocation.Return.Cites and traverse into a child's subtree, or "the parent sees what the child
concluded, not what it was told" stops holding.
```

**This is bundled here** because `assemble.go` is being rewritten for `E1` regardless, and doing it
separately would mean writing the intermediate state twice.

---

## Order and stopping

```text
E1   the interface and the two opaque types. compiles broken until E2 lands.
E2   the kernel calls assemble/project/step and parses. loop green again.
C1   Summary out of graph, View made structural, assemble rebuilt.
E3   the sink, then delete the 34 lines and the 9 Logf calls.
```

**E1 and E2 are one commit** if they cannot compile apart. `C1` and `E3` are separate.

```text
STOP IF   the Rendering/Delivery split forces a decision about what a human adapter or the
          scripted adapter should render. scripted may render NOTHING — a Delivery that carries
          the frame is fine, since scripted bodies read the frame directly. If that does not
          work cleanly, record it rather than inventing a shape.

STOP IF   removing Summary from View requires deciding what else View should expose. The
          requirement is only that Cites is stripped. Anything beyond that is a design call.
```
