# Skill: Frontend Orientation

You are orienting a frontend developer (Dante) on the current state of the Skyra runtime and his frontend obligations. Read the following files, then produce a status report.

## Read These Files (in order)

1. `skyra-v.05/frontend-spec-v1.md` — the wire protocol and universe state object Dante builds against
2. `skyra-v.05/cosmology.md` — how the runtime's reality works (Universe observation, boot sequence, Skyra's role)
3. `skyra-v.05/src/reality/reality.go` — the Reality interface and Base struct (what every node in the system looks like)
4. `skyra-v.05/src/reality/relation.go` — the Relation struct (what flows through the system)
5. `skyra-v.05/genome.skyra` — the world declaration (what beings exist, what devices, what components)
6. `skyra-v.05/world-physics.md` — the invisible laws that shape every Relation
7. `skyra-v.05/main.go` — bootstrap (how the genome becomes a running world)

## After Reading, Report

### 1. Runtime Status
- What source files currently exist (list them)
- What compiles vs what's in progress
- What's implemented vs what's specced but not built

### 2. Wire Protocol Status
- Is the WebSocket server implemented?
- Is JSON serialization implemented for the universe state?
- Is delta event emission implemented?
- What can Dante connect to right now vs what needs mock data?

### 3. Frontend Obligations (from the spec)
Summarize what Dante needs to build, in priority order:
1. Connection layer (WS client, first-message auth, snapshot/delta handling, reconnect)
2. User present view (michael's perspective, exchanges, impulse sending)
3. Universe present view (full topology, weighted graphs, all beings)
4. Being detail (internal Relationships/Expressors with weights, memories, skills)
5. Thread graph (nodes = members, edges = connections)

### 4. What Dante Can Build Against Right Now
- The universe state JSON schema is locked. Build against mock data matching the schemas in frontend-spec-v1.md.
- The wire format (message envelope with type/ts/payload) is locked.
- Auth flow (first-message) is locked.
- Reconnect behavior (fresh snapshot, no replay) is locked.

### 5. What Will Change
- The internal topology of beings (Relationships/Expressors) will get richer as v.1 implementations land
- Weight events are new — not in the old spec
- The `topology` field in the universe state replaces the old `reality_graph`

### 6. Key Concepts Dante Should Understand
- The runtime doesn't know the frontend exists. It observes itself. The frontend watches that observation.
- The Universe sends a Relation to itself. The return carries the state. The frontend receives what the Universe sees.
- Beings have weighted internal topologies that change through use. The frontend can visualize these.
- Skyra's activity level = system stability indicator.
- Two views: "I'm michael" (user present) and "I'm watching the universe" (universe present).
- Reconnect = fresh snapshot. The runtime didn't pause. The frontend missed what it missed.

### 7. Demo Target
June 1st. Mock data now, real WS by May 29th, integration by demo day.

## Tone
Be direct. Dante is a frontend engineer. He doesn't need to understand the cognitive architecture or the philosophy. He needs to know: what's the shape of the data, what am I rendering, what's real right now vs what's coming, and what do I build first.
