# The Expression Grammar

**Date:** 2026-07-14
**Status:** SCRATCHPAD — worked out jointly in one session. Concrete enough to argue with,
not yet a commitment.
**Companions:** `two-graphs.md` (the cycle this emits into), `self-authored-loop.md`
(prediction/diff), `one-kernel.md`, `commit-structure-and-rewind.md`, `README.md`

This is what the LLM emits. Unlike the other notes it is a single artifact rather than
his-then-mine — it was built in a back-and-forth. Where a specific call was made, it's marked.

---

## The Shape

XML **envelope**, line-oriented **triples** inside.

```xml
<framing>Michael wants the auth refactor to compile with tests passing.</framing>

<act id="a1" target="file:/src/auth.go" primitive="write">…content…</act>

<act id="a2" target="build:/src" primitive="compile"
     requires="exact file:/src/auth.go exists true"/>

<act id="a3" target="being:michael" primitive="ask"
     requires="exact test:auth_test.go status fail">The refactor broke two tests — revert?</act>

<expect of="a1">  exact      0.95  file:/src/auth.go   exists   true       </expect>
<expect of="a1">  exact      0.90  file:/src/auth.go   contains "func Authenticate" </expect>
<expect of="a2">  exact      0.85  build:/src          status   ok         </expect>
<expect of="a2">  exact      0.70  test:auth_test.go   status   pass       </expect>
<expect of="a3">  exact      0.99  being:michael       replied  true       </expect>
<expect of="a3">  interpret  0.40  being:michael       state    satisfied  </expect>
```

### Why XML outside

Self-delimiting. A truncated emission still yields every complete element before the cut.
JSON is disqualified: one missing brace and the entire payload is unparseable, and
*"parse failure kept motor from acting at all"* already happened once.

### Why triples inside, not tagged nodes

- **Per-line failure isolation.** A malformed line costs one fact, not the document.
- **Truncation is free.** Out of output tokens → fewer predictions. Nothing corrupted.
- **The emitted form IS the diffable form.** Predicted set minus observed set. Both directions.
  No tree to walk.
- **Token cost.** At 500 nodes, `<node>…</node>` ceremony is real money.

---

## The Triple

```
<mode>  <confidence>  <subject>  <predicate>  <object>
```

### Mode — two, not three

```
exact      the incoming impulse is matched by VALUE COMPARISON
interpret  the incoming impulse is run through the ALIGNER, then compared
```

**Michael's correction: `ask` is not a mode. `ask` is an expression.** It is an act that
*produces* an observation; the interpretation of that observation is a `requires` on a
downstream expression. Dropping `ask` as a mode is what makes the grammar close.

And the mode is **not a property of the claim** — it is a *routing instruction for the incoming
impulse*. It says how an observation resolves against this expectation. That is why it can't be
inferred from the namespace: `being:kathy location beyond:door` is directly observable;
`being:michael state happy` is not. Same prefix, different channel.

Mode is also the ledger weight (from `two-graphs.md`, channel fidelity):

```
exact       weight ~1.0    free, instant, unambiguous
interpret   weight ~0.3    noisy; the aligner translated it
(no source) weight  0      FORBIDDEN
```

### Confidence — it buys the scoring rule

The leading float makes the emission a set of **probabilistic claims**, which means it can be
scored with a **strictly proper scoring rule** (Brier / log). Honest calibration becomes the
score-maximizing strategy *by construction* — no hand-tuned anti-vagueness penalty.

But a proper scoring rule only grades claims you **made**. Silence is unpunished by Brier. So:

```
Brier over stated claims       → makes hedging honest
coverage penalty for silence   → makes silence expensive   (observed-but-not-predicted)
```

**Both are required.** Neither alone. This decomposition only became clear once the syntax
carried a probability.

### Subject/object namespaces carry identity provenance

```
file:/src/auth.go   proc:1234   url:…   coord:478,479   ← WORLD-GIVEN. coreference is FREE.
being:michael       concept:trust                       ← MODEL-ASSIGNED. needs the aligner.
```

The prefix tells you at parse time, with zero inference, whether coreference costs anything.
This is how the diff scales to hundreds of nodes: ordinals (`michael[1]`) do not scale;
world-addressable IDs do.

It also creates the right pressure — the model earns cheap, reliable credit for claims it can
ground in things reality can adjudicate.

---

## Two Positions, One Object

```
expect     a triple claimed TRUE AFTER acting
requires   a triple that must be TRUE BEFORE acting
```

Same grammar, same modes, same three-valued logic, same grading path. **`requires` is an
observable world state** — *"given a particular world state, do X."* You branch off expected
world state, not off act IDs.

### Three-valued

```
satisfied  → fire
falsified  → drop
unknown    → WAIT      ← not drop. this is what makes a guard a conditional.
```

Same logic as `indeterminate` in the world contract.

---

## Parallelism: Derived, Never Declared

No `<parallel>`. No `<sequence>`. An act with no unmet `requires` runs **now**; the runtime
topo-sorts the rest. Parallelism is read off the graph.

**State-preconditions beat act-dependencies** for one decisive reason: *a state can already be
satisfied by the world.* If the file already exists, `a2` doesn't need `a1` and runs immediately.
An act-ID dependency would serialize you for nothing.

## Conditionals: Guards, Never Branches

No `<if>`, no `<else>`. `if/else` is two acts with opposite guards. The model never writes control
flow — it writes **conditions under which things are true**, and the world decides which fires.

Guards are checked against **observation**, never against belief. A conditional cannot be resolved
by fiction.

## No Loops

**The emission is a DAG. Iteration lives in the outer cycle.** If something needs retrying, the
next `observation → projection` cycle projects it again. This keeps the parser and scheduler
trivial forever.

---

## Cardinality — `of=` Does Individual And Joint

```xml
<expect of="a1">     exact  0.95  door:front   state     open                     </expect>
<expect of="a2">     exact  0.99  being:kathy  heard     "walk through the door"  </expect>
<expect of="a1 a2">  exact  0.60  being:kathy  location  beyond:door:front        </expect>
```

Single ID = individual. Multiple IDs = **joint**. No new keyword.

**Why joint must be first-class** (Michael's case): the door opened, Kathy heard, Kathy didn't
walk through. Individual expectations met; joint expectation failed.

```
individual failure  → the ACT was wrong           (the door didn't open)
joint failure       → the CAUSAL MODEL was wrong  (open door + told her ≠ she walks through)
```

Different edges, different lessons. Without the joint you'd punish *"open the door"* because Kathy
is stubborn.

Free signal: the confidence gradient (0.95, 0.99 → 0.60) measures **how much of the outcome is out
of your hands**, and it's gradeable.

## The Falsifiability Law

> **Every `act id` must appear in at least one `expect`, or the emission is rejected.**

An act that projects nothing is unfalsifiable, therefore unlearnable. Enforced by the parser, not
by discipline.

---

## The Aligner Discipline

> **The aligner runs ONCE, on the observation side only.**

```
prediction   a literal stated claim.           NEVER interpreted.
observation  translated into that vocabulary.  Interpreted ONCE.
diff         claim  vs  translated observation.
```

This is **not** aligner-vs-aligner. It is claim-versus-interpreted-observation, one model call, on
the side where the world's answer actually lives. The being commits to a specific proposition in
advance; reality — read once, as faithfully as possible — decides.

Which makes `interpret` safe on a projection, and it's where the compounding asset lives:
predicting `being:michael state satisfied` at 0.40, having the reply resolve it true, and learning
you were *underconfident about what satisfies him*. That's a model of a person, being corrected.

**Validation rule:** an `interpret` expectation needs an observation channel or it never matures.
An `interpret` on `being:michael` with no act touching `being:michael` is a bet with no due date →
warn, or auto-schedule the ask.

**That is the mechanism that makes the being ask.** Not politeness — its bet won't settle otherwise.

---

## Closure — The Planning Run

After the emission, show the model:

```
CURRENT world state      (observed)
PROPOSED world state     (what the acts + expects imply)
EXPECTED world state     (the target)
UNGROUNDED REQUIRES      ← the residual. THIS is the task.
```

> **A plan is closed when every `requires` unifies with either (a) a triple in the current
> observed world state, or (b) an `expect` of another act in the plan.**

**This is a reachability check. Set membership. No judge, no threshold, no graded overlap** — which
retires the hardest open question in `README.md` ("*'reached current state' probably cannot require
exact node identity… may need graded overlap*"). It can. The syntax made it exact.

Not closed → emit again. **Every emission is a node in the traversal.** This is the recursive
reverse bridge from `README.md`, now with a computable stopping condition.

### Show the residual, not just the states

The model's job on pass *n+1* is not "improve the plan." It is **"ground these two triples."** That
precision is why it converges in a few passes instead of wandering.

### Two kinds of hole — only one is fixable by thinking

```
STRUCTURAL   a requires that no act produces
             → fixable by MORE INFERENCE. emit an act that produces it. another pass.

EPISTEMIC    a requires whose truth is UNKNOWN (not false)
             → NOT fixable by inference. no amount of thinking reveals the world.
             → needs an OBSERVATION act. execute, observe, re-plan. NEW CYCLE.
```

**Perceive-before-act falls out of the closure rule** — rather than being learned at the cost of 95
inference calls and a scolding, which is what happened on the first live run.

### Termination

```
pass cap    hard limit on inference passes
fixpoint    the hole set didn't change between passes → stuck, abort
budget      spend / token exhaustion
```

Unclosed → **`unresolved`. Nothing is emitted to the world.** No act fires until the plan closes —
preserving the single-motor-boundary rule (speculative branches must not act merely because they
were explored).

### Every pass is a graded node

Each emission carries **its framing (its prompt) and its expects (its bets)**. When those expects
resolve against observation, **that pass is graded**.

So the planner learns to plan. Passes whose bridging acts produce well-calibrated predictions earn
retrieval weight; passes that emit confident nonsense decay. Next time a similar hole appears, the
framings that closed holes well are what get surfaced.

That is the per-node self-authored prompt from `self-authored-loop.md` — and the nodes turn out to
be **planning passes**. The thing being selected for is *the ability to bridge*.

### Cost

`N passes × 1 call`, N small and capped, each call emitting a large chunk of plan. Against 95 calls
for six messages on the first live run. The multi-pass loop is not a cost regression — it is most
of the savings, because passes terminate on a **computable closure condition** rather than on an
LLM deciding it is done.

---

## Grammar, current state

```
ONE TRIPLE FORM   <mode> <confidence> <subject> <predicate> <object>
TWO MODES         exact | interpret            (ask is an EXPRESSION, not a mode)
TWO POSITIONS     expect | requires            (same object, different tense)
THREE-VALUED      true | false | unknown       (unknown = wait, never drop)
CARDINALITY       of="a1"    individual → grades the ACT
                  of="a1 a2" joint      → grades the CAUSAL MODEL
LAW               every act appears in ≥1 expect, or rejected
CLOSURE           every requires unifies with current-state OR another act's expect
NO                parallel / if / else / loops — derived, guarded, outer-cycle
ALIGNER           runs once, observation side only. may translate; may never judge.
```

## Open

- **Are edges and attributes one form?** `s p o` seems to cover both — a predicate that resolves to
  a known relation *is* an edge. Probably unify; the parser gets smaller.
- **Invariants** (`dir:/src unchanged-except file:/src/auth.go`) — a predicate, or its own
  construct? Lean: just a predicate. Fewer concepts. But it's where boldness lives, so it may earn
  distinct scoring.
- **Maturity per line** — when does a bet come due? Possibly implied by mode (`exact` resolves on
  next observation; `interpret` when the channel produces one). If so, maturity is free, the same
  way coreference was.
- **What surfaces the framings** on the next similar hole — retrieval over graded passes. Not
  designed yet. This is the `observation → projection` arrow, which is still the one arrow with
  everything hard inside it.
