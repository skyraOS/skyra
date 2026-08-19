# identity/04 · Position and Age

**Package:** `identity`

A child position is where a contract lives, where age attaches, and where causality is recorded.
It is a pair, not a minted name.

---

## 1 · The shape

```go
type Position struct {
    Parent ids.EngramID   // the engram that OWNS the contract at this position
    Name   string         // the position's name in that parent's continuation
}

type Version struct {
    Position Position
    Age      uint64
}
```

`../../../reach/age.md`:

```text
VERSION COORDINATE      (owning parent, child position, age)
```

> *"The parent's map key is already the lineage identity, so this needs no new durable type."*

**Position names the owner first.** That ordering is the enforcement of P2 in miniature: there is no
way to name a position without naming whose it is, so a contract can never be spoken of as belonging
to the child.

---

## 2 · Age is positional

The same behavior composed by a different parent under a different contract **ages independently**,
which is required by `engram-contracts-and-composition.md`:

> *"No engram owns the contracts by which every possible caller in the system must use it."*

So age cannot live on the engram, and `Version` cannot be `(EngramID, Age)`. An engram is not one
thing with one history of revisions; it is a behavior that participates in many positions, each with
its own contract and its own version line.

## 3 · What increments

```text
A PARENT OWNS, FOR EACH CHILD POSITION

    signature      what observations this position receives
    return type    what observations this position hands back
    entity         who interprets it
    horizon        how far outward it may see or act

A CHANGE TO ANY OF THEM IS A VERSION INCREMENT.
```

**The continuation is interior. A body rewrite does not increment.**

The rule is derivable rather than enumerated: the four are exactly what a parent declares at a
position, and therefore exactly what another engram's parser can depend on. The continuation is the
part no caller can see.

Which maps cleanly onto this outline's `Child` contract:

```text
signature     → InputContract          contracts/02-child-contract.md
return type   → ReturnContract         ditto
entity        → bound by Decide        kernel/04-decide.md §2
horizon       → bound by Decide        ditto
```

Three of the four are in the contract; the fourth pair is what Decide binds — and `primitives.md`
§3 lists entity and Horizon among the six bindings, which is where `age.md` gets them from.

---

## 4 · FINDING — the namespace has no position

Age attaches to a position. A namespace sits in no parent's map. So **the rule has nothing to attach
to at the outer edge**, and `Version` cannot name a namespace's revision.

`age.md` flags this as open. It lands on the type surface as a concrete gap: the outermost engram in
any chain — the namespace, which `terminology.md` defines as *"a principal's top engram"* — cannot
be given a `Position`, because `Position.Parent` has no value.

The coupling `age.md` records makes it worse rather than academic:

```text
When-outcomes.md §4 makes GC eligibility "disconnected from all namespaces," so if connection is
to a VERSION rather than to the namespace, EACH INCREMENT ORPHANS EVERYTHING attached to the
previous one.
```

Three shapes, none taken:

```text
A  NAMESPACES DO NOT VERSION.       simplest. but then a namespace's accepted returns accumulate
                                    with no revision record, and "an engram can have multiple
                                    remembered versions" has an exception at exactly the place the
                                    user's own history lives.

B  A DISTINCT EPOCH TYPE.           age.md floats this — "whether that is age or a distinct
                                    epoch." keeps Position total, adds a second versioning
                                    concept.

C  Position.Parent IS OPTIONAL.     makes the outer edge a zero value, which is the shape
                                    ../contracts/01-observation.md §2 argues against for Producer.
                                    consistent to reject it here for the same reason.
```

Recorded as **HOLE H11**.

---

## 5 · Concurrent revision has no total order

The second thing `age.md` leaves open, and it bites the type directly:

```text
two live chains revise one position and both are remembered. §3 permits sibling versions, but then
age is NOT TOTALLY ORDERED and "the accepted version" is ambiguous. §1 does not help: two
independently remembered chains have no causal relation.
```

`Age uint64` above assumes a total order. **If sibling versions are permitted, that field is wrong**
— it would need to be a partial order, and "current" would need a resolution rule that does not
exist yet.

This is not a detail deferred to implementation. It decides whether `Version` is a number or a
lattice point, and it is unresolved in the source. Recorded with H11.

---

## 6 · The invariant that must hold whatever shape wins

From `../../../semantics/second-kind-of-credit.md`, which rejects by name *"a back pointer the walk can
follow — if the walk can descend `From`, credit crosses copies and every edit inherits. Whatever
records provenance must be uncitable."*

```text
THE POINTER LIVES ON THE ENGRAM, NEVER IN CONTENT.
THE CREDIT WALK DESCENDS OBSERVATIONS ONLY.
```

Structural, not policy: `Disperse` enumerates `o.Cites()` → `Producer` → `ByReturn`, never reads an
engram field, and `Cites()` cannot scan a non-content field. So no citation can name a predecessor
and no edit can inherit standing.

> **This is the line to hold. Attributing credit across a lineage would restore the global goodness
> prior through a new door, and it would look like an optimization when it arrived.**
