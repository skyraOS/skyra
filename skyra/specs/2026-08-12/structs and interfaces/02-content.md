# 02 · Content

**Package:** `content`   ·   260 code lines

What an observation carries, and the two read-time folds the history partitions by.

## The closed Value set

```go
type Content map[string]Value        // SPEC

type Value interface{ isValue() }    // SPEC

type (
    Str  string       // SPEC
    Num  float64      // SPEC
    Bool bool         // SPEC
    List []Value      // SPEC
    Obj  map[string]Value  // SPEC
)

type Ref ids.ObservationID           // SPEC

type WorldRef struct {               // SPEC
    Host   ids.HostID
    Handle string
}
```

**The set is CLOSED, and that is load-bearing.** `Cites` is derived by scanning content for `Ref`s,
and a scan over an open set cannot be exhaustive. The unexported `isValue()` is what keeps it closed.

## The one genuine addition

```go
type EngramRef struct {              // ADDED  ← the only addition to the semantic model
    ID   ids.EngramID
    Name string
}
```

**Reach — a name that becomes selectable in the frame that receives it.** It is what `discover` hands
back, and it is a distinct type on purpose, because that is the whole enforcement of `grammar.md`'s
line:

```text
DELEGATION   reach arriving from an entity I CHOSE to invoke.   legitimate.
INJECTION    reach arriving in the data channel.                whoever can talk to me picks what
                                                                I can touch.
```

Same bytes, different provenance, **and only the type keeps them apart.** The kernel honours an
`EngramRef` appearing in a RETURN and ignores one in an argument or at ingress.

```text
WHAT IT IS NOT   a Ref. it never enters Cites, so reach carries NO CREDIT and the walk cannot
                 descend through it.
```

> **FOR REVIEW.** This adds a member to the design's most load-bearing closed set, argued from
> `grammar.md` rather than from `interfaces.md`, which lists the Value set without it. It replaced
> something strictly worse — a bare string carrying reach in content — but it deserves a decision.
>
> And it is the direct cause of the wrinkle in `questions.md §15`: because reach is not a `Ref`, a
> judge that picks among candidates is invisible to credit.

## The folds

```go
func (c Content) Cites() []ids.ObservationID   // derived, never declared
func (c Content) Reach() []EngramRef           // only ever consulted on a return
func (c Content) Shape() string                // the structural fold
func (c Content) Render() string               // for an entity to read

type Deref func(ids.ObservationID) (Content, bool)   // ADDED
func Fingerprint(c Content, d Deref) string          // the same-question fold
func FingerprintSet(cs []Content, d Deref) string    // order-independent
```

**`Shape` is settled here and `interfaces.md` calls it "THE ONLY REAL DECISION" in the partitioning
idea.** Defined as the sorted set of field paths paired with value kinds. A nested `Ref` is a shape
element of kind `ref` and is not dereferenced — shape is the form of the answer, and dereferencing
would let a value fact leak into a structural one. List length is a value change, not a structural
one.

**`Fingerprint` dereferences**, because comparing `Ref`s by ID makes every question unique again,
which is what killed the first InvocationKey. Bounded at depth 4; past the bound a ref degrades to
its ID and marks itself, rather than degrading silently.

```text
DIVERGENCE   interfaces.md leans toward "a comparison function rather than a hashed key, so it
             stays soft." this is a key, because a key is what a map needs. the fold is read-time,
             so it can be replaced without migration.
```

## Review notes

`Deref` is a one-line func type that exists so `content` does not import `graph`. Cheap, and it
could be an inline func parameter.

`Cites` scans **inside strings** for embedded `<obs-N>`, which was forced by the first real model
return putting its reference in prose. That widens what is *derived* and changes nothing about what
is *authorized*, since the scope check already refuses any reference the emitter never held.
