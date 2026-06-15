# Episodic Memory — Architecture Summary

## Frame Is the Primitive

A frame is one encoding pulse — the conjunction of what was co-present at that moment, after prediction filtered it. Frames are what get retrieved. Frames are what the write infrastructure produces. Every `WriteToSlice()` automatically captures the write into the current frame. Cells don't know this is happening. Temporal causality falls out for free.

## Episode Is Derived

An episode is a synthesized cap over a sequence of frames — the understanding that formed over those traces. Not a container. Frames point up to episodes; episodes reference frames. Many-to-many. A single frame can belong to N episodes simultaneously, one per segmentation dimension. Without the episode you get deja vu — a strong frame match with no story. With the episode you get "I remember this, and it was part of this, which was about X."

## Multi-Dimensional Segmentation

Episode boundaries aren't inherently temporal. The same cutoff mechanism (prediction error / instability crossing threshold) applies along any dimension — goal arcs, agent relationships, places, tools, emotional patterns. Each dimension segments independently. "What happened with michael" returns a sequence of agent-relationship episodes, each with its own beginning, arc, and resolution. The structure is the narrative.

## Central Store (Hippocampus)

One shared durable store. Frames are born on the slice (transient workspace), graduate to the central store when the traversal closes (durable archive). Episodic is a binding reality with its own logic and indexing, but lives on the slice as write infrastructure — not a node in the traversal graph.

## Distributed Indexes, Shared Store

Each binding reality has its own typed key into the frame layer. Goal indexes by goalID, Motivation by dependencyID, Emotion by valence pattern, Prospective by obligationID. Entity indexes (agent, user, place, device, tool) are shared. Every key resolves to frameIDs. The frameID is the boundary — above it, per-binding-reality doors. Below it, shared infrastructure. All doors open into the same house.

## Multi-Key Associative Retrieval

N independent address spaces query simultaneously. Activation is the product of matches across all active dimensions. More dimensions matching = stronger activation. Partial matches work. The structure is wide but the active surface at any moment is narrow.

## State Modulation (Stress/Emotion Traces)

Every episode carries a stress/emotion fingerprint stamped automatically during writes. Current system state re-ranks retrieved frames via their episode traces — congruent boosted, incongruent suppressed. Can independently activate frames without entity match (pure state-dependent recall).

## Retrieval Is Traversal

The episodic store is topology, not an API. Each binding reality spawns a wave into its own index. The kernel handles scoring and traversal. Results return through observe (Perception Invariant). The kernel doesn't know it's traversing memory.

## Competence Travels on the Wave

Binding realities are types — each has its own specialized logic (Goal tracks setpoints, Prediction runs compare/predict/signal, Motivation checks dependencies). But the cells in the hippocampus are generic. They don't know who they're serving.

The binding reality doesn't traverse into the store itself. It launches a wave with its competence attached as a payload on the impulse. The cells in the store receive the payload and execute it against their own content. Same cell, same frame, same content — evaluated differently depending on which binding reality's payload is on the wave.

```
Goal launches wave with payload: "compare my setpoint against what you hold"
  -> generic cell receives impulse
  -> executes payload against its content
  -> scores, returns result

Motivation launches wave with payload: "check if this dependency still holds"
  -> same generic cell, same content
  -> different payload, different evaluation
  -> different result
```

The specialization travels on the wave, not in the cell. The cell is dumb. The payload is smart. Same pattern as the kernel — the kernel doesn't know what it traverses. Now the cells don't know who they're serving.

## Threshold Descent

Frames surface in activation order — most matched first. The cell that receives the result runs a cheap inference against the payload: is this enough to do the binding reality's work? Two outcomes:

1. **Enough** — structured unwind. The cell does the work the payload describes and the stack unwinds back up through the layers.
2. **Not enough** — continue traversing. The kernel keeps descending into lower-activation matches. More context loads. Next threshold check.

The stack itself is the cognitive process. Each layer is a threshold gate. The deepest you go is determined by how much context you needed before the work could happen. Simple situations unwind fast — first match is enough. Complex situations descend deeper — the being "thinks harder" by traversing further into the index. The payload tells each cell what "sufficient" means for this particular binding reality's question.

## Upgrade the Cognition, Not the Memory

Cells are dumb containers. Competence is on the payload. If the indexing strategy improves, the payload logic improves, the threshold evaluation improves — the existing cells don't change. The memories are the same. The topology is the same. You're upgrading the lens, not rebuilding the library.

This means:
- **No migration.** Better retrieval strategy ships as a payload update. Existing memory store stays intact.
- **Different beings, same store.** A being with a more sophisticated Goal competence retrieves richer context from the same frames a simpler being would skim past. The memory infrastructure is shared. The intelligence is in the payload.
- **Biology parallel.** Neurons don't get replaced when you learn a better retrieval strategy. The connections and content stay. What changes is how you search — better cues, better associations, better evaluation of what you find.

## Two Retrieval Modes, Same Mechanism

**Associative (first wave)** — current attention state activates broadly. The being doesn't choose what comes back — the activation gradient does.

**Direct (think re-descent)** — the binding reality names specific keys from what the first wave returned. Partitioned query, only the intersection lights up. Fast, narrow, targeted.

## Query Correction at the Index Layer

Failed queries don't propagate back immediately. A fast inference call intercepts the miss, knows the schema, reformulates (broaden key, drop a dimension, try adjacent), and retries. Bad query + good data = index fixes locally. Good query + insufficient data = genuine miss, propagates back for think re-descent.

## Think/Work Express Switch

On the express path, each binding reality reads what came back and either:

1. **Think** — result insufficient, re-descend with a narrower or different query. Amplitude attenuation bounds depth.
2. **Work** — result sufficient, do the job. Vote, compute delta, whatever the binding reality's role is.

## Three-Layer Noise Filtering, Nothing Deleted

1. **Amplitude (mechanical)** — kernel activation scoring. No inference, no opinion. Most noise dies here.
2. **Binding reality (inferential)** — relevance decisions linked to the context window that informed them. Traceable.
3. **Suppressed, not removed** — low activation, not gone. State changes can resurface anything.

## Full Retrieval Path

```
Binding reality keys + entity keys
  -> reverse lookup -> candidate frames
    -> episodes read for grounding (what arc was this part of)
      -> stress/emotion trace on episodes re-ranks
        -> binding reality snapshots on each frame
          -> cells (the content)
```

## Full Traversal Shape

```
Coherence (ACC / Frontopolar)
  +-- Children: [Goal, Motivation, Prospective]
       +-- Each child's Children(): its episodic index
            +-- wave enters index -> frames -> episodes for context
            +-- result returns via observe
            +-- Express: think (re-query) or work (vote)
```
