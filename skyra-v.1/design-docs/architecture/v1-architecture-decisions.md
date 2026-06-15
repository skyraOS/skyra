# v.1 Architecture Decisions

## What Skyra Is

A cognitive runtime where every entity implements a single `Space` interface. A Relation (the signal) propagates through a topology of cells. Descent is observation — the being absorbs from its topology. Ascent is expression — the being outputs through the same topology. Weight updates happen on the return path. One pass does perception, cognition, memory, and action.

## The Current Runtime

Go binary. Bubbletea TUI. Debug view on `:3333`. The genome declares beings and their relationships. `Realize()` is the main method of every Space. The world loop routes impulses to the highest-scoring entity by activation.

---

## Architectural Decisions

### 1. The Being Is the Membrane

The being exists one level up from its internal topology. It's not a node in its own graph — it's the boundary. Observe is sensory (reads the world, routes inward). Express is motor (reads internal votes, acts on the world). Same cell, two faces, already in the recursion as the top frame.

**Why:** Sensory and motor need to be in the recursion but face opposite directions. Sensory reads the world. Motor reads the binding realities. Making them the being's Observe and Express dissolves the need for separate cells and keeps them in the existing descent/ascent flow.

### 2. Being / Slice / Relation Split

Three layers, clean ownership:

- **Being (Brain)** — durable. Per-being. The being's specialized binding realities. Thread, Action, Goal, Motivation, Emotion. This is the being's private topology. Persists whether or not a signal is passing through.
- **Slice** — durable. Per-being. Universal activation substrate regardless of DNA. Activation records, traversal count, thresholds, fields, and traversal trees. The physics of traversal — same for every being.
- **Relation** — transient wave. Carries impulse, energy, and references needed to traverse. Doesn't own any being's state. Just the medium.

Ownership is separate from reachability. The Relation holds references so it can traverse realities, but durable state belongs to the Being and its Slice. Brain state is specialized and per-Being. Slice state is universal and per-Being. Relation state exists only for the current wave.

**Why:** Binding realities (Thread, ActionHistory) were sitting on the Relation — shared state holding private history. The being's topology was on the wrong struct. Fields are per-being (different brains produce different imprints) so they belong on the Slice, not the Relation. The being can't own its own traversal — traversal is physics.

### 3. Relationships Map Removed from Skyra

The `Relationships map[string]Space` on the Skyra cell is replaced by `Brain map[string]Space`. Binding realities are cells. Cells have members that are cells. All the way down. Same type everywhere.

**Why:** Relationships was redundant with the binding realities' member maps. Both tracked the same things. The binding realities already have weights, already gate through activation, already learn on the return path. They ARE the topology.

### 4. Five Binding Realities

Each is a cell with members, but each indexes differently:

| Binding Reality | Indexes by | What it answers |
|---|---|---|
| **Thread** | Episodic co-occurrence | What happened together? |
| **Action** | Primitive → instance → result (4 levels) | What happens when I do this? |
| **Goal** | Goal → resolution traces | What serves where I'm going? |
| **Motivation** | Goal → origin traces | Why am I going there? |
| **Emotion** | Affect state | What feels like this? |

**Why:** Each binding reality is a different indexing topology over the same substrate — same as how brain regions are different query strategies over the same neural medium. They fire in parallel, imprint fields onto the Slice, and the activation product across fields is the coincidence detection.

### 5. Motor Is Coincidence Detection

Motor is not a binding reality. It's the being's Express function. Each binding reality outputs a primitive + target (the shared vocabulary: look, listen, engage). Action fires when enough binding realities independently vote for the same primitive + target. The activation product across fields is the gate threshold.

**Why:** The basal ganglia doesn't deliberate — it detects convergence. MSNs fire when multiple cortical regions are simultaneously active on the same action. The activation equation's product already computes this.

### 6. Primitives Are the Shared Language

The primitives (look, listen, engage) give binding realities a common vocabulary to converge on. Every binding reality expresses its opinion as primitive + target + metadata. The primitive is the vote. The metadata is why.

**Why:** Without shared vocabulary, there's no way to detect convergence across binding realities that each track different dimensions.

### 7. Inference Fires at Disagreement, Not High Activation

The current design fires inference (propagate) at high activation. Biology says the opposite. High activation means the weights know the answer — reflex, no inference needed. Low activation means dormant. Inference fires in the middle — where binding fields disagree and weights can't resolve the conflict.

**Why:** Cells at bifurcation points actively compute. Cells deep in attractors just fire. A new being is expensive (everything is at boundary points). A mature being is cheap (deep grooves everywhere). Growth reduces inference cost.

### 8. Inference Owns Recognition, Causality, and Resolution

The error function — "did x resolve y" — cannot be computed from embeddings or weight history alone. Cosine similarity measures relatedness, not resolution. Recognition is its own capability that only the LLM has.

First time: inference recognizes the resolution. Weights record it. Second time: weights carry the recognition. No inference needed. This is how reflexes emerge — inference-assisted resolution becomes weight structure.

**Why:** We chased the loop — cosine doesn't work (proved experimentally, 0.47 for "hello" against "greet the user"), history doesn't capture meaning, prediction needs trained weights. The LLM is the only thing that understands resolution.

### 9. Inference Is Local on the Ascent, Expands to Global

No inference on descent — only on ascent. Error exists locally at each cell (does the cell's content match its members?). Each level fixes locally. As the recursion unwinds, each level sees the already-corrected state from below. Local fixes compound to global coherence.

**Why:** Small inference calls (one cell + its members) instead of one massive context window. Cost is proportional to disruption — no error means free ascent, error at every level means inference at every level. The being pays exactly what the disruption costs.

### 10. A Cell's Content Is Its Setpoint

Every cell has a goal — itself. The content is what the cell is trying to be. Members are the evidence. The error signal is the delta between content and what members say. Goal and content are the same thing.

**Why:** This is Levin's morphogenetic comparison. The target isn't stored externally — it's the coherence of the cell with what's below it. Parent changes (learning), relationships change (confirmation bias), or the parent splits (differentiation).

### 11. Cell Deposit by Signal Convergence

A new cell forms when signals from multiple binding fields converge — high co-activation product. The cell inherits membership in every binding reality that was active when it formed. Same Space instance, referenced by multiple binding realities.

**Why:** Biology doesn't place memories into regions. Traces form where signals intersect. The hippocampus isn't special — it's where co-occurring signals from different areas meet. The mechanism is the same everywhere.

### 12. Hebbian Reinforce

`Reinforce()` should be proportional to activation strength, not a flat tick. Stronger activation = stronger weight update.

**Why:** Synaptic plasticity is proportional to postsynaptic response magnitude. Flat reinforcement treats every firing event as equally important. Biologically wrong and architecturally wasteful.

### 13. Goal and Motivation — Coupled, Separate

Goal (forward — what serves where I'm going) and Motivation (backward — why am I going there) are both indexed by the goal but hold different members and have different error signals. The lifecycle: Motivation activates Goal, Goal pursues, Goal resolves, resolution context feeds Motivation, Motivation activates next Goal.

**Why:** "What to do" and "why to do it" are different questions with different failure modes. Goal error: resolution didn't close the gap. Motivation error: pursuing a goal whose originating conditions no longer exist.

### 14. Three Structural Roles — All Emergent

Nothing is declared. Position in the traversal determines role:

1. **The being** — the membrane. Top frame. Observe/Express.
2. **Binding realities** — cells with members. Indexes. The gate.
3. **Memories** — cells with content. Traces and understandings are the same thing at different points on the binding decay curve.

**Why:** Biology — cell fate is determined by signals, not declarations. A cell that accumulates members becomes a binding reality. A binding reality whose members decay becomes a memory. The promotion gradient.

### 15. Reflexes Are Universal Physics

When weights carry the full chain from impulse to action to completion, the traversal completes without inference. This is a reflex. It lives on the Slice (universal traversal mechanics), not on the DNA. Every being's reflex mechanism is identical.

**Why:** Reflexes weren't designed — they fell out of the architecture. Inference records results into weights, weights eventually carry the recognition, inference stops firing. The transition from deliberation to reflex is the transition from expensive to free.

### 16. A Goal Has Two Faces — Directive and Setpoint

The goal cell emits two distinct things, and the LLM is prompted for both separately:

- **`<intent>`** — the directive. An imperative verb phrase ("greet the user", "ship v.1"). Drives the **proposer** — Goal collapses to the primitive that closes the gap. This is the action-selection face.
- **`<satisfied-when>`** — the setpoint. A declarative completed-state clause ("the user is greeted and has responded", "v.1 is shipped and the round-trip works"). Drives the **comparator** — completion is measured against this. This is the closure-detection face.

A setpoint is a *state*, never an action. "User is greeted" is the setpoint; "greet the user" is the corrective action emitted to reach it (the thermostat's setpoint is "room at 70°", not "heat the room"). This matches the comparator definition in `cell-types.md`: *the goal cell's content is the expected state, the traversal evidence is the actual state.*

A setpoint is also **prescriptive, not predictive.** It is what the being is committed to making true (correct by acting), not what it expects to happen (correct by learning). The tag is `<satisfied-when>`, not `<predicted-...>` — "predict" stays reserved for the anticipatory engine so the two comparators (goal vs. prediction) never blur. (Active-inference framing: one error quantity, two correction levers — act on the world to match the setpoint, or revise belief to match the world. A goal is a prediction the being has chosen to satisfy by action rather than belief-revision.)

**Observed, not asserted.** The real axis is not internal-vs-external — it is whether the setpoint is a state the being *reads back* or one it *declares*. The setpoint must be confirmable by something outside the goal cell's own output:

- **Exterior-confirmed (the default)** — checked against the world's return perception (the exafference channel). Greet, help, answer, ship. The user replies, the file writes, the build passes. The being cannot fake it.
- **Interior-measured (self-regulation)** — checked against a *measured* slice reading. Token budget, coherence delta. Internal, but observed instrumentally — "the working set fits the budget" is read, not claimed.
- **Forbidden** — checked against the being's own efference (its report that it acted). "I greeted → greeted." A self-satisfying setpoint. This is the case the wording must kill.

The two-sides framing the system prompt uses: `<intent>` is the being's side (what it will do); `<satisfied-when>` is the world's side (what shows up when it worked) — which degrades to a measurement of the being's own state only for self-regulation goals, never to a self-report.

**Why:** The word "goal" — and any imperative phrasing — biases the LLM toward action. If the setpoint is phrased as an action, "did I close the gap?" collapses into "did I perform the act?", which is Action's domain (did motor execute), not Goal's — and the completion signal is read off the being's own act instead of off the world. That is the cosine trap (Decision #8) in another form. Routing the imperative into `<intent>` gives the action-bias a legitimate home and leaves `<satisfied-when>` free to be a pure world-state the comparator can measure closure against. Prompt reliability comes from the declarative `<subject> is/has <past-participle>` format plus the observability test ("if you could mark it done just by acting, it is wrong"), which force the state harder than the label alone.

### Canonical prompt block

```
<intent> — what you are going to do. An imperative: a verb and its object.
           This is the action, your side of it.

<satisfied-when> — how you will know it worked. Not a claim you make — a
           condition you observe. Describe what you would notice that confirms
           the intent succeeded: a change in the world, or a measurable reading
           of your own state. Write it as a state that is now true
           ("<subject> has/is <past-participle>"). Test it: if you could mark it
           done just by acting, it is wrong — name the evidence your action
           produces, not the action.
```

---

## Open

- **Experiment .1001:** Can accumulated history determine goal closure through recursive traversal? Branch created, inference removed from runtime. Next: load fake memories, run the experiment.
- **Emotion binding reality:** designed but not implemented.
- **Metadata convergence as reflex trigger:** the reflex isn't just primitives agreeing — it's when the metadata (the why) converges too. Future improvement.
- **Implementation:** the Brain split is specced but not coded. Six phases in `brain-split-implementation.md`.
