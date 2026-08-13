# WHO — Evidence

**Status: OBSERVATION.** Measurements bearing on **WHO**. Commands and their output. **No
conclusions** — what these mean is [`../outcomes/Who-outcomes.md`](../outcomes/Who-outcomes.md).

**The question, from the engram's perspective:** *Who am I?*

**Scope:** who is interpreting me — the entity, and how it is reached.

Sections are numbered within this document and cited as **Who E1**, **Who E2**, …

---

## E1 · The entity is the thing with a track record

`world.md` defines an entity by what accumulates about it, not by what it is:

```text
"An entity is THE THING YOU WANT A SINGLE TRACK RECORD OF."

"If two things should share one history, they are one entity. If a failure in one should not
 count against the other, they are two."

"python3.4 and python3.5 are DIFFERENT ENTITIES, because version changes interpretation.
 Claude and GPT are different entities."
```

**And the code puts identity in the key while leaving place out of it:**

```go
type Invocation struct {
    Entity ids.EntityID   // part of what makes two askings the same question
    Host   ids.HostID     // "RECORDED, NOT KEYED"
}
```

```text
"Same question to Claude and to a human is a DIFFERENT question. Same question to Claude
 reached two ways is the SAME question."
```

So the same expectation under two entities is two questions, and under one entity in two places is
one. Identity discriminates; location does not.

*Quotations verified against `specs/2026-08-09/world.md` and
`skyra-dna-v5/graph/types.go` on this machine.*

## E2 · Reach is plural, and the design already excludes it from the entity

`world.md` decides the invocation key on exactly this:

```text
"Same question to Claude and to a human is a DIFFERENT question. SAME QUESTION TO CLAUDE
 REACHED TWO WAYS IS THE SAME QUESTION — so the host does not enter what makes two askings
 comparable."

"You invoke an entity, not a host. DELIVERY PICKS A HOST ON THE WAY OUT, and the response
 correlates by invocation identity, not by channel. A human messaged on their TV and answering
 on their phone is not two events — it is one invocation arriving back through a different door."
```

And the code implements it:

```go
type Invocation struct {
    Entity ids.EntityID   // KEYED
    Host   ids.HostID     // "RECORDED, NOT KEYED"
}
```

`Host` is excluded from the key **because** reach is plural. If there were one way to reach an
entity, the exclusion would be pointless.

**Where the mapping is said to live:**

```text
"WHICH ENTITIES ARE ON A HOST is a registry fact, held by the runtime and rebuildable from
 configuration. It is not durable structure, for the same reason addresses are not: it changes,
 and the graph is append-only."

"Addresses are not graph structure… Entity → address → network → host is real, mutable,
 rebuildable from configuration, and belongs to the runtime."
```

Stated host-side — *which entities are on a host* — never entity-side.

**And the entity carries none of it:**

```go
// world.md, after every declaration dissolved
type Entity struct { ID EntityID }

// GONE: Reads []Language · Capacity · Presents/Captures modalities · Hosts []HostID
```

**What no type provides today:**

```go
type Host interface {
    ID() ids.HostID
    Execution() ExecutionOps
    Storage() StorageOps
    // ← nothing asks a host what it can reach
}

type Registry struct {
    interpreters map[ids.EntityID]func() Interpreter   // no host anywhere
}
```

*Quotations verified against `specs/2026-08-09/world.md` and `skyra-dna-v5/{graph/types.go,
host/host.go, entity/entity.go}` on this machine.*
