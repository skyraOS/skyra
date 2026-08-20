# Closure And Planning

**Status:** DESIGN · part of [skyra-dna-v3](README.md)

## The closure condition

> **A plan is closed when every `requires` unifies with either**
> **(a) a triple in the current observed world state, or**
> **(b) an `expect` of another act in the plan.**

**This is set membership.** No graded overlap, no threshold, no judge.

That retires the hardest open question in the original planning sketch — *"'reached current state'
probably cannot require exact node identity… may need graded overlap rather than exact graph
identity."* It can be exact. The grammar made it exact: `requires` and `expect` are the same triple
vocabulary over world-addressable IDs, so closure is a reachability check.

Dangling `requires` = a **hole**, and the hole is *named*. You know exactly which triple is
unsupported.

## The planning run

After each emission, show the model:

```text
CURRENT world state      (observed)
PROPOSED world state     (what the acts + expects imply)
TARGET world state       (the goal)
UNGROUNDED REQUIRES      ← the residual. THIS is the task.
```

The model's job on pass *n+1* is not "improve the plan." It is **"ground these two triples."** That
precision is why it converges in a few passes instead of wandering.

**Every emission is a node in the traversal.** Not closed → emit again.

## Two kinds of hole — only one is fixable by thinking

```text
STRUCTURAL   a requires that no act produces
             → fixable by MORE INFERENCE. emit an act that produces it. another pass.

EPISTEMIC    a requires whose truth is UNKNOWN (not false)
             → NOT fixable by inference. no amount of thinking reveals the world.
             → needs an OBSERVATION act. execute, observe, re-plan. NEW CYCLE.
```

**Perceive-before-act falls out of the closure rule** rather than being learned at the cost of 95
inference calls and a scolding, which is what happened on the first live run.

## Termination

```text
pass cap    hard limit on inference passes
fixpoint    the hole set didn't change between passes → stuck, abort
budget      spend / token exhaustion
```

Unclosed → **`unresolved`. Nothing is emitted to the world.**

No act fires until the plan closes. This preserves the single-motor-boundary rule: speculative
branches must not act merely because they were explored.

## Every pass is a graded node

Each emission carries **its framing (its prompt) and its expects (its bets)**. When those expects
resolve against observation, **that pass is graded.**

So the planner learns to plan. Passes whose bridging acts produce well-calibrated predictions earn
retrieval weight; passes that emit confident nonsense decay out. Next time a similar hole appears,
the framings that closed holes well are what get surfaced.

The thing being selected for is **the ability to bridge**.

## The watcher

Plan validity is **computable**, because closure is set membership. So there are no "is my plan
still good?" heuristics and no model call:

> **After every world-state update, re-run the closure check.**

```text
observation → update world state → re-check closure
    closes + unexecuted acts remain   → keep executing
    stops closing                     → replan (locally if the hole is local)
    closes + target reached           → done
    target changed                    → replan
```

A `requires` you believed satisfied that just became **false** opens a hole. That *is* the
invalidation signal.

> **The watcher watches whether the plan still closes.** That is the entire scheduler, and it is
> one predicate.

## Impulse handling

An unrelated impulse — *"want to go to the mall"* arriving mid-refactor:

```text
updates world state                              ✓
matches no expect → no grade                     ✓
purely additive → no penalty                     ✓  (see scoring.md)
does NOT break closure of the refactor plan      ✓  → the refactor keeps running
```

It does one thing: it may introduce a **new target** — a second plan. Now two closed plans compete
for the act budget.

**That is the only place goal-amplitude machinery is needed.** Not to compute a gradient — just to
arbitrate **which closed plan acts next**. A priority function over competing closed plans, not a
potential field.

## Cost

`N passes × 1 call`, N small and capped, each call emitting a large chunk of plan — against 95 calls
for six messages in the first live run.

The multi-pass loop is **not** a cost regression. It is most of the savings, because passes terminate
on a **computable closure condition** rather than on an LLM deciding it is done.

## Open

- **Local bridge repair vs. full reroll** — what makes a hole "local"?
- Whether the target itself is a far `expect` (same type, different distance) or needs its own
  representation. Lean: same type. A standing goal is a far projection; the bridge is the chain of
  nearer ones; the next-step bet is the projection actionable *now*. One type, graded at different
  maturities.
