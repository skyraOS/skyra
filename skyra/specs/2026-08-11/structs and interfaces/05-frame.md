# 05 · The Frame

**Package:** `frame`   ·   107 code lines

What a possibility **sees**. Also a budget boundary, since everything in it ships to the entity's
host on every step.

## The interfaces

```go
type Frame interface {                        // SPEC
    Self() ids.EngramID

    Entity()   ids.EntityID       // who is interpreting ME
    Entities() []ids.EntityID     // who I may ADDRESS. scoped, exactly like names.

    Observations() []*graph.Observation
    Returned()     []*graph.Observation

    Engrams() View                // strips Cites

    Resolve(name string) (*graph.Engram, bool)   // ADDED to the interface
    InScope(id ids.ObservationID) bool           // ADDED to the interface
}

type View interface {                         // SPEC(n) — "EngramView"
    Names() []string
    Summary(name string) (graph.Summary, bool)
}
```

`Resolve` and `InScope` are additions to `interfaces.md`'s `Frame`, and they are the two operations
that make capability-as-scoping executable rather than descriptive:

```text
Resolve   turns a NAME into an engram. you can select what you can name.
InScope   the graph-boundary authorization. was this ID in the emitter's frame?
```

`interfaces.md` describes both as properties of the frame without giving them methods, so something
had to name them. **They belong on the interface rather than in the kernel**, because the perspective
belongs to the possibility.

## The implementation

```go
type Live struct {                            // ADDED — implements Frame
    Engram    *graph.Engram
    Ent       ids.EntityID
    Reach     []ids.EntityID          // inheritable, narrows monotonically
    Obs       []*graph.Observation
    Rets      []*graph.Observation
    Impulses  []*graph.Observation    // PROVISIONAL — buffered to the monitor boundary
    Delegated map[string]*graph.Engram  // reach a child handed back THIS invocation
    Builtin   map[string]*graph.Engram  // discover. nameable in every frame.
    G         *graph.Graph
}

type view struct{ e *graph.Engram; g *graph.Graph }   // ADDED — implements View
```

**Name resolution looks in three places, nearest first:**

```text
CHILDREN    what has WORKED here. formed on citation.
DELEGATED   what a child handed back this invocation. DELEGATION, not injection.
BUILTIN     discover. nameable everywhere — which is AMBIENT AUTHORITY, and deliberate.
```

A name absent from all three is unresolvable, and **there is no lookup of last resort.** That is the
whole enforcement of capability-as-scoping.

## Review notes

`Live` has ten fields and is the widest struct in the tree. Three of them — `Impulses`, `Delegated`,
`Builtin` — are each one mechanism, and `Impulses` is provisional and currently always empty.

`View` exists so an entity cannot reach `Invocation.Return.Cites` and traverse into a child's
subtree. Without it, *the parent sees what the child concluded, not what it was told* stops holding.
`interfaces.md` says its shape "has not been worked through" and that remains true — it currently
strips `Cites` and passes a `graph.Summary` straight through, which means the presentation type from
`03-graph` is visible across this boundary.
