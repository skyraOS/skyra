# CPU structure — the four layers

**Date:** 2026-08-18
**Status: RESEARCH.** Source material. Recorded as found, not yet mapped onto anything in this
project.

---

```text
                    CPU
┌─────────────────────────────────────────┐
│                                         │
│  1. EXECUTION CONTEXT                   │
│                                         │
│     RIP        ← where am I?            │
│     RAX/RBX... ← working state          │
│     FLAGS      ← condition state        │
│     privilege  ← what am I allowed?     │
│                                         │
├─────────────────────────────────────────┤
│                                         │
│  2. INSTRUCTION / ISA                   │
│                                         │
│     opcode                              │
│     operands                            │
│                                         │
│     ADD RAX, RBX                        │
│     MOV RAX, [address]                  │
│     JMP address                         │
│                                         │
├─────────────────────────────────────────┤
│                                         │
│  3. TRANSITION / EXECUTION              │
│                                         │
│     interpret opcode                    │
│     obtain operands                     │
│     apply defined transition            │
│                                         │
├─────────────────────────────────────────┤
│                                         │
│  4. ADDRESSABLE STATE                   │
│                                         │
│     registers                           │
│     memory                              │
│     devices / control state             │
│                                         │
└─────────────────────────────────────────┘
```

---

## Layer 1 expanded — architectural state

```text
CPU ARCHITECTURAL STATE
│
├── General-purpose registers
│     ├── RAX
│     ├── RBX
│     ├── RCX
│     ├── RDX
│     ├── RSI
│     ├── RDI
│     ├── RBP
│     ├── RSP
│     └── R8–R15
│
├── Instruction pointer
│     └── RIP
│
├── Flags
│     └── RFLAGS
│
├── Control registers
│     ├── CR0
│     ├── CR2
│     ├── CR3
│     └── ...
│
└── other architectural registers/state
```

### It is not uniformly writable, and that is the interesting part

The branches of this tree are not the same KIND of state. They differ by who may write them
and by what means:

```text
GENERAL-PURPOSE   written freely in user mode by ordinary instructions. this is the working
                  set, and it is the only branch that behaves the way "a variable" behaves.

RIP               NOT DIRECTLY WRITABLE. there is no `MOV RIP, x`. it changes ONLY as a
                  consequence of control transfer — JMP, CALL, RET, or an entry. POSITION IS
                  A CONSEQUENCE OF TRANSFER, NEVER AN ASSIGNMENT.

RFLAGS            PARTIALLY writable in user mode. condition bits yes; IF and IOPL are
                  privileged, and attempts to change them are ignored or fault by CPL. ONE
                  REGISTER, TWO PRIVILEGE LEVELS INSIDE IT.

CONTROL REGISTERS ring 0 only. MOV to/from CR faults in user mode. CR3 is the page-table
                  base — WRITING IT IS CONFIGURING THE MMU, which is configure-then-exit in
                  one instruction.

CR2               WRITTEN BY THE HARDWARE, not by software: on a page fault the CPU deposits
                  the faulting linear address here. ARCHITECTURAL STATE THAT EXISTS ONLY TO
                  CARRY INFORMATION ABOUT AN EVENT ENTRY.
```

### What that suggests for the causal side

Recorded as questions. None of these are answered:

```text
IS THERE A RIP?    if position in a continuation is architectural state, the RIP rule says it
                   must not be assignable — only moved by DECIDE, RETURN, or an entry. that
                   would make "jump to an arbitrary point in a body" INEXPRESSIBLE RATHER
                   THAN FORBIDDEN, which is the stronger form.

IS THERE A CR2?    every fault in this design currently produces an observation handed back
                   through a return. CR2 is the other pattern: a defined place the substrate
                   deposits fault information, which the handler reads. THE HANDLER DID NOT
                   RECEIVE IT AS AN ARGUMENT.

IS HORIZON A       CR3-shaped: privileged to write, consulted constantly, never an operand.
CONTROL REGISTER?  that fits "ambient, never an operand" better than any field on a struct,
                   and would make narrowing an INSTRUCTION rather than a parameter.

WHAT IS GENERAL-   the free working set — the only branch a body may write at will. if
PURPOSE HERE?      nothing in this design is that, then the design has no registers, and
                   every value lives in "memory." worth knowing either way.
```

---

## Open, not answered here

The four layers are not the same kind of thing as each other, and the mapping onto the causal
transition unit is **not yet derived**. Recorded so it is not mistaken for settled:

```text
LAYER 1 CONTEXT      privilege sits HERE, alongside position and working state — not in the
                     instruction. Which is the concrete form of "privilege is ambient, never
                     an operand." What plays RIP is unresolved.

LAYER 2 ISA          the vocabulary. maps to the eight primitives.

LAYER 3 TRANSITION   interpret · obtain · apply. THREE STEPS INSIDE ONE INSTRUCTION, and the
                     causal-transition-unit encoding currently collapses them into one
                     position (ACT). Whether that collapse is a simplification or a loss is
                     the open question this diagram raises.

LAYER 4 STATE        what is addressable. THE UNANSWERED ONE — this project has no settled
                     account of what causal state is addressable, by whom, and at what
                     granularity.
```
