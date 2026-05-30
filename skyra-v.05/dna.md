# DNA — The Cell Reads Itself

## Competence — Why It's Load-Bearing

The v.1 runtime is one struct (Skyra), one interface (Reality), one traversal method (Realize). 311 lines. Everything — beings, memories, operators, skills — is the same type. The only differentiation is topology: where the cell sits, what its weights are, what relationships it has.

Competence is the mechanism that makes this work. Without it, a cell is whatever type it was created as — forever. You'd need separate structs: Memory, Thread, Think, Act, Specialist. Back to 5k lines. Back to v.05.

With Competence, the cell determines itself every frame from its own weights and the Relation's binding fields. A memory node that accumulates enough weight reads as a specialist next traversal. A dormant node that gets activated reads as a processor. No one relabeled it. It read itself and became what the field says it is.

### Dual Competence Read

The cell reads itself twice per traversal — once before Observe, once before Express. The first read determines intake behavior. Then descent happens: the cell absorbs content, the Relation changes, weights shift. The second read determines output behavior from the post-descent state. A cell that entered as a context node may have picked up enough during descent to express as an integrator on the way back up.

This is biologically correct. A cell that receives a signal undergoes intracellular cascading — phosphorylation, second messengers, transcription factor activation. By the time it's ready to output, its internal state is different from when the signal arrived. The output decision is made from the post-processing state, not the pre-processing state.

One field (`Type`), two reads. The second overwrites the first because the first is already consumed.

### Biology

Biology calls this competence — the cell's pre-existing internal state determines which response pathways are even available before processing the incoming signal. Same signal, different self-read, different outcome. This is not optional. It is the mechanism.

Three layers in biology:

1. **Chromatin landscape** — which genes are accessible vs locked. Long-term identity. A liver cell and a neuron receiving the same signal respond completely differently because different parts of their genome are available.
2. **Active transcription factor network** — what's currently running. Medium-term state.
3. **Bioelectric resting potential (Levin)** — voltage state acts as the upstream self-read, independent of which ion channel produces it. The cell reads its own voltage before deciding what to become.

In Realize, competence is the first step. The Reality reads its own weights and the binding fields on the Relation to determine what it currently is. This gates which Observe and Express branches are available this frame.

There is no declared type. There is no genome label that says "you are an emotional observer." The Reality determines itself from its weights at the moment the Relation arrives. The weights are the bioelectric field. The Relation is the signal. The competence step is the cell asking "what am I right now, given my accumulated experience and what's arriving?"

This means a Reality's type can change frame to frame. Dense emotional weights and an emotionally charged Relation — the Reality expresses as an emotional processor. Same Reality, next traversal, task-heavy Relation and reinforced task weights — it expresses as a task processor. One DNA. The field determines expression. The cell determines itself.

## Observe and Express Are Independently Composable

Every Reality has the same DNA — same interface, same maps. What differentiates a Reality isn't its type. It's the weight field determining which branches of the DNA express. But Observe and Express are not coupled. A Reality can be one thing on the way down and a different thing on the way back up.

Biology confirms this. Receptor expression and effector expression are separate gene regulatory networks composed orthogonally:

- **Same Observe, different Express:** Muscle cells, fat cells, and liver cells all bind insulin (same intake). Muscle stores glycogen, fat makes lipids, liver suppresses glucose production (three different outputs). M1 and M2 macrophages share identical phagocytic intake machinery but M1 outputs inflammatory cytokines while M2 outputs anti-inflammatory.
- **Different Observe, same Express:** Macrophages, neutrophils, and dendritic cells all perform phagocytosis (same output) but are triggered by different receptor families (different intake).

### Three fields per Reality

```
what the cell IS          → weight field (position in topology)
what it does on Observe   → intake profile (composable)
what it does on Express   → output profile (composable)
```

The DNA is two independent switch statements — one for Observe, one for Express. The weights determine which branch fires in each direction.

### Current observe modes

- **Pass-through** — transparent, don't touch the Relation. Dormant nodes.
- **Pick-up** — absorb content from the cell onto the Relation. Costs signal.
- **Propagate** — relay the signal deeper, shape it.

### Current express modes

- **Motor** — act, execute something (bash, API call)
- **Thought** — provider call, inference
- **Memory** — store/retrieve
- **Specialist** — promoted dense region doing synthesis

### System prompt composed from DNA fields

The system prompt isn't a handwritten string. It's assembled from the DNA composition — what this Reality is, what its Observe branch does, what its Express branch does. The Provider receives a prompt composed from all three — derived from the genome, not authored.

The prompt can evolve. If the weights shift and a different Observe branch starts firing, the system prompt changes because the DNA is expressing differently. The prompt is alive.

## One Type — Topology Is the Differentiation

One Reality struct. One Realize method. One Observe. One Express. The behavior emerges from what's in the Relationships map, what's in the Expressors map, and what the weights are. Think isn't a different type from Act — it's the same Reality with different relationships and different expressors. Memory isn't a different type from an operator. Same Reality, different contents, different weights.

The only real differentiation is at the leaf nodes — the base cases that actually execute. The provider calls an LLM. The terminal runs bash. The filesystem writes to disk. Those do different work. Everything above them is the same traversal: route by weight, observe, express, return.

One genome. Placement determines what it becomes.

## The Promotion Gradient

Three levels, all weight-driven. No labels. No type declarations. Just weight dynamics producing layers of abstraction over raw material.

### Memories → Understandings → Specialists

**Memories** — raw cells. The base material. They co-activate frequently enough and a conclusion deposits. The conclusion is itself a cell, linked to everything that was co-present when it formed (co-activation linkage).

**Understandings** — conclusion cells that keep proving useful. The same conclusion keeps forming from the same memories. The cell reinforces through EMA, promotes above its source memories, starts intercepting traffic. An understanding is not labeled. It's a cell whose weight says it belongs above the things that produced it. The understanding is what happens when the same conclusion between nodes keeps firing — the path between them climbs through reinforcement and the conclusion cell earns its position.

**Specialists** — what happens when multiple understandings in the same region keep co-activating. A set of promoted understandings that consistently fire together consolidate into one dense node. The specialist sits above the understandings the way understandings sit above memories.

Each step is the same mechanism — co-activation, reinforcement, promotion. Competence handles it: a cell at understanding-level activation reads as "processor" or "context." A cell at specialist-level activation reads as "specialist" or "integrator." The thresholds in `Competence()` map to this gradient.

### The underlying cells don't die

Promotion doesn't delete. The memories beneath an understanding are still there, still reachable if the signal is strong enough to penetrate past the promoted node. But the promoted node intercepts most traffic. The underlying memories stop getting direct traversal. Their weights decay through starvation — not deletion, just fading. They become the deep memory beneath the understanding.

But if something specific comes in — a signal that resonates with one of the underlying memories but not the promoted node — the signal passes through (transparent to the promoted node) and hits the original memory directly. The detail is still there. It's just quiet until something specific enough asks for it.

### Merge is structural cleanup

The step above understanding is merge — spectral coarsening. When the topology has stabilized and multiple understandings have fully consolidated into a specialist, the underlying nodes that no longer receive direct traffic can be merged. The graph gets smaller without losing structure. Merge is not the process of understanding. Merge is the structural cleanup after understanding has stabilized.

Spectral coarsening (Jin & Loukas, AISTATS 2020) provides the math: merge nodes while preserving the first k eigenvectors of the graph Laplacian. This guarantees that the major pathways of signal flow — the cognitive geography — survive the compression. Co-activation linkage provides the evidence for which nodes belong together. Weight dynamics determine when it's time.

## Promotion Is Activated by the Relation

The promotion gradient isn't triggered by a monitor or a threshold checker. It's triggered by the Relation itself during traversal.

A Relation enters a region. The region is dense — many hits, high convergent activation. The Relation records that density. On the return path, the density evidence is what triggers promotion. The Relation carried the proof that this region is load-bearing enough to deserve a higher expression.

The topology can only differentiate through use. No Relation flowing through a region means no promotion evidence, means no growth. The genome declares what could be. The Relation determines what becomes. One mechanism. The Relation is the morphogenetic signal.

## Promoted Nodes Are Lenses, Not Attractors

The promoted node sits ABOVE its memories in the topology. Promotion moved it there. The Relation descends through the promoted node first, then into the underlying memories.

```
descent enters → dense region → specialist → understandings → memories
```

The promoted node is a gravity well the Relation falls INTO. It's structurally higher in the topology because promotion put it there. The promoted node is a lens — the Relation descends through it, gets filtered and shaped by it, and then hits the layers underneath in the context of that lens. The underlying cells don't change. What changed is what sits above them.

This is the chromatin landscape from competence — the promoted node gates which cells are even accessible this traversal.

### The act of promotion is Levin's attractor

Levin's morphogenetic attractors work the same way. "Head" is an attractor state — a bioelectric prepattern distributed across many cells. When enough cells participate in the pattern, the attractor deepens and the ensemble converges on the morphogenetic goal. The act of convergence IS the promotion — the pattern becomes a structure that sits above its constituent cells and shapes what they express.

The promoted node doesn't know it was promoted. It's a Reality like everything else. Its Competence reads its weights and determines what it is. Promotion is just what the topology looks like after enough traversals reinforced the same region.

## Differentiation Under Activation

A memory node with low activation returns data. Same memory with high activation — enough context flooding in from the traversal — triggers its synthesis behavior. It becomes a pseudo-think layer. Not because someone promoted it. Because its DNA expressed differently under pressure. Same code, different activation, different behavior.

This isn't runaway. The recursion constrains it. A memory-turned-thought-layer produces new activations, but those have to activate through weights to go anywhere. Budget exhaustion and weight exhaustion are the natural brakes. The deeper you go, the more energy it costs. The recursion is the physics.

## Maximum Behavioral Diversity from Minimum Machinery

Same principle as DRY in software — not deduplication, but recognizing which dimensions are actually orthogonal and refusing to couple them. Biology arrived at the same answer because evolution is a compression engine. Redundant machinery costs energy. Cells that share intake behavior don't duplicate the receptor code — they express the same gene regulatory network and compose it with a different output network.
