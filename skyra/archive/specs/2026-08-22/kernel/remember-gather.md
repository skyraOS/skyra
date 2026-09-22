# REMEMBER, distributed — gather the cone to the root

**Date:** 2026-08-22
**Status: WORKING.** REMEMBER's distributed form, the commit half of the distributed possibility
(`../ctu/possibility.md`). The kernel owns REMEMBER, the root, and durability
(`2026-08-18/kernel-ctu-boundary/durable-storage.md`, `.../boundaries.md`); this doc says only what
distribution changes — the WHERE of the walk. Root-only, the cone definition, and load/store are
unchanged.

> **Execution roams; commitment reconverges. When a possibility is partitioned across CTUs, its cone
> is too — so REMEMBER is a GATHER: walk `Causes` across CTU boundaries until it converges at the
> kernel that owns the root, and commit there.**

---

## Why commit is where the cost lands

Shipping a slice is cheap; the price of that is paid here. The cone that a kept observation depended
on may span every CTU the possibility visited. REMEMBER must reassemble it. This is the right trade:
REMEMBER is the rare path (root-only, durable, bulk), and forwarding a slice is the hot path
(`../ctu/possibility.md`, the cost table). Optimize the common case; pay at the uncommon one.

## The gather

```text
START   at the kept observation, on the root's kernel (REMEMBER is root-only: Parent == nil).
WALK    follow Causes references; each hop that names a segment on another CTU is a fetch.
STOP    at roots — a segment with no further Causes.
COMMIT  persist the assembled cone; the kernel decides WHAT (the cone), a driver decides HOW.
```

Content-addressing (`2026-08-18/kernel-ctu-boundary/implementation.md`) is what makes this tractable:
segments are hash-identified, so the walk **dedups** and **skips anything already durable** rather
than re-fetching it. The signed provenance segments the walk follows are `../fabric/signed-provenance.md`.

## Remote slices do not commit

A forwarded slice is not a root, so it cannot REMEMBER — structurally, not by rule (it holds no arena
it *is*; see the arena back-reference in `../ctu/possibility.md`). It produces observations that flow
home; the commit happens once, at the root. That is why the home `arenaID` never needs to travel: the
gather is initiated from home and pulls segments by content-address, and the remote CTU serves a
signed segment by its hash without needing to know the arena.

---

## Open

```text
GATHER FAILURE MODES   a segment on a dead CTU mid-gather; the re-convergence cost of a wide cone;
                       whether the walk is pull-from-root or push-on-completion. sketched, not specified.
REMOTE SLICE'S ARENA   does a forwarded slice carry the home arenaID as an INERT TAG (for gather
                       association) or get a purely LOCAL scope? lean: local scope, and let the gather
                       associate by content-address + cut-edge lineage, so the home arenaID never
                       travels. leans on this walk.
LOST-PARTITION REPLAY  the signed Causes chain is a lineage, so a lost segment can be RE-EXECUTED from
                       its inputs (../fabric/signed-provenance.md). how replay interacts with a
                       gather already in progress is unexplored.
```
