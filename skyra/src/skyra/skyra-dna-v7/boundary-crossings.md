# Overnight build — execution snapshot

**Date:** 2026-08-22
**Status: THROWAWAY IMPLEMENTATION NOTES.** What the overnight run actually executed — the build's
status, next to the code it implemented. The **new information** it surfaced (the interfaces that
did not hold, the documents that collide, the concepts that needed a home, the unresolved items)
was split out to `changes/boundary-crossings.md`. The authority remains `skyra/specs/2026-08-18/`.

## What runs

One synchronous local path now executes:

```text
parent MOVE(request)
  → APPROACH → normalized arrival → GREET
  → kernel ACTIVATE(name, horizon)
  → CTU INTERPRET → SCHEDULE
  → untrusted entity Run(payload) → Content
  → CTU stamps Observation
  → APPROACH → normalized arrival → RECOGNIZE(parent contract)
  → correlated Result
```

The macOS driver also performs the other two privileged operations: DEFINE authenticates a Pin,
and root-only REMEMBER computes the provenance cone and returns Success only after an atomic file
replacement has been flushed.
