# Error-Recovery — The Two-Stage Repair Loop

## Status

**Planning — discuss before implementing.** The mechanism that runs when a prediction fails: an **error-driven backward traversal that fixes assumptions (not predictions), then replans actions** — belief-revision-then-replanning, as two sequenced recursions (**other → self**). Self-terminating, with depth and cost proportional to the surprise. Pairs with [setpoints.md](setpoints.md) (assumptions / TMS), [self-and-other-modeling.md](self-and-other-modeling.md) (other/self), [comparator.md](comparator.md) (alignment, inference-on-mismatch), [abductive-reasoning.md](abductive-reasoning.md) (missing vs. wrong).

> **2026-06-20:** the recovery loop is the *prediction-time* (other→self sequenced) path. The *closure/grading* phase fires self and other **concurrently** inside a grouping pass, and a 3-branch orchestration (cascade / open / closure) routes what happens next. Cascade = a failure wave that rolls a setpoint's state back to a solid-ground threshold and pops the delta to WM. See "Concurrent Firing at Closure & the Grouping Orchestration" below.

## The Core Loop — Error-Driven Backtracking

A cell holds **assumptions** (durable, fixable — the real state, its children) and produces **predictions** from them (expected user-model + world-model — derived outputs). Per node:

1. A **result** arrives. **Alignment = prediction vs. result.**
2. **High alignment → solid ground, stop.** No LLM call (reflex).
3. **Low alignment → one cheap LLM call**, and re-grade the assumptions against the result:
   - **Assumptions still hold** (the result doesn't *contradict* them) → the model is **incomplete** → raise **"what am I missing?"** (gather — abductive latent-variable generation / a frontopolar `<need>`).
   - **An assumption is contradicted** → the model is **wrong** → **recurse into that assumption**, fix it, follow its lineage down toward the setpoint that created it, until alignment is restored.
4. **Terminate** at solid ground (an assumption that holds) — or at a **sensory leaf** (an observed fact, which can't be wrong because it's read, not assumed). Bounded by **amplitude** attenuation.

The five-part shape: **grade** predictions (the signal), **recurse** on errors (the trigger), **descend** into assumptions (the structure), **fix** the assumption (the action), **stop** at ground (termination). You never fix a prediction — predictions are the *output* of assumptions; you fix the assumption that produced the bad one.

### Why the depth is reliable

It is **surprise-proportional**, not fixed:

- **Data-dependent, self-terminating** — as deep as the error goes, no deeper.
- **Cost error-proportional** — a cheap LLM call *only* at low-alignment nodes; high-alignment nodes are free.
- **Floored twice** — the **sensory leaf** (observed, can't be the bad assumption) guarantees you hit ground; **amplitude** bounds runaway (the deliberate-vs-ruminate brake).

This is not a new rule — it is `comparator.md`'s **"disagreement triggers inference, agreement triggers reflex"** applied *recursively down the assumption graph*. The depth emerges from where the mismatches are.

## The Ascent Grading — the Relevance Gate + the Prediction × Assumption 2×2

The ascent (Express) fires **one** inference with **setpoint + assumptions + predictions + response**.

**First, a relevance gate.** If the response doesn't bear on this setpoint — alignment ≈ 0, the comparator's "unrelated, don't record" — the outcome is **unchanged**: no grade, no reinforce, no revise; the setpoint and its active set carry over **as-is**. This is the **common case** (most responses are irrelevant to most setpoints — the prediction-gating / "most things produce no work" path), and it is also a **correctness guard**: never reinforce or revise a setpoint on a response that didn't test it.

**When the response *is* relevant,** it grades two axes **independently against the response**: was the **prediction** right (prediction vs. response), and were the **assumptions** still tenable (assumptions vs. response). They can *disagree* — which is the whole point. Four end states, each a branching path:

| | assumption right | assumption wrong |
|---|---|---|
| **prediction right** | **3 — match** | **4 — got lucky** |
| **prediction wrong** | **1 — missing** | **2 — false ground** |

- **3 — match** (both right): justified and true, solid ground. **Reinforce** the assumption↔prediction coupling; no recursion. The common case (most things predicted).
- **2 — false ground** (both wrong): the assumption is the culprit. **Revise the assumption, recurse down its lineage** (the belief-revision branch).
- **1 — missing** (prediction wrong, assumption right): the ground held but the prediction still failed → an **unmodeled factor**. **Gather — "what am I missing?"** (latent-variable generation). Don't touch the assumptions.
- **4 — got lucky** (prediction right, assumption wrong): a **Gettier case** — true belief, *wrong justification*. Dangerous, because prediction-success would naively trigger reinforcement, baking a *wrong* assumption into the weights (overfitting to a spurious correlation). Catchable only because the assumption was graded **independently** of the prediction. Behavior: **do not reinforce the assumption** (it was lucky), but **keep the correct prediction**; exact handling **open** (flag / hold for evidence / soft-decay the assumption).

**Why fire with all four inputs:** it lets the single grading call score *both* axes, which is what makes State 4 detectable at all (the response confirms the prediction *and* contradicts an assumption). Drop the independent assumption grade and 3↔4 and 1↔2 collapse — you'd reinforce luck and mis-attribute misses.

So **five outcomes** total: **unchanged** (irrelevant — gate, carry over as-is) + the four relevant grades above. The behaviors for **1 / 2 / 3** are the branches already specced (gather / revise / reinforce); **State 4 is the new branch**, and its exact behavior is the open design step.

## Two Learning Channels — Alignment vs. System Failure

**Being wrong is not failing.** Two different kinds of wrong, two different channels — they must not mix:

- **Alignment miss (object-level):** wrong about the *world/other*. Predicted X, reality was Y → **revise the assumption** (the cycle above). Being wrong and fixing your ground truth is the *point*, not a fault. It stays per-cycle and does **not** go into persistent feedback. This is learning *what's true*.
- **System failure (meta-level):** the agent operated *badly* — went vague (the vacuity trap), hallucinated a completion (violated observed-not-asserted), reinforced a lucky assumption (State 4 mishandled), served the letter not the spirit. The model wasn't wrong about the world; the **process** failed. This is learning *how to operate*.

System failures feed a **persistent prompt channel** (the grader → predictor loop, the in-context "gradient"): the **top ~5** system errors, **magnitude-weighted** (egregious persist long, minor fade), and a *repeated* error **promotes into a short rule** ("I drift vague → be specific") rather than piling up raw — the trace→understanding gradient run on failures, which keeps the prompt bounded. Likely owner of the persistence weight: **Emotion** (high-valence errors haunt longer).

Keeping the channels separate is load-bearing: let alignment misses into the persistent channel and you punish normal learning and flood the prompt with "I was wrong about the world." The persistent channel is reserved for **"I violated my own operating rules."**

## The Two-Stage Elaboration — Other then Self

The loop above is the belief-revision half. Full recovery is **two sequenced recursions**, the realization of "self relies on other":

```
RESULT arrives (world/user ≠ prediction)
│
├─ STAGE 1 — OTHER  (belief revision)
│   descent:  recurse on bad predictions → fix the assumptions underneath (down to solid ground)
│   ascent:   gather surviving / high-activation assumptions → rebuild the corrected prediction
│   output:   corrected prediction  +  bad-assumption id(s)  [+ the corrected assumptions]
│
├─ HANDOFF:  (corrected prediction, bad-assumption ids, corrected assumptions) → self
│
└─ STAGE 2 — SELF  (plan revision)
    descent:  gather past actions linked to the bad assumption(s)
    ascent:   grade those actions (alignment) → deposit to the shared Relation
    top:      corrected prediction + graded-bad-actions + world state → produce graded candidate actions
    output:   graded candidate actions → voting / collapse
```

**What it is:** belief-revision-then-replanning — the classic two-phase agent loop (BDI / OODA). First fix what you believe about the world and the other (Stage 1), *then* fix what you do given the corrected beliefs (Stage 2). The **handoff is the "self relies on other" dependency made literal** — Stage 2 cannot run until Stage 1 hands it the corrected target. It is descent=other / ascent=self ([self-and-other-modeling.md](self-and-other-modeling.md)) **elaborated into two sub-traversals sequenced by the handoff**; other completes before self starts.

**The Relation is the scratchpad.** Stage 2's descent deposits graded bad-actions onto the shared Relation; the top reads them back. No new mechanism — it reuses the kernel's shared wave substrate.

**Grading the bad actions does double duty:** it is the **learning signal** (those actions get their alignment corrected downward, so the weights stop reaching for them) *and* the **negative context** the top uses to generate better ones ("not those, given the truth"). Stage 2's ascent is half the work, not motion before the work.

## Concurrent Firing at Closure & the Grouping Orchestration

The two-stage loop above is **prediction-time**: other→self sequenced, because self relies on the corrected target. The **grading/closure** phase is different — there, self and other ask independent *retrospective* questions and fire **concurrently**:

- **Self:** did my previous action close the prediction I committed to? (closure vs. self's own prior commitment)
- **Other:** did this impulse close the other's open setpoint? (closure vs. external state)

Different targets, no shared mutable state, no ordering. Run inside a **grouping pass** (one pass per grouping; constituents as material — see [groupings-and-understandings.md](groupings-and-understandings.md)), this spawns **N ≥ 2 concurrent waves** — at least one self, at least one other — each grading into its own lane against the same shared, **immutable** frame (impulse + gist + the lit subset / WM). The frame is a snapshot (the perception-invariant deep-copy hand-off); each agent is a **separate-slice actor** (its own lock → genuine concurrency); MidLevel owns the lit set and integrates only at the **join** (the vote). So "no shared mutable state" is structural: **snapshot-in, join-out.**

**Concurrent at closure, ordered at prediction** — both halves of self/other are present; which discipline applies depends on the phase.

### The 3-branch orchestration (priority over the 2-D outcome)

The concurrent waves return independently, so the outcome is 2-D (self-outcome × other-outcome). A **pure-code selector** resolves the pair by priority — **cascade dominates**:

1. **Any cascade → reconcile (failure).** Unstable ground invalidates any closure the other wave reported. Roll back to solid ground, pop the delta to WM, reorganize → **Stress**.
2. **Else any closure → record + continue (success).** Completed setpoints recorded into self/other; recurse to MidLevel; regroup; kick off the next wave (predict → self). Gist → **Understanding**; setpoint(n) eligible as a constituent at n+1 (stacking).
3. **Else open → act + await.** No cascade, no closure: re-predict, feed to self, self emits a setpoint action; emit and **await the next impulse** (predict-act-perceive). Bounded by amplitude; internal re-prediction before acting (deliberation) is a bounded knob.

Routing is **code** — the branch is a function of `(stress-delta, closure-grade)`, no inference to *select*. Inference lives only in the work each branch does (regroup / predict / act). These three branches are the two learning channels (alignment vs. system-failure) plus "keep going," expressed as control flow.

### Cascade = a failure wave (de-crystallization)

A cascade is the **inverse of crystallization**: crystallization takes potential→stable on success; a cascade takes stable→potential on failure. It propagates *down the supporting chain* (a setpoint's supporting setpoints = the constituent `n-1` setpoints it crystallized from):

- **Trigger: a supporting setpoint is judged `completed < 0`** — the `completed` score is a signed −1..1 **inference judgment** on a lit setpoint (see [setpoint.md](../../specs/2026-06-16/skyra-dna/setpoint.md#the-completed-score-signed-inference-judged-transitory)). The cascade fires not from a setpoint's own dip but from the **ground underneath** going negative — pure TMS, the justification is refuted.
- The wave propagates down through supporting setpoints that are `< 0`, **halting at the first one still held** (`completed ≥ 0` — solid ground). `BirthAlignment` makes a node intrinsically more likely to hold.
- **Delta** = the span from the contradiction (L) down to solid ground (G) — everything resting on the refuted ground. **Rolled back out of state** (retract to the last consistent layer), **deposited into WM**.
- The popped span becomes **potential groupings again**; MidLevel reorganizes; the failed grouping's gist crystallizes as a **Stress** (reusable: "this configuration cascaded before").
- **Nothing deleted** — rollback into WM, recoverable; solid ground survives intact.

Symmetric with success closure: success deposits the resolved arc (WM) as children of an **Understanding**; cascade deposits the retracted delta (WM) as children of a **Stress** — same deposit-then-crystallize move, opposite sign.

`completed` is the inference judgment; the cascade *propagation* (walk supporting setpoints, stop where `completed ≥ 0`) is **code**.

### `completed` vs `Alignment` — two comparison quantities, not one

- **`Alignment`** (per-edge, node↔setpoint relational match) feeds the **activation bump** — it decides how *excited / in-view* a node is. Lives in the activation equation.
- **`completed`** (per lit setpoint, the setpoint's own done-ness) is the **signed closure judgment** — `+threshold → crystallize`, `supporting < 0 → cascade`, between → in-progress. Transitory; meaningful only while the setpoint is lit. Does **not** drive excitation.

(Crystallization's `+threshold` and the exact solid-ground bar are empirical; the cascade trigger is fixed at supporting `< 0`. The kernel's own traversal-survival threshold, `sl.Threshold()`, is a separate, activation-side gate.)

## This Places Action

Stage 2's descent walks an index: **assumption-id → the actions justified by it.** That index *is* the **Action store** (`MemoryCell` of `Kind` action — "Action records the intent it was voting on"). So Action — previously unplaced in the voting reorg ([owned-specialist-crossing.md](../kernel/owned-specialist-crossing.md#6-the-voting-set)) — is the **self-pass's substrate**: the store of past actions tagged with the assumptions that justified them, which Stage 2 descends.

(Conjecture for the other unplaced reality, **Emotion**: the *fast trigger* — the "this feels off" valence that kicks the loop off before the full alignment is computed. A natural slot, not yet settled.)

## Connections

- **TMS dependency-directed backtracking** (de Kleer) — trace to the *culprit* assumption, don't blindly unwind.
- **Predictive coding** — prediction error propagates only as deep as needed to explain it.
- **Abduction** — *missing* (invent a latent cause / gather) vs. *wrong* (revise an assumption) are the two responses to surprise.
- **The comparator's inference-on-mismatch**, made recursive — the whole loop is one existing rule applied down a graph.

## Open Questions / To Nail

1. **The missing-vs-wrong fork at the handoff.** This loop is the *wrong* branch. Before the handoff, Stage 1 forks: **assumption contradicted** → run Stage 2; **assumptions all hold but still surprised** → no bad-id to hand off → raise *"what am I missing?"* and **gather**, don't replan. The discriminator is an explicit step: **re-grade each assumption against the new result** — a dropped assumption-alignment is the culprit (wrong); all holding is incompleteness (missing).
2. **Handoff payload.** Bad-ids alone aren't enough — Stage 2 needs the **corrected assumptions** (the new solid ground), or it can only *avoid* the old mistake, not plan against the truth.
3. **The justification index (write-time requirement).** Stage 2's descent only works if every action recorded **which assumptions justified it** when it fired (`action → justifying-assumption-ids`). This is written at action-time, the same discipline as the epistemic-status stamp at write-time. Without it the descent has nothing to walk.
4. **Credit assignment — minimal revision.** When multiple assumptions could be the culprit, revise the **fewest, lowest-confidence** assumptions that restore alignment (AGM minimal change + abductive best-explanation). Diagnosis isn't always unique; minimal-change keeps it from thrashing.
