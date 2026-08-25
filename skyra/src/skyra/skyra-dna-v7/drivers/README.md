# drivers

**Date:** 2026-08-20
**Status: THROWAWAY RUNNING.** The macOS backend implements durable storage and local transport.

> **A driver is a pluggable OS backend. It owns host traffic and persistence mechanisms without
> leaking either into the machine above it.**

The traffic side has one logical egress and two physical ingress paths. The filesystem side
implements ACTIVATE and REMEMBER. Host sockets, descriptors, routes, content addressing, and
block transport remain opaque above the driver.

```text
TRAFFIC   APPROACH socket            one logical egress submission point.
          local ingress         traffic routed within this host.
          external ingress      traffic arriving from another host.
          both ingress paths normalize before CTU-level GREET / RECOGNIZE.
STORAGE   ACTIVATE(name)  → a possibility.
          REMEMBER(cone) → durable, "tell me when."
HIDES     sockets, routes, queues; content-addressing, retention, journaling, block transport.
ENABLES   different OS beneath one interface. swap the backend, keep the kernel.
```

Each driver carries its own **port subsystem** and **filesystem subsystem**. The kernel decides
WHAT persists (the cone); the driver decides HOW.

**Design:** `skyra/specs/2026-08-18/network/traffic-routing.md` and
`skyra/specs/2026-08-18/kernel-ctu-boundary/durable-storage.md`.
