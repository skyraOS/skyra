# Open Items We Need Settled

**Archive:** superseded decision ledger; retained for design provenance only.

**Date:** 2026-07-10 · **Status:** SUPERSEDED LEDGER · 2026-07-12

> This ledger belongs to the superseded four-node Schema/Aitia/Maxim/Evidence
> model. Do not resolve or implement its remaining questions as current
> architecture. The active ledger is
> [Minimal Loop Uncertainties](../cause-and-effect/minimal-loop-uncertainties.md),
> owned by the [layered graph](../cause-and-effect/subgraph-and-causal-edge-ownership.md).

Most evidence-side open items discussed on 2026-07-10 are settled and dispersed
into their canonical owning specifications. This file remains the resolution
index and tracks the remaining questions that must be settled in their canonical
owners. It does not define independent architecture.

## Resolution Map

| Concern | Canonical owner |
|---|---|
| Immutable Entity/Schema/Aitia/Maxim/Evidence nodes | [Schema / Aitia / Maxim / Evidence](./schema-aitia-maxim-evidence.md#immutability-and-edge-locality) |
| Schema is exactly Material/Formal/Efficient/Final | [Schema](./schema-aitia-maxim-evidence.md#schema) |
| Aitia inheritance and synthesis | [Schema Composition](./schema-aitia-maxim-evidence.md#schema-composition) |
| Node reuse versus birth by residual magnitude | [Schema Composition](./schema-aitia-maxim-evidence.md#schema-composition) |
| Active Schema references on Slice | [Active Schema State](./schema-aitia-maxim-evidence.md#active-schema-state) |
| Maxim proposal and Evidence ownership | [Maxim And Evidence](./schema-aitia-maxim-evidence.md#maxim-and-evidence) |
| Observation versus intervention | [Directed Affordance Invocation](../directed-affordance-invocation.md#observation-versus-intervention) |
| Tension/resonance and zero-edge defeat | [Signed Edge Strength](./schema-aitia-maxim-evidence.md#signed-edge-strength) |
| Identity/Familiarity EMA behavior | [Ungraded Traversal](./schema-aitia-maxim-evidence.md#ungraded-traversal) |
| Evidence-mass distribution and backward magnitude | [Evidence-Mass Distribution](./schema-aitia-maxim-evidence.md#evidence-mass-distribution) |
| Absolute-magnitude traversal allocation | [Proportional Amplitude Allocation](../proportional-amplitude-allocation.md) |
| Amplitude as context budget and node token cost | [Space / Kernel Interface Reform](../space-kernel-interface-reform.md#cost-budget-and-amplitude) |
| Sequential grading and local error absorption | [Prediction Error And Local Repair](./schema-aitia-maxim-evidence.md#prediction-error-and-local-repair) |
| Multi-wave grading episodes and cascade ownership | [DAG / Kernel Responsibility Boundary](../dag-kernel-responsibility-boundary.md#grading-episodes-and-cascades) |
| Telos parity on evidence side | [Schema](./schema-aitia-maxim-evidence.md#schema) |

## Open Evidence-Side Items

These items are upstream of Praxis and should be settled before the evidence-side
implementation is considered architecturally closed.

### 1. Residual Error And Fit

Define `actual`, `expected`, improvement, absorbed error, and residual error for
each grading stage:

```text
Maxim grade/repair
Aitia grade/repair
Schema reorganization
Entity-to-Schema grade
```

The Scheduler/Board must be able to calculate escalation and termination without
asking an LLM whether a cascade should continue.

### 2. Contested Net-Neutral Edges

Resolve whether an edge with substantial opposing grade history but net strength
near zero is:

```text
dormant:
  zero magnitude receives no ordinary traversal

or still informative:
  polarity tracks signed net direction
  magnitude separately tracks accumulated charge/disagreement
```

The chosen rule must reconcile contested history with absolute-magnitude
traversal.

### 3. Maxim-Owned Evidence Provenance

Define how one observed affordance result becomes Evidence for multiple Maxims
while every Evidence record has exactly one owning Maxim. Settle:

```text
the durable edge/path from Evidence back to its source observation/result
whether each Maxim receives a distinct Evidence receipt
whether observation mass is copied or conserved across those receipts
idempotency when the same result is ingested more than once
```

### 4. Evidence Mass And EMA Learning

Decide whether Evidence mass affects only distribution/cascade propagation or
also affects the strength of an edge's EMA update.

Current equations make one isolated low-mass grade and one isolated high-mass
grade contribute the same EMA step. If that is intentional, record it as an
invariant. Otherwise define how mass affects learning rate, effective sample
count, or update eligibility.

### 5. Explicit Cascade Stages

Resolve the stage model around Schema reorganization. The current canonical order
names:

```text
Maxim -> Aitia -> Entity
```

but Aitia residual may mount Schema reorganization before Entity grading. Define
whether Schema is an explicit Board stage or a bounded repair substage of Aitia,
then align `CurrentStage`, attempt caps, trace spans, replay, and termination with
that decision.

## Deferred, Not Open In This Scope

Praxis is a separate action-side axis. The evidence-side implementation does not
decide:

```text
how Praxis selects among proposed Maxims
how Skyra's interventions are graded
how world returns measure progress toward Telos
reward or utility
action obedience
```

These require a dedicated Praxis specification after the evidence-side model is
implemented. They must not be filled in ad hoc by the current implementation.

## Evidence-Side Status

The canonical graph and ownership boundaries are accepted. The open items above
are remaining evidence-side mechanics, not Praxis questions. Numeric thresholds,
EMA decay constants, retry caps, and context-window amplitude remain
implementation tuning parameters unless later evidence forces a structural
revision.
