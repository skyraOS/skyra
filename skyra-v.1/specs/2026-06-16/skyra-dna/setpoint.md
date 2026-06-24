# Setpoint — Cell Type Spec

**Date:** 2026-06-19 · **Status:** in spec (updated)

## File
- `../../../src/skyra/skyra-dna/setpoint.go`

## What it is

**A setpoint is a prediction that came true.** Not a goal set in advance — a confirmed prediction that has graduated into durable topology. The system builds its model of the world from successful predictions, not from declared intentions.

**Stacking.** A setpoint sits at a level. On a grouping's closure, the grouping's prediction graduates into a new setpoint at level n, the grouping's constituents become its children (n-1), and the grouping's gist crystallizes into a durable **Understanding**. That new setpoint(n) is then eligible to be a constituent of a grouping at n+1 — the pyramid of abstraction forms by resolved understandings becoming raw material for the next round, not by a separate merge pass. "Understanding" is recursive: it caps the level below and is a constituent of the level above. See [groupings-and-understandings.md](../../../design-docs/cognition/groupings-and-understandings.md).

## A setpoint is one node: understanding as content, surroundings as children

This resolves the old "setpoint(n) vs Understanding(n)" question — **one node:**

- **`Content()` = the understanding** — the crystallized gist (what this setpoint *means*).
- **`Children()` = the surroundings snapshot + evidence at closure** — the **entities** involved (person/place/thing) + responses + actions + stresses.

So: *"I predicted x → x came true; here's the gist (content); here's what was around when it closed (children)."* Past-facing only; **no parent/future edge** (the future is Global's grouping index, passed down as context — see [global.md](global.md)). Map keyed by edge identity, so activation is edge-owned.

### The directed cycle (all `children`, not bidirectional)
The entities in a setpoint's children point on to *their* setpoints, forming a cycle through alternating layers — all forward `children` edges:

```
understanding --children--> entity --children--> setpoint(=understanding) --children--> entity --> …
```

This is what makes associative recall a **wave** rather than a hash lookup: a wave fired from the surroundings entities spreads `understanding → entity → understanding …`, bounded by amplitude. See [entities-and-surroundings.md](../../../design-docs/cognition/entities-and-surroundings.md).

### Completion — the write-time invariant
A setpoint is **created when the predicted self AND other both come true** (both faces close). At that completion:
- the understanding crystallizes as the setpoint's content;
- the surroundings snapshot + evidence become its children;
- **stamp `BirthAlignment = closure-grade × global-stress-at-birth`** on the node (set once — the intrinsic excitability);
- **append the setpoint to each involved entity's `Children()`** — the reverse-index edge. *The reverse read depends on this append* (same write-time discipline as the justification index / epistemic stamp).

In-flight predictions are **not** stored here — they live transiently in the WM buffer and only graduate to a setpoint on completion. Failure cascade is the mirror: the retracted delta → WM → children of a **Stress**.

## Edges vs potentials

The wave carries two things through a setpoint (both in the past direction now):

- **Edges (declared adjacency)** — the setpoint's direct past neighbors, the entries in its `Children()` map. The map **key is the edge identity** (`"A→B"`), so the activation record (on the slice) is owned by the edge — A→B and C→B differ.
- **Potentials (computed, not stored)** — transitive reach. A's excitation of a 2-hop cell (A→B→C) is the path composition `ab ⊗ bc`, computed at traversal time via amplitude attenuation. A-specific because the amplitude started at A. Not a declared edge.

**Crystallization (Hebbian):** a potential that grades above the keep-threshold (the same `0.1` that gates adding stimulus to children) becomes a **declared edge** — the source adds the target to its own `Children()` adjacency and the edge-keyed record seeds on the slice. The path collapses into a synapse; next wave the source reaches that cell in one hop.

**History:** scored potentials that *didn't* crystallize (including negatives) persist as edge-keyed records on the slice with **no** adjacency — they hold their value but the kernel never walks them (it only walks `Children()`). Edges that stop earning grades decay below threshold and fall away. The graph forms *and* prunes — the structural-learning channel.

## Interface (`skyra.Space`)
- `Children()` → one map, **keyed by edge identity**, past only: the **entities** involved + responses + actions + stresses (the closure snapshot + evidence). The cell owns the adjacency; the slice owns the edge-keyed activation record (keyed off the map key). No future/parent edge — the future arrives top-down as Global's grouping context.
- `Content()` → the **understanding** (the crystallized gist this setpoint means)
- `Observe` → descent: gate-check against current impulse (under the active grouping context); contribute to activation score
- `Express` → ascent: update alignment on the edge's activation record; vote

## Activation record

Activation scoring is **edge-owned**: the record lives on the slice keyed by the edge (the `Children()` map key), not by the shared target cell — so A→B and C→B score independently. The kernel calls `Activate(visited, traversalCount)` and gets a score back — it never sees the equation, which has **three loci**:

```
bump = ( IndexScore × Alignment ) × stress_modulator × intrinsic
       └── per-edge (record) ──┘   └ global (slice) ┘  └ per-node ┘
```

- `IndexScore` — surroundings overlap (one-shot per cycle); **per-edge**
- `Alignment` — lagged from the previous Frame 2 output for this edge; **per-edge**
- `stress_modulator` — the being's **global** stress (read from the slice, a neuromodulator — not per-edge)
- `intrinsic` — **`BirthAlignment`**, this node's cell-intrinsic excitability (below)

Both Identity (α=0.02) and Familiarity (α=0.1) EMAs receive this bump. See [activation-and-traversal.md](../../../design-docs/kernel/activation-and-traversal.md).

### Birth alignment — set once, at crystallization
A setpoint is born with an **intrinsic** strength, distinct from any edge — the brain's cell-intrinsic excitability / memory-allocation (engram recruitment), or in predictive-coding terms its **precision**. Kept as a **struct** so the constituents stay separable (room to add valence/emotion-at-birth, novelty later); stamped once at completion:

```
Intrinsic { ClosureGrade, StressAtBirth }       // stamped once
BirthAlignment = Intrinsic.Value() = ClosureGrade × StressAtBirth
```

A clean, high-stress closure is born strong (a "flashbulb" memory); a marginal one is born weak. It is **not** an EMA (those are accumulated/per-edge) and **not** stress (that's global) — it's a one-time per-node stamp. It gives a fresh setpoint cold-start confidence (before edges accumulate) and feeds **solid ground** for the cascade (a high-BirthAlignment node is intrinsically solid — the cascade halts on it).

## The `completed` score (signed, inference-judged, transitory)

A **lit** setpoint carries a `completed` score — "how done is this, really," a **signed −1..1 inference judgment** (a judgment call, not arithmetic from edge alignments). It is **distinct from `Alignment`**:

| | Alignment | Completed |
|---|---|---|
| what | node ↔ setpoint relational match | the setpoint's own done-ness |
| source | feeds the **activation bump** (excitation) | **inference judgment** on lit setpoints |
| scope | per-edge | per **lit** setpoint, **transitory** |
| triggers | how excited / in-view | **+threshold → crystallize; supporting `< 0` → cascade**; between → in-progress |

**`completed` does not drive excitation** — excitation (the bump) decides what's lit/in-view; `completed` is a readout *on* the lit setpoints, gated by lit-ness. A setpoint that loses excitation drops out of view ("not important at the moment") regardless of its `completed`. So `completed` is transitory: meaningful only while lit, not maintained out of view.

## Cascade — state rollback, triggered by a supporting setpoint going negative

A setpoint's **state** is its tower of crystallized layers — and its **supporting setpoints** are the constituent `n-1` setpoints in its `Children()` (what it crystallized from; the stacking). A setpoint rests on its supporting setpoints.

**Cascade trigger: a supporting setpoint is judged `completed < 0`.** Not the setpoint's own dip — the **ground underneath** going negative (pure TMS: the justification is refuted). The cascade is a failure wave propagating *down* through supporting setpoints that are `< 0`, halting at the first one still **held** (`completed ≥ 0` — solid ground). `BirthAlignment` makes a node intrinsically more likely to hold.

- **Delta** = the span from the contradiction down to solid ground — everything resting on the refuted ground.
- Rolled back out of state (TMS dependency-directed backtracking), **deposited into WM**. Solid ground survives intact.
- The popped span becomes **potential groupings again** → Global reorganizes → crystallizes as a **Stress** (the failure-understanding).
- **Nothing deleted** — rollback into WM, recoverable. Mirror of success closure (resolved arc → Understanding); cascade (retracted delta → Stress).

`completed` is the inference judgment; the cascade *propagation* (walk supporting setpoints, stop where `completed ≥ 0`) is **code**. See [error-recovery-loop.md](../../../design-docs/cognition/error-recovery-loop.md#cascade--a-failure-wave-de-crystallization).

## Open
- Exact shape of Understanding and Stress as child cells (their own specs TBD)
- How Responses and Actions attach structurally
- Root drive: is it a setpoint (a prediction that always comes true) or membrane `Content`?
- `completed` thresholds (find by experiment): the +bar for crystallize; whether solid-ground is exactly `≥ 0` or a small positive bar. Cascade trigger is fixed at supporting `< 0`.

## Done when
- A setpoint with a confirmed prediction exists, is traversed by the kernel, survives the Phase 1 filter, and participates in the Phase 2 LLM cycle (Frame 1 → Frame 2 → Vote).
