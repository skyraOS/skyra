# Behaviors

**Date:** 2026-08-05
**Status: SPEC.** The enumeration, split by phase. Written before a scheduler, because you
cannot schedule what you have not named.

> **One kernel. All the math is behavior.**

A behavior is a `(Perturb, Observe)` pair coupled to a node for one walk. The kernel runs it and
cannot tell one from another — it has no method that would let it ask.

```text
PERTURB   READS.  decides WHERE. selection, activation, trust, budget, cost, ordering.
                  writes nothing.
OBSERVE   WRITES. decides WHAT IS TRUE. commits to the graph, ACCUMULATES sends,
                  folds. never decides where. it does not touch the world.
```

Full rule, and why eviction depends on it: [phases.md](phases.md).

```text
1  INTERPRET   grades. writes EDGES, never nodes.
2  EXPECT      plans. writes NODES. the only writer of state.
3  ACT         closes by ACK.
4  BOUNDARY    crossings.

   WALK        not a behavior. execution walks HEAT; retrieval walks DO NOT.
```

## Every node asks the same two questions — 08-06

There is no dispatch table keyed on kind × phase. There is one shape, and the answer to the
second question *is* what a phase would have been:

```text
Perturb:
  1. AM I CLOSED?             → yes: promote nothing.
  2. WHAT WOULD CLOSE ME?     → build that space and return it.
```

Filled in:

```text
ACTION, open        what closes me is an ACK.
                    → a send-space. it accumulates the send; the boundary fires it.

EXPECTATION, open   TWO answers, in order:
                    children not all closed        → a walk over the unclosed ones
                    children closed, self not closed → a request: "what would close me?"

ROOT                an expectation like any other. standing, so never closed, so it always
                    descends.
```

The expectation's two answers are what the iterating `Perturb` is for: walk the children, fold,
ask again, and now the answer is different. One frame, two phases, no branching outside the node.

**`Perturb` builds a space; it does not accumulate.** The accumulating happens when that space
runs and folds, in `Observe`. If `Perturb` accumulated directly, an evicted descent would leave
sends queued and the eviction guarantee would be gone.

### Descent advances through what is closed

```text
DESCEND   pass a sibling only if it is CLOSED. stop at the first open one.
          → the frontier is wherever the descent stops.
```

"Cannot advance unless closed" is about moving PAST a node, not descending INTO one. You go into
an open expectation — that is the point; you cannot move beyond one.

Sequential execution falls out with no `requires` field:

```text
E has [a1, a2], parent policy = sequential
  wave 1   descent reaches a1. open. stop. request the send. the ack closes it.
  wave 2   descent passes a1, reaches a2. open. stop. request the send.
```

One action per round trip, legislated by nobody. **Concurrent policy is what turns it off** — the
parent declaring concurrent lets descent pass an unclosed sibling, and both go out in one wave.

### A send is a Space, so the fire record is a transcript hop

Because a send crosses the kernel, it gets a hop, and the hop carries the `WaveID`. So
*"did this action fire, and in which wave"* is a transcript query, and no separate send log is
needed. That is what separates the three states:

```text
ack = −1                                  DIDN'T RUN.        retry, or the action is wrong.
ack = +1, no effects citing W yet         IN FLIGHT.         wait. do not re-fire.
ack = +1, effects arrived, self not closed RAN, DIDN'T WORK. → EXPECT.
```

### Two traversals, and only one heats

```text
EXECUTION traversal   the active subgraph. THIS is attention. it heats.
RETRIEVAL traversal   EXPECT reading history for context. reads only. HEATS NOTHING.
```

Without the split, looking something up makes it likelier to be executed, and the being spirals
into whatever it happened to consult. This is the additive-only rail with a real boundary: a trace
may raise attention; **retrieval may never be the reason attention is there.**

It will erode by accident if it is not guarded — both traversals are the same `Walk` type, and
whether it bumps is one line.

**ASSEMBLE is not on this list.** Its whole job is traversing a frontier until a bound is hit,
which is what a `WALK` does — the bound is a budget instead of a list length. It is a `WALK`
parameterized by cost, not a behavior. [assembly.md](assembly.md) still describes it correctly;
only its category changed.

## Two kinds of Space, and only one is a behavior

`Perturb` returns exactly one Space and the kernel calls it once per frame, so **a behavior
cannot walk a frontier itself** — it would get one candidate and stop. Walking is delegated:

```text
HANDLE   a durable node + a behavior.   the coupling. what a Behavior is attached to.
WALK     a frontier + a fold.           EPHEMERAL ONLY. no node, no behavior, no dispatch.
```

A `WALK` is a **continuation carrying the remaining frontier**:

```text
walk{A,B,C}   P → takes A, returns walk{B,C}
walk{B,C}       P → takes B, returns walk{C}
walk{C}           P → takes C, returns nil        ← base case
                  O → fold C
                O → fold B into what came back
              O → fold A into what came back
```

One hop per member, so the transcript stays the chain it already is and law 1 survives. This is
`kernel-alteration.md` §2's continuation-passing note doing real work, and §4 anticipated it
exactly: *"a grouped Space may implement its behavior as an internal state machine that returns
further ephemeral continuation Spaces. Each continuation still crosses the kernel as one Space."*

**It also retires `MANY ENTRIES`.** That item was waiting on the iterating-`Perturb` decision and
does not need it: a frontier expressed as a chain needs no fan-out.

A `WALK` writes nothing durable — it folds into ephemeral accumulation only. So an evicted walk
has written nothing at all.

## Workers nest, and nesting is what orders them

A worker's `Perturb` sets up; its `Observe` does the work. Post-order then fires the **inner
`Observe` first**, so ordering constraints are expressed as depth:

```text
descend    EXPECT.P  →  INTERPRET.P  →  WALK …
ascend     … WALK.O   →  INTERPRET.O  →  EXPECT.O
```

Nothing can propose before interpreting, because interpretation is deeper. **This is not the
scheduler's job and must not become it** — round-robin decides which execution steps next, never
which cognition happens first. Ordering lives in the structure, where it cannot be tuned.

The cost is that a worker's writes land in one call at the top of its fold rather than
incrementally. That unevenness is real and cheap: **no step contains a model call** (a wave never
calls the model), so the big `Observe` is edge writes bounded by frontier size. Not worth a
mechanism.

---

## 1 · INTERPRET — impulse → support edges

**Merged 08-05.** Interpreting and grading were one call written down twice.

In v3 they could be separate because grading was arithmetic — exact triple comparison, no model.
Here, the residual pass *is* a model call: scoring `relationship ∈ [-1,1]` answers *how does this
bear on that*, which cannot be done without reading both the impulse and the expectation. Grading
was already interpretation.

The mechanical payoff: under *a wave never calls the model*, keeping them apart costs two boundary
crossings per impulse. Merged, one.

**It is a worker, not a walk.** Its `Perturb` does not traverse — it cannot, since one `Perturb`
yields one Space. It builds the frontier and hands off:

```text
P   grab the graph. build the frontier from the impulse's terms and the active position.
    return a WALK over it. reads only; writes nothing; traverses nothing itself.

O   receive what the walk realized — the COMPLETE candidate set, in one piece.
    exact match against an expectation's perceive → WRITE THE SUPPORT EDGE.
    accumulate everything unresolved as residual.
```

**It writes EDGES and never NODES.** Revised 08-06: the interpretation node is gone
([families.md](families.md)), because a stored reading is a claim about the world and therefore
state — which would make the model a writer of state. EXPECT is the only writer of state, and
that division is checkable: anything calling `AddExpectation` is the expectation worker.

So the model's output is now confined to a **scalar on an edge** or a **proposal**. It cannot
file a reading as a fact. A reading can only become an expectation — which is unconstructible
without a perceive — so every claim about the world is falsifiable by construction.

Seeing the whole candidate set at once is the right shape for this job specifically: a reading
should be decided against everything gathered, not committed incrementally as candidates arrive.

Residual comes from both phases — `Perturb` cannot build a frontier from a term with no history
(*"I could not navigate by this"*), and `Observe` cannot match what does not match.

**It is re-entrant, and that is what keeps the count at five.** The boundary's answer arrives as
an impulse of its own. The next pass matches exactly — the answer is in hand — so it writes the
support edges and produces no residual. The consumer wave is the same behavior, not a new one.

### What it emits

Support edges, and nothing else.

```text
"michael said grab me a glass of water"   the impulse. what arrived. not stored as a claim.
        ↓  support [-1,1]
an existing expectation                   the grade. a scalar, not an assertion.
```

*"being:michael wants water"* is a claim about the world and cannot be written here. If the being
is to hold it, EXPECT must create it as an expectation — with a perceive, and therefore wrong-able.

### TEMPORARILY UNGATED — 08-05

**EXPECT runs on every impulse for now, gap or no gap.** The trigger below is specced and not
enforced.

The reason is measurement: a condition nobody has watched fire cannot be calibrated, and the
alternative is tuning it by argument. So the turn runs an unconditional pipeline —
INTERPRET → EXPECT → PURSUE — and the state machine needs no branch.

**But evaluate the gate anyway and record the verdict.** Running ungated tells you how EXPECT
behaves; recording what the gate *would* have said tells you whether the condition is any good.
Skipping the evaluation costs a second run to find out.

```text
RUN      unconditional. every impulse gets INTERPRET and EXPECT.
RECORD   the gap test's verdict per impulse, unused.
COMPARE  what EXPECT did  vs  what the gate would have permitted.
```

Two things to hold while it is ungated. Structure will be **fabricated** — EXPECT with no gap has
no evidence to build from, which is Toy 4's forbidden move, and that is the point of watching it.
And the graph is **append-only**, so the bloat is permanent: these early graphs are disposable, the
way dnav3's per-episode observation graph is, and should be treated that way from the start.

### What counts as a gap — TWO kinds, and 07-27 has only one

`2026-07-27/workers.md` uses low confidence as a proxy for unplaceable:

> relationship ≈ 0 · confidence LOW → it was read and could not be placed → GAP

**The proxy breaks.** "Seatbelt is tight" → "I have my seatbelt on" is read *confidently* and is
still unplaceable — there is no seatbelt node anywhere. Confident reading, no home, and the doc
has no state for it.

The discriminator is **placement**, not confidence:

```text
1  COULD NOT READ IT          low confidence in the reading itself       unintelligible
2  READ IT, IT MATTERS,       high |relationship|, and NO NODE EXISTS    MISSING STRUCTURE
   NOWHERE TO PUT IT
```

Kind 2 is what drives replanning, and sign is irrelevant to it — a confidently-read positive with
no home is the same signal. What matters is that it bears on something active and has nowhere
to attach.

**And this is why negative support alone must not trigger EXPECT.** You try to stand, you do not,
nothing else comes back: no unplaceable impulse, so nothing to build from. Structure invented
there would be fabricated, which is the move [Toy 4](../2026-07-27/toy-runthroughs.md) forbids
outright — *"it may not fabricate intermediate changes."* dnav3's reconciler makes the same
distinction: the structure comes from `build:/ output app.src` — the new observation — not from
the failure being negative. **Failure that surfaces nothing teaches a probability, not a
structure.**

Which puts unusual weight on the perceive. *"Am I standing?"* returns the negative and nothing
else. *"Why am I still seated?"* can return the seatbelt. **An expectation whose perceive only
checks its own closure can fail forever without ever producing a gap.**

### What it may and may not see

07-27 narrowed blindness from *"the assessor sees no expectations"* to this, and the narrowing is
what makes the merge survivable:

```text
MAY SEE      the live expectations. it cannot score against them otherwise.
MAY NOT SEE  the AUTHORED CONFIDENCE on them, or the target.
```

The whole protection is that **the predictor cannot influence the assessor** — the assessor does
not know what number it is meant to hit, so it cannot hit it.

This is **weaker than dnav3 LAW 6**, which withheld the open bets entirely. That cannot survive a
merge: one model call is one context, and expectations in the prompt condition the reading. Stated
here so it is a choice rather than a discovery.

### What the merge costs

**Toy 1 gets worse.** The blind interpreter converging on a wrong reading becomes an interpreter
that can converge toward *whatever is live* — ambiguous input resolving toward the standing
expectation, the being manufacturing a weak version of its own evidence.

Same posture 07-27 took when folding the action worker into the expectation worker: it cannot be
prevented by construction while the merge holds, so it is **detected**. The detector is a pair,
because neither half works alone:

```text
revision rate               falling → converging, OR self-confirming. AMBIGUOUS.
cross-form contradiction    rising while revision falls → the reading is bending
```

Log both from the first run. Neither is derivable after the fact.

## 2 · EXPECT — what closes this node

**This is the expectation worker.** All four of what `2026-07-27/workers.md` gives it:

```text
CREATES   a new expectation, when impulses arrive that nothing accounts for
REFINES   an existing one, by adding a child beneath it
BINDS     actions to expectations, and orders them
AUTHORS   the confidence that a gap will close    ← the child→parent edge (families.md)
```

### Its trigger is MECHANICAL — 08-06

```text
children all closed  AND  self not closed   →  EXPECT
```

Both halves are derived closure. **No model is needed to know it should run** — only to do the
work. That is the first trigger the exact pass can evaluate alone; every other one needed a
judgement.

It satisfies the fabrication guard structurally rather than by rule: children closing means their
perceives *returned something*, so evidence has arrived by construction. Toy 4's forbidden move —
inventing intermediates with nothing new to go on — cannot fire.

It works at a leaf too. An expectation whose only child is its perceive, where the perceive came
back *"not standing"*: perceive closed, self not closed, EXPECT runs. *"I looked, and it isn't
true"* → replan.

### It is re-entrant, so no extra worker places the result

The model's children arrive in a **later wave**, as an impulse:

```text
wave 1   EXPECT requests.  outbox ← "what would close this node, given this history"
wave 2   the answer arrives. EXPECT writes the children.
```

Producing pass, then consuming pass — the same shape INTERPRET has. So `AddExpectation` keeps one
caller and the writer rule stays greppable.

**And the descent finds the node on its own.** The answer does not need to `chroot` back: the
node's children are closed and the node is open, so a descent from the standing root passes
everything closed and *stops there*. The frontier is where the answer belongs.

### ASSEMBLE is the walk EXPECT builds

Not a peer behavior. *"What would close this node"* is answered by a budgeted retrieval over
history through the token index — which is a `Walk` with a budget instead of a list length. That
settles `WHAT IS ASSEMBLE`, and it is a retrieval walk, so **it heats nothing**.

### It alters the TREE. It never builds the STACK.

The durable/ephemeral split makes this a hard line rather than a preference:

```text
EXPECT writes   the expectation tree — durable, append-only
NOTHING writes   the execution stack — it exists only as the shape of a traversal
```

A stack is what you get from walking a tree. If a worker could construct one directly it would be
choosing execution order by hand, which is the scheduler-authority problem in the one place we
have not already closed it.

So growth is picked up by **walking again**, not by patching what is in flight. Append-only makes
that safe — a traversal already underway sees the new material or does not, and both are
consistent ([execution.md](execution.md)). dnav3's reconciler is the precedent: *"plan → act →
observe → is the target met? if not, and something new was learned, RE-PLAN."* Learning mid-run is
absorbed by re-walking, never by editing the stack.

```text
P   build the frontier around where the gap was found — the attachment point and its
    neighbourhood. return a WALK over it. no residual: nothing here is being resolved.

O   AddExpectation, with its perceive. AddAction to bind and order.
```

It ordinarily wraps INTERPRET, and the nesting is what guarantees it cannot propose against a
reading that has not happened yet.

The only behavior that grows the graph's structure.

## 3 · ACT — execute a primitive

```text
P   nil. an action is a leaf.

O   execute the primitive against the world
    effects come back as RAW IMPULSES
```

No residual: execution happens or it does not, and failure is an impulse like any other.
Perceive is not a separate job — it is an act whose primitive reads instead of writes.

Because the effect is in `Observe`, **an evicted wave can never have fired an action.**

## 4 · BOUNDARY — advertise · deliver · accept

```text
P   resolve the mount name
    reach was decided at construction; nothing is checked here

O   spawn the crossing
    the Receipt reports the SEND, never what the target concluded
```

`show-affordance` stays a static registry read. If answering costs an inference call, discovery
stops being cheap enough to do speculatively.

---

## Exact / residual is real — but it is not a phase boundary

Two orthogonal axes:

```text
PHASE        Perturb / Observe    read vs write       WHEN in the walk
RESOLUTION   exact / residual     here vs deferred    WHAT HAPPENS to each item
```

Exact work happens in **both** phases. Residual is spent in **neither** — a wave never calls the
model. So it cannot be a phase split. It is a per-item outcome: each item either resolves, or
accumulates on the stack and is spent at the boundary after the wave.

```text
PERTURB'S RESIDUAL   "I could not build a frontier from this."   ← this IS the gap signal
OBSERVE'S RESIDUAL   "I could not match or score this."
```

That `Perturb` produces residual is the useful part: a term the index cannot place is exactly
*read, and could not be placed* — `workers.md`'s gap, arriving for free rather than needing its
own detector.

## Three kinds of wave

```text
PRODUCERS   accumulate residual, end at a boundary call    INTERPRET
CONSUMERS   fold a boundary answer back into the graph     INTERPRET (re-entrant) · EXPECT
NEITHER     fully mechanical, no residual possible         ACT · BOUNDARY
```

INTERPRET is both, which is not a special case — a producing pass has residual, a consuming pass
does not, and the behavior is the same either way.

## How a behavior gets chosen

Keyed on what kind of node you are standing on, and what the impulse is:

```text
DIRECTORY + impulse       INTERPRET     raw payload, world return, or a boundary answer
DIRECTORY + proposal      EXPECT
FILE      + executing     ACT
MOUNT     + crossing      BOUNDARY
```

**Not the data choosing its own behavior.** The agent owns the dispatch rule; a node cannot name a
behavior or reach the choice. It is a filesystem dispatching on file type through a table the OS
owns, and it stays safe because kind is derived from composition and a file can never become a
directory ([graph.md](graph.md)).

## What a turn looks like

```text
impulse arrives  →  gets its own STACK, always admitted
  INTERPRET   read it, grade what matches, accumulate the rest
              ↓ boundary — the model is called HERE, outside any wave
  INTERPRET   the answer arrives as an impulse; this pass writes the support edges
  EXPECT     only where a gap was found
  ASSEMBLE    gather under budget → prompt
              ↓ boundary
  ACT         execute
              ↓ effects arrive AS RAW IMPULSES
  (a new wave, back to INTERPRET)
```

## What is left for a scheduler

Most of it is already decided, and deliberately dumb:

```text
ADMISSION   always succeeds. at capacity, evict.      Runner.Admit
ORDERING    round-robin. reads nothing but Done().    Runner.Step
COUPLING    the dispatch table above.                 the agent's, not the scheduler's
```

Round-robin must stay dumb: scheduling by trust or activation would let the controller decide
which executions progress, which is the sealing failure one level up.

So a scheduler owns only:

```text
WHEN TO START     nothing yet triggers a wave no impulse asked for. dnav3's matured-bet sweep
                  has no equivalent, because "when the perceive runs" was never decided.
INFERENCE BUDGET  exact-first bounds what inference SEES; nothing bounds how many boundary
                  calls a turn may make.
THE STEP LOOP     nothing calls Runner.Step outside tests. questions.md Q4.
```

---

## Open

```text
COST                 no Node carries a token count. the budget has nothing to subtract, and
                     everything in ASSEMBLE waits on it.

INTERPRETATION CONTENT  questions.md Q10. an interpretation must name something checkable
                     against the expectation tree, or "is there a node for this?" cannot be asked.

IMPULSE KINDS        the dispatch table keys on one, and the set is not defined anywhere.
```
