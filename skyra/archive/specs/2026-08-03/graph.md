# The Expectation Graph

**Date:** 2026-08-03
**Status: PARTLY SUPERSEDED — 08-06.** The composition rule is live and is this doc's real
contribution. The edge and arithmetic sections have been overtaken.

**Supersedes** [2026-07-27/expectations.md](../2026-07-27/expectations.md), which had two node
types and grading at leaves.

```text
STILL LIVE      composition-derived typing (directory | file), the universal perceive,
                graded-not-settled, no empty directories, the stack anchors at directories

SUPERSEDED BY families.md    "every edge is an OBSERVATION" and "edges are records, not
                             scalars." edges carry SCALARS — support [-1,1] or confidence
                             [0..1] — and content lives on nodes. the word "observation" is
                             retired in favour of impulse / interpretation / edge.

SUPERSEDED BY closure.md     everything below treating P as closure. P is a RETRIEVAL
                             quantity; closure is the most recent grade in the current
                             execution, and is +1 or it is not closed.
```

---

## The rule

```text
EVERY DURABLE NODE IS AN EXPECTATION
EVERY EDGE CARRIES A GRADE          ← was "is an OBSERVATION"; see families.md
```

One type of thing to be, one type of thing to connect with. The collapse
`2026-07-27/expectations.md` predicted has happened, though not the way it guessed — it expected
actions to become leaf expectations. Instead actions stop being nodes at all.

## Expectations and actions are a COMPOSITION, not two types

Following the filesystem the whole way down:

```text
EXPECTATION   a DIRECTORY    has entries. can be traversed into. can anchor the stack.
ACTION        a FILE         a leaf. is executed. cannot be traversed into.
```

Unix does not have a directory type and a file type. It has **inodes**, and directory-ness is a
property of what the inode contains — a directory is a file whose content happens to be a list
of entries. Same here, and one better: the distinction is **derived from composition**, so not
even a mode bit is stored.

```text
has entries    → it is a directory → it is an expectation
has none       → it is a file      → it is an action
```

This is [2026-07-27/expectations.md](../2026-07-27/expectations.md)'s *"the typing is derived,
not declared"* arriving at the same place from the other end, and it is the collapse that doc
predicted, licensed by the one condition it named:

> *actions and expectations are separated here for one reason: children of a node must run
> sequentially when they are actions and concurrently when they are expectations… Deferred, not
> forgotten.*

Concurrency is deferred, so the reason was void and the collapse was free. **Updated 08-06:** the
condition came back as the parent's ORDERING POLICY — sequential or concurrent, declared by the
parent — which is what `expectations.md` predicted (*"a flag the parent declares"*) and what
answers kernel §13 Q3.

### What falls out

**No empty directories.** Every expectation carries its own perceive (below), and a perceive is
an action, so every directory contains at least one file. The falsifiability invariant restated:
an expectation with nothing in it is not merely forbidden, it is not a directory.

**Directories are never executed.** A directory's content is its entries. An expectation acts
only through the actions it contains — it has no execution of its own to grade, which is why
grading arrives entirely through its files.

**The stack can only anchor at a directory.** You cannot `cd` into a file. The active expectation
is always a directory, and an action is something you run from where you are standing rather than
somewhere you can stand.

**`Q` has an edge again.** The directory entry *is* the expectation → action edge, so
`Q(E,A) = E[P_{t+1}(E) − P_t(E) | E_t, A]` has a home without a carve-out and without actions
becoming a second node type. This was [questions](questions.md) Q2.

## Every expectation carries its own perceive

> **An expectation always comes with a perceive expectation. Therefore every expectation is
> graded.**

This is [2026-07-27/expectations.md](../2026-07-27/expectations.md)'s falsifiability invariant
promoted from a check to a **constitutive rule**:

```text
BEFORE   an expectation with no perceive in its own sequence, and no child with one,
         cannot be wrong. checkable structurally, so it can be enforced.

NOW      an expectation cannot be constructed without its perceive.
```

The difference is the same move the kernel made. Do not enforce a rule — remove the ability to
break it. There is no unfalsifiable expectation to detect, because there is no way to make one.

### Graded, not settled

**Vocabulary, and it is load-bearing.** The old model *settled* — a bet resolved once, terminally,
and `ResolvedBy` closed the record. This model **grades**, continuously. The evidence set is
append-only, so a new observation can always arrive and `P` moves by accretion forever. Nothing
reaches a final state.

```text
SETTLE   terminal. the record closes. belongs to the old model.
GRADE    continuous. support accumulates. the record never closes.
```

So there is no settler here. What that mechanism was reaching for survives as the perceive
itself — dnav3's version of the idea, which is the ancestor of this rule:

> *the being goes and LOOKS, so absence becomes a fact rather than a silence, and UNKNOWN is a
> waiting room rather than a destination.*

Universal pairing makes the looking mandatory rather than merely wise. What it no longer does is
end anything.

## An ungraded expectation has no consequences

This is where the filesystem analogy stops, deliberately. A file that nothing reads still sits
there costing nothing and meaning nothing. **The graph is supposed to have consequences.**

```text
graded      it moves something. it is part of the causal structure.
ungraded    inert. no support, no position, no effect on anything downstream.
```

Under the universal-perceive rule the second category is **empty**, which is the point. Every
node in the graph is a node that does something, because a node that could not be graded could
not have been built.

It also strengthens [trust](../2026-07-27/trust.md)'s *"ungrounded chains are worth nothing."*
That was an arithmetic property — an expectation with no observation beneath it contributes `0`,
so rumination is suppressed without a rail. Now it is a structural one: there are no ungrounded
expectations to suppress. The perceive guarantees grounding **arrives**, so absence becomes a
graded fact rather than an indefinite silence.

## Edges are records, not scalars — SUPERSEDED

An edge being an observation means the structure *is* the evidence:

```text
you cannot have an edge without having observed something
```

[trust](../2026-07-27/trust.md)'s *"observations are the only source of grounding"* becomes
literally true rather than a discipline. There is no second kind of edge that links two
expectations on some other authority.

Three consequences of edges being observations:

```text
WRITE-ONCE     an observation is immutable once written, so an edge is too.
APPEND-ONLY    new observation = new edge. the edge SET grows and never rewrites.
MULTIGRAPH     several observations may bear on the same pair. they are distinct
               observations, so they are distinct edges, and T sums over both.
```

The arithmetic is unchanged by any of this: `c` is still the product of edge values along a
path, `T = Σ|c|` over **distinct observations**, `G = Σc`, `P = G/|O|`.

## Two things this fixes

### Toy 5's contestation is available again

[Toy 5](../2026-07-27/toy-runthroughs.md)'s 07-30 update declared two of its own pass conditions
unsatisfiable:

> *Under grading-at-leaves, `o*` cannot attach to `E "the route remains safe"` at all — that
> expectation is grounded through its children, not directly. So pass conditions ② and ④ as
> written are unsatisfiable.*

Grading-at-leaves is gone. Every expectation has its own perceive and is therefore graded
directly, so `o*` **can** bear on `E` and **can** make it contested. Conditions ② and ④ go back
to being satisfiable, and revision by accumulated disconfirmation returns alongside revision by
rivalry.

Toy 5 is not closed — ingress is still the open half, and the being still has to be standing
somewhere the observation can land ([execution](execution.md)). But the toy's own workaround is
no longer the only path.

### Closure stops being capped by depth — SUPERSEDED by closure.md

[FLAGS](../2026-07-27/FLAGS.md) — *"an expectation behind a `0.7` edge cannot exceed `P = 0.7`
however completely the change closes."* That was a real defect: a state change that fully
occurred read as partially closed because of where it sat.

With a perceive on every expectation, closure has a **direct channel** that does not inherit path
decay. Inherited grounding still attenuates with depth, which is correct for a ranking signal;
direct grounding does not, which is correct for closure. The two were conflated because only one
channel existed.

## Links — the next frontier, and it is HARD links

Deferred until the first implementation runs, but worth naming now because the analogy already
supplies the mechanism and it is the one [Toy 3](../2026-07-27/toy-runthroughs.md) asked for.

Toy 3's closure says an action occurrence *"may be referenced by several expectations without
being owned by any of them, because actions own nothing."* That is a **hard link** exactly:
several directory entries, one inode, ownership held by nobody and the inode surviving as long
as something points at it.

```text
HARD LINK   several entries, ONE inode. the switch's flip referenced by Elight and Efan,
            executed once, graded once, inherited along each path. this is Toy 3.

SYMLINK     an entry containing a PATH. resolves late, and can dangle. a different and
            weaker thing — a reference to wherever that name currently leads, rather than
            to this occurrence.
```

The two are worth keeping apart from the start. Toy 3 needs the first. The second is a separate
question about whether an expectation may refer to another *by name* and tolerate the target
changing underneath it — which, given the graph is append-only, may not be a thing this design
ever wants.

## Concurrency

**Deferred.** Single-threaded, one stack, for now.

The eventual shape is settled even though the mechanism is not: **many processes, one
filesystem.** Append-only durable state is what makes that safe, and it is already the rule, so
nothing built now has to be unbuilt. Siblings-run-concurrently from
`2026-07-27/expectations.md` is on hold rather than deleted.

---

## Open

Carried in [questions](questions.md). The two that bear directly on this doc:

```text
WHAT THE CHILD EDGE MEANS   the worked example is a task decomposition; the edge is an
                            observation. how those are the same relation is not yet stated,
                            and trust.md multiplies along it.               → Q1

FILE INTO DIRECTORY         typing is derived from composition, so adding an entry to an
                            action silently promotes it to an expectation. feature or bug
                            is unspecified, and nothing currently stops it.  → Q2
```
