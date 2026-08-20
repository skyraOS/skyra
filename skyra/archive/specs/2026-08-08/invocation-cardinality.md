# Invocation Cardinality

**Date:** 2026-08-09
**Status: MERGED 2026-08-09.** It survived. Promoted into
[`2026-08-09/grammar.md`](../2026-08-09/grammar.md) — the group form, the frame boundary, and the
per-member `cite` rule — with the depth/width/cites signal in `credit.md` and the barrier cost kept
open as `questions.md` §8. **Cite those, not this.**

> One thing was underspecified here and had to be settled on merge: whether `cite` attaches to the
> group or to each member. It must be **per member**. Grouping is a scheduling fact; a group-level
> cite would make one observation the cause of N selections, which reintroduces sibling blindness at
> the one place the design has it solved.

> **Every selection awaits. Concurrency is expressed by how many possibilities are selected at
> once, not by a blocking mode.**

## The shape

A selection emits an invocation group. The runtime invokes every member of the group concurrently
and returns their observations together.

```text
SELECT { A }
await A
SELECT { B }

    A → B
```

```text
SELECT { A, B }
await A and B

     ┌→ A
  ───┤
     └→ B
```

The first shape establishes order. The second says that, given the current frame, neither
invocation requires the other's return before it can run.

```text
CARDINALITY = 1    one causal step.
CARDINALITY > 1    a concurrent frontier chosen by the possibility.
```

The kernel does not find, select, or order that frontier. It executes the group the possibility
already emitted.

## Frame boundary

Every member receives observations from the frame as it existed when the group was emitted. Members
cannot see one another's returns. When the whole group completes, its returned observations enter
the caller's frame and may shape the next selection.

This makes the invocation history a partial order without adding `block`, `async`, or `await` modes:

```text
A then B, and B cites A       confirmed dependency.
A then B, but B never cites A possible unnecessary serialization.
A and B in one group          declared return-independence at that point.
```

The last claim is narrow. Concurrent members may share inputs and contribute to the same eventual
return. They are independent only in the sense that neither needs the other's return to begin.

## What this may buy

The possibility decides both selection and available parallelism. Its history then exposes another
learning signal: whether it found the dependency structure of its own work.

```text
DEPTH    the dependencies it believed had to be sequential.
WIDTH    the work it believed could safely proceed together.
CITES    which declared dependencies were actually exercised.
```

No separate local span syscall is required by this shape. The runtime holds one awaited group at a
time; the possibility supplies the next group after the current returns.

## The cost

An invocation group is a barrier. One slow or trapped member delays the next selection even if
other members have already returned enough to continue. This proposal accepts that constraint for
the smaller mechanism. Whether partial completion is ever necessary remains open and should not be
smuggled in as an execution flag without a case that requires it.
