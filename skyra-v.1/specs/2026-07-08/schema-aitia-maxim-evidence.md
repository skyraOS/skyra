# Schema / Aitia / Maxim / Evidence

**Date:** 2026-07-10 · **Status:** ACCEPTED FOR EVIDENCE-SIDE IMPLEMENTATION
**Supersedes:** the evidence-side data model and grading proposals in
[aitia-maxim-telos-schema-praxis.md](./aitia-maxim-telos-schema-praxis.md) and
[open-items-we-need-settled.md](./open-items-we-need-settled.md)

This is the canonical specification for the evidence-side cognitive graph.
Praxis, reward, utility, action obedience, and progress toward Telos are outside
this scope and remain deferred.

## Canonical Graph

The durable memory topology is a cyclic directed graph (CDG):

```text
Entity -> Schema -> Aitia -> Maxim -> Evidence -> Entity
```

The cycle is intentional. A relationship that reaches zero receives no ordinary
traversal, but its immutable target node may later be reached through another
live path around the graph.

## Immutability And Edge Locality

Every durable node is immutable after birth:

```text
Entity
Schema
Aitia
Maxim
Evidence
```

Node identity, kind, and content never change. Revision or reorganization creates
or reuses another immutable node.

Edges are separate graph records. An edge's identity, endpoints, and relationship
kind are immutable. Its context-local activation record is mutable:

```go
type Node struct {
    ID      string
    Kind    NodeKind
    Content any
}

type Edge struct {
    ID     string
    FromID string
    ToID   string
    Kind   EdgeKind
}

type EdgeActivation struct {
    EdgeID      string
    Identity    float64
    Familiarity float64
    GradeCount  int
}
```

Nodes are reused. Edges are not reused across contexts. Reusing an immutable node
creates a fresh edge with fresh activation state; source-edge strength and
history are never copied.

Adjacency lives in graph edge records, not mutable child maps stored inside
immutable nodes.

## Schema

A Schema always has exactly four role-bound Aitia nodes:

```text
Schema
  --material-->  Aitia{CauseKind: Material}
  --formal-->    Aitia{CauseKind: Formal}
  --efficient--> Aitia{CauseKind: Efficient}
  --final-->     Aitia{CauseKind: Final/Telos}
```

Never more and never fewer.

The four positions mean:

```text
Material Cause:
  concrete facts, constraints, resources, and circumstances

Formal Cause:
  organizing pattern, structure, or defining kind of situation

Efficient Cause:
  process, beliefs, events, or actions that produced the state

Final Cause / Telos:
  end, purpose, or successful outcome toward which it is directed
```

Schema membership is immutable. Replacing any position creates a new Schema node
with four fresh Schema-to-Aitia edges. Unchanged Aitia nodes are reused.

Final/Telos uses the same evidence-side transport, edge, grading, and traversal
mechanics as the other three positions. Its special relationship to action and
progress is deferred to Praxis.

## Aitia

An Aitia node is one immutable causal statement occupying one Schema cause role.
It is not itself a four-cause container.

An Aitia participates in exactly two outgoing relationship classes:

```text
Aitia --inherited-aitia--> Aitia
Aitia --maxim-----------> Maxim
```

`inherited-aitia` records causal composition/provenance. `maxim` records an
action-guiding principle proposed from that cause. These edge classes remain
distinct.

## Schema Composition

Schema first tries to reuse existing Aitia nodes. Existing structures receive
the first opportunity to absorb current prediction error.

```text
residual below creation threshold:
  reuse an existing immutable node through a fresh edge

residual above creation threshold after local repair:
  create a new immutable node and fresh edges, then regrade
```

Reuse is decided by functional fit—residual magnitude difference—not fuzzy text
equality or destructive semantic merging.

Cause-generation prompts constrain each immutable cause statement to a grounded
phrase of approximately six to eight words. The worker extracts causal structure
rather than summarizing dialogue and may not add unsupported speculation.

When one Schema position combines several source Aitia nodes, Schema creates one
new same-kind Aitia node containing an explicit semantic synthesis:

```text
A1.Material:
  user has a low battery

A2.Material:
  routes have sparse charging locations

M7.Material:
  user travels under constrained energy resources

M7 --fresh inherited-aitia edge--> A1.Material
M7 --fresh inherited-aitia edge--> A2.Material
```

Composition rules:

```text
all sources match the destination CauseKind
the worker states the synthesis explicitly
unresolved compositions are invalid
content is immutable after birth
source nodes are reused
every source receives a fresh provenance edge
the synthesis is grounded entirely in its sources
```

Near-semantic duplicates may coexist. Their fresh context-local edges compete
through tension, resonance, and proportional traversal. Redundant weak
relationships naturally lose magnitude.

## Active Schema State

Active state is not stored on immutable Schema nodes. Each Entity's active Schema
reference is mutable runtime state owned by the Slice:

```go
type SliceActiveState struct {
    ActiveSchemaByEntity map[string]string // EntityID -> SchemaID
}
```

Schema traversal starts from Entity roots, resolves or composes the Schema, and
commits the active Schema ID on the Slice. Maxim reads its root from that Slice
state.

```text
Entity roots
  -> Schema worker
      -> resolve/compose Schema
      -> Slice.ActiveSchemaByEntity[entityID] = schemaID
          -> Maxim traversal starts at schemaID
```

Reorganization replaces the Slice reference after commit; it never mutates the
prior Schema. A bounded active-state history may retain previous selections.

## Maxim And Evidence

A previous Maxim becomes proposed when traversed. Stored but untraversed Maxims
remain available memory only.

```text
traversed Maxim = proposed
graded edge strength = contextually supported or opposed proposal
Praxis adoption = deferred
```

A Maxim may be related to several Aitia nodes through independent fresh edges.
The Maxim is the sole owner of its Evidence records. One Evidence record has
exactly one owning Maxim.

Evidence is an immutable observation with nonnegative mass. It has no categorical
state or intrinsic polarity:

```text
Evidence content:
  what was observed

Evidence mass:
  nonnegative amount of observed signal

graded relationship:
  contextual signed meaning in [-1, 1]
```

Do not store `support`, `weaken`, `defeat`, or `clarify` categories on Evidence.
Those meanings are derived from graded relationship polarity.

Current evidence sources:

```text
world -> Skyra impulse:
  observation; eligible Evidence

Skyra -> world impulse:
  intervention; not Evidence

world -> Skyra return/result:
  observed consequence; eligible Evidence
```

This prevents Skyra's own emission from confirming the Maxim that caused it.
Praxis intervention grading is deferred.

## Signed Edge Strength

LLM grades are bounded:

```text
-1 <= grade <= 1
```

The grade is recorded and sampled into slow Identity and fast Familiarity EMAs:

```text
Identity    += 0.02 * (grade - Identity)
Familiarity += 0.10 * (grade - Familiarity)

edgeStrength = (Identity + Familiarity) / 2
```

Interpretation:

```text
+1 = absolute resonance
 0 = no net charge
-1 = absolute tension
```

Tension and resonance describe a contextual relationship, not truth, morality,
reward, or intrinsic node value.

Traversal uses nonnegative magnitude while polarity carries meaning:

```text
polarity  = sign(edgeStrength)
magnitude = abs(edgeStrength)
```

Strong positive and negative relationships are both informative and traversable.
A zero edge has no ordinary proportional traversal claim.

Zero is interpreted with history:

```text
zero with insufficient Evidence mass:
  untested

zero after sufficient Evidence mass:
  defeated/neutralized

zero after substantial opposing grades:
  contested but net-neutral; grade history explains why
```

Defeat belongs to an edge, never an immutable node. A zero edge stays defeated.
Its target node may be reached through another live CDG path, producing a fresh
edge in the new context rather than resurrecting the old edge.

## Ungraded Traversal

No grade is not a synthetic grade of zero.

```text
graded traversal:
  update Identity and Familiarity from the signed grade

ungraded traversal:
  do not sample-update either EMA
  leave Identity unchanged

elapsed time:
  lazily decay Familiarity toward zero
```

Canonical lazy-decay shape (the decay coefficient is a tuning parameter):

```text
Familiarity(t) = Familiarity(lastGrade) * decay^(elapsedTicks)
```

This lets recent charge fade while durable Identity remains.

## Evidence-Mass Distribution

For a Maxim with nonnegative Evidence mass `B`, LLM grades determine how `B` is
distributed across participating Aitia-to-Maxim edges:

```text
allocation_i = B * grade_i / sum(abs(grade))
```

If several grade records target one edge during a phase, reduce them to one
evidence-mass-weighted grade before updating its EMAs once:

```text
phaseGrade_e = sum(mass_j * grade_j) / sum(mass_j)
```

After EMA updates commit, the layer's recomputed nonnegative weighted magnitude
becomes distributable mass for the next grading layer. The next layer supplies
fresh signed grades.

```text
nextMass = currentMass * weightedAverage(abs(edgeStrength_i))
```

Mass says how much signal exists. Grade says where it goes and in which signed
direction. EMA state says what the relationship has learned over time.

## Traversal Roots And Grading Order

```text
Schema traversal:
  starts from Entity roots

Maxim traversal:
  starts from active Schema roots stored on the Slice
```

Grading is sequential and commit-barriered:

```text
1. Maxim grading
   Evidence mass -> Aitia-to-Maxim edge grades/EMAs

2. Aitia grading
   committed Maxim results -> Schema-to-Aitia edge grades/EMAs

3. Entity grading
   committed Aitia results -> Entity relationship grades/EMAs
```

Each stage reads the committed output of the previous stage. It cannot consume
its own partial writes during the same stage.

## Prediction Error And Local Repair

Persistent EMA strength and immediate prediction error are separate signals.

```text
error = (actual - expected)
      / (abs(actual) + abs(expected) + epsilon)
```

One logical GradingEpisode may contain several kernel traversals:

```text
grade current layer
  -> local reorganization when error crosses threshold
  -> regrade against the same Evidence snapshot
  -> compute residual error
  -> stop or escalate residual to the next layer
```

Canonical principle:

> Each layer gets one or more chances to absorb prediction error locally. Only
> residual error propagates upward, and the substrate—not the LLM—decides whether
> that becomes a cascade.

Workers see only their local structure, current Evidence, and permitted local
operations. Scheduler/Board owns retries, residual calculation, escalation, and
termination.

Required safeguards (thresholds and attempt caps are tuning parameters):

```text
use one stable Evidence snapshot per episode
accept a reorganization only when fit improves
cap attempts at each layer
stop when residual falls below the layer threshold
escalate when local repair cannot absorb enough error
never let an LLM worker recursively schedule itself
```

A Maxim failure emits residual error, not an Aitia verdict. The Aitia worker uses
the same Evidence snapshot and residual mass to grade the four contextual
Schema-to-Aitia edges. If local Aitia repair cannot absorb the residual, the
substrate mounts Schema reorganization and regrades the replacement Schema.

## Context Budget

For agent cognition, wave amplitude is usable context-token budget. Node cost is
the rendered token size paid from that budget.

```text
wave.Amplitude = usable context token budget
node.Cost       = rendered token count
```

Traversal allocates available amplitude by absolute edge magnitude, checks
affordability, pays node cost, and propagates the remainder. There is no fixed
ancestry depth; context budget, magnitude, node cost, and cycle protection bound
each wave. The durable graph may continue growing over time.

The amplitude knob is tuned below the provider's hard context limit to reserve
room for instructions, transient framing, and output.

## Deferred Praxis Axis

This specification does not decide:

```text
how Praxis selects among proposed Maxims
how interventions are graded
how world results measure progress toward Telos
reward or utility
action obedience
```

These require a separate Praxis specification after the evidence-side model is
implemented.

## Required Invariants

```text
Schema has exactly four role-bound Aitia nodes.
Every durable node is immutable.
Changing node meaning creates or reuses another node.
Edges are context-local and never inherit strength/history.
Evidence is immutable, nonnegative-mass, Maxim-owned, and polarity-free.
Signed meaning belongs to graded relationships.
Traversal uses absolute magnitude and carries polarity separately.
Zero edges receive no ordinary allocation; nodes remain reachable by other paths.
Active Schema references live on the Slice.
No grade is not grade zero.
Only Familiarity decays lazily without new grades.
Only residual prediction error propagates upward.
Scheduler/Board—not LLM workers—owns cascades.
Praxis and Telos-progress mechanics remain deferred.
```
