# Skyra — Current Direction and Code

**Date:** 2026-07-16
**Scope:** Current `src` implementation and the mechanisms recorded in
`specs/2026-07-14`.
**Status:** Project snapshot. The v3 filesystem experiment is implemented. Relational
regions, closure-duration learning, target-rooted kernel planning, learned parsers,
and parser compilation are design direction, not current code.

## Executive summary

Skyra is becoming a graph-based, embodied learning system in which an LLM is a
temporary interpreter at the unresolved edge of experience, not the center of the
runtime.

The project has settled on one experiential loop:

```text
observation -> state change -> action -> observation
     O              dS          A          O'
```

Its three relationships are independently learnable:

```text
P(dS | O)   interpretation: how an observation revises represented state
P(A | dS)   policy:         how Skyra responds to that revision
P(O' | A)   consequence:    what the world returns after the action
```

The current v3 code proves the first important substrate claim: a being can make
falsifiable claims, act in a sandboxed world, settle those claims through targeted
observation, grade them without an LLM judge, and retrieve successful precedent on
later episodes. Its measured learning curves are passes-to-closure and prediction
error.

The current design direction generalizes that experiment. All observation,
state-change, and action nodes become durable roles in one experiential ontology.
Current state is the observation-authored graph itself. Desired state is a target
subgraph in the same vocabulary. Their difference creates an error inside a
relational region, and that region runs a learned playbook to close it.

The emerging objective is not merely "did the action work?" but:

> Given this relational region, which action closes its error and keeps it closed
> for the longest observed duration?

Repeated stable behavior should migrate from model inference to graph retrieval and
eventually to a deterministic compiled parser or reflex. The intended cost curve is:

```text
novel situation    -> model inference
familiar situation -> retrieved playbook
stable situation   -> compiled parser/reflex, potentially zero model calls
```

The immediate design focus is the action/behavior loop: fixed or already-resolved
regions, active/dormant renewal intervals, and closure-duration feedback. Region
identity and observation-side parser learning are named but less mature.

## What is implemented now

### Shared graph substrate

The common runtime contracts are in `src/skyra/space.go`:

- `Space` is the universal traversable cell;
- `Slice` owns per-being traversal records, maps, results, and proper time;
- `Time` realizes a graph wave;
- `Impulse` carries a primitive, target, payload, source, amplitude, and opaque
  cognition route state;
- activation records are keyed by edge identity, allowing the same node to have
  different strength in different contexts.

The relation kernel in `src/skyra/relation/relation.go` performs depth-first graph
waves with amplitude attenuation, branch conservation, activation thresholds,
visited-path termination, post-order expression, and brief slice bookkeeping
locks. It deliberately does not interpret cognitive meaning.

The separate `src/skyra/traverse` package is an even smaller budgeted traversal
substrate. It walks a weighted frontier, prunes unaffordable children, returns the
lit subgraph, and exposes the traversed edges as an eligibility set. This clean,
domain-ignorant traversal is close to the kernel role envisioned by the new specs.

### Embodied spatial world

`src/skyra/world` implements a spatial environment with:

- coordinates and placed agents;
- perception, engagement, movement, bump, and wait primitives;
- radius-bounded perception and delivery;
- overlapping occupancy without losing agent identity;
- walls, paths, and deterministic maze generation;
- results returned as explicit world-authored impulses.

`src/skyra/user-dna` represents the human participant. It performs no inference;
it records inbound and outbound impulses and accumulates known entities.

### DNA v2: integrated but superseded cognition

DNA v2 remains the largest integrated cognitive runtime. It contains:

- sensory intake and utterance decomposition;
- episodic, semantic, and causal graph stores;
- causal expansion and signed causal edges;
- budgeted frame construction;
- final-cause and efficient-cause prediction branches;
- grading, reframing, coherence settlement, and credit assignment;
- a static workflow DAG and concurrent per-impulse board;
- motor collapse into `engage`, `move`, `perceive`, or `wait`.

Its fixed workflow is roughly:

```text
sensory intake
  -> grade prior prediction/action
  -> frame current situation
  -> [final prediction || efficient prediction]
  -> atomic join
  -> motor collapse
```

V2 established useful substrate patterns, especially edge-owned activation,
ephemeral workers over durable memory, bounded traversal, correction retries,
transactional causal commits, concurrent joins, and a cognition-ignorant kernel.

Its cognitive architecture is no longer the direction. It depends heavily on
model-authored causal structure and model-mediated grading, runs an expensive fixed
pipeline, and contains acknowledged placeholder mechanics around closure versus
alignment.

### DNA v3: the implemented learning experiment

DNA v3 is a smaller, single-threaded walking skeleton designed to test a precise
claim before reconnecting to the general kernel.

Its central representation is the `Triple`:

```text
subject | predicate | object
```

Observed facts and expected facts use the same type. Therefore grading is an exact
comparison rather than a model judgment.

The implemented v3 laws include:

1. Observation is the only writer of world state.
2. Learning crosses the boundary between Skyra's expression and the world's return.
3. Exact resolution runs before inference.
4. Interpretation occurs on observation, never during grading.
5. Every action must make a falsifiable, deadline-bearing expectation.
6. Observation must be produced without seeing the being's desired answer.

The current stores are:

```text
ObservationGraph
  append-only observation log
  latest observed value per (subject, predicate)

ExpressionGraph
  emitted framings, actions, and expectations
  reusable action->outcome edges
  reverse index from desired triple to prior expressions
```

An edge estimates a counted conditional probability:

```text
P(world state | reusable action signature)
```

Its grade is the model's log-probability advantage over the edge's learned base
rate. The world, not the model, stamps whether a claim has an exact measurement
channel. Unknown observations do not count as failures.

The v3 filesystem world provides `write`, `exec`, `build`, and targeted `perceive`.
It is sandbox-rooted and reports outcomes without inspecting the being's
expectations. A deliberately hidden multi-file build recipe forces the being to
discover prerequisites and later retrieve them as precedent.

The current v3 cycle is:

```text
target
  -> retrieve precedent and assemble a present
  -> model emits acts + falsifiable expectations
  -> parser validates the emission
  -> structural closure checks every action requirement
  -> runnable actions execute
  -> matured bets trigger targeted perception
  -> observations settle bets through exact comparison
  -> failed targets cause observation-informed replanning
```

This implementation measures:

- passes-to-closure;
- inference calls;
- resolved and correct expectations;
- prediction score over the learned base rate;
- parser rejects.

## The current design direction

### One durable experiential ontology

The mature ontology is one durable graph with three node roles:

```text
observation -> state change -> action -> observation
```

Durability is not what distinguishes the roles. Episodes may define active views,
roots, or scopes, but they must not decide which history survives.

This explicitly supersedes the current experimental `NewEpisode` behavior, which
creates a fresh observation graph while retaining the expression graph. The raw
observation record must remain durable; interpretations may be destroyed and
re-derived, but evidence may not be discarded.

"One graph" should be understood as one authoritative ontology and lineage. The
runtime may still maintain rebuildable projections and indexes for current state,
open setpoints, region membership, precedent, and compiled parsers.

### Current state is the observation graph

There should be no separate current-state store to keep synchronized. Current state
is the latest fold over durable observations.

Target state is a desired subgraph expressed in the same vocabulary. The error is
their deterministic difference:

```text
targeted region != observed region -> error active
targeted region  = observed region -> error closed
```

An observation does not automatically turn cognition. It first has to revise the
represented current state or target state. An observation that changes neither is
history, not a new cognitive turn.

### Target-rooted planning through the kernel

Planning begins at the desired state, not at the latest observation. The intended
traversal is backward regression through learned expressions:

```text
desired state
  <- expressions that produced it
  <- requirements of those expressions
  <- until regressed requirements are contained in observed state
```

The closure condition is:

```text
regressed state subset-of current observed state
```

The lit bridge can then be followed forward from current toward desired. The kernel
owns traversal, ordering, admission, termination, and folding. Skyra owns the
target, current morphology, inference at an empty frontier, and action scheduling.

### Relational regions

The `state change -> action` relationship is learned over relational regions rather
than isolated instances.

A region has two separate aspects:

```text
IDENTITY
  a soft weighted co-occurrence signature
  answers: which region does this subgraph belong to?

BEHAVIOR
  an active/dormant renewal profile
  answers: what works here, how quickly, and for how long?
```

Identity is spatial. Its proposed grounded statistic is co-occurrence lift:

```text
P(A, B) > P(A) * P(B)
```

Membership remains a distribution over overlapping regions, not a hard label.

Behavior is temporal. Each region alternates between:

```text
ACTIVE   target mismatch; corrective service is running; minimize duration
DORMANT  target matches observation; no correction is active; maximize duration
```

Identity and behavior draw boundaries independently. A clean region is where
co-occurrence and shared renewal agree. Confident but erratic closure inside one
identity cluster is evidence that the cluster contains more than one behavioral
region and should eventually split.

The present focus is behavior. Identity formation, soft membership, splitting, and
the feedback seam are deliberately less developed.

### Region as playbook

A region's experiential entries form an executable playbook:

```text
observation -> state change -> action -> observation
```

When a target places the region into error, Skyra looks up the state change that
caused the error and retrieves actions that previously corrected that change.

The state change is not abducted at planning time. It is linked to the observation
that changed represented state. What is learned and graded is which action closes
that recorded change and whether the closure persists.

Thus:

- region identity selects a playbook;
- the causing state change selects an entry inside it;
- traversal runs the stored procedure;
- closure behavior grades and reshapes it.

### Closure duration as the slow objective

The action loop has two timescales:

```text
FAST
  choose the most probable action now from P(A | dS)

SLOW
  measure how long each action keeps the region observably closed
  feed that duration back into P(A | dS)
```

The slow question is:

> Given a relational set, what action makes its error disappear for the longest
> observed duration?

A correction becomes a fast reopening, not a special negative event. An action that
briefly satisfies an underspecified target loses to one whose closure persists. The
same signal can expose over-general regions: converged action probability combined
with high closure-duration variance indicates pooled subcontexts.

The hard guardrail is:

```text
un-probed != closed
```

Skyra may not maximize duration by stopping observation. Closure only counts while
relevant observations continue to support the target. Missing observation is
indeterminate or censored.

The exact duration clock and censoring estimator remain unspecified. The current
lean is a count of relevant observations rather than wall time or unconditional
agent ticks.

### Setpoints and nested control

A relational region is a setpoint with two phases:

```text
SATISFIED  current matches desired condition
VIOLATED   current differs; corrective pursuit becomes active
```

The unified rule is:

> Maximize time matched to the setpoint. Given a violation, minimize violation
> time.

Setpoints may nest across abstraction levels:

```text
alive
  -> hydrated
       -> has water
```

A leaf correction staying dormant is its parent condition remaining satisfied.
Stable recurrence may eventually promote a reactive leaf into a higher-level,
feed-forward maintained setpoint.

Competing setpoints are handled as a procedure rather than a universal scalar
ranking:

```text
detect negative co-variation
  -> search for an action that decouples the conflict
  -> if irreducible, satisfice using prerequisite dependence and urgency
```

Setpoint altitude, weight, promotion, and hard-conflict arbitration remain open.

### Understanding as learned parsing

The observation side treats understanding as parsing external input into the
grounded relational floor.

Natural language, protocols, and individual human idioms are all languages whose
parsers are learned through use. A parse is not graded by syntax or by an LLM judge.
It is graded by downstream behavioral stability:

```text
observation -> parse -> state change -> action -> observed stability
                   ^                              |
                   +---- destroy + error trace ---+
```

Only the derived parse is disposable. The raw observation remains immutable and can
be reparsed later.

This gives interpretation and action the same learning shape: emit relational
structure, observe whether downstream behavior remains stable, retain what works,
and destroy/retry what systematically fails.

Observation-side learning is conceptually established but parked while the action
loop is developed. Its central unsolved problem is blame localization.

### Resolved versus unresolved compute

Compute is not routed by which of the three learning edges is active. It is routed
per fragment:

```text
resolved cleanly     -> deterministic resolution
unresolved residual  -> model inference
```

As an edge accumulates probability and evidence, the fragment migrates from
unresolved to resolved. Learning is this boundary moving. The proposed measurement
is therefore not only fewer passes-to-closure, but a rising resolved fraction per
call.

A high top probability is insufficient for a reflex. The margin over rivals matters:

- a clear winner may execute deterministically;
- close rivals are an ordering/scheduling issue;
- a probability cliff, where learned structure ends, is the frontier where the
  being must infer or act to create new ground.

The strongest inference is expected near `observation -> state change`, where
unresolved human meaning enters. Downstream relationships should increasingly run
as learned reflexes.

### Parser compilation and the path below one model call

The current v3 experiment retains a floor of one model call: precedent can suggest a
plan, but the model confirms that it applies.

The new direction treats sustained stability as a JIT compilation signal. A hot,
stable path becomes a deterministic parser:

```text
LLM        general interpreter; expensive; handles cache misses
parser     compiled behavior; deterministic; handles a known region/path
```

Both sides may compile:

```text
observation -> state change   observation-side parser
state change -> action        compiled playbook/action parser
```

Executing a compiled parser without model confirmation is speculative execution.
It permits zero-call familiar behavior but accepts occasional detected misfires and
fallback. This is only appropriate where misfires are observable and rollback is
actually safe.

The concrete parser representation, compilation threshold, failure attribution,
and acceptable misfire bound remain open.

## Current code versus intended architecture

```text
CURRENT V3                              DIRECTION

separate observation/expression stores one durable experiential ontology
fresh observations per episode         durable observation history; episode as view
model emits complete plans              region playbooks supply familiar plans
brute-force precedent retrieval         target-rooted budgeted traversal
action->outcome probability edges       three separately learned experiential edges
prediction-diff grading                 prediction diff + closure-duration training
one model call minimum                  compiled reflexes may reach zero calls
filesystem task classes                 general relational regions and setpoints
single-threaded experiment              eventual integration with the void kernel
```

The direction retains v3's falsifiable representation, exact grading, independent
world evidence, targeted settlement, and reusable learned edges. It replaces the
temporary storage split and model-centered planning loop.

The direction retains v2's useful mechanics—kernel separation, graph waves,
edge-owned activation, bounded traversal, ephemeral workers, concurrency rails, and
world embodiment—without retaining v2's LLM-judged causal cognition.

## Architectural invariants to preserve

1. **Observation is the only writer of world state.** Plans and actions cannot assert
   their own success.
2. **Unknown is not false.** Missing evidence does not grade an expectation down.
3. **Learning crosses the boundary.** Every grade must point to world-authored
   evidence.
4. **Interpret once, before grading.** The model may not modify the grading path.
5. **Exact first.** Resolve deterministic fragments before exposing a residual to a
   model.
6. **Every learned action is falsifiable.** An action without a measurable expected
   consequence cannot train policy.
7. **Traversal direction does not change edge meaning.** Backward planning may walk a
   consequence edge backward without redefining what was learned.
8. **The kernel remains ignorant of cognition.** It moves and bounds waves; Skyra owns
   meaning, objectives, and scheduling.
9. **Raw observations are immutable; interpretations are replaceable.** Evidence may
   be reparsed but not rewritten.
10. **Un-probed is not closed.** A setpoint cannot earn persistence by avoiding
    observation.

## Main unresolved problems

### Blame localization

An unstable closure is a global alarm over:

```text
observation -> parse -> state change -> action -> world
```

It does not identify the bad link. The design proposes co-variation across repeated
contexts, but it still needs a reliable reparse versus replan versus world-difficulty
decision and enough controlled variation to separate confounded causes.

### Region identity and segmentation

Closure statistics depend on which experiences are pooled into one region, while
behavioral instability is also intended to reshape region boundaries. The identity/
behavior feedback loop, region formation, soft-membership arbitration, and split
operation are not designed.

### Target creation and scheduling

The specs explain what happens after a region is targeted but not yet how a worker
chooses to place a region into target state, how several active setpoints are
scheduled, or where setpoint altitude and weight originate.

### Duration and censoring

Closure duration needs a defined clock, a probing policy, treatment of still-open
and still-closed episodes, normalization across different evidence counts, and an
exploration policy that can retry an abandoned action after the world changes.

### Decision cost and safety

Probability predicts outcomes but does not by itself represent consequence cost,
irreversibility, authority, or catastrophic downside. Speculative zero-call parsers
are safe in the filesystem sandbox only while failure is cheap, observable, and
reversible. A broader world will need an admissibility boundary even if probability
remains the learning weight.

### Physical storage shape

One durable graph is a strong ontology but should not forbid multiple derived
indexes. The implementation needs to distinguish authoritative lineage from
rebuildable projections without recreating competing sources of truth.

## Recommended immediate experiment

The next experiment should isolate the action/behavior loop without learned region
identity or model inference.

Use:

- fixed region identities;
- one observable setpoint per region;
- two or three actions per state change;
- mandatory relevant probes;
- occasional environment changes;
- explicit active and dormant interval recording.

Test whether closure-duration feedback makes `P(A | dS)` converge on the action that
keeps the setpoint satisfied longest.

Include five cases:

1. one clearly superior action;
2. a world change that makes the old winner worse;
3. two hidden subcontexts deliberately pooled into one region;
4. confident but erratic closure that should trigger the split tripwire;
5. an abandoned action becoming useful again, requiring exploration.

Measure:

```text
active/service duration
dormant/closure duration
relevant-observation coverage
censored versus completed intervals
action probability and rival margin
action-selection regret
closure-duration variance within each region
```

If that loop works, region behavior has an empirical foundation. Learned identity can
then be attached to it, followed by observation-side parser learning and finally
parser compilation.

## Bottom line

Skyra's direction is no longer an inference-heavy cognitive pipeline. It is a
self-compiling relational control system:

```text
experience creates durable relational structure
structure forms regions
regions become corrective playbooks
observed closure duration reshapes those playbooks
stable playbooks compile into reflexes
the model remains at the unresolved frontier
```

The code already demonstrates falsifiable prediction-diff learning and precedent
retrieval in a real sandbox. The next architectural claims—relational-region
behavior, closure-duration learning, durable unified lineage, target-rooted kernel
planning, and zero-call compiled reflexes—are coherent but remain to be proven in
that order.
