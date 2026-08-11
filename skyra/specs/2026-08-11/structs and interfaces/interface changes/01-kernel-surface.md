# 01 · The Kernel Surface

**Date:** 2026-08-11
**Status: CHANGE.** What is called on the kernel. Distilled from `../../kernel-revision.md`.

> **Seven of nine methods go. What is left is one channel; everything else was already the grammar.**

---

## The change

```go
// BEFORE — what the build exported
Ask · Ingress · Drive · Accept · Reject · rootFor · Hold · Introspect · Spiral

// AFTER
Ingress(c Content) *Observation      // a CHANNEL. the only thing outside the two forms.
```

Everything else is either one of the two forms with an external caller, or a name.

```text
Drive       IS A SELECT.    cardinality 1, depth 0, no parent. in the build, Drive and a select
                            member call the identical function — kernel.go:173 and kernel.go:330.

Accept      IS A COMMIT.    citing.        credit.md: "the human is simply the outermost parent,
Reject      IS A COMMIT.    not citing.    running the same operation with a different surface."

                            materialization already follows citation, so cite → effects land and
                            don't-cite → effects discard-and-record. No branch to write.

Ask         A COMPOSITION.  ingress + open a root + drive. written for a test rig.
rootFor     INTERNAL.       finding the engram for a recurring ask.
Spiral      A READ.         user-rooted vs kernel-rooted counts. that is introspection.
Hold        A LEAK.         exists only because mint-on-closure needs somewhere to park an engram
                            with identity that has not run. a detail of a RULE, exported.
Introspect  BECOMES A NAME. see below.
```

## introspect becomes a kernel engram

An engram whose entity reads **kernel** state rather than **graph** state.

```text
discover      reads the GRAPH     what HAPPENED       engrams · entities · observations · hosts
introspect    reads the KERNEL    what is HAPPENING   roots · heads · ancestry · bindings
```

`kernel.md` already specifies this — *"a system engram manned by the runtime entity, the same
exemption discover and create already hold"* — and the build made it a method instead.

The read-only `Index` interface already built for `discover` generalizes:

```go
type Index interface{ Engrams() []*graph.Engram }        // built, for discover
type Topology interface{ Heads(SpanID) int; ... }        // same shape, for introspect
```

Bounded by the subtree, unchanged. Never upward, never sideways.

## Ingress is the one thing left, and it is root access

Since `Drive` is a select, **ingress plus drive is an unscoped select** — the root's engram is in
nobody's frame, so there is nothing to check it against. That is why it is the maximum-privilege
operation and why nothing gates it (`questions.md §5`).

---

## Struct changes that go with it

```go
// kernel.Possibility
+ Ancestry []ids.EngramID     // reach may not go upward

// kernel.Kernel
- unminted map[ids.EngramID]*graph.Engram   // the type-level trace of the Hold leak
- Run      *trace.Run                       // see 02

// graph.Graph
+ roots map[ids.EngramID]*Engram   // held APART from what discovery searches
```

**`Ancestry`** — mint-on-closure plus discovery-walks-the-index means a chain can be handed reach to
its own ancestor. Observed: `by-index` returned the root as the resolution of the root's own need,
recursing to the depth bound. Only the kernel holds ancestry, and refusing a binding is execution
topology — no content read, no judgement.

**`roots`** — the ancestry check is necessary and not sufficient. The kernel can *refuse* a binding
but cannot *redirect* the search: a possibility may never look up, and the grammar has no third
channel to carry "not that one, it is above you." So the cut moves upstream — **a root is an ask, not
a capability.** It persists, is reused when the same ask recurs, owns its children, and is not a
search result.

> Mint-on-closure reads as one rule and has **two call sites** — the kernel's commit path and the end
> of an ask. Closing only one leaves roots in the index.

**Root reuse** — not a field, but a behaviour that has to be stated. All learning is topological and
topology accumulates on a *parent's* children; a fresh root per ask means the outermost possibility
arrives with an empty child map every time. Measured: 8/8 wrong choices before, moved after 5 with
reuse.

## What the kernel stops holding

```text
34 LINES OF FORMATTING   indent · shortExpr · oneLine. shortExpr reads Expectation.Expression and
                         oneLine reads Observation.Content, both to truncate for a log. THE ONLY
                         PLACE THE KERNEL READS CONTENT.

9 TRACE CALLS + A        threaded through the loop.
*trace.Run FIELD
```

Resolved by `02-the-entity-boundary.md`: once the crossing is kernel-held and opaque, formatting has
nowhere to live but a sink that is handed events.

---

## Unchanged

```text
THE SPLIT            the kernel holds everything happening; the graph holds only what happened.
CRASH SEMANTICS      nothing to test — no path writes durably outside a commit.
NO SEMANTICS         execution topology only, once the formatting above is gone.
TWO AUTHORIZATIONS   scope and ownership at the graph boundary. grants at the world boundary are
                     STILL NOT BUILT — Root.Grants is declared in interfaces.md and absent here.
```

## Rejected

```text
COLLAPSING ACCEPT AND    both are commits. the difference is whether a reference is carried and
REJECT TO A BOOLEAN      everything downstream falls out of that. a flag names a branch that does
                         not exist.

EXPORTING Hold           if a surface needs it, the RULE is wrong, not the surface.

FILTERING ANCESTORS      discovery is a possibility and cannot see ancestry. doing it there means
INSIDE DISCOVERY         the kernel feeds a possibility upward information, or a possibility looks
                         up. both break something load-bearing.

A THIRD CHANNEL FOR      adding one ends the claim that only two forms parse.
THE KERNEL'S REFUSAL

INTROSPECT AS A          same kind of thing as discover, and kernel.md already says engram.
KERNEL METHOD
```

## Open

```text
GRANTS               the world boundary is ungated. staging works and nothing checks it.
                     introspect:self is the first grant the design needs — and with introspect a
                     name, it is the first NAME to gate rather than the first method.

WHO MAY OPEN A ROOT  Ingress is maximum-privilege and has no gate.

WHETHER Ask SURVIVES a convenience composition. it earns its keep or it goes.

THE ROOTLESS COMMIT  if Accept IS a commit it has nowhere to land — "the outermost parent" has no
                     engram, so nothing owns the root's Cited bool. either the human gets an
                     engram and acceptance is literally an ordinary commit, or it stays kernel-side
                     and the uniformity is approximate. THE LAST THING BETWEEN THIS AND ONE
                     VOCABULARY.

WHAT introspect      not inference. but it joins discover as one of the highest-frequency engrams,
COSTS                and questions.md §7 says system space is where evidential growth breaks
                     first.
```
