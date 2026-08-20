# Implementation notes under the kernel/CTU boundary

**Date:** 2026-08-20
**Status: PROPOSAL.** The mechanism I would put underneath `boundaries.md`. This is one level
below what the conversation was about, and it is separated for exactly that reason — the
boundaries can stand while any of this is replaced. Nothing here is agreed; it is the concrete
machine that would satisfy the two-crossing contract.

> **Identity is content-addressed at the store. The namespace is a mutable label map over an
> immutable content-addressed graph — refs over objects.**

---

## 1 · Identity: provenanced while live, content-addressed at REMEMBER

The question the boundary leaves open is what makes two proposed names "the same type." Two horns,
both bad: a bare name converges but collides on conflicting shapes; a provenanced name never
collides but never converges. Both assume identity is fixed at PROPOSAL.

It is not. Derive the alternative from three things the boundary already fixes:

```text
NEED       convergence (genuinely-same types collapse)  +  no wrong-merge (different shapes never
           fuse)  +  ONE store (REMEMBER, which already writes the cone).
FORCES     identity = a function of STRUCTURE, computed at the single point contributions meet.
           That point is REMEMBER. Identity is the hash of the cone.
```

So a proposed name is **provenanced while it is live** — for the local satisfaction check and the
trace — and **content-addressed at REMEMBER**. The human name is a mutable label pointing at that
hash. This is Unison's identity model, and it is nearly free here because REMEMBER already writes
the cone: hash what you are already writing.

---

## 2 · Structure: refs over objects

```text
NAMESPACE       name → hash        mutable map. kernel-owned. written at REMEMBER.      git refs
DURABLE GRAPH   hash → content     immutable, content-addressed store.                  git objects

THINK       reads a cached name→hash view; proposes provenanced names.
ACTIVATE    name → hash → content.  resolve, then load.                                 checkout
REMEMBER    content → hash; intern by hash; update name→hash.                           commit
```

This is git's architecture, and it lands the two crossings on it exactly: ACTIVATE resolves a ref
to a hash and loads the object; REMEMBER writes an object, computes its hash, and moves the ref.

---

## 3 · The horizon filters the namespace — shared content, scoped resolution

`boundaries.md` SEAM 3 asks how a shared namespace coexists with local, scoped authority. Under
content-addressing they stop competing:

```text
THE MAP        one, shared per namespace.        this is what makes convergence work.
ACCESS         horizon-scoped.                    "THINK may only propose names that resolve"
                                                   = names reachable in name→hash UNDER your
                                                   capability.
```

A global filesystem with per-process permissions. Convergence is a property of the map;
containment is a property of the reach.

---

## 4 · Horizons as attenuable tokens — why BOUND is not a crossing

This is the mechanism `boundaries.md` SEAM 2 needs. Make a horizon a macaroon-shaped capability:

```text
NARROW   add a caveat. LOCAL, offline, unprivileged. no kernel.
WIDEN    remove a caveat / forge a broader grant. IMPOSSIBLE by construction.
ENFORCE  the fabric verifies the token on every MOVE. the MMU check. NOT a trap.
```

If that holds, BOUND resolves into a register-to-register narrowing plus inline fabric
enforcement, and never crosses the kernel/CTU line. This is what has to be true for "exactly two
crossings" to survive, and it is a commitment (tokens, not kernel-held maps), not a free lunch.

---

## 5 · Contract currency: a predicate over hashes

With identity as a hash, a contract is a predicate over hashes, and RECOGNIZE compares hashes in
the frame. The one real fork:

```text
== hash          nominal-tight. unforgeable identity, zero flexibility.
conforms(hash)   structural. flexibility, and it reopens "what does conform mean over hashes."
```

Decidable PER CONTRACT rather than once for the machine — exact identity globally, and each
position's contract chooses whether it wants `==` or `conforms`.

---

## 6 · THINK's cached view

For THINK to be a non-crossing (`boundaries.md`), it must read the namespace without trapping. So
it reads a **cached projection** of name→hash — a TLB over the kernel-owned map, refreshed at
ACTIVATE. THINK hands up names and top-level contracts only, never bodies: it has the directory
entry, not the file. `readdir`, not `open`. This is also why RECOGNIZE must verify on arrival —
the load that could confirm the shape happens at ACTIVATE, after THINK is done.

---

## The seams

Each is a collision between the boundaries in `boundaries.md` and this mechanism — not doubts
about whether content-addressing works, but places where making it satisfy the boundary bites.

```text
SEAM A — HASH IDENTITY IS EXACT, AND CONVERGENCE WANTED APPROXIMATE
  COLLISION   identity = hash of the whole cone (this doc)  ×  convergence is the GOAL
              (boundaries.md, the namespace). A hash changes if any leaf deep in the cone changes,
              so two "morally the same" types with a deep diff NEVER converge. Content-addressing
              gives exact structural identity and nothing for "close enough."
  CONSEQUENCE convergence-by-hash captures only genuine structural sameness. Any approximate
              identity has to be rebuilt as per-contract `conforms` (§5), which means the shared
              global thing (identity) and the flexible thing (conformance) live at different
              layers. That may be correct — but it means the namespace alone does NOT make two
              usefully-similar types interchangeable.

SEAM B — THE COLLISION MOVED TO LABELS
  COLLISION   content never collides (hashes are unique)  ×  the namespace is name → hash, and two
              arenas can bind the same LABEL to different hashes.
  CONSEQUENCE the wrong-merge (malignant, boundaries.md SEAM 3/linker-ODR) is gone; a label
              ambiguity (benign) replaces it. But REMEMBER's write to name→hash now needs a
              policy for one-name-two-hashes — qualify by hash-prefix, last-writer, refuse — and
              the boundary doc's "install a surviving name" does not say which. The malignant
              problem is solved; a benign one is created and left unspecified.

SEAM C — WHICH CONTRACT PREDICATE, AND SEAM 3 DOES NOT GO AWAY
  COLLISION   §5's `==` vs `conforms`  ×  boundaries.md SEAM 3 (global identity meets local
              satisfaction). Choosing `==hash` collapses the two into one (identity IS
              satisfaction) and kills local flexibility; choosing `conforms` keeps them apart and
              inherits SEAM 3 whole.
  CONSEQUENCE the boundary seam is not resolved by content-addressing — it is RELOCATED into the
              contract predicate, where it becomes a per-position choice instead of a machine-wide
              one. Better shaped, not closed.

SEAM D — CACHE STALENESS BECOMES A RESOLUTION FAULT
  COLLISION   THINK reads a CACHED name→hash view so it can be a non-crossing (§6, boundaries.md)
              ×  the authoritative map updates at REMEMBER.
  CONSEQUENCE a name that resolved in THINK's cached view can fail to resolve at ACTIVATE against
              the updated map — a stale-TLB / stale-index fault. The boundary's totality still
              holds (it is the third outcome, a resolution fault, not a stuck state), but the
              machine now has a real invalidation question the two-crossing surface does not
              mention: when and how the cached projection is refreshed, and what a body sees when
              a proposal it was handed no longer resolves.
```
