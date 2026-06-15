# Co-Activation Linkage

## The Problem

When a cell fires an observation or expression, the output isn't born from one memory. It's born from the combination of everything active in the context window at that moment. If the new node only links to its trigger or its host, the topology loses the birth context — the full set of memories that contributed to that thought.

## The Mechanism

The Relation already carries everything needed:

- **Thoughts** — every cell whose content was picked up during descent. The co-active memories.
- **Fields** — every cell's binding activations. Which cells were active, at what level, in which topological layers.

When an observation or expression produces a new node, that node gets edges to everything in Thoughts and Fields at that moment. Not just the triggering memory. Everything that was co-present in the context window.

The context window is already segmented by the memories themselves — each memory has its own slot in Fields with its binding activations. The provenance is structural. You can trace back exactly which memories seeded which observation at what activation level.

## Weight Dynamics Do the Rest

At birth, all edges from the new node to its context are equal. Over time:

- Memories that actually contributed to the observation keep co-occurring with it on future traversals. Their edges strengthen through reinforcement.
- Memories that were coincidentally in the window but not actually contributing stop co-occurring. Their edges fade through decay.

The topology figures out which connections were real and which were noise. No curation step needed. The weight dynamics are the curation.

## Seeding Process

Deposit already creates a new cell and wires it into the topology. Co-activation linkage extends this:

1. Create the new cell (observation, expression, or final deposit)
2. Link it to the host (LastSeen) — where it physically lives
3. Link it to every cell in Thoughts — everything picked up during descent
4. Register it in Fields with the binding activations from the current traversal
5. For each cell in Thoughts, create a reciprocal edge weighted by that cell's activation at the time of creation

This applies to three types of output:

- **Observations** — intermediate products from Observe. "I noticed X." Born from the context window, linked to everything in it.
- **Expressions** — intermediate products from Express. "I think Y follows." Born from the post-descent context, linked to everything in it.
- **Final deposits** — when signal is spent. The terminal memory, linked to its full birth context.

## What This Means

The topology doesn't just remember conclusions. It remembers how it got there. Every observation and expression is a node linked to its birth context. The chain of reasoning is structural — not a log, not a trace, but live nodes with weighted edges to the memories that produced them.

Future traversals through the same region encounter not just the raw memories but the observations and expressions they generated. Those intermediate nodes are lenses — the same way promoted nodes sit above their constituent memories, observations and expressions sit above their birth context and shape what future traversals see.

The topology develops layers of interpretation over its raw material. Memories at the bottom. Observations above them. Expressions above those. Each layer linked to the layer below through co-activation edges that strengthen or decay based on continued relevance.

This is how understanding forms. Not by storing a conclusion. By the topology building weighted structure between the memories that co-produced an insight. The understanding IS the pattern of co-activation. When those memories light up together again, the observation they produced lights up too — because the edges are there.

## Compression Research

The co-activation linkage gives full provenance — what was co-present, at what activation, when every node was born. Most compression techniques compress blind. This compresses with full context. Research existing techniques (context window compression, memory consolidation, graph-based summarization) and map each against the architecture. Some may fit, some won't. Incorporate what matches, discard what doesn't.
