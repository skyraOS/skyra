# Entity

**Date:** 2026-08-11 · revised 2026-08-12
**Status: CHANGE PROPOSAL.** Code only — signatures, types, and what to delete.

**Why:** [`../../semantics/interface-revision.md`](../../semantics/interface-revision.md)

> **Revised 2026-08-12.** `Step` returns the emission AND an opaque artifact. The `Raw string`
> version proposed in `2026-08-12/03` was tried and stopped — see `2026-08-12/stops.md`.

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
    Step(d Delivery) (grammar.Emission, Response, error)
    Monitor(o *graph.Observation)
}
```

## New types

```go
// entity — all three opaque. no method the kernel can call.
type Rendering interface{ isRendering() }
type Delivery  interface{ isDelivery()  }
type Response  interface{ isResponse()  }   // what came back. NIL where nothing crossed.
```

Each entity defines its own unexported implementations. `codex` renders a prompt and returns its
text; `runtime` renders a struct it reads by field access and returns **nil**; `scripted` renders
nothing and returns nil.

## The entity parses its own output

**`Step` returns the emission, not text.** An earlier proposal had it return `Raw string` with the
kernel calling `grammar.Parse`. That was wrong twice over:

```text
IT BREAKS DELEGATION   entity/runtime commits content carrying content.EngramRef, and the grammar
                       has NO SURFACE FORM for one — grammar/parse.go toValue produces exactly
                       Ref, WorldRef or Str. A text round trip destroys reach SILENTLY, and no
                       test in the tree would have caught it.

IT NAMES A MEDIUM      `Raw string` puts a model's shape in an interface that must be indifferent
                       to one — the same leak as host.Handle carrying Stdout · Stderr · Exit,
                       removed as C4.
```

Parsing is a property of entities that speak text. `codex` parses because text is what crosses for
it. `runtime` builds a `Commit` directly and has nothing to parse.

**The two-forms claim survives structurally, not through the parser.** `isEmission()` is unexported,
so only `grammar` can implement `Emission` — an entity can return `Select` or `Commit` and nothing
else, whatever it does internally.

**`Response` is what the kernel HOLDS**, and holding was always the point — `interface-revision`'s
three defects are the invariant being self-reported, recovery being unavailable, and replay
depending on the trace. None needed the kernel to parse. It is the artifact you need to reproduce or
diff a crossing you cannot simply re-run, so it is meaningful exactly where an entity is external
and non-deterministic, and **absent otherwise**.

```text
WATCH   entity/runtime returns nil under that rule and IS NOT DETERMINISTIC — `create` calls
        codex.Author inline, so a model call happens with no crossing recorded. A nil Response
        would make that invisible AND LOOK CORRECT. Same defect as questions.md §2, and the
        standing direction is that the authoring step should be a CHILD, not an inline call.
```

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
+ Events sink   (below)
```

**How the invariant gets checked.** The kernel holds one `Rendering` per step and `Monitor` returns
nothing, so the second side comes from a second interpreter. `Registry` already stores
`func() Interpreter`, so constructing one costs nothing structural.

```text
I₁  the live interpreter, fed Monitor(o) for every return   →  Assemble(x, f_now)  →  R_delta
I₂  a FRESH interpreter, fed nothing                        →  Assemble(x, f_now)  →  R_full
                                                               R_delta.Equal(R_full)
```

Both sides are objects the kernel constructed from inputs it chose. The entity supplies only the
predicate, which it must — comparing opaque things needs one from whoever owns them. A broken
entity can still return true; what is fixed is that it no longer chooses what is compared or when.
Run it behind a flag or in tests: a second full assemble per step is not free against a model.

## The trace sink

```go
type Events interface {
    Opened(span ids.SpanID, e ids.EngramID, ent ids.EntityID, depth int)
    Crossed(span ids.SpanID, e ids.EngramID, ent ids.EntityID, r Rendering, resp Response, err error)
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

`span` on every method is the correlation key the current trace lacks — concurrent possibilities on
one engram are indistinguishable without it. `trace.Run` needs a new surface, not just a new home:
its methods are `Logf`, `Step` and `Write`, none of which match.

Deleting the 34 lines and 9 `Logf` calls in `kernel/kernel.go` depends on this existing first — they
are the only channel the kernel currently has.

## Blocked on something that does not exist

```text
Project(r Rendering, h ids.HostID)   ← NOTHING SUPPLIES h.

Registry is map[EntityID]func() Interpreter — no host anywhere. Kernel.Host is ONE host, held as
a field. Member.Host is written by the grammar and NEVER READ. Invocation.Host is hardcoded to
k.Host.ID() on every invocation.

Project takes the host as an argument, so it cannot run until routing picks one, and routing does
not exist. See kernel.md, "WHERE does not exist".
```
