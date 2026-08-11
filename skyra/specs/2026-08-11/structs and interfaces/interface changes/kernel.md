# Kernel

**Date:** 2026-08-11
**Status: CHANGE PROPOSAL.** Code only — signatures, types, and what to delete.

**Why:** [`../../semantics/kernel-revision.md`](../../semantics/kernel-revision.md)

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

// KEEP — the only method left
func (k *Kernel) Ingress(c content.Content) *graph.Observation
```

`Drive`, `Accept` and `Reject` are the two forms with an external caller, so they are emitted rather
than called. `Introspect` becomes a name — see below.

## Types

```go
// kernel.Possibility
+ Ancestry []ids.EngramID

// kernel.Kernel
- unminted map[ids.EngramID]*graph.Engram
- Run      *trace.Run

// graph.Graph
+ roots map[ids.EngramID]*Engram

// graph — new methods for the above
+ func (g *Graph) MintRoot(e *Engram)
+ func (g *Graph) Roots() []*Engram
  func (g *Graph) Engrams() []*Engram      // now returns capabilities only, never roots
  func (g *Graph) Engram(id) (*Engram, bool)  // resolves across BOTH maps, for the walk
```

## New interface, mirroring `runtime.Index`

```go
// entity/runtime
type Index interface {                       // exists
    Engrams() []*graph.Engram
}

type Topology interface {                    // ADD — what introspect reads
    Heads(from ids.SpanID) int
    Depth(from ids.SpanID) int
    Rooted() (user, kernel int)
}
```

`introspect` is then a system engram like `discover`, manned by the runtime entity, and the
`system:introspect` expression joins the constants in `entity/runtime`.

## Behaviour required, not expressible as a signature

```text
ROOT REUSE            an ask that recurs resolves to the existing root engram rather than minting
                      a new one.

MINT-ON-CLOSURE HAS   the kernel's commit path AND the end of an ask. a root must go to MintRoot
TWO CALL SITES        at both, or it lands back in the discoverable index.

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
