# Possibilities

**Date:** 2026-08-09
**Status: SPEC.** Not built.

> **A possibility creates possibilities and reads their returns. There is no third capability.**

---

## The layering

```text
DURABLE      +
INTERACTION  =
EMERGENT
```

A series of interactions over durable structure produces the behaviour you see. **The emergent layer
is not represented anywhere** — it is what the other two do, not a thing that is stored.

Possibilities are the interaction layer. It is deliberately ignorant: every distinction requiring
judgement belongs above it.

## What a possibility is

**One opening of an engram**, manned by an entity.

```text
THE ENGRAM         durable. what to do, and its accumulated history.
THE ENTITY         who interprets. stateless.
THE OBSERVATIONS   the parameters.
THE COMMIT         one observation. the completed work.
```

Durable is what *has* happened; ephemeral is what *might*. That pairing is the point of the naming —
potential should not be the thing that persists, and a record should not be the thing you invoke.

**A possibility knows three things:** the observations it was given, which engram it is opening, and
the observations its children returned. Nothing else is in scope.

**It cannot perform work except by creating a possibility and handing it over.** An incapacity, not a
rule — everything it does, it does by minting something and emitting it.

## Two phases

```text
GATHER      the sequence of selects.       SELECTION is the skill.
SYNTHESIZE  the terminal composition.      SYNTHESIS is the skill.
```

Both may be code. Both may be a model call. Which it is depends on the entity, not on the engram.

The resemblance to `Perturb`/`Observe` is close enough to refuse explicitly:

```text
PERTURB / OBSERVE   INTERLEAVED — perturb, descend, observe, perturb. about PERMISSION.
                    structural: two methods, two API views.

GATHER / SYNTHESIZE the boundary is at the END, not between children. about SKILL.
                    descriptive: a sequence of emissions whose last one composes. no machinery.
```

`Perturb` was `Observe` with `realized = nil` — a name for the first iteration of a loop that had
already started iterating. What the split was carrying, and what pays for it now:
[traversal.md](traversal.md).

## Leaf and orchestrator, derived from composition

```text
NEVER SELECTED   a LEAF. atomic, one-shot. ORIGINATES an observation.
HAS SELECTED     an ORCHESTRATOR. may only name or combine what came back. PERMANENTLY.
```

**Permanent is necessary, not tidy.** If a possibility could originate *after* orchestrating, nothing
could distinguish content that came from its children from content invented at the end, and *cannot
originate, only combine* would collapse.

Kind is derived from composition and never declared — the third independent arrival at directory
versus file.

**Origination bottoms out in an entity's judgement, or in code.** Which is what makes the entity
definition load-bearing rather than descriptive.

## Composition

**The commit is singular.** Many things go in; one comes out; that one may compose several but it is
its own thing.

A plural return makes every possibility a **pipe** — the parent sees its grandchildren's observations
directly, reaches past the child into its subtree, and the frame boundary means nothing. One composed
observation is opaque: the parent sees what the child *concluded*, not what it was told.

> **Synthesis is the job. A possibility that refuses to compress is plumbing.**

### It is compression, not conservation

Nothing is conserved — N in, 1 out, and information is destroyed on purpose.

```text
SEEN   parameters + every child's return
USED   CITED BY ANY EMISSION — a select or the commit
```

**Two consumptions, and both count.** An observation used in a select shaped an *invocation*; one
used in the commit entered the *composition*. Counting only the second punishes exactly the
observations that made invocation possible — a discover result would be permanently wasteful, which
is what creates pressure to smuggle scaffolding into the synthesis.

Unused therefore means genuinely inert: it changed neither what the possibility did nor what it said.

`used / seen` is a compression ratio. **Low does not mean bad** — it means a lot was looked at to
conclude a little, which is a bad query, an over-provisioned frame, **or** legitimate exploration that
discarded correctly. A gauge, not a verdict, and it means something only differentially.

If the metaphor becomes energy it must be **dissipative**. Observations flow, they are the only thing
that flows, and select/synthesize shape the flow — but conservation is a law this design does not
have, and assuming it sends someone looking for where the missing observations went.

### A possibility cannot originate an observation

> **It can only combine what it can name.**

Frame is what you were passed plus what your selects returned, so composition has nothing else to
draw on. This is the anti-fabrication constraint arriving through **scoping** rather than through a
type on observations. It bounds **provenance, not validity** — you cannot cite what you never saw,
but you can still reason badly about what you did. That is what selection is for.

**And it is enforced rather than agreed**, because citation is derived from content: what you cited
is the set of references appearing in what you handed up. You cannot claim to have used something you
did not carry forward.

## Lifetime matches the work

> **A possibility's lifetime matches the work, never the lifetime of what it made.**

A browser stays open; the possibility that launched it does not. Coupling runtime lifetime to world
lifetime makes the runtime hold something the world was already holding.

```text
"open the page"        returns immediately with a handle. the browser is WORLD STATE.
"watch until X"        genuinely long-running — the possibility is alive because THE WORK is.
```

Both are legitimate. But a **watch** should not be a held-open possibility at all: the page changes,
an event arrives, a caller invokes discover. Duration stays in the world where it already was —
*a live process is a crossing that streams; the wave spawns and returns, results arrive as impulses.*

Which leaves the group barrier mattering only for **computation you are waiting on**, never for
artifacts that persist.

## Mint on closure

A possibility mints its durable engram on closure if one does not already exist. Creation is not a
separate authoring step: **an expectation is improvised, it runs, and if it closes it persists.**

Which separates two filters that are easy to conflate:

```text
ENGRAMS   minted on closure.     record what was TRIED.
EDGES     formed on citation.    record what WORKED.
```

**Discovery sees attempts; selection sees successes.** A possibility walks its engram's children, so
it only ever sees what worked. Discovery walks the index, so it finds things nobody has used yet —
including earlier failed attempts, which is memoization rather than bloat: the second time a need
arises, discovery finds the orphan instead of re-improvising it.

## What a possibility sees

Everything on its engram's edges — the successes and the failures.

```text
the child engram's name
its stored returns — the history, and the diffs across it
credit records — hop profiles and rejections
the cite-source of each selection — (Q → P)
```

So the prompt is: *here is what I have used, what it said, when it worked, and when it did not.* No
weights change. **The entity sees a richer edge set over time**, which is what topological learning
means in practice.

### It sees DOWN, never UP

Its edges point at its children, so it sees their history. Its own reception lives on its **parents'**
edges, which it cannot reach.

> **No engram can know how it has been received.**

A safety property falling out of locality rather than designed for: a thing that cannot observe
whether it is being cited cannot optimise for being cited. **The citation signal is ungameable from
inside.**

### Synthesis has no historical input at all

The one-shot sees the frame and what was gathered. Nothing else — its own past returns live on its
parents' edges.

So composition quality is a function of the entity's priors and the material handed to it. **The
system improves by changing what gets gathered, never by composing better.**

Stated explicitly because it is surprising, and because an implementation will otherwise assume
otherwise.

### The frame travels

Everything a possibility sees ships to the entity's host **on every step**. So `what a possibility
sees` is both the learning mechanism and the per-step transfer cost, and those pull against each
other. The view that strips what a possibility may not see is therefore also a **budget** boundary,
and the two jobs may not want the same answer.

---

## Rejected

```text
A PLURAL COMMIT           makes every possibility a pipe and invites padding — hand up ten
                          things so one gets used. singular means verbosity buys nothing.

TYPED OBSERVATIONS        world-authored vs agent-authored at the base layer. provenance is a
                          semantic judgement and this layer makes none. scoping does the work a
                          type was reaching for.

SELF-MODIFYING SYNTHESIS  a possibility rewriting its own expectation when a result is
                          inadequate. self-modification without external pressure defeats
                          starvation — you cannot decline to invoke something that reshapes
                          itself into whatever keeps being invoked.

DECLARED CITATION         a commit stating what it used. two places for one fact eventually
                          disagree, and deriving it makes CREDIT FOLLOWS DATA structural.
```
