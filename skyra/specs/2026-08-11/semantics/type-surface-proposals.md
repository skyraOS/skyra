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
- heads map[ids.SpanID]int              // A COUNT, and it counts the wrong thing — see below.

// kernel.Possibility
+ Children  []*Possibility              // the LIVE tree. Pending holds *graph.Engram, which is
                                        // the durable child, not the live opening.
+ Draft     graph.Expectation           // what it is working under NOW.
+ Returned  *graph.Observation          // nil = has not committed under the current draft.
+ Revisions int                         // how many times the parent restated it.
```

**Heads become derived, and the counter was wrong for its stated purpose.** It increments on entering
`open()` and decrements on leaving, so a depth-five chain with one leaf running reports five — four
of which are ancestors sitting in `runGroup`. A *stalled head* is therefore indistinguishable from
*a parent legitimately waiting*, which is the eviction decision the number exists for.

```text
head  ==  no live children  AND  Returned == nil
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

`Record` is per-possibility and applied immediately. It has to accumulate on the `Branch` and land in
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
**From:** `roots-and-admission.md`, as revised by `terminology.md`. Not one type in the tree.

**Much smaller than an earlier draft of this doc proposed.** A namespace is an **engram**, so most of
what looked like new machinery is machinery that already exists.

```go
// ids
+ PrincipalID string
+ type Grant string        // prefix form. fs:read:/p/** · spawn:<bin> · introspect:self

// graph — a namespace is an ENGRAM. no new type, one new population.
+ namespaces map[ids.EngramID]*Engram    // held apart from the index, like roots are today

// kernel — THE RACK. flat, keyed by the namespace engram. admin is a row on it.
- type Root struct{ … Engram *graph.Engram … }
+ mounted map[ids.EngramID]*Tree

+ type Tree struct {                 // all live work under one namespace
      Namespace *graph.Engram
      Entity    ids.EntityID         // who is manning it
      Grants    []Grant              // provisioned at sign-in. ONE per tree.
      Reach     []ids.EntityID
      Branches  map[ids.SpanID]*Branch
  }

+ type Branch struct {               // ONE ASK. the transaction boundary.
      Span   ids.SpanID
      Top    *Possibility            // the interior. possibilities nest; BRANCHES DO NOT.
      staged []*Effect               // per ask — accepting one must neither land nor discard
  }                                  // another's writes. this is what forces Branch to exist.

// graph.Engram — for a namespace only
+ Ceiling []Grant                    // DURABLE. the maximum it may ever be mounted with.

// kernel.Kernel
- Builtin map[string]*graph.Engram        // one global mount table for every principal
// frame.Live
- Builtin map[string]*graph.Engram
```

`Builtin` goes with nothing replacing it: **ambient names are children of each namespace, placed at
boot.** `discover` nameable in every frame becomes `discover` being a child everyone has.

**And the select provisions.** This is the mechanism that replaces mounts entirely:

```go
// grammar.Member
+ Grants []Grant     // handed to this child. DEFAULT IS NOTHING.
+ Reach  []ids.EntityID
```

```text
DEFAULT = EVERYTHING I HOLD   inheritance with extra syntax. the act would be decorative.
DEFAULT = NOTHING             least privilege by construction. a child needing fs:read must be
                              given it; a parent that forgets sees a loud failure it learns from.
```

Under-granting is self-correcting — the child fails, it is recorded, selection learns. Over-granting
is bounded by your own ceiling, which is the object-capability guarantee. So *nothing* is safe and
its failure mode is cheap.

**`kernel.go:173` is the live bug this closes.** `Drive` passes `k.Reg.Known()` as the chain's reach
— every principal can name every registered entity. Reach comes from provisioning, never from the
registry, which becomes what it should have been: how the kernel *reaches* an entity it was already
authorized to name.

### What A4 no longer proposes

```text
GONE   namespace.Mount · MountSource · Resolver-with-origin · three-place resolution order ·
       mounts-shadow-earned · an Origin return on frame.Resolve · a Provision field

WHY    mount was five existing mechanisms under one word — a child placed by provisioning ·
       ambient names placed at boot · grants on the chain · entity reach · copy.

       and grants-on-the-reference was specifically wrong: EFFECTS HAPPEN AT LEAVES, so
       authority attached to a name at the top bounds nothing unless every intermediate
       re-narrows it. grants bound the CHAIN. Root.Grants was always the right home.

       PATH hijacking dissolves with mounts — one child map, no shadowing order to get wrong.
       the collision moves to provisioning time, where "who may write into my namespace" is
       already answered by commit-is-a-request.
```

### A5 · The world boundary has no check
**From:** `roots-and-admission.md`, and `kernel-revision.md` already says it: *the second
authorization is NOT BUILT AT ALL.*

```go
// kernel/authorize.go
+ func authorizeEffect(b *Branch, e *Effect) error  // grants. the second authorization.
+ func authorizeGrants(parent, handed []Grant) error // a select may only NARROW
  func ownership(rec *graph.Record) error           // becomes REAL once memory is shared
```

`Stage` holds an effect with no authorization anywhere on the path.

**`authorizeGrants` is the whole of the ocap guarantee** and it is one subset check. Everything else
in the model rests on it holding: boot provisions admin, admin provisions namespaces, a select
provisions its children, and none of the three can widen. `ownership()` is the
documented no-op written to hold this place — *"if engrams ever become shareable, this stops being
free and must become real."* A mount is shareable memory.

---

### A6 · Boot has no method
**From:** the admission bootstrap.

```go
+ func (k *Kernel) Boot(cfg Config) (*graph.Engram, error)   // returns the ADMIN namespace
```

Boot's whole job, and it should stay this short:

```text
1  mint the system engrams — discover · introspect · authenticate · provision · create
2  register the entities that man them
3  mint the ADMIN NAMESPACE, place those engrams as its children, and WRITE ITS CEILING
```

`Select` and `Commit` are **not services** — they are forms, parsed and never initialized. The
distinction matters because a boot that initializes "the two forms" will grow into a second
configuration language.

> **Boot holds the admin namespace's grants. That is the axiom, and it should be one line.**

Same shape as `here` and as `discover` being nameable everywhere: the thing you cannot derive because
you are standing on it. `interfaces.md` already has its ancestor — *the runtime must know how to
reach at least one entity that can author an expectation* — and this grows it by exactly one clause.

Everything else provisions downward from there: admin provisions namespaces, a namespace's chain
provisions its children, and `authorizeGrants` is what makes each step narrow.

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

Correct under `kernel-revision` — it is the one channel. But it cannot stay as written:

```go
// NOW
func (k *Kernel) Ingress(c content.Content) *graph.Observation

// NEEDED — it is MOUNT. attach a namespace to the rack; no content, because the payload is an
// argument to the select that follows.
func (k *Kernel) Mount(ent ids.EntityID, ns *graph.Engram, grants []Grant) (*Tree, error)
func (k *Kernel) Unmount(ns ids.EngramID)
```

The one kernel check is `grants ⊆ ns.Ceiling` — arithmetic, no semantics, and **no live admin
required**, which is what keeps the rack flat. `Ingress` described a model where the method minted an
inbound observation; it no longer does.

**That is the gate `questions.md §5` is missing.** `01` predates every document that says so.

### D2 · `01`'s `Topology.Rooted()` cannot be subtree-bounded

```go
type Topology interface {
    Heads(from ids.SpanID) int
    Depth(from ids.SpanID) int
    Rooted() (user, kernel int)     // ← global. no `from`.
}
```

`kernel-revision` argues introspection is **bounded by the subtree** — down never up, never sideways.
`Rooted()` is a global count with **no node to start from**, and it is `Spiral` surviving under a new
name.

The bound is not a rule to enforce: descending can only reach descendants, so it is a fact about
which way the pointers point. Which makes the shape concrete:

```go
+ type Topology interface {
      Heads(from ids.SpanID) []Head      // leaves with Returned == nil
      Done(from ids.SpanID)  []Head      // leaves with Returned != nil
  }
- Rooted() (user, kernel int)            // global. no `from`. not introspection.
```

`Head` carries span, engram **ID**, entity, host, depth and time-in-Step — **topology, never
content.** The caller dereferences the engram against their own graph. That is also what retires
`shortExpr` and `oneLine`: they exist because the kernel wanted to say something readable and had no
other channel.

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
A4 · A5   the boundary. nothing about it is built and nothing about it is checked, and A5's
          subset check is the whole ocap guarantee in one function.
A3        one function, and the contract is the whole of it.
B1 · B2   two defects, both structural rather than stylistic.
C         seven removals, none blocked by a decision.
D         two questions to answer before 01 is applied.

Read `terminology.md` first. It retires ROOT and MOUNT, and both appear throughout the docs this
one cross-references.
```
