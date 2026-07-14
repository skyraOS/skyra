# Captain's Log

Observations and conclusions about activation dynamics across experiment runs.

Two kinds of entry:

- **Activation phases** (`phase-NN-*.md`) — the core log. Each is a Skyra experiment run:
  what was set up, what was observed, what it means. Numbers-driven (see Format + Rules).
  Experiment numbering runs in series: `.1xxx` was the comparator work (phases 03–06); `.2xxx`
  is the `skyra-dna-v3` prediction-diff learning work, opening at `phase-07` / `.2001`.
- **Field observations** (`observations/NN-*.md`) — qualitative observations of *existing*
  systems (e.g. an agentic harness), kept when they bear directly on the architecture.
  Not experiments, no debug numbers, no next phase — they deliberately break the
  numbers-not-feelings rule and say so in a banner. Filed as standing evidence, cited when
  the relevant design is questioned.

## Format (activation phases)

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
