# Entities and Hosts

**Date:** 2026-08-09
**Status: SPEC.** Not built.

> **One structure, held locally by whoever holds it. There is no third thing behind it.**

---

## Entity

```text
ENTITY   the identity that MANS a possibility. it interprets an engram's expectation and
         produces its selections and its commit.
```

An earlier draft had *entity = a place of execution*, which was one word doing two jobs. **Who
interprets** and **where it runs** are different questions, and only the first changes what gets
produced.

```text
ENTITY   WHO.    interprets. authors behaviour.
HOST     WHERE.  the device. realizes execution.
```

An entity may be reachable several ways and on several hosts. *How* it is reached is runtime
configuration, never a modelled thing — see below.

**Entities are stateless.** Anything accumulating between invocations is world state and must be
carried explicitly — a session handle, a worktree, a page. If state hides inside an entity, two
invocations that look identical give different answers and nothing can see why: `same question`
becomes a lie, and every fold downstream inherits the falsehood.

For entities we build, that is a requirement. For entities we do not control — a human, a remote
agent — hidden state is a fact, and the comparison degrades. **But it degrades visibly:** an engram
whose returns never converge under identical conditions is pointing at hidden state, and the
partitioned history is what makes that legible.

### Identity is behavioural, not physical

`python3.4` and `python3.5` are different entities, because version changes interpretation. Claude
and GPT are different entities. The test is not what something *is*:

> **An entity is the thing you want a single track record of.**

If two things should share one history, they are one entity. If a failure in one should not count
against the other, they are two. Start coarse and **split on evidence** — when the same entity
behaves differently on two hosts, it is under-specified, and the fix is to split it rather than to
record the host as an excuse.

## Host

```text
HOST   the device. where an entity is reachable and where effects land.
```

Which entities are on a host is a **registry** fact, held by the runtime and rebuildable from
configuration. It is not durable structure, for the same reason addresses are not: it changes, and
the graph is append-only.

### A host is reached through capability classes, and drivers implement them

The Linux device-class model: the kernel calls through a class interface, a driver fills it in, and
the kernel never knows what is behind it.

```text
Skyra kernel
    │
    ├── execution     spawn. the one needed first.
    ├── storage       read/write. cheap ops that must not become spawns.
    ├── display  ─┐
    └── audio    ─┴─ the PROJECTION substrate
          ↑
      OS drivers            linux · macos · windows
          ↑
    heterogeneous hosts
          ·
    transport is BENEATH all of this, never a class within it
```

**Why classes rather than one method:**

> **An ops struct with optional members can grow without breaking existing implementations. A
> one-method interface cannot grow at all.**

Which matters more here than usual, because the primitive list is the one surface selection cannot
correct. The class shape does not retire that warning — it makes being wrong **recoverable by
addition** rather than by rewriting every driver.

**And an absent implementation is not a declaration**, which is what lets capability vary without
reintroducing what was dissolved:

```text
A DECLARATION   "this host supports audio." a claim, unverified, made by a registrar.
NOT IMPLEMENTED the driver does not fill that class. A BUILD-TIME FACT. nothing asserted.
```

Capability lands at **interface granularity**: a TV's driver implements display and not storage.
Calling into a class it does not implement fails loudly, the same requirement projection already
carries.

### Why spawn alone is not enough

Spawn is expressively sufficient and **cost-catastrophic**:

```text
read a file    spawn a reader.  a process per read.
get the time   spawn `date`.    absurd.
```

Linux has `read()` as a syscall rather than *spawn a reader* for exactly this reason. Reading a byte
and starting a process are not the same kind of operation, and one interface treating them alike
will be routed around immediately.

### Network is not a class

It is **transport** — how anything remote is reached, not something the kernel invokes.

```text
"fetch this URL"   spawn an HTTP client (EXECUTION), or invoke an HTTP entity.
reaching host B    transport. beneath everything, not beside it.
```

Made a class, it absorbs both, and the kernel starts making network calls that are really
invocations wearing a different hat.

### What to instantiate

**Execution now. Storage soon** — *read a file* becoming *spawn a reader* is the failure that forces
the second one.

**The others are not named yet.** The warning applies per class: each is a hand-designed surface, and
an empty `display` interface is a guess about display nobody has been forced to make. The *pattern*
is what is adopted here; a class earns its place when something cannot be done without it — the same
admission rule the design uses everywhere else, applied to its own plumbing.

**Where entities sit**, since the diagram omits them: entities are **above** this, in the model. You
invoke an entity; the kernel reaches it by spawning through the execution class on some host. The
diagram is plumbing beneath the model, and the two compose rather than compete.

### Ports are returned, never specified

```text
SPECIFIED   the caller says "listen on 8080." it now knows an implementation detail of what it
            spawned, and has to avoid colliding with everything else.
RETURNED    spawn hands back a HANDLE. how to reach the thing is inside it, and the caller
            never learns what a port is.
```

Which is what `WorldRef{Host, Handle}` already does — opaque, host-scoped, and resolving it
elsewhere is an error rather than a lookup. Whether a handle is backed by a port, a socket path, or
a pipe is the platform layer's business.

**The one place a port legitimately appears is inside an expectation.** If something must listen on
8080 because a human configured a firewall around it, that is `--port 8080` in the expression — a
requirement about the world, so it is **content**, not structure.

The test is the usual one: does anything in the model *read* it? Selection does not, credit does
not, discovery does not. So it is not a field. And port exhaustion is a spawn failure, which is why
the interface can stay at one method — everything that can go wrong there shows up as spawn
failing.

**`here` is the axiom.** The host the runtime is executing on is given, not discovered. Everything
else is reached from it — the same shape as the standing root: the thing you cannot derive because
you are standing on it.

### You invoke an entity, not a host

Delivery picks a host on the way out; the response correlates by **invocation identity**, not by
channel. A human messaged on their TV and answering on their phone is not two events — it is one
invocation arriving back through a different door.

**Addresses are not graph structure.** They are volatile and the graph is append-only, so a durable
record of where something *was* reachable is wrong within a week and permanent. Entity → address →
network → host is real, mutable, rebuildable from configuration, and belongs to the runtime.

## You cannot invoke your own grader

The rule was first written as *"a human is not an entity."* Too narrow — a human and a model are
interchangeable as interpreters, and *"ask Michael to find a utensil"* is a human being invoked with
nothing wrong about it. What must never happen is sharper:

> **You cannot invoke the thing that will judge your return.**

The same human is a parent in one place and an interpreter in another. **The constraint is on the
relationship, not on the species.**

```text
DOWN   invoke.           interpreters. capability. reach.
UP     return, or trap.  the grader. never invoked from below.
```

### It is enforced by scoping, not by a rule

A frame contains what was passed into it. The root's caller never passes **itself** down, so it is
not in any frame beneath it, so nothing below can name it. Within one chain the grader is unreachable
**because nobody put it in scope** — not because anything forbids it.

> **The organism has no concept of human — only of graders and interpreters.**

## There are no declarations

An earlier draft had entities declaring what they read, how much they could take, and what
modalities they support. **All of it dissolves, for one reason:**

> **A declaration is made by whoever REGISTERED the entity, not by the entity.** A third-party
> assertion about a capability nobody verified.

That is worse than the expectation-side language tag, not better — at least an author knows what
they wrote. So:

```go
type Entity struct { ID EntityID }
```

`Host` keeps one method — see above — but declares nothing either.

**History replaces every one of them.** *"Python to this entity worked"* is recorded per invocation
and grounded. *"This entity reads Python"* is somebody's guess at registration time.

```text
COLD START IS THE WHOLE COST   send Python to a TV, it fails immediately, one record, no model
                               call. compare a wrong declaration, silently believed forever.
```

**What the return carries instead** is what was *actually done* — rendered as audio, on host X —
rather than what was claimed possible. That accumulates into exactly the knowledge the declarations
were pretending to have, except earned.

## Assemble and project

Two operations, two bounds, two owners:

```text
ASSEMBLE   engram state → a frame.       bounded by the ENTITY. how much it can take.
PROJECT    frame → the host's medium.    bounded by the HOST. what form it can realize.
```

Context window is Claude's wherever it runs. Screen-versus-speaker is the device's whoever is using
it. Software entities are the degenerate case where every host is text over a wire, which is why
projection *looks* entity-shaped and isn't.

**`assemble`, not `select`.** The emission `select` is invocation; this is internal frame assembly at
a different level, and the set already tracks one naming collision it failed to fix in time. The old
lineage's *assembly walk* was exactly this operation.

### Projection must fail loudly

The single implementation requirement everything here rests on:

```text
LANGUAGE MISMATCH   a parser errors. LOUD.
MODALITY MISMATCH   audio to a device with no speaker — nothing happens, the invocation
                    "succeeds," AND NOBODY PERCEIVED ANYTHING.
```

The second is the silent-wrong-answer class this design refuses everywhere. A projector that
silently drops an unsupported form turns an immediate loud failure into a late invisible one, and
the record would read as success.

### Delivery fails immediately; work fails late

Projection happens **at dispatch**, not at return:

```text
DELIVERY   could this be presented at all?   IMMEDIATE. the possibility never opens.
WORK       was the answer any good?          however long it takes.
```

So an undeliverable member returns instantly and **does not block its group** — which matters,
because the group barrier is what would otherwise make this expensive. You do not wait on a human
reading a TV to discover the TV had no speaker.

## Authority follows the select

A page renders on your screen. Someone else wrote it. Whose is it?

```text
THEIR CONTENT     data. arrives as an observation.
YOUR RENDERING    an act on YOUR host, invoked by YOUR engram.
THE ENTITY        the browser. does the work. neither owns nor authorizes it.
```

**If a sender could render into your field directly, they would be choosing your coupling** — *the
coupling is chosen by the agent, never by the data*, at the perceptual layer. So an observation may
carry a modality **hint** and may never determine one; otherwise anything that can put text in front
of you could decide to speak instead.

Reach is what makes this enforceable rather than aspirational: rendering into your host requires
reach to your host, and that is yours to grant. A notification arriving unasked is not a
counterexample — it is a capability granted **once**, at admission, being exercised.

## The OS is beneath the kernel

Neither entity nor host. The **platform layer**:

```text
MODEL       entities (who interprets) · hosts (where it lands)
PLATFORM    os_write · os_spawn · os_open — normalized, per-OS adapters. RUNTIME CODE.
```

Reaching a human through their phone is three roles and only two are modelled: the OS is
**delivery**, the human is the **interpreter**, the phone is **where it lands**.

```text
TRUSTED BASE   the kernel and the platform layer. hand-written, reviewed, never learned about.
ABOVE IT       everything. selected, recorded, starved if wrong.
```

**Variation the platform layer cannot learn about surfaces one level up.** `os_spawn` being slow on
box-A is not something the adapter adapts to — it appears as invocation history on host-A, which is
where hosts are modelled.

```text
WATCH   the os_* list is the ONE hand-designed surface in the system. too narrow and things the
        runtime needs are unreachable; too wide and the kernel starts knowing about Linux.
        everything else here is derived or learned; this is chosen, and a wrong choice cannot
        be corrected by selection.
```

**And `spawn` is not `os_spawn`.** Opening a possibility is a kernel allocation; spawning a process
is a platform call. Different layers, and they must not share the word in code.

## There is no ground truth

> **Truth is derived, not stored — and it is derived BY THE OBSERVER.**

An entity's expectation *of itself* and a discoverer's expectation *of the entity* are one structure
held by two parties — **not two claims about a third thing.** Positing a hidden real capability that
both approximate is a stored fact nothing can ever check, which is worse than the stored facts
already refused, because it is unfalsifiable by construction.

**What follows makes the mechanism smaller.** An observer only ever holds its own expectations and
its own observations. It cannot compare another party's claim to that party's reality, because it has
no access to the interior.

**Discovery is therefore symmetric.** Two structures meet; each derives locally. No privileged
discoverer, no discovered object, nothing forcing convergence. They may diverge permanently and
neither is wrong, because "wrong" would require the third thing.

### Each entity owns its own engrams

There is no global graph. Follow any engram up and you reach a root, and that root belongs to an
entity.

So when I invoke Michael, **the engram is mine** — my record of *"when I asked Michael to find a
utensil, here is what came back and whether I used it."* Michael has his own, which I never see. One
event, two local records, neither authoritative over the other.

Which keeps the credit walk local: `Producer` names an engram in **my** tree, never Michael's
internals. The walk descends my own memory and stops at the crossing — the same frame boundary as
parent-and-child, one level up.

## The world is the residue of what has been exercised

Finding an entity is a possibility whose return happens to be an entity. There is no second
primitive, no registry, and no map.

> **An entity can only find other entities through the interaction of possibilities.**

The graph of who-can-reach-whom is not a structure anyone maintains. It is what is left behind by
what has actually been done.

---

## Rejected

```text
ENTITY AS A PLACE        one word for WHO and WHERE. only the first changes what is produced.

A SHARED SURFACE TYPE    grouping entities by how they are reached — a "model surface," a
                         "python surface." Claude and GPT do not share one, and neither do
                         python3.4 and 3.5. how each is reached is runtime configuration.

ADDRESSES IN THE GRAPH   volatile facts in an append-only store accumulate stale entries
                         forever. a category error, not a preference.

HOSTS IN THE INVOCATION  if one entity answers differently on two hosts, that is evidence it is
KEY                      UNDER-SPECIFIED, not a reason for a column. splitting produces two
                         histories; a field produces one nobody folds over.
                         (EXCEPT where the host is the TARGET rather than the interpreter —
                         two TVs are two hosts and turning on the wrong one matters. see
                         interfaces.md.)

A WORLD OBJECT / MAP     reach is the residue of exercise. a maintained map would be a second
OF AGENTS                place for that answer to live and a second place for it to disagree.

POSSIBILITY AS GROUND    a hidden real capability that advertisement and expectation both
TRUTH                    approximate. unfalsifiable by construction.
```
