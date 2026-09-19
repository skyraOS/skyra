# CPU privilege — mechanism vs policy (a mining seam)

**Date:** 2026-08-27
**Status: RESEARCH / RAW ORE.** Not a decision doc — a captured explanation of how real CPUs split
privilege, dropped here to MINE for Skyra. The "TO MINE" callouts are the veins; the prose is the rock.
Source: the mechanism/policy split (x86 rings, CPL, syscall gates) worked through in conversation.

---

## The one-line answer

**The CPU owns the MECHANISM (enforcement); the kernel owns the POLICY (who gets what).** Neither "is in
charge of privilege" alone — they own different halves. Confusing them ("is it the CPU or the kernel?") is
the same confusion Skyra resolves by splitting `Verify` (mechanism) from `Mint`/`Bound` (policy).

```text
CPU     = the enforcer.  owns CPL bits, fault-on-violation, the numeric ≤ compare.   MECHANISM.  no opinion.
KERNEL  = the policy.    owns the tables + gates the CPU checks against.              POLICY.     all opinion.
```

## What the CPU actually does (mechanism)

- Holds a **Current Privilege Level (CPL)** — 2 bits in the code-segment register, value 0–3. That IS the
  ring. Nothing more than a small integer in a register.
- **Checks** it mechanically: a privileged instruction (load page-table base / `mov cr3`, `hlt`, write a
  control register) FAULTS (#GP) if CPL ≠ 0; a supervisor-only page FAULTS if touched at CPL 3; I/O is
  gated by IOPL. All bit-compares.
- **Traps** on violation — no appeal, no interpretation. The CPU never asks *why*; it sees "CPL 3, page is
  supervisor → FAULT."
- **Compares numerically** — access needs `CPL ≤ required`. This is the "linear decrease": not an actor
  decrementing anything, just **lower number = more privilege** and a `≤` on two integers.

The CPU is a lock + key-reader, **not** a security office.

> **TO MINE #1 — the enforcer has NO opinion.** This is exactly the CTU / `Verify` (MMU). The CTU should
> never *interpret* authority — it does the bit-compare (target ∈ horizon?) and TRAPs. Any "why" living in
> the CTU is a leak. Confirms: outcomes are binary, semantics live elsewhere (think-and-act.md).

## What the kernel actually does (policy)

The kernel is **just software that runs at ring 0.** Its power comes from two facts, not from a per-moment
grant:

- It **writes the tables the CPU checks** — page tables (supervisor vs user), descriptor tables (segment
  privilege), trap gates (the ONLY doors from ring 3 → ring 0). The CPU enforces whatever the tables say;
  the kernel decides what they say.
- It **defines the gates.** User code can't flip its own CPL to 0 (that instruction faults). It executes
  `syscall`, which the CPU turns into a controlled jump **to a kernel-registered address**, raising CPL to
  0 *only at that entry point.* You can become privileged — but only by landing exactly where the kernel
  said you may.

> **TO MINE #2 — privilege is not a ladder you descend; it's a control transfer.** The ring moves in ONE
> hardware step: 3→0 on trap, 0→3 on return. The kernel never lowers itself gradually — it takes ring 0,
> keeps it, and only DROPS to ring 3 to run untrusted code, letting it back up ONLY through gates. Maps to
> our TRAP (mechanical fault → kernel) and the syscall gate = the semantic interpreter / ABI. The "gate is
> the only door up" is our system-and-semantic invariant (the syscall gate = the trap into the kernel).

> **TO MINE #3 — the syscall gate is a KERNEL-CHOSEN landing point, not an arbitrary jump.** Ring 3 can
> only enter ring 0 at addresses the kernel registered. This is the structured-RETURN / `ReturnTo`/`Route`
> idea generalized to the CALL direction: crossings land at KNOWN, pre-registered points, not anywhere.
> Both directions are constrained by pre-registration — worth carrying into the four-interpreter design.

## The bootstrap subtlety (why is the kernel trusted?)

Not because the CPU decided so — because it is **primordial.** The CPU powers on at the highest privilege;
firmware loads the kernel THERE; the kernel is simply the first thing sitting at ring 0, and it defines
everything below by writing the tables. **Trust isn't issued to the kernel — the kernel IS the root**
because it got there first and controls the control room. The CPU faithfully enforces whatever the root
wrote.

> **TO MINE #4 — trust is ROOTED, not granted; the root is whoever boots first at the top.** Directly our
> `boot.go` / `Boot(seed)` and the root entity. The kernel's authority is primordial, established at boot,
> and everything downstream is derived by minting from that root. The chain-of-trust starts at a
> self-asserted root — mirror this in how the root horizon / root capability is established (nobody grants
> the root; it IS the grant-source). Ties to horizon-and-attention (kernel CREATES the horizon).

## The x86 fusion — what Skyra un-welds

x86 **fuses three things** into "ring 0": trusted **and** privileged **and** it's-the-kernel, all welded
together. And the "capability" is just **ambient page-table state** — there's no clean privilege OBJECT
handed around; authority is diffuse global table state.

Skyra **splits what x86 fused** (divergence #4, system-and-semantic.md):

```text
x86                                   Skyra
────────────────────────────────     ──────────────────────────────────────────────────────
ring 0 = trusted = privileged = kernel   TRUST axis (Host/Entity block)  ⟂  AUTHORITY (horizon/capability)
                                         two orthogonal axes, not one welded number
authority = ambient page-table state     authority = an explicit CAPABILITY object (Mint → Horizon)
                                         a thing you hold and hand down, not global table state
kernel sets tables / CPU checks them     kernel MINTS (policy) / CTU Verify checks (mechanism)
                                         the SAME mechanism/policy split, made into two NAMED components
```

> **TO MINE #5 — Skyra's model is x86's implicit split made EXPLICIT and object-based.** "Kernel issues,
> CPU enforces" is *more literally true in Skyra than in x86*, because we pulled mechanism (`Verify`/CTU)
> and policy (`Mint`/kernel) into two named things and made the capability a first-class object instead of
> ambient state. This is a STRENGTH to lean on: capabilities-as-objects (vs ambient authority) is the
> capability-security tradition (seL4, KeyKOS) — a known-good lineage to cite/steal from.

## Loose ore (facts worth keeping, not yet placed)

```text
· CPL lives in the CS register's low 2 bits — privilege rides WITH the running code selector, not a
  separate global. (mirror: authority rides with the Context/Move, via horizon on the frame?)
· #GP (general protection fault) is the catch-all "you violated a protection rule" trap → our TRAP.
· the IDT (interrupt descriptor table) maps trap/interrupt vector → handler address. our failure-KIND
  vectoring (think-and-act.md: "#PF vs #DE", vector points at ACT) is literally an IDT for semantic faults.
· real OSes use only rings 0 and 3; rings 1,2 sit empty — the linear ladder didn't carve real joints
  (see the rings-vs-grid note in system-and-semantic.md).
· privileged transitions are ONE-WAY-gated and atomic (you can't be "half in" ring 0). custody analog:
  a crossing is ACCEPTED or not (ack-and-response.md two-hop) — no half-states.
· user→kernel copies are CHECKED (copy_from_user validates pointers) — the BELIEVED/CHECKED asymmetry
  (Host believed / Entity checked) is exactly this: the kernel never trusts a userspace pointer.
```

> **TO MINE #6 — the IDT is our fault-vector table.** think-and-act's "the CTU vectors to the handler by
> failure KIND" wants a concrete structure: an IDT-shaped table {fault-kind → handler}. Materialize/Remember
> are two entries; the ACT-wake (contract/guard failure) is a third vector. Draw this table.

### #6a — THE KERNEL IS A DIFFERENT LOOKUP, NOT A USERSPACE ENTITY (develops #6)

**Piggybacks directly on #6.** Prior Skyra framing put the kernel as **an entity on the userspace/Entity
block** — reachable through the NORMAL entity path (resolve a target, check the horizon, dispatch
Local/Cut). The CPU says that is wrong: reaching the kernel is a **TRAP through a GATE**, and it is a
**different lookup structure**, because it is a **different kind of work**. And the table it looks up
through IS #6's IDT.

```text
ENTITY dispatch (a peer move)          KERNEL trap (a syscall / fault)
──────────────────────────────        ──────────────────────────────────────────
target IDENTITY → horizon? → driver    fault KIND (a small vector) → handler   ← the IDT (#6)
SPARSE keyspace (many targets)         DENSE keyspace (few vectors: Materialize, Remember, …)
AUTHORIZED (must be in the horizon)    ALWAYS AVAILABLE (the gate is structural, not a capability)
SEMANTIC (a meaning-bearing peer)      MECHANICAL (a meaning-blind service)
a BRANCH (BTB — goes anywhere)         a TRAP (IDT vector — lands at a registered handler)
```

Two consequences that RETIRE "kernel as a userspace entity":

```text
1  THE KERNEL IS NOT IN THE HORIZON — IT IS THE FLOOR UNDER IT. you hold no capability to "call the
   kernel"; the kernel is what CHECKS your capabilities. reaching it is STRUCTURAL (the gate is always
   there), not AUTHORIZED. a horizon entity must be addressable + provisioned; the kernel is neither.

2  IT FIXES A BLOCK INCONSISTENCY. the kernel is BELIEVED (Host / sysblock, ring 0). placing it as an
   entity on the userspace / CHECKED block put the trusted authority on the untrusted, checked side —
   backwards. the trap/gate path is what lets us say the kernel was never a userspace peer.
```

The one nuance to keep honest: microkernels (seL4, #5's lineage) DO give kernel services a message-passing,
**entity-shaped semantic FACE** — but even there the **PATH** to cross in is a trap/gate, never a normal
branch. **Entity-shaped interface, kernel-path mechanism** — not a conflict once face and path are
separated.

> **LANDS:** dispatch-and-decode (TRAP is a distinct arm reached by KIND, not a target resolve — confirm
> the kernel is never a Local/Cut target) · the blocks model (kernel belongs on Host/BELIEVED, not
> Entity/CHECKED) · the four-interpreter grid (a kernel crossing is SystemCall via the gate, not a normal
> address resolve). OPEN: do voluntary syscalls (app requests a service) and involuntary faults (body
> missing) share ONE vector table, or is the syscall ABI a separate gate from the fault IDT?

> **TO MINE #7 — copy_from_user = BELIEVED/CHECKED, concretely.** The kernel VALIDATES every pointer
> crossing from user space (never dereferences a raw userspace address). Our Entity-block-is-CHECKED gains
> a concrete discipline: every observation/payload crossing entity→system is validated at the gate, never
> trusted. A real implementation rule, not just a stance.

## Where each vein probably lands

```text
#1 no-opinion enforcer      → CTU / Verify (ctu.go)            · reinforces think-and-act "CTU owns outcomes"
#2 transfer not ladder      → dispatch-and-decode (TRAP arm)   · syscall gate = the semantic interpreter
#3 kernel-chosen landing    → system-and-semantic (four interp)· pre-registered crossing points, both dirs
#4 trust is rooted at boot  → boot.go / root horizon           · horizon-and-attention (kernel creates it)
#5 capability-as-object     → kernel Mint/Bound + Horizon      · cite seL4/KeyKOS capability lineage
#6 IDT = fault-vector table → think-and-act (the vector table) · Materialize/Remember/ACT-wake as entries
  #6a kernel ≠ userspace entity → dispatch-and-decode + blocks · kernel reached by KIND-vector, not target
#7 copy_from_user           → the entity→system gate           · BELIEVED/CHECKED made a concrete rule
```
