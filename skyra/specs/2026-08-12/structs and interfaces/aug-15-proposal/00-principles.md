# Principles — the test a type must pass

**Date:** 2026-08-15
**Status: DERIVED.** Not new rules. The eleven constraints the design set already enforces, stated
as properties of a *type surface* so that a proposed type can be checked against them mechanically.

> **Every one of these is an INCAPACITY, not a policy. If a rule below is expressible and merely
> forbidden, the type is wrong.**

---

## The eleven

```text
P1   NOTHING MAY APPEAR FROM NOWHERE
     Every value traces to a parameter or a child return. No unbound literals, no ambient graph
     read, no hidden entity memory.
     TEST: can a type be constructed by a caller who was not given its constituents? then no.

P2   MONOTONE NARROWING
     Authority enters at a boundary and can only shrink on the way down. A select may drop; it can
     never mint.
     TEST: is there a constructor for the authority type that does not take a wider one? then no.

P3   PROVENANCE IS SUBSTRATE-KNOWN, NEVER SELF-REPORTED
     A value may not certify its own origin, type, capability, or standing by writing a label into
     content.
     TEST: can content be forged to change how the substrate classifies it? then no.

P4   REMEMBERED ⊆ OCCURRED
     Nothing enters durable state unless the substrate witnessed it occur. Remember selects; it
     never originates.
     TEST: does any durable-write signature accept material the kernel did not itself hold? then no.

P5   TWO FORMS OF FAILURE, NO UNIVERSAL ERROR
     Failure is inability to cross a boundary. Explanation is evidence ABOUT a failure and never
     the thing that causes it.
     TEST: is there an Error value whose absence would make a failure invisible? then no.

P6   THE PARENT OWNS THE BOUNDARY
     Input contract, return contract, parser, and the behavior invoked when the parser refuses.
     A child may not declare what its parent must accept.
     TEST: is any contract reachable from the child's side? then no.

P7   NO SEMANTICS IN THE KERNEL
     The kernel may HOLD anything and READ almost nothing. Holding an opaque blob is fine; reaching
     into it is not. Structural traversal is not interpretation.
     TEST: does the kernel call a method that returns meaning rather than structure? then no.

P8   DOWN, NEVER UP; NEVER SIDEWAYS
     A possibility cannot observe its ancestry, its own reception, or a sibling. This is a fact
     about which way pointers point, not a check anyone runs.
     TEST: is there a field or method from which an ancestor is reachable? then no.

P9   NO TAG READ AT SELECTION
     Credit, age, cost, popularity, and exposure are measures. None may become a selection input,
     because what is used gets selected, gets more support, and suppresses novelty.
     TEST: is the quantity reachable from a frame during Decide? then no.

P10  DERIVED, NEVER STORED
     Two places for one fact will eventually disagree. Cites, heads, distance, hop profiles,
     leaf-ness and used/seen are all read-time computations.
     TEST: is there a setter? then it is stored, and it is probably wrong.

P11  EXACT DESUGARING
     Every language construct has a primitive sequence. Every primitive sequence has a language
     form, or is DISCOVER / ORIGINATE and deliberately has none.
     TEST: can the language express something the primitives cannot, or vice versa? then there are
     two systems wearing one name.
```

```text
SOURCE   P1  continuation-language.md "Nothing may appear from nowhere"
         P2  reach-boundaries-proposal.md "CHILD HORIZON ⊆ PARENT HORIZON"; terminology.md PROVISION;
             energy/README.md "A CHILD MAY SPEND OR SUBDIVIDE IT. IT MAY NOT MINT."
         P3  engram-contracts-and-composition.md "Typing is known observation provenance"
         P4  causality-and-recognition.md; observable-causality.md, both §"two laws of remembrance"
         P5  engram-contracts-and-composition.md §"Failure requires no universal Error object"
         P6  parent-owns-the-return-boundary.md (DECIDED)
         P7  the-crossing.md "Seeing is not reading"; conditions-and-loops.md on traversal
         P8  primitives.md §7; the-collapse.md §"Ancestry, because reach may not go upward"
         P9  second-kind-of-credit.md; age.md §"The constraint"; energy/README.md §"IS ENERGY READ
             AT SELECTION"; gravitate.md §"NO SELECTION SHORTCUT"
         P10 age.md §"A pointer, not a container"; distance.md §"Derived, never stored";
             type-surface-proposals.md C2
         P11 surfaces-and-solidification.md §"Two modes, not two systems"
```

---

## The three that are currently violated in `skyra-dna-v5`

Recorded so the proposal is measured against the same test it sets.

```text
P2   VIOLATED   kernel/kernel.go runGroup folds EngramRef from a child's RETURN into the parent's
                Delegated map. That is minting INTO a frame from below, in the same function whose
                comment says "a select may drop from it but can never mint." The narrowing law and
                the delegation mechanism contradict each other in one file.
                → resolved by 02 §5. This is the finding that produces D3.

P4   VIOLATED   Ingress and runGroup both call G.Store outside a commit. the-collapse.md records
                this and type-surface-proposals.md C6/C7 propose the removal; overnightBuild
                explicitly excludes it as loop work.
                → resolved by 07 §2, where the store is ephemeral until projection.

P7   VIOLATED   shortExpr reads Expectation.Expression, oneLine reads Observation.Content. Two
                lines, both for a log. stops.md E3 confirms they are the whole of it.
                → resolved by 06 §5, the Events sink.
```

---

## The one rule this proposal adds

Not in the set as a rule, though the set follows it everywhere:

```text
P12  A SUM TYPE WHERE THE SET WRITES A LIST OF OUTCOMES
     Three Reach conclusions, four Effect dispositions, six mismatch classes, three recognition
     strengths. Every one is currently prose, and every one has already produced a defect when
     implemented as booleans or an untyped string.
     TEST: does a documented outcome have no name in the type system? then a code path will
     silently occupy it.
```

`Effect`'s two booleans for four states — `type-surface-proposals.md` B1 and `02-defects.md` D2 —
is the worked example. The failed-write state was not merely unhandled; it was **unnameable**, so
nothing could point at it. That is the argument for making every outcome list a closed type.
