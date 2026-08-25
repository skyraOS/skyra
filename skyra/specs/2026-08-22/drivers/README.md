# drivers/ — (no new material in this dated spec)

**Status: PLACEHOLDER.** The distributed-possibility pass (2026-08-22) added nothing to the driver
boundary. The live driver design remains `2026-08-18/kernel-ctu-boundary/durable-storage.md` (the two
crossings as the shared driver abstraction). The one driver-adjacent consequence — that a persistence
driver serves a signed provenance segment by its content-address during a gather — lives with the
mechanism that uses it, `../kernel/remember-gather.md`.
