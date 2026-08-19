# State and storage

**Date:** 2026-08-18
**Status: RESEARCH.** Source material. Recorded as found; the causal mapping below is
questions, not conclusions.

---

```text
STATE / STORAGE
│
├── CPU state
│     └── registers
│           RAX, RBX, RIP, ...
│
├── Memory
│     └── virtual/physical memory
│           address → value
│
└── Persistent storage
      ├── SSD
      ├── hard drive
      ├── flash
      └── etc.
```

---

## The three tiers are reached differently, not just at different speeds

Speed is the boring difference. These are the load-bearing ones:

```text
                REGISTERS              MEMORY                  PERSISTENT STORAGE
HOW REACHED     NAMED. the register     ADDRESSED. the address   NEITHER. no instruction
                is part of the          is a VALUE, and can be   names or addresses it.
                instruction encoding.   computed at runtime.     reached by I/O PROTOCOL.

OPERABLE        YES. `ADD RAX, RBX`     YES. `ADD RAX, [addr]`   NO. NOTHING EXECUTES
DIRECTLY        operates in place.      operates through it.     AGAINST A DISK. it must be
                                                                 brought into memory first.

CAN FAULT       NO.                     YES. translation,        YES, and differently —
                                        permission, absence.     the device may simply fail.

PRIVILEGE       user mode.              user mode, WITHIN what   PRIVILEGED. device access
                                        the MMU permits.         is kernel work.

SURVIVES        not a context switch    not power loss.          YES. that is the whole
                unless saved.                                    point.
```

### The rule that matters most

> **THE CPU CANNOT EXECUTE AN INSTRUCTION AGAINST PERSISTENT STORAGE.**

There is no addressing mode for it, no opcode that reaches it, and no way to phrase the
attempt. Data must be brought into memory before anything can operate on it. That is not a
performance convention — it is a structural impossibility in the instruction set.

### Transitions between tiers each have a name and a cost

```text
REGISTER ↔ MEMORY     LOAD / STORE. unprivileged, ordinary, can fault.
MEMORY ↔ PERSISTENT   I/O. PRIVILEGED, EXPLICIT, AND NEVER IMPLICIT. nothing reaches disk
                      as a side effect of arithmetic.
```

### What is deliberately NOT in this tree

**Caches.** L1/L2/L3 hold state, affect timing enormously, and are invisible to the ISA — no
instruction names a cache line, and no program can observe one architecturally. They are below
the ISA level, in the same category as micro-operations.

That is the test for whether something belongs in this tree at all: **can a program name it?**

---

## Open, for the causal side

```text
IS AN ENGRAM         the match is close. "a possibility cannot be opened without representing
PERSISTENT           an engram" is the same shape as "code must be loaded into memory before
STORAGE?             it executes." if it holds, then NOTHING OPERATES ON AN ENGRAM DIRECTLY,
                     and that becomes inexpressible rather than forbidden.

ARE THERE THREE      registers / memory / disk is three lifetimes with explicit transitions
LIFETIMES HERE?      between them. this design also has three — BRANCH, POSSIBILITY, ENGRAM.
                     whether they line up, or only look like they do, is unchecked.

IS REMEMBER THE      a candidate split that has not been made: REMEMBER moves a value into
STORE, AND PROJECT   live causal history — a STORE, unprivileged, ordinary. PROJECT writes
THE I/O?             the cone at the outer boundary — privileged, explicit, never implicit.
                     those are the two transitions above, and the design may have been
                     treating them as one kind of act.

NAMED, ADDRESSED,    the operand form asks what the allowed forms of a causal operand are.
OR PROTOCOL?         this tree says there are three answers, not two, and that they are
                     distinguished by HOW THE THING IS REACHED rather than what it holds.

WHAT IS THE CACHE?   if the test is "can a program name it," then anything in this design a
                     possibility cannot name is below the ISA and must not appear in any
                     decomposition. WORTH APPLYING TO EVERY FIELD CURRENTLY ON A STRUCT.
```
