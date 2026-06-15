# Stateless Kernel / Slice Ownership Refactor

## Status

**Implementation spec.** This translates the [Active State Invariant](../laws/invariants/active-state-invariant.md) and [Plane Slice Invariant](../laws/invariants/plane-slice-invariant.md) into the concrete kernel changes needed next.

## Core Rule

The kernel traverses inactive Reality and realizes active state into the traversed Reality's plane Slice.

The kernel does not own active state. Every Reality belongs to a slice plane, and `Reality.Slice()` returns the active-state record for that plane. The kernel operates on the Slice returned by the Reality it is traversing.

Slice ownership is plane-dependent:

- the world is a plane and its points return the world's Slice;
- a being is a plane and its cells return the being's Slice;
- a point does not get a unique Slice merely because it is a point; points exist on the world's plane and are realized into the world's Slice.

## Required Changes

### 1. Remove `ActiveState` from `Relation`

The kernel should not own or hold a slice registry.

Current shape:

```go
type Relation struct {
    ActiveState map[string]skyra.Slice
    Amplitude   float64
}
```

Target: remove `ActiveState`. Slice lookup belongs to the traversed actor/layer, not to the kernel.

### 2. Add Plane Slice Access to `Reality`

Every Reality must expose the Slice for the plane it belongs to. This is not per-node slice ownership; it is shared plane state.

Target shape:

```go
type Reality interface {
    ID() string
    Slice() Slice
    NewActivation() ActivationRecord
    Observe(origin Reality, impulse any)
    Express(origin Reality, impulse any) any
    Children() map[string]Reality
    Content() any
}
```

The important ownership rule is not "every point has its own Slice." The rule is: every Reality can return the Slice for its plane. Many Realities may return the same Slice.

### 3. Remove `Register`

If the kernel does not own slice references, it does not register actors.

Remove:

```go
func (r *Relation) Register(beingID string, sl skyra.Slice)
```

Registration wires actors into worlds, and Realities know their plane Slice. The kernel remains mechanism.

### 4. Bind the Slice from `traversed.Slice()`

Current shape:

```go
r.realize(origin, traversed, impulse, r.ActiveState[traversed.ID()])
```

Target shape:

```go
sl := traversed.Slice()
```

The kernel binds the frame from the traversed Reality's plane at wave start.

### 5. Stop Passing Slices Through `realize`

The Slice is not an argument handed down by the caller. It is returned by the traversed Reality and read at the wave boundary.

Target shape:

```go
func (r *Relation) Realize(origin, traversed skyra.Space, impulse any) {
    r.realize(origin, traversed, impulse)
}

func (r *Relation) realize(origin, traversed skyra.Space, impulse any) {
    sl := traversed.Slice()
    if sl == nil {
        return
    }

    result := r.traverse(origin, traversed, impulse, sl)
    r.deliver(origin, result)
}
```

### 6. Fix `Content()` Crossing

Current stale shape:

```go
isl := r.ActiveState[inner.ID()]
go r.realize(origin, inner, impulse, isl)
```

Target shape:

```go
go r.realize(inner, inner, impulse)
```

`Content()` crossing starts an independent wave. The contained Reality owns the spawned wave and supplies its own plane Slice.

### 7. Remove `wg.Wait()`

Parent waves do not wait for child actor waves.

Actor-boundary spawning is asynchronous. The parent wave continues toward its own result. The child wave delivers to its own origin.

### 8. Make `Realize` Return Void

A wave's result path is delivery through `Observe`, not a caller-controlled return value.

Target shape:

```go
func (r *Relation) Realize(origin, traversed skyra.Space, impulse any)
```

`Express` still returns output internally so ascent can form the wave result. `Realize` does not return that result to the caller.

### 9. Update Tests Around Boundary Spawning

The crossing test should enforce the corrected invariant:

- parent receives only the parent wave's result;
- child wave observes its own intake;
- child wave receives its own result delivery;
- parent does not receive the child result;
- parent wave does not wait on child actor completion.

Avoid tests that depend on timing sleeps. Use a synchronization hook in the mock actor if the test needs to prove the child goroutine ran.

## Non-Goal

Do not make every point own a unique Slice.

Points are Realities on the world's plane. They return the world's Slice and are realized into it. A point only returns a different Slice if it becomes part of a different plane; ordinary grid points do not.
