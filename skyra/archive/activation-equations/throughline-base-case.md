# Throughline Base Case

## The Idea

Two base cases for descent, not one:

1. **Signal exhausted** — the wave ran out of energy. Return with what you gathered.
2. **Throughline found** — the accumulated thoughts have coherence. Return before you lose it.

Signal exhaustion is physical. Throughline is cognitive. The being stops descending not just because it ran out of gas, but because it arrived.

## Mechanism

ThoughtDensity already exists — average pairwise cosine similarity across thought vectors on the Relation. Currently only checked on ascent (CompetenceExpress) to decide whether to compress. The throughline base case checks it on descent too. If density crosses a threshold during accumulation, the thoughts are converging. Continuing deeper would dilute the coherence.

```
// during descent, after Observe:
if ThoughtDensity(r, beingID) > throughlineThreshold {
    // stop descending — throughline found
    // skip Relationships traversal, go straight to ascent
}
```

## Implications

- A being that finds its answer early ascends early and is cheaper.
- A being that doesn't find coherence keeps descending until signal dies — more expensive, and the result is probably scattered.
- The throughline base case is both a quality signal and an economics signal.
- Threshold tuning matters — too low and the being stops at shallow agreement, too high and it never triggers.

## Status

Not implemented. Idea stage.
