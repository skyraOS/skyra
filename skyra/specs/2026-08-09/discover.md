# Discover

**Date:** 2026-08-09
**Status: SPEC.** Not built.

> **All ingress goes through discover, which is why nothing can be dropped.**

---

## It needs an expectation, not a signature

Discovery is not a lookup:

> A flat list is raw material. *"There is nothing here that does X"* is a CONSTRUCTED claim — it
> requires knowing what was sought and judging that nothing met it. A function returns a list. Only
> a possibility can return a diagnosis.

## One call, four blanks

An invocation needs four things. `discover` fills in whichever are missing:

```go
discover(obs []Observation, engram *EngramID, entity []EntityID, address []Address)
```

```text
engram  nil   WHAT TO DO
entity  nil   WHO DOES IT
address nil   WHERE
obs     nil   WITH WHAT
```

A complete decomposition of an invocation — and the fourth is what earns the unification.
`discover(obs: nil, engram: X)` is *"what do I have that X needs?"*, which is **retrieval**. The old
set carried that as a separate surface behind `blocked`; here it is the same call.

**Bootstrap is the maximally-nil case**, which is a good sign the decomposition is right:

```go
discover(obs: [{need: "draw a picture"}], engram: nil, entity: nil, address: nil)
```

The seed is still the name; it just arrives as an observation like everything else.

### Three states per parameter, not two

```text
nil               resolve freely.
A CANDIDATE SET   resolve WITHIN this. a constraint, not a blank.
A SINGLE VALUE    fixed. nothing to resolve.
```

Which makes **reach inheritable through arguments**. A parent narrows what a child may resolve to by
passing a set — capability-is-scoping expressed as data rather than as a second mechanism.
`Frame.Entities()` is the default range; a passed set is a narrower one.

### It returns an observation, not a completed invocation

The possibility then emits the select using it. Which matters for credit: **the select cites that
observation**, so discover earns credit exactly when its answer was used, through the mechanism that
already exists. No special case.

### Not one implementation — the implementations are engrams

`discover` is an **orchestrator**, and its children are resolvers:

```text
discover
  ├─ by-index          code. tokenize, query the reverse index.
  ├─ by-recency        code. what has been used lately.
  ├─ by-cooccurrence   code. what usually accompanies these observations.
  └─ create            authors a new expectation.
```

So discover's own history records **which resolver suited which kind of need**, accumulating on its
edges like anything else. The regress stops because `discover` is nameable in every frame — a
builtin name, not something you must discover in order to discover.

### Which dissolves a rule into a strategy

An earlier draft said *create is licensed **only** by select's rejection*, as the thing stopping
fabrication on every miss. It is not a system rule. It is **ordering inside a particular discover
engram's expectation** — try index, try recency, then create.

```text
BEFORE   a rule. asserted, unfalsifiable.
NOW      a strategy. embodied, and BEATEN BY A BETTER ONE if it is wrong.
```

Expectations are immutable, so a given discover has a fixed strategy, and one reaching for `create`
too eagerly is **outcompeted** by one that does not. The guard becomes testable rather than trusted
— and if eager creation turns out to be fine, the eager discover wins and we learn something.

Soft only at cold start: at n=0 there is one seed discover and its strategy is whatever was
hand-written.

## Address is what is genuinely multiple

```text
ENGRAM    genuinely multiple. many expectations can serve one need.
ENTITY    CONSTRAINED BY THE EXPECTATION. python source runs on entities that read python —
          one narrow family. NL runs on many. so it varies, and cannot be counted on.
ADDRESS   ALWAYS multiple. one entity, many hosts.
```

So the useful symmetry is **discover → engrams** and **resolve → addresses**, not entities. An
engram sits in one tree with one owner; the multiplicity it does have comes from what its
expectation is written in, not from ownership.

**Entity partitioning is already handled and needs nothing.** `Invocation.Entity` is part of what
makes two askings the same question, so one engram interpreted by Claude and by a human already
accumulates separate histories. Worth stating because *"the same engram"* reads as one thing and
isn't.

**And provisioning stays outside all of this.** Resolving among *existing* entities is free.
Registering a *new* one is new **reach**, and starvation does not correct reach — so it is gated,
once per capability, never per invocation. It is not a nil-parameter case.

## Two mechanisms behind one call

```text
LOCAL    what do I already have that bears on this.   an index over accumulated structure.
REMOTE   what can this entity do.                     ask it. the index cannot know.
```

An index holds only what a perspective already accumulated, so it structurally cannot find what has
never been interacted with. *"Discovery is purely reverse index"* is true for the first and
impossible for the second.

## Nothing can be silently dropped

Three separate failures traced to one precondition — *an observation in frame is bound, cited, or
routed to discovery, never silently dropped.* It needs no enforcement, because **there is no path by
which an unhandled observation exists.**

```text
A CALLER INVOKES discover(payload, …)
        ↓
discover RESOLVES — its resolvers run in order, and the last one creates.
        ↓
the payload is placed, or it is the seed of new structure. THERE IS NO THIRD OUTCOME.
```

An observation cannot sit unbound in a frame, because an observation does not *arrive* into a frame —
it arrives as **the argument to a discover call**, and discover's contract is to resolve it. The rule
is unexpressible rather than forbidden, which is the form this design wants.

**Any caller is an ingress channel.** A human, a webhook, a file watcher — structurally identical.
Nothing is special about the human except position.

## Refinement grows the asker, not the answerer

When a returned observation is not good enough, the possibility does **not** rewrite its own
expectation.
The failure is cited as the prompt for a new discovery, which may create a new possibility.

`find-utensil` stays generic and useful to everyone; the parent grows a more specific child at exactly
the point where generic was not enough. **Specialization accumulates where it is needed rather than
being pushed back onto whoever was asked.**

There is a cost ladder, and the cheap rungs come first:

```text
1  pass different observations to the SAME possibility     cheap. no structure.
2  select a DIFFERENT existing possibility                 cheap. uses structure.
3  discover → create a NEW one                             expensive. GROWS structure.
```

**Name the need, not the complaint.** The name is what gets tokenized for the index, so a new
possibility must be named for what it does — *find a red utensil* — never for the rejection that
prompted it. The negation is the `cite`, not the name.

**The cost:** child-list growth on the asker. Every successful refinement is a permanent child, so a
much-used parent accumulates many. It is the good kind of growth — everything there worked — but
gather has more to select among over time.

## System possibilities are names, not grammar

`discover` and `create` are reached through the same `<select>` as anything else — the way `write` is
a name you call, not a keyword. **The grammar stays at two forms no matter how many system
possibilities exist.**

And that is what rescues create living inside discover. The caller does not own discover's body and
cannot rewrite it. `malloc` is not the program allocating memory; it is the program asking, and the
program cannot fabricate a page whatever it wants.

```text
THE TRADE   creation gated by a FIXED BODY is cheap and local, and cannot learn.
            creation gated by PRESSURE FROM ABOVE adapts, and regresses to the user.
```

Taking the first avoids a regress where every novel thing travels upward. It is paid for in
adaptivity — **and that cost is paid down below, not absorbed.**

## The runtime is an entity, so system engrams improve

The fixed-body trade said discovery and creation cannot learn. They can, once the runtime is an
**entity** — an identity that mans system spaces and emits like any other.

### Two roles, and they must not touch

```text
AS KERNEL   pairs, drives, applies writes.   NOT SELECTABLE.
AS ENTITY   interprets system expectations.  SELECTABLE, and emits like anything else.
```

If those blur, a possibility that selects the runtime-entity reaches the **write path** — and could
influence what is recorded about itself. That is the one thing the honest-witness split exists to
prevent. Same software, two roles; the separation is the whole safety story.

### Its engrams are system space

`2026-08-06/execution.md`'s rule transfers intact, and was derived for a different architecture:

> **A system-space expectation may be ENTERED FROM user space. It may never be the standing root.**

Without it the agent spirals into self-maintenance — getting better at retrieval, grading how well it
got better at retrieval, and doing no work.

### Who grades it falls out for free

The invoker cites discover's return, or does not. That is the local verdict, on the invoker's engram,
by the same mechanism as everything else.

The old set needed a bespoke perceive here — *"did the process trap again for the same condition"* —
because retrieval had no natural grader. Here it just has a parent. **One fewer special mechanism.**

### What improves is competition, not any body

Expectations are immutable, so no discover engram gets better. What improves is **competition among
discover engrams**: different retrieval strategies accumulating different histories, with selection
moving between them. Same as everywhere.

### And the cost stays bounded

```text
THE QUERY    an index lookup. code. cheap. unchanged.
THE SELECT   judgment among candidates — ALREADY entity-interpreted, because select is the
             grammar and something has to man that space.
```

Discovery does not become an inference call on every lookup. The judgment was always there; making
the runtime an entity only means the engram wrapping it accumulates history like anything else.

**What to watch:** system engrams will be the most-invoked in the graph — every discovery, every
creation. So they are where evidential growth concentrates, and they are exactly the ones that must
stay assemblable.

---

## Rejected

```text
DISCOVER AS A PURE       a lookup cannot construct "there is nothing here that does X," which is
LOOKUP                   the claim that matters. it needs a body.

SELECT AS A CHILD        selection is the grammar — the act of choosing among known
POSSIBILITY              possibilities — not a possibility that gets invoked.

A SEPARATE ENTITY-       finding an entity is a possibility whose return happens to be an
FINDING PRIMITIVE        entity.

CREDIT AS THE INDEX      a tokenized credit structure looks like it could BE the discovery
                         index. an index built from accepted pairs contains only what already
                         worked, so discovery could never surface anything unproven — and
                         unproven structure is where all new material comes from.
```
