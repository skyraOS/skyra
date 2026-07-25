# The Self-Authored Loop

**Date:** 2026-07-13
**Status:** SCRATCHPAD — raw working thought, not yet a design or implementation commitment
**Companion:** `README.md` (Multiple Potential Wells) — the field/gradient math is parked; this
file is the loop it would sit inside.

---

# Part I — Michael

## The Current Code Is A Stopgap

The DAG was never the goal. The workflow definition, the fan-out, the join, the reframe rail —
all of it is scaffolding for a pipeline that shouldn't be a pipeline.

## What I Actually Want

A loop where the model itself defines its own system prompt.

I don't want to tell the model how to think. I want to give it a place where it can correct
itself from previous experience, and learn.

## The Scheduler

Not a DAG. A watcher.

```text
grades dip
    ↓
traversal fires
    ↓
a new action, with better context, fills the screen
```

## Sensory Worker + Watcher

```text
impulse arrives
    ↓
sensory worker receives the impulse
  + the result of traversing PRESENT state and DESIRED state
    ↓
updates both subgraphs
    ↓
watcher (scheduler) sees the change
    ↓
starts a new traversal
```

## Why Planning Came Back

I honestly didn't even want planning. But it's a nice way for the system to grade itself and
learn. Basically it's this:

```text
what did I think I wanted to do
what did I actually do
compute the difference
```

And the more I think about it, planning should be for allowing the model to **parallelize work**.

## Emit A Subgraph At Every Step

A vague plan that turns out right teaches you nothing. That's why every step should emit a
subgraph — a *predicted current state*.

The summary — "the planning" — isn't really the thing I'm after. Although it *is* the system
prompt for that node.

```text
Observe (each step) emits:
  summary                 ← the system prompt for that node
  predicted world state   ← the subgraph; the bet

on reafference from that node:
  sensory grades  predicted world state  vs  actual world state
```

And then we are cooking.

---

# Part II — Claude

Reactions, sharpened claims, and the failure modes I'd expect. Nothing here is a commitment;
it's the stress test.

## Why The Subgraph Is The Whole Design

> A sentence can only be judged. A subgraph can be diffed.

This is the load-bearing consequence of "emit a subgraph at every step," and it's the escape
from the trap the current build is in.

Today a large share of the learning signal is an LLM grading an LLM (`final:grade`'s judge, the
coherence settle, any closure check). Those grades bottom out in an opinion and then get written
into the durable ledger as if they were evidence.

A predicted-state subgraph doesn't need a judge:

```text
predicted:  door --state--> open
observed:   door --state--> closed
            ↓
        set difference
            ↓
          grade
```

No opinion anywhere in that path. Prediction-vs-observation is self-grounding. This is how the
loop stops being self-referential.

**Corollary:** the same argument says the *world* is the only thing allowed to teach. Every grade
should carry provenance — reality-derived or judge-derived — and the two must never silently mix
in the ledger. Cheap now, impossible later.

## Why The Per-Node Prompt Can't Drift

The naive version of "the model writes its own system prompt" — one global document the model
edits, which then conditions the next edit — **always degenerates**. No error signal, so it
flatters itself into mush.

The version in Part I doesn't, and the reason is worth stating precisely:

```text
the summary is per-NODE, not global
its quality is measured by the accuracy of the prediction it produced
well-calibrated nodes get retrieved more; bad ones decay
```

The model never holds the pen on its own prompt. It writes *claims*; the world grades them;
retrieval compiles the highest-graded relevant claims into the next framing. The prompt evolves
by **selection on predictive accuracy**, not by self-editing. That is a completely different
mechanism from prompt self-modification, and it's the one that survives contact with critics.

## The Failure That Will Happen On Day One

**Under-prediction is the degenerate strategy.**

If the grade rewards accuracy, the optimal policy is to predict almost nothing. Emit one
trivially-true node, score 100%, collect the credit. The system will select for exactly the vague
plans this design exists to eliminate — and it will *look* like it's working, because the numbers
go up.

The grade must punish silence. Both directions:

```text
error = (predicted but not observed)     ← you were wrong
      + (observed but not predicted)     ← you were blind
```

The second term forces boldness. A node that predicts nothing has **maximum** error, not zero.

The value of a node is not low surprise — it's **calibration**: did it commit to what mattered,
and was it right. This wants a strictly proper scoring rule, where honest and complete prediction
is the maximizing strategy.

**Scope:** the model is not on the hook for predicting the color of the sky. It is on the hook for
the delta it claims its action causes, plus whatever falls inside the attended frame. Attention
defines the diff's domain — which is what the traversal budget already does.

## The LLM Aligns; The Diff Grades

Predicted `{a glass of water}[3]`, observed `{the cup}[7]`. Something has to decide those corefer,
and that something is a model. Fine — but let it do *only* that.

```text
allowed:     "these two nodes are the same referent"   ← weak, checkable
forbidden:   "you did well"                            ← the opinion, back through the side door
```

Alignment is a weak claim. Judgment is not. Keep the boundary.

## Untested Is Not Wrong

Most descent nodes are never traversed — you only walk one path. Their predictions never come due.

```text
falsified   → alignment (A) down
unresolved  → confidence (N) low; alignment untouched
```

Collapse these and you punish branches for not having been explored. The A/S/F/N store already has
the right shape for this; the distinction just has to be respected.

**Open:** every prediction needs a **maturity condition** — when does the bet come due? Next
observation? When the bridge advances past that node? A timeout? Without one, open bets pile up
ungraded. The current code already has this smell in `LatestUngradedFinal`.

## The Watcher Will Spin Unless It Watches The Right Thing

Traversal *writes to the graph* (descent mints predecessor states; ascent records efficient causes;
success crystallizes edges). A watcher that fires on "the graph changed" triggers itself forever.

So it must not watch the graph. It watches **error** — and the endpoint versions. Which forces an
ownership split:

```text
ENDPOINTS  (CURRENT, DESIRED)   versioned; written by exactly ONE writer (the sensory worker)
BRIDGE / FIELD                  derived, disposable, keyed by (currentV, desiredV); never watched
DURABLE GRAPH                   written only on ascent, only by graded evidence
```

**Consequence:** descent does **not** write DESIRED. The predecessor states it invents belong to the
bridge, not to the goal. If they land in DESIRED, descent mutates its own trigger and the loop spins.

**Level-triggered, not edge-triggered.** Keep a dirty bit + last-reconciled version. Three impulses
arriving during a traversal produce *one* re-run, not three. Compare versions at the emission
boundary; stale-and-diverged is discarded and re-reconciled. That removes the need for a
cancellation policy entirely.

## Cost Scales With Surprise, Not Time

"Grades dip → traversal fires" is a homeostat, and it buys a property a version-watcher doesn't:
when predictions hold, the being is nearly free; when reality diverges, it thinks hard.

That's what makes always-on affordable — and always-on is what the architecture is *for*. It's also
the actual predictive-processing story: you only compute when you're wrong.

## Parallelism Is Free — But Only From A DAG

The parallelism intuition is right, and it's a far better justification for planning than
route-finding is. But it only works if the bridge is a DAG. The sketch in `README.md` describes a
**chain**:

```text
TARGET → final-n → final-(n-1) → ... → CURRENT      ← chains do not parallelize
```

If descent emits one predecessor per hop, you get a critical path and nothing else. If descent may
emit **multiple independent predecessors per hop**, the prerequisite structure *is* the dependency
graph, and what can run concurrently is read straight off it — the way a build system does it.

Parallelism isn't something to design. It's something to avoid destroying.

> **Decide now, while it's free: descent produces a prerequisite DAG, not a predecessor chain.**

## What This Deletes

The board loses its coordination role. No fan-out, no join, no atomic motor guard, no one-reframe
rail — those were all coordination for a pipeline that no longer exists. The Row becomes: impulse
ID, causal predecessor, the endpoint versions it read, the trace.

Two grade genes (`final:grade`, `efficient:grade`) collapse into one closure at the transition
boundary. `Frame.Text` stops being load-bearing — the frame *is* the endpoint pair; the sentence is
a rendering.

**Open question worth answering early:** does the prediction diff subsume the **coherence settle**?
The settle is a multi-pass LLM judge whose entire job is to write alignment onto edges. If alignment
falls out of prediction error instead, the most expensive gene in the system deletes itself — and
the collapse pays for the descent.

## The Standing Principle

Everything above is one rule, applied repeatedly:

> **Don't script the process. Ground the outcome.**

The only thing hard-coded is the loop that turns outcomes into retrieval pressure. Every hand-designed
cognitive choreography — which gene fires when, how the model must annotate, what the settle asks — is
a candidate for being absorbed by the next model. The ledger is not. Cut the ceremony; keep the ledger.

## The Loop, End To End

```text
descent   — each node emits its framing (its prompt) and its predicted state subgraph (its bet)
praxis    — walk the bridge; act at exactly one boundary
reafference — the world returns the actual state subgraph
diff      — the grade, both directions, so silence is punished
retrieval — grades move what gets recalled; recall compiles the next framing
watcher   — error rises, the loop fires again
```

Nothing in it tells the model how to think. Nothing in it bottoms out in an opinion.
