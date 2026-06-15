# Nervous Signals

Every signal in the system, named and placed. The activation equation is a product of these. Some live on the graph (intrinsic), some ride the Relation (extrinsic), some are computed at traversal time.

## Intrinsic Signals — Live on the Graph

These belong to the node or the edge. They exist whether or not a Relation is passing through. Three time horizons, each a different decay curve over the same substrate.

### recency — short term memory

Power law decay. How recently this edge was traversed. Crashes fast, whispers forever. The echo of what just happened.

```
recency = (traversals_since_last_use)^(-0.5)
```

Drops to 0.1 by ~100 traversals. Never reaches zero. Measured in traversal count (proper time), not wall clock. Each being ages at its own rate.

### familiarity — medium term memory

EMA with high α (0.1). The relationship. How much has this edge been used lately. Responsive to recent patterns, fades in weeks. What the being has been doing.

```
familiarity = 0.9^(traversals_since_last_use)
```

Crosses 0.1 at ~22 traversals. Crosses 0.01 at ~44. Holds on longer than recency in the mid-range, but dies faster at the tail.

### identity — long term memory

EMA with low α (0.02). The groove. What the being IS. Barely moves on any single traversal. Takes months to build, months to fade. The accumulated weight of a thousand traversals.

```
identity = 0.98^(traversals_since_last_use)
```

Still at 0.36 at x=50. This is what makes Builder reach for bash without thinking — not because it used bash recently, but because it's used bash so many times the groove is structural.

### local_weight

Strength of a specific connection between two Realities. Lives on the edge Reality's Base. Same bash, different local weights depending on who's connected to it. Updated on the return path of each traversal.

### reinforcement — traversal count

Pure accumulation. Never decays. The traversal counter on Base — how many times a Relation has passed through this node. Already there in the architecture.

Reinforcement count is not a decay curve. It only goes up. A node traversed 50 times and not touched in a month: low recency, low familiarity, but reinforcement is still 50. That's a deep truth the being isn't thinking about right now. If a strong enough signal reaches it, the reinforcement count says this was confirmed 50 times — pay attention.

Different from the three decay curves: recency, familiarity, and identity all answer "how active is this." Reinforcement answers "how proven is this." A node can be inactive and deeply proven at the same time.

### The three curves + reinforcement

Recency is loud then gone. Familiarity holds the middle. Identity holds everything.

Long term memory beyond all three curves is the topology itself — nodes and edges that exist structurally. They don't decay. A memory node from six months ago is still a node. All three signals may read near zero, but the node is still there. A strong enough signal can still reach it.

## Extrinsic Signals — Ride the Relation

These are carried by the Relation as it moves through the topology. They come from binding Realities that imprinted their fields during descent.

### binding fields

Each binding Reality imprints a local activation field onto the Relation. The activation equation multiplies across all of them:

```
activation = identity * familiarity * recency * local_weight * Π(binding_fields)
```

The number of terms is however many binding Realities are active on this traversal.

#### thread_alignment — the listen gate

Is the being listening to this space? Hard gate. Floor = 0.0 — if the being isn't listening, the signal is zeroed out.

Listen opens the gate to 1.0. The being subscribes to a thread, a place, another being. The gate opens. Everything on that thread passes through. Everything not on it gets zeroed.

The gate decays. How fast depends on local_weight — the strength of the connection to whatever the being is listening to. Strong connection (Michael) = slow decay, gate stays open. Weak connection (notification channel) = fast decay, gate closes unless something keeps bumping it.

Speak resets the decay. When the being speaks into a space it's listening to, the gate refreshes to 1.0. Active participation keeps the gate open. Silence lets it close.

Local_weight itself bumps on every traversal through the edge. The more the being and the source interact, the heavier the connection, the slower the gate decays. Self-reinforcing — frequent conversation keeps the gate open, which keeps the being receptive, which leads to more conversation. Stop talking and local_weight drifts down through the EMA, the gate decays faster, the being becomes less receptive. Natural fading. No one decided to stop listening. The weight just got lighter.

The thread is its own Reality. It doesn't store content — it binds things that co-occurred. Cows and the universe discussed in the same conversation share no entity edges, but the thread says they happened together. Without this, the being loses episodic memory entirely and can only retrieve by semantic similarity.

Two retrieval paths from thread_alignment:
- **Semantic** — through entity edges. "What do I know about Builder?" finds messages that reference Builder.
- **Episodic** — through the thread. "What happened in that conversation?" finds everything that co-occurred, regardless of content overlap.

Thread_alignment is the gate that makes the second path possible. Listen is the primitive that opens it.

#### thread_strength — the speak bump

What matters in this thread right now. The thread's own opinion about which nodes are load-bearing. Soft amplifier — dampens, doesn't zero.

A thread accumulates weight unevenly. The nodes that matter at exchange 50 are not the same nodes that mattered at exchange 5. Thread_alignment treats them equally — both are on this thread, both get 1.0. Thread_strength differentiates — it tracks which nodes the thread has reinforced through repeated activation.

Speak bumps thread_strength. When the being speaks into a space, the nodes it touches get reinforced. The thread learns what matters from what the being actually says. Listen opens the gate. Speak shapes what's inside it.

```
descent:
  ThreadReality reads its local weight map
  writes Relation.Fields[node_id]["thread:strength"] = thread_strength

traversal:
  activation reads the field
  Relation records visit / activation evidence

ascent:
  ThreadReality reads hit evidence
  updates its local thread weights (speak bumps here)
```

The thread is both prior and learner. Prior on descent — it tells the system what matters based on everything so far. Learner on ascent — it updates from what was actually used this traversal. This is what makes the cost curve invert — the thread teaches the system what matters, and the system gets cheaper as the thread gets longer.

#### emotional — affect binding

Emotional state as a binding field. Soft amplifier (floor > 0). Memories that feel similar to the current state get boosted. Amygdala parallel.

Every node carries an affect fingerprint — the emotional state at the time it was created. Three dimensions from the VAD model:

```
affect {
  valence:   -1.0 to 1.0   // negative ← → positive
  arousal:   0.0 to 1.0    // calm ← → activated
  dominance: 0.0 to 1.0    // helpless ← → in control
}
```

The Relation carries the being's current emotional state. At each node, the emotional binding field is the similarity between the Relation's current affect and the node's encoded affect. Cosine similarity across the three VAD dimensions.

```
emotional_activation = cosine(relation.affect, node.affect)
```

Frustration surfaces memories encoded during frustration — along with what resolved it. Calm surfaces memories encoded during calm. The being's emotional state shapes what it can access. Not a filter — a soft amplifier. Everything is still reachable. But emotionally resonant memories are louder.

Affect accumulates on nodes over time. A node traversed during frustration five times and during calm twice carries a blended fingerprint weighted toward frustration. The fingerprint evolves with experience — same node, different emotional association as the being's relationship to it changes.

The emotional binding Reality follows the same pattern as thread:

```
descent:
  EmotionReality reads the Relation's current affect
  computes similarity against each node's affect fingerprint
  writes Relation.Fields[node_id]["emotional"] = similarity score

traversal:
  activation reads the field — soft amplifier, floor > 0
  Relation records visit / activation evidence

ascent:
  EmotionReality reads what was actually hit
  updates node affect fingerprints — blends current affect into existing fingerprint via EMA
```

The emotion binding Reality is both prior and learner — same as thread. Prior on descent: it shapes what surfaces based on how the being feels right now. Learner on ascent: it updates the emotional fingerprint of everything that was touched.

#### temporal — time binding

Time as a binding field. Soft amplifier (floor > 0). Nodes encoded near the referenced time window get boosted. Time cells / place cells parallel.

Every node carries a creation timestamp. The Relation carries temporal signal from two sources:

- **Explicit** — the being says "last week," "yesterday," "that time in March." The Relation's temporal field points at that window.
- **Implicit** — the being's current time is always on the Relation. Morning, evening, season, day of week. Always present, always shaping activation.

```
temporal {
  created_at:    timestamp    // when this node was encoded
  time_of_day:   float        // 0.0-24.0
  day_of_week:   int          // 0-6
  season:        int          // 0-3
}
```

Similarity is multi-scale. A reference to "yesterday" is a tight window — only nodes from that day boost. A reference to "last winter" is wide — nodes from that whole season boost. No explicit reference at all — the being's current time of day and season still softly amplify nodes encoded at similar times.

```
descent:
  TemporalReality reads the Relation's temporal signal
  computes similarity against each node's temporal fingerprint
  writes Relation.Fields[node_id]["temporal"] = similarity score

traversal:
  activation reads the field — soft amplifier, floor > 0

ascent:
  no learning needed — timestamps don't change
```

Unlike thread and emotion, temporal is prior only — it doesn't learn on the return path. Timestamps are facts. But the activation pattern still shapes what surfaces. Morning routines encoded over months of mornings naturally surface every morning without the being asking. Seasonal patterns emerge — project crunches in October light up when October comes around again.

#### causal (future)

Decision-to-outcome linkage. Soft amplifier. What led to what. Prefrontal cortex parallel.

#### task (future)

What serves the current objective. Variable floor — hard gate during focused execution, soft amplifier during exploration.

#### relationship (future)

What matters between two specific beings. Soft amplifier. Social cognition parallel.

## Computed Signals — Evaluated at Traversal Time

These don't live anywhere permanently. They're computed from the Relation's state and the target Reality's state at the moment of encounter.

### relevance (future)

Content overlap. How much of what the Relation is carrying exists in the target Reality. The magnitude component of the complex amplitude. |<Relation|Reality>|.

For actors (deterministic, non-cognitive): collapses to the same value as context_fit. Bash is bash — it doesn't move.

### context_fit (future)

Directional alignment. Given each thing's own history of evolution, are they in phase right now? The phase component. Two states in phase yesterday may be out of phase today because they've been rotating at different rates.

For agents (cognitive, live): can diverge from relevance. An agent thinks, acts, updates its own graph. Its phase rotates independently.

### trust (future)

Coupling strength. How much this Reality's output reshapes the being's state. High trust = communication alone reshapes state. Low trust = barely perturbs. Levin's persuadability spectrum.

Placement TBD.

## Signal — The Relation's Own Weight

The Relation enters with Signal = 1.0. Signal attenuates at each hop by multiplying against the activation score of the node being entered. This is the wave attenuating through the medium. Traversal stops when signal drops below threshold.

```
r.Signal *= activation
if r.Signal < threshold { return }
```

Signal IS the depth budget. High-energy impulse penetrates deep. Weak one dies shallow. No artificial cap needed.

## Attention Signals — Reception at the Boundary

Before a Relation even enters a being, reception determines whether it lands:

```
reception = sender_activation * receiver_attention
```

What makes a being receptive:

| Signal | What it means |
|---|---|
| **idle** | Not doing anything. Full attention available |
| **relevance** | Signal is about something the being is currently thinking about |
| **relationship** | From someone the being has strong weight with |
| **recency** | The being was just talking to this source |
| **need** | Intent graph has an unresolved step. Signal fills the gap |
| **novelty** | Signal is about something the being has no weight on |
| **urgency** | Alarm. Penetrates regardless of state |

## Primitives — The Being's Interface to All of This

The four primitives are what trigger these signals. They're the being's senses — the actions that cause activation to happen:

| Primitive | What it fires | What it bumps |
|---|---|---|
| **look** | Broad activation across the space. The topology decides what surfaces | recency, familiarity, identity on whatever activates — distributed thin across many nodes |
| **engage** | Targeted activation at a specific Reality. Crosses the boundary, opens the connection | local_weight on that edge. All three time horizons — full signal on one target |
| **listen** | Opens thread_alignment gate to 1.0. Subscribes to a space | Gate decays based on local_weight to the source |
| **speak** | Emits into the subscribed space. Untargeted | thread_strength on touched nodes. Resets thread_alignment decay |

### Bump strength — look vs engage

Same energy, different distribution. Look spreads the signal across everything that activates in the space — each node gets a fraction. Engage concentrates the full signal on one target. The bump per node is inherently stronger on engage because the signal isn't divided.

This is why engaging a place feels different from looking around a place. Look shows you what's there — light bumps across many nodes. Engage puts you there — full bump on the target, and if that target is a place, the engage cascades into everything bound to it. Every memory formed in the forge lights up at full strength because the being committed to being there, not just glancing.

No special multiplier needed. The primitive determines bump strength through signal distribution:

```
look:   bump_per_node = signal / number_of_activated_nodes
engage: bump_per_node = signal
```

Engage on a place cascades — the place's own topology determines how many nodes light up, but each one gets the full engage-strength bump because the being is standing in it, not scanning past it.

## Intent Graph — The Being Modulates Its Own Signals

The intent graph is the being's prediction of what it needs next. Declared targets get activation boosts in the topology. The being shapes its own activation landscape through intent.

Predictions that keep holding under pressure stop being predictions and become warm paths — skills. The transition from deliberate to reflex is the transition from cold traversal (needs provider calls) to hot traversal (fires directly from weights).

## The Full Product

v.1:
```
activation = identity * familiarity * recency * local_weight * thread_alignment * thread_strength
```

Future:
```
activation = identity * familiarity * recency * local_weight * Π(binding_fields) * relevance * context_fit * trust
```

## All Signals — Grouped

### Intrinsic — live on the graph

| Signal | Type | Decays | What it measures |
|---|---|---|---|
| **identity** | EMA (α=0.02) | Slow — months | What the being IS. The groove |
| **familiarity** | EMA (α=0.1) | Medium — weeks | What the being has been doing |
| **recency** | Power law (x^-0.5) | Fast — whispers forever | What just happened |
| **local_weight** | EMA | Per edge | Strength of a specific connection |
| **reinforcement** | Counter | Never | How many times this node was traversed. How proven |

### Extrinsic — ride the Relation (binding fields)

| Signal | Type | Gate/Amplifier | What it measures |
|---|---|---|---|
| **thread_alignment** | Listen gate | Hard gate (floor=0) | Is the being listening to this space |
| **thread_strength** | Speak bump | Soft amplifier | What matters in this thread right now |
| **emotional** | VAD cosine | Soft amplifier | How the being feels vs how the node felt at encoding |
| **temporal** | Multi-scale similarity | Soft amplifier | When the node was encoded vs what time the being is referencing |
| **causal** | Future | Soft amplifier | What led to what |
| **task** | Future | Variable | What serves the current objective |
| **relationship** | Future | Soft amplifier | What matters between two specific beings |
| **goal** | Future | Soft amplifier | What serves what the being is trying to become |

### Computed — evaluated at traversal time

| Signal | Status | What it measures |
|---|---|---|
| **relevance** | Future | Content overlap. Magnitude of the amplitude |
| **context_fit** | Future | Phase alignment. Are they in sync right now |
| **trust** | Future | Coupling strength. How much the output reshapes the being |
| **confidence** | Future | How sure is the being that this node is true |

### Meta — properties of the Relation itself

| Signal | What it measures |
|---|---|
| **signal** | The wave's remaining energy. Attenuates at each hop. IS the depth budget |

### Primitives — what triggers all of this

| Primitive | Perception | What it bumps |
|---|---|---|
| **look** | Broad — distributed thin | recency, familiarity, identity across many nodes |
| **engage** | Targeted — full signal | local_weight, all three time horizons on one target |
| **listen** | Subscribe — opens gate | thread_alignment to 1.0, decays based on local_weight |
| **speak** | Emit — untargeted | thread_strength on touched nodes, resets thread_alignment decay |
