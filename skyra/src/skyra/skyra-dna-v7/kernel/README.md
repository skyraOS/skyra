# kernel

**Date:** 2026-08-20
**Status: THROWAWAY RUNNING.** See `system.go` and `system_test.go`.

> **The kernel owns the durable graph and the root/lifetime structure. It meets the CTU at
> exactly two crossings, and configures-and-exits everywhere else.**

It is not interposed on execution. It establishes what is permitted, steps out, and is re-entered
only at the two crossings:

```text
ACTIVATE   LOAD.   durable → live. reads durable storage, hands back a possibility, steps out.
REMEMBER   STORE.  live → durable, at ROOT. commits the cone of a contribution, the arena tears
                   down, steps out.
```

```text
OWNS      the durable graph · roots and lifetime · name → meaning (the namespace).
COMPUTES  the cone — reachability from a root. WHAT persists. it cannot delegate this.
DELEGATES persistence. HOW the cone is stored is a driver's problem. see ../drivers/.
```

**Design:** `skyra/specs/2026-08-18/kernel-ctu-boundary/boundaries.md` and `durable-storage.md`.
