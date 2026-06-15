# Captain's Log

Observations and conclusions about activation dynamics across experiment runs.

Each document is an activation phase. Each run gets logged with what was set up, what was observed, and what it means.

## Format

Each log entry:

```
# Phase: [name]

## Setup
What was configured. Which signals were active. What the topology looked like going in.

## Run
What happened. Raw observations from the debug view, logs, and TUI.

## Observations

### [observation name]

Expected:
  What should have happened based on the theory.

Observed:
  What actually happened.

Evidence:
  Exact numbers from debug logs. Activation scores, signal values, traversal counts.

Conclusion:
  What this means. What to change. What to test next.

## Phase Conclusion
What this phase proved or disproved. Architectural changes to make before next phase.
```

## Rules
- Numbers, not feelings. Every observation needs activation scores from the logs.
- One observation per signal anomaly. Don't bundle.
- If a conclusion changes the architecture, link to the commit or the spec change.
- If a conclusion needs another experiment, name the next phase.
