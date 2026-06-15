# Codex Runtime Context

## Purpose

This document is for Codex sessions working on `skyra-v.1`.

Use this to load the current runtime posture quickly. Do not re-litigate the ontology unless the user asks. Codex is primarily execution support; Mythos is used for brainstorming.

## Working Mode

- Be concise.
- Do not dismiss interface names as metaphysics. Evaluate whether they preserve a runtime boundary before challenging the user.
- If a concept sounds strange, check whether it preserves a runtime boundary before challenging the name.
- When asked to implement, read the relevant law/spec first, then code.
- When asked to review, lead with concrete mismatches between docs, tests, and code.

## Current Runtime Shape

The runtime is a traversal system.

- `Space` is durable traversable structure.
- `Relation` / wave is the event in motion.
- `Slice` is active state for a slice plane.
- `Children()` are same-plane topology.
- `Content()` may open another plane.
- `Time` / `Relation` mechanism is stateless with respect to slices.
- `Impulse.Source` may describe where an impulse appears to arise, but it is contextual metadata only. The kernel `origin` remains the wave owner and return-perception root.

Core rule:

> The kernel traverses inactive Space and realizes active state into the traversed Space's plane Slice.

Recording rule:

> DNA computes. Kernel records.

## Decision History

This section exists because current state alone is not convincing. The runtime rules were selected by eliminating designs that collapsed important boundaries.

### Why The `laws/` Folder Exists

Codex created `design-docs/laws/` after the project needed to separate invariants from rules.

Problem: not every load-bearing runtime statement is the same kind of thing.

- An **invariant** must remain true across the runtime. Breaking it means the model is wrong.
- A **rule** is required operational behavior that protects or enforces invariants in a specific area.

Putting everything in `invariants/` made rules pretend to be invariants. Scattering rules elsewhere made them look optional. `laws/` is the parent category that can hold both without flattening them.

Current shape:

```text
design-docs/laws/
  invariants/
  rules/
```

This is not metaphysics. It is a taxonomy that prevents different kinds of runtime commitments from being confused.

### Why The Names Are Load-Bearing

Codex agreed that the names are strange because the problem is strange, not because the project is trying to sound strange.

The names exist because common software nouns collapsed distinctions the runtime needs to preserve:

- `node` was too passive for something that can observe, express, contain, activate, and open another plane.
- `agent` was too narrow for world, memory, thread, point, provider, and being all sharing traversal mechanics.
- `memory` as a subsystem recreated the routing problem of deciding when memory should be invoked.
- `kernel` became misleading when it held `ActiveState`, because that made the mechanism look durable/stateful.
- `invariant` became too broad once operational rules showed up.
- `world` needed the observe split because it is both callable environment and traversable Space.

The current names preserve the boundaries:

- `Space` — durable traversable structure.
- `Relation` / wave — event in motion.
- `Slice` — active state for a plane.
- `Children` — same-plane topology.
- `Content` — possible boundary to another plane.
- `laws` — parent category for invariants and rules.

The test for a name is not whether it sounds ordinary. The test is whether replacing it with a familiar word would erase a runtime boundary.

### Why `Realize` Became Void

Earlier shape:

```go
Realize(...) any
```

Problem: returning a result to the caller creates a chosen result channel. The caller can collect the outcome of its own send as a normal return value.

Current rule: `Realize` returns nothing. The only result path is a return perception wave rooted at the wave's origin.

This preserves the Perception Invariant: what a being sends is chosen; what it perceives is not.

### Why `Content()` Does Not Continue the Parent Wave

Earlier shape:

```go
go realize(parentOrigin, innerReality, impulse)
```

Problem: the inner actor's result returns to the parent wave's origin. That makes actors behave like asynchronous subroutines of other actors.

Current rule:

```go
go realize(innerReality, innerReality, impulse)
```

The contained actor owns the spawned wave. The parent wave continues and receives only its own result. If the inner actor responds, that response is a later chosen send, not a return value.

### Why Parent Waves Do Not Wait

Earlier shape: a wait group tracked the current wave plus spawned actor waves.

Problem: waiting makes actor boundaries dependency boundaries. A's wave can be held hostage by B's interior.

Current rule: crossing `Content()` spawns and does not wait.

This keeps independent actors independent while still allowing perturbation.

### Why `ActiveState` Left the Kernel

Earlier shape:

```go
type Relation struct {
    ActiveState map[string]Slice
}
```

Problem: a slice registry inside the kernel makes the kernel stateful/durable. It blurs mechanism with runtime ownership.

Current rule: the kernel asks the traversed Space for `Slice()`.

The kernel remains mechanism. Slice ownership is plane state.

### Why The Kernel Writes Into The Slice

Earlier confusion: `Point.Express` or other DNA could write its own output into the Slice.

Problem: that moves wave aggregation into DNA. Points, cells, and other Spaces would each decide how traversal-produced active state is recorded, which mixes local expression policy with runtime bookkeeping.

Current rule: DNA computes; kernel records. `Observe` and `Express` compute local behavior and output. The kernel writes activation records, traversal trees, result maps, and other wave-produced active state into the traversed Slice.

The kernel manages Slice mutation during traversal, but it does not own slices durably. The traversed plane owns the Slice.

### Why `Slice()` Is On `Space`

Rejected concern: not every point should own a unique Slice.

Resolution: `Space.Slice()` does not mean "this node owns a private Slice." It means "return the Slice for the plane this Space belongs to."

So:

- world points return the world Slice;
- cells inside a being return the being Slice;
- many Spaces can return the same Slice.

This preserves shared plane state without adding an active/durable object split.

### Why `Children()` And `Content()` Are Different

`Children()` are same-plane traversal. They stay in the same wave, stack, and Slice.

`Content()` may open another plane. If it contains a Space, the kernel starts an independent wave over that Space.

This distinction is the code-level boundary between topology and actor nesting.

### Why `world.Observe` Splits On `self`

The world is both an environment and a Space.

If `world.Observe` always relays to the kernel, then kernel traversal of the world re-enters `world.Observe`, which starts another kernel traversal, recursively.

Current rule:

```go
world.Observe(self, impulse) // self != nil: facilitate action
world.Observe(nil, impulse)  // self == nil: traversal intake only
```

This is an operational rule, not an invariant. It protects the invariants from recursive relay.

### Why Short/Medium/Long Memory Was Rejected

The runtime previously considered separate memory layers.

Problem: arbitrary user input forces a router to decide which memory layer to invoke before the runtime understands the input. That special-cases infinity at the front door.

Current direction: memory, messages, places, and binding realities are all traversable Space. Activation decides what participates. There is no separate "when do I invoke long-term memory" oracle.

## Load These First

Read these before touching kernel/world behavior:

- [Perception Invariant](../laws/invariants/perception-invariant.md)
- [Active State Invariant](../laws/invariants/active-state-invariant.md)
- [Plane Slice Invariant](../laws/invariants/plane-slice-invariant.md)
- [Kernel Recording Invariant](../laws/invariants/kernel-recording-invariant.md)
- [World Observe Split](../laws/rules/world-observe-split.md)
- [Stateless Kernel / Slice Ownership Refactor](../kernel/stateless-kernel-slice-refactor.md)
- [A3 World Observe / Slice Result Stub](../kernel/a3-world-observe-and-result-stub.md)

## Laws Summary

### Perception

Every wave returns its result to its own origin by starting a return perception wave rooted at that origin.

`Realize` does not return a result to the caller. A caller-controlled return value is a second result channel and violates the invariant.

`Impulse.Source` is not the same thing as the wave origin. Source records spatial/event context, such as the world point and coordinates where an impulse arose. Origin records ownership of the result path.

### Active State

The kernel does not traverse the Slice. It traverses Space and writes the active subset into the Slice.

Space is disk space. Slice is active space.

### Plane Slice

`Space.Slice()` returns the Slice for the plane the Space belongs to.

Many Spaces can return the same Slice:

- world points return the world Slice;
- cells inside a being return the being Slice;
- ordinary points do not each own unique slices.

`Children()` stay on the same slice plane. `Content()` may cross to another slice plane.

### Kernel Recording

DNA computes. Kernel records.

`Observe` and `Express` may compute local behavior and return local output, but the kernel writes wave-produced active state into the traversed Slice.

For the world path, `Point.Express` returns `PointResult`; the kernel records it into `worldSlice.Maps()[impulse.ResultMap][pointID]`.

World point results include source metadata so a returned perception can say which world point and coordinate expressed the result without letting that point become the return channel.

### World Observe Split

`world.Observe(self, impulse)` facilitates action.

`world.Observe(nil, impulse)` participates in traversal.

This prevents recursive relay when the kernel is traversing the world as Space.

## Current Code State

Implemented:

- `Space` includes `Slice() Slice`.
- `Relation` no longer owns `ActiveState`.
- `Relation` has no `Register`.
- `Realize(origin, traversed, impulse)` is void.
- `realize` binds state with `traversed.Slice()`.
- `Content()` crossing spawns `go r.realize(inner, inner, impulse)`.
- Parent waves do not wait for child actor waves.
- Parent waves do not receive child actor results.
- `Express` returns output; the kernel records output into Slice maps when the impulse names a result map.

Focused tests:

```sh
go test ./src/skyra/relation ./src/skyra/world
```

Known broad failure:

```sh
go test ./...
```

This still fails because stale/archive packages import old paths or old symbols. Do not treat that as a kernel regression unless the user asks to clean stale packages.

## Next Major Work

A3 is next:

- `world.observe`
- environment reference
- minimal world-being action path
- Slice `Result()` / `SetResult()`

Do not invent the full design while stubbing. Keep A3 grounded in the laws above.

## Things Not To Reintroduce

- Kernel-owned `ActiveState`.
- Durable kernel registry.
- Passing slices through `realize`.
- `Realize` returning a result.
- DNA writing wave-produced result maps directly into a Slice.
- Parent wave waiting on child actor waves.
- Parent wave receiving child actor results.
- Per-point unique slices for ordinary world points.

## Useful Short Forms

- The kernel traverses disk space and brings it into active space.
- DNA computes. Kernel records.
- Children share the plane. Content can open another plane.
- Non-nil self invokes the world. Nil self means the world is being traversed.
- One wave, one origin, one traversal, one result.
