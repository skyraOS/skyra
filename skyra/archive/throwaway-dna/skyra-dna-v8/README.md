# skyra-dna-v8 — present state (the distributed possibility)

**Date:** 2026-08-23
**Status: RUNNING, THROWAWAY.** A boundary-finding build for ONE thread: a possibility
**partitioned across CTUs**. The implementation trail is `boundary-crossings.md`. Design authority
is `skyra/specs/2026-08-22/` (on
the foundation of `skyra/specs/2026-08-18/`) — if code and docs disagree, the docs win.

## The one idea

The CTU **executes a possibility**. The possibility carries its own state, so "run it" needs no other
argument. When it hits a dependency it cannot resolve locally, the CTU **forwards the slice** that
needs the other CTU — it does not park a waiting parent. Execution roams by partitioning; a **MOVE is
the cut** between partitions; each CTU **signs** the provenance it produced; **REMEMBER gathers** the
cone home; **ORIENT** reads a distributed frontier.

## The packages (component layout, matching the spec dirs)

```
causal/    the types that cross boundaries and travel between CTUs (leaf package).
ctu/       Execute + the primitives; Interpreter; Scheduler (push/pull Selection).
kernel/    Activate ; Remember-as-gather ; Define.
drivers/   the host-table persistence half (near-unchanged).
entity/    decode + ALU (unchanged interior).
fabric/    transport: Approach (egress) + Fetch (gather hop) ; Custody (Sign/Verify).
```

## Reading order

```
1  README.md              (this file) — the one idea and the layout.
2  interface-rationale.md  every type, the term it comes from, and where boundaries cross.
3  causal/causal.go        the crossing types (PRIMARY / SECONDARY, cited to the docs inline).
4  the five package files  ctu ▸ kernel ▸ fabric ▸ drivers ▸ entity.
5  claude-ideas.md         the SHAPE is a set of Claude's proposals — attributed, rejectable.
```

## What is mine vs yours

The **terms** are all from the docs. The **shape** — which types exist, what they return, how the
packages split — is a set of Claude's proposals, every one attributed and rejectable in
**`claude-ideas.md`**. The three I'd check first: the cargo modes on `Move` (probably want a sum
type), `MoveID`-vs-`PossibilityID` on the wire (I took a side on a real doc seam), and where `Custody`
lives. Build/vet are clean; that proves the shapes fit together, not that they're right.

## What runs

One local transport connects two CTUs. A possibility PUSHes an attenuated Slice across a MOVE cut,
the receiver resumes it as a non-root child, signs the returned provenance Segment, and RECOGNIZEs
it home by MoveID. Root-only REMEMBER follows content-addressed Causes through Fabric.Fetch and
atomically persists the verified, deduplicated Cone. `boundary-crossings.md` records every interface
that changed and what remains unresolved. The overnight work order is `overnight.md`.
