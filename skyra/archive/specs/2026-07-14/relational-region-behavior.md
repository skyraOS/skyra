# Relational Region — Behavior (the Renewal Profile)

**Status:** MECHANISM (2026-07-15 research). High-level only. **The current design focus.**
**Parent:** [relational-region.md](relational-region.md) · **Objective:**
[closure-duration-objective.md](closure-duration-objective.md) · **Setpoint framing & conflict:**
[setpoints-and-conflict.md](setpoints-and-conflict.md)

Behavior is the *temporal* process: how a region acts over time once you are already inside it. It
says nothing about identity — which region this is — see
[relational-region-identity.md](relational-region-identity.md).

## The shape — a renewal process on the kernel tick

While the being is dealing with a region, the region cycles between two states along the kernel-tick
axis:

```text
                   impulse               impulse
                      │                     │
   ──── ACTIVE ───────┤──── DORMANT ────────┤──── ACTIVE ──── ...
     (below the line) │   (above the line)  │
     working to       │   closed,           │
     target;          │   not active;       │
     service time     │   the good state    │
     MINIMIZE         │   MAXIMIZE          │
```

- **Below the line — ACTIVE.** The region is open and the being is working it toward target. This is
  service time. Minimize it.
- **Above the line — DORMANT.** The region is closed and stays gone. This is the good state.
  Maximize it.

Each cycle is one **entry**:

```text
state change → action → observation → state change
└──────────────── one entry, measured in kernel-tick diffs ───────────────┘
```

The region's behavior is the **vector of these entries** across episodes. Successive entries are
compared by their tick stamps; the diff between them is the metric.

## The objective over this structure

Minimize below (active/service), maximize above (dormant). That is the **closure-duration
objective** — see [closure-duration-objective.md](closure-duration-objective.md) for the objective
itself, its hard guardrail (un-probed ≠ closed), and its inner/outer decomposition. This doc is the
*representation* that objective runs over; that doc is the *objective* over this representation.

## Two axes "dense" hides

"Many entries in the vector" is ambiguous, and the below/above split already disambiguates it:

- **Costly** — below-heavy: expensive to close (long service).
- **Frequent** — above-short: recurs often (short dormancy).

Different region characters, wanting different responses. Keep them separate; "dense" is not one
thing.

## Grounded, not judged

Active→dormant and dormant→active are **observed** transitions — the watcher in
[closure-and-planning.md](../../design-docs/skyra-dna-v3/closure-and-planning.md) already fires on
both (closes → done; a `requires` goes false → reopen). The tick-diffs are intervals between
observed events. No judge; LAW 2 holds by construction.

## Open (for the continued design)

- **Delimiting a cycle.** What exactly starts "active" and ends it. The watcher's close/reopen
  events are the candidates — but the reopen must be attributed to *this* region, which reaches into
  the identity process.
- **Which clock.** Kernel tick (agent), world tick, or count of relevant observations that passed
  without contradiction? Lean (from closure-duration): relevant-observation count — agent-relative
  and censors cleanly when the being stops looking.
- **Censoring.** A still-dormant region has not earned "infinite above"; it is *not-yet-recurred*,
  not durably solved. Survival-shaped, never credited as a win.
- **Exploration.** Pure argmax never re-tries an abandoned action, so it cannot notice the world
  changed back. A short closure (recurrence) is the natural re-try trigger — unspecified.
