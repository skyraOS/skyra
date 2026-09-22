# Projection and Representation

**Date:** 2026-08-15  
**Status: FUTURE RESEARCH.** Not a Skyra proposal or decision. A stub, recorded so two existing
anchors and one existing constraint are not rediscovered later as if new.

---

## The question

A user asks for something and the system produces an interface — buttons, fields, a surface to act
through. What *is* that, in this design?

```text
NOT AT ISSUE   that a possibility can cause it. Decide binds a child; a host class realizes it;
               the effect materializes on acceptance. That path exists.

AT ISSUE       what REPRESENTS it. what the durable record holds, what the user receives, and
               what accumulates when the same need recurs.
```

Deliberately downstream. Nothing here should be settled before the WHERE hole is, because a
projection lands on a host and the design cannot yet name one.

## Two anchors already exist

Both predate the question and neither was written for it.

**The host has an unnamed slot.** `src/skyra/cognitive-agent/skyra-dna-v5/host/host.go`:

```text
Skyra kernel
    ├── execution     spawn.
    ├── storage       read/write.
    ├── display  ─┐
    └── audio    ─┴─ the PROJECTION substrate (NOT NAMED YET)
```

That file also supplies the test for whether something belongs there at all: *"whether anything
INTERPRETS. A driver call reaches MECHANISM with fixed, reviewed semantics; an entity reaches an
INTERPRETER whose behaviour is authored elsewhere."* Rendering a surface is mechanism. Deciding what
the surface should contain is interpretation. The seam runs between them, and it is the same seam
that keeps a filesystem read a driver call and an external API an entity.

**The receipt type is already settled.** `../structs and interfaces/aug-14-int-structs/contracts/01-observation.md`
§4a keeps `content.WorldRef` and gives the reasoning — an address resolves, a `WorldRef` does not,
so it is not an address and P12 does not reach it. Its own doc comment in
`../structs and interfaces/02-content.md` lists **"a rendered page"** as the first example of what one points at.

So a materialized surface already has a place to be referred to from, and the reference already
carries the host it is valid on.

```text
CARRIED FORWARD FROM THAT SECTION — HOLE H14. A program can be closed; the graph cannot forget the
receipt. Append-only, no delete. A projection is exactly the short-lived thing that makes a stale
WorldRef ordinary rather than rare, so whatever is decided here inherits that question rather than
meeting it fresh.
```

## The constraint that is already decided

> **The user should not have to design their own desktop.**

Stated as a preference, but the design has already taken the side that makes it structural:

```text
EXPECTATIONS ARE IMMUTABLE   authored once and re-run.
EVERYTHING ACCUMULATES       around them, never inside them.
ALL LEARNING IS TOPOLOGICAL  selection learns; synthesis does not.
```

Which decides the *kind* of thing a representation may be, before anything about its form:

```text
A SURFACE YOU AUTHOR        synthesis. re-improvised per ask, learning nothing.
A SURFACE THAT ACCUMULATES  selection. a stable expectation with history gathering around it,
                            reached the second time instead of re-derived.
```

The second is the only one consistent with the rest of the system. Note the precedent: `Ask` reuses
a branch when the same ask recurs, and `interface changes/kernel.md` records why — if every ask
minted a fresh root, *"the outermost possibility, the one whose selections matter most, would be the
only thing in the system that can never learn."* A per-ask interface is that defect wearing a
different surface.

## What to investigate

```text
1  IS A PROJECTION A HOST CLASS OR AN ENTITY?   the interpret test above answers it in principle;
                                                whether "render this" is fixed enough in practice
                                                to be a driver call is not obvious.

2  WHAT DOES THE USER RETURN?                   a click is an inbound observation with no producer
                                                — Approach. Whether a bound surface makes the user
                                                a durable ingress channel, and what that does to
                                                "nothing gates ingress," is unexamined.

3  WHAT ACCUMULATES?                            if the surface is an engram, its history is what
                                                makes the second asking better. What a return
                                                looks like for a surface — and whether a surface
                                                can be cited — is the substance of the question.

4  WHERE DOES IT RUN?                           blocked. see H15 and H3's missing third axis.
```

The third is the one that decides whether this is a real subject or an application detail. If
nothing accumulates, a projection is just an effect and needs no representation beyond the receipt
it already has.
