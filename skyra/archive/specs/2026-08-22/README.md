# 2026-08-22 — the distributed possibility, by component

**Status: WORKING.** A component-organized dated spec. Where `2026-08-18/` is organized by *boundary*
(kernel-ctu, entity-ctu) and lays the foundation, this set is organized by *component* (the same split
as the v7 code: ctu, kernel, drivers, entity, causal, fabric) and holds one thread: the **distributed
possibility** — how a possibility crosses hosts. `2026-08-18/` remains the foundation and the current
set for everything else; read it first.

> **The CTU executes a possibility. When it hits a dependency it cannot resolve locally, it forwards
> the slice that needs the other CTU — not a request that waits. Execution roams by partitioning; a
> MOVE is the cut between partitions. Everything else — commit, reflection, trust — is a read over
> that one partition along the same edges.**

---

## Read in this order

```text
1  ctu/possibility.md          THE SPINE. possibility as the unit of execution; ship the slice, not
                               the whole; the possibility tree across hosts; fan-out/join; the arena
                               back-reference; the three distributed reads.
2  ctu/orient.md               the THIRD read: ORIENT as the distributed frontier. termination
                               detection over a diffusing computation; the escalation lead.
3  kernel/remember-gather.md   the COMMIT: REMEMBER as a gather of the cone to the root's kernel.
4  fabric/signed-provenance.md the WIRE: a cut is a MOVE; each CTU signs its provenance segment;
                               unforgeable across a trust boundary (the punch-list's P1.2).
5  causal/terms.md             the shared vocabulary. read when a term is unfamiliar.
   drivers/README.md           placeholder — no new material.
   entity/README.md            placeholder — no new material.
```

## Relationship to 2026-08-18

```text
FOUNDATION      the-quadrant, the eight primitives, the two boundaries, dispatch, traffic-routing,
                horizon-and-root — all in 2026-08-18/. this set does not restate them; it cites them.
RESOLVES        2026-08-18/entity-ctu-boundary/decode-and-run.md's "LOCAL vs DISTRIBUTED — THE LARGEST
                QUESTION", in one direction: partition, forward the slice. (ctu/possibility.md)
SUPERSEDES      2026-08-18/primitive-decomposition/distributed-possibility.md — the single-file draft
                of this thread, now split across the components that own each piece.
LEAD ON         2026-08-18/README.md's ESCALATION gap, from the mechanism side. (ctu/orient.md)
```

## House rules

Inherited from `2026-08-18/README.md` unchanged: status on every doc; record contradictions and name
the other document; end every doc with what is open; derive, don't declare; prefer inexpressible to
forbidden; argue the other side.
