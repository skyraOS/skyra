# mac-os

**Date:** 2026-08-20
**Status: SCAFFOLD.** Nothing implemented.

> **A concrete driver — the shared abstraction implemented on macOS.**

The first backend. It satisfies the two crossings (ACTIVATE, REMEMBER) on this machine and hides
everything about how. It carries its own filesystem subsystem; see
`filesystem-subsystem/`.

```text
IMPLEMENTS   ACTIVATE / REMEMBER for macOS.
CONTAINS     filesystem-subsystem/ — where persistence actually happens.
IS ONE OF    many possible drivers. nothing above it depends on it being macOS.
```

**Design:** `skyra/specs/2026-08-18/kernel-ctu-boundary/durable-storage.md`.
