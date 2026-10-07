# pursue: Event-driven, continuous goal pursuit

pursue is a proposed durable procedure that keeps a goal active over time without requiring an agent to think continuously.

### Structure

It has two semantic ingresses:

1. Intent/control: establish, revise, pause, resume, or cancel a goal.
2. Observations: receive events from any number of sources—user messages, timers, sensors, or other system actors.

Intent describes the desired change. Observed inputs establish the baseline and current state, making the goal evaluable.

```text
Intent + initial observations → evaluable goal
                                      ↑
N event sources → observed state → progress / gap / unknown
                                      ↓
                              bounded next step
                                      ↓
                              persist and wait
```

### Goal evaluation

A pursuit carries a way to derive its target, incorporate observations, and evaluate progress.

For “lose 15 pounds in a year”:

```text
baseline  = initial observed weight
target    = baseline − 15 pounds
progress  = baseline − current observed weight trend
remaining = current observed weight trend − target
```

Missing evidence produces unknown progress, not assumed progress. Goals need not be numeric, but they must have observable completion conditions.

### Execution and novelty

Each relevant event triggers a bounded episode:

- Update the observed state.
- Evaluate progress and whether existing assumptions still hold.
- Run a remembered procedure when applicable.
- Invoke the model when novelty invalidates those assumptions.
- Take an authorized action, ask a question, or do nothing.
- Persist state and return to waiting.

The pursuit persists; individual episodes finish. Events provide evidence, not additional authority.

The foundation is a durable relationship between intent and observed reality: maintain a commitment, measure its progress, and adapt when new evidence changes what should happen next.
