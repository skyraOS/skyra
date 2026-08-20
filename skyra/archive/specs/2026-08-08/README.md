# 2026-08-08

> **RETIRED 2026-08-09.** Its live content was promoted to [`2026-08-09`](../2026-08-09/) and
> decomposed into a spec set. This directory stays as the working record — the reasoning, the wrong
> turns, and the order things were found in. **Cite `2026-08-09`, not this.**
>
> `primitives.md` is the single doc everything here became. `traversal.md` was carried across intact.

**A whiteboard, not a set.** `2026-08-06` stays live and authoritative. Nothing here supersedes
anything there until it is moved there deliberately.

The rule the previous sets were built on applies to this one differently:

> **Strip complexity, do not add it.** Nothing carries across by default. A doc earns entry the
> same way a mechanism does — by being required, not by being coherent with the rest of the set.

`2026-08-06` applied that to docs that had already earned their place. This set is upstream of
that: things are written here **while they are still wrong**, and most of it should die rather
than graduate.

```text
2026-08-06   what is decided. cite it.
2026-08-08   what is being decided. cite nothing from here.
2026-08-07   locality.md was written this way too, and correctly says so in its own header —
             UNDECIDED, held to keep a direction while it is thought about. it belongs to
             this mode, not to the set it is sitting in.
```

---

## On the table

| doc | the question | status |
|---|---|---|
| [primitives](primitives.md) | entity · discover · produce, and the space-as-function signature | **WORKING** |
| [traversal](traversal.md) | should `Perturb` / `Observe` survive as the interface? | **ANSWERED — no.** now the checklist of what the split was carrying |

## What the whiteboard is for

Not to record conclusions. To hold **what is load-bearing on a thing before it is changed**, so a
change is made against the actual list rather than against memory of it.

That list is the only thing worth writing down in advance, because it is what a good argument has
to survive and what a bad one will quietly skip.

## Superseding, not replacing

**Decided 2026-08-09.** The built kernel and graph are `skyra-dna-v4` and have `Perturb`/`Observe`.
This model does not. The fork is resolved by the convention the repository already uses for exactly
this:

```text
KEEP     skyra-dna-v4 stays. it runs, it is tested, and it is the live design's implementation.
MARK     the new direction takes the next version number — skyra-dna-v5.
```

Nothing is deleted and nothing is migrated. A version number is how this project has recorded a
change of direction before, and it costs nothing to be wrong about.

## The standing constraint

The kernel is the one doc that matches running code line for line. So a question about the
traversal interface is a question about the only thing here that is real, and the cost of being
wrong is different from the cost of being wrong about, say, differentiation.

```text
SPEC        wrong is free.
BUILT       wrong has a working implementation and seven tests behind it.
```

Not an argument for keeping anything. An argument for knowing which one you are doing.
