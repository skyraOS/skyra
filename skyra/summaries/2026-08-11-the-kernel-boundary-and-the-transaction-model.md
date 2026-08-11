# Skyra — The Kernel Boundary and the Transaction Model

**Date:** 2026-08-11
**Scope:** The `specs/2026-08-09` set, now ten documents. Complete overview, written to be read cold.
**Status:** WORKING. **Nothing is built.** The live implementation is still `skyra-dna-v4`, whose
spec is `2026-08-06`. This is the `v5` direction.

> **Supersedes [`2026-08-10-engrams-possibilities-and-the-interpreter-split`](2026-08-10-engrams-possibilities-and-the-interpreter-split.md).**
> That one is still accurate as far as it goes — the vocabulary did not change under it. What it
> lacks is everything the kernel now owns: the active/durable split, monitor, impulses, the
> transaction model, permissions as engrams, the trust boundary, and the origination rule restated by
> its harm. Roughly half of what is below postdates it.

> **CONTINUED IN [`2026-08-11-the-first-build-and-the-surface-collapse`](2026-08-11-the-first-build-and-the-surface-collapse.md),
> later the same day.** This document ends by describing what the first build should be. It was then
> built and driven by a model. **Nothing here is superseded** — the vocabulary, the mechanisms and
> the document map are all still correct — but three things below are now settled by evidence rather
> than argument, and should be read with the continuation beside them:
>
> - **`Status: WORKING. Nothing is built.`** is out of date. `skyra-dna-v5` exists and has run.
> - **The central bet** — *all learning is topological* — is described below as argued and never
>   observed. It has now been observed.
> - **The kernel surface** described in §3 is larger than what the build needed. `Drive` is a
>   select and `Accept` is a commit; there is no control plane.

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

The central bet, which everything else serves:

> **All learning is topological.** Expectations never improve. What improves is *which* engram gets
> selected, by *whom*, with *what*. Selection is the only learning mechanism in the system.

## Vocabulary

```text
ENGRAM        DURABLE. a fixed expectation with a history accumulating around it. owned by
              exactly one parent; two parents structurally cannot share one.
POSSIBILITY   EPHEMERAL. one opening of an engram. what MIGHT happen, undetermined until commit.
ENTITY        the identity that MANS a possibility — interprets the expectation, emits.
HOST          the device. where an entity is reachable and where effects land.
KERNEL        pairs an engram with an entity. that pairing opens a possibility.
OBSERVATION   one completed work. the only thing that flows.
CHAIN         a root and everything descending from it. THE TRANSACTION UNIT.
```

**Durable is what has happened; ephemeral is what might.** Potential should not be the thing that
persists, and a record should not be the thing you invoke.

*(Engrams do **not** reconsolidate. The borrowed name carries a lability the design refuses:
retrieval appends, nothing already written is modified.)*

## The document set

| doc | owns |
|---|---|
| `README.md` | set framing, the fork, what the set claims, toy results |
| `grammar.md` | SELECT / COMMIT, cardinality, citation, the five questions, capability as scoping |
| `possibilities.md` | the layering, leaf vs orchestrator, origination, composition, what a possibility sees |
| `kernel.md` | active/durable split, the write chain, grants, the transaction, permissions, monitor, impulses, roots |
| `world.md` | entity, host, capability classes, the trust line, assemble/project, no ground truth |
| `discover.md` | one call four blanks, resolvers, ingress, refinement, the runtime as entity |
| `credit.md` | the write, conditional childhood, the walk, hop profiles, events-not-counts, selection as correction |
| `interfaces.md` | the Go implementation proposal. PROPOSAL, not canon |
| `traversal.md` | what `Perturb`/`Observe` was carrying, and what pays for each item now. CHECKLIST |
| `questions.md` | 14 open items, ordered by how much rests on them. LIVE |

Everything is `SPEC` except the last three. **No doc in the set can be marked `BUILT`** — `2026-08-06`
could, because it matched running code line for line. Nothing here matches anything.

---

# 1 · The grammar

> **Two forms — SELECT and COMMIT. Everything else is a name.**

```text
<select>
  <engramName cite=obsID entity=entID host=hostID> obsID, obsID </engram>
  <engramName cite=obsID entity=entID>             obsID         </engram>
</select>

<commit><observationID></commit>          an orchestrator NAMES
<commit>{ … }</commit>                    a leaf ORIGINATES
```

A possibility emits. It does not spawn and has no API. Because this is a **grammar** rather than an
interface, the incapacity is a *parsing* property — nothing else is expressible, so no rule has to be
remembered.

**Select IS invoke.** There is no separate call form, which makes *selection happens at the call
site* literal.

### Cardinality is the concurrency declaration

```text
CARDINALITY = 1    one causal step.
CARDINALITY > 1    a concurrent frontier chosen by the possibility.
```

No `async`, no `await`, no execution flag. Every group awaits. The runtime does not find, select or
order a frontier — it executes the group already emitted. Members receive the frame **as it was when
the group was emitted** and cannot see one another's returns.

This is also what makes irreducibly joint action representable: two simultaneous releases are a group
of two, with separate engrams and separate credit, rather than both hidden inside one body.

**The accepted cost:** a group is a barrier. One slow member delays the next selection.

### Every emission cites its sources, per member

```text
<engramName cite=obsID> obsID, obsID </engram>
 ^^^^^^^^^^             ^^^^^^^^^^^^
 WHO                    WHAT
```

An observation that supplied the *name* is never bound as an argument, so without `cite` it would
fall through as unused — and it is the observation that made the invocation possible. `cite` also
buys exact **capability provenance**: *how did you come to be able to name this* is answerable for
every invocation ever made.

**Per member, never per group.** Grouping is a scheduling fact and must not become a shared cause;
`(Q → P)` pairs are what recur and carry statistics, and a set does not recur the same way.

```text
CITED FROM AN OBSERVATION   I had to discover this before I could act.
CITED FROM DURABLE          I already had the structure.
```

The ratio between those measures how well-structured a possibility has become.

### The five questions

```text
WHO     the possibility.               named.
WHAT    the observations.              bound as arguments.
WHERE   the entity.                    named — supplied per INVOCATION, so the same possibility
                                       runs in different places without being different.
WHEN    the tick of the expectation.   local. per-cell time, no global clock.
WHY     what prompted it.              the citation.
```

`WHEN` costs global ordering: cross-cell causality is reconstructible only through citation chains,
never timestamps.

### Capability is scoping

> **You can select what you can name; you can name what you were given.**

A possibility holds **names, not content**. Reach does not need a type, it needs a scope — but reach
must not travel in the data channel:

```text
DELEGATION   reach from an entity I CHOSE to invoke.   open() handing back a descriptor.
INJECTION    reach arriving in the data channel.       whoever can talk to me picks what I touch.
```

### One universal signature

```text
pos "pos-name" (obs obs[], entity) { … }
```

One, universally. Nothing per-callee to discover, no compatibility check before invoking.
**Type safety is replaced by learning** — anything may be invoked with anything, and whether that was
useful is what `used/seen` and credit measure.

---

# 2 · Possibilities

```text
DURABLE + INTERACTION = EMERGENT
```

The emergent layer is not represented anywhere. Possibilities are the interaction layer, and are
deliberately ignorant: every distinction requiring judgement belongs above them.

A possibility knows **three things**: the observations it was given, which engram it is opening, and
what its children returned. It cannot perform work except by creating a possibility and handing it
over — an incapacity, not a rule.

### Two phases, and what they are not

```text
GATHER      the sequence of selects.    SELECTION is the skill.
SYNTHESIZE  the terminal composition.   SYNTHESIS is the skill.
```

Explicitly *not* `Perturb`/`Observe`. That split interleaved and was about **permission**; this
boundary is at the end and is about **skill**. `Perturb` was `Observe` with `realized = nil` — a name
for the first iteration of a loop that had already started iterating.

### Leaf and orchestrator, and the origination rule

```text
NEVER SELECTED   a LEAF. atomic, one-shot. ORIGINATES.
HAS SELECTED     an ORCHESTRATOR. may only name or combine. PERMANENTLY.
```

> **Originate only where it could not be mistaken for composition.**

Stated by its **harm** rather than by position, which explains what position cannot. The forfeit
fires at the *first select*, because that is when ambiguity becomes possible. It is *permanent*,
because once children have returned the ambiguity never goes away. **The rule guards against
invention disguised as derivation, not against invention** — which is why a leaf inventing freely is
fine and an orchestrator inventing one field is not.

Kind is derived from composition and never declared.

### Ingress is not origination

Two acts wore one word:

```text
INGRESS   an observation ENTERING the tree from outside.  originated BY DEFINITION.
COMMIT    an observation LEAVING a possibility upward.    originated iff it never selected.
```

A root does not originate a commit — it performs **ingress**. Asking whether an inbound observation
is originated or composed is a category error, because no possibility produced it. So *no parent* and
*no child* are not matched ends of a tree: **one is a channel, the other is a rule about a channel.**

```text
LEAF ORIGINATION   Producer = the leaf engram.        Cites empty.
A CROSSING         Producer = the engram I invoked.   Cites empty.
INGRESS            Producer EMPTY.                    Cites empty.
```

Discriminated by absence rather than a tag. An implementation reading empty-`Cites` as one case will
get it wrong.

**The nameable cost:** an orchestrator cannot say *here are my children's results, plus this note
that is mine*. It must spawn a leaf. That may be right — it keeps composers selectable — but it is a
price, not a necessity.

### Composition is compression, not conservation

The commit is **singular**. A plural return would make every possibility a pipe, letting a parent
reach past its child into the subtree.

```text
SEEN   parameters + every child's return
USED   CITED BY ANY EMISSION — a select or the commit
```

Both consumptions count: one used in a select shaped an *invocation*, one used in the commit entered
the *composition*. Counting only the second would make a discover result permanently wasteful.

`used/seen` is a compression ratio, a **gauge not a verdict**, meaningful only differentially. Low
can mean a bad query, an over-provisioned frame, *or* legitimate exploration that discarded
correctly. Nothing is conserved — information is destroyed on purpose.

### Mint on closure

A possibility mints its durable engram on closure if one does not exist. Creation is not a separate
authoring step: **an expectation is improvised, it runs, and if it closes it persists.**

```text
MINTED ON CLOSURE    an engram exists at all.          records what was TRIED.
FORMED ON CITATION   it becomes a CHILD of the asker.  records what WORKED.
```

**Discovery sees attempts; selection sees successes.** A possibility walks its engram's children, so
it only sees what worked. Discovery walks the index, so it finds orphans nobody has used — which is
memoization, not bloat.

### What a possibility sees — down, never up

Everything on its engram's children: names, stored returns and their diffs, credit records, hop
profiles, the cite-source of each selection. No weights change; **the entity sees a richer child set
over time**, which is what topological learning means in practice.

> **No engram can know how it has been received.**

Its own reception lives on its *parents'* engrams, which it cannot reach. A safety property falling
out of locality: a thing that cannot observe whether it is being cited cannot optimise for being
cited. **The citation signal is ungameable from inside.**

**Synthesis has no historical input at all.** The one-shot sees the frame and what was gathered.
So the system improves by changing what gets gathered, never by composing better.

**And the frame travels** — everything a possibility sees ships to the entity's host on every step,
so what it sees is simultaneously the learning mechanism and the per-step transfer cost.

---

# 3 · The kernel

> **The kernel holds everything that is happening. The graph holds only what happened.**

```text
KERNEL HOLDS   possibilities in flight · frames · entity bindings · buffered impulses ·
               the group barrier · roots and heads · staged effects · the chain registry.
GRAPH HOLDS    engrams · history · children · credit.
```

Not a division of labour — it is what makes three existing claims true rather than aspirational.
*The kernel writes nothing* fails if a running flag lives in the graph. *Mint-on-closure* fails if
unclosed work is discoverable. And **crash semantics collapse to nothing**: the kernel dies,
everything in flight evaporates, and the graph is exactly the set of completed commits. No
reconciliation, no recovery protocol, ever.

### The write chain

```text
POSSIBILITY   emits COMMIT.               AUTHORS.
KERNEL        checks scope + ownership.   AUTHORIZES. writes nothing.
HOST / OS     the storage class.          PERFORMS.
```

The middle claim survives **only because storage is a driver beneath the kernel**. If the kernel held
the store, *writes nothing* would be a policy of restraint rather than an incapacity.

### Two authorizations

```text
THE GRAPH BOUNDARY   emissions.         SCOPE + OWNERSHIP. arithmetic on the frame.
THE WORLD BOUNDARY   host class calls.  GRANTS. not derivable from the frame.
```

For durable writes there is no third check. But name resolution bounds what a possibility can *say*
and nothing about what a spawned process *does* — that gap is what grants fill.

### Grants

> **Authority is granted at the root of a chain and narrows on the way down. Nothing widens it.**

The same shape entity reach already has. **A miss is a question, not an error** — it escalates to the
root; approve performs, deny is an ordinary failed invocation, which makes approval *learnable*.
**An approval is an event and never becomes a grant**, or the permission model widens by use.

**A grant the platform layer cannot realize is a declaration.** If the kernel checks it and then
spawns something the OS does not confine, the grant described nothing and the record reads as
bounded. So the vocabulary is bounded by what a sandbox profile or namespace actually enforces.

### The chain is the transaction

> **A world effect is staged when asked for and realized when the root accepts.**

This repairs an overstatement: `credit.md` says an evicted possibility leaves no trace, *which is
what nothing happened should mean.* True of the graph, **false of the world** — a leaf that spawned a
process that wrote a file has left one.

The **chain** is the unit, not the possibility. Nested aborts have no good answer, and the chain
already has one commit point where authority, acceptance and the human all are.

```text
STORAGE          stage the EFFECT.   an overlay. B reads back what A wrote.
EVERYTHING ELSE  stage the INTENT.   an email cannot be un-sent but can be un-sent-YET.
```

The second is the general case; the overlay is a storage-class convenience. An earlier
stageable/not-stageable split was the wrong cut.

**Materialization follows citation** — the credit walk already computes which invocations contributed
to what was accepted, so one traversal does both. But **a discard is an event, not an absence**:
citation was designed with cheap failure (a mis-ranking), and here the same mistake loses a write.

### Permissions are engrams

A permission is not a special channel. It is an engram, invoked by the kernel; its approval is a
possibility. Which makes the permission model **learnable and competitive** — `ask-the-human`,
`auto-approve-under-X`, `check-a-policy-file` are three engrams for one need and selection moves
between them, exactly like discover's resolvers.

**Kernel-callable is scoping, not a flag.** An engram is kernel-callable iff its name is in the
kernel's root frame. A flag would be a declaration, and the set refuses those.

```text
USER-ROOTED     the ask came from outside. a human at the root.
KERNEL-ROOTED   the kernel originated it. permissions · credit walks · eviction.
```

Two rules keep this from violating *system space may never be the standing root*:

> **A kernel root is derivative — it exists because a user root provoked it, and dies with it.**
> **A kernel-rooted chain can never exceed the grants of the chain that provoked it.**

The second terminates the permission regress *and* blocks escalation through the grantor. **Eviction
is the class to bound** — provoked by a timer rather than a chain, it is the one that could grow with
graph size instead of with work, which is why the registry carries provenance at all.

**Grading falls out**: the effect's observation cites the approval observation, so the ordinary walk
credits the permission engram. No grader invented for system space.

### Monitor

```text
ASSEMBLE   FULL.   at invocation open.      engram state → frame.
MONITOR    DELTA.  at each group closure.   observation → frame'.
```

A subscription whose lifetime is the possibility's, so **an entity carries nothing across
possibilities** and statelessness survives. Monitor does not add state; it moves who maintains it.

```text
IT MUST NOT RETURN        a return value is a third emission channel.
IT FIRES AT GROUP CLOSURE not on member arrival — arrival order is a scheduling accident, and
                          pushing on arrival is partial completion through the back door.
LEAVES NEVER RECEIVE IT   a leaf never selects; there is nothing to subscribe to.
```

> **The invariant: `assemble(frame)` == `assemble(frame₀) + monitor*`.**

If those diverge, the entity holds something the kernel cannot rebuild. It also gives recovery free:
a host dying mid-invocation is a full re-assemble elsewhere.

### Impulses *(provisional)*

**No observation ever changes.** What changes is the world behind a handle. The diff is a diff of a
**referent**, and monitor is where it lands — the missing half of *a live process is a crossing that
streams; results arrive as impulses.*

Self-caused versus world-caused is structural: the kernel knows whether a change followed a commit of
yours in this span. *"I noticed this changed and I did not touch it"* is told, not inferred.

> **An impulse may only concern something already in frame. It cannot introduce a name.**

Otherwise it is injection. Impulses buffer to the same boundary as monitor, and one that reaches an
entity must land in history or replay diverges. Buffered impulses die with the kernel — fine, because
a referent that moved is still moved, so the next assemble re-reads it.

### The only global thing

```text
IT HOLDS     roots · heads · ancestry · bindings.   EXECUTION TOPOLOGY.
IT CANNOT    read content · judge · assign credit.  NO SEMANTICS.
```

A scheduler's view, not a world model. **The locality commitment was about judgement**; execution
topology was never local. And knowing the shape of live chains is not knowing their *order* — ticks
are still per-cell.

From that view and no other: a stalled head, an orphaned root, a runaway, and **kernel-rooted work
outgrowing user-rooted work — the self-maintenance spiral, made countable.**

**Introspection is bounded by the subtree, not by the grammar.** An earlier claim that it must not be
addressable at all named a mechanism where the property was elsewhere: the requirement is *down,
never up*, and asking about your own descendants is down. What it still cannot do is look **upward**
(own reception) or **sideways** (siblings in flight).

---

# 4 · The world

```text
ENTITY   WHO.    interprets. authors behaviour.
HOST     WHERE.  the device. realizes execution.
```

**Entities are stateless.** Identity is behavioural, not physical — `python3.4` and `python3.5` are
different entities. The test: *an entity is the thing you want a single track record of.* Start
coarse and **split on evidence**.

### There are no declarations

> **A declaration is made by whoever REGISTERED the entity, not by the entity.**

Language, capacity, modality, host lists — all dissolved, for that one reason. `type Entity struct{
ID EntityID }`. **History replaces every one of them**: *"Python to this entity worked"* is recorded
and grounded; *"this entity reads Python"* is a guess made at registration. Cold start is the whole
cost and it is cheap.

### Capability classes and the trust line

```text
                    ABOVE — UNTRUSTED
    entities · engrams · expectations · everything selected, recorded, starved
    ═════════════════════════════════════════════════════════════════════════
                    BELOW — TRUSTED BASE

    Skyra kernel
        ├── execution   Spawn.          the one needed first.
        ├── storage     read / write.   cheap ops that must not become spawns.
        ├── display ─┐
        └── audio   ─┴─ the PROJECTION substrate. NOT NAMED YET.
              ↑  OS drivers  ↑  heterogeneous hosts  ·  platform layer: os_spawn · os_write
        transport is BENEATH all of this, never a class within it
```

The Linux device-class model. **An ops struct grows by addition where a one-method interface cannot
grow at all** — which matters because the primitive list is the one surface selection cannot correct.
**An unimplemented class is a build-time fact, not a declaration.**

**Spawn alone is expressively sufficient and cost-catastrophic**: *read a file* becomes *spawn a
reader*.

### The test for which side something is on

Not what effect it has — **whether anything interprets**:

```text
A DRIVER CALL   reaches MECHANISM.      semantics fixed, hand-written, reviewed.
AN ENTITY       reaches an INTERPRETER. behaviour authored elsewhere, learnable, starvable.
```

So a kernel-authorized read and an external API call are **not the same kind of thing**. The
filesystem does not interpret a path; a remote service interprets your request. **An external API is
an entity, never a capability class** — the second independent reason network is not one.

### Spawn is the boundary crossing itself

```text
THE CALL      TRUSTED.     fixed semantics, kernel code.
THE PAYLOAD   UNTRUSTED.   the process does whatever it wants.
```

Every other driver op is trusted end to end. This one exists to release something untrusted, which is
why a grant on it cannot be checked at the call site and must be *realized* by the platform layer.

```text
                    │ NO ESCAPE │ WORLD EFFECT
TRUSTED BEHAVIOUR   │ read      │ write — what STAGING exists for
UNTRUSTED BEHAVIOUR │     —     │ spawn · an external API. UNBOUNDED.
```

**The empty cell is the point.** There is no untrusted-but-harmless call: interpretation and escape
arrive together, which is why the trust line and the effect line keep being mistaken for one line.

### Assemble and project

```text
ASSEMBLE   engram state → a frame.      bounded by the ENTITY. HOW MUCH it can take.
MONITOR    observation → frame'.        same bound, same owner. DELTA.
PROJECT    frame → the host's medium.   bounded by the HOST. WHAT FORM it can realize.
```

**Projection must fail loudly.** Audio to a device with no speaker must error, never silently drop —
otherwise the invocation "succeeds," nobody perceived anything, and the record reads as success.
**The one implementation requirement everything rests on.**

Projection happens **at dispatch**, so an undeliverable member returns instantly and does not block
its group.

### You cannot invoke your own grader

Not *a human is not an entity* — a human and a model are interchangeable as interpreters. The
constraint is on the **relationship**:

```text
DOWN   invoke.          interpreters. capability. reach.
UP     return, or trap. the grader. never invoked from below.
```

Enforced by scoping: the root's caller never passes itself down, so nothing below can name it. **The
organism has no concept of human — only of graders and interpreters.**

### There is no ground truth

An entity's expectation *of itself* and a discoverer's expectation *of the entity* are one structure
held by two parties — **not two claims about a third thing.** Discovery is therefore symmetric; two
parties may diverge permanently and neither is wrong.

**Each entity owns its own engrams.** When I invoke Michael, the engram is *mine* — my record of what
came back and whether I used it. Michael has his own, which I never see.

---

# 5 · Discover

> **All ingress goes through discover, which is why nothing can be dropped.**

```go
discover(obs []Observation, engram *EngramID, entity []EntityID, address []Address)
```

```text
engram  nil   WHAT TO DO      entity  nil   WHO DOES IT
address nil   WHERE           obs     nil   WITH WHAT
```

A complete decomposition of an invocation. `discover(obs: nil, engram: X)` is *"what do I have that X
needs?"* — retrieval, which the old set carried as a separate surface behind `blocked`. Bootstrap is
the maximally-nil case.

Three states per parameter — `nil`, a candidate **set**, a fixed value — which makes reach
inheritable through arguments.

**Discovery is not a lookup.** *"There is nothing here that does X"* is a constructed claim. So
`discover` is an **orchestrator** whose children are resolvers (`by-index`, `by-recency`,
`by-cooccurrence`, `create`), and its history records which resolver suited which need.

Which **dissolves a rule into a strategy**: *create is licensed only by rejection* is not a system
rule, it is ordering inside one discover engram's expectation — and a discover that reaches for
create too eagerly is **outcompeted** rather than forbidden.

**Nothing can be silently dropped**, because an observation does not *arrive* into a frame — it
arrives as the argument to a discover call, whose contract is to resolve it. Unexpressible rather
than forbidden.

**Refinement grows the asker, not the answerer.** A failed return is cited as the prompt for a new
discovery; `find-utensil` stays generic and the parent grows a more specific child. Name the need,
never the complaint — the negation is the `cite`, not the name.

**The runtime is an entity**, so system engrams accumulate history and improve by the ordinary
mechanism. Two roles that must not touch: *as kernel* — pairs, drives, applies writes, not
selectable; *as entity* — interprets system expectations, selectable.

---

# 6 · Credit

> **Errors stay local. Acceptance walks down by hops. Everything lands on an engram.**

**State is written by the parent, at the parent's commit.** It must be the commit, because that is
the first moment `used` exists. Selects write nothing — so **eviction safety is free**, and every
durable write in the system is caused by a commit.

```text
CITED BY ANY EMISSION   → USED.       the efficiency metric.
CITED BY THE COMMIT     → a CHILD.    durable structure.
```

Two deliberately different thresholds: if citation-by-select made children, `discover` would become a
permanent child of everything that ever searched.

> **The durable graph records what WORKED, not what was tried.**

**Everything durable is owned by the parent.** A child engram sits in exactly one parent's map, so
*A never writes into B's memory* stops needing an argument — there is no shared B. Locality becomes
structural, and the no-transfer property follows: a bad expectation is independently rediscovered as
bad by every parent, because there is no "it," only copies.

### The dispersal asymmetry

```text
ERRORS        stay LOCAL.         the immediate parent's verdict, on the child it owns.
ACCEPTANCE    WALKS DOWN by hops. an ephemeral credit possibility walks the chain.
```

Not arbitrary. *"Somewhere along here it broke"* does not distribute — at most one did it. *"This
whole chain worked"* does. And a false positive is self-limiting where **a false negative is not**:
being unselected is absorbing, so nothing earns its way back if it is never tried.

### Events, not counters

> **An engram records the accepted event. Magnitude is derived by counting.**

Append-only without qualification. **Trajectory comes free** — an engram that worked for a year and
failed for a month reads identically to one always mediocre, under a counter. And **set membership
survives**, keeping counterfactual credit approachable rather than foreclosed.

### Hop profiles are the detector

```text
{hop1, hop2, hop3}   validated locally AND at range.
{hop1, hop2, — }     good nearby, never confirmed further out.
{hop1, — , — }       cited by its parent, and nothing else has ever agreed.
```

The third is the self-sustaining subtree, read off directly. **Not one signal weakening with
distance — signals at every scale.**

### Credit follows data, not decisions

```text
PASSED AND CITED     the walk finds it. CREDIT.
PASSED, NOT CITED    the child had it and didn't use it. earns nothing.
NOT PASSED           no data flow. exploratory. UNCITABLE, not merely uncited.
```

Enforced rather than agreed, because citation is derived from the Refs in the commit. Which makes
`Cite` a **grouping key only** — what makes `(Q → P)` statistics possible, carrying no credit itself.

### Selection is the correction mechanism

Fabrication is not prevented, it is **starved**. A child authors none of *which engram, which
observations, which entity, which host*, and does not decide whether it is invoked again.

```text
STARVATION REQUIRES ALTERNATIVES   a parent with one option for a need is not selecting, it is
                                   calling.
```

**The root closes — acceptance is citation.** The human is the outermost parent: the change was kept,
the diff was not reverted. Nobody prompts them to grade anything. And it is now the same event that
**materializes** the chain's staged effects.

```text
PREVENTION   failures are impossible.  bought with complexity.
SELECTION    failures are transient.   bought with dependence on pressure actually arriving.
```

---

# 7 · The implementation shape

`interfaces.md` is a **PROPOSAL**, not canon.

```go
type Engram struct {
    ID          EngramID
    Expectation Expectation          // { Expression string }. immutable. NO language tag.
    History     []*Invocation
    Engrams     map[string]*Engram   // children by name. exactly ONE parent's map.
    ByReturn    map[ObservationID]*Invocation
}

type Observation struct {
    ID       ObservationID
    Content  Content
    Producer EngramID        // EMPTY means INGRESS
    Cites    []ObservationID // DERIVED from Content's Refs. never declared.
}

type Invocation struct {
    Span         SpanID
    Observations []*Observation
    Entity       EntityID       // KEYED
    Host         HostID         // RECORDED, not keyed
    Cite         ObservationID  // a grouping key only
    Return       *Observation
    Cited        bool           // THE LOCAL VERDICT
    Acceptances  []Acceptance   // appended later by the walk
}

type Root struct {                 // a chain. THE TRANSACTION BOUNDARY.
    Possibility *Possibility
    Origin      RootOrigin         // User | Kernel
    Provoker    SpanID
    Grants      []Grant
    staged      []*Effect
}

type Interpreter interface {
    Step(x Expectation, f Frame) Emission
    Monitor(o *Observation)        // void. delta assemble.
}

type Host interface {
    ID() HostID
    Execution() ExecutionOps       // Spawn — the trust boundary itself
    Storage()   StorageOps         // soon
}
```

**Emissions are structs, not interfaces**, because they are data — and the exhaustive type switch is
a guard on the vocabulary: a switch that fails to compile on a third case is better than dispatch
that silently accommodates one.

**Nothing is shared**, and duplication is nearly free because expectations never improve, so there is
nothing to propagate.

**The `Frame` view is not optional.** It strips `Cites`, or an entity could traverse into a child's
subtree and *the parent sees what the child concluded, not what it was told* stops holding.

---

# 8 · Validation

The five adversarial toys in `2026-07-27/toy-runthroughs.md` were run against this model:

```text
Toy 1   stability ≠ correctness                PASSES
Toy 2   record count ≠ independent grounding   DISSOLVES · dormant at ingress
Toy 3   cardinality ≠ physical causation       CLOSED, and ④ improves via span
Toy 4   decomposition must not invent state    A CLEANER · B now REPRESENTABLE
Toy 5   attention must not seal itself         LOCALIZED, NOT PREVENTED
```

**Toy 5 is the second clause of the pass rule** — *represented honestly as unresolved* — not the
first.

Four of five passing is real validation and **is not execution**. A design feels most coherent
exactly when it has been reasoned about for a long time and touched nothing.

---

# 9 · What is open

`questions.md`, ordered by how much rests on them.

```text
 1  NOTHING IS BUILT           the one that dominates. every doc is derived.
 2  THE UNATTRIBUTED TERM      a model call appears in no citation. provenance is exact for
                               observations and absent for priors.
 3  SIBLING BLINDNESS          exact for selections, unsolved for compositions. (the IN-FLIGHT
                               kind is closed on purpose — group members cannot see each other.)
 4  THE TERMINUS HAS NO        nothing above it can reject and nothing fails to place. use is
    PRESSURE                   the positive signal, so use PROTECTS a bad implementation.
 5  NOTHING GATES INGRESS      opening a root is the maximum-privilege operation in the design.
                               NOT the same mechanism as grants. re-entry into a running chain
                               is entirely open.
 6  HISTORY MUST NOT BECOME    the rule is written; a STRUCTURAL form of it has not been found,
    A CACHE                    and every other constraint here is structural.
 7  EVIDENTIAL BLOAT           concentrated in system space — discover and create are the
                               most-invoked engrams and must stay assemblable.
 8  LEAVES LEARN NOTHING       a leaf is as good as the model's priors on the day it was made.
 9  THE GROUP BARRIER          approval escalation makes a chain block on a HUMAN. this arrives
                               earlier than a long-running spawn.
10  GRANTS / THE WORLD         where checked, what is named, content or structure. plus staging
    BOUNDARY                   cost, and CONCURRENT SIBLINGS STAGING CONFLICTING WRITES.
11  WHO MAY SELECT CREATE      scoping, or something else.
12  WHAT PRODUCE IS NOW        may not survive as a peer to discover.
13  ARE ALL KERNEL METHODS     the rule adopted is yes. some may be better as methods.
    ENGRAMS
14  DOES "POSSIBILITY"         may be a perspective label rather than a type.
    SURVIVE AS A WORD
```

## What to attack

**The central bet has never run.** *All learning is topological* is the claim everything rests on,
and it has been argued rather than observed.

**One of two provenance sources is uninstrumentable.** Compared against taint analysis, this
design's provenance is *sound* — nothing is credited that did not flow — and **incomplete in a way
taint systems are not**, because leaf priors emit without being observable. That is §2 stated
precisely, and it is not fixable by a better citation rule.

**Implicit flows were traded away.** *Prompted without being passed* earns nothing. That is the
standard practical taint tradeoff for the standard reason, and it carries the standard consequence:
a chain can be causally driven by something that earns nothing.

**Staging makes the kernel hold a view of the world.** Everything else is bookkeeping over the
runtime's own execution. An overlay per capability class per OS is real machinery nobody has costed,
and conflicting concurrent writes have no answer.

**Approval fatigue is unsolved.** Making the asker an engram makes the *strategy* learnable, not the
human's patience infinite. The gauge is `escalations / invocation` — query complexity, in the active
learning sense — and it is pressure, not prevention.

**The human now occupies both roles in one chain**, grader at the root and interpreter manning a
permission possibility partway down. Permitted, and untested.

## What would most change the assessment

```text
IF SELECTION DOES NOT       everything else is scaffolding for a mechanism that does not work.
CONVERGE
IF THE RECORD CHANNEL       one record per invocation with no write-time collapse either stays
GROWS UNBOUNDED             bounded or the frame stops being assemblable.
IF LEAVES DO MOST OF        then "leaves learn nothing" is not a footnote, it is the system.
THE WORK
```

---

# 10 · Where it stands

```text
2026-08-06   the spec for what RUNS.       v4. Perturb/Observe. still live.
2026-08-09   the spec for the DIRECTION.   v5. not built, not started.
```

The fork is by version: `skyra-dna-v4` stays and runs, the new direction takes `skyra-dna-v5`.
Nothing is deleted and nothing is migrated. **Where this set disagrees with `2026-07-27` or
`2026-08-06`, it has not earned the disagreement**, because it has not run.

### The first build

Unchanged, and it is the only item that moves anything else:

> **The learning substrate with everything else stubbed** — hand-seeded engrams, deterministic
> bodies over child state, one entity, no model, no create, no ingress. It tests the claim everything
> rests on and dodges every other open item.

```text
WHAT IT TESTS   does credit accumulate as described · does selection measurably improve · does
                the hop profile appear on a self-sustaining subtree · does the record channel
                stay bounded · does the TERMINUS case behave as predicted

SCENARIO 1      CONVERGENCE. two engrams for one need, one works 90% and one 30%. does the
                policy find it, and how many episodes does it take?
```

The terminus case is worth building deliberately: **demonstrating a known limit is stronger evidence
the model behaves as understood than any passing case.**

Note what the first build **cannot** reach: it has no ingress, so §5 is invisible to it; no world
effects, so the transaction model and grants are untested; and one entity, so nothing about entity
choice is exercised. Those surface the first time something puts a socket in front of the runtime.
