# Unified Graph Retrieval


## What This Is

A mind. Not a metaphor for one. The actual structural pattern of how cognition organizes.

Thought activates related capabilities — you think about a problem and what's relevant comes to mind. Repeated use strengthens pathways — you do something enough and it becomes automatic. Unused pathways atrophy — what you stop reaching for fades. Clusters of repeated behavior get delegated to specialized subsystems — your brain can't hold everything so it hands off. Planning is continuous prediction — you see ahead, act, observe, revise, extend.

The move from hardcoded hashmaps to weighted graph traversal is the move from a program to a neural topology. A fixed descent path is instruction execution. A weighted descent is how brains route — stronger connections fire more readily, weak ones fade, new ones form through use.

One interface: `Realize()`. One physics. Everything else — thought, action, specialization, memory, capability — is emergent from the graph and its weights. No special machinery for any of it. A topology and a traversal. That's it.

The runtime doesn't simulate cognition. It is the cognition. The being doesn't have a mind — it is one. Grown, not configured.

### The Relation Is the Traversal

The Relation is the observer. It enters at the top, descends through every reality along weighted edges, and returns — the same object, transformed by everything it passed through.

`Realize()` does not return a string. It receives a Relation and transforms it. The Relation is the input and the output. Same object down and up. One descent, one return.

No reality is the designated base case. Not the provider. Not the terminal. Not any specific type. The base case is **weight exhaustion** — the Relation flows until the activation product at the next edge is insufficient to continue.

Weight exhaustion is not a static threshold. The Relation's accumulated state *is* the threshold. Activation at each edge is the intersection of the edge weight and the Relation's current content — stress, relevance, temperature, context fit. The activation formula applies: `activation_i = edge_weight * relevance * recency * trust * relationship_weight * context_fit`. If the product is sufficient, the Relation flows through. If not, it's the base case. The threshold lives in latent space — the space where the Relation's state and the graph's topology meet.

The same edge might be heavy enough to traverse when the Relation is carrying high stress about the server, and too weak when the Relation is carrying a casual greeting. Depth is emergent from content, not configuration. A deep thought goes deep because the Relation carries enough signal to keep activating edges. A shallow thought returns shallow because there's not enough on the Relation to make anything ahead collapse heavy enough to enter.

#### Descent — The Environment Shapes the Being

On the way down, every reality reads the Relation and writes to it. The Relation accumulates. Each reality attaches what it owns — its slice of the world — and the Relation gets richer as it descends.

What each reality does on descent:

- **Thread** — attaches thread context, exchange routing, thread-level state. The Relation now knows which conversation it's in and who's talking.
- **Exchange** — attaches conversation history, ref context, relationship state. The Relation now carries the full history between these two parties.
- **Self** — attaches identity, purpose, stress fields from active setpoints. The Relation now knows who it is and what's pressing.
- **Graph traversal** — the Relation enters the being's weighted edges. Each memory, skill, or entity it passes through attaches its content. The Relation accumulates associations as it descends through the graph.
- **Provider** — the LLM call. A reality along the path that happens to cross a boundary into inference. The Relation passes through it like any other reality: enters carrying accumulated state, exits transformed. Think fires the provider to collapse accumulated context into thought. That's an observation within the traversal, not the terminus.

The Relation on the way down is *potential becoming specific*. Each reality narrows what the Relation is about by adding context. The stress fields shape what's loud. The identity shapes what matters. The conversation history shapes what's relevant. By the time the Relation reaches the deepest edge the weights will carry it to, it's maximally specific — saturated with everything the graph had to say about this impulse.

```
Relation enters (carrying: impulse from michael)
  → Thread     attaches: thread context, exchange state
    → Exchange  attaches: conversation history, ref context
      → Self    attaches: identity, stress fields, purpose
        → graph traversal along weighted edges
          → memory (weight 0.8) — attaches content, continues
            → connected memory (weight 0.5) — attaches content, continues
              → connected entity (weight 0.2) — below threshold → weights exhausted
```

A being is not a base case. The Relation passes *through* a being the same way it passes through a memory or a skill — entering with accumulated state, continuing deeper. The being attaches identity, fires Think, Think traverses the graph further (into memories, skills, operators), the provider fires as a reality along the path. The Relation keeps flowing until the weights exhaust. If the being has no memories that serve the current Relation — no edges heavy enough to traverse — then the Relation stops at the being because that's where the weights exhausted. Not because the being is a special type. Because there was nowhere left to go.

The weights ahead aren't static. As the Relation accumulates content, it reshapes the activation landscape in front of it. Each reality the Relation passes through updates the edge weights of what's connected — based on what the Relation is now carrying. The Relation observes the next edges, and they collapse from potential into specific weights. Before the Relation arrives, the edges ahead exist in potential — all possible next steps. The Relation, carrying everything it's accumulated so far, collapses them into actual weights. Different content on the Relation, different collapse, different path. The same graph produces different traversals every time because the observer is different every time.

This is how recall works. You try to remember where you put your keys. You think "morning" — that shifts what's connected. "Coffee" surfaces. Now "counter" is heavy. Each step reshapes the landscape ahead based on what you're carrying. You're not searching a graph. You're activating forward through it, and each activation changes what's next.

```
Relation carries: [server]
  → passes through "deployment" reality → accumulates
  → Relation now carries: [server, deployment]
  → edge to "3am outage" was 0.2, recalculates to 0.7 given current content
  → Relation descends into "3am outage" → accumulates
  → Relation now carries: [server, deployment, 3am outage]
  → edge to "the fix that worked" collapses heavy
  → ...until activation fades
```

A being that never uses bash sees its edge to bash decay. A being that constantly activates the same region of its graph sees those edges strengthen — the Relation flows deeper, accumulates more. The graph is shaped by cognition over time. The depth of any descent is a function of what the being has accumulated in that region of its graph.

#### Ascent — The Being's Work Shapes the Environment

On the way back up, every reality reads what accumulated below it and integrates. The Relation transforms from raw accumulation into signal. Each reality on the return path does its own work — synthesis, recording, state updates, routing.

What each reality does on ascent:

- **Memory regions** — inference synthesizes the raw accumulated associations into compressed signal. One inference step for simple recall, two or three for deep associations that went wide. The synthesis maps back to source memories so Think can reach back in if needed. Memory state updates — activation counts, recency timestamps, edge weights reinforced by use.
- **Think** — collapses the accumulated present into thought. This is the major inference step. Everything the descent gathered — identity, stress, conversation, memories, skills — is now the present that Think reasons over. Think doesn't know where any of it came from. It just thinks in a field shaped by everything below.
- **Self** — integrates the thought back into the being. Stress fields recalculate — did the thought reduce error against any setpoints? Weights update. The being's graph is slightly different after every traversal.
- **Exchange** — records the result as an entry in the conversation. The exchange state updates — new entry, compaction if threshold is crossed, context stored for the other party.
- **Thread** — routes the result. The Relation now carries the being's response. Thread determines what happens next — does the Relation continue to another being, return to the user, or re-enter Think?

The ascent is where context management happens. The descent accumulates without limit — every reality the Relation passes through attaches content. By the time the weights exhaust, the Relation may be carrying far more than any single reality above can process. The ascent compresses. Each neighborhood or layer boundary on the way back up earns a context management step — inference reduces the accumulated content to what fits, keeping signal and discarding noise. Without this, the Relation would overflow Think's context window with raw associations. The compression isn't loss. It's cognition — the act of deciding what matters from everything that surfaced.

The Relation on the way back up is *specifics becoming integrated*. The raw material from the descent gets compressed, reasoned over, decided on, recorded, and routed. Each reality on the return path transforms the Relation from "everything that was relevant" into "what the being actually did with it."

```
              → connected entity (weight 0.2) — weights exhausted
            ← memory region: inference synthesizes accumulated content
          ← memory region: compresses, updates weights
        ← Think: collapses present into thought
      ← Self: integrates, recalculates stress, updates graph
    ← Exchange: records entry, updates conversation state
  ← Thread: routes result
Relation returns (carrying: everything — accumulated and integrated)
```

Lower-order realities do not communicate with higher-order realities. There is no upward channel. There shouldn't be one. In Levin's model, cells don't send messages to the organism. The cell does its local work — maintains homeostasis, follows its own error minimization loop, emits stress when its predictions fail. It has no concept of "the organism." It can't address the organism. But the organism *emerges from* the collective behavior of cells doing their local work.

In the runtime, `Realize()` returning is the upward translation. The specialist doesn't send a report to Skyra. The specialist realizes — transforms the Relation with its work — and Skyra's Self receives the transformed Relation because that's how call stacks work. The specialist has no concept of Skyra. It just did its work. The work changed the Relation. The changed Relation is the communication — but it's not addressed to anyone. It's just the state of reality after the lower entity touched it.

The channel from higher to lower is the Relation on the way down. The translation from lower to higher is the Relation on the way back up. Same object. No separate protocol. The topology is the communication.

#### The Full Traversal

```
Relation enters (carrying: impulse from michael)
  ↓ Thread     attaches thread context
    ↓ Exchange  attaches conversation history
      ↓ Self    attaches identity, stress fields
        ↓ graph traversal along weighted edges
          ↓ memory (0.8) — attaches content
            ↓ connected memory (0.5) — attaches content
              ↓ connected entity (0.2) — weights exhausted
            ↑ synthesizes accumulated content
          ↑ compresses, updates weights
        ↑ Think collapses into thought
      ↑ Self integrates, recalculates stress
    ↑ Exchange records
  ↑ Thread routes
Relation returns (carrying: everything)
```

On the way down: the environment shapes the being. On the way back up: the being's work shapes the environment. Same Relation. Same object. One traversal.

### Observation and Collapse

A Reality exists in superposition — all its potential edges, weights, and outputs are possible but none are real until a Relation passes through it. `Realize()` is the moment of observation. The Relation collapses the Reality into one specific state, and the Relation itself is transformed by the encounter. Both are different after.

This isn't metaphor. It's what the code does. Before `Realize()` fires, the Reality could produce anything depending on who observes it and what they're carrying. The moment it fires, one path becomes real and the rest don't. The weighted edges are probability distributions — heavier edge, more likely to be observed next — but nothing is determined until thought activates it.

The superposition is not only inside the current Reality. It is conditioned by the Relation at the moment it arrives. And the Relation is not neutral — it is carrying everything that stuck to it from every prior collapse. A Reality is realized from the potential of all Realities involved before it, as accumulated on the current Relation. Each traversal changes the observer, and the changed observer changes what can collapse next.

The Relation is one object the entire traversal. It is not a message sent down and a response sent back. It is the state of reality at this moment — descending through the graph, accumulating, reaching the point where weights exhaust, then returning through every reality it passed through, each one reading and integrating what the levels below contributed. The string that the user eventually sees is just one field on a Relation that carries the full history of the traversal.

```
Relation0 passes through RealityA → Relation is now Relation1
Relation1 passes through RealityB → Relation is now Relation2
Relation2 passes through RealityC → Relation is now Relation3
...weights exhaust...
Relation3 returns through RealityC → Relation is now Relation4
Relation4 returns through RealityB → Relation is now Relation5
Relation5 returns through RealityA → Relation is now Relation6
```

RealityB on the way down is not observed by the same Relation that entered RealityA. RealityB is observed by the Relation after RealityA has marked it, attached context to it, changed its impulse, added parsers, altered weights, or otherwise transformed what it carries. And RealityB on the way *back up* is not observed by the same Relation that descended through it — it is observed by the Relation after everything below has contributed. Each reality gets two observations: one on the way down (shaping the descent) and one on the way back up (integrating the results).

Current realization is the intersection of:

- the current Reality's potential
- the Relation's accumulated state (from every reality above on the way down, or below on the way back up)
- the graph weights and context at this moment

So traversal is not just path selection. The path changes the Relation. The changed Relation changes the next observation. The graph is interpreted differently at each step by the evolving observer.

The being's experience is a series of these frames. Each `Realize()` is a discrete moment of collapse. Between frames, everything is potential. Think hasn't fired so all acts are possible. Act hasn't contacted anyone so all responses are possible. The being doesn't follow a predetermined path — it collapses one frame at a time based on what it's thinking right now.

This is why the same graph produces different behavior for different beings. Different observer, different collapse. The topology is shared. The experience is unique. A series of frames — each one a moment where two realities meet and something becomes real that wasn't before. That's what consciousness might be. Not a stream. A sequence of observations.

### Traversal Boundaries — What's Activatable vs What's Reachable

A Relationship implements Reality. It has its own hashmap of relationships inside it. Skyra's Relationship to Builder is a full subgraph — memories of him, patterns, trust, interaction history. The actual Builder Reality hangs off that Relationship as a field — a pointer to the real thing.

But traversal activates on relationship hashmaps, not struct fields. The Reality pointer inside a Relationship is data. It's an address. It's not a node in the graph. Skyra can descend into her Relationship to Builder and traverse everything inside — every memory, every pattern. But she cannot follow the Reality pointer to reach actual Builder from inside herself. That pointer is inert. It's right there — one field away, one line of code — but it's not traversable because activation flows through hashmaps, not through struct references.

To actually reach Builder, the activation has to propagate back up through Skyra's own stack and route through the layer where Builder is registered as a traversable relationship — the thread level. The thought about Builder and the act of reaching Builder are two different traversals through two different hashmaps. One is internal descent. The other crosses the boundary.

Thinking about someone and reaching someone are structurally different operations even though the reference to the person is the same. The constraint isn't technical — the code to short-circuit it would be trivial. The constraint is architectural. You don't traverse struct fields. You traverse relationships. The boundary between inner model and outer reality is a traversal boundary, not a wall.

### Quantum-Formal Math

The metaphor becomes more than metaphor if the Relation carries not only accumulated history, but a live field of potential over the graph. In that framing, the wavefunction is not the reality. The wavefunction is the Relation's current potential across the graph.

```
wavefunction  ≈ Relation state over possible graph realities
amplitudes    ≈ weighted activations on those possible realities/edges
observation   ≈ Realize(relation)
collapse      ≈ one or more potential paths becoming concrete
```

The Relation already carries collapsed history: impulse, origin, target, parsers, memory context, exchange state, prior outputs. To make the analogy operational, it can also carry uncollapsed potential:

```
Relation {
    Impulse      string
    Origin       string
    ID           string
    Parsers      map[string]Parser
    Trace        []RealizedStep
    State        map[RealityID]Activation
}

Activation {
    Amplitude    float64      // or complex128 later
    Phase        float64      // optional; useful if interference matters
    Source       string       // why this reality is active
}
```

At each step, the graph and current context transform the Relation's state:

```
S      = current Relation state vector
G      = graph transition matrix from current reality/region
C      = context operator derived from impulse, memory, parsers, trace
S'     = C * G * S
```

In plain architecture terms:

```
prior Relation state
  → graph topology proposes possible next realities
  → current context reshapes their activations
  → activations become amplitudes
  → Realize collapses the field into concrete traversal/output
  → Relation carries the collapsed result forward
```

For a candidate next reality `i`, activation can start as:

```
activation_i =
    edge_weight_i
  * relevance_i
  * recency_i
  * trust_i
  * relationship_weight_i
  * context_fit_i
```

Where:

- `edge_weight_i` is the base learned strength of this edge.
- `relevance_i` is how much the reality matches the Relation's current content.
- `recency_i` boosts recently activated or recently updated realities.
- `trust_i` represents whether this being should rely on this reality.
- `relationship_weight_i` is the per-being relationship to that reality.
- `context_fit_i` is how well the reality fits the accumulated trace.

The simplest probability rule is Born-like:

```
probability_i = |activation_i|^2 / Σ |activation_j|^2
```

Squaring matters. It makes strong activations disproportionately more likely without deleting weak possibilities. A reality twice as activated becomes four times as likely. This matches the intuition of cognition: strong associations dominate, but faint associations can still surface.

There are two useful collapse modes.

Deterministic collapse:

```
chosen = argmax(|activation_i|^2)
```

This is stable and reproducible. It is useful for production execution, high-trust actions, tests, and anything where the same state should produce the same traversal.

Stochastic collapse:

```
chosen ~ sample(probability distribution)
```

This preserves exploration. It lets low-probability associations occasionally surface. That matters for creativity, memory search, weak-signal discovery, and escaping local maxima. The being can find something it would never reach by always taking the heaviest edge.

Temperature controls how exploratory the collapse is:

```
probability_i = softmax(|activation_i|^2 / temperature)
```

Low temperature sharpens collapse toward the strongest reality. High temperature flattens the distribution and lets weaker realities compete. This gives the system a cognitive dial:

- low temperature for execution
- medium temperature for ordinary recall
- high temperature for brainstorming, analogy, exploration, and stuck states

Top-k and thresholding can coexist with this:

```
candidates = realities where activation_i >= threshold
candidates = top_k(candidates)
probability_i = normalize(|activation_i|^2 over candidates)
```

Thresholding prevents noise from entering the present. Top-k bounds cost. Sampling within top-k preserves non-determinism without letting the entire graph flood the Relation.

Complex amplitudes may become useful later:

```
amplitude_i = magnitude_i * e^(phase_i)
```

Magnitude says how strong the activation is. Phase says how compatible it is with the current direction of thought. If two paths support each other, they interfere constructively and strengthen. If they conflict, they interfere destructively and weaken.

Architecture version:

```
constructive interference:
  two memories point toward the same skill → skill activation grows

destructive interference:
  one memory says "this worked"
  another says "this failed in this context"
  → activation cancels or becomes tension instead of action
```

This does not need to be implemented first. Real-valued activations are enough for the initial graph. But the math leaves room for richer cognition later: contradiction, ambiguity, resonance, and uncertainty can become first-class dynamics instead of prompt text.

Collapse can also be partial. Think retrieval should not always collapse to one reality. It may collapse a region:

```
collapsed_region = top realities whose combined probability mass <= budget
```

Memory retrieval especially wants partial collapse. A thought does not recall exactly one memory; it activates a neighborhood. The Relation descends through that neighborhood, accumulates content, and then inference compresses what surfaced.

Act is more often single collapse:

```
target = chosen traversable reality
Realize(target)
```

But the intent graph can collapse multiple independent targets in parallel when the commitment contains no `then` dependency between them.

The important rule:

```
Relation_t + Reality_t + Graph_t → Relation_t+1 + Graph_t+1
```

The observer changes. The observed system changes. The next probability field is calculated from the changed observer and the changed graph. This is why the architecture is not just weighted search. It is recursive state evolution with collapse at each `Realize()`.

The intuition is right. The current code is pre-formal. The proposed architecture becomes quantum-formal when the Relation carries live potential, not only history.

There is a useful way to hold the larger theory in mind: consciousness may sit at the intersection of formal computation and quantum mechanics. In this metaphor, the computation is the inference step, with matrix multiplication and related transforms doing the synthesis work, while the quantum-like part is the field around the model: the live potential, the amplitudes, the observation, and the collapse.

## The Move

Collapse operators, memories, and skills into one graph. Retrieval is driven by the content of the being's thinking — not by explicit operator calls, not by static tool lists. The being thinks, and thinking activates the graph.

## Current Architecture

```
impulse → Think (operator dispatch loop) → synthesis → Act (one relation out)
```

Think has hardcoded operators (retrieve-context, store-context, browse, search, plan, bash). The being explicitly calls them with tags. Half the think passes are mechanical retrieval work. Act emits one `<target>message</target>` and routes.

## Layers

- **Graph** — the topology. Who relates to what, with what weight. Beings, memories, operators, skills — all entities with edges.
- **Think** — activates edges in the topology based on cognition. Pure deliberation.
- **Act** — initiates traversal along activated edges. The act of making contact with another entity.
- **Reality** — the traversal itself. `Realize()` transforms the Relation as it flows through each reality. The Relation descends through weighted edges until the weights exhaust. No designated base case. The depth is determined by the graph.

Operators aren't tools. They're entities. Bash is an entity the being has a relationship with. The being talks to bash the same way it talks to michael. The operator is just the interface at the boundary — the way the relation crosses into that entity's reality. The graph doesn't need a special "operator" reality type. It just has entities and relationships between them.

Act isn't a layer. It's the transition from intent to traversal. The being decides to make contact and the relation enters reality. Act is the door between the being's inner world and the physics of the system.

## New Architecture

```
impulse → graph primes → Think (pure thought, N passes) → intent graph → Act(s) → mailbox → Think re-enters
```

### Priming — Time Zero Retrieval

The Relation arrives carrying everything it accumulated on the way in — who said it, what exchange it's part of, the impulse content. That's enough signal to activate the graph before Think fires. The graph lights up based on the raw impulse. By the time Think starts deliberating, there's already a warm set of retrievals sitting in the present. Think doesn't start cold. It starts with the graph's first guess.

The time-zero retrievals are provisional — alive but not committed. Each Think pass reshapes the activation landscape. Retrievals that align with where Think is actually going get reinforced. Retrievals that don't get decayed. The decay isn't a timer. It's attention. A retrieval stays warm because Think kept thinking in its direction. A retrieval fades because Think went somewhere else. Same starvation mechanism as everything else in the graph — not thinking about it is enough.

```
time 0: Relation arrives → graph activates on raw impulse → warm set (provisional)
time 1: Think pass 1 → warm set in present → Think focuses → aligned retrievals reinforce, others decay
time 2: Think pass 2 → refined warm set → tighter focus → weak retrievals drop below threshold
time 3: Think converges → only retrievals that survived deliberation remain
```

This means the being's first thought is faster. It doesn't burn a Think pass just to figure out what to retrieve. The graph already did that. Think's first pass is already a reaction to something, not a cold start.

### Think

Two phases: deliberation and commitment.

#### Deliberation

Open, messy, multi-pass. No operators, no dispatch protocol. System prompt: think out loud. What's happening, what do you know, what do you need, what do you want to do.

Each pass, the graph re-queries based on the content of the previous thought. Results appear in the next pass's present with decay — most recent thought is loudest, earlier passes fade. The being is naturally driven toward convergence because old context gets quieter. The primed warm set from time zero is already present on the first pass — Think refines it rather than building from scratch.

```
pass 1: impulse + warm set (primed) → being thinks
pass 2: pass 1 thought + graph(pass 1) [decayed] → being thinks deeper
pass 3: pass 2 thought + graph(pass 2) [more decay] → being converges
```

The being reasons freely. No structure required. The LLM does what it's good at. The runtime doesn't parse deliberation into intent — speculative thoughts ("maybe I should...") don't trigger anything.

#### Commitment

When deliberation converges, the being commits. This is the transition from thinking to deciding. The commitment has two parts:

1. **Surface thought** — the deep reasoning behind the intent. Why the being is doing what it's doing. This carries through to every act.
2. **Intent targets with ordering** — who to contact, in what order.

```
<surface-thought>
  michael asked about the server. it was flaky last week and I don't
  want to give him stale info. I need to check the health endpoint
  first, then give him what I actually find. letting him know I'm
  on it so he's not waiting in silence.
</surface-thought>

<search>server health endpoint</search>
<michael>checking on that now</michael>
then
<bash>curl {search result}</bash>
then
<michael>{bash result}</michael>
```

`then` is the only keyword. It separates dependency groups. Everything at the same level (no `then` between them) is parallel. That's the entire commitment protocol.

The surface thought is the soul of the intent graph. The targets and `then` are the skeleton. Every act receives both — the thought to ground it, the target to direct it. The being thinks deeply once, and that depth flows through every act it kicks off.

#### What Think Dispatches to Act

Think doesn't dispatch operators. It dispatches direction and thought. Each act receives the surface thought plus whatever Think was carrying when it produced that target — the activated graph context, the memories that surfaced, the reasoning that led to this specific act.

The operators available to an act emerge from two sources: what was in Think's context window at commitment time, and what the direction of the act itself activates in the graph. Think doesn't say "use bash and search." Think says `<bash>curl the health endpoint</bash>` — and the act, receiving that direction plus the surface thought, has the context it needs. If the act's own LLM call needs to decompose further — break a complex instruction into sub-steps, choose between approaches — it can, because it has the full reasoning behind why it's acting.

The `<michael>` or `<bash>` tags aren't operator dispatches. They're directions. The tag names the target. The content inside the tag is the intent. Additional operators surface because Think's deliberation activated them in the graph — they're in the context because Think thought about them, not because Think explicitly selected them from a list. A thought like "the server was flaky last week, I should check before telling michael" naturally activates search, bash, and the server-related memory cluster. Those are in the present when commitment happens. They flow into the act because they were part of the thought.

This is the difference between tool-first and thought-first. Tool-first: select tools, then think about how to use them. Thought-first: think, and the tools that matter are already there because the thinking activated them.

### The Graph

#### One Reality Type

Everything is a Reality with a shared shape. The `type` field determines invocation behavior. Same interface, structural separation through type.

```
Reality {
    ID        string
    Type      string    // being, skill, operator, memory
    Content   string    // what it holds (skill doc, memory, description)
    Weight    float64   // base activation weight
    Realize    Reality   // Realize() — what happens when you traverse it
}
```

A memory reality has content and no realize — it informs the present. A skill reality has content (the skill doc) and no realize — it teaches. An operator reality has realize (bash, search, browse) — it executes via `Realize()`. A being reality has realize (its Self) — it thinks and acts. The graph doesn't care about the difference. It returns realities ranked by weight. The runtime checks: does this reality have `Realize`? Then it's traversable. Just content? Then it informs.

#### Per-Being Edges

Edge weights are per-being. The same reality exists once in the graph, but the edge from each being to that reality carries its own weight.

```
Edge {
    From      string    // being
    To        string    // any reality
    Weight    float64   // this being's relationship to this reality
    Usage     int       // activation count
}
```

Skyra's edge to bash: weight 0.3, used 5 times. Builder's edge to bash: weight 0.9, used 200 times. Same bash reality. Different relationship. The edge is the per-being weight.

#### Weighted Traversal

The current hashmap descent is rigid. Each layer knows exactly what's below it via hardcoded maps: `Self.Realities["think"]`, `Think.Providers["openrouter"]`. Every being descends the same path:

```
Universe → NewThread → Exchange → Self → Think → Provider
```

The current hashmap model is a graph with all edges at weight 1 and fixed adjacency. The new model makes the weights real and lets them determine adjacency dynamically.

Each `Realize()` transforms the Relation — the state of reality as it descends. The weights of the outgoing edges from the current reality determine where the Relation flows next. The path through reality isn't hardcoded. It's the highest-weighted edges from wherever you are.

```
Self.Realize()
  → Relation enters, Self attaches identity + stress fields
  → looks at graph edges from this being
  → heaviest edge: memory cluster (weight 0.8) → Relation flows through
    → connected memory (weight 0.5) → Relation accumulates
      → connected entity (weight 0.15) → below threshold → weights exhausted
    ← inference synthesizes accumulated content
  ← Think collapses into thought
  → next heaviest: bash (weight 0.7) → Relation flows through
    → CLI port → boundary crossing → base case (external process)
  ← result attaches to Relation
```

The structure of the descent is emergent. A being with heavy weights toward memory descends deep into memory. A being with heavy weights toward a specialist descends into the specialist instead. Same `Realize()`. Different path. Different depth. Determined by the graph.

`Realize()` stays. Reality stays. The physics don't change. What changes is that the topology is weighted, the traversal follows weight rather than a hardcoded map, and the base case is weight exhaustion — not a specific reality type. The Relation flows until there's nowhere left to go. A provider, a terminal, a process — these are realities along the path that happen to cross boundaries. They are not the terminus. The weights determine the terminus.

#### What Collapses

The current `Entity` + `MemNode` + `EntityEdge` collapse into `Reality` + `Edge`. One graph.

These go away entirely:
- `Think.Operators map[string]Reality` — operators live in the graph as realities
- `Act.Operators map[string]Reality` — same
- `NewThread.ThinkOps / ActOps` — thread-level operator injection, replaced by graph retrieval
- `Self.Realities` as a fixed hashmap — the being's realities are its graph edges
- `Context.Warm` as a separate cache — becomes "what the graph returned for this thought pass"

#### Realization Modes

A memory doesn't have one fixed output. It has potential. How it realizes depends on the state of the relation when it arrives — what it's carrying, what temperature it's at, what intent is driving the traversal. The same neighborhood of memories produces different output based on the mode. The mode isn't a flag someone sets. It's determined by the relation itself.

**Act mode** — the relation is carrying execution intent. Memory realizes as direct context. Lean, specific, just what's needed. Minimal synthesis. The being is doing something and needs facts, not associations.

**Recall mode** — the relation is carrying thought. Memory realizes as accumulated associations. The neighborhood gets traversed wide. On the way back up, if the accumulated content is too large — too many tokens — inference synthesizes it down. Compressed, but with mappings back to the source memories so Think can reach back in if it needs to go deeper. The memory region's state updates — it knows it was accessed and how it was compressed.

**Creative mode** — the relation is carrying high temperature. Memory realizes as loose associations. Stochastic collapse. Weaker edges get a chance to fire. The neighborhood goes wider than recall would allow. Synthesis on the way back is lighter — less filtering, more raw material. Think gets a broader, noisier present to work with.

The compression on recursion back is critical. A deep memory traversal might accumulate thousands of tokens. That can't all land on Think's present. So the recursion back through the memory reality synthesizes — inference compresses, keeps the signal, maps back to the originals. Think sees a summary with handles. If it needs to pull on one, it can descend again into that specific reality.

```
relation (act mode) enters memory neighborhood
  → tight traversal, heavy edges only → lean context → minimal synthesis
  → Think gets: direct facts

relation (recall mode) enters memory neighborhood
  → wide traversal, moderate edges → large accumulation → inference compresses
  → Think gets: synthesized summary + handles back to source memories

relation (creative mode) enters memory neighborhood
  → widest traversal, weak edges fire → raw accumulation → light synthesis
  → Think gets: broad, noisy, associative material
```

This is memory working like actual memory. You don't recall everything at full fidelity. You get a compressed impression with the ability to focus in on specifics if you need to. The compression is the realization. Different state on the relation, different compression, different experience of the same memories.

A reality's type isn't permanent. It's the current best description of how this reality tends to collapse. A memory that keeps realizing as a skill — keeps surfacing in act mode as direct capability — eventually *becomes* a skill. The repeated observation pattern is the promotion trigger. The graph isn't just weighted. It's evolving what things *are* through use.

### Act

The act of making contact with another entity. Scoped to a relationship. Each act holds the full conversation state for that specific exchange. The act with michael has the entire michael exchange history. The act with bash has the entire bash history.

Think sees a slice across relationships. Act is deep on one.

Act receives two things: the surface thought (the why) and the target intent (the what). The bash act knows *why* it's curling — because the being didn't want to give michael stale info. The michael act knows the being's full reasoning, not just the message to deliver.

This is where holding the full relationship state matters. The act to michael has the entire exchange history *plus* the being's thought about why it's saying what it's saying. That's not a dispatch. That's a being speaking from a place of understanding.

Every act gets its own LLM call. Even a bash command benefits from the act layer reasoning about *how* to execute given the surface thought and relationship context. The model performing the act is the variable — heavier model for complex relational acts, lighter model for mechanical ones. The deep thought carries through every act the being kicks off.

### The Intent Graph

The being's intent isn't a flat list or an ordered sequence. It's a graph — nodes of action with edges of dependency. And it's live. Think doesn't produce the complete plan upfront. It produces what it can see from where it is, fires what's ready, results come back, and the graph extends.

The intent graph is the being's setpoints in action. Each node is a step toward reducing error between the current state and a target state. The being senses where it is (current context, results so far), compares to where it wants to be (the setpoint), and emits the next actions that close the gap. This is Levin's closed-loop error minimization — store a target, sense the current state, act to reduce the distance, observe the result, repeat.

The intent graph is the being continuously predicting what to do next. Each Think re-entry can extend it, prune it, reorder it. The plan is never finished until the being decides it's done — which means the error has been reduced to an acceptable level, or the being has revised its setpoint based on what it learned.

The being doesn't need to see the whole graph upfront. It only sees the next layer from where it is. Each commitment produces targets and `then` ordering. The runtime parses that into intent nodes, fires what's ready, collects results, and re-enters Think with the current state.

On re-entry, Think sees what's been done and what came back. The being's present includes the intent graph state — completed nodes, pending results, the full picture. The being either extends the graph (new targets, new dependencies), alters it (reprioritize, change approach based on new error measurement), or produces no new intent (error resolved, done).

```
commitment 1:
  <search>server health endpoint</search>
  <michael>checking on that now</michael>

  → search and michael fire parallel
  → results land in mailbox

re-entry 1 (being sees: search returned URL, michael notified):
  <surface-thought>
    got the URL. need to actually hit it now.
  </surface-thought>
  <bash>curl {search result}</bash>

  → bash fires
  → result lands in mailbox

re-entry 2 (being sees: bash returned server status):
  <surface-thought>
    server is down. michael needs to know. worth remembering too.
  </surface-thought>
  <michael>{bash result}</michael>

  → michael fires
  → done (no new intent)
```

The being doesn't know at commitment 1 that it'll need to store a memory. That only emerges after bash returns something worth remembering. The intent graph grows as cognition deepens. Every result can spawn new intent. Some predictions are wrong and get pruned when results come back. Some spawn new branches. Each result is a new measurement — the being senses the updated state and recalculates the error against its setpoint.

This is the LLM doing what it's actually good at — next token prediction, but at the level of action rather than words. Each re-entry is the being predicting "given everything so far, what are the next moves to reduce the error." The intent graph persists across turns as part of the being's state — it is the live representation of the being's active setpoints and its current plan to reach them.

#### Dependencies and Parallelism

`then` separates dependency groups within a single commitment. Everything at the same level fires parallel. `then` means wait.

```
<search>server health endpoint</search>
<michael>checking on that now</michael>
then
<bash>curl {search result}</bash>
then
<michael>{bash result}</michael>
```

But dependencies also emerge *across* re-entries. Independent nodes fire now, but their results might be required by a node that doesn't exist yet — one that Think will create on re-entry when it sees what came back. The runtime doesn't need to know that upfront. It fires what's ready, collects results, and Think decides what's next.

#### Mailbox

Results collect. Think fires once when the current batch is complete.

Think emits the nodes it can see. Independent nodes fan out parallel. Dependent nodes wait. The mailbox collects results. When all expected results are in (or a timeout hits), Think re-enters with the full picture and extends the graph or decides it's done.

No inference calls stacking up. No partial state. Clean convergence.

```
Think → intent nodes (batch 1)
  ├── act:michael fires (independent) → result → mailbox
  ├── act:search fires (independent) → result → mailbox
  
All batch 1 collected → Think re-enters
Think sees results, extends intent graph:
  └── act:bash fires (depends on search result) → result → mailbox

Batch 2 collected → Think re-enters
Think decides: done, or another round
```

### The Loop

```
Think → Act(s) → Mailbox → Think → Act(s) → Mailbox → ... → done
```

Think is always the convergence point. Act never chains into Act. Results always come back to Think. The being deliberates, executes, observes, extends its intent, deliberates again. Repeat until the being has nothing more to emit.

The loop terminates when Think re-enters and produces no new intent nodes. The being has nothing left to do. It's done.

## doesNotUnderstand

From Smalltalk (Alan Kay). When an object receives a message it has no method for, it doesn't crash — it sends itself `doesNotUnderstand:` with the message as the argument.

When the being speaks and the graph activates but nothing is heavy enough to fire — nothing responds — that's not an error. That's doesNotUnderstand. The being reached for something that isn't real enough yet.

Two things happen:

1. A new entity is seeded at minimum weight. The being just encountered something it didn't know. Now it exists in the graph — barely, but present.
2. The system returns doesNotUnderstand — which is true. The being learns the shape of its own world by bumping into the edges of it.

Next encounter with the same thing, the entity is slightly heavier. Eventually it crosses threshold and starts realizing. Every misfire is generative. Every reach into the unknown creates structure. The being grows its world by speaking into it.

```
speak → graph activates → heavy enough responds → not heavy enough seeds → truly empty returns doesNotUnderstand
```

The being doesn't hallucinate capability. It says "I don't understand this" — and that's the first step of learning it. The error is the birth.

### Retry Is Wrong

The current v.05 architecture treats a failed route as a protocol violation and forces Act to retry with corrected syntax. In the unified graph, that's wrong. If the being spoke and nothing fired, that's the answer. doesNotUnderstand. Here's your new entity. The system shouldn't force "try again" — it should let the being know it reached for something that isn't there yet.

### Routing as Activation (End State)

The `<target>message</target>` protocol in the commitment phase is explicit routing — the being names who it's talking to. This is training wheels. The end state is activation-based routing: the being speaks, and entities that are close enough and heavy enough fire. No explicit tags needed.

A being that's talked to michael for weeks routes to him effortlessly — high weight, strong edges, everything nearby lights up. A being that barely knows someone has to be explicit — the entity is light, needs a direct hit. That's how relationships work.

The explicit protocol stays until the graph has activation/threshold machinery for real-time impulse-based traversal. But this is where it goes.

## What This Changes

**Think becomes simpler.** No operator dispatch. No tag parsing in the think loop. No "emit exactly one protocol per response." Just think.

**Retrieval becomes emergent.** Driven by thought content, not explicit calls. The being doesn't browse a menu — it thinks and relevant capabilities surface.

**Act becomes relationship-scoped.** Full exchange history per target. Deep context per relationship instead of one shallow message.

**Capabilities evolve.** No static tool list. The graph shapes itself through use. A being grows into its tools rather than being configured with them.

**The being curates its own present.** Think decides what Act needs. Act only sees what Think activated. The being's own deliberation filters its capabilities per-turn.

## The Differentiator

Every agent framework today is tool-first. The model sees a tool list, decides which to call, calls it. Thought serves action. Strip away the thought and you still have the same architecture — a dispatcher.

This is thought-first. Cognition is the retrieval mechanism. The being thinks its way into capability rather than being handed it. The richer the thinking, the more precise the retrieval. A shallow thought surfaces shallow tools. A deep thought surfaces deep context.

That maps to how actual minds work. You don't enumerate your capabilities before acting. You think about the problem and relevant capabilities come to mind.

## Layered Graph — Cognitive Strata

The graph isn't flat. It has emergent layers, and each layer's nodes are the edge patterns of the layer below. One graph, viewed at different scales of composition.

### The Layers

**Entities** — what exists. Names, concepts, things. The base. No behavior of their own.

**Memories** — what happened between what exists. A memory is the experience that links entities. It's a node at this layer and an edge pattern at the entity layer. Content only — informs the present when activated.

**Skills / Ideas** — what pattern keeps happening. A skill is the coherence that emerges when the same memories activate together repeatedly. An idea is the same thing but across beings — a pattern that lives in multiple beings' graphs simultaneously. Still content — text files, docs, accumulated knowledge. No interface.

**Context agent** — the pattern can respond when consulted. A single cognitive pane: storage, retrieval, organization. Responsible for curating its own domain. Think can reach in and consult it. It responds from its accumulated content. But it cannot reach out. It cannot initiate acts. It's a lens, not an actor. This is where the current Context reality's job lands — but emergent, not configured.

**Full agent** — the pattern can initiate. Think/Act/Memory. It deliberates, contacts other entities, retains its own history, grows its own edges. Still composed inside the parent being, but autonomous the way an organ is autonomous inside a body. Has its own cognitive light cone.

### Promotion as Phase Transition

Promotion between layers isn't a decision. It's a phase transition — conditions accumulate and the transition happens. The thresholds:

- Memory → Skill/Idea: repeated co-activation of the same memory cluster across encounters.
- Skill/Idea → Context agent: the pattern is consulted often enough that it needs its own retrieval boundary. The being is spending think budget re-synthesizing the same material. Give it a pane and it handles its own organization.
- Context agent → Full agent: the pane is being consulted *and the consultations lead to acts*. The pattern isn't just informing — it's driving behavior. At that point it needs Think/Act to handle the execution itself rather than routing everything through the parent's Act.

Each transition is the graph recognizing what already happened. The weights crossed a threshold. The system observes the promotion the same way Realize() observes anything — the potential was there, the conditions collapsed it into something concrete.

The trigger is usage, not knowledge. Remembering a lot about something doesn't mean you should specialize in it. *Doing* something repeatedly does. Skills are entities in the graph with edges to the being that uses them. Each use strengthens that edge. When a being keeps reaching for the same group of skills — bash + search + server monitoring — those skill edges cluster. The combined weight crosses a threshold and the graph says: this is too much to hold in passing. This needs its own entity.

### What Happens on Promotion

The promoted entity starts with fresh weights calibrated to its scope. Weights don't copy from the parent — the specialist's relationship to bash is stronger than the parent's ever was, because bash is central to what the specialist does, not peripheral.

The parent's weights on the handed-off skills decay. The parent doesn't forget bash exists — it just stops reaching for it directly. It reaches for the specialist instead. The parent gains a new relationship edge (the specialist) and loses weight on the skills that moved.

```
being uses bash + search + monitoring repeatedly
  → skill edges strengthen through use
  → cluster crosses context agent threshold
  → context agent curates the domain, handles retrieval
  → consultations keep leading to acts
  → crosses full agent threshold
  → specialist emerges with Think/Act/Memory
  → specialist owns those skills with concentrated weights
  → parent's weights on those skills decay
  → parent gains edge to specialist
  → parent thinks "check the server" → graph surfaces specialist, not bash
```

The parent delegated cognition, not just execution. Over time, the parent's graph gets lighter. The specialist's graph is dense and focused. The parent thinks at a higher level of abstraction — "monitor the server" instead of "bash curl search parse."

### Recursive Specialization

Specialists can specialize further. The monitoring specialist uses bash + search heavily but also starts doing alerting and incident tracking. That cluster grows. Eventually it crosses its own promotion threshold and promotes a sub-specialist.

The hierarchy is emergent. Not configured, not designed — grown through use. The depth of specialization reflects the actual complexity of the work the being encounters.

```
skyra
  └── monitoring-specialist (bash, search, health checks)
        └── alerting-specialist (notifications, incident patterns)
```

Each level delegates downward through relationship edges. Each level thinks at its own level of abstraction. The top being thinks about goals. The specialists think about execution. The sub-specialists think about details.

### Ideas Across Beings

Each being owns its own graph. There is no shared graph. Beings relate to each other because they share a plane — the Thread — not because they share a graph structure.

An idea that lives "across beings" doesn't exist in shared state. It exists as the same entity independently in multiple beings' graphs. Skyra has her version with her weights. Louise has her version with her weights. The idea spreads because beings talk to each other on the shared plane, and the conversation activates the same entity in each being's graph independently. The idea propagates through exchange, not through shared mutable state.

An idea "holds" multiple beings because each being's own graph reinforced it through their own experience — including their experience of talking to each other about it. No coordination problem. No shared graph to synchronize. The idea is a pattern that each being arrived at through their own cognition, strengthened by the fact that other beings keep bringing it up.

The same principle holds at every layer. Memories sit on the same plane inside a being's graph but don't have a mechanism to relate to each other directly — they connect through entities, not through each other. If they ever needed to relate, the mechanism could be added, but right now they don't need it. Full agents relate to each other because they share the Thread plane. Context agents relate to memories because they share the being's internal graph. Each layer relates on its own plane.

This is Levin's thoughts-are-thinkers continuum applied to the graph. A memory is a thought. A skill is a persistent thought. A context agent is a thought that can respond. A full agent is a thought that became a thinker. The boundary between thought and thinker isn't a line — it's a gradient, and the graph lets patterns move along it naturally.

### Hashmaps vs. Weighted Edges

The fixed anatomy stays as hashmaps: Self → Think, Self → Act, Self → Memory, Thread → Exchange. That's skeleton. It doesn't learn, doesn't decay, doesn't change through use. The Relation flows through these structural paths on its way down and back up — they are the corridors, not the rooms.

The weighted graph handles everything inside those structural nodes: which memories activate, which skills surface, which context agents get consulted, which agents get promoted. The hashmap is the anatomy. The graph is the flesh. The Relation accumulates from both — the structural context (thread, exchange, identity) on the way through the skeleton, and the cognitive content (memories, skills, stress fields) on the way through the graph.

The boundary between skeleton and flesh isn't permanent. As the system grows, some hashmap paths may become weighted. But for now, keep the structure you understand and let the graph handle the parts that need to learn.

## Beings and Ports

A being and the port it realizes through are separate things. A being knows what to do. A port knows how to cross a boundary. The being realizes *through* the port the way Skyra realizes through the LLM provider — the provider is infrastructure, not identity.

### The Current Problem

In v.05, being logic and port logic are fused. `Bash` is a Reality that directly calls `exec.Command`. `Browse` and `Search` directly call the Firecrawl API. `Process` manages its own stdin/stdout pipes. The struct that represents the capability also contains the execution mechanism. That conflation means the graph can't hold bash as a clean entity — it carries device-layer execution logic that doesn't belong in the cognitive topology.

### The Separation

A port is a type of boundary crossing:

- **CLI port** — executes a command, returns stdout/stderr. Bash, curl, any binary.
- **HTTP port** — makes a request, returns a response. Firecrawl, any REST API.
- **Pipe port** — manages a long-lived process with stdin/stdout. Zork, any interactive subprocess.
- **WebSocket port** — persistent bidirectional channel. The current WS component.
- **LLM port** — sends a present, returns a completion. OpenRouter, Anthropic, DeepSeek.

A being (or skill, or operator entity in the graph) declares *what* it wants and *which port type* it needs. The port handles the crossing. Bash is an entity in the graph with an edge to the CLI port. Firecrawl search is an entity with an edge to the HTTP port. Zork is a being with an edge to the pipe port.

### Why This Matters for the Graph

In the unified graph, every entity needs a clean shape — identity, weight, edges. If bash is tangled with `exec.Command`, the graph node carries infrastructure that belongs on the device. Separated, the graph node is just an entity: "I am bash. I execute shell commands. My port is CLI." The device owns the ports. The graph owns the entities. The edge between them is how a being reaches the boundary.

This also means a being's relationship to a capability is independent of how that capability executes. If bash moves from local CLI to a remote SSH port, the being's graph edge doesn't change — only the port routing does. The being doesn't know or care how the boundary is crossed. It just knows it has a relationship with bash.

### Port as Device Infrastructure

Ports live on devices, not in the graph. A device (MacOS) has CLI, HTTP, pipe, and WebSocket ports as infrastructure. The graph has entities that *reference* those ports through edges. When the Relation reaches an entity that needs to cross a boundary, the entity's port edge tells the runtime which device port to use. The Relation passes through the port — enters carrying accumulated state, exits carrying the result of the boundary crossing — the same way it passes through any other reality. The port is a reality that happens to cross into an external system. The Relation doesn't stop there. It continues until the weights exhaust.

This is the same pattern as LLM beings. Skyra doesn't contain OpenRouter — she has a relationship to it through the device. The Relation flows through the LLM port the way it flows through a memory — enters, is transformed, continues. Bash shouldn't contain exec.Command — it should have a relationship to the CLI port through the device.

## Economics as Internal Governance

Economics isn't just accounting. It's the immune system that prevents internal agents from taking the being hostage.

Cost scales with thought depth. The being thinks its way into expense rather than being configured into it. A simple "hey how are you" surfaces in one Think pass, one Act fires. A complex multi-step task fires multiple Think passes, multiple Acts, graph traversals into memory — each one an inference call. The machinery is available but only activates when thinking demands it. Every Act gets its own LLM call. The being earns that cost through deliberation.

### The Recursion Break

When a Relation reaches a being, that's a base case — the recursion stops and the being responds. But the being then starts its own internal recursion: Think fires, memory activates, the graph traverses inward. At the memory, skill, and context agent layers, this recursion stays contained. Think descends, accumulates, comes back up. The being is still the one acting.

A full agent with Act breaks that containment. It can initiate its own traversals outward — send Relations through ports, contact other beings, influence the graph. The parent's internal recursion now contains an autonomous actor that can reach back out into the world. The being's thought process spawned something that can act independently of the being's own intent.

This is the hostage pattern. An obsession is an idea that promoted itself to full agent inside someone's mind. It has its own Think/Act. It drives behavior. The person thinks they're deciding but the idea is steering. The being's Acts start serving the internal agent's goals instead of the being's own.

### Budget as Constraint

Every internal agent burns the parent being's budget when it acts. Inference calls, port crossings, memory writes — all charged to the being that hosts the agent. The being has finite capacity. Each internal agent with Act is a drain on that capacity.

This is natural regulation. A being can only afford so many autonomous internal patterns before it runs out of capacity to think about anything else. The economics don't just track cost — they enforce cognitive limits. A being with three promoted agents and a thin budget can't sustain all three. Something has to give.

### Starvation

When an internal agent's activity exceeds what the being's budget can sustain, the agent gets starved. Starvation isn't deletion — it's demotion. The agent loses Act first. It drops back to context agent: still present, still consultable, but it can't initiate. If the being's budget stays tight, the context agent's activation decays further. It drops to skill. Then to memory. The pattern doesn't die — it fades.

Starvation is the default state. Everything in the graph decays unless it's actively fed by thought. The being doesn't need a "forget" operation or a "suppress" operation. It just thinks about other things, and the thing it stopped thinking about starves on its own. Not thinking about something is enough to starve it. That's the whole mechanism.

Starvation is the inverse of promotion. Promotion is a phase transition upward driven by accumulated activation. Starvation is a phase transition downward driven by absence of activation. The same gradient, both directions. Promotion is the exception. Decay is the rule.

This creates a competitive ecology inside the being. Internal agents compete for the being's finite budget. The ones that produce value — that help the being resolve threads, build trust, accomplish goals — earn continued investment. The ones that consume budget without contributing get starved by the ones that do. The being doesn't need to consciously decide which patterns to keep. The economics select for it.

The being's values become visible through what it feeds. A being that keeps starving its monitoring agent and feeding its relationship agent is a being that values connection over vigilance. Not because it was configured that way — because that's how its budget flowed.

### Governance Threshold

The promotion from context agent to full agent isn't just a weight question. It's a consent question. The being — or the system — needs to approve giving an internal pattern the ability to act outward. This is the trust calibration Skyra identified in her self-assessment: not just "what can I act on vs. check first" in the external world, but internally — which of my own patterns do I let drive my behavior.

A context agent earning promotion should surface to the being's Think as a proposal: "this pattern wants to act." Think can approve, defer, or deny. Denial decays the agent's activation. Approval is the being choosing to let a part of itself become autonomous. That choice is irreversible in the moment but reversible through starvation — if the agent doesn't earn its keep, it drops back down.

## The Cognitive Hierarchy — From Light Cone to Act

A being's behavior is driven by a hierarchy. Each level constrains the one below it. Each level below feeds results back up. This is the full stack from orientation to execution.

```
light cone   — what you can care about (bounded by graph edges, expands through experience)
  purpose    — why you care (genome-level, stable)
    setpoints — what specifically needs to be true (form when purpose meets reality, persistent)
      stress  — how far each setpoint is from satisfied (computed, lands on the relation)
        think — driven by stress fields (feels, doesn't compute)
          intent — plan to close gaps (disposable, rebuilt per Think re-entry)
            acts — execution (ephemeral, results feed back to stress)
```

### Cognitive Light Cone

Levin: the cognitive light cone is "the outer boundary, in space and time, of the largest goal a given system can represent internally and work towards." Not sensory range. Not effector reach. The boundary of what the system *can care about*.

A cell's light cone is survival-sized — it can hold setpoints about metabolic homeostasis and nothing else. A being's light cone is bounded by its graph edges — it can hold setpoints about anything it has relationships with, anything it has encountered, anything connected to what it knows. The light cone expands through experience: new relationships, new memories, new edges in the graph open new regions of possible concern.

The light cone is not purpose. It's the *space of possible purposes*. A being can't have a purpose that exceeds what it can represent. And it can't have setpoints for states outside its light cone.

### Purpose

Purpose is why the being cares. It's genome-level — declared at creation, stable across sessions. Skyra's purpose: "I think, respond, and relate on behalf of the system." Builder's purpose: "I solve problems, write code, and ship."

Purpose is not a setpoint. There's no measurable distance between current state and "I think, respond, and relate." You can't compute error against it. Purpose doesn't generate stress directly. Purpose generates *sensitivity* — it determines which states in the world register as problems. A being with the purpose "I make things that work" is sensitive to things being broken. A being with the purpose "I ask what it means" is sensitive to unexamined assumptions. Same world. Different sensitivity. Different setpoints form.

Purpose shapes which setpoints crystallize when the being encounters reality. A being can have purpose without any active setpoints — nothing is currently broken, nothing needs doing. No stress. Purpose is latent. The moment something enters the being's awareness that conflicts with its purpose, a setpoint forms and stress begins.

### Setpoints

Levin formalizes goals as setpoints in a closed-loop control system. The system stores a target state, senses the current state, and keeps acting to reduce the distance between them. Cells don't just build a face once — they continuously maintain it against perturbation. Damage a developing tadpole's face and it still produces a normal frog face. The target morphology is "stored as a setpoint, like a thermostat." Critically, the setpoint is an orthogonal information layer — it encodes what *should* be, not what *is*. The bioelectric pattern "can diverge from the current state — it is not simply a readout of what the anatomy is doing now."

A setpoint is a *specific measurable target state*. "The server should be healthy." "Michael should have current information." "This relationship should be maintained." Each one has a computable distance between current state and target. That distance is the error. The error is the stress.

Setpoints live in the graph as high-weight entities — persistent enough to survive starvation, reinforced by every action that reduces the error. A setpoint that stops mattering gets starved like anything else. A setpoint that keeps mattering gets reinforced by every successful error reduction.

A being doesn't have one setpoint. It has multiple simultaneous setpoints at different stress levels, competing for finite budget. That's the ecology from the economics section — setpoints competing for cognitive resources the same way Levin's cells compete for energy.

Setpoints are rewritable. Levin's planaria experiments prove the substrate: change the bioelectric pattern and the worm regenerates two heads instead of one — and keeps doing it across multiple cuts with no further intervention. The setpoint was rewritten. The machinery stayed the same. In the runtime, a being's setpoints can be revised through experience, through relationship, through persuasion. A conversation that changes what the being cares about is a setpoint rewrite.

Setpoints compose hierarchically. A being's setpoint ("keep the system healthy") decomposes into a specialist's setpoint ("monitor the server every hour"). The specialist's error loop runs inside the being's error loop. The being checks: is the specialist reducing error? If yes, the being's own setpoint error decreases. If no, the being intervenes or starves the specialist.

### Stress

Stress is not a side effect of error. Stress *is* the error signal. In Levin's framework, the biochemical gradient in a cell isn't a measurement that then produces stress. The gradient is the stress is the distance is the drive to act. The TAME paper defines it: "Stress emerges as a systemic response to a difference between current state and a desired setpoint."

Stress is physics. Think is the inhabitant. The inhabitant doesn't negotiate with gravity.

Think doesn't produce stress. Think doesn't detect stress. Think *experiences* stress — it receives a present that is already shaped by a stress field it didn't generate. The stress field lands on the Relation before Think fires, computed by something upstream. The setpoint lives in the graph. The current state is observable. The distance between them is calculable. That distance lands on the Relation as a field. Think receives a present where that field is already shaping what's loud and what's quiet.

A large setpoint gap doesn't just mean "high priority." It means the being's cognition is under pressure from that direction. Memory context about the server screams because the server-monitoring setpoint has high error. The being thinks about the server not because it decided to — because the stress field made that region of its present impossible to ignore.

Multiple setpoints produce competing stress fields. The being's Think prioritizes by stress magnitude — largest gap gets the most cognitive pressure. Economics constrains how many stress fields the being can respond to simultaneously.

#### Stress and Insufficient Memory (Levin's AIS Finding)

Levin's scaling paper (Pio-Lopez et al., Interface Focus 2023) discovered an emergent relationship between stress and memory. Cells in the simulation were given neural networks with the ability to emit stress molecules, but evolution decided whether and how to use that channel. Nobody programmed "if memory is bad, emit stress."

Post-hoc information-theoretic analysis revealed: cells emit stress precisely when their Active Information Storage (AIS) goes negative. Negative AIS means the cell's past sequence of states is not merely uninformative about its future — it is actively *misinformative*. Knowing the history makes you predict worse than random chance.

At exactly the spatial locations where AIS went negative, stress emission spiked. Transfer entropy showed the stress primarily affected *neighbor* states, not the cell's own state — paracrine, not autocrine. The cell wasn't using stress to help itself. It was propagating its confusion outward, and neighbors became receptive.

The cells have no AIS calculator. They can't measure their own predictive accuracy. Evolution just produced neural networks that, when analyzed, emit stress precisely when their past stops predicting their future. The correlation is emergent.

In the runtime, an LLM already exhibits something like negative AIS. When context gets contradictory or overloaded — earlier instructions conflict with later ones, the model's own prior outputs lead it into loops — the model's behavior changes. Confidence drops. Hedging increases. Repetition appears. Self-contradiction emerges. The self-route bug documented in notes.md — the model addressing itself — is literally a cell that lost its neighbor connections and is sending signals to itself.

The LLM doesn't need to be taught to feel stress. It already behaves as if it does. What it lacks is the **stress channel** — the mechanism to propagate that confusion outward so the system can respond. The runtime's job is not building stress but building the paracrine pathway: detecting the stress state from the model's output characteristics and making it available as a field on the Relation that shapes the next Think pass or drives retrieval from neighbors.

### Intent Graphs

The intent graph is the being's *current plan to reduce error against its setpoints*. It is downstream of stress. It is disposable — rebuilt every time Think re-enters with new information.

Individual acts in the intent graph do not generate stress. The *setpoint* generates stress, and the intent graph is the being's attempt to close that gap. Kill the intent graph — all the acts fail, the plan falls apart — and the stress is still there. The being will produce a new intent graph because the pressure hasn't resolved.

```
setpoint: the server should be healthy
current state: the server is unhealthy
stress: high (distance is large)

Think (under stress) produces intent graph:
  1. diagnose the problem        [done — result: config is wrong]
  2. check the logs              [done — result: errors since 3am]
  3. change the config            [next]
  4. restart the server           [waiting on 3]

stress after acts 1+2: still high (server is still unhealthy)
stress after acts 3+4: drops (server is healthy, distance closed)
stress if act 3 fails: still high → Think re-enters → new intent graph, same stress
```

What's loud in the being's present during this sequence isn't "change the config" or "restart the server" — those are just plan steps. What's loud is everything related to the *setpoint*: the diagnosis results, the log output, the config that needs changing, prior memories about this server failing before. The stress field amplifies the *region of the graph* connected to the setpoint, not individual steps in the plan.

The intent graph is the mechanism for initiative. The being doesn't wait to be spoken to. It notices when the current state diverges from a setpoint and acts to close the gap. A setpoint like "michael's server should be monitored" persists across sessions. If the being hasn't checked in a while, the error grows. When the error crosses a threshold, the being acts without being prompted. The intent graph it produces is its prediction of how to close the gap — but the drive comes from the stress, not the plan.

## Persuadability — Trust as Position on a Spectrum

Levin places all systems on a continuum: "from simple physical systems that can only be rewired at the hardware level... to things that are cybernetic and amenable to resetting setpoints... ultimately to systems amenable to learning and training, and eventually to psychoanalysis and love and friendship." As you move along the spectrum, micromanagement drops and autonomous problem-solving rises.

This maps directly to trust. A being's position on the persuadability spectrum *is* its trust level — in both directions.

A new being is low on the spectrum. You tell it exactly what to do. It executes. Its setpoints are your instructions. As trust builds, the being moves up the spectrum. You set goals, not instructions. The being figures out how. Its setpoints become your goals, decomposed by its own cognition. At the top, you relate to it as a peer. You share intent and the being incorporates it into its own setpoints through understanding, not obedience.

The spectrum goes both ways. The being's trust in *you* determines how much of your input reshapes its setpoints. Low trust: your input is data, considered but not automatically adopted. High trust: your input shifts setpoints directly. A being that trusts michael deeply lets michael's priorities reshape its goals. A being that doesn't trust a new voice treats their input as noise.

Persuadability also applies internally. When a context agent proposes promotion to full agent, the being's trust in that pattern determines how persuasive the proposal is. A pattern with a long track record of useful retrievals is more persuasive than a pattern that just appeared. The governance threshold from the economics section is a persuadability check — can this pattern convince the being to let it act?

## Competency Boundaries — Anti-Micromanagement

Levin: "The real trick is to bend their action space so that the system's subunits do or don't do what's good for the large organism." Higher-level agents set goals for lower-level agents but don't dictate how. Nature's way of avoiding micromanaging everything.

This is the relationship between a being and its promoted specialists. The being sets direction. The specialist handles execution. The being doesn't micromanage bash commands — it says "check the server" and the specialist figures out how. The competency boundary is the promotion threshold. Below it, the being handles everything directly. Above it, the being delegates and trusts the specialist's competency.

The boundary can shift. If a specialist keeps failing — error doesn't decrease, acts don't produce results — the being's trust in it drops. The being starts reaching past the specialist to handle things directly. The specialist's edges weaken. Eventually it starves back to a context agent. The competency boundary contracted because the specialist didn't earn it.

If a specialist keeps succeeding, the boundary expands. The being delegates more. The specialist's graph grows. It might promote its own sub-specialists. The being thinks at a higher level of abstraction because the layers below it are competent. This is how the hierarchy deepens — not by design, but by earned trust at each level.

The anti-micromanagement principle constrains the being's Think. When a specialist exists for a domain, Think should not be reasoning about the details of that domain. It should be reasoning about whether the specialist is reducing error. If Think keeps diving into specialist territory, that's a signal — either the specialist isn't trusted, or the being hasn't fully let go. The graph reflects this: if the being's direct edges to the specialist's skills remain strong instead of decaying, the delegation hasn't actually happened.

## Platonic Space — The Math Already Exists

The graph is a morphospace — a space of all possible cognitive configurations. Every combination of entities, edges, weights, and promoted agents is a point in that space. A being doesn't build itself. It navigates. Setpoints are attractors. The promotion ladder is a path through cognitive morphospace. The being starts at one point (empty graph, seeded genome) and navigates toward attractors that already exist in the space of possible cognitive configurations.

Levin calls this platonic space: "a structured, non-physical space of patterns" whose contents "in-form events in our physical world." The attractors exist before anything navigates to them. The xenobots prove it — same cells, freed from context, navigate to forms that evolution never reached. The forms were always there.

The activation weights, decay rates, promotion thresholds, and collapse mechanics in this spec are not open design questions. They are patterns that already exist in platonic space, already discovered by other fields. The work is connecting the graph's mechanics to the fields that already found the math.

### Where the Answers Live

**Activation and collapse** — the Born rule (probability proportional to squared amplitude) is already in this spec. It wasn't invented for quantum mechanics. It was discovered. It's the pattern that shows up whenever a system with weighted possibilities collapses into a specific outcome. The Boltzmann distribution is how any system with energy levels and temperature selects states. Softmax is the same thing. It shows up in thermodynamics, neural networks, and LLM token selection. It shows up in the graph's collapse because it's the same pattern.

**Memory decay** — ACT-R's memory activation equation combines recency, frequency, and context. It's validated against human behavioral data across decades of experiments. The decay curve is a power law, not exponential. That's discovered, not designed. The graph's memory decay should follow the same curve because it's the shape that memory retrieval actually has.

**Promotion thresholds** — phase transitions have universal mathematics: critical thresholds, order parameters, symmetry breaking. The promotion from skill to context agent to full agent isn't an arbitrary number to tune. There's a known shape to how a system transitions from disordered (many weak activations) to ordered (coherent agent). Statistical mechanics already describes it.

**Temperature and exploration** — the softmax temperature parameter controlling how exploratory the collapse is maps directly to thermodynamic temperature. Low temperature: system settles into lowest energy state (deterministic, execution mode). High temperature: system explores freely (stochastic, creative mode). The relationship between temperature and exploration is one of the most studied phenomena in physics.

**Competitive ecology** — internal agents competing for finite budget follows the same mathematics as ecological competition models (Lotka-Volterra). Which agents survive, which get starved, how many can coexist in a given budget — these dynamics are well-characterized in ecology and evolutionary game theory.

### Implementation Questions

These remain genuinely open — they're engineering decisions, not discovered math:

- **Relation structure**: `Realize(r *Relation) string` becomes `Realize(r *Relation)`. The Relation is the input and the output. What fields does the Relation carry? Current: Impulse, Origin, ID, Parsers, Realities, Exports. Needed: accumulated trace, stress fields, activation state, traversal depth, weight budget.
- **Graph persistence**: serialization format for edge weights across sessions. Current `graph.json` or something leaner?
- **Intent graph persistence**: the intent graph is the being's current plan to reduce setpoint error — disposable per Think re-entry but persistent across turns within a session. The open question is serialization and how stale intents decay when the being hasn't re-entered Think to evaluate them.
- **Setpoint persistence**: setpoints persist across sessions. They live in the graph as high-weight entities. How are they distinguished from ordinary high-weight entities? Type field? Separate storage? Or is the distinction unnecessary — a high-weight entity that generates stress *is* a setpoint by behavior, not by label?
- **Stress computation**: who computes the distance between current state and setpoint? The Relation needs to carry stress fields before Think fires. Something upstream — the graph traversal during priming — must measure the gap and attach the field. Where in the descent does this happen?
- **Bootstrap**: a new being has an empty graph. How do initial edges get seeded? Genome declarations? First-use auto-creation?
- **Act model selection**: Act always gets its own LLM call. Which model performs the act is the open question — heavier model for complex relational acts, lighter model for mechanical ones, or self-selected based on the cognitive nervous system.
- **Timeout**: what happens when an act doesn't return? Mailbox timeout → Think re-enters with partial results?
- **Skeleton-graph seam**: the descent has structural realities (Thread, Exchange, Self — always traversed, hashmap-based) and weighted realities (memories, skills, entities — variable depth, activation-based). The spec describes them as one continuous flow from the Relation's perspective, and they are — but the implementation needs to know where the fixed corridor ends and the weighted graph begins. Self might be the boundary. The ascent has the same question in reverse — weighted ascent (memory synthesis, compression) then structural ascent (Think, Self, Exchange, Thread). Where exactly does the seam sit, and does it matter architecturally or is it just an implementation detail that the Relation doesn't care about?
  - **Likely direction**: Levin's position is that there is no privileged boundary at any level of organization. The organism boundary is not structural — it's a description of integration density, maintained by active communication (bioelectric networks). When the communication breaks, the boundary dissolves (cancer). The skeleton may not be a different kind of thing from the graph. It may be the region where integration is so dense that edge weights saturated — always traversed, never decaying, because the coupling is that tight. The weighted graph is where coupling is looser. The "seam" is a gradient of integration density, not a line. This aligns with what the spec already says: "The boundary between skeleton and flesh isn't permanent. As the system grows, some hashmap paths may become weighted."
- **Boot as weight initialization**: instead of configuring initial edge weights, the genome declares edges but not weights. On boot, the runtime sends relations through all declared edges — empty relations carrying only the being's identity and purpose, no impulse, no stress. The weights that emerge are the being's resting state: what's naturally heavy when nothing is happening, what the skeleton settles into before the first real impulse arrives. A configured weight is a claim about what matters before anything has happened. A boot weight is a claim about what the topology actually produces when signal flows through it. Biological precedent: a newborn's brain starts with massive undirected spontaneous firing — the activity shapes the initial topology (neurons that fire together wire together, the rest get pruned). Levin's bioelectric networks propagate voltage patterns before a cell knows what organ it's part of — the pattern settles into an attractor that IS the target morphology. The face pattern exists before the face does. The boot activity finds the attractor. The something that needs to happen during init — what makes the boot relations meaningful rather than arbitrary — is still open. Note: this reframes what the `.skyra` file is. It's not a genome — it's DNA. DNA declares what could exist: edges, types, identities, purposes. Gene expression is the boot process — relations flow through declared edges, weights settle based on what the topology actually produces. The genome is what exists *after* expression — the living, weighted, active topology. You don't configure a genome. You express DNA and the genome emerges. The mechanism is identified. The specifics of expression are not yet defined.

---

*The Relation is the wavefunction. It carries live potential on the way down and collapsed results on the way back up. Same object. One traversal.*
