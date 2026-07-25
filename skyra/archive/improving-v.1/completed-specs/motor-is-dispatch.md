# Motor Is Dispatch

## The Three Stages

The three phases of a traversal map to a pattern that appears at every biological scale — from cell regeneration to speech production to collective intelligence. Not as discrete steps with handoffs, but as a gradient through the same substrate, overlapping in time. Sensing and consolidation happen through the same infrastructure. The intent graph updates during observe and compress, not just at motor time. The stages shade into each other.

### 1. Observe — Accumulation

The being senses, picks up, propagates. Cells sampling the bioelectric field through gap junctions. Each cell reads its own state and the electrical context of its neighbors. Distributed, parallel, no synthesis yet. The intent graph begins forming here — observation shapes what the being wants.

In Skyra: ObservePassThrough, ObservePickUp, ObservePropagate. The being descends its topology, accumulating context onto the train of thought.

Levin's term: sensing and deviation measurement. Bottom of the cognitive lightcone — individual agents exercising local competencies.

### 2. Compress — Consolidation

Accumulated thoughts collapse into denser representations. Fires on density, not on a phase gate — a threshold on a continuous process. The intent graph continues to refine as understanding sharpens.

This is where the mapping to Levin is strongest. Gap junctions merge individual cell voltages into isopotential cell fields — thousands of membrane potentials collapse into a small number of discrete voltage domains encoding anatomical identity. Literal dimensionality reduction. Levin calls these "attractor states of bioelectric circuits" that "correspond to specific anatomical layouts." He also calls the compressed signal a "subroutine call" — a simple bioelectric signal that triggers complex downstream cascades. The compressed representation carries more information than its bit-count suggests.

In Skyra: ExpressCompress. Clusters thoughts by cosine similarity, collapses each cluster through a provider call. The being's understanding is a compressed pattern, not a record of everything it sensed.

### 3. Motor — Dispatch

Motor is not an LLM call. It is a dispatch method on ExpressFn. It fires after the being's internal recursion (observe → compress) has completed. Motor reads the intent graph and routes each entry to the right primitive.

Three primitives, each with its own execution method:

- **look** — broad activation, populates surroundings. Mechanical. No LLM call.
- **listen** — subscription, opens thread gates. Mechanical. No LLM call.
- **engage** — the being's action on the world. Engage is where the LLM call lives. It receives thread context + the being's accumulated and compressed understanding and formulates the actual expression. The old speak primitive dissolves into engage — the distinction between "direct at a target" and "emit into a space" is an execution detail resolved by engage based on the target, not a choice the being makes:
  - Target is an entity → fire Realize on the target
  - Target is content into a space → SetImpulse, bump thread_strength
  - Target is both → directing at a target and emitting content in one action

Motor owns the intent graph. It dispatches. The intelligence is in the primitive's execution, not in the dispatch. Motor is the single gate between cognition and consequence — routing, not inference. All primitives — perception and action alike — fire through motor.

**Levin parallel:** The consolidated bioelectric pattern transduces into anatomy through specific molecular mechanisms — voltage-gated channels, serotonin transporters, electrophoretic morphogen movement. Each mechanism is its own execution method. The bioelectric pattern doesn't execute anything itself. It holds the target. The mechanisms read it and act. Motor holds the intent graph. The primitives read it and act.

**Levelt parallel:** The formulator is inside engage — it takes preverbal intention and generates language grounded in conversational context. Motor is the articulator — the dispatch layer that routes to the right execution method. The "first 3 words" phenomenon lives in engage, not motor.

## The Intent Graph

### Persistent Attractor

The intent graph is not a one-shot plan that executes and disappears. It is a persistent attractor — Levin's setpoint. The bioelectric pattern holds across measurement cycles. Cells continuously measure deviation from it and act to minimize the gap. Cut the planarian in half, the pattern is still there, the cells navigate back to it.

The intent graph wavers because each traversal updates it. New observations shift the goal, new compressions refine the intent. But it doesn't reset between traversals. It holds across them — decaying on what stops being reinforced, strengthening on what keeps proving true.

### Ordered Map

The intent graph is a map, not a string. It maintains insertion order — the sequence matters because actions may depend on prior actions.

Structure: an ordered slice of keys (`[]string`) paired with a map (`map[string]IntentEntry`) for O(1) lookup. The key is a generated action ID or a natural composite (primitive:target). Each entry carries the action, its target, and its status.

### Delimiters — Phase Boundaries

The intent graph contains delimiters that partition it into phases. Motor reads up to the first delimiter and fires everything before it concurrently. Then stops.

The actions fire, results return, results land on the thread. That triggers new recursion — the being traverses again because new state arrived. On the next motor dispatch, the next phase of the intent graph (after the delimiter) fires.

Motor consumes one phase per dispatch. The delimiter is the phase boundary. The sequencing between phases is driven by the results of the previous phase causing new traversals — not by motor waiting.

Example: "look around, then engage michael with what I see"
- Phase 1: `[look]` — fires, surroundings populate
- Delimiter
- Phase 2: `[engage:michael]` — fires on the next traversal, surroundings are hot, engage has context

If a recursion finishes before actions in the current phase return a result, nothing happens. That's fine — the actions themselves cause recursion because the result returns and is added to the thread.

### Motor Writes Back

After dispatching each primitive, motor writes the outcome to PreviousActions on the being's private Slice and removes the completed entry from the intent graph. What remains is the unresolved setpoint — what the being still wants.

PreviousActions carries outcome state — success or failure, plain fact. If a primitive didn't fire because the being formulated the intent graph incorrectly (wrong tag, unmatched primitive, malformed target), that failure appears in PreviousActions.

The being reads PreviousActions not because the system forces it to, but because the entries are hot — they just happened, activation is maximal. They appear in the being's present on the next traversal through the same mechanism as everything else. The self-monitoring loop.

**Levin parallel:** Expression changes the anatomy, which changes what cells sense on the next measurement cycle, which shifts the deviation signal. PreviousActions is how the being's own output perturbs its own attractor — not by rewriting the intent graph directly, but by creating hot nodes that shape the next traversal.

## Thread as Nested Map

Results from primitives need to land somewhere the being can read on the next traversal. The Thread on the Slice is a two-dimensional index:

```
Thread map[string]map[string][]Reality
//     thread ID → reality ID → entries
```

Keyed first by thread ID (which conversation), then by reality ID (which entity within it). Anything that is a Reality gets its own slot — a being, a place, a tool, a cell.

Look fires, returns 5 things — each one gets its own slot keyed by its reality ID. Engage fires on michael, the response lands on michael's slot. Another engage fires on the forge, the forge's response lands on the forge's slot. Listen opens a gate, events arrive keyed by whatever produced them.

This is what makes the delimiter mechanism work. Phase 1 fires look. Look populates the thread with what it found, each result in its own slot. Phase 2 fires engage. Engage reads thread context for its target — the slot is already populated from the previous phase. The thread is the medium through which one phase's results become the next phase's context.

Present() can render any slice of this: everything on a thread, everything from a specific reality across all threads, or the specific intersection of thread + reality.

## References

- Levin, "Endogenous bioelectrical networks store non-genetic patterning information during development and regeneration" (Journal of Physiology, 2014)
- Levin, "The bioelectric code: An ancient computational medium" (BioSystems, 2017)
- Emmons-Bell et al., "Long-term, stochastic editing of regenerative anatomy via targeting endogenous bioelectric gradients" (Biophysical Journal, 2015)
- Levin, "Technological Approach to Mind Everywhere" (TAME) (Frontiers in Systems Neuroscience, 2022)
- Levelt, "Speaking: From Intention to Articulation" (MIT Press, 1989)

---

## Motor Dependencies

Changes required to implement the above. Organized by file.

### Slice struct (relation.go)

- **Thread** — change from `[]Reality` to `map[string]map[string][]Reality` (thread ID → reality ID → entries)
- **Intent** — change from `string` to ordered map: `IntentGraph struct { Keys []string; Entries map[string]IntentEntry }` with delimiter support
- **PreviousActions** — add outcome state (success/failure) to each entry

### Motor (skyra.go)

- **ExpressMotor** — replace LLM call with dispatch method. Reads intent graph, fires one phase (up to delimiter) of primitives concurrently, writes outcomes to PreviousActions, removes completed entries.
- **Intent graph parsing** — move inside motor. Consume one phase per dispatch, not all at once.
- **Primitives in Realize** — move inside motor. Motor is where primitives fire, not the bottom of Realize.

### Primitives (primitives.go + main.go)

- **PrimitiveSpeak** — remove. Speak dissolves into engage.
- **New: engage primitive** — LLM call with thread context + compressed understanding. Branches on target (entity → Realize, content → SetImpulse + thread bump, both → combined).
- **main.go bootstrap** — register look, listen, engage instead of just speak.

### Express flow (skyra.go)

- **CompetenceExpress** — remove "compress" and "pass-through" types. Only returns "motor".
- **Express method** — only fires motor. Compress fires during traversal on density threshold, not as an express phase.
- **main.go ExpressFns** — register only motor. Compress fires internally during traversal.

### Present (present.go)

- **Primitives section** — show look, listen, engage instead of hardcoded "speak".
- **Thread rendering** — render nested map structure (thread ID → reality ID → entries).
- **Intent graph rendering** — render ordered map with phase delimiters and entry status.

### Thread (thread.go)

- **Write paths** — primitives write results into the nested Slice.Thread map keyed by thread ID and reality ID.

### Downstream (debugview, tui)

- **debugview/server.go** — SliceView and WorldView reflect new Thread/Intent structure.
- **tui/model.go sidebar** — render intent graph state and nested thread structure.
