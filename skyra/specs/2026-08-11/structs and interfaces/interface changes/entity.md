# Entity

**Date:** 2026-08-11
**Status: CHANGE PROPOSAL.** Code only — signatures, types, and what to delete.

**Why:** [`../../semantics/interface-revision.md`](../../semantics/interface-revision.md)

---

## The interface

```go
// BEFORE
type Interpreter interface {
    Step(x graph.Expectation, f frame.Frame) (grammar.Emission, error)
    Monitor(o *graph.Observation)
}

// AFTER
type Interpreter interface {
    Assemble(x graph.Expectation, f frame.Frame) (Rendering, error)
    Project(r Rendering, h ids.HostID) (Delivery, error)
    Step(d Delivery) (grammar.Emission, error)
    Monitor(o *graph.Observation)
}
```

## New types

```go
// entity — opaque. no method the kernel can call.
type Rendering interface{ isRendering() }
type Delivery  interface{ isDelivery()  }
```

Each adapter defines its own unexported implementation. `codex` renders a prompt; `runtime` renders
a struct it reads by field access; `scripted` renders nothing.

## Types

```go
// entity/codex.Interpreter
- base      map[ids.ObservationID]bool
- delta     map[ids.ObservationID]bool
- Divergent []string
- Run       *trace.Run

// entity/runtime.Entity
- Run *trace.Run

// kernel.Kernel
+ the assemble/monitor comparison, using what it now holds
+ Events sink   (see below)
```

## The trace sink

Replaces the three independent `*trace.Run` handles.

```go
// kernel — handed one, never reaches into state
type Events interface {
    Opened(span ids.SpanID, e ids.EngramID, ent ids.EntityID, depth int)
    Crossed(span ids.SpanID, e ids.EngramID, ent ids.EntityID, r Rendering, raw string, err error)
    Emitted(span ids.SpanID, e ids.EngramID, em grammar.Emission)
    Closed(span ids.SpanID, e ids.EngramID, out ids.ObservationID)
    Refused(span ids.SpanID, e ids.EngramID, why error)
}
```

Nil-safe, as `trace.Run` already is, so the substrate tests pass nothing and nothing traces. All
rendering and truncation moves behind this interface, which is what lets the 34 lines in
[kernel.md](kernel.md) be deleted.

`span` on every method is the correlation key the current trace lacks — concurrent possibilities on
one engram are indistinguishable without it.
