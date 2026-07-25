# Directed Affordance Invocation

**Date:** 2026-07-08 · **Status:** ACTIVE RUNTIME DRAFT / PARTIALLY SUPERSEDED
**Parents:** [affordance-record.md](../../design-docs/being/affordance-record.md),
[perceive-build-uncertainties.md](../2026-07-07/perceive-build-uncertainties.md),
[slice-based-dispatch.md](../2026-07-03/slice-based-dispatch.md)

> **Supersession notice (2026-07-12):** This document still owns affordance
> invocation/result topology and the observation/intervention distinction.
> Maxim-owned Evidence, separate Evidence-node ownership, and coupling to the
> old Schema model are superseded. Observed events now become observation-born
> `CausalNode`s and act as Evidence through weighted causal relationships, as
> defined by the
> [layered graph](./cause-and-effect/subgraph-and-causal-edge-ownership.md).

This spec pins the common data model for world affordances (`perceive`, `move`,
`engage`) and cognitive/DAG affordances (`midlevel:frame`, `other:predict`,
`self:predict`, `motor:collapse`). The abstraction is not "world action." The
abstraction is:

> a space exposes an affordance; invoking it is a directed crossing from one
> endpoint to another.

The current code partially implements this. `perceive` records are closest to the
target spine. `engage` and motor actions still use older shapes, which causes the
frame to split Skyra's own concrete actions away from the recent exchange.

---

## Core Rule

Every affordance invocation is a `MemoryCell` record with two required endpoint
fields in its metadata:

```text
affordanceRecord
  metadata:
    from
    to
```

The record's containing process owns the invocation. In the short term, some
records may still hang directly under the active gist during migration. The target
shape is:

```text
grouping
  -> gist
    -> proposalRecord
      -> affordanceRecord
        -> affordanceResultRecord
          -> entityOrSpace
```

The invocation record represents the act of trying. The result record represents
what came back or what actually happened. Entities and spaces are attached to the
result because the result is what confirms the content of the world.

Result entities route back to the episode:

```text
entityOrSpace -> grouping
```

This makes all concrete crossings graph-equivalent. Their payloads differ, but
their topology does not.

The proposal record is not just a one-shot prediction. It is the durable process
container for a possible trajectory. A proposal may create multiple concrete
affordance invocation records over time:

```text
gist
  -> selfProposalRecord
    -> perceiveRecord
      metadata: {from: skyra, to: world}
      -> perceiveResultRecord
        -> skyra
        -> world
    -> moveRecord
      metadata: {from: skyra, to: kitchen}
      -> moveResultRecord
        -> skyra
        -> kitchen
    -> engageRecord
      metadata: {from: skyra, to: michael}
      -> engageResultRecord
        -> skyra
        -> michael
```

Under the gist edge, the proposal means "this process belongs to the current
episode framing." Under the proposal edge, each affordance record means "this
concrete step was emitted by this process."

---

## Invocation Versus Result

Affordance invocation and affordance result are separate records.

```text
proposal/process record
  -> affordance invocation record
      -> affordance result record
          -> entities/spaces
```

The invocation record is the attempt:

```text
engageRecord:
  metadata:
    primitive: engage
    from: skyra
    to: michael
    payload: "hello"
    timestamp: t
    status: pending/resulted
```

The result record is the observed consequence:

```text
engageResultRecord:
  content:
    delivered: true
    heard: true
    timestamp: t+1
  children:
    skyra
    michael
```

This distinction keeps intended targets separate from confirmed world content.
Traversal should care primarily about the content/result graph, not about routing
metadata on the attempted invocation.

### Observation Versus Intervention

Affordance direction determines evidence eligibility, not a categorical Evidence
label:

```text
world -> Skyra impulse:
  observation; eligible to birth an observed CausalNode

Skyra -> world impulse:
  intervention; never Evidence by itself

world -> Skyra return/result:
  observed consequence; eligible to birth an observed CausalNode
```

An emitted invocation must not confirm the process that emitted it. Its returned
world result may later be observed. Praxis/Telos progress and reward/utility are
separate deferred mechanics.

Examples:

```text
perceiveRecord = Skyra tried to perceive world/current-place
perceiveResultRecord = the world returned surroundings
perceiveResultRecord -> skyra at coord
perceiveResultRecord -> michael at coord

engageRecord = Skyra tried to engage Michael with an utterance
engageResultRecord = the world reported delivered/blocked/heard
engageResultRecord -> skyra
engageResultRecord -> michael
```

The invocation may carry `from`/`to` as metadata because the command must know
where it is aimed. The durable traversal-relevant entity edges belong under the
result record because the result is what confirms the graph.

---

## Universal Landing Spine

Every incoming impulse needs an active grouping and active gist before durable
affordance records are appended.

This is a universal intake invariant, not a perceive-specific workaround.

```text
before appending invocation/result records:
  if active groupID is missing OR active gistID is missing:
    create blank grouping
    create blank gist
    mark gist/grouping as needs-framing
```

Target blank shape:

```text
grouping{status: needs-framing}
  -> gist{text: "", status: needs-framing}
```

Then the incoming impulse lands under that spine:

```text
grouping{status: needs-framing}
  -> gist{text: "", status: needs-framing}
    -> proposal/process record
      -> affordance invocation record
        -> affordance result record
          -> entities/spaces
```

Midlevel should be handed the existing blank spine with an explicit framing note:

```text
There is currently no active gist for this impulse. Please create/sharpen one.
```

Midlevel then edits/sharpens the blank gist instead of creating the first durable
spine after the fact.

Current code only does this consistently on the perceive-return ingest path via
`ensureDefaultGrouping`. External engage cold-start still records into the entity
index first and asks midlevel to create a grouping later. That should be migrated
to the universal landing spine rule.

---

Because self and other proposals may split and later point back at the same
underlying entity or affordance record, traversal needs an explicit join/merge
rule. When two active branches converge on the same durable node, amplitude should
not be treated as two unrelated observations forever, and it should not be
silently discarded. The target traversal property is:

```text
split:
  parent amplitude divides across outgoing active branches

join:
  convergent path amplitudes remerge at the shared node
```

The merged node still needs to retain path provenance for grading/explanation:

```text
sharedNode
  reachedBy:
    selfProposal -> engageRecord -> engageResultRecord -> michael
    otherProposal -> engageRecord -> engageResultRecord -> michael
  mergedAmplitude: f(pathA, pathB)
```

Full same-wave convergence is deferred. The first implementation uses strict
path-local visited-once traversal, redistributes allocation across still-open
edges, and records blocked-convergence provenance. Do not choose a merge function
ad hoc during implementation. See
[proportional-amplitude-allocation.md](./proportional-amplitude-allocation.md#merge-interaction).

---

## One Node Type

The runtime should reuse the existing `MemoryCell` struct as the single durable
graph node type. Do not introduce separate graph-node structs for
`EngageRecord`, `PerceiveRecord`, `MoveRecord`, `EngageResultRecord`, etc.

`MemoryCell` is an implementation carrier for the canonical immutable `Node`.
Node identity, kind, and content are immutable after birth. Relationships are
separate edge records; adding topology must not mutate a node-owned child map in
the target model.

The differences between those records should be expressed by:

```text
MemoryCell.ID
MemoryCell.AffordanceKind / Kind
MemoryCell.Content
external Edge records
topological position
```

The current code calls this role tag `Kind`:

```go
type MemoryCell struct {
    id      string
    kind    Kind
    content any
}

type EdgeRecord struct {
    id, fromID, toID string
    kind             EdgeKind
}
```

The target naming should move toward `AffordanceKind` so the abstraction is
clearer:

```text
AffordanceKind = what role/affordance this memory cell represents
```

Examples:

```text
proposal
engage
engage:result
perceive
perceive:result
move
move:result
self:predict
other:predict
midlevel:frame
```

This is a naming/coherence change, not a request to create many node types.
`MemoryCell` remains the record.

---

## Record Shape

The `MemoryCell` is the affordance invocation record or result record.

```text
AffordanceKind / Kind = affordance name or result kind
Content = typed payload/result for that affordance
Immutable invocation metadata/content:
  from
  to
External edge:
  invocation -> affordanceResultRecord
```

Examples of `Kind`:

```text
engage
perceive
move
wait
bump
midlevel:frame
other:predict
self:predict
motor:collapse
```

`KindAction` is legacy. It currently marks the motor-side proposal realization
record used by `self:grade`, but the target model should remove the special
`KindAction` role. A motor action should become a normal affordance invocation
cell:

```text
proposalRecord
  -> MemoryCell{AffordanceKind: engage}
      -> MemoryCell{AffordanceKind: engage:result}
```

During migration, `KindAction` may remain as compatibility for grading. The end
state is: no separate `action` kind for motor actions; use the concrete
affordance kind (`engage`, `perceive`, `move`, `wait`) plus topology.

The word "affordance" is allowed to cover both world affordances and cognitive
affordances. What they have in common is not domain semantics. What they have in
common is the directed invocation shape.

---

## Endpoint Semantics

Each invocation has exactly one `from` and exactly one `to` in its metadata.

`from` is the endpoint where the invocation originates.

`to` is the endpoint the invocation is directed at.

For some affordances, `to` is an agent. For others, it is a world, place, space,
process, or cognitive role.

```text
engage:
  from = speaker/actor
  to   = recipient

perceive:
  from = perceiver
  to   = perceived world/place/space

move:
  from = mover
  to   = destination/place

midlevel:frame:
  from = sensory/membrane/current episode
  to   = midlevel

other:predict:
  from = midlevel/current gist
  to   = other-model

self:predict:
  from = midlevel/current gist
  to   = self-model

motor:collapse:
  from = self proposal
  to   = motor/world boundary
```

Additional participant edges may exist later on result records, but they are not
the base invocation model. The base invocation metadata is `from` + `to`; the
base traversal content model is `invocation -> result -> entities`.

Optional participant edges:

```text
observed:<id>
recipient:<id>
location:<id>
object:<id>
device:<id>
result-source:<id>
```

These must not replace `from` and `to`.

---

## Transient Envelope

The runtime already has the right central envelope shape:

```go
type Impulse struct {
    Primitive string
    TargetID  string
    ResultMap string
    Payload   any

    Source    *ImpulseSource
    Return    bool
}

type ImpulseSource struct {
    SpaceID string
    PointID string
    Coord   Coord
}
```

The universal mapping is:

```text
Primitive      = affordance name / invocation kind
Source.SpaceID = invocation metadata from
TargetID       = invocation metadata to
Payload        = carried content or return result
Return         = this impulse is reafference / a return from an earlier invocation
```

The words do not need to match. `Source`/`TargetID` already have the same semantics
as invocation metadata `from`/`to`.

Universal rules:

1. Every new impulse must set `Source.SpaceID`.
2. Every new impulse must set `TargetID`.
3. `Primitive` is always the affordance name, including cognitive/DAG affordances.
4. `Source.SpaceID` and `TargetID` are not limited to people. They may name agents,
   worlds, places, spaces, devices, or cognitive processes.
5. Durable affordance records are minted from the impulse envelope:

```text
invocationRecord.Kind          = impulse.Primitive
invocationRecord.Content       = typed invocation payload/metadata
invocationRecord.Metadata.from = impulse.Source.SpaceID
invocationRecord.Metadata.to   = impulse.TargetID

resultRecord.Kind              = impulse.Primitive + ":result"
resultRecord.Content           = impulse.Payload or typed result payload
```

Examples:

```text
michael -> skyra engage:
  Primitive      = engage
  Source.SpaceID = michael
  TargetID       = skyra
  Payload        = "can u see me?"

skyra -> world perceive:
  Primitive      = perceive
  Source.SpaceID = skyra
  TargetID       = world/current-place

midlevel -> self self:predict:
  Primitive      = self:predict
  Source.SpaceID = midlevel
  TargetID       = self

self -> motor motor:collapse:
  Primitive      = motor:collapse
  Source.SpaceID = self
  TargetID       = motor
```

Missing `TargetID` is legacy debt. Empty target should not be introduced in new
affordance paths.

### Route Belongs On The Board

`Route any` is workflow/control state, not world/content state.

It should live on the board/work item, not on `Impulse`.

Current code may use `Impulse.Route` as a courier for structured routing
consequences:

```text
LLM output:
  <match>2</match>

parsed routing consequence:
  Route{Kind: verdict, Matched: true, Score: 0.30}
```

That route is not the raw LLM output. It is the parsed control meaning that tells
the workflow what should happen next.

Target split:

```text
Impulse:
  domain/event envelope

Board/work item:
  workflow/control state
  route/verdict/reframe/handoff

MemoryCell result records:
  durable content/evidence
```

Compatibility rule:

```text
Remove Route from Impulse.
New route semantics should be written to the board/work item.
Only write route data into durable memory when a worker intentionally records it
as result/evidence content.
```

### Amplitude Belongs On Spacetime

Amplitude is kernel wave-start metadata, not domain/event content.

Target:

```go
type Spacetime struct {
    Origin    Space
    Traversed Space
    Impulse   Impulse
    Amplitude float64
}
```

`Impulse` says what event/content is moving. `Spacetime` says how the time kernel
realizes that event through a traversed space.

---

## Current Mismatch

### Perceive

Current perceive ingest is close:

```text
grouping -> gist -> perceiveRecord -> entity -> grouping
```

But it records each perceived entity as its own perceive record. The target model
separates the outbound invocation from the returned result:

```text
grouping
  -> gist
    -> selfProposalRecord
      -> perceiveRecord
        metadata: {from: skyra, to: world/current-place}
        -> perceiveResultRecord
          -> skyra
          -> world/current-place
          -> observed entities
```

Observed entities can remain payload initially or become result participant edges.
The key rule is: intended `from`/`to` belong on the invocation metadata; confirmed
entities belong under the result.

### Engage

Current external engage is split:

```text
entity(michael) -> InteractionRecord
gist            -> InteractionRecord
entity(michael) -> grouping
```

Target shape:

```text
grouping
  -> gist
    -> otherProposalRecord
      -> engageRecord
        metadata: {from: michael, to: skyra}
        -> engageResultRecord
          -> michael
          -> skyra

michael -> grouping
skyra   -> grouping
```

Skyra's outbound engage must use the exact same shape:

```text
grouping
  -> gist
    -> selfProposalRecord
      -> engageRecord
        metadata: {from: skyra, to: michael}
        -> engageResultRecord
          -> skyra
          -> michael
```

If an external event arrives before Skyra has inferred the other's proposal, the
system should still preserve the target topology by creating an implicit/unknown
`otherProposalRecord` and attaching the observed affordance record below it.
Later other-prediction may refine, grade, or relabel that proposal.

### Incoming Before Gist

An incoming impulse can arrive before any active gist exists. That should not
force the system to abandon the unified topology.

Target behavior:

```text
incoming impulse
  -> ensure grouping
  -> ensure gist
  -> ensure implicit proposal record
  -> append affordance invocation record
      -> append affordance result record
          -> entities
```

The first gist may be weak or empty:

```text
grouping{cold-start}
  -> gist{text:""}
    -> otherProposalRecord{implicit, status:unknown}
      -> engageRecord
        metadata:{from:michael, to:skyra}
        -> engageResultRecord
          -> michael
          -> skyra
```

Midlevel can later edit the gist once it has enough context. The important rule is
that early sensory intake should create a valid landing spine, not a separate
side-channel shape that has to be reconciled later.

### Motor Action

Current motor action hangs under the self proposal:

```text
gist -> selfPrediction -> actionRecord
```

That preserves the proposal/action realization link, but the record is still a
legacy motor-local shape. Conceptually, `actionRecord`, `engageRecord`,
`perceiveRecord`, and `moveRecord` are all concrete affordance invocation records.
The target is one durable record that can be reached through both process and
episode topology.

Target shape:

```text
gist
  -> selfProposalRecord
    -> engageRecord
      metadata: {from: skyra, to: michael}
      -> engageResultRecord
        -> skyra
        -> michael
```

The proposal may create more than one concrete action over its lifetime:

```text
gist
  -> selfProposalRecord
    -> perceiveRecord
      -> perceiveResultRecord
    -> moveRecord
      -> moveResultRecord
    -> engageRecord
      -> engageResultRecord
```

Execution grading should eventually grade at two levels:

```text
step grade     = did this concrete affordance execute?
proposal grade = did the long-running process accomplish its intent?
gist grade     = did the episode framing remain coherent?
```

---

## Rendering Rule

`Recent exchange` must render from affordance invocation records, not only from
`InteractionRecord`s in the entity index.

Minimal exchange-renderable affordances:

```text
engage
move
perceive
wait
bump
```

The renderer should be able to produce:

```text
[t1] michael -> skyra engage: skyra hello can u see me?
[t2] skyra -> world perceive
[t3] skyra -> michael engage: I can see you.
```

This prevents Skyra's own actions from appearing only as self-referential frame
body material while Michael's utterances appear in the exchange transcript.

---

## Traversal Rule

The frame traversal may still render affordance records as context, but a single
event should not appear as two unrelated concepts.

Bad shape:

```text
Recent exchange:
  michael: hello

Frame body:
  self proposes: tell michael hello
  action: engage michael hello
```

Better shape:

```text
Recent exchange:
  michael -> skyra engage: hello
  skyra -> michael engage: hello

Frame body:
  Framing: greeting Michael
  prior self prediction: tell michael hello
```

Predictions are cognitive affordance invocations. Concrete world actions are
world affordance invocations. They can be connected, but they should not collapse
into one record.

---

## Migration Plan

1. Introduce the graph grammar in tests/specs before runtime migration.
2. Treat proposal records as process containers under the active gist:
   `gist -> proposalRecord`.
3. Add an `EngageRecord` payload and make inbound external engage write:
   `gist -> otherProposalRecord -> engageRecord -> engageResultRecord -> entities`.
4. Make outbound motor `engage` write:
   `gist -> selfProposalRecord -> engageRecord -> engageResultRecord -> entities`.
5. Make outbound motor `perceive` write:
   `gist -> selfProposalRecord -> perceiveRecord -> perceiveResultRecord -> entities`.
6. Add invocation metadata fields for `from` and `to` on perceive/engage/move
   invocation records.
7. Change `Recent exchange` to render from affordance invocation records.
8. Teach execution grading to treat the new affordance invocation record as the
   realization marker, replacing the legacy `KindAction` dependency.
9. Move cognitive DAG steps toward the same invocation/result shape after world
   affordances are stable.

---

## Deferred Affordance-Migration Questions

These do not change the canonical layered cognitive graph. They are migration or
future Praxis/participant-role decisions.

1. What is the canonical `to` endpoint for `perceive`: `world`, current place,
   mounted space, or perception boundary cell?
2. Does each observed entity become an extra participant edge immediately, or stay
   inside the perceive payload until participant roles are generalized?
3. Should `midlevel:frame`, `other:predict`, `self:predict`, and
   `motor:collapse` be stored in the same episode spine on day one, or should the
   first migration cover only world affordances?
4. Should the old `InteractionRecord` collapse into `EngageRecord`, or remain as
   a compatibility payload during migration?
5. Does `entity -> grouping` apply to both endpoints for every affordance, even
   when one endpoint is a process role such as `motor`?
6. What fields distinguish an unknown/implicit `otherProposalRecord` from an
   explicit prediction made by `other:predict`?

Same-wave convergence is not an open first-implementation choice: use strict
path-local visited-once traversal, redistribute over still-open edges, and record
blocked convergence in traversal provenance. A richer merge model is deferred
and must receive its own specification before implementation.
