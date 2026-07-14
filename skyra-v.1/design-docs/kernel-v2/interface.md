# The Space Contract

**Status:** DESIGN · part of [kernel-v2](README.md)
**Lifts from:** `specs/2026-07-08/space-kernel-interface-reform.md` (un-superseded — see README)

## The problem the reform found

Today's `Space` splits traversal discovery across two methods:

```go
Observe(origin Space, impulse any)      // mutates hidden state
Children() map[string]Space             // the kernel asks what Observe prepared
Express(origin Space, impulse any) any
```

> *"`Observe` must mutate hidden state if it wants to change what the kernel traverses, and the
> kernel later asks `Children()` for the traversal surface."*

And it blurs two genuinely different things:

```text
DURABLE GRAPH CHILDREN     the edges that exist in memory
PER-WAVE TRAVERSAL SURFACE the edges THIS WAVE lights
```

For a `MemoryCell`, `Children()` means durable edges. For an ephemeral worker there are no durable
edges — its children are *the roots selected for this wave*. One method, two meanings.

## The fix

**`Observe` returns the traversal surface directly.**

```go
type Space interface {
    ID() string
    Slice() Slice
    NewActivation() ActivationRecord

    // Cost is this node's weight against the budget, IN THE BUDGET'S UNITS. The kernel only
    // reads it; it has no idea what the units are.
    Cost(origin Space, impulse any) int

    // Observe: WHAT SHOULD THE KERNEL LOOK AT NEXT?
    //
    // Returns the cell's ONWARD DISTRIBUTION for this wave: the edges it lights, each with a
    // relative weight, keyed by edge identity. nil = leaf.
    //
    // The cell owns the policy. The kernel only scales.
    Observe(origin Space, impulse any) map[string]Alloc

    // Express: WHAT HAPPENED AFTER TRAVERSAL?
    //
    // The post-order fold. It SEES ITS CHILDREN'S RESULTS — which the current signature does not
    // allow, and which is the difference between a fold and a side effect.
    Express(origin Space, impulse any, children []*Result) any

    Agents() map[string]Space
    Content() any
}

// Alloc is one entry in the onward distribution: a child, and its RELATIVE weight.
type Alloc struct {
    Target   Space
    Weight   float64 // magnitude. the kernel scales budget by it and does NOT normalize.
    Polarity int     // +1 / −1. MEANING, not transport. see transport.md.
}
```

```text
Observe = what should the kernel look at next?
Express = what should happen after traversal?
```

No hidden state. No second method to keep in sync. The wave's surface is a **return value**, which
is what lets it be a function of the wave.

## Two changes from the 2026-07-08 draft

**1. `Alloc`, not bare `Space`.** The original returns `map[string]Space` and has the kernel look up
edge strength in the slice's activation records. `traverse.go` already does better: the cell hands
over its own weights.

That **subsumes `proportional-amplitude-allocation.md` completely**. The cell decides how to
distribute; the kernel multiplies. The kernel never learns what a weight *means* — not alignment, not
probability, not relevance, not decay.

**2. `Express` receives its children.** The current signature is:

```go
Express(origin Space, impulse any) any     // cannot see what came back
```

which is fine when `Express` is a side-effecting cell method and **useless when it is a real fold**.
Groundedness, plan selection, and scatter-gather rollup all require the parent to see what its
children returned. This is the one signature change the reform did not anticipate, and it is
non-negotiable.

## The DAG is not the executor

From `dag-kernel-responsibility-boundary.md`, and it is still exactly right:

```text
TODAY (wrong)                      INTENDED
DAG creates worker                 DAG:    choose step, mount the ephemeral node,
worker.Observe does the WORK               attach observe/express, hand it to the kernel
orchestrator reads condition       KERNEL: Observe → traverse → Express
DAG routes next worker             WORKER: Observe QUERIES/ACTIVATES.
                                           Express DOES and COMMITS.
```

> **Observe asks: what is available / active here?**
> **Express says: what happened after traversal?**

v2 inverted this — the genes did their substantive work in `Observe` and `Express` was vestigial,
which is precisely why the kernel became scenery and the orchestrator became the real executor.

**Getting this right is the whole point of the revision.** If `Observe` still does the work, the
kernel is still decoration.

## Durable children versus the traversal surface

They are different, and the reform's insistence on separating them is what makes the scheduler
possible:

```text
Children()  the DURABLE graph. what exists. (memory, topology, adjacency.)
Observe()   the PER-WAVE surface. what THIS wave lights, and how brightly.
```

A retrieval wave and a present wave standing on the *same* cell light *different* edges with
*different* weights. That is only expressible if the surface is computed per wave — which is what
returning it from `Observe` buys.

See [scheduler.md](scheduler.md).
