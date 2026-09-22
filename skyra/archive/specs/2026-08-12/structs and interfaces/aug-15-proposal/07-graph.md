# The graph

**Date:** 2026-08-15
**Status: PROPOSAL.** This is where decision D4 lives.

> **The graph witnesses only what Remember retained. It has ONE mutator, and everything that used
> to be a write is now a member of the thing that mutator takes.**

---

## 1 · The live causal record — ephemeral, kernel-held

Before the graph, the thing the graph is projected *from*. `causality-and-recognition.md` requires
it and nothing types it.

```go
package kernel

// Event is one witnessed occurrence inside a live branch. Append-only, ephemeral, dies with the
// kernel. The kernel may hold and connect these identifiers WITHOUT INTERPRETING CONTENT.
type Event struct {
	ID     ids.EventID
	After  []ids.EventID // EXPLICIT PREDECESSOR EDGES. a causal graph, not one global timeline.
	Kind   EventKind
	Engram ids.EngramID
	Inputs []ids.ObservationID
	Output ids.ObservationID
}

type EventKind uint8
const (
	Received EventKind = iota // an observation entered this frame
	Decided                   // a position was bound
	Realized                  // Reach made it live, or did not
	Returned                  // a child handed something back
	Recognized                // a parser rendered a verdict
	Revised                   // the parent restated the draft
	Remembered                // this parent retained it
)
```

```text
WHY A GRAPH AND NOT A SEQUENCE   a single sequence number preserves order within one possibility,
                                 but selections run children CONCURRENTLY. Arrival order is a
                                 scheduling accident and must not become meaning. So the record
                                 preserves SEQUENCE, CAUSES, and EMITTED POSITION — three facts, of
                                 which only the first is a number.
```

```text
SOURCE   causality-and-recognition.md §"Method one: Remember preserves causality", including the
         illustrative Event shape, which I have kept nearly verbatim and marked as illustrative
         there and here.
```

---

## 2 · `Cone`, and the one mutator — decision D4

```go
package graph

// Cone — INVENTED. The set says Remember "projects one retained cone" and never names the thing.
// It is the backward-reachable closure from one chosen observation, over the live causal record.
type Cone struct {
	Span     ids.SpanID
	Chosen   ids.ObservationID
	Events   []Event                // every occurrence reachable backward from Chosen
	Observations []*Observation      // the ones those events produced or received
	Positions map[Position]Version   // which version each position was acting through
}

// Project IS THE DURABLE WRITE. There is exactly one, and it happens once per branch.
func (g *Graph) Project(c Cone) error
```

```text
WHAT COLLAPSED INTO IT

    Mint         an engram persists because its cone reached the boundary. not a separate act.
    MintRoot     a branch is a POSITION, not a node. there is nothing root-shaped to mint.
    Store        observations live in ephemeral kernel state until projection. this closes the
                 two writes-outside-a-commit that type-surface-proposals.md C6/C7 identify and
                 that the-collapse.md's "unchanged" block wrongly claims do not exist.
    Apply        one record per commit becomes one projection per branch.
    Disperse     the walk still runs — but ONCE, inside Project, over the retained cone.
    Hold         nothing is minted before projection, so there is nothing to park.
```

**This is `closure.md` implemented rather than restated.** *Closure is the only causal durable
append. A commit produces an observation; it does not persist anything.* And mint-on-closure was
always named for the right event and implemented at the wrong one — v5 mints at every commit, which
under this vocabulary is minting at **return**.

```text
WHAT IT BUYS   crash semantics collapse to nothing, for real this time. The kernel dies, everything
               in flight evaporates, and the graph is exactly the set of projected branches. No
               half-written state, no reconciliation, no recovery protocol.

WHAT IT COSTS  A LONG CHAIN PERSISTS NOTHING UNTIL THE END. Crash safety is total and progress
               safety is zero. closure.md calls this THE REAL TRADE and refuses checkpoints
               because checkpointing is partial completion wearing a different name.
               → future-research/kubernetes-state-and-recovery.md exists for exactly this.
```

> **QUESTION 7.1 — `Project` takes material the kernel assembled, which is what P4 requires. But
> `Cone` is large.** It carries every event and observation in the closure, which for an hour-long
> branch is the whole hour. A signature that hands the graph a value that big is honest about the
> transaction and terrible about memory. The alternative — the graph pulling from a kernel-side
> iterator — inverts the dependency and lets the graph reach into live state, which is worse.
> **Taken: the big value**, because the ugliness is the trade being made visible rather than a new
> problem.

---

## 3 · What the graph holds

```go
type Graph struct {
	engrams      map[ids.EngramID]*Engram
	namespaces   map[ids.NamespaceID]*Engram // held apart. A NAMESPACE IS AN ENGRAM.
	observations map[ids.ObservationID]*Observation
	discards     []Discard
}

// Invocation is one occurrence of one child position, durable after projection.
type Invocation struct {
	ID       ids.InvocationID
	Span     ids.SpanID
	Position Position
	Version  int

	Entity ids.EntityID
	Host   ids.HostID // RECORDED, NOT KEYED. reach is plural, so the host does not enter the key.

	Inputs []*Observation
	Cite   ids.ObservationID
	Return *Observation

	Remembered  bool         // THE LOCAL VERDICT. determined by what the parent retained.
	Acceptances []Acceptance // appended by the walk, carrying hop counts.
}

type Acceptance struct {
	Span ids.SpanID
	Hop  int
}
```

```text
NAMESPACES ARE HELD APART   for the same reason roots were in v5, and the reason survives the
                            rename: A NAMESPACE IS AN ASK, NOT A CAPABILITY. Its expectation is
                            the best lexical match for its own subtree's need, so offering it to
                            Think means every branch can rediscover itself. It still persists, is
                            still reused, still owns its children, still carries all the learning —
                            and is not a search result.

                            v5 learned this the hard way: MINT-ON-CLOSURE READS AS ONE RULE AND HAS
                            TWO CALL SITES. Under D4 there is one, which is the structural fix.
```

---

## 4 · The walk, unchanged in shape and moved in time

```go
// Disperse runs ONCE per branch, inside Project. It follows CONSTITUENTS, not durable structure:
// observations are instances, so the walk lands only on invocations that actually produced this
// result.
func (g *Graph) disperse(c Cone)
```

```text
O(n · depth) BY DESIGN   an engram deep in a chain is visited once per accepting ancestor, and
                         that repetition is what populates the hop profile. IT COSTS NO MODEL
                         CALL — which is only an argument because tokens are the currency.

ERRORS STAY LOCAL        descending a rejection would be one evidential origin becoming N records.
                         And the asymmetry is not arbitrary: a false positive is self-limiting,
                         a false negative is absorbing, because nothing earns its way back if it
                         is never tried.

THE VERDICT IS FINAL     before, a parent's verdict was rendered the moment it committed, on
                         partial information, and could not be revised. Now a parent may revise
                         until the branch lands, which is what makes "the durable graph records
                         what worked, not what was tried" literally true.
```

**This is the answer to the optimistic-dispersal problem** in `skyra-dna-v5`, where `Apply` calls
`Disperse` at every commit so a rejected branch has already written `Remembered` on every invocation
and attached every child. Under D4 a rejected branch writes nothing at all.

---

## 5 · Effects

```go
type Effect struct {
	Span        ids.SpanID
	Engram      ids.EngramID
	Op          EffectOp
	Disposition Disposition // AN ENUM. two booleans expressed four states and one was unnameable.
}

type Disposition uint8
const (
	Staged Disposition = iota // held, not performed
	Landed                    // the driver wrote it
	Discarded                 // the walk never reached it, or the branch was rejected. RECORDED.
	Failed                    // reached, attempted, the driver errored. RECORDED.
)
```

`02-defects.md` D2 and B1 are the same defect at two levels: a failed write was left *neither*
`Landed` nor `Discarded` and recorded no discard, occupying a fourth state the booleans could not
name. An enum makes forgetting one a compile error.

```text
EVERY STAGED EFFECT IS EXPLICITLY DISPOSED.   A discard is an EVENT, never a silent no-op —
                                              citation was designed with cheap failure and a lost
                                              write is not that.

MATERIALIZATION FOLLOWS CITATION.             the effects that land are exactly the ones the walk
                                              reaches, for the same reason they earn credit.
                                              One traversal, two jobs.
```

---

## 6 · The reads

```go
func (g *Graph) Engram(ids.EngramID) (*Engram, bool)
func (g *Graph) Observation(ids.ObservationID) (*Observation, bool)
func (g *Graph) Engrams() []*Engram   // DISCOVERABLE. capabilities, never namespaces.
func (g *Graph) Discards() []Discard

// Index is everything the ENTITY role may see. Read-only BY CONSTRUCTION — there is no *Graph in
// that package, therefore no Project, no disperse. The separation is a type, not a promise.
type Index interface{ Engrams() []*Engram }
```

```text
DERIVED, NOT STORED   HopProfile · UsedSeen · Partitions · Cites · heads · distance.
                      type-surface-proposals.md C2 removes the UsedSeen TYPE for the right reason:
                      "a type for it invites someone to store it." P10.
```

> **QUESTION 7.2 — the second kind of credit still has no subject, and D4 sharpens it.**
> `second-kind-of-credit.md` records that a superseded engram never closes, never mints, and
> evaporates — taking the **authored negative** (`REMOVED`: a parent looked at it and took it out)
> with it. That is the strongest signal in the system and the transaction model destroys it at the
> moment of creation. `age.md` argues version retention recovers it. **Under D4 that only works if
> `Cone.Positions` carries superseded versions**, and a cone by construction contains only what
> contributed. So the type above **cannot** carry the negative. Either `Project` takes a second
> argument the cone does not include — which weakens `REMEMBERED ⊆ OCCURRED` into something with an
> exception — or the signal stays lost. → 09 F15. **This is the sharpest unresolved tension in the
> proposal.**

> **QUESTION 7.3 — is there a second durable append path for cost?**
> `distance.md` decides the owner and leaves the mechanism open: *"Does the kernel have to append
> cost directly to the namespace even when no causal chain is remembered?"* If yes, `Project` is
> not the only mutator and D4 is false as stated. The cost account has a **different subject** —
> the namespace, not the cone — and a rejected branch still spent. `energy/README.md` notes that
> deciding this probably decides the budget question too. **Not taken.** → 09 F4.
