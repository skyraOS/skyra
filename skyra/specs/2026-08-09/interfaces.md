# Interfaces

**Date:** 2026-08-09
**Status: PROPOSAL.** The implementation shape for `skyra-dna-v5`, being iterated. Not canon — where
this disagrees with the rest of the set, the set wins until this is settled.

> **The kernel pairs an engram with an entity. That pairing opens a possibility.**

---

## Vocabulary

```text
ENGRAM              DURABLE. the trace. a fixed expectation with a history accumulating
                    around it. owned by exactly one parent.

POSSIBILITY         EPHEMERAL. one opening of an engram. what MIGHT happen, undetermined
                    until it commits.

ENTITY              the identity that MANS a possibility. it interprets the engram's
                    expectation and produces its selections and its commit.

HOST                the device. realizes execution. entities register on hosts; hosts have
                    a capacity for entities.

KERNEL              pairs an engram with an entity, drives the loop, and hands the record
                    to the graph.
```

**Durable is what has happened; ephemeral is what might.** That pairing is the point of the naming —
potential should not be the thing that persists, and a record should not be the thing you invoke.

**One caveat carried with `engram`:** engram theory includes **reconsolidation**, where retrieval
makes a memory labile and rewritable. **That does not transfer.** Append-only means a replay adds
records and never modifies existing ones; only the aggregate drifts. Take the part that describes the
mechanism, refuse the part that imports a law this design does not have.

**Also settled here so it is visible rather than assumed:** the emission is `select`, because
*selection* is load-bearing throughout the set — selection is the skill, selection happens at the
call site, selection is the correction mechanism. *Exploration* names the phase, not the act.
`commit` replaces `return`, because in every language a reader knows `return` means *hand back a
value, no side effect*, and this one **writes**.

---

## The durable layer

```go
package graph

type EngramID      string
type ObservationID string
type EntityID      string
type HostID        string
type SpanID        string
type Language      string   // "nl" · "python3.11" · "bash" · "sql"
```

### Engram

```go
// An Engram is the durable trace: a fixed expectation with a history accumulating
// around it, and children of its own.
//
// It does NOT reconsolidate. Retrieval appends; nothing already written is modified.
type Engram struct {
    // Identity. An engram does NOT hold its name — that is the parent's map key.
    // ID is what it IS; the name is how it is REACHED.
    ID EngramID

    Expectation Expectation
    History     []*Invocation

    // Its own children, keyed by name. An engram lives inside exactly ONE parent's
    // map, so two parents structurally cannot share one.
    Engrams map[string]*Engram

    // The credit walk's entry point. Derived, rebuildable, deletable.
    ByReturn map[ObservationID]*Invocation
}
```

**Nothing is shared, and that is structural rather than a rule.** *A never writes to B* stops needing
an argument — there is no shared B. Which makes the no-transfer property a consequence rather than a
concession: a bad expectation is independently rediscovered as bad by every parent, because there is
no "it," only copies.

**Duplication is nearly free**, and that is what makes this work. Every parent needing a capability
holds its own copy. That would be expensive if expectations improved and improvements had to
propagate — **they don't.** Expectations never change; only selection learns. Sharing an immutable
thing buys storage and nothing else.

It also unifies two things that used to be separate mechanisms:

```text
FIND    search the index, COPY an expectation in.   → a local engram
BUILD   author a new one.                           → a local engram
```

Same outcome; they differ only in where the expectation came from.

### Expectation

```go
// An Expectation is what an engram is FOR. Its EXPRESSION is the form it takes.
//
// Immutable. Expectations never improve — they are authored once and re-run, and
// everything that accumulates accumulates AROUND them.
//
// It carries NO LANGUAGE TAG. See below.
type Expectation struct {
    Expression string
}
```

### There is no pairing check, and that is correct

An earlier draft put a `Language` tag here so the kernel could refuse a mismatched pairing. That was
wrong, and the reason matters more than the field:

> **The tag would be a self-report, and the kernel would act on it without verifying.**

Whoever authored the expectation declares what it is; the kernel trusts the declaration. That is the
one thing this design refuses everywhere — and the verification that would matter, *is this actually
valid Python*, can only be performed by Python. So the check catches honest mislabeling, which is
rare, and passes confident garbage, which is not.

**Handing a compiler natural language should fail.** The entity reading it and failing **is** the
observation — grounded, where the tag was decoration.

**And the attribution is already right**, which is what makes this closed rather than merely
tolerated:

```text
THE INVOKER   chose the pairing. knowing the syntax has always been the caller's job.
THE RECORD    lands on the INVOKER'S engram.
THE LEARNING  is the invoker's selection improving.
```

Nobody has to decide where the blame goes. It is where credit already goes.

Discovery does not need a tag either. It does not *filter*, it **ranks** — and the credit records
rank it, so the system learns which expectations work with which entities by trying. That records
what actually worked rather than what was claimed, which is the same trade as type-safety-replaced-
by-learning.

**The cost, named:** early on, discovery pairs wrongly and fails often. Those failures are cheap — a
compiler erroring costs no model call — and cold start already applies everywhere else.

### Observation

```go
// An Observation is one completed work.
type Observation struct {
    ID      ObservationID
    Content Content

    // WHICH ENGRAM returned this — a specific child of a specific parent, which is
    // exactly the precision the walk needs.
    Producer EngramID

    // DERIVED from Content: every Ref appearing in it. Cached, rebuildable.
    // NEVER declared — see below.
    //
    // EMPTY means it ORIGINATED — an entity's judgement at a leaf, or a crossing.
    Cites []ObservationID
}

type Content map[string]Value

type Value interface{ isValue() }

type Str  string
type Num  float64
type Bool bool
type List []Value
type Obj  map[string]Value

// The one that matters: an observation embedded in another's content.
type Ref ObservationID

// A pointer INTO THE WORLD — a rendered page, a running session, a file. It carries
// its host, and resolving it anywhere else is an ERROR, never a lookup: a path can
// resolve on another host to a DIFFERENT THING WITH THE SAME NAME, and a silent
// wrong answer is worse than a crash.
type WorldRef struct {
    Host   HostID
    Handle string
}
```

### The return path is always structured

```text
IN    BEHAVIOUR.  polymorphic — code, prose, a query, a URL. whatever the entity reads.
OUT   EVIDENCE.   uniform. always structured.
```

**Behaviour is polymorphic because there are many ways to act; evidence is uniform because there is
one way to learn.** Every mechanism downstream of a return runs on observations — the *same question*
comparison, composition, the credit walk, shape partitioning. If observations varied in kind, each
would need cases. One shape means one mechanism.

**The conforming happens in the adapter, not the entity.** A shell does not emit structure; its
adapter wraps stdout into `{stdout, exit}`. Marshal in, structure out — both halves of the same job,
and adapters are already runtime code looked up by entity ID.

> **Structure is not correctness. A well-formed return is not a good return.**

Requiring structure buys comparability, never truth. The temptation will be to treat conformance as
validation — which is *"the tests passed"* in new clothes, and the `run-tests` terminus case is what
happens when someone yields to it. Judgement still bottoms out at use, at the root.

### Cites is derived, never declared

Scan the content for `Ref`s. That is the citation list.

Two places for one fact would eventually disagree — a commit could declare it cited A while its
content references B, and nothing would catch it. Deriving it also makes **credit follows data**
structural instead of a rule we agreed to:

```text
IN YOUR OUTPUT       there was data flow. cited, BY CONSTRUCTION.
NOT IN YOUR OUTPUT   influence at most. exploratory. UNCITABLE, not merely uncited.
```

You can no longer claim to have used something you did not carry forward.

### Nesting makes attribution granularity follow composition granularity

```text
STRUCTURED   {summary: …, source: <obsA>, contrast: <obsB>}
             a parent citing .source flows credit to A ALONE.

BLENDED      prose merging A and B into one field.
             credit is coarse — because attribution genuinely is not there.
```

That does not solve sibling blindness. It narrows it wherever a composer kept the parts separable,
and where it did not, coarse credit is the honest answer rather than a defect.

**A leaf is what authors this.** An orchestrator may only name. A composer is a leaf that receives
A, B, C as arguments and references them — which is how nesting gets in without breaking origination
rights.

**`Producer` being a field is what keeps `ObservationID` opaque.** The alternative was encoding it so
the walk could parse it out, letting one traversal's convenience override a rule the set holds
everywhere.

### Invocation

```go
// One asking, filled in three phases — which is why it is durable rather than
// transient.
type Invocation struct {
    // DETERMINED at invocation, WRITTEN at commit. Nothing is written during
    // exploration, or an evicted possibility leaves a trace.
    Span         SpanID
    Observations []*Observation
    Entity       EntityID
    Cite         ObservationID // what prompted reaching for this. A GROUPING KEY ONLY.

    // WRITTEN at commit.
    Return *Observation
    Cited  bool // did the parent's commit cite it. THE LOCAL VERDICT.

    // APPENDED LATER by the credit walk. The pending part.
    Acceptances []Acceptance
}

type Acceptance struct {
    Span SpanID // which execution accepted. A LABEL — the walk is bounded by
                // constituents being instances, not by this.
    Hop  int    // 1 = the parent's parent accepted. N = accepted N levels up.
}
```

**The invocation carries the ENTITY, not the host.** Same question to Claude and to Michael is a
*different* question. Same question to Claude reached two ways is the *same* question. And if one
entity gives different answers on two hosts, that is not a reason to add the host — it is evidence
**the entity is under-specified**, and the fix is to split it.

---

## Entities and hosts

```go
// An Entity is an IDENTITY. Nothing else.
type Entity struct{ ID EntityID }

// A Host is reached through CAPABILITY CLASSES, filled in by a per-OS driver. The kernel
// calls through the class and never knows what is behind it.
//
// Constitutive rather than declared: something with no execution class is not a host, it
// is a device we cannot use.
type Host interface {
    ID() HostID
    Execution() ExecutionOps   // spawn. the only class instantiated now.
    Storage() StorageOps       // read/write. SOON — see below.
    // display · audio: the projection substrate. NOT NAMED YET.
}

type ExecutionOps interface {
    Spawn(x Expectation, f Frame) (Handle, Streams, error)
}
```

**Classes rather than one method**, because an ops struct grows by addition where a single interface
cannot grow at all — and the primitive list is the one surface selection cannot correct.

**Spawn alone is expressively sufficient and cost-catastrophic**: *read a file* becomes *spawn a
reader*, a process per read. Linux has `read()` as a syscall for exactly that reason.

**An unimplemented class is not a declaration** — it is a build-time fact, so capability varies
without reintroducing the claims that dissolved. Calling into one fails loudly.

**Network is not a class.** It is transport, beneath everything — *fetch this URL* is either spawning
an HTTP client or invoking an HTTP entity, and a network class would absorb both.

**Ports are returned, never specified.** `Spawn` hands back a `Handle`; whether it is backed by a
port, a socket path, or a pipe is the platform layer's business, and the caller never learns. A port
that genuinely matters — a firewall was configured around 8080 — is `--port 8080` **inside the
expression**, which is content, not structure.

**Every declaration dissolved**, and for one reason: a declaration is made by whoever *registered*
the entity, not by the entity. A third-party assertion about a capability nobody verified — worse
than the expectation-side language tag, because at least an author knows what they wrote.

```text
GONE   Reads []Language · Capacity · Presents/Captures modalities · Hosts []HostID
```

**History replaces all of it.** *"Python to this entity worked"* is recorded per invocation and
grounded; *"this entity reads Python"* is a guess made at registration time. Cold start is the whole
cost, and it is cheap: send Python to a TV, it fails immediately, one record, no model call.

```text
GRAPH     entities and hosts as IDENTITIES.
RUNTIME   how to reach each one — addresses, adapters, which entities are on which host.
          mutable, rebuildable from configuration, NEVER durable.
```

**You invoke an entity, not a host.** Delivery picks a host on the way out; the response correlates
by invocation identity, not by channel. A human messaged on their TV and answering on their phone is
one invocation arriving back through a different door.

### The OS is beneath all of it

```text
MODEL       entities · hosts
PLATFORM    os_write · os_spawn · os_open — normalized, per-OS adapters. RUNTIME CODE.
```

**Two different surfaces, easy to conflate:**

```text
PLATFORM LAYER   what the runtime needs to EXIST.     memory, threads, io. per-OS, hand-written.
HOST INTERFACE   what the runtime needs to USE a host. Spawn. uniform across all of them.
```

For `here` they overlap — the local host's `Spawn` is implemented *via* the local platform layer. For
a remote host it is implemented by something on the far side that receives and spawns locally. Same
interface, different implementation, which is exactly why it can be uniform.

Neither entity nor host. Reaching a human through their phone is three roles and only two are
modelled: the OS is **delivery**, the human is the **interpreter**, the phone is **where it lands**.

```text
TRUSTED BASE   the kernel and the platform layer. hand-written, reviewed, never learned about.
ABOVE IT       everything. selected, recorded, starved if wrong.
```

**`spawn` is not `os_spawn`.** Opening a possibility is a kernel allocation; spawning a process is a
platform call. They must not share the word in code.

### Bootstrap

The one place the design requires something be handed to it rather than growing it:

> **The runtime must know how to reach at least one entity that can author an expectation. The
> root's frame must name it.**

That is all. The runtime entity is already present — `here` is the axiom — so it can act via the
platform layer without any registered actor. Everything past this is created on demand, so nothing
about the agent's structure is decided in advance.

## The execution layer

### The kernel

```go
// ONE kernel. Not one per entity — differences in how an entity is reached live in
// runtime adapters, never as separate types here.
type Kernel struct{ /* graph, entity registry, span minting */ }

// Run pairs an engram with an entity and drives the loop. It does NOT check the
// pairing — a mismatch is an ordinary failure, observed and recorded on the
// invoker's engram.
func (k *Kernel) Run(e *Engram, ent EntityID, obs []*Observation) (*Observation, Record)
```

### The runtime is also an entity, and the two roles must not touch

The runtime mans system spaces — discovery, creation — so their engrams accumulate history and
improve by the ordinary mechanism. That pays down the fixed-body trade in
[discover.md](discover.md).

```text
AS KERNEL   pairs, drives, applies writes.   NOT SELECTABLE.
AS ENTITY   interprets system expectations.  SELECTABLE, and emits like anything else.
```

**This is a constraint on the code, not a note.** If a possibility that selects the runtime-entity
can reach the write path, it can influence what is recorded about itself — the one thing the
honest-witness split exists to prevent. The entity role must hold no reference the kernel role uses
to write.

```go
```

Open a possibility · `Step(expectation, frame)` · on `Select`, run the group concurrently and append
to `Returned` · on `Commit`, mint the observation and close.

### One method, and it is the execution boundary

```go
// Step is THE ONLY METHOD IN THE DESIGN. Everything else is data crossing it.
//
// Nothing executes without asking: a possibility cannot invoke, it can only emit a
// request and be invoked in turn.
type Interpreter interface {
    Step(x Expectation, f Frame) Emission
}
```

**Marshalling is what interpreting is** — rendering the frame into whatever the entity actually
reads:

```text
A MODEL     frame → a prompt.
PYTHON      frame → stdin JSON, argv, env.
SHELL       frame → env vars.
A HUMAN     frame → a message.
```

**Neither kind needs a declared signature.** A code expectation reads the frame by *field access*; a
model expectation reads it by *attention*. Both duck-type over what they were handed, which is what
one universal signature requires.

```text
WATCH   a code expectation reading a missing field fails LOUDLY; a model given the same frame
        just does worse QUIETLY. same missing input, very different credit records.
```

### Assemble and project

```text
ASSEMBLE   engram state → a frame.       bounded by the ENTITY. HOW MUCH it can take.
PROJECT    frame → the host's medium.    bounded by the HOST. WHAT FORM it can realize.
```

Two operations, two bounds, two owners. Neither is declared — both are learned by trying, and the
return records what was *actually* done rather than what was claimed possible.

**`assemble`, not `select`** — the emission `select` is invocation; this is internal frame assembly
at another level, and the set already tracks one naming collision it failed to fix in time.

```text
PROJECTION MUST FAIL LOUDLY   audio to a device with no speaker must ERROR, never silently
                              drop. otherwise the invocation "succeeds," nobody perceived
                              anything, and the record reads as success. THE ONE
                              IMPLEMENTATION REQUIREMENT THIS RESTS ON.

DELIVERY vs WORK              projection happens AT DISPATCH. an undeliverable member returns
                              instantly and does not block its group — you do not wait on a
                              human reading a TV to learn the TV had no speaker.
```

### Two emissions

```go
type Emission interface{ isEmission() }

// Select is an INVOCATION GROUP. Cardinality is the concurrency declaration: one
// member is a sequential step, N members run together, and every group awaits.
type Select struct{ Members []Member }

type Member struct {
    Engram       string          // the NAME, resolved in frame
    Cite         ObservationID   // PER MEMBER, never per group
    Entity       EntityID        // resolved in frame. WHO interprets the child.
    Observations []ObservationID
}

// Commit closes the possibility and is THE ONLY DURABLE WRITE.
//
// No Cites field: citation is DERIVED from the Refs appearing in Content.
type Commit struct {
    // Exactly one of these. See leaf and orchestrator.
    Content  Content        // a LEAF originates.
    Observed ObservationID  // an ORCHESTRATOR names.
}
```

Emissions are **structs, not interfaces**, because they are data. The exhaustive type switch is a
**guard on the vocabulary**: two forms is the design's central claim, and a switch that fails to
compile on a third case is better than dispatch that silently accommodates one.

```text
SELECT   ephemeral write. accumulate onto the possibility. dies with it.
COMMIT   durable write.   dump to the graph. outlives everything.
```

> **Every durable write in the system is caused by a commit.** Credit fires when an ancestor commits;
> mint-on-closure is at commit; an inbound observation is not recorded until the possibility that
> used it commits. Eviction safety stops being a property you verify and becomes one you cannot
> violate.

### Leaf and orchestrator, derived from composition

```text
NEVER SELECTED   a LEAF. atomic, one-shot. ORIGINATES an observation.
HAS SELECTED     an ORCHESTRATOR. may only name or combine what came back. PERMANENTLY.
```

**Permanent is necessary, not tidy.** If a possibility could originate *after* orchestrating, nothing
could distinguish content that came from its children from content invented at the end, and *cannot
originate, only combine* would collapse.

Kind is derived from composition and never declared — the third independent arrival at directory
versus file.

**Origination bottoms out in an entity's judgement, or in code**, which is what makes the entity
definition load-bearing rather than descriptive.

### A possibility is what the frame is

```go
// A Possibility is one opening of an engram. The KERNEL holds this.
type Possibility struct {
    Engram *Engram
    Entity EntityID

    Observations []*Observation // the parameters
    Returned     []*Observation // what previous groups handed back
    Selected     bool           // has it forfeited origination
}

// Frame is what a POSSIBILITY sees. An entity is handed it to interpret; the
// perspective belongs to the possibility. Same object, restricted — it
// cannot reach what is not here, so DOWN NEVER UP stops being a rule.
type Frame interface {
    Self() EngramID

    Entity()   EntityID     // who is interpreting ME
    Entities() []EntityID   // who I may ADDRESS. scoped, exactly like names.

    Observations() []*Observation
    Returned()     []*Observation

    // Strips Cites. A possibility sees what its children CONCLUDED, never what they
    // were told.
    Engrams() EngramView
}
```

**`Entities()` is not optional.** A member names an entity and that name must resolve somewhere.
Without it, either entity names resolve globally — a registry, which `world.md` refuses — or a
possibility can only run children on its own entity, which makes delegation impossible. Entity reach
is **inheritable**: a parent hands down a subset, which is how a limb gets narrower authority than
what spawned it.

**The view is not optional either.** Handing an entity the raw engram lets it reach
`Invocation.Return.Cites` and traverse into a child's subtree, and *the parent sees what the child
concluded, not what it was told* stops holding.

```text
AN ENTITY   sees its own engram's children. the return observation, NOT what went into it.
THE WALK    sees the causal graph. a system engram with a fixed expectation — the same
            exemption discover and create already have.
```

---

## Credit

Two lookups, and the durable structures exist to serve them:

```text
OBSERVATION → CITES      reverse causality. where to descend.
OBSERVATION → PRODUCER   which engram to descend into.
```

```text
walk(observation, hop):
    for each c in observation.Cites:
        inv := engramFor(c.Producer).ByReturn[c.ID]
        inv.Acceptances = append(inv.Acceptances, Acceptance{Span, Hop: hop})
        walk(c, hop+1)
```

### Credit follows data, not decisions

> Passing an observation to a child creates a **dependency**. Being *prompted* by one without passing
> it is influence with no data flow — that is exploration, and it earns nothing.

**Already enforced rather than needing a rule:** you can only cite what you can name, and only name
what you were given. So the constituent chain the walk follows *is* the data-flow chain.

```text
PASSED AND CITED     in the constituents. the walk finds it. CREDIT.
PASSED, NOT CITED    the child had it and didn't use it. correct that it earns nothing.
NOT PASSED           no data flow. exploratory. not creditable and not reachable.
```

Which makes `Cite` a **grouping key only** — what makes `(Q → P)` statistics possible and what gets
pruned — carrying no credit of its own.

### The history partitions, and the diff is what carries the signal

Two returns are only comparable if they answered the same question **and came back in the same
shape**. So the history folds twice:

```text
BY QUESTION   which asking this was.
BY SHAPE      what kind of answer came back.
```

```text
{tests: 480} vs {tests: 500}      comparable. THE WORLD MOVED.
{tests: 480} vs {error: timeout}  not comparable. SOMETHING STRUCTURAL HAPPENED.
```

A shape change is its own event, distinct from a value change and usually more important. Errors
partition away from successes without anyone modelling errors.

**Within a partition, the DIFF is the informative object** — not the returns. Learning reads diffs
over an engram's history: `480 → 500` is the signal; two absolute numbers are not.

**Both folds are read-time.** Shape is derived from content, and *same question* is already a
comparison rather than a stored key, so this adds no field and no migration. Changing what counts as
the same shape invalidates nothing.

**And nothing declares the shape**, because it is a property of neither side alone — the same
expectation under Claude and under python3.11 returns different shapes. It belongs to **the
pairing**, so it can only be observed. The partitioned history *is* the knowledge of what this
pairing returns.

Same posture as dropping the language tag, one level up: do not assert what can only be found out.

### Span is a label, not a bound

Span filtering was a hard constraint while the walk descended **durable child structure**, which
holds everything an engram ever did. Descending **constituents** does not have that problem:
observations are *instances*, so the walk lands only on invocations that actually produced this
result.

## The kernel writes nothing

```text
kernel   pairs, drives, and produces a RECORD of what was emitted. NO POLICY.
graph    applies that record as writes. the witness, and the only thing with policy —
         engram formation, the local verdict, credit dispersal.
```

If an interpreter performed the write it could omit it, which is the self-reporting defect the
honest-witness argument exists to prevent. v4's kernel wrote nothing; this one keeps that by moving
the policy out rather than by having none to move.

---

## Open in this proposal

```text
WHAT "SHAPE" MEANS      the history partitions by it, so it needs a definition. the set of
                        field paths? paths plus value kinds? does a nested Ref count as a
                        shape element or as a value? THE ONLY REAL DECISION in the
                        partitioning idea.

REFERENCE COMPARISON    "same question" must DEREFERENCE Refs and compare their content, not
                        compare ref IDs — comparing by identity makes every question unique
                        again, which is what killed the first InvocationKey. dereferencing
                        walks a chain, so comparison depth is a real cost on deeply nested
                        content.

WORLD REFERENCE         the observation is durable; the page it points at is not. the browser
LIFETIME                closes and the reference is stale while the record still reads as a
                        success. this is the third lifetime the old set already hit — graph,
                        process, and the artifact that outlives the process.

ARE OBSERVATIONS        expectations carry no kind tag, and the argument that killed it applies
TYPED                   here too: a kind tag would be SELF-REPORTED, and the consumer finds out
                        by trying. the parallel cuts toward NO. not settled — noted so the
                        argument is present when this comes back.

THE ROOT'S EXPECTATION  every engram has one, and the root is created by a human's invocation.
                        either the root is a special case with none, or its expectation is
                        whatever was typed — which makes the human an entity reading `nl`.
                        consistent, but it should be chosen.

EXPECTATION PROVENANCE  FIND copies an expectation from the index; BUILD authors one.
                        recording which would help discovery rank candidates — and is exactly
                        BORROWED TRUST, which the set forbids from supplying support. if
                        recorded, it must be readable by discovery and invisible to credit.

ID MINTING              EngramID, ObservationID and SpanID are all globally unique and all
                        minted by the kernel. probably one mechanism. NOTE that global
                        uniqueness means `Producer` resolves across entity trees — the walk
                        is bounded by CITES BEING EMPTY at a crossing, not by ID scope, and
                        someone will otherwise reach for scoping to "fix" a boundary that is
                        already held.

WHO MINTS SpanID        "the outer loop." the outer loop does not exist yet.

WHAT COUNTS AS THE      a comparison function rather than a hashed key, so it stays soft.
SAME QUESTION           order-independent over observations is the current lean, since one
                        universal signature gives position no meaning.

EngramView's SHAPE      stripping Cites is the requirement. what else a possibility should and
                        should not see through it has not been worked through. IT IS ALSO A
                        BUDGET boundary — the frame ships to the entity's host on EVERY step,
                        so what a possibility sees is also what gets transferred, forever.

COMPOSERS COMPETE,      if composition happens by invoking a composer, and composers are
THEY DO NOT LEARN       leaves, then composition becomes SELECTABLE but no composer improves.
                        smaller than "synthesis now learns," and the difference matters.

THE os_* INTERFACE      the ONE hand-designed surface in the system. too narrow and things the
                        runtime needs are unreachable; too wide and the kernel starts knowing
                        about Linux. everything else here is derived or learned — this is
                        CHOSEN, and a wrong choice cannot be corrected by selection.

MODALITY CHOICE         rendering is mechanism and should be dumb. but WHICH modality, when a
                        host offers three, is a judgement — probably the same answer as host
                        resolution (dumb default, select a projector if you care), except the
                        choice is where it earns its keep, not the rendering.
```
