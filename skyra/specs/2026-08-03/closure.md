# Closure — and what `P` is actually for

**Date:** 2026-08-06
**Status: RESOLVED.** Supersedes [2026-07-27/truth.md](../2026-07-27/truth.md)'s claim that `P`
measures closure. It does not, and the two were never the same question.

> **Closure is one observation, this execution. `P` is a lifetime fold, for retrieval.**

---

## Two questions that were being asked of one number

```text
CLOSURE   did the perceive I just ran come back?    ONE observation. THIS execution. no history.
P         how reliably has this closed before?      lifetime fold. RANKS. settles nothing.
```

Asking the retrieval quantity to answer the execution question is what made time keep biting. It
cannot be fixed by weighting, because it is not a weighting problem — the two questions want
different inputs.

### The failure that exposed it

```text
T1   perceive: door open   → +1
T2   perceive: door shut   → −1

P = G/|O| = 0        reads CONTESTED. the door is shut.
```

A lifetime average has no *right now* in it, and no amount of recency weighting makes an average
into a current reading. Closure needed something else, and the something else turns out to be
trivial: **run the perceive and read what comes back.**

## `trust.md` said this and `truth.md` contradicted it

> Trust is… a **RANKING** signal — attention, retrieval, what to look at next.
> **It is not a truth value and may not settle anything.**

`truth.md` then had `P` — the same fold, normalized — measuring how much of a change had closed,
which is settling. That is the contradiction, and it resolves in `trust.md`'s favour:

> **The entire fold is a retrieval quantity. Nothing derived from history closes anything.**

### Which takes the normalizer off the critical path

If `P` only ranks, it only has to be **monotone** — ordering is what matters, not scale.
`truth.md`'s two objections still affect ranking *quality*:

```text
the mean punishes breadth      one strong + nine weak positives reads as barely supported
the scale is unpinned          0.333 means different things on [-1,1] and [0,1]
```

Neither gates execution any more. That question has been open since 07-27 and just stopped
blocking anything.

## Closed is +1. There is no threshold.

```text
CLOSED       the most recent grade in this execution is +1
NOT CLOSED   anything else
```

Not a cutoff on a continuum — a single grade, read off the ephemeral space. Which means **two
kinds of grade exist, and only one of them can close anything:**

```text
PERCEIVE / ACK      ±1.       DEFINITE. the world answered. this is what closes.
INTERPRETATION      [-1,1].   graded bearing. feeds support → P → RETRIEVAL. closes nothing.
```

A perceive asks something the world can answer outright — *is the door open?* — so its return is
definite. `relationship ∈ [-1,1]` answers a different question, *how much does this bear on
that*, and that one is for ranking.

### −1 is a finding, not an absence

```text
+1          CLOSED.
−1          DEFINITELY NOT CLOSED. the look confirmed the negative — this is information,
            and it is what makes "children closed, self not closed" fire.
0 / absent  DID NOT LOOK, or could not tell. NOT the same as −1.
```

Collapsing the last two would lose the distinction the whole design turns on: *absence became a
fact* versus *nobody went and looked*.

### And it explains the standing root mechanically

`affordance.go` decides measurability by whether the world exposes a perceive for a predicate. A
**standing** expectation has no exact channel — so **its perceive can never return +1**, so it can
never close. That was a definitional claim; now it is a mechanical consequence of the same rule
everything else uses.

## Closure is ephemeral, and closure EVENTS are durable

```text
DURING EXECUTION   closure lives on the ephemeral spaces. computed, not stored.
AT CLOSURE         the event is written. append-only. this is the history P folds over.
THE NODE           is never marked closed. there is no flag, and nothing has to un-set one.
```

`P` does not travel with the ephemeral spaces. It is computed *from* the durable graph, for
deciding what to retrieve — and it plays no part in deciding whether anything is done.

## There is no reopened state — you re-execute

An expectation that cannot close searches history, finds a prerequisite that closed recently,
and **runs it again.** Its perceive answers. Nobody marks it open, nobody watches for staleness,
nothing revises a stored fact.

This is the append-only discipline reaching closure:

> **You do not revise a fact. You observe again.**

## What this commits the design to

The durable graph holds **structure and history, and no claim about the present.**

```text
"has the door usually been open?"    cheap. fold the history. that is P.
"is the door open right now?"        you have to LOOK. there is no stored answer.
```

That is further than dnav3 took LAW 1 — its `ObservationGraph.Resolve` folds observations to
answer *is X true*, which is a stored present that later reads trust instead of looking. Here
there is not one. More expensive, and it makes the perceive genuinely the only channel to the
world rather than the channel that populates a cache.

## Two constraints on what may enter closure

**An action's ack must not propagate.** `stand_up` closing means it *ran*, not that it worked. If
that inherited upward, the being would close "I am standing" because it performed the motion —
`2026-07-27/expectations.md`'s original failure, verbatim:

> The being concludes it is logged in **because it performed the keystrokes** — it has graded its
> own procedure and called that success.

```text
CHILD EXPECTATIONS   may bear on a parent's closure — they carry world evidence
ACTIONS              may NOT. an ack is a fact about execution, never about the world
```

**Only perceives put world-evidence into the fold.** That is the falsifiability invariant with a
mechanical reason behind it: without it, closure would be computable without ever leaving the
agent.

---

## What this changes about Q1

[questions.md](questions.md) Q1 — *what does the child edge assert* — was marked **blocks the
arithmetic**, because `trust.md` multiplies along it.

It still needs an answer, but the stakes dropped: the arithmetic it blocks is **retrieval
ranking**, not execution. A wrong reading of the child edge produces worse retrieval order, not
wrong closure.

The candidate, from 08-06: the child edge is the **authored confidence** — *"I am x% sure this
child completes me"* — and support multiplies through it, so a child grounded at `0.8` beneath a
`0.3` confidence contributes `0.24`. That gives the multiplication a meaning rather than an
inherited formula: *how much of this child's grounding is mine to claim.*

## Open

```text
RE-EXECUTION COST       "cannot close → search history → run the prerequisite again" has no
                        bound. a chain of stale prerequisites could re-run a long way back.

WHAT A CLOSURE EVENT CARRIES   enough for P to fold it, and enough for the search that finds
                        "this closed recently." at minimum a node, a wave, and a time.
```
