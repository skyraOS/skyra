# Behaviors

**Date:** 2026-08-06
**Status: SPEC.** Carries `2026-08-03/behaviors.md` and absorbs `2026-08-03/assembly.md` —
ASSEMBLE is not a behavior, it is a walk EXPECT builds, so it is a section here rather than a doc.

> **One kernel. All the math is behavior.**

A behavior is a `(Perturb, Observe)` pair coupled to a node for one walk. The kernel runs it and
cannot tell one from another.

```text
PERTURB   READS.  decides WHERE. selection, activation, trust, budget, cost, ordering.
                  writes nothing.
OBSERVE   WRITES. decides WHAT IS TRUE. commits to the graph, ACCUMULATES sends, folds.
                  never decides where. it does not touch the world.
```

Full rule, and why eviction depends on it: [phases.md](phases.md).

```text
1  INTERPRET   grades. writes EDGES, never nodes.
2  EXPECT      plans. writes NODES. the only writer of state.
3  ACT         closes by ACK.
4  BOUNDARY    crossings.

   WALK        not a behavior. execution walks HEAT; retrieval walks DO NOT.
```

---

## Every node asks the same two questions

There is no dispatch table keyed on kind × phase. There is one shape, and the answer to the second
question *is* what a phase would have been:

```text
Perturb:
  1. AM I CLOSED?             → yes: promote nothing.
  2. WHAT WOULD CLOSE ME?     → build that space and return it.
```

```text
ACTION, open        what closes me is an ACK.
                    → a send-space. it accumulates the send; the boundary fires it.

EXPECTATION, open   TWO answers, in order:
                    children not all closed          → a walk over the unclosed ones
                    children closed, self not closed → a request: "what would close me?"

ROOT                an expectation like any other. standing, so never closed, so it always
                    descends.
```

The expectation's two answers are what the iterating `Perturb` is for: walk the children, fold, ask
again, and now the answer is different. One frame, two phases, no branching outside the node.

**`Perturb` builds a space; it does not accumulate.** The accumulating happens when that space runs
and folds, in `Observe`. If `Perturb` accumulated directly, an evicted descent would leave sends
queued and the eviction guarantee would be gone.

### Descent advances through what is closed

```text
DESCEND   pass a sibling only if it is CLOSED. stop at the first open one.
          → the frontier is wherever the descent stops.
```

"Cannot advance unless closed" is about moving PAST a node, not descending INTO one.

Sequential execution falls out with no `requires` field:

```text
E has [a1, a2], parent policy = sequential
  wave 1   descent reaches a1. open. stop. request the send. the ack closes it.
  wave 2   descent passes a1, reaches a2. open. stop. request the send.
```

One action per round trip, legislated by nobody. **Concurrent policy turns it off** — the parent
declaring concurrent lets descent pass an unclosed sibling, and both go out in one wave.

### A send is a Space, so the fire record is a transcript hop

A send crosses the kernel, so it gets a hop, and the hop carries the `WaveID`. *"Did this action
fire, and in which wave"* is a transcript query; no separate send log is needed.

```text
ack = −1                                   DIDN'T RUN.        retry, or the action is wrong.
ack = +1, no effects citing W yet          IN FLIGHT.         wait. do not re-fire.
ack = +1, effects arrived, self not closed RAN, DIDN'T WORK.  → EXPECT.
```

### Two traversals, and only one heats

```text
EXECUTION traversal   the active subgraph. THIS is attention. it heats.
RETRIEVAL traversal   EXPECT reading history for context. reads only. HEATS NOTHING.
```

Without the split, looking something up makes it likelier to be executed, and the being spirals
into whatever it happened to consult. **Retrieval may never be the reason attention is there.**

It will erode by accident if unguarded — both are the same `Walk` type, and whether it bumps is
one line.

### Retrieval competence is itself an expectation, and it is private

Retrieval can be done well or badly, so it is something the agent should get better at — which in
this model means an expectation with a perceive, graded like anything else.

**It does not violate the heat rule.** The rule protects *content* from gaining activation by being
looked at:

```text
THE CONTENT WALKED   config paths, prior plans, history.  MUST NOT HEAT.
THE RETRIEVAL NODE   "am I good at finding things here."  a different node. may heat.
```

Heating *"I am good at finding config paths"* does not heat the config path. A competence node is
not content. The walk it performs still heats nothing.

**Private is structural, not a preference.** A shared retrieval expectation is a node two agents
both write, and EXPECT-is-the-only-writer is a per-agent rule — sharing it creates a mutable node
with two owners. It would also not work: good retrieval strategy depends on the shape of *your*
graph.

These belong in **their own mount**, so meta-expectations do not compete with world expectations for
activation and are not swept into differentiation.

### Its perceive is the re-trap

Constitutive: no expectation without one, and the obvious candidate is the bad one.

```text
WEAK    "did the process eventually succeed?"    long, indirect, confounded. a process
                                                 fails for reasons unrelated to what it
                                                 was handed.

SHARP   "did the process trap AGAIN for the      local, immediate, falsifiable. a re-trap
        same condition?"                         on the same need means retrieval missed.
```

Cheap, and genuinely about the retrieval rather than about the work. Without something like it the
retrieval node is unfalsifiable decoration — the failure `shape.md` names for confidence edges,
reappearing on a node.

**The first internal competence modelled as an expectation.** Others follow — interpretation
quality, budget allocation — and each one is more meta-structure in the graph. The mount is what
keeps that from polluting attention.

## Two kinds of Space, and only one is a behavior

`Perturb` returns one Space per call, so **a behavior cannot walk a frontier itself.** Walking is
delegated:

```text
HANDLE   a durable node + a behavior.   the coupling.
WALK     a frontier + a fold.           EPHEMERAL ONLY. no node, no behavior, no dispatch.
```

A `WALK` holds a cursor and yields members one at a time. It writes nothing durable, so an evicted
walk has changed nothing at all.

**Members are opened with a leaf behavior.** Opening them with the namespace default makes each
member run a nested turn — the walk recurses, the tree explodes, and EXPECT fires once per nested
turn. A walk *visits* its frontier; it does not re-enter it.

## Workers nest, and nesting is what orders them

A worker's `Perturb` sets up; its `Observe` does the work. Post-order fires the **inner `Observe`
first**, so ordering is depth:

```text
descend    EXPECT.P  →  INTERPRET.P  →  WALK …
ascend     … WALK.O  →  INTERPRET.O  →  EXPECT.O
```

**This is not the scheduler's job and must not become it** — round-robin decides which execution
steps next, never which cognition happens first. Ordering lives in the structure, where it cannot
be tuned.

---

## 1 · INTERPRET — impulse → support edges

Interpreting and grading are one call: scoring `relationship ∈ [-1,1]` answers *how does this bear
on that*, which cannot be done without reading both.

```text
P   grab the graph. build the frontier from the impulse and the active position.
    return a WALK over it. reads only; traverses nothing itself.

O   receive the COMPLETE candidate set in one piece.
    exact match against an expectation's perceive → WRITE THE SUPPORT EDGE.
    accumulate everything unresolved as residual.
```

**It writes EDGES and never NODES.** A stored reading is a claim about the world and therefore
state — see [shape.md](shape.md).

**It is re-entrant.** The boundary's answer arrives as an impulse of its own; the next pass matches
exactly and writes the support edges, producing no residual. The consumer wave is the same
behavior.

### There is no exact pass on a raw payload

Matching is a comparison, a raw payload has no structure to compare, and giving it one is a model
call. dnav3 got pure comparison at grading only because its decomposer had already produced a
triple — judgement first, comparison after. Merged, there is no earlier pass.

```text
STRUCTURED     world returns, boundary answers.   exact comparison is real here.
UNSTRUCTURED   user text, another agent's prose.  RESIDUAL by definition.
```

So the whole frontier coming back as residual is the **correct** answer for raw input, not a
placeholder.

### What counts as a gap — two kinds

`2026-07-27/workers.md` uses low confidence as a proxy for unplaceable. The proxy breaks:
*"seatbelt is tight"* is read **confidently** and is still unplaceable.

```text
1  COULD NOT READ IT          low confidence in the reading itself       unintelligible
2  READ IT, IT MATTERS,       high |relationship|, and NO NODE EXISTS    MISSING STRUCTURE
   NOWHERE TO PUT IT
```

Kind 2 drives replanning, and sign is irrelevant to it. **Negative support alone must not trigger
EXPECT** — you tried to stand, you did not, nothing else came back: no unplaceable impulse, nothing
to build from, and structure invented there is Toy 4's forbidden fabrication. Failure that surfaces
nothing teaches a probability, not a structure.

Which puts weight on the perceive. *"Am I standing?"* returns the negative and nothing else. *"Why
am I still seated?"* can return the seatbelt. **An expectation whose perceive only checks its own
closure can fail forever without ever producing a gap.**

### What it may and may not see

```text
MAY SEE      the live expectations. it cannot score against them otherwise.
MAY NOT SEE  the AUTHORED CONFIDENCE on them, or the target.
```

The protection is that **the predictor cannot influence the assessor** — it does not know what
number it is meant to hit. This is weaker than dnav3 LAW 6, which withheld the bets entirely; that
cannot survive a merge, since one call is one context.

**The cost, detected not prevented:** the interpreter can converge toward whatever is live. Log
both, from the first run, because neither is derivable afterward:

```text
revision rate               falling → converging, OR self-confirming. AMBIGUOUS.
cross-form contradiction    rising while revision falls → the reading is bending
```

### The immune system belongs here — a direction

**Not built.** Recorded because the home is already obvious and choosing it later would be worse.

Detecting *"this does not belong"* is a grading judgement on incoming material, which is what
INTERPRET is. It also already carries the only self-monitoring in the design — revision rate and
cross-form contradiction, logged to catch the interpreter bending toward whatever is live. An immune
system is that instinct generalized from *is my reading drifting* to *is something acting on me.*

```text
DETECTION   INTERPRET's. it reads everything arriving and scores it.
RESPONSE    NOT INTERPRET'S. it writes edges, never nodes, and quarantining or killing
            is an action. detection here, response through EXPECT and BOUNDARY.
```

That split is the same one that keeps INTERPRET from being a writer of state, so the immune system
does not get an exemption from it.

**One signal to keep from the start**, since it is cheap and not derivable afterward: a burst of
novel irrevocable requests is itself an observation. Rate of first-use irrevocable actions is a
perceive on the planner's health, which makes flooding a detectable event rather than a successful
one.

## 2 · EXPECT — what closes this node

The expectation worker: CREATES, REFINES, BINDS, and AUTHORS the confidence that a gap will close.

### Two triggers, both mechanical

```text
children all closed AND self not closed   →  REPLAN
observation load exceeds allocated budget →  DIFFERENTIATE   (differentiation.md)
```

Both halves of the first are derived closure — **no model is needed to know it should run**, only
to do the work. It satisfies the fabrication guard structurally: children closing means their
perceives *returned something*, so evidence has arrived by construction.

It works at a leaf. An expectation whose only child is its perceive, where the perceive came back
*"not standing"*: perceive closed, self not closed, EXPECT runs.

```text
P   build the frontier around the attachment point — a RETRIEVAL walk, budgeted.
O   AddExpectation, with its perceive. AddAction to bind and order.
```

### It is re-entrant, so nothing else places the result

```text
wave 1   EXPECT requests.  outbox ← "what would close this node, given this history"
wave 2   the answer arrives. EXPECT writes the children.
```

**And the descent finds the node on its own.** The node's children are closed and the node is open,
so a descent from the standing root passes everything closed and stops there. The frontier is where
the answer belongs.

### Acquiring an action it does not have

Not a third trigger — part of planning. EXPECT decides what would close a node, and sometimes the
action required does not exist in the graph.

```text
RETRIEVE   walk for an existing action. the ordinary path.
     ↓     nothing found
FIND       search a catalog of proven blocks. BIND one as a new action node.
     ↓     nothing found
BUILD      generate an adapter. THE FALLBACK, and currently not permitted.
```

`design-docs/world/act-service.md` argued this before the current model existed: *"Let the AI reach
for the most proven thing. If it can't find one, then make it itself."* Find outranks build because
a proven block carries trust accumulated from other people's use, and generation starts at zero
every time.

**But borrowed trust is not grounding.**

> A catalog's claim that a block is proven is not grounding. A perceive still has to fire.

The same rule that governs a process's self-report, generalized to a supplier. A proven block may
set the **authored confidence** prior. It may not supply **support**, which only comes from a
perceive firing in this agent's world. So finding gets a better starting bet, never a shorter path
to closure.

### Discovery is a retrieval. It must not become a capability.

The failure to avoid, given that a gateway can front a thousand tools behind one endpoint:

```text
GATEWAY AS AN ACTION NODE   the process gets ONE tool that reaches a thousand apps.
                            permission-by-exposure is gone and it is a deny list again.

CATALOG AT PLAN TIME        EXPECT retrieves over the catalog. what gets BOUND is a
                            specific tool, its own action node, its own Q edge.
```

Discovery and execution are different phases, which is the split this doc already runs on. **The
gateway belongs to the first and may never appear in the second.**

### A description is an expectation, checked immediately after firing

Discovered descriptions are third-party text. A description is a **claim about the world** — *"this
sends an email"* — and this model has exactly one way to hold such a claim:

> Every claim about the world is falsifiable by construction, because creating one is the only way
> to state one.

So a description does not enter as trusted text. It enters as **the expectation the tool is graded
against**, and the check runs immediately after the invocation rather than at leisure: interception
already sees the observation side, so what the tool did is compared against what it said it would
do, and a revertable mismatch is reverted.

**This buys supplier calibration for free.** A catalog whose descriptions overclaim produces high
authored confidence against low assigned support on the same node — `actions.md`'s calibration
comparison, pointed at a vendor.

A description may still inform selection. It may never be read as a directive — `THE COUPLING IS
CHOSEN BY THE AGENT, NEVER BY THE DATA`, at the affordance registry.

### Revertable and irrevocable

Grading after the fact only helps if the effect can be undone, so the split has to exist. **It is
derived, not declared** — BOUNDARY already owns crossings, so *"does this leave the runtime"* is
read off the action's mount and cannot disagree with the structure.

```text
REVERTABLE     worktree effects inside the confinement. throw the clone away.

IRREVOCABLE    ↳ CROSSES THE BOUNDARY   sent, posted, paid. read off the mount.
               ↳ WRITES THE GRAPH       append-only. LOCAL, AND STILL FINAL.
```

**Confinement is what makes the first row true.** `rm -rf` inside a confined clone is undoable
because the clone is disposable; without that decision, "local" would not imply "revertable" at all.

The second irrevocable class is the one that is easy to miss: EXPECT writes nodes, nothing is ever
removed, and a bad expectation is worse than a bad send — a send is one event, an expectation shapes
every future plan that retrieves it.

### The gate is at admission, not invocation

> **A tool is reviewed when it enters the runtime. Not every time it is used.**

Per-request approval is the standard way human-in-the-loop fails: it floods, and a flooded gate is a
click-through. Reviewing once, at the point a capability is granted, costs the user one decision per
tool instead of one per action.

```text
NOT THE SAME ADMISSION   impulse admission must not consult attention — FIFO, never ranked.
                         TOOL admission is a capability grant and is deliberately gated.
                         two different doors; do not let the vocabulary merge them.
```

```text
COVERS         the catalog as an entry point for NEW capability. nothing arrives unseen.
DOES NOT       misuse of a capability already granted.
```

The residual is the ordinary posture of any assistant holding credentials: having accepted *"this
agent may send email,"* the remaining question is whether the **first binding** of an irrevocable
tool to a new expectation deserves its own gate. Undecided, and the natural next decision.

### It alters the TREE. It never builds the STACK.

A stack is what you get from walking a tree. A worker that could construct one directly would be
choosing execution order by hand. Growth is picked up by **walking again** — append-only makes that
safe, and dnav3's reconciler is the precedent: *"if the target is not met, and something new was
learned, RE-PLAN."*

### TEMPORARILY UNGATED

**EXPECT runs on every impulse for now, gap or no gap.** A condition nobody has watched fire cannot
be calibrated, so the turn runs an unconditional pipeline and the state machine needs no branch.

```text
RUN      unconditional.
RECORD   the gap test's verdict per impulse, unused.
COMPARE  what EXPECT did  vs  what the gate would have permitted.
```

While ungated, structure will be **fabricated** — that is the point of watching it — and the graph
is append-only, so the bloat is permanent. These early graphs are disposable and should be treated
that way from the start.

## 3 · ACT — execute a primitive

```text
P   nil. an action is a leaf.
O   ACCUMULATE the send. the boundary fires it.
    effects come back as RAW IMPULSES.
```

No residual: execution happens or it does not, and failure is an impulse like any other. Perceive
is not a separate job — it is an act whose primitive reads instead of writes.

Because nothing fires inside the wave, **an evicted wave can never have acted.**

## 4 · BOUNDARY — advertise · deliver · accept

```text
P   resolve the mount name. reach was decided at construction; nothing is checked here.
O   accumulate the crossing. the Receipt reports the SEND, never what the target concluded.
```

`show-affordance` stays a static registry read. If answering costs an inference call, discovery
stops being cheap enough to do speculatively.

---

## The assembly walk

Absorbed from `2026-08-03/assembly.md`. Not a behavior — **a `WALK` with a budget**, built by
EXPECT to answer *what would close this node*. A **retrieval** walk, so it heats nothing.

### Three jobs, and one is a different wave

```text
ASSEMBLY    walk the activated region, gather until the budget is spent.  MECHANICAL.
RENDERING   turn what was gathered into the prompt.                       MECHANICAL.
GRADING     score what came back.                                         A DIFFERENT WAVE.
```

Grading cannot be the assembly wave's `Observe`: at that moment nothing has come back, because what
would be graded *is* the response, and the response does not exist until this wave produces the
prompt that asks for it.

### Two directions, opposite

```text
BUDGET         DESCENDS.  it must be known BEFORE deciding whether to promote.
ACCUMULATION   ASCENDS.   a post-order fold; it does not exist until unwind.
```

An accumulation at a frame is only what was gathered *beneath* it, and on descent nothing beneath
has run. So the budget is carried **down** instead, and the descent decision needs no total:

```text
child_budget = (parent_budget − parent_cost) × weight
descend iff  child_budget ≥ child_cost
```

Per-execution state — budget remaining, tokens spent, what has been gathered — lives **on the
Stack**. Shared across frames within one execution is not the shared-state problem: single-threaded,
one owner, and the execution is the unit of eviction.

### The reverse index is activation, and it is a cache

```text
THE GRAPH   authoritative.
THE INDEX   DERIVED. rebuildable. deletable. cheap to be wrong about.
```

`2026-07-27/workers.md` already names it: *activation SELECTS which expectations an impulse is
relevant to — **tokens***. So the index belongs to `Perturb`, and activation is **additive only**:
it raises what is in view and never dims what arrived.

---

## Exact / residual is not a phase boundary

```text
PHASE        Perturb / Observe    read vs write       WHEN in the walk
RESOLUTION   exact / residual     here vs deferred    WHAT HAPPENS to each item
```

Exact work happens in **both** phases; residual is spent in **neither**, because a wave never calls
the model. It is a per-item outcome: each item resolves, or accumulates on the stack and is spent
at the boundary afterward.

```text
PERTURB'S RESIDUAL   "I could not build a frontier from this."   ← this IS the gap signal
OBSERVE'S RESIDUAL   "I could not match or score this."
```

**Anything that has to leave the agent folds to the root and goes out from there** — two kinds of
leaving, *doing* and *asking*, and the wave does neither.

## What a turn looks like

```text
impulse arrives  →  the one stack, entered at the standing root
  INTERPRET   read it, grade what matches, accumulate the rest
  EXPECT      (ungated: always)
              ↓ boundary — the model is called HERE, outside any wave
  ACT         accumulate sends
              ↓ boundary fires them
              ↓ effects arrive AS RAW IMPULSES
  (a new wave, back to INTERPRET)
```

## What is left for a scheduler

```text
ADMISSION   FIFO. never consults attention.
ORDERING    round-robin, and with one stack there is nothing to rotate.
COUPLING    each node names a CONDITION; the agent owns what answers it.

WHEN TO START     nothing yet triggers a wave no impulse asked for.
INFERENCE BUDGET  exact-first bounds what inference SEES; nothing bounds how many boundary
                  calls a turn may make.
THE STEP LOOP     nothing drives the runner outside tests.
```
