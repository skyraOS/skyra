# Memory

## What This Is

Memory is not stored on an entity. Memory is an entity.

`remember` is an entity registered in the world at startup. It implements Entity — same interface as everything else. It takes a Relation and returns results. Same protocol.

`remember` is not exempt from its own operations. It retains its own compression decisions, recalls its own history when asked, and compresses itself when it grows too large.

---

## Who Talks to Remember

The inner entity. Not the outer entity.

The inner entity has `remember` in its relationship list and routes to it with `<>` like any peer. The inner entity calls `remember` to pull context beyond its 10-pair working memory window. The outer entity never addresses `remember` — memory access is the inner entity's job.

---

## Operations

### Recall

The inner entity asks `remember` for relevant context.

```
<> remember ~recall ~about <topic>
<> remember ~recall ~from <entity>
<> remember ~recall ~type <trace|salience|tension|understanding>
<> remember ~recall ~after <timestamp>
<> remember ~recall ~before <timestamp>
<> remember ~recall ~strength <min>
```

Filters compose. Without filters, relevance is weighted by type and recency — understanding surfaces first, traces surface last.

### Retain

An entity routes an artifact to `remember` to make it permanent.

```
<> remember ~retain <artifact>
```

Retained artifacts have a type:

- **trace** — a record of what happened in an exchange. Lowest weight. First to be compressed away.
- **salience** — something noticed as significant. Mid-weight. Survives one compression cycle before it either becomes understanding or is discarded.
- **tension** — an unresolved question or contradiction. High weight. Persists through compression until resolved.
- **understanding** — a conclusion reached, something now known. Highest weight. Survives indefinitely.

The shape of when and how the inner entity retains is not yet decided.

### Compress

Compression when the retained store grows too large.

```
<> remember ~compress
```

`remember` reads the origin entity's retained artifacts, derives understanding artifacts from traces and salience, retains those, and discards what no longer needs to be carried. The entity doesn't decide what to keep — `remember` does.

Memory quality is not measured by how much can be recalled. It is measured by how little the present entity needs to carry without breaking.

### Forget

Delete retained artifacts before a given timestamp.

```
<> remember ~forget ~before <timestamp>
```

All artifacts retained by the origin entity older than the given value are permanently removed. This is the only delete operation in the memory system.

---

## Retained Artifact Family

Every artifact retained by `remember` has a structure:

**Trace**
```
type: trace
core:           the exchange or event
entities:       who was involved
thread:         thread ID if applicable
strength:       how significant the exchange was
```

**Salience**
```
type: salience
core:           what was noticed
context:        what was happening when it was noticed
strength:       how strongly it registered
```

**Tension**
```
type: tension
core:           the unresolved question or contradiction
origin:         where it came from
strength:       how much it is still active
```

**Understanding**
```
type: understanding
core:           the conclusion
derived_from:   trace or salience artifacts that produced it
strength:       confidence in the understanding
```

---

## What This Is Not

- `remember` does not store raw conversation history. That lives in the exchange map on the thread. `remember` stores derived understanding.
- `remember` does not decide what an entity believes. It surfaces what the entity has retained. The entity reasons over it.
- Memory is not a database. It is an entity with judgment.
