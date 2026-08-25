# Kernel entry point

**Date:** 2026-08-18
**Status: RESEARCH.** Source material. Recorded as found, not yet mapped onto anything in this
project.

---

## The note

> On x86, there are special privileged CPU mechanisms/registers and memory structures that the
> kernel configures for exactly this purpose.
>
> For exceptions and interrupts, the big one is the **IDT — Interrupt Descriptor Table**.

---

## What the IDT is, factually

```text
WHAT       a table of 256 gate descriptors, indexed by VECTOR NUMBER. each entry names a
           handler and the privilege at which it runs.

WHERE      the IDTR register holds its base and limit. loaded with LIDT — A PRIVILEGED
           INSTRUCTION. user mode cannot point the machine at a different table.

WHEN       on an exception or interrupt the CPU indexes the IDT by vector, switches stack
           and privilege level, and transfers control to the handler. THE PROGRAM IS NOT
           CONSULTED AND CANNOT DECLINE.
```

This is configure-then-exit made concrete: **the kernel writes the table once, and thereafter
the CPU performs the transfer.** The kernel is not running during the dispatch. It is the
destination of a transfer the hardware carries out on its behalf.

## There is more than one entry mechanism, and the difference matters

```text
INVOLUNTARY   exceptions and interrupts. page fault, divide by zero, timer, device IRQ.
              THE PROGRAM DID NOT ASK. dispatched through the IDT.

VOLUNTARY     SYSCALL / SYSRET. the program deliberately requests kernel service.
              DOES NOT USE THE IDT — the target lives in MSRs (IA32_LSTAR and friends),
              a fast path added precisely because the IDT route was too slow for a
              mechanism programs invoke on purpose, constantly.
```

Supporting structure, same category — configured once, then used by hardware without the
kernel running: the **TSS** supplies the kernel stack pointer (RSP0) used when privilege
changes, so the trap lands on a stack the user could not have corrupted.

---

## The mental model

```text
                  CPU
                   │
          CONTROL TRANSFER
             /           \
            /             \
      EVENT ENTRY       REQUEST ENTRY
          │                 │
     "something          "software
      happened"           asks"
          │                 │
          ▼                 ▼
     exception /          syscall
      interrupt
          │                 │
          └───────┬─────────┘
                  ▼
            privileged
             execution
```

Both arrive at the same place. **They do not arrive on the same terms**, and the difference is
the useful part:

```text
                    REQUEST ENTRY              EVENT ENTRY
WHO CHOSE THE       the caller.                NOBODY. it can land mid-instruction.
MOMENT
RETURN ADDRESS      the caller's next           the faulting point — WHICH THE HANDLER IS
                    instruction. known.         NOT OBLIGED TO RESUME.
RESUMPTION          expected. that is what      NOT GUARANTEED. the handler may terminate,
                    a syscall is.               retry, or unwind past the frame entirely.
SHAPE IN SOURCE     A CALL.                     NOT VISIBLE IN THE SOURCE AT ALL.
```

---

## Mapping this design onto it

```text
REQUEST ENTRY — the possibility asked

  GREET    yes, and A SPECIAL CASE: the requester is not running yet, so it is closer to
              exec than to a syscall. the only entry with no prior context to return to.
  THINK       yes. asks what is reachable. reachability is bounded by privilege, so the
              question cannot be answered outside the kernel.
  ORIENT      yes. asks for live topology, which user code cannot see.
  REACH       yes, and THE CLEAREST SYSCALL IN THE SET. deliberate, crosses a protection
              boundary, expects to come back.
  DECIDE      SOMETIMES. a call to an already-reached target is pure user mode and enters
              nothing. one that must cross CONTAINS a request entry. same opcode, two paths
              — which is the solidification split, now with a name.

EVENT ENTRY — nobody asked

  HORIZON VIOLATION      a general protection fault. the privilege check belongs ON the
                         operation, and the trap is involuntary.
  RECOGNIZE MISMATCH     a fault discovered at the return boundary. nothing requested it.
  REACH FAILURE          NOTE THE SPLIT: the request was voluntary, the failure is not.
                         exactly `MOV RAX, [addr]` — an ordinary user instruction whose
                         page fault is event entry.
  COMPILATION FAILURE    the creator pop. nobody asked for it.
```

### The finding this produces

**This design currently routes every one of those through a return value.** `Realize` hands
back `(observation, matched, error)`, which is REQUEST-ENTRY SHAPE — and it is being used for
things that are structurally event entry.

That matters because of one row in the table above: a request entry resumes the caller, and an
event entry **is not obliged to**. Escalation-as-unwinding is event entry — the frame does not
resume. A normal child return is not — the parent continues exactly where it stopped.

So a horizon violation arriving as a `bool` on a return value is a category error of the same
kind as storing the verdict was. It is not a value the caller inspects. It is a transfer the
caller does not control.

### The second finding: there are TWO tests, not one

`THINK` fails the execution test — it changes no state — and is still kernel API. That looked
like an inconsistency. It is not. **The two columns are sorted by two independent questions:**

```text
DOES IT CHANGE STATE?     sorts EXECUTION ISA membership.
DOES IT REQUIRE           sorts KERNEL API membership.
PRIVILEGE?
```

`THINK` and `ORIENT` fail the first and pass the second — privileged reads. `RECOGNIZE` fails
BOTH, and is not request entry either, since nothing calls it. It is the SOURCE of an event
entry rather than a destination, which is consistent with it belonging to neither column.

---

## Open, not answered here

```text
TWO ENTRIES,       this project has one story for crossing into privilege, and hardware has
NOT ONE            two with different costs and different meanings. REACH looks voluntary —
                   a deliberate request. A RECOGNIZE MISMATCH LOOKS INVOLUNTARY — a fault
                   nobody asked for. Whether those are the same mechanism here is unasked.

WHO WRITES         the IDT is written by the kernel and unwritable by user code, and that
THE TABLE          asymmetry is what the whole scheme rests on. THERE IS NO EQUIVALENT
                   ASYMMETRY IN THIS DESIGN YET — the containment hole, showing up again in
                   a second place.

THE HANDLER IS     a vector is an INDEX, not an address the faulting code chose. escalation
NOT CHOSEN BY      here currently pops to a parent, which IS chosen by structure. whether a
THE FAULTING       causal fault should dispatch through a table rather than up a chain is
PARTY              a real fork, and nothing has argued it.
```
