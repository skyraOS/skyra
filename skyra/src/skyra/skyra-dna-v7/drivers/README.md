# drivers

**Date:** 2026-08-20
**Status: SCAFFOLD.** Nothing implemented.

> **A driver is a pluggable OS backend behind the shared abstraction. The shared abstraction is
> the two crossings — ACTIVATE and REMEMBER — and nothing else crosses.**

This is the driver-is-opaque pattern (`blk_mq_ops`) lifted above the filesystem: the whole OS,
its filesystem subsystem included, is the opaque thing. The kernel hands a driver a load or a
cone-to-store and does not look inside.

```text
GIVEN     ACTIVATE(name)  → a possibility.
          REMEMBER(cone)  → durable, "tell me when."
HIDES     content-addressing, retention mechanism, journaling, block transport — all internal.
ENABLES   different OS beneath one interface. swap the backend, keep the kernel.
```

Each driver carries its **own filesystem subsystem** (see `mac-os/filesystem-subsystem/`). The
kernel decides WHAT persists (the cone); the driver decides HOW.

**Design:** `skyra/specs/2026-08-18/kernel-ctu-boundary/durable-storage.md`.
