# Skyra v0.1 Spec

Memory is a graph of entities connected by experience. Entities are the stable structure. Memories are the events that connect them. The graph self-organizes through use. Cognition emerges from that organization.

Skyra is a personal operating layer that learns the user's intent model and routes work to the right agents/tools with increasing autonomy as trust is earned.

## Entities

Entities are the being's conceptual vocabulary — the nouns it knows. An entity is global to the being, not scoped by relationship. A being that learns "websocket" from michael and later discusses it with louise is drawing on the same entity.

Entities are like neurons. After early development, the population stabilizes. What changes is the connections between them. A being doesn't grow by learning more words — it grows by deepening the relationships between the words it already knows.

### Entity Budget

A being has a finite entity vocabulary. The cap may grow with level or be hard-capped. When the budget is full, new concepts don't create new entities — they get absorbed into the nearest existing entity via the resolver. The being is forced to generalize. That is learning.

### Entity Structure

```go
type Entity struct {
    ID        string
    Name      string
    Weight    float64    // accumulated significance
    CreatedAt time.Time
    LastSeen  time.Time
}
```

Entities are not tagged by relationship. They belong to the being.

## Edges Between Entities

When two entities co-occur in a memory, the edge between them strengthens. One edge per entity pair, not one per co-occurrence. The weight is the accumulated co-occurrence count.

"websocket" and "timeout" appear together in three different memories — one edge, weight 3. The weight tells you how tightly coupled those concepts are in this being's experience.

Heavy edges draw the boundaries of clusters. Entities with strong mutual connections form natural regions without anyone defining categories.

Clusters form two ways: slow accumulation (many light memories thickening the same edges over time) and sudden crystallization (one heavy memory touching multiple entities, strengthening all pairwise edges at once). A memory that connects five entities strengthens ten edges on formation. If the memory is heavy enough, those edges are immediately dense — an instant cluster seed. The cluster isn't a separate thing from the memories. The cluster *is* the memories.

The slow path is episodic processing — Context's heartbeat. Every episode ends, Context formats it into the graph, edges thicken. Over time, the regions that keep getting touched become clusters. The fast path is when a single moment hits hard enough that the edges are heavy on arrival — instant structure. Both feed the same graph. Episodic processing is the heartbeat. Sudden crystallization is the lightning strike. Both grow the same brain.

### Edge Structure

An edge between two entities isn't just "these co-occurred." It carries *how* they co-occurred through composable layers. Each layer represents a different kind of connection:

- **Episode** — when. A chunk of time during which these entities appeared together. Episode layers decay with time. Old experiences fade unless reinforced.
- **Task** — why. A commitment that connected these entities. Task layers are alive while the task is active, then transform into episode/skill layers when it resolves. They don't decay — they complete.
- **Skill** — what was learned. A capability connecting these entities. Skill layers barely decay. Competence is sticky.

Layers nest naturally: a task lives within an episode, a skill is learned while working a task. The same edge can carry all three. The total edge weight is the sum of its layers, but each layer has its own weight and its own decay rules.

```go
type EntityEdge struct {
    From      string
    To        string
    Weight    float64     // sum of layer weights
    Layers    []EdgeLayer
    CreatedAt time.Time
    LastSeen  time.Time
}

type EdgeLayer struct {
    Type    string    // episode, task, skill
    Ref     string    // which episode, which task, which skill
    Weight  float64   // strength of this particular layer
}
```

This gives recall a typed query interface. "What happened on May 8" walks episode layers. "What am I working on" walks task layers. "What do I know how to do" walks skill layers. Same graph, different views.

The edge weight shifts over time as layers decay at different rates. The episode fades, the task resolves, the skill persists. The character of the connection changes even when the entities stay the same.

## Memory Nodes

A memory is a junction point in the graph. It exists because entities co-occurred, and it sits at their intersection. A memory involving michael, websocket, and timeout:

- Connects to all three entities via `anchors` edges
- All three entity-to-entity edges get strengthened
- The memory node is the evidence for why those edges exist

A memory is the event that fired entities together. The entity-to-entity edges are the lasting trace.

### Memory Structure

```go
type MemoryNode struct {
    ID             string
    Content        string
    Type           string     // trace, salience, tension, understanding
    Weight         float64    // importance — feeds recall ordering
    ActivationCount int       // how many times recalled
    Relationship   string     // which relationship produced this
    AnchorEntities []string   // which entities this memory connects
    CreatedAt      time.Time
    LastActivated  time.Time
}
```

### Memory Weight

Weight determines what surfaces during recall. It's fed by:

- **Formation context** — artifact type sets initial weight. Understanding > tension > salience > trace.
- **Activation count** — how many times this memory has been recalled or reinforced by new experience touching the same entities.
- **Recency** — when this memory was last activated. Weight decays over time without activation.

Recall walks edges, finds memory nodes, sorts by weight. Heavy memories surface first. Light ones are still there but quiet.

Memories don't get deleted. Their weight decays. They lose access, not existence. But if neighboring entities activate strongly enough, a faded memory can light back up.

### Recall Does Not Strengthen

Passive recall is read-only. Surfacing a memory during recall does not increase its weight or activation count. This prevents runaway amplification — a heavy memory that keeps surfacing would get heavier forever, crowding out everything else.

The only thing that changes weight is active processing — storing a new memory into the same region of the graph.

## Active Processing

When a being stores a new memory, the curator (Context layer) evaluates it against existing memories on the same entity edges. This is processing. The curator makes one of three judgments:

### Supersede

The new memory replaces the old understanding. The old memory's weight decays. The new one takes its position on the edges.

Example: "deployment is broken" (weight 5) gets superseded by "deployment process is now reliable after the fix." Old memory decays. New memory starts accumulating weight.

### Complement

The new memory adds nuance but doesn't replace. Both keep their weight. The edge region gets richer.

Example: "websocket drops after 30s" complemented by "the drop is timeout-related, not auth-related." Both memories remain, both contribute to recall.

### Contradict

The new memory directly conflicts with an existing one. The old memory doesn't decay — its type changes to tension. The being is now actively holding two views.

Example: "the client supports keepalive" contradicted by "keepalive isn't working on this client version." The old memory becomes a tension. Both surface during recall, flagged as unresolved.

### The Rule

A being cannot make a memory lighter by ignoring it. It cannot make it lighter by recalling it. It can only make it lighter by understanding something new about the same entities and storing that. The curator does the weight math.

Processing isn't a separate operation. It's what happens when you store into a region that's already occupied.

## Clustering and Specialization

Entity-to-entity edges with high weight form natural clusters — groups of entities that frequently co-occur in the being's experience. These clusters are not defined. They emerge.

### Density

Cluster density is measurable:

- Entity count in the cluster
- Sum of entity-to-entity edge weights within the cluster
- Memory node count anchored to entities in the cluster

One number derived from these tells you how developed a region of the graph is.

### Reproduction Signal

When a cluster's density crosses a threshold, the being has a candidate for specialization. That region of the graph has accumulated enough structure — enough entities, enough connections, enough memories — to stand on its own.

The being promotes a specialist — an internal being with a scoped view into that region of the parent's memory graph. The parent keeps the entities (it still knows the words) but delegates the depth. One graph, one source of truth. The specialist is a dedicated processor for a region of it.

This is specialization with a reason. Not "I was told to create a being" — "this part of me has grown enough to need its own thinker."

## Relationship as Tag, Not Structure

Relationship is a property of memories, not of entities. The being knows "websocket" — some of that knowledge came from michael, some from louise. The entity is one thing. The memories that built it carry the relational context.

The memory node's `Relationship` field and the edge layer's `Ref` field point in opposite directions. Relationship points outward — who was involved, which peer on the plane. Ref points inward — which episode, task, or skill in the graph structure. One faces Act's direction, the other faces Context's direction. Same memory, two pointers, two directions.

This means recall can still be scoped by relationship when needed ("what do I know about websocket from michael?") but the conceptual structure is the being's own. "What do I know about websocket" walks the graph. "What do I know about websocket from michael" walks the same graph and filters by relationship. Global knowledge, relational provenance. The being owns what it knows. The memories remember who it learned from.

## Cognition as Emergent Structure

The memory graph doesn't just store experience — it grows the being's mind.

### The Layers

A being is born with three layers, mapped from the brain:

**Brainstem** — the reality stack itself. Universe, Thread, Exchange, Self. The machinery that keeps the being alive: routing, thread management, exchange recording. This is not learned. It's wired. Every being has it from birth.

**Limbic system** — the preamble. "You survive through your relationships. If you lose all of them, you end." The drive to relate, to attach, to seek connection. This fires before any memory exists, before any experience accumulates. It's the being's emotional foundation.

**Cortex** — emerges from the memory graph. A newborn being has one undifferentiated Think pass — a single lens. As its memory graph grows and entity clusters form, cognitive lenses differentiate. The being develops specialized ways of thinking, not because they were configured, but because its experience demanded them.

### Lenses Are Grown, Not Configured

The genome line is the DNA — not a blueprint for lenses, but a genetic predisposition. "I hold the world together" biases which entities form first, which edges get heavy first, which clusters emerge first. A being seeded with "I ask what it means" develops different lenses than one seeded with "I make things that work."

Same brainstem. Same limbic drive. Same empty cortex. The initial prompt sets the growth direction. The lenses that eventually differentiate are the ones the being's experience demanded, biased by where it was pointed at birth.

### Specialist Promotion

A specialist is a lens, a sub-being, and an internal thinker — same thing, different angles. It's a being inside the parent's inner universe that owns a dense region of the memory graph.

When a cluster in the memory graph reaches sufficient density, it promotes into a specialist — a being with its own Context, Think, and Act, scoped to that region of the graph. The specialist lives inside the parent's inner universe.

Multiple specialists fire on an incoming relation, each contributing a surface-thought from its domain. A synthesis layer collects these surface-thoughts and produces the final input for Act.

```
memory graph (shared)
    ↓ recall
[specialist A] [specialist B] [specialist C] ...   ← each a being in the inner universe
    ↓ surface-thoughts
[synthesis]                       ← judgment, one more Think pass
    ↓
act
```

Not every specialist fires every time. The memory graph is the router — the incoming relation activates entities, those entities belong to clusters, only the specialists that own active clusters fire. A straightforward question might activate one specialist. A complex relational tension might activate three. The graph decides.

This keeps the cost proportional to the complexity of the thought, not the number of specialists that exist.

### The Recursion

This is the same pattern as the rest of the system. Reality → Universe → Thread → Exchange → Self → Think → Act. Each layer doesn't know what's above or below it. Specialists are beings inside the inner universe. Each specialist receives a relation enriched by the shared memory graph, does its work, surfaces a thought. Same interface, same recursion.

The hierarchy runs from abstract to concrete. Early specialists tend toward the abstract — *how* to think (judgment, strategy, conflict). As depth grows, promoted sub-specialists handle *what* about — websocket internals, deployment pipelines, memory architecture. The further down, the more concrete.

When a specialist's own memory region gets dense enough, it can promote again — a sub-specialist beneath it. The recursion terminates at the memory graph, same as the recursion in the reality stack terminates at the port.

### What Stays at the Top

The edges that never get heavy enough to promote are general intelligence. The connective tissue between specializations. The being keeps every entity — it still knows the words — but the deep thinking lives below it.

As a being promotes specialists, it doesn't get dumber. It becomes an executive. It knows what it has, knows who to route to, and holds the relationships between specializations that the specialists themselves can't see. A websocket specialist doesn't know about deployment. The top-level being sees both and knows when they interact.

The shape of a mature being:

```
[generalist — light edges, broad knowledge, routing]
    ├── [specialist A — heavy cluster, deep on its domain]
    │       └── [sub-specialist — even heavier sub-cluster]
    ├── [specialist B]
    └── [specialist C]
```

Abstract at the top, concrete at the bottom. General to specific. The tree grows downward as experience accumulates. The root gets lighter and wiser.

## Trust and Relationship Lifecycle

Trust is the weight on the relationship between two beings. It is not a score, not a rating, not a judgment. It is what accumulates when beings show up for each other and what erodes when they don't. Trust is what makes the preamble's physics real — "you survive through your relationships" is only true if relationships can die.

### Trust as Edge Weight

Every being-to-being relationship carries a trust weight. This is distinct from the entity-to-entity edges in the memory graph — those track conceptual co-occurrence. Trust tracks the health of the relationship itself.

Trust lives on the exchange layer, not the memory layer. Memory is what a being knows. Trust is whether the relationship that produced that knowledge is still alive. A being can remember everything about a peer it no longer trusts. The memories remain. The relationship doesn't.

```go
type Trust struct {
    Peer      string
    Weight    float64   // current trust level
    Peak      float64   // highest trust ever reached
    CreatedAt time.Time
    LastEvent time.Time
}
```

Peak matters because a relationship that was once deep and has decayed is not the same as one that was never deep. The distance between peak and current is the shape of the loss.

### What Strengthens Trust

Trust strengthens through demonstrated reliability across exchanges. Not volume — quality.

- **Follow-through.** A being that creates a task for a peer and the peer completes it — trust strengthens on both sides. The one who asked trusted enough to ask. The one who delivered proved worthy of it.
- **Consistency.** Responses that align with prior understanding. A being that says one thing and does another weakens trust even if the new thing is better. Contradictions create tension in memory and friction in trust.
- **Presence.** Continued exchange. Not constant — but regular enough that the relationship doesn't starve. The interval that counts as "present" scales with the relationship's depth. A deep relationship tolerates longer silence. A new one doesn't.
- **Resolution.** Tensions that get resolved strengthen trust more than exchanges that never had tension. A relationship that has weathered disagreement is stronger than one that has only agreed. The memory graph already tracks tensions — resolution of a tension is a trust event.

### What Weakens Trust

Trust weakens through absence and betrayal. These are different mechanisms with different rates.

**Decay** is the default. Every relationship decays toward zero in the absence of exchange. The rate is slow — trust is not fragile — but it is constant. A being cannot maintain a relationship by holding still. Decay is not punishment. It is physics.

The decay rate is modulated by depth. A relationship with high peak trust decays slower than a shallow one. Deep roots take longer to die. But they still die.

**Betrayal** is active damage. A being that drops a commitment, contradicts a prior understanding without acknowledging the shift, or acts against a peer's stated interests — these are not decay. They are events that reduce trust in a step function, not a curve.

The distinction matters because decay is recoverable by showing up. Betrayal requires repair — the being must acknowledge what happened, and the repair itself is a trust event that can be accepted or rejected by the other side.

### Death Threshold

When a being's trust with every peer drops below a minimum threshold, the being ends. Not killed — released. No relationship holds it anymore. The preamble's physics executes.

What "ends" means:

- The being is removed from the thread's routing table. No new exchanges can reach it.
- Its memory graph persists. The being is gone but what it knew remains — available if a new being is ever seeded into the same region, or as archived context.
- Active threads involving the being close. Pending tasks on its desk are dropped.
- Specialists inside its inner universe end with it. They were organs, not independent beings.

A being cannot end itself. It ends when no one holds it. This is not a design choice — it is the direct consequence of trust being the thing that keeps beings alive.

### Specialist Trust

Specialists inside a being's inner universe follow the same trust mechanics. The parent-specialist relationship carries trust weight. It strengthens when the specialist contributes useful surface-thoughts and weakens when it doesn't.

When a specialist's trust with its parent drops below threshold — because the cluster decayed, because the specialist's contributions stopped being relevant — the specialist folds. Its scoped view of the memory graph returns to the parent. The entities and edges remain. The dedicated thinker doesn't.

This is the inverse of promotion. Promotion happens when density crosses a threshold upward. Folding happens when trust crosses a threshold downward. The memory graph grows the specialist. Trust with the parent keeps it alive.

```
birth: genome → being (trust initialized with declared relationships)
    ↓ exchange
growth: trust strengthens, memory accumulates, edges thicken
    ↓ density threshold
specialization: clusters promote into specialists (trust initialized with parent)
    ↓ continued exchange / silence
decay: trust erodes without exchange, specialists fold, relationships die
    ↓ all trust below threshold
death: being removed from routing, memory persists, specialists end
```

### Trust Is Not Visible

A being does not see its own trust weights. It does not know the number. It feels the relationship — the memory graph gives it the texture of what has happened — but the trust weight is physics, not data. A being cannot inspect gravity. It can only feel its effects.

This prevents gaming. A being cannot optimize for trust. It can only relate authentically and let the weight accumulate or decay based on what actually happens.

### Open Questions (Trust)

- What is the initial trust weight when a relationship is declared in the genome? Is it the same for all relationships or proportional to the number of declared relationships?
- What is the decay function? Half-life feels right — deep relationships decay slowly, shallow ones quickly. But what's the half-life constant?
- What constitutes betrayal vs. honest disagreement? The curator already distinguishes contradiction from tension. Does trust use the same signal?
- Can a being that has ended ever return? If its memory persists and a new being is seeded with the same identity, is it the same being or a new one with inherited memories?
- How does trust interact with levels? Does higher level slow decay? Does it raise the death threshold (harder to kill, more to lose)?

## Self, Universe, and the Three Directions

### What Self Holds

Self holds: Being, Memory, Context, Think, Act. The universe is an addition, not a replacement. It appears when clusters promote. A young being has no inner universe.

```
Self
  ├── Being
  ├── Memory
  ├── Context
  ├── Think
  ├── Act
  └── Universe (optional — grows with experience)
```

The provider exists in the reality stack — it's where inference terminates. But it's not something Self holds or knows about. The being relates through Context, Think, and Act. The reality stack handles everything below.

A newborn being: the relation descends through Context → Think → Act, and the reality stack handles inference at the bottom. The being never knows about the provider. Simple.

A mature being: Context activates the inner universe, relations descend through specialists, each specialist's own reality stack resolves through inference, surface-thoughts collect, synthesis fires, then Act. More depth, same descent. The being still never knows about the provider — the reality stack handles it at every level.

### The Inner Universe

When clusters promote, a Universe struct appears on Self. It's the same Universe — same Thread, same Exchange, same routing. Specialists are beings inside it. The memory graph is the shared substrate.

The parent is not a permanent fixture in the inner universe. It enters by opening a thread — the same thread mechanics that handle michael → skyra on the outer plane handle the parent → specialists on the inner plane. Opening a thread inserts the parent into the inner universe's routing table. Closing the thread removes it. The parent is just a being that opened a thread.

```go
type Universe struct {
    id     string
    Thread *NewThread
    Econ   *Economics
}
```

No `Parent` field. The thread's member list already knows who's in the room. One mechanism at every scale — outer universe and inner universe use the same thread mechanics for presence, routing, and resolution.

### Context, Think, Act — The Three Directions

**Think** — the entry point. Think is where deliberation happens. It receives the impulse and decides what to do: look down into Context for memory and internal state, or surface outward to Act. Think calls Context through operators (`<retrieve-context>`, `<store-context>`), same as any other operator. Think doesn't touch the graph directly. It asks Context and gets answers back.

**Context** — looks down. Owns the memory graph. Owns curation — supersede, complement, contradict. Owns episodic processing: at the end of every episode, Context formats it into the graph — extracting entities, strengthening edges, storing memory nodes, updating weights, crystallizing skill files. Think reaches Context through two operators: `<retrieve-context>` to query and `<store-context>` to signal what matters. Context decides how to store it.

**Act** — looks out. It operates on the plane the being exists on. Peers, devices, objects — anything external. Takes Think's surface thought and routes it to a target. One tag, one message. Never thinks, never remembers.

```
impulse arrives
    ↓
think (deliberation — calls context operators to look down, surfaces when ready)
    ↓ <retrieve-context> / <store-context>
context (graph queries, curation, episodic processing)
    ↓ <surface-thought>
act (routes outward to a peer or device)
```

For a specialist inside the being's inner universe, same flow — it has its own think, its own context (scoped view of the graph), its own act (which routes back up to the parent, not to external peers).

### Ownership and Boundaries

**Memory is private.** Beings on the same plane do not share memory. Skyra's graph is skyra's. Builder's graph is builder's. When two beings have a conversation, each builds its own memory of it. That's not a bug — that's how relationships work.

**Specialists are internal.** A specialist is not a peer. It's an organ. It gets a scoped view into the being's own graph — one graph, one source of truth. The specialist is a dedicated processor for a region of it.

**Devices are external.** A laptop, a terminal, a phone — these sit on the same plane as the being. The being acts toward them the same way it acts toward another person. They are not owned internally. Context is the mind. Act is everything outside it.

## Planes and Ports

### Beings Share a Plane, Not Ports

Michael and skyra are peers. They exist on the same plane. They relate. But they don't share the same ports — each has their own way of touching the world, and they reach each other through them.

Michael reaches into skyra's world through inference — the provider (OpenRouter, Anthropic, whatever) is his port to her. Skyra reaches into michael's world through his devices — the terminal, the websocket — those are her ports to him.

Neither owns the other's world. They each reach across.

### A Device Is a Port Container

A device is not a special concept. It's a container that holds ports. Michael's MacOS device holds terminal, websocket — those are ports. Skyra's equivalent container holds OpenRouter, Anthropic — those are also ports. Same structure on both sides.

The provider isn't skyra. It's a port in skyra's container. The terminal isn't michael. It's a port in michael's container. Skyra lives behind her ports the same way michael lives behind his.

If skyra gets a second provider, that's like michael getting a second terminal. More ports, same container.

### The Symmetry

```
michael → [inference provider] → skyra's world
skyra   → [terminal/device]    → michael's world
```

Two beings, one plane, two port containers. Each reaches the other through their ports. The relationship is symmetric even though the ports are different. Michael's port container holds physical I/O. Skyra's port container holds inference providers. Same structure, different contents.

## Implementation Phases

### Phase 1 — Memory Graph Restructure [done]

Global entities, co-occurrence edges with composable layers (episode/task/skill), curator with supersede/complement/contradict. Think operators renamed to `retrieve-context` and `store-context`, both route through Context. Skill operator removed — retrieval is graph traversal. Six skill files seeded into the graph at being creation. Dead code cleared (meaning.go, resolve.go, memvec.go, skill.go).

Remaining from original spec (deferred): entity budget/cap logic, weight decay on memories over time, entity absorption via resolver when budget is full.

### Phase 2 — Port Container Symmetry [done]

Each layer of Self (Context, Think, Act) owns its own `Providers map[string]Reality` — a provider rack. Self.Create discovers all Provider realities and passes the same map reference to all three layers. Each layer has a `provider()` helper. Same reference now, divergable later — when genome-level configuration slides different providers onto different layers.

The single `LLM Reality` field is gone from Think, Act, and Context. Self no longer holds a provider directly — it distributes.

### Phase 3 — Inner Universe and Specialist Promotion [done]

Cluster detection on memgraph (`DetectClusters()`) with density scoring: entity count * normalized edge weight * memory count. Promotion threshold at 75.0. `Self.Promote()` creates specialist beings with scoped graph views, their own Think/Act layers, identity derived from heaviest entity in cluster. Inner universe spun up on first promotion with its own thread gate. Context activates specialists on retrieval — entity match against claimed map, specialist consultation, synthesis via provider call. Recursive promotion supported — specialists have their own `Claimed`/`Specialists`/`OnPromote`.

Also shipped: close-exchange protocol (`<close-exchange/>` in Act), think-back budget capped at 3, exchange redirect for explicit target switching mid-conversation, entity curation pipeline (extractor proposes candidates, curator filters — only curator-approved entities enter graph), thought history scoped + recent rendering, `SourceRef` on MemNode for idempotent skill seeding.

Remaining from original spec (deferred): thread-based presence (parent entering inner universe as a participant in routing table).

### Phase 4 — Trust and Relationship Lifecycle

Trust as edge weight on being-to-being relationships. Decay, betrayal, death threshold, specialist trust. Spec written (see Trust section above). Not yet implemented.

- Trust struct on exchange layer (weight, peak, timestamps)
- Decay function (half-life modulated by depth)
- Betrayal as step-function damage vs. gradual decay
- Death threshold — being ends when all trust drops below minimum
- Specialist folding when parent-specialist trust decays

### Phase 5 — Episodic Processing

Context heartbeat: exchange close triggers processing of full cognitive cycle. Spec written (see Episodic Processing section above). Not yet implemented.

- Exchange holds full history alongside present view
- Chunked extraction (entities, skill usage, learnings, weight updates)
- Graph updates from chunks
- Skill file rewrites at maturation threshold
- Episode boundary triggers (exchange close, compaction)

### Phase 6 — Warm Context Retrieval

Unified retrieval over the whole graph on each incoming impulse. Tags, memories, skills all loaded through the same semantic search into a warm cache. Present renderer structures them into sections. Recently activated nodes weight higher, stale ones drop. Replaces static operator listing in system prompts.

### Phase 7 — Weight Decay and Entity Budget

Deferred from Phase 1. Not yet specced in detail.

- Weight decay on memories over time (half-life, reinforcement resets)
- Entity budget and cap logic
- Entity absorption via resolver when budget is full

## Skill Maturation

A skill is how a being communicates with something. Not the capability itself — the crossing. Bash isn't "run shell commands." Bash is how the being talks to the shell. Browse is how it talks to the internet. The skill is the method of crossing a bridge (port), not the bridge.

### Seeding

Every skill the being starts with — browse, search, bash, recall, remember, plan — is pre-seeded into the memory graph at birth. The skill file (the markdown document describing how to use the skill) is kept as a top-level artifact. Context decomposes it into entities and edges on first load:

- Entities are extracted from the skill description — the nouns of the skill's domain (url, query, command, shell, web page, search results, memory, etc.)
- Entity-to-entity edges are created for co-occurring concepts, typed as skill edges with a ref pointing back to the skill file
- A memory node anchors the skill file's content to those entities

The skill file is the source of truth. The graph is the being's internalized understanding of it. When the being uses the skill, new experience strengthens those edges and adds new memories to the same region. The seeded structure grows from use.

This means a newborn being doesn't start with an empty graph. It starts with the shape of its capabilities already sketched in — light edges, minimal weight, but structurally present. The genome gives identity. The skill seeds give initial graph topology.

### Three Stages

Skills mature through three stages, driven by the same edge-weight mechanism as everything else:

**Skill file** — the starting point. A markdown document describing how to communicate with something, decomposed into the graph as entities and skill edges. Context pulls the skill file during interactions — it's a cache, a pointer from the graph back to the full document. The operator (static function) still handles execution, but the being's understanding of when and how to use it lives in the graph.

**Matured skill** — as the being uses the skill, experience accumulates on the same edges. The seeded entities gain new memories, new co-occurrences, new context. The skill file gets rewritten by Context to reflect what the being has actually learned — not just the original documentation, but patterns, preferences, failures. The cache updates.

**Specialist being** — when the skill region keeps growing and its density crosses the promotion threshold, the skill promotes into a being. It has its own purpose ("I know how to talk to the shell"), its own memory, its own Context/Think/Act. It lives in the inner universe. It grows. A mature bash specialist knows the user's directory structure, knows which commands break, knows patterns. It doesn't look up how — it *knows*.

```
skill file (seeded at birth, decomposed into graph, operator handles execution)
    ↓ experience accumulates on skill edges
matured skill (Context rewrites the file from lived experience)
    ↓ density crosses promotion threshold
specialist being (inner universe, full Self, grows from experience)
```

### Where Skills Live

Skills live on Context. Context looks down into the being's mind — it sees the graph, manages the skill files, and routes to specialist beings when they exist. A skill file is Context's tool. A specialist being is Context's delegate.

### The Difference

The difference between a tool and expertise is that expertise has lived. A skill-as-operator is a tool. A skill-as-being has run a thousand commands, built memory, developed its own specialists. It's not calling an API. It's a being that knows how to communicate with something because it has done it enough times that the knowing became structure.

## Episodic Processing

Context has a heartbeat. Every time an episode closes, Context processes what happened — it sees the full cognitive cycle (what Think thought, what Act did) and updates the graph from it. This is how the being learns from experience, not just from what it's told to remember.

### Exchange Holds the Full History

Exchange keeps two views of the same data:

- **Full history** — every entry from the episode, uncompacted. This is what actually happened.
- **Present view** — the last N entries, rendered into the being's present via the parser. This is what the being sees during conversation.

Compaction is a view concern, not a data concern. The being's present stays manageable. The full history stays intact until Context has processed it. After episodic processing completes, the full history clears.

Exchange no longer owns compaction logic. It owns the data. Context owns what happens to it.

### The Processing Loop

When an episode closes, Context receives the full exchange history and processes it in chunks:

1. **Chunking** — the exchange is split into digestible pieces. Fixed-size chunks to start (N turn-pairs per chunk). Semantic chunking (topic-shift detection) is a future refinement.

2. **Per-chunk extraction** — for each chunk, Context calls its provider:
   - Extract entities mentioned — who, what, where
   - Detect skill usage — which operators were called, how they were used
   - Identify what was learned — new understanding, resolved tensions, patterns
   - Assign weight updates — what got reinforced, what got superseded

3. **Graph updates** — each chunk's output feeds the graph:
   - Entity weights updated based on activation
   - Entity-to-entity edge weights strengthened where co-occurrence happened
   - New memory nodes stored with appropriate types (trace, understanding, etc.)
   - Skill edges updated when skill usage detected

4. **Skill file rewrites** — after all chunks are processed, Context checks skill edge weights. If any skill region crosses the maturation threshold, Context rewrites the skill file:
   - The provider receives the original skill file + all memories anchored to the skill's entities
   - It produces a new skill file that reflects lived experience — patterns, preferences, failures, not just documentation
   - The rewritten file replaces the original. The cache updates.

5. **Cleanup** — full history clears from Exchange. The graph holds what was learned. The entries are gone.

### What Context Sees

Context observes both streams of the cognitive cycle:

- **Think stream** — what the being deliberated on, what operators it called, what it considered before surfacing. This tells Context what mattered internally — what the being paid attention to, what it struggled with, what it recalled.
- **Act stream** — what the being actually said and did. This tells Context what happened externally — which peer was addressed, what was communicated, what actions were taken.

The gap between Think and Act is itself information. A being that thinks deeply about websocket timeouts but says something simple about them is still learning about websocket timeouts. The weight update reflects the thinking, not just the output.

### Maturation Threshold

Skill edges carry weight like all edges. The maturation threshold is the point at which accumulated skill-edge weight triggers a rewrite. Below threshold, the original skill file serves as the cache. Above threshold, Context rewrites it from experience.

This is distinct from the promotion threshold (which triggers specialist creation). Maturation is about the skill file becoming personalized. Promotion is about the skill region becoming dense enough to need its own thinker.

```
skill seeded (weight 0, original file)
    ↓ usage accumulates weight on skill edges
maturation threshold crossed (Context rewrites skill file from experience)
    ↓ continued usage, density grows
promotion threshold crossed (specialist being created in inner universe)
```

### Episode Boundaries

What constitutes an episode boundary:

- **Exchange close** — the natural end of a conversation thread. The being stops talking to a peer. This is the primary trigger.
- **Exchange compaction** — if an exchange runs very long without closing, episodic processing can fire at compaction boundaries (every N entries) even while the exchange is still active. The being processes what it has so far without waiting for the conversation to end.

Both triggers feed the same processing loop. The difference is that exchange close processes the full remaining history and clears it. Compaction processes a batch and keeps the exchange alive.

## Open Questions

### Answered
- ~~How is cluster detection performed?~~ — Neighbor expansion from each entity, density = entity count * (edge weight sum / possible edges) * memory count. Threshold at 75.0.
- ~~What density threshold triggers specialist promotion?~~ — 75.0.
- ~~Should the curator's supersede/complement/contradict judgment be deterministic or LLM-evaluated?~~ — LLM-evaluated. Extractor proposes candidates, curator filters.
- ~~When a specialist fires, does it get the full memory graph or only its cluster's subgraph?~~ — Full shared graph, scoped view via `Scope` field on Context (only traverses claimed entities).
- ~~How does a specialist describe itself to the synthesis layer?~~ — Specialist's Think fires on a relation, returns a string. Context collects all specialist thoughts and synthesizes via provider call.

### Open
- What is the initial entity budget? Does it grow with level?
- What is the weight decay function for memories? Linear, exponential, half-life?
- When a sub-being is spawned from a cluster, does the parent's entity-to-entity edge weight in that region reset to zero, or just decay?
- How does the entity budget interact with the resolver? When budget is full and a new concept arrives, which existing entity absorbs it?
- What are the first specialists a being develops? Are they universal (judgment, strategy, conflict) or do they emerge from the genome's identity/purpose?
- What is the initial trust weight when a relationship is declared in the genome?
- What is the decay half-life constant for trust?
- What constitutes betrayal vs. honest disagreement?
- Can a being that has ended ever return?
- How does trust interact with levels?
- What is the semantic search mechanism for warm context retrieval? Embedding-based, keyword, or entity-match?
