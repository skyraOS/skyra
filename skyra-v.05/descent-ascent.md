# Descent Is Additive, Ascent Is Subtractive

## Two Different Physics

The descent doesn't budget. The descent follows weights until they exhaust. Activation determines depth, nothing else. The descent is unconstrained — accumulate everything the weights say matters.

The ascent is subtractive. After the base case hits, the Relation starts returning upward. Now the token budget constraint kicks in. Each Express level looks at what's on the Relation, looks at the desired token count, and compresses toward it. Each level strips away what doesn't need to propagate up. The memory region used 50 memories to synthesize one insight — the 50 stay down there, the insight goes up.

The weights decide what's relevant (descent). The token budget decides what survives (ascent). The observer's context window shapes the return, not the observation.

This is why compaction never gets hit as a panic step. The context window never overflows because the ascent is lossy by design. Each Express level is a compression boundary — not because someone wrote a compaction strategy, but because that's what Express does. The loss is the cognition.

Every other system lets context accumulate in a flat buffer and panics when it's full. This never accumulates in a flat buffer. The topology is the compression. The structure is the strategy.

## The Cost Curve Inverts

Every other system gets worse as conversations get longer. They accumulate context in a flat buffer, hit the window, and panic-compact — sliding window, summarization, arbitrary truncation. Each one is a guess about what to throw away. The longer the conversation, the worse the guesses get.

This system gets better. Every traversal reinforces what mattered and starves what didn't. By the 50th exchange the topology has opinion — dense neighborhoods where the being has deep experience, sparse regions where it doesn't. The compression isn't a strategy applied to a buffer. It's the shape of the graph itself. The being doesn't decide what to keep. What survived the weight dynamics IS what matters.

The cold start is expensive. Early traversals haven't built up weight differentiation — everything looks roughly equal, providers fire where they don't need to, the topology hasn't learned what matters yet.

But the cost curve inverts. Flat-buffer systems start cheap and get expensive (or degrade). This starts expensive and gets cheap — because the topology learns where to spend inference and where to skip.

The distributed collapse is the mechanism. Every node on the ascent is a compression boundary. Each one only passes up what's load-bearing from its perspective. By the time the result reaches the top frame it's been compressed through every layer it passed through — not once by a summarizer, but at every level by a node that knows its own neighborhood. The loss is distributed across the topology instead of concentrated in one compaction step.
