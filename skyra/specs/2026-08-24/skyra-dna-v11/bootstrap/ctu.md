# CTU — bootstrap

**Date:** 2026-08-24
**Status: SCOPING — in progress.** What the CTU is at boot, and how it comes up under the kernel.
Builds on `kernel.md`, `../ownership.md`, `../research/capability-table-and-horizon.md`, and the v7
drivers example (`src/skyra/skyra-dna-v7/drivers`).

---

## Role

The **execution core.** Nearly-stateless; runs whatever possibility is resident. **One CTU per
endpoint**, many possible.

```text
OWNS       execution — Drive, dispatch, the moves.
CARRIES    OS-specific DRIVER mechanisms (its "capabilities") — present but INERT.
HOLDS      only what is HANDED to it — the HORIZON in each move, and cut-IDs (fds).
NO AMBIENT AUTHORITY: it never knows its own capabilities; it acts only on the horizon handed to it.
```

## Capability (driver) vs horizon — two different things

```text
CAPABILITY (driver)   a MECHANISM the CTU carries (OS-specific). PRESENT but INERT. the CTU HAS it,
                      does NOT know it, cannot self-fire it.
HORIZON               the AUTHORITY handed WITH a move. references registered capabilities. the CTU
                      knows ONLY this.
```

So **even the CTU's own drivers require a horizon to touch** — no ambient authority, ever. Anchor:
**CHERI** (capability hardware) — every access must present a capability; there is no ambient authority
even in silicon. The CTU is CHERI-shaped: mechanisms in hand, nothing fires without the horizon —
stricter than an ordinary CPU, which fires any unit for any instruction.

## The CTU is the OS-abstraction — the driver seam

The CTU abstracts the machine through the **driver interface** (`Activate`/`Remember` for storage,
`Approach`/ingress for traffic — see the v7 drivers). The interface is **OS-independent**; the backend
is **OS-specific**.

```text
HOSTED (first)     an abstraction layer over the host OS — v7's macOS backend. the realistic start.
BAREMETAL-capable  a baremetal backend behind the SAME interface. the CTU CAN go to the metal; need not.
"SWAP THE BACKEND, KEEP THE KERNEL" (v7) — this line IS the OS-independence split.
```

## The three layers

```text
1. CTU DRIVERS   inert mechanisms the CTU carries. OS-specific. cannot self-fire.
2. KERNEL TABLE  those drivers registered as capabilities (sysblock). the kernel's inventory.
3. HORIZON       per-move authority handed to the CTU. references table entries.
```

The CTU **carries (1), acts on (3); the kernel owns (2).** The horizon works **identically** whether
the target is a driver, a host, or an entity — the CTU's own drivers are not special; they are
capabilities that happen to live on that CTU, reached the same way as anything else.

## Boot order and the handshake — the CTU is SECOND

The CTU cannot provision itself (self-provision = self-authorization, forbidden), so it comes up under
the kernel:

```text
1. the OS-specific CTU binary starts, carrying its drivers → abstracts the machine. NO authority yet.
2. HANDSHAKE — the CTU ADVERTISES its drivers; the kernel READS them (device-enumeration / PCI-like)
   and REGISTERS them into the sysblock. the CTU never uses that inventory itself. (this handshake IS
   the kernel running ACQUIRE on the CTU — discover/test/admit; see `sequence.md` C.)
3. the kernel provisions horizons; a move arrives WITH its horizon; the CTU Drives within it.
4. a reach beyond the horizon FAULTS (page fault) → the kernel provisions or denies.
```

## The one inherent thing — the bootstrap primitive

**CORRECTION (supersedes an earlier note):** the CTU has **NO inherent horizon** — even its own drivers
are horizon-gated. So the bootstrap loop-breaker is NOT an inherent local horizon.

```text
INHERENT       only the HANDSHAKE CHANNEL to the kernel — the trap door. like a CPU's trap instruction
               being inherent (you can always trap; everything else needs a capability).
PROVISIONED    everything else — including its own drivers for real work — after the kernel enumerates.
```

## Open

```text
· the minimal HANDSHAKE CHANNEL — co-located it is trivially local; remote, some born-in way to find
  the kernel. (the bootstrap primitive to pin next.)
· how the CTU AUTHENTICATES at handshake (which endpoint/principal it serves).
· durability / re-provisioning on restart.
· what an "endpoint" is at boot, and who assigns it.
```
