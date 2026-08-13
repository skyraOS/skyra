# 08 · The Kernel

**Package:** `kernel`   ·   540 code lines (245 load-bearing)

Holds everything that is **happening**. All of it ephemeral, all of it dies with the process.

> See `../kernel-revision.md` for the argument that the *method* surface collapses to one channel and
> two forms. This document is the *type* surface, which is a different cut.

## The kernel

```go
type Kernel struct {                          // SPEC
    G    *graph.Graph      // the witness
    Reg  *entity.Registry
    Host host.Host
    IDs  *ids.Minter
    Run  *trace.Run        // ← outside the design. see 09-trace.

    Builtin map[string]*graph.Engram   // discover. AMBIENT AUTHORITY, deliberate.

    MaxDepth int           // a RUNAWAY is depth or breadth past a bound
    MaxSteps int

    roots    map[ids.SpanID]*Root
    heads    map[ids.SpanID]int              // the concurrent frontier
    unminted map[ids.EngramID]*graph.Engram  // identity without closure
    userRooted, kernelRooted int             // the spiral, made countable
}
```

**Everything below `Builtin` is ephemeral by design.** The kernel dies, it all evaporates, and the
graph is exactly the set of completed commits — no half-written state, no reconciliation, no recovery
protocol.

```text
FOR REVIEW   `unminted` is the type-level trace of the Hold leak. it exists only because
             mint-on-closure needs somewhere to park an engram that has identity and has not run.
             kernel-revision.md argues the METHOD should go; the FIELD is the same question.
             if a surface needs it, the rule is wrong, not the surface.

             `Run` puts a trace handle in the component that is supposed to hold no semantics.
```

## The possibility

```go
type Possibility struct {                     // SPEC
    Engram *graph.Engram
    Entity ids.EntityID
    Name   string

    Observations []*graph.Observation
    Returned     []*graph.Observation
    Selected     bool         // the origination forfeit. set at the FIRST select, PERMANENTLY.

    Impulses []*graph.Observation   // PROVISIONAL. currently always empty.
    Pending  []*graph.Pending       // its whole life, written in one go at commit
    Delegated map[string]*graph.Engram
    Reach     []ids.EntityID        // narrows monotonically
    Ancestry  []ids.EngramID        // ADDED — reach may not go upward
}
```

**`Ancestry` is the one field the build forced in.** Mint-on-closure plus discovery-walks-the-index
means a chain could be handed reach to its own ancestor. The kernel is the right holder because it
already holds ancestry, and refusing a binding is execution topology — no content read, no judgement
made.

`Pending` accumulating across the whole life and being written **in one go at commit** is what makes
eviction safety fall out rather than being built.

## The transaction

```go
type RootOrigin int                           // SPEC(n)
const (UserRooted RootOrigin = iota; KernelRooted)

type Root struct {                            // SPEC
    Span     ids.SpanID
    Engram   *graph.Engram
    Origin   RootOrigin
    Provoker ids.SpanID       // empty only for the maintenance class
    staged   []*Effect
}

type Effect struct {                          // SPEC(n)
    Engram    ids.EngramID    // materialization follows CITATION — the walk must reach this
    Class     string          // "storage"
    Path      string
    Body      []byte
    Handle    string
    Landed    bool
    Discarded bool
}
```

**The root is the transaction boundary, not the possibility.** Nested transactions raise *what does
it mean for a child to commit into a parent that later aborts*, and there is no good answer.

`Materialize` realizes exactly what the credit walk reached and **discards the rest, recording every
discard**. One traversal, two jobs.

```text
FOR REVIEW   Effect has 7 fields for one implemented class. Class is a string, which is the shape
             of a thing that wants to be an interface once there is a second class. Landed and
             Discarded are mutually exclusive booleans — a three-state disposition (staged ·
             landed · discarded) that got two flags.

             Grants is ABSENT from Root. interfaces.md declares it. the world boundary is entirely
             ungated: staging works and nothing checks it.
```

## Not types, but the surface

```go
func (k *Kernel) Ingress(c content.Content) *graph.Observation
func (k *Kernel) Ask(ask string, ent ids.EntityID, in []*graph.Observation) (*Root, *graph.Observation, error)
func (k *Kernel) Drive(r *Root, ent ids.EntityID, in []*graph.Observation) (*graph.Observation, error)
func (k *Kernel) Accept(r *Root, o *graph.Observation) error
func (k *Kernel) Reject(r *Root, why string)
func (k *Kernel) Introspect(from ids.SpanID) (heads int, ok bool)
func (k *Kernel) Spiral() (user, kernelRooted int)
func (k *Kernel) Hold(e *graph.Engram)
```

Seven of these eight should not exist. `../kernel-revision.md` is the argument: `Drive` is a select,
`Accept` is a commit, `Ask` is a composition, `Spiral` is a read, `Hold` is a leak.
