# The Scheduler

**Status:** DESIGN · part of [kernel-v2](README.md). **New** — not in `specs/2026-07-08/`.

## Why it has to exist

The reform says `Observe` returns the traversal surface. Good. But **what surface?**

A cell cannot answer that from its own type, because the same cell behaves differently depending on
**what the wave is for**:

```text
                    RETRIEVAL wave        PRESENT wave            PERCEIVE wave
triple cell         OR — posting list     spread to expressions   leaf
expression cell     AND — its requires    spread to its expects   —
place cell          —                     —                       adjacency
```

A triple cell is an **OR node** in a retrieval wave and a **leaf** in a present wave. It has no way
to know which.

> **Behavior = (cell type × wave purpose).**

That is a two-dimensional dispatch, and neither dimension lives in the cell alone. So `Observe` and
`Express` become **thin**, and delegate:

```go
func (c *TripleCell) Observe(origin Space, imp any) map[string]Alloc {
    return c.sched.Surface(PurposeOf(imp), c)      // the scheduler knows what THIS wave is for
}

func (c *TripleCell) Express(origin Space, imp any, kids []*Result) any {
    return c.sched.Fold(PurposeOf(imp), c, kids)
}
```

## The wave envelope

Something has to ride every wave so `PurposeOf(imp)` can work:

```text
WAVE ENVELOPE   waveID · purpose · visited set · budget
```

And note the **visited set** is on it. That is not incidental — a walk that crosses a device boundary
must carry its visited set or it loops forever (`one-kernel.md`). The trace-context a distributed
walk needs and the dispatch key a scheduler needs are **the same object**.

## Which makes the job board load-bearing

The board's `Purpose` field (`workflow | present | perceive`) is not observability. **It is the
dispatch key.** The board is not bookkeeping *about* the traversal — it is *part of* it.

```text
Row {
    WaveID      the key
    Purpose     ← THE DISPATCH KEY. not a label.
    WorkflowID  self-referencing at the root
    ...
}
```

## The risk: the scheduler becomes the new kernel

Every behavior in the system now routes through one object. That is a **god object waiting to
happen**, and if cognition leaks into it, you have reintroduced exactly what the void kernel was
protecting against — a hand-designed cognitive choreography, sitting one layer up, wearing a
different hat.

So, the line:

> **The scheduler DISPATCHES. The cell IMPLEMENTS.**
>
> It is a **routing table**, not a library of algorithms.

```text
ALLOWED     "in a retrieval wave, a triple uses the OR frontier and the OR fold"
FORBIDDEN   knowing what GROUNDED means
            knowing what a GRADE is
            knowing what a TARGET is
            computing anything
```

The expression cell knows how to AND its children. The triple cell knows how to OR them.
**Groundedness is just the value that flows up the fold** — nobody has to *know* it.

### The smell

> **If the scheduler grows a method that COMPUTES something rather than SELECTS something, that is
> the leak.**

Watch for it. It is the failure mode this whole architecture has been arranged to avoid, and it will
arrive disguised as convenience.

## Where the skeleton plugs in

The three brute-force functions in `skyra-dna-v3` already have the right shape — deliberately:

```text
Closed()      → the AND/OR fold.        expression cells AND; triple cells OR; groundedness flows up.
Precedent()   → the OR frontier.        the posting list IS the triple cell's onward distribution.
present()     → the budgeted spread.    tokens as the budget; attention is affordability.
```

Three functions, three scheduler policies, **zero changes to the kernel** — because the kernel never
knew what they meant.

## When to build it

**Not yet.** The skeleton's one-hop retrieval should drive `passes-to-closure` to 1 for
single-expression tasks (it already does: `[2 1]`) and **stall** on tasks needing a multi-step
bridge, because a flat posting-list lookup cannot see a chain.

**That stall is the signal.** It is a falsifiable prediction about when the kernel becomes necessary,
and it means the decision gets made by data rather than by appetite.

Build it when the number says to.
