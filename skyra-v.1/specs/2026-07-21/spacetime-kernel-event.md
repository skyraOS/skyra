# Spacetime: The Concrete Kernel Event

**Date:** 2026-07-21  
**Status:** ARCHITECTURAL DECISION · pulled forward from the 2026-07-08 kernel
interface reform; not yet implemented.  
**Scope:** `skyra.Time`, `skyra.Space`, and the value handed to the kernel when a
wave begins.  
**Supersedes only:** the positional
`Time.Realize(origin, traversed, impulse)` signature.  
**Does not supersede:** the current kernel-v2 traversal, scheduler, or
`Observe`/`Express` decisions.

> **SUPERSEDED IN PART (2026-07-23) — `Amplitude` is not a `Spacetime` field.**
> [kernel-alteration](../2026-07-22/kernel-alteration.md) §1a makes amplitude
> **space-owned** (it *is* the excitation field; `amplitude = phase + magnitude`),
> read from the space rather than carried by the kernel event. What this spec got
> RIGHT stands — amplitude is *not* a token budget, the kernel does *not* divide it
> across branches, `child budget = parent remainder × edge weight` is WRONG. What is
> superseded is only its *location*: the `Amplitude float64` field on `Spacetime`,
> and the "Field semantics → `Amplitude`" section. The kernel event carries the
> **impulse payload only**. See the marked sections below.

**Parents:**
[Space / Kernel Interface Reform](../2026-07-08/space-kernel-interface-reform.md) ·
[Directed Affordance Invocation](../2026-07-08/directed-affordance-invocation.md) ·
[kernel-v2 contract](../../design-docs/kernel-v2/interface.md) ·
[kernel-v2 traversal](../../design-docs/kernel-v2/traversal.md)

## Decision

The concrete event of the time kernel realizing an impulse through a space is a
`Spacetime` value.

```go
type Spacetime struct {
	Origin    Space
	Traversed Space
	Impulse   any
	Amplitude float64 // SUPERSEDED 2026-07-23 — amplitude is space-owned (excitation field),
	                   // not a kernel wave-start field. see kernel-alteration §1a.
}

type Time interface {
	ID() string
	Realize(st Spacetime)
}
```

The fields already exist in the live call, but only as positional arguments and
an `Amplitude` field carried on `Impulse`. The reform gives the event a concrete
identity and places kernel wave-start state on that event.

```text
Impulse     what is moving; domain/event content
Space       what can be traversed
Time        the kernel that realizes one ordered wave
Spacetime   this impulse beginning this traversal from this owner
```

`Spacetime` is a value, not a new agent, graph node, or durable ontology object.
It is the complete kernel input for one wave.

## Why the bundle matters

The current signature permits the three parts of one event to travel separately:

```go
Realize(origin, traversed Space, impulse any)
```

That hides an invariant: none of those values describes a kernel event by itself.
An impulse does not occur in the kernel without a traversed space, and a traversal
does not occur without an owner to whom perception belongs.

The bundle makes illegal partial calls harder to express and gives schedulers,
membranes, traces, and distributed crossings one value to hand off without
reconstructing positional arguments.

## Field semantics

### `Origin`

`Origin` owns the wave. It is where the result may return as involuntary
perception. It is not the sender described by `Impulse.Source`, and it is not a
visited marker.

### `Traversed`

`Traversed` is the `Space` whose slice and per-wave surface the kernel walks.
DNA and World may both satisfy `Space`; `Time` does not need to know which one it
was handed.

### `Impulse`

`Impulse` is kernel-opaque domain content. It may describe perception, action,
cognition routing, or another event type. The time kernel may transport it but
must not interpret its cognitive meaning.

### `Amplitude` — SUPERSEDED 2026-07-23

> **`Amplitude` is not a `Spacetime` field.** [kernel-alteration](../2026-07-22/kernel-alteration.md)
> §1a makes amplitude **space-owned** — it is the excitation field, `amplitude = phase +
> magnitude`, held in the spaces and read by `Observe`. The kernel event carries the impulse
> payload only. The question this section left open (*"may `Amplitude` be deleted from
> `Spacetime`?"*) is answered: yes — it does not live on the kernel event. It moves to the
> **space**, not onto `Impulse` (that prohibition still holds).

What this section got RIGHT, and which the correction preserves:

```text
WRONG   impulse.Amplitude = traversal execution state mixed into domain content   (still WRONG)
WRONG   child budget = parent remainder x edge weight                             (still WRONG — kernel does no math)
RIGHT   amplitude is NOT a token budget, NOT divided by the kernel                (still RIGHT)
RIGHT   weights order traversal; the environment decides when it is full          (still RIGHT)
```

The single change: **amplitude's home is the space, not `Spacetime`.** Original text of this
section retained below for the record.

`Amplitude` is the initial strength assigned to this wave. If activation
attenuates during traversal, that attenuation begins from this value.

Amplitude is not part of the impulse because the same impulse can be realized
with different wave strengths. It is also not a token budget and must not be
divided among branches as an affordability allocation.

```text
WRONG   impulse.Amplitude = traversal execution state mixed into domain content
WRONG   child budget = parent remainder x edge weight
RIGHT   spacetime.Amplitude = initial wave strength
RIGHT   weights order traversal; the global environment decides when it is full
```

If the completed kernel-v2 traversal proves that no independent wave-strength
value is required, `Amplitude` may be deleted from `Spacetime`. It must not move
back onto `Impulse`.

## The 100k context ceiling is separate

`Spacetime.Amplitude` does not enforce the inference-window limit.

The context limit is one global invariant over the materialized inference
request:

```text
system
+ mandatory current state
+ admitted active memory
+ rendering overhead
+ reserved output
+ safety margin
<= 100,000 tokens
```

During traversal, the environment owns this through the current kernel-v2
contract:

```go
Done() bool          // the present has reached its allowed envelope
Admit(s Space) bool  // the node fits; account for its token weight
```

Weights order which nodes are offered first. `Admit` accounts for their token
weights. `Done` stops the walk. No branch receives a private portion of the
context window.

Immediately before inference, the complete serialized request must still be
counted and rejected or rebuilt if it exceeds the hard limit. Traversal
accounting is selection; final request accounting is enforcement.

## `Time.Realize` remains void

Bundling the inputs does not create a chosen return channel:

```go
Realize(st Spacetime)
```

still returns nothing.

The intra-wave post-order fold and the inter-wave result path remain distinct:

```text
INTRA-WAVE
    Express sees child results during ascent

INTER-WAVE
    a completed result returns to Origin as a new perception Spacetime
```

A caller may not synchronously collect the outcome of its own send. Restoring the
Perception Invariant therefore means constructing a new return-perception event,
not adding a return value to `Time.Realize`.

```go
time.Realize(Spacetime{
	Origin:    origin,
	Traversed: origin,
	Impulse:   returnPerception,
	Amplitude: completed.Amplitude,
})
```

The exact return scheduling remains part of kernel-v2. The invariant here is the
shape of the crossing.

## Scheduler boundary

The scheduler may select and emit the next `Spacetime` value. `Time` executes it.

```text
SCHEDULER
    decides which wave should exist next
    emits Spacetime

TIME
    traverses the supplied Space
    folds the lit subgraph
    knows no cognition
```

Wave purpose, target semantics, closure, grading, and world meaning must not
become fields that `Time` interprets. They may ride inside a kernel-opaque
envelope used by the relevant spaces and scheduler.

## Agent target versus wave direction

For goal regression, the agent's direction and the wave's direction remain
inverse:

```text
agent target       desired state
wave origin        desired state
wave target        current state
```

`Spacetime.Origin` still means wave ownership, not necessarily the first graph
node used by a domain-specific search. If kernel-v2 needs both concepts, it must
name them separately rather than silently overloading `Origin`.

This is an explicit warning because the existing API uses `Origin` as the wave
owner for return perception, while the regression documents sometimes use
"wave origin" to mean the desired-state search root. Those are not automatically
the same `Space`.

## Migration

The mechanical implementation order is:

1. Add `Spacetime` beside `Time` and `Space` in `src/skyra/spacetime/space.go`.
2. Change `Time.Realize` to `Realize(Spacetime)`.
3. Change `relation.Relation.Realize` and its internal entry point to accept the
   value.
4. ~~Move initial amplitude from `Impulse` onto `Spacetime`.~~ **[SUPERSEDED 2026-07-23]**
   Amplitude is space-owned (excitation field), not carried on `Spacetime` — remove it from
   `Impulse` and do not add it to the kernel event. See kernel-alteration §1a.
5. Update every wave spawn, fake kernel, harness, and test.
6. Restore return perception by spawning a new `Spacetime`, when the broader
   kernel-v2 return-path work lands.
7. Land the `Observe`-returns-surface and `Express`-sees-children reform as a
   separate change; `Spacetime` does not require those signatures to change in
   the same commit.

Keeping the first migration mechanical matters. It establishes the event type
without simultaneously changing traversal behavior.

## Tests implied by the decision

1. A wave invoked through `Spacetime` produces the same traversal as the current
   positional call.
2. `Origin` remains the wave owner and `Impulse.Source` remains contextual sender
   metadata; swapping them fails the test.
3. DNA and World can both be supplied as `Traversed` without a kernel branch on
   their concrete type.
4. Two `Spacetime` values may carry the same impulse with different initial
   amplitudes without mutating the impulse.
5. `Time.Realize` has no return value.
6. A completed result can reach the owner only through a separately recorded
   return-perception wave.
7. The kernel never type-asserts cognitive payloads in order to choose traversal
   semantics.
8. The hard inference-window test counts the complete serialized request and
   cannot be bypassed by setting a larger amplitude.

## Bottom line

`Time` is the kernel. `Space` is what it traverses. `Spacetime` is one concrete
event joining them.

The change is small in code but important in meaning: ownership, location,
content, and initial wave state stop traveling as loosely associated arguments.
The kernel receives one complete event, realizes it without cognitive knowledge,
and returns nothing except through the system's perception path.
