# Horizon and attention — the frame, and who creates it

> **PARTLY SUPERSEDED (2026-09-06):** [focus-and-move.md](../2026-09-01/focus-and-move.md) names
> per-node available choices Focus and makes them explicit on the live move, revising the no-stored-field
> rule below. Choices are distinct from committed dependencies. Horizon remains kernel-created and
> workflow-scoped. The kernel-free hot-path claim is also superseded by
> [the kernel/scheduler ownership decision](../2026-09-01/engram-loading-and-fault-return.md).

**Date:** 2026-08-27
**Status: DECISION (this phase).** Resolves the open child-horizon / `bound()` question the v11 CTU build
surfaced. Supersedes the CTU-side narrowing Codex implemented.

---

## Two layers, two owners

```text
HORIZON    the TOTAL permission set of a ROOT / WORKFLOW. workflow-scoped, not per-node — the whole
           possibility tree runs under one horizon. AMBIENT to the workflow; a node never manipulates it.
           CREATED BY THE KERNEL (provisioning). = what the workflow COULD do.

ATTENTION  PER-NODE, created by THINK during PLANNING: a node's direct children in the frame — "what the
           child can do in the frame." lives WITHIN the horizon. the children (the candidate set) are ON
           the Context frame (see children-and-the-frame.md); Work[Issued] is only the ACTIVATED subset.
           = what a node can engage; what it IS engaging is Work[Issued]. (REFINED: children are stored on
           the frame, not derived from Work — the base case must SEE them to choose.)
```

`attention ⊆ horizon` still holds — but they are created at different times by different owners: the
kernel bounds the horizon (authority), THINK lays out the attentions (plan structure). The two are not
the same operation seen twice; they are distinct.

## No narrowing this phase

There is **no parent→child horizon narrowing** right now. The workflow shares the root's horizon; THINK
plans each node's attention within it. A cut does **not** carry a narrowed subset — the child runs under
the workflow's horizon. (A real narrowing may arrive later — e.g. at a cross-workflow / cross-entity
boundary — but it is deferred, not this phase.)

## `bound` is a KERNEL method that CREATES a horizon

The decisive correction: **the CTU must never construct a horizon.** The spine is `human owns policy ·
kernel mints/provisions · CTU executes`. A CTU synthesizing a capability set is minting authority it does
not have.

```text
WAS  (ctu.go) cutSeed() → bound(parent.Horizon, target, action)
     the CTU narrows a parent's horizon into a child seed. an AUTHORITY LEAK — the CTU is creating a
     capability set.

IS   (kernel) bound(...) Horizon
     the KERNEL creates the workflow horizon (provisioning). the CTU only runs within what it is handed;
     it never narrows or mints. cutSeed no longer builds a horizon.
```

So `bound` keeps its name but flips owner and meaning: not a CTU attenuation of an existing horizon, but
the kernel's *creation* of one for a root/workflow.

## What this changes in the build

```text
· MOVE bound() out of ctu.go and onto the kernel (a horizon-creation / provisioning method).
· cutSeed no longer calls bound / no longer narrows — a cut carries the ticket + request + return route,
  and the child runs under the workflow horizon (this phase: the root's, carried from the top).
· CTU-side code holds NO horizon-construction. it verifies offline against the horizon it was handed.
· ownership check restored: only the kernel creates horizons.
```

## What did NOT change

- The **hot path** stays kernel-free at run time: the CTU still verifies the carried horizon OFFLINE.
  Horizon *creation* is the kernel's (cold/provision); horizon *checking* is the CTU's (hot).
- **Attention is still not a stored field** — it's THINK's plan structure, and its runtime-active subset
  is `Work[Issued]`. Nothing about the `Attention`-field removal is reversed.
- The horizon still rides with the workflow so the hot path can verify without a kernel round-trip.

## Open / deferred

```text
· the LATER narrowing — at a cross-entity / cross-workflow boundary, does the child get its OWN horizon
  (provisioned for its principal) rather than the caller's? that is the ticket-vs-child-horizon split;
  parked until the cold/provision path is built.
· the exact kernel `bound` signature (root/principal → Horizon) — settle when the kernel path is combed.
```
