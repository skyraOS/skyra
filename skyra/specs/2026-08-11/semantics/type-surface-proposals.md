# Type Surface — proposals from the semantic layer

**Date:** 2026-08-11
**Status: REVIEW.** Every item is a proposal to go through, not a decision. Nothing here has been
applied to the code, and two items exist only to say *do not add this yet*.

> **The `interface changes` directory covers two semantic documents out of eight.** `01` implements
> `kernel-revision`, `02` implements `interface-revision`. **`open-until-commit`,
> `roots-and-admission` and `second-kind-of-credit` have no type-surface proposal at all** — and
> those three are the ones that change the durable model.

Items are numbered so they can be referred to. `A` is *the semantics require this and nothing
exists*; `B` is *an existing type is wrong*; `C` is *remove*; `D` is *the existing change proposals
do not hold*.

---

## A · Required by the semantics, no type exists

### A1 · A possibility must survive its commit
**From:** `open-until-commit.md`

```go
// kernel.Kernel
- heads map[ids.SpanID]int              // A COUNT. the live tree is unrepresentable.
+ // Root holds the tree; heads is derived by walking it.

// kernel.Possibility
+ Children  []*Possibility              // the LIVE tree. Pending holds *graph.Engram, which is
                                        // the durable child, not the live opening.
+ Draft     graph.Expectation           // what it is working under NOW.
+ Revisions int                         // how many times the parent restated it.
```

**`Draft` is the load-bearing one.** `Engram.Expectation` is durable and immutable; a possibility
being refined is working under something that is neither yet. There is currently nowhere for that to
live, so **refinement-in-place is not expressible at all**.

`Revisions` is the only available signal that an expectation was hard to state, and it is a number
the kernel already has.

### A2 · Closure, not commit, is the durable write
**From:** `open-until-commit.md`

```go
// graph
  func (g *Graph) Apply(r *Record)      // called from kernel.commit — MOVES to closure
+ func (g *Graph) Close(rs []*Record)   // one call, the whole chain, then Disperse ONCE
```

`Record` is per-possibility and applied immediately. It has to accumulate on the `Root` and land in
one go, or *every durable write is caused by the root's commit* is false in code.

**And `Possibility.Selected` changes meaning.** The forfeit is expectation-scoped now, so a revision
resets it. One line, large consequence, and it belongs written beside the field:

```go
Selected bool   // the origination forfeit. set at the first select UNDER THE CURRENT DRAFT.
                // a revision resets it — the parent superseded what came before, and only the
                // party the rule protects can do that.
```

### A3 · Copy has no function
**From:** `second-kind-of-credit.md` — the decision is made and nothing implements it.

```go
// graph
+ func (g *Graph) Copy(e *Engram) *Engram
```

Contract, which is where *never copy the credit* becomes structural rather than remembered:

```text
NEW ID
SAME Expectation
SAME children map                    shallow — this is REFINEMENT depth
invocation records COPIED with       Cited = false · Acceptances = nil
SAME observation objects             not re-minted. the decision of 2026-08-11.
```

**Deliberately not added:** a `From` back pointer. `second-kind-of-credit` lists it as undecided, and
the type surface is exactly where an undecided thing gets decided by accident.

### A4 · The permission model — nothing exists
**From:** `roots-and-admission.md`. Not one type in the tree.

```go
// ids
+ PrincipalID string

// a new package — namespace
+ type Namespace struct {
      Principal ids.PrincipalID
      Mounts    map[string]*Mount
  }
+ type Mount struct {
      Name   string
      Engram *graph.Engram
      Grants []Grant
      Source MountSource
  }
+ type MountSource int          // Ambient · Granted · Inherited
+ type Grant string             // prefix form. fs:read:/p/** · spawn:<bin> · introspect:self
+ type Resolver interface {
      Resolve(name string) (*graph.Engram, bool)
      Names() []string
  }

// kernel.Root
- Engram *graph.Engram          // roots-and-admission: A ROOT IS A NAMESPACE
+ NS     *namespace.Namespace
+ Grants []Grant                // interfaces.md declares this. it was never built.

// kernel.Kernel
- Builtin map[string]*graph.Engram    // one global mount table
// frame.Live
- Builtin map[string]*graph.Engram
+ Mounts  Resolver                    // per principal
```

**And a behaviour change, not just a rename.** `frame.Live` resolves *children → delegated →
builtin*. `roots-and-admission` requires:

```text
1  DELEGATED   this invocation only.
2  MOUNTS      granted.
3  CHILDREN    earned.
```

Earned entries shadowing mounts is **PATH hijacking** — a principal authors its own `write` and
everything provisioned for it routes through structure it wrote.

**No `Provision` field is proposed**, because *what a memory provision is* still has no shape and a
field would be inventing one.

### A5 · The world boundary has no check
**From:** `roots-and-admission.md`, and `kernel-revision.md` already says it: *the second
authorization is NOT BUILT AT ALL.*

```go
// kernel/authorize.go
+ func authorizeEffect(r *Root, e *Effect) error    // grants. the second authorization.
  func ownership(rec *graph.Record) error           // becomes REAL once memory is shared
```

`Root.Stage` holds an effect with no authorization anywhere on the path. `ownership()` is the
documented no-op written to hold this place — *"if engrams ever become shareable, this stops being
free and must become real."* A mount is shareable memory.

---

## B · Existing types that are wrong

### B1 · `Effect` has two booleans for four states
**From:** `kernel.md`'s rule that every staged effect is explicitly disposed.

```go
- Landed    bool
- Discarded bool
+ Disposition Disposition   // Staged · Landed · Discarded · Failed
```

**This fixes a live defect rather than tidying.** `root.go:142` — a write that *fails* is left
neither `Landed` nor `Discarded` and records no discard, so it silently occupies a fourth state the
booleans cannot name. Unreachable is handled; failed is not. An enum makes forgetting it a compile
error.

### B2 · `Step` cannot supply what `Events.Crossed` asks for
**From:** `interface-revision.md`, against `interface changes/02`.

`02` proposes both of these:

```go
Step(d Delivery) (grammar.Emission, error)
Crossed(span, e, ent, r Rendering, raw string, err error)
```

The kernel calls `Crossed` and **has no `raw`** — nothing returns it. And `interface-revision`'s own
defect list includes *"the raw response is parsed inside the adapter"*, which `02` does not move.

```go
+ Step(d Delivery) (Raw, error)     // Raw = the response. THE KERNEL PARSES IT.
```

Parsing the grammar is not reading content — the grammar is the kernel's own vocabulary and it
already type-switches on the result. This completes the argument `interface-revision` makes, supplies
`Crossed`, and makes replay diffable in **both** directions rather than one.

Note that `grammar.ErrNoEmission` already carries `Raw` on *failure*. The kernel holds the response
when parsing fails and not when it succeeds, which is the asymmetry this removes.

---

## C · Removals the semantics justify

```text
C1  graph.Summary       marshalling is entity-side. under interface-revision, Assemble is where a
                        child set becomes a prompt — so a PRESENTATION type in the DURABLE package
                        is misplaced. it also leaks across frame.View, which exists to prevent
                        exactly that kind of reach.

C2  graph.UsedSeen      two ints and a ratio. a function.

C3  graph.Diff          one method on Partition. KEEP Partition — it earns it; the run showed it
                        separating a right pairing from two wrong ones with an identical shape.

C4  host.Handle         cut to {Host, Ref}. Stdout · Stderr · Exit is a shell's shape inside a
                        class interface, and Spawn has never run. world.md's warning about the one
                        hand-designed surface applies exactly: a guess nobody was forced to make.

C5  entity.Entity       interfaces.md's `struct{ ID EntityID }`, never implemented. roots-and-
                        admission confirms it should not be — an entity has NO owner and no
                        authority of its own, because authority is on the MOUNT. the ID is the
                        whole type.

C6  runGroup's          conversation-runtime-mismatch row 1, RESOLVED toward the conversation:
    G.Store(ret)        errors are ephemeral. also one of the two writes outside a commit.

C7  Ingress's           the other one. credit.md: an inbound observation is not recorded until the
    G.Store(o)          possibility that used it commits.
```

C6 and C7 together retire the claim in `kernel-revision.md`'s *unchanged* block that no path writes
durably outside a commit — which is currently false, and false in a way `conversation-runtime-
mismatch` recorded separately without either noticing the other.

---

## D · Where the existing change proposals do not hold

### D1 · `01` keeps `Ingress` with its current signature

Correct under `kernel-revision` — it is the one channel. But `roots-and-admission` makes ingress **a
select resolved against the root's namespace**, so it cannot stay:

```go
func (k *Kernel) Ingress(c content.Content) *graph.Observation
```

It needs the principal and the profile, and **that is the gate `questions.md §5` is missing.** `01`
was written before `roots-and-admission` existed.

### D2 · `01`'s `Topology.Rooted()` cannot be subtree-bounded

```go
type Topology interface {
    Heads(from ids.SpanID) int
    Depth(from ids.SpanID) int
    Rooted() (user, kernel int)     // ← global. no `from`.
}
```

`kernel-revision` argues introspection is **bounded by the subtree** — down never up, never sideways.
`Rooted()` is a global count with no bound available, and it is `Spiral` surviving under a new name.

Either it is not introspection and belongs to whatever reads the spiral out of band, or the subtree
bound is not absolute. Both are decisions; neither has been made.

---

## What is deliberately absent

```text
NO `From` BACK POINTER      second-kind-of-credit lists it undecided. adding a field picks an
                            answer.

NO DEEP-COPY PRODUCER       same doc, same reason. refinement does not hit it; discovery does.
RULE

NO `Provision` FIELD        "you may accumulate HERE" has no HERE yet.

NO ESCALATION TYPE          a grant miss escalates by opening a KERNEL-ROOTED chain, and
                            Root.Origin / Root.Provoker already express that. nothing new is
                            needed until a permission engram exists to be selected.
```

---

## Reading order, if going through this against the code

```text
A1 · A2   the loop. everything else in A sits on them.
A4 · A5   the boundary. nothing about it is built and nothing about it is checked.
A3        one function, and the contract is the whole of it.
B1 · B2   two defects, both structural rather than stylistic.
C         seven removals, none blocked by a decision.
D         two questions to answer before 01 is applied.
```
