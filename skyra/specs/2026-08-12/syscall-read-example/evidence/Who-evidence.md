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
                      1 unique   2 indep.    3 expression   4 returns   5 no back-channel
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

2   measured this session — see E3. Every candidate has a NAME; what separates them is
    whether reaching it produces a party with its own failure domain.

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

## E3 · Why the criterion is not "has an address"

The first form of criterion 2 was *"has at least 1 unique address"*, on the reasoning that a
linked library has none. **That is false, and the measurement is one line:**

```console
$ python3 -c "ctypes.CDLL('/opt/homebrew/opt/openssl@3/lib/libssl.3.dylib')"
```

A path was named and a transport ran. Identical in shape to reaching curl:

```text
curl      /usr/bin/curl                                    fork(2) + execve(59)
libssl    /opt/homebrew/opt/openssl@3/lib/libssl.3.dylib   dlopen → open(5) + mmap(197)
```

`dlopen` is **not** in Darwin's syscall table — it is a dyld function over `open` and `mmap`.

**The runtime pointer misled the first attempt.** `SSL_CTX_new` lands at a different address every
launch, which looked like "no stable address":

```console
run 1:  SSL_CTX_new @ 0x000100731ef4
run 2:  SSL_CTX_new @ 0x000105321ef4
run 3:  SSL_CTX_new @ 0x000101101ef4
```

That is ASLR. The *address* of a library is its **path**, which is exactly as stable and shareable
as `/usr/bin/curl`. And nothing is "listening" at `/usr/bin/curl` either — both are inert files
until a transport runs, so that test discriminated nothing.

**What actually differs is what the transport produces:**

```console
before dlopen:  pid 19388
after  dlopen:  pid 19388                 ← SAME PROCESS
after  fork:    pid 19388 spawned 19389   ← a SECOND process
```

```text
execve · connect   →  a HANDLE (an fd) to a party that is not you. its own address space,
                      its own crash, its own lifetime.
dlopen             →  a POINTER, which is an offset INTO you. no new pid. you are simply
                      larger.
```

**The test, in one question:** *can it fail without taking you with it?*

```text
mongod crashes         → connection error. you keep running.        ✓
Claude returns 503     → an error observation. you keep running.    ✓
a spawned curl faults  → nonzero exit status. you keep running.     ✓
libssl faults          → YOU FAULT. same process.                   ✗
```

*Every console block above was run on this machine.*

## E4 · A path and a network address are the same object: a NAME IN A NAMESPACE

Every candidate reached this session was named, and each name belongs to a namespace that **is**
its resolution procedure:

```text
/usr/bin/curl                       the FILESYSTEM namespace    → open · execve
api.anthropic.com:443               DNS + IP + port             → socket · connect
/tmp/mongodb-27017.sock             the filesystem again        → socket · connect
/opt/homebrew/…/libssl.3.dylib      the filesystem              → open · mmap
```

Note that two different namespaces resolve through the filesystem to **different transports** —
the name alone does not tell you how to reach it; the namespace does.

**URI notation states the namespace outright, as a scheme:**

```text
file:///usr/bin/curl
https://api.anthropic.com/v1/messages
unix:///tmp/mongodb-27017.sock
```

```text
THE SCHEME IS THE RESOLUTION PROCEDURE.
```

This is the same shape the design already uses for authority — `terminology.md` makes NAMESPACE a
first-class durable object (*"an engram. a principal's top engram. children are what they can
select"*), and capability is name resolution: you can select what you can name, and you can name
what you were given.

**And it is why criterion 2 could not be "has an address":** everything here has a name, libraries
included. Naming is universal, so it discriminates nothing — which is exactly what E3 measured.
