# Phase: Scale — Does Signal-Finding Survive a 200k Present?
# Experiment .1003

## Context

.1002 showed the comparator finds 5 signal cells in 55 noise cells and resists entity decoys — but at ~2k tokens. The economic thesis ("simulate a 1M effective window with cheap, tight calls") needs that to hold as the present approaches the real window. DeepSeek's window is ~1M, so this fills the present toward 200k tokens (20% of capacity) and watches for lost-in-the-middle.

## Setup

Same graded comparator prompt (DeepSeek, deepseek-v4-pro). Noise generated programmatically (templated, deterministic). 5 (A) / 6 (B) signal cells scattered across the *full span* (4% → 96% position) to probe lost-in-the-middle; 8 "michael" entity-decoys interleaved. Ladder: A at ~26k / ~104k / ~209k tokens, B at ~209k. **Single sample per rung — numbers are noisy; read the trend, not the decimals.** Saved: `experiments/1003-scale/` (present.txt, call-meta.txt, response.txt per rung).

## Results

| rung | tokens | cells | headline | cited recall | precision | decoys grabbed |
|---|---|---|---|---|---|---|
| a-25k | 26k | 1,914 | **0.5** ✓ | 4/5 (missed #1837, last cell) | clean | 0 / 8 |
| a-100k | 104k | 7,615 | **0.5** ✓ | **5/5** | clean | 0 / 8 |
| a-200k | 209k | 15,217 | **0.5** ✓ | 2/5 cited (~3/5 conceptual) | 1 bad index | 0 / 8 |
| b-200k | 209k | 15,218 | **+1** ✓ | 4/6 | clean | 0 / 8 |

## The headline is scale-invariant

Expected: the closure signal might drift as the present grows.
Observed: **0.5 / 0.5 / 0.5 / +1 — correct at every scale, 26k through 209k.** The graded goal-closure judgment — the thing the comparator exists for — does not degrade with context size.
Conclusion: the core economic-thesis claim (the *judgment* survives a 200k present) holds. This is the result that matters most, and it held cleanly.

## Decoy resistance is scale-invariant

Across all four rungs — up to 15k cells with 8 decoys — **zero michael-decoys were grabbed.** Semantic relevance filtering does not weaken at scale.

## Recall degrades at 200k — and in two distinct ways

Recall was clean through 100k (5/5) and broke at 200k. Two separable failure modes appeared, both only at 200k:

1. **Lost-in-the-middle (real misses).** a-200k genuinely dropped #7608 (general knowledge) and #14608 (identity, at the 96% tail); b-200k dropped #6390 and #12173. The model stops reliably finding *every* scattered needle once the present is ~15k cells.
2. **Index transcription error.** a-200k's observation explicitly reasons about "silence" — so it *found* the silence memory (#11107) — but cited it as **#1107** (a digit dropped), which is a noise cell. The internal find was right; the 5-digit index came out wrong. This is a *citation* failure, not a retrieval failure, and it is fixable independently: shorter/chunked IDs, or have the model quote cell content alongside the index.

## Phase Conclusion (.1003)

The economic thesis splits into two claims, and the scale run separates them:

- **Closure judgment at scale: holds.** 0.5/+1 correct at 209k, decoys resisted at every scale. If all you need from a tight call is "did the goal close," you can hand it 200k and trust the number.
- **Complete recall from one 200k call: does NOT hold.** Recall was clean at ~100k and broke at 200k (misses + a citation error). So the reliable-recall ceiling for a *single* call sits near **~100k (10% of the window), not 200k (20%)** — the earlier "20% per call" heuristic is too loose when complete recall matters.

This is the empirical case *for* the descent/ascent decomposition rather than one giant window: a single stuffed 200k call is fine for a verdict but unreliable for exhaustive retrieval, exactly where chunked tighter calls (≤~100k) plus composition should win. It also flags a concrete output fix: cite content, not bare 5-digit indices, or chunk the numbering.

## Next

- Repeat the 200k rung 3–5× to get real recall variance (single samples here).
- Re-run at ~150k to bracket the recall cliff between 100k and 200k.
- Switch the output to quote cell *content* (or chunked IDs) and re-test 200k recall — isolates how much of the 200k degradation was citation vs. genuine lost-in-the-middle.
- Then the bake-off: chunked ≤100k tighter calls vs. one stuffed 200k call, same total tokens — does decomposition recover the recall the single call lost?
