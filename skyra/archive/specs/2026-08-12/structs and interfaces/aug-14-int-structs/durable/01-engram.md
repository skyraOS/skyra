# durable/01 · The Engram

**Package:** `durable`

Contracts plus continuation. P1, and the type should make anything else hard to add.

---

## 1 · The shape

```go
type Engram struct {
    ID ids.EngramID

    // THE TWO SEMANTIC PARTS. P1.
    Contracts    map[string]*contract.Child   // keyed by child position name
    Continuation contract.Continuation        // nil until solidified

    // THE TWO SLOTS. fixed at solidification. ../creator-and-compiler.md §2, §3.
    Creator  ids.EngramID   // authored the body. EMPTY FOR A LEAF — no creation half.
    Compiler Address        // WHERE the compiler is. it compiles the causal code and calls
                            // the kernel. it holds no protocol methods of its own.

    // THE TYPE THIS ENGRAM RETURNS, exposed to the kernel as a string.
    // what it EXPECTS BACK per position lives in Contracts[pos].Return — same string, other side.
    Returns string

    // HISTORY
    History []*Invocation
    Prev    *Engram                           // the version chain. NEVER in content.

    // DERIVED
    Engrams  map[string]*Engram                    // children, keyed by name
    ByReturn map[ids.ObservationID]*Invocation     // the walk's entry point
}

```

```text
ENGRAM = CHILD CONTRACTS + COMPILED CONTINUATION
```

Everything after those two fields is a slot, history, or derived. `ByReturn` keeps the status
`../../03-graph.md` gives it and `../../../reach/age.md` reaffirms: derived, rebuildable, deletable.

### The compiler compiles; it does not hold the protocol

**REVISED 08-15 — an earlier pass in this session put `Decide`, `Recognize`, and `Remember` on a
`Compiler` interface. That was wrong and is reverted.**

```text
DECIDE · REMEMBER · RECOGNIZE   ARE KERNEL METHODS. kernel/01-protocol.md.
THE COMPILER                    compiles the causal code and CALLS them.
```

The causal syntax compiles to syscalls, the way any language compiles to syscalls. A compiler that
held the causal forms would be extending causal structure, which `primitives.md` reserves:
*"Decide and Remember... are the only two primitives that change the live causal structure."*

So `Compiler` is an **address** — where that compiler lives — not an interface. It connects to the
process/endpoint model in `../creator-and-compiler.md` §9 rather than to the kernel's method set.

### Types are strings the kernel compares

```text
THE PARENT STORES   Contracts[position].Return — the childType it expects back.   A STRING.
THE CHILD EXPOSES   Returns — the type it produces.                               A STRING.
THE KERNEL          compares them. that is the whole of Recognize.
```

Two sides of one comparison, so there is no second map — `ChildTypes` would have duplicated
`Contracts[pos].Return`, and two places for one fact eventually disagree.

The child returns; **the first thing the kernel does is call `Recognize`**, and the kernel already
holds both strings, so it never touches content to do it.

```text
THIS DOES NOT BREACH P4.  the label is on the ENGRAM — durable, substrate-held — not written into
                          observation CONTENT by the child. A child cannot forge what it does not
                          author. That is the distinction P4 is actually about.
```

### Compilers compose, one per syntax

```text
├── the runtime's compiler    skyra continuation language
├── python's compiler
└── further syntaxes as they earn a place
```

Each compiles its own syntax down to the same kernel calls. The selector is not learned but
**determined** — the body's syntax picks its compiler, per `../creator-and-compiler.md` §3.

```text
⚠ CONTESTED   these two fields put something on an Engram besides contracts and continuation,
              which P1 says is all an engram IS. Either P1 is wrong, or the slots describe how an
              engram is REACHED rather than what it is — the second preserves P1 and is probably
              right, since the compiler is DETERMINED by the body's syntax rather than chosen.
              NOT RULED. see ../00-principles.md P1.
```

## 2 · What is absent, and why

```text
NO Expectation FIELD    it lives on the PARENT'S SIGNATURE for the position, not on the child.
                        ../contracts/02-child-contract.md §3. A free-floating expectation on the
                        engram would be the child declaring what it is, which P4 refuses — and
                        putting it on the position is what answers H12.
                        REAL DIVERGENCE FROM ../../03-graph.md. see §5.

NO Kind FIELD           P6. leaf and orchestrator are positions.
                        LEAF == len(Contracts) == 0.

NO Age FIELD            age is positional. ../identity/04-position.md §2.

NO Horizon FIELD        authority is not a property of durable behavior. it is provisioned
                        downward per decision.
```

Both are the same error: putting a per-invocation fact on the durable object, where it becomes a
default and then a policy.

```text
REVERSED 08-15 — "NO Entity FIELD"

This document previously refused an interpreter field, on the grounds that an entity is bound per
Decide and an engram is not owned by one. ../creator-and-compiler.md §3 overturns it: `Entity`
dissolved, and the COMPILER is determined by the body's syntax rather than chosen per invocation.
A python body has one compiler forever, so binding it per Decide was never meaningful — and the
old shape let a parent bind an incompatible interpreter, which is now inexpressible.

THE OLD REASONING STILL HOLDS FOR THE THING IT WAS ABOUT. Nothing per-invocation goes on the
engram. The compiler is not per-invocation; that was the mistake.
```

---

## 3 · One parent, structurally

```text
AN ENGRAM LIVES IN EXACTLY ONE PARENT'S MAP, so two parents structurally cannot share one.
```

That is what makes locality structural rather than argued: *A never writes into B's memory, A writes
into the child engram A owns.*

```text
CARRIED FOR REVIEW, unresolved since 08-11:

conversation-runtime-mismatch.md flags that engrams are currently reused BY POINTER AND ID, where
the conversation has reuse COPYING the expectation into a new locally owned engram. Copy-on-reuse
would make the one-parent rule true BY CONSTRUCTION rather than by discipline, and interfaces.md
already argues duplication is nearly free because expectations never improve.
```

**Under the new model this question sharpens rather than resolves.** If reuse copies, then the
copy's `Prev` pointer and the original's version line diverge, and `age.md`'s DEEP COPY note applies
— it explicitly leaves `Producer` across a copied subtree undecided, while noting that versioning is
*refinement*, which shares children rather than copying them, and so does not hit the problem.

```text
VERSIONING   shares children. does not hit the copy question.
REUSE        may or may not copy. still open. THE TWO MUST NOT BE CONFLATED LATER —
             age.md records this explicitly so they are not.
```

---

## 4 · The invocation

```go
type Invocation struct {
    Span         ids.SpanID
    Position     Position                      // WHERE in the parent. new.
    Version      Version                       // which version this ran through. new.
    Observations []*observation.Observation    // exactly what was supplied
    Entity       ids.EntityID                  // KEYED
    Host         ids.HostID                    // RECORDED, NOT KEYED
    Cite         ids.ObservationID             // the candidate observation that prompted the choice
    Return       *observation.Observation
    Remembered   bool                          // was RESULT.Cited
    Acceptances  []Acceptance
}
```

Mostly carried from `../../03-graph.md`. Three changes:

```text
ADDED Position   the emitted position inside a concurrent group, which
                 observable-causality.md lists among what the substrate MUST preserve.

ADDED Version    which version governed this invocation. observable-causality.md's REVISION —
                 "which draft governed each invocation."

RENAMED Cited    → Remembered. the local verdict is now named for the operation that renders it.
                 same field, same meaning.
```

`Entity` keyed and `Host` recorded-not-keyed carries over unchanged, and matters: two askings are
the same question if the entity matches, regardless of where it ran. That is what makes the
same-question fold work across hosts.

**`Cite` changes meaning without changing type.** In the runtime it is *"a GROUPING KEY ONLY"* with
`questions.md §15` attached. Under the new model it is the candidate observation Decide selected —
which is what `../kernel/03-think.md` §5 requires, so that the thought which authored a behavior stays
a constituent of its execution.

---

## 5 · Reconciliation with the runtime

```text
graph.Engram.Expectation   →  REMOVED. expectation moves into parent-owned contracts. §2.
graph.Engram.History       →  unchanged.
graph.Engram.Engrams       →  unchanged.
graph.Engram.ByReturn      →  unchanged, still derived.
graph.Expectation          →  the struct with "NO LANGUAGE TAG. a tag would be a self-report."
                              the REASONING survives verbatim into the return contract; the
                              standalone type does not.
graph.Invocation.Cited     →  Remembered.
                          +  Position, Version added.
```

The expectation removal is the one to weigh. `../../03-graph.md` calls these five types *"the
smallest thing in the tree and the most load-bearing,"* and this takes a field off the most
load-bearing of them.

```text
THE ARGUMENT FOR    P4 and P2. a contract is owned by the parent; an expectation on the child is
                    the child stating what it is, which is the thing typing must not rest on.

THE ARGUMENT AGAINST  an engram with no expectation is hard to READ. discovery searched
                      expectations lexically, and a root's expectation IS the ask — which is the
                      whole argument in ../../../semantics/the-collapse.md for holding roots apart
                      from what discovery walks.

NOT RESOLVED. If discovery needs a lexically searchable expectation on the engram, then either it
searches contracts instead, or the field comes back as derived-and-uncited. RECORDED AS H12.
```
