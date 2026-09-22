# ctu

**Date:** 2026-08-20
**Status: THROWAWAY RUNNING.** See `machine.go` and `machine_test.go`.

> **The CTU executes primitives over causal state. One per endpoint. It is not the kernel and it
> is not the fabric.**

It owns the **live possibility tree**, resolves its own payloads, runs its own bodies, and issues
MOVEs. The driver normalizes local and external ingress before the CTU admits a REQUEST through
GREET or a RESPONSE through RECOGNIZE. Issuing a MOVE is **unprivileged**. The CTU traps to the
kernel at exactly two points — ACTIVATE and REMEMBER — and is otherwise on its own.

```text
OWNS      the live possibility tree; the execution registers of whatever runs now.
DOES      execute. resolve payloads, run bodies, admit work and returns, walk the tree.
DOES NOT  own durable storage, name-meaning, or lifetime. those are the kernel's.
```

**Design:** `skyra/specs/2026-08-18/primitive-decomposition/causal-transition-unit/` and
`skyra/specs/2026-08-18/kernel-ctu-boundary/boundaries.md`.
