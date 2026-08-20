# Expectations and Actions — the Shape

**Date:** 2026-07-30
**Status:** MODEL SETTLED. Ordering and the intended collapse are noted below.

**Supersedes** the immediate/consequent typing explored on 07-29. That distinction is not
gone — it is **derived** now rather than declared, which is why the type is not stored.

---

## The model

```text
an EXPECTATION has   0:*  actions
                     0:*  expectations
                     and at least ONE of the two

an ACTION has        ≥1   expectation
                     it is always a LEAF
                     it owns nothing
```

An expectation owns every edge it points at — at its actions and at its child expectations.
It never owns an edge that points at it. Actions own nothing at all, which is what makes them
leaves rather than a second kind of node.

## Grading happens only at leaves

> **The observation worker grades leaf actions. Everything above a leaf inherits.**

This replaces the earlier statement that observations attach to expectations. The graph gains
one layer and the arithmetic is unchanged:

```text
Observation ── support ──▶ Action ── support ──▶ Expectation ── support ──▶ Expectation
```

`T`, `G`, `|O|`, `P`, and coherence are still folds along that path, still products of the edge
scalars, still derived at evaluation and never persisted. See [trust](trust.md).

## The typing is derived, not declared

```text
expectation WITH actions      graded through its own execution
expectation with NO actions   grounded ONLY through its children — strict inheritance
```

Nothing to store. An expectation that has no actions cannot be graded directly, and that is not
a rule imposed on it — it is a consequence of grading happening at leaves and it having none.

This is what the immediate/consequent split was reaching for. Cardinality gets it for free.

## The falsifiability invariant

> **An expectation with no `perceive` in its own sequence, and no child that has one, cannot
> be wrong.**

Consider what happens without it:

```text
E "I am logged in"
   actions: [type_username, type_password, hit_enter]
```

Every action closes. The expectation inherits closure. The being concludes it is logged in
**because it performed the keystrokes** — it has graded its own procedure and called that
success. Add the look and the world gets a vote:

```text
E "I am logged in"
   actions: [type_username, type_password, hit_enter, perceive_screen]
```

`perceive` is the only channel by which an expectation can be wrong about the world rather than
about its own execution. This is checkable structurally, so it can be enforced rather than
hoped for.

The mechanism already exists on the substrate — `edge.go`:

> *at maturity the settler goes and LOOKS, so absence becomes a fact rather than a silence, and
> UNKNOWN is a waiting room rather than a destination.*

Grading-at-leaves is what makes that settler mandatory instead of merely wise.

## Order

```text
SIBLINGS   concurrent      state is parallel
CHAINS     sequential      execution is serial
TRAVERSAL  post-order depth-first
```

Children complete, then the parent's own sequence runs, then it folds upward:

```text
exRoot ──▶ exL1 ──▶ exL2
  │          │        │
  a5         a3       a1
  a6         a4       a2
  a7
```

Execution order is `a1 a2 · a3 a4 · a5 a6 a7`. `exL1` and `exL2` are chained, so they are
sequential. Were they siblings, they would run concurrently.

**No scheduler is required.** This is the kernel's existing shape — `Perturb` descends,
execution happens, `Observe` folds the result up. The pipeline *is* the traversal. See
[kernel-alteration](kernel-alteration.md) §4.

Order is **structural**, not by rank. Trust influences what gets *created* and what gets
*looked at*; it must not decide what runs next, or the executor becomes one more place attention
can gate itself.

## One action, several expectations

An action may be referenced by `≥1` expectations. It executes **once**, is graded **once**, and
is inherited by each expectation that references it.

That is what makes multi-effect physical action representable without duplication:

```text
E "the switch changes open → closed"          actions: [flip]
   ├── E "the light changes off → on"          actions: [perceive_light]
   └── E "the fan changes stopped → spinning"  actions: [perceive_fan]
```

One physical execution. Two separately falsifiable children with separate positions. The light
may close while the fan does not, and neither a compound expectation nor a duplicated act is
needed. Credit splitting stays deleted, because grading happens per leaf action and each
expectation inherits only along its own path.

## Motivation is derived

An observation that justifies creating an expectation needs no `motivated-by` edge. It has a
**temporal signature** instead:

```text
at arrival        activation HIGH · support ≈ 0     nothing accounts for it
                  → the expectation worker creates a target for it
after creation    support strongly SIGNED           it now has a home
```

The observation did not change. The field grew a place for it. Motivation is therefore *an
observation whose polarity moved because the field changed underneath it* — a delta over the
append-only log, stored nowhere.

It is the mirror of revision rate: revision is a reading displaced by new evidence, motivation
is evidence finding a reading that did not previously exist.

**This requires re-scoring.** An observation scored once, on arrival, against a field that
lacked its target will sit at ≈0 forever and never acquire its edge. When the expectation worker
creates or refines, the observations that triggered it must be re-presented to the observation
worker. The substrate permits this — the evidence set is append-only — but it has to be an
explicit step or the mechanism silently does not work.

## The intended collapse

Actions and expectations are separated here for one reason: children of a node must run
sequentially when they are actions and concurrently when they are expectations.

That is a property of the **grouping**, not of the node. `kernel-alteration.md` §2 already puts
it there — an ephemeral Space privately carries *"sequence, fan-out, or join behavior."*

So the distinction is expected to dissolve: actions become leaf expectations, and
sequential-vs-concurrent becomes a flag the parent declares. One type, closed under composition,
the way `Entity` was in v0.4.

**Deferred, not forgotten.** Keeping them separate makes the concurrency rule easier to reason
about while it is still being settled.

---

## Open

```text
THE COLLAPSE          when sequential/concurrent becomes grouping behavior, actions
                      and expectations are one type. what, if anything, is lost

PARENT ORDER          the parent's own sequence runs after its children. whether a
                      parent may interleave instead is not decided

INERT NODES           "at least one of the two" prevents a node with nothing beneath
                      it. it does not prevent a subtree with no perceive anywhere —
                      the falsifiability invariant has to be checked, not assumed

WINDOW                what bounds an action's support prediction. sequential execution
                      within an expectation suggests the next action taking over, but
                      concurrent siblings mean two windows can be open at once
```
