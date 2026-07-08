# Outstanding Reform Tracks

**Date:** 2026-07-08 · **Status:** DRAFT
**Parents:** [directed-affordance-invocation.md](./directed-affordance-invocation.md),
[dag-kernel-responsibility-boundary.md](./dag-kernel-responsibility-boundary.md),
[proportional-amplitude-allocation.md](./proportional-amplitude-allocation.md)

This spec collects the remaining reform tracks after settling the universal
affordance record model and cold-start landing spine.

The settled ground:

```text
MemoryCell is the one durable graph node / record.
AffordanceKind / Kind differentiates role.
The target chain is:

gist
  -> proposal/process record
      -> affordance invocation record
          -> affordance result record
              -> entities/spaces
```

Every incoming impulse must land on an active grouping/gist. If none exists,
create a blank `needs-framing` grouping/gist first, then hand that spine to
midlevel for sharpening.

---

## 3. Self / Other Reform

This replaces the narrower "ActionRecord collapse" item.

The current code treats self and other as prediction/grade workers with legacy
record shapes:

```text
self:
  gist -> self Prediction -> KindAction

other:
  gist -> other Prediction distribution
  later external engage grades the prior distribution
```

The target model should treat self and other as process/proposal records that own
affordance invocations and results:

```text
gist
  -> selfProposal/process
      -> engage/perceive/move invocation
          -> result

gist
  -> otherProposal/process
      -> observed/inferred engage/perceive/move invocation
          -> result
```

### KindAction Deprecation

`KindAction` is legacy. It currently exists because `self:grade` finds the latest
ungraded self prediction by looking for a `KindAction` child.

Target:

```text
self:grade reads:
  selfProposal/process
    -> affordance invocation
        -> affordance result

not:
  selfPrediction
    -> KindAction
```

Migration:

```text
phase 1:
  motor writes both legacy KindAction and new invocation/result chain

phase 2:
  self:grade reads the invocation/result chain

phase 3:
  remove KindAction and ActionRecord as special motor records
```

### Other Proposal Reform

Inbound external actions should not remain entity-index-only events. They should
land under an `otherProposal/process`.

Cold start:

```text
gist{needs-framing}
  -> otherProposal{implicit, status:unknown}
      -> engage invocation
          -> engage result
              -> michael
              -> skyra
```

Later `other:predict` can refine or replace the implicit process with an explicit
proposal/distribution. The exact relationship is still open:

```text
- edit the implicit otherProposal,
- link explicit predictions under it,
- or supersede it with a new proposal record.
```

### Self Proposal Reform

Self proposals should become long-running process containers, not one-shot
prediction strings.

Target:

```text
selfProposal/process
  -> perceive invocation
      -> perceive result
  -> move invocation
      -> move result
  -> engage invocation
      -> engage result
```

This allows step-level grading and process-level grading to separate:

```text
step grade:
  did this concrete affordance execute?

proposal/process grade:
  did the process accomplish the intent?
```

---

## 4. DAG / Kernel Boundary

Existing spec: [dag-kernel-responsibility-boundary.md](./dag-kernel-responsibility-boundary.md)

Current dnav2 genes mostly do the work inside `Observe`, and the orchestrator
reads `root.Condition` afterward.

Target:

```text
DAG:
  creates/mounts ephemeral worker

Observe:
  query / activation / intake strategy

kernel:
  traverses the worker's exposed graph surface

Express:
  performs or commits the affordance work
  writes invocation/result records
  reports condition
```

This is not urgent for the first data-model migration, but it should be fixed
before the DAG semantics harden.

---

## 5. Impulse Envelope Universalization

The runtime `Impulse` envelope is the right transient carrier:

```go
type Impulse struct {
    Primitive string
    TargetID  string
    ResultMap string
    Payload   any

    Source    *ImpulseSource
    Amplitude float64
    Return    bool
}
```

Mapping:

```text
Primitive      = AffordanceKind / invocation kind
Source.SpaceID = invocation metadata from
TargetID       = invocation metadata to
Payload        = carried content or return/result payload
Return         = reafference/result path
```

Universal rule:

```text
new impulses should set Source.SpaceID and TargetID
```

### Route Moves To The Board

`Route` is workflow/control state. It should live on the board/work item, not on
the domain/event impulse.

Current `Impulse.Route` is a temporary courier for parsed routing consequences,
such as verdicts and reframe requests. It is not the raw LLM output.

Target:

```text
Impulse:
  primitive/source/target/payload/return/amplitude

Board/work item:
  route/verdict/reframe/handoff/grouping/status

Durable MemoryCell result:
  content/evidence that workers intentionally record
```

Migration:

```text
keep Impulse.Route as compatibility bridge
write new route state to board/work item
remove Impulse.Route after worker handoff no longer needs it
```

---

## 6. Traversal Amplitude Split

**Status:** Settled conceptually; implementation tabled until the data-model and
kernel-interface changes are ready.

Existing spec: [proportional-amplitude-allocation.md](./proportional-amplitude-allocation.md)

First implementation target:

```text
strict visited-once
proportional split over open children
redistribute amplitude that would have gone to already-visited children
record blocked convergence metadata
```

Current code splits amplitude equally across candidate children. Target code
should split by normalized positive edge product.

---

## 7. First Implementation Plan

The first code slice should be small and compatible.

Proposed order:

```text
1. Add universal landing-spine helper.
2. Make external engage use the landing spine.
3. Add MemoryCell affordance invocation/result writing helpers.
4. Make motor write outbound engage/perceive/move invocation records.
5. Keep legacy KindAction until self:grade is migrated.
6. Make sensory/reafference attach result records.
7. Update recent exchange rendering to read invocation/result chains.
8. Migrate self:grade away from KindAction.
9. Remove KindAction/ActionRecord special path.
```

The likely first safe commit is:

```text
external engage and motor engage/perceive create new MemoryCell chains,
while existing InteractionRecord and KindAction paths remain for compatibility.
```
