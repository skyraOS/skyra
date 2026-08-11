# The Second Kind of Credit

**Date:** 2026-08-11
**Status: OPEN.** Parked deliberately. One thing is agreed and nothing else is decided. This doc
exists so the question stops being rediscovered and so the mechanism is not picked by accident.

> **Everything the set calls credit lands on an invocation of an engram, in one tree, owned by one
> parent. This is credit on an OBSERVATION, aggregated across mounts. It has a different subject,
> and the design has no category for it.**

---

## Where it came from

Copy-and-edit. A parent revises by copying an engram, changing it, and citing the result. The rule
that fell out immediately:

> **Copy the structure. Never copy the credit.**

Otherwise editing launders a track record — take a well-accepted engram, change one word, arrive with
hop profiles you did not earn. `interfaces.md` already names that danger from the other side, calling
recorded provenance *borrowed trust, which the set forbids from supplying support.*

But refusing to copy the credit **creates a second question rather than closing one**. If a copy earns
its own standing, and the observations it carries are the same observations that appear in other
copies elsewhere, then something is true across the system that no local record can see.

## The subject is different, and that is the finding

```text
WHAT THE SET HAS   credit on an INVOCATION OF AN ENGRAM.
                   credit.md: "everything lands on an engram."

                   Cited        the local verdict. structural — the attachment IS the signal.
                   Acceptances  appended by the walk, carrying hop counts.
                   (Q → P)      selection credit, via the per-member cite.

                   ALL THREE are scoped to one parent's tree.

WHAT THIS IS       credit on an OBSERVATION, across mounts.
                   observations are currently THE DATA THAT FLOWS. never a subject.
```

It is not a fourth signal in an existing category. **There is no existing category for it.**

And it is unreachable by anything that exists, by construction: every mechanism is scoped to one
parent's tree, the walk terminates at crossings, and nothing aggregates across roots. That is not an
oversight — it is what makes credit local, and locality is load-bearing.

---

## The one thing agreed

> **It is a MEASURE, not a CREDENTIAL. It may be read by reflection. It may never be read at
> selection.**

Because the other reading is already refused by name:

> `credit.md` — **A TAG READ AT SELECTION TIME.** a global goodness prior. what is used gets
> selected, gets more universality, gets selected more — and it suppresses exactly the unproven
> material that contradiction arrives as.

An observation carrying a global track record, readable when choosing, **is that tag.** So the fork is
settled before the mechanism, and in the safe direction.

```text
STANDING     "should I invoke this again?"   per parent. A SELECTION INPUT. exists today.
DEPENDENCE   "how much rests on this?"       global. NOT a selection input. does not exist.
```

### Which suggests it is exposure, not quality

The best available reading, and it is the inverse of what the number looks like:

> An observation copied into forty engrams that all cite it looks like **forty independent
> confirmations**. It is **one evidential origin wearing forty records.**

That is Toy 2 exactly — *record count ≠ independent grounding*. `README.md` marks Toy 2 **DISSOLVED,
dormant at ingress**, because all ingress goes through discover. **Copy reopens it**, because copy is
a second route from one origin to many records, and it is not at ingress.

So the measurement, if built, reads as a **risk concentration**: if this observation is wrong,
everything downstream of every copy is wrong *together*, and no local walk anywhere can see that.
Same shape as `questions.md §4`'s terminus — *both wrong, and wrong together, because they read one
instrument.*

**It is not decided that this is what it means.** It is the only reading found so far that does not
collapse into the rejected tag.

---

## DECIDED, 2026-08-11 — what a copy mints

> **A copy mints a NEW ENGRAM and carries the ORIGINAL observations forward.** New invocation
> records referencing the same observation objects. Nothing else is minted.

Taken because a decision was needed to move, and it is narrower than it first looked.

### The original is UNREACHABLE, not merely uncredited

An earlier draft of this doc warned that carrying observations forward makes *credit flow backward*.
**That hazard does not exist**, and the reason is structural rather than a matter of degree.

```text
human ─▶ A ─▶ B          B returns obs1, obs2. they are A's.

copy A ──▶ A′            A′ carries obs1, obs2.
A′ creates C, passes both, C returns, A′ returns, the human commits.
```

The walk descends `Cites` and resolves `engramFor(c.Producer)`. Every observation in that chain was
produced by **A′**, **B**, or **C**. None was produced by **A**.

> **A produced no observation in this chain, so no citation anywhere names it as a producer, so the
> walk has no edge into it. It is not part of the chain and cannot be reached.**

Not *credit lands elsewhere* — **the question cannot arise.** Which is the same property that bounds
the walk everywhere else: it follows constituents, and constituents are instances.

### Copy depth

**In-chain refinement is shallow.** Same observations, same children. `A′` and `A` name one `B`, and
that is fine because `A` is superseded — it never closes, so it never mints, so two engrams naming
one child is transient and evaporates at closure.

> **This decision depends on `open-until-commit.md`. If possibilities ever close on return again, a
> copy leaves two PERSISTED parents sharing a child, and locality stops being structural.**

**Discovery copies the whole subtree**, which is a different depth for the same word.

```text
REFINEMENT   copy the engram. share the children.        the original is superseded.
DISCOVERY    copy the engram AND everything below it.    the original stays live elsewhere.
```

The one detail that follows and is not settled: a carried observation holds `Producer`, and if a
deep copy leaves that pointing at the **original** child, the walk exits the copied subtree into
someone else's tree. Either the copy re-points producers into the copies — which is minting new
observations by another name — or the boundary needs stating. See *still not decided*.

## Still NOT decided

```text
PRODUCER ACROSS A DEEP COPY       discovery copies a subtree. do carried observations keep
                                  Producer pointing at the ORIGINAL children, or is it
                                  re-pointed into the copies? the first lets a walk leave the
                                  copied subtree; the second is minting new observations under
                                  another name. REFINEMENT DOES NOT HIT THIS — it shares
                                  children rather than copying them.

                                  AND IT IS NOT ONLY A CORRECTNESS QUESTION. the runtime is
                                  DISTRIBUTED — one runtime spanning devices, entities global,
                                  search global, and probably a graph per device. so keeping the
                                  original Producer can mean the credit walk leaves the local
                                  graph: over the wire, or into a dangling ID that stops
                                  silently. re-pointing keeps it local at the cost the decision
                                  avoided. NOT BEING CALLED NOW — recorded so the distribution
                                  half is present when it is.

WHERE THE BACK POINTER LIVES      on the engram · on the observation · both · nowhere

WHAT IT AGGREGATES OVER           one mount · a principal · the whole graph

WHO MAY READ IT                   reflection, agreed. WHAT reflection is, not agreed.
```

---

## Its sibling, which is the same class

An **edit is the only negative signal in the system with an author.** Everything else is an absence —
not selected, not cited, starved — and a silence cannot distinguish *bad* from *never tried*. An edit
is a parent actively taking something out.

```text
possibilities.md has two filters. this is a third, and the strongest:

TRIED     it closed.
WORKED    it was cited.
REMOVED   a parent looked at it and took it out.   AUTHORED. NOT AN ABSENCE.
```

Both signals share the property that makes them one problem: **neither is visible from inside a
chain.** Whatever reads them runs over the whole graph, after the fact, and it is not selection.

### And the copy decision costs this signal

A superseded original never closes, so it never mints, so **it evaporates at closure and the removed
branch is gone.** The authored negative — the thing that makes this signal worth more than an absence
— is exactly what the transaction model discards.

```text
TO KEEP THE SIGNAL      something superseded must persist despite not closing, which contradicts
                        mint-on-closure — the rule that makes the index safe to discover from,
                        because nothing in it is a promise.

TO KEEP THE RULE        the strongest negative signal in the system is destroyed at the moment
                        it is created.
```

Unresolved, and it is the sharpest tension the copy decision introduces. A candidate that is *not*
being adopted here: persist superseded engrams somewhere that discovery cannot see — which is the
shape `MintRoot` already uses for roots, so the mechanism exists. It is not adopted because nothing
yet reads it, and the admission rule applies.

That is the case for a reflection pass existing at all — and this doc is not the place that argues
for one. Recorded because if it is ever built, these two are what it would have to work with, and
neither exists.

---

## Rejected already, so it is not rediscovered

```text
CREDIT FLOWING BACK       an original accumulating standing when its copies succeed. it IS the
TO AN ORIGINAL            global goodness prior, and it makes editing an inheritance mechanism.

A GLOBAL TAG READ         credit.md rejects it by name. the feedback loop is the argument, and
AT SELECTION              observation-level global credit is the loop at its purest.

COPY COUNT AS QUALITY     how often something was REACHED FOR is legitimate and discovery may
                          rank on it — interfaces.md permits provenance that is "readable by
                          discovery and invisible to credit." it says nothing about whether the
                          thing WORKED, and must not be read as if it did.

A BACK POINTER THE        if the walk can descend `From`, credit crosses copies and every edit
WALK CAN FOLLOW           inherits. whatever records provenance must be uncitable.
```

## What would settle it

```text
A CASE THAT DEMANDS IT    nothing currently needs this number. it was derived, not required.
                          the set's own admission rule applies: a mechanism earns its place
                          when something cannot be done without it.

KNOWING WHAT REFLECTION   "read by reflection, never at selection" is only a real constraint if
IS                        reflection exists. right now it names a pass nobody has specified,
                          which makes the constraint currently unfalsifiable.

THE COPY/EDIT DECISION    what is minted determines whether the link is free or must be
                          recorded. that decision is being made elsewhere and this follows it.
```

**The honest status: this is a question the design created for itself by refusing to copy credit, and
refusing was right.** Nothing in the system currently needs the answer.
