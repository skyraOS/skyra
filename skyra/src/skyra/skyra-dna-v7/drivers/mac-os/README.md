# mac-os

**Date:** 2026-08-20
**Status: THROWAWAY RUNNING.** Atomic file persistence and local APPROACH are implemented.

> **A concrete driver — the OS-facing abstraction implemented on macOS.**

The first backend. It supplies the host traffic paths and satisfies the two durable crossings
(ACTIVATE, REMEMBER) on this machine. It carries its own port and filesystem subsystems.

```text
IMPLEMENTS   ACTIVATE / REMEMBER for macOS; traffic surface still to be decomposed.
CONTAINS     port-subsystem/       — APPROACH egress, routing, local/external ingress.
             filesystem-subsystem/ — where persistence actually happens.
IS ONE OF    many possible drivers. nothing above it depends on it being macOS.
```

**Design:** `skyra/specs/2026-08-18/network/traffic-routing.md` and
`skyra/specs/2026-08-18/kernel-ctu-boundary/durable-storage.md`.
