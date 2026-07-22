# Thesis Proposal — Working Scaffold

**Degree:** MS, Computer Science
**Status:** pre-writing. This is the argument skeleton, not the proposal prose.
**Subject system:** `skyra-dna-v3` (`skyra-v.1/src/skyra/skyra-dna-v3/`)
**Date opened:** 2026-07-21

---

## The question

> **Does substituting a computed prediction–observation diff for an LLM-judged
> grade reduce planning cost across repeated exposure to a task, and does the
> reduction survive a memory ablation?**

This is the narrow, defensible form. It is deliberately not the research program
(below, under Future Work). It is a single claim with enumerable failure modes.

### Why this phrasing, specifically

"Substituting X **for** Y" defines a controlled experiment. The baseline is not an
external agent with a hundred confounds — it is *this architecture with one part
swapped*. Same substrate, same world, same task, same retrieval and rendering. One
independent variable: how the grade is produced. That is the cleanest comparison the
system admits, and it is the reason the narrow question beats the broad one for an MS.

---

## Operational definitions

The question is only falsifiable once these are fixed. Settle them before writing.

- **The grade (treatment).** A computed diff between a committed, falsifiable
  expectation and a subsequent observation of world state. No model in the grading
  path. Accumulates on edges as `P(world state | expression)` with an EMA prior.
- **The grade (baseline).** An LLM judge scoring the same expectation against the
  same outcome, feeding the *same* edge/EMA machinery — only the source of the number
  changes. See Build Debt: this module does not exist yet.
- **Planning cost.** Primary: **passes-to-closure** (planning cycles until the target
  is observed-met). Architecture-neutral — both arms plan identically. Secondary:
  total inference calls, and wall-clock. See Metrics for why passes is primary.
- **Repeated exposure.** N episodes of one task class, fresh world each episode,
  persistent expression graph. Only what was *learned* carries across.
- **Memory ablation.** The durable graph / accumulated edges disabled. Prediction:
  planning cost stops falling — isolating the graph as the mechanism.

---

## Experimental design — a 2×2 factorial

|                   | **Memory on**                          | **Memory off**              |
|-------------------|----------------------------------------|-----------------------------|
| **Computed diff** | cost falls — *the thesis*              | cost flat — *ablation*      |
| **LLM judge**     | cost falls less / noisier / higher floor — *the comparison* | cost flat — *control* |

- **Main effect (clause 1):** computed-diff vs LLM-judge, memory on. Does the diff
  reduce planning cost relative to the judge?
- **Ablation (clause 2):** memory on vs off, computed-diff arm. Does the reduction
  survive removing the graph? If cost still falls with memory off, the graph is not
  the mechanism and the thesis is wrong.

Every cell has a predicted outcome that can be falsified. Run all four; the fourth
(judge, memory off) is a near-free control that completes the factorial.

---

## Metrics — two orthogonal axes

The result is not "the diff learns better." It is **Pareto dominance** across two axes.

1. **Capability axis — passes-to-closure.** Is the learning signal as good?
2. **Efficiency axis — cost + latency.** Given equal quality, what does producing the
   signal take?

The strong claim: *at least as good on capability, strictly better on cost and
latency.* This framing also defuses the rigged-comparison objection — the efficiency
win is on a **separate axis** from the capability claim, not a restatement of it.

**The equal-curve case is not a null result.** If the curves overlap, capability is
equal and the entire interest of the thesis moves to the efficiency axis — where the
computed diff dominates by construction (grading is arithmetic; the judge is a network
call per grade). So the thesis is interesting whether the curves separate *or* overlap:
separation is a capability win, overlap is an efficiency win, and the diff dominates in
both. State this explicitly — it removes the fear that "same curve" kills the result.

### Reporting rules (so the numbers survive scrutiny)

- **Neutral unit + concrete unit for each metric.**
  - Cost: grading calls per closure (neutral, primary) *and* dollars (concrete).
  - Latency: grading calls in the critical loop (neutral) *and* wall-clock (concrete).
- **Passes-to-closure is primary, not total calls.** The diff grades for free; the
  judge costs a call per grade. Measuring total calls hands the diff a trivial win and
  invites the rigged-comparison charge. Lead with passes (fair); report cost/latency
  as the separate efficiency result.
- **Isolate grading latency from confounds.** API latency is noisy (network, provider
  load, retries). The deterministic, un-arguable statement is structural: the diff adds
  **zero** calls to the trusted loop; the judge adds **one per grade**. Report the
  call-count delta as the result; report wall-clock as its concrete instantiation with
  variance bars.
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

- **An LLM-judge grading head for dnav3.** The counterfactual must exist and must be a
  fair swap: drops in where the computed diff sits, feeds the same edge/EMA machinery,
  everything downstream identical. dnav2's judge was a different architecture and does
  not count. Well-scoped (~1–2 days), but the single largest build item — scope it in
  the proposal so the comparison does not read as hand-wavy.

---

## Experiment order — fail fast on the one that matters

1. **Judge vs diff, memory on, on a hard-enough task — FIRST.** This is the load-
   bearing comparison. If the curves separate, there is a thesis. If they overlap, the
   task is too easy to distinguish the mechanisms (or the efficiency-only story is the
   thesis) — learn this now, while there is runway to make the world harder, not in
   month five.
2. Memory ablation (computed-diff arm).
3. Variance band across seeds for every cell.
4. Second task shape, to show the curve is not task-specific.

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

The novel core has no clean ancestor: **an ungameable grade produced by keeping the
model out of its own grading path** (LAWs 4 and 6, and
`TestPromptDoesNotLeakTheScoringRule`). SOAR/ACT-R never had a component with an
incentive to look good; dnav3 does, and builds structure specifically to neutralize it.

---

## What would falsify the thesis

- Planning cost does not fall in the computed-diff arm → the compounding claim dies.
- Cost falls but survives the memory ablation → the graph is not the mechanism; the
  LLM alone is doing the work.
- The judge arm matches the diff on passes *and* the efficiency gap is negligible →
  no result on either axis (unlikely — grading cost is structural — but stated for
  completeness).

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
