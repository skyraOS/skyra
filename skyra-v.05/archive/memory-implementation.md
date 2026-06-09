# Memory — Go Rewrite Plan

Skyra's memory layer. Rewritten from NeuroNote's graph architecture in Go. Lives on the inner being plane — the being actively reads and writes memory as part of thinking.

## What We're Taking from NeuroNote

1. **Entity extraction** — rule-based regex + dictionary spotting (the 90% case)
2. **Entity resolution** — alias table, abbreviation expansion, fuzzy matching, confidence ranking
3. **Graph storage** — Apache AGE on Postgres, Cypher queries, typed nodes and edges
4. **Graph sync** — delete-and-replace by source, durable edges (SYNONYM_OF, SUBTOPIC_OF)
5. **Concept meta-classification** — LLM identifies synonym/subtopic pairs, writes durable edges

## What We're Dropping

- TipTap editor, note CRUD, block tree, markdown import/export
- Next.js frontend, D3 visualization (Lens handles this later)
- spaCy integration (Go has no good binding; rule-based is enough)
- sentence-transformers (use deterministic embeddings or call an external API)
- FastAPI, SQLAlchemy, Alembic (Go replaces all of this)
- Media/asset handling
- Rate limiting, CORS, API key middleware
- Background job store (Skyra's main loop is already synchronous)

## Architecture

### Reality Interface

Memory is a `Reality`. It implements `ID()`, `Create()`, `Realize()`.

```
type Memory struct {
    id        string
    db        *pgx.Pool
    graph     *Graph
    extractor *Extractor
    resolver  *Resolver
}
```

- `Create()` — connects to Postgres, ensures AGE graph exists, seeds alias table
- `Realize()` — the being calls this to read or write memory

### Two Modes: Read and Write

The inner being controls memory through its impulse. Memory parses the impulse to determine mode:

**Write** — being says "remember: [content]" or structured `~memory [content]`
1. Extract entities from content (rule-based)
2. Resolve entities against alias table (normalize, fuzzy match)
3. Extract relations (verb pattern matching → typed edges)
4. Sync to AGE graph (delete-and-replace by source being + timestamp)
5. Upsert concept registry
6. Return confirmation

**Read** — being says "recall: [query]" or memory is queried automatically by Think
1. Extract entities from query
2. Resolve to canonical forms
3. Query AGE graph — variable-length path traversal from matched entities
4. Return graph neighborhood as text (entities, relations, confidence)

**Passive** — Think fires memory on every relation, memory attaches relevant context
1. Extract entities from the current impulse
2. If any resolve to known concepts, query their 1-hop neighborhood
3. Attach as a parser on the relation — the being gets memory context in its present without asking

### Package Structure

```
memory/
├── go.mod
├── main.go              # standalone test harness
├── memory.go            # Reality interface implementation
├── graph/
│   ├── graph.go         # AGE connection, graph init, Cypher execution
│   ├── sync.go          # delete-and-replace, node/edge upsert
│   └── query.go         # neighborhood traversal, concept lookup
├── extract/
│   ├── entities.go      # rule-based entity extraction (dictionary + regex)
│   ├── relations.go     # verb pattern → typed relation extraction
│   └── keyphrases.go    # keyphrase scoring (optional, can defer)
├── resolve/
│   ├── resolver.go      # orchestrator: normalize → alias → fuzzy → rank
│   ├── normalize.go     # text normalization (lowercase, strip, collapse)
│   ├── alias.go         # alias table lookup + upsert
│   ├── fuzzy.go         # string similarity matching
│   └── ranking.go       # multi-layer confidence ranking
├── schema/
│   ├── migrations.go    # Postgres + AGE schema setup (programmatic, no Alembic)
│   └── seed.go          # seed terms, initial alias entries
└── present/
    └── render.go        # render graph neighborhood as text for the being's present
```

### Graph Schema (AGE)

Same node/edge types as NeuroNote, stripped to what matters for memory:

**Nodes:**
- `Entity` — id (slug), name, kind, confidence, created_at, updated_at
- `Memory` — id, source_being, content, content_hash, created_at
- `Being` — id, name (tracks who remembered what)

**Edges:**
- `MENTIONS` — Memory → Entity (mention_text, confidence, source_being)
- `RELATES_TO` — Entity → Entity (relation_type: IS_A, PART_OF, CAUSES, USES, PRODUCES, CONTRASTS_WITH)
- `SYNONYM_OF` — Entity ↔ Entity (durable, no source)
- `SUBTOPIC_OF` — Entity → Entity (durable, no source)
- `REMEMBERED_BY` — Memory → Being

### Dependencies

**Go packages:**
- `github.com/jackc/pgx/v5` — Postgres driver (supports AGE raw SQL)
- `github.com/agnivade/levenshtein` or similar — fuzzy string matching
- No ORM. Raw SQL + Cypher via `ag_catalog.cypher()`.

**Infrastructure:**
- PostgreSQL 16 with Apache AGE extension
- pgvector extension (for future embedding search, can defer)
- Docker Compose for local dev

### Migration Strategy (Programmatic)

No Alembic. Go runs schema setup on startup:

1. `CREATE EXTENSION IF NOT EXISTS age`
2. `LOAD 'age'`
3. `SET search_path = ag_catalog, "$user", public`
4. `SELECT create_graph('skyra_memory')` if not exists
5. Create relational tables: `entity_aliases`, `concept_registry`, `extraction_cache`
6. Create pgvector table (deferred)

## Build Order

### Phase 1 — Graph foundation
- [ ] `go.mod`, pgx connection, AGE initialization
- [ ] `graph.go` — Cypher execution wrapper with proper escaping
- [ ] `sync.go` — upsert node, upsert edge, delete by source
- [ ] `query.go` — fetch neighborhood (1-hop, 2-hop), fetch by entity ID
- [ ] `migrations.go` — schema setup on startup
- [ ] Test: manually insert nodes/edges, query them back

### Phase 2 — Entity extraction
- [ ] `entities.go` — dictionary lookup + title-case + acronym + domain pattern regex
- [ ] `normalize.go` — lowercase, strip punctuation, collapse whitespace
- [ ] `relations.go` — sentence split, verb extraction, typed relation mapping
- [ ] Test: extract entities and relations from sample text

### Phase 3 — Entity resolution
- [ ] `alias.go` — alias table CRUD (Postgres), exact match lookup
- [ ] `fuzzy.go` — Levenshtein / sequence matching with threshold
- [ ] `ranking.go` — multi-layer scoring (alias 1.0, fuzzy 0.82, embedding 0.74)
- [ ] `resolver.go` — orchestrate: normalize → alias → fuzzy → rank → resolve or unresolved
- [ ] `seed.go` — initial seed terms from config
- [ ] Test: resolve "ML" → "machine learning", "backprop" → "backpropagation"

### Phase 4 — Memory Reality
- [ ] `memory.go` — implement Reality interface
- [ ] `render.go` — graph neighborhood → text for being's present
- [ ] Wire into Skyra's Think: inner being can read/write memory
- [ ] Test: full loop — being writes memory, being reads it back, present includes memory context

### Phase 5 — Concept meta-classification
- [ ] LLM call to classify synonym/subtopic pairs among new concepts
- [ ] Write durable edges to AGE
- [ ] Concept registry to avoid re-classification
- [ ] Test: write three related memories, verify taxonomy edges appear

### Phase 6 — Integration
- [ ] Docker Compose: Postgres + AGE alongside Skyra
- [ ] Genome syntax: `physics ~name memory ~type memory ~device postgres`
- [ ] Wire passive mode: every relation through Think gets memory context attached
- [ ] Verify: being remembers across exchanges, memory shapes responses

## What This Doesn't Cover Yet

- **Memory budget** — the finite active window from world-physics.md. Needs confidence decay, access recency, triage mechanism. Build after the graph is live and we can see what accumulates.
- **Embedding search** — pgvector nearest-neighbor for semantic recall. Deferred until we need it (rule-based + alias resolution handles most cases).
- **Lens integration** — D3 visualization of the memory graph. NeuroNote's frontend could be adapted later, or Lens builds its own.
- **Multi-being memory** — shared vs. private memory spaces. The graph supports it (REMEMBERED_BY edges), but governance rules need design.
