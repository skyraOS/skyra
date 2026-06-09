# Episodic Encoding — Working Memory to Hippocampus

## The Gap

We have when episodes switch (prediction delta + stress modulation). We don't have how working memory contents flow into episodic binding — the gating, the filtering, and the structure of what gets recorded.

## Three-Layer Retrieval Structure

The hippocampus is a three-layer inverted index. Each layer fans out through a different key. One direction of ownership. No back-pointers.

```
Layer 1 — Typed Attention Maps (person/place/thing)
       ↓ inverted index lookup
Layer 2 — Episode
       ↓ 1..*
Layer 3 — Frame
       ↓ 1..*
    Binding Realities (the conjunction — what was bound)
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

### Layer 2 — Episode

An episode owns an ordered sequence of frames. One episode, many frames. The episode is the chapter — what happened across a sustained context.

```
Episode
  ├── Frame 0
  ├── Frame 1
  ├── Frame 2
  └── Frame 3
```

The typed attention maps at Layer 1 get you to episodes. From there you walk down into frames.

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
Attention (goal, agent, user, place, device, tool)
  → typed inverted index scores per type
    → episodes (ranked by activation product across types)
      → frames within those episodes (ordered sequence)
        → binding realities on each frame (the conjunction)
          → cells (the content)
```

You enter through attention. You exit with cells. Everything in between is index traversal.

### Structural Extension (future)

Conjunction hash cross-linking — hashing the binding pattern of a frame to find other frames with the same structural shape across unrelated episodes — is a natural extension of this structure. Not needed for v.1. The three-layer lookup is sufficient. When the pattern proves it needs structural recurrence detection, the conjunction hash adds a lateral link at Layer 3 without changing the ownership model.

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

GPS is the first binding key that comes from the physical world, not the runtime. The runtime doesn't generate location — the device reports it. Place is a sensory input that flows through Observe like everything else. The Membrane receives GPS the same way it receives keyboard input. The episodic index doesn't care where the signal came from. Place, device, and tool are all Realities — same interface, same activation. They just happen to map to physical context instead of cognitive context.

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

## Open Questions

- What defines the frame boundary within a traversal? Each depth level? Each push/pop? A prediction sub-threshold?
- How many frames does a typical episode hold before segmentation?
- Should frame ordering be stored for sequence matching, or is set overlap enough?
- Do Layer 1 attention scores impact traversal strength inside the episode, or only episode selection?
