# Primitives

**Date:** 2026-08-08
**Status: WORKING.** Reasoned from first principles in session. Nothing here is settled enough to
build from, and much of what it replaces is still standing in `2026-08-06`.

> **Two forms — SELECT and RETURN. Everything else is a name.**

---

## The layering

```text
DURABLE      +
INTERACTION  =
EMERGENT
```

A series of interactions over durable structure produces the behavior you see. The emergent layer is
**not represented anywhere** — it is what the other two do, not a thing that is stored.

The primitives below are the interaction layer. It is deliberately ignorant: every distinction
requiring judgement belongs above it.

## The grammar

A space emits. It does not spawn, and it has no API.

```text
<select cite=obsID><possibilityID> obsID, obsID, obsID </possibility></select>

<return cite=obsID,obsID><observationID></return>
```

```text
SELECT   name a possibility, bind observations to it.   AN INVOCATION.
RETURN   name one observation.                          THE VALUE UP.
```

That is the entire action vocabulary. **Select IS invoke** — there is no separate call form, which
makes *"selection happens at the call site"* literal rather than descriptive.

### Every emission cites its sources

A select has **two positions, and they are different roles**:

```text
<select><possibilityID> obsID, obsID </possibility></select>
         ^^^^^^^^^^^^^  ^^^^^^^^^^^^
         WHO            WHAT
```

An observation that supplied the *name* is never bound as an argument and never composed, so
without a citation it falls through as unused — and that is the observation that made the
invocation possible at all.

`cite` fixes it in the grammar rather than in the metric. It also buys something larger: **capability
provenance is exact.** *How did you come to be able to name this?* is answerable for every invocation
the system ever made. Derived attribution cannot do that — two observations carrying the same ID are
indistinguishable, and you cannot tell which grant was exercised.

A name may also come from durable structure rather than an observation, and the distinction is a
signal:

```text
CITED FROM AN OBSERVATION   I had to discover this before I could act.
CITED FROM DURABLE          I already had the structure.
```

A possibility that must discover before every action has internalized nothing, and the ratio between
the two measures how well-structured it has become.

### The five questions an emission answers

```text
WHO     the possibility.                    named.
WHAT    the observations.                   bound as arguments.
WHERE   the entity.                         named.
WHEN    the tick of the expectation.        local.
WHY     what prompted it.                   the citation.
```

**`WHERE` resolves the entity question.** Entity sits in every signature identically, which is why it
never needed discovering — the caller supplies it per *invocation*, so the same possibility runs in
different places without being a different possibility. It was never about the callee.

**`WHEN` makes time per-cell**, which lands proper time — ownerless since the Slice died — and keeps
it consistent with locality: no global clock, no component holding a model of the whole. **The cost
is that events cannot be globally ordered.** Cross-cell causality is reconstructible only through
citation chains, never through timestamps.

**`WHY` is split**, and collapsing it would lose the half that matters:

```text
WHO ASKED       the invoking possibility.   structural. free — the kernel knows the emitter.
WHAT PROMPTED   the observation cited.      emitted. the capability trail.
```

Because it is a grammar rather than an interface, the incapacity is a **parsing** property, not an
API property. A space cannot do anything but emit these two forms, and nothing else is expressible.

### Everything is by ID

A space holds **names, not content**. It can only name what is in frame: what it was passed, plus
what its own selects returned.

> **Capability is name resolution. You can select what you can name; you can name what you were
> given.**

This is what a second `entity[]` parameter was reaching for and does better. Reach does not need its
own type — it needs a scope. A space cannot address what was never put in its frame.

## What a space is

```text
pos "pos-name" (obs obs[], entity) {
    // collection behaviour — the while loop
    // return behaviour — one shot
}
```

```text
THE POSSIBILITY    the function name. durable.
THE OBSERVATIONS   the parameters.
THE ENTITY         where this runs. same slot in every signature.
THE BODY           TWO BLOCKS. a loop that selects, and a one-shot that composes.
THE RETURN         one observation. the completed work.
```

**Every possibility has this signature.** Not one per callee — one, universally. So there is no
per-possibility contract to discover, nothing for discovery to return beyond the name, and no
compatibility check before invoking.

> **Type safety is replaced by learning.** Anything may be invoked with anything; whether that was
> useful is what `used/seen` and credit measure.

**A space knows three things:** the observations it was given, what kind of possibility it is, and
the observations its children returned. Nothing else is in scope.

### Two phases, and they are not `Perturb` / `Observe`

```text
GATHER      the sequence of selects.       SELECTION is the skill.
SYNTHESIZE  the terminal composition.      SYNTHESIS is the skill.
```

Both may be code. Both may be a model call. **Both are owned by the durable cell**, because both are
things a possibility gets better at.

The resemblance to the phase split is close enough to be worth refusing explicitly:

```text
PERTURB / OBSERVE   INTERLEAVED — perturb, descend, observe, perturb. about PERMISSION.
                    structural: two methods, two API views.

GATHER / SYNTHESIZE the boundary is at the END OF THE BODY, not between children. about SKILL.
                    descriptive: a sequence of emissions whose last one composes. no machinery.
```

`Perturb` was `Observe` with `realized = nil` — a name for the first iteration of a loop that had
already started iterating. What the split was carrying, and what now has to pay for it:
[traversal.md](traversal.md).

## Entity

```text
ENTITY   a place of execution. WHERE something can run.
```

Not a taxonomy of things in the world. The question it answers is *where can I execute this*, which
is mechanical and therefore belongs in this layer.

**`here` is the axiom.** The device it is executing on is given, not discovered. One primitive
entity; everything else is reached from it — the same shape as the standing root, the thing you
cannot derive because you are standing on it.

### You cannot invoke your own grader

The rule was first written as *"a human is not an entity"* — the human is always upstream, and if it
were reachable the agent could dispatch work to its own judge.

**That version is too narrow.** A human and a model are interchangeable as bodies: *"ask Michael to
find a utensil"* is a human being invoked, and nothing is wrong with it. What must never happen is
narrower and sharper:

> **You cannot invoke the thing that will judge your return.**

The same human is a parent in one place and a body in another. The constraint is on the
**relationship**, not on the species.

```text
DOWN   invoke.           bodies. capability. reach.
UP     return, or trap.  the grader. never invoked from below.
```

A clarifying question is not the agent calling its parent — it is the agent yielding, and something
upstream possibly answering. The grader sits on the upward channel only, so it never appears in a
frame beneath it and the hierarchy cannot be inverted from inside.

> **The organism has no concept of human — only of graders and bodies.**

## Possibility

Entities have possibilities. **A possibility has the same structure as an expectation.** It is not a
different substrate underneath.

```text
THE ENTITY'S       its expectation OF ITSELF.       local. its own.
THE DISCOVERER'S   its expectation OF THE ENTITY.   local. its own.
```

One structure, two holders. **Not two claims about a third thing.** An advertisement is not a claim
*about* a capability — it is the entity's own expectation, made visible. That is why it has the same
shape. It is one.

### There is no ground truth

> **Truth is derived, not stored — and it is derived BY THE OBSERVER.**

Positing a hidden real capability that affordance and expectation both approximate is a stored fact
nothing can ever check. Worse than the stored facts already refused, because it is unfalsifiable by
construction.

**What follows, and it makes the mechanism smaller:** an observer only ever holds its own
expectations and its own observations. It cannot compare another party's claim to that party's
reality, because it has no access to the interior. An overclaiming entity is caught because *my*
structure — which its advertisement helped build — keeps failing against *my* observations. Its
inner state never enters the computation.

**Discovery is therefore symmetric.** Two structures meet; each derives locally. No privileged
discoverer, no discovered object, nothing forcing convergence. They may diverge permanently and
neither is wrong, because "wrong" would require the third thing.

## Discover

**Discover returns. It does not produce.** Produce is separate.

Discovery is not a lookup, and this is why it needs a body rather than a signature:

> A flat list is raw material. *"There is no possibility here that does X"* is a CONSTRUCTED claim —
> it requires knowing what was sought and judging that nothing met it. A function returns a list.
> Only a space can return a diagnosis.

So Discover is a **self-contained possibility**. Its work:

```text
QUERY    run the index. candidates out.
SELECT   the grammar, not a child. judge the candidates — pick one, or reject all.
CREATE   licensed ONLY by select's rejection.
```

Create is not a peer that may fire whenever. It is conditioned on select having looked and refused,
which is what stops fabrication on every miss — and cold start means nearly everything misses.

### Two mechanisms behind one call

```text
LOCAL    what do I already have that bears on this.   an index over accumulated structure.
REMOTE   what can this entity do.                     ask it. the index cannot know.
```

An index holds only what a perspective already accumulated, so it structurally cannot find what has
never been interacted with. *"Discovery is purely reverse index"* is true for the first and
impossible for the second.

### Finding an entity is not a separate primitive

Finding an entity is a possibility whose return happens to be an entity. No second call.

> **The entity can only find other entities through the interaction of possibilities.**

No registry, no map of the world. The graph of who-can-reach-whom is the **residue of what has been
exercised**, never a structure someone maintains.

### The empty case is the bootstrap

**Nothing needs to exist before the system starts, and no observation is required to begin.**

```text
expect "draw a picture" (obs null, entity)
```

Traced at n=0: discover queries an empty index and finds nothing. Select has nothing to select and
rejects. **Create is licensed by that rejection** — so the first invocation always creates.

> **The seed is the NAME, not an observation.** It is the query for discovery and the specification
> for creation, both. Observations are refinement, not requirement.

This is what answers the objection that conditional childhood cannot bootstrap — that nothing
attaches until something works, and nothing works because there is no structure to gather from. The
empty case was already covered by create-on-rejection; it just had not been traced.

### All ingress goes through discover, which is why nothing can be dropped

**Decided 2026-08-09.** Three separate failures traced to one precondition — *an observation in
frame is bound, cited, or routed to discovery, never silently dropped* — which was stated as a rule
with nothing enforcing it.

It does not need enforcing, because there is no path by which an unhandled observation exists.

```text
A CALLER INVOKES discover(payload, entity)
        ↓
discover QUERIES · SELECTS or REJECTS · CREATES ON REJECTION
        ↓
the payload is placed, or it is the seed of new structure. THERE IS NO THIRD OUTCOME.
```

An observation cannot sit unbound in a frame, because an observation does not *arrive* into a frame
— it arrives as **the argument to a discover call**, and discover's contract is to resolve it. The
rule becomes unexpressible rather than forbidden, which is the form this design wants.

**Any caller is an ingress channel.** A human, a webhook, a file watcher — structurally identical.
Nothing is special about the human except position.

### The grader is not nameable from below

*The human is a possibility* and *you cannot invoke the thing that will judge your return* look like
they collide. They do not, and the reason is the scoping rule already in place.

A frame contains what was passed into it. The root's caller never passes **itself** down, so it is
not in any frame beneath it, so nothing below can name it. Same person may be a body in one place
and a grader in another; within one chain, the grader is unreachable **because nobody put it in
scope** — not because a rule forbids it.

### System possibilities are names, not grammar

`discover` and `create` are reached through the same `<select>` as anything else — the way `write` is
a name you call, not a keyword. The grammar stays at two forms no matter how many system
possibilities exist.

**And that is what rescues create living inside discover.** The caller does not own Discover's body
and cannot rewrite it. `malloc` is not the program allocating memory; it is the program asking, and
the program cannot fabricate a page whatever it wants.

```text
THE TRADE   creation gated by a FIXED BODY is cheap and local, and cannot learn.
            creation gated by PRESSURE FROM ABOVE adapts, and regresses to the user.
```

Taking the first avoids a regress where every novel thing travels upward. It is paid for in
adaptivity, and that should stay visible rather than get absorbed.

## Composition

**The return is singular.** Many things go in; one comes out; that one may compose several but it is
its own thing.

A plural return makes every space a **pipe** — the parent sees its grandchildren's observations
directly, reaches past the child into its subtree, and the frame boundary means nothing. One composed
observation is opaque: the parent sees what the child *concluded*, not what it was told.

> **Synthesis is the job. A space that refuses to compress is plumbing.**

### It is compression, not conservation

Nothing is conserved — N in, 1 out, and information is destroyed on purpose.

```text
SEEN   parameters + every child's return
USED   CITED BY ANY EMISSION — a select or the return
```

**Two consumptions, and both count.** An observation used in a select shaped an *invocation*; one
used in the return entered the *composition*. Defining `used` as composition-only punishes exactly
the observations that made invocation possible — a discover result would be permanently wasteful,
which is what pulls toward smuggling scaffolding into the synthesis. It should not go there. It
should be counted where it was actually spent.

Unused then means genuinely inert: it changed neither what the space did nor what it said.

`used / seen` is a compression ratio. Low does not mean bad: it means the space looked at a lot to
conclude a little, which is a bad query, an over-provisioned frame, **or** legitimate exploration
that discarded correctly. A gauge, not a verdict, and it only means something differentially — the
same possibility over time, or two competitors on the same task.

### If the metaphor becomes energy, it is dissipative

Observations flow, they are the only thing that flows, and select/synthesize are the two operations
that shape the flow. Geometry is fair for that. **Conservation is not** — N in, 1 out, information
destroyed on purpose. A conservative reading reintroduces a law this design does not have, and sends
someone looking for where the missing observations went.

### The invariant that falls out

> **A space cannot originate an observation. It can only combine what it can name.**

Frame is what you were passed plus what your selects returned, so composition has nothing else to
draw on. Origination happens only where observations enter from outside: upstream, or a system
possibility that actually crossed to the world.

This is the anti-fabrication constraint arriving through **scoping** rather than through a type on
`obs`. It bounds **provenance, not validity** — you cannot cite what you never saw, but you can still
reason badly about what you did. That is what selection is for.

**Mechanically:** a composed observation records its direct constituents. Direct only — the full
chain is a walk, not a stored list, or an observation at depth seven carries its entire subtree.
Cheap, and it makes any conclusion traceable back to what entered from outside.

## The two skills, and why only one is graded

Both depend on the same two inputs, which is why they intersect:

```text
observations passed at invocation   ← from above
observations returned by children   ← from below
```

**One measurement at the parent, read two ways:**

```text
AGGREGATE over all inputs   → did I gather well.        THIS space's SELECTION.
PER INPUT, was it used      → was that answer worth it. THAT CHILD'S EDGE.
```

The second is the correction I would otherwise get wrong here, and it is settled below: a per-input
verdict does not reach the child. It lands on **the parent's edge to it**, so what improves is the
parent's selection, never the child's composition.

**Singular return is what keeps this honest.** A plural return invites padding — hand up ten things
so one gets used. One observation is used or discarded, so verbosity buys nothing. The residual risk
is *generic* rather than verbose, and the hierarchy absorbs it: a parent composing from vague inputs
produces a vague return, which its own parent discards. Pressure propagates rather than stopping at
the liar.

### Selection learns. Synthesis does not.

```text
SELECTION   graded by its children's efficiency, on edges it owns.   LEARNS.
SYNTHESIS   its verdict lands on someone else's edge.                DOES NOT.
```

Selection's loop closes locally — it passes observations, sees the efficiency, adjusts, and nothing
else has to change. **So learning starts at selection, and ends there.** All learning is topological:
what changes is which possibilities get invoked and under what conditions, never how any body
composes.

It also predicts where the system is thin. Selection is meaningless at a leaf — no children, nothing
to select among — so **leaves are pure synthesis and therefore learn nothing at all.** They are as
good as the model's priors on the day they were created, permanently. The root is nearly pure
selection and is where almost all the learning lives.

## Selection, and why the base layer needs no provenance split

There is no world-authored vs agent-authored distinction on observations. Provenance is a semantic
judgement and this layer makes none.

Fabrication is not prevented. It is **starved**:

```text
WHICH POSSIBILITY    what gets to run
WHICH OBSERVATIONS   what it sees — and therefore what it can name
WHICH ENTITY         where it runs
```

A child authors neither, and does not decide whether it is invoked again. **Selection happens at the
call site, not at the return** — a child controls what it says and still does not control whether it
is asked. The signal is the possibility and the observations passed into it; one that cannot adapt
under them is starved.

Pressure comes from above, and terminates at the user.

```text
STARVATION REQUIRES ALTERNATIVES   a parent with exactly one possibility for a need is not
                                   selecting, it is calling — and an unadaptable possibility
                                   with no competitor is invoked forever.
```

Competing siblings stop being a nice property of append-only and become the thing starvation runs on.

### The trade, stated plainly

```text
PREVENTION   failures are impossible.  bought with complexity.
SELECTION    failures are transient.   bought with dependence on pressure actually arriving.
```

The base layer does not have to enforce truth; it does have to let pressure **reach**. Two things
carry it: `used/seen` is a local discriminator that works at every level with no model call and no
upstream judgement — it measures waste rather than truth, but it ranks alternatives where nothing
else has arrived yet. And hop profiles mean acceptance is recorded **at every range**, so depth does
not attenuate the signal, it changes which part of the profile is populated.

What remains true is the limit below: a possibility at the terminus of accepted chains has no
pressure at all, because nothing above it can reject.

## The write

> **State is written by the parent, at the parent's return. It is the only write in the system.**

It has to be the parent's return, because that is the first moment `used` exists — during the body
the space does not yet know what it will compose from.

**Selects write nothing.** The body is pure gathering. The terminal composition emits one observation
upward and one set of writes downward.

```text
EVICTION SAFETY, FOR FREE   a space cut off mid-body has written nothing, so nothing happened.
                            the phase rule engineered this with a permission split. here it is
                            just where the information lives.
```

So the write covers the space's **whole life at once** — which children were invoked, what was
passed to each, which returns were used. An evicted space leaves no trace at all, which is what
*nothing happened* should mean.

### Invocation registers nothing. Being used does.

```text
CITED BY ANY EMISSION   → counts as USED.      the efficiency metric.
CITED BY THE RETURN     → becomes a CHILD.     durable structure.
```

The two thresholds are deliberately different. If citation-by-select made children, `expect` would
become a permanent child of everything that ever searched and the tree would be mostly scaffolding.
Composition-gating keeps it meaningful: find-utensil is cited by the picture and attaches; the
discover observation that *named* find-utensil is not, and does not.

> **The durable graph records what WORKED, not what was tried.**

Which retires the bloat problem rather than tolerating it. Exploration costs nothing permanent — a
space may try five things, use one, and only the one attaches. Append-only stops being a liability
because the filter runs **before** the write instead of never.

"Used in synthesis" means **cited by the return**, not textually embedded in it. Otherwise every
dependency that was transformed rather than quoted would be dropped.

### The local positive is structural. Everything else is a record.

```text
CITED BY THE PARENT      THE EDGE EXISTS. the attachment IS the signal — nothing is written.
ACCEPTED AT RANGE        an appended event on that edge, carrying its hop count.
NOT CITED / REJECTED     an appended record on that edge.
```

The first costs nothing to store because existence carries it. The other two are records, and they
are on the same object, which is what lets a body read all three together.

Earlier drafts framed this as *positive structural, negative evidential*. That was true when the only
positive was local. Long-range acceptance is also positive **and** is a record, so the cut is between
**local** and **at range**, not between good and bad.

### Everything durable is owned by an edge, and the edge is owned by the parent

> **A does not write to B's cell. A writes to A's own edge to B.**

An earlier draft put failure records on the child's cell and defended a cell being written by many
parents. Edge ownership means it never happens: no shared mutable state, no contention, nothing to
defend. **Locality becomes structural instead of argued.**

It also fixes an asymmetry that was backwards. With records on the cell, negatives kept their
conditions and positives lost them — and gather, which reads positives to decide *when* to invoke,
was reading the one signal stripped of context. On the edge, both are scoped to the relationship.

**What it commits to: synthesis does not learn.** If B's failures live on A's edge, B never sees
them. What improves is A's selection — *stop invoking B in these conditions* — and, when that is not
enough, A grows a better sibling. That is *refinement grows the asker, not the answerer* applied to
credit, and it means **all learning is topological.** No mechanism is needed inside a body, which is
consistent with bodies being code or a fixed model.

The cost is **no transfer**. A possibility that is bad for everyone is independently rediscovered as
bad by every parent. That is the price of having no shared state.

### The edge carries the return it cited

**Decided 2026-08-09.** An earlier draft kept only the fact of use, so an edge recorded that A used
B and never what B said. That made the system a complete record of **what it did** and no record of
**what it saw** — and two of the five toys fail on exactly that.

```text
STORED   the observation B returned, on the A→B edge, at the moment A cited it.
```

**What it buys.** Across-time comparison becomes possible at all. Ten stable readings can be
compared to an eleventh that differs, which is Toy 1's provenance condition. And *480 passed* can be
compared to *500 passed* — the **absence** case, where a thing stops happening, nothing is edited,
and no event fires to be piped in. No ingress mechanism reaches that; only stored history does.

### Store on change, not on every citation

Otherwise the positive channel acquires the unbounded growth that was the stated danger on the
negative one — content over instances rather than counts over types.

```text
IDENTICAL TO THE LAST   increment a count. store nothing new.
DIFFERENT               store it. this is a CHANGE EVENT.
```

The edge then holds a **change history** rather than a log, bounded by how often the answer actually
differs — which is the only part anyone wants. Ten identical readings record honestly as *stable,
n=10*; the eleventh that differs is a change, and a change is exactly the contested signal.

### It is history. It is never a cache.

> **A stored return may be read as WHAT B SAID BEFORE. It may never substitute for INVOKING B NOW.**

This is the way the decision could silently undo a core commitment. `closure.md`: *"has the door
usually been open"* is cheap, *"is the door open right now"* requires looking. An edge carrying B's
last return makes it trivially easy to read the answer instead of asking — which is the stored
present the design refuses, arriving as an optimization.

The rule wants to be structural rather than remembered, and nobody has proposed how.

## Credit

> **Credit assignment is a space, not a mechanism. It disperses on every close, and it lands on
> edges.**

### Nothing is ever paid to a space

Spaces are ephemeral and hold nothing durable, so no space can receive credit — including a credit
space. There is no regress to stop, and no exemption is needed for one.

### Every close disperses. The root is not special.

A space closes by synthesizing, and synthesizing **is** rendering verdicts on its children — it
decided what to cite. So every close already holds a full set of local verdicts, and dispersal is
the same recursive operation at every level. The root is the outermost instance, not a different
mechanism.

```text
ERRORS        stay LOCAL.         the immediate parent's verdict, on its own edge.
ACCEPTANCE    WALKS DOWN by hops. an ephemeral credit space walks the dependency chain.
```

**A chain that never reaches its root is therefore not ungraded — it is locally graded only.** Deep
work still receives verdicts from every level above it that closed. What is missing is the
longest-range grounding, not all grounding, so the system runs on partial global results.

### Why the asymmetry is not arbitrary

```text
FAILURE   "somewhere along here it broke."   DOES NOT DISTRIBUTE. at most one edge did it, and
                                             marking all of them asserts more than is known.
SUCCESS   "this whole chain worked."         DISTRIBUTES. every edge genuinely participated in
                                             something that worked. no attribution problem.
```

Conjunctive success distributes; conjunctive failure does not. Descending a *rejection* would be
Toy 2 in the credit channel — one evidential origin becoming N records — and the design already
refuses that shape. Descending an acceptance is not the same claim.

There is a consequence argument underneath it. A false positive is self-limiting: a useless edge
picking up credit from a chain that worked anyway will also sit in chains that fail. **A false
negative is not.** An edge punished for another's failure gets selected less, and *being unselected
is absorbing* — nothing can earn its way back if it is never tried. In a design where starvation is
the correction mechanism, distributing negative evidence is structurally more damaging than
distributing positive.

**What a rejection therefore means** is *"somewhere along this chain it failed"* — not *"E is bad."*
Which is exactly what licenses the parent to discover an alternative rather than to blame anyone, and
is the refinement move arriving at credit.

### Keyed by span, and conditional childhood is what makes that cheap

An edge exists only if a return cited it, so **the edge set for a span is already the causal
skeleton** — everything tried and discarded never became an edge. The credit walk therefore
traverses only what actually contributed, with no filtering step of its own.

That is a payoff from a decision made for an unrelated reason: the filter that keeps the graph clean
is also what makes the walk cheap and correct.

**The verdict is not owned by the span.** It was considered — one record, N edges joining against it
— and rejected: a body sees everything on its edges and nothing else, so a join against a span
verdict would make long-range standing a *global read during gather*. The walk writes to each edge
instead. N writes is simply what the walk costs, and it costs no model call.

**Span is the one global handle in the design.** Names resolve in frame, ticks are per-cell, writes
are local — a span crosses cells by construction. *Learning is local writes* still holds, because
the span is an address and the writes remain local. Something must mint span IDs, and that is the
outer loop.

### The edge stores EVENTS, not a value

> **An edge records the accepted event. Magnitude is derived by counting.**

That keeps it append-only without qualification — a counter is a mutation, an appended event is not.
Nothing has to be rebuilt, cached, or reconciled.

```text
KEY     the accepted event
VALUE   the HOP COUNT — how far up the acceptance originated
```

Three reads from one record shape:

```text
THE OBSERVATION   the condition. what prompted this selection — (Q → P).
THE HOP           the range at which it was accepted.
THE COUNT         the magnitude, by summing.
```

**Trajectory comes free.** The log is ordered, so *recent* versus *ever* is a comparison over the
same structure — no windowing scheme, no retention policy. This is what a bare counter would have
destroyed: a count says how often and never whether something is getting worse, and an edge that
worked for a year and has failed for a month would read identically to one that was always mediocre.

**And set membership survives.** A count collapses it — you would know P was accepted forty times
and never *alongside what*. With events, `{A,B,C} succeeded` and `{A,B,D} failed` are both
recoverable, so a successful subgraph can be rebuilt after the fact.

That does not solve sibling blindness; single-failure attribution is still unavailable. It makes it
**approachable later rather than permanently foreclosed** — the set differences that counterfactual
methods run on are retained. Counts would have thrown them away.

### Hop profiles are the detector

The repeated visiting is not overhead — it is one fact per accepting ancestor, and they are
genuinely different facts:

```text
G cited by F               → G's edge:  hop 1
E accepts F, walks down    → G's edge:  hop 2
D accepts E, walks down    → G's edge:  hop 3
```

So the detector is a **shape**, not two numbers held against each other:

```text
{hop1, hop2, hop3}   validated locally AND at range.
{hop1, hop2, — }     good nearby, never confirmed further out.
{hop1, — , — }       cited by its parent, and nothing else has ever agreed.
```

The third profile is the self-sustaining subtree, and it reads off directly.

This is the answer to depth attenuation, and it is better than the problem it solved. Not one signal
weakening with distance — **signals at every scale.**

### It replaces four things that are no longer needed

```text
ELIGIBILITY TRACES      the edge written during the span IS the trace. already durable,
                        already keyed, already the right set. nothing transient to clear.
HELD-OPEN SPACES        the signal lands on the edge, which persists. the space's mortality
                        was never the problem — it was the wrong object to watch.
FORWARD-FACING WRITERS  same.
NODE TAGS               an optimization that avoids walking the chain and pays in lost
                        context. the walk is bounded by depth and cheap.
```

**A tag read at selection time would be a different thing and a dangerous one.** `search.md` already
forbids it: activation may order what already matched and may never multiply it, or what is active is
retrieved more, enters more plans, and stays active. A global goodness prior on selection is that
pathology renamed — and it falls the wrong way, since contradictory evidence is usually about
something not currently favored.

### Credit on selections, and why it is exact

A select cites **one** observation. A return cites **many**. That asymmetry decides how tractable
each is:

```text
SELECTION     ONE cause.    credit assignment is EXACT.
COMPOSITION   MANY causes.  credit assignment is the unsolved sibling-blindness problem.
```

The attribution problem that has no cheap answer for returns simply does not arise for selections —
because of arity, not cleverness.

**And it supplies the condition an edge lacks.** The edge says *A used P*. The citation says *A used
P **because of O***. Typed through O's producer, that is `(Q → P)`, which recurs and therefore
carries statistics where a raw observation pair never could.

> **The pairwise correlation table is not a separate mechanism. It is credit on selections**,
> derived from the citation already in the grammar.

### Pruning is behavioural, not structural

Append-only forbids deletion, so nothing is removed — the selection simply stops being made.

And it is finer than the edge. `(Q → P)` can be pruned while `(R → P)` is kept, so a possibility
that is wrong when one thing prompted it stays available when another does. Edge-level pruning would
discard that distinction.

### Three constraints

```text
PROPAGATION MUST BE     B's durable children include everything B ever formed, in every context.
SPAN-FILTERED           unfiltered descent sprays a verdict onto edges from unrelated
                        invocations. span is not only how the chain is found — IT IS WHAT
                        BOUNDS THE RECURSION.

VOLUME IS NOT THE       an edge deep in a chain is visited once per ACCEPTING ANCESTOR, so the
PROBLEM IT LOOKED       work is O(n · depth). that was called superlinear overhead and it is
LIKE                    not overhead — it is what POPULATES THE HOP PROFILE, one fact per
                        ancestor. and the walk costs NO MODEL CALL: it is graph traversal,
                        which is free in the only currency this system spends.

RANKING MAY NOT         a tokenized credit structure looks like it could BE the discovery
REPLACE THE INDEX       index. it cannot. an index built from accepted pairs contains only
                        what has already worked, so discovery could never surface anything
                        unproven — and unproven structure is where all new material comes
                        from. INDEX OVER NODES for reach; CREDIT FOR ORDERING; search.md's
                        rule composes them — order what matched, never multiply it.
```

### What this addresses

An edge attaches when cited, and citation is usage. Follow the chain up and **usage was the only
internal signal in the system** — so a subtree deep enough that no boundary signal reached it would
sustain itself on being followed, internally coherent and undetectable from inside. Not weak
correction at depth: an active reinforcement process that does not require correctness.

Long-range records are what break it. Accuracy still enters at only two places — an entity crossing
and a human keeping a change — but the acceptance walk is what carries those inward, and **the hop
profile is what makes their absence visible**: an edge holding `{hop1}` and nothing beyond it has
been followed and never confirmed by anything.

### The root closes after all — acceptance is citation

This section previously said the pattern does not close at the root, because nothing invokes the
root and so nothing grades its synthesis. That was wrong.

> **The human's synthesis is acceptance. Accepting a change IS citation; rejecting is not-citing.**

So the human is simply the outermost parent, running the same operation as every other parent with a
different surface. No new mechanism is needed at the entry point.

**And it is the good kind of signal, because it is observed rather than reported:**

```text
COMPUTED    the change was kept. the diff was not reverted. the file still stands.
NOT ASKED   nobody prompts the human to grade anything.
```

Which keeps the entry point consistent with the rest of the design — *a self-report is not
grounding* — and means the whole pressure model rests on something a runtime can actually see.

---

## What a body sees

**Decided 2026-08-09.** Everything on its edges — the successes and the failures.

```text
the child possibility's name
its stored returns — the change history, with counts
credit at SHORT range (was it cited) and LONG range (did the chain hold)
the cite-source of each selection — (Q → P)
```

So the prompt is: *here is what I have used, what it said, when it worked, and when it did not.* No
weights change. **The model sees a richer edge set over time**, which is what topological learning
means in practice.

### A body sees DOWN, never UP

Its edges point at its children, so it sees their history. Its own reception lives on its **parents'**
edges, which it cannot reach.

> **No possibility can know how it has been received.**

That is a safety property falling out of locality rather than being designed for: a thing that cannot
observe whether it is being cited cannot optimise for being cited. **The citation signal is
ungameable from inside.**

### Synthesis has no historical input at all

The one-shot block sees the frame and what was gathered. Nothing else — its own past returns live on
its parents' edges.

So composition quality is a function of the model's priors and the material handed to it,
permanently. **The system improves by changing what gets gathered, never by composing better.**

Stated explicitly because it is surprising, and because an implementation will otherwise assume
otherwise.

## Open

```text
EVIDENTIAL BLOAT        THE LIKELIEST PLACE THE BLOAT WIN LEAKS BACK. structural growth is
                        filtered — nothing attaches unless the return cited it. the NEGATIVE
                        channel has no such filter, failures outnumber successes heavily, and
                        the fastest-growing edges are on the most-invoked possibilities, which
                        are exactly the ones that must stay assemblable. counts over types
                        converge where content over instances does not, and the volume bound on
                        credit propagation helps — but neither has been checked.

THE UNATTRIBUTED TERM   every emission cites its sources, so any conclusion traces back to what
                        entered from outside. A MODEL CALL DOES NOT. its contribution appears in
                        no citation, so provenance is exact for observations and absent for
                        priors — in a design whose whole posture is suspicion of unfalsifiable
                        claims. nothing currently distinguishes an observation composed from
                        cited material from one substantially authored by the model.

HISTORY MUST NOT        the edge now carries the return it cited, which makes reading a stored
BECOME A CACHE          answer instead of invoking trivially available. that is the stored
                        present the design refuses. the rule is written; a STRUCTURAL form of
                        it — one that makes the substitution unexpressible rather than
                        forbidden — has not been found.

SIBLING BLINDNESS       exact for selections, unsolved for compositions. a rejected chain does
                        not say which of A+B+C caused it, and pairwise cannot see A+C working
                        while A+B+C fails. two independent routes reached this limit, which
                        suggests it belongs to cheap credit assignment generally.

VARIATION HAS ONE       competition selects among alternatives; it cannot run on a population
SOURCE                  of one. the only things that grow the population are placement failure
                        and ingress. a possibility that is the TERMINUS of accepted chains has
                        neither — nothing above it can reject, and nothing fails to place — so
                        it is unfalsifiable from inside. correction there is external and
                        discrete, not gradual.

BINDING                 the grammar shows the call and the return, but not where a returned
                        observation acquires an ID IN THE CALLER'S FRAME. something mints that
                        name and puts it in scope, or the next select has nothing to refer to.

SIGNATURES              `<possibilityID> obsID, obsID` is positional, so a caller must know the
                        callee's parameter shape. it can only have learned that from discovery —
                        so discovery returns SHAPES, not just IDs.

WHAT PRODUCE IS NOW     nothing registers on invocation any more, and creation is reached by name
                        through select. `produce` may not survive as a peer syscall at all —
                        what is left of it is the write at return, which the parent performs.

DOES "POSSIBILITY"      if the structure is identical to an expectation, the word may only be
SURVIVE                 marking WHOSE it is and whether it is visible — a perspective label, not
                        a type. the signature may never need it.

THE WORD "SYSCALL"      2026-08-06/syscalls.md uses it for process → runtime: blocked, exit, kill.
                        these are entity-level. the word is doing two jobs, and this set already
                        tracks one naming collision that was not fixed in time.

WHO MAY SELECT CREATE   create is licensed by discover's rejection. whether that is enforced by
                        scoping — create is simply not nameable outside discover's frame — or by
                        something else, is undecided.
```
