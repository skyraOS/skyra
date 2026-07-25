# Two Graphs — The Causal Graph Dissolves

**Date:** 2026-07-14
**Status:** SCRATCHPAD — raw working thought, not yet a design or implementation commitment
**Companions:** `self-authored-loop.md` (the prediction/diff loop), `one-kernel.md`,
`commit-structure-and-rewind.md`, `README.md`

This is the largest deletion in the set. It removes the causal layer as a separate thing.

---

# Part I — Michael

## Two Knowledge Graphs

```text
observed world state KG     what I perceived
action world state KG       what I did / expressed
```

## The Cycle

```text
projected world state subgraph  is linked to    action world state subgraph
action world state subgraph     is tied back to observed world state subgraph
observed world state subgraph   then creates    projected world state subgraph
```

So: two KGs. One for expression. One for observation.

## The Consequence

**The causal graph and the knowledge graph dissolve. The causal graph IS the graph.**

```text
projection  ->  expression
expression  ->  observation
observation ->  projection
```

---

# Part II — Claude

## Why It Works: A Diff Requires One Type

This is the mechanism, and it's the reason the collapse is real rather than cosmetic.

**You cannot subtract two things of different kinds.** That was the actual defect in the causal
layer: a `CausalNode` is an *annotated sentence*; a perception is *world state*. Two different
kinds of object. So comparing them required a judge — and the judge is the thing poisoning the
ledger (see `self-authored-loop.md`, "why the subgraph is the whole design").

In the two-graph cycle, a **projection and an observation are the same kind of object**: both are
world-state subgraphs, same vocabulary, same node types. The diff becomes a set operation.

```text
BEFORE   causal node (annotated sentence)  vs  perception (world state)   → needs a JUDGE
AFTER    projected subgraph                vs  observed subgraph          → SET DIFFERENCE
```

That type-match is the whole win.

## Causality Stops Being Asserted

Today the model must **claim** causality: the annotation grammar (`michael[1] {asked about}[1→2]`),
retain/compose/requery, `expandCausalBeliefs` issuing two more calls to assert efficient and final
beliefs. The LLM is the *author* of the causal structure — which makes the causal structure an
opinion.

In the cycle, **the runtime records causality and the model never claims it.** "I projected P, I
emitted E, I then observed O." The links are simply there, because that is the order things
happened.

The causal graph is not a layer over the knowledge graph. It is **the shape of the loop through
it.**

So the four causes stop being annotations and become **positions in the cycle**:

```text
projection    = FINAL       what I am aiming at
expression    = EFFICIENT   what I did
observation   = MATERIAL    what actually is
```

Nobody has to say so. It is structural.

### What this deletes

```text
CausalNodes            (as a separate store)
AitiaSet               (present / efficient / final triple)
the annotation grammar (ordinals, verb-group braces, edge marks)
validateStructuralOp   (retain / compose / requery)
expandCausalBeliefs    (two inference calls per turn)
structuralCandidateIDs / renderCausalStructure / the offered-candidate apparatus
```

That is most of the 95 calls from the first live run.

## The Sharper Cut: Mine vs The World's

Michael's line is observe / action. One notch over is a better boundary:

```text
MINE          projections + actions     things I generated
THE WORLD'S   observations              things reality returned
```

A projection is my bet. An action is my emission. Both originate in me. **Only observation crosses
in from outside.**

Draw the line there and an invariant falls out that is worth more than any discipline enforced by
hand:

> **Learning only ever happens at the boundary between the two graphs.**

A grade is *by definition* a cross-boundary diff. Which means **you structurally cannot learn from
your own output** — the LLM-grading-LLM failure becomes a **type error**, not a rule anyone has to
remember. The grounding guarantee stops being a principle and becomes physics.

This also subsumes the grade-provenance note from `one-kernel.md`: provenance is no longer a field
to be maintained on a grade — it is which side of the boundary the evidence came from.

## Where The Grade Lands

```text
projection ──emitted──> action ──returned──> observation
     └───────────────── diff ───────────────────┘
```

The diff lands **on the edge**: *this projection, via this action, produced this much error.*

Which is exactly where activation already lives. A/S/F/N was always edge-owned. Now **the edges ARE
the causal claims**, and they carry their own accuracy. No new machinery is required — the economy
that already exists is the one that grades the loop.

## What This Has NOT Solved

### 1. `observation → projection` is one arrow with the entire cognitive act inside it

Framing, retrieval, goal-setting, inference — all of it lives in that single edge. That is correct
to draw that way, but it means the simplification has not *removed* the hard part. It has
**isolated** it.

```text
projection  ->  expression      collapse (motor). easy.
expression  ->  observation     the world's job (reafference). free.
observation ->  projection      ← EVERYTHING HARD IS HERE
```

All remaining design lives on that arrow.

### 2. Where does the standing goal live?

A projection as drawn is a **one-step bet** — *"the door will open."* A goal is a **standing
target** — *"Michael has water"* — that must survive many cycles. Both are projected states. They
have different lifetimes and different maturity conditions. Collapse them and the goal is
overwritten every cycle, which is the exact continuity bug that started this whole line of work.

Proposed resolution: **same type, different distance.**

```text
standing target   a FAR projection          matures on satisfaction
the bridge        a chain of nearer ones    matures as it is crossed
next-step bet     the projection I can act on NOW    matures on the next observation
```

One type, one graph, graded at different maturities. That is where DESIRED (from
`self-authored-loop.md`) fits without introducing a fourth kind of thing.

## Open

- Is a projection's *maturity condition* a property of the projection itself (a due-date it carries),
  or is it inferred from its distance along the bridge?
- Coreference across the boundary is still a model call (`{a glass of water}[3]` vs `{the cup}[7]`).
  The aligner survives this collapse. Keep it demoted — it may say *"these corefer,"* never *"you
  did well."*
- Does the action graph need its own node types at all, or is an action just an observation of
  myself? (If the world reports my own effects back to me, maybe expression is observed too — and
  there is only one graph with a provenance bit.)
- What is the residual role of `Frame.Text`? Probably: a rendering of the current projection, for
  humans. Not load-bearing.
