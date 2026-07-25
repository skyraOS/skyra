# Runtime Response to Frontend Spec

Dante — read through your spec. It's clean and the core principle is right: the frontend renders the runtime, it doesn't model it. Here's where the runtime actually is and how your spec maps to it.

---

## 0. The thing to understand first

The runtime exposes **N+1 graphs**.

**N thread graphs** — one per thread. Each thread has its own members and edges showing who talked to who within that thread. A user can have multiple threads running simultaneously, each with a different topology.

**1 reality graph** — global, singular. The recursive composition structure of the entire universe — what contains what. NewThread contains beings, each being contains Think and Act, Think contains operators (Recall, Remember, Skill), all the way down. This graph changes when a new being is grown mid-flight.

Your spec has one graph (LogosMap). The runtime has many. The thread graphs are the relationship layer — who is talking to who. The reality graph is the structural layer — what the universe is made of. The frontend renders all of them.

---

## 1. Vocabulary mapping

Your spec uses different names for the same things. Let's lock these now so the Zod schemas match the Go structs.

| Frontend spec | Runtime term | What it is |
|---|---|---|
| Logos | **Being** | An entity in the system — LLM-backed or user-backed. Has a name, description, relationships. |
| Logos (world/operator variant) | **Reality** | The core interface. Everything implements it: `ID()`, `Create(*Relation)`, `Realize(*Relation) string`. Beings, exchanges, threads, devices — all Realities. |
| LogosMap | **Thread graphs** | Each Thread has its own graph — members and edges (who talked to who within that thread). Multiple threads means multiple topology graphs, each with different participants and connections. `NewThread.Threads map[string]*Thread`. |
| (no equivalent) | **Reality graph** | One global graph — the recursive composition structure of the entire universe. What contains what. Every Reality can contain other Realities. NewThread contains Self, Self contains Think and Act, Think contains Recall/Remember/Skill. No other system exposes this. |
| Relation (your spec) | **Relation** | Same name, same concept. The message bus. Carries: `Origin` (who sent it), `ID` (who it's going to), `ThreadID`, `Impulse` (the message content), `Parsers` (context attachments), `Realities` (available beings/errors). |
| RelationStream | **Exchange / Conversation** | `Exchange` holds `Exchanges map[string]*Conversation`. Each Conversation has `Parties [2]string`, `Entries []Entry`, `Active bool`. Entries are append-only — `{From, Content}`. This is your stream. |
| Shortcut | Not yet in runtime | Product concept. We'll get there. |

**Recommendation:** Use the runtime names in the protocol. Your React components can be called whatever you want (`<Logos />` is fine as a component name), but the JSON on the wire should say `being`, `relation`, `conversation`, `thread` — not `logos`, `logos_map`, etc. One vocabulary across the stack.

---

## 2. Runtime architecture (what you're rendering)

The runtime is a recursive descent. A relation enters at the top and descends through layers, each layer adding context or routing, until it hits a being and comes back up.

```
User input
  → NewThread (creates/finds thread, checks access, detects system ops like `grow`)
    → Exchange (finds/creates conversation, enforces context crossing via <ref> tags, records entries)
      → Self (the being's container — holds inner + outer layers)
        → Think (inner layer — private thought, operators: recall/remember/skill, 5-pass budget, persistent thought history)
        → Act (outer layer — protocol enforcement, emits <target>message</target>, routes to peers)
      ← response string
    ← response string (recorded as entry)
  ← NewThread loops if the response targets another being
```

Key thing for the frontend: **NewThread owns the loop.** When skyra talks to michael, then michael's response goes to skyra, then skyra decides to talk to louise — that's all one loop iteration in NewThread. The frontend doesn't need to manage this. It just receives events as they happen.

**N+1 graphs, one universe.** Each thread has its own topology graph — who talked to who within that thread. Multiple threads means multiple relationship graphs, each with different participants and edges. On top of that, one global reality graph shows the composition structure — what contains what, all the way down. The frontend renders all of them.

---

## 3. What the runtime exposes — the universe state

The frontend receives the full state of the universe. Every Reality emits its state. Nothing is hidden — this is god mode. Privacy is a being-level concern; the universe view has no perspective.

The top-level object pushed to the frontend:

```
universe {
  beings[]
  threads[]
  exchanges[]
  economics{}
}
```

### beings[]

Source: `NewThread.Beings` map + `Being` struct + `Self`/`Think`/`Act` layers on each

```json
{
  "name": "skyra",
  "type": "llm",
  "identity": "I hold the world together.",
  "purpose": "I think, respond, and relate on behalf of the system.",
  "peers": ["michael", "louise"],
  "status": "active",
  "device": "openrouter",
  "layers": {
    "think": {
      "budget": 5,
      "operators": ["recall", "remember", "skill"],
      "history": [
        {
          "peer": "michael",
          "thought": "he's asking about memory architecture — I should talk to louise about this",
          "ts": 1714650000
        }
      ]
    },
    "act": {
      "operators": ["plan"]
    }
  },
  "memories": {
    "items": [
      { "filename": "1714650000.md", "content": "michael prefers direct answers" }
    ],
    "skills": [
      { "name": "code-review", "content": "when reviewing code, focus on..." }
    ]
  }
}
```

Full thought history content, full memory content, full skill content. The frontend sees everything a being has thought, remembered, and learned.

### threads[]

Source: `NewThread.Threads` map

```json
{
  "id": "a1b2c3d4",
  "created_by": "michael",
  "active": true,
  "members": ["michael", "skyra", "louise"],
  "edges": [
    { "from": "michael", "to": "skyra" },
    { "from": "skyra", "to": "louise" }
  ]
}
```

Edges are append-only. A new edge means a new being-to-being conversation happened within this thread. The frontend renders this directly as a graph — nodes are members, edges are connections.

### exchanges[]

Source: `Exchange.Exchanges` map

```json
{
  "key": "michael:skyra",
  "parties": ["michael", "skyra"],
  "active": true,
  "entries": [
    { "index": 0, "from": "michael", "content": "what do you think about memory?", "ts": 1714650000 },
    { "index": 1, "from": "skyra", "content": "memory is what compounds", "ts": 1714650005 }
  ],
  "context": {
    "skyra": "context brought from louise (entries 0-3):..."
  }
}
```

Every exchange between every pair of beings — not just the user's conversations. The frontend can show skyra:louise exchanges that the user (michael) never directly participated in.

### economics{}

Source: `Economics.Fields` map

```json
{
  "fields": {
    "inference_calls": 42,
    "tokens_used": 18500
  }
}
```

Not yet wired into the main loop. When it is, this tracks token spend, inference call count, etc.

### Error

```go
type Error struct {
    Message string
}
```

Errors ride the relation bus via `r.Realities["error"]`. When Exchange blocks a context crossing or a being isn't found, this is what surfaces. Errors are recoverable — the runtime routes them back to the origin being for correction.

---

## 4. Proposed message envelope (adjusted)

Two modes:

### Full snapshot — on connect and reconnect

The entire universe state object as defined above. One push, everything the frontend needs to render from zero.

```json
{
  "id": "string",
  "type": "universe",
  "ts": 0,
  "payload": {
    "beings": [...],
    "threads": [...],
    "exchanges": [...],
    "economics": {...},
    "reality_graph": {...}
  }
}
```

### Deltas — on change

Individual events as they happen:

```json
{
  "id": "string",
  "type": "entry" | "edge" | "being" | "thought" | "memory" | "error",
  "ts": 0,
  "payload": {}
}
```

- `entry` — new exchange entry appended (from, content, conversation key, index)
- `edge` — new edge in thread graph (from, to, thread_id)
- `being` — new being grown mid-flight (full being snapshot)
- `thought` — new thought surfaced in Think (being name, peer, thought content)
- `memory` — new memory written via remember (being name, filename, content)
- `reality` — reality graph changed (new being grown adds a subtree)
- `error` — system error (origin, message). Errors are recoverable — the runtime routes them back to the origin being for correction.

### reality_graph — the composition structure

The reality graph shows what the universe is made of — how realities nest inside each other. Self-similar recursive structure: every node is a Reality, every node can contain other Realities.

```json
{
  "id": "newthread",
  "type": "NewThread",
  "children": [
    { "id": "exchange", "type": "Exchange", "children": [] },
    {
      "id": "skyra", "type": "Self",
      "children": [
        { "id": "skyra-being", "type": "Being", "children": [] },
        {
          "id": "think", "type": "Think",
          "children": [
            { "id": "recall", "type": "Recall", "children": [] },
            { "id": "remember", "type": "Remember", "children": [] },
            { "id": "skill", "type": "Skill", "children": [] }
          ]
        },
        {
          "id": "act", "type": "Act",
          "children": [
            { "id": "plan", "type": "Plan", "children": [] }
          ]
        }
      ]
    }
  ]
}
```

This is the thread graph (who talked to who) plus the reality graph (what contains what). Two graphs, one universe. The frontend can render both — topology view for relationships, tree/nested view for structure.

The delta types map directly to the universe state sections. The frontend applies each delta to its local copy of the universe. On reconnect, full snapshot resyncs everything.

The state is small enough that full snapshots on every change would work too for alpha. Deltas are the clean path.

---

## 5. Answers to your open questions

**Auth on the WS connection** — First message. Token in query string leaks to logs. Header is cleaner but some WS clients don't support custom headers. First-message auth with a handshake response before any state flows is the cleanest.

**Delta vs snapshot for topology** — Deltas. The thread graph is append-only (edges and members only add, never remove). Full snapshot on initial connect and on reconnect. No periodic snapshots needed.

**Backpressure** — Server buffers. The runtime is not high-throughput — beings think for seconds at a time. A slow frontend won't cause pressure. If it does later, we drop duplicate snapshots (keep latest), never drop relation events.

**Reconnect semantics** — Server replays missed relation events since the client's last acknowledged `id`. Thread graph gets a full snapshot on reconnect. This means the frontend needs to send `ack` for relation events and track its last received `id`.

---

## 6. What doesn't exist yet (backend work needed)

1. **WebSocket server** — The runtime is currently in-process only (stdin/stdout via MacOS device). We need a WS device that implements the Reality interface and bridges to the frontend. This replaces MacOS as the user's device. A `Lens` reality with WS infrastructure already exists (`src/reality/lens.go`) — needs to be rewritten to emit the universe state object.

2. **Event emission** — Currently the runtime doesn't emit events to an observer. We need hooks at: Exchange (entry appended), NewThread (edge spread, being grown), Self (think/act transitions — optional, but useful for showing "being is thinking" state).

3. **Serialization** — No JSON serialization exists for any runtime struct. Straightforward but needs to be written.

~~Timestamps on entries~~ — Done. `Entry` struct now has `Time time.Time`, set on append.

~~Being device field~~ — Done. `Being` struct now has `Device string`, set from `~device` in genome.

The WS device is the big one. Everything else is small.

---

## 7. Two views, not three regions

Your spec describes three regions: user-owned, being-held, overlap zone. The runtime doesn't work that way. There are two views:

**User present** — michael's view as a being in the system. His active exchange, his conversation entries, his peers, what's in front of him right now. This is the same present the runtime builds for any being — scoped and perspectival. The user posts relations, the being responds, entries are append-only. No turns, no request/response. Michael can post three times in a row. Skyra can talk to louise for ten exchanges before coming back.

**Universe present** — god mode. Every being with full inner state (thoughts, memories, skills), every thread graph, every exchange between any pair of beings, the reality graph showing composition structure, economics. No perspective. No privacy. The full state of everything.

The user toggles between "I'm michael" and "I'm watching the universe." One is a being's view. The other is the view from nowhere.

The three-region model collapses into the user present — it's just one being's exchange rendered however the frontend wants. The universe present is the new thing your spec didn't have.

---

## 8. Your timeline from the backend side

| Week | What we'll have ready |
|---|---|
| 1 (May 1–7) | Message schemas locked (this doc is the start). Runtime struct serialization. |
| 2 (May 8–14) | WS device implemented. Event hooks on Exchange entries. You can connect and receive live relation events from a real runtime. |
| 3 (May 15–21) | Thread delta events. Being snapshot on connect. Full reconnect/replay. Real beings in staging. |
| 4 (May 22–28) | Stable. Bug bash coordination. |

This lines up with your spec — you build against mock in weeks 1–2, we give you real WS in week 2, you cut over in week 3.

---

## 9. One thing to flag

Your spec says "the interaction surface is two-sided." In the runtime, it's potentially multi-sided. Skyra can be in conversations with michael AND louise simultaneously. The frontend needs to handle: which conversation is the user looking at? The LogosMap (thread graph) node selection → RelationStream flow you described handles this naturally. Just be aware that a being's "state" isn't one conversation — it's all of their active conversations.
