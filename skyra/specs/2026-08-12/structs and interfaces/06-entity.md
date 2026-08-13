# 06 · Entities

**Packages:** `entity` (35) · `entity/codex` (302) · `entity/runtime` (258) · `entity/scripted` (28)

**WHO interprets.** An entity is an identity and nothing else — every declaration dissolved, because
a declaration is made by whoever *registered* the entity, not by the entity.

## The boundary

```go
type Interpreter interface {                              // SPEC
    Step(x graph.Expectation, f frame.Frame) (grammar.Emission, error)
    Monitor(o *graph.Observation)
}
```

**`Step` is the only method that produces anything.** Everything else is data crossing it. Nothing
executes without asking: a possibility cannot invoke, it can only emit a request and be invoked in
turn.

**`Monitor` returns nothing, by constraint.** A return value would be a third emission channel
outside SELECT / COMMIT, and one method with a return type ends the claim that only two forms parse.

```text
DIVERGENCE   interfaces.md writes `Step(x, f) Emission` with no error. but the grammar's incapacity
             is a PARSING property, so a response that parses as neither form has to be
             representable — and the set itself says such a thing becomes "an ordinary failed
             invocation, observed and recorded." something must carry it across. returning it keeps
             the failure OUTSIDE the two forms rather than inventing a third to hold it.
```

## The registry

```go
type Registry struct {                                    // ADDED — SPEC(n) in world.md prose
    interpreters map[ids.EntityID]func() Interpreter
}

func (r *Registry) Register(id ids.EntityID, mk func() Interpreter)
func (r *Registry) Open(id ids.EntityID) (Interpreter, error)
func (r *Registry) Known() []ids.EntityID
```

**A constructor per entity, not an instance**, and that is what makes statelessness structural.
Anything accumulating between invocations is world state and must be carried explicitly; if state hid
inside an interpreter, two invocations that look identical would give different answers and *same
question* becomes a lie.

`Open` fails **loudly** on an unreachable entity. An unreachable entity is not a silent no-op.

## Three implementations

### codex — a model

```go
type Interpreter struct {                                 // ADDED — implements the interface
    ID        ids.EntityID
    Run       *trace.Run
    base      map[ids.ObservationID]bool   // assemble(frame₀)
    delta     map[ids.ObservationID]bool   // what monitor folded in
    Divergent []string                     // recorded, not swallowed
}

func Assemble(x graph.Expectation, f frame.Frame) string   // marshalling IS interpreting
func System() string                                        // the grammar. fixed text.
func Author(need string) (name, expression string, err error)
```

`base`/`delta`/`Divergent` exist to hold the invariant live: **`assemble(frame) == assemble(frame₀) +
monitor*`**, which `interfaces.md` says to build both of and compare. It holds on every step of every
run — trivially, because nothing else in a frame can change mid-possibility. It becomes a real
constraint the moment impulses exist.

`Author` is the **bootstrap**: the one place the design requires something be handed to it. It is
also `questions.md §2` made concrete — the authored expression appears in no citation.

### runtime — code, manning system space

```go
type Index interface {                                    // ADDED
    Engrams() []*graph.Engram
}

type Author func(need string) (name, expression string, err error)   // ADDED
type Mint   func(x graph.Expectation) *graph.Engram                  // ADDED

type Entity struct {                                      // ADDED
    ID    ids.EntityID
    Index Index
    Auth  Author
    Mint  Mint
    Run   *trace.Run
}
```

> **An entity is not handed a graph.** The package holds no `*graph.Graph`, so `Apply`,
> `Mint`-into-index and `Disperse` are absent from it. An entity emits; the kernel drives; the graph
> writes. Nothing here is hidden from anything — an entity holding the graph would simply have
> stopped being an entity.
>
> **One gap:** the graph is passed in *as* an `Index`, so a type assertion would recover it. That is
> a boundary at the declaration site, not an airtight one, and wrapping the graph in a struct that
> does not embed it closes it. Code hygiene, not a design property.

**An engram is not an inference call.** This entity mans `discover`, `by-index`, `by-recency` and
`create`, and only the last reaches a model. Measured: 23 + 6 = 29 invocations with zero inference.

### scripted — deterministic, for the substrate

```go
type Body  func(f frame.Frame, step int) (grammar.Emission, error)   // ADDED
type Table map[string]Body                                           // ADDED

type Interpreter struct {                                            // ADDED
    T    Table
    step int
    Seen []*graph.Observation   // what Monitor was handed, IN ORDER
}
```

Not a mock — an entity like any other, whose judgement happens to be a table. `Seen` exists so a test
can assert monitor fires in **emitted** order under a rig where the first-emitted member returns
last.

## Review notes

`Author` and `Mint` are func types that exist so `entity/runtime` holds no inference wiring and no
graph write path. Both are one line and both are doing real separation work.

The `Entity` struct in `runtime` shares a name with `graph`'s notion of an entity identity and with
`interfaces.md`'s `type Entity struct{ ID EntityID }` — **which is not implemented anywhere.** The ID
is used directly as `ids.EntityID`. Worth deciding whether the one-field struct earns existing.
