# Age

**Date:** 2026-08-14
**Status: DECIDED — the rule. PROPOSAL — where it surfaces.** Age is defined in
[`../syscall-read-example/outcomes/When-outcomes.md`](../syscall-read-example/outcomes/When-outcomes.md)
§2, which says an engram's age advances "on acceptance of a new version" without saying what makes a
revision a new version. This supplies that predicate, and places the result on Reach's return.

> **A parent owns four things for each child position. A change to any of them is a new version of
> that position. Reach reports which version a possibility is acting through; only Remember makes it
> durable.**

---

## The rule

```text
A PARENT OWNS, FOR EACH CHILD POSITION

    signature      what observations this position receives
    return type    what observations this position hands back
    entity         who interprets it
    horizon        how far outward it may see or act

A CHANGE TO ANY OF THEM IS A VERSION INCREMENT.
```

The continuation is interior. A body rewrite does not increment.

This is derivable rather than enumerated: the four are exactly what a parent declares at a position,
and therefore exactly what another engram's parser can depend on. The continuation is the part no
caller can see.

`../continuation-language/continuation-language.md`: *"THE PARENT owns this engram's input and
return contracts"* — and acceptance test 14, *"the LLM cannot change the supplied method signature,
child signatures, or return contract."* Entity and Horizon join them because
`../continuation-language/primitives.md` §3 lists both among what Decide binds, and Decide is the
parent acting.

## Age is positional

Age resolves to a position, not to a free-floating engram.

```text
VERSION COORDINATE      (owning parent, child position, age)
```

The parent's map key is already the lineage identity, so this needs no new durable type. The same
behavior composed by a different parent under a different contract ages independently — which is
required by `../continuation-language/engram-contracts-and-composition.md`: *"No engram owns the
contracts by which every possible caller in the system must use it."*

## A pointer, not a container

A version carries a pointer to its previous self. History is the chain; walking it is the read.

```text
REJECTED   engramID → [v0, v1, v2 …]     a container. bookkeeping.
TAKEN      v2 → v1 → v0                  the chain IS the history.
```

Same reason `Cites` is derived rather than declared — *"two places for one fact would eventually
disagree"* — and the same status `ByReturn` already carries: derived, rebuildable, deletable.

### The invariant this requires

`../semantics/second-kind-of-credit.md` rejects by name *"a back pointer the walk can follow — if
the walk can descend `From`, credit crosses copies and every edit inherits. Whatever records
provenance must be uncitable."*

The version pointer satisfies that structurally rather than by policy:

```text
THE POINTER LIVES ON THE ENGRAM, NEVER IN CONTENT.
THE CREDIT WALK DESCENDS OBSERVATIONS ONLY.
```

`Disperse` enumerates `o.Cites()` → `Producer` → `ByReturn`. It never reads an engram field, and
`Cites()` cannot scan a non-content field. So no citation can name a predecessor and no edit can
inherit standing.

**This is the line to hold.** Attributing credit across a lineage would restore the global goodness
prior through a new door, and it would look like an optimization when it arrived.

## Where it surfaces: Reach

Reach's caller-facing conclusions gain a version coordinate on success only.

```text
POSSIBILITY               live — and the version this possibility is acting through.
                          SAME AGE   an existing version was resolved.
                          NEW AGE    this decision authored a revision. PROVISIONAL.

BEYOND CURRENT REACH      no version. nothing was reached.
BEYOND YOUR HORIZON       no version. nothing was reached.
```

**You have to reach something in order to age it.** The failure cases carry no version because
nothing became live, which is `When-outcomes.md` §2's *"does not advance: rejected versions,
ephemeral attempts."*

### Why not Decide

Decide is where the parent authors the change, so it is the tempting home. It is wrong for the
reason `../continuation-language/primitives.md` §4 already gives:

> *"Reach happens before the possibility exists. An undeliverable decision is not a target
> possibility which opened and failed; it is potential the system could not make live."*

A version returned by Decide would be a version for something that may never exist. Reach is also
forbidden from altering any of the four — *"it may not substitute another entity, expectation, or
input"* — so Reach reporting the version cannot corrupt it.

### Three moments, and only one is durable

```text
DECIDE      the parent authors the change.              nothing versioned.
REACH       the composition goes live.                  the version is RESOLVED for this possibility.
REMEMBER    the chain survives to the outer boundary.   the version becomes DURABLE.
```

This keeps `When-outcomes.md` §2 intact — *"only remembrance turns that revision into the engram's
next durable version"* — and gives §3 its mechanism: **current is possibility-local, and Reach is
where a possibility acquires it.** A rejected chain leaves the position at its prior age.

## Reading the history

The chain is walked by kernel privilege, and it is durable past rather than live topology, so it
belongs beneath Think rather than to Orient. It surfaces at two bounds:

```text
BENEATH THINK   bounded, ranked version candidates, like any other Think return.
                possibility-facing. part of execution.

REFLECTION      the full chain, unbounded, after the fact.
                not possibility-facing.
```

The split is required by the bound Think already carries and by
`../continuation-language/continuation-language.md`, which puts *arbitrary graph lookup* under what
the language may not express. A possibility must not receive an unbounded dump of its own lineage.
Reflection remains a pass nobody has specified.

## What this settles

```text
WHAT MAKES A REVISION A VERSION      a change to signature, return type, entity, or horizon.
WHAT AGE IS ATTACHED TO              a position, not a free engram.
HOW HISTORY IS STORED                a pointer chain, not a container.
WHERE A POSSIBILITY LEARNS ITS       Reach, on success only.
VERSION
```

It also recovers a signal `../semantics/second-kind-of-credit.md` records as destroyed. That
document's sharpest tension is that a superseded engram never closes, never mints, and evaporates —
taking the **authored negative** with it:

```text
TRIED     it closed.
WORKED    it was cited.
REMOVED   a parent looked at it and took it out.   AUTHORED, NOT AN ABSENCE.
```

Retaining versions preserves `REMOVED`. This is not an added cost: `When-outcomes.md` §3 already
requires it, since *"an engram can have multiple remembered versions"* is impossible if supersession
evaporates the predecessor.

## Open

```text
THE NAMESPACE                 age is positional and a namespace sits in no parent's map, so the
                              rule has nothing to attach to at the outer edge. Whether a namespace
                              versions on accepted returns — and whether that is age or a distinct
                              epoch — is unresolved. Note the coupling: `When-outcomes.md` §4 makes
                              GC eligibility "disconnected from all namespaces," so if connection
                              is to a VERSION rather than to the namespace, each increment orphans
                              everything attached to the previous one.

CONCURRENT REVISION           two live chains revise one position and both are remembered. §3
                              permits sibling versions, but then age is not totally ordered and
                              "the accepted version" is ambiguous. §1 does not help: two
                              independently remembered chains have no causal relation.

DEEP COPY                     `second-kind-of-credit.md` leaves Producer across a copied subtree
                              undecided. Versioning is refinement, which shares children rather
                              than copying them, so it does not hit this — recorded so the two are
                              not conflated later.

WHAT READS AGE                resolution, agreed. Nothing else is specified, and the constraint
                              below is what keeps that from drifting.
```

## The constraint

> **Age is a resolution coordinate, never a selection input.**

High age may mean refined. It may equally mean this position keeps breaking. Reading it as quality
at selection time is the global goodness prior `credit.md` rejects by name, and it arrives wearing a
version number.
