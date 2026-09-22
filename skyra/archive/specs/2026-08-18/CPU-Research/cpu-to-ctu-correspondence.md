# CPU → CTU correspondence

**Date:** 2026-08-19
**Status: WORKING.** The correspondence is the author's. Everything under "Open items" is
derived from it and is QUESTIONS, not conclusions — including one re-classification it forces
on the causal-transition-unit README.

---

## The correspondence

```text
CPU                         CTU

GENERAL REGISTERS           OBSERVATION REGISTERS
RAX, RBX, ...               O0, O1, ...
│                           │
└─ working values           └─ working observations /
                               observation references

RIP                         EXECUTION POSITION
│                           │
└─ current instruction      └─ current instruction /
                               position in realized graph

RFLAGS                      CAUSAL CONDITIONS
│                           │
└─ conditions produced      └─ conditions produced
   by execution                by causal transitions

INTERRUPT / EXCEPTION       EVENT / FAULT ENTRY
│                           │
└─ exceptional event        └─ timeout, resolution fault,
   transfers control           authority fault, mismatch,
                               external failure, etc.

PRIVILEGE / CONTROL         AUTHORITY CONTEXT
│                           │
└─ what execution may       └─ which children/resources/
   access                      transitions are reachable
```

---

## The bucketing

> **CTU REGISTERS hold the execution state of the possibility that is CURRENTLY RUNNING.
> OBSERVATIONS live in memory, inside the possibility. RIP operates sequentially WITHIN a
> possibility.**

The load-bearing part is that **a CPU has one register file, not one per process.** The
register file is whoever is running right now; everything that merely EXISTS is in memory. An
earlier version of these notes gave each possibility its own registers, which is not how the
hardware works and made the tiers impossible to sort.

### State lives in two places, at different times

```text
RUNNING      execution state is in the CTU registers.
SUSPENDED    it has been spilled back into the possibility.
             THE POSSIBILITY IS THE SAVE AREA.
```

A possibility waiting on a child cannot be in the register file — the child is using it. This
is a context switch, and it explains something that has been odd about the `Possibility`
struct: it reads as a grab-bag of unrelated fields. **So does `task_struct`.** It reads that
way because it IS a saved context, and a saved context is heterogeneous by definition —
registers, flags, authority, and position frozen together.

DECIDE therefore carries a cost the current model does not show — save, transfer, restore —
which belongs in its `writes:` section.

### One register file means one thing executes at a time

This bucketing makes the CTU **a single core**, and that is worth deciding rather than
inheriting:

```text
ONE CTU     sequential by construction. a blocking DECIDE is CORRECT, not a compromise, and
            concurrency is simply not expressible. HONEST.
MANY CTUs   concurrency exists, and every multicore question arrives at once: shared memory,
            coherence, which CTU owns which possibility, ordering between them.
```

The decomposition README records a hazard — that a CPU core blocks, so this frame would make
a blocking DECIDE feel validated rather than examined. **This bucketing reverses that.** It
converts the concurrency hole from a vague gap into a countable architectural decision, and
one that probably attaches to `RuntimeID`, since how many CTUs exist is a fact about a machine.

---

## Open items

### 1. CAUSAL CONDITIONS gives the verdict a home — and moves RECOGNIZE

`CMP` writes no general register and computes nothing anyone keeps. It sets flags; you branch;
the comparison is gone. **The result of a comparison is never stored.** That is `CMP; JE`, and
it is what "the verdict is a verb, not a noun" has been describing without a mechanism.

The consequence, stated plainly:

> `CMP` DOES change state — it writes RFLAGS. If causal conditions are architectural state,
> then RECOGNIZE writes architectural state and PASSES the execution test.

**This contradicts the README**, which files RECOGNIZE under SUBSTRATE on the grounds that it
changes no state. That placement was downstream of there being no flags register. Adding one
relocates it into the ISA, alongside a real instruction class: operations whose only write is
to condition state.

The apparent tension — an unforgeable comparison performed by an unprivileged instruction —
dissolves:

```text
THE OPERANDS   are protected. the durable strings cannot be written by either party.  MMU.
THE COMPARE    is unprivileged. `CMP RAX, [protected]` is ordinary user mode.         ISA.
```

Unforgeability comes from WHERE THE OPERANDS LIVE, never from the instruction being
privileged. Both earlier arguments were right about different halves.

**STILL OPEN:** flags or fault? Hardware does both — some conditions set a flag you branch on,
some raise a fault that unwinds. The dividing line is WHETHER EXECUTION CAN REASONABLY
CONTINUE. This design's repair story says a parent can respond to a mismatch in place, which
argues FLAG. Note that "mismatch" is currently also listed under EVENT / FAULT ENTRY above.
It cannot be both.

### 2. What is in a register, and what is in the frame

Observations live in the possibility's memory. What a register holds is therefore **a
reference into that memory**, not the observation itself — which settles the hedge in the
correspondence above ("working observations / observation references") in favour of
references.

The consequence is not cosmetic:

```text
IF A REGISTER HELD THE VALUE       reading it could not fault.
IF IT HOLDS A REFERENCE            dereferencing CAN fault — unreachable, revoked, gone.
                                   THE `reads:` SECTION OF EVERY DECOMPOSITION MUST ACCOUNT
                                   FOR THIS.
```

Registers still behave as a calling convention at a transfer — `RDI, RSI, RDX, RCX, R8, R9`
is where a call's arguments go on System V AMD64, and O0..On is that. But they are not the
storage; the frame is.

```text
STILL OPEN — FINITENESS   architectural finiteness is what makes a register a register; it
                          is why register allocation is a discipline. Six args then the
                          stack, on x86-64. Does a causal call have bounded arity, and where
                          does overflow spill? THE BUCKETING DOES NOT DECIDE THIS.

STILL OPEN — ADDRESSING   registers are NAMED; memory is reached by a COMPUTED address. If
                          observations sit at offsets in a frame, something computes those
                          offsets — an addressing mode. NONE OF THE EIGHT PRIMITIVES HAS ONE.
```

### 3. EXECUTION POSITION is carrying two registers

RIP operates sequentially within a possibility. It says which instruction. It does **not** say
your call depth — `RSP`/`RBP` do, and crossing between possibilities is not a RIP increment
but a full context switch.

```text
INSTRUCTION POSITION   where in this body. RIP. SEQUENTIAL WITHIN ONE POSSIBILITY, and NOT
                       ASSIGNABLE — moved only by transfer.
FRAME IDENTITY         which possibility, and what is beneath it. RSP/RBP.
```

This design has the second (`PossibilityID` + `Parent`) and **is missing the first entirely.**
Which is the same gap RETURN exposed: with no position inside a continuation there is no early
return, no conditional return, and no resumption point to come back to.

Under the bucketing this is sharper than a missing field. **A saved context with no saved
instruction position cannot be resumed** — so either possibilities never resume after a child
returns, which contradicts the whole blocking model, or the position exists and has not been
written down.

### 4. AUTHORITY CONTEXT is page tables, not the privilege ring

Hardware has both and they are different: `CPL` is a LEVEL (0–3); page tables are a MAP of what
is reachable. "Which children/resources/transitions are reachable" is unambiguously the map.

That matters beyond classification. **Page tables are installed by the kernel, unwritable by
the process, and checked by hardware on every access without the kernel running.** If Horizon
is CR3-shaped, that is the first CONSTRUCTIVE answer to the containment hole rather than
another sighting of it — enforcement performed by the substrate on every access, configured
once at Greet and never consulted again.

It also predicts a shape: narrowing becomes an INSTRUCTION that installs a new map for a
child, not a field passed on a call.

### 5. The memory tier — the possibility IS memory

`state-and-storage.md` has three tiers: registers · memory · persistent. The correspondence
had registers and control state and then stopped. **The bucketing fills the gap: memory is
the possibility.**

```text
REGISTERS     execution state of whatever is running now.
MEMORY        THE POSSIBILITY. addressable, faultable, volatile, not durable.
PERSISTENT    the engram. nothing operates on it directly; it must be opened first.
```

Which gives the three homeless fields somewhere to live:

```text
Observations                 locals and arguments, at offsets.
Syntax                       CODE — loaded from persistent storage, because nothing executes
                             from disk. this is why a body must be brought into a possibility
                             before it runs.
ChildBranches · Actions      accumulated record. addressable, not durable.
```

And it makes a possibility **a stack frame with an address space** — locals at offsets, code
loaded in, a return path, a save area for the context. `os-parallels` §19 reached the same
place from an entirely different direction, which is the useful kind of agreement.

## What this does not address

Nothing here touches the containment gap except item 4, which proposes a SHAPE for it and no
mechanism.

**The concurrency hazard recorded in the decomposition README is no longer live in the form it
was written.** That note warned this frame would make a blocking DECIDE feel validated rather
than examined, because a CPU core blocks too. The bucketing removes the camouflage: one
register file means one thing runs, so blocking is a CONSEQUENCE OF THERE BEING ONE CTU rather
than a property of the analogy. The question survives, but as a countable decision — how many
CTUs — instead of a hidden assumption.

What remains genuinely unaddressed: FINITENESS of the register file, the ADDRESSING MODE that
reaches an observation at an offset, whether a mismatch sets a flag or raises a fault, and
where INSTRUCTION POSITION is written down.
