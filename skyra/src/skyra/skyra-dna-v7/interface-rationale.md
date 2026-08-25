# Interface rationale — skyra-dna-v7

**Date:** 2026-08-22
**Status: THROWAWAY.** These interfaces exist to be read in the morning for **where boundaries
cross**. Every term is drawn from `skyra/specs/2026-08-18/`; anything not a doc term is declared
and justified below. Bodies are the overnight build's job — this is interfaces only.

---

## New terminology declared (the only two)

The instruction was: no new terminology, and if a struct needs a new term, justify it. Two things
here are not verbatim doc terms:

```
`causal` (package name)   from "causal state" — causal-transition-unit/README.md, "the CTU executes
                          primitives over CAUSAL STATE." It is a LEAF package holding exactly the
                          types that cross boundaries, so ctu/kernel/drivers/entity can share them
                          with no import cycle. Not a new concept — a home for existing ones.

`Outcome` (type name)     from move.md's "three outcomes": SUCCESS / FAILURE / NOT-YET. The three
                          VALUES are doc terms; they needed a Go type to sit under. The name is the
                          only invention, and it names nothing new.
```

**And one thing deliberately NOT invented:** the conversation floated "system block / user block"
for the two registries, but that never landed in a doc. The docs say **host table / entity table**
(`dispatch.md`), so the code uses those. If you want system/user block, it has to enter a doc first.

Everything else — Move, Header, Target, Horizon, Caveat, SysInstruction, Payload, Correlation,
Observation, Content, Provenance, Contract, Potential, Goal, Frame, Possibility, Outstanding, Name,
Engram, Cone, Account, Pin, Resource, EligibleSet, Interpreter, Scheduler, Entity, Driver, Kernel,
CTU — is a doc term, cited inline in the source.

---

## The types, and why each carries the type it does

### The MOVE (`causal`, from move.md)

```
Header         enum Request|Response.   two values → exactly two ports. it is ENVELOPE because the
                                        receiving CTU must read it to route to GREET vs RECOGNIZE.
Target         struct{Endpoint}.        ENVELOPE. the user-instruction address. minimal on purpose —
                                        machine:unit addressing (traffic-routing.md) isn't decomposed.
Horizon        struct{Root,Caveats}.    ENVELOPE. the authority. a Root name + an append-only caveat
                                        list — the narrowing token from horizon-and-root §4.
SysInstruction *pointer, optional.      ENVELOPE, CTU-only. POINTER because it is optional — nil = the
                                        common case (scheduler free). Present = a narrowing constraint.
Payload        struct{Body []byte}.     OPAQUE. a byte body the CTU never reads — user-space content.
Correlation    struct{ID}.              matches a Response to its Request.
Outcome        enum Success|Failure|NotYet.  the return of Move/Greet/Remember. NOT a bool — NotYet is
                                        the third state a bool cannot carry (move.md).
```

**The mirror is visible in the struct:** `Payload` is the entity's private field (opaque to CTU);
`SysInstruction` is the CTU's private field (never reaches the entity). Each space, one private
field — `move.md`'s "two instructions, one per space."

### Execution interior (`causal`, from decode-and-run.md / dispatch.md)

```
EligibleSet   struct{[]Resource}.   what INTERPRET produces — candidates, not a target. already
                                    filtered by the quadrant + any SysInstruction.
Resource      struct{Name,Trusted,Kernel}.  Trusted and Kernel are bools, but they are PROVENANCE,
                                    not runtime flags: Trusted = which table it came from; Kernel =
                                    is it THE named kernel. dispatch.md: "named, not flagged."
Content       struct{Bytes}.        RAW entity output. deliberately a DIFFERENT type from Observation.
Provenance    struct{Resource,Invocation}.  what the CTU stamps onto Content. the entity cannot make one.
```

### Causal state (`causal`, from continuation-language.md / primitive-details.md)

```
Observation   Content + Provenance.  the trust boundary in one type: an entity yields Content; the CTU
                                     produces the Observation by stamping Provenance. THE CROSSING.
Contract      struct{Name}.          parent-owned expectation; RECOGNIZE checks an Observation against it.
Potential     struct{Signature}.     a method-signature string — contract-shaped, no values, no prose.
Goal, Frame   what THINK is given.   Frame = observations in scope + the return contract.
Possibility   struct{Parent,Frame,Outstanding}.  the live process. Parent==nil IS the root/arena — a
                                     POSITION, not a kind (primitive-details.md).
Outstanding   struct{Count}.         custody you do not hold. also ORIENT's read.
```

### Durable side (`causal`, from kernel-ctu-boundary/ / primitive-details.md)

```
Name     struct{Value}.       an entry in the kernel-owned namespace (name → meaning).
Engram   Name+Contracts.      durable behaviour = child contracts + compiled continuation (opaque here).
Cone     struct{[]Engram}.    everything an engram depends on — what REMEMBER writes. "chain is the cone."
Account  struct{Root}.        durable identity a Pin unlocks; owns a namespace.
Pin      struct{Secret}.      authenticates an Account, unlocks its root Horizon.
```

---

## The interfaces, and the boundary each draws

```
CTU (ctu)          the primitives. Move/Greet/Recognize/Think/Bound/Orient, typed. Bound is
                   ATTENUATE only (narrowing); the DEFINE half names the kernel and is not a method.
Interpreter (ctu)  Interpret(Move) → EligibleSet. reads the interface, NEVER the payload.
Scheduler (ctu)    Schedule(EligibleSet, operands) → Resource. where/when. local-vs-distributed OPEN.
Kernel (kernel)    Activate(Name,Horizon)→Possibility ; Remember(Observation)→Cone. the two crossings.
Driver (drivers)   Activate(Name)→Engram ; Remember(Cone)→Outcome. the host-table persistence half.
Entity (entity)    Run(operands,Payload)→Content. returns CONTENT, not Observation — the trust line.
```

---

## Where boundaries cross — read these first

These are the type-flows I'd look at in the morning; several are already-recorded seams.

```
1  CONTENT vs OBSERVATION      entity.Run returns Content; only the CTU makes an Observation (stamps
                               Provenance). If anything hands an Observation straight out of an entity,
                               the trust boundary leaked. (decode-and-run.md)

2  HEADER / CORRELATION HOME    move.md puts Header + Correlation in the PAYLOAD; traffic-routing.md
                               puts them in the ENVELOPE. The code puts them in the envelope (routing
                               needs them) and marks the SEAM on Payload. Genuinely unresolved.

3  TWO ACTIVATEs                kernel.Activate(Name,Horizon)→Possibility and drivers.Activate(Name)→
                               Engram are DIFFERENT crossings at different layers (CTU→kernel vs
                               kernel→driver). Watch that the Engram→Possibility step stays the
                               kernel's, not the driver's. (durable-storage.md)

4  BOUND SPLITS ACROSS TABLES   ctu.Bound only ATTENUATEs (narrowing, local). DEFINE names the
                               kernel, so it is kernel.Define — the authority-tier write that ties
                               Account + Pin. If a Define ever appears on the CTU, that's the
                               crossing to question. (horizon-and-root §2a, boundaries.md SEAM 2)

5  REMEMBER FLUSH RACE          drivers.Remember→Outcome: Success MUST mean durable, or teardown races
                               the flush and a crash loses the arena. (decode-and-run.md open / block-layer read)

6  SCHEDULER SCOPE              Scheduler.Schedule returns one Resource. Whether the eligible set may
                               span HOSTS (distributed) or only local units is the LARGEST open —
                               left undecided on purpose. (decode-and-run.md)

7  SYS-INSTRUCTION REACH        SysInstruction must stay ⊆ Horizon — it narrows selection, never widens
                               authority. If placement ever reaches past the horizon, that's the leak.
```
