# Divergence — the semantic layer against `skyra-dna-v5`

**Date:** 2026-08-13
**Status: MEASUREMENT.** Supersedes `specs/2026-08-11/semantics/divergence.md`, which measured the
same gap two days earlier and smaller.

> **08-11 measured a design that had moved ahead of its code. 08-12/13 moved it further, and in a
> direction the code has no shape for at all.**

---

## What 08-11 already recorded, still true

```text
1  open() RETURNS ON COMMIT       the code kills the possibility; closure.md keeps it live until
                                  the branch commits.
2  Root HOLDS AN ENGRAM           it should be a position, not a node.
3  Builtin IS GLOBAL              not per-principal.
4  NO COPY, NO EDIT               second-kind-of-credit.md specifies both; neither exists.
5  Interpreter IS TWO METHODS     and the crossing is entity-held — see the-crossing.md.
6  NO GRANTS · PROFILES ·         terminology.md's whole authority model is unbuilt.
   MOUNTS · NAMESPACES ·
   PRINCIPALS
```

None of these closed. The overnight build addressed defects, not semantics.

---

## What 08-12/13 added to the gap

```text
7   THE SEVEN PRIMITIVES     approach · think · decide · reach · remember · recognize · orient.
                             The code exports Ingress · Ask · Drive · Accept · Reject · Hold ·
                             Introspect · Spiral. Five map by rename; REACH and RECOGNIZE have
                             no code at all.

8   REACH DOES NOT EXIST     no resolver, no endpoint table, and host.Host cannot be asked what
                             it can reach. Registry maps EntityID → constructor with no host
                             anywhere. Invocation.Host is written once as k.Host.ID() and READ
                             NOWHERE — verified by grep.

9   NO DEVICE IDENTITY       Where decision 2 says identity is a hash over hardware fields read
                             at boot. Nothing reads any of them.

10  ENTITY CRITERIA ARE      the five are a definition, not a type. Nothing checks them, and
    UNCHECKED                criterion 2 could not be checked at first contact anyway — it needs
                             a failure and needs surviving it.

11  DISTANCE IS UNMEASURED   ../energy/distance.md proposes Orient × Endpoint × Experience. Nothing times a
                             reach. Causal distance is the exception: HopProfile() already
                             computes it.

12  APPROACH HAS NO GATE     and authenticate — specified in terminology.md since 08-11 — has no
                             successor in ../continuation-language/primitives.md and no code. See terminology.md's
                             addendum.
```

---

## The one place code is ahead of the semantics

```text
CAUSAL DISTANCE   graph/types.go computes HopProfile() — {hop1,hop2,hop3} validated at range —
                  and ../energy/distance.md only named it on 08-13. The mechanism preceded the word.
```

Worth noting because it is the only row that runs this direction, and because it is evidence the
credit design was right about something before there was vocabulary for it.

---

## One implementation question remains

```text
questions.md §16    The seven-method surface is decided. The implementation still has to determine
                    how much shared machinery Think, Decide, and Reach use beneath that interface.
```

The interface decision no longer depends on that build; only the internal decomposition does.

---

## What this measurement says

The 08-11 reading was *"the code did its job and the design moved."* That still holds, and the
distance is now larger in a specific way worth naming:

```text
08-11's GAP   the code implements an OLDER VERSION of the same shape.
              open() returning on commit is the same operation at the wrong moment.

08-13's GAP   REACH and RECOGNIZE are not older versions of anything. There is no line of code
              that is a worse implementation of reach, because there is nothing to be worse.
```

A gap of the first kind is repaired by editing. A gap of the second kind is repaired by building —
and building is also the only thing that will settle §16, because both branches of that fork read
plausibly on paper.
