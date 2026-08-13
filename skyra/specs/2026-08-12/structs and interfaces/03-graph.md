# 03 · The Graph

**Package:** `graph`   ·   344 code lines, the largest in the tree

The durable layer. It holds only what **happened** — nothing that is happening is representable.

## The durable model

Five types, 52 lines total. **The smallest thing in the tree and the most load-bearing.**

```go
type Expectation struct {          // SPEC
    Expression string              // NO LANGUAGE TAG. a tag would be a self-report.
}

type Engram struct {               // SPEC
    ID          ids.EngramID       // what it IS; the NAME is the parent's map key
    Expectation Expectation
    History     []*Invocation
    Engrams     map[string]*Engram              // its children, keyed by name
    ByReturn    map[ids.ObservationID]*Invocation  // the walk's entry point. derived.
}

type Observation struct {          // SPEC
    ID       ids.ObservationID
    Content  content.Content
    Producer ids.EngramID          // EMPTY means INGRESS
}

type Invocation struct {           // SPEC
    Span         ids.SpanID        // determined at invocation, WRITTEN AT COMMIT
    Observations []*Observation
    Entity       ids.EntityID      // KEYED — part of what makes two askings the same question
    Host         ids.HostID        // RECORDED, NOT KEYED
    Cite         ids.ObservationID // a GROUPING KEY ONLY  ← see questions.md §15
    Return       *Observation      // written at commit
    Cited        bool              // THE LOCAL VERDICT
    Acceptances  []Acceptance      // APPENDED LATER by the walk
}

type Acceptance struct {           // SPEC
    Span ids.SpanID                // a LABEL, not a bound
    Hop  int                       // 1 = the parent's parent accepted
}
```

**An engram lives in exactly one parent's map**, so two parents structurally cannot share one. That
is what makes locality structural rather than argued: *A never writes into B's memory, A writes into
the child engram A owns.*

```text
FOR REVIEW   conversation-runtime-mismatch.md flags this: engrams are currently reused BY POINTER
             AND ID, where the conversation has reuse COPYING the expectation into a new locally
             owned engram. copy-on-reuse would make the one-parent rule true by construction rather
             than by discipline, and interfaces.md already argues duplication is nearly free
             because expectations never improve.
```

## The witness

```go
type Graph struct {                // ADDED — and the biggest unnamed object in interfaces.md
    engrams      map[ids.EngramID]*Engram   // what DISCOVERY walks
    roots        map[ids.EngramID]*Engram   // held APART. a root is an ask, not a capability.
    observations map[ids.ObservationID]*Observation
    discards     []Discard
}
```

`interfaces.md` says *"the graph applies that record as writes… the witness, and the only thing with
policy"* and never declares it. So the component holding **all** the policy — engram formation, the
local verdict, credit dispersal — was the one type the spec never wrote down.

**Roots are held apart from what discovery searches.** Mint-on-closure plus discovery-walks-the-index
means every root is a search result, and a root's expectation *is* the ask, so it is the best lexical
match for its own chain's need. See `../kernel-revision.md`.

## The write

```go
type Record struct {               // SPEC(n) — "a RECORD of what was emitted"
    Span        ids.SpanID
    Engram      *Engram
    Name        string
    Invocations []*Pending
    Return      *Observation
}

type Pending struct {              // ADDED — scaffolding
    Child *Engram
    Name  string
    Inv   *Invocation
}

type Discard struct {              // ADDED — kernel.md demands a discard be a recorded EVENT
    Span   ids.SpanID
    Engram ids.EngramID
    Detail string
}
```

`Apply(*Record)` is **the only durable write**, and it holds four policies: formed-on-citation, the
local verdict, history append, and dispersal. `Disperse` is the credit walk; `Reached` is the same
traversal used for materialization — one walk, two jobs.

## The read-time folds

**Five types to produce four numbers. The clearest thing to cut in the tree.**

```go
type Partition struct {            // ADDED
    Question string
    Shape    string
    Entity   ids.EntityID
    Invs     []*Invocation
}

type Diff struct {                 // ADDED
    From, To ids.ObservationID
    Fields   map[string][2]string
}

type UsedSeen struct{ Used, Seen int }   // ADDED

type Summary struct {              // ADDED — pure presentation
    Name        string
    Engram      *Engram
    Invocations int
    CitedCount  int
    Hops        map[int]int
    Partitions  []*Partition
}
```

Every one is derived at read time — no field, no migration, and changing any definition invalidates
nothing already written, exactly as `credit.md` requires. **And "it is only a read-time view" is
precisely the argument that lets five of them accumulate.**

```text
Partition   EARNS IT. the two folds are the knowledge of what a pairing returns, and the run
            showed it separating a right pairing from two wrong ones with an identical shape.

Diff        borderline. "within a partition the DIFF is the informative object" is a real claim,
            but it is one method on Partition rather than a type.

UsedSeen    two ints and a ratio. a function would do.

Summary     belongs in the entity adapter. it exists because something had to render a child set
            into a prompt, which is MARSHALLING, and marshalling is what interpreting is.
```
