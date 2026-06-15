# Phase: Coherence vs Compaction — Does Goal-Relative Curation Beat Panic Summarization?
# Experiment .1004  (QUEUED — not first priority)

## Status

**Specced, not scheduled.** First priority remains porting the comparator onto the void kernel so the present comes from real traversal (see `conclusions.md` → Next). This is the *validation* experiment for the "a constant coherence call beats straight compaction" claim, to run once a continuous coherence manager exists in a loop. Listed here so the design is captured while the reasoning is fresh.

## Context

.1001–.1003 showed one coherence call is a cheap, goal-relative, scale-robust context manager. The stronger claim — that running it *every turn* beats the standard panic-compaction every agent uses — is so far an **argument**, four-part:

- goal-relative vs. goal-blind (scores against the active setpoint, not "keep the gist")
- continuous vs. panic (runs every turn; nothing accumulates to a cliff)
- suppress vs. delete (low-relevance cells stay in the topology, can resurface)
- traceable vs. opaque (emits its reasoning)

This experiment turns that argument into a number.

## The Question

Over a long, multi-turn task at a fixed context budget, does **coherence-managed** context produce better **downstream** output than **straight compaction**?

## Design — two arms, same task, same budget, same end-task model

- **Arm A — Compaction (baseline).** Standard buffer + summarize-on-overflow. What every agent does today.
- **Arm B — Coherence-managed.** Every turn, the coherence call scores each cell's relevance against the active setpoint; low-alignment cells suppressed (not deleted); the kept working set held ≤~100k (the reliable-recall ceiling from .1003).

Hold constant: the model that does the *end task*, the total context budget, and the task itself. The only variable is the context strategy.

## Task — the crux case

A multi-turn task where success hinges on a detail that is **introduced early, irrelevant at the time, and made load-bearing later.** This is precisely where the two strategies diverge:

- Compaction summarizes the early detail away (it looked irrelevant when the window filled).
- Coherence suppresses-not-deletes, so the detail is still in the topology to resurface when the later turn makes it matter.

Example shape: a long working session where a constraint mentioned in turn 2 ("the client can't deploy on Fridays") only becomes decisive at turn 40.

## Metric

- **Primary:** downstream task success — objective, not retrieval-against-a-key. Did the final output honor the early-but-later-critical detail?
- **Secondary:** tokens spent per arm; count of load-bearing details each arm lost.

## What would falsify the claim

If Arm B does **not** beat Arm A on downstream success at equal budget — or only matches it at higher cost — then "coherence beats compaction" is unsupported, and goal-relative curation is a wash versus summarization. A real possibility; the experiment must be able to return it.

## Honesty rails

- Same end-task model in both arms — isolate the *context strategy*, not the model.
- Design the early-irrelevant/later-critical detail so compaction is *genuinely tempted* to drop it (not a strawman summary).
- Objective/blind scoring of the final output.
- The resurfacing mechanism Arm B depends on must actually exist and be exercised — otherwise Arm B degrades to a worse compaction.

## Dependency chain

kernel port (comparator on a real traversed frame) → continuous coherence manager (the loop) → **this experiment.** Cannot run before the loop exists; the loop's dynamics (cascading suppression, the self-bounding ≤~100k constraint) are themselves untested and may surface their own findings first.
