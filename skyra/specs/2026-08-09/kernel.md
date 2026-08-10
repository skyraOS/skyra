# The Kernel

**Date:** 2026-08-09
**Status: SPEC.** Not built.

> **The kernel holds everything that is happening. The graph holds only what happened.**

---

## The split

```text
KERNEL HOLDS   possibilities in flight · frames · entity bindings · buffered impulses ·
               the group barrier · roots and heads.        ALL EPHEMERAL.

GRAPH HOLDS    engrams · history · edges · credit.
               nothing that is HAPPENING. only what happened.
```

This is not a division of labour. **It is what makes three claims already made in this set true
rather than aspirational.**

**"The kernel writes nothing"** holds only under a strict split. If in-flight work lived in the graph
you would mark a possibility running and later unmark it — a durable write, performed by the one
component forbidden one. *Commit is the only durable write* stops being true the moment active work
is representable.

**Mint-on-closure needs it.** An improvised expectation existing in the graph before it closed would
be discoverable mid-flight, and something could be selected that has never proven it can close.
The split is what makes engrams a record of things that **completed**.

**Crash semantics collapse to nothing.** The kernel dies, everything in flight evaporates, and the
graph is exactly the set of completed commits. No half-written state, no reconciliation pass, no
recovery protocol — ever. One leak and a recovery story is owed for the life of the system.

It is also what makes hosts fungible. **The kernel holds the binding, not the host**, so a host dying
mid-invocation is a re-assemble elsewhere rather than a lost possibility.

## It facilitates a write; it does not write

Three roles, no overlap, one per layer:

```text
POSSIBILITY   emits COMMIT.               AUTHORS.
KERNEL        checks scope + ownership.   AUTHORIZES. writes nothing.
HOST / OS     the storage class.          PERFORMS.
```

**The claim survives only because storage is a driver beneath the kernel.** If the kernel held the
store, *writes nothing* would be a fiction — it would be the writer, with a policy of restraint. The
OS sitting underneath makes it an incapacity again, which is what every other constraint in this set
reduces to.

## Two authorizations, not one

The kernel authorizes twice, at two boundaries, and they are easy to collapse into one and wrong to.

```text
THE GRAPH BOUNDARY   emissions.            SCOPE + OWNERSHIP. arithmetic on the frame.
THE WORLD BOUNDARY   host class calls.     GRANTS. not derivable from the frame.
```

### The graph boundary is arithmetic

Capability is name resolution, so both checks are structural:

```text
SCOPE       was this ID in the emitter's frame?     select AND commit.
OWNERSHIP   is this engram's parent the writer?     commit only.
```

Total, cheap, no policy, nothing to configure. **For durable writes there is no third check** — the
frame is the whole authority, and needing more would mean something was made expressible that should
not be.

### The world boundary is not

An earlier draft stopped at the two above and claimed completeness. **That was wrong.** Name
resolution bounds what a possibility can *say*; it bounds nothing about what a spawned process
*does*. `spawn` reaches the world through the execution class, and no arithmetic on the frame has an
opinion about which files it opens.

```text
COMMIT A REF TO /etc/passwd     the frame decides. already handled.
SPAWN SOMETHING THAT WRITES     the frame is silent. THE GAP.
/etc/passwd
```

So there is a second authorization, and it lives at the **driver boundary** rather than at emission —
because the kernel cannot know in advance what a process will do, only gate the call that starts it.

## Grants flow down from the root

*Status: **DEFERRED**. The shape is settled; the vocabulary is not, and nothing needs it before the
first build has world effects at all.*

> **Authority is granted at the root of a chain and narrows on the way down. Nothing widens it.**

```text
root      granted   fs:write:/work/**  ·  net:out  ·  introspect:self
  └─ A    narrowed  fs:write:/work/a/**
       └─ B         CANNOT recover /work/b. the narrowing is one-way.
```

Which is not a new mechanism — it is the shape entity reach already has. `interfaces.md` records that
a parent hands down a **subset** of `Entities()`, *which is how a limb gets narrower authority than
what spawned it*. World grants are the same object with a different alphabet, and the kernel is the
right holder because it already holds the chain and the ancestry.

**Monotonic narrowing is the whole rule.** A select may drop grants; it can never mint one. So the
authority of any possibility is derivable from its ancestry, and the root is the only place authority
enters the system — which is also where the human is.

### A grant the platform layer cannot enforce is a declaration

The one thing that would make this decoration:

> **If the kernel checks a grant and then spawns a process the OS does not confine, the grant
> described nothing.** The process does whatever the OS permits, and the record reads as bounded.

That is the same silent-wrong-answer class as a projector dropping an unsupported modality. So a
grant is only real if it maps onto something the platform layer actually realizes — a restricted
user, a sandbox profile, a mount namespace. **The vocabulary of grants is therefore bounded by what
the platform layer can enforce, not by what is convenient to express**, and it is one more reason
the `os_*` surface is the one that has to be gotten right.

### What is undecided

Whether grants are checked per driver call, realized once at spawn, or both. Whether they name
resources (`/work/**`) or capabilities (`fs:write`) or both. And whether a grant is content or
structure — the port precedent in `world.md` says content is the default answer, and this may be the
exception, since **credit must never read it**. See [questions.md](questions.md).

## Monitor

The frame is **maintained**, not re-shipped.

```text
ASSEMBLE   FULL.   at invocation open.       engram state → frame.
MONITOR    DELTA.  at each group closure.    observation → frame'.
```

Monitor is the incremental form of assemble, not a new primitive — and entity-bounded for the same
reason: folding an observation into a Claude context and into a Python process are different
operations, and neither is the kernel's business.

**It is a subscription whose lifetime is the possibility's lifetime.** The kernel opens it when it
pairs engram with entity and closes it at commit. No new lifetime is introduced; it is the one the
kernel already owns.

### Entities are still stateless, stated precisely

> **An entity carries nothing across possibilities.**

Within one it obviously holds a context that grows — that was already true the moment the frame
shipped on every step. Monitor does not add the state. It moves who maintains it: **kernel push,
rather than the entity re-reading a full frame and diffing it itself.**

And the reception property survives on the scoping. Within one invocation an entity sees only what
that possibility sees — its own children's returns. **Down, never up.** Monitor cannot reach a
parent's edge because the subscription is dead before the parent ever forms one.

### Three constraints

```text
IT MUST NOT RETURN        a return value is an emission channel outside SELECT / COMMIT. the
                          grammar's incapacity is a PARSING property, and one method with a
                          return type ends it.

IT FIRES ON GROUP         not on member arrival. otherwise context is shaped by arrival order,
CLOSURE                   which is a scheduling accident — and it is partial completion through
                          the back door: the entity cannot ACT early, but its context was
                          already built A-before-B.

LEAVES NEVER RECEIVE IT   a leaf never selects, so there is nothing to subscribe to. monitor is
                          strictly an orchestration channel and goes quiet exactly where the
                          one-shot happens.
```

Members are pushed in **emitted order**, never completion order — which is what keeps the barrier
real and makes the sequence reconstructible from the durable record, so re-running a possibility
rebuilds the same context.

### The invariant

> **`assemble(frame)` and `assemble(frame₀) + monitor*` must produce the same context.**

If those can diverge, the entity holds something the kernel cannot reconstruct, and statelessness is
gone in practice whatever this doc says. It also supplies recovery for free: a host dies
mid-invocation, the kernel re-assembles in full from the engram and the returns it already has, and
the possibility continues.

**Monitor is an optimization over reconstructible state, or it is a liability.** Cheap to assert in
the implementation — build both and compare.

### What it is for

`spaces.md` records the tension: what a possibility sees is simultaneously the learning mechanism and
the per-step transfer cost, and the two pull against each other. **A delta protocol relieves it
rather than picking a side.** It is not required for the first build.

## Impulses

*Status within this doc: **PROVISIONAL**. It closes something open rather than adding surface, but
nothing has forced it yet.*

**No observation ever changes.** Identity is `(producer, span, position)` and append-only holds. What
changes is **the world behind a handle** an observation names — `world.md` already put duration out
there: *open the page* returns immediately with a handle, and the browser is world state.

So the diff is a diff of a **referent**, and the observation is what makes it nameable.

This is the missing half of a mechanism already present. `spaces.md` left the watch case as *a live
process is a crossing that streams; results arrive as impulses* — and never said where an impulse
lands. **Monitor is where.** An impulse folds into an entity's context mid-invocation with no select,
no new possibility, and no polling.

### Self-caused versus world-caused is structural

```text
"I noticed this changed, and I did not touch it."
```

The kernel **knows** — did this follow a commit of mine in this span, or did it not? The entity is
told rather than deciding, the same move as citation-derived-from-content.

### The one constraint

> **An impulse may only concern something already in frame. It cannot introduce a name.**

```text
CHANGE NOTICE     on a handle I already hold.     delegation, refreshed. fine.
NEW OBSERVATION   pushed at me unsolicited.       INJECTION — whoever can reach my entity picks
                                                  what I can touch.
```

That is the line `grammar.md` draws for reach arriving in the data channel, and it applies here
unchanged. Keep it and capability-as-scoping survives: the world may say what you hold has moved, it
may not hand you something new to hold.

### Two consequences

**Impulses land on the same boundary as monitor** — buffered to group closure. Otherwise the world
interrupts at arbitrary points in an entity's reasoning and arrival timing becomes semantic. There
should be one boundary in this design, not two.

**An impulse that reaches an entity must land in history**, or the invariant above is false, replay
diverges, and the entity holds something unreconstructible. Recording it is what keeps the diff an
optimization instead of a side channel.

Buffered impulses are active work, so they die with the kernel — and that is fine, for a reason worth
stating: **an impulse is a notice that a referent moved, and the referent is still moved.** The next
assemble reads current world state. It is a re-read, not a lost event. An impulse only needs to reach
history when the invocation it informed actually commits.

### What it closes

Staleness is currently unrepresentable — a handle in a frame may be dead and nothing in the design
can say so. See `interfaces.md`, *world reference lifetime*.

## The kernel is the only global thing

Which collides on its face with *no global clock, no component holding a model of the whole*. The
resolution is a boundary, not an exception:

```text
IT HOLDS     roots · heads · ancestry · bindings.        EXECUTION TOPOLOGY.
IT CANNOT    read content · judge · assign credit.       NO SEMANTICS.
```

**A scheduler's view, not a world model.** It knows *that* a chain runs and where its head is; it
cannot know what any of it means, because meaning lives in observations it cannot interpret and
credit lives on edges it does not write. **The locality commitment was about judgement.** Execution
topology was never local, and pretending otherwise means nobody can schedule.

The time claim survives intact: **knowing the shape of live chains is not knowing their order.** The
kernel can say two heads are alive; it still cannot say which tick came first, because ticks are
per-cell.

### Roots and heads

```text
MANY ROOTS        independent asks alive at once.
MANY HEADS/ROOT   the concurrent frontier — cardinality > 1 branches the chain.

    root₁ ──────┬── head
                └── head
    root₂ ────────── head
```

From that view, and from no other:

```text
A STALLED HEAD    has not moved.
AN ORPHANED ROOT  the asker is gone.
A RUNAWAY         depth or breadth past a bound.
```

**These are eviction decisions, and nothing else in the system can make them.** A possibility sees
down; an engram sees its own edges; no entity spans invocations. Eviction requires exactly the view
the kernel has and nobody else does. That is a justification for globality, not a convenience.

### The report goes to the root

*Parent pressure is ultimately the user*, so **the root of a chain is whoever asked.** Reporting head
state to the root is reporting to the asker — and it is looking *down*, so it does not touch
*no engram can know how it has been received*.

> **Introspection is kernel-mediated. It is bounded by the subtree, never by the grammar.**

An earlier draft said it must not be addressable from inside the grammar at all. **Too strong** — it
named a mechanism where the property was elsewhere. What the separation actually requires is
`DOWN, NEVER UP`, and asking about your own descendants is *down*.

```text
TO THE ASKER      out of band, on the channel the ask arrived on. the root IS whoever asked.

TO A POSSIBILITY  a system engram manned by the runtime entity — the same exemption discover
                  and create already hold — answering ONLY about its own subtree.
```

The kernel holds ancestry, so that bound is **exact and cheap**: it already knows which heads
descend from the caller. An orchestrator asking whether its own subtree is stuck is a legitimate
question and there is no other way to ask it.

**Two things it still cannot do**, and both are what mattered:

```text
UPWARD    a possibility cannot see its ancestors' state, so it cannot learn its own reception.
          the citation signal stays ungameable from inside.

SIDEWAYS  it cannot see other roots, or siblings in its own group. no coordination in flight.
```

**Who may ask is a permission**, and it is the introspection case of grants above — `introspect:self`
is a grant like any other, narrowing down the chain. Which is why the answer to *what may ask the
kernel what it is doing* is not a structural rule: it is the first grant the design actually needs.

## What it costs

**You cannot ask the graph what the system is doing right now.** Every operational question — what is
running, what is stuck, how deep — is a kernel-side facility, out of band, unavailable to anything
inside the model. This will be tempting to violate the first time someone wants a dashboard built
out of engrams.

**Sibling coordination becomes structurally impossible, not merely unimplemented.** Two members of a
group can learn nothing about each other while running, because the only place that information
exists is kernel-side and unaddressable.

That is believed correct — it is what a barrier means — and it is **decided here on purpose** rather
than discovered later. Note that it is *not* the sibling blindness in `questions.md §3`: that one is
about credit assignment across a composition after the fact, and it stays open. This is about
coordination during flight, and it is closed.

---

## Rejected

```text
ACTIVE WORK IN THE      a running flag on an engram. it is a durable write by the component
GRAPH                   forbidden one, it makes unclosed work discoverable, and it buys a
                        recovery protocol that is otherwise never needed.

MONITOR AS A            an entity accumulating observations across possibilities. it would gain
STANDING SUBSCRIPTION   state, and an entity seeing traffic that cites its own output can learn
                        its own reception — which is the one thing that makes the citation
                        signal ungameable from inside.

MONITOR RETURNING       a third emission channel, outside a grammar whose whole claim is that
A VALUE                 only two forms parse.

PUSHING GROUP MEMBERS   arrival order is a scheduling accident, replay diverges, and partial
AS THEY LAND            completion arrives as a side effect rather than as a decision.

IMPULSES CARRYING       the world handing an entity a name it was never given. injection —
NEW NAMES               whoever can reach an entity would choose what it can touch.

UNBOUNDED               live state reachable by select, over the whole graph. the subtree bound
INTROSPECTION           is what keeps DOWN NEVER UP; without it a possibility could read its own
                        reception. the engram itself is fine — the reach was the problem.

GRANTS AS DECLARATIONS  a permission the platform layer does not realize. the process does
                        whatever the OS allows and the record reads as bounded — the same silent
                        success a projector produces when it drops a modality.

GRANTS THAT WIDEN       a select handing down authority it did not have. authority enters at the
                        root and narrows; anything else means ancestry stops determining reach.
```
