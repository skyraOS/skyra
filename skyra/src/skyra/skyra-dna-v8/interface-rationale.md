# Interface rationale — skyra-dna-v8 (the distributed possibility)

**Date:** 2026-08-23
**Status: THROWAWAY.** These interfaces exist to be read in the morning for **where boundaries
cross** in the DISTRIBUTED possibility model. Every term is drawn from `skyra/specs/2026-08-22/`
(on the foundation of `skyra/specs/2026-08-18/`); anything not a doc term is declared and justified
below. This is v7's interface set evolved for one thread only — a possibility partitioned across
CTUs. Bodies are the overnight build's job.

---

## What changed from v7, in one paragraph

A possibility is now the **unit of execution and of transport**. The CTU gains `Execute(*Possibility)`
— the loop that steps a possibility and **forwards a slice** when a dependency is non-local. A MOVE
is the **cut** between partitions: its request leg carries a `Slice` (push) or a pull instruction;
its response leg carries an `Observation` at a join. `Correlation` is **dissolved** into `MoveID`
(the cut edge). Provenance now crosses as a **signed `Segment`**; `REMEMBER` **gathers** the cone
across CTUs; `ORIENT` returns a distributed **`Frontier`**; the scheduler's WHERE becomes a
**push/pull `Selection`**; and a new `fabric` package owns egress, segment fetch, and sign/verify.

---

## New terminology declared

The rule: no new terminology, and any struct that needs a new term must justify it. Carried from v7:
`causal` (package, from "causal state") and `Outcome` (type name for move.md's three outcomes). New
or newly-load-bearing in v8:

```
fabric (package)      DOC TERM, not new — 2026-08-18/README.md: "the CTU executes... the FABRIC
                      transports between them." v7 kept transport inside the build; v8 gives it the
                      package the docs already name.

Custody (interface)   from "custody transfer" (MOVE's definition). Names the SIGN/VERIFY half — what
                      makes custody of a signed segment unforgeable across a cut. Derived, not new
                      concept.

Signature (type)      the mechanism noun for "each CTU SIGNS its segment" (signed-provenance.md).
                      "Signed" is the doc's; Signature is held by the completed Segment. Names nothing new.

SegmentRef (type)     a CONTENT-ADDRESS handle — from content-addressing (implementation.md,
                      remember-gather.md "fetch ... dedup'd by content-address"). "Ref" is the handle.

Selection (type)      DOC TERM — decode-and-run.md's "eligibility-then-SELECTION." Its two variants
                      name the doc's push/pull:
  RunHere               PULL — "request the resource here" (ctu/possibility.md). Descriptive name.
  SendTo                PUSH — "forward the slice to the resource" (ctu/possibility.md). Descriptive name.

MoveID (type)         from terms.md — "the per-edge form ... the v7 build's MoveID"; Correlation
                      dissolves into it (changes/changes.md). Acknowledged by the spec, so used.

PossibilityID (type)  from terms.md's "possibility identity." "ID" is the Go noun.

Execute (CTU method)  from ctu/possibility.md — "the CTU EXECUTES a possibility." A verb the doc uses;
                      it is the LOOP, explicitly NOT a ninth primitive (claude-ideas.md).
```

Doc terms used as-is (cited inline in source): Possibility, Slice, Arena, Segment, Frontier, Move,
Header, Target, Horizon, Caveat, SysInstruction, Payload, Provenance, Contract, Observation, Content,
Potential, Goal, Frame, Name, Engram, Cone, Account, Pin, Resource, EligibleSet, and the primitives
Move/Greet/Recognize/Think/Bound/Orient/Activate/Remember/Define plus Approach/Fetch.

---

## The types the distributed model adds or changes — and why each carries what it does

### The possibility and its slice (`causal`, from ctu/possibility.md, terms.md)

```
Possibility   +ID +Arena +Frontier,   ID = the shared identity across fragments; Arena = the
              +Name +Horizon +Payload DENORMALIZED arenaID back-reference; Name/Horizon/Payload are
                                       the executable work and authority the by-value proposal omitted;
                                       Frontier = this fragment's in-flight cut edges (ORIENT's read).
Slice         Of,Frame,Horizon,       a PARTITION that travels. Horizon is ATTENUATED to the slice —
              Payload.                 least-knowledge. Of = which possibility it belongs to (crosses
                                       on the forward). NO Arena — the home arenaID does NOT travel
                                       (ctu/possibility.md: "a forwarded slice never learns it"); the
                                       receiver gives the resumed child a LOCAL scope at GREET.
```

### The cut edge (`causal`, from move.md, fabric/signed-provenance.md)

```
Move   +ID(MoveID) +Slice +Returned,   ID replaces Correlation (dissolved). Slice = request-leg push
       -Correlation.                    cargo; Payload = pull-leg instruction; Returned = response-leg
                                        observation. THREE cargo modes, one per leg — SEAM: a Cargo
                                        sum type is cleaner (claude-ideas.md).
MoveID        struct{Value}.            an opaque WIRE token. SEAM: logical identity is (possibility
                                        identity, edge) — but PossibilityID is off the wire
                                        (signed-provenance.md), so the token is edge-scoped and mapped
                                        to a possibility LOCALLY at the issuer.
PossibilityID struct{Value}.            LOCAL identity; crosses only inside a Slice, never on a response.
```

### Signed provenance (`causal` + `fabric`, from signed-provenance.md, remember-gather.md)

```
Segment       Provenance,By,Signature,  the unit carried by Observation across a cut and fetched by
              Ref.                       the gather. By = which CTU signed it; Ref = its content-address.
Signature     struct{Bytes}.            makes it unforgeable; filled by Custody.Sign, checked by .Verify.
SegmentRef    struct{Hash}.             content-address; Provenance.Causes is now []SegmentRef (the
                                        lineage the gather walks), not inlined []Provenance as in v7.
Cone          struct{[]Segment}.        REMEMBER's output is now the GATHERED signed segments.
```

### The reads (`causal`, from orient.md, remember-gather.md)

```
Frontier   struct{[]MoveID}.   ORIENT's local read — the in-flight cut edges. whole-view composes by
                               propagation over these, not by a global observer.
Selection  RunHere|SendTo.     SCHEDULE's WHERE, now push/pull. RunHere{Resource} = run local;
                               SendTo{Target} = forward the slice.
```

---

## The interfaces, and the boundary each draws

```
CTU (ctu)        Execute(pointer) + Move/Greet/Recognize/Think/Bound/Orient. Execute is the loop; Greet now
                 RESUMES a slice into a child possibility; Recognize now VERIFIES a signed segment.
Interpreter/     Interpret(Move)→EligibleSet ; Schedule(EligibleSet,operands)→Selection (push/pull).
 Scheduler (ctu)
Kernel (kernel)  Activate(Name,Horizon)→Possibility ; Remember(root,keep)→Cone (GATHER) ; Define(Pin).
Driver (drivers) Activate(Name)→Engram ; Remember(Cone)→Outcome. Cone is now signed segments.
Entity (entity)  Run(operands,Payload)→Content. unchanged — the interior is untouched.
Fabric (fabric)  Approach(from,Move)→Outcome (egress) ; Fetch(SegmentRef)→Segment (gather hop).
Custody (fabric) Sign(Segment)→(completed Segment,Outcome) ; Verify(Segment)→Outcome. trust half.
HorizonMinter    Mint, held only by the kernel. Authority = bearer Attenuate + public VerifyHorizon.
SegmentVerifier public Verify only, held by transport and gather; cannot sign for a CTU.
```

---

## Where boundaries cross — read these first

```
1  SLICE BECOMES POSSIBILITY   ctu.Greet returns a *Possibility resumed from Move.Slice. Its Parent
                               is the incoming MoveID, because the parent possibility does not cross.
                               This is the child-activation the v7 build had no home for — and it must NOT route
                               through kernel.Activate (that is durable load, not migration handoff).
                               If a forwarded slice ever loads via the kernel, the crossing leaked.

2  MOVEID ON THE WIRE          MoveID crosses; PossibilityID and Arena do NOT (signed-provenance.md).
                               The response routes by (Target.Endpoint, MoveID); the owning possibility
                               is resolved LOCALLY. If PossibilityID appears on a response, least-
                               knowledge leaked. THE seam to watch.

3  SIGN / VERIFY AT THE CUT    entity.Run→Content (raw); the CTU stamps Provenance and Custody.Sign
                               makes a Segment; a downstream ctu.Recognize calls Custody.Verify before
                               admitting it. If Recognize trusts an unverified segment, the trust
                               boundary is open.

4  REMEMBER IS A GATHER        kernel.Remember walks Provenance.Causes ([]SegmentRef) via fabric.Fetch
                               to the root, dedup'd by content-address. Root-only stays STRUCTURAL
                               (Parent==nil). Watch that a non-root's Arena back-reference never
                               becomes a commit handle. (remember-gather.md)

5  PUSH vs PULL                Scheduler.Schedule returns Selection: RunHere (pull) or SendTo (push).
                               A SendTo forwards a Slice; a RunHere runs a local Resource. This is the
                               one decision that makes execution distributed. (ctu/possibility.md)

6  ORIENT HAS NO GLOBAL VIEW   ctu.Orient returns only this fragment's Frontier. Any "whole
                               possibility" answer is a composition/cut, never an instant. If Orient
                               ever returns a global set directly, a global observer was smuggled in.
                               (ctu/orient.md)

7  FRONTIER FIDELITY           the Frontier a caller reads is a CUT, not an instant — an in-flight
                               MoveID may be counted in neither end. Approximate for scheduling;
                               ack-disciplined for termination. (ctu/orient.md open)

8  CARGO MODES ON MOVE         Slice / Payload / Returned are mutually exclusive by leg. A malformed
                               Move with two set is the shape a Cargo sum type would make impossible.
```
