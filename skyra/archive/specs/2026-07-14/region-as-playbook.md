# Region as Playbook — the Operational Framing

**Status:** MECHANISM (2026-07-15 research). High-level, and a **cleaner operational restatement**
of the behavior side — captured to return to.
**Companions:** [relational-region-behavior.md](relational-region-behavior.md) ·
[setpoints-and-conflict.md](setpoints-and-conflict.md) ·
[closure-duration-objective.md](closure-duration-objective.md) ·
[traversal-as-excitation.md](traversal-as-excitation.md) ·
[../../design-docs/skyra-dna-v3/the-cycle.md](../../design-docs/skyra-dna-v3/the-cycle.md)

> **2026-07-18 refinement:** The newer traversal-as-excitation direction treats current state and
> target state as potentially derived roles inside one active interference pattern, not necessarily
> as separate objects. The current/target language below remains an operational description of the
> observed and directional sides of the pattern while that representation is worked out. See
> [traversal-as-excitation.md](traversal-as-excitation.md).

A more concrete framing of the same behavior mechanism, with one simplification: **there is no
current-state to maintain.**

## Current state is not maintained — it *is* the state graph

There is no separate "current state" representation to keep in sync. Current state is the **state
graph** itself — the fold over observation nodes (= world state in the README). It is always there,
authored only by observation. Nothing to maintain; it just *is* the record.

The tighter working statement is:

> **The active field is the local relational workspace in which current state, target state, and
> the presently lit bridge between them coexist.**

The **relational field** contains the durable regions and their histories. The **active field** is
the temporary working subset recruited into cognition. What precisely constitutes current state
and target state inside that workspace—and how observation decay, target excitation, and traversal
produce them—remains open here rather than being fixed prematurely.

Target state and current state are the **same kind of object** — both are subgraphs in one
vocabulary (from [the-cycle.md](../../design-docs/skyra-dna-v3/the-cycle.md): a projection and an
observation are world-state subgraphs of the same type). That is *why* current state needs no store
of its own: a target is just a region of the graph flagged as desired, and the current value of that
region is already in the graph. The **error is the diff between them.**

## Targeting creates error

A relational state is either just part of current state (untargeted, quiescent) **or** a worker has
placed a region of the state graph into **target state**.

- Placing a region into target = **creating an error**: for that region, current ≠ target.
- Error is not discovered by scanning — it is *created by the act of targeting*.
- This is the setpoint's violated phase, operationalized: violated = *a worker put a target on it.*
  Untargeted = satisfied. Targeted = in error = active.

## Two modes — normal running and error

A region is always present in current state and **running normally** — it exists, it is maintained,
it is watching. That is the baseline: no corrective work, but not gone either.

Being **placed into target state** flips it into the error mode. Only then does the corrective
playbook fire.

```text
NORMAL (current state)   region present, maintained, no gap    → no corrective work
ERROR  (target placed)   current ≠ target                      → run the playbook
```

Normal running is *monitoring, not polling* — a region re-checks when a relevant observation
arrives, per "staleness comes from perception, not decay." It does not burn cycles watching itself
with no input.

## The objective, restated

- A region in error has one goal: **minimize time in error.**
- On exit — when current matches target — the region **stamps that bounded activity with a tick.**
- Those tick-stamped intervals are the renewal profile of
  [relational-region-behavior.md](relational-region-behavior.md); "minimize time in error" is the
  minimize-below (minimize pursuit) of [closure-duration-objective.md](closure-duration-objective.md).

So the whole active phase is a single bounded, tick-stampable event: **target placed → error → work
→ match → tick.** No maintained comparison; a created interval that closes and gets stamped.

## The pipeline — the model only translates

The one model call is a **translation**, not an interpretation:

```text
observation ──(cheap, stable model)──► relational structure
relational structure ──(deterministic diff vs. the graph)──► state change
state change ──(playbook)──► action
```

The model turns natural language into relational structure and is told **nothing else** — not the
objective, not the playbook, not that errors exist. Meaning is not the model's job: the **state
change is the diff** between the translated structure and the graph, computed deterministically. An
agent doing local work without knowing why it does it.

That flips the model requirement at this layer: we need **stability, not intelligence** — a cheap
model (Haiku, or a constrained parser) whose output is *reliable* run-to-run. And it must be, because
**learning happens on stable observations**: if the same input yields different relational structures
each time, every edge downstream is built on sand. A creative model here is a bug, not a feature.
LAW 3 makes it both cheap and stable — exact-first resolves the world-addressable parts, so the model
only touches the residual, which is less to vary on.

This refines the earlier "compute flows to interpretation." The interpretation isn't a big model
call — it is a deterministic diff against a *learned* graph. The compute is amortized **into the
graph**, not paid per-call. The persistent model role shrinks to stable translation; the expensive
interpretation is only cold-start, and it migrates to a diff once the region is learned.

## The region is a playbook

A relational region contains the loop entries:

```text
observation → state change → action → observation
```

Given a target set within the region, the region **runs the playbook** — the stored
`obs → ΔS → action → obs` procedure that has closed errors in that region before. So:

- the region's *content* is a **runnable procedure** (its precedent / retrieval surface),
- **identity** (which region) selects *which playbook*,
- **behavior** (minimize time in error, tick-stamp on exit) *grades and shapes* the playbook.

"Which playbook do I run, and did running it close the error fast and keep it closed" is the whole
loop, per region.

### The lookup is keyed on the causing state change — which is *known*

In error, the playbook is not consulted as "region → best action." It is consulted as:

> **What state change put this region into error — and what action fixes that?**

This is **not a diagnosis and not a judgment.** By LAW 1 only observation writes world state, so the
state change that put the region into error was written by *some observation* — and that link is
recorded at observation time (LAW 4). The cause is always an observation, **never an action**, and
it is *known*, not abduced. "What caused this error" reads a logged fact; the playbook maps that
known state change to its correction.

The same region-error still takes different fixes for different causes — "Kathy is thirsty" because
she drank the last glass (refill) versus because she just ran (electrolytes) is one region, two
recorded state changes, two corrections. The playbook is indexed by the **state change**, not just
the region. Nothing here is an opinion to grade; what gets graded is the correction (`ΔS → action`),
by whether it closes the error and keeps it closed.

## What this buys

- **Drops a store** — no separate current-state representation; it is the state graph.
- **Error becomes an event** — created by targeting, bounded, tick-stampable — not a standing
  comparison that must be maintained.
- **Region = playbook** unifies *which region am I in* (identity) with *what do I run* (the stored
  procedure). Retrieval and region-membership become the same lookup.

## To return to

- How a worker decides to **place a region into target state** — the targeting/scheduling trigger.
- **Playbook selection under soft membership** — a subgraph carries a distribution over regions
  ([relational-region-identity.md](relational-region-identity.md)); which playbook runs.
- **Running the playbook = traversal** — relationship to the kernel/loop that executes
  `obs → ΔS → action → obs`.
