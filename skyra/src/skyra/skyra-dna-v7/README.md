# skyra-dna-v7

**Date:** 2026-08-20
**Status: SCAFFOLD.** Structure only — nothing here is implemented. v7 is the first build of the
CTU pivot, and it is **not a revision of v6.** v6 (`../cognitive-agent/skyra-dna-v6/`) is a
pre-pivot snapshot and stays where it is.

> **The design is upstream. This tree is where it gets built, and it points back at the design
> rather than restating it.**

The authoritative design is `skyra/specs/2026-08-18/`, and in particular
`skyra/specs/2026-08-18/kernel-ctu-boundary/`. If code here disagrees with those documents, the
documents win until they are revised.

---

## The three parts

```text
ctu/       EXECUTES. one per endpoint. owns the live possibility tree, issues MOVEs, runs bodies.
           unprivileged. traps to the kernel only at the two crossings.

kernel/    owns the DURABLE GRAPH and the root/lifetime structure. privileged. meets the CTU at
           exactly two crossings — ACTIVATE (load) and REMEMBER (store) — and nowhere else.

drivers/   pluggable OS backends behind the shared abstraction (the two crossings). the kernel
           delegates persistence here. each driver carries its own filesystem subsystem.
```

The one contract between the first two: **exactly two crossings.** See
`kernel-ctu-boundary/boundaries.md`.
