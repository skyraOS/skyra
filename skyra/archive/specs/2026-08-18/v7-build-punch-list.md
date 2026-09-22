# Punch-list from the v7 build

**Date:** 2026-08-22
**Status: ASSESSMENT.** The design to-dos the overnight v7 interface build surfaced. Source:
`src/skyra/skyra-dna-v7/changes/boundary-crossings.md` (the implementer's findings), consolidated
against this set. Not new canon — a work list. Each item names the doc that OWNS the fix.

> **The interfaces that did not hold are the signal. Four are real design gaps; one is a semantics
> call; several were settled well by the build and should fold back in.**

---

## P1 — blocks the core (resolve first)

```text
1  THE POSSIBILITY TREE HAS NO INTERFACE
   WHAT    ACTIVATE carries no parent operand; child-linking, the live tree, and fan-out/fan-in
           are unbuilt. The build ran ONE synchronous hop (request → entity → response); THINK,
           composition, and the reservation station are not exercised.
   OWNS    primitive-details.md (ACTIVATE · ROOT-as-position · the possibility tree)
           + decode-and-run.md (fan-out via APPROACH, fan-in via RECOGNIZE, the reservation station).
   DECIDE  how ACTIVATE names a parent and links a child; where the live tree lives; whether the
           reservation station is a typed thing. This is the recursion that makes it a machine.

2  PROVENANCE MUST SURVIVE THE FABRIC, UNFORGEABLY
   WHAT    an entity yields Content; the local CTU stamps Observation; a FOREIGN parent CTU must
           RECOGNIZE that provenance without MOVE reading the payload and without trusting the
           sender's word. The build carries the stamp beside the MOVE, locally only, and flags it.
   OWNS    traffic-routing.md ("what correlates a response" · custody/wire) + decode-and-run.md
           (provenance) + the-quadrant.md (cross-machine trust).
   DECIDE  where the stamp travels, and what makes it unforgeable across a trust boundary — signing,
           the horizon, or correlation-derived. THE BIGGEST OPEN: it is the distributed-trust
           question in concrete form.
```

## P2 — shape the interfaces asked for

```text
3  THE FRAME NEEDS A CONTRACT-AT-POSITION LOOKUP
   WHAT    RECOGNIZE checks the parent's frame-resident contract at the CORRELATED position, but
           Frame holds one return contract and no correlation→contract map. The build added an
           Expect(correlation, contract) ledger as a stand-in.
   OWNS    continuation-language.md §1–2 (what the frame supplies) + primitive-details.md (RECOGNIZE).
   DECIDE  the real frame shape — how a correlated position maps to its contract. Don't leave it a
           side ledger.

4  GREET NEEDS A NORMALIZED RETURN PATH
   WHAT    a request must carry WHERE its one response goes; GREET(Move) had no source, so the build
           made it GREET(from, …), `from` supplied by ingress.
   OWNS    traffic-routing.md (ingress normalization · addressing) + primitive-details.md (GREET).
   DECIDE  whether the return path is envelope addressing (normalized at ingress) or part of the
           correlation — and record it, because it touches the header/correlation seam below.
```

## P3 — a semantics call to make

```text
5  CONTRACT SATISFACTION: NOMINAL OR STRUCTURAL
   WHAT    the build made "satisfies" = expected name equals producing name (pure nominal — the
           tight end of a fork the specs already name).
   OWNS    boundaries.md SEAM 3 + continuation-language.md §7–8 + implementation.md §5 (== vs conforms).
   DECIDE  is a contract `==` (nominal-tight, what the build did) or `conforms` (structural)? This
           one is decidable PER POSITION, not machine-wide.
```

---

## Resolved by the build — fold back into the specs

Settled well; absorb them so they are not re-litigated.

```text
OUTCOME ON EVERY FALLIBLE CROSSING   totality made explicit per-primitive (Activate/Remember/Define/
                                     Run/Interpret/Schedule), not just MOVE.  → move.md, primitive-details.md
REMEMBER(root, keep)                 root-only becomes expressible (Parent==nil); durability failure
                                     reportable.                              → primitive-details.md
RESOURCE IS A SUM TYPE, not bools    realizes "named, not flagged" with no impossible states. → dispatch.md
TARGET IS TWO-LEVEL (endpoint+name)  realizes the traffic-routing addressing split. → traffic-routing.md, move.md
DEFINE IS A KERNEL METHOD            authority-tier write; ties account+pin.   → horizon-and-root.md (Model A)
```

---

## Still open, and it clusters with P1.2 / P2

```text
HEADER / CORRELATION HOME   the build kept them in the ENVELOPE because routing forced it; move.md
                            still says PAYLOAD. Unresolved. Resolve alongside #2 and #4 — they all
                            touch the envelope.  → move.md vs traffic-routing.md
```
