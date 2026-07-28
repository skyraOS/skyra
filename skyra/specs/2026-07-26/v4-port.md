# v4 Port

**Date:** 2026-07-22
**Status:** WORKING DOC. Open. Nothing here is decided; §2 is measured from the code, §3
onward is proposal and argument.

**Reads on:** [dnav3-engine-and-the-ephemeral-worker](dnav3-engine-and-the-ephemeral-worker.md)
(the loop anatomy and the three non-mapping problems — not repeated here),
[traversal-implementation-plan](../../design-docs/kernel/traversal-implementation-plan.md)
(this is its Phase 3, against a substrate it did not anticipate),
[slice-based-dispatch](../2026-07-03/slice-based-dispatch.md),
`src/skyra/spacetime/traverse/traverse.go` (the kernel that exists and is unused).

---

## What the port is

> v3's engine is hand-written control flow over an implicit graph. v4 runs the same
> behavior as workers over an explicit one, on the `skyra.Space` / `skyra.Time` interfaces.

What must NOT change: the six laws, the grading arithmetic, and the two numbers. A port that
moves `PassesToClosure` or `GradeSum` has changed the experiment, not the implementation.

What is allowed to change: where iteration lives, what bounds it, and what can be inspected
mid-flight.

---

## 1. The finding that motivates this

**The code already contains three budget-bounded traversals, hand-rolled, in three
different shapes, none of which use the kernel written for exactly this.**

| # | Where | Bound | Selection rule |
|---|---|---|---|
| 1 | `propagateActivation` (`activation_field.go`) | `HopLimit: 2` | `crossing = activation × edge.P × edge.confidence()`, cut at `ε = 0.05` |
| 2 | `SelectMemoryDetailed` (`memory.go:188`) | `MemoryTokens: 2048` | greedy by `Density() = Activation / TokenCount`, mandatory first |
| 3 | `traverse.Assemble` (`traverse/traverse.go`) | caller budget | `child = (budget − cost) × weight`, descend iff `child ≥ child.Cost()` |

(1) and (2) run back-to-back on every planning pass and together **are frame assembly** —
walk the graph under a budget, keep what earns its space, render it. That is the exact job
`traverse.Assemble` was built for and has never been wired to.

(3) is Phase 0 of a plan that says its own Phase 3 is "collapse onto `skyra.Space`."

So the port is not introducing traversal into v3. **It is deleting two of three
implementations.**

---

## 2. The cycles the current code takes

Measured, not estimated. `→` marks a cycle that is already a traversal wearing a `for`.

### Per external impulse — `ObserveImpulse`, no model call, no tick advance

| Cycle | Bound | Iterates | Per iteration |
|---|---|---|---|
| write | `len(impulse)` | impulse triples | `appendObservation` → log append + fold + span index |
| decay | 1 | all five activation maps | `λ=0.5`, drop below `ε` |
| bump + retrieve | `len(written) × 3` | subject/predicate/object | bump entity; heat every span in that term's posting |
| → **propagate** | `HopLimit=2` | warm spans × expressions × expects | `activation × P × confidence`, cut at `ε` |
| grade | `len(written) × bets-at-key` | matching open bets | 6 gates, then EMA fold |
| target derive | groups | activated measured edges | grouping + confidence check |

### Per `Run` — the three nested loops

| Level | Bound | Body |
|---|---|---|
| **L1 reconcile** | 6 | `targetMet` (per target triple → `Obs.Resolve`) · `plan()` · `express()` |
| **L2 closure** | 6 | `Closed()` · `present()` · **1 model call** · `Parse` · `Validate` · `Stamp` · `Place` |
| **L3 ascent** | 8 | `Runnable()` per expression · `World.Do` per act · `Complete` · effects · `Settle()` |

### The sub-cycles inside those, which are the ones that actually cost

| Cycle | Site | Bound | Note |
|---|---|---|---|
| `Closed` | L2, every pass | `plans × acts × requires` | builds `projected` set, then `Obs.Resolve` per hole |
| → **`activateQuery`** | L2, inside `present` | 2 hops | *a second full propagation, per pass* |
| → **`activeMemoryNodes`** | L2, inside `present` | warm spans × observations + expressions | builds every candidate node, then sorts |
| → **`SelectMemoryDetailed`** | L2, inside `present` | `len(candidates) log n` + 2 passes | greedy knapsack to 2048 tokens |
| `Runnable` | L3, every round, every expression | `acts × requires` | `Obs.Resolve` per require |
| `Complete` | L3, per act | `len(expects) × len(Of)` | arming sweep |
| effects | L3, per act | `len(out.Effects)` | each is a full `observeInSpan` → write + activate + **grade** |
| `Settle` | L3, every round | `open bets + open interpretations` | `Matured` check; matured → `World.Perceive` → more `observeInSpan` |

### Multiplied out — worst case for one `Run`

```text
model calls        6 × 6                     = 36
present builds     6 × 6                     = 36   (each: 1 propagation + full candidate
                                                     build + knapsack over ~2048 tokens)
propagations       36 (planning) + 1 per observation written
World.Do           6 × 8 × acts-per-plan
World.Perceive     6 × 8 × matured-open-bets
grade sweeps       once per observation written, at every site that writes one
```

**36 model calls and 36 frame assemblies is the ceiling for a single target.** The
observed number is far lower — that gap *is* the retrieval curve — but the ceiling is set
by three unrelated constants that nothing reconciles.

### What this inventory shows

1. **`present` is the expensive thing, and it runs inside the innermost paid loop.** Each
   call re-propagates the field and re-runs the knapsack. Nothing caches across passes,
   even though passes within one `plan()` differ only by the holes.
2. **Grading is diffuse.** It runs at every observation-write site — intake, act effects,
   settle returns — which is correct (it is a fold on the write path, §5a of the engine
   doc) but means "when does grading happen" has no single answer in the current code.
3. **The three bounds are unrelated to each other and to cost.** 6 × 6 × 8 is not a budget;
   it is three guesses. A wave budget would make them one number with a meaning.
4. **Two propagations per pass.** `activateQuery` inside `present` re-does work
   `activateImpulse` already did at intake. Nobody decided this; it fell out of both paths
   needing a warm field.

---

## 3. Workers

Moved to its own doc — **[workers.md](workers.md)** — as a flat list of workers and
responsibilities, since it is the artifact that will keep growing.

Derived from §2: one worker per cycle with a distinct input, output, and failure mode.
**17 workers: 2–3 call a model, 2 cross the world boundary, 2 are write-path, the rest are
pure arithmetic.** (Two collapse to 15 once the excitation field owns its own decay and
propagation — see [interaction-primitives](interaction-primitives.md) §6.)

That ratio is the argument for the port. Most are deterministic, individually testable, and
individually inspectable — and today they are anonymous stretches of three functions.

Two are genuinely new: `sensory:interpret` (PerceivedActs → Beliefs) and the recording
half of `plan:articulate` (which Beliefs were in play, and how hot). Every other worker is
behavior that already exists in v3 and is being given a name and a boundary. Vocabulary in
[v4-data-model](v4-data-model.md).

---

## 4. What must survive

| Thing | Why |
|---|---|
| LAW 1 — observation the only writer | `Obs.Write` stays the single write path; workers return outcomes, they do not write world state |
| LAW 2 / LAW 4 — model out of the grading path | `grade:*` are write-path, not routed |
| LAW 5 — every bet gets settled | needs `Deadline` + `indeterminate`, which v3 does not have (§5b) |
| Proper time | `Slice.Count`, bumped once per wave, per agent. **Not** the being's motor activity. See [time](time.md). |
| `PassesToClosure` / `Closures` / `GradeSum` | pin a baseline run BEFORE porting; there is nothing to compare to afterwards |

---

## 4a. The cognitive agent does not sit on the substrate — REQUIRED WORK

The v4 tree asserts a relationship the code does not have.

```text
spacetime           → debug, itself
cognitive-agent     → NOTHING
distributed-agent   → spacetime, debug
user-agent          → spacetime, debug
```

**`cognitive-agent/skyra-dna-v3` imports zero spacetime.** Every other agent sits on the
substrate; cognition sits beside it. So `Space`, `Slice`, `Time`, and `Impulse` — the
contracts the rest of the runtime is built on — were never adopted by the thing the runtime
exists to run.

That is not an incidental gap. It is the cause of four things already recorded separately:

| symptom | recorded in |
|---|---|
| no `Slice`, so no proper time — the clock is a bare `int` on `Being` | [time](time.md) |
| no waves, so iteration is three nested `for` loops with hand-set bounds | §1–§2 above |
| no `Impulse`, so nothing rides the wave and dispatch is statement order | [dnav3-engine](dnav3-engine-and-the-ephemeral-worker.md) §5b |
| no kernel, so three budgeted traversals were hand-rolled instead of one | §1 above |

Each was diagnosed as its own problem. They are one problem: **the substrate contracts were
never reintroduced into the agent.**

### What this makes required rather than optional

The port is not "adopt a kernel because waves are tidier." It is **restoring a boundary the
rest of the tree already honours**:

```text
Space      the agent becomes traversable — the precondition for a device-place subtree
Slice      proper time, the traversal workspace, and the locking discipline
Time       the kernel invokes; nothing runs unless invoked
Impulse    routing state rides the wave instead of living in local variables
```

`distributed-agent` and `user-agent` already implement these. Cognition is the outlier.

### The order this implies

`Slice` first. It is the smallest contract, it carries proper time (which everything
measured depends on — [time](time.md) §3e), and it does not require the loops to be
restructured yet. `Space` and `Time` follow, and only then does the wave conversion make
sense.

**Open:** whether `Slice` is adopted whole or proper time gets a narrower home first — see
[time](time.md) §3a.

---

## 5. Open questions

> **Partially resolved 2026-07-24.** Q1 and Q2 below are closed by later docs; Q3–Q6 remain
> open. This doc is otherwise a valid *measurement of the v3 code* (§2) and stays as reference.

1. ~~**One transition table or three?**~~ **RESOLVED.** There is one wave with two phases,
   `Perturb` (descent) and `Observe` (ascent), not three nested loops with 6/6/8 bounds. See
   [perturb-observe-traversal](perturb-observe-traversal.md) and
   [kernel-alteration](kernel-alteration.md).
2. ~~**`traverse.Assemble` vs the knapsack — "the real technical fork."**~~ **RESOLVED — it
   was not a fork.** Affordability *is* local density; the kernel does zero math and each
   Space's `Observe` owns allocation-with-affordability.
   See [kernel-alteration §9](kernel-alteration.md) and
   [settlement-and-surprise §7b](settlement-and-surprise.md). "attenuation vs density-greedy"
   was a false dichotomy — magnitude-as-weight makes affordability price-per-token.
3. **How much of the substrate contract is adopted, and in what order** (§4a) — `Slice`
   alone, or `Space`/`Time`/`Impulse` together.
4. **`Realm` as leaf contract, or replaced by a Space subtree?** Bears directly on
   macAgentPort: a subtree can express "the Mac contains spaces containing affordances";
   `Realm` cannot.
5. **Is the port done before or after macAgentPort exists?** Porting against the sandbox
   keeps the experiment measurable throughout; porting against the Mac proves the mount
   but loses the control.
6. **Concurrency.** v2's board assumes `go`-spawned waves. v3 is single-threaded and its
   counters are unguarded. Adopt or defer?

---

## 6. Cheapest first move

Not a conversion. **Record a baseline.** `cmd/dnav3 -reps N` against the post-05:15 metric
definition, output kept. Without it, question 5 in the engine doc's §6 is unanswerable and
every later claim about "the port preserved behavior" is unfalsifiable — which is a poor
look for this system in particular.
