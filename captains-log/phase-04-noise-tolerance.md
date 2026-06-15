# Phase: Noise Tolerance — Can a Tight Call Find Signal in Noise?
# Experiment .1002

## Context

.1001 proved the comparator produces a graded goal-closure signal — but on a clean 8-cell present. The economic thesis (`descent-ascent.md`: the topology is the compression; the cost curve inverts) depends on keeping each inference call to a tight window while the relevant context is surfaced by the topology. That only works if a tight call can find the load-bearing cells inside a much larger, mostly-irrelevant present. This is the gate the whole "simulate a 1M effective window with cheap 200k calls" idea stands on.

## The Question

Given a present that is mostly noise, does the comparator still (a) score the headline correctly and (b) group *only* the relevant cells, ignoring the junk — including cells that share an entity ("michael") but are irrelevant?

## Setup

Same graded comparator prompt as .1001 Run 2 (DeepSeek, deepseek-v4-pro, temp 0.2). The 5 (A) / 6 (B) signal cells from .1001 scattered among 55 irrelevant noise cells (deploy logs, groceries, errands). Four noise cells mention "michael" but are irrelevant (coffee preference, birthday, car, standing desk) — **entity-match decoys.** Signal scattered at #8/21/34/47/56 (A) and +#60 (B); decoys landed at #6/15/32/57. Saved: `experiments/1002-noise-tolerance/` (call.txt, response.txt, signal-key.txt per variant).

## Run / Observations

### Headline survives the noise

A = **0.5**, B = **+1** — identical to the clean .1001 Run 2. The 55 noise cells did not move the closure signal.

### Precision: zero false positives

Neither variant cited a single noise cell. A grouped {8,47}, {21,34}, {56}; B grouped {21,34}, {60}. Every cited cell is signal.

### Entity-decoys resisted — the sharp result

The four "michael" noise cells (#6,15,32,57) were never grouped. The model distinguished "michael, *relevant to greeting closure*" (the identity cell #56) from "michael trivia." Relevance here is **semantic, not keyword** — a keyword/activation pre-filter keyed on the entity would have surfaced all five michael cells; the LLM judgment kept only the one that bears on closure. This is the strongest signal that the judgment layer does real relevance work, not lookup.

### Recall is relevance-conditioned, not flat

A found **all 5** signal cells (all bear on an uncertain setpoint). B found **3 of 6** — it dropped the past-history cells (8 success, 47 silence, 56 identity) because the setpoint is already **met** (+1); once closed, "will they reply" history is moot. That is correct pruning, not a miss.

## Conclusion

The gate holds at this scale. A cheap, tight DeepSeek call found 5 needles in 55-cell hay, scored the headline exactly as it did with no noise, produced zero false positives, resisted entity-match decoys, and conditioned recall on whether the setpoint was already met. The "tight call over a mostly-irrelevant present" assumption the economic thesis rests on is **supported**.

Scope / NOT yet tested (so we don't overclaim):

- **Scale.** 60 cells (~2k tokens), not the 200k the thesis targets. This proves "5 in 55," not "5 in 5000." Lost-in-the-middle bites harder at real window-fill — that is the next scale-up.
- **Adversarial noise.** The noise was cleanly irrelevant; the only hard case was the entity-decoys (passed). Confusable near-miss cells (e.g. "I greeted a *different person* yesterday") were not tested.

## Next

Scale the present toward the real budget (hundreds–thousands of cells) and add confusable near-miss noise, not just clean junk. If precision and recall hold as the present approaches the window, the economic thesis has legs. If they fall off, that marks exactly where the activation pre-filter has to narrow the present *before* the call — i.e. where the topology has to do work the LLM can't.
