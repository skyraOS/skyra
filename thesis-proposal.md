# Thesis Proposal — Working Scaffold

**Degree:** MS, Computer Science
**Status:** pre-writing. This is the argument skeleton, not the proposal prose.
**Subject system:** `skyra-dna-v3` (`skyra-v.1/src/skyra/skyra-dna-v3/`)
**Date opened:** 2026-07-21

---

## The question

> **Does an agent whose learning signal is a computed prediction–observation diff
> reduce its planning cost across repeated exposure to a task class — and is the
> durable, observation-grounded graph the mechanism of that reduction?**

This is the narrow, defensible form. It is deliberately not the research program
(below, under Future Work). It is a single mechanistic claim with an ablation that
can kill it.

### Why this phrasing, specifically

Two clauses, and the second does the work. Clause 1 is a curve: cost falls with
exposure. Clause 2 asks whether the accumulated graph *causes* that fall, and it is
answered by removing the graph and watching the curve flatten.

Both comparisons are **internal** — the system against its own earlier state, and the
system against itself with one component disabled. No external agent, no second
architecture, no confounds imported from someone else's substrate. Same world, same
task, same retrieval and rendering, because it is literally the same binary.

### Why there is no LLM-judge baseline

An earlier draft phrased this as *substituting* a computed diff **for** an LLM-judged
grade, which requires a judge arm. That framing does not survive contact with the
architecture:

- **The judge has nowhere to stand.** v3 expectations are triples with exact
  addresses. Handing a judge a predicted triple and an observed triple asks it to
  perform string equality — a task it cannot win and was never needed for. The
  comparison is rigged by construction, and a reader will see that immediately.
- **It is not one independent variable.** The judge was necessary in v2 *because*
  predictions were free text. Representation and grading method are a package deal:
  free-text-plus-judge and triple-plus-diff differ in two things at once, so "only
  the source of the number changes" is simply false.
- **It cannot falsify anything.** The result is knowable in advance — equal accuracy
  at best, plus latency and variance. An experiment whose outcome is known before it
  runs is not evidence.

The signal-quality question the judge was meant to answer is answered better, and
deterministically, by the noise sweep below. See Related Work for where the judge
argument *does* belong.

---

## Operational definitions

The question is only falsifiable once these are fixed. Settle them before writing.

- **The grade.** A computed diff between a committed, falsifiable expectation and a
  subsequent observation of world state. No model in the grading path. Accumulates on
  edges as `P(world state | expression)` with an EMA prior.
- **Planning cost.** Primary: **passes-to-closure** (planning cycles until the target
  is observed-met). Secondary: total inference calls, and wall-clock. See Metrics for
  why passes is primary.
- **Repeated exposure.** N episodes of one task class, fresh world each episode,
  persistent expression graph. Only what was *learned* carries across.
- **Memory ablation.** The durable graph / accumulated edges disabled. Prediction:
  planning cost stops falling — isolating the graph as the mechanism.
- **Grade corruption (the noise sweep).** `Truth` is flipped with probability *r*
  immediately before the edge update — *after* settlement, so nothing in the loop
  changes except the quality of the learning signal. *r* is the dial that replaces
  the judge arm, and it must be seeded so every run replays exactly.

---

## Experimental design — one ablation, one dose-response

### 1. The ablation (clause 2) — the load-bearing experiment

|                   | **Memory on**             | **Memory off**             |
|-------------------|---------------------------|----------------------------|
| **Computed diff** | cost falls — *the thesis* | cost flat — *the ablation* |

Two cells. If cost still falls with the graph disabled, the graph is not the
mechanism, the LLM alone is doing the work, and the thesis is dead. This is the only
experiment that can kill it, so it gets the seeds and the scrutiny.

### 2. The noise sweep (signal quality) — dose-response, no LLM

Flip `Truth` with probability *r* before the edge update. Sweep, memory on,
everything else fixed:

```text
r = 0      the thesis curve
r = 0.1    degraded, still falling
r = 0.25   substantially flattened
r = 0.5    uninformative on a binary — should land on the memory-off curve
```

Predicted result: monotone degradation, with `r = 0.5` collapsing onto the ablation.

That bracket is the whole point. **Any grader whatsoever — an LLM judge included —
sits at some unmeasured *r* on this axis.** The sweep does not merely replace the
judge comparison, it *subsumes* it: one binary result against one judge at one
unknown error rate is strictly less information than the curve that contains it. And
it is deterministic, reproducible, and free.

---

## Metrics — two axes, one measured and one structural

1. **Capability axis — passes-to-closure.** Does the curve fall, and does it stop
   falling under ablation? *Measured*, with variance bands.
2. **Efficiency axis — grading cost.** *Analytic, not empirical.* The diff adds
   **zero** calls to the trusted loop; any model-in-the-loop grader adds one per
   grade. This needs no experiment — it is a property of the construction, and it is
   un-arguable precisely because no benchmark is involved.

Splitting the axes this way removes the rigged-comparison objection entirely: there
is no comparison to rig. The capability claim is measured against the system's own
earlier state; the efficiency claim is a structural fact about where the network
calls are.

**A flat curve is falsification, not a null result** — and it is the outcome this
design is built to detect early rather than in month five. If the curve falls but the
noise sweep fails to degrade it, that is *also* informative and *also* bad: something
other than the graded signal is driving the improvement, and the ablation says what.

### Reporting rules (so the numbers survive scrutiny)

- **Neutral unit + concrete unit for each metric.**
  - Cost: grading calls per closure (neutral, primary) *and* dollars (concrete).
  - Latency: grading calls in the critical loop (neutral) *and* wall-clock (concrete).
- **Passes-to-closure is primary, not total calls.** Passes measures whether the
  being got better. Total calls conflates that with how the grade is produced. Lead
  with passes; report calls and wall-clock separately as the efficiency result.
- **State grading cost structurally, not as a benchmark.** API latency is noisy
  (network, provider load, retries) and a wall-clock number invites an argument about
  the harness. The un-arguable statement is structural: **zero** calls added to the
  trusted loop. Report that; report wall-clock only as its concrete instantiation,
  with variance bars.
- **Variance, or it is an anecdote.** `.2001` is n=1 in seed. Every cell needs a band
  (≥8–10 seeds) or "falls" and "falls less" are not distinguishable.

### The architectural argument the metrics unlock

A grading path that requires a network call is in a different complexity class than one
that is arithmetic: it cannot run in a tight real-time loop, cannot be replayed
deterministically, and cannot be audited offline. This is the operational
justification for LAW 3 (exact-first, arithmetic in the trusted path). Cost and latency
are what turn "my thing is faster" into "my thing is a different kind of system."

---

## Build debt the question creates

**None.** This is the main practical consequence of dropping the judge arm: every
experiment above runs on the current substrate.

- The ablation needs a flag disabling durable edge accumulation.
- The noise sweep needs one seeded flip on `Truth` immediately before the edge update
  — `being.go`, at the `edge.Observe(truth, b.tick)` call site. A few lines.
- Everything else is instrumentation that already exists: `Metrics.Resolved`,
  `Metrics.Correct`, `Metrics.GradeSum`, and passes-to-closure.

The previous draft carried an LLM-judge grading head for dnav3 as "~1–2 days and the
single largest build item." It is now zero days, and that runway goes to seeds and a
second task shape — which is where it buys actual defensibility.

---

## Experiment order — fail fast on the one that matters

1. **The ablation, on a hard-enough task — FIRST.** Memory on vs memory off. This is
   the load-bearing experiment and the only one that can kill the thesis. If cost does
   not fall with memory on, or falls just as well with it off, learn that now — while
   there is runway to make the world harder — not in month five.
2. **Variance bands** across seeds for both cells. `.2001` is n=1 in seed; below
   ≥8–10 seeds, "falls" and "flat" are not distinguishable claims.
3. **The noise sweep**, once the ablation has separated. It is only interesting if
   there is a curve to degrade.
4. **A second task shape**, to show the curve is not task-specific.

---

## Related work — position, do not ignore

Several results have named ancestors. Raising them first is the difference between
"reinvented 1986 unknowingly" and "knows the field." The comparison is favorable.

- **SOAR chunking** (Laird, Rosenbloom, Newell, 1986). Passes-to-closure `5→1` is
  structurally the same speedup-learning curve. **The difference:** chunking compiles
  resolutions of impasses reachable through hand-authored operators; it does not
  acquire structure never encoded. dnav3 recovers a recipe *never in the prompt* from
  the world's failure reports — an LLM proposes candidate structure, counted edges
  validate it against observation. That is an answer to the knowledge-acquisition
  bottleneck that was not available in 1986.
- **ACT-R activation** (Anderson; Collins & Loftus, 1975). The activation field —
  EMA decay, bumps from the current impulse, bounded spreading over edges — is close to
  base-level plus spreading activation. Position as convergent, then cite.
- **Model-based RL** (Sutton & Barto). "The edge is a conditional probability, the EMA
  the prior" is a tabular transition model with a learning rate. Say so.
- **Active inference / predictive processing** (Friston). observation → state change →
  action → observation with prediction error as the signal. Phase 27's own invariant
  already gestures here.
- **LLM-as-judge** (Zheng et al., 2023, and the self-evaluation literature). This is
  where the judge belongs — as *position*, not as an experimental arm. The argument is
  architectural and needs no run: a judged grade bottoms out in an opinion with no
  independent source, and the documented biases (self-preference, position, verbosity)
  are exactly what a component with an incentive to look good will exploit. dnav2 is
  the lineage evidence — that architecture was built, run, and hit precisely this
  ceiling, which is why v3 exists. Cite the history; do not re-run it.

The novel core has no clean ancestor: **an ungameable grade produced by keeping the
model out of its own grading path** (LAWs 4 and 6, and
`TestPromptDoesNotLeakTheScoringRule`). SOAR/ACT-R never had a component with an
incentive to look good; dnav3 does, and builds structure specifically to neutralize it.

---

## What would falsify the thesis

- **Planning cost does not fall** → the compounding claim dies outright.
- **Cost falls but survives the memory ablation** → the graph is not the mechanism;
  the LLM alone is doing the work. The sharpest failure, and the reason the ablation
  runs first.
- **The noise sweep does not degrade the curve** → the improvement is not coming from
  the graded signal, so the stated mechanism is wrong even though the curve is real.
- **The curve falls on only one task shape** → a speedup specific to that task, not a
  property of the mechanism.

---

## Scope boundary — what this thesis is NOT

Not defended here, held as context or future work: the globe/world design, the QM
phases, world nesting, the compression research (`replacing-triples-with-compression`),
the kernel/wave, deep-bridge retrieval. The world is **instrumentation** — an honest
oracle whose only job is to say no in a way the agent did not author. It is designed
from the claim, frozen, then run. It is not the subject.

---

## Future work — the research program this is a first brick in

> Under what conditions can lived experience — accumulated as counted, observation-
> grounded relationships — substitute for parameter updates as the mechanism by which
> an agent's capability grows?

This is the broad, PhD-shaped question. It requires holding the claim against fine-
tuning, RL, and continual-learning literature, and multi-year runway. Recorded here so
the narrow thesis knows what it is a first brick in — not attempted now.

---

## The one discipline this whole document depends on

Freeze the grading path and measure it. The project's method (overbuild → run →
let the number condemn it → delete) has revised the substrate five times in five
months and has no built-in exit. A thesis needs a frozen artifact. Freeze **one
component — the grade** — and let the rest of the program keep moving around it. The
work that makes this a thesis is running experiments against a frozen substrate and
reading the five load-bearing papers. It is less fun than building. That is the job.
