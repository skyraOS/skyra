# The parent owns the return boundary

**Date:** 2026-08-13  
**Status: DECIDED.**

> **The child owns the observation it returns. The parent owns the boundary that decides whether
> that observation can continue the parent's program.**

For a child position expecting one type, the boundary is generic string matching:

```text
expected type == returned type    → 1
expected type != returned type    → 0
```

The expected type belongs to the parent-owned child contract. The returned type belongs to the
child observation. This comparison is `Recognize`.

## Protocol ownership and composition

> **The kernel owns the `Recognize` protocol. Entities and engrams compose under it.**

`Recognize` is a kernel method because the kernel owns delivery, ordering, the return boundary, and
the recognition protocol. Inside that method, the kernel calls the relevant parent entity and
engram behavior. Those compositions interpret the observation; the kernel does not acquire their
semantics merely because it owns the protocol.

```text
child returns observation
        ↓
kernel delivers it to the parent's return boundary
        ↓
kernel.Recognize(parent contract, returned observation)
        ↓
Recognize calls entity and engram methods
        ↓
their composition produces match or mismatch
        ↓
kernel records the result and preserves its causality
```

```text
1    bind the observation, Remember it locally, and continue.

0    the parent fails to consume the return. Wake the parent at that child position so its entity
     can reconcile the mismatch.
```

The child does not own the type boundary and does not receive the failure. The kernel owns and runs
the recognition protocol; entities and engrams supply the semantic work within it. The kernel
delivers the returned observation with its position and provenance, witnesses the parent's
subsequent behavior, and preserves the resulting causality.

> **A type mismatch is the parent's failure, not the child's. The parent sees it and reconciles.**
