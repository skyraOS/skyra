# One Kernel — Cognition And World On The Same Traversal

**Date:** 2026-07-13
**Status:** SCRATCHPAD — raw working thought, not yet a design or implementation commitment
**Companions:** `self-authored-loop.md` (the cognitive loop), `README.md` (potential wells),
`distributed systems design docs/` (overview-spec, ephemeral-world-workers)

---

# Part I — Michael

## What An Agent Is In This Runtime

An **agent** is a series of spaces that belong to the same underlying *thing* — a bounded group
of specialized spaces.

That's the definition. It is not "a thing that thinks."

## World Agents Are Not Cognitive

World agents are strictly spaces that cognitive agents move over, do work on, mount, etc.

There is not one agent. There are many. Right now exactly one of them has cognition.

```text
skyra-agent    cognition — nucleus, genes, learning        (moves)
world-agent    spaces — topology, state, affordances       (is moved over)
device-agent   spaces — a bounded group of device spaces   (is mounted)
```

The skyra-agent moves across the world and mounts other world-agents / devices.

## The Kernel Might Be The Same Kernel

Kubernetes has the same pattern, and it mirrors Observe/Express traversal.

We could potentially use the **same kernel** we use for the skyra-agent for the world-agent —
where the world does distributed-systems computation with it.

---

# Part II — Claude

## First, A Correction I Owe The World Layer

I had been talking about the world as if it *grades*. It does not. `ephemeral-world-workers.md`
invariant 6 forbids it and it's right:

```text
world workers      determine WHAT HAPPENED     → Outcome
cognitive workers  determine WHAT IT MEANS     → grade
```

The world returns `succeeded / blocked / denied / stale / unavailable / cancelled / partial /
indeterminate`. That is a report, not a verdict. It has never seen Skyra's prediction, and it
must never see it.

**And that is exactly why it is trustworthy.** The world has no stake in the bet. It cannot be
flattered and it cannot be gamed, *because it doesn't know what answer was hoped for*. An LLM
judge — which sees the prediction it is grading — can never have that property. Independence is
the entire value.

```text
cognition makes the bet   (private)
world says what happened  (independent, no stake)
cognition diffs           (the grade)
```

The two layers never touch. Grading is 100% cognition-side.

Related: `denied` is **a peer agent refusing**, not a permission check Skyra runs on itself. Same
for `blocked` and `unavailable`. Those are another agent's answers, which is what makes a mount a
crossing rather than an API call.

## The Kernel Already Is A Distributed-Systems Kernel

```text
child_budget = (parent_budget − parent_cost) × weight
descend iff  child_budget ≥ child_cost
```

**That is gRPC deadline propagation.** A request carries a deadline, each hop spends some and
passes the remainder down, and refuses a call it cannot afford to finish. Same arithmetic, minus
the weight.

`traverse.go` already says why, in a comment that turned out to be more prophetic than intended:

> Cost is the node's weight against the budget, **in the budget's units** (tokens for cognition,
> 1 for a world hop).

Milliseconds is just a third unit. The budget axis is **polymorphic**:

```text
tokens        cognition — how much thought this walk can afford
hops          spatial   — how far a perceive reaches
milliseconds  network   — how long a distributed call may take
```

And the deeper identity: **Observe/Express is scatter-gather.** Descend = fan the request out,
attenuating budget. Ascend = fold partial results back. That's MapReduce; that's a query planner;
that's k8s walking owner references. A distributed read *is* a budgeted traversal with a
post-order fold.

## Share The Kernel, Not The Lifecycle

The traversal fits wherever the computation is *"spread a bounded budget over a graph and fold the
results back"*:

```text
FITS                                    DOES NOT FIT
perceive (radius = budget)              invocation lifecycle
discovery (walk the network)              resolve → validate → commit is a TRANSACTION
scatter-gather across devices             (ordering, preconditions, rollback — not a fold)

                                        the reconcile loop
                                          the watcher CALLS a traversal when error rises;
                                          it is not itself one
```

So: **one kernel, two lifecycles above it.** That is `ephemeral-world-workers.md`'s "behavioral
symmetry, not ontology collapse" — sharpened. The symmetry lives *in the kernel*. The asymmetry
lives in what sits on top.

## Three Things Distributed Use Breaks In `traverse.go`

**1. There is no failure channel.** Today `walk` either descends or it doesn't. There is no
"descended and got an error." In cognition an unlit node is simply absent from the frame — fine.
In a distributed walk, an unreachable device is a **partial failure**, and `Express` must be able
to fold honestly over a hole. This is invariant 7 (*transport is hidden, failure is not*) and the
kernel currently cannot express it.

```text
Result needs a failure channel
Express needs to see the hole
```

**2. The visited-once rail does not cross machines.** `visited map[string]bool` is an in-memory,
per-walk set. The moment a walk crosses a device boundary it must travel with the impulse or the
walk loops. Same problem as a distributed trace context, same solution — carry it on the wave —
but it has to be *decided*, not discovered at 2am.

**3. Descent is sequential — and this is the one that matters.** `walk` iterates children one at a
time and recurses. Depth-first, single-threaded. Fine (arguably good — deterministic) for
cognition. For perceiving across four devices it is a latency catastrophe: they'd be hit serially,
each one's deadline burning while the previous runs.

## The Kicker

The kernel needs **concurrent descent with a join at the fold**.

That is the *exact same missing feature* that planning-parallelism needs (see
`self-authored-loop.md` — descent must produce a prerequisite DAG, not a chain, and independent
prerequisites must be walkable concurrently).

```text
concurrent descent + join at the Express fold
        ↓                           ↓
skyra pursues independent      the world fans a perceive
sub-goals in parallel          across N devices without serializing
```

**One change to the kernel. Both payoffs.**

This is the strongest evidence so far that these are not two projects. The world is not a system
built *after* cognition works — it is the same traversal primitive with a different budget unit and
a different lifecycle bolted on top, and the single feature it forces into the kernel is the feature
cognition already wanted.

## Open

- Does the reconcile watcher live above both lifecycles, or does each agent-kind get its own?
- Concurrent descent breaks determinism (today siblings are visited in sorted edge-key order).
  Is determinism worth preserving as an option — a `sequential` mode for tests and replay?
- Where does the visited-set / trace context ride? On the `Impulse`, or on a new invocation envelope?
- Failure folding: does one failed child fail the parent's `Express`, or does the parent decide?
  (Almost certainly: the parent decides — the cell owns the policy, the kernel only mechanizes.)
