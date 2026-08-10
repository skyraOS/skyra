# Skyra — SELECT / RETURN and the Possibility Model

**Date:** 2026-08-08 · revised 2026-08-09 with the credit model

> **SUPERSEDED 2026-08-10** by
> [`2026-08-10-engrams-possibilities-and-the-interpreter-split`](2026-08-10-engrams-possibilities-and-the-interpreter-split.md).
> **The vocabulary here is obsolete and will mislead:** `possibility` meant the durable thing,
> `return` meant the value up, and entities were places of execution. All three changed. Kept as
> the working record of how the model looked before the rename.
**Scope:** Two working sessions, reasoned from first principles, plus one round of external review.
**Status:** WORKING. Nothing here is built. The live design remains `specs/2026-08-06`; this is
`specs/2026-08-08`, a whiteboard set. Written to be read cold, for critique.

---

## What the system is trying to be

An agent runtime built on one loop: something arrives, it is graded against an expectation, that
expectation determines the next action, that action determines the next expectation.

The commitments that constrain every decision, inherited and not re-argued:

```text
TRUTH IS DERIVED, NOT STORED   the graph holds structure and history, never a claim about the
                               present. "is the door open" requires looking.

APPEND-ONLY                    nothing is rewritten. a walk in flight cannot be standing on
                               something that ceased to exist.

SAFE BY INCAPACITY             constraints are held by removing the ability to violate them,
                               not by rules someone must remember.

NO SELF-REPORT AS GROUNDING    a thing's claim about its own success is an expression, not
                               evidence. something must independently look.

LEARNING IS LOCAL WRITES       caching may be global reads. no component holds a model of the
                               whole tree.
```

The prior design had a kernel with two traversal methods — `Perturb` (descend, read-only) and
`Observe` (ascend, write) — four named behaviors, and a durable graph of expectations. That kernel is
implemented and tested. **This work concluded the two-method interface should not survive.**

---

## The model

### Three layers

```text
DURABLE + INTERACTION = EMERGENT
```

The emergent layer is not represented anywhere. It is what the other two do.

### The grammar

A space emits. It has no API and cannot spawn. Two forms, and that is the entire action vocabulary:

```text
<select cite=obsID><possibilityID> obsID, obsID </possibility></select>

<return cite=obsID,obsID><observationID></return>
```

```text
SELECT   name a possibility, bind observations to it.   AN INVOCATION.
RETURN   name ONE observation.                          THE VALUE UP.
```

Because it is a grammar rather than an interface, the incapacity is a **parsing** property. A model
authoring a body can express these two things and nothing else.

Every emission **cites its sources**. A select has two positions — the possibility named (*who*) and
the observations bound (*what*) — and an observation that supplied the *name* is never bound and
never composed, so without citation it falls through as unused despite having made the invocation
possible. Citation also makes capability provenance exact: *how did you come to be able to name
this* is answerable for every invocation ever made.

Five things an emission answers:

```text
WHO     the possibility.               named.
WHAT    the observations.              bound as arguments.
WHERE   the entity.                    named — a parameter of the INVOCATION, not the callee.
WHEN    the tick of the expectation.   local. no global clock, no global ordering.
WHY     what prompted it.              the citation.
```

### A space is a function

```text
exp exp-name (obs obs[])  →  return obs
```

It knows three things: the observations it was given, what kind of possibility it is, and what its
children returned. Its body is a sequence of selects; its return is one composed observation.

```text
GATHER      the sequence of selects.    SELECTION is the skill.
SYNTHESIZE  the terminal composition.   SYNTHESIS is the skill.
```

Not a phase split in the old sense — the boundary is at the end of the body rather than between
children, and it is about skill rather than permission.

### Capability is scoping

A space holds names, not content, and can only name what is in frame — its parameters plus what its
own selects returned.

> **You can select what you can name; you can name what you were given.**

This replaced an explicit `entity[]` capability parameter. Reach does not need a type; it needs a
scope.

### Entity and possibility

An entity is **a place of execution**. `here` — the device the runtime is on — is the one axiom;
everything else is reached from it. There is no registry: *the graph of who-can-reach-whom is the
residue of what has been exercised.*

A possibility has **the same structure as an expectation**, not a different substrate. An entity's
advertisement of itself and a discoverer's expectation of it are one structure held by two parties —
**not two claims about a third thing.** There is no ground truth: an observer holds only its own
expectations and its own observations, and can never compare another party's claim to that party's
reality. An overclaiming entity is caught because *my* structure keeps failing against *my*
observations.

---

## The mechanisms

### One write, at the parent's return

State is written by the parent, at the parent's return — the first moment it knows what it composed
from. Selects write nothing.

Eviction safety falls out for free: a space cut off mid-body has written nothing. The prior design
engineered this with a read/write permission split; here it is simply where the information lives.

### Invocation registers nothing. Being used does.

```text
CITED BY ANY EMISSION   → counts as USED.      the efficiency metric.
CITED BY THE RETURN     → becomes a CHILD.     durable structure.
```

> **The durable graph records what WORKED, not what was tried.**

This retires a problem the previous design had conceded — that fabricated structure is permanent
under append-only. Exploration now costs nothing permanent; the filter runs before the write.

### Everything durable is owned by an edge, owned by the parent

> **A does not write to B's cell. A writes to A's own edge to B.**

An earlier draft put failure records on the child's cell. Edge ownership means no shared mutable
state and no contention — locality becomes structural rather than argued. It also fixes an
asymmetry that ran backwards: on a cell, negatives kept their conditions and positives lost them,
and gather (which reads positives to decide *when* to invoke) was reading the signal stripped of
context.

**It commits to synthesis not learning.** If B's failures live on A's edge, B never sees them. What
improves is A's selection, and when that is not enough, A grows a better sibling. All learning is
topological; no mechanism is needed inside a body. The cost is **no transfer** — a possibility bad
for everyone is independently rediscovered as bad by every parent.

### Credit

> **Credit assignment is a space, not a mechanism. It disperses on every close, and it lands on
> edges.**

Nothing is ever paid to a space — spaces are ephemeral and hold nothing durable, so there is no
regress and no exemption is needed.

**Every close disperses; the root is not special.** A space closes by synthesizing, and synthesizing
*is* rendering verdicts on its children. So every close already holds a full set of local verdicts,
and dispersal is the same recursive operation at every level.

```text
ERRORS   stay LOCAL.         the immediate parent's verdict, on its own edge.
CREDIT   flows BACK FURTHER. recursively, down the subtree that produced the observation.
```

A chain that never reaches its root is therefore **locally graded, not ungraded**. The system runs on
partial global results.

**Keyed by span** — and conditional childhood is what makes that cheap. An edge exists only if a
return cited it, so the edge set for a span is already the causal skeleton. Span is the one global
handle in the design; the writes remain local, so *learning is local writes* still holds.

**Records carry scope**, and that is the gain rather than the bookkeeping:

```text
SHORT RANGE   immediate · cheap · precise about attribution — but measures only USAGE.
LONG RANGE    delayed · coarse · no attribution — but actually GROUNDED.
```

Not one signal attenuating with distance — **signals at every scale.** An edge reading strong nearby
and weak at range *is* a self-sustaining subtree, visible with no additional detector.

This replaced four things: eligibility traces (the edge written during the span *is* the trace),
held-open spaces, forward-facing writers, and node tags.

Two constraints: propagation **must** be span-filtered, or a rejection sprays onto edges from
unrelated invocations — span bounds the recursion, it isn't only how the chain is found. And volume
is O(subtree) per close at every level, bounded informationally: stop descending when the verdict
does not change what is already recorded.

### Selection is the correction mechanism

Fabrication is not prevented, it is **starved**. A parent controls which possibility runs, what it
sees, and where it runs; a child authors none of these and does not decide whether it is invoked
again. **Selection happens at the call site, not at the return.**

Pressure terminates at the user. **The human's synthesis is acceptance** — accepting a change is
citation, rejecting is not-citing — so the human is the outermost parent running the same operation.
The signal is computed, not reported: a diff kept, a file not reverted.

The rule that keeps this from inverting is not about humans:

> **You cannot invoke the thing that will judge your return.**

A human is a parent in one place and a body in another. The constraint is on the relationship.

### Refinement grows the asker, not the answerer

A space does not rewrite its own expectation when a result is inadequate — that is self-modification
without external pressure, and it defeats starvation. Instead the failure is cited as the prompt for
a new discovery. `find-utensil` stays generic; the parent grows a more specific child.

```text
1  pass different observations to the SAME possibility     cheap. no structure.
2  select a DIFFERENT existing possibility                 cheap. uses structure.
3  discover → create a NEW one                             expensive. GROWS structure.
```

### The empty case is the bootstrap

`expect "draw a picture"(obs null, entity)` at n=0: discovery finds nothing, selection rejects, and
**create is licensed by that rejection.** The seed is the **name** — the query for discovery and the
specification for creation, both. Observations are refinement, not requirement.

---

## Why the two-method interface was dropped

```text
Perturb(origin, impulse)            → Space
Observe(origin, impulse, realized)  → Space
```

**`Perturb` is `Observe` with `realized = nil`** — the split names the first iteration of a loop that
had already begun iterating.

```text
EVICTION SAFETY        → falls out of the write happening only at return.
NOTHING FIRES IN-WAVE  → same.
API READ/WRITE SPLIT   → GONE. replaced by: there is only one write, and it is last.
POST-ORDER ORDERING    → survives.
RETURN PATH IS TREE    → survives and improves: typing the return as an observation removes the
                         continuation-vs-value ambiguity the two methods carried.
ONE-SHAPE DISPATCH     → survives. two emissions, no dispatch table.
```

---

## The sharpest criticism received, and what was done about it

External review made the point that the design ran on **usage as its only internal signal**. An edge
attaches when cited; citation is usage; a parent's own survival depends on being cited by *its*
parent. Follow that up and nothing internal measures correctness. A subtree deep enough that no
boundary signal reaches it would sustain itself on being followed — internally coherent and
undetectable from inside. Not weak correction at depth: an **active reinforcement process that does
not require correctness.**

The scoped credit model above is the response. Long-range records carry the boundary signals inward,
and the divergence between short-range and long-range readings makes their absence visible.

**It should not be read as a refutation.** Accuracy still enters at exactly two places — an entity
crossing, and a human keeping a change. The mechanism moves those signals around better; it does not
create more of them. Whether they propagate far enough is empirical and untested.

---

## What to attack

**1 · Nothing has been tested against a case.** All of this is derived. The `2026-07-27` set has five
toy scenarios whose job is breaking designs, and this was run against none of them. Toy 2 ("distinct
is not necessarily independent") looks most likely to bite, since correlation over possibilities may
be a fan-in count in new clothes.

**2 · The unattributed term.** Every emission cites its sources, so conclusions trace back to what
entered from outside — **except a model call**, whose contribution appears in no citation. Provenance
is exact for observations and absent for priors, in a design whose posture is suspicion of
unfalsifiable claims. Nothing distinguishes an observation composed from cited material from one
substantially authored by the model.

**3 · Successful content is destroyed.** An edge records that A used B, never what B said. The
design's answer is re-observe rather than recall, but an internal composition **cannot** be re-asked
— the body may be a model call and the frame is gone. A successful path is not replayable even in
principle, and this loss falls on the valuable side.

**4 · Credit volume.** Dispersal is O(subtree) per close at every level, superlinear in chain size
and worst where chains are deepest. The informational bound is plausible and unchecked.

**5 · Sibling blindness.** A rejected chain does not say which member caused it; pairwise correlation
cannot see A+B+C failing while A+C works. Two independent routes reached this limit, which suggests
it is a property of cheap credit assignment rather than of one implementation.

**6 · `create` is under-specified.** It is licensed by selection's rejection and lives in a
system-owned body, so the caller asks rather than makes — the `malloc` argument. But `malloc` returns
undifferentiated memory, and `create` must return something with *competence*. What authors the new
body, and from what, is unanswered.

**7 · Loose ends.** The relationship to the built kernel is unstated — it has `Perturb`/`Observe` and
this model does not, and nobody has said whether that kernel changes or this is a parallel track. Two
vocabularies coexist with no mapping. Binding — where a returned observation acquires a name in the
caller's frame — is unspecified.

---

## What would most change the assessment

```text
THAT SELECTION BOOTSTRAPS ALONE   its loop closes locally, so learning can start without
                                  synthesis improving — which, under edge ownership, it never
                                  does. if this is wrong, nothing learns.

THAT THE BOUNDARY SIGNALS REACH   correction enters at two points only. scoped credit carries
                                  them inward. how far is untested, and it is the question the
                                  whole design now rests on.

THAT THE FILTER DOES NOT STARVE   conditional childhood keeps the graph clean by attaching only
                                  what worked. whether enough works, early and at depth, to grow
                                  usable structure is untested.
```
