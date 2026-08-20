# The Kernel

**Date:** 2026-08-03
**Status: BUILT.** `src/skyra/spacetime/kernel` — `space.go`, `kernel.go`, `kernel_test.go`.

**Supersedes** [2026-07-27/kernel-alteration.md](../2026-07-27/kernel-alteration.md). That doc
was the design; this is what it became, with its superseded vocabulary stripped and three of its
open questions closed. Where the two disagree, this one is right, because this one runs.

> **The kernel executes the Space it is given, records that it did, and returns the record.
> Nothing else.**

---

## 1 · The interface

```go
type Space interface {
    ID() string
    Perturb(origin Space, impulse any) Space
    Observe(origin Space, impulse any, realized Space) Space
}
```

```text
Perturb    returns one ephemeral executable Space, or nil
Observe    receives one realized Space and returns one concrete Space, or nil
ID         the transcript identity — READ TO RECORD, never to decide
```

Both traversal methods are closed over `Space`. Descent and ascent never leave the algebra, so
nothing the kernel touches can be a thing it has to classify.

`ID` is the only accessor, and it exists for the transcript. The kernel reads it to write it
down and for no other purpose — there is no branch anywhere in the implementation that keys on
an ID's value.

**Nil from `Perturb` is the descent base case.** It means *I promoted nothing* — a fixed point
of the promotion, never a judgement about what kind of thing the target is. `Observe` still runs
there, with a nil realized Space, so a leaf gets to say what it concretely is.

**`Express` is not a traversal method.** The name stays reserved for the cognitive inference
operation and its durable record.

## 2 · What the kernel cannot do

Not *does not* — **cannot**. There is no method on `Space` that would let it:

```text
enumerate children          interpret grouping         select or order a frontier
subtract or divide          compare magnitudes         allocate or attenuate
read a cost                 test affordability         infer meaning
```

This is the same discipline `2026-07-27` stated as a rule, held instead by construction. The old
eight-method kernel had `Children() map[string]Space` and therefore *had* to be trusted not to
interpret the map. This one is handed one value with one behavior boundary, and a value it
cannot open is a value it cannot misuse.

The kernel is also **struct-agnostic and stays that way**. Nothing in the package names an
expectation, an action, a support value, or a grade. It moves Spaces and writes down IDs.

## 3 · The whole traversal

**Updated 08-06.** The frames moved off Go's stack into a value — `Stack` — so an execution can be
inspected, stepped, and dropped. `Step` advances it by exactly ONE call into a Space:

```text
FOLD      a child finished; hand its realization to this frame's Observe.
          the frame does NOT pop — it will be asked for another continuation.
DESCEND   Perturb promoted something; push it.
EXHAUST   Perturb promoted nothing; mark the base case.
FINISH    pop, and hand this frame's value up.
```

Uniform granularity is what makes stepping fair: `EXHAUST` and `FINISH` are separate so a leaf
costs one call like everything else, not two.

**`Perturb` ITERATES.** A frame is asked repeatedly and pops only when it promotes nothing — which
is what lets one frame run a pipeline, each phase seeing what the previous one wrote. The cost is
that **a Space must eventually return nil**; one that keeps handing back the same child never
terminates. That was free under a one-shot `Perturb` and is now the Space's responsibility.

**The hop is recorded BEFORE the continuation runs.** A wave that panics or is cut short still
leaves evidence of what it had reached.

### The return path is the tree

`Observe` returns a `Space`, so the fold happens on the stack. There is nothing to assemble and
no second traversal, which is what deleted the entire result-tree apparatus the old kernel
carried — `Node`, `TraversalPair`, `Survivors`, `NextChild`, the tree double-buffer.

Descent is every `Perturb` on the way down; ascent is every `Observe` on the way up, each
receiving what the one below it realized:

```text
perturb:a  perturb:b  perturb:c  ·  observe:c←nil  observe:b←c  observe:a←b
```

## 4 · The transcript, and why the kernel returns it

```go
type Transcript struct {
    WaveID string
    Hops   []Hop
}

type Hop struct {
    SpaceID  string
    Parent   int    // index of the hop that promoted this one; −1 at the root
    Depth    int
    Promoted bool   // Perturb returned a continuation
    Realized bool   // Observe returned a concrete Space
}
```

`WaveID` is what a crossing records when one agent spawns a wave at another, and what a returning
effect cites — the edge of the wave TREE, since a wave never spans agents.

### The kernel is the only honest witness

Runtime Spaces disappear when a wave completes, so something must record them before they go.
That something **may not be the Space itself**:

> **A Space that records its own execution can omit itself.**

Then the audit record is self-reported — the same defect `2026-07-27/FLAGS.md` names for
confidence, *"the one number in the system with no blind assessor"* — except here it is the
record you would use to check everything else. The split that keeps it honest:

```text
the Space controls WHAT it returns
the kernel controls THAT the return is recorded
```

The kernel is the only thing that saw every `Perturb` return. So it witnesses, and it hands the
witness statement back. **This is why the kernel needs no write access to anything.**

### A nil realization does not delete the hop

The two artifacts keep different authority, and this is where it shows:

```text
TRANSCRIPT              authoritative — every Space actually executed
Observe's return        derived — what survived and was realized
```

An observed omission does not erase the corresponding member from the transcript or declare it
false. A Space whose `Observe` returns nil is still recorded as having run.

### It is a SPINE, not a full record of composition

**Updated 08-06.** This section said the spine was a CHAIN, because one `Perturb` meant one child.
`Perturb` now ITERATES — a frame is asked repeatedly and can promote several children in
sequence — so the transcript is a **tree**, and `Depth` and `Parent` are STORED on the `Hop`
rather than derived from position.

That is a cost paid deliberately: it is what lets one frame run a pipeline (set up, fold, act on
what the fold produced), which is how a turn sequences its phases without a scheduler.

The spine claim below is unchanged and still holds — the kernel records only what it was HANDED,
never what a Space contains.

`2026-07-27/kernel-alteration.md` §10 asked the transcript to carry *"grouped member IDs and
boundaries."* Law 8 says the kernel never enumerates grouped members. **Both cannot hold.** Law
8 wins: the kernel reports only what it was handed, and composition inside a grouping Space is
invisible here by construction.

Grouped members still never cross the kernel, and a group that re-enters does so as its own wave
with its own transcript. What changed is only that a frame may promote several children in
sequence, so the record branches and position no longer carries depth or parentage.

If member detail is ever needed, the Space has to volunteer it, and **that field alone goes back
to being self-reported.** Worth marking as an explicit exception rather than blurring this.

## 5 · The Perception Invariant, held

```text
FORBIDDEN   the caller collects the cognitive OUTCOME of its own send
FINE        the caller collects a mechanism record that the wave ran, and where
```

The test is whether a being could use the return to read back what it thought without perceiving
it. IDs and booleans: no. A payload, a `Content()`, or an `Observe` return: **yes**, and at that
moment the transcript becomes a results channel and the invariant is broken.

> **Every field on `Hop` is a string or a bool. That is a constraint on the type, not a
> description of it.**

Two entry points, and the difference is not a convention:

```text
Realize          returns the Transcript. CANNOT reach the concrete result.
RealizeConcrete  returns (Space, Transcript). the assembly boundary, and only it.
```

`Realize` is what a cell calls. `RealizeConcrete` is what the boundary calls — the one caller
entitled to the root's concrete return, because it is what writes that return to the payload log
as the exact grounding of a call. Keeping them apart makes *who may see the outcome* a property
of which method you are able to name, rather than a rule someone has to remember.

## 6 · Concurrency, dissolved

`kernel-alteration.md` §13 Q4 asked about Slice locking if grouped members run concurrently, and
`2026-07-27/FLAGS.md` promoted it to *"Slice concurrency is live again."*

It is gone, and not by being solved:

```text
the kernel holds NO state between waves
each Realize gets its own private wave value
the transcript is RETURNED, never written to shared state
```

Two waves over the same Space cannot collide, because there is nothing shared to collide over.
The contended write the question was about does not exist.

**This does not make concurrency free.** It relocates it. Whatever holds the durable edge set is
still written by concurrent waves, and 07-27 already picked the shape that survives it —
write-once action edges, immutable support values, an append-only evidence set. Append-only needs
no read lock and only a trivial append lock. The guarantee holds **only if nothing
mutable-in-place lands on the durable substrate**: derived magnitudes must live on the ephemeral
node and die with the wave.

## 7 · Interface laws

1. `Perturb` returns exactly one executable Space or nil.
2. The returned Space owns all grouping and group behavior.
3. Nil from `Perturb` is the descent base case, and `Observe` still runs.
4. Every executed Space is recorded before its continuation runs.
5. `Observe` receives exactly one realized Space or nil.
6. `Observe` returns exactly one concrete Space or nil.
7. The kernel never enumerates grouped members.
8. The kernel performs no cognitive arithmetic.
9. The kernel writes nothing. It returns its record.
10. The transcript carries identity and mechanism. Never content.

## 8 · What the tests hold

`kernel_test.go`, and each one is a claim rather than a coverage exercise:

```text
descent then post-order ascent          the return path IS the tree
nil Perturb is base case, still observes termination is a fixed point, not a verdict
transcript is the spine                 chain-positional depth and parentage
nil realization still records the hop   the two artifacts keep different authority
grouped members never cross the kernel  3 members run · 1 hop recorded
only the boundary reaches the concrete  the invariant, as method reachability
concurrent waves share no state         no shared transcript to corrupt
```

The spine test is the one to watch. **If the kernel ever gains the ability to branch, it
fails** — which is where that should surface.

## 9 · Open

```text
GROUPING BEHAVIOR   §13 Q3 — how a Space encodes sequential, concurrent, and join behavior
                    over its own members. it lives inside the Space; its shape is undecided.
                    this is where expectations.md's "siblings concurrent, chains sequential"
                    has to land.

PROPER TIME         the old Slice carried the wave count and nothing owns it now.

THE DURABLE SUBSTRATE  the kernel writes nothing, so something else holds the edge set.
                    append-only, but where it lives and who may write it is open.

ORIGIN AND IMPULSE  both are still `any` at the boundary. the exact types may be refined;
                    the meanings in §1 may not.
```
