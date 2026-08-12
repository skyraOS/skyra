# Kernel

**Date:** 2026-08-11
**Status: CHANGE PROPOSAL.** Code only — signatures, types, and what to delete.

**Why:** [`../../semantics/kernel-revision.md`](../../semantics/kernel-revision.md) ·
[`roots-and-admission`](../../semantics/roots-and-admission.md) ·
[`terminology`](../../semantics/terminology.md)

> **Revised 2026-08-11, later.** `Ingress` is now `Mount`, and `Topology` lost `Rooted()`. Read
> `terminology.md` first — **ROOT** is retired throughout, and what it named is a **BRANCH**.

---

## Methods

```go
// REMOVE
func (k *Kernel) Ask(ask string, ent ids.EntityID, in []*graph.Observation) (*Root, *graph.Observation, error)
func (k *Kernel) Drive(r *Root, ent ids.EntityID, in []*graph.Observation) (*graph.Observation, error)
func (k *Kernel) Accept(r *Root, o *graph.Observation) error
func (k *Kernel) Reject(r *Root, why string)
func (k *Kernel) Introspect(from ids.SpanID) (heads int, ok bool)
func (k *Kernel) Spiral() (user, kernelRooted int)
func (k *Kernel) Hold(e *graph.Engram)
func (k *Kernel) rootFor(ask string) *graph.Engram          // unexported; delete with Ask

// REPLACE — Ingress described a model where the method minted an inbound observation. It
// does not: the payload is an argument to the select that follows.
- func (k *Kernel) Ingress(c content.Content) *graph.Observation

// KEEP — the only exported methods. attach a namespace to the rack, and detach it.
+ func (k *Kernel) Mount(ent ids.EntityID, ns *graph.Engram, grants []Grant) (*Tree, error)
+ func (k *Kernel) Unmount(ns ids.EngramID)
```

`Drive`, `Accept` and `Reject` are the two forms with an external caller, so they are emitted rather
than called. `Introspect` becomes a name — see below.

`Mount`'s one check is **`grants ⊆ ns.Ceiling`** — arithmetic, no semantics, and **no live admin
required**, which is what keeps the rack flat.

## Types

```go
// kernel.Possibility
+ Ancestry []ids.EngramID

// kernel.Kernel
- unminted map[ids.EngramID]*graph.Engram
- Run      *trace.Run
- roots    map[ids.SpanID]*Root
- heads    map[ids.SpanID]int              // counts live FRAMES, not the frontier
+ mounted  map[ids.EngramID]*Tree          // THE RACK. flat, keyed by namespace. admin is a row.

// graph.Graph
+ namespaces map[ids.EngramID]*Engram      // held apart from the index, as roots are today

// graph.Engram — for a namespace only
+ Ceiling []Grant                          // DURABLE. the max it may ever be mounted with.

// graph — new methods for the above
+ func (g *Graph) MintNamespace(e *Engram)
+ func (g *Graph) Namespaces() []*Engram
  func (g *Graph) Engrams() []*Engram      // capabilities only, never namespaces
  func (g *Graph) Engram(id) (*Engram, bool)  // resolves across BOTH maps, for the walk
```

**`Tree` and `Branch` replace `Root`**, and the full shape is in
[`type-surface-proposals`](../../semantics/type-surface-proposals.md) `A4`. A tree is all live work
under one namespace; a branch is one ask and is the transaction boundary.

## New interface, mirroring `runtime.Index`

```go
// entity/runtime
type Index interface {                       // exists
    Engrams() []*graph.Engram
}

type Topology interface {                    // ADD — what introspect reads
    Heads(from ids.SpanID) []Head            // live leaves: Returned == nil
    Done(from ids.SpanID)  []Head            // finished leaves, still live and revisable
}

type Head struct {                           // TOPOLOGY ONLY. never the expectation.
    Span   ids.SpanID
    Engram ids.EngramID                      // an ID. the caller dereferences it themselves.
    Entity ids.EntityID
    Host   ids.HostID
    Depth  int
    Since  time.Duration                     // time in Step — what makes STALLED detectable
}
```

`Rooted() (user, kernel int)` is **not** on this interface. It is a global count with no node to
start from, and introspection is bounded by the subtree — which is not a rule but a fact about which
way the pointers point. It is `Spiral` under a new name and belongs to whatever reads the spiral out
of band.

`introspect` is then a system engram like `discover`, manned by the runtime entity, and the
`system:introspect` expression joins the constants in `entity/runtime`.

## Behaviour required, not expressible as a signature

```text
NAMESPACES ARE NOT    a namespace persists and is never offered to discovery. mint-on-closure has
DISCOVERABLE          TWO CALL SITES — the commit path and the end of an ask — and closing only
                      one leaves it in the index.

HEADS ARE DERIVED     walk the live tree. a head is a leaf with no live children and no return
                      under the current draft. THE COUNTER COUNTS THE WRONG THING: it ticks on
                      entering open(), so a depth-five chain with one leaf running reports five.

REACH MAY NOT GO      when folding a returned EngramRef into a frame, refuse any engram already in
UPWARD                the possibility's Ancestry.
```

## Code to delete

```text
kernel/kernel.go   indent · shortExpr · oneLine                        34 lines
kernel/kernel.go   9 Run.Logf calls
```

`shortExpr` reads `Expectation.Expression` and `oneLine` reads `Observation.Content`. They are the
only places the kernel reads content. Replaced by the sink in [entity.md](entity.md).
