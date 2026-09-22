# durable/02 · The Graph

**Package:** `durable`

The witness. It holds only what happened, and under the new model it is written from exactly one
place.

---

## 1 · The shape

```go
type Graph struct {
    mu           sync.RWMutex
    engrams      map[ids.EngramID]*Engram
    namespaces   map[ids.EngramID]*Engram          // held APART from what Think walks
    observations map[ids.ObservationID]*observation.Observation
    discards     []Discard
}
```

Carried from `../../03-graph.md`, with `roots` renamed to `namespaces` per
`../../../semantics/terminology.md`'s retirement of the word:

```text
RETIRED   "ROOT". it implied a special node and was not one. every wrong turn — root-as-engram,
          root-as-namespace, root-as-hinge — came from that implication.
```

**Held apart is the substantive part, and it survives the rename.** Mint-on-closure plus
Think-walks-the-index means every namespace is otherwise a search result, and a namespace's
expectation *is* the ask — so it would be the best lexical match for its own chain's need. See
`../../../semantics/the-collapse.md`.

## 2 · One writer

```go
// the ONLY durable mutation.
func (g *Graph) Project(cone Cone) error
```

This is the change. `../../03-graph.md` records five current write paths — `Ingress` and group
completion call `Store`, closure calls `Mint` or `MintRoot`, effect handling calls `RecordDiscard`,
and `Apply` stores the returned observation.

```text
CURRENT     Apply(*Record), plus four other durable mutations.
PROPOSED    Project(Cone), and nothing else.
```

`Record` and `Pending` do not survive. They were the shape of *a record of what was emitted* applied
at commit; under `Remember`, what is applied is a cone selected at the outer boundary, and the
staging they scaffolded happens in ephemeral kernel state instead.

```text
REMOVED   graph.Record · graph.Pending · Apply · Store · Mint · MintRoot
KEPT      Discard, and RecordDiscard folds into Project. see kernel/06-remember.md §4.
```

**This is the largest single divergence from the runtime in this directory**, and it is precisely
what `../../../overnightBuild/README.md` fenced off — *"DO NOT CHANGE when durable writes happen."*
That fence was correct for that work order and does not bind a design pass.

---

## 3 · What the graph structurally cannot hold

The type is shaped so that P11 needs no discipline:

```text
NO CONTRACT TYPE       a stored contract edge would say only that two things MIGHT compose. it
                       would go stale and be mistaken for evidence that composition succeeded.
                       "CONTRACT AS GRAPH EDGE — potential looks like experience."

NO LIVENESS FIELD      nothing that is HAPPENING is representable. no head, no session, no rack.
                       "LIVE ATTEMPT AS DURABLE EDGE — exploration looks accepted before anyone
                       remembers it."

NO Horizon             authority is ephemeral and provisioned per decision. the CEILING is durable
                       and lives on the namespace engram, which is a different fact.

NO CANDIDATE           a candidate is potential. it enters the graph only if Decided, Reached, and
                       Remembered, at which point it is an Invocation.
```

The third defect that document names is the one no type can prevent:

```text
DURABLE EDGE AS TYPE   what worked once becomes the only thing considered compatible later.
```

That is a reading error, not a storage error, and the guard against it is the same as the guard
against reading age as quality: nothing in `Think`'s inputs may be a durable edge presented as a
compatibility fact.

---

## 4 · The read-time folds

> **SUPERSEDED WHOLE, 08-15 — UNRATIFIED.** See
> [`../diffs-and-the-possibility.md`](../diffs-and-the-possibility.md). `Partition` attributes
> returning to a *pairing*, which does not run. The diff belongs to the possibility, reached through
> the span an observation already carries. **`Partition` and the entire fold chain are deleted** —
> `Partitions`, `diff`, `Shape`, `Fingerprint`, `FingerprintSet`, `Deref`, and `content/question.go`,
> none of which has a caller once `Summary` is cut.
>
> Everything below is retained as the record of what was proposed, including the compiler-key
> revision and the record-versus-key split. Both are moot: with no cross-invocation key there is
> nothing to key and nothing to protect.

Carried, with the cut `../../README.md` already argues for.

```go
type Partition struct {
    Question string       // Compiler + FingerprintSet(inputs)      ← was Entity + …
    Shape    string
    Compiler ...          // ← was Entity ids.EntityID. see below.
    Invs     []*Invocation
}

type diff struct {                 // unexported, deliberately
    From, To ids.ObservationID
    Fields   map[string][2]string
}
```

### The interpreter key changed — 08-15

`Entity` dissolved into a **creator slot** and a **compiler slot**
([`../creator-and-compiler.md`](../creator-and-compiler.md) §2). Only one of them belongs in this
key:

```text
THE COMPILER   present at every invocation. determines what shape comes back.   ← THE KEY
THE CREATOR    absent at runtime. authored the body once and left.              NOT the key
```

The original justification survives intact and just gains a more precise subject —
`2026-08-09/interfaces.md`: *"the same expectation under Claude and under python3.11 returns
different shapes."* Claude and python3.11 are **compilers**.

### And the key may be dissolving entirely

`../creator-and-compiler.md` §3 argues that once a body solidifies, its syntax pins its compiler —
one engram, one compiler, forever. If that holds:

```text
BEFORE SOLIDIFICATION   the behavior is prose. many compilers possible. THE KEY EARNS ITS PLACE.
AFTER SOLIDIFICATION    the compiler is fixed. every invocation shares it, and keying on it
                        partitions nothing.
```

Which would make interpreter-keying a **pre-solidification mechanism** that quietly stops doing
work rather than one that stays load-bearing.

### The record is finer than the key, deliberately

`../creator-and-compiler.md` §9 requires a `WorldRef` to carry process identity — `address:pid` —
so a silently restarted process behind a stable address is a **visible** discontinuity rather than
an invisible one.

**That granularity must not reach this key.**

```text
IN THE RECORD   address:pid.   a restart is visible. continuity is never falsely claimed.
IN THE KEY      address only.  a restart is THE WORLD MOVING — which this fold exists to absorb.
```

If pid entered the key, every restart would fork a new partition and there would be nothing left to
diff against. The fold would fragment exactly the history that is supposed to show the world moving.

Both are honest and they answer different questions:

```text
THE RECORD asks   did continuity hold?          it must never claim more than it witnessed.
THE FOLD asks     are these comparable?         it must group across changes that do not matter.
```

Which is the general form of what the contract does — `2026-08-09/interfaces.md`:

```text
{tests: 480} vs {tests: 500}      comparable.      THE WORLD MOVED.        nothing wakes.
{tests: 480} vs {error: timeout}  not comparable.  SOMETHING STRUCTURAL.   the parent wakes.
```

**A contract is a variance absorber.** It states how far the world may move before anyone has to
care. A leaf is where that matters most, because a leaf's child *is* the world — and the world
changes on every single invocation. If a parent can map the change back into its own type without
altering anything, the chain never wakes.

```text
UNRESOLVED, and it is open-questions.md H16's surviving consequence 2 — "what Partitions() keys on
when the interpreter is a stack." Two candidate answers, neither taken:

  A  KEEP THE COMPILER KEY.  harmless once fixed — a constant key costs nothing and keeps the
     fold correct across the prose phase, where it does matter.

  B  KEY ON THE VERSION instead.  ../../../reach/age.md gives each position an age, and a solidified
     body's revisions are what actually separate incomparable returns after the compiler stops
     varying. But age.md's own constraint bites: "age is a resolution coordinate, NEVER a selection
     input" — and a partition key is arguably neither, which is exactly why this needs deciding
     rather than assuming.
```

```text
Partition   EARNS IT. the two folds are the knowledge of what a pairing returns, and the run
            showed it separating a right pairing from two wrong ones with an identical shape.

diff        returned by Partition.Diffs(), unexported so callers may inspect results without
            naming and persisting the derived shape.

Summary     GONE. pure presentation. it existed because something had to render a child set into
            a prompt, which is MARSHALLING, and marshalling is what interpreting is. It belongs
            in the entity adapter, which ../../README.md already says.
```

Every fold is derived at read time, so no migration is required when a definition changes. That
property is why the fold layer is the safest place in the design to be wrong.

---

## 5 · The credit walk

`Disperse` survives unchanged in mechanism, and its bound is the thing to protect:

```text
IT ENUMERATES   o.Cites() → Producer → ByReturn
IT NEVER READS  an engram field.
```

Which is what makes the version pointer uncitable (`../identity/04-position.md` §6) and what makes
reach carry no credit. Two consequences worth stating together, because they are the same
mechanism seen from two sides:

```text
THE WALK IS BOUNDED BY CITES BEING EMPTY at a crossing, never by ID scope.
THE WALK DESCENDS OBSERVATIONS ONLY, never lineage.
```

```text
CHANGED BY THIS OUTLINE, and it needs review: under provenance typing
(../contracts/01-observation.md §4), a reach observation IS an observation with a Producer, where
EngramRef deliberately was not a Ref and "never enters Cites, so reach carries NO CREDIT and the
walk cannot descend through it."

If a candidate observation is citable, the walk CAN now descend through a thinking behavior. That
may be correct — questions.md §15's invisible judge becomes visible — or it may re-open the credit
path EngramRef was shaped to close. IT IS NOT THE SAME QUESTION AS VISIBILITY, and this outline
does not settle it. RECORDED AS H13.
```

---

## 6 · Acceptance

```go
type Acceptance struct {
    Span ids.SpanID   // a LABEL, not a bound
    Hop  int          // 1 = the parent's parent accepted
}
```

Carried unchanged. `Hop` is what `../../../energy/distance.md` and `observable-causality.md` both build
on — *"MEASURE VALIDATION by causal distance through successive remembrances"* — and it is the only
quantity in the durable model that says anything about how far a result travelled.

```text
CHILD RESULT REMEMBERED BY PARENT    correct enough for that parent, in that invocation.
RESULT SURVIVES SEVERAL HOPS         accepted at causal distance.
RESULT REMEMBERED BY THE USER        correct for that namespace at that event.
```

**The graph does not certify one global truth. It witnesses the verdict rendered by a perspective.**
If a principal accepts something factually wrong, the record remains truthful — *this principal
accepted this result, produced by this causal cone, at this time.*
