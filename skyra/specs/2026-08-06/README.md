# 2026-08-06

**A consolidation, not a fork.** `2026-08-03` is retired: its live content moved here, its
superseded content stays behind as history. Nothing should cite it going forward except where the
reasoning matters.

The rule is inherited from `2026-07-27` and applied to the docs as well as the mechanisms:

> **Strip complexity, do not add it.** Nothing carries across by default. A doc earns entry the
> same way a mechanism does — by being required, not by being coherent with the rest of the set.

Applying that to nine docs produced six. The compression was real: `graph.md` led with a rule that
had become false, and `assembly.md` described a behavior that turned out to be a walk.

`syscalls.md` was added afterward, the first doc to enter this set by being required rather than
carried.

---

## The loop

Unchanged since 07-27. It is why anything here exists.

```text
an impulse comes in
    │
    ▼
it is graded against an expectation
    │
    ▼
given that expectation, what is the next ACTION
    │
    ▼
given that action, what is the next EXPECTATION
    │
    └──────────────────────────▶ (an impulse comes in)
```

## In this set

| doc | owns | status |
|---|---|---|
| [kernel](kernel.md) | the substrate. `Perturb` / `Observe` closed over `Space`, the returned transcript, and what the kernel cannot do. | **BUILT** — `src/skyra/spacetime/kernel` |
| [shape](shape.md) | one family of state, three structural states, two edge kinds, and the only writer. | **SETTLED** |
| [execution](execution.md) | durable vs ephemeral, the anchor, **ONE STACK**, interception, the worktree's lifetime, and the program/process direction. | **FOUNDATION** |
| [syscalls](syscalls.md) | the process-facing surface. **two traps out, one signal in, no grading verb.** | **SPEC** |
| [phases](phases.md) | **`Perturb` reads. `Observe` writes.** and why eviction depends on it. | **RULE** |
| [behaviors](behaviors.md) | the four, the one question every node asks, and the assembly walk. | **SPEC** |
| [closure](closure.md) | **closed is +1, this execution. `P` is retrieval.** | **RESOLVED** |
| [differentiation](differentiation.md) | structure from overload. EXPECT's second trigger. | **WORKING TOWARD** |
| [questions](questions.md) | decisions nobody has made. renumbered. | **LIVE** |

## What happened to 2026-08-03

```text
ENTERED INTACT   kernel · phases · closure · differentiation

MERGED           graph + families        → shape.md
                 assembly                → a section of behaviors.md

REWRITTEN        execution   two rules were sitting under a "working toward" heading and
                             apply to what is built today — see below
                 questions   six of ten were answered and still listed as open
                 behaviors   absorbed assembly; the dispatch table was superseded by
                             "every node asks the same two questions"

DID NOT ENTER    graph.md    its live half is the composition argument, which is now the
                             opening of shape.md. its headline rule — "every edge is an
                             observation" — had become false, and a banner three lines down
                             does not fix a headline.
```

**The two rules that were mis-filed as directions**, now in `execution.md` proper:

```text
TWO BUDGETS      ONE STACK limits ATTENTION, not limbs. one subgraph active does not mean
                 one process alive.

INTERCEPTION     route and record, never filter. "in charge of attention" is true for
                 ordering and delivery, and must not become true for admission.
```

## What is built

```text
src/skyra/spacetime/kernel                     the kernel. trampoline, iterating Perturb,
                                               returned transcript.
src/skyra/cognitive-agent/skyra-dna-v4/graph   namespace, handles, walks, four behaviors
                                               with PURSUE stubbed. one turn runs.
```

Three states that were rules are now unconstructible: an expectation without a perceive, an edge
without a grade, an action grown into an expectation.

## The critical path

```text
Q1   THE BOUNDARY   nothing spends the outbox, so no support flows, the fold has never
                    run, and calibration is never computed.

Q14  COST           no Node carries a token count. it unblocks the assembly walk,
                    differentiation, AND retrieval — one quantity, three dependents.
```

Nothing else on the questions sheet moves until one of those does. **Take Q1 first** — COST is
engineering, the boundary is validation, and building further on arithmetic that has never executed
is what would have to be unwound.

## Still the live reasoning

`2026-07-27` — trust, actions, workers, dynamics, and the five toys. It has not been re-derived
here. `truth.md` is superseded by [closure](closure.md); `expectations.md` by [shape](shape.md).
