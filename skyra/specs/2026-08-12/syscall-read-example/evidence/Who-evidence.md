# WHO — Evidence

**Status: OBSERVATION.** Measurements bearing on **WHO**. Commands and their output. **No
conclusions** — what these mean is [`../outcomes/Who-outcomes.md`](../outcomes/Who-outcomes.md).

**The question, from the engram's perspective:** *Who am I?*

**Scope:** who is interpreting me — the entity, and what makes it one identity.

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
