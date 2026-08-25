# The quadrant — system/user × privilege

**Date:** 2026-08-21
**Status: WORKING, and foundational.** The quadrant STRUCTURE is settled and is the frame the rest
of the set instantiates. Terminology at the edges (the entity/driver/kernel umbrella) is still
open. Converged over conversation on 2026-08-21.

> **Two axes, not one. TRUST is the system/user boundary. PRIVILEGE is the durable/authority
> boundary. They are independent, and unbundling them is what the whole design turns on.**

Real operating systems weld trust and privilege into a single ring boundary — user code is both
untrusted AND unprivileged, kernel code is both trusted AND privileged. This design separates
them, and almost everything else follows from the separation.

---

## The two axes

```text
TRUST      SYSTEM SPACE vs USER SPACE.  is it the machine, or does it run ON the machine?
                                        the firmest line in OS design.
PRIVILEGE  PRIVILEGED vs UNPRIVILEGED.  may executing it cause a durable / authority write?
```

---

## The quadrant

```text
                              ▲  PRIVILEGED
                              │
        SYSTEM SPACE          │          USER SPACE
   ┌───────────────────────┐  │  ┌────────────────────────────┐
   │ KERNEL / root         │  │  │ ACCOUNT / user             │
   │ ring 0. EXECUTES       │  │  │ holds a pin-minted token.  │
   │ durable writes,        │  │  │ AUTHORIZES writes to its   │
   │ system-wide. the TCB.  │  │  │ own graph.                 │
   └───────────────────────┘  │  └────────────────────────────┘
 SYSTEM ─────────────────────●──────────────────────────────── USER ▶
   ┌───────────────────────┐  │  ┌────────────────────────────┐
   │ CTU + trusted drivers │  │  │ ENTITIES                   │
   │ the dispatcher +      │  │  │ foreign code. output       │
   │ unprivileged hardware │  │  │ stamped + checked.         │
   └───────────────────────┘  │  └────────────────────────────┘
                              │
                              ▼  UNPRIVILEGED
```

```text
TOP-LEFT      KERNEL / root   — system + privileged. INTRINSIC privilege. executes durable writes
                               system-wide. it is the trusted computing base.
BOTTOM-LEFT   CTU + drivers   — system + unprivileged. the dispatcher and the trusted hardware it
                               runs directly. traps up for durable I/O.
TOP-RIGHT     ACCOUNT / user  — user + privileged. AUTHENTICATED, SELF-SCOPED privilege. holds a
                               pin-minted root token; authorizes writes to its OWN graph, which the
                               kernel executes. this is a logged-in user, not a violation.
BOTTOM-RIGHT  ENTITIES        — user + unprivileged. foreign execution hardware; output validated.
```

The whole trust line runs **through every endpoint**: the entity (user) and its CTU (system) sit
at the same address, separated by it — exactly like a process and the kernel co-resident on one
box.

---

## The two axes have different natures

This is what keeps the quadrant firm, and it is the correction that took the longest to reach:

```text
TRUST      is INTRINSIC — WHAT YOU ARE. in the TCB or not. you cannot gain it.
PRIVILEGE  is HELD      — WHAT YOU HOLD. a capability. you can gain it (the pin) or drop it (narrow).
```

So **authentication grants privilege, not trust.** The pin is a vertical elevator on the USER side:
it lifts a user from bottom-right to top-right (scoped privilege) and NEVER moves them left. An
authenticated account is still user space — its output is still stamped and checked. Identity is
what the pin establishes; trustworthiness of code is not for sale.

---

## The two lines are the two crossings

```text
THE TRUST LINE (vertical)     system ↔ user. crossing it is ADMISSION + VALIDATION — GREET /
                              RECOGNIZE, then stamp + contract-check. see entity-ctu-boundary/.

THE PRIVILEGE LINE (horizontal) unprivileged ↔ privileged, WITHIN system space. crossing it is the
                              durable TRAP — ACTIVATE / REMEMBER, the two crossings. see
                              kernel-ctu-boundary/boundaries.md.
```

Because the axes are unbundled, the classic syscall (which crosses both at once) **splits into two
independent crossings.**

---

## A user-initiated durable write takes the L — three cells, two crossings, no shortcut

```text
ACCOUNT (top-right, user)     ①  authorizes REMEMBER, presents its token
     │  ① cross the TRUST line   →  the CTU admits and validates the token
     ▼
CTU (bottom-left, system)     ②  traps up for the durable write
     │  ② cross the PRIVILEGE line →  the kernel verifies scope and executes
     ▼
KERNEL (top-left, system)         performs the write into the account's OWN graph
```

The account **authorizes**, the system **admits and validates**, the kernel **executes.** The
token links top-right to top-left; it is presented by the user and verified by the kernel — the
same shape as `write(2)`.

---

## The one firm rule: there is no diagonal

```text
FORBIDDEN   USER space → PRIVILEGED SYSTEM space, directly. that is user code in ring 0 —
            privilege escalation, the thing every OS exists to prevent.
REQUIRED    a user reaching privilege takes the L: cross into the system (validated), then the
            system crosses to privilege. the token AUTHORIZES a write; it never TELEPORTS the user
            into the kernel.
```

And it is **inexpressible**, not merely forbidden, for reasons already built elsewhere in the set:

```text
TOKENS NARROW ONLY   you attenuate a root token DOWN (drop its durable authority, top-right →
                     bottom-right), never climb up.
ONLY THE PIN MINTS   a privileged root token comes from authentication and nowhere else — and it is
                     self-scoped, never system-wide. no token grants the KERNEL's kind of privilege.
```

---

## How the rest of the set instantiates this

This document is the hub; the others are its edges.

```text
dispatch.md                          THE TRUST AXIS, MADE CONCRETE. two lookup tables (host = trusted
                                     hardware, entity = untrusted) and the kernel NAMED, not flagged.
kernel-ctu-boundary/boundaries.md    THE PRIVILEGE LINE. the two crossings, load/store, and BOUND's
                                     DEFINE/ATTENUATE split (an authority-tier write).
entity-ctu-boundary/decode-and-run.md THE TRUST LINE, interior. how a user entity's output is
                                     admitted and validated into system space.
network/traffic-routing.md           THE TRUST LINE, transport. how envelopes reach GREET/RECOGNIZE.
kernel-ctu-boundary/horizon-and-root.md  THE USER COLUMN + THE ELEVATOR. the pin, the account, the
                                     token that is the horizon, and why narrowing only drops.
```

---

## Open

```text
TERMINOLOGY        the umbrella word for "anything the CTU dispatches to." working split: ENTITIES
                   = umbrella; DRIVERS = the trusted (system) column; KERNEL = the privileged
                   drivers; foreign ENTITIES = the user column. Not yet fixed.

DRIVER PRIVILEGE   some drivers are privileged (storage → they are the kernel), some are not
                   (compute → trusted helpers). exactly where a given driver sits on the vertical
                   axis is per-driver and not yet enumerated.

ACCOUNT SHAPE      whether the account IS a root entity holding a token, or a distinct principal
                   object the root entity carries. §top-right treats them as the same; unconfirmed.

CROSS-ACCOUNT      one account acting on another's graph (delegation) is a user→user question the
                   quadrant does not yet address. deferred with third-party caveats in
                   horizon-and-root.md.
```
