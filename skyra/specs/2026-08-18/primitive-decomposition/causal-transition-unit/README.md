# Causal transition unit

**Date:** 2026-08-18, rewritten 2026-08-19
**Status: OPEN.** A frame, a form, and the primitives run through it. The rewrite reflects a
structural change: MOVE became the only transport primitive, and the eight-primitive ISA
listed in the first version did not survive. What replaced each of them is recorded below
rather than deleted.

---

## What it is — TWO UNITS, not one

> **CAUSAL TRANSITION UNIT** — the abstract machine responsible for executing the system's
> irreducible transitions over causal state.

A MACHINE, not a format. It occupies the position a CPU occupies: the thing that actually
performs transitions.

```text
CPU     executes instructions over MEMORY state.
CTU     executes primitives  over CAUSAL state.
```

**But transport is not execution, and conflating them was an error the first version made.**

```text
MOVE IS THE ACT OF WRITING AN INSTRUCTION SOMEWHERE. It is not the act of performing one.
The computation still has to happen, and something has to be the thing that performs it.
```

So there are two units, and the fabric between them:

```text
CTU       EXECUTES. one per endpoint. resolves its own payloads, runs its own body,
          performs RECOGNIZE on arrival, decides what to retain.
FABRIC    TRANSPORTS between CTUs. ONE INSTRUCTION — MOVE. reads the envelope, refuses on
          authority, and never looks inside.
```

This is not an invention. Tagged-token dataflow machines are built exactly this way — a
token queue and matching store that route, and separate processing elements that compute.

**THE KERNEL IS NEITHER.** That is the correction the whole frame exists to make. The kernel
was being asked to execute everything; a kernel that executes everything is not a kernel, it
is an interpreter wearing the name.

---

## Where the frame came from

Hardware, read for its split rather than its speed.

An ISA is a CPU's vocabulary of machine instructions — `MOV`, `ADD`, `CMP`, `JMP`, `CALL`,
`RET`, `SYSCALL`, `HLT`, control-register ops. **They are all members of the ISA.** `MOV`/
`ADD`/`JMP` are not a smaller set the others reduce to. The CPU may break an instruction into
micro-operations internally, but that is BELOW the ISA level and no program can name it.

Alongside that vocabulary sits a privilege split. Applications run in user mode; the kernel
runs in kernel mode. Ordinary computation — arithmetic, comparison, branching, permitted
memory access — **executes on the CPU directly.**

> **THE KERNEL DOES NOT INSPECT EVERY INSTRUCTION.**

It is not interposed. It is TRAPPED INTO. It configures the enforcement machinery once, steps
out, and control transfers to it only when something requires privilege. Memory protection is
the clearest case: the MMU checks every access, in hardware, and the kernel is consulted on
none of them. It only wrote the page tables.

```text
INTERPOSED      every step passes through the kernel, which inspects and forwards.
                the model this project had been assuming.
CONFIGURE       the kernel establishes what is permitted, exits, and is re-entered only
AND EXIT        on privilege, fault, or interrupt. THE MODEL HARDWARE ACTUALLY USES.
```

The second is what solidification already describes — primitive mode crosses per step, a
compiled continuation does not cross on the happy path. That was being treated as an
optimization. Under this frame it is the architecture, and the optimization reading was the
mistake.

---

## The two tests

The first version of this document ran ONE test and applied it to two columns, which produced
question marks that were artefacts rather than ambiguities.

```text
DOES IT CHANGE STATE?        sorts EXECUTION membership.
  > "An execution primitive is an irreducible operation by which the execution substrate
  >  changes its underlying state."

DOES IT REQUIRE PRIVILEGE?   sorts what must be REFUSED BY THE FABRIC rather than performed
                             by an endpoint.

CAN A PROGRAM NAME IT?       sorts what is an instruction at all. Caches fail it. So does
                             REACH — nothing ever called it, which is why it kept refusing to
                             become a method.
```

Two words in the first test carry the weight:

```text
IRREDUCIBLE   nothing below it is nameable. a decomposition that bottoms out in another
              primitive has not finished; one that bottoms out in mechanism has.
CHANGES       an operation that changes no state is not an execution primitive, WHATEVER
STATE         ELSE IT IS.
```

---

## THE EXERCISE — the decomposition form

Every primitive is written out in the same form. This is how an ISA reference specifies an
instruction:

```text
OPCODE

operands:
    role
        allowed forms

reads:
    state/value(s) required to execute

transition:
    irreducible rule applied

writes:
    state changed by the transition
```

### The worked example

`ADD`, decomposed. A causal primitive written out should read like this:

```text
ADD

operands:
    destination
        allowed forms:
            register
            memory

    source
        allowed forms:
            register
            memory
            immediate

reads:
    value(destination)
    value(source)

transition:
    value(destination) + value(source)

writes:
    destination
    flags
```

### Why these four sections

```text
OPERANDS ARE NAMES,      an operand is what the instruction NAMES. a read is a value
READS ARE VALUES         OBTAINED. `ADD RAX, [mem]` has two operands, and dereferencing
                         happened between naming and reading. THE READ IS THE STEP THAT CAN
                         FAULT, which is why it may not be folded into the operand.

TRANSITION IS AN         `value(destination) + value(source)` — not prose. IRREDUCIBILITY IS
EXPRESSION               DEMONSTRATED, NOT CLAIMED.

WRITES ARE NOT A         ADD writes `destination` AND `flags`, and flags are returned to
RETURN VALUE             nobody. "what the caller receives" and "what changed" are different
                         questions.

THE FORM IS CHECKABLE    every symbol in `transition` must appear in `reads`. every target in
                         `writes` must be an operand or named state.
```

### Three sections the causal machine added

The form grew when it met real primitives. Each addition was forced, not chosen:

```text
faults:        ADD cannot fault; MOV can page-fault, and ISA references carry an Exceptions
               section for exactly that. THESE PRIMITIVES FAULT CONSTANTLY.

fires when:    NO CPU FORM HAS THIS. On a von Neumann machine an instruction executes when
               reached. Under dataflow it waits for its operands, which produces a THIRD
               OUTCOME the success/failure code cannot carry:
                   SUCCESS   it crossed.
                   FAILURE   it did not.
                   NOT YET   operands unresolved. NOTHING HAPPENED AND NOTHING WENT WRONG.

carried,       MOVE names `payload` as an operand and never reads it. NO CPU INSTRUCTION DOES
never read     THIS — except block copy, `REP MOVSB`, which moves bytes it never interprets.
               That is transport, and it is what makes the end-to-end principle STRUCTURAL
               rather than a rule: MOVE cannot read the payload because nothing in MOVE is
               defined over it.
```

### No privilege field — derive it

Hardware does not let an instruction declare its own privilege; the state it touches
determines whether it traps. So privilege FALLS OUT of `writes:`, and classification becomes a
CONSEQUENCE of each decomposition rather than an input to it.

### What `allowed forms` predicts

Register · memory · **immediate**. An immediate is a value carried inside the instruction
itself — no address, no provenance, appearing from nowhere. **That is what P1 forbids.** So
causal operands admit NO IMMEDIATES, except at a leaf, where the crossing is the origination
point. `Leaf` is already the field governing that forfeit.

---

## The architecture, as it now stands

```text
                        ONE TRANSPORT PRIMITIVE

                              MOVE
                    custody transfer, both directions
                              │
              ┌───────────────┼───────────────┐
              │               │               │
           target          horizon         payload
              │               │               │
         resolved by     resolved by      produced by
          PERCEIVE          BOUND          the body
              │               │               │
        "what is live"  "what is        OPAQUE. request | response
                         permitted"      + correlation + body.
                                         ENDPOINT BUSINESS.

                       ENVELOPE          PAYLOAD
                    the fabric reads   nothing between the
                    target and horizon endpoints looks inside
```

`PERCEIVE` and `BOUND` are not peers of MOVE. **They are what make a MOVE eligible to fire.**

---

## Where the original eight went

Nothing was dropped silently. Each has a successor:

```text
DECIDE       → the REQUEST header value.
RETURN       → the RESPONSE header value.
REACH        → SPLIT THREE WAYS, which is why it never sat still: BOUND constrains the walk,
               PERCEIVE walks and reports, MOVE walks and crosses. One word for three
               relationships to one axis.
RECOGNIZE    → AT THE RECEIVER, on arrival, against a frame-resident contract. The party that
               bears the consequence of a bad match performs the check — self-enforcing in a
               way boundary-side checking never was.
REMEMBER     → an endpoint act on receipt. Payload business.
THINK        → PERCEIVE (substrate: what is live) + RANK (semantic: which is best, and always
               was entity work). The seam is why THINK never sorted cleanly.
ORIENT       → an endpoint reading its own outstanding set.
APPROACH     → UNPLACED. See below.
```

---

## To decompose

```text
move.md      DONE. the only transport primitive.
perceive.md  resolves `target`. walks and reports without crossing.
bound.md     resolves `horizon`. larger space + authority → Horizon. CR3-shaped, and likely
             request entry: seccomp and capability-dropping only ever NARROW and are still
             syscalls, because the map lives where the process cannot write it.
```

Each must:

```text
1  FILL THE FORM             operands · reads · transition · writes · faults · fires when.
                             every symbol in `transition` traced to a read.
2  NAME THE IRREDUCIBLE ACT  in one sentence, in terms of state, not intent.
3  DERIVE THE PRIVILEGE      from what it WRITES. never declared.
4  SAY WHAT IS BELOW IT      the mechanism it decomposes into, and why that mechanism is not
                             nameable by a program. micro-operations, not opcodes.
5  BREAK ITS OWN             argue the other reading. a classification that survives being
   CLASSIFICATION            argued against is worth something; one never tested is a guess
                             with a box drawn round it.
```

---

## The holes

```text
THE BOOTSTRAP IS         every MOVE needs a live connection from PERCEIVE; PERCEIVE needs a
CIRCULAR                 Horizon from BOUND; BOUND needs an authority that APPROACH
                         establishes. Every machine solves this from OUTSIDE the instruction
                         set — a reset vector, a boot address, a root capability handed in.
                         This design will need the equivalent, and "smallest surface" cannot
                         help here.

ORDERING IS A            RESOLVED. Dataflow fires on dependency alone, so independent MOVEs
DEPENDENCY, NOT A        may fire in any order — which is what makes concurrency native. Both
SCHEDULE                 orders exist: PROGRAM ORDER is a specification in the continuation
                         language, EXECUTION ORDER is dataflow, and the execution pointer is a
                         WATERMARK that retires in order. Writing `A; B` does not order them;
                         B follows A only if B DEPENDS on A. Ordering for effect alone must be
                         written as a dependency, which puts it where it can be seen.

VALIDATION IS LATE       the fabric rejects on target and horizon, never on payload. A
                         malformed payload is discovered at the endpoint, after transport was
                         spent. The standard price of end-to-end.

CONTAINMENT              A CPU may let user code run unaccompanied because the MMU CONTAINS
                         it — physically, whatever the code does. Endpoints here are
                         addresses, possibly python, on machines nobody in this design
                         controls. Carrying `horizon` in the envelope switches the model from
                         AMBIENT AUTHORITY to CAPABILITY PASSING, which is the right choice
                         for a distributed system — ambient authority does not survive
                         leaving one box. It is not yet a mechanism.
```
