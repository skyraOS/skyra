# Kernel Interface — Execute the Returned Space

**Date:** 2026-07-23  
**Status:** DECIDED. `Perturb` returns one executable ephemeral Space.
`Observe` receives one realized Space and returns one concrete Space or nil. The
kernel performs no cognitive math and never interprets a child collection.

**Reads on:** [perturb-observe-traversal](perturb-observe-traversal.md) for the
representation semantics, [traversal-spine](../2026-07-22/traversal-spine.md) for the field
being walked, and [tokenizer](tokenizer.md#the-concretesegmented-cycle) for the
concrete/segmented distinction.

---

## 1. The interface

```go
type Space interface {
    Perturb(origin Space, impulse any) Space
    Observe(origin Space, impulse any, realized Space) Space
}
```

The exact origin and impulse types may still be refined. These meanings may not:

```text
Perturb    returns one ephemeral executable Space or nil
Observe    receives one realized Space and returns one concrete Space or nil
Space      owns its grouped data and the behavior governing that grouping
```

`Expression` remains the name of a cognitive inference operation and durable
record. `Express` is not a traversal method.

The kernel never receives a map, slice, iterator, or tuple of children.

---

## 2. A Space is data with behavior

The value returned by `Perturb` is not a passive frontier description. It is an
executable continuation:

> **Execute this data with this behavior.**

An ephemeral Space may privately carry:

```text
grouped segmented members
their ordering and boundaries
sequence, fan-out, or join behavior
CreationPhase
amplitude/evidence phase
allocation and affordability state
cut and assembly policy
```

The group is first-class. It is not flattened into smaller nodes and handed to
the kernel for interpretation.

This combines several familiar patterns:

```text
Command / Composite       data, grouping, and behavior travel together
Continuation passing      Perturb returns the next executable unit
Interpreter / trampoline  the kernel repeatedly executes returned Spaces
post-order fold           Observe realizes the result on unwind
event sourcing             the Slice records the ephemeral execution
```

The canonical project term is **executable continuation Space**.

---

## 3. The kernel does zero cognitive math

> **The kernel executes the Space it is given. Nothing else.**

The kernel:

1. calls `Perturb`;
2. records the returned ephemeral Space on the Slice;
3. executes that Space when it is non-nil;
4. calls `Observe` with the realized Space returned by that execution;
5. returns the resulting concrete Space upward.

The kernel does not:

```text
enumerate children
interpret grouping
subtract or divide
compare magnitudes
read Cost
allocate amplitude
test affordability
select or order a frontier
apply CreationPhase visibility
render a prompt
infer meaning
rebuild a concrete result from the Slice
```

A nil Space returned by `Perturb` is the descent base case. `Observe` still runs
at that point with a nil realized Space.

---

## 4. Kernel code shape

```go
func execute(current Space, origin Space, impulse any, slice *Slice) Space {
    next := current.Perturb(origin, impulse)
    slice.RecordPerturbation(current, next)

    var realized Space
    if next != nil {
        realized = execute(next, origin, impulse, slice)
    }

    return current.Observe(origin, impulse, realized)
}
```

This is structural pseudocode. A composite returned Space privately schedules,
sequences, fans out, joins, or cuts its grouped members through its own behavior.
The kernel does not special-case those structures.

A grouped Space may implement its behavior as an internal state machine that
returns further ephemeral continuation Spaces. Each continuation still crosses
the kernel as one Space.

---

## 5. `Perturb` owns descent and grouping

`Perturb` answers:

> Which executable Space should the kernel run next?

It owns every calculation required to answer:

```text
propagation
group formation
allocation
affordability
cost
member ordering
conservation or broadcast policy
phase-sensitive routing
```

Candidate behavior:

```text
Perturb:
    read this Space's amplitude and phase
    inspect durable relationships
    select eligible segmented material
    preserve the meaningful grouping and boundaries
    attach the behavior governing that group
    return one ephemeral executable Space
```

The returned Space may contain many internal members. The kernel sees one value
with one behavior boundary.

---

## 6. `Observe` owns concrete realization

`Observe` answers:

> What concrete Space did this perturbed Space realize?

It receives the one realized Space produced by executing the continuation. It
may:

```text
return its concrete counterpart
compose the realized group's internal results
apply CreationPhase visibility
cut grouped results under a context bound
construct a grounding header with surviving detail
return nil
```

The result flows up the existing stack. The kernel never performs a second
traversal of the Slice and never unfolds the grouped result.

At an inference assembly boundary, the root's concrete return may be written
directly to the payload log as the exact grounding of the call.

---

## 7. Durable relationships are not the traversal interface

A storage implementation may still expose durable graph relationships, but that
API is separate from `Space`:

```text
durable relationships    everything the source could reach
returned Space           what this perturbation grouped and chose to execute
```

Returning a child collection would force the kernel to assign meaning to that
collection. Returning a Space tells the kernel only to execute the contained data
with its contained behavior.

---

## 8. Phase ownership

The kernel carries an opaque impulse. Spaces may inspect the parts of that
impulse their behavior understands.

Two phases remain distinct:

```text
amplitude/evidence phase    observed or projected status in the field
CreationPhase               BeliefPhase or ActionPhase for this worker
```

The kernel reads neither.

During `Observe`, the ephemeral Space applies `CreationPhase` visibility:

```text
same-kind prior material    traversable during Perturb; concretely transparent
different-kind material     eligible for concrete realization
```

See
[creation phase controls concrete visibility](perturb-observe-traversal.md#6a-creation-phase-controls-concrete-visibility).

---

## 9. Amplitude, affordability, and density

Amplitude is space-owned:

```text
amplitude = evidence phase + magnitude
```

It is read, allocated, and written by Spaces, not carried or scaled by the
kernel.

At the assembly boundary:

```text
budget     = model context window
Cost       = concrete record's token count
magnitude  = candidate share of attention
```

Price-per-token selection is the returned assembly Space's own
allocation-with-affordability behavior:

```text
allocation = incoming share attenuated by magnitude
keep iff allocation covers Cost
```

There is no kernel knapsack. The ephemeral grouping Space owns execution order,
and its assembly weights conserve when the context window must bound its total
material.

---

## 10. Ephemeral execution, durable transcript

Runtime Spaces may disappear as soon as a wave completes. The Slice therefore
records an immutable transcript before each continuation disappears:

```text
SpaceID
parent SpaceID
grouped member IDs and boundaries
BehaviorID + version
CreationPhase
amplitude/evidence phase
perturbation ID
```

The two durable artifacts have different authority:

```text
Slice
    authoritative perturbation transcript
    every ephemeral Space actually executed

Observe return / payload log
    derived concrete Space
    what survived and was realized
```

An observed omission does not erase the corresponding member from the Slice or
declare it false.

---

## 11. Migration contract

```text
old Observe(origin, impulse)                 → Perturb(origin, impulse) Space
old Express(origin, impulse) value           → Observe(origin, impulse, realized) Space
old Perturb() child collection               → one executable ephemeral Space
old kernel child enumeration                 → recursive execution of returned Space
old hidden result-tree assembly              → Space-owned grouping and concrete return
```

Mechanical renaming is incomplete unless the return becomes an executable
grouping and the kernel stops interpreting collections.

---

## 12. Interface laws

1. `Perturb` returns exactly one executable Space or nil.
2. The returned Space owns all grouping and group behavior.
3. Nil from `Perturb` is the descent base case.
4. Every non-nil returned Space is recorded before execution.
5. `Observe` receives exactly one realized Space or nil.
6. `Observe` returns exactly one concrete Space or nil.
7. Same-kind visibility is enforced inside phase-aware Spaces.
8. The kernel never enumerates grouped members.
9. The kernel performs no cognitive arithmetic.
10. The Slice stores enough immutable information to audit an ephemeral Space
    after runtime disposal.

---

## 13. Open implementation questions

1. Whether concrete and segmented Spaces share one Go interface or refine it
   through subinterfaces.
2. The minimal immutable transcript required to reconstruct an ephemeral
   grouping.
3. How a grouping Space encodes sequential, parallel, and join behavior.
4. Concurrency and Slice locking if grouped members run concurrently.
5. Whether nil is sufficient or deserves an explicit option/result type.
6. Whether `CreationPhase` lives directly on a typed impulse or behind a narrower
   capability interface.

---

## 14. One-paragraph statement

`Perturb` returns one ephemeral executable Space rather than a collection of
children. That Space contains the selected data, its grouping, phase, and the
behavior governing its execution. The kernel records an immutable description
of the Space, executes it, and passes its one realized return into `Observe`.
`Observe` produces one concrete Space or nil. All grouping, scheduling,
allocation, affordability, phase visibility, selection, and composition belongs
to Spaces; the kernel is only their interpreter.
