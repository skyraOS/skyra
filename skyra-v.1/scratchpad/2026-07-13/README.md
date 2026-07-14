# Multiple Potential Wells

**Date:** 2026-07-13  
**Status:** SCRATCHPAD — raw working thought, not yet a design or implementation commitment

```text
Multiple potential wells

↓

Superposed field

↓

Gradient

↓

Motion
```

## Durable Graph And Active Scratchpad

The durable graph stores long-lived attractors, memories, and learned structure.

The blackboard/scratchpad stores active discrepancies, unresolved tensions, and current estimates.

The LLM proposes ways to reduce those discrepancies.

The substrate decides which discrepancies become important enough to recruit new attractors or actions.

## Open Question: Local Downhill Score

Given a current state and a target, what computation assigns every neighboring edge a local downhill score?

## Reverse Traversal From Active Final Causes

Suppose the gradient is produced by traversal that starts from the active final causes as children and traverses backward into their currently active subgraph: the current state.

The traversed subgraph represents the final state — all active goals — reaching back toward the current state.

```text
active final causes / goals
        ↓ reverse traversal
causal paths capable of reaching them
        ↓
currently active subgraph
        ↓
local edge scores available to motor
```

The traversal constructs the potential field. Differences between neighboring node potentials form the local gradient.

For every active final cause `G`, begin with its current amplitude:

```text
potential(G) = goal amplitude
```

Propagate that value backward through relationships that could lead to it:

```text
contribution(path to G) =
    goalAmplitude
    × causalEdgeStrengths
    × decay(distance)
```

With several active final causes, their fields superpose:

```text
potential(node) =
    Σ contribution(node → goalᵢ)
```

At a current node `S`, an available edge to neighboring node `N` receives a local directional score:

```text
directionScore(S → N) =
    potential(N) - potential(S) - transitionCost(S → N)
```

With potential interpreted as desirability, a positive score moves toward a stronger goal basin. If represented as energy instead, negate the potential and the same motion becomes literally downhill.

```text
active final causes
        ↓
reverse causal traversal
        ↓
assign goal-conditioned potential to reachable nodes
        ↓
intersect with the current active subgraph
        ↓
compare neighboring potentials
        ↓
motion / action
```

The material/formal container remains fixed during ordinary operation. Active final causes project gradients backward through it, and motor follows locally available gradients. Ordinary changes in the field may change action without rewriting the container. Only an exceptional positive surprise delta may change the material/formal summary itself.

The system does not need to generate a complete route. It needs trustworthy causal edges and active final causes; reverse traversal turns those into local direction.

## Two Active Subgraphs: Current And Target

Maintain two active subgraphs:

```text
CURRENT — the state Skyra presently inhabits
TARGET  — the desired final state represented by all active goals
```

Each cognitive emission is itself a subgraph rather than only an action string.

Given the target state, traverse its reverse indexes looking for causal predecessors. If no usable final causes emerge, emit a new predecessor-state subgraph, `final-1`, together with a causal summary representing the relationship to the subgraph just created or traversed.

Then repeat:

```text
TARGET / final-n
        ↓ reverse-index or emit predecessor
final-(n-1)
        ↓
final-(n-2)
        ↓
...
        ↓
CURRENT
```

The process continues until the reverse traversal reaches or sufficiently overlaps the current-state subgraph. The stopping condition may need to be graded overlap rather than exact graph identity.

This is backward causal construction during descent:

```text
descent:
  target state
    -> retrieve or emit its prerequisite state
    -> make the later state a final cause of the prerequisite
    -> continue until the current state is reached
```

At the current state, `Express` begins. Expression is Praxis: it follows the constructed bridge forward and executes.

```text
ascent:
  current state
    -> execute the locally available transition
    -> observe the realized result
    -> record the action/result as an efficient cause
    -> continue toward the target
```

The compact symmetry is:

```text
descent creates final causes
ascent creates efficient causes
```

The backward pass says what each prior state is oriented toward. The forward pass discovers what actually produced each later state.

## Praxis Error And Local Retraversal

The intended next state and realized result must remain separate.

For a planned transition:

```text
current state S(k)
    -> intended final cause S(k+1)
    -> emitted action
    -> observed result R(k+1)
```

After Praxis runs, grade the emitted action and observed result against the intended final cause:

```text
grade(R(k+1) | intended S(k+1))
```

If the grade is above threshold, the transition closes without error. The realized action/result becomes efficient-cause evidence for the reached state.

If the grade is below threshold:

1. Preserve the desired end-state subgraph.
2. Preserve the failed result as observed current material; do not overwrite it with the intended state.
3. Record the bad grade and a reason/discrepancy describing the miss.
4. Reverse-traverse again from the desired end state toward the failed result or the predecessor immediately before the failed transition.
5. Rebuild only the affected bridge where possible.
6. Run Praxis again from that point with the failure feedback present in the active scratchpad.

```text
desired end state
        ↓ reverse traversal
revised predecessor chain
        ↓
failed observed state + reason for failure
        ↓ Express / Praxis
new action
```

This is local replanning rather than discarding the target or rebuilding all learned structure after every mistake.

## What A Repeated Traversal Should Strengthen

A node reached by two traversals has become more important, but failure should not automatically make the failed route more desirable.

Keep at least three signals distinct:

```text
familiarity / frequency   repeated traversal happened
alignment / success       the transition moved toward its intended final
surprise / uncertainty    the expected transition failed or remains unresolved
```

Repeated failure may increase salience, attention, or recruitment pressure while decreasing the failed edge's positive alignment. This makes the unresolved region a stronger attractor for cognition without making the same failed action a stronger attractor for motor.

On a successful retry, the new transition earns efficient-cause strength. The old failed transition remains as negative or contradictory evidence rather than being deleted.

## Bounds And Unresolved Bridges

If reverse indexes do not reach the current state, emitted predecessor subgraphs must be bounded by depth, inference budget, and minimum progress. Failure to bridge the two active subgraphs is a legitimate unresolved state. The system must not invent an arbitrarily long causal chain merely because traversal is required to close.

## Impulse-Keyed Workflow, End-State-Keyed Field

Workflow should remain keyed by impulse. The board answers transactional questions:

```text
which impulse is being processed?
what work did it recruit?
what snapshot did it read?
what completed, failed, or remains in flight?
what prior impulse/action caused it?
```

The main cognitive field should be keyed by the active end state, not by the impulse.

The end state may not belong on the board at all. It is an actively maintained target subgraph on the being's Slice or working-memory substrate. Many impulses may arrive, register, and complete while the same end-state subgraph remains active.

```text
Board
  impulse X -> transaction against end-state version 4
  impulse Y -> transaction against end-state version 4
  impulse Z -> transaction against end-state version 5

Slice / active field
  current-state subgraph
  active end-state subgraph, version 5
  reverse-traversed potential field joining them
```

An impulse does not own the goal. It may provide evidence that reinforces, weakens, satisfies, contradicts, or proposes changing the actively maintained end state.

## End-State Versioning

Treat the active end state as versioned state:

```text
ActiveEndState
  ID
  Version
  Root causal-node IDs
  Active goal amplitudes
  ChangedAt
  Cause / triggering impulse
```

Each impulse row records the end-state version it observed. This makes stale work detectable without copying the end state into every board row.

```text
ImpulseRow
  ImpulseID
  PreviousImpulseID
  EndStateVersionRead
  CurrentStateVersionRead
  Work status / trace
```

## End-State Change Invalidates The Field

A committed change to the active end-state subgraph invalidates the goal-conditioned potential field and forces reverse traversal again:

```text
end-state subgraph changes
        ↓
increment end-state version
        ↓
invalidate prior potential/traversal snapshot
        ↓
reverse-traverse from the new active finals
        ↓
intersect/reconnect with current state
        ↓
publish a new field snapshot
```

An end-state change does not automatically justify changing the durable graph. First retraverse the graph that already exists. Only if the new target cannot form a sufficiently supported bridge to current state should cognition propose missing predecessor subgraphs or new causal relationships.

```text
change goal
  -> retraverse existing structure
  -> grade reachability / closure
  -> repair or grow graph only where the bridge is missing
```

This separates three events:

```text
impulse registration   transaction/lifecycle event on the board
end-state commit       active-target state change on the Slice
graph learning         durable topology change justified by traversal/evidence
```

## Concurrent End-State Changes

Praxis and inference operate against immutable field snapshots. If the end state changes while an action is being prepared, the action's snapshot version is compared with the active version before emission.

Possible outcomes:

- still aligned under the new end state: allow emission;
- stale but harmless or irreversible work already began: record it and grade the result;
- stale and no longer aligned: suppress/recompute before external emission.

The exact cancellation policy remains open, but version identity must survive into the action record so reafference is graded against the end state that actually justified the action.

## Amplitude As Possible End-State Participation

Instead of using amplitude as a representation of context-window size, amplitude may represent how strongly a possible end state participates in the active field.

```text
active final G1  amplitude .60
active final G2  amplitude .25
active final G3  amplitude .15
```

Each possible end state launches a reverse traversal with its amplitude. The amplitude is propagated through supported predecessor relationships and contributes to the superposed field over current state.

```text
amplitude(G) = current participation of possible end state G

fieldContribution(node, G) =
    amplitude(G)
    × causal path strength
    × traversal decay
```

Changing an end state's amplitude reshapes the field even when neither endpoint subgraph nor durable topology changes:

```text
possible end-state amplitudes change
        ↓
invalidate field snapshot
        ↓
reverse-traverse again
        ↓
new local action gradient
```

Context capacity should then be a separate quantity:

```text
amplitude  = how much a possible end state participates
budget     = how much traversal/context can be afforded
cost       = what a node or edge consumes from that budget
```

This avoids making a goal less important merely because its representation is expensive, and avoids treating a larger context window as stronger desire or likelihood.

Amplitude may not be identical to probability. Possible interpretations include:

- confidence that the end state is reachable;
- commitment or desire for the end state;
- current activation/salience of that end state;
- allocated participation mass in the active field.

These meanings should not be silently collapsed. A likely state is not necessarily desirable, and a desired state is not necessarily believed reachable.

A provisional decomposition could keep three values distinct:

```text
amplitude     nonnegative participation magnitude
polarity      attractive or repulsive orientation
confidence    estimated reachability / epistemic support
```

Whether all three are required remains open.

Mutually exclusive final states may normalize their amplitudes within one choice set. Compatible simultaneous goals should not necessarily compete for a fixed sum; they may each project a field and superpose.

## Working Memory As Maintained Endpoint Knowledge Graphs

At least one central part of working memory is the actively maintained pair:

```text
CURRENT STATE knowledge subgraph
DESIRED STATE knowledge subgraph
```

Every impulse tasks a memory worker with maintaining both endpoint graphs. The worker interprets the impulse as a delta to current state and, where warranted, as a proposal or update to desired state.

Example: while Skyra is in the middle of a large cognitive task, Michael says:

> Oh, and can you grab me a glass of water from the counter?

The memory worker updates current state with something like:

```text
michael --wants--> water
michael --requested--> skyra
skyra --obligated/proposed-to-fetch--> water
water --located-at--> counter
```

It proposes or activates a desired-state subgraph such as:

```text
michael --has--> water
```

or, if firmly warranted:

```text
michael --is-drinking--> water
```

The distinction matters: "Michael has water" follows more directly from "grab me a glass of water" than "Michael is drinking water." Several candidate desired states may be represented with different amplitudes rather than forcing one interpretation immediately.

After the endpoint update, reverse traversal runs again:

```text
impulse
  -> memory worker updates CURRENT
  -> memory worker proposes/reweights DESIRED
  -> commit new endpoint versions
  -> invalidate the prior field
  -> reverse-traverse from all active desired finals
  -> publish a new superposed field over current state
```

The new request does not necessarily replace the large cognitive task already in progress. It introduces another possible end state into the maintained target graph. Its amplitude determines how strongly it participates alongside existing goals.

```text
existing cognitive end state  amplitude A
water-delivery end state      amplitude B
        ↓
superposed reverse field
        ↓
local action gradient
```

The memory worker maintains the endpoint representations; it does not unilaterally decide action. The substrate decides whether the proposed desired-state change is admitted, how much amplitude it receives, whether it conflicts with existing finals, and whether the resulting field recruits action or additional cognition.

This gives working memory a concrete responsibility:

> Maintain an evidence-grounded representation of where the being is and a versioned, amplitude-weighted representation of where it is currently oriented to go.

## Observe Builds The Reverse Bridge Recursively

The reverse bridge can be built through repeated kernel `Observe` passes.

Start traversal at a target-state node:

```text
Observe(target)
  -> run the target's reverse index
  -> retrieve causal predecessor nodes
  -> expose those nodes as edge-keyed children to the kernel
```

The kernel selects and visits the exposed predecessor nodes. Their own `Observe` pass repeats the operation:

```text
Observe(final-n)
  -> retrieve known predecessors
  -> construct candidate predecessor subgraphs where needed
  -> summarize each candidate subgraph
  -> expose the candidates to the kernel
```

Rinse and repeat until a candidate predecessor state reaches or sufficiently overlaps the actively maintained current-state subgraph. At that boundary, descent is complete and ascent/`Express` begins.

```text
TARGET
  Observe -> predecessors
    Observe -> predecessors
      Observe -> predecessors
        ...
          CURRENT STATE reached
        Express
      Express
    Express
  Express
```

## Experienced And Counterfactual Candidate Subgraphs

When a known reverse index does not supply enough structure, an `Observe` pass may construct at least two kinds of candidate predecessor subgraph:

```text
experience-conditioned candidate
  includes relevant prior experience and learned causal structure

experience-independent candidate
  attempts the predecessor without relying on that prior experience
```

The first is exploitation: what previously learned structure says could lead here. The second is exploration or counterfactual generation: what could lead here if the previous experience is incomplete, misleading, or absent.

Each candidate is a complete summarized state subgraph:

```text
CandidateState
  summary / material-formal representation
  causal-node members
  relationship edges
  provenance
  amplitude / field participation
```

The candidates are placed into the wave's edge-keyed traversal map and handed to the kernel. The kernel remains content-agnostic: it receives candidate spaces and their edge identities, applies traversal mechanics, and records what was visited.

The map should be wave-local or working-memory state first. A generated candidate is a potential, not immediately learned durable topology. Praxis success and reafferent evidence may later crystallize it into the durable graph; failure leaves graded negative or unresolved history.

## Observe/Express Responsibilities

```text
Observe:
  move backward from desired effect toward possible causes
  retrieve known predecessors first
  generate missing predecessor subgraphs when necessary
  expose an edge-keyed candidate map

Express:
  fold the completed reverse bridge forward
  compare competing branches
  assemble the Praxis route
  record intended final relationships
  execute from current state toward target
```

External side effects should not occur independently in every branch's `Express`. Branch expression should first produce scores, proposals, and route structure. A single selected Praxis/motor boundary performs the externally visible action. Otherwise speculative branches could all act merely because post-order traversal visited them.

## Reaching Current State

"Reached current state" probably cannot require exact node identity. The target-side predecessor subgraph and current-state graph may describe the same situation at different granularity.

A provisional closure check may grade:

```text
overlap(candidate predecessor, CURRENT)
causal compatibility(candidate predecessor, CURRENT)
unexplained residual
```

Descent stops when overlap and compatibility cross their thresholds and residual falls below its threshold. It also stops unresolved on traversal budget, depth, visited-once, or no-progress limits.

The resulting traversal tree is the proposed causal bridge. Descent populated final-cause orientation; ascent/Praxis tests the bridge and supplies realized efficient-cause evidence.

## The Planner Is Familiar; The Substrate Is The Wager

This is not a claim to have invented planning. The model largely rederives backward planning plus forward action and error correction.

The architectural wager is that planning and action are forced to inhabit the same underlying substrate.

```text
current-state representation
  -> desired-state representation
  -> reverse causal bridge
  -> action proposal
  -> affordance invocation
  -> world result
  -> observed next state
  -> grade / evidence
```

These are not separate planner, controller, tool, and memory data models connected by ad hoc messages. They are subgraphs and causal records in one traversal grammar.

Action is therefore not a side channel leaving cognition. An action invocation and its result remain part of the same graph that justified the action. Reafference can directly compare:

```text
the current state that was read
the active final that oriented traversal
the causal bridge that supported the move
the action that was emitted
the state that reality returned
```

The same relationship used for planning can then be strengthened, contradicted, or left unresolved by action.

The claim that this is a better substrate remains an empirical hypothesis. A useful invariant is:

> Every externally emitted action must be traceable through the substrate to the current-state version, active end-state version, traversed causal path, intended transition, and observed result that graded it.

If one of those links exists only in an orchestration side channel, action has escaped the substrate and the unification is incomplete.
