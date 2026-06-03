# Godot Relation Frontend Mechanism

Skyra's frontend should not own the world. The `Relation` owns the world.

Godot's job is to render the current relation, show what matters, and send user
acts back into the runtime. The Go process remains the source of truth for
state, traversal, threads, slices, inference, and logs.

## Core Boundary

```text
Go Skyra Runtime                         Godot Frontend
----------------                         --------------
Relation                                 Scene tree
World map                                Entity nodes
Threads                                  Edges / bindings
Fields                                   Visual weights
Slices                                   Per-being panels
Impulse                                  Input focus / event surface
Traversal logs                           Motion / highlights / timelines
Provider calls                           Thinking / pending states
```

The frontend is a projection of the relation. If Godot disappears and
reconnects, it should be able to rebuild itself from the latest relation
snapshot.

## Why Godot Fits

Godot is useful when the interface is a living state space instead of a static
dashboard. Skyra's world has entities, relationships, activations, threads,
attention, memory, and traversal. Those map naturally to a scene graph:

- A `Reality` becomes a visible node.
- A `Thread` becomes a binding edge, orbit, trail, layer, or channel.
- A `Slice` becomes a private state panel for a being.
- `Fields` become visual force, brightness, proximity, edge weight, or gating.
- `Impulse` becomes the currently active input/event moving through the world.
- Traversal events become animation.

Godot is too heavy if the frontend only needs chat, tables, and forms. It is a
good fit if the frontend should feel like the relation is alive.

## Activation-Based Space

The end goal is not a literal map where a being occupies one physical location.
Skyra space should be activation-based.

This projection is for global relation state, not the user slice. The spatial
field should show the shared world: entities, threads, fields, activation,
traversal, and co-activation. A `Slice` is private per-being state. It can be
inspected or summarized in the frontend, but it should not be treated as the
global map.

In this model, space is a projection of relevance:

```text
distance   = relational / activation distance
brightness = current importance
overlap    = shared attention or co-activation
motion     = changing relation state
```

A being does not stand in one place. A being emits an activation field. Its hot
zone is what it currently cares about, remembers, is bound to, or is traversing.
When beings overlap hot zones, that overlap is where relation happens.

This makes multiple engagement and multiple subscriptions natural:

- Multiple engaged nodes become several active peaks.
- Multiple subscriptions become persistent field influences.
- A focused being has a tight hot zone.
- A diffuse being has a wider cloud.
- Shared attention appears as overlapping heat.

The frontend should avoid implying that the being has one canonical location.
It should instead show weighted relation, attention, and co-activation.

Useful mappings:

- `Slice.Surroundings` -> what the being currently sees
- `Thread.Members` -> subscribed or bound attention
- `Relation.Fields` -> field modifiers and gates
- `Reality.Activation(rel)` -> heat, size, brightness, or pull
- traversal enter/exit -> pulse through the field
- thoughts -> temporary marks or particles in the field
- goal/intent -> attractors that shape the field

Godot is viable here because it can render a live field instead of a literal
room. The coordinates are presentation, not truth. The truth remains the
relation graph and its activation values.

## Agent Layer And Memory Pane

The 3D scene should separate agents from the memory/relation surface.

Agents do not live in memory space. They project activation into it.

Use two coupled spaces:

```text
Agent layer
  who is acting, observing, listening, thinking, speaking

Memory pane
  what is currently active in global relation/memory state
```

Agents can be placed above, beside, or around the pane. Their position is a
view/control anchor, not their literal location in memory. The user can move
agents around if useful, because placement belongs to the projection layer.

Each agent should have a stable color or visual theme. That same color appears
on the memory pane where the agent is currently activating, traversing, or
subscribed.

```text
Skyra agent: blue-white
  blue-white glow on pane = Skyra activation

Michael/user agent: warm gold
  gold glow on pane = user-originated activation

Multiple agents touching the same area
  blended or layered glow = co-activation
```

The memory pane is not all memory. It is the currently active global surface:
active nodes, recent nodes, subscribed regions, nearby context, and relation
hot zones. Inactive memory can sink, dim, or remain outside the current pane
projection.

The pane should feel like shallow 3D relief rather than a flat graph:

- hot nodes rise upward
- active regions glow outward
- threads appear as arcs, currents, or rings
- traversal appears as pulses across the surface
- stale memory sinks or fades
- overlapping agent fields create blended light pools

This creates three visual layers:

```text
above / beside pane
  agents, beings, observers

on pane
  active memory nodes and relation surfaces

between them
  attention beams, projected fields, traversal paths
```

The agent layer can expose per-being state such as idle, listening, traversing,
thinking, speaking, signal, load, and subscriptions. The pane remains the global
relation projection.

```text
Relation
  actual state

Activation Field
  computed projection of importance

Spatial Layout
  visual coordinates derived from activation, threads, and similarity

Godot Scene
  animated rendering of that projection
```

## Data Flow

Godot should receive state in two modes:

1. Initial snapshot

   Sent when the frontend connects or reconnects. It contains enough state to
   rebuild the scene.

2. Incremental events

   Sent while the system runs. These are smaller patches that explain what just
   changed.

```text
Godot connects
  -> Go sends relation.snapshot
  -> Godot creates scene nodes

User acts in Godot
  -> Godot sends input/event to Go
  -> Go updates Relation
  -> Go emits relation events
  -> Godot updates visuals
```

The frontend should not poll the entire relation every frame. It should apply
events and occasionally request a full resync.

## Transport

Use WebSocket for the live path.

HTTP can still serve:

- `/api/state` for a full snapshot
- `/api/logs` or file-backed debug views
- health/status endpoints

WebSocket should carry:

- relation snapshots
- entity updates
- thread updates
- slice updates
- traversal events
- provider call lifecycle events
- user/device input

## Event Shapes

The protocol should stay small and typed.

```json
{
  "type": "relation.snapshot",
  "relationId": "abc",
  "origin": "michael",
  "impulse": "hello",
  "entities": [],
  "threads": [],
  "slices": []
}
```

```json
{
  "type": "traversal.enter",
  "beingId": "skyra",
  "nodeId": "skyra",
  "activation": 1.0,
  "competence": "propagate"
}
```

```json
{
  "type": "thread.field",
  "threadId": "thread:skyra:main",
  "nodeId": "skyra",
  "alignment": 1.0,
  "strength": 0.92
}
```

```json
{
  "type": "slice.update",
  "beingId": "skyra",
  "signal": 0.98,
  "load": 421,
  "thoughtCount": 3,
  "goal": "Maintain system coherence",
  "intent": "<speak>...</speak>"
}
```

```json
{
  "type": "device.input",
  "beingId": "michael",
  "content": "what do you see?"
}
```

## Importance Model

Godot should not invent importance. It should render importance from the
relation.

Useful signals:

- `Reality.Activation(rel)`
- `Skyra.IdentityWeight`
- `Skyra.FamiliarityWeight`
- `Recency`
- `Thread.Members`
- `Relation.Fields`
- `Slice.Signal`
- `Slice.Load`
- `Slice.Thoughts`
- current traversal enter/exit events
- provider pending/result/error events

Visual rules can be frontend-owned, but the inputs should come from Go.

Examples:

- High activation means brighter, larger, or closer.
- Thread strength means thicker edge or stronger orbit.
- Low signal means fading or reduced motion.
- Provider pending means pulsing/thinking state.
- A newly traversed node gets a short highlight.

## Ownership Rules

Go owns:

- relation mutation
- routing
- activation
- traversal
- threads
- slices
- inference
- command handling
- persistence/debug logs

Godot owns:

- scene layout
- animation
- camera
- hit testing
- visual filtering
- user input affordances
- local-only display preferences

Shared protocol owns:

- event names
- JSON schemas
- reconnect behavior
- versioning

## Minimal Implementation Path

1. Keep the existing `/api/state` snapshot.
2. Add a WebSocket endpoint, probably `/api/events`.
3. Emit typed traversal events from the same places that currently call
   `debug.Log` and `debug.Being`.
4. Add a small Godot client that connects, requests a snapshot, and renders
   entities as nodes.
5. Render threads as edges.
6. Render slices as side panels.
7. Send Godot input back as `device.input`, then route it through the same path
   as terminal input.

## Design Constraint

The frontend should be replaceable.

Godot is a strong client for a world-like frontend, but the mechanism should not
depend on Godot. Any client that can consume snapshots and events should be able
to render the relation. This keeps Skyra's architecture clean:

```text
Relation first.
Projection second.
Client last.
```
