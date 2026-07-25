# Skyra Substrate And Project Direction

**Date:** 2026-07-08
**Scope:** Whole project substrate, not DNA-specific.

This is a snapshot of what Skyra appears to be becoming after the v.05 lineage,
the v.1 runtime, and the current July reform tracks.

## Thesis

Skyra is a long-lived agentic substrate, not an agent pipeline.

Most agent systems put the model at the center and build orchestration around it:
memory retrieval, planning, tools, routing, output parsing, and logs. Skyra's
direction is different. The model is one affordance inside a living traversal
runtime. The center is the world substrate: spaces, relations, slices, activation,
memory topology, affordance crossings, prediction, surprise, and durable history.

The project is trying to answer:

```text
What is the smallest runtime physics that lets an AI being inhabit continuity?
```

Not just remember facts. Not just call tools. Inhabit a world where actions,
relationships, predictions, corrections, trust, tasks, memory, and identity
compound over time.

The current north star is that the user should feel the universe changing after
every conversation. The user is not merely an input source. The user is a
world-shaping participant inside the same durable universe.

## Lineage

The older v.05 architecture treated everything as recursive Reality traversal.
A Relation descended through worlds, beings, exchanges, tools, memory, and
providers. The product shape was creature-like: many beings, recursive worlds,
relationships, and ports. It already resembled normal agentic architecture, but
its modules were expressed as topology rather than as a conventional stack.

v.05 discovered the important product ontology:

```text
being
relationship
world
thread
exchange
memory
tool
provider
recursive descent/ascent
```

Its strength was that present context emerged from the path the Relation traveled.
The descent built the present. The ascent integrated the result. Tools, beings,
threads, and providers were all realities in one recursive structure.

The pressure that broke v.05 was memory. A traversal path can build a present, but
a long-lived being needs to know what durable structure that present leaves
behind. It needs a topology that can grow, decay, reframe, correct itself, and
remain inspectable.

v.1 is the answer to that pressure. It preserves the v.05 intuition, but moves
down a layer:

```text
v.05:
  recursive beings/worlds/products

v.1:
  substrate physics for memory, activation, prediction, action, and result
```

The evolution is not a rejection of v.05. It is v.05's creature world being
refactored into a cellular substrate.

## Current Substrate

The current runtime is organized around a few recurring concepts.

### Space

`Space` is the inhabitant/interface. A space may be a world, point, being,
memory cell, worker, or affordance surface. The current interface still has some
legacy coupling, but the target meaning is stable:

```text
Observe = expose what the kernel can traverse next
Express = commit or return what happens after traversal
```

Durable graph nodes such as `MemoryCell` have durable children. Ephemeral workers
should expose per-wave traversal surfaces. The July reform separates these two
ideas so hidden mutable `Children()` state does not become the universal kernel
API.

### Time / Kernel

The kernel is transport, not cognition.

It should own:

```text
- one wave starts with a budget/amplitude
- one wave records on one slice
- call Observe
- traverse returned edge-keyed spaces
- enforce visited-once and affordability
- call Express post-order
- record the traversal tree and traversed edges
```

It should not own:

```text
- what a memory means
- what relevance means
- what probability means
- how trust, alignment, identity, or stress are computed
- why a worker exposed one child instead of another
```

That boundary is important. If the kernel stays small, the system can support many
kinds of beings without baking one cognitive theory into traversal mechanics.

### Spacetime

The current reform introduces `Spacetime` as the concrete event of a wave:

```go
type Spacetime struct {
    Origin    Space
    Traversed Space
    Impulse   Impulse
    Amplitude float64
}
```

This is a clean split:

```text
Impulse   = domain/event envelope
Spacetime = kernel wave-start metadata
```

Amplitude belongs to the wave, not the impulse. A message, perception, action
result, or cognitive affordance invocation is an event. The amount of traversal
energy given to that event is a property of the wave moving through the substrate.

### Slice

The Slice is durable per-being runtime state.

It is not just scratch space. It is the being's local memory manager and working
set: activation records, traversal history, active grouping state, thresholds,
stress-like signals, and indexes. The same world can be interpreted differently
by different beings because each being has its own slice.

This is one of Skyra's core architectural commitments:

```text
world topology may be shared
experience is per-being
```

### MemoryCell

`MemoryCell` is becoming the one durable graph node / record.

The system should not proliferate separate durable structs for every cognitive
role. Instead:

```text
MemoryCell.ID
MemoryCell.Kind
MemoryCell.Content
MemoryCell.Children
metadata / typed payload
```

differentiate gist, process, invocation, result, entity, prediction, reminder,
and other durable records.

This keeps the graph inspectable and allows many roles to share one traversal
grammar.

### Activation

Activation is edge-owned and being-relative.

The important move is that a node does not globally "have activation." A being's
slice owns activation records for edges and traversal history. The same memory can
be loud for one being and quiet for another.

Traversal should allocate amplitude proportionally across open edge products,
not split energy equally. Stronger active edges should receive more of the wave.
Weak edges are not deleted, but they should not receive the same transport budget
as stronger edges.

The emerging rule:

```text
split:
  parent amplitude divides across open children by normalized edge product

join:
  future convergence should preserve that multiple paths reached the same node
```

The current stable first implementation can keep visited-once traversal and
record blocked convergence as metadata.

## Affordance Grammar

The July reform is converging on one durable graph grammar for both world actions
and cognitive actions:

```text
gist
  -> proposal/process record
      -> affordance invocation record
          -> affordance result record
              -> entities/spaces
```

This is the most important current direction.

An invocation is the attempt. A result is what reality returned. The distinction
matters because intended target and confirmed world content are not the same
thing.

Examples:

```text
perceive invocation:
  Skyra tries to perceive the world

perceive result:
  the world reports Skyra and Michael at coordinates

engage invocation:
  Skyra tries to send an utterance to Michael

engage result:
  the world reports delivered, blocked, heard, or otherwise observed
```

Entities and spaces attach to results because results confirm world content.
Invocations carry route metadata because commands need direction.

This grammar should apply to:

```text
world affordances:
  perceive, move, engage

cognitive affordances:
  midlevel:frame, other:predict, self:remind, motor:collapse, grade
```

That gives Skyra one record model for acting in the world and acting inside its
own cognition.

## Universal Landing Spine

Every incoming impulse needs somewhere durable to land before it is understood.

The target invariant:

```text
before appending invocation/result records:
  if no active grouping/gist exists:
    create blank grouping
    create blank gist
    mark them needs-framing
```

Then incoming evidence lands first:

```text
grouping{needs-framing}
  -> gist{text:"", needs-framing}
      -> process
          -> invocation
              -> result
                  -> entities/spaces
```

Midlevel then sharpens the blank gist.

This is a major correction. Cold start should not be a special case where events
live outside the graph until framing catches up. Novel evidence should land on a
minimal spine immediately, then be framed.

## Self And Other

The self/other split is being reframed.

The older shape:

```text
other:
  predicts next external message distribution
  grades later incoming impulse against that distribution

self:
  predicts Skyra's next action
  motor collapses that prediction into KindAction
  self grades the action against reafference
```

The emerging shape:

```text
other = expected state of the world
self  = remembered policy/reminder field for bridging present to expected state
motor = concrete affordance invocation
result = reality correction
```

This is cleaner than "self predicts action, other predicts user." Other becomes
the predictive model of everything not-self: user state, conversation state,
world state, tool state, project trajectory, likely next pressure. Self becomes
the skill/reminder field conditioned on that expected world state.

One good grammar:

```text
present gist
  -> other/world expected state
      -> self reminders conditioned on that state
          -> motor collapse
              -> affordance invocation
                  -> affordance result
```

Self reminders are not commands. They are recalled policies or skill hints:

```text
under this gist,
given this expected world state,
remember this way of acting
```

That lets motor produce concrete actions without forcing self to be a brittle
one-shot instruction generator.

## Null Graph Bootstrapping

A null graph should not ask self to act first.

The correct cold-start sequence is:

```text
incoming impulse
  -> grade prior other/world expectation
      -> no prior
  -> deposit evidence under active/default gist
  -> midlevel sharpens gist
  -> other/world creates expected-state record from this evidence
  -> self attaches reminders conditioned on expected state
  -> motor may collapse an invocation
```

No-prior is not the same as contradicted. It is novelty, not failure.

The first incoming impulse becomes seed evidence for an implicit other/world
process. Other creates the first expected-state record. Self then has a world
state to condition on instead of inventing behavior from nothing.

Target cold-start shape:

```text
gist
  -> otherProcess{implicit|seeded}
      -> otherExpectedState{source: incoming impulse}
      -> gradeResult{kind: no-prior}

  -> selfProcess{seeded}
      -> selfReminder{conditionedOn: otherExpectedState}
      -> selfReminder{conditionedOn: otherExpectedState}
```

On the next impulse, the previous expected state can be graded as held, partial,
or contradicted.

## Scheduler / Board Direction

The current `dnav2` orchestrator and gene conditions are a useful bridge, but the
target design wants scheduling separated from cognition.

The scheduler should own workflow decisions:

```text
which wave runs next?
which process is open?
what work item is pending?
what happens after this Express result?
```

Spaces and workers should own local cognition:

```text
what evidence do I expose?
what process/result do I write?
what next Spacetime values do I emit?
```

This prevents the scheduler from becoming the old heavy orchestrator. It should
coordinate waves and board state, not encode self/other/memory logic.

## Directional Assessment

The project is directionally coherent.

The repeated pattern across the history is:

```text
large explicit architecture
  -> pressure discovers a smaller invariant
  -> code collapses toward that invariant
  -> product ambition survives at a lower substrate level
```

v.05 had the product world: beings, relationships, ports, recursive reality, many
creatures. v.1 is discovering the durable physics that makes those creatures
long-lived rather than theatrical.

The current reforms are especially good because they reduce special cases:

```text
one durable node type
one affordance invocation/result grammar
one landing spine
one wave envelope
one edge-keyed traversal surface
one self/other process grammar
```

This is how the system can become lighter while becoming more capable. It is not
less architecture. It is architecture moving from named modules into reusable
physics.

## Main Risks

The main risk is descending forever.

There is always a deeper substrate: graph, edge, wave, amplitude, scheduler,
process, expectation, reminder, memory, identity. The way to avoid endless descent
is to protect the current invariants and implement them before inventing another
layer.

The highest-value next implementation targets are:

```text
1. introduce Spacetime and move Amplitude off Impulse
2. make Observe return edge-keyed traversal surfaces
3. preserve MemoryCell as the single durable record type
4. implement the universal landing spine for all incoming impulses
5. migrate self/other toward process + expected-state + reminder records
6. write invocation/result chains while maintaining legacy compatibility
7. keep Scheduler/Board thin
```

The other risk is making `other` too narrow. If other only means "the user," the
architecture collapses back toward chatbot prediction. Other should mean expected
state of the world: user, environment, conversation, tools, project, social field,
and likely next pressure.

The final risk is making self too commanding. Self should not immediately become
"the action planner." It is better as a reminder/skill field that accumulates
ways of bridging present state to expected world state. Motor can collapse those
pressures into concrete affordance invocations.

## Short Form

Skyra is trying to build the ideal long-lived agentic architecture by refusing to
put the model at the center.

The center is a traversal substrate where:

```text
world events land durably,
gists frame them,
other/world expectation predicts where reality is going,
self reminders recall how to move under that expectation,
motor invokes affordances,
results correct the graph,
activation changes what comes to mind next,
and the user's participation grows the universe over time.
```

That is the project.

