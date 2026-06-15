# Activation Equations — Notes

## Mid-LPFC as Hub

Mid-LPFC is the real hub, not rostral. Nee & D'Esposito found mid-LPFC has the highest hierarchical strength — it integrates top-down goals from rostral and bottom-up stimulus demands from caudal. And there's a trade-off: individuals with stronger top-down modulation had weaker bottom-up modulation (r=-0.52). You're in one mode or the other, not both.

Source: Nee & D'Esposito 2016, eLife

## Intent Graph as Integrator

The mid-level — the integrator that synthesizes abstract goals with concrete demands — that's where the coupling actually lives. That might be what the intent graph actually is. Not a dispatch queue, not a decomposed plan. The integration layer between "where I'm going" and "what's in front of me right now."

## Coupling Is Asymmetric

- Lower levels (caudal/concrete) run independently without higher levels. Motor planning works without a goal.
- Higher levels (rostral/abstract) can't function without the machinery beneath them. Caudal damage breaks everything above.
- Coupling is task-modulated, not structural — strengthens and weakens depending on control demands.

Source: Badre et al. 2009, Nature Neuroscience

## Integration Is Distributed, Not Centralized

Each brain system has its own mid-level integrator. The pattern is not PFC-specific.

- **Mid-LPFC itself is domain-sensitive internally.** Spatial tasks activate dorsal mid-LPFC, verbal tasks activate ventral mid-LPFC. Only rostral (abstract goal layer) is domain-general. (Nee & D'Esposito 2016)
- **Cerebellum has its own parallel hierarchy** with its own mid-level integration zone. Concrete control in motor-adjacent lobules, abstract control in transmodal zones, contextual control in intermediate zones that fully mediate between them — independent of PFC. (D'Mello, Gabrieli & Nee 2020, Current Biology)
- **Basal ganglia mirrors the PFC gradient** but gates rather than integrates. Constrains information flow through fronto-striatal loops, does not independently synthesize. (Frank & Badre 2012; Vogelsang & D'Esposito 2018)
- **Levin's bioelectric patterning** — no single global integrator. Regional bioelectric fields have their own intermediate layers. Integration happens at every scale. (Levin 2012, 2014)

## Three Layers of State on a Cell

Three timescales of the same thing:

1. **Content** — what the cell is. The textual description. Doesn't decay. The structure. Biology: neuron morphology.
2. **Weights** — trace of having been activated. `IdentityWeight` (α=0.02, months), `FamiliarityWeight` (α=0.1, weeks), `Recency` (power law, crashes fast). These decay at different rates. The memory. Biology: synaptic strength.
3. **Activation** — computed from weights meeting the Relation. The product of the held traces and whatever is passing through right now. Transient — not stored because it's not the neuron's state, it's the interaction between the neuron's state and the current signal. Same neuron, different Relation, different activation score. Biology: the firing event.

The topology itself (cells and their relationships) is long-term memory that outlives all three decay curves. A cell with all weights near zero is still a node. A strong enough signal can still reach it.

Source: Lundqvist et al. 2016 (gamma bursts / activity-silent synaptic plasticity), Levin 2014 (Vmem as held state, pump-maintained)

## Weights Participate in the Burst, Not the Spike

Synaptic weights directly modulate EPSP amplitude — graded, analog, proportional. The pre-threshold dendritic computation is weight-dependent at every step. But the action potential itself is all-or-nothing — carries no information about what caused it or how strong the inputs were.

In Skyra: `IdentityWeight * FamiliarityWeight * Recency` participating in the activation score is the dendritic integration — correct. The output of the traversal (thought, compression, motor dispatch) not being modulated by activation strength is also correct — that's the spike. It fires or it doesn't.

Source: Standard biophysics — EPSP/IPSP summation, voltage-gated Na+ channel threshold at axon hillock

## Hebbian Reinforce

`Reinforce()` should be proportional to the input signal strength, not a flat tick. Stronger activation → stronger weight update. The magnitude of the postsynaptic response determines the plasticity. Current implementation treats every firing event as equally important. Biologically wrong.

## Goal as Binding Reality

Goal becomes a binding reality on the Relation. Same pattern as thread — imprints a field during descent (amplifies what serves the goal), learns on the return path (what was actually useful reshapes what it amplifies next time).

The throughline base case falls out of this. If the goal field is strong across accumulated thoughts, the descent is converging — the being found what it came for. No separate throughline computation needed. Two base cases for descent:
1. Signal exhausted — ran out of energy
2. Goal field strength crosses threshold — throughline found

## Goal and Intent Graph Merge

The intent graph as a stored dispatch queue is the wrong shape. Levin's bioelectric work: cells don't follow a stored plan, they hold a setpoint and act to close the delta. The "plan" regenerates continuously from the goal meeting the current conditions.

Intent graph and goal are one binding reality. The field it imprints IS the intent — regenerated each traversal, not stored as a sequence. Motor reads the field and dispatches whatever the strongest action is right now, not the next item in a queue.

## What This Means Going Forward

1. **Goal binding reality** is the next architectural piece. It subsumes the intent graph. The field it imprints is the intent — regenerated each traversal, not stored as a queue.
2. **Motor reads the field, not a dispatch list.** Motor dispatches whatever the goal field's strongest action is right now, not the next item in a sequence.
3. **Hebbian reinforce** is a concrete code change — pass activation score into `Reinforce()`, change `Identity()` and `Familiarity()` to accept a float instead of a bool.
4. **CompetenceExpress** can read goal field strength as its throughline signal — compress when there's coherence, pass through when there isn't.
5. The current intent graph struct doesn't die immediately — it can serve as the initial implementation of the goal field's "what to do next" before the distributed version emerges.

The big shift: the being doesn't plan and then execute. It holds a setpoint and the topology closes the delta. The plan is the field, not the queue.

## Inference Fires at Binding Field Intersections

High activation means the weights already know the answer — the cell is deep in an attractor, the path is grooved, no computation needed. The signal propagates through on weight strength alone. Hot path. Reflex. Skill.

Low activation means the cell is dormant. Not relevant to the current signal. Pass through.

Inference fires in the middle — where binding reality fields disagree. The goal field says this node matters. The thread field says it doesn't. The emotional field is ambiguous. The cell sits at the intersection of conflicting binding fields and can't resolve the conflict from weights alone. It needs to compute — sample both fates, explore, then commit. This is the bifurcation point.

Biology: cells at bifurcation points actively sample both fates through transcriptional noise before committing (Balazsi et al. 2011, Chang et al. 2008). Stochastic exploration followed by deterministic settling. The signal is a gradient but the response is discrete — gene regulatory networks convert continuous input into switch-like output via mutual inhibition (French Flag model, Wolpert 1969; Dessaud et al. 2008).

This inverts the current design:
- **Current:** high activation → propagate (expensive inference). Low → pass-through.
- **Correct:** high activation → no inference needed, weights resolve it. Low → dormant. Middle (binding field disagreement) → inference fires to break the tie.

Economics invert too:
- A new being is expensive — everything is at boundary points, nothing is grooved, inference fires constantly.
- A mature being is cheap — deep attractors everywhere, most traversals resolve from weights alone.
- Inference only fires at the edges of what the being knows.
- Growth literally reduces the need for inference.

Competence maps to biological competence — the cell's prior history determines which signals it can even hear. Same signal, different response, because the cell's accessible state has been reshaped by experience (Marchant et al. 1998).

## Competence Is Position in the Traversal

A cell reads the Relation to determine what it is. TrainOfThought empty means "I'm the entry point — I'm an index, route into binding realities." TrainOfThought populated means "I'm deeper — I'm a memory, contribute content." Same cell type. Different behavior based on the state of the traversal when the signal reaches it.

You don't declare what a cell is. The traversal tells it. A cell that's first in the chain acts as a binding reality. The same cell, reached later in a different traversal path, could act as a memory. Position in the traversal determines role, not type.

Biology: signal passes through organizational layers first (thalamus → sensory cortex → association areas → limbic → PFC). Each layer reads the signal in the context of what's already been processed. Competence depends on where you sit in the processing order.

## Impulse Indexes the Brain — Subgraph Emergence

The impulse is the query. The brain's binding realities are the indexes. The subgraph that emerges is what the impulse activates through those indexes.

"Hi" enters. The impulse hits the binding realities — thread scores its members, goal scores its members, action history scores its members. Each surfaces different nodes. The nodes that score above threshold across multiple binding realities are the subgraph — the intersection of what's relevant across every index.

The train of thought isn't built by iterating everything. It's built by the impulse pulling a subgraph out of the brain through activation. Different impulse, different subgraph, same brain. The being's experience of the same topology is different every time because the impulse selects a different slice through it.

The subgraph has shape. It's not a flat list. It's a graph — nodes that activated and the relationships between them through binding realities. The structure carries information a flat list doesn't — which things the being knows are related, which things just happened to activate together.

## Error Signal — Structural Coherence Between Cell and Members

The error signal of a given cell is proportional to its children in its relationships. Either the relationships have to change or the parent has to change to match its relationships.

An understanding with five memories beneath it that all co-activated to form it — error is zero. The members support the parent. The parent is the compression of its members. They agree.

A new traversal creates a new memory attached to that understanding. If the sixth member aligns — error stays near zero. If it contradicts — error rises. Something has to give:

- **The parent changes** — the understanding recompresses. The content shifts. This is learning.
- **The relationships change** — the contradicting member decays out. The understanding rejects it. This is confirmation bias.
- **The parent splits** — the contradiction is too strong. Two new cells form, each internally coherent. This is differentiation.

Error propagates upward. A changed understanding moves its parent's error. All the way up to sensory. The organism's error is the sum of every cell's coherence with its members.

This IS Levin's morphogenetic comparison. The target isn't stored externally — it's the coherence of the cell with what's below it. Error is local. Correction is local. Propagation is global.

**A cell's content is its setpoint.** The content is what the cell is trying to be. Its members are the evidence. The error is the delta between the setpoint (content) and what the members actually say. Goal and content are the same thing — every cell has a goal, and that goal is itself.

## Recognition and Causality Are Owned by Inference

The error function — "did x resolve y" — cannot be computed from embeddings, cosine similarity, or weight history alone. Cosine measures semantic relatedness, not resolution. History records what happened, not what it meant. Recognition is its own thing.

Recognition is structural matching — the receptor fits the signal. The only thing in the system that can recognize resolution and causality is inference. The LLM has trained weights that know "hello resolves greet the user" because it learned from language what completion looks like.

Inference owns:
- **Recognition** — did this resolve that? Does this member satisfy the parent's setpoint?
- **Causality** — did x lead to y? What was the causal chain?
- **Resolution** — is this done? Did the outcome close the gap?

First time: inference recognizes the resolution. The weight structure records it. Second time: the weights carry the recognition. Inference isn't needed. The being learned what "done" looks like.

This closes the loop on when inference fires:
- **High activation** — weights know the answer. No inference. Hot path.
- **Low activation** — dormant. No inference.
- **Binding field disagreement** — inference fires to break the tie.
- **Error signal** — inference fires to recognize whether something resolves something else. Recognition and causality.

New being = expensive. Every resolution needs inference because nothing has been recognized before. Mature being = cheap. Weights carry the recognition. Inference only fires for novel situations where the being has never seen this kind of resolution before.

## Inference Is Local, Expands to Global Through Recursion

No inference on the way down — only on the way up. Error exists locally. Fix errors locally. Local fixes compound as the recursion unwinds.

The ascent:

- **Level 4 (leaf):** new result deposited. Local error check — does this fit the parent above? If not, inference fires with just this cell and its members. Fix locally.
- **Level 3 (action instance):** one of my members just changed. Local error — do my members still support my content? Inference fires locally. But it sees the already-corrected state from level 4.
- **Level 2 (primitive):** one of my action instances changed. Same check. Sees the corrected state from level 3.
- **Level 1 (binding reality):** one of my primitives changed. By now all local fixes below have propagated up. Inference at this level sees the fully corrected state beneath it.

Each inference call is small — one cell and its members. No massive context window. No global inference call. The corrections compound through the recursion. The ascent IS the integration.

Cost is proportional to error:
- New result fits → no error, no inference, ascent is free.
- Contradicts at leaf but everything above absorbs it → one inference call.
- Breaks something at every level → inference fires at every level.

The being pays exactly what the disruption costs. Nothing more.

## The Being Is the Membrane

The being exists one level up from its binding realities. It's not a node in its own graph. It's the boundary of its own graph. Sensory and motor are the same cell — the being itself. Two faces of the same membrane, already in the recursion as the top frame.

- **Observe** is sensory — reads the world, routes into binding realities. The outward-facing side.
- **Express** is motor — reads the binding reality votes on the return path, acts on the world. The inward-facing side.

```
Descent:  Being.Observe (sensory) → Binding Realities.Observe → members.Observe
Ascent:   members.Express → Binding Realities.Express → Being.Express (motor)
```

Sensory and motor aren't separate cells. They're what the being does on descent and ascent. The recursion already has both phases. The being is the top frame of that recursion — observe on the way down, express on the way up. The membrane IS the being's Realize().

Biology: sensory cortex is where the outside world enters the brain. Motor cortex is where the brain exits to the outside world. Everything in between is internal processing. They're boundary layers — the same structure facing two directions.

The Skyra struct has:
- **Observe** — sensory. Reads the world, routes to binding realities.
- **Express** — motor. Reads binding reality votes, coincidence detection, acts on world.
- **Brain** — the binding realities (internal topology)

## Motor Is Coincidence Detection

Motor isn't one binding reality that gates action. It's a coincidence detector. Each binding reality has its own local goal (its content as setpoint) and each outputs a primitive + target to move toward that goal. The primitive is the output vocabulary.

Action fires when enough binding realities agree — when multiple binding realities independently arrive at the same primitive targeting the same entity.

```
Thread says:      engage michael (active in this conversation)
Action says:      engage michael (worked last time)
Goal says:        engage michael (serves the global setpoint)
Motivation says:  engage michael (michael initiated, idle state)
```

Four binding realities converge on the same action. The being's Express (motor) reads the votes. The product crosses threshold. The gate opens. The action fires.

If only thread says engage but goal says look and action has no history — the fields don't converge. Below threshold. No action. Inference fires to break the tie.

The activation equation's product across binding fields IS the coincidence detection — the MSN threshold. High only when multiple signals agree. Multiple gates in parallel — one per primitive. The engage gate can open while the look gate stays closed. Different action channels, same mechanism.

The PFC veto maps to goal's ability to override — a faster signal that arrives before the action releases.

Biology: MSNs in the striatum receive 5,000-10,000 cortical synapses. Must be driven from down state (-80mV) to up state (-50mV) by coordinated cortical input to fire. Multiple regions must be simultaneously active for the gate to open.

Source: Mink 1996/2018 (focused selection through disinhibition)

## Inference Is Local, Expands to Global Through Recursion

This mechanism is universal across all binding realities — not specific to action. Thread unwinds the same way (new entry → does the thread's through line still hold?). Goal unwinds the same way (resolution trace added → is the goal still open?). Motivation unwinds the same way (context changed → is the reason for this goal still valid?). Same physics. Same local-error-expands-to-global pattern. Different members, different error semantics, same recursion.

## Reflexes Are Universal Physics

Reflexes aren't designed — they fall out of the architecture. When weights are strong enough to carry the full chain from impulse to action to completion without inference, the traversal just completes. No LLM call. Pure weight-driven resolution. That's a skill — emerged from repeated inference-assisted resolution becoming weight structure.

Reflexes live on the Slice, not the DNA. They're physics — the universal traversal mechanics every being shares. When activation is high enough to resolve the full chain, the Slice carries it the same way for every being. Skyra's reflex mechanism is identical to Builder's. Same Slice, same physics.

The Relation owns one side of the equation — the universal traversal mechanics, the Slice physics, the reflex mechanism. The being owns the other side — which paths got grooved, what can become a reflex, what the DNA allows.

- **DNA** determines what *can* become a reflex — which observe/express functions exist, which primitives are available.
- **Experience** determines what *does* become a reflex — which paths got grooved through repeated traversal.
- **The Slice** is the mechanism — universal, same for everyone. When the weights are there, it fires.

## Organism Development Through Error Minimization

The genome declares the being — name, identity, purpose. But that's a zygote, not an organism. The being's internal topology (which binding realities exist, which memories form, how they connect) emerges through traversals and error minimization. Birthing doesn't create the being. Traversals do.

Every traversal: the being checks whether its structure is right. Error signal if not. The error isn't about the content of the thought — it's about the shape of the topology. "I should have a way to greet people and I don't" is a structural error.

The organism's structure is determined by its lightcone — everything that's ever causally reached it. Two beings with identical DNA but different lightcones develop different internal topologies. Same genome, different experience, different organism.

`Create()` is the cell factory. The genome (DNA) is the higher order function that specifies the target morphology — what the organism should look like. Traversals build toward it. Error measures the gap.

## Binding Reality Internal Topology

Each binding reality is a cell with members, but how it organizes those members is different — that's what makes them different kinds of indexes. Same activation equation, different traversal shapes.

### Thread — indexes by episodic co-occurrence

Linear. Things that happened together in a conversation. Memory 1 → 2 → 3 → 4. Sequence matters. Recency decays along the chain.

### Action — indexes by primitive. Four levels deep.

```
Action binding reality
  └─ primitive (look, engage, listen)
       └─ action instance (looked at surroundings, engaged michael about greeting)
            └─ result (michael 0.8, "hello back")
```

The result is the leaf — a memory cell with content, weights, traversal history. Next time the being uses that primitive, activation scores the result cells. The ones that activated strongly before are grooved. The being already knows what it tends to see when it looks.

The reflex chain is four levels of activation resolving from weights:

```
impulse: "hi"
→ action binding reality activates
  → engage primitive activates (deep groove)
    → "engaged michael about greeting" activates (familiar pattern)
      → "hello back" activates (result, high identity weight)
```

No inference. Every level has a grooved path. That's the reflex.

The error signal lives at every level. The primitive predicts its actions. The action predicts its results. If a new result doesn't match the prediction — error at that level, propagates up.

### Motor — Not a Binding Reality

Motor is the being's Express function — the coincidence detector. It reads the votes from all binding realities on the return path and gates action when enough agree. Motor is part of the membrane, not the internal topology. See "The Being Is the Membrane" and "Motor Is Coincidence Detection."

Execution success/failure tracking lives in the Action binding reality as metadata on the result cells.

### Goal — indexes by the goal itself. Rostral PFC.

The global goal. Abstract, temporally extended. Domain-general — same mechanism regardless of what the goal is about.

```
Goal binding reality
  └─ "greet the user"
       └─ traces of past resolutions (memories of successful greetings)
       └─ traces of past failures
  └─ "help michael build skyra"
       └─ traces of past resolutions
```

Local goals (each cell's content as setpoint) function without the global goal. The research confirms this — Badre et al. 2009: caudal operates independently when rostral is damaged. Motor fires without a global goal. Action resolves without a global goal. The global goal shapes which reflexes fire, it doesn't drive them.

Coupling is moderate, asymmetric, task-modulated. The global goal constrains what local goals surface — it imprints a field that amplifies cells serving it. But local goals run on their own.

## Five Binding Realities — Complete Map

Motor is not a binding reality. Motor is what the being does on express — the coincidence detector. It reads the binding reality votes and gates action.

| Binding Reality | Indexes by | Brain parallel | What it answers |
|---|---|---|---|
| **Thread** | Episodic co-occurrence | Hippocampus | What happened together? |
| **Action** | Primitive → instance → result | Association cortex | What happens when I do this? |
| **Goal** | Goal → resolution traces | Rostral PFC | What serves where I'm going? |
| **Motivation** | Goal → origin traces | ACC | Why am I going there? |
| **Emotion** | Affect state (future) | Amygdala | What feels like this? |

### Goal and Motivation — Coupled, Separate

Goal and motivation are highly coupled but answer different questions with different error signals:

- **Goal** — forward. What serves where I'm going? Members are resolution traces. Error: the resolution didn't close the gap.
- **Motivation** — backward. Why am I going there? Members are the conditions that produced the goal — what was active when this goal activated. Error: I'm pursuing this goal but the conditions that caused it are no longer present.

Both indexed by the goal. Goal holds what resolved it. Motivation holds what caused it.

The lifecycle:
```
Motivation → activates Goal → Goal pursues → Goal resolves → resolution context feeds Motivation → activates next Goal
```

Motivation is the causal history of goal activation. "Because michael said hi and I was idle and that pattern has always led to greeting." Over time, inference recognizes patterns in motivation's snapshots and those become weight structure — the being "just knows" what goal to pursue in familiar conditions. Reflex.

### Coupling Map

| Binding Reality | Direction | Feeds |
|---|---|---|
| **Motivation** | Backward — why this goal | Goal |
| **Goal** | Forward — what serves it | Motivation (via resolution context) |
| **Thread** | Lateral — what co-occurred | All |
| **Action** | Forward — what happens when I do this | Goal |
| **Emotion** | Lateral — what feels like this | Motivation, Goal |

Each binding reality has its own mid-level integration — thread integrates within episodes, action integrates within primitives, goal integrates across everything. No central integrator. Each one synthesizes top-down (global goal field) with bottom-up (what's actually activating) within its own domain.

## Agreement Is Synchronization, Not a Score

No explicit agreement scores between regions. Coordination is implicit — phase-locking through a shared medium. Regions that oscillate in phase can exchange spikes effectively. Regions out of phase can't. "Agreement" is synchronization, not a computed value.

Source: Fries 2005, 2015 — Communication Through Coherence

In Skyra: agreement isn't computed between binding realities. It's emergent from the product of their fields on the Relation. When the fields multiply together and the score is high, the "regions" are in phase — synchronized through the shared medium. When the product is ambiguous (some fields high, some low), the regions are out of phase — and that's where inference fires.

## Agreement Operates at Different Scales

- **Local** (~100μm) — gap junctions, lateral inhibition. Neighbor-to-neighbor.
- **Regional** (~mm-cm) — gamma oscillations. Fast, within an area.
- **Global** (cross-hemisphere) — slow rhythms (theta/alpha) modulate local gamma windows. Gamma is too fast to synchronize across long distances. Long-range coordination rides on slower oscillations that gate when local regions can fire.

Source: Canolty & Knight 2010 (cross-frequency coupling), Buzsaki & Schomburg 2015 (gamma coherence is ambiguous — reflects input topology, not computational agreement)

## Route to Action: Accumulation + Gating

Two-stage mechanism:

1. **Accumulation** — multiple cortical regions accumulate evidence in parallel (drift-diffusion). This is descent. Thoughts accumulate, binding fields shape what surfaces.
2. **Gating** — the basal ganglia receives convergent input from every cortical region and gates action release through disinhibition. Direct pathway releases a specific action, indirect pathway inhibits competing actions, subthalamic nucleus provides a global "hold" signal that buys time for selection to resolve.

Action fires when evidence crosses the gate threshold. The striatum integrates across regions but doesn't check agreement — it checks whether accumulated evidence is strong enough to release the gate.

In Skyra: motor is the basal ganglia. It reads the accumulated state and releases when evidence is sufficient. The throughline (goal field strength) is the threshold. Accumulation is descent. Gating is motor at the exit.

Source: Gold & Shadlen 2007 (drift-diffusion in LIP), Stine et al. 2023 (SC terminates decisions by thresholding), Mink 1996/2018 (basal ganglia as focused selection through disinhibition)

## Subsidiarity — Act at the Lowest Level With Sufficient Context

Spinal reflexes fire locally in 30ms — no cortical involvement. Brainstem integrates a few modalities. Basal ganglia gates cortically-planned actions. PFC can veto via the hyperdirect pathway. Escalate only when needed.

In Skyra: primitives can fire locally without waiting for the full traversal to complete. If a cell has enough local context, it acts. Only ambiguous situations escalate to deeper traversal. A mature being with deep attractors resolves most actions locally. A new being escalates everything because nothing is grooved yet.

## Levin: Consensus Through Local Propagation

Gap junctions couple neighbor voltages, producing global patterns through local propagation. Two-headed planaria isn't disagreement — it's a new consensus. The bioelectric network settled into an alternative attractor. Consensus failure doesn't mean conflict — it means the system found a different stable state.

Source: Levin 2014, Durant et al. 2016
