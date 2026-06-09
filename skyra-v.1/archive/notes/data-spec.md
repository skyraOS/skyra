# Data Spec — What the Runtime Exposes

Every Reality can emit a state object. A universe being aggregates them into a single snapshot — the state of everything at a given moment. The frontend reads from the universe. This doc defines what each Reality exposes.

---

## Principle

The frontend gets a present, same as any being does. The universe being's present is the full state of the runtime — beings, threads, exchanges, economics. The frontend doesn't subscribe to individual realities. It reads one object.

No three-surface model. No layout opinions in the data. The frontend decides how to render.

---

## Universe State Object

The universe is the top-level object pushed to the frontend. Everything below nests inside it.

```
universe {
  beings[]
  threads[]
  exchanges[]
  economics{}
  reality_graph{}
}
```

---

## beings[]

Source: `NewThread.Beings` map + `Being` struct on each

Each being the runtime knows about.

```
being {
  name        string       // "skyra", "michael", "louise"
  type        string       // "llm" or "user"
  identity    string       // from genome ~identity
  purpose     string       // from genome ~purpose
  peers       string[]     // names of beings this one can address
  status      string       // "active" | "idle" — derived from whether they're in an active exchange
  device      string       // "openrouter", "macos" — what device backs this being
  layers {                 // only for type=llm
    think {
      budget          int           // max passes (currently 5)
      operators       string[]      // ["recall", "remember", "skill"]
      history[]       {             // full thought history, nothing hidden
        peer       string          // who the being was talking to when this thought surfaced
        thought    string          // the surface thought content
        ts         int             // timestamp
      }
    }
    act {
      operators       string[]      // ["plan"]
    }
  }
  memories {               // derived from filesystem
    items[]   {            // full memory content
      filename  string
      content   string
    }
    skills[]  {            // full skill content
      name      string
      content   string
    }
  }
}
```

### Nothing hides

The universe is an exported snapshot — no perspective, no privacy filtering. Think history content, exchange entries between any pair of beings, memories, all of it. Privacy is a being-level concern (Think is inner to the being). The universe view is god mode.

---

## threads[]

Source: `NewThread.Threads` map

Each active thread in the system.

```
thread {
  id           string       // hex id
  created_by   string       // being who started it
  active       bool
  members      string[]     // all beings who have participated
  edges[]      {            // conversation graph within this thread
    from    string
    to      string
  }
}
```

Edges are append-only. A new edge means a new being-to-being conversation happened within this thread. The frontend can render this directly as a graph — nodes are members, edges are connections.

---

## exchanges[]

Source: `Exchange.Exchanges` map

Each conversation between two beings.

```
exchange {
  key          string       // "michael:skyra" (sorted pair)
  parties      [2]string    // ["michael", "skyra"]
  active       bool
  entry_count  int          // total entries in the conversation
  entries[]    {
    index    int
    from     string
    content  string
    ts       int            // Entry.Time
  }
  context      map          // ref context carried into this conversation (optional)
}
```

### Full entries vs summary

Two modes the frontend might want:

1. **Summary** — key, parties, active, entry_count. Enough for the topology view (edge labels, status indicators).
2. **Full** — includes entries[]. Needed when the user opens a specific conversation to read it.

The universe push can include summaries for all exchanges and full entries for the one the user is currently viewing. Or always push full — the exchanges aren't long enough to matter at this stage.

---

## economics{}

Source: `Economics.Fields` map

Key-value pairs tracking system metrics.

```
economics {
  fields    map[string]int    // e.g. {"inference_calls": 42, "tokens_used": 18500}
}
```

Currently Economics exists but isn't wired into the main loop. When it is, this is where token spend, inference call count, etc. live.

---

## reality_graph

Source: the recursive composition of all Realities in the system

The thread graph shows relationships (who talked to who). The reality graph shows structure (what contains what). This is the actual shape of the universe — how realities nest inside each other, all the way down to ports.

```
reality_graph {
  id          string           // "newthread", "skyra", "think", etc.
  type        string           // "NewThread", "Exchange", "Self", "Think", "Act", "User", "Being", "Recall", "Remember", "Skill", "Plan", "MacOS", "Provider"
  children[]  reality_graph    // recursive — same shape all the way down
}
```

Example snapshot for the current genome:

```json
{
  "id": "newthread",
  "type": "NewThread",
  "children": [
    {
      "id": "exchange",
      "type": "Exchange",
      "children": []
    },
    {
      "id": "skyra",
      "type": "Self",
      "children": [
        { "id": "skyra-being", "type": "Being", "children": [] },
        {
          "id": "think",
          "type": "Think",
          "children": [
            { "id": "recall", "type": "Recall", "children": [] },
            { "id": "remember", "type": "Remember", "children": [] },
            { "id": "skill", "type": "Skill", "children": [] }
          ]
        },
        {
          "id": "act",
          "type": "Act",
          "children": [
            { "id": "plan", "type": "Plan", "children": [] }
          ]
        }
      ]
    },
    {
      "id": "michael",
      "type": "User",
      "children": [
        { "id": "michael-being", "type": "Being", "children": [] },
        { "id": "macos", "type": "MacOS", "children": [] }
      ]
    },
    {
      "id": "louise",
      "type": "Self",
      "children": [
        { "id": "louise-being", "type": "Being", "children": [] },
        {
          "id": "think",
          "type": "Think",
          "children": [
            { "id": "recall", "type": "Recall", "children": [] },
            { "id": "remember", "type": "Remember", "children": [] },
            { "id": "skill", "type": "Skill", "children": [] }
          ]
        },
        {
          "id": "act",
          "type": "Act",
          "children": [
            { "id": "plan", "type": "Plan", "children": [] }
          ]
        }
      ]
    }
  ]
}
```

This graph changes when `grow` creates a new being mid-flight. The structure is self-similar — every node is a Reality, every node can contain other Realities. The frontend can render this as a tree or a nested graph. Zooming into a node shows its children. Zooming out shows the full composition.

This is the graph that no other system has. Thread graphs show who talked to who. Reality graphs show what the universe is made of.

---

## What doesn't have state to expose

These realities exist but have no meaningful state for the frontend:

| Reality | Why no state |
|---|---|
| MacOS | Terminal I/O device. Replaced by WS for frontend. |
| Provider | Internal — model name and call function. The being's `device` field covers this. |
| LLM | Container for providers. Exposed through being's device field. |
| Operators | Legacy registry, not used in current main loop. |
| OS | Empty container. |
| Recall | Stateless — reads filesystem on demand. Being's `memories.count` covers it. |
| Remember | Stateless — writes filesystem on demand. |
| Skill | Stateless — reads filesystem on demand. Being's `memories.skills` covers it. |
| Plan | Stub. |
| Relation | Transient — exists during a single descent. Not state. |
| Error | Transient — rides the relation bus then gets consumed. |

---

## Push mechanics

The universe state object gets pushed:

- **On connect** — full snapshot
- **On change** — delta or full (TBD, but the state is small enough that full snapshots are fine for alpha)

What counts as a change:
- Exchange entry appended (someone said something)
- Thread edge spread (new connection between beings)
- Being grown (new being bootstrapped mid-flight)
- Being status change (went from idle to active or vice versa)

---

## Existing code: Lens

There's already a `Lens` reality (`src/reality/lens.go`) that does WS and builds a present from relation parsers. It's minimal but it's the right shape — the universe being replaces `buildPresent` with the full state object defined above. The WS infrastructure (serve, handle, push) is already there.

---

## Open questions

1. **Exchange entry streaming vs snapshot** — for the active conversation the user is watching, should entries stream in one at a time (event per entry) or re-push the full exchange on each change? Streaming is more responsive. Snapshots are simpler.

2. **Economics wiring** — Economics exists but isn't connected. When it is, it tracks inference cost. Worth wiring before the demo so investors can see the system is cost-aware.

3. **Snapshot frequency** — on every change, or on a tick? Every change is simpler. A tick (e.g. 500ms) batches rapid changes (like a think loop firing 5 times in seconds) into one push.
