# Episodic Encoding — Working Memory to Hippocampus

> **Reconciliation (2026-06-23).** Vocabulary drift; the retrieval mechanics
> (frames, episodes, typed indexes, retrieval-as-traversal, state modulation) hold
> and this doc is still the reference for them. Read **"Coherence" → MidLevel**,
> **"binding realities" → the self/other agents + entity cells**. Its core WM claim
> ("working memory is not a separate buffer — it is the activated subset… in the
> focus of attention," §"Working Memory Is the Attention Map") is **consistent with
> the current model** (WM = the lit-up subgraph). The `goalID/agentID → []episode`
> typed indexes are now realized as **entity cells whose `Children()` are the
> setpoints stamped on them** (see [entities-and-surroundings.md](entities-and-surroundings.md)).

## The Gap

We have when episodes switch (prediction delta + stress modulation). We don't have how working memory contents flow into episodic binding — the gating, the filtering, and the structure of what gets recorded.

## Working Memory Is the Attention Map

Working memory and attention are the same structure. One map on the slice, not two separate systems.

What the docs previously described as two things are the same thing from two angles:
- **Attention** — what's currently active, what penetrates the membrane, what binding realities read
- **Working memory** — what's currently held, small capacity, decays fast, raw recent state before absorption into traces

"What you're attending to" and "what you're holding in working memory" are the same set. Small capacity is attention being bounded. Fast decay is attention fading. The attention map on the slice IS working memory.

All binding realities read from it:
- **Episodic** scores overlap against it and absorbs from it when items decay out
- **Prediction** compares against it ("what actually arrived")
- **Motivation** monitors dependencies in it passively (if the dependency cell is in the attention map, Motivation reads it for free — the "passive mode" from the dependency graph)
- **Emotion** pattern-matches the current signal shape from it
- **Goal** reads the delta from it

Matches Cowan's embedded-processes model: working memory is not a separate buffer — it is the activated subset of long-term memory currently in the focus of attention.

## Frames Are Recorded Working Memory

A frame is a snapshot of the attention map at one encoding pulse. Working memory is the live state. The frame is the recorded state.

When Episodic encodes, it captures the conjunction of what was co-present in the attention map after prediction filtered it. That capture is the frame. Encoding IS taking a frame of the current attention map.

The frame is what makes working memory durable. Without it, attention is transient — it decays and is gone. The frame is the deposit. There is no separate "absorption" step from working memory into episodic — the encoding pulse that creates a frame is the absorption.

A frame is not a full copy of working memory. Prediction gates what gets into the frame — only the mismatch after prediction subtraction passes through. That's why some frames are dense (lots of surprise, most of the attention map encoded) and some are thin (mostly predicted, little passes).

A frame is born in one episode, but the cells it references can appear in frames across many episodes. The frame owns the binding pattern — which cells were co-present. The cells themselves are shared across the topology. Same cell, different conjunctions, different episodes.

## Retrieval Structure — Frame Is the Primitive

The frame is the atom of episodic memory. The episode is a derived structure — a synthesized understanding that contextualizes frames, not a container that owns them.

Frames are what get retrieved. Episodes are what ground the retrieval. The frame is the hit — the specific moment that matches. The episode is the explanation — the arc it belonged to, what it was about. You don't need the episode to find the frame, but you need the episode to understand what the frame meant.

This inverts the ownership model. Frames don't belong to one episode. A frame points up to N episodes (one per segmentation dimension). Episodes don't own frames — they contextualize them. Episodes are synthesized caps over frames, the same way understandings sit above memories in the promotion gradient. The frame is the trace. The episode is the understanding that formed over those traces.

```
Layer 1 — Typed Indexes (per-binding-reality + entity)
       ↓ inverted index lookup
Layer 2 — Frame (the primitive — one encoding pulse)
       ↓ points up to
Layer 3 — Episode(s) (synthesized context — one per dimension)
       ↓ carries stress/emotion trace
    Binding Realities on the frame (the conjunction — what was bound)
```

### Layer 1 — Typed Attention Maps

Attention is based on a thing. Each type gets its own inverted index map. Each scores independently based on attention. The activation product across all matching types gives the final score.

```
goalID   → []episodeID
agentID  → []episodeID
userID   → []episodeID
placeID  → []episodeID
deviceID → []episodeID
toolID   → []episodeID
```

Entry point for all retrieval. The being's current attention IS the query — active goals, present agents, current place, active device, active tool. Each typed map returns episodes where that thing was present. Episodes matching multiple types score highest.

These scores structure what the agent sees in the frame. They may also impact traversal strength (open question).

### State Modulation — Stress and Emotion Traces

Every episode carries a stress/emotion fingerprint — the state of the system when the episode was active. This trace is captured automatically as part of the episodic write infrastructure. When a binding reality writes to the slice, the current stress level and emotion valence get stamped onto the episode trace. Same dumb mechanism as frame binding — cells don't know it's happening.

The trace is not an index. It's a property on the episode that gets compared against current system state during activation scoring. Two roles:

**Modulator (primary)** — entity index activates candidate episodes, the state trace re-ranks them. Congruent episodes get boosted, incongruent get suppressed. Fast — just a distance calculation on already-retrieved candidates. The entity index did the expensive narrowing; the state trace is a cheap re-ranking pass.

```
people/places/things → reverse lookup → candidate episodes
  → each candidate has a stress/emotion trace
  → current system state scores against those traces
  → congruent episodes boosted, incongruent suppressed
```

**Independent activator (rare)** — stress/emotion can surface episodes without entity match. Pure state-dependent recall. "I feel like this and I don't know why." Scans traces without the entity index first. Expensive, but this is the path for state-dependent memory that has no cognitive trigger — just the feeling.

This is state-dependent memory from biology. Bower's mood-congruent memory, state-dependent learning. You recall things better when you're in the same state as when they were encoded. Stressed now? Episodes encoded under high stress activate more strongly. Calm now? Calm episodes surface. The being's current state shapes what it remembers, not just what it's attending to.

### Episode — Synthesized Context

An episode is a synthesized cap over a sequence of frames — the understanding that formed over those traces. Not a container. A frame points up to its episodes; an episode references its frames. Many-to-many across segmentation dimensions.

```
Frame 7 points up to:
  ├── Temporal episode 12
  ├── Goal episode 4 (ship v.1)
  ├── Agent episode 7 (michael)
  └── Place episode 2 (coffee shop)
```

The episode provides grounding: what was this frame about, what arc was the being in, what was the larger context. Without the episode you get déjà vu — a strong frame match with no story around it. With the episode you get "I remember this, and it was part of this larger thing, which was about X."

Episodes carry stress/emotion traces and provide narrative structure. Frames are retrieved first; episodes are read for context.

### Layer 3 — Frame

A frame is one encoding pulse — the conjunction of what was co-present at that moment, after prediction filtered it. A frame holds binding realities. One frame, many binding realities.

```
Frame
  ├── Binding Reality A (goal state at this pulse)
  ├── Binding Reality B (agent state at this pulse)
  ├── Binding Reality C (emotion valence at this pulse)
  └── ...
```

Content lives on the cells that the binding realities point to. The frame owns the binding pattern — which binding realities were active and what they were bound to at this pulse.

### The Full Lookup

```
Binding reality keys + entity keys (goal, agent, user, place, device, tool)
  → typed inverted index scores per type
    → frames (ranked by activation product across types)
      → binding realities on each frame (the conjunction)
        → cells (the content)
      → episodes (read for grounding — what arc was this frame part of)
        → stress/emotion trace on the episode re-ranks
```

You enter through keys. You hit frames. You read episodes for context. You exit with cells. The frame is the retrieval target. The episode is the explanation.

### Distributed Indexes, Shared Store

The hippocampus is shared memory with per-binding-reality indexes on top. The boundary is the frameID.

**Per-binding-reality (above frameID)** — each binding reality has its own typed key into the frame layer. Each binding reality owns its own door into the store:

```
goalID        → []frameID   (Goal's door)
dependencyID  → []frameID   (Motivation's door)
valencePattern → []frameID  (Emotion's door)
obligationID  → []frameID   (Prospective's door)
agentID       → []frameID   (entity index)
userID        → []frameID   (entity index)
placeID       → []frameID   (entity index)
deviceID      → []frameID   (entity index)
toolID        → []frameID   (entity index)
```

**Shared (below frameID)** — once you're at the frame, you're in shared infrastructure. The frame, its binding reality snapshots, and the episodes it points up to — all the same store, same structure, regardless of which binding reality's key got you there. Episodes provide grounding context and carry stress/emotion traces for re-ranking.

The indexes are distributed. The store is centralized. The frameID is the boundary between "my door" and "shared house." Each binding reality doesn't need its own copy of episodic memory — it just needs its own key type into the shared layer. The entity index (people/places/things) is one set of keys. Each binding reality adds its own domain-specific key. All resolve to frameIDs and from there it's the same traversal.

### Retrieval as Traversal

The episodic store is traversable topology — not a retrieval API. Each binding reality child of Coherence spawns a wave into its own index. The kernel handles scoring and traversal. Results return through observe (Perception Invariant). No special memory retrieval mechanism — just waves into structure.

On the express path, the binding reality reads what came back and has two options:

1. **Think (re-descent)** — the result isn't sufficient. Spawn another wave with a different or narrower query. The kernel calls observe again. Amplitude attenuation is the natural constraint — you can't query forever.
2. **Work** — the result is sufficient. Do the job — vote, compute delta, whatever this binding reality does.

```
Coherence (ACC / Frontopolar)
  └── Children: [Goal, Motivation, Prospective]
       └── Each child's Children(): its episodic index
            └── wave enters index → episodes → frames
            └── result returns via observe
            └── Express: think (re-query) or work (vote)
```

The kernel doesn't know it's traversing memory — it's traversing spaces, scoring amplitude, delivering results. Same mechanism all the way down.

### Associative vs Direct Query

Two retrieval modes, same index, same store, same kernel traversal. The difference is what the impulse carries.

**Associative (first wave)** — current attention state activates the index broadly. Whatever matches surfaces. The being doesn't choose what comes back — the activation gradient does. This is the default retrieval path.

**Direct (think re-descent)** — the binding reality names specific keys. `agentID:michael + goalID:help-with-calendar → []episodeID`. The query is partitioned — only the intersection of those named keys lights up. Fast, narrow, targeted. The hippocampus doesn't know the difference — it just activates what matches the keys it received.

This is how the being "pulls on a thread." First wave surfaces a memory. The binding reality reads it, notices something — a name, a goal, a place — and re-descends with those specific keys. That lights up a different slice of episodes. The being follows a chain of associations not by magic but by direct queries triggered by what the last result contained. Each re-descent is bounded by amplitude attenuation. The depth of retrieval is paid for in energy, not limited by a hardcoded iteration count.

### Query Correction at the Index Layer

Failed queries don't propagate back up immediately. The index layer intercepts the miss, fires a fast inference call that knows the schema, and retries with a corrected query before returning a failure.

```
wave hits index → no match
  → fast inference: knows index schema, knows what was queried
  → reformulates: broaden key, drop a dimension, try adjacent entity
  → retry → result or genuine miss
  → only genuine misses propagate back up
```

The inference call is cheap — it knows the index shape (what keys exist, what combinations are valid, what's adjacent to what was queried). It doesn't need the binding reality's context or goals. It just needs to turn a near-miss into a hit.

The binding reality never sees the reformulation. It either gets a result or gets a real miss. The correction is internal to the index layer — same pattern as everything else. Dumb mechanism, the binding reality doesn't know the index fixed itself.

This separates two failure modes:
- **Bad query, good data** — the index layer handles it. Broaden the key, drop a dimension, try adjacent. Fast, local, no round trip.
- **Good query, insufficient data** — genuine miss. Propagates back to the binding reality. The "think" re-descent handles this — the binding reality changes strategy, asks a fundamentally different question.

### Noise Filtering — Three Layers, Nothing Deleted

As memory grows, retrieval returns more candidates. Filtering is necessary. But filtering is lossy, and LLMs carry biases (completion bias, conventional bias, consensus bias — see `design-docs/bias/`). An LLM deciding "this doesn't matter" is the conventional bias in action. The thing that doesn't fit the current frame might be the most important signal.

Three layers of filtering, from mechanical to inferential:

1. **Amplitude (coarse, mechanical)** — the kernel's activation scoring. Low activation falls below threshold naturally. No inference, no opinion. Most noise dies here. This is bias-free — the kernel doesn't know what it's filtering, it just scores amplitude.

2. **Binding reality (fine, inferential)** — each binding reality makes relevance decisions on its retrieved context. Inference-based, but every filtering decision is linked to the context window it was made on. The decision and the retrieval that informed it are bound together — traceable. If later evidence says "that thing you dismissed was important," you can re-traverse the context window that led to the dismissal and understand why.

3. **Nothing gets deleted** — filtered means suppressed, not removed. Low activation, not gone. Suppressed signals remain in the store at low activation. If the being's state changes — new goal, new stress level, new entity present — state modulation can resurface things that were suppressed under different conditions. The being never permanently loses signal.

No single system is the arbiter of relevance. Amplitude handles the bulk mechanically. Binding realities handle the fine-grained filtering with traceable decisions. State changes can reverse any suppression. The balance between "filter the noise" and "it might actually matter" is maintained by never deleting and by making the filtering state-dependent — what's noise under one set of conditions is signal under another.

**Durability resolution:** Episodic binding happens on the slice (writes automatically captured into frames during traversal), but the durable storage is the central store. Frames are born on the slice, then graduate to the central store when the traversal closes. The slice is the workspace, the store is the archive.

### Structural Extension (future)

Conjunction hash cross-linking — hashing the binding pattern of a frame to find other frames with the same structural shape across unrelated episodes — is a natural extension of this structure. Not needed for v.1. The three-layer lookup is sufficient. When the pattern proves it needs structural recurrence detection, the conjunction hash adds a lateral link at Layer 3 without changing the ownership model.

## Prediction Indexing

The Prediction binding reality owns a reverse index on the slice:

```
predictionID → {frame, episode}
```

Same retrieval pattern as Episodic (`cellID → []episode`) and Motivation (`goalID → []cellID`). Each binding reality owns its own reverse index — the index is how it answers questions without iterating.

### Ownership (forward)

```
Episode owns → Frames (ordered sequence)
Frame owns → Predictions (set at that encoding pulse)
```

A prediction is born at a specific encoding pulse — one frame. The frame holds both sides: what was predicted and what actually came through (the attention map snapshot after filtering). The delta is derivable from the two.

### Slice State

The slice holds current prediction state:

- `CurrentPrediction` — what's predicted right now, awaiting comparison with the next signal
- `CurrentEpisodeID` — the active episode
- `CurrentFrameID` — the active frame

These give Prediction's three runtime reads a concrete path:
- "Which episode am I in?" → `CurrentEpisodeID`
- "What actually arrived?" → the attention map on the slice
- "What did I predict?" → `CurrentPrediction`

### Historical Lookup

"When I predicted something like this before, what actually happened?" The reverse index finds past frames where similar predictions were set. Read the delta from those frames. This is the pattern history: shrinking delta = learning this context, growing = model is stale, consistent highs on a signal type = blind spot.

## Prediction Lives in Observe

Prediction is not a separate step between working memory and binding realities. Prediction is what Observe does.

When motor fires, it sends an **efference copy** — a copy of the motor command — back to the sensory path. This is how the brain distinguishes self-generated signals from external input (Von Holst & Mittelstaedt, 1950):

- **Reafference** — sensory input caused by your own action
- **Exafference** — sensory input from the external world

The brain subtracts predicted reafference from total input. What remains after subtraction is the real news from outside. That's why you can't tickle yourself — your motor sends the prediction, sensory receives the actual touch, delta is zero, suppressed.

### The Membrane's Observe Has Two Modes

Observe needs **competence** — two distinct ObserveFn commands based on what kind of signal is arriving:

**ObserveFn: efference (self-signal)** — motor just fired. Receive the prediction payload. "Here's what I expect to happen as a result of my action." This primes the binding realities before external input arrives:
- Episodic pre-activates frames from past episodes where similar actions led to similar outcomes
- Motivation pre-evaluates whether predicted response satisfies or breaks dependencies
- Emotion shifts valence based on predicted outcome (anticipation, anxiety)
- Action pre-loads the next likely primitive

**ObserveFn: exafference (other-signal)** — external input arrived. Compare against the primed prediction. Subtract. What remains is surprise.

### The Flow

```
Motor fires
    ↓
Efference copy → Observe (self-signal mode)
    ↓ prediction primes binding realities
World responds
    ↓
Actual input → Observe (other-signal mode)
    ↓ compare against prediction, subtract
    ↓ only surprise passes
Surroundings updated with prediction error
    ↓
Binding realities process the error
    ↓
Episodic binds the conjunction
Motivation checks dependencies
Action records what fired
Emotion tags valence
```

### Prediction Fires on Both Sides

Prediction brackets every exchange:

- **Before action** — "If I say this, B will probably say that." The prediction is set.
- **After input** — B responds. Compare actual against expected. Delta is the encoding signal.

The being is not idle between act and receive. The prediction reshaped the brain's interior. By the time actual input arrives, the binding realities are already configured for what was expected. A wrong prediction is disruptive because the entire brain pre-configured for something that didn't happen. The unwinding of that primed state IS the surprise.

Prediction is the gate between working memory and episodic encoding. It doesn't just trigger episode boundaries — it filters what gets encoded at all. The boundary is the extreme case where delta is large enough to flush the buffer and start a new episode. Below that threshold, prediction still filters every frame — deciding what's novel enough to pass through to binding.

**Prediction doesn't gate whether you CAN process. It gates how much you HAVE to.**

No prediction = everything is mismatch = everything gets through. The comparator has nothing on one side. That's maximum surprise. Full encoding. That's why first experiences are vivid and why cold start is expensive — no predictions, every frame is dense, every binding reality gets flooded. A mature being is cheap because predictions suppress most of working memory. Only mismatch passes. Sparse frames. Efficient encoding.

## Frames — The Encoding Pulse

An episode is not one flat snapshot. It is a sequence of frames. Each frame is one encoding window — what was co-present at that moment, after prediction filtered it.

```
Episode
  ├── Frame 0: cells A, B, C co-present
  ├── Frame 1: cells A, D, E co-present
  ├── Frame 2: cells D, F co-present
  └── Frame 3: cells A, B, F co-present
```

Cell A appears in three frames but was bound with different conjunctions each time. The episode knows the sequence. The frames know the binding.

Prediction filters at each frame boundary, not once per episode. "What's novel enough to encode right now" is a per-frame question. Some frames are dense (lots of mismatch, lots gets through). Some frames are thin (mostly predicted, little gets encoded). Stress modulates how sensitive that per-frame filter is.

### Biology Mapping

This maps to theta-gamma coupling in the hippocampus:
- The episode is the theta wave — the chapter
- Each frame is a gamma burst within the theta cycle — the sentence
- The hippocampus pulses: open, encode what's here, close. 4-8 Hz theta, 4-7 gamma bursts per cycle
- Without pulsing, you get smear — everything bleeds into everything. Discrete frames create stronger binding signals than "everything active during this traversal"

## Physical Context as Binding Keys

Place, device, and tool extend the typed attention maps beyond the runtime into the physical world:

```
Place (GPS — the coffee shop, the office, home)
  └── Device (the laptop, the phone, the whiteboard)
       └── Tool (the terminal, the browser, the calendar)
```

Each is a typed map in Layer 1. The coffee shop activates every episode where the being was at those coordinates. The laptop narrows to episodes where that device was the interface. The terminal narrows further to episodes where that tool was in use.

They compose multiplicatively. Place + device + tool is three concurrent context signals. Episodes matching all three score highest. Episodes matching only place score lower but still surface. The gradient gives you "I vaguely remember doing something here" (place only) versus "I know exactly what I was working on" (place + device + tool).

GPS is the first binding key that comes from the physical world, not the runtime. The runtime doesn't generate location — the device reports it. Place is a sensory input that flows through Observe like everything else. The Membrane receives GPS the same way it receives keyboard input. The episodic index doesn't care where the signal came from. Place, device, and tool are all Spaces — same interface, same activation. They just happen to map to physical context instead of cognitive context.

## Episode Boundaries

Prediction delta is still the mechanical trigger for segmentation. Stress modifies the sensitivity:

```
effective_segmentation_threshold = base_threshold / (1 + stress)
```

At stress = 0, normal threshold. As stress rises, smaller prediction deltas trigger new episodes. More fragmentation, denser encoding per frame.

When a boundary fires:
1. Current episode closes
2. Motivation decides which goals carry forward vs close with the old episode
3. New episode begins — fresh frame sequence
4. Closed episode becomes a retrieval target through the typed attention maps

### Multi-Dimensional Segmentation

Episode cutoff (prediction error / instability crossing threshold) is the core mechanism — but it's not inherently temporal. Temporal is just one dimension, and likely the default, but the same cutoff mechanism applies along any dimension:

- A goal's pursuit has its own coherent arc and boundary, independent of the clock
- A relationship with a particular agent can shift and mark its own episode boundary
- A place, a tool, an emotional pattern — each can be its own segmentation axis

The result is overlapping episode structures over the same underlying frames. A single frame can belong to multiple episodes simultaneously — one per dimension — each with boundaries determined by that dimension's own instability signal, using the same cutoff mechanism.

```
Frame at 3pm Tuesday belongs to:
  ├── Temporal episode 12    (the afternoon work session)
  ├── Goal episode 4         (third attempt at this subtask of "ship v.1")
  ├── Agent episode 7        (the conversation with michael that changed direction)
  └── Place episode 2        (second time working at the coffee shop)
```

Same frame, different episodes, different arcs, different boundaries. Each dimension segments independently because each dimension has its own instability signal.

This connects directly to the per-binding-reality indexes (Goal by goalID, Prospective by obligationID, etc.). Those indexes don't just retrieve loosely-matching frames — they define their own segmentations along that key's dimension, returning coherent arcs rather than scattered hits.

Retrieval payoff: "what happened with michael" doesn't return a flat set of frames ranked by activation. It returns a sequence of agent-relationship episodes, each with its own beginning, arc, and resolution. The structure IS the narrative. You don't reconstruct it — you retrieve it.

## Open Questions

- What defines the frame boundary within a traversal? Each depth level? Each push/pop? A prediction sub-threshold?
- How many frames does a typical episode hold before segmentation?
- Should frame ordering be stored for sequence matching, or is set overlap enough?
- Do Layer 1 attention scores impact traversal strength inside the episode, or only episode selection?
- What is the concrete shape of the attention map on the slice? Flat (`map[string]float64`, cellID → strength) or typed (separate maps per goal/agent/place/device/tool)?
