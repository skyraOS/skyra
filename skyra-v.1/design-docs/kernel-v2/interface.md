# The Contract

**Status:** DESIGN · part of [kernel-v2](README.md)
**Lifts from:** `specs/2026-07-08/space-kernel-interface-reform.md` (un-superseded — see README)

The reform targets **`skyra.Time` and `skyra.Space`** — the kernel that actually runs (`relation.go`).
Not `traverse.go`, which is an isolated Phase-0 sketch that nothing ever called.

## The problem the reform found

Today `Space` splits traversal discovery across two methods:

```go
Observe(origin Space, impulse any)      // must MUTATE HIDDEN STATE to change the surface
Children() map[string]Space             // the kernel later asks what Observe prepared
Express(origin Space, impulse any) any
```

And it blurs two genuinely different things:

```text
DURABLE GRAPH CHILDREN      the edges that EXIST in memory
PER-WAVE TRAVERSAL SURFACE  the edges THIS WAVE lights, and how probable they are
```

For a `MemoryCell`, `Children()` means durable edges. For an ephemeral worker there are no durable
edges — its children are *the roots selected for this wave*. **One method, two meanings.**

## The fix

**`Observe` returns the traversal surface directly. `Children()` is deleted.**

```go
type Space interface {
    ID() string
    Slice() Slice

    // Observe: WHAT SHOULD THE KERNEL LOOK AT NEXT?
    //
    // Returns this cell's onward distribution FOR THIS WAVE: the edges it lights, each with
    // the PROBABILITY that it reaches the wave's target. nil = leaf.
    //
    // The cell owns the policy. The kernel only orders by the number and never asks what it
    // means.
    Observe(origin Space, impulse any) map[string]Alloc

    // Express: WHAT HAPPENED AFTER TRAVERSAL?
    //
    // The post-order fold. It SEES ITS CHILDREN'S RESULTS — which the current signature does
    // not allow, and which is the difference between a fold and a side effect.
    Express(origin Space, impulse any, children []*Result) any

    Agents() map[string]Space
    Content() any
}

// Alloc is one entry in the onward distribution.
type Alloc struct {
    Target Space
    Weight float64 // P. "how likely is this edge to reach the wave's target." [0,1]. NO SIGN.
}
```

```text
Observe = what should the kernel look at next?
Express = what should happen after traversal?
```

No hidden state. No second method to keep in sync. **The wave's surface is a return value**, which is
what lets it be a function of *the wave* rather than of the cell alone.

## Three changes from the 2026-07-08 draft

**1. `Alloc`, not bare `Space`.** The draft returns `map[string]Space` and has the kernel look up
edge strength in the slice's activation records. The cell should hand over its own weights instead —
which **subsumes `proportional-amplitude-allocation.md` entirely.**

**2. No polarity.** The draft carries signed magnitude, because v2's `A` was a judge's opinion on
`[-1, 1]`. **In v3 the weight IS `P`, a probability.** Probabilities have no sign. `0.1` means
*"unlikely,"* not *"negatively aligned."*

**3. No `Cost`.** The draft adds `Cost(origin, impulse) int` for affordability. **There is no budget**
(see [traversal.md](traversal.md)) — so there is nothing to afford. The stop condition is
`env.Done()`, and accounting (tokens consumed, nodes admitted) lives in `env.Admit()`, computed by the
cell as the present is built.

## `Time` — the kernel itself

```go
type Time interface {
    ID() string
    // Realize runs a wave. VOID: a wave's output is written to the traversed Slice and
    // returned to the origin as a PERCEPTION wave. A return value would be a second,
    // chosen results channel, which the Perception Invariant forbids.
    Realize(origin, traversed Space, impulse any)
}
```

**This stays void — and the fold does not conflict with it.** They are different mechanisms:

```text
INTRA-WAVE   Express(origin, impulse, children) any    the ascent. the parent SEES its children.
INTER-WAVE   Realize(origin, traversed, impulse)       VOID. results return as PERCEPTION.
```

The fold flows up *inside* one wave. The wave's *result* comes back to the origin as a
return-perception wave. **The Perception Invariant survives the fold intact** — and restoring it
(suspended at `relation.go:56`, with results currently carried by `go kernel.Realize` side channels)
is **part of this reform, not separate from it.** The side channels are the reason the kernel became
scenery.

## The DAG is not the executor

From `dag-kernel-responsibility-boundary.md`, still exactly right:

```text
TODAY (wrong)                      INTENDED
DAG creates worker                 DAG:    choose the step, mount the ephemeral node,
worker.Observe does the WORK               attach observe/express, hand it to the kernel
orchestrator reads condition       KERNEL: Observe → traverse → Express
DAG routes next worker             WORKER: Observe QUERIES / ACTIVATES.
                                           Express DOES and COMMITS.
```

> **Observe asks: what is available / active here?**
> **Express says: what happened after traversal?**

v2 inverted this — the genes did their substantive work in `Observe` and `Express` was vestigial,
which is *precisely* why the kernel became scenery and the orchestrator became the real executor.

**If `Observe` still does the work, the kernel is still decoration.** Getting this right is the whole
point of the revision.

## Durable children versus the traversal surface

They are different, and separating them is what makes the scheduler possible:

```text
the DURABLE graph   what EXISTS.        memory, topology, adjacency.
Observe()           what THIS WAVE lights, and how probably.
```

A retrieval wave and a present wave standing on the *same* cell light *different* edges with
*different* probabilities. **That is only expressible if the surface is computed per wave** — which is
exactly what returning it from `Observe` buys.

See [scheduler.md](scheduler.md).
