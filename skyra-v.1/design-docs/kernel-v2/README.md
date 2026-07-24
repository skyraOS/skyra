# kernel-v2

**Status:** DESIGN — nothing built. Companion to [skyra-dna-v3](../skyra-dna-v3/README.md) and
[distributed-agent-v2](../distributed-agent-v2/README.md).
**Lifts from:** `specs/2026-07-08/` — the reform was written, marked superseded, and never built.

## The uncomfortable fact

> **The central kernel has not been used for two generations.**

- **v2** suspended the Perception Invariant (`relation.go:56`) and carried results forward through
  `go kernel.Realize` side channels. The forward pipeline was the executor; the kernel was scenery.
- **v3's walking skeleton has NO traversal at all** — `Closed()`, `Precedent()`, `present()` are flat
  scans. Brute force, correct at this scale, and honest about it.
- **`traverse.go` was never the kernel.** Its own header: *"intentionally isolated from relation.go
  and the world — the collapse onto `skyra.Space` is Phase 3."* **Phase 3 never happened.** A
  budgeted traversal was written, in isolation, and never called by anything.

So the kernel is a design that has never run. Not degraded — **unexercised**.

## What is being reformed

```text
REFORM     skyra.Time  +  skyra.Space        ← the ACTUAL kernel (relation.go)
FOLD IN    traverse.go's mechanics            ← the Phase-3 collapse that never landed
DELETE     traverse.go                        ← once absorbed
RESTORE    the Perception Invariant           ← suspended in relation.go:56
```

## The reform is un-superseded

Every kernel-reform doc in `specs/2026-07-08/` carries a 2026-07-12 supersession banner, and every
one points at the same place: *"superseded by the layered graph —
`cause-and-effect/subgraph-and-causal-edge-ownership.md`."*

**v3 dissolves that layer.** `CausalNodes`, `AitiaSet`, the annotation grammar, causal-belief
expansion — deleted, because causality stopped being asserted and became *the shape of the loop*. The
thing that superseded the reform no longer exists.

**The reform stands, and it was right.**

| spec (2026-07-08) | what survives |
|---|---|
| `space-kernel-interface-reform.md` | **`Observe` returns the traversal surface.** The central fix. |
| `dag-kernel-responsibility-boundary.md` | **The DAG is not the executor.** Observe *queries*; Express *commits*. |
| `proportional-amplitude-allocation.md` | The bug it found is real. Its fix is **superseded** — see below. |
| `directed-affordance-invocation.md` | An invocation is a **directed crossing, from → to**. |

## The one law

> **The kernel carries no cognition.**

It walks a graph and folds what comes back. It does not know what a triple is, what *grounded* means,
what a grade is, or that AND and OR exist.

This is not hygiene. It is what puts the kernel on the right side of the bitter lesson: a mechanism
for **persistence and credit across time** is something a model structurally cannot do for itself.
Hand-designed *cognitive choreography* gets absorbed by the next model. A void kernel does not.

**The moment cognition leaks into the kernel, it becomes choreography, and it dies.**

## What changed since 2026-07-08

**1. Polarity is dead.** The old model carried signed alignment because v2's `A` was a judge's
opinion on `[-1, 1]`. **In v3 the weight IS `P` — a probability.** Probabilities have no sign. `0.1`
means *"unlikely to happen,"* not *"negatively aligned."* Delete it.

**2. Budget does not exist.** No `budget − cost`, no affordability check, no proportional split.
**Amplitude is a STOP CONDITION** — *traverse until this is met* — and the cell computes it as the
present is built.

**3. Weights order; they do not portion.** A context window is **global**. There is nothing to split
between branches; the whole present must fit. So weights are **priority**, not portion, and the walk
is **best-first**: expand the most probable, fill until the condition trips.

**4. The nodes are STATES. The edges are EXPRESSIONS.** *"The kernel moves from one previous state to
the next state. That's its job."* Goal regression, not a bipartite triple/expression walk.

**5. The wave's origin and target are the INVERSE of the agent's.**

```text
THE AGENT'S TARGET   the DESIRED state.
                     ↓ the substrate materializes the nodes
THE WAVE'S ORIGIN    the DESIRED state.   ← the wave STARTS here
THE WAVE'S TARGET    the CURRENT state.   ← the wave is trying to REACH here
```

**6. There is no second wave.** The lit subgraph of the retrieval walk **is** the present.
`traverse.go` already said it: *"the tree IS the frame."*

## The parts

| doc | what it settles |
|---|---|
| [interface.md](interface.md) | the `Space` + `Time` contract. Observe returns the surface; Express sees its children. |
| [traversal.md](traversal.md) | states and expressions, regression, best-first, `Done()`. Replaces `transport.md`. |
| [scheduler.md](scheduler.md) | behavior = (cell × wave purpose). the dispatch table, and the god-object risk. |
| [bootstrap.md](bootstrap.md) | cold start. materializing a target that has no nodes. standing vs concrete targets. |

## When to build it

**Not yet.** The skeleton's one-hop retrieval should drive `passes-to-closure` to 1 for
single-expression tasks (it already does: `[2 1]`) and **stall** on multi-step bridges, because a
flat posting-list lookup cannot see a chain.

**That stall is the signal.** A falsifiable prediction about when the kernel becomes necessary — so
the decision gets made by data rather than by appetite.

## Open

- **Concurrent descent with a deterministic replay mode.** Required by rewind; also what the world's
  device fan-out needs. One change, both payoffs.
- **A typed failure channel.** `BUDGET-EXHAUSTED` / `NO-PRECEDENT` / `CYCLE` / `UNREACHABLE` — and
  `NO-PRECEDENT` is not a failure at all. **It is the wave asking the model a question.**
- **Memoize on key.** Two acts often require the same state. The walk becomes a DAG, not a tree.
- **AO\*.** Best-first over an AND/OR graph is genuinely harder than A\*. See
  [traversal.md](traversal.md) — v0 sidesteps it, deliberately and explicitly.
