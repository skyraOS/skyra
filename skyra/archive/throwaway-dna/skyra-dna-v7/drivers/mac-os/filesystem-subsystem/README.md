# filesystem-subsystem

**Date:** 2026-08-20
**Status: THROWAWAY RUNNING.** The mechanism currently lives in the parent `driver.go`: one
atomically replaced JSON graph file, flushed before REMEMBER returns Success.

> **The filesystem subsystem inside the macOS driver. This is where a cone actually becomes
> durable, and where all the storage mechanism the kernel refused to know about lives.**

Given a name, it loads. Given a cone, it persists it atomically and reports when durable.
Everything here is internal to the driver and invisible above the shared abstraction:

```text
CONTENT-ADDRESSING   identity of a durable object. one proposal is refs-over-objects.
RETENTION MECHANISM  storing only what was handed down, collecting the rest. the kernel decided
                     WHAT (the cone); this decides HOW it is kept.
JOURNALING           the cone is many objects and must land all-or-nothing. a torn commit leaves
                     a dangling engram, which the model forbids.
BLOCK TRANSPORT      the actual reads and writes to the device.
```

**Design:** `skyra/specs/2026-08-18/kernel-ctu-boundary/durable-storage.md` for the boundary, and
`implementation.md` for one concrete way to fill this in (content-addressing, journaling). Both
are proposals — this subsystem is free to satisfy the interface however macOS makes easiest.
