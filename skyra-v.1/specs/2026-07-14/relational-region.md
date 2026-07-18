# Relational Region — Identity and Behavior

**Status:** MECHANISM (2026-07-15 research). High-level only — no implementation.
**Companions:** [the-learning-loop.md](the-learning-loop.md) ·
[closure-duration-objective.md](closure-duration-objective.md) · children:
[relational-region-behavior.md](relational-region-behavior.md),
[relational-region-identity.md](relational-region-identity.md)

The `state change → action` axis is graded over **relational regions**, not single instances. This
doc defines the region and — more load-bearing than the definition — **partitions it into two
processes that must be designed separately.**

## The definition

> **A relational region = a weighted co-occurrence signature (identity) + a renewal profile
> (behavior), with the boundary drawn where the two agree.**

## Two processes, not one

Resolving *which region you are in* is not the same act as *behaving while inside it*. They are
different processes, run off different signals, and conflating them is the trap.

```text
IDENTITY   spatial.   which entities + edges is this?          answers MEMBERSHIP.
           "given a subgraph, what region does it belong to?"
           diff = co-occurrence above the independence baseline.
           → relational-region-identity.md

BEHAVIOR   temporal.  how does the region act over time?        answers the OBJECTIVE.
           the renewal profile: active → dormant → active.
           diff = kernel-tick intervals (active vs dormant).
           → relational-region-behavior.md   (objective: closure-duration-objective.md)
```

Identity answers *where am I*. Behavior answers *what do I do here, and did it hold*. Both are "the
region." They answer different questions and are measured off different signals — which is exactly
why they get separate specs.

## The seam — where the two meet

Each process draws a region boundary, and a clean region is where they **agree**:

- co-occurrence says *these args appear together*,
- shared renewal says *these args close and recur together*.

Where they **disagree** is the split signal: a co-occurrence cluster that shows *confident-but-erratic*
closure (the tripwire in [closure-duration-objective.md](closure-duration-objective.md)) is one
spatial region hiding two behaviors, and it wants cutting. Regions are never hand-defined; they
crystallize where "appears together" and "resolves together" coincide.

## The feedback loop (named, not designed)

Identity and behavior almost certainly feed back — identity selects which behavior applies;
behavior's splits reshape identity's boundaries. That loop is real and out of scope here. **We
design them separately first**, precisely so the loop has two clean endpoints to connect later.
Current focus: **behavior**.

## Open

- The reconciliation mechanism (the seam) — how agreement/disagreement is computed and acted on.
- Whether identity or behavior is prior, or neither.
- The feedback loop itself.
