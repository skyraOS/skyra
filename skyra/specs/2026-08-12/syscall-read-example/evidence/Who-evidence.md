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

## E2 · The five, tested against real entities

```text
                      1 unique   2 address   3 expression   4 returns   5 no back-channel
Claude Opus 5            ✓           ✓            ✓             ✓              ✓
curl 8.7.1               ✓           ✓            ✓             ✓              ✓
mongod                   ✓           ✓            ✓             ✓              ✓
python3.14               ✓           ✓            ✓             ✓              ✓
an external HTTP API     ✓           ✓            ✓             ✓              ✓
─────────────────────────────────────────────────────────────────────────────────────
OpenSSL 3.2              ✓           ✗            ✓             ✓              ✓
```

**Where each criterion comes from:**

```text
1   terminology.md: "registered per runtime; KNOWN ACROSS RUNTIMES BY GOSSIP."
    anything gossiped must be comparable across runtimes.
    world.md: "python3.4 and python3.5 are DIFFERENT ENTITIES, because version changes
    interpretation."

2   measured this session — every entity reached had an address and a transport:
      curl      /usr/bin/curl            fork(2) + execve(59)
      Claude    api.anthropic.com:443    socket(97) + connect(98)
      mongod    /tmp/mongodb-27017.sock  socket + connect
    A linked library has neither, and no failure domain of its own.

3   interfaces.md: "IN  BEHAVIOUR. polymorphic — code, prose, a query, a URL."
    and: "Handing a compiler natural language should fail. The entity reading it and
    failing IS the observation."

4   interfaces.md: "OUT  EVIDENCE. uniform. always structured."
    world.md: "an entity is THE THING YOU WANT A SINGLE TRACK RECORD OF" — with no return,
    nothing is minted and that record is empty.

5   world.md: "Anything accumulating between invocations is world state and must be CARRIED
    EXPLICITLY — a session handle, a worktree, a page. If state hides inside an entity, two
    invocations that look identical give different answers and nothing can see why:
    `same question` becomes a lie."
```

**Three things were wearing the word "state", and only one of them breaks anything:**

```text
SUBSTANCE           weights · the collection · the stdlib. constant across your invocations.
                    WHAT THE ENTITY IS. never a problem.

WORLD STATE         what your own effects wrote. changes between invocations, but the runtime
                    wrote it and holds the handle. already modeled — staged, materialized on
                    citation.

HIDDEN CONTINUITY   the entity remembering YOUR last invocation through a channel you do not
                    hold. ← the only one that breaks `same question`.
```

*Quotations verified against `specs/2026-08-09/{world.md, interfaces.md}` and
`specs/2026-08-11/semantics/terminology.md`. The addresses and transports were measured on this
machine this session; the OpenSSL row is reasoning from the same measurements, not a separate one.*
