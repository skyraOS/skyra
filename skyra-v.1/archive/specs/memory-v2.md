# Memory v2

## What Changed

Memory v1 was operator-based: `remember` as a standalone Reality the inner being routes to. That stays. What changes is the underlying architecture — a graph that makes memory structural, relationship-scoped, and capable of growing new beings.

The v.01 two-tier trust model (observational vs committed, user-approved promotion) is gone. Beings own their own memory. No approval gate. The economics layer (levels, XP) governs weight — not a user.

---

## Core Principle

Memory is a graph. Nodes are things the being knows. Edges are how those things relate. Both carry weight. Weight is earned through exchange, not assigned.

A being accumulates memory through experience. When a neighborhood of that graph coheres — when a cluster of nodes and edges becomes dense enough to stand on its own — the being grows a new being. A new neighbor. `grow` is how memory becomes alive.

Truth is derived, not stored. There is no "current state" field. The being reasons over edges, weights, and history to conclude what is true now.

---

## Retrieval Structure

Two layers of control:

**Layer 1 — Relationship.** Memory is scoped by relationship, same as desk. When a being thinks about a conversation with michael, it sees memory from the michael neighborhood. When it thinks about philosopher, it sees that neighborhood.

**Layer 2 — Entity.** Within a relationship, entities are the anchors. Entities are the things that keep coming up — concepts, tools, patterns, decisions. They emerge from conversation and become the pegs memories hang on. Retrieval doesn't return "all memories with michael" — it returns "memories with michael about this thing."

```
michael ↔ skyra (relationship)
  │
  ├── websockets (entity, weight: 0.8)
  │   ├── mem: "michael wants browser bridge"
  │   ├── mem: "shipped WS component on may 2"
  │   └── mem: "port 8080 default"
  │
  ├── task economy (entity, weight: 0.9)
  │   ├── mem: "desk holds tasks by relationship"
  │   ├── mem: "acceptance routing: giver accepts, not the being"
  │   └── mem: "plan is inner operator on think"
  │
  └── routing (entity, weight: 0.5)
      └── mem: "subscribe pattern, never mutate relation"
```

Entities earn their weight through frequency and recency. High-weight entities surface first. Low-weight ones fade. Passive mode extracts entities from the current impulse, matches them against the relationship's entity anchors, and only pulls those neighborhoods. The impulse itself selects which slice of memory is relevant.

Entities can appear across multiple relationships. An entity exists once in the graph with edges into multiple relationship neighborhoods. Cross-relationship recall is possible but scoped by default.

---

## Graph Structure

### Nodes

Every node has: id, type, content, vector, weight, created_at, last_seen_at.

**entity** — a named thing. Person, place, concept, tool. Extracted from exchange. Aliases resolve to one canonical id. Entities are the anchor layer — memories hang off them, retrieval routes through them.

**memory** — a retained artifact. The four-member family, unchanged since v.02:

- **trace** — what happened. Factual, non-interpretive. The grounding artifact. Lowest weight. First to compress.
- **salience** — what mattered. What carried weight or attention. Mid-weight. Survives one compression cycle before becoming understanding or being discarded.
- **tension** — what remained unresolved. Open edges, conflicting signals, incomplete. High weight. Persists until resolved.
- **understanding** — what it meant. Derived interpretation. Highest weight. Survives indefinitely.

Core distinction: trace preserves occurrence. The other three preserve consequence. All four are derived from exchange, but trace stays on the factual side of the line.

Every memory node carries:
- **anchor_set** — entity ids and relationship pairs that ground the artifact structurally. This is how retrieval finds it.
- **context_artifacts** — references to earlier artifacts that shaped this one's formation. An understanding points back to the traces it was derived from. This preserves the layered nature of experience.
- **trust_at_formation** — the being's cognitive trust judgment at the moment the artifact was formed. Not copied from a relationship, not averaged — a snapshot of interpretive posture, frozen permanently on the artifact. Later trust movement doesn't change it.

**fact** — a piece of knowledge the being holds as true. Derived from understanding nodes that stabilize.

### Edges

Every edge has: from, to, type, weight, history[], created_at, last_seen_at.

Edges carry weight histories. A weight shifting from 0.9 to 0.4 over time is data — it tells the being how a relationship is changing.

Two nodes can have multiple edges between them. Relationships are rarely one thing.

| type | meaning |
|---|---|
| `relates_to` | general semantic relationship |
| `part_of` | entity is part of another |
| `causes` | one thing leads to another |
| `alias_of` | name resolves to canonical entity |
| `derived_from` | understanding derived from traces/salience |
| `mentions` | memory mentions an entity |
| `in_relationship` | entity belongs to a relationship neighborhood |

Edge types are not exhaustive. New types emerge as relationships are observed.

---

## Neighborhoods and Growth

A neighborhood is a subgraph scoped to a relationship. Dense neighborhoods — many nodes, many edges, high aggregate weight — represent deep knowledge.

When a neighborhood coheres into something that can stand independently:

```
exchange → memory accumulates → neighborhood densifies
  → entities cluster, edges strengthen, understandings stabilize
  → being recognizes: this neighborhood can stand on its own
  → being grows a new being (grow command)
  → new being = new neighbor in the graph
```

The economics layer (XP, levels) provides the developmental signal. A being with more exchange, more resolved tasks, more accumulated weight has the substrate to grow meaningfully. Growth is not gated by approval — it is gated by having something real to grow from.

---

## The Three Layers

```
Memory (long-term — graph, all relationships, all history)
    ↓
Context (medium-term — managed window, current relationship, LLM-curated)
    ↓
Think (short-term — single exchange, budget of 5 passes)
```

### Memory (long-term)

The full graph. Every artifact the being has ever retained, across all relationships. Persisted to disk. Never fully loaded — too large. Think never queries it directly.

### Context (medium-term)

The working memory layer. A managed window between Memory and Think. Context holds the slice of Memory relevant to the active relationship — loaded entity neighborhoods, recent understandings, active tensions.

Context is managed by an LLM because deciding what's relevant is a judgment call. It's not "most recent" — it's "what matters for this conversation right now." Context watches the impulse, matches against entity anchors, loads relevant neighborhoods from Memory, and evicts what's stale.

When the relationship shifts (michael → philosopher), Context swaps its window.

Context is like RAM. Memory is disk. Think is the register file.

### Think (short-term)

Think sees what Context has loaded. It doesn't reach into Memory. The inner operators (remember, recall) write to and read from Context, which syncs back to Memory.

---

## Operators

### remember (write)

The being retains an artifact during thinking. Writes to Context, which syncs to Memory.

```
<remember>
  <type>understanding</type>
  <content>michael prefers direct answers without preamble</content>
  <entities>michael</entities>
</remember>
```

1. Extract entities from content (rule-based: dictionary + regex + title-case)
2. Resolve entities against alias table (normalize, fuzzy match)
3. Create memory node with anchor_set, context_artifacts, trust_at_formation
4. Create edges: memory → entities (mentions), memory → relationship neighborhood (in_relationship)
5. Update entity weights based on frequency and recency
6. Node lives in Context immediately, syncs to Memory graph

### recall (read)

The being queries memory during thinking. Reads from Context first, reaches into Memory if Context doesn't have enough.

```
<recall>
  <about>websocket architecture</about>
  <type>understanding</type>
</recall>
```

1. Extract entities from query
2. Resolve to canonical forms
3. Search Context window first — already-loaded neighborhoods
4. If insufficient, query Memory graph: semantic search (vector similarity) + graph traversal (neighborhood walk)
5. Rank results by: weight, recency, type priority (understanding > tension > salience > trace)
6. Load results into Context for future passes

Filters compose. Without filters, relevance is weighted by type and recency.

### Context management (passive)

Context runs on every relation, before Think fires. It is the passive layer.

1. Extract entities from current impulse
2. Match against entity anchors in the current relationship's neighborhood
3. Load matching neighborhoods from Memory into the Context window
4. Evict stale neighborhoods that haven't been referenced recently
5. Attach loaded context as parser — Think sees it in its present without asking

---

## Compression

When memory grows too large, the being compresses.

```
<compress>
  <relationship>michael</relationship>
</compress>
```

Memory reads the relationship neighborhood, derives understanding nodes from clusters of traces and salience, retains those, discards what no longer needs to be carried.

Compression order: traces first, then salience that didn't become understanding. Tensions persist until resolved. Understandings persist indefinitely.

Memory quality is measured by how little the being needs to carry without breaking.

---

## Implementation

### Embeddable, Pure Go

No external database. The graph lives in-process:

- **Graph storage** — adjacency lists with typed edges, weight histories. Serializable to disk (JSON or binary).
- **Vector index** — HNSW or flat index for semantic search. Embeddings via external API call (same provider infrastructure).
- **Entity extraction** — rule-based: dictionary lookup, title-case detection, acronym patterns, regex.
- **Entity resolution** — normalize → alias table → fuzzy match (Levenshtein) → rank.
- **Persistence** — serialize graph to disk on shutdown, load on startup. Per-being, per-relationship files.

### On Self

Memory and Context are both Realities on Self, alongside Desk and Think.

```go
type Memory struct {
    id        string
    Graph     *MemoryGraph
    Extractor *Extractor
    Resolver  *Resolver
}

type Context struct {
    id        string
    LLM       Reality
    Loaded    map[string][]Artifact  // entity anchor → loaded artifacts
    Active    string                 // current relationship
    Memory    *Memory                // reference to long-term store
}
```

- Context sits between Memory and Think
- Context manages the window — loads, evicts, scopes by relationship
- Think subscribes to Context (not Memory)
- Think has remember/recall as inner operators — they write to/read from Context
- Context syncs back to Memory

### Package Structure

```
src/reality/
├── memory.go      — Memory Reality, long-term graph store
├── context.go     — Context Reality, LLM-managed working window
├── memgraph.go    — in-process graph: nodes, edges, adjacency, traversal
├── memvec.go      — vector index for semantic search
├── extract.go     — rule-based entity extraction
├── resolve.go     — entity resolution pipeline
```

---

## What This Is Not

- Memory does not store raw conversation history. That lives in Exchange.
- Memory does not require user approval. The being owns it.
- Memory is not a database exposed to the world. It is internal to the being.
- Memory does not decide what the being believes. It surfaces what the being has retained. The being reasons over it.
