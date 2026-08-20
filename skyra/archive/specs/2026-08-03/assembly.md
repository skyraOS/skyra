# Assembly — retrieval, rendering, and where grading is not

**Date:** 2026-08-05
**Status: SPEC. REVISED 08-06** — ASSEMBLE is not a behavior. It is a `Walk` with a budget,
built by EXPECT to answer *"what would close this node."* Everything below about how it works is
unchanged; only its category is.

> **A retrieval walk, two directions, and a third job that belongs to a different wave.**

It is a RETRIEVAL walk, which means it **heats nothing** — an execution walk is attention, a
retrieval walk is reading. Without that split, looking something up makes it likelier to be
executed ([behaviors.md](behaviors.md)).

---

## The problem it started from

A reverse index over tokens: tokenized observations point at tokenized nodes, which point at
tokenized observations. The walk keeps going until the context window is full, and then the
"real work" happens over what was gathered.

Two behaviors were proposed and would not come apart cleanly:

```text
BEHAVIOR 1   perturb   create traversal behavior, grab the graph
             observe   grade the observation against the context window on unwind

BEHAVIOR 2   perturb   bring tokens onto accumulation, check count, promote if under budget
             observe   turn tokens into NL
```

They will not separate because **three different jobs are being packed into two behaviors**, and
one of the three cannot happen in this wave at all.

## 1 · Three jobs, not two

```text
ASSEMBLY    walk the activated region, gather until the budget is spent.  MECHANICAL.
RENDERING   turn what was gathered into the prompt.                       MECHANICAL.
GRADING     score what came back against what was expected.               A DIFFERENT WAVE.
```

**Grading cannot be Behavior 1's `Observe`.** At the moment the assembly wave unwinds, nothing
has come back — the model has not been called. There is nothing to grade against, because what
would be graded IS the response, and the response does not exist until after this wave has
produced the prompt that asks for it.

The sequence is:

```text
assemble        descend, gather, budget-bounded
render          ascend, fold to NL
CALL THE MODEL  at the assembly boundary — RealizeConcrete, the one caller entitled to the root's
                concrete return (kernel.md §5)
                            ↓
                the response ARRIVES AS A RAW IMPULSE
                            ↓
grade           A NEW WAVE, with its own stack, its own WaveID
```

That is not a workaround. It is the same rule every crossing already follows: a result comes back
as an impulse, never as a return value. The model is a boundary like any other, and the being
does not get to collect the outcome of its own send.

**So grading is out of scope for this behavior entirely.** Trying to put it on the unwind is what
made the split feel wrong.

## 2 · Two directions, and they are opposite

This is the real reason the loop would not close.

```text
BUDGET         DESCENDS.  it must be known BEFORE deciding whether to promote.
ACCUMULATION   ASCENDS.   it is a post-order fold and does not exist until unwind.
```

"Bring tokens onto accumulation, check the count, promote if under budget" tries to make one
mechanism do both — but the accumulation at a frame is only the part gathered *beneath* it, and on
descent nothing beneath it has run yet. The total is never available at the moment the decision
needs it.

The fix is not to make accumulation available early. It is to **carry the budget down instead**,
so the descent decision needs no total at all:

```text
child_budget = (parent_budget − parent_cost) × weight
descend iff  child_budget ≥ child_cost
```

That is already in the tree — `spacetime/traverse/traverse.go`, the Phase-0 substrate — and it is
what `2026-07-27/kernel-alteration.md` §9 describes as *"allocation = incoming share attenuated by
magnitude; keep iff allocation covers Cost."* Nothing new has to be invented.

Once the budget descends, the two directions stop competing:

```text
PERTURB   reads the remaining budget, decides what to promote.   Never reads accumulation.
OBSERVE   folds what a child realized into this frame.           Never decides anything.
```

## 3 · The missing tier

There is nowhere to put the running budget, and that is a real gap rather than an oversight to
work around.

```text
Namespace   DURABLE        the graph
???         PER-EXECUTION  budget remaining, tokens spent, the assembled set
Handle      PER-FRAME      position, coupling, local fold
```

`Handle` holds `Accumulation`, and handles are created per frame — so per-execution state has no
home. This is what *"a handle should be 1:1 with a call stack"* is sensing: the 1:1 thing is
missing, not misplaced.

**It goes on the Stack, and it is safe there.** State shared across frames *within one execution*
is not the shared-state problem: single-threaded, one owner, and the execution is already the unit
of eviction. Nothing crosses between stacks.

```text
per-execution   ON THE STACK      budget, token count, what has been gathered
per-frame       ON THE HANDLE     this node's own fold
```

## 4 · The reverse index is activation, and it is a cache

Token → nodes containing that token. That is not the graph and must not be confused with it:

```text
THE GRAPH   authoritative. every edge carries a grade (families.md).
THE INDEX   DERIVED. rebuildable. deletable. cheap to be wrong about.
```

Same status `dnav3` gives `BuildResolutionIndex` — *"a materialized cache over resolved-to
commits."* It exists so a walk does not have to start from the root every time.

And it is exactly what `2026-07-27/workers.md` already names:

> `ACTIVATION   SELECTS   which expectations this observation is RELEVANT to     tokens`

So the reverse index is the activation mechanism, it belongs to `Perturb` (selection is Perturb's,
`kernel.md` §5), and it never touches `Observe`.

**One rail carried forward.** Activation here is additive only — it raises what is in view and may
never dim what arrived. Attention narrows *within* an execution; it does not decide which
executions exist. That separation is what keeps Toy 5 contained (`questions.md`, 08-04).

## 5 · What it actually is

One walk, not two behaviors:

```text
THE ASSEMBLY WALK

  Perturb    consult the index for the activated frontier
             read the remaining budget off the stack
             promote the next affordable node, or nil when the budget is spent
             nil is the base case — the window is full, and that is not a failure

  Observe    fold the child's realization into this frame's accumulation
             at the root, that fold IS the rendered prompt
```

Behavior 1's *"create traversal behavior, grab the graph"* is not a separate behavior — it is the
root frame's first `Perturb`. Behavior 2's *"turn tokens into NL"* is not a separate behavior — it
is the top of the fold. What looked like two is one, and what looked like the second half of it
was grading, which is a different wave.

---

## Open

```text
COST            nothing carries a token count yet. Node has no Cost, and the budget has
                nothing to subtract. This is the first thing to build.

WHAT RENDERS    "fold to NL" is stated, not specified. what a concrete Space carries, and
                how the root turns it into a prompt, is undecided.

THE INDEX       when it is built, when it is invalidated, and whether it is per-namespace
                or per-execution.
```
