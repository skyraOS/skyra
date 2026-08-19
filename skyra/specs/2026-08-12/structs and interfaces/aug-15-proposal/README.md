# Type surface — 2026-08-15 proposal

**Date:** 2026-08-15
**Status: PROPOSAL.** A complete type surface derived from the `2026-08-12/` design set. Every
decision here is traceable to a document in that set; where the set does not converge, this
proposal **picks**, states why, and leaves the fork underneath the section it affects.

> **Read only the dated design set to write this.** `structs and interfaces/` — including
> `aug-14-int-structs/` — was deliberately not read, so that this is an independent derivation
> rather than a revision of one. Where the two agree, that agreement is evidence. Where they
> disagree, the disagreement is worth more than either document alone.

---

## What this proposes

```text
00-principles.md              the eleven rules a type here must satisfy. the test, not the surface.
01-identity-and-authority.md  ids · From · Horizon · Ceiling · Position. the narrowing law, once.
02-observation-and-contract.md Content · Observation · Provenance · Contract · Verdict.
03-engram-and-continuation.md Engram = child contracts + compiled continuation. the language's types.
04-frame.md                   what a possibility sees, and what it structurally cannot.
05-kernel-protocol.md         the surface. FIVE methods for SEVEN meanings.
06-entity-crossing.md         Assemble · Project · Step · Monitor. the opaque types.
07-graph.md                   one mutator. projection of a cone.
08-host-and-reach.md          capability classes, endpoints, and the unbuilt half.
09-open-forks.md              every question in this proposal, collected, ranked.
10-escalation-and-the-       ADDED 2026-08-15. the direction failure travels when repair
   world-boundary.md         exhausts, and why the surface is on the world side.
```

---

## The shape, in one page

```text
                        APPROACH
                  outside → namespace
                  establishes From · Horizon
                            │
                            ▼
                     ┌─ POSSIBILITY ─────────────────────────┐
                     │                                        │
                     │   THINK(entity[], expect[], obs[])     │  potential
                     │        → []Candidate                   │
                     │                                        │
                     │   DECIDE(position, candidate)          │  causal, down
                     │        ├ internal REACH                │
                     │        └ → Realization                 │
                     │                                        │
                     │   ORIENT()                             │  live topology
                     │        → []Head                        │
                     │                                        │
                     │   REMEMBER(observation)                │  causal, up
                     │        ├ internal RECOGNIZE            │
                     │        └ → Verdict                     │
                     └────────────────────────────────────────┘
                            │
                            ▼
                  OUTER REMEMBER projects one cone
                            │
                            ▼
                        GRAPH.Project
```

Five methods. Seven meanings. `Reach` and `Recognize` are protocols with **public result types and
no caller-facing method** — see [`05-kernel-protocol.md`](05-kernel-protocol.md) §1.

---

## The four decisions this proposal makes that the set does not

Stated up front because everything else follows from them, and because each is a place a reasonable
reader could go the other way.

```text
D1   FIVE METHODS, NOT SEVEN.        Reach and Recognize are never called by a possibility.
                                     The set says so in three places and calls them methods anyway.
                                     → 05 §1

D2   HORIZON ABSORBS GRANTS.         Grant · Reach · Energy are three alphabets of ONE
                                     monotonically narrowing object. Ceiling is the durable
                                     maximum. ONE subset check, used at three scales.
                                     → 01 §3

D3   REACH NEVER TRAVELS IN CONTENT. Delegation is provisioned by Decide, not returned by a child.
                                     This dissolves EngramRef, and with it the `Raw`-cannot-be-text
                                     stop in overnightBuild/stops.md.
                                     → 02 §5

D4   THE GRAPH HAS ONE MUTATOR.      Project(cone). Mint · Store · Apply · Disperse · MintRoot all
                                     collapse into it, because closure is the only durable append.
                                     → 07 §2
```

---

## What this proposal does NOT do

```text
NO SYNTAX          continuation-language.md leaves the concrete syntax open and this respects that.
                   03 types the COMPILED form, not the source form.

NO OUTCOMES        syscall-read-example/outcomes/ is Michael-owned. Where a type depends on one of
                   those decisions, it cites it. Where an outcome is empty (WHY, WHEN evidence),
                   nothing is inferred into it.

NO LIFETIME CALL   overnightBuild fenced "when a possibility dies." This proposal types both
                   readings and marks the fork. → 09 F1.

NO IMPLEMENTATION  no storage, transport, serialization, concurrency mechanism, or package layout
                   beyond what a signature forces.
```

---

## Traceability

Every section carries a `SOURCE` block naming the documents it derives from. If a type here has no
source, it is an invention and is marked `INVENTED` with an argument for why the set requires it.
There are four:

```text
Position         01 §4   age.md defines the coordinate; nothing types it.
Verdict          02 §6   language-improvements.md requires classification beside the binary.
Realization      05 §4   the three Reach conclusions are a sum type in prose, nowhere in code.
Cone             07 §2   Remember projects one; the set never names the thing projected.
Focus            10 §3   the bounded repair frame. described four times, typed nowhere.
Exhaustion       10 §4   what makes a parent stop repairing. the set supplies no moment at all.
```
