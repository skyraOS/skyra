# Compression

## What We Already Have

The architecture arrived at the same mechanisms the literature describes independently. This is confirmation, not discovery.

**EMA decay is synaptic homeostasis.** Tononi & Cirelli's Synaptic Homeostasis Hypothesis (2003-2014): during waking, synapses strengthen through learning. During sleep, global downscaling renormalizes all synapses proportionally. Weakly-tagged connections fall below threshold and die. Strongly-tagged ones survive at reduced absolute strength but higher relative signal-to-noise. `Reinforce()` and `Decay()` are this mechanism. Proportional decay that preferentially eliminates low-signal connections while preserving ratio structure.

**Descent/ascent is the hippocampal-neocortical transfer.** McClelland, McNaughton & O'Reilly's Complementary Learning Systems (1995): two systems — hippocampus for fast, sparse, episodic storage; neocortex for slow, distributed, semantic integration. Hippocampal replay gradually transfers episodes into neocortical structure. The raw signal enters at the leaves (episodic). Each compression boundary on ascent moves it toward semantic integration. The topology is the neocortex.

**Every traversal is memory reconsolidation.** Nader et al. (2000): retrieved memories become labile — they can be modified during re-storage. Each retrieval is a chance to compress further or update with new context. The signal doesn't just read the graph. It rewrites weights. Every traversal is a reconsolidation event.

**Distributed compression beats centralized compression.** Hierarchical Cascade Framework (2025): distributing compression across stages, each operating at different granularity, outperforms monolithic compression while using 97% fewer FLOPs. The ascent does this by default — every node is a compression boundary.

## What We Can Use

### Spectral Coarsening

Jin & Loukas (AISTATS 2020). Merge nodes while preserving the first k eigenvectors of the graph Laplacian. This preserves community structure and diffusion properties — how signal flows through the graph doesn't change even though the graph got smaller.

Co-activation linkage defines which nodes are spectrally related. Cells that decayed to similar weights and connect to the same downstream cells are merge candidates. The graph gets smaller without losing structure. The being's topology consolidates over time — many memories of the same pattern become one dense node. This is how understanding forms structurally. Not by summarizing. By merging.

### Optimal Forgetting

Nagy, Torok & Orban (PLoS Computational Biology, 2020). The most important paper for this architecture.

When you compress episodes through a generative model (prior knowledge of structure), the distortions aren't errors. They're systematic: gist preservation, schema-consistent confabulation, detail loss. The same distortions human memory produces. Forgetting is optimal compression, not failure.

The topology IS the generative model. Compression through it will naturally produce these properties. The being won't remember exactly what was said. It will remember the shape of what happened — gist, not transcript. And the distortions will be consistent with the being's existing weight structure. It will remember what fits its experience, which is exactly what humans do.

This reframes the entire approach to memory. Don't try to preserve signal. Let the topology compress it lossily. The loss IS the cognition.

### Decision-Centric Rate-Distortion

(arxiv 2605.10870, 2025). Measures memory quality not by reconstruction fidelity but by decision quality — how well can the agent act given the compressed memory? You don't need to preserve the signal. You need to preserve what matters for downstream action.

Weight dynamics already encode "what matters" through reinforcement. High-weight structure is what the being has used successfully. Compression should preserve that, not high-fidelity reconstruction of everything. The being that forgets details but retains what helped it act is better than the being that remembers everything and can't find what matters.

### Local Preserve/Discard Decisions

LLMLingua-2 (Microsoft, ACL 2024) reframes compression as binary token classification — a local scorer decides preserve or discard using bidirectional context. Their implementation assumes a flat buffer (incompatible). But the principle maps directly: each node on the ascent makes a local preserve/discard decision about what passes up. The node knows its own neighborhood. It knows what's load-bearing from its perspective. It decides.

No central compactor. No summarizer. Each node compresses what came from below using its own local context. The ascent is a cascade of local preserve/discard decisions.

## What to Ignore

- **Sliding window** — positional eviction, no semantic signal. A hack.
- **Chunked summarization** — assumes flat buffer, destroys topology structure.
- **Retrieval-augmented compression** — flat chunking into summary vectors. Doesn't respect the graph.
- **Positional eviction** (StreamingLLM) — keeps attention sinks plus recent tokens, drops everything else. Purely mechanical.

All of these assume a flat buffer. The architecture doesn't have a flat buffer. It has a topology. Techniques built for buffers don't transfer.

## The Frame

The literature confirms: the architecture is doing the right thing. Distributed compression through a weighted topology produces better results than centralized compression of a flat buffer. Forgetting through structure produces systematic, useful distortion — gist, not garbage. Weight dynamics are synaptic homeostasis. The descent/ascent is the hippocampal-neocortical transfer.

Two concrete techniques to incorporate: spectral coarsening for graph consolidation (many similar nodes merge into one dense node), and optimal forgetting as the theoretical frame for why lossy compression through topology isn't a limitation — it's the mechanism.

## Two Traversal Modes

The activation equation answers "what's relevant" — it scores relationships by weight, recency, and context, and only follows what fires. This is what the being experiences during `Realize()`. It's how cognition works.

But merge needs a different question: "what's nearby." Structural neighborhood — which nodes are connected, regardless of current activation. A blind breadth-first walk that ignores weights and returns everything within N hops. The v.05 `MemoryGraph.Neighbors()` did exactly this — BFS from a node, collect everything found.

Both are needed. Activation-driven traversal for cognition. Structure-driven traversal for merge candidate identification. The activation equation tells the being what matters right now. The BFS tells the compression step what's structurally related regardless of current relevance. Spectral coarsening operates on the structural neighborhood, not the activation landscape.
