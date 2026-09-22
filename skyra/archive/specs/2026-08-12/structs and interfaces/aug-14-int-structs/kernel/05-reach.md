# kernel/05 · Reach

**Package:** `kernel`, internal

The hidden realization phase of Decide. Not a peer of Think, and not callable.

---

## 1 · Unexported, by design

```go
// internal. no possibility holds a reference to this.
func (k *Kernel) reach(d Decision) Realization
```

```text
RECEIVES   the decision · runtime-supplied From · narrowed Horizon.
RETURNS    the possibility, or a bounded impossibility observation.
```

`primitives.md` §4: *"Reach is not directly exposed to the possibility as a peer of Think. It is the
hidden realization phase of Decide."*

**Thinking about reach and reaching are different operations**, and only the first is possibility-
facing:

```text
THINK-REACH-ENTITY   uses experience to propose expectations through which an entity might
                     compose. CREATES POTENTIAL ONLY. reachable through Think.

REACH                resolves current mechanism for the one composition Decide selected.
                     CREATES THE POSSIBILITY. reachable through nothing.
```

---

## 2 · Three caller-facing conclusions, and only three

```go
type Realization struct {
    Outcome Outcome
    Version Version                  // populated on Live only. see §4.
    Because observation.Observation  // bounded. never endpoint-specific.
}

type Outcome uint8

const (
    Live               Outcome = iota  // the decided composition is live
    BeyondCurrentReach                 // Horizon permitted it; no eligible relationship worked now
    BeyondYourHorizon                  // the composition exceeded the authority provisioned
)
```

The two failure conclusions are deliberately coarse. Everything sharper stays below the boundary:

> *"Endpoint-specific evidence remains below the Reach boundary, while its caller receives the
> conclusion at the level where it can make another decision."*

```text
ADDRESSES AND TRANSPORT FAILURES DO NOT ESCAPE UPWARD AS APPLICATION VOCABULARY.
```

A failed endpoint and a working fallback remain causal observations *beneath* Reach, where hidden
composable behavior can learn from them. That is where reach experience accumulates, and it is why
the coarseness upward is not a loss of information — it is a placement of it.

## 3 · What Reach resolves, and what it may not touch

```text
ENTITY       stable identity; keyed in history.       ← Reach RESOLVES it, never substitutes it
HOST         a place from which endpoints are available.
ADDRESS      the name being resolved.                  ← below the boundary. P12.
TRANSPORT    the mechanism used to contact it.         ← below the boundary. P12.
```

**An entity is an identity and does not carry its endpoints. Hosts hold endpoints.** That separation
is what lets an expectation survive movement without embedding a stale endpoint in what the behavior
is.

P9 in full:

```text
REACH MAY NOT SUBSTITUTE   the engram · the entity · the inputs · the child position · the Horizon.
                           any substitution creates a possibility other than the one decided.
```

Enforced by `Decision` having no exported fields (`04-decide.md` §3) rather than by a check inside
`reach`.

---

## 4 · Reach is where a possibility learns its version

From `../../../reach/age.md`, which is DECIDED on the rule:

```text
A PARENT OWNS, FOR EACH CHILD POSITION
    signature · return type · entity · horizon
A CHANGE TO ANY OF THEM IS A VERSION INCREMENT.
```

The continuation is interior; a body rewrite does not increment.

**The version surfaces on Reach's return, on success only:**

```text
POSSIBILITY            live — and the version this possibility is acting through.
                       SAME AGE   an existing version was resolved.
                       NEW AGE    this decision authored a revision. PROVISIONAL.

BEYOND CURRENT REACH   no version. nothing was reached.
BEYOND YOUR HORIZON    no version. nothing was reached.
```

> *"You have to reach something in order to age it."*

### Why not Decide

Decide is where the parent authors the change, so it is the tempting home, and it is wrong for the
reason `primitives.md` §4 gives:

> *"Reach happens before the possibility exists. An undeliverable decision is not a target
> possibility which opened and failed; it is potential the system could not make live."*

A version returned by Decide would be a version for something that may never exist. And because
Reach is forbidden from altering any of the four, Reach reporting the version cannot corrupt it.

```text
DECIDE      the parent authors the change.             nothing versioned.
REACH       the composition goes live.                 the version is RESOLVED for this possibility.
REMEMBER    the chain survives to the outer boundary.  the version becomes DURABLE.
```

Only the third is durable. A rejected chain leaves the position at its prior age.

---

## 5 · The constraint on `Version`

```go
type Version struct {
    Position Position
    Age      uint64
}
```

> **Age is a resolution coordinate, never a selection input.**

`age.md` states the danger precisely: high age may mean refined, and it may equally mean this
position keeps breaking. Reading it as quality at selection time is the global goodness prior
`credit.md` rejects by name, *"and it arrives wearing a version number."*

**In the type surface this means `Version` must not be reachable from `ThinkArgs` or `Candidate`.**
If a candidate could carry an age, ranking on it becomes one line of code away. It is on
`Realization`, which is produced after selection is over.

---

## 6 · The version chain is a pointer, not a container

```text
REJECTED   engramID → [v0, v1, v2 …]     a container. bookkeeping.
TAKEN      v2 → v1 → v0                  the chain IS the history.
```

Same reason `Cites` is derived rather than declared — two places for one fact would eventually
disagree — and the pointer lives on the engram, never in content, so the credit walk cannot descend
it. See `../durable/01-engram.md` §3 and `../contracts/01-observation.md` §3.

---

## 7 · What Reach recovers

`../../../semantics/second-kind-of-credit.md` records a signal as destroyed: a superseded engram never
closes, never mints, and evaporates — taking the **authored negative** with it.

```text
TRIED     it closed.
WORKED    it was cited.
REMOVED   a parent looked at it and took it out.   AUTHORED, NOT AN ABSENCE.
```

Retaining versions preserves `REMOVED`. Not an added cost: `When-outcomes.md` §3 already requires
it, since *"an engram can have multiple remembered versions"* is impossible if supersession
evaporates the predecessor.
