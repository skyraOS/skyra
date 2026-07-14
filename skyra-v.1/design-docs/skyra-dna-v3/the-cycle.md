# The Cycle — Two Graphs, One Loop

**Status:** DESIGN · part of [skyra-dna-v3](README.md)

## Two graphs

```text
MINE          projections + actions      things I generated
THE WORLD'S   observations               things reality returned
```

Michael's cut was observe / action. One notch over is the load-bearing line: **what came from me**
versus **what came from outside me.** A projection is my bet; an action is my emission. Both
originate in me. Only observation crosses in.

## The loop

**Write it starting from observation.** The being does not begin by wanting something; it begins by
being told something.

```text
observation  ->  projection  ->  expression  ->  observation
```

```text
observation   what actually is                  MATERIAL cause
projection    what I am aiming at / expect      FINAL cause
expression    what I did                        EFFICIENT cause
```

**The causes are positions in the loop, not annotations.** Nobody emits them. The runtime records
*"I observed O, I projected P, I emitted E, I then observed O′"* and the links are simply there,
because that is the order things happened.

Written this way, the hard arrow is where it belongs — **first**. `observation → projection` is the
one edge with the entire cognitive act inside it (see [retrieval.md](retrieval.md)). The other two
are cheap: projection → expression is a collapse, and expression → observation is the world's job.

## The two graphs are mutually dependent — the loop is circular

It is tempting to say the expression graph is the asset and the observation graph is just the
transcript. **That is wrong**, and the error matters.

```text
learning TAKES PLACE in the expression graph        grades live on expects and framings

learning IS ONLY POSSIBLE via the observation graph an expect's claim REFERS TO observed state;
                                                     its grade IS a diff against an observation

the observation graph IS SHAPED BY the expression graph
                                                     you only observe what you acted to observe
```

The observation log is **not** a neutral recording of the world. What got observed is a function of
where the being *looked*, and where it looked was determined by its own perceive acts. Two beings in
the same room produce different observation logs, because they attend to different things. **The log
is already a fact about the being.**

So the dependency is not hierarchical, it is **circular** — and it is circular in exactly the shape
of the loop above. Neither graph is prior. Neither is disposable.

Strip the observations and the expression graph is numbers with no provenance: unfalsifiable,
unauditable, unreproducible. Which is precisely the ungrounded thing this architecture exists to
prevent.

### Which makes LAW 2 auditable

Every observation that resolves a bet points back at the bet it resolved. That link is not
bookkeeping — it is the evidence for the grade.

> **Every grade must name the observation that produced it. A grade with no observation behind it
> is, by definition, self-adjudicated.**

This turns the law from a discipline into a **query**: walk the expression graph, find any grade
whose `ResolvedBy` is empty, and you have found a place where the being graded itself. You can scan
for the poison.

> **The causal graph is not a layer over the knowledge graph. It is the shape of the loop through
> it.**

This is why v2's entire causal apparatus dissolves: `CausalNodes`, `AitiaSet`, the annotation
grammar, `validateStructuralOp`, `expandCausalBeliefs`. The model never *claims* causality, so
causality is never an opinion.

## Why the collapse works: a diff requires one type

The real defect in v2 was a **type mismatch**. A `CausalNode` is an annotated sentence; a perception
is world state. You cannot subtract one from the other — so comparing them required a **judge**, and
the judge is what poisoned the grades.

```text
v2   causal node (annotated sentence)  vs  perception (world state)   → needs a JUDGE
v3   projected subgraph                vs  observed subgraph          → SET DIFFERENCE
```

A projection and an observation are **the same kind of object**: world-state subgraphs, same
vocabulary, same node types. The diff becomes arithmetic.

## LAW 1 — Observation is the only writer

> **Expression has no write access to world state. Observation is the only writer.**

Writing a file **projects** `file:… exists true`. The observation graph does **not** contain
`file:… exists true` until something came back and said so.

This kills the most damaging agent failure in production: *"I wrote the file"* → proceed as if it
exists → build three steps on top → the whole chain is fiction. Structurally impossible here. The
being can be wrong, but it **cannot believe its own output into reality**.

### What the rule buys

The observation graph is a **fold over the observation log**. That is its definition, and four properties
fall out of it:

```text
deterministic   same observation log → same observation graph
replayable      rewind = fold to time T
auditable       every node traces to the observation that produced it
fiction-free    nothing self-authored can ever enter
```

Event-sourcing does not have to be designed for the world half. It **is** the fold.

### The resulting epistemics

```text
"I emitted `write file X`"   expression graph. self-authored. TRUE BY CONSTRUCTION —
                             I always know what I did.

"file X exists"              observation graph. ONLY after something told me so.
                             I never know what RESULTED until I look.
```

This gives `indeterminate` (see [ephemeral-world-workers](../world/)) a natural home: *I emitted it,
I don't know whether it committed, so my observation graph claims nothing.* Perceive to resolve.

### Evidence the rule is right

The first live run's one demonstrable piece of learning was: *after moving on stale location
information and being criticized, Skyra changed behavior to perceive first, then move.* It spent 95
inference calls and a human's disapproval to learn, badly, what this invariant provides by
construction.

### The cost, and why it is already paid

The objection: now you must *look* at everything you care about, and a fifty-effect action needs
fifty perceptions.

It doesn't, because **`Outcome.Effects` is reported by the world, not by the being.** That report
crossed the boundary — it *is* an observation, at fidelity ~1.0, and it carries all fifty effects in
one return. Perception is only needed where the world couldn't tell you.

## LAW 2 — Learning happens only across the boundary

> A grade is **by definition** a cross-boundary diff.

Which means **grading your own output is a type error**, not a rule anyone has to remember. The
LLM-judging-LLM failure becomes structurally unrepresentable.

This also subsumes grade-provenance: provenance is not a field to maintain — it is **which side of
the boundary the evidence came from.**

## Where the grade lands

```text
projection ──emitted──> action ──returned──> observation
     └───────────────── diff ───────────────────┘
```

On the **edge**: *this projection, via this action, produced this much error.*

Which is where edge-owned activation already lives. **The edges ARE the causal claims, and they
carry their own accuracy.** The economy grades the loop with no new machinery.

## What this does not solve

**`observation → projection` is one arrow with the entire cognitive act inside it.**

```text
projection  ->  expression      collapse. easy.
expression  ->  observation     the world's job. free.
observation ->  projection      ← EVERYTHING HARD IS HERE
```

The collapse **isolates** the hard part rather than removing it. See [retrieval.md](retrieval.md) —
that arrow is retrieval over graded precedent, and it is the only place the model does real work.

## Relationships

- [world-assertions-and-policy.md](../cognition/world-assertions-and-policy.md) — the same
  observed/expected split, written 2026-06-29. v3's representation is that document's.
- [prediction-falsifiability-invariant.md](../laws/invariants/prediction-falsifiability-invariant.md)
  — still binding.
- [expression-grammar.md](expression-grammar.md) — what gets emitted into this loop.
