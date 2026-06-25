# Global — Active-Probe Grouping (Improvement Spec)

**Date:** 2026-06-25 · **Status:** improvement / **post-MVP**. The MVP grouping pass
(single inference call) in [global.md](global.md) ships **first**; this is the planned
upgrade, layered on once the loop runs and groupings can be measured.

> Captures the 2026-06-25 design conversation. Not on the step-2 critical path — build the
> passive MVP, close the loop, then bolt this on.

## The shift — the LLM drives the wave, instead of receiving it

The MVP grouping pass is **passive**: one inference call scores the impulse against the
active groupings over the lit set (one shot). This spec makes Global an **active driver**:
it reads the frame, **decides which nodes to probe**, spawns waves into them, grades what
returns, and groups *based on what it found*. The LLM steers its own attention instead of
passively receiving whatever the topology surfaced — **"perceive is active sensing" turned
inward** (the being's own attention model, applied to memory).

**Why it's better:** it directly attacks the **frame lock-in** seam
([../../../captains-log/observations/02-ancestor-failure-modes.md](../../../captains-log/observations/02-ancestor-failure-modes.md)
#3 — "relevant *dormant* structure is invisible because relevance is judged vs. the *active*
field"). An LLM that can **probe dormant nodes before grouping** sees structure the passive
surface never lit → groupings reflect exploration, not just what happened to be warm. The
LLM reorganizes itself rather than being a passive recipient.

## The mechanism — uses parts that already exist

1. **Passed competence (the "operators").** The probe wave **carries Global's
   grade-competence**; each probed node **brackets its own Observe/Express with it**
   (first-on-Observe, last-on-Express — the cell-execution contract in
   [../../../design-docs/kernel/activation-and-traversal.md](../../../design-docs/kernel/activation-and-traversal.md)).
   Kernel-agnostic — the kernel never reads the payload, just delivers the wave. So a node
   gets graded **without knowing what grading is**. This is "the methods are the operators of
   the traversal."

2. **Mini-BSP under kernel-void.** Probe waves can't return (no ascent — the kernel returns
   void). So they **write their grades to the cognitive-core slice**, and Global **joins by
   reading the slice after a barrier**. Shape:

   ```
   read frame  →  spawn concurrent probe waves  →  BARRIER  →  read grades  →  (iterate?)  →  group  →  express/commit
   ```

   A **mini-BSP inside Global's Observe superstep** — fractal with the outer two-pass loop
   (see [working-memory.md](../../../design-docs/cognition/working-memory.md)). The
   justification index (`intent → preds`) and the grades stay on the **cognitive-core slice**,
   not an agent's private slice.

3. **The flow.**
   - **Observe:** the LLM reads the frame (entities involved, the impulse, the active
     groupings, candidate setpoints).
   - LLM **identifies nodes to dig into** (entities or other memory nodes).
   - the system **spawns a wave onto those nodes' children**, carrying the grade-competence.
   - the **Express of those waves grades** → appended to the slice.
   - Global **reads the grades → groups** (informed by the probe) → commits.

## The termination controller — when deliberation stops and action fires

The active loop **must** be bounded, or it is exactly the **"attention dynamics became an
intractable control problem"** that killed OpenCog/ECAN (failure-modes #2). The controller is
the **drift-diffusion / evidence-accumulation-to-bound** model with an **urgency signal**
(Ratcliff; collapsing bounds — Cisek & Churchland): evidence accumulates toward a decision
bound; an urgency signal collapses the bound over time so a choice is forced even under
uncertainty.

### The ladder (driven by alignment dynamics + an urgency bound)

```
probe → grade (alignment, the existing −1..1 comparator metric)
  alignment improving        → keep probing (cheap tier)
  stalled HIGH               → pull up, commit (confident)            ← "make a call"
  stalled LOW                → escalate model tier, re-probe (cascade)
  stalled LOW at top tier    → resolve OUTWARD (act / ask / gather)    ← "outwardly resolve"
  + urgency bound (time → stress) collapses each pass → ALWAYS terminates
```

### Components

- **Alignment plateau (the core stop signal).** Stop when the **marginal improvement per
  probe** falls below threshold over a window of *n* probes — "no more information here." It
  rides the metric you **already produce** (alignment), so it is same-mechanism-different-job,
  no new machinery. The **stall *value* routes the branch**: stall **high** = converged →
  commit; stall **low** = stuck.
- **Tier cascade.** Escalate the model **only on stall-low**, never per-pass (bumping every
  iteration burns the cheap-call economics). Cheap-first, escalate on low confidence — the
  **same pattern the membrane already uses** (resonance-first, inference only when borderline)
  and the FrugalGPT model-cascade idea. Capped at the top tier → terminates.
- **Urgency bound (the hard termination guarantee).** Time-since-last-response is **not a
  separate timer** — it is a **stress input**, and stress already modulates thresholds/budget
  (the conflict-monitor compression). Stress **collapses the decision bound** → commit on
  weaker convergence under pressure, explore more when relaxed (the speed–accuracy tradeoff,
  done emergently rather than imperatively). Because the bound descends every pass, the loop
  **cannot run forever** even if alignment never converges — this is the guarantee against the
  ECAN failure.
- **Outward resolution.** Stuck-low **even at the top tier** = internal information exhausted
  → the correct move is to **act or ask** (gather external evidence). Not a fallback — the
  right cognitive response to genuine uncertainty. (This is the principled trigger for "should
  outwardly resolve.")
- **The stall threshold is itself stress-modulated** — under pressure, a shorter window and
  looser stall-tolerance fire earlier. This is where time (idea 1) and alignment (idea 3)
  mechanically meet.

## Guards (non-negotiable)

- **Budget** — amplitude / depth / probe-count cap, wired to the **explore/exploit knob**
  (grouping width: narrow grouping leaves more spillover to explore; wide exploits).
- **Termination** — the urgency bound guarantees halt; the alignment plateau gives the
  principled early stop; the tier cascade is finite.
- **Linearity** — each probe wave stays **within its node's own subgraph** (shared graph read
  by reference, never re-walked per probe) → keeps it O(N + E), not O(N²).

## Prior art

- **BSP / Pregel** — supersteps, barrier sync, **vote-to-halt** termination.
- **ReAct / agentic (iterative) retrieval** — the LLM as a search *policy* deciding what to
  probe next.
- **Active inference** — act to reduce your own uncertainty.
- **Drift-diffusion model** (Ratcliff) + **collapsing bounds / urgency** (Cisek, Churchland) —
  the stop controller.
- **Model cascades** (FrugalGPT) — cheap-first, escalate on low confidence.

## Relationship to the MVP

| | MVP ([global.md](global.md)) | This spec |
|---|---|---|
| Grouping pass | passive — one inference call (lit-set vs. groupings) | active — LLM-driven probe loop |
| Driver | the topology surfaces; the LLM receives | the LLM steers; topology is the substrate |
| Termination | n/a (single call) | the DDM ladder (plateau + cascade + urgency) |
| Build order | **first** — closes the loop, lets the comparator score a real frame | **after** — once the loop runs and groupings are measurable |
| Metric | alignment | **same** alignment — no new metric |

## Open

- Plateau window *n* and the marginal-gain threshold (find by experiment).
- The tier ladder — which models (cheap → mid → frontier; the inference package already has
  several backends).
- The shape of the **stress → bound-collapse** curve.
- Whether the probe loop and the outer two-pass BSP share **one** barrier mechanism.
- Per-cycle cost ceiling (the ~100k recall / cheap-call economics from the captain's log).

## Further out — the cell-type collapse (blackboard)

A direction this loop points at, **far down the line** (recorded so it isn't lost; **not** to
build now):

- **Collapse the structural cell types into one.** Today membrane / global / self / other are
  distinct types. The end state is **one polymorphic cell that reads the slice and assumes its
  role from state** — the same move `MemoryCell` already made (entity/setpoint/understanding
  are `Kind` *roles* of one type; [../../../design-docs/being/dna.md](../../../design-docs/being/dna.md),
  "one type, topology differentiates"), extended to the structural cells.
- **Indexes live on the slice** (not per cell); cells become uniform readers/interpreters over
  it. The named pattern is the **blackboard architecture** (Hearsay-II): a shared store +
  opportunistic agents that trigger on its state.
- **The reactive loop.** Global keeps **swapping the grouping until confident**, while self/
  other **continuously read it and re-fire on grouping change** (gated: they don't fire until a
  grouping exists). Global-writes / agents-read is **single-writer/multiple-reader** — the
  clean, lock-free corner (atomic publish the grouping; readers take the latest), *not* the full
  concurrency problem.
- **What makes it stable:** the **DDM stop controller above**. The blackboard's historical
  weakness (and ECAN's) is the *control problem* — agents thrashing, never settling. The
  alignment-plateau + urgency bound is exactly what halts the swap loop so the re-firing
  converges. This direction and this spec are one arc.
- **The discipline that keeps it clean:** the role-behaviors must be **operators / passed
  competence (data the uniform cell runs)**, NOT a hardcoded `switch(role)` — otherwise it
  reintroduces the behavior switch `dna.md` forbids ("`Kind` is NOT a behavior switch"). The
  cell stays a uniform *interpreter*; the slice holds the *program*. Building toward operators
  (already where passed-competence leads) keeps the eventual collapse cheap.
- **Timing — this is a *compression*, and the method is overbuild → then compress.** The
  distinct cells are the overbuild; "they're all just read-slice-assume-role" is the compress.
  Collapse only when that shared seam is **undeniable** in the built code — doing it now is
  guessing the abstraction before the duplication reveals it. (Open alt, also deferred: all
  indexes on every cell vs. all on the slice.)

## Done when

- Global runs an **active probe loop**: reads frame → identifies nodes → spawns
  competence-carrying probe waves → reads grades off the slice → groups.
- The loop **terminates** via the DDM ladder (alignment plateau + tier cascade + urgency
  bound).
- Groupings **measurably improve** over the MVP passive pass — the falsifiable claim worth
  measuring (a sibling to the `.1004` compaction-vs-coherence experiment).
