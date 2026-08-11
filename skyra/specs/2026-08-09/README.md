# 2026-08-09

**A spec set, promoted from the `2026-08-08` whiteboard.** The rule that governed the promotion is
the one inherited from `2026-07-27` and applied to the docs as well as the mechanisms:

> **Strip complexity, do not add it.** Nothing carries across by default. A doc earns entry the same
> way a mechanism does — by being required, not by being coherent with the rest of the set.

What that stripped: the session narration. *"An earlier draft did X and was wrong"* is right for a
whiteboard and noise in a spec. What it kept: the **rejected alternatives**, one block per doc, one
line each with the reason — because those stop rediscovery, and several of them were rediscovered
more than once already.

## Nothing here was built

```text
2026-08-06   the spec for what RUNS.  v4. Perturb/Observe. still live for the code that exists.
2026-08-09   the spec for the DIRECTION. v5. not built, not started.
```

`2026-08-06` could mark `kernel.md` **BUILT** because it matched running code line for line. **No doc
in this set can say that.** Every one is SPEC or below, and six confident documents should not be
allowed to imply otherwise.

**On 2026-08-11 it was built** — `skyra-dna-v5`, twelve packages, driven by a real model. The docs
above are left exactly as they were written, because what they got wrong is only legible against what
they said. [build.md](build.md) is the report, and it is the only doc in the set with evidence
behind it.

> **The substrate holds.** Credit accumulates, hop profiles appear at range, and selection measurably
> improves with no weights. What is genuinely broken is narrower: `discover` cannot earn credit under
> the rules as written, a chain could rediscover itself, and refinement has no memory of what it
> already rejected.
>
> **Decomposition was not tested.** Every ask in the run was something the model already knew, so
> routing was never necessary — a bad measurement, not a result.

The fork is resolved by the convention this repository already uses: `skyra-dna-v4` stays and runs;
the new direction takes `skyra-dna-v5`. Nothing is deleted and nothing is migrated.

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
| [grammar](grammar.md) | SELECT / COMMIT, citation, the five questions, capability as scoping, the universal signature. | **SPEC** |
| [possibilities](possibilities.md) | the layering, what a possibility is, leaf vs orchestrator, composition, mint-on-closure, what a possibility sees. | **SPEC** |
| [kernel](kernel.md) | the active/durable split, the write chain, grants, the chain as transaction, permissions as engrams, monitor, impulses, roots and heads. | **SPEC** |
| [world](world.md) | entity, host, `here`, graders vs interpreters, affordance, no ground truth. | **SPEC** |
| [discover](discover.md) | the two mechanisms, create-on-rejection, ingress, the bootstrap, refinement. | **SPEC** |
| [credit](credit.md) | the write, conditional childhood, parent ownership, the walk, hop profiles, events-not-counts, selection as correction. | **SPEC** |
| [interfaces](interfaces.md) | the implementation shape for `skyra-dna-v5`. types, emissions, the frame. | **PROPOSAL** |
| [traversal](traversal.md) | what `Perturb`/`Observe` was carrying, and what pays for each item now. | **CHECKLIST** |
| [questions](questions.md) | decisions nobody has made, ordered by how much rests on them. | **LIVE** |
| [build](build.md) | what happened when it was built and run. what held, what broke, what was measured badly. | **BUILT** |

## What the set claims

```text
TWO FORMS               a possibility emits SELECT or COMMIT. incapacity is a PARSING property.
CARDINALITY IS THE      concurrency is how many possibilities one select names. no async, no
CONCURRENCY DECLARATION await, no execution flag. every group awaits.
CAPABILITY IS SCOPING   you can select what you can name; you can name what you were given.
NO GROUND TRUTH         one structure held locally by each observer. no third thing behind it.
ONE WRITE               at the parent's return. eviction safety falls out rather than being built.
NOTHING HAPPENING IS    the kernel holds all active work; the graph holds only what completed.
REPRESENTABLE           crash semantics collapse to nothing.
THE CHAIN IS THE        world effects are staged when asked for and realized when the root
TRANSACTION             accepts. materialization follows the same walk credit does.
NO SPECIAL CHANNELS     a permission is an engram and its approval is a possibility, so
                        approval strategies compete like everything else.
WHAT WORKED, NOT        a child engram forms on citation. exploration costs nothing permanent.
WHAT WAS TRIED
ALL LEARNING IS         selection learns; synthesis does not. possibilities are selected less,
TOPOLOGICAL             never improved.
ERRORS LOCAL,           conjunctive success distributes; conjunctive failure does not.
ACCEPTANCE BY HOPS
```

## The critical path

```text
Q1   BUILD SOMETHING   every doc here is derived and nothing has been run. the first build is
                       the learning substrate with everything else stubbed — it tests the claim
                       the rest depends on and dodges every other open item.
```

That is the only thing on the sheet that moves anything else. See [questions.md](questions.md).

## Validation

The five adversarial toys in `2026-07-27/toy-runthroughs.md` were run against this model and the
results recorded there beside the older lineage's:

```text
Toy 1   stability ≠ correctness                PASSES
Toy 2   record count ≠ independent grounding   DISSOLVES · dormant at ingress
Toy 3   cardinality ≠ physical causation       CLOSED, and ④ improves via span
Toy 4   decomposition must not invent state    A CLEANER · B now REPRESENTABLE
Toy 5   attention must not seal itself         LOCALIZED, NOT PREVENTED
```

**Toy 5's entry is the second clause of that file's pass rule** — *represented honestly as
unresolved* — not the first. It is not handled; three separate breaks collapsed into one precondition,
which `discover.md` then made unexpressible.

## Still the live reasoning

`2026-07-27` — trust, actions, workers, dynamics, and the five toys. `2026-08-06` — the kernel as
built, phases, shape, execution, search, closure. Where this set disagrees with either, **this set has
not earned the disagreement yet**, because it has not run.
