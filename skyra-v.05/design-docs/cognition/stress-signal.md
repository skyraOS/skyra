# Stress Signal

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

## Cross-Binding Disagreement

Stress can emerge from inner communication across binding realities. It rises when different binding realities report incompatible interpretations of the same active pattern.

Example:

```
Goal: keep pursuing
Motivation: reason is uncertain
Episodic: weak match / no precedent
Action: prior attempts failed
Prediction: expected next state did not arrive
```

No single signal here is "stress" by itself. Stress is the combined pressure of the disagreement.

## Slice Ownership

Stress lives on the Slice because it is universal per-being activation state.

The Brain owns specialized binding realities. Each binding reality can contribute stress components. The Slice aggregates the components into a current stress signal and stores stress traces across traversals.

The Relation does not own stress. A Relation can carry current-wave pressure such as energy cost or buffered arrivals, but durable stress history belongs to the Slice.

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
