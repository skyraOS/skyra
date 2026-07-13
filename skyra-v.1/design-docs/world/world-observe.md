# World Observe & On-Demand Fabric

## Status

**Spec — precedes implementation.** This is the design for the World as a `Space` and for `World.Observe` / `Point.Observe`, the methods we are about to write. It synthesizes the points work from this session. The [conducting fork](#the-fork-resolved-conducting-space) is resolved — space conducts — so what remains is implementation and tuning.

## Scope

What this covers:

- the World as a concrete `Space` that owns the grid
- `Observe` as the method that **builds space on demand** rather than reading a pre-built grid
- lazy, sparse point construction — no eager million-point allocation
- what is durable vs. ephemeral in the world
- the resolved conducting-space decision that defines how far a single observation can propagate

What this does **not** re-derive (already canon — referenced, not repeated):

- the `self`-vs-`nil` branching of `Observe` → [world-observe-split.md](../laws/rules/world-observe-split.md)
- which Slice a point's frame belongs to → [plane-slice-invariant.md](../laws/invariants/plane-slice-invariant.md)
- disk → active realization → [active-state-invariant.md](../laws/invariants/active-state-invariant.md)
- observer-relative, stale-until-re-observed state → [local-relativity-invariant.md](../laws/invariants/local-relativity-invariant.md)

## The Point Is the World's DNA

The world is an organism, and it has exactly one cell type: the **Point**. Every point runs the same methods — `Observe` builds fabric, `Express` executes the primitive, `NewActivation` makes it conduct, `Children` resolves neighbors. That shared code is the world's **DNA**. A materialized point is a **cell** expressing it.

This is the being/world parallel, exact (see [dna.md](../../archive/design-docs/dna.md)):

| | DNA (shared, durable code) | Differentiation (placement) |
|---|---|---|
| **Being** | cell-type methods — Observe / Express / competence | where the cell sits in the brain topology |
| **World** | the Point's methods | the coordinate `"x,y"` |

It matches the standing seed — *DNA is the interface/methods, not the `.skyra` file; the file is placement.* For the world, the coordinate is the placement; the Point's code is the DNA.

Two consequences shape how this gets built:

- **The world is homogeneous tissue.** A being is heterogeneous — membrane, binding realities, memories, several types. The world is one type repeated, differentiated only by position. That uniformity is why space is lawful where a mind is not.
- **The Point's methods are the physics.** The same DNA at every point means one law everywhere, over a field whose state (weight, occupancy) varies by coordinate — which is what a physical law *is*. So `Observe` / `Express` / `NewActivation` are conduction, attenuation-as-distance, propagation, and primitive execution. **Writing them is writing the laws of this world, not assembling a grid.**

And it settles the ephemeral question below: **points are ephemeral.** A cell is transient by definition; the DNA is what persists. A point is conjured for one wave from coordinate + occupancy + activation, then discarded — its durable state already lives in the world Slice (`Records`) and the occupancy map, re-keyed by the coordinate that re-conjures it. Caching points would be caching cells; the architecture caches DNA and lets cells come and go.

## The Two Roles of `World.Observe`

The World is both an environment a being acts *through* and a `Space` the kernel *traverses*. The [World Observe Split](../laws/rules/world-observe-split.md) keeps them apart by origin presence:

- `world.Observe(self, impulse)` — non-nil → a being is acting through the world → **facilitate**: `kernel.Realize(self, pointOf(self), impulse)`.
- `world.Observe(nil, impulse)` — nil → the world is being traversed as space → **participate**: intake, never relay.

This spec is about the second role and its consequence for points. The facilitation role is fully specified by the split rule.

## Owner vs. Traversed Root

`Realize(origin, traversed, impulse)` carries two independent axes:

- **origin** — the wave owner; the actor that perceives the result.
- **traversed** — the root Space the wave walks.

Thought and action differ by whether those axes coincide:

```go
// thought: root the wave in the actor's interior
kernel.Realize(self, self, impulse)

// action: root the wave in the actor's world location
world.Observe(self, impulse)
kernel.Realize(self, world.pointOf(self), impulse)
```

Thought roots in the actor. Action roots in the actor's location. Delivery always returns to the actor.

The world is the translator from owner to location. It owns authoritative occupancy, so `pointOf(self)` is a world lookup: "where does this actor stand?" The returned point belongs to the world plane and returns the world Slice, so an outward action builds active state into world space while still delivering the result to `self`.

## On-Demand Fabric: Observe Builds, Children Reads

The central mechanism. The kernel reads a cell's neighbors **after** it observes the cell — always, every cell:

```go
traversed.Observe(nil, impulse)      // build
r.scoreChildren(rootNode, sl)        // reads .Children()
...
childNode.Cell.Observe(nil, impulse) // build
r.scoreChildren(childNode, sl)       // reads .Children()
```

`scoreChildren` is the only reader of `Children()`, and it always runs after `Observe()`. So `Observe` is the build step and `Children` is the read step. A point does not pre-exist with its neighbors wired. It is **realized when a wave arrives and asks**: during `Observe(nil, …)` the point computes its neighbor coordinates, materializes them against the World's store, and populates the view that `Children()` then hands to the kernel.

This is `Realize` as measurement at the grid scale — *Time realizes Space*. Before a wave reaches a coordinate it is potential (a string that could be a point); `Observe` collapses it into actual fabric. The no-op `Point.Observe` becomes the most load-bearing world method, and the one that names the physics.

The wave materializes fabric exactly as far as it propagates. Amplitude attenuates per depth (`relation.go`); when score drops below threshold the wave stops, and the grid simply does not exist past where any wave has been. The observable horizon is a consequence of physics, not a bookkeeping pass.

## Lazy Construction

`CreateGrid()` today eagerly allocates 1,000,000 `*Point` structs and wires 8 neighbors into a map on every one — hundreds of MB, nearly all dormant, built before a single wave runs. That contradicts the [Active State Invariant](../laws/invariants/active-state-invariant.md) ("at boot nothing is lit") and is a v.05 placeholder.

The model:

- **One sparse store on the World** — `map[string]*Point` (or `sync.Map`) keyed `"x,y"`. Only touched points exist. This is the single hashmap, not a million 8-entry maps.
- **Neighbors are computed, not stored** — a point's 8 Moore neighbors are coordinates derived on demand and resolved against the store (`getOrCreate` each). `GridSize` survives only as the **wrap modulus** (`(x+dx+GridSize)%GridSize`), not as an allocation bound. Toroidal space: homogeneous, no edges, every point has exactly 8.
- **Concurrency** — lazy makes the store shared-mutable, so materialization uses `LoadOrStore` (or a guarded map). Contention is low (only first-touch of a new coordinate) but real; eager construction avoided it by being immutable after boot. This is the cost lazy pays.

## Durable vs. Ephemeral

The session's key un-mixing: **the durable world state is not the points.** It is two things, and the `Point` struct is neither.

| | What | Where | Lifetime |
|---|---|---|---|
| **Occupancy** | which coords hold a being/content | placement map on the World, `coord → Space` | durable, tiny (a few beings) |
| **Activation** | which coords have been observed, with weight | the **world Slice's `Records`**, keyed `"x,y"` | durable, grows with exploration |
| **The Point** | a coordinate's realized view for one wave | conjured in `Observe` from coord + occupancy + activation | **ephemeral** |

The `Records` keyset *is* the active (observed) subset; its complement is potential. `recordFor` already creates a record on first sight, so the world Slice fills itself in as waves explore — it becomes the durable trace of everywhere a wave has been. The point ID stays `"x,y"`, so an ephemeral point re-keys the same durable record next traversal.

**Resolved — points are ephemeral.** A cell is transient; the DNA (the Point's methods) is what persists (see [The Point Is the World's DNA](#the-point-is-the-worlds-dna)). A point is conjured per wave and discarded; its durable state already lives in the world Slice and the occupancy map, re-keyed by `"x,y"`. The world Slice lock serializes world-waves, so re-conjuring cannot race.

## The World Slice

Per the [Plane Slice Invariant](../laws/invariants/plane-slice-invariant.md): every point returns the **same world Slice** — points are separate Spaces in one plane, not a million actors. `point.Slice()` returns the shared world Slice; the kernel binds it once at the wave's root and writes every traversed point's record there. The `worldSlice` shared in `CreateGrid` was never vestigial — it is the world's frame, and `main.go` already roots a wave at a point that reads it.

## Proposed Shapes

Interface-faithful — `World` and `Point` implement the existing `skyra.Space` (`ID, Slice, NewActivation, Observe, Express, Children, Content`) and nothing else. Struct internals below are not new interface methods.

```go
type World struct {
    kernel    skyra.Time              // for facilitation: Realize(self, pointOf(self), impulse)
    slice     skyra.Slice             // the shared world frame
    points    sync.Map                // "x,y" -> *Point   (the sparse fabric)
    placement map[string]skyra.Space  // "x,y" -> occupant (durable occupancy)
    size      int                     // wrap modulus
}

func (w *World) Observe(origin skyra.Space, impulse any) {
    if origin != nil {            // a being acts through me — facilitate
        w.kernel.Realize(origin, w.pointOf(origin), impulse)
        return
    }
    // nil: I am being traversed as space — intake only, never relay
}

func (w *World) Children() map[string]skyra.Space { /* entry points: occupants / propagation origin */ }
```

```go
type Point struct {
    id        string                  // "x,y" — stable across traversals
    x, y      int
    world     *World                  // back-ref: resolves neighbors against the store
    neighbors map[string]skyra.Space  // the view built during Observe (transient)
    content   skyra.Space             // occupant, sourced from placement when conjured
}

func (p *Point) Slice() skyra.Slice { return p.world.slice }   // shared plane

func (p *Point) Observe(origin skyra.Space, impulse any) {
    if origin == nil {            // being traversed — build the fabric
        p.neighbors = p.world.neighborsOf(p.x, p.y) // compute 8 coords, getOrCreate
        return
    }
    // non-nil: delivery/perceive. Points don't originate waves yet — latent (no-op).
}

func (p *Point) Children() map[string]skyra.Space { return p.neighbors } // read what Observe built
func (p *Point) Content() any                     { return p.content }

func (p *Point) NewActivation() skyra.ActivationRecord {
    return newSpaceActivation() // baseline conductance; fresh points start at minimum non-zero weight
}
```

`NewActivation()` returns a non-nil baseline record — the resolved [conducting-space](#the-fork-resolved-conducting-space) decision. It must be non-zero for fresh points, or the wave could never enter unseen space.

## The Fork, Resolved: Conducting Space

**Decision: space conducts.** New activations flow as far as the impulse strength lets them. An empty point is added to the tree like any other point, each hop attenuates the signal, and each branch divides the remaining signal across candidate children. Same operation everywhere — there is no empty-vs-occupied branch. An empty point is just space with no occupant.

Mechanically:

- `Point.NewActivation()` returns a non-nil baseline record, so every point — occupied or not — participates in `scoreChildren`.
- The limiter is amplitude, already in the kernel: `score = activation × childAmp`. `childAmp` attenuates by `amplitudeDecay` per hop and is divided across candidate branches. When score falls below threshold the wave stops. A strong impulse reaches far; a weak one stays local; a wide wave dissipates faster than a narrow one. **Reach belongs to the impulse and aperture, not the primitive.**
- `neighborsOf` materializes all 8 Moore neighbors, empty included — they are real points the wave can cross.
- Termination still holds from the existing rails: amplitude attenuation cuts off descent, visited-once prevents revisiting a point on the current path. Infinite potential space, finite materialized neighborhood.

This is "distance is activation" made literal: distance is how many hops a wave's strength survives, relative to the observer (see [local relativity](../laws/invariants/local-relativity-invariant.md)). Branching does not create energy: a wider survivor set splits the attenuated signal, so spread has a cost. And it leaves room for the topology to learn — points warm with traversal (the `BaseActivation` EMAs), so frequently-traveled directions conduct better over time and preferred paths carve themselves through space. Geometry emerges from use, not from coordinates.

**One constraint this imposes.** A never-observed point's activation record must return a *non-zero* baseline on first contact, or the wave could never enter unseen space — a zero score fails threshold immediately, and conducting space silently collapses back to being-to-being. Fresh points start at **minimum weight, not zero** — the "reach for the unknown seeds a cell at minimum weight" rule from the locked decisions. The exact baseline relative to `amplitudeDecay` and threshold is a tuning parameter, not a fork.

## Reached vs. Rooted

A point materialized as another wave's neighbor has been **reached** (a leaf in that wave's tree) but has not **rooted** a wave of its own — it has built no view from itself, holds no perspective. Occupancy and reach do not imply a point has observed. See the [Local Relativity Invariant](../laws/invariants/local-relativity-invariant.md#reached-vs-rooted).

## Part 2: Express Executes

`Observe(nil, …)` builds the world fabric on descent. `Express(origin, …)` executes the payload on ascent.

That split keeps the world boring:

```text
Point.Observe(nil, impulse)
  build local space
  compute neighbors
  expose Children()

Point.Express(origin, impulse)
  parse the impulse
  execute the primitive at this point
  return the primitive result to the kernel
```

The kernel does not parse primitives. It carries the impulse, owns traversal, calls `Express` during ascent, and records the returned output into the traversed Slice. The world point owns world execution because only the point/world knows occupancy, adjacency, placement, and primitive meaning at that layer.

The agent owns impulse construction. Raw API input, user text, or provider output is not parsed by the world into intent. The being's DNA interprets its input and emits a concrete motor signal:

```go
skyra.Impulse{
    Primitive: "engage",
    Payload:   ...,
    Source:    nil, // filled by the world boundary when the actor acts from a point
    Amplitude: ...,
}
```

The world receives an already-formed `Impulse`. It does not decide what the agent meant. It only decides whether that motor signal can execute under world rules and what consequence follows.

The world boundary may attach source metadata to the impulse before handing it to the kernel:

```go
type Impulse struct {
    Primitive string
    ResultMap string
    Payload   any
    Source    *ImpulseSource
    Amplitude float64
    Return    bool
}

type ImpulseSource struct {
    SpaceID string
    PointID string
    Coord   Coord
}

type Coord struct {
    X int
    Y int
}
```

`Source` is spatial/event context, not ownership. It says where the impulse appears to arise, such as `SpaceID: "world"`, `PointID: "3,4"`, `Coord: {X: 3, Y: 4}`. The kernel call still carries `origin` separately, and that `origin` remains the owner of the return perception path.

Invalid impulses are therefore not corrected by the environment. If a being emits a malformed primitive, bad payload, impossible target, or invalid amplitude, the world returns a failed consequence. That failure is delivered back to the origin through the normal perception path. The agent absorbs the error as its own bad motor output and can learn from it.

Biological shape:

```text
external stimulus -> organism interprets -> motor impulse -> environment executes -> consequence returns
```

Not:

```text
external stimulus -> environment decides what the organism meant
```

This is where primitives load:

- `look` — read local / nearby world state from the point's frame.
- `engage` — commit through the current world path and interact with the target if the wave lands.
- `move` — update world occupancy if the primitive is allowed and the destination is valid.
- `speak` / later communication primitives — emit through the world boundary or current thread.

Exact primitive parsing is deferred. The invariant-level shape is not:

```go
Express(...) returns PrimitiveRequest
kernel dispatches it
```

The shape is:

```go
Point.Express(origin, impulse) executes the world primitive
Point.Express returns the primitive result
kernel records that result into the traversed Slice
kernel later returns the wave result through origin
```

So `Express` is execution at the layer being traversed. The kernel remains transport.

## Part 2: Slice Result Field

Primitive execution needs a place to accumulate the wave's local result without making `Realize` return a value.

Add result state to `Slice`:

```go
type Slice interface {
    ...
    Result() any
    SetResult(any)
}
```

The result belongs to the traversed Slice because that is the active record for the plane being walked. For an outward action rooted at `pointOf(self)`, the traversed Slice is the world Slice, so world primitive execution returns world results and the kernel records them there ([Kernel Recording Invariant](../laws/invariants/kernel-recording-invariant.md)). The wave then returns its result to the owner by rooting a perception wave at the origin ([Perception Invariant](../laws/invariants/perception-invariant.md)):

```go
// kernel records each point's Express output into the world Slice:
worldSlice.Maps()[impulse.ResultMap][pointID] = pointResult
// then the wave returns by rooting a perception wave at the origin:
Realize(origin, origin, returnImpulse)
```

This preserves the Perception Invariant:

- `Realize` returns nothing.
- The caller does not synchronously receive a result.
- Delivery is a return perception wave rooted at the origin: `Realize(origin, origin, resultImpulse)` — not a callback.

This also keeps ownership clean:

- The **world** owns primitive execution (the point's `Express`).
- The **world Slice** owns the active record of that execution (`Tree()` + result maps).
- The **kernel** owns recording and the return perception wave.
- The **origin** owns perception of the returned result (its own inward wave).

## Tree Ownership And Delivered Results

For an outward action:

```go
kernel.Realize(self, pointOf(self), impulse)
```

ownership splits like this:

```text
origin    = self          // who perceives the result
traversed = world point   // what Space is walked
slice     = world Slice   // where active state is recorded
```

So the traversal tree belongs to the **world Slice**, not to the actor:

```go
worldSlice.Tree().Current
```

The actor does not own the tree just because it owns the wave. It owns the delivered perception of the result. The world owns the active record of the world traversal.

Result state follows the same rule, and the kernel — not the point — records it. The world Slice holds result maps addressed by the impulse's `ResultMap`, then by cell ID:

```go
worldSlice.Maps()[impulse.ResultMap][pointID]   // e.g. Maps()["perception"]["3,4"]
```

For the first pass, a result map is keyed by point ID:

```go
type PerceptionResult map[string]PointResult
```

Each point contributes its local expression, and the kernel records it:

```text
Point.Express → PointResult{PointID: "3,4", Source: ..., Content: ..., ...}
kernel records → worldSlice.Maps()["perception"]["3,4"] = that PointResult
```

`PointResult.Source` records the point and coordinate that expressed the result. It allows Skyra's returned perception to include "this came from there" without making that point the perception origin.

That keeps the two records separate, both owned by the traversed Slice and both written by the kernel:

- `Tree()` — the structure the kernel traversed.
- `Maps()[ResultMap]` — the facts expressed from the traversed points, addressed by the impulse's `ResultMap`.

The result returns to the actor as the payload of a return perception wave — never an `origin.Observe` callback:

```go
Realize(origin, origin, skyra.Impulse{
    Primitive: "perceive",
    Payload:   worldSlice.Maps()[impulse.ResultMap], // a skyra.PerceptionResult
    Return:    true,
})
```

The invariant is the same either way: the caller does not receive a return value; the result is perceived through `Observe`.

Open implementation detail: whether `SetResult` overwrites a single latest result, appends to a per-wave result list, or stores a structured accumulator. The first pass should use the smallest shape that lets one traversal produce one deliverable result. Current leaning: latest traversal result as `PerceptionResult`.

## Current Concurrency Rule

For the first implementation, a single world point does not originate multiple concurrent outward waves.

Short form:

> One point, one outgoing world wave at a time.

This is a current runtime rule, not a permanent invariant. It keeps the first result model boring:

- one point-originated action is recorded as one traversal tree on the world Slice;
- one point-originated action is recorded as one result map on the world Slice (the kernel records the returned output; the point only expresses it — see [Kernel Recording Invariant](../laws/invariants/kernel-recording-invariant.md));
- no two actions from the same physical location race to define that point's current world result.

This does not make the whole world single-threaded. Different points / actors may still originate independent waves, and actor-interior waves spawned through `Content()` remain independent. Return perception to the actor is a separate concern: the wave returns by rooting a perception wave at the origin (`Realize(origin, origin, …)` — the [Perception Invariant](../laws/invariants/perception-invariant.md)), serialized at the origin's Slice, and may later grow mailbox / scheduled-delivery semantics.

Deferred decision:

> Whether one point may originate multiple concurrent waves is left open.

If v.1 later needs multi-wave emission from one location, the result model must become explicit about queueing, per-wave IDs, or another accumulator shape. Do not solve that in the first pass.

## Open

- **Fresh-point baseline** — the exact minimum-weight value a new point's record returns, relative to `amplitudeDecay` and threshold. Sets how far a unit impulse reaches into unseen space. A tuning number, best found by running it (captain's-log style), not by argument.
- **Occupancy home** — placement map on the World (assumed here) vs. content set on the materialized point from another source.
- **`World.Children()` contents** — what the world returns when traversed as a whole: all occupants, nearby occupants, or the single entry point a signal arrived at.
- **Slice result shape** — latest result, result list, or structured accumulator.
- **Multiple waves from one point** — deferred; current implementation behaves as one outgoing world wave per point at a time.
- **The world originating its own waves** — would make `world.Observe(world, …)` (delivery) and `world.Observe(self, …)` (facilitate) indistinguishable by presence; the split rule's [latent edge](../laws/rules/world-observe-split.md#latent-edge-not-yet-live). Not live yet.

## Relationships

- [world-observe-split.md](../laws/rules/world-observe-split.md) — the `self`/`nil` branching this spec builds on.
- [active-state-invariant.md](../laws/invariants/active-state-invariant.md) — active state is built by traversal, not pre-allocated.
- [plane-slice-invariant.md](../laws/invariants/plane-slice-invariant.md) — points share the world Slice.
- [local-relativity-invariant.md](../laws/invariants/local-relativity-invariant.md) — stale, observer-relative state; reached vs. rooted.
- [world-kernel-coupling.md](../kernel/world-kernel-coupling.md) — the World struct, primitives, and dispatch context.
- [slice-and-world.md](slice-and-world.md) — Space Is Local; how a being acts through the world.
