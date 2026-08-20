# The Phase Rule

**Date:** 2026-08-05
**Status: RULE.** Short on purpose. This is the thing to follow when writing any behavior.

> Entered `2026-08-06` intact. One rule, unchanged since it was written.

> **`Perturb` READS. `Observe` WRITES.**

---

## The rule

```text
PERTURB    reads durable state and per-execution state.  WRITES NOTHING.
           decides WHERE to go. returns one Space, or nil.

OBSERVE    writes. commits to the graph, ACCUMULATES world effects, folds.
           decides WHAT IS TRUE. never decides where.
```

Sharper, because "read" and "write" both need a boundary:

```text
PERTURB    may read anything. may write nothing — not the graph, not the world.
OBSERVE    may write. reads ONLY WHAT IT IS HANDED — its accumulation and the child's
           realization. it does not go looking.
```

This replaces the looser "`Observe` decides nothing." `Observe` decides plenty; what it never
does is **select**. Selection is `Perturb`'s and only `Perturb`'s.

## Why it is load-bearing, not tidy

**Eviction is safe at any point during descent.**

The runner evicts under capacity pressure, and an evicted stack's accumulation dies unwritten
(`kernel/runner.go`). That is only a clean operation if descent has changed nothing. Under this
rule it has:

```text
EVICTED MID-DESCENT   nothing was written. nothing happened. the transcript says how far it got.
EVICTED MID-ASCENT    a PREFIX of the writes landed. append-only makes a prefix consistent —
                      each write is complete in itself, there are just fewer of them.
```

If `Perturb` could write, eviction would be a rollback problem, and there is nothing to roll back
with. It also means **an evicted wave never fires an action**: a file's primitive runs in
`Observe`, so a stack cut off on the way down cannot have flipped the switch.

That property is worth more than the symmetry.

## It holds across every job

```text
                PERTURB (read)                          OBSERVE (write)
INTERPRET   build the frontier, return a WALK       commit interpretations, write support edges
EXPECT     build the frontier, return a WALK       AddExpectation, with its perceive
ACT         nil — an action is a leaf               execute the primitive against the world
BOUNDARY    resolve the mount                       spawn the crossing
WALK        take one member, return the rest        fold into accumulation (ephemeral only)
```

The pattern that recurs: **`Perturb` finds, `Observe` commits.** INTERPRET's exact pass locates a
resolution on the way down and writes it on the way up. That is not a workaround for the rule; it
is what makes descent free to be abandoned.

## The API split

Behaviors do not get the namespace. They get a view, and which view depends on the phase:

```text
Reader     Root · Node · Observation · Entries · index lookups
Writer     Reader, plus AddExpectation · AddAction · support writes
```

```go
type Behavior interface {
    ID() string
    Perturb(h *Handle, g Reader, origin kernel.Space, impulse any) kernel.Space
    Observe(h *Handle, g Writer, origin kernel.Space, impulse any, realized kernel.Space) kernel.Space
}
```

**Structural, not conventional.** `Perturb` has no writer to call, so the rule cannot be broken by
forgetting it — the same move as the kernel having no method that would let it classify, and
`AddExpectation` having no way to omit a perceive.

The durable API therefore splits cleanly in two, which is the split worth maintaining:

```text
NOT BEHAVIORS   the graph's own methods. Ground · AddExpectation · AddAction · Node · Root ·
                Observation. these are what a namespace can do to itself.

BEHAVIORS       consume those methods, through a Reader or a Writer depending on phase.
```

## What it does not settle

```text
WHERE INFERENCE GOES   a wave never calls the model; the residual is collected and spent at
                       the assembly boundary, after the wave (behaviors.md). so neither phase
                       is "the inference phase" — Perturb finds what it cannot resolve,
                       Observe accumulates it, and the boundary asks.

READS DURING ASCENT    "Observe reads only what it is handed" is a rule, and there will be a
                       case that wants to look something up on the way back. when one appears,
                       it is evidence the walk went down the wrong path — not a reason to
                       widen Observe.
```

---

**Paths:** the graph package now lives at
`src/skyra/cognitive-agent/skyra-dna-v4/graph`.
