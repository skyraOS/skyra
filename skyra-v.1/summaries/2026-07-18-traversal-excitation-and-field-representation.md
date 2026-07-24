# Skyra — Traversal Excitation and Field Representation

**Date:** 2026-07-18
**Scope:** Current code, the still-valid foundations from the 2026-07-16 project snapshot, and the
new traversal/field direction developed through 2026-07-18.
**Status:** Project snapshot. Traversal-as-excitation, interference-pattern regions, recurrent carry,
and compression-based field representation are working mechanisms, not implemented architecture.
The traversal base case and collapse boundary are explicitly unsettled.

## Executive summary

Skyra high level is an distributed system abstraction over an individuals devices. The aim is to give an agent a world to inhabit where the world is your devices.

The 3 current components are:
Skyra-cognitive-agent
Skyra-distributed-agent
skyra-kernel

This summery will mostly focus on Skyra-cognitive-agent

Skyra remains a graph-based embodied learning system whose durable truth comes only from
observation. The LLM is still moving away from the center of the runtime and toward the unresolved
edge, where it may serve as a general collapse function until stable patterns compile into parsers
or reflexes.

The new direction changes the proposed mechanics of cognition:

```text
old working picture:
    current state -> planned bridge -> target state

new working picture:
    interaction + carried structures
        -> one traversal
        -> interference patterns
        -> field representation
        -> collapse/carry + expression
        -> returned observation
        -> next traversal
```

Excitation may not be a separately maintained cognitive variable. A cell is excited insofar as it
participates in the currently realized traversal. Currently excited observations and structures
form the root set of one traversal. As the kernel builds the tree, cells inspect the
traversal-so-far, record their occurrences on the Slice, and expose traversal-local frontiers.
Repeated occurrence of the same durable cell through several paths creates structural convergence.

The completed traversal is the active field:

> **The active field is the local relational workspace in which current state, target state, and
> the presently lit bridge between them coexist.**

That sentence remains useful, but current and target may not be separate objects. They may be roles
inside one active relational pattern:

```text
observed role       what the interaction/world currently establishes
directional role    what the interaction requests, requires, or leaves unresolved
behavioral role     expressions associated with transforming the pattern
```

A relational region may likewise not begin as a durable container. It may first be an interference
pattern created where traversal paths overlap. A recurring pattern can later compress into a
durable region, and stable region behavior can compile into a parser or reflex.

The representation of the field has therefore become central. It must supply Skyra's analogue of
phase: the relational combination rule that determines whether overlapping path contributions
reinforce, compose, conflict, remain independent, or remain unresolved. The kernel records paths
and convergence; it does not interpret them.

The immediate research task is to understand compression well enough to represent these overlap
patterns without erasing provenance, contradiction, unresolved direction, or observable closure.

## Foundations retained from the previous direction

Several commitments from the 2026-07-16 snapshot remain load-bearing.

### Observation is the only writer of durable world state

An expression, plan, model response, or traversal result cannot establish that the world changed.
Only a returned observation may do so.

```text
expression attempts a transition
    -> world returns an observation
    -> observation changes durable evidence
```

This remains the primary epistemic boundary.

### Unknown is not false

Missing evidence may require perception, waiting, or interpretation. It may not be treated as a
failed claim. Any future field representation must preserve:

```text
true | false | unknown
```

or an equivalent distinction.

### Raw observations are durable; interpretation is revisable

An observation's content remains immutable. Its interpretation, significance, region membership,
and role in a traversal may change.

```text
CONTENT
    what the durable observation recorded

SIGNIFICANCE
    what role an occurrence of that observation plays in traversal T
```

This makes reparsing safe: an interpretation can be destroyed and rebuilt without destroying the
evidence that crossed the boundary.

### Exact first; inference sees the unresolved residual

World-addressable identity, exact bindings, known structure, and deterministic relations should be
resolved before an LLM is recruited. As patterns stabilize, more of the field should migrate from
model inference into traversal, compressed regions, and compiled parsers.

### Learning remains across the observation boundary

The durable learning chain remains:

```text
observation -> state change -> action -> observation
```

Its relationships remain conceptually distinct:

```text
P(state change | observation)      interpretation
P(action | state change)           policy
P(next observation | action)       consequence
```

The new field model changes how these relations become active and combine. It does not permit a
model to grade its own expression.

### The kernel remains ignorant of cognition

The kernel owns traversal mechanics:

- roots;
- ordering;
- admission and budget;
- cycle prevention;
- tree construction;
- termination;
- the fold.

Cells and the field representation own relational meaning, compatibility, frontier behavior, and
the interpretation of convergence.

## Current implemented code

### Shared substrate

`src/skyra/spacetime/space.go` currently defines:

- `Space`, the universal traversable cell;
- `Slice`, per-being traversal records, maps, tree, result, and proper time;
- `Time`, the wave-realization interface;
- `Impulse`, the event envelope;
- edge-keyed activation records.

`src/skyra/spacetime/relation/relation.go` implements one rooted depth-first wave with:

- path-local cycle prevention;
- amplitude attenuation and branch conservation;
- thresholded admission;
- brief Slice locks;
- post-order `Express`;
- tree and result recording.

The present kernel keeps the becoming traversal stack locally and writes the completed root tree at
the end. Cells cannot yet inspect a Slice-resident tree as it is being constructed.

The shared `Space` contract is currently:

```go
Observe(origin Space, impulse any)
Children() map[string]Space
```

`Observe` returns nothing, and the kernel reads `Children()` afterward. This encourages cells to
communicate a per-wave frontier by mutating hidden child state.

The desired conceptual contract is closer to:

```text
Observe(traversal context) -> traversal-local frontier
```

The exact Go signature remains open.

The standalone `src/skyra/spacetime/traverse` package already has a `Frontier`-like contract and budgeted
subgraph assembly. It is mechanically closer to the new direction, but it uses a global
visited-once set that would currently suppress convergence through a second root/path.

### DNA v2

DNA v2 remains the most integrated implementation of:

- sensory intake;
- episodic and causal memory;
- static workflow routing;
- concurrent prediction branches;
- action collapse;
- edge activation;
- correction retries;
- bounded graph traversal.

Its useful substrate mechanics remain relevant. Its fixed, inference-heavy cognitive pipeline and
model-authored causal grading are not the current direction.

### DNA v3

DNA v3 remains the implemented falsifiable-learning experiment:

- observations and expectations share a relational `Triple`;
- only observations update world state;
- expectations carry deadlines and actively obtain settlement evidence;
- exact comparison grades outcomes;
- reusable action-to-outcome edges learn counted probabilities;
- precedent reduces passes-to-closure in a sandboxed filesystem world.

V3 proves that independently observed consequences can replace an LLM judge. Its present split
between observation and expression stores, brute-force planning, explicit target loop, and
single-call floor are experimental arrangements rather than the mature field architecture.

## Traversal is the excitation mechanism

The working principle is:

> **A cell is excited insofar as it participates in the currently realized traversal.**

Excitation is structural participation, not necessarily a free-standing scalar assigned before the
wave.

### One traversal, many roots

Every excited observation does not get an independent traversal. The currently excited
observations and carried structures form the root set of one traversal:

```text
excited roots
  O1  O2  O3  carried-pattern-P
   \   |   |       /
       one traversal
            |
            v
     one active morphology
```

The kernel may represent this as a forest or a synthetic root. The synthetic root is a mechanical
device, not another cognitive object.

### Observe interacts with the traversal-so-far

When the kernel visits a cell, the cell checks whether that same durable cell already occurs in the
tree currently being built.

```text
kernel visits cell
    -> cell reads traversal-so-far
    -> records its current-tree occurrence on the Slice
    -> changes traversal-local behavior if appropriate
    -> returns its onward frontier
    -> kernel extends the tree
```

"Exists elsewhere" here means occurrence inside the current traversal tree. It does not mean a
semantic search for a similar durable observation.

Keep separate:

```text
TRAVERSAL CONVERGENCE
    the same durable cell is reached through several paths in this traversal

REPRESENTATIONAL MATCHING
    different structures are compressed or interpreted as equivalent
```

The first is a kernel/Slice mechanism. The second belongs to the field representation.

### Tallies before theory

The smallest Slice record may be:

```text
observation-O1      1
state-change-S4     2
action-A7           3
observation-O9      1
```

This records only that `action-A7` occurred through three paths in the current tree. It does not
prematurely mean confidence, relevance, conjunction, or priority.

If an experiment requires it, occurrences may later retain:

```text
root
path
tick
relational role
```

Start with the tally and preserve more only when information loss is demonstrated.

### Convergence must survive tree mechanics

Several tree appearances may refer to one durable cell:

```text
O1 -> A
O2 -> A
```

The Slice must retain that two paths converged on `A`. A global visited-once rail would erase the
second path. Cycle prevention should therefore remain path-local:

```text
do not revisit the same cell on one active path
```

while allowing the same cell to appear through other roots or paths.

## Context is reconstructed, not appended

The root set is analogous to a context window, but the resulting context is not:

```text
Context(t+1) = Context(t) + Observation(t+1)
```

It is reconstructed:

```text
Roots(t+1) =
    Excite(previous field, new interaction, carried patterns)

Field(t+1) =
    Traverse(Roots(t+1))
```

A new interaction may completely change the significance of an old observation:

```text
O0@T0 -> confirmation
O0@T1 -> contradiction
O0@T2 -> background
```

The durable content of `O0` remains unchanged. Its current relational role changes because it now
overlaps with different paths.

Therefore:

> **A context is not a collection of observations. It is the relational morphology produced when
> currently excited observations traverse together.**

The eventual model-facing context is a budgeted rendering of that morphology, not the morphology
itself.

## The field representation must supply the phase analogue

The interference analogy is useful as a source of mechanics, not as a claim that Skyra is quantum:

```text
QM-like idea                  Skyra mechanism

multiple paths                traversal paths from several roots
superposition                 several candidate structures remain live
constructive interference     compatible paths reinforce shared structure
destructive interference      incompatible paths weaken, block, or split structure
interference pattern          resulting activation morphology across the field
measurement/collapse          budgeted selection and committed expression
decoherence/decay             old traversal contributions leave the active field
```

An occurrence count supplies magnitude-like information. It does not say whether two contributions
are compatible.

The field/region representation must answer:

> **What happens when two traversal paths try to bind incompatible values into the same relational
> position?**

Conceptually:

```text
Combine(A, B)
    -> reinforce | compose | conflict | independent | unresolved
```

Possible relational behavior:

```text
same binding        -> reinforce
compatible bindings -> compose
opposing bindings   -> conflict or split
unrelated bindings  -> remain independent
unknown binding     -> remain unresolved
```

This combination rule is Skyra's analogue of phase. It belongs to the field representation, not
the void kernel. No decision has been made that it requires complex amplitudes, triples, signed
weights, or any other particular syntax.

## Relational regions emerge from interference

The behavior belongs to the resulting overlap pattern, not necessarily to an individual path:

```text
path from O0 --+
               +--> overlap pattern --> region --> behavior
path from O1 --+
```

A region may first be transient:

```text
REGION OCCURRENCE
    one interference pattern in one traversal
```

Repeated structurally and behaviorally similar occurrences may compress into:

```text
DURABLE REGION
    a reusable representation of the recurring pattern
```

This reverses the earlier assumption that durable regions already exist and are merely activated.
Some regions may be retrieved from prior compression; genuinely novel ones may first appear only as
field patterns.

## Recurrent traversal and carry

The emerging recurrence is:

```text
incoming interaction It
+ structures carried from T(t-1)
            |
            v
        root set Rt
            |
            v
        traversal Tt
            |
            v
   interference patterns Ft
            |
            v
       collapse Ct
        /          \
       v            v
carry into T(t+1)  expression At
                         |
                         v
                   observation I(t+1)
```

At cold start:

```text
roots = incoming interaction
```

On later turns:

```text
roots = new interaction + compressed structures carried from the previous traversal
```

Without collapse and a carry budget, recursively retaining every path would make the field grow
without bound.

## Significance delta as episode segmentation

The difference between successive traversals may provide an endogenous segmentation signal.

```text
same overlap patterns retain similar significance
    -> segment may continue

patterns reorganize, split, reverse role, or stop carrying
    -> candidate segment boundary
```

The segmentation signal is not just node-set difference. It includes changes in:

- which paths converge;
- the roles observations play;
- region birth, split, and disappearance;
- conflict versus reinforcement;
- what collapse carries forward.

However:

```text
pattern disappeared != interaction resolved
```

A pattern may disappear because it:

- decayed;
- lost the context budget;
- was compressed away;
- was suppressed;
- was superseded;
- or was observably resolved.

Only the last provides closure evidence.

## Current and target are becoming derived roles

The active field may not require separate `CurrentState` and `TargetState` objects.

```text
active relational pattern
    observed role
        what observations establish

    directional role
        what an interaction requests, requires, or leaves unresolved

    behavioral role
        expressions associated with transforming the pattern
```

A human request already contains a directional relation. Copying that relation into a separate
target store may not add mechanism.

Likewise, current state may simply be the observed side of the active pattern rather than a
separate maintained object.

The current/target language remains useful as a projection:

```text
current = observed role
target  = directional role
```

But it may not deserve two substrate-level node types, stores, or traversal origins.

## Removing target does not remove the objective

If target is no longer explicit, Skyra must still answer:

- Why was one action selected?
- What observable return counts as success?
- When does the interaction stop recruiting expression?
- How is an expression graded?
- How is resolution distinguished from decay?
- What prevents collapse from dropping unresolved direction?

The objective may move into the relational pattern and collapse function, but it cannot disappear.

The field representation must distinguish:

```text
RESOLVED
    returned observation established resolution

DROPPED
    budget, decay, or compression removed active participation

UNRESOLVED
    direction remains but is not currently selected

SUPERSEDED
    later observation revised or canceled the direction
```

Only `RESOLVED` supplies closure evidence.

Closure duration and setpoint behavior from the previous direction may still fit, but they must be
restated over recurring patterns rather than assumed target objects:

```text
pattern becomes observably resolved
    -> closure begins

relevant observation recreates the unresolved pattern
    -> closure ends
```

This preserves the slow learning question:

> Which expression causes this relational tension to remain observably resolved longest?

## Collapse remains open

Two operations may be hiding under the term:

### Representational collapse

```text
large traversal field
    -> compressed structures fitting the context budget
```

### Behavioral collapse

```text
live candidate behaviors
    -> committed expression/action
```

The field representation is expected to collapse or compile most behavioral ambiguity once a
pattern becomes stable. A deterministic parser may produce an action directly. An LLM may handle
novel or unresolved fields.

It remains open whether these are:

- two stages;
- one combined function;
- or different implementations of the same interface.

Neither collapse may write world state. It may carry structures or emit expression; returned
observation remains the only durable evidence of consequence.

## Express should select structures, not isolated top nodes

The model-facing context has a finite budget, but independently selecting the highest-scoring nodes
could destroy the relationships that make them meaningful.

Express or representational collapse should select complete structures such as:

- an overlap pattern;
- a compressed region with current bindings;
- a contradiction;
- a complete supporting path;
- an unresolved residual;
- a candidate expression and the observations that graded it.

Repeated convergence should compress:

```text
action A reached through ten compatible paths
```

into one represented structure plus its support, rather than consume the context budget by
rendering `A` ten times.

The eventual objective is:

> Preserve the smallest field representation sufficient to make the next grounded decision.

## Why compression is now central

A multi-root traversal can grow as:

```text
roots
    x historical paths
    x repeated shared cells
    x overlapping patterns
```

Compression must keep the field finite while preserving:

- durable evidence pointers;
- repeated structure;
- current bindings;
- path convergence;
- contradiction;
- unresolved residuals;
- action-relevant distinctions;
- returned observations used for grading.

Three levels are relevant:

### Structural deduplication

```text
same durable cell reached ten times
    -> one identity + occurrence tally
```

This can be lossless.

### Template or grammar compression

Repeated subgraphs become parameterized patterns:

```text
Michael requests water
Kathy requests juice

becomes:

requester requests resource
```

with occurrence-specific bindings.

### Behavioral compression

Differences may be discarded only while they do not change prediction, expression, or observable
resolution.

```text
irrelevant distinction
    -> may compress away

distinction changes which expression produces durable resolution
    -> must remain a slot or cause a split
```

Confident but erratic resolution is evidence that a compression pooled behaviorally different
cases.

The strongest current candidate definition is:

> **A relational region is a compressed representation of traversal structure that preserves the
> distinctions necessary to combine paths, predict observations, select expressions, and evaluate
> resolution.**

Minimum Description Length, grammar compression, sufficient statistics, and predictive compression
are relevant bodies of ideas, but no encoding has been selected.

## The base case is explicitly unsettled

Do not yet define the traversal base case.

Prior candidates assumed:

- reaching an explicit target;
- satisfying a target predicate;
- meeting current morphology during backward planning;
- or reaching a resolved field pattern.

The move from separate current/target objects to observed and directional roles inside one pattern
changes the stopping question.

The future base case must at least distinguish:

```text
resolved by observation
unresolved but actionable
unresolved and requiring perception or inference
conflicting
dropped or decayed without resolution
```

The field/compression representation must make these states expressible before the base case can be
chosen.

## Current architectural picture

```text
1. An interaction enters the relational field.

2. Compressed structures carried from the prior traversal may join it as roots.

3. The kernel performs one multi-root traversal.

4. Each cell reads the traversal-so-far, records its occurrence, and exposes a frontier.

5. Multiple paths converge, compose, or conflict.

6. Their interference pattern is the active field.

7. Novel patterns are transient region occurrences; repeated ones may compress into durable regions.

8. Collapse selects structures under budget, determines carry, and may produce expression.

9. Expression acts without writing world state.

10. Returned observation introduces the next interaction.

11. Significance delta across traversals supplies a candidate episode boundary.

12. Stable field-to-expression behavior may compile into a parser/reflex.
```

## Open questions

### Root admission and decay

- What makes an observation or carried pattern eligible as a root?
- What decays: root eligibility, occurrence contribution, edge activation, or all three?
- Does target/directional tension affect admission without becoming a separate target object?

### Field representation

- What is the concrete relational position into which path contributions bind?
- What supplies the reinforce/compose/conflict/independent/unresolved combination rule?
- How are AND support, OR support, and redundant evidence distinguished?
- Can compression preserve contradiction and provenance without retaining the whole raw tree?

### Traversal mechanics

- How does a cell inspect the tree while it is being built?
- Is an occurrence tally enough?
- How does repeated occurrence affect frontier without traversal-order dependence?
- What deterministic ordering or commutative merge is required?

### Collapse and expression

- Does field compression itself resolve most behavioral competition?
- Is the LLM one collapse implementation and a parser another?
- Who chooses the carry set?
- Does Express commit one action, one next step, or a runnable partial order?
- How is an unresolved but unselected pattern kept from disappearing?

### Segmentation and learning

- What significance delta constitutes a new episode?
- How is disappearance distinguished from resolution?
- How is closure duration measured over recurring patterns?
- When does a transient interference pattern earn a durable region?
- When does a stable region earn a compiled parser?

### Base case

- Intentionally unsettled until the field representation can express resolution, actionability,
  uncertainty, conflict, decay, and supersession.

## Recommended next conceptual work

Before implementation, construct several small traversal examples by hand.

For each example:

1. Choose two or three interaction roots.
2. Draw the paths the kernel would build.
3. Mark repeated durable cells.
4. Record occurrence tallies.
5. Identify compatible and incompatible bindings.
6. Draw the resulting interference pattern.
7. Decide what structure should carry into the next traversal.
8. Add a new interaction and observe how prior significance changes.
9. Identify what was resolved versus merely dropped.
10. Compress repeated patterns into a candidate region.

Use examples where:

- paths reinforce;
- paths contradict;
- two roots jointly afford behavior neither affords alone;
- a new observation reverses an old observation's significance;
- a pattern disappears through decay rather than resolution;
- a recurring pattern deserves compression;
- a compressed pattern later fails because it discarded a behaviorally important distinction.

This should expose the requirements for the field representation before triples, templates,
compressed graphs, or another syntax are selected.

## Bottom line

Skyra's current direction is a recurrent relational field:

```text
interaction perturbs carried structure
    -> traversal produces excitation
    -> overlapping paths create field patterns
    -> field representation combines and compresses them
    -> collapse carries structure and emits expression
    -> returned observation reconstructs the field
```

The durable substrate remains grounded in observation. The active context is reconstructed through
traversal. Relational regions emerge as recurring interference patterns. Compression turns those
patterns into bounded reusable structure. Parsers may eventually collapse stable structures
directly into behavior, leaving the LLM for novel or unresolved fields.

The important discipline is unchanged: expression may act, but only observation may establish what
happened. The new work is to design a field representation capable of preserving overlap,
contradiction, unresolved direction, and resolution strongly enough that traversal, segmentation,
collapse, and learning can all be defined over it.
