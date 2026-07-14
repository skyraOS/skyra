# The Wave

**Status:** DESIGN · part of [skyra-dna-v3](README.md)

The kernel does **two** things. It does not do the diff and it does not do closure — those are set
operations and must never get a traversal.

```text
WAVE 1   builds the WORKFLOW              triggered by a final-state change
WAVE 2   builds the PRESENT for a node    triggered by a worker inside that workflow
```

Same algorithm, different starting point, different frontier, different budget unit. The kernel
knows none of it.

## The wave spreads in both directions

```text
        prior observations   ←——  node  ——→   expected states
        (what led here)                        (where this goes)
              ↑                                       ↑
        backward spread                        forward spread
        history / precedent                    goal / plan
```

Both budget-attenuated from the node, **in the same wave**. The present is the union of what was
affordable in each direction: history you can afford, plus goal you can afford.

This is `DerivePresent` (v.04, v.05) with a real economy underneath it.

## Budget units

The budget axis is polymorphic. Each wave owns a different unit:

```text
WAVE 1     INFERENCE CALLS       how much planning can be afforded
WAVE 2     CONTEXT TOKENS        how much of the graph fits in the window
PERCEIVE   HOPS / MILLISECONDS   how far into the world a reach can go
```

"Tokens for cognition" was always about **wave 2**. The present is a prompt, and a prompt has a hard
size:

> **The present is what fit in the budget. Attention is affordability.**

Spread a token budget from the node's triples outward; what you can afford is what the model sees;
what you cannot afford, it does not know about.

## The graph the planning wave walks

Not cells — **triples and expressions, alternating**. A bipartite AND/OR graph:

```text
triple (a hole)                          ← OR  node: ANY prior expression could ground it
  └── expression that produced it        ← AND node: ALL of its requires must be grounded
        └── its requires (triples)       ← OR  nodes
              └── expressions...
                    └── triple satisfied by observed world state → LEAF (closure)
```

This is backward chaining / goal regression, which is what "reverse traversal from the finals"
actually is once written down.

**The existing `Alloc` weight semantics already encode AND/OR:**

```text
OR  node (triple)      weights SUM TO 1   conserve — choosing among alternatives
AND node (expression)  weights ALL 1      broadcast — every child is needed
```

Conserve-vs-broadcast — written to separate cognition from space — turns out to be
disjunction-vs-conjunction. The kernel needs no new concept: the cell's `Frontier` decides, and the
kernel stays oblivious.

## The frontier is a query, not a stored map

```text
triple's frontier       = the POSTING LIST for that triple    ← index lookup, ranked by
                                                       retrieval_weight × overlap
expression's frontier   = its Requires                        ← stored on the expression
world place's frontier  = its adjacency                       ← stored topology
```

`Frontier(origin, impulse) map[string]Alloc` already permits this. The retrieval frontier is
*computed by index intersection at traversal time* and the kernel never knows.

## The plan loop (descent)

Runs inside the workflow goroutine. **The floor is one LLM call, never zero.**

```text
STATE CHANGE
  ↓
1  INDEX     the target's ungrounded requires → gather precedent           (FREE)
2  PRESENT   bidirectional wave; precedent + world state                   (FREE)
3  PLAN      ONE LLM CALL — mandatory. framing + expression.
             may adopt retrieved precedent wholesale.
4  DRAIN     place every parsed node (the placer worker)                   (FREE)
5  INDEX     new holes may have precedent → place exact matches            (FREE)
6  CLOSED?   → EXPRESS
   FIXPOINT? → abort UNRESOLVED
   BUDGET?   → abort UNRESOLVED
   else      → back to 2 (another call)
```

**Steps 1, 2, 4, 5 are free. Only step 3 costs money.** The entire game is making step 5 do more work
over time so step 3 runs fewer times.

### Drain fully, check closure after

The LLM may emit more than the plan needs. **Place all of it anyway** — the extra nodes carry
`expects`, and expects are gradeable bets. Discarding them discards learning signal, and placement is
free. Closure is checked *after* the drain, never during.

### The index runs BEFORE the first call

That is what lets call #1 close the plan once the being knows the task. If the index only ran after
the first emission, the model would be inventing a bridge it already had.

## Why the floor is one call, not zero

> **The irreducible LLM call is the FRAMING, not the plan.**

Precedent can supply the plan. It cannot supply the judgment that **this precedent applies here**. A
state change means something new happened, and deciding what it means — does the target still hold,
is this the same situation, does what worked before work now — *is* the `observation → projection`
arrow, and that arrow always has the model in it.

It is also the guard against the failure a zero-call path would create: an agent blind-replaying
precedent, confidently doing the wrong thing because the situation *looked* the same. **Nothing is
executed on precedent alone.**

### What the call is for changes with experience

```text
day 1     invent the structure AND the particulars           → N calls
day 100   instantiate known structure with new particulars   → 1 call
```

The model's job shifts from **generating a bridge** to **binding a known bridge to this instance**.
That is a smaller job, which is why one call suffices.

### The metric

```text
passes-to-closure:   N  →  1
                           ↑  fully learned for this task class
```

A bounded floor is a far more interpretable curve than an asymptote at zero.

> **Deferred: reflexes.** A very probable path may eventually skip the framing call entirely. That
> is a future optimization and is explicitly **not** in scope. The floor stays at 1.

## Termination

```text
pass cap    hard ceiling on LLM calls per plan
FIXPOINT    the hole set is unchanged after a full drain → the model is not making
            progress → abort. THIS is the real rail; a cap alone lets it spin N times
            emitting nothing useful.
budget      spend exhausted
```

**Unresolved is a legitimate outcome.** Nothing is emitted to the world, and the being says so.

## The express path (ascent)

Closure reached → walk the bridge forward and execute.

```text
fire Runnable() acts      →  motor (no model, no present — pure execution)
observations return       →  decomposer → world state (LAW 1) → grader
```

**Descent creates final causes. Ascent creates efficient causes.** The descent said what each state
is *for*; the ascent discovers what actually *produced* it.

## Attention is a goal, not a mechanism

The world-agent **holds impulses until an agent perceives them** — they are not pushed (see
[world-agent-v2](../world-agent-v2/README.md)). Which raises the obvious objection: *Michael talks
and Skyra never listens.*

The answer is not a push channel and not a peripheral-vision hack:

> **Bootstrap the current-state and expected-state graphs with an ACTIVE BRIDGE between them. Then
> perception is already bootstrapped.**

The being perceives because it has a **standing target that requires perceiving**. It hears Michael
not because an interrupt fired, but because *knowing what Michael is doing is something it wants*,
and the plan to get it includes looking. A bridge that never closes.

```text
attention is not a mechanism.  ATTENTION IS A GOAL.
```

No special-case machinery, no interrupt path, nothing bolted onto the loop — it is the ordinary
workflow wave, running against a target the genome seeded.

**And the genome already has the slot.** v.05's `being ~name skyra ~purpose I think, respond, and
relate on behalf of the system` — **purpose IS the bootstrap target state.** The thing that was
always going to come back is the thing that seeds the standing bridge.

## The workflow wave is a live goroutine

It stays alive **while world state agrees with the plan**. It owns its own node-firing loop: it
fires planners as nodes become realizable and motors as acts become runnable. The only thing outside
it is the kill switch.

### The kill condition — precisely

```text
expected change (an expect resolved TRUE)   → plan continues. this IS the plan working.
purely additive change (news)               → plan continues. incompleteness is not a sin.
a bad bet nothing depended on               → GRADE it. do not kill. it was a cheap loss.
a LOAD-BEARING triple CONTRADICTED          → KILL the goroutine. replan.
```

**Load-bearing = some unexecuted act `requires` it.** That is exactly "closure broke" — already the
watcher's predicate, computable, no heuristics. A 0.4 `interpret` bet failing is a lesson; a
satisfied `requires` going false is an emergency.

### The restart is not a reset

The killed wave's plan **is still in the expression graph**. A new wave rediscovers it through the
same reverse index — the nodes are there, their expects are there, their grades are there. The being
does not start over; it re-derives, with one more piece of evidence than it had before.

This is why the bridge is *derived* state and never needs to be preserved: it can always be found
again.

## Cost scales with surprise — literally

Steady state is **one call per unmodeled state change**. The wave only dies when reality does
something the plan did not account for:

```text
well-modeled world   →  plan survives  →  few kills   →  few calls
surprising world     →  plan dies      →  many kills  →  many calls
```

The being's running cost is **proportional to how wrong it is**. That is the homeostat from
`one-kernel.md`, no longer an aspiration but an arithmetic consequence of the floor-of-1 and the kill
condition.

## Three changes the kernel needs

**1. A typed failure channel.** Not `error` — the fold must distinguish, because the responses differ:

```text
BUDGET-EXHAUSTED   couldn't afford to explore     → retry with more budget
NO-PRECEDENT       the index came back empty      → a GENERATION REQUEST. not a failure —
                                                    it is the wave asking the model a question.
CYCLE              already visited                → dead, prune
UNREACHABLE        distributed: node didn't answer → fold partially, and honestly
```

**2. Memoize on triple key.** Two acts often require the same state. Memoizing turns the walk from a
tree into a **DAG** — correct (same triple, same subtree) and a large constant-factor saving. The
visited set becomes a memo table, not a rail.

**3. Concurrent descent, deterministic fold.** OR children are independent; AND children are
independent; everything is parallel. The fold is AND/OR — commutative and associative — so
determinism costs only sorted edge keys and a `(score, key)` tie-break.

## Open

- Is exact-match precedent re-runnable as-is? Lean: **placeable only when its own `requires` are also
  satisfied in current world state** and its payload was fully determined. Otherwise it is a
  near-match and goes into the present for the model to adapt. That keeps direct placement honest —
  you only replay a step when the situation it ran in is genuinely the same situation.
- The placer is where **parallel width** is born, and width is money. Throttle it against the spend
  account rather than letting the DAG's shape dictate cost.
