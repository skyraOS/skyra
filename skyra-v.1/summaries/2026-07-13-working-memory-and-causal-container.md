# Skyra — Current Runtime, Working Memory, And Causal Containers

**Date:** 2026-07-13  
**Scope:** Current `skyra-dna-v2` implementation, the first live world run, and
the design direction discovered from it.  
**Status:** Project snapshot. Sections marked **Design** or **Rough idea** are not
implemented.

## Executive summary

Skyra currently runs as a fixed, inference-heavy cognitive pipeline over a
durable in-process graph. Every observed impulse is decomposed into a causal
node, the current situation is retained/composed/requeried, efficient and final
causal beliefs are expanded, the episode is coherence-settled, efficient- and
final-cause branches run concurrently, and motor collapses the efficient-cause
proposal into a world primitive. The **efficient-cause** lane proposes and
grades Skyra's action; the **final-cause** lane predicts and grades the
anticipated external outcome.

The first sustained live run proved two things at once:

1. The substrate can learn behaviorally meaningful structure. After moving on
   stale location information and receiving criticism, Skyra changed behavior to
   perceive first and then move.
2. The fixed pipeline is too expensive and brittle to remain the architecture.
   Six external messages plus three perception returns produced 95 inference
   calls over roughly 28.5 minutes. Parsing or causal-expansion failure can keep
   motor from acting at all.

The emerging replacement is a two-speed working-memory runtime:

```text
fast:
  register observed material -> assess/regrade meaning -> motor reads snapshot

slow, only after a surprise shock:
  revise the material/formal account -> regroup -> settle -> atomic commit
```

The newest rough idea sharpens the contents of working memory into causal
containers: many efficient causes feed one stable material/formal container,
which maintains one active final cause. Ordinary cognition may search for and
replace the final cause, but only an exceptional positive delta in surprise may
rewrite the material/formal summary.

## What is implemented now

### Substrate

The runtime is built around:

- `skyra.Space`: the universal graph-cell interface;
- `skyra.Time`: the wave-realization/kernel interface;
- per-being `Slice` state containing activation, maps, traversal trees, and
  results;
- edge-keyed activation, so the same node can carry different strength through
  different relationships;
- a relation kernel that performs bounded, visited-once graph waves;
- a separate budgeted traversal used to assemble cognitive frames;
- `MemoryCell` as the durable graph-node representation;
- world primitives for `engage`, `move`, `perceive`, and `wait`.

The model is deliberately split between mechanism and cognition. The kernel
transports waves and records traversal; DNA-owned genes decide what things mean.

### DNA v2

`DNA` owns the being's durable cognitive state:

```text
node/entity registry
episode and frame graph
activation store
causal-node and causal-edge indexes
signed edge strengths
query budgets
registered genes
inference and motor seams
```

Genes are selected into disposable per-wave `EphemeralRoot` transcripts. This
keeps step-local mutable state out of the shared genome and permits concurrent
workers.

### Current fixed workflow

For an external message:

```text
sensory:intake
  -> final:grade
  -> midlevel:frame
  -> [final:predict || efficient:predict]
  -> board join
  -> motor:collapse
```

For an action return:

```text
sensory:intake
  -> efficient:grade
  -> midlevel:frame
  -> predictions
  -> motor
```

For a perception return:

```text
sensory:intake
  -> sensory:ingest
  -> midlevel:frame
  -> predictions
  -> motor
```

The board coordinates concurrent prediction branches, permits one reframe per
impulse, and fires motor exactly once after the join.

### Current causal representation

Every external observation is decomposed into an annotated, observation-born
`CausalNode`. Midlevel chooses one structural operation:

```text
retain    keep the active material/formal representation
compose   extend it using offered causal nodes
requery   replace it with a standalone representation
```

The active situation is represented by an `AitiaSet` containing a present
material/formal causal node and expanded efficient/final causal beliefs. These
beliefs are persisted as causal nodes connected by typed causal records. Signed
edge grades, query budgets, provenance, and evidence references exist around
that graph.

The episode graph is a separate, traversal-oriented spine containing frames,
interactions, perceptions, predictions, proposals, and realized actions.

### Current learning behavior

The runtime currently has several learning/readjustment channels:

- final-cause predictions are graded against the next external message;
- efficient-cause proposals are graded against action reafference;
- edge activation records accumulate frequency/alignment-like strength;
- coherence settle grades episode items against a fixed frame and prunes
  negative items from that settlement pass;
- query budgets adjust how much specificity efficient and final causal workers
  may emit;
- active causal sets are replaced atomically rather than mutated in place.

## What the live run demonstrated

The observed interaction progressed through greeting, location reporting,
movement, criticism of a failed/stale movement attempt, behavioral correction,
an explanation/apology, and a maze request.

The important behavioral sequence was:

```text
Michael asks Skyra to move to him
  -> Skyra moves using the available location
  -> Michael reports the move did not work and says Skyra should look first
  -> Skyra perceives
  -> Skyra moves using the updated perception
  -> later Skyra explains that it relied on the prior observation
```

This is real evidence that the graph, framing, and action history can influence
subsequent behavior. The system did not merely repeat the correction verbatim;
it inserted a perception before the next move.

The same run exposed the present economic and structural limits:

- 95 inference calls were started across approximately 28.5 minutes;
- strict annotated-sentence parsing caused repeated retries over ordinary glue
  words such as `about`, `toward`, `along`, and `before`;
- a causal expansion can fail after retries and terminate the impulse before
  motor;
- a retained frame can still trigger several expensive settle passes;
- a perception return re-enters nearly the whole cognitive pipeline;
- the episode accumulated greetings, location work, movement, apology, and maze
  work into one large settlement context;
- low-probability predictions count as simply `held`, losing the difference
  between confident expectation and a barely covered surprise;
- observed and composed nodes can duplicate nearly identical semantic content;
- process termination leaves in-flight inference records without terminal trace
  events.

The final maze perception retained the current frame, ran three settle passes,
and then started concurrent efficient/final causal calls. The process was
manually stopped while those calls were in flight.

## Design: the working-memory loop

The fixed pipeline is proposed to dissolve into a working-memory substrate.
Working memory is the bounded active field of causal nodes on the Slice. All
cognitive acts traverse that field.

An incoming impulse is registered into working memory. Registration, not the
successful completion of framing, becomes motor's go-signal. Motor becomes a
sovereign subscriber to working-memory snapshots instead of the final stage
behind a prediction join.

The intended two-speed behavior is:

```text
fast loop, per registration:
  observe/decompose -> register -> assess -> act from current snapshot

slow loop, only when substrate dynamics demand it:
  recruit midlevel -> revise frame -> expand/settle -> atomic WM commit
```

This directly addresses the live failure mode: cognition may degrade or fail
without agency becoming mute.

The prediction branches are expected to converge with the causal graph. The
final-cause lane predicts the anticipated outcome in the world; the
efficient-cause lane proposes Skyra's causal action toward it. This naming
uses the causal roles the branches perform rather than perspective-relative
labels.

This design is not implemented and still depends on kernel reform, snapshot
history, working-memory bounds, motor coalescing, self-engage maturation, and a
precise surprise signal.

## Rough idea: causal containers

The newest model gives a working-memory situation the following shape:

```text
efficient causes                 material/formal              active final

A ---------+
B ---------+----> M ----------------------------------------> D
C ---------+
     * : 1                     1 : 1                         1 : 1
```

- A container `M` may have many efficient causes.
- `M` has one current material/formal summary version.
- `M` has at most one active final cause.
- Several final candidates may be attempted and retained as history, but only
  one is actively asserted.
- The final cause represents the currently extracted meaning, direction, or
  outcome of the situation.

### Relational grading

A final candidate is graded independently against every efficient cause:

```text
grade(D | A)
grade(D | B)
grade(D | C)
```

Its provisional aggregate support is a normalized amplitude-weighted score:

```text
score(D) = sum(amplitude(Ei) * grade(D | Ei))
           -----------------------------------
                    sum(abs(amplitude(Ei)))
```

Normalization prevents containers with more causes from receiving artificial
confidence merely because they contain more nodes.

Provisional interpretation:

```text
score >= +0.35              supported final / closure
score <= -0.35              strongly defeated final
-0.35 < score < +0.35       weak or unresolved
```

Magnitude may drive attention, but sign determines whether a final is supported
or refuted.

### Meaning search does not rewrite M

Ordinary cognition holds `M` fixed:

```text
new efficient evidence arrives
  -> regrade active final
  -> if weak, propose another final against the same M
  -> activate a candidate only if it earns positive closure
  -> otherwise remain explicitly unresolved
```

The search must be bounded. Failure to find a meaning is not permission to
invent one, and it is not permission to rewrite the material/formal account.

### Surprise alone authorizes M revision

`M` changes only after an exceptional positive jump in surprise. A provisional
signal is:

```text
support(t)       = support of the active final
surprise(t)      = 1 - support(t)
deltaSurprise(t) = surprise(t) - surprise(t-1)
```

Material/formal revision is allowed only when:

```text
deltaSurprise(t) > dynamicShockThreshold
```

The comparison is directional. A large reduction in surprise means the existing
situation became better explained and should not cause a rewrite.

The cognitive tension becomes:

```text
incoming material introduces unexplained amplitude
inner cognition tries to repress ordinary surprise by finding a better final
exceptional unresolved surprise recruits a new material/formal summary
```

### Bootstrap

When an observed impulse has no home, a new container is created around the
fact of its occurrence:

```text
efficient set = {the observed impulse}
summary       = "the impulse occurred"
```

This is identity grounding: the consequence of saying something is, first, that
it was said. It should carry special `given`, `birth`, or `identity` provenance
rather than act like learned causal evidence. Otherwise self-grounding could
reinforce itself indefinitely.

### Summary versions

The container identity remains stable, but shock-driven summaries are versioned:

```text
M/v1  "Michael said something"
M/v2  "Michael requested movement"
M/v3  "Michael requested movement after a failed attempt"
```

Grades retain the version against which they were made. A slow-loop reframe
atomically changes the active summary version and its active final.

## Built versus proposed

| Capability | Current state |
|---|---|
| Observation-born causal nodes | Built |
| Retain/compose/requery material/formal framing | Built |
| Efficient and final causal expansion | Built |
| Signed causal edges and query budgets | Built |
| Episode traversal and coherence settle | Built |
| Concurrent efficient/final causal branches | Built |
| Board join before motor | Built |
| World perception, movement, and engagement | Built and exercised |
| Bounded causal-node working-memory buffer | Design |
| Registration as motor go-signal | Design |
| Motor as sovereign snapshot subscriber | Design |
| Surprise-recruited midlevel | Design |
| Predictions absorbed into causal beliefs | Design |
| Many-to-one causal container | Rough idea |
| One active final per container | Rough idea |
| Per-efficient-cause grading of a final | Rough idea |
| Surprise-gated material/formal summary versions | Rough idea |

## Safest migration path

This direction should not be implemented as one rewrite. A low-risk sequence is:

1. Preserve richer live-run instrumentation and establish latency/behavior
   baselines.
2. Decouple motor eligibility from successful causal expansion so a cognition
   failure cannot silence Skyra.
3. Add immutable/versioned working-memory snapshots alongside existing episodes.
4. Introduce causal-container records in shadow mode, populated from current
   active sets without changing behavior.
5. Record per-efficient-cause final grades and evaluate candidate aggregate
   scores in shadow mode.
6. Run the surprise monitor in shadow mode against live traces and calibrate it
   before it gains authority to rewrite summaries.
7. Gate material/formal revision on the monitor while retaining the old pipeline
   as an observable fallback.
8. Move motor to snapshot subscription and add ordering/coalescing semantics.
9. Absorb the efficient/final branch records into the general causal-container
   representation only after behavioral equivalence is demonstrated.
10. Retire the board join and fixed workflow after the new loop owns their
    invariants.

## Unresolved decisions

The following remain open and should not be smuggled into implementation as
defaults:

- the exact meaning of a final cause: predicted outcome, intended direction,
  extracted meaning, or a typed union;
- the owner and normalization of efficient-cause amplitude;
- the final-closure threshold and whether it adapts per container;
- surprise estimator, history window, shock floor, and calibration constants;
- working-memory admission, capacity, eviction, and retention of negative
  evidence;
- candidate-search budget and explicit unresolved behavior;
- hysteresis between close final candidates;
- whether efficient causes can retire without a material/formal rewrite;
- representation of edge-to-edge evidence and causal closure;
- stale slow-loop commit prevention when newer impulses register;
- motor ordering when several registrations arrive rapidly;
- bootstrap provenance and protection from self-reinforcement;
- the migration boundary between causal containers and the broader Praxis
  design.

## Current north star

The current implementation proved that Skyra can alter behavior through its
causal and episodic substrate. The next architecture should preserve that
learning while removing the requirement that every cognitive specialist finish
successfully before the being may act.

The emerging invariant is:

> Many efficient causes feed one stable material/formal container. The
> container searches for one positively supported final cause. Ordinary
> cognition may revise that final, but only an exceptional increase in surprise
> may revise what the situation itself is understood to be.
