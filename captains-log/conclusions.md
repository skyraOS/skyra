# Conclusions — Comparator Experiments (.1001–.1003)

Three experiments, one cheap judge (DeepSeek `deepseek-v4-pro`), hand-built presents. Run June 14 2026.

## The arc

- **.1001 — Goal closure** (`phase-03`). Alignment against an observable `satisfied-when` is a real, graded goal-closure signal: 0.5 (greeting sent) → 1.0 (reply arrived), where cosine sat flat at 0.47. It does not self-satisfy — the greeting alone never reads as done, because the setpoint's exterior hook ("…and has replied back") is unmet. The Run-1 ±1 polarization was a prompt gap (poles defined, middle not), fixed in Run 2.
- **.1002 — Noise tolerance** (`phase-04`). 5 signal cells in 55 noise + 4 entity-decoys: found all 5, zero false positives, zero decoys grabbed. Relevance is **semantic, not keyword** — a filter keyed on the entity would have surfaced all five "michael" cells; the judge kept only the one bearing on closure.
- **.1003 — Scale** (`phase-05`). Filled the present to ~200k tokens (15k cells) on DeepSeek's 1M window. The **judgment is scale-invariant** (headline correct 26k→209k, decoys resisted at every scale). **Complete recall degrades past ~100k** — clean at 100k (5/5), broke at 200k (lost-in-the-middle misses + a 5-digit index transcription error: found the silence cell #11107, cited #1107).

## What it adds up to

The comparator is **two things in one cheap call**: a goal-closure **verdict** (the headline alignment) and a goal-relative **context manager** (the context-management half). Both held under pressure:

- **Verdict** — graded, observed-not-asserted, scale-invariant to 200k.
- **Curation** — finds signal in noise, zero false positives, semantic (not keyword) relevance. This is the goal-relative, continuous, non-destructive, traceable alternative to panic compaction: it scores relevance *against the active setpoint*, runs every turn instead of firing once at the cliff, *suppresses rather than deletes* (so suppressed context can resurface when the goal shifts), and emits its reasoning.

This is the first empirical support for the architecture's core thesis — `descent-ascent.md`: the topology is the compression, the cost curve inverts — and for the economic thesis that cheap tight calls over a topology can replace one stuffed expensive call. The load-bearing sub-claims (cheap, reliable, goal-relative, scale-robust judgment) held.

## The number that matters

**Reliable-recall ceiling ≈ ~100k per call (10% of a 1M window).** 200k is fine for a verdict, unreliable for exhaustive recall. Keep each call ≤~100k and compose — which is itself the empirical case *for* descent/ascent decomposition over one giant window.

## The honest ledger — what is NOT shown

- **Single samples.** Decimals are noisy; only the trends are load-bearing.
- **Hand-built presents.** No traversal, activation, frames, or write-time domain stamps were exercised. We validated the **judgment layer**, not the **machinery** that assembles a present from a real topology.
- **One-shot, not continuous.** A coherence call scored relevance *once*. Running it as a constant context manager in a loop — cascading suppression, the self-bounding ≤~100k constraint — is untested, and that is where the hard dynamics live.
- **"Beats compaction" is an argument, not a measurement.**

## Next

**First priority (unchanged):** stop hand-building the present — port a minimal being (membrane + Goal) onto the void kernel so the frame comes from real *traversal*, and re-run the comparator on it. That tests the machinery, not just the judge.

**Hint at what follows:** these results make a specific, falsifiable claim worth measuring — that a *constant coherence call* beats *straight compaction* as a context manager. Turning that from argument into measurement is the next experiment (`.1004`, specced in `phase-06`), to run once the manager exists in a loop. Queued, not first priority.
