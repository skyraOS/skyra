# Kernel

**Date:** 2026-08-11 · revised 2026-08-12
**Status: CHANGE PROPOSAL.** Code only — signatures, types, and what to delete.

**Why:** [`../../semantics/the-collapse.md`](../../semantics/the-collapse.md) ·
[`roots-and-admission`](../../../2026-08-11/semantics/roots-and-admission.md) ·
[`terminology`](../../semantics/terminology.md)

> **Revised 2026-08-11, later.** `Ingress` is now `Mount`, and `Topology` lost `Rooted()`. Read
> `terminology.md` first — **ROOT** is retired throughout, and what it named is a **BRANCH**.
>
> **Revised 2026-08-12.** A section added on WHERE, which is not vestigial in the code — it is
> absent. Nothing routes, and three things would have to exist before anything could.

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
+ func (k *Kernel) Mount(ent ids.EntityID, ns *graph.Engram, grants []Grant) (*Session, error)
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
+ mounted  map[ids.EngramID]*Session          // THE RACK. flat, keyed by namespace. admin is a row.

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

**`Session` replaces `Root`**, and the full shape is in
[`type-surface-proposals`](../../semantics/type-surface-proposals.md) `A4`. A session is all live work
under one namespace. **A branch is not a type** — it is a possibility with no parent, and staged
effects are partitioned by its span on the session.

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

## WHERE does not exist

Not "partly built" and not "vestigial". There is no host resolution anywhere in the tree.

```go
type Registry struct {
    interpreters map[ids.EntityID]func() Interpreter   // NO HOST. anywhere.
}

type Kernel struct {
    Host host.Host        // ONE host. a field, not a table. no method takes a HostID.
}

type Member struct {
    Host ids.HostID       // written by the grammar. NEVER READ — checked every reference.
}

Invocation.Host = k.Host.ID()   // hardcoded to the single local host on every invocation.
```

`world.md` says *"you invoke an entity, not a host; delivery picks a host on the way out."* There is
no way out to pick on. `Registry.Open(id)` hands back an interpreter with zero host involvement.

Three things are missing, in dependency order:

```text
1  AN ENTITY → HOSTS TABLE    which hosts can reach entity e, plus capacity. a REGISTRY FACT —
                              runtime state, rebuildable from configuration, never durable.
                              routing reads (entity, capacity) and NOTHING ELSE, or it is
                              reading semantics.

2  RESIDENCY, PER ENTITY      where the implementation IS, which is independent of reachability:
                              a remote host can REACH salesdb@mac-A over the network and cannot
                              BE where the data is. affinity is this, not a reachability count.
                              needs 1 to mean anything.

3  Spawn AS A PLACE           see below.
```

## `Spawn` names a medium, twice

```go
// CODE
Spawn(argv []string, dir string) (Handle, error)

// interfaces.md
Spawn(x Expectation, f Frame) (Handle, Streams, error)
```

`argv`/`dir` is Unix command-line vocabulary inside a capability class — the same leak as
`host.Handle` carrying `Stdout · Stderr · Exit`, which `C4` removed. The return was cut and the
parameters were left.

And the spec's version carries the **work**, which is one layer too clever if spawn means *give me a
place*: the work is the engram, and the engram is virtual.

```go
+ Spawn() (Handle, error)     // "somewhere to do work." no argv, no expectation, no frame.
```

Any argument describing the work drags semantics below the boundary with it.

```text
NOT SETTLED — FROM THE 08-12 CONVERSATION, RECORDED SO IT IS NOT DECIDED BY ACCIDENT

THE CUT           engrams = work that can be done VIRTUALLY. host APIs = work that CANNOT.
                  a syscall cannot be separated from the machine it is on, so it cannot be an
                  engram: Invocation.Host is RECORDED, NOT KEYED, so anything whose answer
                  depends on the machine is under-specified by construction and the fold has no
                  slot for what makes it differ.

WHAT FOLLOWS      a host would be REACH + SPAWN, and StorageOps would not be on it. Two messages
                  gave read/write opposite homes — first "an entity is its read(), write()", then
                  "host APIs = work that cannot be done virtually". WHICH STANDS IS NOT DECIDED,
                  and it decides whether host.StorageOps survives at all.

WHAT DOES NOT     engrams stop at the entity. `salesdb@mac-A` is an entity — it interprets — and
CHANGE EITHER WAY the syscalls inside it never surface as engrams. The instruction is virtual;
                  the read is bound.
```

## Code to delete

```text
kernel/kernel.go   indent · shortExpr · oneLine                        34 lines
kernel/kernel.go   9 Run.Logf calls
```

`shortExpr` reads `Expectation.Expression` and `oneLine` reads `Observation.Content`. They are the
only places the kernel reads content. Replaced by the sink in [entity.md](entity.md).
