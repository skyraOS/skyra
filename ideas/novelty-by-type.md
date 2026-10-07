# Novelty by type: a parent's known types (draft)

Date: 2026-10-07. Status: thinking, not agreed. Builds on 28 (the parent owns
the check), 30 (novelty is a miss on the type), 32 (types and refinements are
the interpreter's; fallback at an edge).

## The rule

> **At each edge, a parent knows a set of types. A type is a shape, and the
> properties the parent says hold for it.** A return whose shape is not in
> the set, or whose shape is there but whose properties do not hold, wakes
> the parent: a novelty interrupt. What the parent writes is kept, so the
> next return of that type does not wake it.

The order is fixed: **shape first, then properties.** Properties are written
against a shape, so they cannot let through a shape nobody has seen.

```text
child returns v at edge e
  shape(v) ∉ known(e)                    → wake: unseen
  shape(v) = T.shape, T.holds(v) false   → wake: outside
  shape(v) = T.shape, T.verdict = refuse → refuse, without waking
  shape(v) = T.shape, T.holds(v) true    → the compiled body goes on
```

This is a polymorphic inline cache with guards: the shape is the cache key,
the properties are the guard, and a wake is the slow path that fills the
cache.

## What the interpreter already has

| The idea | Built (`controller/interp`) | Missing |
|---|---|---|
| type = shape | `types.Of`: H(name, contract, shape); `proc.EdgeType` | none, for now |
| the parent says what holds | `Refinement.On[node]`: a CEL predicate, written once at remember (H9) | one predicate per edge rather than one per shape; written only at remember |
| unseen shape → wake | `Refinement.Edges[node]`: exactly **one** shape hash; a mismatch → `fallback` | only one shape per edge |
| shape first | **No.** In `refusedEdge`, a CEL predicate on an edge *replaces* the shape check (`continue`) | an unseen shape passes a loose predicate without waking |
| wake | `fallback`: invokes the original actor with `resume {command, ran, refused}` | the model takes over for good (`pos.Fallback`); its answer is final, so it cannot teach the procedure and hand control back |
| kept | `SetRefinement`, only during remember | nothing is kept from a wake at run time |
| fan-out | `refusedEdge` returns the first refusal | every novel edge in a round should go to one wake (30) |

## 1. Known types per edge

```go
type Known struct {
    Shape   string // hash: never shown to a model
    Shown   string // the presentation
    Holds   string // CEL over this shape; "" means the shape is enough
    Verdict string // "admit" | "refuse"
}
// Refinement.Edges: node → []Known (was node → one hash); On folds into Holds.
```

- Migration: today's single `Edges[node]` hash and `On[node]` predicate
  become one `Known{Verdict: admit}`.
- The input is handled the same way: `Input` becomes `[]Known` over the
  header.
- **CEL can be typed per shape.** A `Known` declares its shape to CEL as an
  object type, so `returns.x.txet` fails at compile time instead of being an
  error at run time. This settles 32's open question "Typing JSON in CEL".
- **`each` edges have element types.** An array's shape is the set of its
  distinct element shapes, so one odd element makes the whole list novel.
  The known types of an `each` node should be its elements' types, and only
  the odd element wakes the parent.

## 2. Demand: what the rest of the body reads from the edge

The interpreter can compute, for each node, which paths of its return the
downstream templates read. `proc.Refs` already finds the root of each
binding; the change is to keep the full path:
`{"$": [node, "parts", ...]}` → `$.parts`. That is the edge's **demand**.

A new shape either meets the demand (every path it needs is present, with the
same leaf kind) or it does not. This is width subtyping, measured against
what the body actually uses rather than a declared type. The interpreter
computes it; the model never asserts it.

## 3. The wake

The resume payload gains a `novelty` section, made of names and
presentations only:

```json
{"resume": {"command": …, "ran": […],
  "novelty": [{
    "at": "summarize_chunk[2]",
    "kind": "unseen",
    "type": "summarize_chunk → {error: string, retry_after: number}",
    "known": ["summarize_chunk → {text: string}"],
    "delta": {"added": ["$.error", "$.retry_after"], "removed": ["$.text"]},
    "demand": ["$.text"],
    "meets_demand": false
  }]}}
```

- `kind` is `unseen` (no known shape) or `outside` (a known shape whose
  properties are false, or hit an error).
- `delta` is `types.Diff` against the nearest known shape. It already
  exists, so the model sees "Summary, but with `error` added and `text`
  removed", not just an unfamiliar blob.
- Every novel edge in the round goes into one wake.

## 4. What the parent answers

| Answer | Allowed when | Effect |
|---|---|---|
| `admit {holds}` | `unseen`, and the shape meets the demand | adds a `Known` type; the frame **re-plans over the same steps** and the body continues. Nothing runs twice |
| `widen {holds}` | `outside` | replaces that type's predicate, after the case check (section 5) |
| `refuse {cause}` | either | records `Known{Verdict: refuse}`; this time and every later time the procedure returns refused, without waking. This is 30's "refusing is a route" |
| `take` | either | today's fallback: the model goes on from this edge, and its return is the procedure's. Nothing is kept, because a branch needs versioning |

Today `Position.Fallback bool` means "the model owns this frame now". That
becomes `Waking`: the frame waits on the wake's answer, and only `take` turns
into the old fallback. A shape that does not meet the demand can only be
`take` or `refuse`. Running the compiled body on it would fail at a binding.
Handling it properly is a branch: remember the `take` strand once versioning
exists.

## 5. Checking an answer (the interpreter, deterministically)

1. The CEL compiles against the shape, its cost is bounded, and it accepts
   the value that caused the wake without hitting an error. These are the
   checks `accepts` already makes at remember.
2. `admit` meets the demand.
3. **The log is the test suite (28).** The interpreter keeps the cases for
   each edge in its own state (`edges/<procedure>/<node>`: content hash,
   value, verdict, deduplicated). A `widen` is run against every recorded
   case of that type. A verdict that flips is shown back to the model, which
   acknowledges it on purpose, and the flip is recorded. An `admit` adds a
   new shape, so it can only widen and needs no case check. Changing an
   existing type's predicate needs one.

## 6. Authority and safety

- Only the actor the procedure was remembered from answers a wake, and its
  answer is a recorded return in the procedure's frame. A procedure lives
  only in the Horizon that remembered it, so the model widens a refinement
  under the same Horizon it was written in. F10 holds.
- Widening is bounded the way 32 says a loose refinement is: the body still
  emits through intake.
- Putting the shape first closes the current hole: no predicate can admit an
  unseen shape.
- The wake shows the refused value, which is where injected text sits. An
  injected `admit {holds: "true"}` is bounded by the Horizon and by the
  demand check, but the adversarial suite should aim at it.
- Races: two frames hit the same new shape. An `admit` identical to one
  already stored is a no-op. A different predicate for the same shape
  becomes a `widen`, with the case check. The races are the actor's (29).

## 7. Order

1. Shape first, always: a predicate no longer skips the shape check. Small,
   and it fixes a real hole.
2. `Edges` becomes `[]Known` per node, with the migration.
3. Demand: full paths from templates.
4. The `novelty` section in the wake: kinds, delta, demand, every novel edge
   in the round.
5. `Waking` and the answers `admit`, `refuse`, `take`; re-plan over the same
   steps.
6. The case log and `widen` with flips.
7. Element types for `each`.

## Open questions

- **Keying.** Per procedure edge (node), as now, or per (parent, child
  contract), shared by every node that calls that child. Per edge is
  simpler, and demand is per edge anyway.
- **Identity.** The name the value came back through, as now (`reference` for
  narrowings), or the registration hash 30 wanted, so that re-registering
  `twitter` makes its old shapes novel again.
- **Content novelty.** 30's `previous()` / `changed(path)` / `diff()` ("wake
  me when the text changes") are properties too: `previous` could join the
  CEL environment. That is novelty in content, not type; it may belong to a
  later step.
- **A budget for wakes.** How many wakes per frame or per edge before the
  procedure gives up and `take` becomes the default.
- **Deterministic parents** (the user actor, agent) cannot answer, so their
  known types are declared at registration, as in 30.
