# kernel-v2

**Status:** DESIGN — nothing built. Companion to [skyra-dna-v3](../skyra-dna-v3/README.md) and
[world-agent-v2](../world-agent-v2/README.md).
**Lifts from:** `specs/2026-07-08/` — the reform was written, marked superseded, and never built.

## The uncomfortable fact

> **The central kernel has not been used for two generations.**

- **v2** suspended the Perception Invariant (`relation.go:56`) and carried results forward through
  `go kernel.Realize` side channels. The forward pipeline was the executor; the kernel was scenery.
- **v3's walking skeleton has NO traversal at all.** `Closed()`, `Precedent()`, and `present()` are
  flat scans and one-hop lookups. Brute force. Correct at this scale, and honest about it.

So the kernel is a design that **has never run.** Not a degraded kernel — an unexercised one. That is
what this revision fixes.

## The reform is un-superseded

`specs/2026-07-08/` contains four kernel-reform documents, and every one of them carries a
supersession banner dated 2026-07-12 pointing at the same place:

> *"superseded by the layered graph — `cause-and-effect/subgraph-and-causal-edge-ownership.md`"*

**v3 dissolves that layer.** `CausalNodes`, `AitiaSet`, the annotation grammar, causal-belief
expansion — all deleted, because causality stopped being asserted by the model and became *the shape
of the loop*. The thing that superseded the kernel reform no longer exists.

**So the reform stands, and it was right.**

| spec (2026-07-08) | what survives |
|---|---|
| `space-kernel-interface-reform.md` | **`Observe` returns the traversal surface directly.** The central fix. |
| `dag-kernel-responsibility-boundary.md` | **The DAG is not the executor.** Observe queries; Express commits. |
| `proportional-amplitude-allocation.md` | **Split budget by edge magnitude, not equally.** Signed polarity carried separately. |
| `directed-affordance-invocation.md` | **An invocation is a directed crossing, from → to.** |

## The one law

> **The kernel carries no cognition.**

It spreads a budget over a frontier and folds what comes back. It does not know what a triple is,
what *grounded* means, what a grade is, or that AND and OR exist. Every one of those lives above it.

This is not hygiene. It is what puts the kernel on the right side of the bitter lesson: a mechanism
for **persistence and credit across time** is something a model structurally cannot do for itself,
no matter how capable. Hand-designed *cognitive choreography* gets absorbed by the next model. A void
kernel does not.

**The moment cognition leaks into the kernel, it becomes the choreography, and it dies.**

## The parts

| doc | what it settles |
|---|---|
| [interface.md](interface.md) | the `Space` contract. `Observe` returns the surface. `Express` sees its children. |
| [transport.md](transport.md) | budget, cost, affordability, proportional allocation, signed magnitude |
| [scheduler.md](scheduler.md) | behavior = (cell × wave purpose). the dispatch table, and the god-object risk. |

## What changed since 2026-07-08

The reform was written before the scheduler existed, and before v3. Three things sharpen it:

**1. `Observe` should return weights, not bare children.** The 07-08 spec returns
`map[string]Space` and has the kernel read edge strength out of the slice's activation records.
`traverse.go` already does better: `map[string]Alloc{Target, Weight}` — the cell hands the kernel its
own onward distribution.

That **subsumes proportional-amplitude-allocation entirely**: the cell decides the weights, the
kernel scales by them, and the kernel never learns what a weight *means*.

**2. Conserve-vs-broadcast is AND-vs-OR**, and it was already in the code.

```text
weights SUM TO 1     conserve   → choosing among alternatives   → an OR node
weights ALL 1        broadcast  → every child is needed         → an AND node
```

That distinction was written to separate cognition from space. It turns out to be disjunction versus
conjunction. **The kernel needs no new concept** — it multiplies budget × weight and stays blind.

**3. A cell cannot know what it is on its own.** The same triple cell is an **OR node** in a
retrieval wave and a **leaf** in a present wave. Behavior is `(cell type × wave purpose)`, which is
why the scheduler exists — and why the job board's `Purpose` field is **load-bearing dispatch**, not
observability.

## Open

- **Concurrent descent with a deterministic replay mode.** Everything is parallel (OR children are
  independent; AND children are independent), and the fold is commutative, so determinism costs only
  sorted edge keys and a `(score, key)` tie-break. Required by rewind; also what the world's
  device fan-out needs. **One change, both payoffs.**
- **A typed failure channel.** Not `error`. `BUDGET-EXHAUSTED` / `NO-PRECEDENT` / `CYCLE` /
  `UNREACHABLE` — because the responses differ, and `NO-PRECEDENT` is not a failure at all: **it is
  the wave asking the model a question.**
- **Memoize on key.** Two acts often require the same state. The walk becomes a **DAG**, not a tree —
  correct, and a large constant-factor saving.
- **When to build it.** The skeleton's one-hop retrieval should make `passes-to-closure` fall to 1
  for single-expression tasks and **stall** for multi-step bridges. That stall is the signal. Build
  the kernel when the number says to, not before.
