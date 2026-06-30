# Stress Signal

> **Reconciliation (2026-06-23).** Vocabulary drift — the reasoning holds. Read
> **"Coherence" → MidLevel**, **"binding realities" → the self/other agents**.
> Placement: stress lives on the being's slice but **not** on the `skyra.Slice`
> kernel interface (see "Slice Ownership"). The `stress = persistence × instability`
> model predates the `completed` (notDone) score and isn't yet reconciled with it —
> persistence overlaps "an open prediction kept being re-presented"; treat the exact
> relationship as open.

## The Insight

Stress is not an unresolved goal. A long-running goal can be calm if its dependencies are valid, progress is predictable, and the being knows how to continue.

Stress is unresolved pressure plus instability. It appears when the active goal persists while the substrate reports threat to resolution.

```
stress = persistence x instability
```

Persistence keeps the question alive. Instability makes it costly.

## What Stress Is Not

- Not raw duration. A goal can persist for many traversals without stress.
- Not failure alone. A failed action can be informative and low-stress if the system has alternatives.
- Not emotion. Emotion stores valence around patterns. Stress is the pressure signal that can cause memory to encode more strongly.
- Not a separate planner. Stress does not choose actions directly. It changes salience, compression pressure, and memory formation.

## Persistence

Persistence measures how much the substrate is continuing to carry a goal.

Sources:

- **Traversal persistence** — how many traversals the goal has remained active.
- **Episode persistence** — how many episode boundaries the goal survived.
- **Motivational carry-forward** — how many times Motivation chose to slide the goal into a new episode instead of closing or abandoning it.

Persistence is not bad. It is the weight of continued concern.

## Instability

Instability measures whether the active concern is becoming harder to resolve.

Sources:

- **Goal error trend** — progress delta is flat or worsening.
- **Dependency uncertainty** — Motivation dependencies have faded from Surroundings or failed verification.
- **Prediction error** — expected incoming signal did not arrive, or prediction delta is rising.
- **Action failure** — primitives fail, return unmatched, or repeatedly do not change the state.
- **Field disagreement** — binding realities disagree about the same active pattern.
- **Load pressure** — too many active cells or too much content remains uncompressed.
- **Energy pressure** — Relation energy is running low before resolution.
- **Concurrency pressure** — buffered Relations accumulate while inference is gated.
- **Episode churn** — episode boundaries keep triggering without closure.

Instability is the signal that persistence is no longer cheap.

## Coherence (ACC) as Stress Center

Coherence is the ACC — the general mismatch detector. But the ACC in biology doesn't just detect conflict — it allocates cognitive control resources. Coherence is not a detector that hands deltas to a separate stress aggregator. Coherence IS the stress center. It finds the mismatch, reads the magnitude, and routes energy based on the type of mismatch.

Three mismatch domains, each routing energy differently:

- **Content mismatch** (setpoint vs members) → energy to contradicting cells. Topology pressure. Split or absorb.
- **Prediction mismatch** (predicted vs actual) → energy to the failing goal. Goal dominance. Gather-before-commit.
- **Articulation mismatch** (intent vs expression) → energy to motor/action loop. Self-correction.

Stress on the slice is the readable state of what MidLevel is currently doing — not a separate aggregation step. MidLevel writes it; cells read it. The instability components in the formula below are MidLevel's outputs, not independent signals collected from everywhere. (Historical note: "Coherence" = MidLevel; "binding realities" = the self/other agents.)

The error signal IS energy. Coherence signals the mismatch magnitude, that magnitude becomes activation energy on the relevant pathway. A failing goal gets louder because the error signal is energy, not because a rule amplified it. Suppression of other goals is the shadow of amplification — finite energy, one goal running hot, everything else gets less. Resolution drops the activation, attention redistributes automatically. No explicit restore step.

### The Mechanism — Per-Goal Delta Pass

Coherence is a dumb method. It takes prediction errors and propagates the network. The mechanism is:

1. **Stimulus arrives** — lands in the attention map on the slice.
2. **Read active goals** — all goals currently tracked on the slice.
3. **Cheap per-goal comparison** — for each goal in isolation: does this stimulus close the gap or widen it? Goal setpoint vs stimulus. Directional delta: positive (closer), negative (further), zero (irrelevant). No inference. Activation math.
4. **Coherence aggregates** — reads the collection of per-goal deltas. Net direction, magnitude, which goals moved which way.
5. **Coherence routes energy** — based on the aggregate.

Step 3 is N cheap passes, one per active goal. Not one big evaluation. Each pass is the goal's content (setpoint) against what just arrived.

### Aggregate Patterns

The aggregation is where behavior comes from. One stimulus can help one goal and hurt another simultaneously. Coherence sees both.

- **All deltas positive** → low stress, proceed, reinforce. Everything got closer.
- **Mixed deltas** → some goals helped, some hurt. Conflict. Energy routes to the hurt goals. This tension IS the stress signal.
- **All deltas negative** → high stress. Everything got worse. Gather-before-commit. Frontopolar activates, questions generate.
- **One large negative, rest neutral** → single goal under threat. Goal dominance — that goal runs hot, others starve temporarily. Resolution drops it back.

**Status: directionally correct.** The formula and components below still hold as a shape — persistence × instability — but instability may collapse from a sum of independent components to Coherence's magnitude across its three domains.

## Cross-Binding Disagreement

Stress can emerge from inner communication across binding realities. It rises when different binding realities report incompatible interpretations of the same active pattern.

Example:

```
Goal: keep pursuing
Motivation: reason is uncertain
Episodic: weak match / no precedent
Action: prior attempts failed
Prediction: expected next state did not arrive
Coherence: mismatch magnitude high across multiple domains
```

No single signal here is "stress" by itself. Stress is the combined pressure of the disagreement.

## Slice Ownership

Stress lives on the being's slice because it is universal per-being activation state — a **global neuromodulator** (one value per being, not per-edge). It modulates the activation bump (the `stress_modulator` term), episode-segmentation sensitivity, and the cascade-stop reading.

> **Placement (2026-06-23):** stress is held on the **concrete being-slice**, *not* on
> the `skyra.Slice` **kernel interface** — the kernel is mechanism and must not know a
> cognition signal. Cells read it via a capability assertion, never the kernel contract.
> See [activation-and-traversal.md](../kernel/activation-and-traversal.md#stress-is-global-not-per-edge).

The cells contribute stress components (MidLevel and its agents, on cascade / prediction-error); the slice holds the aggregated current signal and stress traces across traversals.

The Relation does not own stress. A Relation can carry current-wave pressure such as energy cost or buffered arrivals, but durable stress history belongs to the slice.

## Stress Components

First-pass shape:

```go
type StressState struct {
    GoalPersistence        float64
    EpisodePersistence     float64
    MotivationalCarry      float64

    GoalErrorTrend         float64
    DependencyUncertainty  float64
    PredictionError        float64
    ArticulationFailure    float64
    ActionFailure          float64
    FieldDisagreement      float64
    LoadPressure           float64
    EnergyPressure         float64
    ConcurrencyPressure    float64
    EpisodeChurn           float64

    Current                float64
}
```

The exact weights are DNA-defined. Different beings can experience pressure differently.

## Formula

Start simple:

```
persistence =
  traversal_persistence
  + episode_persistence
  + motivational_carry

instability =
  goal_error_trend
  + dependency_uncertainty
  + prediction_error
  + articulation_failure
  + action_failure
  + field_disagreement
  + load_pressure
  + energy_pressure
  + concurrency_pressure
  + episode_churn

stress = persistence * instability
```

Clamp or normalize to `0.0–1.0` for use as a field.

## Timing

Stress should update on ascent, after binding realities have expressed their interpretation of the traversal.

Order:

1. Relation descends through activated cells.
2. Binding realities write fields and record observations.
3. Ascent reconstructs Current traversal tree.
4. Goal, Motivation, Prediction, Action, Episodic report deltas.
5. Slice aggregates stress from those deltas.
6. Stress modifies memory formation and future activation.

Stress depends on the completed traversal. It should not read a half-built Current tree.

## Effect on Memory Formation

Stress is an encoding modifier.

Low stress:

- light trace
- normal reinforcement
- weak or no compression

High stress with resolution:

- strong understanding
- strong action linkage
- reinforce dependencies that were re-grounded

High stress with failure:

- tension memory
- warning pattern
- action dampening

High stress from dependency loss:

- episodic marker around what faded
- Motivation dependency check gets stronger next time

High stress with later relief:

- store the full arc: pressure, attempted actions, resolution
- this is where useful coping/reflex behavior can form

## Relationship to Emotion

Stress is pressure. Emotion is valence.

Stress says:

```
this pattern is costly / unstable / unresolved
```

Emotion says:

```
approach, avoid, urgency, relief, threat, reward
```

Emotion can read stress as an input, but stress should not become emotion automatically. A stressful pattern can resolve positively and become confidence. A low-stress pattern can still have negative valence.

## Relationship to Motivation

Motivation tracks whether the reason for a goal still holds. Dependency uncertainty is one stress component.

Important distinction:

```
dependency absent from Surroundings = uncertainty
dependency checked and invalid = failure
```

Uncertainty raises epistemic goals. Failure can cause Goal abandonment. Both can contribute to stress, but they are different states.

## Relationship to Inference

Stress can raise the priority of inference, but should not be the only trigger.

Inference still fires at disagreement, contradiction, compression pressure, or unresolved recognition boundaries. Stress tells the runtime how urgent or salient the boundary is.

High stress can:

- preserve more context in the inference prompt
- increase compression priority
- cause stronger memory encoding after the result
- slow motor firing until dependencies are checked
- **bias primitive weighting toward perception over action** — over-communicate, ask more questions, gather more signal before committing. Stress shifts motor proposals from `engage` toward `perceive`. The being becomes cautious: wrong predictions mean the world-model is stale, so the response is to observe more before acting. This is the gather-before-commit reflex.

## Audit Record

Every decision record should be able to explain stress:

```
stress: 0.72
persistence:
  traversal: 0.30
  episode: 0.20
  motivation_carry: 0.10
instability:
  dependency_uncertainty: 0.40
  action_failure: 0.20
  prediction_error: 0.15
  field_disagreement: 0.25
effect:
  stored tension memory
  raised epistemic goal: look:missing_clue
```

Stress should be inspectable, not mystical.

## Test Room Expectations

Escape Room:

- stress spikes when a needed clue disappears
- stress drops when the being discovers a way to preserve or re-ground clues
- resolution creates a strong understanding/action memory

Maze Room:

- stress rises when orientation dependencies degrade
- stress drops when the being re-grounds location

Decay Room:

- stress tracks scarcity pressure
- useful prioritization should reduce stress over repeated runs

Social Room:

- stress rises when coordination dependencies conflict
- stress should not rise merely because multiple agents are active

Repeat Room:

- repeated success should reduce stress for the same pattern
- repeated unresolved failure should form tension memory or avoidance

## Stress as Episodic Segmentation Pressure

Stress modulates how aggressively Episodic segments. It lowers the threshold for "this is a new episode."

- **Low stress** → episodes run long, lazy boundaries, smooth continuity between traversals
- **Rising stress** → episodes fragment more often, each fragment encodes stronger (more gets packed per frame)
- **High stress sustained** → rapid fragmentation, dense encoding, poor cross-episode coherence. The being "can't think straight" — segmenting too fast to build a throughline.

This matches biology: under cortisol, hippocampal encoding goes into overdrive. More gets recorded, but the recordings are fragmented. Stressful memories are vivid but disjointed — lots of strong traces, poor continuity.

The subjective experience of "time speeding up" under stress maps directly to this: more episode boundaries per unit of real time means more chapters, shorter chapters. Each frame holds more because encoding is amplified. From the being's perspective, more happened.

### Mechanic

Prediction delta is still the mechanical segmentation trigger. Stress modifies the sensitivity:

```
effective_segmentation_threshold = base_threshold / (1 + stress)
```

At stress = 0, the threshold is normal. As stress rises, smaller prediction deltas are enough to trigger a new episode. The segmentation machinery doesn't change — just the bar for when it fires.

### Stress is not a binding reality

Stress doesn't need its own binding reality. It's a modifier that reads from binding reality outputs and writes to Episodic's segmentation threshold. It sits on the Slice as aggregated state — universal, not specialized.

## Open Questions

- Does stress decay continuously, per traversal, or only after resolution?
- Should stress be stored per goal, per episode, or globally on the Slice?
- How much does stress lower or raise activation thresholds?
- Can high stress suppress motor until Motivation dependencies are checked?
- What is the minimum useful component set for the first implementation?
