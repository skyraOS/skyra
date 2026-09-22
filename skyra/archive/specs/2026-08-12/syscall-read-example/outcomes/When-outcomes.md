# WHEN — Outcomes

**Status: MICHAEL ONLY.** Decisions about **WHEN**, taken from [`../example.md`](../example.md), on
the evidence in [`../evidence/When-evidence.md`](../evidence/When-evidence.md).

**The question, from the engram's perspective:** *When am I executed?*

**Scope:** when am I executed — my tick, and what I can be sequenced against.

---

## 1 · WHEN is established when the causal chain is remembered

**Decided 2026-08-13.**

An occurrence does not acquire a durable **when** merely by executing. While the frame is live,
events have an ephemeral causal order. Remembering the causal chain accepts that order into state
and establishes when its engrams occurred relative to one another.

```text
OCCURRENCE     creates ephemeral causal order
REMEMBRANCE    accepts a causal chain and establishes durable WHEN
REJECTION      leaves no durable WHEN
```

**WHEN is causal position, not wall-clock time.** A clock may be retained as an observation, but it
does not establish the engram's order. The remembered causal chain does.

## 2 · Age is the version of a given engram

**Decided 2026-08-13.**

Age is versioning for one particular engram. An engram's age increments when a new version of that
engram is accepted. It is not the number of other engrams accepted after it, and it is not elapsed
wall-clock time.

```text
AGE(e)            the accepted version of engram e
ADVANCES ON       acceptance of a new version of e
DOES NOT ADVANCE  rejected versions, ephemeral attempts, or acceptance of another engram
```

An attempted revision can occur without changing age. Only remembrance turns that revision into the
engram's next durable version.

Age therefore answers *which accepted version of this engram?* It does not sequence unrelated
engrams. Their relative **when** comes from a remembered causal relation, if one exists.

## 3 · Current is resolved by a possibility, not stored in the durable graph

**Decided 2026-08-13.**

An engram can have multiple remembered versions, but the durable graph does not contain an
intrinsically **current** version. Current is reflected in a given possibility: it is the version
that possibility presently resolves and acts through.

```text
AGE       a durable version coordinate belonging to a particular engram
CURRENT   a possibility-local resolution of which version is in use
```

Current does not mean oldest, newest, or globally preferred. Different live possibilities may
resolve different versions of the same engram. Treating one version as current inside the durable
graph would erase that distinction and turn contextual use into global mutable state.

## 4 · Skyra garbage collection begins at namespace disconnection

**Decided 2026-08-13. Definition only; deletion behavior remains open.**

Skyra's version of garbage collection applies when a remembered causal chain is no longer connected
to a namespace. A chain does not become collectible merely because it is old, superseded, or not
current in a particular possibility.

```text
CONNECTED TO A NAMESPACE       durably reachable; not garbage
DISCONNECTED FROM ALL NAMESPACES
                               eligible for Skyra garbage collection
```

This defines the reachability boundary, not the final storage operation. Whether collection means
standard deletion, archival, compaction, or some other treatment is deliberately unresolved.
