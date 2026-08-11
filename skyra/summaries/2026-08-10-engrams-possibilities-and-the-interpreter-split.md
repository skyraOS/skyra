# Skyra — Engrams, Possibilities, and the Interpreter Split

**Date:** 2026-08-10
**Scope:** The `specs/2026-08-09` set, promoted from the `08-08` whiteboard and worked through to an
implementation proposal.
**Status: SUPERSEDED.** Nothing is built.

> **Superseded by [`2026-08-11-the-kernel-boundary-and-the-transaction-model`](2026-08-11-the-kernel-boundary-and-the-transaction-model.md).**
> The vocabulary here is still correct — nothing was renamed under it — so this remains readable.
> What it lacks is roughly half the current design: the kernel's active/durable split, monitor,
> impulses, the chain as a transaction, permissions as engrams, the trust boundary and the
> interpretation test, and the origination rule restated by its harm. Read the newer one.

> **Supersedes [`2026-08-08-select-return-and-the-possibility-model`](2026-08-08-select-return-and-the-possibility-model.md).**
> That summary's vocabulary is obsolete — `possibility` meant the durable thing, `return` meant the
> value up, entities were places. All three changed.

---

## What the system is trying to be

An agent runtime built on one loop: something arrives, it is graded against an expectation, that
expectation determines the next action, that action determines the next expectation.

The commitments that constrain every decision, inherited and not re-argued:

```text
TRUTH IS DERIVED, NOT STORED   the graph holds structure and history, never a claim about the
                               present. "is the door open" requires looking.
APPEND-ONLY                    nothing is rewritten.
SAFE BY INCAPACITY             constraints are held by removing the ability to violate them.
NO SELF-REPORT AS GROUNDING    a thing's claim about its own success is an expression, not
                               evidence.
LEARNING IS LOCAL WRITES       caching may be global reads.
```

## Vocabulary

```text
ENGRAM        DURABLE. a fixed expectation with a history accumulating around it. owned by
              exactly one parent; two parents structurally cannot share one.
POSSIBILITY   EPHEMERAL. one opening of an engram. what MIGHT happen, undetermined until commit.
ENTITY        the identity that MANS a possibility — interprets the expectation, emits.
HOST          the device. where an entity is reachable and where effects land.
KERNEL        pairs an engram with an entity. that pairing opens a possibility.
```

**Durable is what has happened; ephemeral is what might.** Potential should not be the thing that
persists, and a record should not be the thing you invoke.

*(Engrams do **not** reconsolidate. The borrowed name carries a lability the design refuses:
retrieval appends, nothing already written is modified.)*

---

## The model

### Two forms

```text
<select>
  <engramName cite=obsID entity=entID host=hostID> obsID, obsID </engram>
  <engramName cite=obsID entity=entID>             obsID         </engram>
</select>

<commit><observationID></commit>     an orchestrator NAMES
<commit>{ … }</commit>               a leaf ORIGINATES
```

```text
SELECT   ephemeral write. accumulate onto the possibility. dies with it.
COMMIT   durable write.   the ONLY durable write in the system.
```

The two-form vocabulary *is* the durable/ephemeral split. **Every durable write is caused by a
commit**, so eviction safety stops being a property you verify and becomes one you cannot violate.

**Cardinality is the concurrency declaration** — one member is a sequential step, N run together,
every group awaits. No `async`, no flags. `cite` is **per member, never per group**, or selection
credit stops being exact.

### A possibility is data plus an interpreter

```text
pos(obs[], entity) { … }
```

One universal signature — no per-callee contract to discover, and **type safety replaced by
learning**. An expectation is *data* (`Expression`, a string); the entity is what turns it into
behaviour. `Step` is the only method in the design.

```text
NEVER SELECTED   a LEAF. atomic, one-shot. ORIGINATES an observation.
HAS SELECTED     an ORCHESTRATOR. may only name or combine. PERMANENTLY.
```

Permanent because otherwise nothing could distinguish content that came from children from content
invented at the end. **Origination bottoms out in an entity's judgement, or in code.**

**A possibility's lifetime matches the work, never the lifetime of what it made.** A browser stays
open; the possibility that launched it does not.

### Observations

```text
Observation{ ID, Content, Producer EngramID, Cites []ObservationID }
```

`Content` is structured — comparable, renderable into a prompt, bindable into code inputs,
composable. It can embed `Ref`s to other observations, so **attribution granularity follows
composition granularity**: a parent citing `.source` flows credit to that constituent alone.

**`Cites` is derived from content, never declared.** Which makes *credit follows data* structural:
you cannot claim to have used something you did not carry forward.

**Returns are always structured; inputs are polymorphic.** Behaviour is many-formed because there
are many ways to act; evidence is uniform because there is one way to learn.

---

## Credit

**The parent writes, at its commit.** One write covering the possibility's whole life. Selects write
nothing, so an evicted possibility leaves no trace.

```text
CITED BY ANY EMISSION   → counts as USED.     the efficiency metric.
CITED BY THE COMMIT     → becomes a CHILD.    durable structure.
```

> **The durable graph records what WORKED, not what was tried.**

**Everything durable is owned by the parent.** A child engram's history is the record of *that parent
invoking it* — no shared state, so *A never writes into B's memory* needs no enforcement. The cost is
**no transfer**: a bad expectation is independently rediscovered by every parent.

**Which commits to synthesis not learning.** All learning is topological — what changes is which
engrams get invoked under what conditions, never how any body composes.

### Dispersal is asymmetric

```text
ERRORS       stay LOCAL.         the immediate parent's verdict.
ACCEPTANCE   WALKS DOWN by hops. an ephemeral credit possibility walks the constituent chain.
```

Conjunctive success distributes; conjunctive failure does not — *"somewhere along here it broke"* is
at most one edge, and marking all of them is Toy 2 in the credit channel. There is a consequence
argument underneath: a false positive is self-limiting, **a false negative is absorbing**, because
nothing earns its way back if it is never tried.

**Hop profiles are the detector**, and it reads off as a shape:

```text
{hop1, hop2, hop3}   validated locally AND at range.
{hop1, — , — }       cited by its parent, and nothing else has ever agreed. ← self-sustaining.
```

**Records are events, not counters** — which preserves trajectory and set membership, keeping
sibling blindness *approachable* rather than foreclosed. The history partitions by question and by
return shape, and **within a partition the diff is the informative object.**

Span is a **label**, not a bound: the walk follows constituents, which are instances, so it cannot
stray.

---

## Entities, hosts, and the OS

```go
type Entity struct{ ID EntityID }

// Reached through CAPABILITY CLASSES, filled in by a per-OS driver.
type Host interface {
    ID() HostID
    Execution() ExecutionOps   // spawn. the only class instantiated now.
    Storage() StorageOps       // read/write. soon.
    // display · audio: the projection substrate. not named yet.
}
```

**No declarations survived.** Not language, not capacity, not modality — because *a declaration is
made by whoever registered the entity, not by the entity*. A third-party claim nobody verified.
History replaces all of it; cold start is the whole cost and it is cheap.

**Execution is constitutive, not declared** — something with no execution class is not a host. But
spawn alone is expressively sufficient and **cost-catastrophic**: *read a file* becomes *spawn a
reader*, a process per read. Linux has `read()` as a syscall for exactly that reason.

**So hosts follow the Linux device-class model.** The kernel calls through a class interface, a
per-OS driver fills it in, and the kernel never knows what is behind it:

```text
kernel ──┬── execution ──┐
         ├── storage     ├── OS drivers ── heterogeneous hosts
         └── display/audio (the projection substrate)
                         ·
         transport is BENEATH all of this, never a class within it
```

Two properties earn the shape. **An ops struct grows by addition** where a one-method interface
cannot grow at all — which matters because the primitive list is the one surface selection cannot
correct. And **an unimplemented class is not a declaration**, it is a build-time fact, so capability
varies without reintroducing the claims that dissolved.

**Network is not a class** — it is transport. *Fetch this URL* is either spawning an HTTP client or
invoking an HTTP entity, and a network class would absorb both and blur the entity boundary.

**The OS is beneath all of it** — hand-written per-OS drivers, neither entity nor host, not modelled
at all.

```text
TRUSTED BASE   the kernel and the platform layer. hand-written, reviewed, never learned about.
ABOVE IT       everything. selected, recorded, starved if wrong.
```

**Assemble and project** are two operations with two owners: assemble is engram state → a frame,
bounded by the entity; project is frame → the host's medium, bounded by the host. **Projection must
fail loudly** — audio to a device with no speaker must error, or the invocation "succeeds" and
nobody perceived anything.

**Authority follows the select.** A page renders on your screen: their content, *your* rendering, on
your host, invoked by your engram. If a sender could render directly they would be choosing your
coupling.

---

## Discover

One call, four blanks — a complete decomposition of an invocation:

```go
discover(obs []Observation, engram *EngramID, entity []EntityID, address []Address)
```

```text
engram nil   WHAT TO DO      entity nil   WHO DOES IT
address nil  WHERE           obs nil      WITH WHAT   ← retrieval, which the old set carried
                                                         behind `blocked`
```

Bootstrap is the **maximally-nil case**. Three states per parameter — nil, a candidate set, a fixed
value — which makes reach **inheritable through arguments**.

**Not one implementation.** `discover` is an orchestrator whose children are resolvers, so its own
history records which resolver suited which need. Which dissolves a rule into a strategy: *create is
licensed only by rejection* becomes **ordering inside a particular discover's expectation**, and a
version that creates too eagerly is outcompeted rather than forbidden.

**The runtime is an entity**, so system engrams improve by the ordinary mechanism — but its two roles
must not touch: as kernel it drives and writes and is **not selectable**; as entity it interprets and
is.

**The syscall surface dissolved.** `blocked` → `discover(obs: nil)`. `exit` → commit. `kill` →
eviction. Three special forms standing in for what the ordinary vocabulary already covers.

---

## Validation

The five adversarial toys were run against the model and the results recorded in
`specs/2026-07-27/toy-runthroughs.md`, beside the older lineage's:

```text
Toy 1   stability ≠ correctness                PASSES
Toy 2   record count ≠ independent grounding   DISSOLVES · dormant at ingress
Toy 3   cardinality ≠ physical causation       CLOSED, and ④ improves via span
Toy 4   decomposition must not invent state    A CLEANER · B now REPRESENTABLE
Toy 5   attention must not seal itself         LOCALIZED, NOT PREVENTED
```

**Toy 5's entry is the second clause of that file's pass rule** — *represented honestly as
unresolved* — not the first.

---

## What to attack

**1 · Nothing has been built.** All of it is derived. `2026-08-06` could mark `kernel.md` BUILT
because it matched running code; nothing here matches anything. The first build should be the
learning substrate with everything stubbed — hand-seeded engrams, deterministic bodies, one entity,
no model, no create, no ingress — because it tests the claim everything rests on and dodges every
other open item.

**2 · The unattributed term.** Every emission cites its sources, so conclusions trace to what entered
from outside — **except a model call**, whose contribution appears in no citation. Provenance is
exact for observations and absent for priors, in a design whose posture is suspicion of unfalsifiable
claims.

**3 · Sibling blindness.** Exact for selections (one cause), unsolved for compositions (many). Two
independent routes reached this limit, which suggests it belongs to cheap credit assignment
generally.

**4 · The terminus has no pressure.** Something at the end of accepted chains has nothing above it
that can reject and nothing that fails to place. `run-tests` invoked directly, returning "480
passed," cited, accepted by a human reading the same output — short range positive, long range
positive, both wrong *together* because they read one instrument.

**5 · History must not become a cache.** The rule is written; a structural form of it has not been
found, and every other constraint in the set is structural.

**6 · Evidential bloat, concentrated in system space.** `discover` and `create` are invoked on every
discovery and every creation, so they will be the largest engrams by a wide margin — and exactly the
ones that must stay assemblable. Which makes it cheap to measure: watch two engrams, not the graph.

**7 · The capability classes** are the one hand-designed surface in the system. Everything else is
derived or learned; those lists are **chosen**, and a wrong choice cannot be corrected by selection.
The class shape makes being wrong recoverable by *addition* rather than by rewriting every driver —
which is a mitigation, not an answer. Only `execution` is instantiated; the rest earn their place
when something cannot be done without them.

---

## What would most change the assessment

```text
THAT SELECTION BOOTSTRAPS ALONE   its loop closes locally, so learning can start without
                                  synthesis improving — which, under parent-ownership, it never
                                  does. if this is wrong, nothing learns.

THAT THE BOUNDARY SIGNALS REACH   correction enters at two points only — an entity crossing and
                                  a human keeping a change. how far they propagate is untested,
                                  and the whole design now rests on it.

THAT THE FILTER DOES NOT STARVE   conditional childhood keeps the graph clean by attaching only
                                  what worked. whether enough works, early and at depth, to grow
                                  usable structure is untested.
```
