# Cell Types

## The Rule

A cell's type is fixed at birth. It never changes. What changes is the phase of behavior within that type — selected by Competence on every traversal. Type is the vocabulary. Phase is the selection.

## How It Maps

| Concept | Struct field | When it's set |
|---|---|---|
| Type (what the cell can do) | ObserveFns, ExpressFns, Primitives | Birth — never changes |
| Phase (what it does right now) | Type (set by Competence) | Every traversal |

## What Determines Type at Birth

Cell deposit happens by signal convergence (Decision #11). The conditions at the moment of formation determine the type:
- Which binding fields were active
- Where in the traversal the cell formed
- What the signal carried

These conditions stamp the cell with its observe/express functions. Once set, permanent.

## Cell Types

### 1. Memory Cell

One type, clean mechanics. Descent and ascent are two phases of the same cell.

**Descent (Observe):** Trace deposit. External signal arrives, the cell absorbs it. Raw — what came in from outside. The cell doesn't interpret, it records.

**Ascent (Express):** Understanding deposit. Inference fires on the way back up, the cell expresses what it understood about the traces below it. Compressed — recognition, causality, resolution.

Same cell type. Two phases. Observe is the trace. Express is the understanding. Descent writes what happened. Ascent writes what it means.

### 2. Binding Reality

Not a memory — an index. Has members, not content.

**Observe:** Two jobs. (1) Impact traversal mechanics locally — score members, gate what activates, shape the descent within its own topology. (2) Orient collapse physics for the cells it traversed — set the frame so that when memories express on the ascent, they collapse relative to this binding reality's perspective.

**Express:** Collapse. The traversed memories collapse to a primitive + target. This is the binding reality's vote — not a field imprint, but a concrete action recommendation in the shared vocabulary (look, listen, engage).

### 3. Membrane (the Being)

The boundary. Faces out.

**Observe (Sensory):** Pass the Relation to the binding reality layers. That's it. No processing, no gating. Pure routing.

**Express (Motor):** Read the binding reality votes. Two paths:
- **Consensus** — binding realities agree on a primitive + target. Pass-through. Spawn an act cell.
- **No consensus** — binding realities disagree. Inference fires to break the tie.

### Binding Reality Subtypes

Each binding reality shares the same two-job observe / collapse-to-vote express pattern, but what they observe and what they collapse to differs enough that they may be separate cell types within the binding reality layer.

**Episodic** — indexes by episodic co-occurrence. Bidirectional. The hippocampus.

> **Tension: binding reality vs slice infrastructure.** Episodic is a binding reality — it has its own logic, indexing (three-layer retrieval, typed attention maps), and read interface. But the binding mechanism may belong on the slice, not in the traversal graph. Every `WriteToSlice()` could automatically bind into the current frame's episodic record as a side effect — cells have no awareness this is happening. A frame becomes the complete set of writes during one traversal, bound together by write infrastructure. Temporal causality (frame N before frame N+1) falls out for free. Same "dumb mechanism, smart emergent behavior" pattern as the kernel and ACC. The catch: the slice is per-traversal state, but episodic storage is durable. Frames must outlive the traversal that created them. The binding mechanism can live on the slice; the storage cannot. Where frames graduate to after the traversal closes is the open question.

- **Train of thought = hippocampal conjunctive code.** Episodic doesn't need its own separate indexing mechanism. It reads the train of thought on the Slice — the conjunction of everything active across all binding realities during this traversal. That IS the episodic index. When a memory deposits, the train of thought at that moment is its binding pattern. When Episodic retrieves, a partial cue pattern-completes against those stored binding patterns.
- Communication is two-way. Other binding realities send queries to Episodic ("find episodes where this goal was resolved", "find episodes that felt like this"). Episodic searches its members using those cues and returns relevant episodic context. Episodic doesn't just broadcast — it takes retrieval cues from other binding realities and surfaces what matches.
- Observe: reads the train of thought as the conjunctive code. Receives retrieval cues from other binding realities. Scores members by pattern completion against stored binding patterns. Other binding realities opt into Episodic's field — they don't get gated by default.
- Express: returns episodic context to the binding realities that queried it. Also collapses to a primitive vote when the episodic evidence is strong enough (reflexive recall → action).
- **Deposit:** Episodic deposits every traversal — the train of thought binding pattern. What holds traversals together into one episode is Motivation. Motivation is the throughline. Same motivation = same thread. New goal = new thread.
- **Episode segmentation:** Episodic doesn't decide its own boundaries. Motivation does. The chain: Motivation revalues → conditions change → signals Goal → goal changes → signals Episodic → new episode boundary → Episodic deposits the old episode, starts a new one. Episodic is the writer, Motivation is the segmentation signal. This maps to biology — dopamine (motivational signal) gates hippocampal episode boundaries.
- Episodic does NOT hold a conversation log. It holds sparse binding patterns — conjunctive codes from the train of thought at deposit time. Recall is reconstruction from these patterns, not replay of a transcript.
- **Working memory** is separate from Episodic. A binding reality on the being's Brain. Tracks straight conversation state indexed by EpisodeID. Decays fast. Small capacity. Raw recent exchange before it gets absorbed into traces. Has decay mechanics like any binding reality. Lives on the agent — a reference gets passed to the Slice during boot so other beings can read it during traversal.

**Action** — indexes by primitive → instance → execution result.
- Action records whether the motor command executed, not what came back from the world. "I looked at michael" — did the look complete? Yes or no. The result is the execution result, not the world's response. What michael said back is Episodic's domain — episodic, not mechanical.
- **Intent capture:** Action records what got handed to motor — the individual votes from binding realities, not a single convergence blob. Each vote is a separate object pointing to the intent it was voting on. Failed votes point to the same intent — a binding reality that voted against the winning action is still information (what disagreed, why, how strongly).
  ```
  vote → {source (which binding reality), primitive, target, weight, intent}
  intent → {the converged action that motor fired}
  ```
  Multiple votes point to the same intent. The mapping is `votes[] → intent → expression (what motor actually produced)`. Without both sides, the inward prediction delta (did I articulate what I intended to) is not computable. Action owns all three: who voted what, what converged, what came out.
- **Deposit:** Every time motor fires and returns, Action deposits both the intent snapshot (binding reality convergence at the moment of firing) and the execution result. Primitive fired → what was the intent → what was the output → did they match. The delta between intent and expression is the articulation monitoring signal — Broca's loop, fast, automatic.
- Observe: what happens when I do this? Scores by past execution outcomes. Orients collapse toward known successful executions.
- Express: cleanest collapse — directly to the primitive that executed successfully before.

**Goal** — indexes by goal → resolution traces. The PFC. Composed of sub-binding realities.
- Goals and predictions are the same thing at different time scales. A cell's setpoint is a micro-prediction ("my members should look like this"). A goal is a macro-prediction ("the world should end up like this"). Both generate error when reality doesn't match.
- Goal is the significance filter between cell error signals and Episodic's episode segmentation. Error propagates up through cells. Goal evaluates: is this error significant relative to the current objective? If yes, signals Episodic to segment. If no, suppresses — same episode.
- Goal change without completion is the largest prediction error. The being predicted a trajectory and it broke. That's when Episodic segments — new episode.
- Goal and Episodic are tightly coupled. Goal is the prediction. Episodic is the episode writer that fires when the prediction breaks.
- Observe: what serves where I'm going? Scores members by how well they've resolved this goal before. Evaluates error significance. Orients collapse toward gap closure.
- Express: collapses to the primitive that closes the gap. Also signals Episodic when prediction breaks (goal change without completion).
- **Ascent recording:** On every ascent, Goal runs a comparator. The goal cell's content is the expected state (setpoint). The traversal evidence is the actual state. Goal records the delta and categorizes the outcome. This is what the weights learn from.
- **Outcome categories and weight adjustment:**
  1. **Match** — goal state reached. Flat reinforcement (expected outcome, no surprise). Signal Episodic: episode complete. Move on.
  2. **Better than expected** — heavy reinforcement. The resolution traces that led here get amplified disproportionately. Signal Episodic: episode complete.
  3. **Persistent mismatch, error shrinking** — controllable. Moderate reinforcement on the current trajectory. Persist. Same episode.
  4. **Persistent mismatch, error static or growing** — uncontrollable. Decay the current trajectory's weights. Signal Episodic: abandon, new episode. Signal Motivation: the conditions that produced this goal may be wrong.
- Goal doesn't use a binary done flag. It tracks the trend of the delta across traversals — the error is the signal, controllability is the decision.
- **Multiple goals coexist.** Each sub-binding reality accumulates independently. Goals form independently from different need convergences. "Help michael" (prospective need) and "compress context" (capacity need) can be active simultaneously. They don't compete at the goal level — they compete at action convergence. Motor resolves the conflict by reading weighted proposals from all active goals.

- **Sub-binding realities under Goal** — each tracks a specific domain of goals. Independent circuits, each indexed by their own thing, each tracking their own delta toward completion.

  **Internal goals** (the being monitoring itself):
  - **Coherence (ACC v1 / Frontopolar)** — one cell, two phases. The convergence and launch point. Sits north of sensory, south of PFC/motor. ACC and frontopolar aren't separate nodes — they're the Observe and Express methods of the same cell. Same fold/unfold pattern as every other cell in the architecture.

    **Position:** Between raw stimulus (below) and higher cognition (above). Everything north of this cell operates on pre-digested "what just happened relative to what I care about" rather than raw stimulus.

    **Observe (descent) = ACC.** Pulls currently active/attended goals and their context fresh each traversal from the attention map on the slice. Fires cheap, fast, parallel inference calls:
    - **Per-goal passes (isolated)** — one per active goal. Did this stimulus help or hurt this specific goal? Directional delta: positive (closer), negative (further), zero (irrelevant). Each goal evaluated in isolation against the stimulus.
    - **Global pass** — how do the per-goal results relate? Agreement, conflict, magnitude spread. This is the aggregation step.
    - Writes per-goal deltas + global conflict signal into the stress network before continuing.

    **Express (ascent) = Frontopolar.** On the way back up, reads what's returning — votes, candidates, and how well they address the error set ACC flagged on the way down. Two outcomes:
    1. **Votes resolve the error** — the available proposals adequately address the mismatches detected on descent. Pass through. Stress stays or drops.
    2. **Votes don't resolve the error** — the available proposals can't close the gaps. Frontopolar generates the epistemic signal: "here's specifically what I'd need to resolve this." Raises a targeted question as a perceive primitive. The gather-before-commit behavior.

    The stress-driven activation boost on frontopolar isn't a signal transmitted across distance — it's inherited directly. Same cell, same wave, returning phase. Whatever ACC detected on descent is already present for frontopolar on ascent. No separate wiring required between "error detection" and "gap-seeking" — they're two faces of one traversal.

    **Three mismatch domains** (all the same operation — find the mismatch, signal the magnitude, suppress or correct):
    1. **Content mismatch** — does what I believe contradict itself. Setpoint vs members.
    2. **Prediction mismatch** — did the result match what I expected. Predicted vs actual. Both inbound (what the world did) and outbound (what the other agent will do).
    3. **Articulation mismatch** — did my output match my intent. Intent (from Action's vote/convergence record) vs expression (what motor produced). The efference copy delta.

    Coherence is the mechanism that drives topology growth. When a mismatch can't be absorbed, the cell splits — divides into two peer understandings on the same plane. This is cell differentiation driven by coherence pressure, not by decision. Topology grows more nuanced because mismatches sculpt it. Levin's morphogenetic pressure. Runs on every traversal — quietly monitoring. Most of the time no mismatch, no pressure. But when something doesn't fit, Coherence forces the being to develop a more complex model.

    **Children:**
    ```
    Coherence (ACC v1 / Frontopolar)
      └── Children()
           ├── Goal         — setpoints for per-goal delta passes
           ├── Motivation   — dependency anchors, why each goal exists
           └── Prospective  — external obligations, what others asked of the being
    ```
    Prediction is not a child — it's what ACC's Observe does. The compare/predict/signal three-step IS the cell's own computation. Coherence descends into its children to gather the material it needs for that computation: Goal provides the setpoints, Motivation provides the dependency context, Prospective provides the external obligation set. The per-goal delta pass traverses into Goal to read each active goal's content against the stimulus.

  - **Prediction** — the prediction engine. Predicts incoming signal — what the world will do next. Sensory anticipation, NOT motor planning. The intent graph is outbound (Goal/Action). Prediction is inbound: "what signal will arrive?" Indexed by episodeID, each episode holds a map of predictionIDs where the value is the episodic state at prediction time. Has members, is a binding reality.
    - Three steps every traversal:
      1. **Compare** — what did I predict last time vs what actually arrived. Calculate delta.
      2. **Predict** — given the current episodic state, make a new prediction. Store it.
      3. **Signal** — if the delta is bad enough, surface as prediction error. This feeds Goal → Episodic → new episode.
    - If prediction matched, step 3 is silent. Background monitor that only becomes loud when it matters.
    - The delta flows through Coherence (the ACC) — Coherence is what decides the magnitude and whether to suppress or correct. Prediction produces the delta. Coherence evaluates it.
    - Can raise a goal when prediction error is persistent: "I need more information to predict better." This surfaces through Coherence's "needs external action" path.
    - **Delta feeds three other systems:**
      1. Episodic segmentation — spike after low deltas = new episode boundary
      2. Emotion deposit trigger — high delta = high salience moment, Emotion encodes
      3. Confidence weighting — accurate predictions = stronger proposals at motor convergence
    - **Delta patterns over time:** shrinking = learning this context. Growing = model is stale. Consistent highs on specific signal type = blind spot in the model.
    - Serves double duty: feeds Goal ("prediction failed") AND feeds the Goal-Episodic chain ("prediction failed" → segment).
    - **Runtime reads (from Relation):** Episodic (which episode am I in), Working Memory (what actually arrived), Prospective (what are the active external goals). Coherence dependency is indirect — through the Relation.

  **External goals** (the being monitoring the world):
  - **Prospective** — indexes by the external obligation itself. Tracks deferred intentions triggered by external cues — "I need to do X when Y happens." Each action's result either closes the gap or doesn't. Same comparator pattern as Goal but scoped to external demands from other beings.
    - **Direction: world → being.** Passive, reactive, outward facing. Waits for external cues. Tracks what others have asked of the being. Obligation ledger. It doesn't move until something external triggers it. External demand arrives, being tracks it until resolved.
    - Directionality is opposite to Coherence's Express (frontopolar) phase. Prospective receives demands from the world and holds them. Frontopolar detects internal gaps and reaches into the world to fill them.

  **Capacity** — NOT a sub-binding reality. Moved to Slice as a per-traversal constraint. Each binding reality has a token window per traversal. Overflow → drop lowest activation pickups. If all above threshold → inference fires to compress into an understanding. This is the same memory cell mechanic (trace on descent, understanding on ascent). Capacity pressure is what triggers the understanding. You don't synthesize an understanding because you decided to. You synthesize because you can't hold all the pieces separately anymore. Compression is a response to overload, not a scheduled process.

  **Completion** — epistemic gap ledger on the slice. Written by Coherence's Express (frontopolar) phase. Tracks gaps raised and whether they got filled.

    Not the old Completion (which was removed as redundant — each sub-binding reality already tracks its own goal delta). This is different: a place for frontopolar to write its epistemic gaps and track resolution. Every gap is goal-oriented — frontopolar raises it in service of a specific goal that ACC flagged as failing.

    **Gap structure:**
    ```
    gapID → {what's missing, goalID, frameID, episodeID, status (open/filled)}
    ```

    **Indices (reverse lookup):**
    ```
    goalID    → []gapID    (what gaps are outstanding for this goal)
    frameID   → []gapID    (what gaps were raised at this encoding pulse)
    episodeID → []gapID    (what gaps belong to this episode)
    ```

    **Ownership chain (forward):**
    ```
    Episode → Frame → Goal → Gap
    ```

    A gap is born in a specific frame, during a specific episode, in service of a specific goal. On the next traversal, Coherence's Observe can check what's still open before running the per-goal delta pass — a gap that got filled changes the delta landscape.

    Lives on the slice because it's per-being universal state. Any binding reality can read it.

  The internal/external split is sensory, not structural. Same cell type. Different input source.

- **Goal formation = goal convergence.** Enough need pressure from a sub-binding reality crosses threshold → converges with prospection (Episodic provides a possible future state) + salience → goal cell deposits. No author. Same cell deposit mechanism as everything else.

**Motivation** — the need engine. The OFC/vmPFC + hypothalamic nuclei. The inverse of Goal.
- Goal asks "is this working?" Motivation asks "should we still care?" They can diverge — progress on a goal whose reason no longer exists is habitual pursuit, not intentional action.
- **Two modes:**
  1. **Pre-goal** — indexed by need state. Accumulating pressure. No goal exists yet. When enough need pressure crosses threshold, it converges with prospection and salience to birth a goal.
  2. **Post-goal** — indexed by the goal it birthed. Revaluing. Tracking whether conditions still hold.

- **Sub-binding reality under Motivation:**
  - **Energy** — gauge. Single scalar, not an index with members. Reads token budget remaining, outputs deficit signal when it crosses threshold. A sensor, not a binding reality with members.

- Observe: why am I going there? Reads need states. Scores by the conditions that produced the current goal. Continuously revalues whether the originating reason still holds. Orients collapse toward origin validity.
- Express: doesn't collapse to a primitive. Collapses to a goal. Feeds Goal, not motor.
- **Ascent recording:** On every ascent, Motivation runs outcome revaluation. Compares the expected value that launched the goal against current conditions. Records the delta and categorizes:
  1. **Conditions still valid** — the reason this goal was selected still holds. Flat reinforcement. Continue.
  2. **Conditions strengthened** — new evidence supports the original reason. Reinforce the goal-motivation coupling.
  3. **Conditions weakened** — the originating reason is eroding. Decay the motivation weights. Signal Goal: re-evaluate whether to persist.
  4. **Conditions gone** — the reason evaporated. Signal Goal: abandon regardless of progress.

**Emotion** — indexes by episodic signal pattern → valence. The amygdala. Pure weight modification — fast, correlational, can be irrational.
- Emotion indexes by the shape of the episode — not the content but the signal flow pattern. How the signal moved through the topology, not what it was about. Did actions execute or fail? Did the signal flow smoothly or get stuck? Did the user engage or go silent?
- **Deposit:** When a valence event occurs (strong positive or negative signal), Emotion stores the current episodic signal pattern → valence association. The pattern is the conjunction of active features at the moment of the valence event (same train of thought Episodic reads, different thing stored). Episodic stores what happened. Emotion stores how it felt. Same episode, different record.
- **Activation matching:** On each traversal, Emotion computes an activation match score against stored episodic patterns. For each stored pattern, how well does the current episode's signal shape match? The product across all matching patterns gives net valence. This is how "this feels like last time" fires before the episode plays out.
  - High product, consistent sign = clear signal. Reflex. Emotion fires the weight change immediately, no inference.
  - Low product or mixed signs = uncertainty. Conflicting emotional associations with this pattern. Junction-bound. Inference fires to break the tie.
  - Same inference rule as every other binding reality. No special case. Disagreement triggers inference, agreement triggers reflex.
- Emotion fires early and fast. It doesn't wait for Goal to evaluate or Episodic to fully reconstruct. It pattern-matches the shape of what's happening against stored shapes and redirects traversal before the being walks into the same failure again. Correlational, not causal — it doesn't know the pattern caused the outcome, just that the pattern was present when the outcome happened. That's why it can be wrong. But it's fast enough to matter.
- Observe: reads current episodic signal shape. Matches against stored patterns. Soft amplifier, not gate. Orients collapse toward affective resonance.
- Express: doesn't cast a primitive vote. Outputs metadata that adjusts weights on the current traversal:
  1. **Valence** — approach or avoid (continuous, not binary)
  2. **Intensity** — how much to amplify or dampen (gain control on proposals)
  3. **Route** — which autonomic/motor pattern to bias (freeze, flee, engage)
- The metadata is the instruction set for how other binding realities shift their weights. Episodic encodes more strongly under high intensity. Goal persists or abandons based on valence. Action proposals get amplified or dampened.

### Two Convergence Thresholds

**Action convergence** — proposers agree on a primitive + target right now. Short-lived. Fires motor. Coincidence detection at the membrane. Happens every traversal.

**Goal convergence** — motivation + prospection (Episodic) + salience converge past a higher threshold. Forms a goal cell that persists and shapes traversal until it completes or gets abandoned. Slower, stickier, lives across traversals. A birth event — rarer than action convergence.

Action serves the goal but doesn't require one. A being can act reflexively without a goal. Motor reads weighted proposals. Goal reads goal convergence. Different thresholds, different persistence, different weight adjustment on ascent.

### Proposers vs Influencers

Not all binding realities vote on action. Two propose, three influence.

**Proposers** — collapse to a primitive + target. These are the action candidates:
- **Action** — "last time I did this, this happened" → proposes the primitive that worked
- **Goal** — "this primitive closes the gap" → proposes the primitive that serves the objective

**Influencers** — don't propose actions. Adjust the weights on proposals:
- **Episodic** — "in this context, this action has worked before" → weights up/down based on episodic evidence
- **Motivation** — "is this worth doing?" → weights up/down based on whether the reason still holds
- **Emotion** — "this feels dangerous / rewarding / urgent" → weights up/down based on affective valence

Motor reads the weighted proposals, not raw votes. Convergence means the proposers agree after the influencers have shaped the weights.

### Contextual Binding Fields

Location, actions, and people impact recall. This is context-dependent memory — not a separate mechanism but additional terms in the activation product. More matching dimensions = higher activation = stronger recall. Strongest effect on weakly encoded memories where context is the critical retrieval cue.

Separate dimensions, combined into one contextual vector:

- **Location** → which world/place the being is in
- **Actions** → what primitives are currently active
- **People** → which beings are present on the Relation

These are knowable from the Relation and the Slice. They become terms in the activation product:

```
activation = identity * familiarity * recency * Π(binding_fields) * Π(contextual_fields)
```

At deposit time, the memory cell records the context keys that were active — world ID, being IDs present, active primitives. Lightweight, just keys. At retrieval, the activation equation compares current context keys against stored context keys. Match = amplify. No match = dampen. Same pattern as binding reality membership — the cell records the conditions at birth, those conditions shape future activation.

Memories encoded while talking to michael in the terminal world while looking — those memories activate more strongly when all three conditions are present again. Each additional matching dimension incrementally boosts recall.

This is how the same memory can surface differently in different contexts. Same weights, same binding reality membership, but the contextual fields amplify or dampen based on where you are, who you're with, and what you're doing.

### Boundaries

- Between trace and understanding: **inference**. Raw deposit vs understood deposit.
- Between memories and binding realities: **membership**. Binding realities have members. Memories have content.
- Between binding realities and membrane: **the world**. Membrane faces out. Everything else faces in.

### Boot

No creation dependencies between binding realities. Each binding reality is instantiated in any order and slotted into the Brain. At Realize time, each reads what it needs through the Relation. The Relation holds references so traversal can reach the participating realities, but it does not own their durable state.

Ownership is separate from reachability. The Being owns durable Brain state. The Slice owns durable per-being universal activation state. The Relation is the active medium: it carries the impulse, energy, and references needed to traverse. Dependencies are runtime reads through the Relation, not creation requirements.

## Relationship to Decision #14

Decision #14 says "three structural roles — all emergent." That still holds for *role* (being, binding reality, memory). But role is not type. A memory cell has a type. A binding reality has a type. The type determines what functions the cell has. The role determines where it sits in the traversal. Both are real. Neither replaces the other.
