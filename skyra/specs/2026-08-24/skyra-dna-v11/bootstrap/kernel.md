# Kernel — bootstrap

**Date:** 2026-08-24
**Status: SCOPING — in progress.** What the kernel is at boot, and what seeds it. Builds on
`../ownership.md` and `../research/capability-table-and-horizon.md`.

---

## Role — the AUTHORITY, not a coordinator

The **root of authority inside the system.** It **authorizes**; it does **not** coordinate.
Coordination is *entity behavior* — the system's participants working together — and the kernel's only
role is to **provision the capabilities that make a given coordination authorized.** It enables; it
never directs.

```text
ANCHOR   seL4 microkernel — a capability AUTHORITY. user-space services do the coordination, holding
         capabilities. NOT a Kubernetes control plane (which schedules / orchestrates).
v7       "the kernel decides WHAT (the cone); a driver decides HOW."
```

It is **OS-independent** — it touches the world only through CTUs. Owns:

```text
IDENTITY               who the principal is; authenticates.
THE FULL TABLE         all capabilities (sysblock/host + entity) — registration.
MINT                   authority origination.
THE BARRIER            ingress/egress + the in-flight socket table.
SEPARATION HANDLERS    the trap surface; executes the human's policy.
```

## Boot order — the kernel is FIRST

A CTU can't execute without a provisioned horizon, and provisioning is the kernel's. So the kernel must
exist and be seeded **before** any CTU runs. Kernel first, CTUs second.

## What seeds the kernel — the human, out-of-band

Nothing *inside* the system authorizes the kernel — its authority is the **human's**, configured
out-of-band, before the system runs:

```text
SEED = the human's ROOT IDENTITY + the ROOT POLICY (risk taxonomy + tolerance).
```

Same shape as an OS trusting `init` / root because they were set up out-of-band. **The bottom of the
bootstrap is the human at configuration time.**

## Discovery — the handshake registers a CTU's drivers

When a CTU handshakes, the kernel **reads** its drivers (device-enumeration / PCI-like) and
**registers** them into the **sysblock** (host table). The kernel now knows what that CTU can do; the
CTU does not.

This handshake **IS acquire**, run by the kernel on the CTU (discover → test → admit) — so **acquire is
the universal onboarding primitive**: THINK acquires endpoints; the kernel acquires CTUs. The kernel
runs it with its own **seed authority**, not a provisioned horizon (no loop). See `sequence.md` C.

```text
driver MECHANISM   OS-specific, CTU-local. "the HOW."
driver REGISTRATION  the capability entry in the sysblock — kernel-owned. "that it exists, BELIEVED."
```

Mechanism vs authority, once more: the CTU carries the mechanism; the kernel owns the registration.

## After seeding

```text
· provisions CTUs — hands each move a horizon (a view of the table).
· handles traps — page faults (provision or deny), authority faults (ask_permission / mint).
· mints — registers capabilities into the full table.
```

## Open

```text
· what "seeding" concretely IS — the format of root identity + root policy; where it lives.
· ONE kernel vs ONE-PER-NODE — DECIDED: one per host (`sequence.md` B). the distributed kernel
  (kernel-to-kernel) is deferred — "the kernel is the distributed part of the system," later.
· how the kernel proves ITS identity to a CTU (mutual auth at handshake).
· the DISCOVERY / enumeration protocol — how the kernel reads a CTU's drivers.
```
