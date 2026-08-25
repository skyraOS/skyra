# Ownership — human, kernel, CTU

**Date:** 2026-08-24
**Status: WORKING MODEL.** The first-pass answer to the open "who owns privilege" question. Resolves
the split three ways; the remaining edges (below) stay open. Builds on `skyra-dna-v10`'s D3 (principal
vs mechanism), D5 (privilege is the trap axis), and `potentials/user-as-privilege-boundary.md`.

---

## The split

```text
HUMAN    OWNS POLICY. the risk tolerance and the criteria for granting. authors it; is the escalation
         target when policy says "ask." never executes anything.
KERNEL   EXECUTES the human's policy (mechanically — system-administered). owns the PRIVILEGED
         SEPARATION HANDLERS (the trap surface). performs the privileged OPERATIONS — MINT / writes —
         over the CTU registries. NEUTRAL: it holds no policy of its own; it only executes the human's.
CTU      OWNS EXECUTION. Drive, dispatch, the moves. holds the registries and READS them to dispatch.
         RAISES a trap on a fault; never decides, never writes a registry, never mints.
```

One line: **the human owns the policy; the kernel executes it and owns the privileged handlers +
registry writes; the CTU owns execution and raises the trap.**

---

## How they meet

```text
CTU → KERNEL    a move faults (unadmitted name). the CTU marks the trap and PARKS. it does not decide.
KERNEL ← HUMAN  the kernel's separation handler EXECUTES the human's policy: known → auto-grant,
                novel-beyond-tolerance → wake the human. THINK / the model never decide.
KERNEL → CTU    on grant, the kernel performs the privileged operation (MINT) over the CTU registry;
                control returns; the CTU re-drives and READS the now-updated registry.
```

The registry is **CTU-resident state**, but its **write is the kernel's alone**. The CTU reads it to
dispatch (execution); the kernel writes it to admit (origination). Same read/write split as everywhere.

---

## Why this holds the existing decisions

```text
D3 (principal vs mechanism)   refined: the PRINCIPAL is the HUMAN (owns policy); the MECHANISM is the
                              KERNEL (executes policy + mints). the split still stands — the human's
                              policy is never baked into the kernel; the kernel executes it neutrally.
D5 (privilege = trap axis)    the kernel OWNS the trap / separation handlers. the trap axis IS the
                              kernel's handler surface — not a peer block.
READ vs WRITE                 CTU reads registries (execution); kernel writes them (privileged mint).
```

The Skyra inversion worth stating: an OS bakes policy into the kernel. **Skyra hands policy to the
human and the kernel merely executes it.** That is the North Star — the human as world-shaping
participant — made mechanical: the kernel is the human's neutral executor, not its own authority.

---

## Still open

```text
POLICY RESIDENCY   where the human's policy physically lives, and how it is configured and consulted
                   (the risk taxonomy + tolerance config). the kernel executes it — but reads it from where?
DIRECT vs AGENT    does the kernel execute the human's policy directly, or consult a resident user-root
                   possibility that holds it? (kernel-with-policy vs a principal the kernel serves.)
MULTI-USER         whose policy when there are several principals? how does the kernel pick, and can
                   authority be delegated? (D3 anticipated this.)
REGISTRY CROSS     the registry is CTU-resident but kernel-written — the exact mechanism by which a
                   kernel operation writes CTU state (and whether that is one CTU or many).
```
