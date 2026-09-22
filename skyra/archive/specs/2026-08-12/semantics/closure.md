# Open Until the Root Commits

**Date:** 2026-08-11
**Status: PROPOSAL.** Revises the possibility lifetime in `2026-08-09/possibilities.md` and
`credit.md`. Nothing here is built.

> **A possibility RETURNS when it has an observation. It CLOSES when the root commits.**
>
> **Those were one word, and almost everything expensive came from that.**

> **CARRIED FORWARD** from `specs/2026-08-11/semantics/open-until-commit.md`.
>
> **Status here: CARRIED, RENAMED.** This document decided the model
> `../continuation-language/primitives.md` now calls
> **`Remember`** — *"Closure is the only causal durable append. A commit produces an observation;
> it does not persist anything"*, one append per chain rather than one per possibility. `Remember` is this
> plus `Accept`, unified into one recursive operation.
>
> **The decision is 08-11's; only the name is new.**

---

## Why this exists

`build.md`'s refinement finding cost more to fix than anything else in the run, and the fix was four
rules no document states:

> the failure must travel as an **argument**; an orchestrator must **forward** what it was handed;
> *"what failed"* is **ambiguous** between the work that was wrong and the suggestion that was
> disliked; and the exclusion set must be **cumulative**, with the caller's frame as the only memory.

All four exist to carry information back **down** to a possibility that had already died. The
diagnosis it was written under:

> A possibility is one opening, so a fresh `discover` cannot know its own last answer was rejected.

**If it never closed, it knows.** The accumulated observations are still in its frame. Nothing has to
travel anywhere, because nothing left.

---

## The change

```text
BEFORE   commit → produce an observation → THE POSSIBILITY DIES.
AFTER    commit → produce an observation → THE POSSIBILITY STAYS LIVE.
         the ROOT's commit closes the whole tree.
```

### Three events, and the set has two words for them

```text
COMMIT   the emission. produces an observation and hands it up. STILL LIVE, still re-drivable.
RETURN   the same event, seen by the parent.
CLOSE    THE ROOT COMMITTED. materialize · disperse · mint · die.
```

Every *commit closes the possibility* line in the set needs auditing against this, including the
build's loop, which returns from `open()` on commit.

### Which moves the durable append

> **Closure is the only causal durable append. A commit produces an observation; it does not persist
> anything.**

`credit.md`'s *every durable append in the system is caused by a commit* becomes **caused by the
root's commit** — one causal append per chain rather than one per possibility. The namespace cost
account may require a distinct append path; that question remains open in `../energy/distance.md`.

And **mint-on-closure was always named for the right event and implemented at the wrong one.** The
build mints at every commit, which under this vocabulary is minting at *return*. The name has been
correct since it was written.

Eviction safety gets stronger rather than weaker: an abandoned chain now leaves **literally nothing**,
in the graph and in the world, because neither channel wrote.

---

## Refinement in place

The AI returns something wrong.

```text
THE USER DOES NOT DENY IT.
THE USER DOES NOT PASS A FAILURE BACK DOWN.
```

There is no failure. **It did not fail yet** — the user has not committed and has not returned. What
the user does instead is revise:

```text
1  change the expectation.        the user has authority over it; it is theirs.
2  create a new observation.
3  send it down into the SAME LIVE POSSIBILITY.
```

And what the possibility sees is a sequence, not an event:

```text
the old expectation
    → the observations accumulated under it
        → the new expectation
            → the new observation
```

**The system knows a change happened. It does not conclude the earlier observations were bad.** That
distinction is the whole reason the old mechanism needed *"what failed" is ambiguous* written down —
it was trying to encode, in a failure argument, something that is plainly legible as a sequence.

The dependency chain is unchanged. The possibility still cites what it used. It still returns. **The
authority still lies with the user**, and nothing became durable while any of this happened.

> **This is refining a workflow through natural language. When the root commits, what worked becomes
> history and the rest is forgotten.**

---

## Expectations are drafts until closure

Refinement-in-place looks like it breaks immutability. It does not, and the resolution is already in
the set:

```text
interfaces.md   "Immutable. Expectations never improve — authored once and re-run."
possibilities.md "an expectation is IMPROVISED, it runs, and if it closes it persists."
```

**Improvised means draft.** An expectation was always mutable until closure; staying open only
lengthens the drafting window past the first return.

> **Immutability is a property of COMMITTED expectations, not of expressions in flight.**

Stated because *immutable* currently reads absolute, and an implementation will otherwise freeze the
expression at open.

---

## What this dissolves

```text
THE FOUR MISSING RULES     failure-as-argument · forwarding · the ambiguity · the cumulative
                           exclusion set. all four carried information DOWN to something that
                           had died. nothing dies.

THE REJECTION MEMORY       "short-term, held for the life of the asking, passed to discovery as
ITEM                       a parameter" — build.md's attack list, item 3. it is the frame,
                           which was always held for the life of the asking.

TWO STAGING MECHANISMS     kernel.md stages WORLD EFFECTS to root acceptance. memory now stages
                           to the same point, by the same walk. one rule, two alphabets.
```

## What this resolves

`conversation-runtime-mismatch.md` records five mismatches and says it does not choose. **Four are
chosen by this change and the fifth by *discovery is always a copy*:**

```text
1  errors remain ephemeral                    nothing durable until the root commits.
2  only successful structure matters          what worked becomes history; the rest is forgotten.
3  the node authors a better expectation      it revises in place. nothing is passed back because
                                              nothing returned to pass.
4  reuse copies the expectation               DISCOVERY IS ALWAYS A COPY.
5  the user is the ultimate acceptance        dispersal happens once, at the root.
   boundary
```

That note is no longer a list of tensions. It is a specification of the conversation side, and the
runtime is what disagrees with it.

---

## Credit becomes batch, and the verdict becomes final

`credit.md`: *every commit disperses. the root is not special.*

**False as written now.** Dispersal happens once, for the whole tree, at closure.

I think this is an improvement rather than a cost:

```text
BEFORE   a parent's verdict is rendered the moment it commits, on partial information, and
         cannot be revised.
AFTER    a parent may revise until the root lands. THE VERDICT IT RENDERS IS FINAL.
```

Which is also what makes *the durable graph records what worked, not what was tried* literally true.
Before, it recorded what worked **at the moment each parent stopped looking**.

**What does not change:** the asymmetry. Errors stay local, acceptance walks down by hops, and the
walk is the same walk. Only when it runs moves.

---

## What it costs

```text
EVERY POSSIBILITY IN A     memory, and for model entities HELD CONTEXT. monitor already keeps a
CHAIN IS LIVE FOR THE      context current, so it may be the same context that would have been
CHAIN'S DURATION           re-assembled anyway — but nobody has measured either, and a long
                           chain now holds all of it at once.

A LONG CHAIN HOLDS         one causal append per chain means a chain doing an hour of work
EVERYTHING UNTIL THE END   persists nothing for an hour. crash safety is total and progress
                           safety is zero. THIS IS THE REAL TRADE.

THE GROUP BARRIER          "every group awaits" now means awaits RETURNS, not closures. the
                           barrier is unchanged; the word underneath it moved.
```

The middle one is the one to argue about. It is the same shape as a long-running transaction in any
database, and the usual answer — checkpoints — is a partial-completion mechanism, which
`questions.md §9` says must not be smuggled in without a case that demands it. **This may be the
case that demands it**, and it should be recognized as such rather than solved early.

---

## Rejected

```text
COMMIT AS CLOSURE          one word for two events, and the source of the refinement defect,
                           the rejection-memory gap, and provisional verdicts.

PASSING FAILURES BACK      it did not fail. the user did not commit and did not return, so
DOWN                       there is no failure event to carry — only a revision, which is
                           legible as a sequence without anything being encoded.

FREEZING THE EXPRESSION    immutability applies to COMMITTED expectations. mint-on-closure
AT OPEN                    already said an expectation is improvised until it closes.

CHECKPOINTS, FOR NOW       a long chain persisting nothing is the real cost, and checkpointing
                           is partial completion wearing a different name. name the case before
                           adding the mechanism.
```

## Open

```text
WHEN DOES A POSSIBILITY    if nothing closes until the root, what evicts a subtree that is
STOP BEING RE-DRIVABLE     finished and will never be revised? the kernel holds it all. a
                           stalled head is detectable; a DONE head is not.

WHAT A REVISION COSTS      RESOLVED 2026-08-11 — a revision DOES restore origination, and the
  → RESOLVED               forfeit is scoped to the EXPECTATION rather than to the possibility.
                           the objection was that the forfeit would be escapable by revising.
                           It is not, because THE PARENT REVISES, NOT THE POSSIBILITY —
                           possibilities.md already rejects SELF-MODIFYING SYNTHESIS by name,
                           and a parent revising is the external pressure that rejection says is
                           missing. Only the party the rule protects can waive it.
                           see type-surface-proposals A2.

DOES THE ROOT'S COMMIT     the root is a namespace, not an engram (roots-and-admission.md), so
HAVE A HOME                "the root commits" needs a subject. the principal commits INTO the
                           namespace. that is consistent, and it is not written anywhere.

REPLAY                     the monitor invariant holds that assemble(f) == assemble(f₀) +
                           monitor*. a revision sequence is now part of f. replay must
                           reproduce the revisions, in order, or diverge.
```
